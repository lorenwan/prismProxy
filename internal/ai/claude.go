package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ClaudeProvider Claude 提供商
type ClaudeProvider struct {
	config       *ClaudeConfig
	client       *http.Client
	defaultModel string
}

// claudeRequest Claude API 请求
type claudeRequest struct {
	Model     string          `json:"model"`
	Messages  []claudeMessage `json:"messages"`
	MaxTokens int             `json:"max_tokens"`
	Stream    bool            `json:"stream"`
	System    string          `json:"system,omitempty"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// claudeResponse Claude API 响应
type claudeResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage *struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// claudeStreamEvent 流式事件
type claudeStreamEvent struct {
	Type  string `json:"type"`
	Delta *struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"delta,omitempty"`
}

// NewClaudeProvider 创建 Claude 提供商
func NewClaudeProvider(config *ClaudeConfig) *ClaudeProvider {
	model := config.Model
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}

	return &ClaudeProvider{
		config:       config,
		client:       &http.Client{},
		defaultModel: model,
	}
}

// Name 提供商名称
func (p *ClaudeProvider) Name() string {
	return string(ProviderClaude)
}

// IsAvailable 检查是否可用
func (p *ClaudeProvider) IsAvailable() bool {
	return p.config != nil && p.config.APIKey != ""
}

// Chat 同步聊天
func (p *ClaudeProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if !p.IsAvailable() {
		return nil, fmt.Errorf("Claude 提供商未配置")
	}

	model := p.defaultModel
	if req.Model != "" {
		model = req.Model
	}

	// 分离 system 消息
	var system string
	var messages []claudeMessage
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			system = msg.Content
		} else {
			messages = append(messages, claudeMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	apiReq := claudeRequest{
		Model:     model,
		Messages:  messages,
		MaxTokens: 4096,
		Stream:    false,
	}
	if system != "" {
		apiReq.System = system
	}

	body, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	baseURL := p.config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var apiResp claudeResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if apiResp.Error != nil {
		return nil, fmt.Errorf("API 错误: %s", apiResp.Error.Message)
	}

	if len(apiResp.Content) == 0 {
		return nil, fmt.Errorf("无响应内容")
	}

	result := &ChatResponse{
		Content:  apiResp.Content[0].Text,
		Provider: p.Name(),
		Model:    model,
	}

	if apiResp.Usage != nil {
		result.Usage = &Usage{
			PromptTokens:     apiResp.Usage.InputTokens,
			CompletionTokens: apiResp.Usage.OutputTokens,
			TotalTokens:      apiResp.Usage.InputTokens + apiResp.Usage.OutputTokens,
		}
	}

	return result, nil
}

// StreamChat 流式聊天
func (p *ClaudeProvider) StreamChat(ctx context.Context, req *ChatRequest) (<-chan *StreamChunk, error) {
	if !p.IsAvailable() {
		return nil, fmt.Errorf("Claude 提供商未配置")
	}

	model := p.defaultModel
	if req.Model != "" {
		model = req.Model
	}

	var system string
	var messages []claudeMessage
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			system = msg.Content
		} else {
			messages = append(messages, claudeMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	apiReq := claudeRequest{
		Model:     model,
		Messages:  messages,
		MaxTokens: 4096,
		Stream:    true,
	}
	if system != "" {
		apiReq.System = system
	}

	body, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	baseURL := p.config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("API 返回错误状态: %d", resp.StatusCode)
	}

	chunks := make(chan *StreamChunk, 100)

	go func() {
		defer resp.Body.Close()
		defer close(chunks)

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
			}

			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")

			var event claudeStreamEvent
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				continue
			}

			switch event.Type {
			case "content_block_delta":
				if event.Delta != nil && event.Delta.Text != "" {
					chunks <- &StreamChunk{
						Content:  event.Delta.Text,
						Provider: p.Name(),
					}
				}
			case "message_stop":
				chunks <- &StreamChunk{Done: true, Provider: p.Name()}
				return
			}
		}
	}()

	return chunks, nil
}
