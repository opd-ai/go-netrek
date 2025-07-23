package network

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sony/gobreaker"

	"github.com/opd-ai/go-netrek/pkg/config"
)

// TestNetworkService_Execute tests basic circuit breaker execution
func TestNetworkService_Execute(t *testing.T) {
	envConfig := &config.EnvironmentConfig{
		CircuitBreakerMaxRequests:         3,
		CircuitBreakerInterval:            60 * time.Second,
		CircuitBreakerTimeout:             30 * time.Second,
		CircuitBreakerMaxConsecutiveFails: 5,
	}

	ns := NewNetworkService(envConfig)
	ctx := context.Background()

	t.Run("successful operation", func(t *testing.T) {
		err := ns.Execute(ctx, func() error {
			return nil // Success
		})
		if err != nil {
			t.Errorf("Expected nil error, got %v", err)
		}

		if ns.GetState() != gobreaker.StateClosed {
			t.Errorf("Expected circuit breaker to be closed, got %v", ns.GetState())
		}
	})

	t.Run("failed operation", func(t *testing.T) {
		testError := errors.New("test error")
		err := ns.Execute(ctx, func() error {
			return testError
		})

		if err == nil {
			t.Error("Expected error, got nil")
		}

		// Circuit should still be closed after one failure
		if ns.GetState() != gobreaker.StateClosed {
			t.Errorf("Expected circuit breaker to be closed after one failure, got %v", ns.GetState())
		}
	})
}

// TestNetworkService_CircuitBreakerTrip tests that the circuit breaker trips after consecutive failures
func TestNetworkService_CircuitBreakerTrip(t *testing.T) {
	envConfig := &config.EnvironmentConfig{
		CircuitBreakerMaxRequests:         3,
		CircuitBreakerInterval:            60 * time.Second,
		CircuitBreakerTimeout:             1 * time.Second, // Short timeout for testing
		CircuitBreakerMaxConsecutiveFails: 3,               // Low threshold for testing
	}

	ns := NewNetworkService(envConfig)
	ctx := context.Background()
	testError := errors.New("test failure")

	// Cause enough failures to trip the circuit
	for i := 0; i < 3; i++ {
		err := ns.Execute(ctx, func() error {
			return testError
		})
		if err == nil {
			t.Errorf("Expected error on attempt %d, got nil", i+1)
		}
	}

	// Circuit should now be open
	if ns.GetState() != gobreaker.StateOpen {
		t.Errorf("Expected circuit breaker to be open after failures, got %v", ns.GetState())
	}

	// Next operation should fail immediately due to open circuit
	err := ns.Execute(ctx, func() error {
		t.Error("Operation should not be called when circuit is open")
		return nil
	})

	if err == nil {
		t.Error("Expected error when circuit is open, got nil")
	}
}

// TestNetworkService_CircuitBreakerRecovery tests circuit breaker recovery
func TestNetworkService_CircuitBreakerRecovery(t *testing.T) {
	envConfig := &config.EnvironmentConfig{
		CircuitBreakerMaxRequests:         2,
		CircuitBreakerInterval:            60 * time.Second,
		CircuitBreakerTimeout:             100 * time.Millisecond, // Very short for testing
		CircuitBreakerMaxConsecutiveFails: 2,
	}

	ns := NewNetworkService(envConfig)
	ctx := context.Background()
	testError := errors.New("test failure")

	// Trip the circuit
	for i := 0; i < 2; i++ {
		ns.Execute(ctx, func() error { return testError })
	}

	// Verify circuit is open
	if ns.GetState() != gobreaker.StateOpen {
		t.Errorf("Expected circuit breaker to be open, got %v", ns.GetState())
	}

	// Wait for circuit to move to half-open
	time.Sleep(150 * time.Millisecond)

	// Try a successful operation to close the circuit
	err := ns.Execute(ctx, func() error {
		return nil // Success
	})
	if err != nil {
		t.Errorf("Expected successful operation, got error: %v", err)
	}

	// Circuit should be closed again after successful operation
	// Note: After a successful operation in half-open state, circuit may stay half-open
	// until the interval passes. This is normal gobreaker behavior.
	state := ns.GetState()
	if state != gobreaker.StateClosed && state != gobreaker.StateHalfOpen {
		t.Errorf("Expected circuit breaker to be closed or half-open after recovery, got %v", state)
	}
}

// TestNetworkService_ExecuteWithRetry tests retry logic with exponential backoff
func TestNetworkService_ExecuteWithRetry(t *testing.T) {
	envConfig := &config.EnvironmentConfig{
		CircuitBreakerMaxRequests:         3,
		CircuitBreakerInterval:            60 * time.Second,
		CircuitBreakerTimeout:             30 * time.Second,
		CircuitBreakerMaxConsecutiveFails: 10, // High threshold to avoid tripping during test
	}

	ns := NewNetworkService(envConfig)
	ctx := context.Background()

	t.Run("eventual success", func(t *testing.T) {
		attempt := 0
		testError := errors.New("temporary failure")

		err := ns.ExecuteWithRetry(ctx, func() error {
			attempt++
			if attempt < 3 {
				return testError // Fail first two attempts
			}
			return nil // Succeed on third attempt
		})
		if err != nil {
			t.Errorf("Expected eventual success, got error: %v", err)
		}

		if attempt != 3 {
			t.Errorf("Expected 3 attempts, got %d", attempt)
		}
	})

	t.Run("all retries fail", func(t *testing.T) {
		attempt := 0
		testError := errors.New("persistent failure")

		err := ns.ExecuteWithRetry(ctx, func() error {
			attempt++
			return testError // Always fail
		})

		if err == nil {
			t.Error("Expected error after all retries fail, got nil")
		}

		if attempt != 3 {
			t.Errorf("Expected 3 attempts, got %d", attempt)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		// Cancel after a short delay
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		err := ns.ExecuteWithRetry(ctx, func() error {
			return errors.New("failure")
		})

		if err == nil {
			t.Error("Expected error due to context cancellation, got nil")
		}

		if ctx.Err() == nil {
			t.Error("Expected context to be cancelled")
		}
	})
}

// TestNetworkService_GetState tests state inspection methods
func TestNetworkService_GetState(t *testing.T) {
	envConfig := &config.EnvironmentConfig{
		CircuitBreakerMaxRequests:         3,
		CircuitBreakerInterval:            60 * time.Second,
		CircuitBreakerTimeout:             30 * time.Second,
		CircuitBreakerMaxConsecutiveFails: 5,
	}

	ns := NewNetworkService(envConfig)

	// Initial state should be closed
	if ns.GetState() != gobreaker.StateClosed {
		t.Errorf("Expected initial state to be closed, got %v", ns.GetState())
	}

	// Counts should be empty initially
	counts := ns.GetCounts()
	if counts.Requests != 0 || counts.TotalSuccesses != 0 || counts.TotalFailures != 0 {
		t.Errorf("Expected empty counts initially, got %+v", counts)
	}
}

// TestNetworkService_Configuration tests configuration validation
func TestNetworkService_Configuration(t *testing.T) {
	testCases := []struct {
		name        string
		config      *config.EnvironmentConfig
		expectPanic bool
	}{
		{
			name: "valid configuration",
			config: &config.EnvironmentConfig{
				CircuitBreakerMaxRequests:         3,
				CircuitBreakerInterval:            60 * time.Second,
				CircuitBreakerTimeout:             30 * time.Second,
				CircuitBreakerMaxConsecutiveFails: 5,
			},
			expectPanic: false,
		},
		{
			name: "zero max requests",
			config: &config.EnvironmentConfig{
				CircuitBreakerMaxRequests:         0,
				CircuitBreakerInterval:            60 * time.Second,
				CircuitBreakerTimeout:             30 * time.Second,
				CircuitBreakerMaxConsecutiveFails: 5,
			},
			expectPanic: false, // gobreaker should handle this gracefully
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tc.expectPanic {
					t.Errorf("Unexpected panic: %v", r)
				} else if r == nil && tc.expectPanic {
					t.Error("Expected panic but none occurred")
				}
			}()

			ns := NewNetworkService(tc.config)
			if ns == nil {
				t.Error("Expected valid NetworkService, got nil")
			}
		})
	}
}
