package ai

import "context"

// Provider AI 提供商接口
type Provider interface {
	// Chat 同步聊天请求
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

	// StreamChat 流式聊天请求
	StreamChat(ctx context.Context, req *ChatRequest) (<-chan *StreamChunk, error)

	// IsAvailable 检查提供商是否可用
	IsAvailable() bool

	// Name 提供商名称
	Name() string
}
