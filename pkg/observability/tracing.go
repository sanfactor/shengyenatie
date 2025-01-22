package observability

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Span represents a single operation within a trace
type Span struct {
	TraceID     string
	SpanID      string
	ParentID    string
	Name        string
	StartTime   time.Time
	EndTime     time.Time
	Tags        map[string]string
	Events      []SpanEvent
	Status      SpanStatus
}

// SpanEvent represents an event within a span
type SpanEvent struct {
	Time    time.Time
	Name    string
	Message string
	Tags    map[string]string
}

// SpanStatus represents the status of a span
type SpanStatus int

const (
	StatusOK SpanStatus = iota
	StatusError
)

// Tracer manages distributed tracing
type Tracer struct {
	spans    map[string]*Span
	mu       sync.RWMutex
	sampler  Sampler
}

// Sampler determines if a trace should be sampled
type Sampler interface {
	ShouldSample(traceID string) bool
}

// NewTracer creates a new tracer
func NewTracer(sampler Sampler) *Tracer {
	return &Tracer{
		spans:   make(map[string]*Span),
		sampler: sampler,
	}
}

// StartSpan starts a new span
func (t *Tracer) StartSpan(ctx context.Context, name string, opts ...SpanOption) (*Span, context.Context) {
	span := &Span{
		TraceID:   generateTraceID(),
		SpanID:    generateSpanID(),
		Name:      name,
		StartTime: time.Now(),
		Tags:      make(map[string]string),
		Status:    StatusOK,
	}

	// Apply options
	for _, opt := range opts {
		opt(span)
	}

	// Check if we should sample this trace
	if !t.sampler.ShouldSample(span.TraceID) {
		return nil, ctx
	}

	t.mu.Lock()
	t.spans[span.SpanID] = span
	t.mu.Unlock()

	return span, context.WithValue(ctx, spanKey{}, span)
}

// EndSpan ends a span
func (t *Tracer) EndSpan(span *Span) {
	if span == nil {
		return
	}

	span.EndTime = time.Now()
}

// AddEvent adds an event to a span
func (t *Tracer) AddEvent(span *Span, name string, message string, tags map[string]string) {
	if span == nil {
		return
	}

	event := SpanEvent{
		Time:    time.Now(),
		Name:    name,
		Message: message,
		Tags:    tags,
	}

	span.Events = append(span.Events, event)
}

// SetError sets the span status to error
func (t *Tracer) SetError(span *Span, err error) {
	if span == nil {
		return
	}

	span.Status = StatusError
	span.Tags["error"] = err.Error()
}

// SpanOption configures a span
type SpanOption func(*Span)

// WithParent sets the parent span
func WithParent(parent *Span) SpanOption {
	return func(s *Span) {
		if parent != nil {
			s.ParentID = parent.SpanID
			s.TraceID = parent.TraceID
		}
	}
}

// WithTags adds tags to the span
func WithTags(tags map[string]string) SpanOption {
	return func(s *Span) {
		for k, v := range tags {
			s.Tags[k] = v
		}
	}
}

type spanKey struct{}

// Helper functions for generating IDs
func generateTraceID() string {
	return fmt.Sprintf("trace-%d", time.Now().UnixNano())
}

func generateSpanID() string {
	return fmt.Sprintf("span-%d", time.Now().UnixNano())
}
