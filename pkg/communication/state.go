package communication

import (
	"context"
	"sync"
	"time"
)

// StateEntry represents a single state value with metadata
type StateEntry struct {
	Value     interface{}
	Version   int64
	UpdatedAt time.Time
}

// StateStore manages distributed state
type StateStore struct {
	states map[string]StateEntry
	mu     sync.RWMutex
}

// NewStateStore creates a new state store
func NewStateStore() *StateStore {
	return &StateStore{
		states: make(map[string]StateEntry),
	}
}

// Set updates a state value
func (ss *StateStore) Set(key string, value interface{}) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	currentEntry, exists := ss.states[key]
	var version int64 = 1
	if exists {
		version = currentEntry.Version + 1
	}

	ss.states[key] = StateEntry{
		Value:     value,
		Version:   version,
		UpdatedAt: time.Now(),
	}
}

// Get retrieves a state value
func (ss *StateStore) Get(key string) (StateEntry, bool) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	entry, exists := ss.states[key]
	return entry, exists
}

// Watch monitors a key for changes
func (ss *StateStore) Watch(ctx context.Context, key string) (<-chan StateEntry, error) {
	updates := make(chan StateEntry, 1)
	
	// Initial state
	if entry, exists := ss.Get(key); exists {
		updates <- entry
	}

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		defer close(updates)

		var lastVersion int64 = -1

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if entry, exists := ss.Get(key); exists && entry.Version > lastVersion {
					lastVersion = entry.Version
					updates <- entry
				}
			}
		}
	}()

	return updates, nil
}
