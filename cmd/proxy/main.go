package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"prismproxy/internal/api"
	"prismproxy/internal/proxy"
	"prismproxy/internal/storage"
	"prismproxy/internal/websocket"
)

func main() {
	// 数据库路径
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./prismproxy.db"
	}

	// 初始化存储
	store, err := storage.NewStorage(dbPath)
	if err != nil {
		log.Fatalf("[FATAL] 初始化存储失败: %v", err)
	}
	defer store.Close()

	// 执行数据库迁移
	if err := store.RunMigrations(); err != nil {
		log.Fatalf("[FATAL] 数据库迁移失败: %v", err)
	}

	// 加载 CA 证书
	caCert, err := os.ReadFile("./ca.crt")
	if err != nil {
		log.Fatalf("[FATAL] 加载 CA 证书失败: %v", err)
	}
	caKey, err := os.ReadFile("./ca.key")
	if err != nil {
		log.Fatalf("[FATAL] 加载 CA 私钥失败: %v", err)
	}

	// 初始化代理
	proxyServer, err := proxy.NewProxy(caCert, caKey, store)
	if err != nil {
		log.Fatalf("[FATAL] 初始化代理失败: %v", err)
	}

	// 初始化 WebSocket Hub
	hub := websocket.NewHub()
	go hub.Run()

	// 初始化 API
	apiService := api.NewAPI(store, hub)

	// 创建 Gin 路由
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// 注册 API 路由
	apiService.RegisterRoutes(r)

	// 代理服务器端口
	proxyAddr := ":8080"
	// API 服务端口
	apiAddr := ":8081"

	// 启动 API 服务
	go func() {
		log.Printf("[INFO] API 服务启动: %s", apiAddr)
		if err := r.Run(apiAddr); err != nil {
			log.Fatalf("[FATAL] API 服务启动失败: %v", err)
		}
	}()

	// 启动代理服务
	log.Printf("[INFO] 代理服务启动: %s", proxyAddr)
	proxyHTTP := &http.Server{
		Addr:    proxyAddr,
		Handler: proxyServer,
	}
	if err := proxyHTTP.ListenAndServe(); err != nil {
		log.Fatalf("[FATAL] 代理服务启动失败: %v", err)
	}
}
