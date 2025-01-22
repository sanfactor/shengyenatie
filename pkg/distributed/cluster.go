package distributed

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/user/modulox/pkg/communication"
	"github.com/user/modulox/pkg/types"
)

// ClusterConfig contains configuration for the distributed cluster
type ClusterConfig struct {
	Address     string
	HeartbeatInterval time.Duration
	NodeTimeout      time.Duration
}

// Cluster manages a collection of distributed nodes
type Cluster struct {
	config    ClusterConfig
	nodes     map[string]*Node
	client    *communication.AgentClient
	mu        sync.RWMutex
}

// NewCluster creates a new distributed cluster
func NewCluster(config ClusterConfig) (*Cluster, error) {
	client, err := communication.NewAgentClient(config.Address, "cluster")
	if err != nil {
		return nil, fmt.Errorf("failed to create agent client: %w", err)
	}

	cluster := &Cluster{
		config: config,
		nodes:  make(map[string]*Node),
		client: client,
	}

	// Start heartbeat monitoring
	go cluster.monitorHeartbeats()

	return cluster, nil
}

// RegisterNode registers a new node with the cluster
func (c *Cluster) RegisterNode(node *Node) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.nodes[node.config.ID]; exists {
		return fmt.Errorf("node already registered: %s", node.config.ID)
	}

	c.nodes[node.config.ID] = node

	// Publish node registration event
	return c.client.PublishEvent(context.Background(), "node_registered",
		fmt.Sprintf("Node %s registered with cluster", node.config.ID),
		map[string]string{
			"node_id":  node.config.ID,
			"address": node.config.Address,
		})
}

// GetNode returns a node by ID
func (c *Cluster) GetNode(id string) (*Node, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	node, exists := c.nodes[id]
	if !exists {
		return nil, fmt.Errorf("node not found: %s", id)
	}

	return node, nil
}

// GetHealthyNodes returns a list of healthy nodes
func (c *Cluster) GetHealthyNodes() []*Node {
	c.mu.RLock()
	defer c.mu.RUnlock()

	healthy := make([]*Node, 0)
	for _, node := range c.nodes {
		if node.status == StatusHealthy {
			healthy = append(healthy, node)
		}
	}

	return healthy
}

// ScheduleTask schedules a task on the most suitable node
func (c *Cluster) ScheduleTask(ctx context.Context, task string, requirements types.TaskRequirements) (string, error) {
	// Find suitable node based on requirements and load
	node := c.findSuitableNode(requirements)
	if node == nil {
		return "", fmt.Errorf("no suitable node found for task")
	}

	// Execute task on selected node
	result, err := node.ExecuteTask(ctx, requirements.AgentID, task)
	if err != nil {
		return "", fmt.Errorf("task execution failed: %w", err)
	}

	return result, nil
}

// findSuitableNode finds the most suitable node for a task
func (c *Cluster) findSuitableNode(requirements types.TaskRequirements) *Node {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var bestNode *Node
	var lowestLoad float64 = 1.0

	for _, node := range c.nodes {
		if node.status != StatusHealthy {
			continue
		}

		// Check if node meets requirements
		if !c.nodeMatchesRequirements(node, requirements) {
			continue
		}

		// Calculate load factor
		loadFactor := float64(node.load) / float64(node.capacity)
		if loadFactor < lowestLoad {
			lowestLoad = loadFactor
			bestNode = node
		}
	}

	return bestNode
}

// nodeMatchesRequirements checks if a node meets task requirements
func (c *Cluster) nodeMatchesRequirements(node *Node, requirements types.TaskRequirements) bool {
	// Check if node has required agent
	if requirements.AgentID != "" {
		_, exists := node.agents[requirements.AgentID]
		if !exists {
			return false
		}
	}

	// Check if node has required tags
	if len(requirements.Tags) > 0 {
		nodeTags := make(map[string]bool)
		for _, tag := range node.config.Tags {
			nodeTags[tag] = true
		}

		for _, requiredTag := range requirements.Tags {
			if !nodeTags[requiredTag] {
				return false
			}
		}
	}

	return true
}

// monitorHeartbeats monitors node health through heartbeats
func (c *Cluster) monitorHeartbeats() {
	ticker := time.NewTicker(c.config.HeartbeatInterval)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		for id, node := range c.nodes {
			if time.Since(node.lastPing) > c.config.NodeTimeout {
				node.status = StatusUnhealthy
				c.client.PublishEvent(context.Background(), "node_unhealthy",
					fmt.Sprintf("Node %s marked as unhealthy", id),
					map[string]string{"node_id": id})
			}
		}
		c.mu.Unlock()
	}
}

// Close closes the cluster and all its nodes
func (c *Cluster) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var errs []error
	for _, node := range c.nodes {
		if err := node.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if err := c.client.Close(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing cluster: %v", errs)
	}

	return nil
}
