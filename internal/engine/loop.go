package engine

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/elihe999/go-claw-demo/internal/provider"
	"github.com/elihe999/go-claw-demo/internal/schema"
	"github.com/elihe999/go-claw-demo/internal/tools"
)

// AgentEngine 是微型 OS 的核心驱动
type AgentEngine struct {
	provider provider.LLMProvider
	registry tools.Registry

	// WorkDir (工作区): 借鉴 OpenClaw 的理念，Agent 必须有一个明确的物理边界
	WorkDir string
}

func NewAgentEngine(p provider.LLMProvider, r tools.Registry, workDir string) *AgentEngine {
	return &AgentEngine{
		provider: p,
		registry: r,
		WorkDir:  workDir,
	}
}

// Run 启动 Agent 的生命周期
func (e *AgentEngine) Run(ctx context.Context, userPrompt string) error {
	log.Printf("[Engine] 引擎启动，锁定工作区: %s\n", e.WorkDir)

	// 1. 初始化会话的 Context (上下文内存)
	// 在真实的场景中，这里会由动态 Prompt 组装器加载 AGENTS.md。目前我们先硬编码。
	contextHistory := []schema.Message{
		{
			Role:    schema.RoleSystem,
			Content: "You are go-tiny-claw, an expert coding assistant. You have full access to tools in the workspace.",
		},
		{
			Role:    schema.RoleUser,
			Content: userPrompt,
		},
	}

	turnCount := 0

	// 2. The Main Loop: 心跳开始 (标准的 ReAct 循环)
	for {
		turnCount++
		log.Printf("========== [Turn %d] 开始 ==========\n", turnCount)

		// 获取当前挂载的所有工具定义
		availableTools := e.registry.GetAvailableTools()

		// 向大模型发起推理请求 (包含 Reasoning)
		log.Println("[Engine] 正在思考 (Reasoning)...")
		responseMsg, err := e.provider.Generate(ctx, contextHistory, availableTools)
		if err != nil {
			return fmt.Errorf("模型生成失败: %w", err)
		}

		// 将模型的响应完整追加到上下文历史中
		contextHistory = append(contextHistory, *responseMsg)

		// 如果模型回复了纯文本，打印出来 (这通常是它的思考过程，或是最终结果)
		if responseMsg.Content != "" {
			fmt.Printf("🤖 模型: %s\n", responseMsg.Content)
		}

		// 3. 退出条件判断
		// 如果模型没有请求任何工具调用，说明它认为任务已经完成，跳出循环。
		if len(responseMsg.ToolCalls) == 0 {
			log.Println("[Engine] 任务完成，退出循环。")
			break
		}

		// 4. 执行行动 (Action) 与 获取观察结果 (Observation)
		log.Printf("[Engine] 模型请求调用 %d 个工具...\n", len(responseMsg.ToolCalls))

		// ext
		results := make([]schema.Message, len(responseMsg.ToolCalls))
		var wg sync.WaitGroup

		for i, toolCall := range responseMsg.ToolCalls {
			wg.Add(1)
			
			go func(index int, call schema.ToolCall) {
				defer wg.Done()
				
				result := e.registry.Execute(ctx, call)
				
				results[index] = schema.Message{
					Role:       schema.RoleUser,
					Content:    result.Output,
					ToolCallID: call.ID,
				}
			}(i, toolCall)
		}
		
		wg.Wait()
		contextHistory = append(contextHistory, results...)

		// 循环回到开头，模型将带着新加入的 Observation 继续它的下一轮思考...
	}
	return nil
}
