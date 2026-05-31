package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"time"

	"prismproxy/internal/storage"
)

// handleHTTP 处理普通 HTTP 请求
func (p *Proxy) handleHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// 读取请求体
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "读取请求体失败", http.StatusInternalServerError)
		log.Printf("[ERROR] 读取请求体失败: %v", err)
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(reqBody))

	// 创建转发请求
	req, err := http.NewRequest(r.Method, r.URL.String(), bytes.NewReader(reqBody))
	if err != nil {
		http.Error(w, "创建请求失败", http.StatusInternalServerError)
		log.Printf("[ERROR] 创建请求失败: %v", err)
		return
	}
	req.Header = r.Header

	// 获取服务器 IP
	host, _, _ := net.SplitHostPort(req.Host)
	serverIP := resolveIP(host)

	// 转储原始请求
	rawReq, _ := httputil.DumpRequestOut(req, true)

	// 发送请求到目标服务器
	resp, err := p.transport.RoundTrip(req)
	if err != nil {
		http.Error(w, "转发请求失败", http.StatusServiceUnavailable)
		log.Printf("[ERROR] 转发请求失败: %v", err)
		return
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "读取响应体失败", http.StatusInternalServerError)
		log.Printf("[ERROR] 读取响应体失败: %v", err)
		return
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(respBody))

	// 转储原始响应
	rawResp, _ := httputil.DumpResponse(resp, true)

	duration := time.Since(startTime).Milliseconds()
	log.Printf("[INFO] 收到响应: %s (耗时 %dms)", resp.Status, duration)

	// 保存抓包数据
	p.saveTraffic(req, resp, serverIP, reqBody, respBody, rawReq, rawResp, duration)

	// 返回响应给客户端
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// saveTraffic 保存抓包数据到存储
func (p *Proxy) saveTraffic(req *http.Request, resp *http.Response, serverIP string,
	reqBody, respBody, rawReq, rawResp []byte, durationMs int64) {

	reqHeaders, _ := json.Marshal(req.Header)
	respHeaders, _ := json.Marshal(resp.Header)

	data := &storage.TrafficData{
		Method:      req.Method,
		URL:         req.URL.String(),
		Host:        req.Host,
		ServerIP:    serverIP,
		ContentType: resp.Header.Get("Content-Type"),
		StatusCode:  resp.StatusCode,
		DurationMs:  durationMs,
		ReqHeaders:  string(reqHeaders),
		ReqBody:     reqBody,
		RespHeaders: string(respHeaders),
		RespBody:    respBody,
		RawReq:      rawReq,
		RawResp:     rawResp,
	}

	if err := p.storage.SaveTraffic(data); err != nil {
		log.Printf("[ERROR] 保存抓包数据失败: %v", err)
	}
}

// resolveIP 解析域名 IP
func resolveIP(host string) string {
	addrs, err := net.LookupIP(host)
	if err != nil || len(addrs) == 0 {
		return ""
	}
	return addrs[0].String()
}
