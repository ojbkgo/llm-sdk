package api

// Role 定义消息的角色类型
type Role string

const (
	// RoleSystem 系统消息角色
	RoleSystem Role = "system"
	// RoleUser 用户消息角色
	RoleUser Role = "user"
	// RoleAssistant 助手消息角色
	RoleAssistant Role = "assistant"
)

// Message 定义对话消息
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

// Request 定义请求参数
type Request struct {
	// 必填字段
	Messages []Message `json:"messages"`
	Model    string    `json:"model"`

	// 可选参数
	Temperature      *float64 `json:"temperature,omitempty"`
	TopP             *float64 `json:"top_p,omitempty"`
	MaxTokens        *int     `json:"max_tokens,omitempty"`
	PresencePenalty  *float64 `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`
	Stop             []string `json:"stop,omitempty"`
	Stream           bool     `json:"stream,omitempty"`

	// 自定义字段，用于提供商特定的参数
	ExtraParams map[string]interface{} `json:"-"`
}

// Response 定义完整响应
type Response struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// ResponseChunk 定义流式响应的数据块
type ResponseChunk struct {
	ID      string        `json:"id"`
	Object  string        `json:"object"`
	Created int64         `json:"created"`
	Model   string        `json:"model"`
	Choices []ChunkChoice `json:"choices"`
}

// Choice 定义响应中的选择
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// ChunkChoice 定义流式响应中的选择
type ChunkChoice struct {
	Index        int     `json:"index"`
	Delta        Message `json:"delta"`
	FinishReason string  `json:"finish_reason,omitempty"`
}

// Usage 定义令牌使用情况
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
