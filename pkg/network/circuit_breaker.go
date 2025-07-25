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
	const maxRetries = 3
	const baseDelay = 1 * time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		err := ns.Execute(ctx, operation)
		if err == nil {
			return nil
		}

		if ns.shouldSkipRetries(ctx, err, attempt, maxRetries) {
			return err
		}

		if ns.isLastAttempt(attempt, maxRetries) {
			return ns.handleFinalAttemptFailure(ctx, maxRetries, err)
		}

		if err := ns.waitBeforeRetry(ctx, attempt, baseDelay, maxRetries, err); err != nil {
			return err
		}
	}

	return fmt.Errorf("unexpected exit from retry loop")
}

// shouldSkipRetries checks if retries should be skipped due to circuit breaker state.
func (ns *NetworkService) shouldSkipRetries(ctx context.Context, err error, attempt, maxRetries int) bool {
	if ns.breaker.State() == gobreaker.StateOpen {
		ns.logger.LogWithContext(ctx, slog.LevelWarn, "circuit breaker is open, skipping retries",
			"attempt", attempt+1,
			"max_retries", maxRetries,
		)
		return true
	}
	return false
}

// isLastAttempt checks if this is the final retry attempt.
func (ns *NetworkService) isLastAttempt(attempt, maxRetries int) bool {
	return attempt == maxRetries-1
}

// handleFinalAttemptFailure logs and returns error for the final failed attempt.
func (ns *NetworkService) handleFinalAttemptFailure(ctx context.Context, maxRetries int, err error) error {
	ns.logger.LogWithContext(ctx, slog.LevelError, "all retry attempts failed",
		"attempts", maxRetries,
		"final_error", err,
	)
	return fmt.Errorf("max retries (%d) exceeded: %w", maxRetries, err)
}

// waitBeforeRetry implements exponential backoff delay with context cancellation support.
func (ns *NetworkService) waitBeforeRetry(ctx context.Context, attempt int, baseDelay time.Duration, maxRetries int, err error) error {
	delay := ns.calculateBackoffDelay(attempt, baseDelay)

	ns.logger.LogWithContext(ctx, slog.LevelWarn, "operation failed, retrying",
		"attempt", attempt+1,
		"max_retries", maxRetries,
		"delay", delay,
		"error", err,
	)

	select {
	case <-time.After(delay):
		return nil
	case <-ctx.Done():
		return fmt.Errorf("retry cancelled: %w", ctx.Err())
	}
}

// calculateBackoffDelay computes the exponential backoff delay for the given attempt.
func (ns *NetworkService) calculateBackoffDelay(attempt int, baseDelay time.Duration) time.Duration {
	return time.Duration(attempt+1) * baseDelay
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
