package proxy

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"prismproxy/internal/storage"
)

// handleConnect 处理 HTTPS CONNECT 请求（MITM 模式）
func (p *Proxy) handleConnect(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] 处理 CONNECT: %s", r.URL.Host)

	// 劫持连接
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "不支持连接劫持", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// 发送连接建立响应
	_, err = clientConn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
	if err != nil {
		log.Printf("[ERROR] 发送连接建立响应失败: %v", err)
		clientConn.Close()
		return
	}

	// 获取主机名
	host, _, _ := net.SplitHostPort(r.URL.Host)

	// 生成动态证书
	siteCert, err := p.getCert(host)
	if err != nil {
		log.Printf("[ERROR] 生成证书失败 %s: %v", host, err)
		clientConn.Close()
		return
	}

	// 建立 TLS 连接（服务端模拟目标站点）
	// 注意: 这里是调试工具的核心功能，用于抓取 HTTPS 流量
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*siteCert},
		NextProtos:   []string{"http/1.1"},
	}
	clientTlsConn := tls.Server(clientConn, tlsConfig)
	if err := clientTlsConn.Handshake(); err != nil {
		log.Printf("[ERROR] TLS 握手失败: %v", err)
		clientTlsConn.Close()
		return
	}
	defer clientTlsConn.Close()

	// 检查协议
	state := clientTlsConn.ConnectionState()
	if state.NegotiatedProtocol != "http/1.1" && state.NegotiatedProtocol != "" {
		log.Printf("[WARN] 客户端协商了不支持的协议 '%s'，使用原始隧道模式", state.NegotiatedProtocol)
		p.tunnel(clientTlsConn, r.URL.Host)
		return
	}

	log.Printf("[INFO] 成功建立 MITM 连接: %s", r.URL.Host)

	// 循环处理请求
	reader := bufio.NewReader(clientTlsConn)
	for {
		startTime := time.Now()
		clientTlsConn.SetReadDeadline(time.Now().Add(5 * time.Minute))

		req, err := http.ReadRequest(reader)
		if err != nil {
			if err != io.EOF {
				log.Printf("[WARN] 读取客户端请求失败: %v", err)
			}
			return
		}

		// 读取请求体
		reqBody, err := io.ReadAll(req.Body)
		if err != nil {
			log.Printf("[ERROR] 读取 HTTPS 请求体失败: %v", err)
			return
		}
		req.Body = io.NopCloser(bytes.NewBuffer(reqBody))

		// 修正请求 URL
		req.URL.Scheme = "https"
		req.URL.Host = req.Host

		// 转储原始请求
		rawReq, _ := httputil.DumpRequestOut(req, true)

		// 连接到目标服务器
		// 注意: InsecureSkipVerify 仅用于调试工具，生产环境应配置正确的证书验证
		destConn, err := tls.Dial("tcp", r.URL.Host, &tls.Config{
			InsecureSkipVerify: true, // 调试工具专用，允许抓取任意 HTTPS 流量
		})
		if err != nil {
			log.Printf("[ERROR] 连接目标服务器失败 %s: %v", r.URL.Host, err)
			return
		}

		// 获取服务器 IP
		serverIP := ""
		if destConn.RemoteAddr() != nil {
			serverIP, _, _ = net.SplitHostPort(destConn.RemoteAddr().String())
		}

		// 发送请求到目标服务器
		if err := req.Write(destConn); err != nil {
			log.Printf("[ERROR] 写入请求到目标服务器失败: %v", err)
			destConn.Close()
			return
		}

		// 检查 WebSocket 升级
		if strings.ToLower(req.Header.Get("Upgrade")) == "websocket" {
			log.Printf("[INFO] 检测到 WebSocket 升级，切换到原始隧道模式")
			go io.Copy(destConn, clientTlsConn)
			io.Copy(clientTlsConn, destConn)
			return
		}

		// 读取响应
		resp, err := http.ReadResponse(bufio.NewReader(destConn), req)
		if err != nil {
			log.Printf("[ERROR] 读取目标服务器响应失败: %v", err)
			destConn.Close()
			return
		}

		// 读取响应体
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("[ERROR] 读取 HTTPS 响应体失败: %v", err)
			return
		}
		resp.Body = io.NopCloser(bytes.NewBuffer(respBody))

		// 转储原始响应
		rawResp, _ := httputil.DumpResponse(resp, true)

		duration := time.Since(startTime).Milliseconds()
		log.Printf("[HTTPS] 拦截响应: %s (耗时 %dms)", resp.Status, duration)

		// 保存抓包数据
		p.saveHTTPSTraffic(req, resp, serverIP, reqBody, respBody, rawReq, rawResp, duration)

		// 返回响应给客户端
		if err := resp.Write(clientTlsConn); err != nil {
			log.Printf("[ERROR] 写入响应到客户端失败: %v", err)
		}
		destConn.Close()
	}
}

// saveHTTPSTraffic 保存 HTTPS 抓包数据
func (p *Proxy) saveHTTPSTraffic(req *http.Request, resp *http.Response, serverIP string,
	reqBody, respBody, rawReq, rawResp []byte, durationMs int64) {

	reqHeaders, _ := json.Marshal(req.Header)
	respHeaders, _ := json.Marshal(resp.Header)

	data := &storage.TrafficData{
		Method:      req.Method,
		URL:         req.URL.String(),
		Host:        req.Host,
		ServerIP:    serverIP,
		ContentType: resp.Header.Get("Content-Type"),
		StatusCode:  resp.StatusCode,
		DurationMs:  durationMs,
		ReqHeaders:  string(reqHeaders),
		ReqBody:     reqBody,
		RespHeaders: string(respHeaders),
		RespBody:    respBody,
		RawReq:      rawReq,
		RawResp:     rawResp,
	}

	if err := p.storage.SaveTraffic(data); err != nil {
		log.Printf("[ERROR] 保存 HTTPS 抓包数据失败: %v", err)
	}
}

// tunnel 原始隧道模式（用于不支持的协议）
func (p *Proxy) tunnel(clientConn net.Conn, host string) {
	destConn, err := net.Dial("tcp", host)
	if err != nil {
		log.Printf("[ERROR] 隧道模式连接目标失败: %v", err)
		return
	}
	go io.Copy(destConn, clientConn)
	io.Copy(clientConn, destConn)
}
