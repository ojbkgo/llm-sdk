package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ojbkgo/llm-sdk/pkg/api"
	"github.com/ojbkgo/llm-sdk/pkg/models"
	"github.com/ojbkgo/llm-sdk/pkg/providers/openai"
)

func main() {
	// 从环境变量获取API密钥
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("请设置OPENAI_API_KEY环境变量")
		os.Exit(1)
	}

	// 创建OpenAI客户端
	client, err := openai.NewClient(func(options *api.ClientOptions) {
		options.APIKey = apiKey
	})
	if err != nil {
		fmt.Printf("创建客户端失败: %v\n", err)
		os.Exit(1)
	}

	// 准备请求
	temperature := 0.7
	request := &api.Request{
		Model: models.GPT35Turbo,
		Messages: []api.Message{
			{
				Role:    api.RoleSystem,
				Content: "你是一个有帮助的AI助手。",
			},
			{
				Role:    api.RoleUser,
				Content: "你好，请介绍一下自己。",
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
		fmt.Printf("AI回复: %s\n", response.Choices[0].Message.Content)
	} else {
		fmt.Println("未收到回复")
	}

	fmt.Printf("使用的令牌数: %d (提示: %d, 完成: %d)\n",
		response.Usage.TotalTokens,
		response.Usage.PromptTokens,
		response.Usage.CompletionTokens,
	)
}
