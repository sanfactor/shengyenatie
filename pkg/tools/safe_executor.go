package tools

import (
	"context"
	"fmt"
	"reflect"
)

// SafeExecutor provides type-safe tool execution
type SafeExecutor struct {
	registry *ToolRegistry
}

// NewSafeExecutor creates a new safe executor
func NewSafeExecutor(registry *ToolRegistry) *SafeExecutor {
	return &SafeExecutor{
		registry: registry,
	}
}

// ExecuteWithType runs a tool with strict type checking
func (se *SafeExecutor) ExecuteWithType(ctx context.Context, name string, input interface{}, outputType reflect.Type) (interface{}, error) {
	result, err := se.registry.ExecuteTool(name, input)
	if err != nil {
		return nil, err
	}

	// Verify output type
	resultValue := reflect.ValueOf(result)
	if !resultValue.Type().AssignableTo(outputType) {
		return nil, fmt.Errorf("tool returned invalid type: expected %v, got %v", outputType, resultValue.Type())
	}

	return result, nil
}

// ValidateInput checks if input matches tool's expected input type
func (se *SafeExecutor) ValidateInput(name string, input interface{}) error {
	tool, err := se.registry.tools[name]
	if err != nil {
		return fmt.Errorf("tool not found: %s", name)
	}

	inputType := reflect.TypeOf(input)
	expectedType := reflect.TypeOf(tool).In(0)

	if !inputType.AssignableTo(expectedType) {
		return fmt.Errorf("invalid input type: expected %v, got %v", expectedType, inputType)
	}

	return nil
}
