package provider

const defaultOpenRouterBaseURL = "https://openrouter.ai/api/v1/"

// NewOpenRouterProvider creates an OpenAI-compatible provider for OpenRouter.
// Set OPENROUTER_API_KEY before calling it.
func NewOpenRouterProvider(model string) *OpenAIProvider {
	return newOpenAICompatibleProvider("OPENROUTER_API_KEY", defaultOpenRouterBaseURL, model)
}
