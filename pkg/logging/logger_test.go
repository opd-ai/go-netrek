package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestNewLogger(t *testing.T) {
	logger := NewLogger()
	if logger == nil {
		t.Fatal("NewLogger() returned nil")
	}
	if logger.Logger == nil {
		t.Fatal("Logger.Logger is nil")
	}
}

func TestLogLevelFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected slog.Level
	}{
		{"debug level", "DEBUG", slog.LevelDebug},
		{"info level", "INFO", slog.LevelInfo},
		{"warn level", "WARN", slog.LevelWarn},
		{"warning level", "WARNING", slog.LevelWarn},
		{"error level", "ERROR", slog.LevelError},
		{"lowercase debug", "debug", slog.LevelDebug},
		{"mixed case", "Info", slog.LevelInfo},
		{"invalid level", "INVALID", slog.LevelInfo},
		{"empty value", "", slog.LevelInfo},
	}

	originalLevel := os.Getenv("NETREK_LOG_LEVEL")
	defer os.Setenv("NETREK_LOG_LEVEL", originalLevel)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("NETREK_LOG_LEVEL", tt.envValue)
			level := getLogLevelFromEnv()
			if level != tt.expected {
				t.Errorf("getLogLevelFromEnv() = %v, want %v", level, tt.expected)
			}
		})
	}
}

func TestCorrelationID(t *testing.T) {
	t.Run("generate correlation ID", func(t *testing.T) {
		id1 := GenerateCorrelationID()
		id2 := GenerateCorrelationID()
		
		if id1 == "" {
			t.Error("GenerateCorrelationID() returned empty string")
		}
		if id2 == "" {
			t.Error("GenerateCorrelationID() returned empty string")
		}
		if id1 == id2 {
			t.Error("GenerateCorrelationID() returned duplicate IDs")
		}
		if len(id1) != 16 { // 8 bytes = 16 hex characters
			t.Errorf("GenerateCorrelationID() returned wrong length: %d", len(id1))
		}
	})

	t.Run("context with correlation ID", func(t *testing.T) {
		ctx := context.Background()
		expectedID := "test-correlation-id"
		
		ctx = WithCorrelationID(ctx, expectedID)
		actualID := GetCorrelationID(ctx)
		
		if actualID != expectedID {
			t.Errorf("GetCorrelationID() = %q, want %q", actualID, expectedID)
		}
	})

	t.Run("context without correlation ID", func(t *testing.T) {
		ctx := context.Background()
		id := GetCorrelationID(ctx)
		
		if id != "" {
			t.Errorf("GetCorrelationID() = %q, want empty string", id)
		}
	})

	t.Run("auto-generate correlation ID", func(t *testing.T) {
		ctx := context.Background()
		ctx = WithCorrelationID(ctx, "")
		
		id := GetCorrelationID(ctx)
		if id == "" {
			t.Error("WithCorrelationID() with empty string should auto-generate ID")
		}
		if len(id) != 16 {
			t.Errorf("Auto-generated correlation ID has wrong length: %d", len(id))
		}
	})
}

func TestSanitizeAttributes(t *testing.T) {
	tests := []struct {
		name     string
		attr     slog.Attr
		expected string
	}{
		{
			"password field",
			slog.String("password", "secret123"),
			"[REDACTED]",
		},
		{
			"token field",
			slog.String("auth_token", "bearer-token"),
			"[REDACTED]",
		},
		{
			"secret field",
			slog.String("api_secret", "my-secret"),
			"[REDACTED]",
		},
		{
			"normal field",
			slog.String("username", "testuser"),
			"testuser",
		},
		{
			"case insensitive password",
			slog.String("PASSWORD", "secret123"),
			"[REDACTED]",
		},
		{
			"partial match authorization",
			slog.String("authorization_header", "Bearer token"),
			"[REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeAttributes(nil, tt.attr)
			if result.Value.String() != tt.expected {
				t.Errorf("sanitizeAttributes() = %q, want %q", result.Value.String(), tt.expected)
			}
		})
	}
}

func TestLoggerMethods(t *testing.T) {
	var buf bytes.Buffer
	
	// Create a logger that writes to our buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := &Logger{slog.New(handler)}
	
	ctx := WithCorrelationID(context.Background(), "test-id-123")

	t.Run("info logging", func(t *testing.T) {
		buf.Reset()
		logger.Info(ctx, "test info message", "key", "value")
		
		var logEntry map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
			t.Fatalf("Failed to parse log JSON: %v", err)
		}
		
		if logEntry["msg"] != "test info message" {
			t.Errorf("Expected message 'test info message', got %v", logEntry["msg"])
		}
		if logEntry["level"] != "INFO" {
			t.Errorf("Expected level 'INFO', got %v", logEntry["level"])
		}
		if logEntry["correlation_id"] != "test-id-123" {
			t.Errorf("Expected correlation_id 'test-id-123', got %v", logEntry["correlation_id"])
		}
		if logEntry["key"] != "value" {
			t.Errorf("Expected key 'value', got %v", logEntry["key"])
		}
	})

	t.Run("error logging", func(t *testing.T) {
		buf.Reset()
		testErr := errors.New("test error")
		logger.Error(ctx, "test error message", testErr, "context", "test")
		
		var logEntry map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
			t.Fatalf("Failed to parse log JSON: %v", err)
		}
		
		if logEntry["msg"] != "test error message" {
			t.Errorf("Expected message 'test error message', got %v", logEntry["msg"])
		}
		if logEntry["level"] != "ERROR" {
			t.Errorf("Expected level 'ERROR', got %v", logEntry["level"])
		}
		if logEntry["error"] != "test error" {
			t.Errorf("Expected error 'test error', got %v", logEntry["error"])
		}
	})

	t.Run("debug logging", func(t *testing.T) {
		buf.Reset()
		logger.Debug(ctx, "debug message", "debug_key", "debug_value")
		
		var logEntry map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
			t.Fatalf("Failed to parse log JSON: %v", err)
		}
		
		if logEntry["level"] != "DEBUG" {
			t.Errorf("Expected level 'DEBUG', got %v", logEntry["level"])
		}
	})

	t.Run("warn logging", func(t *testing.T) {
		buf.Reset()
		logger.Warn(ctx, "warning message", "warn_key", "warn_value")
		
		var logEntry map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
			t.Fatalf("Failed to parse log JSON: %v", err)
		}
		
		if logEntry["level"] != "WARN" {
			t.Errorf("Expected level 'WARN', got %v", logEntry["level"])
		}
	})
}

func TestWrapError(t *testing.T) {
	t.Run("wrap nil error", func(t *testing.T) {
		result := WrapError(nil, "context")
		if result != nil {
			t.Errorf("WrapError(nil) should return nil, got %v", result)
		}
	})

	t.Run("wrap error with context", func(t *testing.T) {
		originalErr := errors.New("original error")
		wrapped := WrapError(originalErr, "additional context")
		
		expectedMsg := "additional context: original error"
		if wrapped.Error() != expectedMsg {
			t.Errorf("WrapError() = %q, want %q", wrapped.Error(), expectedMsg)
		}
		
		// Test that the original error is preserved
		if !errors.Is(wrapped, originalErr) {
			t.Error("WrapError() should preserve original error")
		}
	})

	t.Run("wrap error with formatted context", func(t *testing.T) {
		originalErr := errors.New("original error")
		wrapped := WrapError(originalErr, "context with %s and %d", "string", 42)
		
		expectedMsg := "context with string and 42: original error"
		if wrapped.Error() != expectedMsg {
			t.Errorf("WrapError() = %q, want %q", wrapped.Error(), expectedMsg)
		}
	})
}

func TestLogWithoutCorrelationID(t *testing.T) {
	var buf bytes.Buffer
	
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := &Logger{slog.New(handler)}
	
	ctx := context.Background() // No correlation ID
	logger.Info(ctx, "test message")
	
	logOutput := buf.String()
	if strings.Contains(logOutput, "correlation_id") {
		t.Error("Log should not contain correlation_id when none is set in context")
	}
}
