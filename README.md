# LLM SDK

一个用 Go 语言编写的轻量级 LLM (大型语言模型) SDK，提供统一的接口访问多种 LLM 服务提供商。

## 特性

- 统一的 API 接口设计，支持多种 LLM 提供商
- 完整的错误处理和重试机制
- 支持同步和流式响应
- 强大的 SSE (Server-Sent Events) 流式处理能力
- 灵活的配置选项
- 支持嵌入向量生成

## 支持的提供商

- OpenAI (GPT-3.5, GPT-4)
- Anthropic (Claude 3 系列)
- DeepSeek (DeepSeek Chat, DeepSeek Coder, DeepSeek Llama)
- Google (Gemini Pro, Gemini Ultra)

## 安装

```bash
go get github.com/ojbkgo/llm-sdk
```

## 快速开始

### OpenAI 聊天示例

```go
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
	}
}
```

### Anthropic 聊天示例

```go
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
				Content: "你是一个专业、友好的AI助手。",
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
		fmt.Printf("Claude回复: %s\n", response.Choices[0].Message.Content)
	}
}
```

### DeepSeek 聊天与代码示例

```go
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

	// 准备聊天请求
	temperature := 0.7
	request := &api.Request{
		Model: models.DeepSeekChat,
		Messages: []api.Message{
			{
				Role:    api.RoleSystem,
				Content: "你是一个专业、友好的AI助手。",
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
		fmt.Printf("DeepSeek回复: %s\n", response.Choices[0].Message.Content)
	}
}
```

### Google Gemini 示例

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ojbkgo/llm-sdk/pkg/api"
	"github.com/ojbkgo/llm-sdk/pkg/models"
	"github.com/ojbkgo/llm-sdk/pkg/providers/gemini"
)

func main() {
	// 从环境变量获取API密钥
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		fmt.Println("请设置GEMINI_API_KEY环境变量")
		os.Exit(1)
	}

	// 创建Gemini客户端
	client, err := gemini.NewClient(func(options *api.ClientOptions) {
		options.APIKey = apiKey
	})
	if err != nil {
		fmt.Printf("创建客户端失败: %v\n", err)
		os.Exit(1)
	}

	// 准备请求
	temperature := 0.7
	request := &api.Request{
		Model: models.GeminiPro,
		Messages: []api.Message{
			{
				Role:    api.RoleSystem,
				Content: "你是一个专业、友好的AI助手。",
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
		fmt.Printf("Gemini回复: %s\n", response.Choices[0].Message.Content)
	}
}
```

## 高级用法

### 多提供商流式处理示例

```go
// 创建多个提供商的客户端
openaiClient, _ := openai.NewClient(func(o *api.ClientOptions) { o.APIKey = openaiKey })
anthropicClient, _ := anthropic.NewClient(func(o *api.ClientOptions) { o.APIKey = anthropicKey })
deepseekClient, _ := deepseek.NewClient(func(o *api.ClientOptions) { o.APIKey = deepseekKey })
geminiClient, _ := gemini.NewClient(func(o *api.ClientOptions) { o.APIKey = geminiKey })

// 创建相同的请求
request := &api.Request{
    Model: models.DeepSeekChat, // 根据客户端选择对应模型
    Messages: []api.Message{
        {Role: api.RoleUser, Content: "简单介绍一下Go语言的特点"},
    },
}

// 发送流式请求
stream, err := client.CompleteStream(ctx, request)
if err != nil {
    // 处理错误
}

// 使用统一的流处理接口处理不同提供商的响应
err = api.NewStreamProcessor().ProcessWithHandler(stream, func(text string) error {
    fmt.Print(text) // 实时输出
    return nil
})
```

### 配置客户端选项

```go
client, err := openai.NewClient(
	func(options *api.ClientOptions) {
		options.APIKey = apiKey
		options.BaseURL = "https://your-proxy-server.com/v1"
		options.Timeout = 60  // 超时时间(秒)
		options.MaxRetries = 3
	},
)
```

### 流式响应

#### 基本流式处理

```go
// 发送流式请求
stream, err := client.CompleteStream(ctx, request)
if err != nil {
	// 处理错误
}
defer stream.Close()

// 逐块接收响应
for {
	chunk, err := stream.Recv()
	if err == io.EOF {
		break
	}
	if err != nil {
		// 处理错误
		break
	}
	
	// 处理响应块
	if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
		fmt.Print(chunk.Choices[0].Delta.Content)
	}
}
```

#### 使用流处理器 (推荐用法)

```go
// 发送流式请求
stream, err := client.CompleteStream(ctx, request)
if err != nil {
	// 处理错误
}

// 创建流处理器并设置回调选项
processor := api.NewStreamProcessor()
err = processor.Process(stream, &api.StreamOptions{
	// 当收到文本块时的回调
	OnText: func(text string) error {
		fmt.Print(text) // 实时打印
		return nil
	},
	// 当流处理完成时的回调
	OnComplete: func(err error) {
		if err != nil {
			fmt.Printf("\n处理过程中出错: %v\n", err)
		} else {
			fmt.Printf("\n处理完成!\n")
		}
	},
	// 自动关闭流
	AutoClose: true,
})
```

#### 简化的流处理

```go
// 只关注文本内容
err = api.NewStreamProcessor().ProcessWithHandler(stream, func(text string) error {
	fmt.Print(text)
	return nil
})
```

#### 收集完整内容

```go
// 从流式响应中收集完整内容
content, err := api.CollectFullContent(stream)
if err != nil {
	// 处理错误
}
fmt.Printf("收集到的完整内容: %s\n", content)
```

#### 输出到写入器

```go
// 将流式响应直接输出到标准输出
err = api.StreamToWriter(stream, os.Stdout)
if err != nil {
	// 处理错误
}
```

### 生成嵌入向量

```go
// 生成文本的嵌入向量
embedding, err := client.Embedding(ctx, "这是一段需要生成嵌入向量的文本")
if err != nil {
	// 处理错误
}

// 使用嵌入向量进行相似度计算等操作
fmt.Printf("嵌入向量维度: %d\n", len(embedding))
```

## 项目结构

```
/llm-sdk
  /pkg
    /api        # 核心接口定义
    /providers  # 不同LLM提供商实现
      /openai
      /anthropic
      /deepseek
      /gemini   
    /models     # 模型定义与参数
    /utils      # 通用工具函数
  /examples     # 使用示例
    /anthropic  # Anthropic示例
    /deepseek   # DeepSeek示例
    /streaming  # 流式处理示例
    /multi_provider_streaming  # 多提供商流式处理示例
```

## 已实现功能

- [x] 统一的API接口设计
- [x] OpenAI 提供商支持
- [x] Anthropic 提供商支持
- [x] DeepSeek 提供商支持
- [x] Google Gemini 提供商支持
- [x] 流式响应支持（SSE）
- [x] 嵌入向量支持

## 待实现功能

- [ ] 函数调用支持
- [ ] 多模态输入支持
- [ ] 缓存机制

## 许可证

MIT 