package llm

import (
	"context"
)

// Provider defines the interface for language model providers
type Provider interface {
	// Complete generates a completion for the given prompt
	Complete(ctx context.Context, prompt string) (string, error)
	
	// Embed generates embeddings for the given text
	Embed(ctx context.Context, text string) ([]float32, error)
}

// ProviderConfig contains configuration for LLM providers
type ProviderConfig struct {
	ModelName    string
	Temperature  float64
	MaxTokens    int
	APIKey       string
	BaseURL      string
	ExtraParams  map[string]interface{}
}

// BaseProvider provides common functionality for LLM providers
type BaseProvider struct {
	config ProviderConfig
}

// NewBaseProvider creates a new instance of BaseProvider
func NewBaseProvider(config ProviderConfig) *BaseProvider {
	return &BaseProvider{
		config: config,
	}
}
