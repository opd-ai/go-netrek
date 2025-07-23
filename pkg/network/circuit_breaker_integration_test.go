package network

import (
	"testing"

	"github.com/opd-ai/go-netrek/pkg/event"
)

// TestCircuitBreakerIntegration tests circuit breaker integration with GameClient
func TestCircuitBreakerIntegration(t *testing.T) {
	// Create an event bus for the client
	eventBus := event.NewEventBus()

	// Create a client with circuit breaker
	client := NewGameClient(eventBus)

	// Verify that the client has a network service
	if client.networkService == nil {
		t.Fatal("Expected client to have networkService, got nil")
	}

	// Verify initial circuit breaker state is closed
	if client.networkService.GetState().String() != "closed" {
		t.Errorf("Expected initial circuit breaker state to be closed, got %v", client.networkService.GetState())
	}
}

// TestCircuitBreakerFailureHandling tests that circuit breaker handles connection failures gracefully
func TestCircuitBreakerFailureHandling(t *testing.T) {
	// Create an event bus for the client
	eventBus := event.NewEventBus()

	// Create a client with circuit breaker
	client := NewGameClient(eventBus)

	// Try to connect to a non-existent server
	// This should trigger the circuit breaker retry logic
	err := client.Connect("localhost:99999", "TestPlayer", 0)

	// Connection should fail, but gracefully due to circuit breaker
	if err == nil {
		t.Error("Expected connection to fail for non-existent server")
	}

	// Circuit breaker should still be operational
	if client.networkService == nil {
		t.Error("Expected networkService to remain available after failed connection")
	}
}
