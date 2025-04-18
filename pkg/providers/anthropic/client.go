package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ojbkgo/llm-sdk/pkg/api"
	"github.com/ojbkgo/llm-sdk/pkg/utils"
)

// Client 实现了Anthropic的API客户端
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	maxRetries int
	apiVersion string
}

// 默认配置
const (
	defaultBaseURL    = "https://api.anthropic.com"
	defaultTimeout    = 60 * time.Second
	defaultMaxRetries = 3
	defaultAPIVersion = "2023-06-01"
)

// NewClient 创建一个新的Anthropic客户端
func NewClient(options ...api.ClientOption) (api.LLMClient, error) {
	clientOptions := &api.ClientOptions{
		BaseURL:    defaultBaseURL,
		Timeout:    int(defaultTimeout.Seconds()),
		MaxRetries: defaultMaxRetries,
	}

	// 应用选项
	for _, option := range options {
		option(clientOptions)
	}

	// 验证必要的配置
	if clientOptions.APIKey == "" {
		return nil, api.NewError(api.ErrorTypeAuthentication, "API密钥不能为空", 0, nil)
	}

	// 创建HTTP客户端
	httpClient := &http.Client{
		Timeout: time.Duration(clientOptions.Timeout) * time.Second,
	}
	if clientOptions.HTTPClient != nil {
		if client, ok := clientOptions.HTTPClient.(*http.Client); ok {
			httpClient = client
		}
	}

	return &Client{
		apiKey:     clientOptions.APIKey,
		baseURL:    clientOptions.BaseURL,
		httpClient: httpClient,
		maxRetries: clientOptions.MaxRetries,
		apiVersion: defaultAPIVersion,
	}, nil
}

// Complete 发送请求并获取完整的响应
func (c *Client) Complete(ctx context.Context, request *api.Request) (*api.Response, error) {
	// 验证请求
	if err := validateRequest(request); err != nil {
		return nil, err
	}

	// 准备请求体
	reqBody, err := json.Marshal(adaptRequest(request))
	if err != nil {
		return nil, api.NewError(api.ErrorTypeInvalidRequest, "无法序列化请求", 0, err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/messages", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, api.NewError(api.ErrorTypeConnection, "创建HTTP请求失败", 0, err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", c.apiKey)
	req.Header.Set("Anthropic-Version", c.apiVersion)

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, api.NewError(api.ErrorTypeConnection, "HTTP请求失败", 0, err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, api.NewError(api.ErrorTypeServer, "读取响应失败", resp.StatusCode, err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		var anthropicErr AnthropicError
		if err := json.Unmarshal(body, &anthropicErr); err != nil {
			return nil, api.NewError(api.ErrorTypeServer, fmt.Sprintf("API错误(状态码: %d)", resp.StatusCode), resp.StatusCode, nil)
		}
		return nil, mapAnthropicError(&anthropicErr, resp.StatusCode)
	}

	// 解析响应
	var anthropicResp AnthropicResponse
	if err := json.Unmarshal(body, &anthropicResp); err != nil {
		return nil, api.NewError(api.ErrorTypeServer, "解析响应失败", resp.StatusCode, err)
	}

	return adaptResponse(&anthropicResp), nil
}

// CompleteStream 发送请求并获取流式响应
func (c *Client) CompleteStream(ctx context.Context, request *api.Request) (api.ResponseStream, error) {
	// 验证请求
	if err := validateRequest(request); err != nil {
		return nil, err
	}

	// 设置流式标志
	reqCopy := *request
	reqCopy.Stream = true

	// 准备请求体
	reqBody, err := json.Marshal(adaptRequest(&reqCopy))
	if err != nil {
		return nil, api.NewError(api.ErrorTypeInvalidRequest, "无法序列化请求", 0, err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/messages", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, api.NewError(api.ErrorTypeConnection, "创建HTTP请求失败", 0, err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", c.apiKey)
	req.Header.Set("Anthropic-Version", c.apiVersion)
	req.Header.Set("Accept", "text/event-stream")

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, api.NewError(api.ErrorTypeConnection, "HTTP请求失败", 0, err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		var anthropicErr AnthropicError
		if err := json.Unmarshal(body, &anthropicErr); err != nil {
			return nil, api.NewError(api.ErrorTypeServer, fmt.Sprintf("API错误(状态码: %d)", resp.StatusCode), resp.StatusCode, nil)
		}
		return nil, mapAnthropicError(&anthropicErr, resp.StatusCode)
	}

	return &anthropicResponseStream{
		reader:    utils.NewSSEReader(resp.Body),
		rawReader: resp.Body,
	}, nil
}

// Embedding 获取文本的嵌入向量
func (c *Client) Embedding(ctx context.Context, input string) ([]float32, error) {
	// Anthropic 目前还没有公开的嵌入接口，所以这里返回未实现错误
	return nil, api.NewError(api.ErrorTypeUnknown, "Anthropic暂不支持嵌入功能", 0, nil)
}

// 验证请求参数
func validateRequest(request *api.Request) error {
	if request == nil {
		return api.NewError(api.ErrorTypeInvalidRequest, "请求不能为空", 0, nil)
	}
	if request.Model == "" {
		return api.NewError(api.ErrorTypeInvalidRequest, "模型不能为空", 0, nil)
	}
	if len(request.Messages) == 0 {
		return api.NewError(api.ErrorTypeInvalidRequest, "消息不能为空", 0, nil)
	}

	// 验证是否为有效的Anthropic模型
	validModels := map[string]bool{
		"claude-3-opus":   true,
		"claude-3-sonnet": true,
		"claude-3-haiku":  true,
		"claude-instant":  true,
		"claude-2":        true,
	}

	if !validModels[request.Model] {
		return api.NewError(api.ErrorTypeInvalidRequest, fmt.Sprintf("无效的Anthropic模型: %s", request.Model), 0, nil)
	}

	return nil
}

// AnthropicResponse 定义Anthropic API的响应结构
type AnthropicResponse struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Model        string         `json:"model"`
	Role         string         `json:"role"`
	Content      []ContentBlock `json:"content"`
	StopReason   string         `json:"stop_reason"`
	StopSequence string         `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// ContentBlock 定义消息内容块
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// AnthropicError 定义Anthropic API的错误响应
type AnthropicError struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// 将SDK的请求格式转换为Anthropic的格式
func adaptRequest(request *api.Request) map[string]interface{} {
	// 提取系统消息和用户消息
	var systemPrompt string
	var messages []map[string]interface{}

	for _, msg := range request.Messages {
		if msg.Role == api.RoleSystem {
			systemPrompt = msg.Content
		} else {
			// 转换为Anthropic的消息格式
			messages = append(messages, map[string]interface{}{
				"role":    string(msg.Role),
				"content": msg.Content,
			})
		}
	}

	// 构建请求
	req := map[string]interface{}{
		"model":    request.Model,
		"messages": messages,
	}

	// 添加系统提示（如果有）
	if systemPrompt != "" {
		req["system"] = systemPrompt
	}

	// 添加可选参数
	if request.Temperature != nil {
		req["temperature"] = *request.Temperature
	}
	if request.TopP != nil {
		req["top_p"] = *request.TopP
	}
	if request.MaxTokens != nil {
		req["max_tokens"] = *request.MaxTokens
	}
	if len(request.Stop) > 0 {
		req["stop_sequences"] = request.Stop
	}
	if request.Stream {
		req["stream"] = request.Stream
	}

	// 添加其他自定义参数
	for k, v := range request.ExtraParams {
		req[k] = v
	}

	return req
}

// 将Anthropic的响应格式转换为SDK的通用格式
func adaptResponse(anthropicResp *AnthropicResponse) *api.Response {
	// 提取文本内容
	var content string
	for _, block := range anthropicResp.Content {
		if block.Type == "text" {
			content += block.Text
		}
	}

	// 构建Choice
	choices := []api.Choice{
		{
			Index: 0,
			Message: api.Message{
				Role:    api.RoleAssistant,
				Content: content,
			},
			FinishReason: anthropicResp.StopReason,
		},
	}

	return &api.Response{
		ID:      anthropicResp.ID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   anthropicResp.Model,
		Choices: choices,
		Usage: api.Usage{
			PromptTokens:     anthropicResp.Usage.InputTokens,
			CompletionTokens: anthropicResp.Usage.OutputTokens,
			TotalTokens:      anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens,
		},
	}
}

// 将Anthropic的错误映射到SDK的错误类型
func mapAnthropicError(anthropicErr *AnthropicError, statusCode int) *api.Error {
	errType := api.ErrorTypeUnknown

	switch anthropicErr.Error.Type {
	case "invalid_request_error":
		errType = api.ErrorTypeInvalidRequest
	case "authentication_error":
		errType = api.ErrorTypeAuthentication
	case "permission_error":
		errType = api.ErrorTypeAuthentication
	case "rate_limit_error":
		errType = api.ErrorTypeRateLimit
	case "server_error":
		errType = api.ErrorTypeServer
	}

	return &api.Error{
		Type:       errType,
		Message:    anthropicErr.Error.Message,
		StatusCode: statusCode,
	}
}

// anthropicResponseStream 实现流式响应接口
type anthropicResponseStream struct {
	reader    *utils.SSEReader
	rawReader io.ReadCloser
}

// AnthropicStreamResponse 定义Anthropic API的流式响应结构
type AnthropicStreamResponse struct {
	Type         string                 `json:"type"`
	Message      AnthropicStreamMessage `json:"message,omitempty"`
	ContentBlock *AnthropicContentBlock `json:"content_block,omitempty"`
	Delta        *AnthropicContentDelta `json:"delta,omitempty"`
	Index        int                    `json:"index,omitempty"`
}

// AnthropicStreamMessage 定义Anthropic流式消息结构
type AnthropicStreamMessage struct {
	ID           string                  `json:"id"`
	Type         string                  `json:"type"`
	Role         string                  `json:"role"`
	Content      []AnthropicContentBlock `json:"content"`
	Model        string                  `json:"model"`
	StopReason   string                  `json:"stop_reason,omitempty"`
	StopSequence string                  `json:"stop_sequence,omitempty"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage,omitempty"`
}

// AnthropicContentBlock 定义Anthropic内容块结构
type AnthropicContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// AnthropicContentDelta 定义Anthropic内容增量结构
type AnthropicContentDelta struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// Recv 实现ResponseStream接口，读取下一个响应块
func (s *anthropicResponseStream) Recv() (*api.ResponseChunk, error) {
	event, err := s.reader.ReadEvent()
	if err != nil {
		if err == io.EOF {
			return nil, io.EOF
		}
		return nil, api.NewError(api.ErrorTypeServer, "读取SSE事件失败", 0, err)
	}

	// 如果不是data字段，或者数据为空，则跳过
	if event.Data == "" {
		return s.Recv() // 递归调用直到获取到有效数据或EOF
	}

	// 解析JSON数据
	var streamResp AnthropicStreamResponse
	if err := json.Unmarshal([]byte(utils.ParseSSEData(event.Data)), &streamResp); err != nil {
		return nil, api.NewError(api.ErrorTypeServer, "解析流式响应失败", 0, err)
	}

	// 判断事件类型
	switch streamResp.Type {
	// 消息完成事件
	case "message_stop":
		return nil, io.EOF

	// 内容块事件
	case "content_block_delta":
		if streamResp.Delta == nil || streamResp.Delta.Type != "text" {
			return s.Recv() // 非文本内容，继续获取下一个事件
		}
		choices := []api.ChunkChoice{
			{
				Index: streamResp.Index,
				Delta: api.Message{
					Role:    api.RoleAssistant,
					Content: streamResp.Delta.Text,
				},
			},
		}

		return &api.ResponseChunk{
			ID:      "", // Anthropic流式API不在每个块中提供ID
			Object:  "chat.completion.chunk",
			Created: time.Now().Unix(),
			Model:   "", // 同样，模型信息仅在消息完成后提供
			Choices: choices,
		}, nil

	// 内容块开始事件
	case "content_block_start":
		if streamResp.ContentBlock == nil || streamResp.ContentBlock.Type != "text" {
			return s.Recv() // 非文本内容，继续获取下一个事件
		}
		// 通常这个事件不包含实际文本内容，可以跳过
		return s.Recv()

	// 消息开始事件
	case "message_start":
		// 消息开始事件不包含内容，可以跳过
		return s.Recv()

	// 消息完成事件
	case "message_delta":
		// 如果消息包含停止原因，返回EOF
		if streamResp.Message.StopReason != "" {
			return nil, io.EOF
		}
		// 否则继续接收
		return s.Recv()

	// 未识别的事件类型
	default:
		return s.Recv()
	}
}

// Close 关闭流
func (s *anthropicResponseStream) Close() error {
	return s.rawReader.Close()
}
