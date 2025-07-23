// pkg/config/env_config_test.go
package config

import (
	"os"
	"testing"
	"time"
)

// createValidConfig creates a valid EnvironmentConfig for testing
func createValidConfig() *EnvironmentConfig {
	return &EnvironmentConfig{
		ServerAddr:      "localhost",
		ServerPort:      4566,
		MaxClients:      32,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		UpdateRate:      20,
		TicksPerState:   3,
		UsePartialState: true,
		WorldSize:       10000.0,
		// Circuit Breaker Configuration
		CircuitBreakerMaxRequests:         3,
		CircuitBreakerInterval:            60 * time.Second,
		CircuitBreakerTimeout:             30 * time.Second,
		CircuitBreakerMaxConsecutiveFails: 5,
		// Resource Management Configuration
		MaxMemoryMB:           500,
		MaxGoroutines:         100,
		ShutdownTimeout:       30 * time.Second,
		ResourceCheckInterval: 10 * time.Second,
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"NETREK_SERVER_ADDR",
		"NETREK_SERVER_PORT",
		"NETREK_MAX_CLIENTS",
		"NETREK_READ_TIMEOUT",
		"NETREK_WRITE_TIMEOUT",
		"NETREK_UPDATE_RATE",
		"NETREK_TICKS_PER_STATE",
		"NETREK_USE_PARTIAL_STATE",
		"NETREK_WORLD_SIZE",
	}

	for _, key := range envVars {
		originalEnv[key] = os.Getenv(key)
		os.Unsetenv(key)
	}

	// Restore environment after test
	defer func() {
		for key, value := range originalEnv {
			if value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	t.Run("DefaultValues", func(t *testing.T) {
		config, err := LoadConfigFromEnv()
		if err != nil {
			t.Fatalf("LoadConfigFromEnv() failed: %v", err)
		}

		// Test default values
		if config.ServerAddr != "localhost" {
			t.Errorf("Expected ServerAddr 'localhost', got '%s'", config.ServerAddr)
		}
		if config.ServerPort != 4566 {
			t.Errorf("Expected ServerPort 4566, got %d", config.ServerPort)
		}
		if config.MaxClients != 32 {
			t.Errorf("Expected MaxClients 32, got %d", config.MaxClients)
		}
		if config.ReadTimeout != 30*time.Second {
			t.Errorf("Expected ReadTimeout 30s, got %v", config.ReadTimeout)
		}
		if config.WriteTimeout != 30*time.Second {
			t.Errorf("Expected WriteTimeout 30s, got %v", config.WriteTimeout)
		}
		if config.UpdateRate != 20 {
			t.Errorf("Expected UpdateRate 20, got %d", config.UpdateRate)
		}
		if config.TicksPerState != 3 {
			t.Errorf("Expected TicksPerState 3, got %d", config.TicksPerState)
		}
		if !config.UsePartialState {
			t.Errorf("Expected UsePartialState true, got %v", config.UsePartialState)
		}
		if config.WorldSize != 10000.0 {
			t.Errorf("Expected WorldSize 10000.0, got %f", config.WorldSize)
		}
	})

	t.Run("EnvironmentOverrides", func(t *testing.T) {
		// Set environment variables
		os.Setenv("NETREK_SERVER_ADDR", "192.168.1.100")
		os.Setenv("NETREK_SERVER_PORT", "8080")
		os.Setenv("NETREK_MAX_CLIENTS", "64")
		os.Setenv("NETREK_READ_TIMEOUT", "45s")
		os.Setenv("NETREK_WRITE_TIMEOUT", "60s")
		os.Setenv("NETREK_UPDATE_RATE", "30")
		os.Setenv("NETREK_TICKS_PER_STATE", "5")
		os.Setenv("NETREK_USE_PARTIAL_STATE", "false")
		os.Setenv("NETREK_WORLD_SIZE", "15000.0")

		config, err := LoadConfigFromEnv()
		if err != nil {
			t.Fatalf("LoadConfigFromEnv() failed: %v", err)
		}

		// Test environment overrides
		if config.ServerAddr != "192.168.1.100" {
			t.Errorf("Expected ServerAddr '192.168.1.100', got '%s'", config.ServerAddr)
		}
		if config.ServerPort != 8080 {
			t.Errorf("Expected ServerPort 8080, got %d", config.ServerPort)
		}
		if config.MaxClients != 64 {
			t.Errorf("Expected MaxClients 64, got %d", config.MaxClients)
		}
		if config.ReadTimeout != 45*time.Second {
			t.Errorf("Expected ReadTimeout 45s, got %v", config.ReadTimeout)
		}
		if config.WriteTimeout != 60*time.Second {
			t.Errorf("Expected WriteTimeout 60s, got %v", config.WriteTimeout)
		}
		if config.UpdateRate != 30 {
			t.Errorf("Expected UpdateRate 30, got %d", config.UpdateRate)
		}
		if config.TicksPerState != 5 {
			t.Errorf("Expected TicksPerState 5, got %d", config.TicksPerState)
		}
		if config.UsePartialState {
			t.Errorf("Expected UsePartialState false, got %v", config.UsePartialState)
		}
		if config.WorldSize != 15000.0 {
			t.Errorf("Expected WorldSize 15000.0, got %f", config.WorldSize)
		}
	})
}

func TestValidateEnvironmentConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *EnvironmentConfig
		expectError bool
		errorField  string
	}{
		{
			name:        "ValidConfig",
			config:      createValidConfig(),
			expectError: false,
		},
		{
			name: "EmptyServerAddr",
			config: func() *EnvironmentConfig {
				c := createValidConfig()
				c.ServerAddr = ""
				return c
			}(),
			expectError: true,
			errorField:  "ServerAddr",
		},
		{
			name: "InvalidServerPortTooLow",
			config: func() *EnvironmentConfig {
				c := createValidConfig()
				c.ServerPort = 1023
				return c
			}(),
			expectError: true,
			errorField:  "ServerPort",
		},
		{
			name: "InvalidServerPortTooHigh",
			config: func() *EnvironmentConfig {
				c := createValidConfig()
				c.ServerPort = 65536
				return c
			}(),
			expectError: true,
			errorField:  "ServerPort",
		},
		{
			name: "InvalidMaxClientsTooLow",
			config: func() *EnvironmentConfig {
				c := createValidConfig()
				c.MaxClients = 0
				return c
			}(),
			expectError: true,
			errorField:  "MaxClients",
		},
		{
			name: "InvalidMaxClientsTooHigh",
			config: func() *EnvironmentConfig {
				c := createValidConfig()
				c.MaxClients = 1001
				return c
			}(),
			expectError: true,
			errorField:  "MaxClients",
		},
		{
			name: "InvalidReadTimeoutTooShort",
			config: func() *EnvironmentConfig {
				c := createValidConfig()
				c.ReadTimeout = 500 * time.Millisecond
				return c
			}(),
			expectError: true,
			errorField:  "ReadTimeout",
		},
		{
			name: "InvalidReadTimeoutTooLong",
			config: func() *EnvironmentConfig {
				c := createValidConfig()
				c.ReadTimeout = 2 * time.Minute
				return c
			}(),
			expectError: true,
			errorField:  "ReadTimeout",
		},
		{
			name: "InvalidUpdateRateTooLow",
			config: func() *EnvironmentConfig {
				c := createValidConfig()
				c.UpdateRate = 0
				return c
			}(),
			expectError: true,
			errorField:  "UpdateRate",
		},
		{
			name: "InvalidUpdateRateTooHigh",
			config: func() *EnvironmentConfig {
				c := createValidConfig()
				c.UpdateRate = 101
				return c
			}(),
			expectError: true,
			errorField:  "UpdateRate",
		},
		{
			name: "InvalidWorldSizeTooSmall",
			config: func() *EnvironmentConfig {
				c := createValidConfig()
				c.WorldSize = 999.0
				return c
			}(),
			expectError: true,
			errorField:  "WorldSize",
		},
		{
			name: "InvalidWorldSizeTooLarge",
			config: func() *EnvironmentConfig {
				c := createValidConfig()
				c.WorldSize = 100001.0
				return c
			}(),
			expectError: true,
			errorField:  "WorldSize",
		},
		{
			name: "InvalidCircuitBreakerMaxRequestsTooLow",
			config: func() *EnvironmentConfig {
				c := createValidConfig()
				c.CircuitBreakerMaxRequests = 0
				return c
			}(),
			expectError: true,
			errorField:  "CircuitBreakerMaxRequests",
		},
		{
			name: "InvalidCircuitBreakerIntervalTooShort",
			config: func() *EnvironmentConfig {
				c := createValidConfig()
				c.CircuitBreakerInterval = 500 * time.Millisecond
				return c
			}(),
			expectError: true,
			errorField:  "CircuitBreakerInterval",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEnvironmentConfig(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected validation error, but got none")
				} else if validationErr, ok := err.(*ValidationError); ok {
					if validationErr.Field != tt.errorField {
						t.Errorf("Expected error for field '%s', got error for field '%s'", tt.errorField, validationErr.Field)
					}
				} else {
					t.Errorf("Expected ValidationError, got %T: %v", err, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error, but got: %v", err)
				}
			}
		})
	}
}

func TestApplyEnvironmentOverrides(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"NETREK_SERVER_ADDR",
		"NETREK_SERVER_PORT",
		"NETREK_MAX_CLIENTS",
		"NETREK_UPDATE_RATE",
		"NETREK_WORLD_SIZE",
	}

	for _, key := range envVars {
		originalEnv[key] = os.Getenv(key)
		os.Unsetenv(key)
	}

	// Restore environment after test
	defer func() {
		for key, value := range originalEnv {
			if value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	// Set test environment variables
	os.Setenv("NETREK_SERVER_ADDR", "testhost")
	os.Setenv("NETREK_SERVER_PORT", "9999")
	os.Setenv("NETREK_MAX_CLIENTS", "100")
	os.Setenv("NETREK_UPDATE_RATE", "50")
	os.Setenv("NETREK_WORLD_SIZE", "20000.0")

	// Create initial game config
	gameConfig := DefaultConfig()

	// Apply environment overrides
	err := ApplyEnvironmentOverrides(gameConfig)
	if err != nil {
		t.Fatalf("ApplyEnvironmentOverrides failed: %v", err)
	}

	// Verify overrides were applied
	expectedServerAddr := "testhost:9999"
	if gameConfig.NetworkConfig.ServerAddress != expectedServerAddr {
		t.Errorf("Expected ServerAddress '%s', got '%s'", expectedServerAddr, gameConfig.NetworkConfig.ServerAddress)
	}

	if gameConfig.NetworkConfig.ServerPort != 9999 {
		t.Errorf("Expected ServerPort 9999, got %d", gameConfig.NetworkConfig.ServerPort)
	}

	if gameConfig.MaxPlayers != 100 {
		t.Errorf("Expected MaxPlayers 100, got %d", gameConfig.MaxPlayers)
	}

	if gameConfig.NetworkConfig.UpdateRate != 50 {
		t.Errorf("Expected UpdateRate 50, got %d", gameConfig.NetworkConfig.UpdateRate)
	}

	if gameConfig.WorldSize != 20000.0 {
		t.Errorf("Expected WorldSize 20000.0, got %f", gameConfig.WorldSize)
	}
}

func TestGetEnvHelperFunctions(t *testing.T) {
	// Test getEnvOrDefault
	os.Setenv("TEST_STRING", "test_value")
	if result := getEnvOrDefault("TEST_STRING", "default"); result != "test_value" {
		t.Errorf("getEnvOrDefault: expected 'test_value', got '%s'", result)
	}
	if result := getEnvOrDefault("NONEXISTENT", "default"); result != "default" {
		t.Errorf("getEnvOrDefault: expected 'default', got '%s'", result)
	}
	os.Unsetenv("TEST_STRING")

	// Test getEnvAsIntOrDefault
	os.Setenv("TEST_INT", "42")
	if result := getEnvAsIntOrDefault("TEST_INT", 10); result != 42 {
		t.Errorf("getEnvAsIntOrDefault: expected 42, got %d", result)
	}
	if result := getEnvAsIntOrDefault("NONEXISTENT", 10); result != 10 {
		t.Errorf("getEnvAsIntOrDefault: expected 10, got %d", result)
	}
	os.Setenv("TEST_INT", "invalid")
	if result := getEnvAsIntOrDefault("TEST_INT", 10); result != 10 {
		t.Errorf("getEnvAsIntOrDefault with invalid value: expected 10, got %d", result)
	}
	os.Unsetenv("TEST_INT")

	// Test getEnvAsBoolOrDefault
	os.Setenv("TEST_BOOL", "true")
	if result := getEnvAsBoolOrDefault("TEST_BOOL", false); result != true {
		t.Errorf("getEnvAsBoolOrDefault: expected true, got %v", result)
	}
	if result := getEnvAsBoolOrDefault("NONEXISTENT", false); result != false {
		t.Errorf("getEnvAsBoolOrDefault: expected false, got %v", result)
	}
	os.Setenv("TEST_BOOL", "invalid")
	if result := getEnvAsBoolOrDefault("TEST_BOOL", false); result != false {
		t.Errorf("getEnvAsBoolOrDefault with invalid value: expected false, got %v", result)
	}
	os.Unsetenv("TEST_BOOL")

	// Test getEnvAsFloatOrDefault
	os.Setenv("TEST_FLOAT", "3.14")
	if result := getEnvAsFloatOrDefault("TEST_FLOAT", 1.0); result != 3.14 {
		t.Errorf("getEnvAsFloatOrDefault: expected 3.14, got %f", result)
	}
	if result := getEnvAsFloatOrDefault("NONEXISTENT", 1.0); result != 1.0 {
		t.Errorf("getEnvAsFloatOrDefault: expected 1.0, got %f", result)
	}
	os.Setenv("TEST_FLOAT", "invalid")
	if result := getEnvAsFloatOrDefault("TEST_FLOAT", 1.0); result != 1.0 {
		t.Errorf("getEnvAsFloatOrDefault with invalid value: expected 1.0, got %f", result)
	}
	os.Unsetenv("TEST_FLOAT")

	// Test getEnvAsDurationOrDefault
	os.Setenv("TEST_DURATION", "5s")
	if result := getEnvAsDurationOrDefault("TEST_DURATION", time.Second); result != 5*time.Second {
		t.Errorf("getEnvAsDurationOrDefault: expected 5s, got %v", result)
	}
	if result := getEnvAsDurationOrDefault("NONEXISTENT", time.Second); result != time.Second {
		t.Errorf("getEnvAsDurationOrDefault: expected 1s, got %v", result)
	}
	os.Setenv("TEST_DURATION", "invalid")
	if result := getEnvAsDurationOrDefault("TEST_DURATION", time.Second); result != time.Second {
		t.Errorf("getEnvAsDurationOrDefault with invalid value: expected 1s, got %v", result)
	}
	os.Unsetenv("TEST_DURATION")
}
