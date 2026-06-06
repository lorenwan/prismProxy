package proxy

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"net/http"
	"sync"

	"prismproxy/internal/storage"
)

// Proxy 代理服务器
type Proxy struct {
	ca        tls.Certificate
	certCache sync.Map
	storage   *storage.Storage
	transport *http.Transport
}

// NewProxy 创建新的代理实例
func NewProxy(caCert, caKey []byte, store *storage.Storage) (*Proxy, error) {
	ca, err := tls.X509KeyPair(caCert, caKey)
	if err != nil {
		return nil, err
	}

	// 解析 CA 证书
	ca.Leaf, err = x509.ParseCertificate(ca.Certificate[0])
	if err != nil {
		return nil, err
	}

	p := &Proxy{
		ca:      ca,
		storage: store,
		transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	log.Println("[INFO] 代理服务器初始化完成")
	return p, nil
}

// ServeHTTP 实现 http.Handler 接口
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] 收到请求: %s %s %s", r.Method, r.Host, r.URL.String())

	if r.Method == http.MethodConnect {
		p.handleConnect(w, r)
	} else {
		p.handleHTTP(w, r)
	}
}
