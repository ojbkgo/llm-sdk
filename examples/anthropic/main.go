package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ojbkgo/llm-sdk/pkg/api"
	"github.com/ojbkgo/llm-sdk/pkg/models"
	"github.com/ojbkgo/llm-sdk/pkg/providers/anthropic"
)

func main() {
	// 从环境变量获取API密钥
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("请设置ANTHROPIC_API_KEY环境变量")
		os.Exit(1)
	}

	// 创建Anthropic客户端
	client, err := anthropic.NewClient(func(options *api.ClientOptions) {
		options.APIKey = apiKey
	})
	if err != nil {
		fmt.Printf("创建客户端失败: %v\n", err)
		os.Exit(1)
	}

	// 准备请求
	temperature := 0.7
	request := &api.Request{
		Model: models.Claude3Sonnet,
		Messages: []api.Message{
			{
				Role:    api.RoleSystem,
				Content: "你是一个专业、友好且具有创造力的AI助手。请用中文回答问题。",
			},
			{
				Role:    api.RoleUser,
				Content: "你好！请解释一下你是谁，以及你能帮我做什么。",
			},
		},
		Temperature: &temperature,
	}

	// 发送请求
	ctx := context.Background()
	response, err := client.Complete(ctx, request)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		os.Exit(1)
	}

	// 打印响应
	if len(response.Choices) > 0 {
		fmt.Printf("Claude回复: %s\n", response.Choices[0].Message.Content)
	} else {
		fmt.Println("未收到回复")
	}

	fmt.Printf("使用的令牌数: %d (提示: %d, 完成: %d)\n",
		response.Usage.TotalTokens,
		response.Usage.PromptTokens,
		response.Usage.CompletionTokens,
	)
}
