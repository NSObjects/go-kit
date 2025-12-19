// Package health provides component health checking.
package health

import (
	"context"
	"sync"
	"time"
)

// Status represents the health status of a component.
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
)

// Check represents a single health check result.
type Check struct {
	Name    string        `json:"name"`
	Status  Status        `json:"status"`
	Message string        `json:"message,omitempty"`
	Latency time.Duration `json:"latency"`
}

// Checker is the interface for health checks.
type Checker interface {
	Name() string
	Check(ctx context.Context) Check
}

// Registry holds all registered health checkers.
type Registry struct {
	mu       sync.RWMutex
	checkers []Checker
}

// NewRegistry creates a new health check registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// Register adds a health checker to the registry.
func (r *Registry) Register(checker Checker) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checkers = append(r.checkers, checker)
}

// CheckAll runs all health checks and returns results.
func (r *Registry) CheckAll(ctx context.Context) []Check {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make([]Check, 0, len(r.checkers))
	for _, checker := range r.checkers {
		results = append(results, checker.Check(ctx))
	}
	return results
}

// OverallStatus returns the overall health status.
func (r *Registry) OverallStatus(ctx context.Context) Status {
	checks := r.CheckAll(ctx)
	for _, check := range checks {
		if check.Status == StatusUnhealthy {
			return StatusUnhealthy
		}
		if check.Status == StatusDegraded {
			return StatusDegraded
		}
	}
	return StatusHealthy
}
