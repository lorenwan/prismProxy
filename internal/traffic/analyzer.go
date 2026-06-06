package traffic

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Analyzer 流量分析器
type Analyzer struct {
	manager *Manager
}

// NewAnalyzer 创建新的分析器
func NewAnalyzer(manager *Manager) *Analyzer {
	return &Analyzer{manager: manager}
}

// AnalyzeTransaction 分析单条流量
func (a *Analyzer) AnalyzeTransaction(tx *Transaction) *AnalysisResult {
	result := &AnalysisResult{
		TransactionID: tx.ID,
		Timestamp:     time.Now(),
	}

	// 分析请求
	result.RequestAnalysis = a.analyzeRequest(tx.Request)

	// 分析响应
	result.ResponseAnalysis = a.analyzeResponse(tx.Response)

	// 性能分析
	result.PerformanceAnalysis = a.analyzePerformance(tx)

	// 安全分析
	result.SecurityAnalysis = a.analyzeSecurity(tx)

	return result
}

// analyzeRequest 分析请求
func (a *Analyzer) analyzeRequest(req *RequestData) *RequestAnalysis {
	if req == nil {
		return nil
	}

	analysis := &RequestAnalysis{
		BodySize:    req.BodySize,
		ContentType: req.ContentType,
		HeaderCount: len(req.Headers),
	}

	// 检测请求类型
	analysis.RequestType = detectRequestType(req.ContentType)

	// 检测敏感信息
	analysis.SensitiveInfo = detectSensitiveInfo(req)

	return analysis
}

// analyzeResponse 分析响应
func (a *Analyzer) analyzeResponse(resp *ResponseData) *ResponseAnalysis {
	if resp == nil {
		return nil
	}

	analysis := &ResponseAnalysis{
		StatusCode:  resp.StatusCode,
		BodySize:    resp.BodySize,
		ContentType: resp.ContentType,
		HeaderCount: len(resp.Headers),
	}

	// 状态分类
	analysis.StatusCategory = categorizeStatus(resp.StatusCode)

	// 检测缓存
	analysis.CacheInfo = detectCacheInfo(resp.Headers)

	return analysis
}

// analyzePerformance 分析性能
func (a *Analyzer) analyzePerformance(tx *Transaction) *PerformanceAnalysis {
	analysis := &PerformanceAnalysis{
		DurationMs: tx.DurationMs,
	}

	// 性能等级
	if tx.DurationMs < 100 {
		analysis.Grade = "excellent"
	} else if tx.DurationMs < 300 {
		analysis.Grade = "good"
	} else if tx.DurationMs < 1000 {
		analysis.Grade = "acceptable"
	} else if tx.DurationMs < 3000 {
		analysis.Grade = "slow"
	} else {
		analysis.Grade = "very_slow"
	}

	// 请求大小
	analysis.RequestSize = tx.Request.BodySize
	analysis.ResponseSize = tx.Response.BodySize

	return analysis
}

// analyzeSecurity 分析安全性
func (a *Analyzer) analyzeSecurity(tx *Transaction) *SecurityAnalysis {
	analysis := &SecurityAnalysis{
		Issues: []SecurityIssue{},
	}

	// 检查 HTTPS
	if tx.Scheme != "https" {
		analysis.Issues = append(analysis.Issues, SecurityIssue{
			Severity:    "medium",
			Name:        "非 HTTPS 请求",
			Description: "请求使用 HTTP 明文传输，存在安全风险",
		})
	}

	// 检查敏感头信息
	sensitiveHeaders := []string{"Authorization", "Cookie", "X-API-Key"}
	for _, header := range sensitiveHeaders {
		if tx.Request.Headers.Get(header) != "" {
			analysis.Issues = append(analysis.Issues, SecurityIssue{
				Severity:    "info",
				Name:        "敏感头信息",
				Description: fmt.Sprintf("请求包含敏感头信息: %s", header),
			})
		}
	}

	// 检查响应中的安全头
	securityHeaders := map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY 或 SAMEORIGIN",
		"X-XSS-Protection":          "1; mode=block",
		"Strict-Transport-Security": "HSTS",
	}
	for header, expected := range securityHeaders {
		if tx.Response.Headers.Get(header) == "" {
			analysis.Issues = append(analysis.Issues, SecurityIssue{
				Severity:    "low",
				Name:        "缺少安全头",
				Description: fmt.Sprintf("响应缺少安全头 %s (建议: %s)", header, expected),
			})
		}
	}

	// 安全评分
	score := 100
	for _, issue := range analysis.Issues {
		switch issue.Severity {
		case "critical":
			score -= 30
		case "high":
			score -= 20
		case "medium":
			score -= 10
		case "low":
			score -= 5
		}
	}
	if score < 0 {
		score = 0
	}
	analysis.Score = score

	return analysis
}

// AnalysisResult 分析结果
type AnalysisResult struct {
	TransactionID       int64                `json:"transaction_id"`
	Timestamp           time.Time            `json:"timestamp"`
	RequestAnalysis     *RequestAnalysis     `json:"request_analysis,omitempty"`
	ResponseAnalysis    *ResponseAnalysis    `json:"response_analysis,omitempty"`
	PerformanceAnalysis *PerformanceAnalysis `json:"performance_analysis,omitempty"`
	SecurityAnalysis    *SecurityAnalysis    `json:"security_analysis,omitempty"`
}

// RequestAnalysis 请求分析
type RequestAnalysis struct {
	RequestType   string          `json:"request_type"`
	BodySize      int64           `json:"body_size"`
	ContentType   string          `json:"content_type"`
	HeaderCount   int             `json:"header_count"`
	SensitiveInfo []SensitiveInfo `json:"sensitive_info,omitempty"`
}

// ResponseAnalysis 响应分析
type ResponseAnalysis struct {
	StatusCode     int        `json:"status_code"`
	StatusCategory string     `json:"status_category"`
	BodySize       int64      `json:"body_size"`
	ContentType    string     `json:"content_type"`
	HeaderCount    int        `json:"header_count"`
	CacheInfo      *CacheInfo `json:"cache_info,omitempty"`
}

// PerformanceAnalysis 性能分析
type PerformanceAnalysis struct {
	DurationMs   int64  `json:"duration_ms"`
	Grade        string `json:"grade"`
	RequestSize  int64  `json:"request_size"`
	ResponseSize int64  `json:"response_size"`
}

// SecurityAnalysis 安全分析
type SecurityAnalysis struct {
	Score  int             `json:"score"`
	Issues []SecurityIssue `json:"issues"`
}

// SecurityIssue 安全问题
type SecurityIssue struct {
	Severity    string `json:"severity"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// SensitiveInfo 敏感信息
type SensitiveInfo struct {
	Type     string `json:"type"`
	Location string `json:"location"`
	Count    int    `json:"count"`
}

// CacheInfo 缓存信息
type CacheInfo struct {
	HasCacheControl bool   `json:"has_cache_control"`
	CacheControl    string `json:"cache_control,omitempty"`
	HasETag         bool   `json:"has_etag"`
	ETag            string `json:"etag,omitempty"`
	HasLastModified bool   `json:"has_last_modified"`
	LastModified    string `json:"last_modified,omitempty"`
}

// detectRequestType 检测请求类型
func detectRequestType(contentType string) string {
	ct := strings.ToLower(contentType)
	switch {
	case strings.Contains(ct, "json"):
		return "json"
	case strings.Contains(ct, "xml"):
		return "xml"
	case strings.Contains(ct, "form"):
		return "form"
	case strings.Contains(ct, "multipart"):
		return "multipart"
	case strings.Contains(ct, "text"):
		return "text"
	case strings.Contains(ct, "html"):
		return "html"
	default:
		return "binary"
	}
}

// categorizeStatus 状态码分类
func categorizeStatus(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "success"
	case code >= 300 && code < 400:
		return "redirect"
	case code >= 400 && code < 500:
		return "client_error"
	case code >= 500:
		return "server_error"
	default:
		return "unknown"
	}
}

// detectSensitiveInfo 检测敏感信息
func detectSensitiveInfo(req *RequestData) []SensitiveInfo {
	var infos []SensitiveInfo

	body := string(req.Body)

	// 检测邮箱
	if strings.Contains(body, "@") && strings.Contains(body, ".") {
		infos = append(infos, SensitiveInfo{
			Type:     "email",
			Location: "body",
		})
	}

	// 检测手机号（简化检测）
	if len(body) > 0 {
		for _, prefix := range []string{"13", "14", "15", "16", "17", "18", "19"} {
			if strings.Contains(body, prefix) {
				infos = append(infos, SensitiveInfo{
					Type:     "phone",
					Location: "body",
				})
				break
			}
		}
	}

	// 检测密码字段
	passwordFields := []string{"password", "passwd", "pwd", "secret", "token"}
	for _, field := range passwordFields {
		if strings.Contains(strings.ToLower(body), field) {
			infos = append(infos, SensitiveInfo{
				Type:     "credential",
				Location: "body",
			})
			break
		}
	}

	return infos
}

// detectCacheInfo 检测缓存信息
func detectCacheInfo(headers map[string][]string) *CacheInfo {
	info := &CacheInfo{}

	if cc := getHeader(headers, "Cache-Control"); cc != "" {
		info.HasCacheControl = true
		info.CacheControl = cc
	}

	if etag := getHeader(headers, "ETag"); etag != "" {
		info.HasETag = true
		info.ETag = etag
	}

	if lm := getHeader(headers, "Last-Modified"); lm != "" {
		info.HasLastModified = true
		info.LastModified = lm
	}

	return info
}

// getHeader 获取 header 值
func getHeader(headers map[string][]string, key string) string {
	if values, ok := headers[key]; ok && len(values) > 0 {
		return values[0]
	}
	// 尝试不区分大小写
	for k, values := range headers {
		if strings.EqualFold(k, key) && len(values) > 0 {
			return values[0]
		}
	}
	return ""
}

// ToJSON 转换为 JSON
func (a *AnalysisResult) ToJSON() ([]byte, error) {
	return json.Marshal(a)
}
