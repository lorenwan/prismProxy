package cert

import "time"

// CAInfo CA 证书信息
type CAInfo struct {
	Subject      string    `json:"subject"`
	SerialNumber string    `json:"serial_number"`
	NotBefore    time.Time `json:"not_before"`
	NotAfter     time.Time `json:"not_after"`
	Fingerprint  string    `json:"fingerprint"`
	IsLoaded     bool      `json:"is_loaded"`
}

// CertInfo 域名证书信息
type CertInfo struct {
	Domain       string    `json:"domain"`
	SerialNumber string    `json:"serial_number"`
	NotBefore    time.Time `json:"not_before"`
	NotAfter     time.Time `json:"not_after"`
	Issuer       string    `json:"issuer"`
	IsValid      bool      `json:"is_valid"`
}

// CAConfig CA 配置
type CAConfig struct {
	Subject    string `json:"subject"`
	KeyLength  int    `json:"key_length"`
	ValidYears int    `json:"valid_years"`
}

// DefaultCAConfig 默认 CA 配置
func DefaultCAConfig() *CAConfig {
	return &CAConfig{
		Subject:    "CN=PrismProxy CA,O=PrismProxy,C=US",
		KeyLength:  2048,
		ValidYears: 10,
	}
}

// CertConfig 证书配置
type CertConfig struct {
	Domain     string `json:"domain"`
	ValidYears int    `json:"valid_years"`
}

// DefaultCertConfig 默认证书配置
func DefaultCertConfig(domain string) *CertConfig {
	return &CertConfig{
		Domain:     domain,
		ValidYears: 1,
	}
}
