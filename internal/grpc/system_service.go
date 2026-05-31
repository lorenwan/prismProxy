package grpc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"google.golang.org/grpc"

	"prismproxy/internal/rules"
	"prismproxy/internal/traffic"
	pb "prismproxy/proto/gen/go"
)

// SystemServiceImpl SystemService gRPC 实现
type SystemServiceImpl struct {
	pb.UnimplementedSystemServiceServer
	traffic *traffic.Manager
	rules   *rules.Engine
}

// RegisterSystemServiceImpl 注册 SystemService
func RegisterSystemServiceImpl(s *grpc.Server, trafficMgr *traffic.Manager, rulesEngine *rules.Engine) {
	pb.RegisterSystemServiceServer(s, &SystemServiceImpl{
		traffic: trafficMgr,
		rules:   rulesEngine,
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
	// TODO: 从配置文件或数据库加载设置
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
	}, nil
}

// UpdateSettings 更新设置
func (s *SystemServiceImpl) UpdateSettings(ctx context.Context, req *pb.Settings) (*pb.Settings, error) {
	// TODO: 保存设置到配置文件或数据库
	return req, nil
}

// StartProxy 启动代理
func (s *SystemServiceImpl) StartProxy(ctx context.Context, req *pb.Empty) (*pb.ProxyStatusResponse, error) {
	// TODO: 实现代理启停
	return &pb.ProxyStatusResponse{
		Running: true,
		Addr:    "0.0.0.0:8888",
	}, nil
}

// StopProxy 停止代理
func (s *SystemServiceImpl) StopProxy(ctx context.Context, req *pb.Empty) (*pb.ProxyStatusResponse, error) {
	// TODO: 实现代理启停
	return &pb.ProxyStatusResponse{
		Running: false,
	}, nil
}

// GenerateCA 生成 CA 证书
func (s *SystemServiceImpl) GenerateCA(ctx context.Context, req *pb.Empty) (*pb.CAGenerateResponse, error) {
	certPath := filepath.Join(".", "certs", "ca.crt")

	// TODO: 实现 CA 证书生成
	return &pb.CAGenerateResponse{
		Success:  true,
		CertPath: certPath,
	}, nil
}

// ExportCA 导出 CA 证书
func (s *SystemServiceImpl) ExportCA(ctx context.Context, req *pb.Empty) (*pb.CAExportResponse, error) {
	certPath := filepath.Join(".", "certs", "ca.crt")

	certData, err := os.ReadFile(certPath)
	if err != nil {
		return &pb.CAExportResponse{
			Error: "CA 证书文件不存在，请先生成证书",
		}, nil
	}

	return &pb.CAExportResponse{
		CertData: certData,
	}, nil
}

// 确保未使用的导入被使用
var _ = fmt.Sprintf
