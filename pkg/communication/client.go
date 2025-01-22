package communication

import (
	"context"
	"fmt"
	"time"

	pb "github.com/user/modulox/pkg/pb"
	"google.golang.org/grpc"
)

// AgentClient provides a high-level client for agent communication
type AgentClient struct {
	conn   *grpc.ClientConn
	client pb.AgentServiceClient
	agentID string
}

// NewAgentClient creates a new agent client
func NewAgentClient(address, agentID string) (*AgentClient, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return &AgentClient{
		conn:    conn,
		client:  pb.NewAgentServiceClient(conn),
		agentID: agentID,
	}, nil
}

// Close closes the client connection
func (c *AgentClient) Close() error {
	return c.conn.Close()
}

// ExecuteTask sends a task execution request
func (c *AgentClient) ExecuteTask(ctx context.Context, task string, metadata map[string]string) (string, error) {
	req := &pb.ExecuteRequest{
		AgentId:  c.agentID,
		Task:     task,
		Metadata: metadata,
	}

	resp, err := c.client.Execute(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to execute task: %w", err)
	}

	return resp.Result, nil
}

// StreamEvents subscribes to agent events
func (c *AgentClient) StreamEvents(ctx context.Context, eventTypes []string) (<-chan *pb.Event, error) {
	req := &pb.EventRequest{
		AgentId:    c.agentID,
		EventTypes: eventTypes,
	}

	stream, err := c.client.StreamEvents(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to stream events: %w", err)
	}

	events := make(chan *pb.Event, 100)
	go func() {
		defer close(events)
		for {
			event, err := stream.Recv()
			if err != nil {
				return
			}
			select {
			case events <- event:
			case <-ctx.Done():
				return
			}
		}
	}()

	return events, nil
}

// PublishEvent publishes an event
func (c *AgentClient) PublishEvent(ctx context.Context, eventType, payload string, metadata map[string]string) error {
	event := &pb.Event{
		Type:        eventType,
		Payload:     payload,
		SourceAgent: c.agentID,
		Timestamp:   time.Now().Unix(),
		Metadata:    metadata,
	}

	resp, err := c.client.PublishEvent(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("failed to publish event: %s", resp.Error)
	}

	return nil
}

// SyncState synchronizes state with the server
func (c *AgentClient) SyncState(ctx context.Context, key, value string) (int64, error) {
	req := &pb.SyncRequest{
		AgentId: c.agentID,
		Key:     key,
		Value:   value,
	}

	resp, err := c.client.SyncState(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("failed to sync state: %w", err)
	}

	if !resp.Success {
		return 0, fmt.Errorf("failed to sync state: %s", resp.Error)
	}

	return resp.Version, nil
}
