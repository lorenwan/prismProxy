package cert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// CertManager 证书管理器
type CertManager struct {
	mu       sync.RWMutex
	caCert   *x509.Certificate
	caKey    *ecdsa.PrivateKey
	certPool *CertStore
	config   *CAConfig
}

// NewCertManager 创建证书管理器
func NewCertManager(store *CertStore) *CertManager {
	return &CertManager{
		certPool: store,
		config:   DefaultCAConfig(),
	}
}

// GenerateCA 生成自签名根 CA
func (m *CertManager) GenerateCA() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 生成 ECDSA 私钥
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("生成 CA 私钥失败: %w", err)
	}

	// 创建证书模板
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("生成序列号失败: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   "PrismProxy CA",
			Organization: []string{"PrismProxy"},
			Country:      []string{"US"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(m.config.ValidYears, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            1,
	}

	// 自签名
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("创建 CA 证书失败: %w", err)
	}

	// 解析证书
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return fmt.Errorf("解析 CA 证书失败: %w", err)
	}

	m.caCert = cert
	m.caKey = privateKey

	return nil
}

// LoadCA 加载 CA 证书
func (m *CertManager) LoadCA(certPEM, keyPEM []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 解析证书
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return fmt.Errorf("解析 CA 证书 PEM 失败")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("解析 CA 证书失败: %w", err)
	}

	// 解析私钥
	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return fmt.Errorf("解析 CA 私钥 PEM 失败")
	}

	privateKey, err := x509.ParseECPrivateKey(keyBlock.Bytes)
	if err != nil {
		return fmt.Errorf("解析 CA 私钥失败: %w", err)
	}

	m.caCert = cert
	m.caKey = privateKey

	return nil
}

// IssueCert 签发域名证书
func (m *CertManager) IssueCert(domain string) (*tls.Certificate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.caCert == nil || m.caKey == nil {
		return nil, fmt.Errorf("CA 未初始化")
	}

	// 检查缓存
	if cert := m.certPool.GetCached(domain); cert != nil {
		return cert, nil
	}

	// 生成域名私钥
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("生成域名私钥失败: %w", err)
	}

	// 创建证书模板
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("生成序列号失败: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   domain,
			Organization: []string{"PrismProxy"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(1, 0, 0), // 1 年有效期
		KeyUsage:  x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},
		DNSNames: []string{domain},
	}

	// 使用 CA 签发证书
	certDER, err := x509.CreateCertificate(rand.Reader, template, m.caCert, &privateKey.PublicKey, m.caKey)
	if err != nil {
		return nil, fmt.Errorf("签发证书失败: %w", err)
	}

	// 解析证书
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("解析证书失败: %w", err)
	}

	// 保存到存储
	m.certPool.SaveCert(domain, cert)

	// 构建 tls.Certificate
	tlsCert := &tls.Certificate{
		Certificate: [][]byte{certDER, m.caCert.Raw},
		PrivateKey:  privateKey,
	}

	// 缓存
	m.certPool.CacheCert(domain, tlsCert)

	return tlsCert, nil
}

// GetCAInfo 获取 CA 信息
func (m *CertManager) GetCAInfo() *CAInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.caCert == nil {
		return &CAInfo{
			IsLoaded: false,
		}
	}

	// 计算指纹
	fingerprint := sha256.Sum256(m.caCert.Raw)

	return &CAInfo{
		Subject:      m.caCert.Subject.String(),
		SerialNumber: m.caCert.SerialNumber.String(),
		NotBefore:    m.caCert.NotBefore,
		NotAfter:     m.caCert.NotAfter,
		Fingerprint:  hex.EncodeToString(fingerprint[:]),
		IsLoaded:     true,
	}
}

// ExportCACert 导出 CA 证书（PEM 格式）
func (m *CertManager) ExportCACert() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.caCert == nil {
		return nil, fmt.Errorf("CA 未初始化")
	}

	return pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: m.caCert.Raw,
	}), nil
}

// ExportCAKey 导出 CA 私钥（PEM 格式）
func (m *CertManager) ExportCAKey() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.caKey == nil {
		return nil, fmt.Errorf("CA 未初始化")
	}

	keyBytes, err := x509.MarshalECPrivateKey(m.caKey)
	if err != nil {
		return nil, fmt.Errorf("序列化私钥失败: %w", err)
	}

	return pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyBytes,
	}), nil
}

// IsCAInitialized CA 是否已初始化
func (m *CertManager) IsCAInitialized() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.caCert != nil && m.caKey != nil
}

// ClearCerts 清除所有域名证书缓存
func (m *CertManager) ClearCerts() {
	m.certPool.ClearCache()
}
