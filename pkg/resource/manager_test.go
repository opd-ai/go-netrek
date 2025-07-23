// pkg/resource/manager_test.go
package resource

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/opd-ai/go-netrek/pkg/config"
)

func TestNewResourceManager(t *testing.T) {
	config := &config.EnvironmentConfig{
		MaxMemoryMB:           500,
		MaxGoroutines:         100,
		ShutdownTimeout:       30 * time.Second,
		ResourceCheckInterval: 10 * time.Second,
	}

	rm := NewResourceManager(config)

	if rm.maxMemoryMB != 500 {
		t.Errorf("Expected MaxMemoryMB 500, got %d", rm.maxMemoryMB)
	}
	if rm.maxGoroutines != 100 {
		t.Errorf("Expected MaxGoroutines 100, got %d", rm.maxGoroutines)
	}
	if rm.shutdownTimeout != 30*time.Second {
		t.Errorf("Expected ShutdownTimeout 30s, got %v", rm.shutdownTimeout)
	}
	if rm.checkInterval != 10*time.Second {
		t.Errorf("Expected CheckInterval 10s, got %v", rm.checkInterval)
	}

	// Clean shutdown
	rm.Shutdown(context.Background())
}

func TestResourceManager_StartGoroutine(t *testing.T) {
	config := &config.EnvironmentConfig{
		MaxMemoryMB:           500,
		MaxGoroutines:         3, // Small limit for testing
		ShutdownTimeout:       5 * time.Second,
		ResourceCheckInterval: 1 * time.Second,
	}

	rm := NewResourceManager(config)
	defer rm.Shutdown(context.Background())

	ctx := context.Background()
	var wg sync.WaitGroup

	// Test normal goroutine creation
	for i := 0; i < 3; i++ {
		wg.Add(1)
		err := rm.StartGoroutine(ctx, "test-goroutine", func(ctx context.Context) {
			defer wg.Done()
			time.Sleep(100 * time.Millisecond) // Short work
		})

		if err != nil {
			t.Errorf("Expected no error for goroutine %d, got: %v", i, err)
		}
	}

	// Test goroutine limit exceeded
	err := rm.StartGoroutine(ctx, "exceeding-goroutine", func(ctx context.Context) {
		time.Sleep(100 * time.Millisecond)
	})

	if err == nil {
		t.Error("Expected error when exceeding goroutine limit")
	}

	// Wait for goroutines to finish
	wg.Wait()

	// Wait a bit for counters to update
	time.Sleep(50 * time.Millisecond)

	// Verify count is back to 0
	if count := rm.GetGoroutineCount(); count != 0 {
		t.Errorf("Expected goroutine count 0, got %d", count)
	}
}

func TestResourceManager_StartGoroutinePanicRecovery(t *testing.T) {
	config := &config.EnvironmentConfig{
		MaxMemoryMB:           500,
		MaxGoroutines:         10,
		ShutdownTimeout:       5 * time.Second,
		ResourceCheckInterval: 1 * time.Second,
	}

	rm := NewResourceManager(config)
	defer rm.Shutdown(context.Background())

	ctx := context.Background()
	done := make(chan bool, 1)

	// Start a goroutine that panics
	err := rm.StartGoroutine(ctx, "panicking-goroutine", func(ctx context.Context) {
		defer func() { done <- true }()
		panic("test panic")
	})

	if err != nil {
		t.Errorf("Expected no error starting goroutine, got: %v", err)
	}

	// Wait for panic recovery
	select {
	case <-done:
		// Good, goroutine finished
	case <-time.After(1 * time.Second):
		t.Error("Goroutine did not finish within timeout")
	}

	// Wait for counter to update
	time.Sleep(50 * time.Millisecond)

	// Verify count is back to 0 (panic was recovered)
	if count := rm.GetGoroutineCount(); count != 0 {
		t.Errorf("Expected goroutine count 0 after panic recovery, got %d", count)
	}
}

func TestResourceManager_CheckMemoryUsage(t *testing.T) {
	// First test with reasonable limit
	config := &config.EnvironmentConfig{
		MaxMemoryMB:           1000, // Reasonable limit
		MaxGoroutines:         10,
		ShutdownTimeout:       5 * time.Second,
		ResourceCheckInterval: 1 * time.Second,
	}

	rm := NewResourceManager(config)
	defer rm.Shutdown(context.Background())

	// Allocate some memory to ensure memory usage is recorded
	data := make([]byte, 1024*1024) // 1MB
	_ = data

	// Memory check should pass with reasonable limit
	err := rm.CheckMemoryUsage()
	if err != nil {
		t.Errorf("Expected memory check to pass with reasonable limit, got: %v", err)
	}

	// Verify memory usage was recorded (should be > 0)
	usage := rm.GetMemoryUsage()
	if usage <= 0 {
		t.Errorf("Expected memory usage to be > 0, got %d MB", usage)
	}

	// Test failure case by directly checking against current usage
	if usage > 0 {
		// Set a limit below current usage
		rmLow := &ResourceManager{
			maxMemoryMB: usage - 1, // Set limit below current usage
		}

		err = rmLow.CheckMemoryUsage()
		if err == nil {
			t.Error("Expected memory check to fail with limit below current usage")
		}
	}
}

func TestResourceManager_GetResourceStats(t *testing.T) {
	config := &config.EnvironmentConfig{
		MaxMemoryMB:           500,
		MaxGoroutines:         10,
		ShutdownTimeout:       5 * time.Second,
		ResourceCheckInterval: 1 * time.Second,
	}

	rm := NewResourceManager(config)
	defer rm.Shutdown(context.Background())

	// Allocate some memory to ensure memory usage is recorded
	data := make([]byte, 1024*1024) // 1MB
	_ = data

	// Check memory to populate stats
	rm.CheckMemoryUsage()

	stats := rm.GetResourceStats()

	if stats.MaxMemoryMB != 500 {
		t.Errorf("Expected MaxMemoryMB 500, got %d", stats.MaxMemoryMB)
	}
	if stats.MaxGoroutines != 10 {
		t.Errorf("Expected MaxGoroutines 10, got %d", stats.MaxGoroutines)
	}
	if stats.MemoryUsageMB == 0 {
		t.Error("Expected memory usage to be recorded in stats")
	}
	if stats.LastMemoryCheck.IsZero() {
		t.Error("Expected LastMemoryCheck to be set")
	}
}

func TestResourceManager_StartAndShutdown(t *testing.T) {
	config := &config.EnvironmentConfig{
		MaxMemoryMB:           500,
		MaxGoroutines:         10,
		ShutdownTimeout:       5 * time.Second,
		ResourceCheckInterval: 100 * time.Millisecond, // Fast for testing
	}

	rm := NewResourceManager(config)

	// Test start
	err := rm.Start()
	if err != nil {
		t.Errorf("Expected no error starting resource manager, got: %v", err)
	}

	// Test double start (should error)
	err = rm.Start()
	if err == nil {
		t.Error("Expected error when starting already running resource manager")
	}

	// Wait a bit for monitoring loop to run
	time.Sleep(200 * time.Millisecond)

	// Test shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = rm.Shutdown(ctx)
	if err != nil {
		t.Errorf("Expected no error during shutdown, got: %v", err)
	}

	// Test double shutdown (should be safe)
	err = rm.Shutdown(ctx)
	if err != nil {
		t.Errorf("Expected no error during second shutdown, got: %v", err)
	}
}

func TestResourceManager_ShutdownTimeout(t *testing.T) {
	config := &config.EnvironmentConfig{
		MaxMemoryMB:           500,
		MaxGoroutines:         10,
		ShutdownTimeout:       200 * time.Millisecond, // Increased timeout for better testing
		ResourceCheckInterval: 1 * time.Second,
	}

	rm := NewResourceManager(config)

	// Start the resource manager monitoring loop
	err := rm.Start()
	if err != nil {
		t.Fatalf("Failed to start resource manager: %v", err)
	}

	ctx := context.Background()
	stopChan := make(chan struct{})

	// Start a goroutine that runs indefinitely until stopped
	err = rm.StartGoroutine(ctx, "long-running", func(ctx context.Context) {
		// Keep running until explicitly stopped
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-stopChan:
				return
			case <-ticker.C:
				// Keep running
			}
		}
	})

	if err != nil {
		t.Errorf("Expected no error starting goroutine, got: %v", err)
	}

	// Give the goroutine time to start
	time.Sleep(50 * time.Millisecond)

	// Verify the goroutine is actually tracked
	count := rm.GetGoroutineCount()
	if count == 0 {
		t.Error("Expected goroutine count > 0, but got 0")
	}
	t.Logf("Goroutine count before shutdown: %d", count)

	// Shutdown should timeout because the goroutine will still be running
	// Use a context with no timeout to rely on the internal shutdown timeout
	start := time.Now()
	err = rm.Shutdown(context.Background())
	elapsed := time.Since(start)

	t.Logf("Shutdown took: %v, error: %v", elapsed, err)

	// Should timeout after approximately the shutdown timeout
	if err == nil {
		t.Error("Expected shutdown to timeout")
	}

	// Should take at least as long as the shutdown timeout minus some tolerance
	if elapsed < 150*time.Millisecond {
		t.Errorf("Shutdown finished too quickly: %v, expected at least 150ms", elapsed)
	}

	// Clean up - stop the goroutine
	close(stopChan)
	time.Sleep(100 * time.Millisecond)
}

func TestResourceManager_ConcurrentGoroutineAccess(t *testing.T) {
	config := &config.EnvironmentConfig{
		MaxMemoryMB:           500,
		MaxGoroutines:         50, // Enough for concurrent test
		ShutdownTimeout:       5 * time.Second,
		ResourceCheckInterval: 1 * time.Second,
	}

	rm := NewResourceManager(config)
	defer rm.Shutdown(context.Background())

	ctx := context.Background()
	var wg sync.WaitGroup
	numWorkers := 20

	// Launch multiple goroutines concurrently
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			err := rm.StartGoroutine(ctx, "concurrent-worker", func(ctx context.Context) {
				time.Sleep(50 * time.Millisecond)
			})
			if err != nil {
				t.Errorf("Worker %d failed to start goroutine: %v", id, err)
			}
		}(i)
	}

	wg.Wait()

	// Wait for all goroutines to finish
	time.Sleep(200 * time.Millisecond)

	// Verify all goroutines finished
	if count := rm.GetGoroutineCount(); count != 0 {
		t.Errorf("Expected goroutine count 0, got %d", count)
	}
}

// Benchmark tests
func BenchmarkResourceManager_StartGoroutine(b *testing.B) {
	config := &config.EnvironmentConfig{
		MaxMemoryMB:           500,
		MaxGoroutines:         1000,
		ShutdownTimeout:       5 * time.Second,
		ResourceCheckInterval: 10 * time.Second,
	}

	rm := NewResourceManager(config)
	defer rm.Shutdown(context.Background())

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rm.StartGoroutine(ctx, "bench-goroutine", func(ctx context.Context) {
				// Minimal work
			})
		}
	})
}

func BenchmarkResourceManager_GetGoroutineCount(b *testing.B) {
	config := &config.EnvironmentConfig{
		MaxMemoryMB:           500,
		MaxGoroutines:         100,
		ShutdownTimeout:       5 * time.Second,
		ResourceCheckInterval: 10 * time.Second,
	}

	rm := NewResourceManager(config)
	defer rm.Shutdown(context.Background())

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rm.GetGoroutineCount()
		}
	})
}
