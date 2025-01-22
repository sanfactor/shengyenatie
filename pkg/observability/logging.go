package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// LogLevel represents the severity of a log entry
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp time.Time         `json:"timestamp"`
	Level     LogLevel         `json:"level"`
	Message   string          `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	TraceID   string          `json:"trace_id,omitempty"`
	SpanID    string          `json:"span_id,omitempty"`
}

// Logger manages structured logging
type Logger struct {
	output io.Writer
	mu     sync.Mutex
}

// NewLogger creates a new logger
func NewLogger(output io.Writer) *Logger {
	if output == nil {
		output = os.Stdout
	}
	return &Logger{output: output}
}

// Log writes a log entry
func (l *Logger) Log(ctx context.Context, level LogLevel, msg string, fields map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   msg,
		Fields:    fields,
	}

	// Add tracing context if available
	if span, ok := ctx.Value(spanKey{}).(*Span); ok {
		entry.TraceID = span.TraceID
		entry.SpanID = span.SpanID
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Marshal to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling log entry: %v\n", err)
		return
	}

	// Write to output
	l.output.Write(append(data, '\n'))
}

// Helper methods for different log levels
func (l *Logger) Debug(ctx context.Context, msg string, fields map[string]interface{}) {
	l.Log(ctx, DEBUG, msg, fields)
}

func (l *Logger) Info(ctx context.Context, msg string, fields map[string]interface{}) {
	l.Log(ctx, INFO, msg, fields)
}

func (l *Logger) Warn(ctx context.Context, msg string, fields map[string]interface{}) {
	l.Log(ctx, WARN, msg, fields)
}

func (l *Logger) Error(ctx context.Context, msg string, fields map[string]interface{}) {
	l.Log(ctx, ERROR, msg, fields)
}
