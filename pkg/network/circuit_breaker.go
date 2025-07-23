// Package network provides circuit breaker functionality for reliable network operations.
// This implements the circuit breaker pattern to prevent cascading failures and
// provide graceful degradation during network outages.
package network

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/sony/gobreaker"

	"github.com/opd-ai/go-netrek/pkg/config"
	"github.com/opd-ai/go-netrek/pkg/logging"
)

// NetworkService wraps network operations with circuit breaker functionality.
// It provides retry logic, exponential backoff, and failure isolation to prevent
// cascading failures in the game networking layer.
type NetworkService struct {
	breaker *gobreaker.CircuitBreaker
	logger  *logging.Logger
	config  *config.EnvironmentConfig
}

// NetworkOperation represents a function that performs a network operation.
// It should return an error if the operation fails.
type NetworkOperation func() error

// NewNetworkService creates a new NetworkService with circuit breaker configured
// from environment settings. The circuit breaker will prevent cascading failures
// by isolating failing network operations.
func NewNetworkService(envConfig *config.EnvironmentConfig) *NetworkService {
	logger := logging.NewLogger()

	settings := gobreaker.Settings{
		Name:        "netrek-network",
		MaxRequests: uint32(envConfig.CircuitBreakerMaxRequests),
		Interval:    envConfig.CircuitBreakerInterval,
		Timeout:     envConfig.CircuitBreakerTimeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// Trip the circuit if we have too many consecutive failures
			return counts.ConsecutiveFailures >= uint32(envConfig.CircuitBreakerMaxConsecutiveFails)
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			logger.Info(context.Background(), "circuit breaker state changed",
				"name", name,
				"from", from,
				"to", to,
			)
		},
	}

	return &NetworkService{
		breaker: gobreaker.NewCircuitBreaker(settings),
		logger:  logger,
		config:  envConfig,
	}
}

// Execute runs a network operation through the circuit breaker.
// If the circuit is open (too many failures), it will return an error immediately.
// If the circuit is closed or half-open, it will attempt the operation.
func (ns *NetworkService) Execute(ctx context.Context, operation NetworkOperation) error {
	_, err := ns.breaker.Execute(func() (interface{}, error) {
		return nil, operation()
	})
	if err != nil {
		ns.logger.LogWithContext(ctx, slog.LevelError, "circuit breaker execution failed",
			"error", err,
			"state", ns.breaker.State(),
		)
		return fmt.Errorf("circuit breaker: %w", err)
	}

	return nil
}

// ExecuteWithRetry runs a network operation with retry logic and exponential backoff.
// It will attempt the operation multiple times with increasing delays between attempts.
// The circuit breaker state is checked before each retry attempt.
func (ns *NetworkService) ExecuteWithRetry(ctx context.Context, operation NetworkOperation) error {
	maxRetries := 3
	baseDelay := 1 * time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		err := ns.Execute(ctx, operation)
		if err == nil {
			// Operation succeeded
			return nil
		}

		// Check if this is a circuit breaker error (circuit is open)
		if ns.breaker.State() == gobreaker.StateOpen {
			// Circuit is open, no point in retrying
			ns.logger.LogWithContext(ctx, slog.LevelWarn, "circuit breaker is open, skipping retries",
				"attempt", attempt+1,
				"max_retries", maxRetries,
			)
			return err
		}

		// If this is the last attempt, return the error
		if attempt == maxRetries-1 {
			ns.logger.LogWithContext(ctx, slog.LevelError, "all retry attempts failed",
				"attempts", maxRetries,
				"final_error", err,
			)
			return fmt.Errorf("max retries (%d) exceeded: %w", maxRetries, err)
		}

		// Calculate delay with exponential backoff
		delay := time.Duration(attempt+1) * baseDelay
		ns.logger.LogWithContext(ctx, slog.LevelWarn, "operation failed, retrying",
			"attempt", attempt+1,
			"max_retries", maxRetries,
			"delay", delay,
			"error", err,
		)

		// Wait before retrying, respecting context cancellation
		select {
		case <-time.After(delay):
			// Continue to next retry
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		}
	}

	return fmt.Errorf("unexpected exit from retry loop")
}

// GetState returns the current state of the circuit breaker.
// Useful for monitoring and debugging purposes.
func (ns *NetworkService) GetState() gobreaker.State {
	return ns.breaker.State()
}

// GetCounts returns the current failure/success counts of the circuit breaker.
// Useful for monitoring and metrics collection.
func (ns *NetworkService) GetCounts() gobreaker.Counts {
	return ns.breaker.Counts()
}
