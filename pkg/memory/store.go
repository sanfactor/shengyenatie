package memory

import (
	"context"

	"github.com/user/go-ai-framework/pkg/types"
)

// VectorStore defines the interface for vector storage and retrieval
type VectorStore interface {
	// Store saves vectors to the store
	Store(ctx context.Context, vectors []types.Vector) error
	
	// Query retrieves the k nearest vectors to the query vector
	Query(ctx context.Context, vector types.Vector, k int) ([]types.Vector, error)
}

// BaseStore provides a basic implementation of VectorStore
type BaseStore struct {
	vectors []types.Vector
}

// NewBaseStore creates a new instance of BaseStore
func NewBaseStore() *BaseStore {
	return &BaseStore{
		vectors: make([]types.Vector, 0),
	}
}

// Store implements VectorStore.Store
func (b *BaseStore) Store(ctx context.Context, vectors []types.Vector) error {
	b.vectors = append(b.vectors, vectors...)
	return nil
}

// Query implements VectorStore.Query
// Note: This is a naive implementation. Real implementations should use proper vector similarity search
func (b *BaseStore) Query(ctx context.Context, vector types.Vector, k int) ([]types.Vector, error) {
	// In a real implementation, this would perform proper vector similarity search
	if k > len(b.vectors) {
		k = len(b.vectors)
	}
	return b.vectors[:k], nil
}
