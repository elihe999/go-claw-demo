package provider

import "os"

const defaultAgnesBaseURL = "https://apihub.agnes-ai.com/v1/"

// NewAgnesOpenAIProvider 基于 OpenAI V3 SDK，指向 Agnes AI 兼容端点。
// 环境变量：AGNES_API_KEY（必填），AGNES_BASE_URL（可选，覆盖默认网关）。
func NewAgnesOpenAIProvider(model string) *OpenAIProvider {
	baseURL := os.Getenv("AGNES_BASE_URL")
	if baseURL == "" {
		baseURL = defaultAgnesBaseURL
	}
	return newOpenAICompatibleProvider("AGNES_API_KEY", baseURL, model)
}
