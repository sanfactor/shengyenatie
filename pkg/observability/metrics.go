package observability

import (
	"context"
	"sync"
	"time"
)

// MetricType represents the type of metric
type MetricType int

const (
	Counter MetricType = iota
	Gauge
	Histogram
)

// Metric represents a single metric
type Metric struct {
	Name        string
	Type        MetricType
	Value       float64
	Labels      map[string]string
	Timestamp   time.Time
}

// MetricsCollector manages metric collection
type MetricsCollector struct {
	metrics map[string][]Metric
	mu      sync.RWMutex
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics: make(map[string][]Metric),
	}
}

// RecordMetric records a new metric
func (mc *MetricsCollector) RecordMetric(ctx context.Context, metric Metric) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	metric.Timestamp = time.Now()
	mc.metrics[metric.Name] = append(mc.metrics[metric.Name], metric)
}

// GetMetrics returns metrics for a given name
func (mc *MetricsCollector) GetMetrics(name string) []Metric {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return mc.metrics[name]
}

// Counter represents a cumulative metric
type Counter struct {
	name   string
	value  float64
	labels map[string]string
	mc     *MetricsCollector
	mu     sync.Mutex
}

// NewCounter creates a new counter metric
func (mc *MetricsCollector) NewCounter(name string, labels map[string]string) *Counter {
	return &Counter{
		name:   name,
		labels: labels,
		mc:     mc,
	}
}

// Inc increments the counter by 1
func (c *Counter) Inc() {
	c.Add(1)
}

// Add adds the given value to the counter
func (c *Counter) Add(value float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.value += value
	c.mc.RecordMetric(context.Background(), Metric{
		Name:      c.name,
		Type:      Counter,
		Value:     c.value,
		Labels:    c.labels,
		Timestamp: time.Now(),
	})
}
