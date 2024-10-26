package processor

import (
	"fmt"

	"github.com/kris-hansen/comanda/utils/models"
)

var originalDetectProvider models.DetectProviderFunc

func init() {
	// Store the original function
	originalDetectProvider = models.DetectProvider

	// Override with test version
	models.DetectProvider = func(modelName string) models.Provider {
		providers := []models.Provider{
			NewMockProvider("OpenAI"),
			NewMockProvider("Anthropic"),
		}

		for _, provider := range providers {
			if provider.SupportsModel(modelName) {
				return provider
			}
		}
		return nil
	}
}

// Restore the original DetectProvider function
func restoreDetectProvider() {
	models.DetectProvider = originalDetectProvider
}

// MockProvider implements the models.Provider interface for testing
type MockProvider struct {
	name       string
	configured bool
	verbose    bool
	apiKey     string
}

func NewMockProvider(name string) *MockProvider {
	return &MockProvider{
		name: name,
	}
}

func (m *MockProvider) Name() string {
	return m.name
}

func (m *MockProvider) SupportsModel(modelName string) bool {
	validModels := map[string][]string{
		"OpenAI": {
			"gpt-4",
			"gpt-4o",
			"gpt-4o-mini",
			"o1-preview",
			"o1-mini",
		},
		"Anthropic": {
			"claude-2",
			"claude-instant",
		},
	}

	if models, ok := validModels[m.name]; ok {
		for _, validModel := range models {
			if modelName == validModel {
				return true
			}
		}
	}
	return false
}

func (m *MockProvider) Configure(apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("API key is required")
	}
	m.configured = true
	m.apiKey = apiKey
	return nil
}

func (m *MockProvider) SendPrompt(model, prompt string) (string, error) {
	if !m.configured {
		return "", fmt.Errorf("provider not configured")
	}
	if !m.SupportsModel(model) {
		return "", fmt.Errorf("unsupported model: %s", model)
	}
	return "mock response", nil
}

func (m *MockProvider) SetVerbose(verbose bool) {
	m.verbose = verbose
}
