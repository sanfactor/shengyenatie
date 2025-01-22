package tools

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/user/modulox/pkg/types"
)

// ToolRegistry manages tool registration and discovery
type ToolRegistry struct {
	tools      map[string]types.Tool
	mu         sync.RWMutex
	validators map[string]func(interface{}) error
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools:      make(map[string]types.Tool),
		validators: make(map[string]func(interface{}) error),
	}
}

// RegisterTool adds a tool to the registry with type validation
func (tr *ToolRegistry) RegisterTool(name string, tool types.Tool, validator func(interface{}) error) error {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	if _, exists := tr.tools[name]; exists {
		return fmt.Errorf("tool already registered: %s", name)
	}

	tr.tools[name] = tool
	if validator != nil {
		tr.validators[name] = validator
	}

	return nil
}

// ExecuteTool runs a tool with type-safe input validation
func (tr *ToolRegistry) ExecuteTool(name string, input interface{}) (interface{}, error) {
	tr.mu.RLock()
	tool, exists := tr.tools[name]
	validator := tr.validators[name]
	tr.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	if validator != nil {
		if err := validator(input); err != nil {
			return nil, fmt.Errorf("input validation failed: %w", err)
		}
	}

	return tool.Execute(input)
}

// DiscoverCapabilities returns all registered tool capabilities
func (tr *ToolRegistry) DiscoverCapabilities() []types.Capability {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	capabilities := make([]types.Capability, 0, len(tr.tools))
	for name, tool := range tr.tools {
		capabilities = append(capabilities, types.Capability{
			Name:        name,
			Description: tool.GetDescription(),
		})
	}

	return capabilities
}
