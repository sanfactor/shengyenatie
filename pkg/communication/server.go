package communication

import (
	"context"
	"fmt"
	"net"
	"sync"

	pb "github.com/user/modulox/pkg/pb"
	"google.golang.org/grpc"
)

// AgentServer implements the gRPC server for agent communication
type AgentServer struct {
	pb.UnimplementedAgentServiceServer
	messageBus *MessageBus
	eventSys  *EventSystem
	stateStore *StateStore
	mu        sync.RWMutex
}

// NewAgentServer creates a new agent server instance
func NewAgentServer() *AgentServer {
	return &AgentServer{
		messageBus: NewMessageBus(),
		eventSys:  NewEventSystem(),
		stateStore: NewStateStore(),
	}
}

// Execute implements AgentService.Execute
func (s *AgentServer) Execute(ctx context.Context, req *pb.ExecuteRequest) (*pb.ExecuteResponse, error) {
	// Forward task to appropriate agent and return response
	result, err := s.executeTask(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute task: %w", err)
	}

	return &pb.ExecuteResponse{
		Result:   result,
		Metadata: req.Metadata,
	}, nil
}

// StreamEvents implements AgentService.StreamEvents
func (s *AgentServer) StreamEvents(req *pb.EventRequest, stream pb.AgentService_StreamEventsServer) error {
	// Create event channel for this agent
	eventCh := s.messageBus.Subscribe(req.AgentId)
	defer close(eventCh)

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case msg := <-eventCh:
			event := &pb.Event{
				Type:        msg.Type,
				Payload:     msg.Content.(string),
				SourceAgent: msg.From,
				Timestamp:   msg.Timestamp.Unix(),
				Metadata:    msg.Metadata,
			}
			if err := stream.Send(event); err != nil {
				return err
			}
		}
	}
}

// PublishEvent implements AgentService.PublishEvent
func (s *AgentServer) PublishEvent(ctx context.Context, event *pb.Event) (*pb.PublishResponse, error) {
	msg := Message{
		Type:      event.Type,
		Content:   event.Payload,
		From:      event.SourceAgent,
		Metadata:  event.Metadata,
	}

	if err := s.messageBus.Publish(ctx, event.SourceAgent, msg); err != nil {
		return &pb.PublishResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.PublishResponse{Success: true}, nil
}

// SyncState implements AgentService.SyncState
func (s *AgentServer) SyncState(ctx context.Context, req *pb.SyncRequest) (*pb.SyncResponse, error) {
	s.stateStore.Set(req.Key, req.Value)
	
	entry, _ := s.stateStore.Get(req.Key)
	return &pb.SyncResponse{
		Success: true,
		Version: entry.Version,
	}, nil
}

// Start starts the gRPC server
func (s *AgentServer) Start(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	server := grpc.NewServer()
	pb.RegisterAgentServiceServer(server, s)
	
	return server.Serve(listener)
}

// Helper function to execute tasks
func (s *AgentServer) executeTask(ctx context.Context, req *pb.ExecuteRequest) (string, error) {
	// TODO: Implement task execution logic
	// This should integrate with the workflow system
	return fmt.Sprintf("Executed task for agent %s: %s", req.AgentId, req.Task), nil
}
