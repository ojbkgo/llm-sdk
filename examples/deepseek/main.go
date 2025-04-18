package main

import (
	"context"
	"fmt"
	"os"

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

	// 分别展示正常聊天和代码生成两个示例
	fmt.Println("=== 示例 1: 基础聊天 ===")
	chatExample(client)

	fmt.Println("\n=== 示例 2: 代码生成 ===")
	codeExample(client)

	fmt.Println("\n=== 示例 3: 嵌入向量 ===")
	embeddingExample(client)
}

// 基础聊天示例
func chatExample(client api.LLMClient) {
	// 准备请求
	temperature := 0.7
	request := &api.Request{
		Model: models.DeepSeekChat,
		Messages: []api.Message{
			{
				Role:    api.RoleSystem,
				Content: "你是一个专业、友好且具有创造力的AI助手。请用中文回答用户问题。",
			},
			{
				Role:    api.RoleUser,
				Content: "你好！请介绍一下自己和你的能力。",
			},
		},
		Temperature: &temperature,
	}

	// 发送请求
	ctx := context.Background()
	response, err := client.Complete(ctx, request)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	// 打印响应
	if len(response.Choices) > 0 {
		fmt.Printf("DeepSeek回复: %s\n", response.Choices[0].Message.Content)
	} else {
		fmt.Println("未收到回复")
	}

	fmt.Printf("使用的令牌数: %d (提示: %d, 完成: %d)\n",
		response.Usage.TotalTokens,
		response.Usage.PromptTokens,
		response.Usage.CompletionTokens,
	)
}

// 代码生成示例
func codeExample(client api.LLMClient) {
	// 准备请求
	temperature := 0.1 // 代码生成通常使用较低的温度
	request := &api.Request{
		Model: models.DeepSeekCoder,
		Messages: []api.Message{
			{
				Role:    api.RoleSystem,
				Content: "你是一个专业的编程助手，擅长生成高质量、可运行的代码。请直接给出代码，不需要额外解释。",
			},
			{
				Role:    api.RoleUser,
				Content: "请用Go语言编写一个简单的Web服务器，提供一个REST API接口，能够接收POST请求并返回JSON响应。",
			},
		},
		Temperature: &temperature,
	}

	// 发送请求
	ctx := context.Background()
	response, err := client.Complete(ctx, request)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	// 打印响应
	if len(response.Choices) > 0 {
		fmt.Printf("DeepSeek代码: %s\n", response.Choices[0].Message.Content)
	} else {
		fmt.Println("未收到回复")
	}

	fmt.Printf("使用的令牌数: %d (提示: %d, 完成: %d)\n",
		response.Usage.TotalTokens,
		response.Usage.PromptTokens,
		response.Usage.CompletionTokens,
	)
}

// 嵌入向量示例
func embeddingExample(client api.LLMClient) {
	ctx := context.Background()
	text := "这是一段用于测试嵌入向量生成的文本。"

	embedding, err := client.Embedding(ctx, text)
	if err != nil {
		fmt.Printf("获取嵌入向量失败: %v\n", err)
		return
	}

	fmt.Printf("生成的嵌入向量维度: %d\n", len(embedding))
	if len(embedding) > 5 {
		fmt.Printf("前5个维度的值: %v\n", embedding[:5])
	}
}
