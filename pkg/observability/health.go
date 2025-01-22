package observability

import (
	"context"
	"sync"
	"time"
)

// HealthStatus represents the health status of a component
type HealthStatus struct {
	Status    string
	Message   string
	Timestamp time.Time
	Details   map[string]interface{}
}

// HealthChecker manages health checks
type HealthChecker struct {
	checks   map[string]HealthCheck
	statuses map[string]HealthStatus
	mu       sync.RWMutex
}

// HealthCheck is a function that performs a health check
type HealthCheck func(context.Context) HealthStatus

// NewHealthChecker creates a new health checker
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checks:   make(map[string]HealthCheck),
		statuses: make(map[string]HealthStatus),
	}
}

// RegisterCheck registers a new health check
func (hc *HealthChecker) RegisterCheck(name string, check HealthCheck) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.checks[name] = check
}

// RunChecks runs all registered health checks
func (hc *HealthChecker) RunChecks(ctx context.Context) map[string]HealthStatus {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	for name, check := range hc.checks {
		hc.statuses[name] = check(ctx)
	}

	return hc.statuses
}

// GetStatus returns the current health status
func (hc *HealthChecker) GetStatus(name string) (HealthStatus, bool) {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	status, exists := hc.statuses[name]
	return status, exists
}

// IsHealthy returns true if all checks are healthy
func (hc *HealthChecker) IsHealthy() bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	for _, status := range hc.statuses {
		if status.Status != "healthy" {
			return false
		}
	}

	return true
}

// StartMonitoring starts periodic health checks
func (hc *HealthChecker) StartMonitoring(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			hc.RunChecks(ctx)
		}
	}
}
