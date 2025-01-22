package communication

import (
	"context"
	"sync"
	"time"
)

// Message represents a communication unit between agents
type Message struct {
	ID        string
	From      string
	To        string
	Content   interface{}
	Timestamp time.Time
	Type      string
	Metadata  map[string]interface{}
}

// MessageBus handles message routing between agents
type MessageBus struct {
	subscribers map[string][]chan Message
	mu          sync.RWMutex
}

// NewMessageBus creates a new message bus instance
func NewMessageBus() *MessageBus {
	return &MessageBus{
		subscribers: make(map[string][]chan Message),
	}
}

// Subscribe registers a subscriber for a specific topic
func (mb *MessageBus) Subscribe(topic string) chan Message {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	ch := make(chan Message, 100)
	mb.subscribers[topic] = append(mb.subscribers[topic], ch)
	return ch
}

// Publish sends a message to all subscribers of a topic
func (mb *MessageBus) Publish(ctx context.Context, topic string, msg Message) error {
	mb.mu.RLock()
	subscribers := mb.subscribers[topic]
	mb.mu.RUnlock()

	for _, ch := range subscribers {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ch <- msg:
		default:
			// Non-blocking send to prevent slow subscribers from blocking publishers
			continue
		}
	}
	return nil
}
