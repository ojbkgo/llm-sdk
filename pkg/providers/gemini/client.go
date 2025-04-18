package gemini

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

// Client 实现了Google Gemini的API客户端
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	maxRetries int
}

// 默认配置
const (
	defaultBaseURL    = "https://generativelanguage.googleapis.com/v1"
	defaultTimeout    = 30 * time.Second
	defaultMaxRetries = 3
)

// NewClient 创建一个新的Gemini客户端
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

	// 创建URL，包含API密钥
	endpoint := fmt.Sprintf("%s/models/%s:generateContent?key=%s", c.baseURL, request.Model, c.apiKey)

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, api.NewError(api.ErrorTypeConnection, "创建HTTP请求失败", 0, err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

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
		var geminiErr GeminiError
		if err := json.Unmarshal(body, &geminiErr); err != nil {
			return nil, api.NewError(api.ErrorTypeServer, fmt.Sprintf("API错误(状态码: %d)", resp.StatusCode), resp.StatusCode, nil)
		}
		return nil, mapGeminiError(&geminiErr, resp.StatusCode)
	}

	// 解析响应
	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, api.NewError(api.ErrorTypeServer, "解析响应失败", resp.StatusCode, err)
	}

	return adaptResponse(&geminiResp, request.Model), nil
}

// CompleteStream 发送请求并获取流式响应
func (c *Client) CompleteStream(ctx context.Context, request *api.Request) (api.ResponseStream, error) {
	// 验证请求
	if err := validateRequest(request); err != nil {
		return nil, err
	}

	// 设置流式标志
	reqCopy := *request

	// 准备请求体
	reqBody, err := json.Marshal(adaptStreamRequest(&reqCopy))
	if err != nil {
		return nil, api.NewError(api.ErrorTypeInvalidRequest, "无法序列化请求", 0, err)
	}

	// 创建URL，包含API密钥和流参数
	endpoint := fmt.Sprintf("%s/models/%s:streamGenerateContent?key=%s&alt=sse",
		c.baseURL, reqCopy.Model, c.apiKey)

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, api.NewError(api.ErrorTypeConnection, "创建HTTP请求失败", 0, err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
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

		var geminiErr GeminiError
		if err := json.Unmarshal(body, &geminiErr); err != nil {
			return nil, api.NewError(api.ErrorTypeServer, fmt.Sprintf("API错误(状态码: %d)", resp.StatusCode), resp.StatusCode, nil)
		}
		return nil, mapGeminiError(&geminiErr, resp.StatusCode)
	}

	return &geminiResponseStream{
		reader:    utils.NewSSEReader(resp.Body),
		rawReader: resp.Body,
		model:     request.Model,
	}, nil
}

// Embedding 获取文本的嵌入向量
func (c *Client) Embedding(ctx context.Context, input string) ([]float32, error) {
	endpoint := fmt.Sprintf("%s/models/embedding-001:embedContent?key=%s", c.baseURL, c.apiKey)

	reqBody, err := json.Marshal(map[string]interface{}{
		"content": map[string]interface{}{
			"parts": []map[string]interface{}{
				{
					"text": input,
				},
			},
		},
	})
	if err != nil {
		return nil, api.NewError(api.ErrorTypeInvalidRequest, "无法序列化请求", 0, err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, api.NewError(api.ErrorTypeConnection, "创建HTTP请求失败", 0, err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

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
		var geminiErr GeminiError
		if err := json.Unmarshal(body, &geminiErr); err != nil {
			return nil, api.NewError(api.ErrorTypeServer, fmt.Sprintf("API错误(状态码: %d)", resp.StatusCode), resp.StatusCode, nil)
		}
		return nil, mapGeminiError(&geminiErr, resp.StatusCode)
	}

	// 解析嵌入响应
	var embedResp struct {
		Embedding struct {
			Values []float32 `json:"values"`
		} `json:"embedding"`
	}

	if err := json.Unmarshal(body, &embedResp); err != nil {
		return nil, api.NewError(api.ErrorTypeServer, "解析嵌入响应失败", resp.StatusCode, err)
	}

	return embedResp.Embedding.Values, nil
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

	return nil
}

// GeminiError 定义Gemini API的错误响应
type GeminiError struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

// GeminiResponse 定义Gemini API的响应结构
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
			Role string `json:"role"`
		} `json:"content"`
		FinishReason  string `json:"finishReason"`
		Index         int    `json:"index"`
		SafetyRatings []struct {
			Category    string `json:"category"`
			Probability string `json:"probability"`
		} `json:"safetyRatings,omitempty"`
	} `json:"candidates"`
	PromptFeedback struct {
		SafetyRatings []struct {
			Category    string `json:"category"`
			Probability string `json:"probability"`
		} `json:"safetyRatings,omitempty"`
	} `json:"promptFeedback,omitempty"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata,omitempty"`
}

// GeminiStreamResponse 定义Gemini API的流式响应结构
type GeminiStreamResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
			Role string `json:"role"`
		} `json:"content"`
		FinishReason  string `json:"finishReason"`
		Index         int    `json:"index"`
		SafetyRatings []struct {
			Category    string `json:"category"`
			Probability string `json:"probability"`
		} `json:"safetyRatings,omitempty"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata,omitempty"`
}

// 将SDK的请求格式转换为Gemini的格式
func adaptRequest(request *api.Request) map[string]interface{} {
	// 将消息转换为Gemini格式
	contents := []map[string]interface{}{}

	for _, msg := range request.Messages {
		content := map[string]interface{}{
			"role": mapRole(msg.Role),
			"parts": []map[string]interface{}{
				{
					"text": msg.Content,
				},
			},
		}
		contents = append(contents, content)
	}

	// 构建请求
	req := map[string]interface{}{
		"contents": contents,
	}

	// 添加生成参数
	generationConfig := map[string]interface{}{}

	if request.Temperature != nil {
		generationConfig["temperature"] = *request.Temperature
	}
	if request.TopP != nil {
		generationConfig["topP"] = *request.TopP
	}
	if request.MaxTokens != nil {
		generationConfig["maxOutputTokens"] = *request.MaxTokens
	}
	if len(request.Stop) > 0 {
		generationConfig["stopSequences"] = request.Stop
	}

	if len(generationConfig) > 0 {
		req["generationConfig"] = generationConfig
	}

	// 添加安全设置
	safetySettings := []map[string]string{
		{
			"category":  "HARM_CATEGORY_HARASSMENT",
			"threshold": "BLOCK_NONE",
		},
		{
			"category":  "HARM_CATEGORY_HATE_SPEECH",
			"threshold": "BLOCK_NONE",
		},
		{
			"category":  "HARM_CATEGORY_SEXUALLY_EXPLICIT",
			"threshold": "BLOCK_NONE",
		},
		{
			"category":  "HARM_CATEGORY_DANGEROUS_CONTENT",
			"threshold": "BLOCK_NONE",
		},
	}

	req["safetySettings"] = safetySettings

	return req
}

// 为流式请求适配请求格式
func adaptStreamRequest(request *api.Request) map[string]interface{} {
	req := adaptRequest(request)
	// 添加流式标志
	req["streamGenerationConfig"] = map[string]interface{}{
		"streamContentTokens": true,
	}
	return req
}

// 将Gemini的响应格式转换为SDK的通用格式
func adaptResponse(geminiResp *GeminiResponse, modelName string) *api.Response {
	// 提取文本内容
	choices := []api.Choice{}

	for i, candidate := range geminiResp.Candidates {
		var content string
		for _, part := range candidate.Content.Parts {
			content += part.Text
		}

		choices = append(choices, api.Choice{
			Index: i,
			Message: api.Message{
				Role:    api.RoleAssistant,
				Content: content,
			},
			FinishReason: candidate.FinishReason,
		})
	}

	return &api.Response{
		ID:      "", // Gemini不提供ID
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   modelName,
		Choices: choices,
		Usage: api.Usage{
			PromptTokens:     geminiResp.UsageMetadata.PromptTokenCount,
			CompletionTokens: geminiResp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      geminiResp.UsageMetadata.TotalTokenCount,
		},
	}
}

// 将Gemini的错误映射到SDK的错误类型
func mapGeminiError(geminiErr *GeminiError, statusCode int) *api.Error {
	errType := api.ErrorTypeUnknown

	switch geminiErr.Error.Code {
	case 400:
		errType = api.ErrorTypeInvalidRequest
	case 401, 403:
		errType = api.ErrorTypeAuthentication
	case 429:
		errType = api.ErrorTypeRateLimit
	case 500, 502, 503:
		errType = api.ErrorTypeServer
	}

	return &api.Error{
		Type:       errType,
		Message:    geminiErr.Error.Message,
		StatusCode: statusCode,
		Code:       geminiErr.Error.Status,
	}
}

// 将SDK的角色映射到Gemini的角色
func mapRole(role api.Role) string {
	switch role {
	case api.RoleUser:
		return "user"
	case api.RoleSystem, api.RoleAssistant:
		return "model"
	default:
		return "user"
	}
}

// geminiResponseStream 实现流式响应接口
type geminiResponseStream struct {
	reader    *utils.SSEReader
	rawReader io.ReadCloser
	model     string
	chunkID   int
}

// Recv 实现ResponseStream接口，读取下一个响应块
func (s *geminiResponseStream) Recv() (*api.ResponseChunk, error) {
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
	var streamResp GeminiStreamResponse
	if err := json.Unmarshal([]byte(utils.ParseSSEData(event.Data)), &streamResp); err != nil {
		return nil, api.NewError(api.ErrorTypeServer, "解析流式响应失败", 0, err)
	}

	// 如果没有候选项，继续接收
	if len(streamResp.Candidates) == 0 {
		return s.Recv()
	}

	// 转换为SDK的通用格式
	choices := []api.ChunkChoice{}

	for _, candidate := range streamResp.Candidates {
		// 检查是否有结束原因
		if candidate.FinishReason != "" {
			// 返回一个带有结束原因的空内容块
			choices = append(choices, api.ChunkChoice{
				Index:        candidate.Index,
				Delta:        api.Message{Role: api.RoleAssistant},
				FinishReason: candidate.FinishReason,
			})
			continue
		}

		// 提取文本内容
		var content string
		for _, part := range candidate.Content.Parts {
			content += part.Text
		}

		if content != "" {
			choices = append(choices, api.ChunkChoice{
				Index: candidate.Index,
				Delta: api.Message{
					Role:    api.RoleAssistant,
					Content: content,
				},
			})
		}
	}

	// 如果没有有效内容，继续接收
	if len(choices) == 0 {
		return s.Recv()
	}

	s.chunkID++

	return &api.ResponseChunk{
		ID:      fmt.Sprintf("chunk-%d", s.chunkID),
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   s.model,
		Choices: choices,
	}, nil
}

// Close 关闭流
func (s *geminiResponseStream) Close() error {
	return s.rawReader.Close()
}
