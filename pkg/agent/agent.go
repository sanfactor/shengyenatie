package agent

import (
	"context"

	"github.com/user/modulox/pkg/types"
)

// Agent defines the interface for an AI agent
type Agent interface {
	// Execute runs the agent with the given input
	Execute(ctx context.Context, input string) (string, error)
	
	// AddTool adds a new tool to the agent's capabilities
	AddTool(tool types.Tool) error
	
	// GetCapabilities returns the list of agent's capabilities
	GetCapabilities() []types.Capability
}

// BaseAgent provides a basic implementation of the Agent interface
type BaseAgent struct {
	tools        []types.Tool
	capabilities []types.Capability
}

// NewBaseAgent creates a new instance of BaseAgent
func NewBaseAgent() *BaseAgent {
	return &BaseAgent{
		tools:        make([]types.Tool, 0),
		capabilities: make([]types.Capability, 0),
	}
}

// AddTool implements Agent.AddTool
func (b *BaseAgent) AddTool(tool types.Tool) error {
	b.tools = append(b.tools, tool)
	b.capabilities = append(b.capabilities, types.Capability{
		Name:        "tool",
		Description: tool.GetDescription(),
	})
	return nil
}

// GetCapabilities implements Agent.GetCapabilities
func (b *BaseAgent) GetCapabilities() []types.Capability {
	return b.capabilities
}
