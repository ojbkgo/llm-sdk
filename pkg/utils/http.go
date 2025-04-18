package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/ojbkgo/llm-sdk/pkg/api"
)

// HTTPConfig 定义HTTP客户端配置
type HTTPConfig struct {
	Timeout    time.Duration
	MaxRetries int
	RetryDelay time.Duration
}

// DefaultHTTPConfig 返回默认的HTTP配置
func DefaultHTTPConfig() HTTPConfig {
	return HTTPConfig{
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
	}
}

// DoHTTPRequest 发送HTTP请求并处理错误和重试
func DoHTTPRequest(
	ctx context.Context,
	client *http.Client,
	method string,
	url string,
	body interface{},
	headers map[string]string,
	config HTTPConfig,
) ([]byte, int, error) {
	var bodyBytes []byte
	var err error

	// 如果有请求体，将其序列化为JSON
	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, 0, api.NewError(api.ErrorTypeInvalidRequest, "无法序列化请求体", 0, err)
		}
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, 0, api.NewError(api.ErrorTypeConnection, "创建HTTP请求失败", 0, err)
	}

	// 设置默认的Content-Type
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// 设置自定义请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 执行请求并处理重试
	var resp *http.Response
	var retryCount int
	var lastErr error

	for retryCount = 0; retryCount <= config.MaxRetries; retryCount++ {
		if retryCount > 0 {
			select {
			case <-ctx.Done():
				return nil, 0, api.NewError(api.ErrorTypeTimeout, "请求超时", 0, ctx.Err())
			case <-time.After(config.RetryDelay * time.Duration(retryCount)):
				// 指数退避重试
			}
		}

		resp, err = client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		// 读取响应体
		defer resp.Body.Close()
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}

		// 如果是服务器错误（5xx），则重试
		if resp.StatusCode >= 500 && retryCount < config.MaxRetries {
			lastErr = api.NewError(api.ErrorTypeServer, "服务器错误", resp.StatusCode, nil)
			continue
		}

		return respBody, resp.StatusCode, nil
	}

	// 如果所有重试都失败
	if lastErr != nil {
		if apiErr, ok := lastErr.(*api.Error); ok {
			return nil, apiErr.StatusCode, apiErr
		}
		return nil, 0, api.NewError(api.ErrorTypeConnection, "HTTP请求失败", 0, lastErr)
	}

	return nil, 0, api.NewError(api.ErrorTypeUnknown, "未知错误", 0, nil)
}

// MakeAuthHeader 创建认证头
func MakeAuthHeader(apiKey string, authType string) map[string]string {
	headers := make(map[string]string)

	switch authType {
	case "bearer":
		headers["Authorization"] = "Bearer " + apiKey
	case "x-api-key":
		headers["x-api-key"] = apiKey
	default:
		headers["Authorization"] = "Bearer " + apiKey
	}

	return headers
}
