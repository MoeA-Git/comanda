package models

// ModelConfig represents configuration options for model calls
type ModelConfig struct {
	Temperature float64
	MaxTokens   int
	TopP        float64
}

// Provider represents a model provider (e.g., Anthropic, OpenAI)
type Provider interface {
	Name() string
	SupportsModel(modelName string) bool
	SendPrompt(modelName string, prompt string) (string, error)
	Configure(apiKey string) error
	SetVerbose(verbose bool)
}

// DetectProviderFunc is the type for the provider detection function
type DetectProviderFunc func(modelName string) Provider

// DetectProvider determines the appropriate provider based on the model name
var DetectProvider DetectProviderFunc = defaultDetectProvider

// defaultDetectProvider is the default implementation of DetectProvider
func defaultDetectProvider(modelName string) Provider {
	providers := []Provider{
		NewAnthropicProvider(),
		NewOpenAIProvider(),
		NewOllamaProvider(),
	}

	for _, provider := range providers {
		if provider.SupportsModel(modelName) {
			return provider
		}
	}
	return nil
}
