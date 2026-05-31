package traffic

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"time"
)

// Collector 流量收集器
type Collector struct {
	manager    *Manager
	maxBodySize int64
}

// NewCollector 创建新的收集器
func NewCollector(manager *Manager, maxBodySize int64) *Collector {
	if maxBodySize <= 0 {
		maxBodySize = 10 * 1024 * 1024 // 默认 10MB
	}
	return &Collector{
		manager:     manager,
		maxBodySize: maxBodySize,
	}
}

// CollectRequest 收集请求数据
func (c *Collector) CollectRequest(r *http.Request) (*Transaction, error) {
	// 读取请求体
	reqBody, err := io.ReadAll(io.LimitReader(r.Body, c.maxBodySize))
	if err != nil {
		return nil, err
	}
	r.Body = io.NopCloser(bytes.NewBuffer(reqBody))

	// 转储原始请求
	rawReq, _ := httputil.DumpRequestOut(r, true)

	// 解析 URL
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	tx := &Transaction{
		Timestamp:  time.Now(),
		Method:     r.Method,
		URL:        r.URL.String(),
		Host:       r.Host,
		Path:       r.URL.Path,
		Scheme:     scheme,
		Port:       r.URL.Port(),
		ClientAddr: r.RemoteAddr,
		Request: &RequestData{
			Headers:     r.Header,
			Body:        reqBody,
			BodySize:    int64(len(reqBody)),
			ContentType: r.Header.Get("Content-Type"),
			Raw:         rawReq,
		},
		Response: &ResponseData{},
	}

	return tx, nil
}

// CollectResponse 收集响应数据
func (c *Collector) CollectResponse(tx *Transaction, resp *http.Response) error {
	// 读取响应体
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, c.maxBodySize))
	if err != nil {
		return err
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(respBody))

	// 转储原始响应
	rawResp, _ := httputil.DumpResponse(resp, true)

	tx.Response = &ResponseData{
		StatusCode:  resp.StatusCode,
		StatusText:  resp.Status,
		Headers:     resp.Header,
		Body:        respBody,
		BodySize:    int64(len(respBody)),
		ContentType: resp.Header.Get("Content-Type"),
		Raw:         rawResp,
	}

	// 解析服务器 IP
	if resp.Request != nil {
		host, _, _ := net.SplitHostPort(resp.Request.Host)
		if host == "" {
			host = resp.Request.Host
		}
		tx.ServerIP = resolveIP(host)
	}

	// 计算耗时
	tx.DurationMs = time.Since(tx.Timestamp).Milliseconds()

	return nil
}

// SaveAndNotify 保存并通知
func (c *Collector) SaveAndNotify(tx *Transaction) error {
	return c.manager.SaveTransaction(tx)
}

// resolveIP 解析域名 IP
func resolveIP(host string) string {
	addrs, err := net.LookupIP(host)
	if err != nil || len(addrs) == 0 {
		return ""
	}
	return addrs[0].String()
}

// MarshalTransaction 序列化事务为 JSON
func MarshalTransaction(tx *Transaction) ([]byte, error) {
	return json.Marshal(tx)
}

// UnmarshalTransaction 反序列化事务
func UnmarshalTransaction(data []byte) (*Transaction, error) {
	var tx Transaction
	err := json.Unmarshal(data, &tx)
	return &tx, err
}

// LogTransaction 记录事务日志
func LogTransaction(tx *Transaction) {
	log.Printf("[TRAFFIC] %s %s %d %dms",
		tx.Method, tx.URL, tx.Response.StatusCode, tx.DurationMs)
}
