package tools

import (
	"encoding/json"
	"fmt"
	"plugin"
	"reflect"
	"sync"

	"github.com/user/modulox/pkg/types"
)

// ToolPlugin represents a dynamically loaded tool
type ToolPlugin struct {
	Name        string
	Description string
	Execute     func(input interface{}) (interface{}, error)
}

// PluginManager manages dynamic tool plugins
type PluginManager struct {
	plugins map[string]*ToolPlugin
	mu      sync.RWMutex
}

// NewPluginManager creates a new plugin manager
func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make(map[string]*ToolPlugin),
	}
}

// LoadPlugin loads a tool plugin from a .so file
func (pm *PluginManager) LoadPlugin(path string) error {
	p, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open plugin: %w", err)
	}

	// Load plugin metadata
	metadataSymbol, err := p.Lookup("ToolMetadata")
	if err != nil {
		return fmt.Errorf("plugin metadata not found: %w", err)
	}

	metadata, ok := metadataSymbol.(*ToolPlugin)
	if !ok {
		return fmt.Errorf("invalid plugin metadata type")
	}

	// Load execute function
	executeSymbol, err := p.Lookup("Execute")
	if err != nil {
		return fmt.Errorf("execute function not found: %w", err)
	}

	execute, ok := executeSymbol.(func(interface{}) (interface{}, error))
	if !ok {
		return fmt.Errorf("invalid execute function type")
	}

	metadata.Execute = execute

	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.plugins[metadata.Name] = metadata

	return nil
}

// GetPlugin retrieves a loaded plugin by name
func (pm *PluginManager) GetPlugin(name string) (*ToolPlugin, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	plugin, exists := pm.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", name)
	}
	return plugin, nil
}
