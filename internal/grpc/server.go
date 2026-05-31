package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"prismproxy/internal/ai"
	"prismproxy/internal/cert"
	"prismproxy/internal/codegen"
	"prismproxy/internal/collection"
	"prismproxy/internal/debugger"
	"prismproxy/internal/diff"
	"prismproxy/internal/environment"
	"prismproxy/internal/perf"
	"prismproxy/internal/rewrite"
	"prismproxy/internal/rules"
	"prismproxy/internal/script"
	"prismproxy/internal/search"
	"prismproxy/internal/storage"
	"prismproxy/internal/traffic"
)

// Server gRPC 服务器
type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener

	// 模块引用
	traffic     *traffic.Manager
	rules       *rules.Engine
	debugger    *debugger.Debugger
	collection  *collection.Manager
	runner      *collection.Runner
	environment *environment.Manager
	rewrite     *rewrite.Engine
	ai          *ai.Service
	codegen     *codegen.Generator
	scriptStore *script.ScriptStore
	scriptEngine *script.ScriptEngine
	diffEngine  *diff.DiffEngine
	perfAnalyzer *perf.PerfAnalyzer
	certManager *cert.CertManager
	certStore   *cert.CertStore
	searchEngine *search.SearchEngine
	filterStore *search.FilterStore
	storage     *storage.Storage
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int
}

// NewServer 创建 gRPC 服务器
func NewServer(cfg ServerConfig, store *storage.Storage, trafficMgr *traffic.Manager, rulesEngine *rules.Engine, debuggerMgr *debugger.Debugger, collectionMgr *collection.Manager, runner *collection.Runner, envMgr *environment.Manager, rewriteEngine *rewrite.Engine, aiSvc *ai.Service, codegenGen *codegen.Generator, scriptStore *script.ScriptStore, scriptEng *script.ScriptEngine, diffEng *diff.DiffEngine, perfAnaly *perf.PerfAnalyzer, certMgr *cert.CertManager, certSt *cert.CertStore, searchEng *search.SearchEngine, filterSt *search.FilterStore) (*Server, error) {
	// 创建监听器
	addr := fmt.Sprintf(":%d", cfg.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("监听 %s 失败: %w", addr, err)
	}

	// 创建 gRPC 服务器 (带拦截器)
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggingInterceptor,
			panicRecoveryInterceptor,
		),
		grpc.ChainStreamInterceptor(
			streamLoggingInterceptor,
		),
	)

	s := &Server{
		grpcServer:   grpcServer,
		listener:     lis,
		traffic:      trafficMgr,
		rules:        rulesEngine,
		debugger:     debuggerMgr,
		collection:   collectionMgr,
		runner:       runner,
		environment:  envMgr,
		rewrite:      rewriteEngine,
		ai:           aiSvc,
		codegen:      codegenGen,
		scriptStore:  scriptStore,
		scriptEngine: scriptEng,
		diffEngine:   diffEng,
		perfAnalyzer: perfAnaly,
		certManager:  certMgr,
		certStore:    certSt,
		searchEngine: searchEng,
		filterStore:  filterSt,
		storage:      store,
	}

	// 注册所有服务
	s.registerServices()

	return s, nil
}

// registerServices 注册所有 gRPC 服务
func (s *Server) registerServices() {
	// 注册 TrafficService
	RegisterTrafficServiceImpl(s.grpcServer, s.traffic)
	log.Println("[INFO] 已注册 TrafficService")

	// 注册 RulesService
	if s.rules != nil {
		RegisterRulesServiceImpl(s.grpcServer, s.rules)
		log.Println("[INFO] 已注册 RulesService")
	}

	// 注册 BreakpointsService
	if s.debugger != nil {
		RegisterBreakpointsServiceImpl(s.grpcServer, s.debugger)
		log.Println("[INFO] 已注册 BreakpointsService")
	}

	// 注册 RewritesService
	if s.rewrite != nil {
		RegisterRewritesServiceImpl(s.grpcServer, s.rewrite)
		log.Println("[INFO] 已注册 RewritesService")
	}

	// 注册 CollectionsService
	if s.collection != nil {
		RegisterCollectionsServiceImpl(s.grpcServer, s.collection, s.runner)
		log.Println("[INFO] 已注册 CollectionsService")
	}

	// 注册 EnvironmentsService
	if s.environment != nil {
		RegisterEnvironmentsServiceImpl(s.grpcServer, s.environment)
		log.Println("[INFO] 已注册 EnvironmentsService")
	}

	// 注册 AIService
	if s.ai != nil {
		RegisterAIServiceImpl(s.grpcServer, s.ai, s.traffic)
		log.Println("[INFO] 已注册 AIService")
	}

	// 注册 SystemService
	RegisterSystemServiceImpl(s.grpcServer, s.traffic, s.rules)
	log.Println("[INFO] 已注册 SystemService")

	// 注册 CodeGenService
	if s.codegen != nil {
		RegisterCodeGenServiceImpl(s.grpcServer, s.codegen)
		log.Println("[INFO] 已注册 CodeGenService")
	}

	// 注册 ScriptsService
	if s.scriptStore != nil {
		RegisterScriptsServiceImpl(s.grpcServer, s.scriptStore, s.scriptEngine)
		log.Println("[INFO] 已注册 ScriptsService")
	}

	// 注册 DiffService
	if s.diffEngine != nil {
		RegisterDiffServiceImpl(s.grpcServer, s.diffEngine)
		log.Println("[INFO] 已注册 DiffService")
	}

	// 注册 PerfService
	if s.perfAnalyzer != nil {
		RegisterPerfServiceImpl(s.grpcServer, s.perfAnalyzer)
		log.Println("[INFO] 已注册 PerfService")
	}

	// 注册 CertService
	if s.certManager != nil {
		RegisterCertServiceImpl(s.grpcServer, s.certManager, s.certStore)
		log.Println("[INFO] 已注册 CertService")
	}

	// 注册 SearchService
	if s.searchEngine != nil {
		RegisterSearchServiceImpl(s.grpcServer, s.searchEngine, s.filterStore)
		log.Println("[INFO] 已注册 SearchService")
	}
}

// Start 启动 gRPC 服务器
func (s *Server) Start() error {
	log.Printf("[INFO] gRPC 服务器启动，监听 %s", s.listener.Addr())
	return s.grpcServer.Serve(s.listener)
}

// Stop 优雅关闭 gRPC 服务器
func (s *Server) Stop() {
	log.Println("[INFO] 正在关闭 gRPC 服务器...")
	s.grpcServer.GracefulStop()
	log.Println("[INFO] gRPC 服务器已关闭")
}

// GrpcServer 获取底层 gRPC 服务器 (用于 grpc-web 包装)
func (s *Server) GrpcServer() *grpc.Server {
	return s.grpcServer
}

// loggingInterceptor 日志拦截器 (记录每个 RPC 的方法名和耗时)
func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	// 调用处理器
	resp, err := handler(ctx, req)

	// 记录日志
	duration := time.Since(start)
	statusCode := codes.OK
	if err != nil {
		if st, ok := status.FromError(err); ok {
			statusCode = st.Code()
		}
	}
	log.Printf("[gRPC] %s %s %v", info.FullMethod, statusCode, duration)

	return resp, err
}

// streamLoggingInterceptor 流式 RPC 日志拦截器
func streamLoggingInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	start := time.Now()

	err := handler(srv, ss)

	duration := time.Since(start)
	statusCode := codes.OK
	if err != nil {
		if st, ok := status.FromError(err); ok {
			statusCode = st.Code()
		}
	}
	log.Printf("[gRPC-Stream] %s %s %v", info.FullMethod, statusCode, duration)

	return err
}

// panicRecoveryInterceptor panic 恢复拦截器
func panicRecoveryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[ERROR] gRPC panic recovered in %s: %v", info.FullMethod, r)
			err = status.Errorf(codes.Internal, "内部错误: %v", r)
		}
	}()

	return handler(ctx, req)
}
