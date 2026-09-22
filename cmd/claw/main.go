package main

import (
	"context"
	"log"
	"os"

	"github.com/elihe999/go-claw-demo/internal/engine"
	"github.com/elihe999/go-claw-demo/internal/provider"
	"github.com/elihe999/go-claw-demo/internal/schema"
	"github.com/elihe999/go-claw-demo/internal/tools"
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

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("未加载 .env（将回退到系统环境变量）: %v", err)
	}
	workDir, _ := os.Getwd()
	// 1. 初始化真实的 Provider大脑
	// 可切换：NewZhipuOpenAIProvider / NewZhipuClaudeProvider / NewAgnesOpenAIProvider
	llmProvider := provider.NewZhipuOpenAIProvider("glm-4.7-flash")
	// llmProvider := provider.NewAgnesOpenAIProvider("agnes-2.0-flash")

	// 3. 初始化真实的 Tool Registry
	registry := tools.NewRegistry()

	// 4. 将真实的 ReadFile 工具挂载到注册表中
	readFileTool := tools.NewReadFileTool(workDir)
	writeFileTool := tools.NewWriteFileTool(workDir)
	bashTool := tools.NewBashTool(workDir)
	registry.Register(readFileTool)
	registry.Register(writeFileTool)
	registry.Register(bashTool)

	// 实例化核心引擎，关闭慢思考阶段，享受 YOLO 急速模式
	eng := engine.NewAgentEngine(llmProvider, registry, workDir, false)
	// 发起一个需要连贯物理动作的任务
	prompt := ` 请帮我执行以下操作： 1. 用 bash 查看一下我当前电脑的 Go 版本。 2. 帮我写一个简单的 helloworld.go 文件，输出 "Hello, go-tiny-claw!"。 3. 用 bash 编译并运行这个 go 文件，确认它能正常工作。 `
	err := eng.Run(context.Background(), prompt)
	if err != nil {
		log.Fatalf("引擎运行崩溃: %v", err)
	}
}
