package ai

import "time"

// ProviderType AI 提供商类型
type ProviderType string

const (
	ProviderOpenAI ProviderType = "openai"
	ProviderClaude ProviderType = "claude"
	ProviderOllama ProviderType = "ollama"
)

// Config AI 配置
type Config struct {
	// 默认提供商
	DefaultProvider ProviderType `json:"default_provider"`

	// OpenAI 配置
	OpenAI *OpenAIConfig `json:"openai,omitempty"`

	// Claude 配置
	Claude *ClaudeConfig `json:"claude,omitempty"`

	// Ollama 配置
	Ollama *OllamaConfig `json:"ollama,omitempty"`
}

// OpenAIConfig OpenAI 配置
type OpenAIConfig struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url,omitempty"` // 自定义 API 地址
	Model   string `json:"model,omitempty"`    // 默认模型
}

// ClaudeConfig Claude 配置
type ClaudeConfig struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url,omitempty"`
	Model   string `json:"model,omitempty"`
}

// OllamaConfig Ollama 配置
type OllamaConfig struct {
	BaseURL string `json:"base_url"` // Ollama 服务地址
	Model   string `json:"model"`    // 模型名称
}

// ChatMessage 聊天消息
type ChatMessage struct {
	Role    string `json:"role"`    // system, user, assistant
	Content string `json:"content"`
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Messages []ChatMessage `json:"messages"`
	Model    string        `json:"model,omitempty"` // 可选，覆盖默认模型
	Stream   bool          `json:"stream"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	Content   string `json:"content"`
	Provider  string `json:"provider"`
	Model     string `json:"model"`
	Usage     *Usage `json:"usage,omitempty"`
}

// Usage token 用量
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamChunk 流式响应块
type StreamChunk struct {
	Content  string `json:"content"`
	Done     bool   `json:"done"`
	Provider string `json:"provider"`
}

// AnalysisResult 流量分析结果
type AnalysisResult struct {
	Summary    string        `json:"summary"`
	Issues     []Issue       `json:"issues,omitempty"`
	Suggestions []Suggestion `json:"suggestions,omitempty"`
}

// Issue 发现的问题
type Issue struct {
	Severity string `json:"severity"` // high, medium, low
	Type     string `json:"type"`     // security, performance, error
	Title    string `json:"title"`
	Detail   string `json:"detail"`
}

// Suggestion 优化建议
type Suggestion struct {
	Category string `json:"category"` // performance, security, best_practice
	Title    string `json:"title"`
	Detail   string `json:"detail"`
}

// SecurityReport 安全检测报告
type SecurityReport struct {
	RiskLevel string         `json:"risk_level"` // critical, high, medium, low, safe
	Findings  []SecurityFinding `json:"findings"`
	Summary   string         `json:"summary"`
}

// SecurityFinding 安全发现
type SecurityFinding struct {
	Severity    string `json:"severity"`
	Category    string `json:"category"` // injection, auth, exposure, etc
	Title       string `json:"title"`
	Description string `json:"description"`
	Location    string `json:"location"` // 请求/响应中的位置
	Remediation string `json:"remediation"` // 修复建议
}

// TestCase 生成的测试用例
type TestCase struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Method      string            `json:"method"`
	URL         string            `json:"url"`
	Headers     map[string]string `json:"headers,omitempty"`
	Body        string            `json:"body,omitempty"`
	Assertions  []Assertion       `json:"assertions"`
}

// Assertion 测试断言
type Assertion struct {
	Type     string `json:"type"` // status, header, body, time
	Target   string `json:"target"`
	Operator string `json:"operator"` // eq, contains, matches, lt, gt
	Value    string `json:"value"`
}

// AnalysisTimeRange 分析时间范围
type AnalysisTimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}
