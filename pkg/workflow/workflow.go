package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/user/modulox/pkg/agent"
	"github.com/user/modulox/pkg/types"
	"github.com/user/modulox/pkg/communication"
)

// Workflow defines the interface for agent workflow orchestration
type Workflow interface {
	// Execute runs the workflow with the given task
	Execute(ctx context.Context, task string) (string, error)
	// AddAgent adds an agent to the workflow
	AddAgent(agent agent.Agent) error
}

// SequentialWorkflow implements sequential execution of agents
type SequentialWorkflow struct {
	agents  []agent.Agent
	results chan types.WorkflowResult
}

// NewSequentialWorkflow creates a new sequential workflow
func NewSequentialWorkflow() *SequentialWorkflow {
	return &SequentialWorkflow{
		agents:  make([]agent.Agent, 0),
		results: make(chan types.WorkflowResult),
	}
}

// Execute implements Workflow.Execute for sequential processing
func (w *SequentialWorkflow) Execute(ctx context.Context, task string) (string, error) {
	var finalResult string
	var err error

	for i, agent := range w.agents {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
			result, execErr := agent.Execute(ctx, task)
			if execErr != nil {
				return "", execErr
			}
			// For sequential workflow, each agent's input is previous agent's output
			task = result
			if i == len(w.agents)-1 {
				finalResult = result
			}
		}
	}

	return finalResult, err
}

// AddAgent implements Workflow.AddAgent
func (w *SequentialWorkflow) AddAgent(a agent.Agent) error {
	w.agents = append(w.agents, a)
	return nil
}

// MixtureWorkflow implements parallel execution with result aggregation
type MixtureWorkflow struct {
	agents     []agent.Agent
	aggregator agent.Agent
	results    chan types.WorkflowResult
}

// NewMixtureWorkflow creates a new mixture workflow
func NewMixtureWorkflow(aggregator agent.Agent) *MixtureWorkflow {
	return &MixtureWorkflow{
		agents:     make([]agent.Agent, 0),
		aggregator: aggregator,
		results:    make(chan types.WorkflowResult),
	}
}

// Execute implements Workflow.Execute for parallel processing
func (w *MixtureWorkflow) Execute(ctx context.Context, task string) (string, error) {
	// Create event publisher
	client, err := communication.NewAgentClient("localhost:50051", "mixture-workflow")
	if err != nil {
		return "", fmt.Errorf("failed to create event client: %w", err)
	}
	defer client.Close()

	// Publish workflow start event
	err = client.PublishEvent(ctx, "workflow_start",
		fmt.Sprintf("Starting mixture workflow with %d agents", len(w.agents)),
		map[string]string{"num_agents": fmt.Sprintf("%d", len(w.agents))})
	if err != nil {
		return "", fmt.Errorf("failed to publish start event: %w", err)
	}

	var wg sync.WaitGroup
	results := make([]string, len(w.agents))
	errors := make(chan error, len(w.agents))

	// Execute all agents in parallel
	for i, agent := range w.agents {
		wg.Add(1)
		go func(index int, a agent.Agent) {
			defer wg.Done()

			// Publish agent start event
			client.PublishEvent(ctx, "agent_start",
				fmt.Sprintf("Starting agent %d: %s", index+1, a.GetName()),
				map[string]string{"agent_index": fmt.Sprintf("%d", index+1)})

			result, err := a.Execute(ctx, task)
			if err != nil {
				client.PublishEvent(ctx, "agent_error",
					fmt.Sprintf("Agent %d failed: %v", index+1, err),
					map[string]string{"agent_index": fmt.Sprintf("%d", index+1)})
				errors <- fmt.Errorf("agent %d failed: %w", index, err)
				return
			}

			// Store result in synchronized state
			version, err := client.SyncState(ctx,
				fmt.Sprintf("agent_%d_result", index+1), result)
			if err != nil {
				errors <- fmt.Errorf("failed to sync state: %w", err)
				return
			}

			// Publish agent complete event
			client.PublishEvent(ctx, "agent_complete",
				fmt.Sprintf("Agent %d completed", index+1),
				map[string]string{
					"agent_index": fmt.Sprintf("%d", index+1),
					"state_version": fmt.Sprintf("%d", version),
					"result_length": fmt.Sprintf("%d", len(result)),
				})

			results[index] = result
		}(i, agent)
	}

	// Wait for all agents to complete
	wg.Wait()
	close(errors)

	// Check for errors
	select {
	case err := <-errors:
		client.PublishEvent(ctx, "workflow_error",
			fmt.Sprintf("Workflow failed: %v", err),
			nil)
		return "", err
	default:
	}

	// Publish aggregation start event
	client.PublishEvent(ctx, "aggregation_start",
		"Starting result aggregation",
		map[string]string{"num_results": fmt.Sprintf("%d", len(results))})

	// Aggregate results using the aggregator agent
	aggregatedInput := fmt.Sprintf("Aggregate the following results:\n%s", stringSliceToString(results))
	finalResult, err := w.aggregator.Execute(ctx, aggregatedInput)
	if err != nil {
		client.PublishEvent(ctx, "aggregation_error",
			fmt.Sprintf("Aggregation failed: %v", err),
			nil)
		return "", fmt.Errorf("aggregation failed: %w", err)
	}

	// Publish workflow complete event
	client.PublishEvent(ctx, "workflow_complete",
		"Mixture workflow completed successfully",
		map[string]string{"final_result_length": fmt.Sprintf("%d", len(finalResult))})

	return finalResult, nil
}

// AddAgent implements Workflow.AddAgent
func (w *MixtureWorkflow) AddAgent(a agent.Agent) error {
	w.agents = append(w.agents, a)
	return nil
}

// Helper function to convert string slice to string
func stringSliceToString(slice []string) string {
	result := ""
	for i, s := range slice {
		result += s
		if i < len(slice)-1 {
			result += "\n"
		}
	}
	return result
}
