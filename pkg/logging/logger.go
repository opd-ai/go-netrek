// Package logging provides structured logging capabilities for the go-netrek application.
// It wraps Go's standard slog package to provide consistent logging patterns with
// correlation IDs, error context preservation, and security-conscious log formatting.
package logging

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// Logger wraps slog.Logger to provide application-specific logging functionality
// with correlation ID support and security-conscious formatting.
type Logger struct {
	*slog.Logger
}

// NewLogger creates a new Logger instance with JSON output and configurable level.
// The log level can be controlled via the NETREK_LOG_LEVEL environment variable.
// Valid levels: DEBUG, INFO, WARN, ERROR. Defaults to INFO.
func NewLogger() *Logger {
	level := getLogLevelFromEnv()
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:       level,
		ReplaceAttr: sanitizeAttributes,
	})
	return &Logger{slog.New(handler)}
}

// LogWithContext logs a message with automatic correlation ID extraction from context.
// If a correlation ID exists in the context, it will be included in the log entry.
func (l *Logger) LogWithContext(ctx context.Context, level slog.Level, msg string, args ...any) {
	// Extract correlation ID from context if present
	if correlationID := GetCorrelationID(ctx); correlationID != "" {
		args = append(args, "correlation_id", correlationID)
	}
	l.Log(ctx, level, msg, args...)
}

// Info logs an informational message with context.
func (l *Logger) Info(ctx context.Context, msg string, args ...any) {
	l.LogWithContext(ctx, slog.LevelInfo, msg, args...)
}

// Warn logs a warning message with context.
func (l *Logger) Warn(ctx context.Context, msg string, args ...any) {
	l.LogWithContext(ctx, slog.LevelWarn, msg, args...)
}

// Error logs an error message with context and proper error formatting.
func (l *Logger) Error(ctx context.Context, msg string, err error, args ...any) {
	if err != nil {
		args = append(args, "error", err.Error())
	}
	l.LogWithContext(ctx, slog.LevelError, msg, args...)
}

// Debug logs a debug message with context.
func (l *Logger) Debug(ctx context.Context, msg string, args ...any) {
	l.LogWithContext(ctx, slog.LevelDebug, msg, args...)
}

// correlationIDKey is the context key for correlation IDs
type correlationIDKey struct{}

// WithCorrelationID adds a correlation ID to the context.
// If no correlation ID is provided, a new one will be generated.
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	if correlationID == "" {
		correlationID = GenerateCorrelationID()
	}
	return context.WithValue(ctx, correlationIDKey{}, correlationID)
}

// GetCorrelationID extracts the correlation ID from the context.
// Returns empty string if no correlation ID is present.
func GetCorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value(correlationIDKey{}).(string); ok {
		return id
	}
	return ""
}

// GenerateCorrelationID creates a new random correlation ID.
func GenerateCorrelationID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// getLogLevelFromEnv determines the log level from environment variables.
func getLogLevelFromEnv() slog.Level {
	levelStr := strings.ToUpper(os.Getenv("NETREK_LOG_LEVEL"))
	switch levelStr {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// sanitizeAttributes removes or masks sensitive data from log attributes.
// This prevents accidental logging of passwords, tokens, or other sensitive information.
func sanitizeAttributes(groups []string, a slog.Attr) slog.Attr {
	key := strings.ToLower(a.Key)

	// List of sensitive keys that should be masked
	sensitiveKeys := []string{
		"password", "passwd", "pwd",
		"token", "auth", "authorization",
		"secret", "key", "private",
		"cookie", "session",
	}

	for _, sensitive := range sensitiveKeys {
		if strings.Contains(key, sensitive) {
			return slog.Attr{
				Key:   a.Key,
				Value: slog.StringValue("[REDACTED]"),
			}
		}
	}

	return a
}

// WrapError wraps an error with additional context information.
// This preserves the original error while adding descriptive context.
func WrapError(err error, context string, args ...any) error {
	if err == nil {
		return nil
	}
	if len(args) > 0 {
		context = fmt.Sprintf(context, args...)
	}
	return fmt.Errorf("%s: %w", context, err)
}
