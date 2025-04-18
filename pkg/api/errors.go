package api

import (
	"fmt"
)

// ErrorType 定义错误类型
type ErrorType string

const (
	// ErrorTypeAuthentication 认证错误
	ErrorTypeAuthentication ErrorType = "authentication_error"
	// ErrorTypeInvalidRequest 无效请求错误
	ErrorTypeInvalidRequest ErrorType = "invalid_request_error"
	// ErrorTypeRateLimit 速率限制错误
	ErrorTypeRateLimit ErrorType = "rate_limit_error"
	// ErrorTypeServer 服务器错误
	ErrorTypeServer ErrorType = "server_error"
	// ErrorTypeTimeout 超时错误
	ErrorTypeTimeout ErrorType = "timeout_error"
	// ErrorTypeConnection 连接错误
	ErrorTypeConnection ErrorType = "connection_error"
	// ErrorTypeUnknown 未知错误
	ErrorTypeUnknown ErrorType = "unknown_error"
)

// Error 定义SDK错误
type Error struct {
	Type       ErrorType `json:"type"`
	Message    string    `json:"message"`
	StatusCode int       `json:"status_code,omitempty"`
	Param      string    `json:"param,omitempty"`
	Code       string    `json:"code,omitempty"`
	RawError   error     `json:"-"`
}

// Error 实现error接口
func (e *Error) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("llm-sdk: %s error: %s (status code: %d)", e.Type, e.Message, e.StatusCode)
	}
	return fmt.Sprintf("llm-sdk: %s error: %s", e.Type, e.Message)
}

// Unwrap 获取原始错误
func (e *Error) Unwrap() error {
	return e.RawError
}

// NewError 创建一个新的SDK错误
func NewError(errType ErrorType, message string, statusCode int, rawErr error) *Error {
	return &Error{
		Type:       errType,
		Message:    message,
		StatusCode: statusCode,
		RawError:   rawErr,
	}
}
