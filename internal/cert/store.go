package cert

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// CertStore 证书存储
type CertStore struct {
	mu      sync.RWMutex
	certs   map[string]*x509.Certificate
	cache   map[string]*tls.Certificate
	ordered []string
}

// NewCertStore 创建证书存储
func NewCertStore() *CertStore {
	return &CertStore{
		certs:   make(map[string]*x509.Certificate),
		cache:   make(map[string]*tls.Certificate),
		ordered: make([]string, 0),
	}
}

// SaveCert 保存证书
func (s *CertStore) SaveCert(domain string, cert *x509.Certificate) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.certs[domain]; !exists {
		s.ordered = append(s.ordered, domain)
	}

	s.certs[domain] = cert
}

// GetCert 获取证书
func (s *CertStore) GetCert(domain string) (*x509.Certificate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cert, exists := s.certs[domain]
	if !exists {
		return nil, fmt.Errorf("证书不存在: %s", domain)
	}

	return cert, nil
}

// ListCerts 列出所有证书
func (s *CertStore) ListCerts() []CertInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var certs []CertInfo
	now := time.Now()

	for _, domain := range s.ordered {
		cert := s.certs[domain]
		if cert == nil {
			continue
		}

		certs = append(certs, CertInfo{
			Domain:       domain,
			SerialNumber: cert.SerialNumber.String(),
			NotBefore:    cert.NotBefore,
			NotAfter:     cert.NotAfter,
			Issuer:       cert.Issuer.String(),
			IsValid:      now.After(cert.NotBefore) && now.Before(cert.NotAfter),
		})
	}

	return certs
}

// DeleteCert 删除证书
func (s *CertStore) DeleteCert(domain string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.certs[domain]; !exists {
		return fmt.Errorf("证书不存在: %s", domain)
	}

	delete(s.certs, domain)
	delete(s.cache, domain)

	// 从有序列表中移除
	for i, d := range s.ordered {
		if d == domain {
			s.ordered = append(s.ordered[:i], s.ordered[i+1:]...)
			break
		}
	}

	return nil
}

// CacheCert 缓存 TLS 证书
func (s *CertStore) CacheCert(domain string, cert *tls.Certificate) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cache[domain] = cert
}

// GetCached 获取缓存的 TLS 证书
func (s *CertStore) GetCached(domain string) *tls.Certificate {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.cache[domain]
}

// ClearCache 清除缓存
func (s *CertStore) ClearCache() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cache = make(map[string]*tls.Certificate)
}

// HasCert 是否有指定域名的证书
func (s *CertStore) HasCert(domain string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.certs[domain]
	return exists
}

// Count 获取证书数量
func (s *CertStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.certs)
}

// GetFingerprint 获取证书指纹
func (s *CertStore) GetFingerprint(domain string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cert, exists := s.certs[domain]
	if !exists {
		return "", fmt.Errorf("证书不存在: %s", domain)
	}

	fingerprint := sha256.Sum256(cert.Raw)
	return hex.EncodeToString(fingerprint[:]), nil
}

// IsExpired 证书是否过期
func (s *CertStore) IsExpired(domain string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cert, exists := s.certs[domain]
	if !exists {
		return false, fmt.Errorf("证书不存在: %s", domain)
	}

	return time.Now().After(cert.NotAfter), nil
}

// IsValid 证书是否有效
func (s *CertStore) IsValid(domain string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cert, exists := s.certs[domain]
	if !exists {
		return false, fmt.Errorf("证书不存在: %s", domain)
	}

	now := time.Now()
	return now.After(cert.NotBefore) && now.Before(cert.NotAfter), nil
}
