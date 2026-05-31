package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// OllamaProvider Ollama 提供商
type OllamaProvider struct {
	config       *OllamaConfig
	client       *http.Client
	defaultModel string
}

// ollamaRequest Ollama API 请求
type ollamaRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ollamaResponse Ollama API 响应
type ollamaResponse struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
	Done  bool `json:"done"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage,omitempty"`
	Error string `json:"error,omitempty"`
}

// NewOllamaProvider 创建 Ollama 提供商
func NewOllamaProvider(config *OllamaConfig) *OllamaProvider {
	model := config.Model
	if model == "" {
		model = "qwen2.5:7b"
	}

	return &OllamaProvider{
		config:       config,
		client:       &http.Client{},
		defaultModel: model,
	}
}

// Name 提供商名称
func (p *OllamaProvider) Name() string {
	return string(ProviderOllama)
}

// IsAvailable 检查是否可用（通过健康检查）
func (p *OllamaProvider) IsAvailable() bool {
	if p.config == nil || p.config.BaseURL == "" {
		return false
	}

	// 简单检查服务是否在线
	resp, err := p.client.Get(p.config.BaseURL + "/api/tags")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// Chat 同步聊天
func (p *OllamaProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if p.config == nil || p.config.BaseURL == "" {
		return nil, fmt.Errorf("Ollama 提供商未配置")
	}

	model := p.defaultModel
	if req.Model != "" {
		model = req.Model
	}

	apiReq := ollamaRequest{
		Model:  model,
		Stream: false,
	}
	for _, msg := range req.Messages {
		apiReq.Messages = append(apiReq.Messages, ollamaMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	body, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var apiResp ollamaResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if apiResp.Error != "" {
		return nil, fmt.Errorf("API 错误: %s", apiResp.Error)
	}

	result := &ChatResponse{
		Content:  apiResp.Message.Content,
		Provider: p.Name(),
		Model:    model,
	}

	if apiResp.Usage != nil {
		result.Usage = &Usage{
			PromptTokens:     apiResp.Usage.PromptTokens,
			CompletionTokens: apiResp.Usage.CompletionTokens,
			TotalTokens:      apiResp.Usage.PromptTokens + apiResp.Usage.CompletionTokens,
		}
	}

	return result, nil
}

// StreamChat 流式聊天
func (p *OllamaProvider) StreamChat(ctx context.Context, req *ChatRequest) (<-chan *StreamChunk, error) {
	if p.config == nil || p.config.BaseURL == "" {
		return nil, fmt.Errorf("Ollama 提供商未配置")
	}

	model := p.defaultModel
	if req.Model != "" {
		model = req.Model
	}

	apiReq := ollamaRequest{
		Model:  model,
		Stream: true,
	}
	for _, msg := range req.Messages {
		apiReq.Messages = append(apiReq.Messages, ollamaMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	body, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

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

			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}

			var apiResp ollamaResponse
			if err := json.Unmarshal(line, &apiResp); err != nil {
				continue
			}

			if apiResp.Message.Content != "" {
				chunks <- &StreamChunk{
					Content:  apiResp.Message.Content,
					Provider: p.Name(),
				}
			}

			if apiResp.Done {
				chunks <- &StreamChunk{Done: true, Provider: p.Name()}
				return
			}
		}
	}()

	return chunks, nil
}
