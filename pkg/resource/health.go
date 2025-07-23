// pkg/resource/health.go
package resource

import (
	"context"
	"fmt"
)

// ResourceHealthCheck provides health check functionality for the resource manager.
type ResourceHealthCheck struct {
	manager *ResourceManager
}

// NewResourceHealthCheck creates a new health check for the resource manager.
func NewResourceHealthCheck(manager *ResourceManager) *ResourceHealthCheck {
	return &ResourceHealthCheck{
		manager: manager,
	}
}

// Name returns the name of this health check.
func (r *ResourceHealthCheck) Name() string {
	return "resource"
}

// Check verifies that resource usage is within acceptable limits.
func (r *ResourceHealthCheck) Check(ctx context.Context) error {
	stats := r.manager.GetResourceStats()

	// Check memory usage
	if stats.MemoryUsageMB > stats.MaxMemoryMB {
		return fmt.Errorf("memory usage %dMB exceeds limit %dMB",
			stats.MemoryUsageMB, stats.MaxMemoryMB)
	}

	// Check goroutine usage (warn at 80% of limit)
	goroutineThreshold := int64(float64(stats.MaxGoroutines) * 0.8)
	if stats.GoroutineCount > goroutineThreshold {
		return fmt.Errorf("goroutine count %d exceeds 80%% threshold (%d/%d)",
			stats.GoroutineCount, goroutineThreshold, stats.MaxGoroutines)
	}

	return nil
}
