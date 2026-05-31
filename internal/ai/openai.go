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

// OpenAIProvider OpenAI 提供商
type OpenAIProvider struct {
	config       *OpenAIConfig
	client       *http.Client
	defaultModel string
}

// openaiRequest OpenAI API 请求
type openaiRequest struct {
	Model    string          `json:"model"`
	Messages []openaiMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// openaiResponse OpenAI API 响应
type openaiResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// openaiStreamResponse 流式响应
type openaiStreamResponse struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

// NewOpenAIProvider 创建 OpenAI 提供商
func NewOpenAIProvider(config *OpenAIConfig) *OpenAIProvider {
	model := config.Model
	if model == "" {
		model = "gpt-4o-mini"
	}

	return &OpenAIProvider{
		config:       config,
		client:       &http.Client{},
		defaultModel: model,
	}
}

// Name 提供商名称
func (p *OpenAIProvider) Name() string {
	return string(ProviderOpenAI)
}

// IsAvailable 检查是否可用
func (p *OpenAIProvider) IsAvailable() bool {
	return p.config != nil && p.config.APIKey != ""
}

// Chat 同步聊天
func (p *OpenAIProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if !p.IsAvailable() {
		return nil, fmt.Errorf("OpenAI 提供商未配置")
	}

	model := p.defaultModel
	if req.Model != "" {
		model = req.Model
	}

	// 构建请求
	apiReq := openaiRequest{
		Model:  model,
		Stream: false,
	}
	for _, msg := range req.Messages {
		apiReq.Messages = append(apiReq.Messages, openaiMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	body, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 发送请求
	baseURL := p.config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var apiResp openaiResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if apiResp.Error != nil {
		return nil, fmt.Errorf("API 错误: %s", apiResp.Error.Message)
	}

	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("无响应内容")
	}

	result := &ChatResponse{
		Content:  apiResp.Choices[0].Message.Content,
		Provider: p.Name(),
		Model:    model,
	}

	if apiResp.Usage != nil {
		result.Usage = &Usage{
			PromptTokens:     apiResp.Usage.PromptTokens,
			CompletionTokens: apiResp.Usage.CompletionTokens,
			TotalTokens:      apiResp.Usage.TotalTokens,
		}
	}

	return result, nil
}

// StreamChat 流式聊天
func (p *OpenAIProvider) StreamChat(ctx context.Context, req *ChatRequest) (<-chan *StreamChunk, error) {
	if !p.IsAvailable() {
		return nil, fmt.Errorf("OpenAI 提供商未配置")
	}

	model := p.defaultModel
	if req.Model != "" {
		model = req.Model
	}

	apiReq := openaiRequest{
		Model:  model,
		Stream: true,
	}
	for _, msg := range req.Messages {
		apiReq.Messages = append(apiReq.Messages, openaiMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	body, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	baseURL := p.config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)

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
			if data == "[DONE]" {
				chunks <- &StreamChunk{Done: true, Provider: p.Name()}
				return
			}

			var streamResp openaiStreamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				continue
			}

			if len(streamResp.Choices) > 0 && streamResp.Choices[0].Delta.Content != "" {
				chunks <- &StreamChunk{
					Content:  streamResp.Choices[0].Delta.Content,
					Provider: p.Name(),
				}
			}
		}
	}()

	return chunks, nil
}
