package api

import (
	"context"
)

// LLMClient 定义了与语言模型交互的统一接口
type LLMClient interface {
	// Complete 发送请求并获取完整的响应
	Complete(ctx context.Context, request *Request) (*Response, error)

	// CompleteStream 发送请求并获取流式响应
	CompleteStream(ctx context.Context, request *Request) (ResponseStream, error)

	// Embedding 获取文本的嵌入向量
	Embedding(ctx context.Context, input string) ([]float32, error)
}

// ResponseStream 定义了流式响应的接口
type ResponseStream interface {
	// Recv 接收下一个响应块，当没有更多响应时返回io.EOF错误
	Recv() (*ResponseChunk, error)

	// Close 关闭流
	Close() error
}

// Provider 定义了LLM提供商的接口
type Provider interface {
	// NewClient 创建提供商特定的客户端实现
	NewClient(options ...ClientOption) (LLMClient, error)
}

// ClientOption 定义了客户端配置选项
type ClientOption func(options *ClientOptions)

// ClientOptions 包含所有客户端配置
type ClientOptions struct {
	APIKey     string
	BaseURL    string
	HTTPClient interface{} // 使用时可以转换为具体的HTTP客户端类型
	Timeout    int
	MaxRetries int
}
