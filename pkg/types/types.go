package types

// Capability represents a specific ability that an agent can perform
type Capability struct {
	Name        string
	Description string
	Parameters  map[string]interface{}
}

// Tool represents a function that an agent can use
type Tool interface {
	// Execute runs the tool with given input
	Execute(input interface{}) (interface{}, error)
	// GetDescription returns information about the tool
	GetDescription() string
}

// Vector represents an embedding vector
type Vector struct {
	ID       string
	Values   []float32
	Metadata map[string]interface{}
}

// NodeStatus represents the status of a distributed node
type NodeStatus struct {
	ID         string
	Address    string
	Load       int
	Capacity   int
	Status     int
	LastPing   time.Time
	AgentCount int
}

// TaskRequirements specifies requirements for task execution
type TaskRequirements struct {
	AgentID string
	Tags    []string
	MinCPU  float64
	MinMem  int64
}

// WorkflowResult represents the result of a workflow execution
type WorkflowResult struct {
	AgentID     string
	Output      string
	Error       error
	Metadata    map[string]interface{}
}
