package rewrite

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"prismproxy/internal/rules"
)

// Engine 重写引擎
type Engine struct {
	store   *Store
	matcher *rules.Matcher
	mu      sync.RWMutex
	rules   []*RewriteRule
}

// NewEngine 创建新的重写引擎
func NewEngine(db *sql.DB) *Engine {
	return &Engine{
		store:   NewStore(db),
		matcher: rules.NewMatcher(),
	}
}

// Init 初始化重写引擎
func (e *Engine) Init() error {
	if err := e.store.Init(); err != nil {
		return err
	}

	return e.reload()
}

// reload 重新加载规则
func (e *Engine) reload() error {
	rules, err := e.store.GetEnabled()
	if err != nil {
		return err
	}

	e.mu.Lock()
	e.rules = rules
	e.mu.Unlock()

	log.Printf("[INFO] 重写引擎加载 %d 条规则", len(rules))
	return nil
}

// ProcessRequest 处理请求重写
func (e *Engine) ProcessRequest(req *http.Request) (*http.Request, *http.Response, error) {
	e.mu.RLock()
	rules := e.rules
	e.mu.RUnlock()

	if len(rules) == 0 {
		return req, nil, nil
	}

	// 匹配规则
	matched := e.matcher.MatchRules(toRuleSlice(rules), req.Method, req.URL.String(), req.Host, req.Header)
	if len(matched) == 0 {
		return req, nil, nil
	}

	newReq := req.Clone(req.Context())
	var resp *http.Response

	// 按优先级执行匹配的重写规则
	for _, rule := range matched {
		rewriteRule := findRewriteRule(rules, rule.ID)
		if rewriteRule == nil {
			continue
		}

		for _, action := range rewriteRule.Actions {
			if !action.Where.IsRequest() {
				continue
			}

			var err error
			newReq, resp, err = e.executeAction(action, newReq, nil)
			if err != nil {
				log.Printf("[WARN] 执行重写动作失败: %v", err)
				continue
			}
		}
	}

	return newReq, resp, nil
}

// ProcessResponse 处理响应重写
func (e *Engine) ProcessResponse(req *http.Request, resp *http.Response) (*http.Request, *http.Response, error) {
	e.mu.RLock()
	rules := e.rules
	e.mu.RUnlock()

	if len(rules) == 0 || resp == nil {
		return req, resp, nil
	}

	// 匹配规则
	matched := e.matcher.MatchRules(toRuleSlice(rules), req.Method, req.URL.String(), req.Host, req.Header)
	if len(matched) == 0 {
		return req, resp, nil
	}

	newResp := *resp
	newResp.Header = make(http.Header)
	for k, v := range resp.Header {
		newResp.Header[k] = v
	}

	// 按优先级执行匹配的重写规则
	for _, rule := range matched {
		rewriteRule := findRewriteRule(rules, rule.ID)
		if rewriteRule == nil {
			continue
		}

		for _, action := range rewriteRule.Actions {
			if !action.Where.IsResponse() {
				continue
			}

			var err error
			_, respPtr, err := e.executeAction(action, req, &newResp)
			if err != nil {
				log.Printf("[WARN] 执行重写动作失败: %v", err)
				continue
			}
			if respPtr != nil {
				newResp = *respPtr
			}
		}
	}

	return req, &newResp, nil
}

// executeAction 执行重写动作
func (e *Engine) executeAction(action RewriteAction, req *http.Request, resp *http.Response) (*http.Request, *http.Response, error) {
	switch action.Type {
	case RewriteAddHeader:
		return e.executeAddHeader(action, req, resp)
	case RewriteRemoveHeader:
		return e.executeRemoveHeader(action, req, resp)
	case RewriteReplaceHeader:
		return e.executeReplaceHeader(action, req, resp)
	case RewriteReplaceBody:
		return e.executeReplaceBody(action, req, resp)
	case RewriteReplaceURL:
		return e.executeReplaceURL(action, req)
	case RewriteMapLocal:
		return e.executeMapLocal(action, req)
	case RewriteMapRemote:
		return e.executeMapRemote(action, req)
	default:
		return req, resp, nil
	}
}

// executeAddHeader 添加 Header
func (e *Engine) executeAddHeader(action RewriteAction, req *http.Request, resp *http.Response) (*http.Request, *http.Response, error) {
	if resp != nil {
		resp.Header.Add(action.Key, action.Value)
		return req, resp, nil
	}
	req.Header.Add(action.Key, action.Value)
	return req, nil, nil
}

// executeRemoveHeader 删除 Header
func (e *Engine) executeRemoveHeader(action RewriteAction, req *http.Request, resp *http.Response) (*http.Request, *http.Response, error) {
	if resp != nil {
		resp.Header.Del(action.Key)
		return req, resp, nil
	}
	req.Header.Del(action.Key)
	return req, nil, nil
}

// executeReplaceHeader 替换 Header
func (e *Engine) executeReplaceHeader(action RewriteAction, req *http.Request, resp *http.Response) (*http.Request, *http.Response, error) {
	if resp != nil {
		resp.Header.Set(action.Key, action.Value)
		return req, resp, nil
	}
	req.Header.Set(action.Key, action.Value)
	return req, nil, nil
}

// executeReplaceBody 替换 Body
func (e *Engine) executeReplaceBody(action RewriteAction, req *http.Request, resp *http.Response) (*http.Request, *http.Response, error) {
	if resp != nil {
		resp.Body = io.NopCloser(strings.NewReader(action.Value))
		resp.ContentLength = int64(len(action.Value))
		return req, resp, nil
	}
	req.Body = io.NopCloser(strings.NewReader(action.Value))
	req.ContentLength = int64(len(action.Value))
	return req, nil, nil
}

// executeReplaceURL 替换 URL
func (e *Engine) executeReplaceURL(action RewriteAction, req *http.Request) (*http.Request, *http.Response, error) {
	newReq := req.Clone(req.Context())
	newURL, err := url.Parse(action.Value)
	if err != nil {
		return req, nil, err
	}
	newReq.URL = newURL
	newReq.Host = newURL.Host
	return newReq, nil, nil
}

// executeMapLocal 映射到本地文件
func (e *Engine) executeMapLocal(action RewriteAction, req *http.Request) (*http.Request, *http.Response, error) {
	if action.Target == "" {
		return req, nil, nil
	}

	data, err := os.ReadFile(action.Target)
	if err != nil {
		return req, nil, err
	}

	resp := &http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(data)),
		Request:    req,
	}

	// 根据文件扩展名设置 Content-Type
	ext := strings.ToLower(action.Target[strings.LastIndex(action.Target, ".")+1:])
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
func (e *Engine) executeMapRemote(action RewriteAction, req *http.Request) (*http.Request, *http.Response, error) {
	if action.Target == "" {
		return req, nil, nil
	}

	newReq := req.Clone(req.Context())
	remoteURL, err := url.Parse(action.Target)
	if err != nil {
		return req, nil, err
	}
	newReq.URL = remoteURL
	newReq.Host = remoteURL.Host

	return newReq, nil, nil
}

// CreateRule 创建重写规则
func (e *Engine) CreateRule(rule *RewriteRule) error {
	if err := e.store.Create(rule); err != nil {
		return err
	}
	return e.reload()
}

// GetRule 获取重写规则
func (e *Engine) GetRule(id string) (*RewriteRule, error) {
	return e.store.GetByID(id)
}

// ListRules 获取重写规则列表
func (e *Engine) ListRules() ([]*RewriteRule, error) {
	return e.store.List()
}

// UpdateRule 更新重写规则
func (e *Engine) UpdateRule(rule *RewriteRule) error {
	if err := e.store.Update(rule); err != nil {
		return err
	}
	return e.reload()
}

// DeleteRule 删除重写规则
func (e *Engine) DeleteRule(id string) error {
	if err := e.store.Delete(id); err != nil {
		return err
	}
	return e.reload()
}

// ToggleRule 切换重写规则状态
func (e *Engine) ToggleRule(id string) (bool, error) {
	enabled, err := e.store.Toggle(id)
	if err != nil {
		return false, err
	}
	if err := e.reload(); err != nil {
		return false, err
	}
	return enabled, nil
}

// ReorderRules 重新排序重写规则
func (e *Engine) ReorderRules(ids []string) error {
	if err := e.store.Reorder(ids); err != nil {
		return err
	}
	return e.reload()
}

// ImportRules 导入重写规则
func (e *Engine) ImportRules(data []byte) (int, error) {
	var rules []*RewriteRule
	if err := json.Unmarshal(data, &rules); err != nil {
		return 0, err
	}

	count := 0
	for _, rule := range rules {
		if err := e.store.Create(rule); err != nil {
			log.Printf("[WARN] 导入重写规则失败: %v", err)
			continue
		}
		count++
	}

	if err := e.reload(); err != nil {
		return count, err
	}

	return count, nil
}

// ExportRules 导出重写规则
func (e *Engine) ExportRules() ([]byte, error) {
	rules, err := e.store.List()
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(rules, "", "  ")
}

// toRuleSlice 将 RewriteRule 转换为 rules.Rule 用于匹配
func toRuleSlice(rewriteRules []*RewriteRule) []*rules.Rule {
	result := make([]*rules.Rule, len(rewriteRules))
	for i, r := range rewriteRules {
		result[i] = &rules.Rule{
			ID:      r.ID,
			Enabled: r.Enabled,
			Match:   r.Match,
		}
	}
	return result
}

// findRewriteRule 根据 ID 查找重写规则
func findRewriteRule(rules []*RewriteRule, id string) *RewriteRule {
	for _, r := range rules {
		if r.ID == id {
			return r
		}
	}
	return nil
}
