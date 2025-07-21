package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/opd-ai/go-netrek/pkg/config"
	"github.com/opd-ai/go-netrek/pkg/engine"
	"github.com/opd-ai/go-netrek/pkg/network"
)

// TestHealthCheckIntegration tests the health check system with real game components
func TestHealthCheckIntegration(t *testing.T) {
	// Create a test game configuration
	gameConfig := config.DefaultConfig()
	gameConfig.MaxPlayers = 2

	// Create game and server
	game := engine.NewGame(gameConfig)
	server := network.NewGameServer(game, gameConfig.MaxPlayers)

	// Setup health checker with real components
	healthChecker := NewHealthChecker()

	// Add real health checks
	healthChecker.AddCheck(NewGameEngineHealthCheck(
		func() bool { return server.GetGameRunning() },
	))

	healthChecker.AddCheck(NewNetworkHealthCheck(
		func() string { return server.GetListenerAddress() },
	))

	t.Run("health checks before server start", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		health := healthChecker.CheckHealth(ctx)

		// Game engine should be unhealthy (not started yet)
		if health.Checks["game_engine"].Status != "unhealthy" {
			t.Error("Game engine should be unhealthy before server start")
		}

		// Network should be unhealthy (not listening yet)
		if health.Checks["network"].Status != "unhealthy" {
			t.Error("Network should be unhealthy before server start")
		}

		// Overall status should be unhealthy
		if health.Status != "unhealthy" {
			t.Error("Overall status should be unhealthy before server start")
		}
	})

	// Start the server on a random port for testing
	go func() {
		if err := server.Start("localhost:0"); err != nil {
			t.Errorf("Failed to start test server: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	t.Run("health checks after server start", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		health := healthChecker.CheckHealth(ctx)

		// Both checks should be healthy now
		if health.Checks["game_engine"].Status != "healthy" {
			t.Error("Game engine should be healthy after server start")
		}

		if health.Checks["network"].Status != "healthy" {
			t.Error("Network should be healthy after server start")
		}

		// Overall status should be healthy
		if health.Status != "healthy" {
			t.Errorf("Overall status should be healthy after server start, got: %s", health.Status)
		}
	})

	t.Run("liveness endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		healthChecker.LivenessHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]string
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response["status"] != "alive" {
			t.Errorf("Expected status 'alive', got %s", response["status"])
		}
	})

	t.Run("readiness endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/ready", nil)
		w := httptest.NewRecorder()

		healthChecker.ReadinessHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var response HealthStatus
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Status != "healthy" {
			t.Errorf("Expected status 'healthy', got %s", response.Status)
		}
	})

	// Clean up
	server.Stop()
}

// TestHealthCheckWithFailures tests health check behavior when components fail
func TestHealthCheckWithFailures(t *testing.T) {
	healthChecker := NewHealthChecker()

	// Add a check that will fail
	failingCheck := &mockHealthCheck{
		name:    "failing_component",
		healthy: false,
		err:     fmt.Errorf("component is down"),
	}

	healthChecker.AddCheck(failingCheck)

	t.Run("readiness endpoint with failures", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/ready", nil)
		w := httptest.NewRecorder()

		healthChecker.ReadinessHandler(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("Expected status code %d, got %d", http.StatusServiceUnavailable, w.Code)
		}

		var response HealthStatus
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Status != "unhealthy" {
			t.Errorf("Expected status 'unhealthy', got %s", response.Status)
		}

		if response.Checks["failing_component"].Status != "unhealthy" {
			t.Error("Failing component should be marked as unhealthy")
		}

		if response.Checks["failing_component"].Message == "" {
			t.Error("Failing component should have an error message")
		}
	})
}

// TestMemoryHealthCheckIntegration tests memory health check with real memory stats
func TestMemoryHealthCheckIntegration(t *testing.T) {
	healthChecker := NewHealthChecker()

	// Add memory check with very high limit (should pass)
	memoryCheck := NewMemoryHealthCheck(10000, getCurrentMemoryMB) // 10GB limit
	healthChecker.AddCheck(memoryCheck)

	t.Run("memory check with high limit", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		health := healthChecker.CheckHealth(ctx)

		if health.Checks["memory"].Status != "healthy" {
			t.Errorf("Memory check should be healthy with high limit, got: %s",
				health.Checks["memory"].Message)
		}
	})

	// Remove the previous check and add one with very low limit (should fail)
	healthChecker.RemoveCheck("memory")

	// Use a mock function that returns high memory usage
	mockHighMemory := func() int64 { return 100 }              // 100MB usage
	lowMemoryCheck := NewMemoryHealthCheck(50, mockHighMemory) // 50MB limit
	healthChecker.AddCheck(lowMemoryCheck)

	t.Run("memory check with low limit", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		health := healthChecker.CheckHealth(ctx)

		if health.Checks["memory"].Status != "unhealthy" {
			t.Error("Memory check should be unhealthy with low limit")
		}

		if health.Status != "unhealthy" {
			t.Error("Overall status should be unhealthy due to memory limit")
		}
	})
}
