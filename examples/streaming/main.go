package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ojbkgo/llm-sdk/pkg/api"
	"github.com/ojbkgo/llm-sdk/pkg/models"
	"github.com/ojbkgo/llm-sdk/pkg/providers/deepseek"
)

func main() {
	// 从环境变量获取API密钥
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		fmt.Println("请设置DEEPSEEK_API_KEY环境变量")
		os.Exit(1)
	}

	// 创建DeepSeek客户端
	client, err := deepseek.NewClient(func(options *api.ClientOptions) {
		options.APIKey = apiKey
	})
	if err != nil {
		fmt.Printf("创建客户端失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== 演示 1: 基本流式输出 ===")
	basicStreamingExample(client)

	fmt.Println("\n=== 演示 2: 带进度显示的流式输出 ===")
	progressStreamingExample(client)

	fmt.Println("\n=== 演示 3: 完整内容收集 ===")
	collectFullContentExample(client)

	fmt.Println("\n=== 演示 4: 使用StreamProcessor接口 ===")
	streamProcessorExample(client)
}

// 基本流式输出示例
func basicStreamingExample(client api.LLMClient) {
	// 准备请求
	temperature := 0.7
	request := &api.Request{
		Model: models.DeepSeekChat,
		Messages: []api.Message{
			{
				Role:    api.RoleSystem,
				Content: "你是一个助手，请尽量简短地回复。",
			},
			{
				Role:    api.RoleUser,
				Content: "简单介绍一下Go语言的特点。",
			},
		},
		Temperature: &temperature,
	}

	// 发送流式请求
	ctx := context.Background()
	stream, err := client.CompleteStream(ctx, request)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer stream.Close()

	// 逐块接收响应
	fmt.Println("实时输出:")
	for {
		chunk, err := stream.Recv()
		if err != nil {
			// 检查是否是正常的EOF
			if err.Error() == "EOF" {
				break
			}
			fmt.Printf("接收响应块出错: %v\n", err)
			break
		}

		// 从响应块中提取文本内容
		if len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Delta.Content
			if content != "" {
				fmt.Print(content)
			}
		}
	}
	fmt.Println("\n")
}

// 带进度显示的流式输出示例
func progressStreamingExample(client api.LLMClient) {
	// 准备请求
	temperature := 0.7
	request := &api.Request{
		Model: models.DeepSeekChat,
		Messages: []api.Message{
			{
				Role:    api.RoleSystem,
				Content: "你是一个助手，请用20-30个字回复。",
			},
			{
				Role:    api.RoleUser,
				Content: "介绍一下什么是人工智能。",
			},
		},
		Temperature: &temperature,
	}

	// 发送流式请求
	ctx := context.Background()
	stream, err := client.CompleteStream(ctx, request)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer stream.Close()

	// 用于动画显示的字符
	spinChars := []string{"|", "/", "-", "\\"}
	spinIndex := 0

	// 使用一个字符串构建器来收集完整的响应
	var fullResponse strings.Builder

	// 逐块接收响应
	fmt.Print("生成中: ")
	for {
		chunk, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			fmt.Printf("接收响应块出错: %v\n", err)
			break
		}

		// 从响应块中提取文本内容
		if len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Delta.Content
			if content != "" {
				fullResponse.WriteString(content)

				// 更新动画
				fmt.Printf("\r生成中: %s %d字 ", spinChars[spinIndex], fullResponse.Len())
				spinIndex = (spinIndex + 1) % len(spinChars)

				// 添加一个短暂的延迟，使动画效果更明显
				time.Sleep(100 * time.Millisecond)
			}
		}
	}

	// 打印完整响应
	fmt.Printf("\r完成! 总共生成 %d 字的内容:\n%s\n", fullResponse.Len(), fullResponse.String())
}

// 完整内容收集示例
func collectFullContentExample(client api.LLMClient) {
	// 准备请求
	temperature := 0.7
	request := &api.Request{
		Model: models.DeepSeekChat,
		Messages: []api.Message{
			{
				Role:    api.RoleSystem,
				Content: "你是一个助手，回答应该简洁明了。",
			},
			{
				Role:    api.RoleUser,
				Content: "用一句话描述云计算的优势。",
			},
		},
		Temperature: &temperature,
	}

	// 发送流式请求
	ctx := context.Background()
	stream, err := client.CompleteStream(ctx, request)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	// 使用SDK提供的工具函数收集完整内容
	content, err := api.CollectFullContent(stream)
	if err != nil {
		fmt.Printf("收集内容失败: %v\n", err)
		return
	}

	fmt.Printf("收集到的完整内容:\n%s\n", content)
}

// 使用StreamProcessor接口示例
func streamProcessorExample(client api.LLMClient) {
	// 准备请求
	temperature := 0.7
	request := &api.Request{
		Model: models.DeepSeekChat,
		Messages: []api.Message{
			{
				Role:    api.RoleSystem,
				Content: "你是一个助手，请用分点形式回答。",
			},
			{
				Role:    api.RoleUser,
				Content: "给出三点建议，如何提高编程效率。",
			},
		},
		Temperature: &temperature,
	}

	// 发送流式请求
	ctx := context.Background()
	stream, err := client.CompleteStream(ctx, request)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	// 创建一个字符串构建器来收集完整内容
	var builder strings.Builder

	// 创建流处理器并设置回调选项
	processor := api.NewStreamProcessor()
	err = processor.Process(stream, &api.StreamOptions{
		// 当收到文本块时的回调
		OnText: func(text string) error {
			fmt.Print(text)           // 实时打印
			builder.WriteString(text) // 同时收集
			return nil
		},
		// 当流处理完成时的回调
		OnComplete: func(err error) {
			if err != nil {
				fmt.Printf("\n处理过程中出错: %v\n", err)
			} else {
				fmt.Printf("\n\n处理完成! 总共收集了 %d 字的内容\n", builder.Len())
			}
		},
		// 自动关闭流
		AutoClose: true,
	})

	if err != nil {
		fmt.Printf("流处理失败: %v\n", err)
	}
}
