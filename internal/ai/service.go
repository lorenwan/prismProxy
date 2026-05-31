package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"prismproxy/internal/traffic"
)

// Service AI 服务主控
type Service struct {
	config    *Config
	providers map[ProviderType]Provider
	mu        sync.RWMutex
}

// NewService 创建 AI 服务
func NewService(config *Config) *Service {
	s := &Service{
		config:    config,
		providers: make(map[ProviderType]Provider),
	}

	// 初始化提供商
	s.initProviders()

	return s
}

// initProviders 初始化所有配置的提供商
func (s *Service) initProviders() {
	if s.config.OpenAI != nil {
		s.providers[ProviderOpenAI] = NewOpenAIProvider(s.config.OpenAI)
	}
	if s.config.Claude != nil {
		s.providers[ProviderClaude] = NewClaudeProvider(s.config.Claude)
	}
	if s.config.Ollama != nil {
		s.providers[ProviderOllama] = NewOllamaProvider(s.config.Ollama)
	}
}

// getProvider 获取指定提供商，未指定则使用默认
func (s *Service) getProvider(pt ProviderType) (Provider, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if pt == "" {
		pt = s.config.DefaultProvider
	}

	p, ok := s.providers[pt]
	if !ok {
		return nil, fmt.Errorf("未找到提供商: %s", pt)
	}

	if !p.IsAvailable() {
		return nil, fmt.Errorf("提供商不可用: %s", pt)
	}

	return p, nil
}

// Chat 同步聊天
func (s *Service) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	p, err := s.getProvider(ProviderType(req.Model))
	if err != nil {
		// Model 不是 provider，使用默认
		p, err = s.getProvider("")
		if err != nil {
			return nil, err
		}
	}

	return p.Chat(ctx, req)
}

// StreamChat 流式聊天
func (s *Service) StreamChat(ctx context.Context, req *ChatRequest) (<-chan *StreamChunk, error) {
	p, err := s.getProvider(ProviderType(req.Model))
	if err != nil {
		p, err = s.getProvider("")
		if err != nil {
			return nil, err
		}
	}

	return p.StreamChat(ctx, req)
}

// GetAvailableProviders 获取可用的提供商列表
func (s *Service) GetAvailableProviders() []ProviderType {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var available []ProviderType
	for pt, p := range s.providers {
		if p.IsAvailable() {
			available = append(available, pt)
		}
	}
	return available
}

// UpdateConfig 更新配置
func (s *Service) UpdateConfig(config *Config) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config = config
	s.providers = make(map[ProviderType]Provider)
	s.initProviders()
}

// ==================== 流量分析 ====================

// AnalyzeTraffic 分析流量
func (s *Service) AnalyzeTraffic(ctx context.Context, txs []*traffic.Transaction) (*AnalysisResult, error) {
	if len(txs) == 0 {
		return &AnalysisResult{Summary: "无流量数据"}, nil
	}

	prompt := buildAnalyzerPrompt(txs)

	req := &ChatRequest{
		Messages: []ChatMessage{
			{Role: "system", Content: analyzerSystemPrompt},
			{Role: "user", Content: prompt},
		},
	}

	resp, err := s.Chat(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("AI 分析失败: %w", err)
	}

	result, err := parseAnalysisResult(resp.Content)
	if err != nil {
		return &AnalysisResult{Summary: resp.Content}, nil
	}

	return result, nil
}

func buildAnalyzerPrompt(txs []*traffic.Transaction) string {
	var sb strings.Builder
	sb.WriteString("请分析以下 HTTP 流量数据：\n\n")

	limit := 50
	if len(txs) < limit {
		limit = len(txs)
	}

	for i, tx := range txs[:limit] {
		sb.WriteString(fmt.Sprintf("## 请求 %d\n", i+1))
		sb.WriteString(fmt.Sprintf("- 方法: %s\n", tx.Method))
		sb.WriteString(fmt.Sprintf("- URL: %s\n", tx.URL))
		sb.WriteString(fmt.Sprintf("- 状态码: %d\n", tx.Response.StatusCode))
		sb.WriteString(fmt.Sprintf("- 耗时: %dms\n", tx.DurationMs))

		if tx.Request.ContentType != "" {
			sb.WriteString(fmt.Sprintf("- Content-Type: %s\n", tx.Request.ContentType))
		}
		sb.WriteString("\n")
	}

	if len(txs) > limit {
		sb.WriteString(fmt.Sprintf("... 共 %d 条记录，仅展示前 %d 条\n", len(txs), limit))
	}

	return sb.String()
}

func parseAnalysisResult(content string) (*AnalysisResult, error) {
	var result AnalysisResult

	if err := json.Unmarshal([]byte(content), &result); err != nil {
		start := strings.Index(content, "{")
		end := strings.LastIndex(content, "}")
		if start >= 0 && end > start {
			jsonStr := content[start : end+1]
			if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
				return nil, err
			}
			return &result, nil
		}
		return nil, err
	}

	return &result, nil
}

const analyzerSystemPrompt = `你是一个 HTTP 流量分析专家。请分析提供的流量数据，识别潜在问题并提供优化建议。

请以 JSON 格式返回分析结果：
{
  "summary": "整体分析摘要",
  "issues": [
    {
      "severity": "high|medium|low",
      "type": "security|performance|error",
      "title": "问题标题",
      "detail": "问题详情"
    }
  ],
  "suggestions": [
    {
      "category": "performance|security|best_practice",
      "title": "建议标题",
      "detail": "建议详情"
    }
  ]
}

分析要点：
1. 性能问题：响应时间过长、频繁请求、大文件传输
2. 安全风险：敏感信息泄露、未加密传输、可疑请求
3. 错误模式：频繁的 4xx/5xx 错误、超时
4. 最佳实践：RESTful 规范、缓存使用、压缩`

// ==================== 安全检测 ====================

// SecurityCheck 安全检测
func (s *Service) SecurityCheck(ctx context.Context, tx *traffic.Transaction) (*SecurityReport, error) {
	if tx == nil {
		return &SecurityReport{RiskLevel: "safe", Summary: "无请求数据"}, nil
	}

	prompt := buildSecurityPrompt(tx)

	req := &ChatRequest{
		Messages: []ChatMessage{
			{Role: "system", Content: securitySystemPrompt},
			{Role: "user", Content: prompt},
		},
	}

	resp, err := s.Chat(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("安全检测失败: %w", err)
	}

	report, err := parseSecurityReport(resp.Content)
	if err != nil {
		return &SecurityReport{
			RiskLevel: "unknown",
			Summary:   resp.Content,
		}, nil
	}

	return report, nil
}

func buildSecurityPrompt(tx *traffic.Transaction) string {
	var sb strings.Builder
	sb.WriteString("请检测以下 HTTP 请求的安全性：\n\n")

	sb.WriteString("## 请求信息\n")
	sb.WriteString(fmt.Sprintf("- 方法: %s\n", tx.Method))
	sb.WriteString(fmt.Sprintf("- URL: %s\n", tx.URL))
	sb.WriteString(fmt.Sprintf("- Host: %s\n", tx.Host))

	if tx.Request.ContentType != "" {
		sb.WriteString(fmt.Sprintf("- Content-Type: %s\n", tx.Request.ContentType))
	}

	sb.WriteString("\n## 请求头\n")
	for k, v := range tx.Request.Headers {
		if !isSensitiveHeader(k) {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", k, strings.Join(v, ", ")))
		} else {
			sb.WriteString(fmt.Sprintf("- %s: [已隐藏]\n", k))
		}
	}

	if len(tx.Request.Body) > 0 {
		sb.WriteString("\n## 请求体\n")
		body := string(tx.Request.Body)
		if len(body) > 1000 {
			body = body[:1000] + "...(已截断)"
		}
		sb.WriteString(body)
	}

	sb.WriteString("\n## 响应信息\n")
	sb.WriteString(fmt.Sprintf("- 状态码: %d\n", tx.Response.StatusCode))
	sb.WriteString(fmt.Sprintf("- Content-Type: %s\n", tx.Response.ContentType))

	if len(tx.Response.Body) > 0 {
		sb.WriteString("\n## 响应体\n")
		body := string(tx.Response.Body)
		if len(body) > 1000 {
			body = body[:1000] + "...(已截断)"
		}
		sb.WriteString(body)
	}

	return sb.String()
}

func isSensitiveHeader(name string) bool {
	sensitive := []string{
		"authorization", "cookie", "set-cookie",
		"x-api-key", "x-auth-token", "proxy-authorization",
	}
	lower := strings.ToLower(name)
	for _, s := range sensitive {
		if lower == s {
			return true
		}
	}
	return false
}

func parseSecurityReport(content string) (*SecurityReport, error) {
	var report SecurityReport

	if err := json.Unmarshal([]byte(content), &report); err != nil {
		start := strings.Index(content, "{")
		end := strings.LastIndex(content, "}")
		if start >= 0 && end > start {
			jsonStr := content[start : end+1]
			if err := json.Unmarshal([]byte(jsonStr), &report); err != nil {
				return nil, err
			}
			return &report, nil
		}
		return nil, err
	}

	return &report, nil
}

const securitySystemPrompt = `你是一个 Web 安全专家。请检测提供的 HTTP 请求是否存在安全风险。

请以 JSON 格式返回检测结果：
{
  "risk_level": "critical|high|medium|low|safe",
  "summary": "整体风险评估",
  "findings": [
    {
      "severity": "critical|high|medium|low",
      "category": "injection|auth|exposure|misconfiguration|other",
      "title": "问题标题",
      "description": "问题描述",
      "location": "问题位置（请求头/请求体/URL等）",
      "remediation": "修复建议"
    }
  ]
}

检测要点：
1. SQL 注入：URL 参数、请求体中的 SQL 特征
2. XSS：反射型/存储型 XSS 特征
3. 敏感信息泄露：密码、token、密钥明文传输
4. CSRF：缺少 CSRF token
5. 路径遍历：../ 等特征
6. 命令注入：系统命令特征
7. 不安全的 HTTP 方法：PUT/DELETE 未授权
8. 缺少安全头：HSTS、CSP、X-Frame-Options 等`

// ==================== 测试用例生成 ====================

// GenerateTests 生成测试用例
func (s *Service) GenerateTests(ctx context.Context, txs []*traffic.Transaction) ([]*TestCase, error) {
	if len(txs) == 0 {
		return nil, nil
	}

	prompt := buildTestgenPrompt(txs)

	req := &ChatRequest{
		Messages: []ChatMessage{
			{Role: "system", Content: testgenSystemPrompt},
			{Role: "user", Content: prompt},
		},
	}

	resp, err := s.Chat(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("生成测试失败: %w", err)
	}

	tests, err := parseTestCases(resp.Content)
	if err != nil {
		return nil, fmt.Errorf("解析测试用例失败: %w", err)
	}

	return tests, nil
}

func buildTestgenPrompt(txs []*traffic.Transaction) string {
	var sb strings.Builder
	sb.WriteString("请根据以下 HTTP 请求生成 API 测试用例：\n\n")

	limit := 10
	if len(txs) < limit {
		limit = len(txs)
	}

	for i, tx := range txs[:limit] {
		sb.WriteString(fmt.Sprintf("## 请求 %d\n", i+1))
		sb.WriteString(fmt.Sprintf("- 方法: %s\n", tx.Method))
		sb.WriteString(fmt.Sprintf("- URL: %s\n", tx.URL))
		sb.WriteString(fmt.Sprintf("- 状态码: %d\n", tx.Response.StatusCode))

		if len(tx.Request.Body) > 0 {
			body := string(tx.Request.Body)
			if len(body) > 500 {
				body = body[:500] + "..."
			}
			sb.WriteString(fmt.Sprintf("- 请求体: %s\n", body))
		}

		if len(tx.Response.Body) > 0 {
			body := string(tx.Response.Body)
			if len(body) > 500 {
				body = body[:500] + "..."
			}
			sb.WriteString(fmt.Sprintf("- 响应体: %s\n", body))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func parseTestCases(content string) ([]*TestCase, error) {
	var tests []*TestCase

	if err := json.Unmarshal([]byte(content), &tests); err != nil {
		start := strings.Index(content, "[")
		end := strings.LastIndex(content, "]")
		if start >= 0 && end > start {
			jsonStr := content[start : end+1]
			if err := json.Unmarshal([]byte(jsonStr), &tests); err != nil {
				return nil, err
			}
			return tests, nil
		}
		return nil, err
	}

	return tests, nil
}

const testgenSystemPrompt = `你是一个 API 测试专家。请根据提供的 HTTP 请求生成测试用例。

请以 JSON 数组格式返回测试用例：
[
  {
    "name": "测试名称",
    "description": "测试描述",
    "method": "GET|POST|PUT|DELETE",
    "url": "请求 URL",
    "headers": {
      "Content-Type": "application/json"
    },
    "body": "请求体（如需要）",
    "assertions": [
      {
        "type": "status|header|body|time",
        "target": "断言目标（如状态码、header 名、JSON 路径）",
        "operator": "eq|contains|matches|lt|gt",
        "value": "期望值"
      }
    ]
  }
]

测试类型：
1. 正向测试：验证正常请求和响应
2. 边界测试：空值、超长字符串、特殊字符
3. 错误测试：无效参数、未授权访问
4. 性能测试：响应时间断言

每个请求生成 2-3 个测试用例，覆盖不同场景。`

// ==================== AI 助手 ====================

// ChatAssistant AI 助手对话
func (s *Service) ChatAssistant(ctx context.Context, message string, history []ChatMessage) (<-chan *StreamChunk, error) {
	messages := []ChatMessage{
		{Role: "system", Content: assistantSystemPrompt},
	}

	// 限制历史长度
	if len(history) > 20 {
		history = history[len(history)-20:]
	}
	messages = append(messages, history...)

	messages = append(messages, ChatMessage{
		Role:    "user",
		Content: message,
	})

	req := &ChatRequest{
		Messages: messages,
		Stream:   true,
	}

	chunks, err := s.StreamChat(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("启动对话失败: %w", err)
	}

	return chunks, nil
}

const assistantSystemPrompt = `你是 PrismProxy 的 AI 助手，专注于 HTTP 调试和网络分析。

你的能力：
1. 解答 HTTP/HTTPS 相关问题
2. 分析请求/响应格式
3. 解释状态码和错误信息
4. 提供网络调试建议
5. 协助编写正则表达式和 JSONPath
6. 解释 TLS/SSL 证书问题
7. 优化 API 性能建议

回复要求：
- 简洁明了，直接回答问题
- 使用中文回复
- 技术术语保持英文
- 必要时提供代码示例
- 如果不确定，坦诚说明`
