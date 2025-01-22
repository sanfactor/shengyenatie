package communication

import (
	"context"
	"sync"
)

// Event represents a system event
type Event struct {
	Type     string
	Payload  interface{}
	Metadata map[string]interface{}
}

// EventHandler is a function that processes events
type EventHandler func(context.Context, Event) error

// EventSystem manages event distribution
type EventSystem struct {
	handlers map[string][]EventHandler
	mu       sync.RWMutex
}

// NewEventSystem creates a new event system
func NewEventSystem() *EventSystem {
	return &EventSystem{
		handlers: make(map[string][]EventHandler),
	}
}

// RegisterHandler adds an event handler for a specific event type
func (es *EventSystem) RegisterHandler(eventType string, handler EventHandler) {
	es.mu.Lock()
	defer es.mu.Unlock()
	es.handlers[eventType] = append(es.handlers[eventType], handler)
}

// EmitEvent broadcasts an event to all registered handlers
func (es *EventSystem) EmitEvent(ctx context.Context, event Event) error {
	es.mu.RLock()
	handlers := es.handlers[event.Type]
	es.mu.RUnlock()

	var wg sync.WaitGroup
	errCh := make(chan error, len(handlers))

	for _, handler := range handlers {
		wg.Add(1)
		go func(h EventHandler) {
			defer wg.Done()
			if err := h(ctx, event); err != nil {
				errCh <- err
			}
		}(handler)
	}

	wg.Wait()
	close(errCh)

	// Return first error if any occurred
	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}
