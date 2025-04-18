package deepseek

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ojbkgo/llm-sdk/pkg/api"
)

// Client 实现了DeepSeek的API客户端
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	maxRetries int
}

// 默认配置
const (
	defaultBaseURL    = "https://api.deepseek.com/v1"
	defaultTimeout    = 30 * time.Second
	defaultMaxRetries = 3
)

// NewClient 创建一个新的DeepSeek客户端
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
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, api.NewError(api.ErrorTypeConnection, "创建HTTP请求失败", 0, err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

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
		var deepseekErr DeepSeekError
		if err := json.Unmarshal(body, &deepseekErr); err != nil {
			return nil, api.NewError(api.ErrorTypeServer, fmt.Sprintf("API错误(状态码: %d)", resp.StatusCode), resp.StatusCode, nil)
		}
		return nil, mapDeepSeekError(&deepseekErr, resp.StatusCode)
	}

	// 解析响应
	var deepseekResp DeepSeekResponse
	if err := json.Unmarshal(body, &deepseekResp); err != nil {
		return nil, api.NewError(api.ErrorTypeServer, "解析响应失败", resp.StatusCode, err)
	}

	return adaptResponse(&deepseekResp), nil
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
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, api.NewError(api.ErrorTypeConnection, "创建HTTP请求失败", 0, err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
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

		var deepseekErr DeepSeekError
		if err := json.Unmarshal(body, &deepseekErr); err != nil {
			return nil, api.NewError(api.ErrorTypeServer, fmt.Sprintf("API错误(状态码: %d)", resp.StatusCode), resp.StatusCode, nil)
		}
		return nil, mapDeepSeekError(&deepseekErr, resp.StatusCode)
	}

	return &deepseekResponseStream{
		reader: resp.Body,
	}, nil
}

// Embedding 获取文本的嵌入向量
func (c *Client) Embedding(ctx context.Context, input string) ([]float32, error) {
	reqBody, err := json.Marshal(map[string]interface{}{
		"model": "deepseek-embedding", // 默认嵌入模型
		"input": input,
	})
	if err != nil {
		return nil, api.NewError(api.ErrorTypeInvalidRequest, "无法序列化请求", 0, err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/embeddings", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, api.NewError(api.ErrorTypeConnection, "创建HTTP请求失败", 0, err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

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
		var deepseekErr DeepSeekError
		if err := json.Unmarshal(body, &deepseekErr); err != nil {
			return nil, api.NewError(api.ErrorTypeServer, fmt.Sprintf("API错误(状态码: %d)", resp.StatusCode), resp.StatusCode, nil)
		}
		return nil, mapDeepSeekError(&deepseekErr, resp.StatusCode)
	}

	// 解析嵌入响应
	var embedResp struct {
		Object string `json:"object"`
		Data   []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &embedResp); err != nil {
		return nil, api.NewError(api.ErrorTypeServer, "解析嵌入响应失败", resp.StatusCode, err)
	}

	if len(embedResp.Data) == 0 || len(embedResp.Data[0].Embedding) == 0 {
		return nil, api.NewError(api.ErrorTypeServer, "未收到有效的嵌入结果", resp.StatusCode, nil)
	}

	return embedResp.Data[0].Embedding, nil
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

	// DeepSeek验证模型先注释掉，因为模型可能会更新
	// 实际使用中最好添加模型验证

	return nil
}

// DeepSeekResponse 定义DeepSeek API的响应结构
type DeepSeekResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// DeepSeekError 定义DeepSeek API的错误响应
type DeepSeekError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Param   string `json:"param"`
		Code    string `json:"code"`
	} `json:"error"`
}

// 将SDK的请求格式转换为DeepSeek的格式
func adaptRequest(request *api.Request) map[string]interface{} {
	// DeepSeek的API格式与OpenAI类似，这里可以直接适配
	req := map[string]interface{}{
		"model":    request.Model,
		"messages": request.Messages,
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
	if request.PresencePenalty != nil {
		req["presence_penalty"] = *request.PresencePenalty
	}
	if request.FrequencyPenalty != nil {
		req["frequency_penalty"] = *request.FrequencyPenalty
	}
	if len(request.Stop) > 0 {
		req["stop"] = request.Stop
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

// 将DeepSeek的响应格式转换为SDK的通用格式
func adaptResponse(deepseekResp *DeepSeekResponse) *api.Response {
	choices := make([]api.Choice, len(deepseekResp.Choices))
	for i, choice := range deepseekResp.Choices {
		choices[i] = api.Choice{
			Index: choice.Index,
			Message: api.Message{
				Role:    api.Role(choice.Message.Role),
				Content: choice.Message.Content,
			},
			FinishReason: choice.FinishReason,
		}
	}

	return &api.Response{
		ID:      deepseekResp.ID,
		Object:  deepseekResp.Object,
		Created: deepseekResp.Created,
		Model:   deepseekResp.Model,
		Choices: choices,
		Usage: api.Usage{
			PromptTokens:     deepseekResp.Usage.PromptTokens,
			CompletionTokens: deepseekResp.Usage.CompletionTokens,
			TotalTokens:      deepseekResp.Usage.TotalTokens,
		},
	}
}

// 将DeepSeek的错误映射到SDK的错误类型
func mapDeepSeekError(deepseekErr *DeepSeekError, statusCode int) *api.Error {
	errType := api.ErrorTypeUnknown
	switch deepseekErr.Error.Type {
	case "invalid_request_error":
		errType = api.ErrorTypeInvalidRequest
	case "authentication_error":
		errType = api.ErrorTypeAuthentication
	case "rate_limit_error":
		errType = api.ErrorTypeRateLimit
	case "server_error":
		errType = api.ErrorTypeServer
	}

	return &api.Error{
		Type:       errType,
		Message:    deepseekErr.Error.Message,
		StatusCode: statusCode,
		Param:      deepseekErr.Error.Param,
		Code:       deepseekErr.Error.Code,
	}
}

// deepseekResponseStream 实现流式响应接口
type deepseekResponseStream struct {
	reader io.ReadCloser
	buffer []byte
}

// Recv 实现ResponseStream接口，读取下一个响应块
func (s *deepseekResponseStream) Recv() (*api.ResponseChunk, error) {
	// 这里是简化实现，实际应该解析SSE事件流
	// 真实实现需要处理SSE格式的数据流，如 data: {...} 格式
	return nil, api.NewError(api.ErrorTypeUnknown, "流式响应功能尚未完全实现", 0, nil)
}

// Close 关闭流
func (s *deepseekResponseStream) Close() error {
	return s.reader.Close()
}
