// pkg/resource/health_test.go
package resource

import (
	"context"
	"testing"
	"time"

	"github.com/opd-ai/go-netrek/pkg/config"
)

func TestResourceHealthCheck_Name(t *testing.T) {
	config := &config.EnvironmentConfig{
		MaxMemoryMB:           500,
		MaxGoroutines:         100,
		ShutdownTimeout:       5 * time.Second,
		ResourceCheckInterval: 10 * time.Second,
	}

	rm := NewResourceManager(config)
	defer rm.Shutdown(context.Background())

	check := NewResourceHealthCheck(rm)

	if check.Name() != "resource" {
		t.Errorf("Expected name 'resource', got %s", check.Name())
	}
}

func TestResourceHealthCheck_Check_Healthy(t *testing.T) {
	config := &config.EnvironmentConfig{
		MaxMemoryMB:           1000, // High limit
		MaxGoroutines:         100,  // High limit
		ShutdownTimeout:       5 * time.Second,
		ResourceCheckInterval: 10 * time.Second,
	}

	rm := NewResourceManager(config)
	defer rm.Shutdown(context.Background())

	// Update memory stats
	rm.CheckMemoryUsage()

	check := NewResourceHealthCheck(rm)
	err := check.Check(context.Background())

	if err != nil {
		t.Errorf("Expected healthy check to pass, got error: %v", err)
	}
}

func TestResourceHealthCheck_Check_MemoryUnhealthy(t *testing.T) {
	// Create a manager with very low memory limit
	config := &config.EnvironmentConfig{
		MaxMemoryMB:           1, // Very low limit to trigger failure
		MaxGoroutines:         100,
		ShutdownTimeout:       5 * time.Second,
		ResourceCheckInterval: 10 * time.Second,
	}

	rm := NewResourceManager(config)
	defer rm.Shutdown(context.Background())

	// Allocate memory to ensure we exceed the limit
	data := make([]byte, 2*1024*1024) // 2MB, which exceeds 1MB limit
	_ = data

	// Update memory stats - should exceed limit
	rm.CheckMemoryUsage()

	check := NewResourceHealthCheck(rm)
	err := check.Check(context.Background())

	if err == nil {
		t.Error("Expected health check to fail due to memory limit")
	}
}

func TestResourceHealthCheck_Check_GoroutineUnhealthy(t *testing.T) {
	config := &config.EnvironmentConfig{
		MaxMemoryMB:           1000,
		MaxGoroutines:         5, // Low limit for testing
		ShutdownTimeout:       5 * time.Second,
		ResourceCheckInterval: 10 * time.Second,
	}

	rm := NewResourceManager(config)
	defer rm.Shutdown(context.Background())

	ctx := context.Background()

	// Start enough goroutines to exceed 80% threshold (4+ out of 5)
	for i := 0; i < 5; i++ {
		err := rm.StartGoroutine(ctx, "test-goroutine", func(ctx context.Context) {
			time.Sleep(200 * time.Millisecond) // Keep running during check
		})
		if err != nil {
			// Expected when we hit the limit
			break
		}
	}

	// Wait a moment for goroutines to start
	time.Sleep(50 * time.Millisecond)

	check := NewResourceHealthCheck(rm)
	err := check.Check(context.Background())

	if err == nil {
		t.Error("Expected health check to fail due to goroutine threshold")
	}

	// Wait for goroutines to finish
	time.Sleep(250 * time.Millisecond)
}
