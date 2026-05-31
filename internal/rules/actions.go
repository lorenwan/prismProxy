package rules

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// ActionExecutor 动作执行器
type ActionExecutor struct{}

// NewActionExecutor 创建新的动作执行器
func NewActionExecutor() *ActionExecutor {
	return &ActionExecutor{}
}

// Execute 执行规则动作
func (e *ActionExecutor) Execute(rule *Rule, req *http.Request, resp *http.Response) (*http.Request, *http.Response, error) {
	switch rule.Action.Type {
	case ActionMapLocal:
		return e.executeMapLocal(rule, req)
	case ActionMapRemote:
		return e.executeMapRemote(rule, req)
	case ActionModifyRequest:
		return e.executeModifyRequest(rule, req)
	case ActionModifyResponse:
		return e.executeModifyResponse(rule, resp)
	case ActionBlock:
		return req, e.executeBlock(rule), nil
	case ActionDelay:
		return e.executeDelay(rule, req, resp)
	case ActionMock:
		return req, e.executeMock(rule), nil
	default:
		return req, resp, nil
	}
}

// executeMapLocal 映射到本地文件
func (e *ActionExecutor) executeMapLocal(rule *Rule, req *http.Request) (*http.Request, *http.Response, error) {
	localPath := rule.Action.LocalPath
	if localPath == "" {
		return req, nil, fmt.Errorf("本地路径为空")
	}

	// 读取本地文件
	data, err := os.ReadFile(localPath)
	if err != nil {
		return req, nil, fmt.Errorf("读取本地文件失败: %w", err)
	}

	// 创建响应
	resp := &http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(data)),
		Request:    req,
	}

	// 设置 Content-Type
	ext := strings.ToLower(localPath[strings.LastIndex(localPath, ".")+1:])
	switch ext {
	case "json":
		resp.Header.Set("Content-Type", "application/json")
	case "xml":
		resp.Header.Set("Content-Type", "application/xml")
	case "html":
		resp.Header.Set("Content-Type", "text/html")
	case "js":
		resp.Header.Set("Content-Type", "application/javascript")
	case "css":
		resp.Header.Set("Content-Type", "text/css")
	default:
		resp.Header.Set("Content-Type", "application/octet-stream")
	}

	return req, resp, nil
}

// executeMapRemote 映射到远程 URL
func (e *ActionExecutor) executeMapRemote(rule *Rule, req *http.Request) (*http.Request, *http.Response, error) {
	remoteURL := rule.Action.RemoteURL
	if remoteURL == "" {
		return req, nil, fmt.Errorf("远程 URL 为空")
	}

	// 替换 URL
	newReq := req.Clone(req.Context())
	newReq.URL, _ = url.Parse(remoteURL)
	newReq.Host = newReq.URL.Host

	return newReq, nil, nil
}

// executeModifyRequest 修改请求
func (e *ActionExecutor) executeModifyRequest(rule *Rule, req *http.Request) (*http.Request, *http.Response, error) {
	modify := rule.Action.Modify
	if modify == nil {
		return req, nil, nil
	}

	newReq := req.Clone(req.Context())

	// 添加 Header
	for k, v := range modify.AddHeaders {
		newReq.Header.Add(k, v)
	}

	// 设置 Header
	for k, v := range modify.SetHeaders {
		newReq.Header.Set(k, v)
	}

	// 删除 Header
	for _, k := range modify.RemoveHeaders {
		newReq.Header.Del(k)
	}

	// 修改 Query 参数
	q := newReq.URL.Query()
	for k, v := range modify.AddQuery {
		q.Add(k, v)
	}
	for k, v := range modify.SetQuery {
		q.Set(k, v)
	}
	for _, k := range modify.RemoveQuery {
		q.Del(k)
	}
	newReq.URL.RawQuery = q.Encode()

	// 替换 Body
	if modify.BodyReplace != "" {
		newReq.Body = io.NopCloser(strings.NewReader(modify.BodyReplace))
		newReq.ContentLength = int64(len(modify.BodyReplace))
	}

	return newReq, nil, nil
}

// executeModifyResponse 修改响应
func (e *ActionExecutor) executeModifyResponse(rule *Rule, resp *http.Response) (*http.Request, *http.Response, error) {
	modify := rule.Action.Modify
	if modify == nil || resp == nil {
		return nil, resp, nil
	}

	newResp := *resp
	newResp.Header = make(http.Header)
	for k, v := range resp.Header {
		newResp.Header[k] = v
	}

	// 添加 Header
	for k, v := range modify.AddHeaders {
		newResp.Header.Add(k, v)
	}

	// 设置 Header
	for k, v := range modify.SetHeaders {
		newResp.Header.Set(k, v)
	}

	// 删除 Header
	for _, k := range modify.RemoveHeaders {
		newResp.Header.Del(k)
	}

	// 替换 Body
	if modify.BodyReplace != "" {
		newResp.Body = io.NopCloser(strings.NewReader(modify.BodyReplace))
		newResp.ContentLength = int64(len(modify.BodyReplace))
	}

	return nil, &newResp, nil
}

// executeBlock 阻止请求
func (e *ActionExecutor) executeBlock(rule *Rule) *http.Response {
	block := rule.Action.BlockResponse
	if block == nil {
		block = &BlockSpec{
			StatusCode: 403,
			Body:       "Blocked by PrismProxy",
		}
	}

	resp := &http.Response{
		StatusCode: block.StatusCode,
		Status:     fmt.Sprintf("%d %s", block.StatusCode, http.StatusText(block.StatusCode)),
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(block.Body)),
	}

	for k, v := range block.Headers {
		resp.Header.Set(k, v)
	}

	return resp
}

// executeDelay 延迟请求
func (e *ActionExecutor) executeDelay(rule *Rule, req *http.Request, resp *http.Response) (*http.Request, *http.Response, error) {
	if rule.Action.DelayMs > 0 {
		time.Sleep(time.Duration(rule.Action.DelayMs) * time.Millisecond)
	}
	return req, resp, nil
}

// executeMock Mock 响应
func (e *ActionExecutor) executeMock(rule *Rule) *http.Response {
	modify := rule.Action.Modify
	if modify == nil {
		return &http.Response{
			StatusCode: 200,
			Status:     "200 OK",
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("{}")),
		}
	}

	body := modify.BodyReplace
	if body == "" {
		body = "{}"
	}

	// 尝试解析 JSON 格式化
	var prettyJSON bytes.Buffer
	if json.Indent(&prettyJSON, []byte(body), "", "  ") == nil {
		body = prettyJSON.String()
	}

	resp := &http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}

	resp.Header.Set("Content-Type", "application/json")

	for k, v := range modify.SetHeaders {
		resp.Header.Set(k, v)
	}

	return resp
}
