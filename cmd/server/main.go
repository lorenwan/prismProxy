package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/improbable-eng/grpc-web/go/grpcweb"

	"prismproxy/internal/ai"
	"prismproxy/internal/cert"
	"prismproxy/internal/codegen"
	"prismproxy/internal/collection"
	"prismproxy/internal/debugger"
	"prismproxy/internal/diff"
	"prismproxy/internal/environment"
	"prismproxy/internal/grpc"
	"prismproxy/internal/perf"
	"prismproxy/internal/proxy"
	"prismproxy/internal/rewrite"
	"prismproxy/internal/rules"
	"prismproxy/internal/script"
	"prismproxy/internal/search"
	"prismproxy/internal/storage"
	"prismproxy/internal/traffic"
	"prismproxy/internal/websocket"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("[INFO] PrismProxy gRPC 服务器启动中...")

	// 初始化 SQLite 存储
	dbPath := getEnv("DB_PATH", "./prismproxy.db")
	store, err := storage.NewStorage(dbPath)
	if err != nil {
		log.Fatalf("[FATAL] 初始化存储失败: %v", err)
	}
	defer store.Close()

	// 执行数据库迁移
	if err := store.RunMigrations(); err != nil {
		log.Fatalf("[FATAL] 数据库迁移失败: %v", err)
	}

	// 初始化 WebSocket Hub
	hub := websocket.NewHub()
	go hub.Run()

	// 初始化各模块
	trafficMgr := traffic.NewManager(store, hub)
	rulesEngine := rules.NewEngine(store.DB)
	debuggerMgr := debugger.NewDebugger(store.DB, hub)
	collectionStore := collection.NewStore(store.DB)
	collectionMgr := collection.NewManager(collectionStore)
	collectionRunner := collection.NewRunner()
	envStore := environment.NewStore(store.DB)
	envMgr := environment.NewManager(envStore)
	rewriteEngine := rewrite.NewEngine(store.DB)
	codegenGen := codegen.NewGenerator()
	aiSvc := ai.NewService(&ai.Config{})

	// 初始化新模块
	scriptStore := script.NewScriptStore(store.DB)
	scriptStore.Init()
	scriptEngine := script.NewEngine()
	diffEngine := diff.NewEngine()
	perfAnalyzer := perf.NewAnalyzer(store.DB)
	certStore := cert.NewCertStore()
	certManager := cert.NewCertManager(certStore)
	searchEngine := search.NewSearchEngine(store.DB)
	filterStore := search.NewFilterStore(store.DB)
	filterStore.Init()

	// 初始化代理服务器（默认不启动）
	proxyPort := getEnvInt("PROXY_PORT", 8888)
	proxyServer := proxy.NewServer(proxy.Config{
		ListenAddr: "0.0.0.0",
		Port:       proxyPort,
	})

	// 创建系统代理管理器
	systemProxy := proxy.NewSystemProxy(fmt.Sprintf("0.0.0.0:%d", proxyPort))

	// 创建代理控制器
	proxyCtrl := &grpc.ProxyController{
		StartFunc: func() error {
			return proxyServer.Start()
		},
		StopFunc: func() error {
			return proxyServer.Stop()
		},
		StatusFunc: func() (bool, string) {
			return proxyServer.IsRunning(), proxyServer.GetAddr()
		},
	}

	// 创建 gRPC 服务器
	grpcPort := getEnvInt("GRPC_PORT", 9090)
	srv, err := grpc.NewServer(
		grpc.ServerConfig{Port: grpcPort},
		store,
		trafficMgr,
		rulesEngine,
		debuggerMgr,
		collectionMgr,
		collectionRunner,
		envMgr,
		rewriteEngine,
		aiSvc,
		codegenGen,
		scriptStore,
		scriptEngine,
		diffEngine,
		perfAnalyzer,
		certManager,
		certStore,
		searchEngine,
		filterStore,
		proxyCtrl,
		systemProxy,
	)
	if err != nil {
		log.Fatalf("[FATAL] 创建 gRPC 服务器失败: %v", err)
	}

	// 创建 gRPC-Web 包装器
	grpcWebServer := grpcweb.WrapServer(srv.GrpcServer(),
		grpcweb.WithCorsForRegisteredEndpointsOnly(false),
		grpcweb.WithOriginFunc(func(origin string) bool {
			return true // 开发阶段允许所有来源
		}),
	)

	// HTTP 服务器同时处理 gRPC-Web 和普通 HTTP
	httpPort := getEnvInt("HTTP_PORT", 8080)
	httpServer := &http.Server{
		Addr: fmt.Sprintf(":%d", httpPort),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 健康检查
			if r.URL.Path == "/health" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"ok","service":"prismproxy"}`))
				return
			}

			// gRPC-Web 请求
			if grpcWebServer.IsGrpcWebRequest(r) || grpcWebServer.IsAcceptableGrpcCorsRequest(r) {
				grpcWebServer.ServeHTTP(w, r)
				return
			}

			// 普通 HTTP 请求 (可以扩展为 REST API 或静态文件服务)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"not found"}`))
		}),
	}

	// 在 goroutine 中启动 gRPC 服务器
	go func() {
		if err := srv.Start(); err != nil {
			log.Printf("[ERROR] gRPC 服务器退出: %v", err)
		}
	}()

	// 在 goroutine 中启动 HTTP 服务器
	go func() {
		log.Printf("[INFO] HTTP 服务器启动，监听 :%d (gRPC-Web + HTTP)", httpPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[ERROR] HTTP 服务器退出: %v", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	sig := <-quit
	log.Printf("[INFO] 收到信号 %v，正在关闭...", sig)

	// 创建关闭超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 关闭 HTTP 服务器
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("[ERROR] HTTP 服务器关闭失败: %v", err)
	}

	// 关闭代理服务器
	if proxyServer.IsRunning() {
		if err := proxyServer.Stop(); err != nil {
			log.Printf("[ERROR] 代理服务器关闭失败: %v", err)
		}
	}

	// 关闭 gRPC 服务器
	srv.Stop()

	log.Println("[INFO] PrismProxy gRPC 服务器已关闭")
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt 获取整数环境变量
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if n, err := strconv.Atoi(value); err == nil && n > 0 {
			return n
		}
	}
	return defaultValue
}
