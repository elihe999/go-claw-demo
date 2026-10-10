package provider

const defaultSenseNovaBaseURL = "https://token.sensenova.cn/v1/"

// NewSenseNovaProvider creates an OpenAI-compatible provider for SenseNova.
// Set SENSENOVA_API_KEY before calling it.
func NewSenseNovaProvider(model string) *OpenAIProvider {
	return newOpenAICompatibleProvider("SENSENOVA_API_KEY", defaultSenseNovaBaseURL, model)
}
