// pkg/resource/manager.go
package resource

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/opd-ai/go-netrek/pkg/config"
	"github.com/opd-ai/go-netrek/pkg/logging"
)

// ResourceManager manages system resources like memory and goroutines
// to prevent resource exhaustion and enable graceful shutdown.
type ResourceManager struct {
	maxMemoryMB     int64
	maxGoroutines   int64
	shutdownTimeout time.Duration
	checkInterval   time.Duration

	// Atomic counters for thread-safe access
	goroutineCount int64
	memoryUsageMB  int64

	// Control channels and state
	ctx     context.Context
	cancel  context.CancelFunc
	done    chan struct{}
	mu      sync.RWMutex
	running bool
	logger  *logging.Logger

	// Metrics for monitoring
	lastMemoryCheck    time.Time
	lastGoroutineCheck time.Time
}

// NewResourceManager creates a new resource manager with the given configuration.
func NewResourceManager(config *config.EnvironmentConfig) *ResourceManager {
	ctx, cancel := context.WithCancel(context.Background())

	rm := &ResourceManager{
		maxMemoryMB:        config.MaxMemoryMB,
		maxGoroutines:      int64(config.MaxGoroutines),
		shutdownTimeout:    config.ShutdownTimeout,
		checkInterval:      config.ResourceCheckInterval,
		ctx:                ctx,
		cancel:             cancel,
		done:               make(chan struct{}),
		logger:             logging.NewLogger(),
		lastMemoryCheck:    time.Now(),
		lastGoroutineCheck: time.Now(),
	}

	return rm
}

// Start begins the resource monitoring loop.
func (rm *ResourceManager) Start() error {
	rm.mu.Lock()
	if rm.running {
		rm.mu.Unlock()
		return fmt.Errorf("resource manager already running")
	}
	rm.running = true
	rm.mu.Unlock()

	// Start the monitoring goroutine
	go rm.monitoringLoop()

	rm.logger.Info(rm.ctx, "Resource manager started",
		"max_memory_mb", rm.maxMemoryMB,
		"max_goroutines", rm.maxGoroutines,
		"check_interval", rm.checkInterval,
	)

	return nil
}

// StartGoroutine safely starts a new goroutine with resource tracking.
// It returns an error if the goroutine limit would be exceeded.
func (rm *ResourceManager) StartGoroutine(ctx context.Context, name string, fn func(context.Context)) error {
	// Check goroutine limit before starting
	current := atomic.LoadInt64(&rm.goroutineCount)
	if current >= rm.maxGoroutines {
		rm.logger.Warn(ctx, "Goroutine limit exceeded",
			"current", current,
			"limit", rm.maxGoroutines,
			"name", name,
		)
		return fmt.Errorf("goroutine limit exceeded: %d/%d", current, rm.maxGoroutines)
	}

	// Increment counter before starting
	atomic.AddInt64(&rm.goroutineCount, 1)

	go func() {
		// Ensure we decrement on exit
		defer atomic.AddInt64(&rm.goroutineCount, -1)

		// Panic recovery
		defer func() {
			if r := recover(); r != nil {
				rm.logger.Error(ctx, "Goroutine panic",
					fmt.Errorf("panic: %v", r),
					"name", name,
				)
			}
		}()

		// Execute the function
		fn(ctx)
	}()

	return nil
}

// CheckMemoryUsage checks current memory usage against limits.
func (rm *ResourceManager) CheckMemoryUsage() error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	currentMB := int64(m.Alloc / 1024 / 1024)
	atomic.StoreInt64(&rm.memoryUsageMB, currentMB)
	rm.lastMemoryCheck = time.Now()

	if currentMB > rm.maxMemoryMB {
		return fmt.Errorf("memory usage %dMB exceeds limit %dMB", currentMB, rm.maxMemoryMB)
	}

	return nil
}

// GetGoroutineCount returns the current number of tracked goroutines.
func (rm *ResourceManager) GetGoroutineCount() int64 {
	return atomic.LoadInt64(&rm.goroutineCount)
}

// GetMemoryUsage returns the current memory usage in MB.
func (rm *ResourceManager) GetMemoryUsage() int64 {
	return atomic.LoadInt64(&rm.memoryUsageMB)
}

// GetResourceStats returns current resource usage statistics.
func (rm *ResourceManager) GetResourceStats() ResourceStats {
	return ResourceStats{
		GoroutineCount:     rm.GetGoroutineCount(),
		MaxGoroutines:      rm.maxGoroutines,
		MemoryUsageMB:      rm.GetMemoryUsage(),
		MaxMemoryMB:        rm.maxMemoryMB,
		LastMemoryCheck:    rm.lastMemoryCheck,
		LastGoroutineCheck: rm.lastGoroutineCheck,
	}
}

// ResourceStats contains resource usage statistics.
type ResourceStats struct {
	GoroutineCount     int64     `json:"goroutine_count"`
	MaxGoroutines      int64     `json:"max_goroutines"`
	MemoryUsageMB      int64     `json:"memory_usage_mb"`
	MaxMemoryMB        int64     `json:"max_memory_mb"`
	LastMemoryCheck    time.Time `json:"last_memory_check"`
	LastGoroutineCheck time.Time `json:"last_goroutine_check"`
}

// Shutdown gracefully stops the resource manager and waits for all goroutines to finish.
func (rm *ResourceManager) Shutdown(ctx context.Context) error {
	rm.mu.Lock()
	if !rm.running {
		rm.mu.Unlock()
		return nil // Already shut down
	}
	rm.running = false
	rm.mu.Unlock()

	rm.logger.Info(ctx, "Shutting down resource manager")

	// Signal shutdown
	rm.cancel()

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, rm.shutdownTimeout)
	defer cancel()

	// Wait for monitoring loop to stop
	select {
	case <-rm.done:
		// Monitoring loop stopped
	case <-shutdownCtx.Done():
		rm.logger.Warn(ctx, "Resource manager monitoring loop did not stop gracefully")
	}

	// Wait for all tracked goroutines to finish
	return rm.waitForGoroutines(shutdownCtx)
}

// waitForGoroutines waits for all tracked goroutines to finish or timeout.
func (rm *ResourceManager) waitForGoroutines(ctx context.Context) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		count := rm.GetGoroutineCount()
		if count == 0 {
			rm.logger.Info(ctx, "All tracked goroutines finished")
			return nil
		}

		select {
		case <-ticker.C:
			// Continue waiting
			rm.logger.Debug(ctx, "Waiting for goroutines to finish",
				"remaining", count,
			)
		case <-ctx.Done():
			remaining := rm.GetGoroutineCount()
			rm.logger.Warn(ctx, "Shutdown timeout exceeded with goroutines still running",
				"remaining", remaining,
			)
			return fmt.Errorf("shutdown timeout: %d goroutines still running", remaining)
		}
	}
}

// monitoringLoop runs periodic resource checks.
func (rm *ResourceManager) monitoringLoop() {
	defer close(rm.done)

	ticker := time.NewTicker(rm.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rm.performResourceChecks()
		case <-rm.ctx.Done():
			rm.logger.Info(rm.ctx, "Resource monitoring loop stopping")
			return
		}
	}
}

// performResourceChecks executes periodic resource usage checks.
func (rm *ResourceManager) performResourceChecks() {
	// Check memory usage
	if err := rm.CheckMemoryUsage(); err != nil {
		rm.logger.Error(rm.ctx, "Memory limit exceeded", err,
			"current_mb", rm.GetMemoryUsage(),
			"limit_mb", rm.maxMemoryMB,
		)
	}

	// Update goroutine check timestamp
	rm.lastGoroutineCheck = time.Now()

	// Log current resource usage (debug level)
	rm.logger.Debug(rm.ctx, "Resource usage check",
		"goroutines", rm.GetGoroutineCount(),
		"max_goroutines", rm.maxGoroutines,
		"memory_mb", rm.GetMemoryUsage(),
		"max_memory_mb", rm.maxMemoryMB,
	)
}
