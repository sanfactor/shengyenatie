package distributed

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/user/modulox/pkg/agent"
	"github.com/user/modulox/pkg/communication"
	"github.com/user/modulox/pkg/types"
)

// NodeConfig contains configuration for a distributed node
type NodeConfig struct {
	ID          string
	Address     string
	ClusterAddr string
	Tags        []string
}

// Node represents a single node in the distributed system
type Node struct {
	config    NodeConfig
	client    *communication.AgentClient
	agents    map[string]agent.Agent
	capacity  int
	load      int
	status    NodeStatus
	lastPing  time.Time
	mu        sync.RWMutex
}

// NodeStatus represents the current status of a node
type NodeStatus int

const (
	StatusUnknown NodeStatus = iota
	StatusHealthy
	StatusOverloaded
	StatusUnhealthy
)

// NewNode creates a new distributed node
func NewNode(config NodeConfig) (*Node, error) {
	client, err := communication.NewAgentClient(config.ClusterAddr, config.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent client: %w", err)
	}

	return &Node{
		config:    config,
		client:    client,
		agents:    make(map[string]agent.Agent),
		capacity:  100, // Default capacity
		status:    StatusHealthy,
		lastPing:  time.Now(),
	}, nil
}

// RegisterAgent registers an agent with the node
func (n *Node) RegisterAgent(a agent.Agent) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.load >= n.capacity {
		return fmt.Errorf("node is at capacity")
	}

	id := a.GetName()
	n.agents[id] = a
	n.load++

	// Publish agent registration event
	return n.client.PublishEvent(context.Background(), "agent_registered",
		fmt.Sprintf("Agent %s registered on node %s", id, n.config.ID),
		map[string]string{
			"agent_id": id,
			"node_id":  n.config.ID,
		})
}

// ExecuteTask executes a task on an agent
func (n *Node) ExecuteTask(ctx context.Context, agentID string, task string) (string, error) {
	n.mu.RLock()
	agent, exists := n.agents[agentID]
	n.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("agent not found: %s", agentID)
	}

	// Publish task start event
	err := n.client.PublishEvent(ctx, "task_start",
		fmt.Sprintf("Starting task on agent %s", agentID),
		map[string]string{
			"agent_id": agentID,
			"node_id":  n.config.ID,
		})
	if err != nil {
		return "", fmt.Errorf("failed to publish start event: %w", err)
	}

	// Execute task
	result, err := agent.Execute(ctx, task)
	if err != nil {
		n.client.PublishEvent(ctx, "task_error",
			fmt.Sprintf("Task failed on agent %s: %v", agentID, err),
			map[string]string{
				"agent_id": agentID,
				"node_id":  n.config.ID,
			})
		return "", fmt.Errorf("task execution failed: %w", err)
	}

	// Publish task completion event
	n.client.PublishEvent(ctx, "task_complete",
		fmt.Sprintf("Task completed on agent %s", agentID),
		map[string]string{
			"agent_id": agentID,
			"node_id":  n.config.ID,
		})

	return result, nil
}

// GetStatus returns the current node status
func (n *Node) GetStatus() types.NodeStatus {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return types.NodeStatus{
		ID:        n.config.ID,
		Address:   n.config.Address,
		Load:      n.load,
		Capacity:  n.capacity,
		Status:    int(n.status),
		LastPing:  n.lastPing,
		AgentCount: len(n.agents),
	}
}

// UpdateStatus updates the node's status
func (n *Node) UpdateStatus() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.lastPing = time.Now()
	if float64(n.load)/float64(n.capacity) > 0.8 {
		n.status = StatusOverloaded
	} else {
		n.status = StatusHealthy
	}
}

// Close closes the node and its connections
func (n *Node) Close() error {
	return n.client.Close()
}
