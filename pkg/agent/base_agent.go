package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/user/modulox/pkg/llm"
	"github.com/user/modulox/pkg/memory"
	"github.com/user/modulox/pkg/tools"
	"github.com/user/modulox/pkg/types"
)

// BaseAgentConfig contains configuration for base agent
type BaseAgentConfig struct {
	Name        string
	Description string
	Provider    llm.Provider
	Memory      memory.VectorStore
	Registry    *tools.ToolRegistry
}

// BaseAgent provides a complete implementation of the Agent interface
type BaseAgent struct {
	config      BaseAgentConfig
	tools       *tools.ToolRegistry
	executor    *tools.SafeExecutor
	memory      memory.VectorStore
	provider    llm.Provider
	mu          sync.RWMutex
}

// NewBaseAgent creates a new base agent instance
func NewBaseAgent(config BaseAgentConfig) *BaseAgent {
	executor := tools.NewSafeExecutor(config.Registry)
	return &BaseAgent{
		config:   config,
		tools:    config.Registry,
		executor: executor,
		memory:   config.Memory,
		provider: config.Provider,
	}
}

// Execute implements Agent.Execute
func (b *BaseAgent) Execute(ctx context.Context, input string) (string, error) {
	// First, check memory for relevant context
	embedding, err := b.provider.Embed(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to create embedding: %w", err)
	}

	vectors, err := b.memory.Query(ctx, types.Vector{Values: embedding}, 5)
	if err != nil {
		return "", fmt.Errorf("failed to query memory: %w", err)
	}

	// Build context from memory
	context := buildContext(vectors)

	// Generate completion with context
	prompt := fmt.Sprintf("Context:\n%s\n\nInput: %s", context, input)
	completion, err := b.provider.Complete(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate completion: %w", err)
	}

	// Store the interaction in memory
	b.memory.Store(ctx, []types.Vector{{
		ID:     fmt.Sprintf("interaction_%d", time.Now().UnixNano()),
		Values: embedding,
		Metadata: map[string]interface{}{
			"input":  input,
			"output": completion,
		},
	}})

	return completion, nil
}

// AddTool implements Agent.AddTool
func (b *BaseAgent) AddTool(tool types.Tool) error {
	return b.tools.RegisterTool(tool.GetDescription(), tool, nil)
}

// GetCapabilities implements Agent.GetCapabilities
func (b *BaseAgent) GetCapabilities() []types.Capability {
	return b.tools.DiscoverCapabilities()
}

// Helper function to build context from memory vectors
func buildContext(vectors []types.Vector) string {
	var context string
	for _, v := range vectors {
		if input, ok := v.Metadata["input"].(string); ok {
			if output, ok := v.Metadata["output"].(string); ok {
				context += fmt.Sprintf("Q: %s\nA: %s\n\n", input, output)
			}
		}
	}
	return context
}
