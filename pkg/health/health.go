// Package health provides health check functionality for the go-netrek server.
// It implements HTTP endpoints for liveness and readiness probes that are
// essential for production deployment and monitoring.
package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// HealthCheck defines the interface for individual health checks.
// Each component can implement this interface to provide its health status.
type HealthCheck interface {
	// Name returns the unique name of this health check
	Name() string
	// Check performs the health check and returns an error if unhealthy
	Check(ctx context.Context) error
}

// HealthStatus represents the overall health status of the application.
type HealthStatus struct {
	Status string                     `json:"status"`
	Checks map[string]ComponentHealth `json:"checks"`
}

// ComponentHealth represents the health status of an individual component.
type ComponentHealth struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// HealthChecker manages and executes health checks for the application.
type HealthChecker struct {
	checks map[string]HealthCheck
	mu     sync.RWMutex
}

// NewHealthChecker creates a new health checker instance.
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checks: make(map[string]HealthCheck),
	}
}

// AddCheck registers a new health check with the health checker.
// If a check with the same name already exists, it will be replaced.
func (hc *HealthChecker) AddCheck(check HealthCheck) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.checks[check.Name()] = check
}

// RemoveCheck removes a health check by name.
func (hc *HealthChecker) RemoveCheck(name string) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	delete(hc.checks, name)
}

// CheckHealth executes all registered health checks and returns the aggregated status.
// The overall status is "healthy" only if all individual checks pass.
func (hc *HealthChecker) CheckHealth(ctx context.Context) HealthStatus {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	status := HealthStatus{
		Status: "healthy",
		Checks: make(map[string]ComponentHealth),
	}

	// Execute all health checks
	for name, check := range hc.checks {
		if err := check.Check(ctx); err != nil {
			status.Status = "unhealthy"
			status.Checks[name] = ComponentHealth{
				Status:  "unhealthy",
				Message: err.Error(),
			}
		} else {
			status.Checks[name] = ComponentHealth{
				Status: "healthy",
			}
		}
	}

	return status
}

// LivenessHandler provides a simple liveness probe endpoint.
// This endpoint returns 200 OK if the application is running and able to handle requests.
// It's used by orchestrators to determine if the application should be restarted.
func (hc *HealthChecker) LivenessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]string{"status": "alive"}
	json.NewEncoder(w).Encode(response)
}

// ReadinessHandler provides a readiness probe endpoint that executes all health checks.
// This endpoint returns 200 OK if the application is ready to serve traffic,
// or 503 Service Unavailable if any health check fails.
// It's used by load balancers to determine if traffic should be routed to this instance.
func (hc *HealthChecker) ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	// Create context with timeout for health checks
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	health := hc.CheckHealth(ctx)

	w.Header().Set("Content-Type", "application/json")

	if health.Status == "healthy" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(health)
}

// GameEngineHealthCheck implements HealthCheck for the game engine.
type GameEngineHealthCheck struct {
	gameRunning func() bool
}

// NewGameEngineHealthCheck creates a health check for the game engine.
func NewGameEngineHealthCheck(gameRunning func() bool) *GameEngineHealthCheck {
	return &GameEngineHealthCheck{
		gameRunning: gameRunning,
	}
}

// Name returns the name of this health check.
func (g *GameEngineHealthCheck) Name() string {
	return "game_engine"
}

// Check verifies that the game engine is running properly.
func (g *GameEngineHealthCheck) Check(ctx context.Context) error {
	if !g.gameRunning() {
		return fmt.Errorf("game engine is not running")
	}
	return nil
}

// NetworkHealthCheck implements HealthCheck for network connectivity.
type NetworkHealthCheck struct {
	listenerAddr func() string
}

// NewNetworkHealthCheck creates a health check for network connectivity.
func NewNetworkHealthCheck(listenerAddr func() string) *NetworkHealthCheck {
	return &NetworkHealthCheck{
		listenerAddr: listenerAddr,
	}
}

// Name returns the name of this health check.
func (n *NetworkHealthCheck) Name() string {
	return "network"
}

// Check verifies that the network listener is active.
func (n *NetworkHealthCheck) Check(ctx context.Context) error {
	addr := n.listenerAddr()
	if addr == "" {
		return fmt.Errorf("network listener is not active")
	}
	return nil
}

// MemoryHealthCheck implements HealthCheck for memory usage monitoring.
type MemoryHealthCheck struct {
	maxMemoryMB    int64
	getMemoryUsage func() int64
}

// NewMemoryHealthCheck creates a health check for memory usage.
func NewMemoryHealthCheck(maxMemoryMB int64, getMemoryUsage func() int64) *MemoryHealthCheck {
	return &MemoryHealthCheck{
		maxMemoryMB:    maxMemoryMB,
		getMemoryUsage: getMemoryUsage,
	}
}

// Name returns the name of this health check.
func (m *MemoryHealthCheck) Name() string {
	return "memory"
}

// Check verifies that memory usage is within acceptable limits.
func (m *MemoryHealthCheck) Check(ctx context.Context) error {
	currentMB := m.getMemoryUsage()
	if currentMB > m.maxMemoryMB {
		return fmt.Errorf("memory usage %dMB exceeds limit %dMB", currentMB, m.maxMemoryMB)
	}
	return nil
}
