package grpc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"runtime"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"prismproxy/internal/cert"
	pb "prismproxy/proto/gen/go"
)

// CertServiceImpl CertService gRPC 实现
type CertServiceImpl struct {
	pb.UnimplementedCertServiceServer
	manager *cert.CertManager
	store   *cert.CertStore
}

// RegisterCertServiceImpl 注册 CertService
func RegisterCertServiceImpl(s *grpc.Server, manager *cert.CertManager, store *cert.CertStore) {
	pb.RegisterCertServiceServer(s, &CertServiceImpl{manager: manager, store: store})
}

// GetCAInfo 获取 CA 信息
func (s *CertServiceImpl) GetCAInfo(ctx context.Context, req *pb.Empty) (*pb.CAInfo, error) {
	info := s.manager.GetCAInfo()
	return caInfoToProto(info), nil
}

// GenerateCA 生成 CA 证书
func (s *CertServiceImpl) GenerateCA(ctx context.Context, req *pb.Empty) (*pb.CAInfo, error) {
	if err := s.manager.GenerateCA(); err != nil {
		return nil, status.Errorf(codes.Internal, "生成 CA 证书失败: %v", err)
	}

	info := s.manager.GetCAInfo()
	return caInfoToProto(info), nil
}

// ExportCA 导出 CA 证书
func (s *CertServiceImpl) ExportCA(ctx context.Context, req *pb.Empty) (*pb.ExportCAResponse, error) {
	certPEM, err := s.manager.ExportCACert()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "导出 CA 证书失败: %v", err)
	}

	keyPEM, err := s.manager.ExportCAKey()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "导出 CA 私钥失败: %v", err)
	}

	return &pb.ExportCAResponse{
		CertPem: certPEM,
		KeyPem:  keyPEM,
	}, nil
}

// IssueCert 签发域名证书
func (s *CertServiceImpl) IssueCert(ctx context.Context, req *pb.IssueCertRequest) (*pb.IssueCertResponse, error) {
	domain := req.GetDomain()
	if domain == "" {
		return nil, status.Errorf(codes.InvalidArgument, "域名不能为空")
	}

	tlsCert, err := s.manager.IssueCert(domain)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "签发证书失败: %v", err)
	}

	// 从 TLS 证书中提取信息
	notBefore := time.Time{}
	notAfter := time.Time{}
	serialNumber := ""
	if len(tlsCert.Certificate) > 0 {
		// 尝试解析证书获取详细信息
		certInfo, err := s.store.GetCert(domain)
		if err == nil && certInfo != nil {
			notBefore = certInfo.NotBefore
			notAfter = certInfo.NotAfter
			serialNumber = certInfo.SerialNumber.String()
		}
	}

	return &pb.IssueCertResponse{
		Domain:       domain,
		SerialNumber: serialNumber,
		NotBefore:    notBefore.Format(time.RFC3339),
		NotAfter:     notAfter.Format(time.RFC3339),
	}, nil
}

// ListCerts 列出所有域名证书
func (s *CertServiceImpl) ListCerts(ctx context.Context, req *pb.Empty) (*pb.CertListResponse, error) {
	certs := s.store.ListCerts()

	items := make([]*pb.CertInfoProto, len(certs))
	for i, c := range certs {
		items[i] = certInfoToProto(c)
	}

	return &pb.CertListResponse{Certs: items}, nil
}

// DeleteCert 删除域名证书
func (s *CertServiceImpl) DeleteCert(ctx context.Context, req *pb.CertDeleteRequest) (*pb.Empty, error) {
	if err := s.store.DeleteCert(req.GetDomain()); err != nil {
		return nil, status.Errorf(codes.Internal, "删除证书失败: %v", err)
	}

	return &pb.Empty{}, nil
}

// CheckCert 检查证书状态
func (s *CertServiceImpl) CheckCert(ctx context.Context, req *pb.CertCheckRequest) (*pb.CertCheckResponse, error) {
	domain := req.GetDomain()
	if domain == "" {
		return nil, status.Errorf(codes.InvalidArgument, "域名不能为空")
	}

	exists := s.store.HasCert(domain)
	resp := &pb.CertCheckResponse{
		Domain: domain,
		Exists: exists,
	}

	if exists {
		isValid, _ := s.store.IsValid(domain)
		isExpired, _ := s.store.IsExpired(domain)
		fingerprint, _ := s.store.GetFingerprint(domain)

		resp.IsValid = isValid
		resp.IsExpired = isExpired
		resp.Fingerprint = fingerprint
	}

	return resp, nil
}

// ClearCerts 清除所有证书缓存
func (s *CertServiceImpl) ClearCerts(ctx context.Context, req *pb.Empty) (*pb.Empty, error) {
	s.manager.ClearCerts()
	return &pb.Empty{}, nil
}

// GetTrustStatus 获取 CA 证书信任状态
func (s *CertServiceImpl) GetTrustStatus(ctx context.Context, req *pb.Empty) (*pb.TrustStatusResponse, error) {
	info := s.manager.GetCAInfo()
	if info == nil || !info.IsLoaded {
		return &pb.TrustStatusResponse{
			Trusted:  false,
			Platform: runtime.GOOS,
		}, nil
	}

	// 计算 CA 指纹
	fingerprint := ""
	if rawCert := s.manager.GetCARaw(); rawCert != nil {
		hash := sha256.Sum256(rawCert)
		fingerprint = hex.EncodeToString(hash[:])
	}

	return &pb.TrustStatusResponse{
		Trusted:        true,
		Platform:       runtime.GOOS,
		CaFingerprint:  fingerprint,
	}, nil
}

// === proto ↔ Go 转换函数 ===

func caInfoToProto(info *cert.CAInfo) *pb.CAInfo {
	if info == nil {
		return &pb.CAInfo{IsLoaded: false}
	}

	return &pb.CAInfo{
		Subject:      info.Subject,
		SerialNumber: info.SerialNumber,
		NotBefore:    info.NotBefore.Format(time.RFC3339),
		NotAfter:     info.NotAfter.Format(time.RFC3339),
		Fingerprint:  info.Fingerprint,
		IsLoaded:     info.IsLoaded,
	}
}

func certInfoToProto(info cert.CertInfo) *pb.CertInfoProto {
	return &pb.CertInfoProto{
		Domain:       info.Domain,
		SerialNumber: info.SerialNumber,
		NotBefore:    info.NotBefore.Format(time.RFC3339),
		NotAfter:     info.NotAfter.Format(time.RFC3339),
		Issuer:       info.Issuer,
		IsValid:      info.IsValid,
	}
}

// 确保 CertServiceImpl 实现了 pb.CertServiceServer
var _ pb.CertServiceServer = (*CertServiceImpl)(nil)
