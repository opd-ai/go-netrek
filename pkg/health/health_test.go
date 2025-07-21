package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"
)

// mockHealthCheck implements HealthCheck for testing
type mockHealthCheck struct {
	name    string
	healthy bool
	err     error
}

func (m *mockHealthCheck) Name() string {
	return m.name
}

func (m *mockHealthCheck) Check(ctx context.Context) error {
	if !m.healthy {
		if m.err != nil {
			return m.err
		}
		return fmt.Errorf("mock health check failed")
	}
	return nil
}

// slowHealthCheck implements HealthCheck with configurable delay for testing timeouts
type slowHealthCheck struct {
	name    string
	healthy bool
	delay   time.Duration
}

func (s *slowHealthCheck) Name() string {
	return s.name
}

func (s *slowHealthCheck) Check(ctx context.Context) error {
	select {
	case <-time.After(s.delay):
		if !s.healthy {
			return fmt.Errorf("slow health check failed")
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func TestNewHealthChecker(t *testing.T) {
	hc := NewHealthChecker()
	if hc == nil {
		t.Fatal("NewHealthChecker() returned nil")
	}
	if hc.checks == nil {
		t.Error("checks map not initialized")
	}
}

func TestHealthChecker_AddCheck(t *testing.T) {
	hc := NewHealthChecker()

	check := &mockHealthCheck{name: "test", healthy: true}
	hc.AddCheck(check)

	if len(hc.checks) != 1 {
		t.Errorf("Expected 1 check, got %d", len(hc.checks))
	}

	if hc.checks["test"] != check {
		t.Error("Check not properly stored")
	}
}

func TestHealthChecker_RemoveCheck(t *testing.T) {
	hc := NewHealthChecker()

	check := &mockHealthCheck{name: "test", healthy: true}
	hc.AddCheck(check)
	hc.RemoveCheck("test")

	if len(hc.checks) != 0 {
		t.Errorf("Expected 0 checks after removal, got %d", len(hc.checks))
	}
}

func TestHealthChecker_CheckHealth(t *testing.T) {
	tests := []struct {
		name     string
		checks   []*mockHealthCheck
		expected string
	}{
		{
			name:     "no checks - healthy",
			checks:   []*mockHealthCheck{},
			expected: "healthy",
		},
		{
			name: "all healthy",
			checks: []*mockHealthCheck{
				{name: "check1", healthy: true},
				{name: "check2", healthy: true},
			},
			expected: "healthy",
		},
		{
			name: "one unhealthy",
			checks: []*mockHealthCheck{
				{name: "check1", healthy: true},
				{name: "check2", healthy: false},
			},
			expected: "unhealthy",
		},
		{
			name: "all unhealthy",
			checks: []*mockHealthCheck{
				{name: "check1", healthy: false},
				{name: "check2", healthy: false},
			},
			expected: "unhealthy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc := NewHealthChecker()

			for _, check := range tt.checks {
				hc.AddCheck(check)
			}

			ctx := context.Background()
			status := hc.CheckHealth(ctx)

			if status.Status != tt.expected {
				t.Errorf("Expected status %s, got %s", tt.expected, status.Status)
			}

			if len(status.Checks) != len(tt.checks) {
				t.Errorf("Expected %d check results, got %d", len(tt.checks), len(status.Checks))
			}

			for _, check := range tt.checks {
				result, exists := status.Checks[check.name]
				if !exists {
					t.Errorf("Check result for %s not found", check.name)
					continue
				}

				expectedStatus := "healthy"
				if !check.healthy {
					expectedStatus = "unhealthy"
				}

				if result.Status != expectedStatus {
					t.Errorf("Check %s: expected status %s, got %s", check.name, expectedStatus, result.Status)
				}
			}
		})
	}
}

func TestHealthChecker_CheckHealthWithTimeout(t *testing.T) {
	hc := NewHealthChecker()

	// Create a slow check that respects context timeout
	slowCheck := &slowHealthCheck{
		name:    "slow",
		healthy: true,
		delay:   100 * time.Millisecond,
	}

	hc.AddCheck(slowCheck)

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	status := hc.CheckHealth(ctx)

	// The check should fail due to timeout
	if status.Status != "unhealthy" {
		t.Errorf("Expected unhealthy status due to timeout, got %s", status.Status)
	}

	result, exists := status.Checks["slow"]
	if !exists {
		t.Fatal("Slow check result not found")
	}

	if result.Status != "unhealthy" {
		t.Errorf("Expected unhealthy status for slow check, got %s", result.Status)
	}
}

func TestHealthChecker_LivenessHandler(t *testing.T) {
	hc := NewHealthChecker()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	hc.LivenessHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["status"] != "alive" {
		t.Errorf("Expected status 'alive', got %s", response["status"])
	}
}

func TestHealthChecker_ReadinessHandler(t *testing.T) {
	tests := []struct {
		name               string
		checks             []*mockHealthCheck
		expectedStatusCode int
		expectedStatus     string
	}{
		{
			name:               "healthy service",
			checks:             []*mockHealthCheck{{name: "test", healthy: true}},
			expectedStatusCode: http.StatusOK,
			expectedStatus:     "healthy",
		},
		{
			name:               "unhealthy service",
			checks:             []*mockHealthCheck{{name: "test", healthy: false}},
			expectedStatusCode: http.StatusServiceUnavailable,
			expectedStatus:     "unhealthy",
		},
		{
			name:               "no checks",
			checks:             []*mockHealthCheck{},
			expectedStatusCode: http.StatusOK,
			expectedStatus:     "healthy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc := NewHealthChecker()

			for _, check := range tt.checks {
				hc.AddCheck(check)
			}

			req := httptest.NewRequest("GET", "/ready", nil)
			w := httptest.NewRecorder()

			hc.ReadinessHandler(w, req)

			if w.Code != tt.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatusCode, w.Code)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", contentType)
			}

			var response HealthStatus
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if response.Status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, response.Status)
			}
		})
	}
}

func TestGameEngineHealthCheck(t *testing.T) {
	tests := []struct {
		name        string
		gameRunning bool
		expectError bool
	}{
		{
			name:        "game running",
			gameRunning: true,
			expectError: false,
		},
		{
			name:        "game not running",
			gameRunning: false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := NewGameEngineHealthCheck(func() bool {
				return tt.gameRunning
			})

			if check.Name() != "game_engine" {
				t.Errorf("Expected name 'game_engine', got %s", check.Name())
			}

			err := check.Check(context.Background())

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestNetworkHealthCheck(t *testing.T) {
	tests := []struct {
		name         string
		listenerAddr string
		expectError  bool
	}{
		{
			name:         "network active",
			listenerAddr: "localhost:8080",
			expectError:  false,
		},
		{
			name:         "network inactive",
			listenerAddr: "",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := NewNetworkHealthCheck(func() string {
				return tt.listenerAddr
			})

			if check.Name() != "network" {
				t.Errorf("Expected name 'network', got %s", check.Name())
			}

			err := check.Check(context.Background())

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestMemoryHealthCheck(t *testing.T) {
	tests := []struct {
		name         string
		maxMemoryMB  int64
		currentMemMB int64
		expectError  bool
	}{
		{
			name:         "memory usage within limit",
			maxMemoryMB:  100,
			currentMemMB: 50,
			expectError:  false,
		},
		{
			name:         "memory usage at limit",
			maxMemoryMB:  100,
			currentMemMB: 100,
			expectError:  false,
		},
		{
			name:         "memory usage exceeds limit",
			maxMemoryMB:  100,
			currentMemMB: 150,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := NewMemoryHealthCheck(tt.maxMemoryMB, func() int64 {
				return tt.currentMemMB
			})

			if check.Name() != "memory" {
				t.Errorf("Expected name 'memory', got %s", check.Name())
			}

			err := check.Check(context.Background())

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkHealthChecker_CheckHealth(b *testing.B) {
	hc := NewHealthChecker()

	// Add multiple health checks
	for i := 0; i < 10; i++ {
		check := &mockHealthCheck{
			name:    fmt.Sprintf("check%d", i),
			healthy: true,
		}
		hc.AddCheck(check)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hc.CheckHealth(ctx)
	}
}

func BenchmarkHealthChecker_LivenessHandler(b *testing.B) {
	hc := NewHealthChecker()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		hc.LivenessHandler(w, req)
	}
}

// Helper function to get current memory usage in MB
func getCurrentMemoryMB() int64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return int64(m.Alloc / 1024 / 1024)
}
