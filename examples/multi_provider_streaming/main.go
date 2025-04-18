package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ojbkgo/llm-sdk/pkg/api"
	"github.com/ojbkgo/llm-sdk/pkg/models"
	"github.com/ojbkgo/llm-sdk/pkg/providers/anthropic"
	"github.com/ojbkgo/llm-sdk/pkg/providers/deepseek"
	"github.com/ojbkgo/llm-sdk/pkg/providers/gemini"
	"github.com/ojbkgo/llm-sdk/pkg/providers/openai"
)

func main() {
	promptText := "用简短的话介绍一下自己的特点和优势"

	fmt.Println("===多提供商流式输出对比===")
	fmt.Printf("提示: %s\n\n", promptText)

	// 尝试所有提供商的流式输出
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey != "" {
		fmt.Println("=== OpenAI (GPT-3.5) ===")
		streamFromProvider("openai", openaiKey, models.GPT35Turbo, promptText)
	} else {
		fmt.Println("=== OpenAI (未设置API密钥) ===")
	}

	if anthropicKey := os.Getenv("ANTHROPIC_API_KEY"); anthropicKey != "" {
		fmt.Println("\n=== Anthropic (Claude 3 Haiku) ===")
		streamFromProvider("anthropic", anthropicKey, models.Claude3Haiku, promptText)
	} else {
		fmt.Println("\n=== Anthropic (未设置API密钥) ===")
	}

	if deepseekKey := os.Getenv("DEEPSEEK_API_KEY"); deepseekKey != "" {
		fmt.Println("\n=== DeepSeek (Chat) ===")
		streamFromProvider("deepseek", deepseekKey, models.DeepSeekChat, promptText)
	} else {
		fmt.Println("\n=== DeepSeek (未设置API密钥) ===")
	}

	if geminiKey := os.Getenv("GEMINI_API_KEY"); geminiKey != "" {
		fmt.Println("\n=== Google (Gemini Pro) ===")
		streamFromProvider("gemini", geminiKey, models.GeminiPro, promptText)
	} else {
		fmt.Println("\n=== Google (未设置API密钥) ===")
	}
}

// streamFromProvider 使用指定提供商进行流式输出
func streamFromProvider(provider, apiKey, model, prompt string) {
	var client api.LLMClient
	var err error

	// 根据提供商创建客户端
	switch provider {
	case "openai":
		client, err = openai.NewClient(func(options *api.ClientOptions) {
			options.APIKey = apiKey
		})
	case "anthropic":
		client, err = anthropic.NewClient(func(options *api.ClientOptions) {
			options.APIKey = apiKey
		})
	case "deepseek":
		client, err = deepseek.NewClient(func(options *api.ClientOptions) {
			options.APIKey = apiKey
		})
	case "gemini":
		client, err = gemini.NewClient(func(options *api.ClientOptions) {
			options.APIKey = apiKey
		})
	default:
		fmt.Printf("不支持的提供商: %s\n", provider)
		return
	}

	if err != nil {
		fmt.Printf("创建客户端失败: %v\n", err)
		return
	}

	// 准备请求
	temperature := 0.7
	request := &api.Request{
		Model: model,
		Messages: []api.Message{
			{
				Role:    api.RoleSystem,
				Content: "你是一个助手，回答应该简洁、准确。请说中文。",
			},
			{
				Role:    api.RoleUser,
				Content: prompt,
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

	// 使用StreamProcessor处理流式输出
	var builder strings.Builder

	err = api.NewStreamProcessor().Process(stream, &api.StreamOptions{
		OnText: func(text string) error {
			fmt.Print(text)           // 实时输出
			builder.WriteString(text) // 收集完整内容
			return nil
		},
		OnComplete: func(err error) {
			if err != nil {
				fmt.Printf("\n处理过程中出错: %v\n", err)
			}
		},
		AutoClose: true,
	})

	if err != nil {
		fmt.Printf("\n流处理失败: %v\n", err)
	} else {
		fmt.Printf("\n\n总共生成字符数: %d\n", builder.Len())
	}
}
