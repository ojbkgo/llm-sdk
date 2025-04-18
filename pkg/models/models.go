package models

// 定义不同提供商的模型常量

// OpenAI 模型
const (
	// GPT4 是OpenAI的GPT-4模型
	GPT4 = "gpt-4"
	// GPT4Turbo 是OpenAI的GPT-4 Turbo模型
	GPT4Turbo = "gpt-4-turbo"
	// GPT4TurboPreview 是OpenAI的GPT-4 Turbo预览版
	GPT4TurboPreview = "gpt-4-turbo-preview"
	// GPT4o 是OpenAI的GPT-4o模型
	GPT4o = "gpt-4o"
	// GPT35Turbo 是OpenAI的GPT-3.5 Turbo模型
	GPT35Turbo = "gpt-3.5-turbo"
	// GPT35TurboInstruct 是OpenAI的GPT-3.5 Turbo Instruct模型
	GPT35TurboInstruct = "gpt-3.5-turbo-instruct"
	// TextEmbeddingAda002 是OpenAI的嵌入模型
	TextEmbeddingAda002 = "text-embedding-ada-002"
	// TextEmbedding3Small 是OpenAI的小型嵌入模型
	TextEmbedding3Small = "text-embedding-3-small"
	// TextEmbedding3Large 是OpenAI的大型嵌入模型
	TextEmbedding3Large = "text-embedding-3-large"
)

// Anthropic 模型
const (
	// ClaudeHaiku 是Anthropic的轻量级Claude模型
	ClaudeHaiku = "claude-haiku"
	// ClaudeSonnet 是Anthropic的中型Claude模型
	ClaudeSonnet = "claude-sonnet"
	// ClaudeOpus 是Anthropic的高能力Claude模型
	ClaudeOpus = "claude-opus"
	// Claude3Haiku 是Anthropic的Claude 3 Haiku模型
	Claude3Haiku = "claude-3-haiku"
	// Claude3Sonnet 是Anthropic的Claude 3 Sonnet模型
	Claude3Sonnet = "claude-3-sonnet"
	// Claude3Opus 是Anthropic的Claude 3 Opus模型
	Claude3Opus = "claude-3-opus"
)

// Google 模型
const (
	// GeminiPro 是Google的Gemini Pro模型
	GeminiPro = "gemini-pro"
	// GeminiProVision 是Google的Gemini Pro Vision模型
	GeminiProVision = "gemini-pro-vision"
	// GeminiUltra 是Google的Gemini Ultra模型
	GeminiUltra = "gemini-ultra"
)

// DeepSeek 模型
const (
	// DeepSeekCoder 是DeepSeek的代码模型
	DeepSeekCoder = "deepseek-coder"
	// DeepSeekChat 是DeepSeek的通用聊天模型
	DeepSeekChat = "deepseek-chat"
	// DeepSeekLlama270B 是DeepSeek的70B大模型
	DeepSeekLlama270B = "deepseek-llama-70b"
	// DeepSeekLlama7B 是DeepSeek的7B模型
	DeepSeekLlama7B = "deepseek-llama-7b"
	// DeepSeekMoE 是DeepSeek的稀疏MoE模型
	DeepSeekMoE = "deepseek-moe"
	// DeepSeekEmbedding 是DeepSeek的嵌入模型
	DeepSeekEmbedding = "deepseek-embedding"
)

// ModelInfo 存储模型相关信息
type ModelInfo struct {
	ID           string
	Provider     string
	MaxTokens    int
	InputPrice   float64 // 每1000个输入token的价格（美元）
	OutputPrice  float64 // 每1000个输出token的价格（美元）
	Capabilities []string
}

// 模型能力常量
const (
	CapabilityChat      = "chat"
	CapabilityVision    = "vision"
	CapabilityFunction  = "function"
	CapabilityEmbedding = "embedding"
	CapabilityCoding    = "coding"
)

// GetModelInfo 返回指定模型的信息
func GetModelInfo(modelID string) *ModelInfo {
	if info, ok := modelRegistry[modelID]; ok {
		return &info
	}
	return nil
}

// 模型注册表
var modelRegistry = map[string]ModelInfo{
	GPT4: {
		ID:           GPT4,
		Provider:     "openai",
		MaxTokens:    8192,
		InputPrice:   0.03,
		OutputPrice:  0.06,
		Capabilities: []string{CapabilityChat, CapabilityFunction},
	},
	GPT4o: {
		ID:           GPT4o,
		Provider:     "openai",
		MaxTokens:    128000,
		InputPrice:   0.005,
		OutputPrice:  0.015,
		Capabilities: []string{CapabilityChat, CapabilityVision, CapabilityFunction},
	},
	GPT35Turbo: {
		ID:           GPT35Turbo,
		Provider:     "openai",
		MaxTokens:    16385,
		InputPrice:   0.0015,
		OutputPrice:  0.002,
		Capabilities: []string{CapabilityChat, CapabilityFunction},
	},
	Claude3Opus: {
		ID:           Claude3Opus,
		Provider:     "anthropic",
		MaxTokens:    200000,
		InputPrice:   0.015,
		OutputPrice:  0.075,
		Capabilities: []string{CapabilityChat, CapabilityVision},
	},
	Claude3Sonnet: {
		ID:           Claude3Sonnet,
		Provider:     "anthropic",
		MaxTokens:    200000,
		InputPrice:   0.003,
		OutputPrice:  0.015,
		Capabilities: []string{CapabilityChat, CapabilityVision},
	},
	Claude3Haiku: {
		ID:           Claude3Haiku,
		Provider:     "anthropic",
		MaxTokens:    200000,
		InputPrice:   0.00025,
		OutputPrice:  0.00125,
		Capabilities: []string{CapabilityChat, CapabilityVision},
	},
	GeminiPro: {
		ID:           GeminiPro,
		Provider:     "google",
		MaxTokens:    32768,
		InputPrice:   0.00125,
		OutputPrice:  0.00125,
		Capabilities: []string{CapabilityChat, CapabilityFunction},
	},
	GeminiUltra: {
		ID:           GeminiUltra,
		Provider:     "google",
		MaxTokens:    32768,
		InputPrice:   0.00375,
		OutputPrice:  0.01125,
		Capabilities: []string{CapabilityChat, CapabilityVision, CapabilityFunction},
	},
	DeepSeekCoder: {
		ID:           DeepSeekCoder,
		Provider:     "deepseek",
		MaxTokens:    16000,
		InputPrice:   0.0005,
		OutputPrice:  0.0015,
		Capabilities: []string{CapabilityChat, CapabilityCoding},
	},
	DeepSeekChat: {
		ID:           DeepSeekChat,
		Provider:     "deepseek",
		MaxTokens:    8000,
		InputPrice:   0.001,
		OutputPrice:  0.002,
		Capabilities: []string{CapabilityChat},
	},
	DeepSeekLlama270B: {
		ID:           DeepSeekLlama270B,
		Provider:     "deepseek",
		MaxTokens:    32000,
		InputPrice:   0.002,
		OutputPrice:  0.006,
		Capabilities: []string{CapabilityChat},
	},
	DeepSeekEmbedding: {
		ID:           DeepSeekEmbedding,
		Provider:     "deepseek",
		MaxTokens:    0,
		InputPrice:   0.0001,
		OutputPrice:  0,
		Capabilities: []string{CapabilityEmbedding},
	},
}
