package proxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

// Config 代理配置
type Config struct {
	ListenAddr  string
	Port        int
	EnableHTTPS bool
	EnableMITM  bool
}

// Server HTTP 代理服务器
type Server struct {
	config    Config
	server    *http.Server
	running   bool
	mu        sync.RWMutex
	trafficCh chan<- *TrafficEvent
}

// TrafficEvent 流量事件
type TrafficEvent struct {
	Method    string
	URL       string
	Host      string
	Status    int
	Duration  time.Duration
	Size      int64
	Timestamp time.Time
}

// NewServer 创建代理服务器
func NewServer(config Config) *Server {
	return &Server{
		config: config,
	}
}

// SetTrafficChannel 设置流量事件通道
func (s *Server) SetTrafficChannel(ch chan<- *TrafficEvent) {
	s.trafficCh = ch
}

// Start 启动代理服务器
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("代理已在运行中")
	}

	addr := fmt.Sprintf("%s:%d", s.config.ListenAddr, s.config.Port)

	s.server = &http.Server{
		Addr:    addr,
		Handler: s,
	}

	// 启动 HTTP 代理
	go func() {
		log.Printf("[INFO] HTTP 代理启动，监听 %s", addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[ERROR] 代理服务器错误: %v", err)
		}
	}()

	s.running = true
	return nil
}

// Stop 停止代理服务器
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return fmt.Errorf("代理未在运行")
	}

	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.server.Shutdown(ctx); err != nil {
			log.Printf("[WARN] 代理关闭错误: %v", err)
		}
	}

	s.running = false
	log.Println("[INFO] HTTP 代理已停止")
	return nil
}

// IsRunning 检查是否运行中
func (s *Server) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetAddr 获取监听地址
func (s *Server) GetAddr() string {
	return fmt.Sprintf("%s:%d", s.config.ListenAddr, s.config.Port)
}

// ServeHTTP 实现 http.Handler 接口
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// CONNECT 方法（HTTPS 代理）
	if r.Method == http.MethodConnect {
		s.handleConnect(w, r)
		return
	}

	// 普通 HTTP 代理
	s.handleHTTP(w, r, start)
}

// handleHTTP 处理 HTTP 请求
func (s *Server) handleHTTP(w http.ResponseWriter, r *http.Request, start time.Time) {
	// 转发请求
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	// 移除代理相关头
	r.RequestURI = ""
	r.Header.Del("Proxy-Connection")
	r.Header.Del("Proxy-Authenticate")
	r.Header.Del("Proxy-Authorization")

	resp, err := client.Do(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		s.sendTrafficEvent(r, 502, time.Since(start), 0)
		return
	}
	defer resp.Body.Close()

	// 复制响应头
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)

	// 复制响应体
	size, _ := io.Copy(w, resp.Body)

	// 发送流量事件
	s.sendTrafficEvent(r, resp.StatusCode, time.Since(start), size)
}

// handleConnect 处理 CONNECT 方法（HTTPS 隧道）
func (s *Server) handleConnect(w http.ResponseWriter, r *http.Request) {
	// 获取目标地址
	targetAddr := r.URL.Host
	if _, _, err := net.SplitHostPort(targetAddr); err != nil {
		targetAddr = net.JoinHostPort(targetAddr, "443")
	}

	// 建立到目标的连接
	targetConn, err := net.DialTimeout("tcp", targetAddr, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer targetConn.Close()

	// 告诉客户端连接已建立
	w.WriteHeader(http.StatusOK)

	// 获取底层连接
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer clientConn.Close()

	// 双向转发
	done := make(chan struct{})
	go func() {
		io.Copy(targetConn, clientConn)
		close(done)
	}()
	io.Copy(clientConn, targetConn)
	<-done
}

// sendTrafficEvent 发送流量事件
func (s *Server) sendTrafficEvent(r *http.Request, status int, duration time.Duration, size int64) {
	if s.trafficCh == nil {
		return
	}

	event := &TrafficEvent{
		Method:    r.Method,
		URL:       r.URL.String(),
		Host:      r.Host,
		Status:    status,
		Duration:  duration,
		Size:      size,
		Timestamp: time.Now(),
	}

	// 非阻塞发送
	select {
	case s.trafficCh <- event:
	default:
		// 通道满，丢弃事件
	}
}
