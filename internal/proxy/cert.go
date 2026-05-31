package proxy

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"time"
)

// getCert 获取或生成动态证书
func (p *Proxy) getCert(host string) (*tls.Certificate, error) {
	// 从缓存获取
	if cert, ok := p.certCache.Load(host); ok {
		return cert.(*tls.Certificate), nil
	}

	// 生成新证书
	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: host},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	// 设置 IP 或 DNS
	if ip := net.ParseIP(host); ip != nil {
		template.IPAddresses = []net.IP{ip}
	} else {
		template.DNSNames = []string{host}
	}

	// 生成私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// 使用 CA 签发证书
	derBytes, err := x509.CreateCertificate(
		rand.Reader, template, p.ca.Leaf,
		&privateKey.PublicKey, p.ca.PrivateKey,
	)
	if err != nil {
		return nil, err
	}

	cert := &tls.Certificate{
		Certificate: [][]byte{derBytes, p.ca.Certificate[0]},
		PrivateKey:  privateKey,
	}

	// 缓存证书
	p.certCache.Store(host, cert)
	return cert, nil
}
