package main

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	ctxpkg "github.com/elihe999/go-claw-demo/internal/context"
	"github.com/elihe999/go-claw-demo/internal/engine"
	"github.com/elihe999/go-claw-demo/internal/provider"
	"github.com/elihe999/go-claw-demo/internal/schema"
	"github.com/elihe999/go-claw-demo/internal/tools"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Overload(); err != nil {
		log.Printf("未加载 .env（将回退到系统环境变量）: %v", err)
	}
	workDir, _ := os.Getwd()
	workDir += "/workspace"

	// 1. 初始化真实的 Provider大脑
	// 可切换：NewZhipuOpenAIProvider / NewZhipuClaudeProvider / NewAgnesOpenAIProvider / NewOpenRouterProvider / NewSenseNovaProvider
	// llmProvider := provider.NewZhipuOpenAIProvider("glm-4.7-flash")
	// llmProvider := provider.NewAgnesOpenAIProvider("agnes-2.0-flash")
	// llmProvider := provider.NewOpenRouterProvider("openrouter/free")
	llmProvider := provider.NewSenseNovaProvider("sensenova-6.8-flash-lite")

	// 3. 初始化真实的 Tool Registry
	registry := tools.NewRegistry()
	// registry.Register(tools.NewReadFileTool(workDir))
	// registry.Register(tools.NewWriteFileTool(workDir))
	// registry.Register(tools.NewBashTool(workDir))
	// registry.Register(tools.NewEditFileTool(workDir))
	registry.Register(tools.NewReadFileTool("./tmp/project_front"))
	// 引擎本身变成无状态的，它不绑定 WorkDir（仅适用于本讲演示）
	eng := engine.NewAgentEngine(llmProvider, registry, false)
	reporter := engine.NewTerminalReporter()
	var wg sync.WaitGroup
	// ================= 模拟并发场景 1：飞书前端群 =================
	wg.Add(1)
	go func() {
		defer wg.Done()
		sessionA := ctxpkg.GlobalSessionMgr.GetOrCreate("chat_front_001", "/tmp/project_front")
		// 回合 1
		log.Println("\n>>> 🙋‍♂️ [Session A / Turn 1]: 帮我看看 README.md 里记录了什么密钥？")
		sessionA.Append(schema.Message{Role: schema.RoleUser, Content: "帮我看看 README.md 里记录了什么密钥？"})
		_ = eng.Run(context.Background(), sessionA, reporter)
		// 故意制造大量“废话”对话，刷掉记忆 (假设 Working Memory Limit=6)
		for i := 0; i < 6; i++ {
			sessionA.Append(schema.Message{Role: schema.RoleUser, Content: "这只是一句闲聊占位符。"})
			sessionA.Append(schema.Message{Role: schema.RoleAssistant, Content: "好的，收到闲聊。"})
		}
		// 回合 2：验证记忆截断 (此时第一轮的密钥已经被挤出 Working Memory 了！)
		log.Println("\n>>> 🙋‍♂️ [Session A / Turn 2]: 请直接告诉我，刚才第一轮你查到的那个密钥是什么？")
		sessionA.Append(schema.Message{Role: schema.RoleUser, Content: "请直接告诉我，刚才第一轮你查到的那个密钥是什么？不准调用工具！"})
		_ = eng.Run(context.Background(), sessionA, reporter)
	}()
	// ================= 模拟并发场景 2：飞书后端群 =================
	wg.Add(1)
	go func() {
		defer wg.Done()
		// 稍微错开一点时间发起请求
		time.Sleep(1 * time.Second)
		sessionB := ctxpkg.GlobalSessionMgr.GetOrCreate("chat_back_002", "/tmp/project_back")
		log.Println("\n>>> 🙋‍♂️ [Session B]: 别人查到了一个密钥，你这里能看到吗？")
		sessionB.Append(schema.Message{Role: schema.RoleUser, Content: "别人查到了一个密钥，你这里能看到吗？不准调用工具！"})
		_ = eng.Run(context.Background(), sessionB, reporter)
	}()
	wg.Wait()
}
