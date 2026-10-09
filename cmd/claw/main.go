package main

import (
	"context"
	"log"
	"os"

	"github.com/elihe999/go-claw-demo/internal/engine"
	"github.com/elihe999/go-claw-demo/internal/provider"
	"github.com/elihe999/go-claw-demo/internal/tools"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("未加载 .env（将回退到系统环境变量）: %v", err)
	}
	workDir, _ := os.Getwd()
	workDir += "/workspace"

	// 1. 初始化真实的 Provider大脑
	// 可切换：NewZhipuOpenAIProvider / NewZhipuClaudeProvider / NewAgnesOpenAIProvider / NewOpenRouterProvider
	llmProvider := provider.NewZhipuOpenAIProvider("glm-4.7-flash")
	// llmProvider := provider.NewAgnesOpenAIProvider("agnes-2.0-flash")
	// llmProvider := provider.NewOpenRouterProvider("openrouter/free")

	// 3. 初始化真实的 Tool Registry
	registry := tools.NewRegistry()
	registry.Register(tools.NewReadFileTool(workDir))
	registry.Register(tools.NewWriteFileTool(workDir))
	registry.Register(tools.NewBashTool(workDir))
	registry.Register(tools.NewEditFileTool(workDir))

	// // 实例化核心引擎，关闭慢思考阶段，享受 YOLO 急速模式
	eng := engine.NewAgentEngine(llmProvider, registry, workDir, false)
	// // 发起一个需要连贯物理动作的任务
	// prompt := ` 请帮我执行以下操作： 1. 用 bash 查看一下我当前电脑的 Go 版本。 2. 帮我写一个简单的 helloworld.go 文件，输出 "Hello, go-claw!"。 3. 用 bash 编译并运行这个 go 文件，确认它能正常工作。 `
	// err := eng.Run(context.Background(), prompt)
	// if err != nil {
	// 	log.Fatalf("引擎运行崩溃: %v", err)
	// }

	// D7
	// 【新增挂载】
	reporter := engine.NewTerminalReporter()
	prompt := ` 我需要在当前目录下新建一个 ping.go，提供一个简单的 http ping 接口。 写完之后，帮我把代码用 git 提交一下。 `
	err := eng.Run(context.Background(), prompt, reporter)
	if err != nil {
		log.Fatalf("引擎运行崩溃: %v", err)
	}
}
