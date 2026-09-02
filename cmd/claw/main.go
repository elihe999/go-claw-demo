package main

import (
	"context"
	"log"
	"os"

	"github.com/elihe999/go-claw-demo/internal/engine"
	"github.com/elihe999/go-claw-demo/internal/provider"
	"github.com/elihe999/go-claw-demo/internal/schema"
	"github.com/joho/godotenv"
)

// 升级版 Mock Provider
type mockProvider struct {
	turn int
}

func (m *mockProvider) Generate(ctx context.Context, msgs []schema.Message, tools []schema.ToolDefinition) (*schema.Message, error) {
	// 如果工具列表为空，说明这是引擎发起的 Phase 1: Thinking 阶段
	if len(tools) == 0 {
		return &schema.Message{
			Role:    schema.RoleAssistant,
			Content: "【推理中】目标是检查文件。我不能直接盲猜，我需要先调用 bash 工具执行 ls 命令，看看当前目录下有什么，然后再做定夺。",
		}, nil
	}

	// 如果工具列表不为空，说明这是 Phase 2: Action 阶段
	m.turn++
	if m.turn == 1 {
		// 第一轮 Action：顺着刚才的 Thinking，精准调用工具
		return &schema.Message{
			Role:    schema.RoleAssistant,
			Content: "我要执行我刚才计划的步骤了。",
			ToolCalls: []schema.ToolCall{
				{ID: "call_123", Name: "bash", Arguments: []byte(`{"command": "ls -la"}`)},
			},
		}, nil
	}

	// 第二轮 Action：直接总结退出
	return &schema.Message{
		Role:    schema.RoleAssistant,
		Content: "根据工具返回的结果，我看到了 main.go，任务圆满完成！",
	}, nil
}

type mockRegistry struct{}

func (m *mockRegistry) GetAvailableTools() []schema.ToolDefinition {
	return []schema.ToolDefinition{
		{
			Name:        "get_weather",
			Description: "获取指定城市的当前天气情况。",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"city": map[string]interface{}{
						"type": "string",
					},
				},
				"required": []string{"city"},
			},
		},
	}
}

func (m *mockRegistry) Execute(ctx context.Context, call schema.ToolCall) schema.ToolResult {
	log.Printf(" -> [Mock 工具执行] 获取 %s 的天气中...\n", call.Name)
	return schema.ToolResult{ToolCallID: call.ID, Output: "API 返回：今天是晴天，气温 25 度。", IsError: false}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("未加载 .env（将回退到系统环境变量）: %v", err)
	}
	if os.Getenv("ZHIPU_API_KEY") == "" {
		log.Fatal("请先在 .env 中配置 ZHIPU_API_KEY，或导出该环境变量")
	}
	workDir, _ := os.Getwd()
	// 1. 初始化真实的 Provider大脑
	// 可切换：NewZhipuOpenAIProvider / NewZhipuClaudeProvider / NewAgnesOpenAIProvider
	// llmProvider := provider.NewZhipuOpenAIProvider("glm-4.7-flash")
	llmProvider := provider.NewAgnesOpenAIProvider("agnes-2.0-flash")
	// 2. 注入伪造的工具注册表
	registry := &mockRegistry{}
	// 3. 实例化并运行引擎，开启 EnableThinking = true (开启慢思考阶段！)
	eng := engine.NewAgentEngine(llmProvider, registry, workDir, true)
	// 设定测试任务
	prompt := "我想去北京跑步，帮我查查天气适合吗？"
	err := eng.Run(context.Background(), prompt)
	if err != nil {
		log.Fatalf("引擎运行崩溃: %v", err)
	}
}
