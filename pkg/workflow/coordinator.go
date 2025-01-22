package workflow

import (
	"context"
	"fmt"
	"sync"

	"github.com/user/modulox/pkg/agent"
	"github.com/user/modulox/pkg/communication"
)

// Coordinator manages collaboration between multiple agents
type Coordinator struct {
	workflows map[string]Workflow
	client    *communication.AgentClient
	mu        sync.RWMutex
}

// NewCoordinator creates a new coordinator instance
func NewCoordinator(serverAddr string) (*Coordinator, error) {
	client, err := communication.NewAgentClient(serverAddr, "coordinator")
	if err != nil {
		return nil, fmt.Errorf("failed to create agent client: %w", err)
	}

	return &Coordinator{
		workflows: make(map[string]Workflow),
		client:    client,
	}, nil
}

// RegisterWorkflow adds a new workflow to the coordinator
func (c *Coordinator) RegisterWorkflow(name string, w Workflow) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.workflows[name] = w

	// Publish workflow registration event
	c.client.PublishEvent(context.Background(), "workflow_registered",
		fmt.Sprintf("Registered workflow: %s", name),
		map[string]string{"workflow_name": name})
}

// ExecuteWorkflow runs a specific workflow by name
func (c *Coordinator) ExecuteWorkflow(ctx context.Context, name string, task string) (string, error) {
	c.mu.RLock()
	workflow, exists := c.workflows[name]
	c.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("workflow not found: %s", name)
	}

	// Publish workflow execution start event
	err := c.client.PublishEvent(ctx, "workflow_execution_start",
		fmt.Sprintf("Starting execution of workflow: %s", name),
		map[string]string{
			"workflow_name": name,
			"task_length":   fmt.Sprintf("%d", len(task)),
		})
	if err != nil {
		return "", fmt.Errorf("failed to publish start event: %w", err)
	}

	// Execute workflow
	result, err := workflow.Execute(ctx, task)
	if err != nil {
		// Publish error event
		c.client.PublishEvent(ctx, "workflow_execution_error",
			fmt.Sprintf("Workflow %s failed: %v", name, err),
			map[string]string{"workflow_name": name})
		return "", fmt.Errorf("workflow execution failed: %w", err)
	}

	// Publish completion event
	c.client.PublishEvent(ctx, "workflow_execution_complete",
		fmt.Sprintf("Workflow %s completed successfully", name),
		map[string]string{
			"workflow_name": name,
			"result_length": fmt.Sprintf("%d", len(result)),
		})

	return result, nil
}

// Close closes the coordinator and its connections
func (c *Coordinator) Close() error {
	return c.client.Close()
}

// Error types
type WorkflowError string

func (e WorkflowError) Error() string { return string(e) }

const (
	ErrWorkflowNotFound = WorkflowError("workflow not found")
)
