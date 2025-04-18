package api

import (
	"io"
)

// StreamHandler 是一个流式响应处理器函数类型
type StreamHandler func(chunk *ResponseChunk) error

// StreamOptions 定义了流式响应的配置选项
type StreamOptions struct {
	// OnChunk 当接收到新的响应块时被调用
	OnChunk StreamHandler

	// OnComplete 当流处理完成时被调用
	OnComplete func(err error)

	// OnText 当接收到纯文本内容时被调用（便于直接处理文本内容）
	OnText func(text string) error

	// AutoClose 是否在接收完所有事件后自动关闭流，默认为true
	AutoClose bool
}

// DefaultStreamOptions 返回默认的流式选项
func DefaultStreamOptions() *StreamOptions {
	return &StreamOptions{
		AutoClose: true,
	}
}

// StreamProcessor 是流式响应处理器的接口
type StreamProcessor interface {
	// Process 处理流式响应
	Process(stream ResponseStream, options *StreamOptions) error

	// ProcessWithHandler 使用简单的文本处理函数处理流式响应
	ProcessWithHandler(stream ResponseStream, handler func(text string) error) error
}

// DefaultStreamProcessor 是一个默认的流式处理器实现
type DefaultStreamProcessor struct{}

// NewStreamProcessor 创建一个新的流式处理器
func NewStreamProcessor() StreamProcessor {
	return &DefaultStreamProcessor{}
}

// Process 实现了流式响应处理
func (p *DefaultStreamProcessor) Process(stream ResponseStream, options *StreamOptions) error {
	if options == nil {
		options = DefaultStreamOptions()
	}

	defer func() {
		if options.AutoClose {
			stream.Close()
		}
	}()

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			if options.OnComplete != nil {
				options.OnComplete(nil)
			}
			return nil
		}
		if err != nil {
			if options.OnComplete != nil {
				options.OnComplete(err)
			}
			return err
		}

		// 调用块处理回调
		if options.OnChunk != nil {
			if err := options.OnChunk(chunk); err != nil {
				if options.OnComplete != nil {
					options.OnComplete(err)
				}
				return err
			}
		}

		// 如果有纯文本处理回调，提取并传递文本内容
		if options.OnText != nil && len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Delta.Content
			if content != "" {
				if err := options.OnText(content); err != nil {
					if options.OnComplete != nil {
						options.OnComplete(err)
					}
					return err
				}
			}
		}
	}
}

// ProcessWithHandler 实现了简化的流式处理，只关注文本内容
func (p *DefaultStreamProcessor) ProcessWithHandler(stream ResponseStream, handler func(text string) error) error {
	return p.Process(stream, &StreamOptions{
		OnText:    handler,
		AutoClose: true,
	})
}

// 以下是流式响应相关的便捷函数

// CollectFullContent 从流式响应中收集完整内容
func CollectFullContent(stream ResponseStream) (string, error) {
	var fullContent string
	err := NewStreamProcessor().ProcessWithHandler(stream, func(text string) error {
		fullContent += text
		return nil
	})
	return fullContent, err
}

// StreamToWriter 将流式响应输出到一个io.Writer
func StreamToWriter(stream ResponseStream, writer io.Writer) error {
	return NewStreamProcessor().ProcessWithHandler(stream, func(text string) error {
		_, err := writer.Write([]byte(text))
		return err
	})
}
