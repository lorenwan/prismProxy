package grpc

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"prismproxy/internal/cert"
	"prismproxy/internal/proxy"
	"prismproxy/internal/rules"
	"prismproxy/internal/storage"
	"prismproxy/internal/traffic"
	pb "prismproxy/proto/gen/go"
)

// ProxyController 代理控制器（回调模式）
type ProxyController struct {
	StartFunc  func() error
	StopFunc   func() error
	StatusFunc func() (running bool, addr string)
}

// SystemServiceImpl SystemService gRPC 实现
type SystemServiceImpl struct {
	pb.UnimplementedSystemServiceServer
	traffic    *traffic.Manager
	rules      *rules.Engine
	store      *storage.Storage
	certMgr    *cert.CertManager
	proxyCtrl  *ProxyController
	systemProxy *proxy.SystemProxy
}

// RegisterSystemServiceImpl 注册 SystemService
func RegisterSystemServiceImpl(s *grpc.Server, trafficMgr *traffic.Manager, rulesEngine *rules.Engine, store *storage.Storage, certMgr *cert.CertManager, proxyCtrl *ProxyController, systemProxy *proxy.SystemProxy) {
	pb.RegisterSystemServiceServer(s, &SystemServiceImpl{
		traffic:     trafficMgr,
		rules:       rulesEngine,
		store:       store,
		certMgr:     certMgr,
		proxyCtrl:   proxyCtrl,
		systemProxy: systemProxy,
	})
}

// GetStatus 获取系统状态
func (s *SystemServiceImpl) GetStatus(ctx context.Context, req *pb.Empty) (*pb.SystemStatus, error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	status := &pb.SystemStatus{
		Version:          "0.1.0",
		GoVersion:        runtime.Version(),
		Os:               runtime.GOOS,
		Arch:             runtime.GOARCH,
		MemoryUsageBytes: int64(m.Alloc),
		GoroutineCount:   int32(runtime.NumGoroutine()),
	}

	// 获取代理状态
	if s.proxyCtrl != nil && s.proxyCtrl.StatusFunc != nil {
		running, addr := s.proxyCtrl.StatusFunc()
		status.ProxyRunning = running
		status.ProxyAddr = addr
	}

	// 获取流量统计
	if s.traffic != nil {
		stats, err := s.traffic.GetStats()
		if err == nil && stats != nil {
			status.TrafficCount = stats.TotalRequests
		}
	}

	// 获取规则统计
	if s.rules != nil {
		stats, err := s.rules.GetStats()
		if err == nil && stats != nil {
			status.RuleCount = int32(stats.TotalRules)
		}
	}

	return status, nil
}

// GetSettings 获取设置
func (s *SystemServiceImpl) GetSettings(ctx context.Context, req *pb.Empty) (*pb.Settings, error) {
	if s.store == nil {
		return s.defaultSettings(), nil
	}

	settings, err := s.store.GetSettings()
	if err != nil {
		log.Printf("[WARN] 获取设置失败: %v，使用默认值", err)
		return s.defaultSettings(), nil
	}

	return &pb.Settings{
		Proxy: &pb.ProxySettings{
			Port:        int32(settings.ProxyPort),
			ListenAddr:  settings.ProxyListenAddr,
			EnableHttps: settings.ProxyEnableHTTPS,
			EnableMitm:  settings.ProxyEnableMITM,
		},
		Theme:            settings.Theme,
		Language:         settings.Language,
		EnableTrafficLog: settings.EnableTrafficLog,
		MaxTrafficCount:  int32(settings.MaxTrafficCount),
		TrafficTtlHours:  int32(settings.TrafficTTLHours),
	}, nil
}

// UpdateSettings 更新设置
func (s *SystemServiceImpl) UpdateSettings(ctx context.Context, req *pb.Settings) (*pb.Settings, error) {
	if s.store == nil {
		return req, nil
	}

	settings := &storage.Settings{
		Theme:            req.GetTheme(),
		Language:         req.GetLanguage(),
		EnableTrafficLog: req.GetEnableTrafficLog(),
		MaxTrafficCount:  int(req.GetMaxTrafficCount()),
		TrafficTTLHours:  int(req.GetTrafficTtlHours()),
	}

	if req.GetProxy() != nil {
		settings.ProxyPort = int(req.GetProxy().GetPort())
		settings.ProxyListenAddr = req.GetProxy().GetListenAddr()
		settings.ProxyEnableHTTPS = req.GetProxy().GetEnableHttps()
		settings.ProxyEnableMITM = req.GetProxy().GetEnableMitm()
	}

	if err := s.store.SaveSettings(settings); err != nil {
		return nil, status.Errorf(codes.Internal, "保存设置失败: %v", err)
	}

	return req, nil
}

// StartProxy 启动代理
func (s *SystemServiceImpl) StartProxy(ctx context.Context, req *pb.Empty) (*pb.ProxyStatusResponse, error) {
	if s.proxyCtrl == nil || s.proxyCtrl.StartFunc == nil {
		return &pb.ProxyStatusResponse{
			Running: false,
			Error:   "代理控制器未配置",
		}, nil
	}

	if err := s.proxyCtrl.StartFunc(); err != nil {
		return &pb.ProxyStatusResponse{
			Running: false,
			Error:   err.Error(),
		}, nil
	}

	running, addr := false, ""
	if s.proxyCtrl.StatusFunc != nil {
		running, addr = s.proxyCtrl.StatusFunc()
	}

	return &pb.ProxyStatusResponse{
		Running: running,
		Addr:    addr,
	}, nil
}

// StopProxy 停止代理
func (s *SystemServiceImpl) StopProxy(ctx context.Context, req *pb.Empty) (*pb.ProxyStatusResponse, error) {
	if s.proxyCtrl == nil || s.proxyCtrl.StopFunc == nil {
		return &pb.ProxyStatusResponse{
			Running: false,
			Error:   "代理控制器未配置",
		}, nil
	}

	if err := s.proxyCtrl.StopFunc(); err != nil {
		return &pb.ProxyStatusResponse{
			Running: true,
			Error:   err.Error(),
		}, nil
	}

	return &pb.ProxyStatusResponse{
		Running: false,
	}, nil
}

// EnableSystemProxy 启用系统代理
func (s *SystemServiceImpl) EnableSystemProxy(ctx context.Context, req *pb.Empty) (*pb.SystemProxyStatus, error) {
	if s.systemProxy == nil {
		return &pb.SystemProxyStatus{
			Enabled: false,
			Error:   "系统代理管理器未配置",
		}, nil
	}

	if err := s.systemProxy.Enable(); err != nil {
		return &pb.SystemProxyStatus{
			Enabled: false,
			Error:   err.Error(),
		}, nil
	}

	_, addr := s.proxyCtrl.StatusFunc()
	return &pb.SystemProxyStatus{
		Enabled:   true,
		ProxyAddr: addr,
	}, nil
}

// DisableSystemProxy 禁用系统代理
func (s *SystemServiceImpl) DisableSystemProxy(ctx context.Context, req *pb.Empty) (*pb.SystemProxyStatus, error) {
	if s.systemProxy == nil {
		return &pb.SystemProxyStatus{
			Enabled: false,
			Error:   "系统代理管理器未配置",
		}, nil
	}

	if err := s.systemProxy.Disable(); err != nil {
		return &pb.SystemProxyStatus{
			Enabled: true,
			Error:   err.Error(),
		}, nil
	}

	return &pb.SystemProxyStatus{
		Enabled: false,
	}, nil
}

// GetSystemProxyStatus 获取系统代理状态
func (s *SystemServiceImpl) GetSystemProxyStatus(ctx context.Context, req *pb.Empty) (*pb.SystemProxyStatus, error) {
	if s.systemProxy == nil {
		return &pb.SystemProxyStatus{
			Enabled: false,
		}, nil
	}

	enabled, err := s.systemProxy.IsEnabled()
	if err != nil {
		return &pb.SystemProxyStatus{
			Enabled: false,
			Error:   err.Error(),
		}, nil
	}

	proxyAddr := ""
	if s.proxyCtrl != nil && s.proxyCtrl.StatusFunc != nil {
		_, proxyAddr = s.proxyCtrl.StatusFunc()
	}

	return &pb.SystemProxyStatus{
		Enabled:   enabled,
		ProxyAddr: proxyAddr,
	}, nil
}

// GenerateCA 生成 CA 证书
func (s *SystemServiceImpl) GenerateCA(ctx context.Context, req *pb.Empty) (*pb.CAGenerateResponse, error) {
	if s.certMgr == nil {
		return &pb.CAGenerateResponse{
			Success: false,
			Error:   "证书管理器未配置",
		}, nil
	}

	// 生成 CA 证书
	if err := s.certMgr.GenerateCA(); err != nil {
		return &pb.CAGenerateResponse{
			Success: false,
			Error:   fmt.Sprintf("生成 CA 证书失败: %v", err),
		}, nil
	}

	// 导出并保存到文件
	certDir := filepath.Join(".", "certs")
	os.MkdirAll(certDir, 0755)
	certPath := filepath.Join(certDir, "ca.crt")

	certPEM, err := s.certMgr.ExportCACert()
	if err != nil {
		return &pb.CAGenerateResponse{
			Success: false,
			Error:   fmt.Sprintf("导出 CA 证书失败: %v", err),
		}, nil
	}

	if err := os.WriteFile(certPath, certPEM, 0644); err != nil {
		return &pb.CAGenerateResponse{
			Success: false,
			Error:   fmt.Sprintf("保存 CA 证书失败: %v", err),
		}, nil
	}

	return &pb.CAGenerateResponse{
		Success:  true,
		CertPath: certPath,
	}, nil
}

// ExportCA 导出 CA 证书
func (s *SystemServiceImpl) ExportCA(ctx context.Context, req *pb.Empty) (*pb.CAExportResponse, error) {
	if s.certMgr == nil {
		return &pb.CAExportResponse{
			Error: "证书管理器未配置",
		}, nil
	}

	certData, err := s.certMgr.ExportCACert()
	if err != nil {
		return &pb.CAExportResponse{
			Error: fmt.Sprintf("导出 CA 证书失败: %v", err),
		}, nil
	}

	return &pb.CAExportResponse{
		CertData: certData,
	}, nil
}

// defaultSettings 返回默认设置
func (s *SystemServiceImpl) defaultSettings() *pb.Settings {
	return &pb.Settings{
		Proxy: &pb.ProxySettings{
			Port:        8888,
			ListenAddr:  "0.0.0.0",
			EnableHttps: true,
			EnableMitm:  true,
		},
		Theme:            "dark",
		Language:         "zh",
		EnableTrafficLog: true,
		MaxTrafficCount:  10000,
		TrafficTtlHours:  72,
	}
}
