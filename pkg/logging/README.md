# Logging Package

The logging package provides structured logging capabilities for the go-netrek application, building on Go's standard `log/slog` package with additional features for production-ready applications.

## Features

- **Structured JSON Logging**: All log entries are formatted as JSON for easy parsing and analysis
- **Correlation ID Support**: Automatic tracking of requests through correlation IDs in context
- **Security-Conscious**: Automatic redaction of sensitive data like passwords and tokens
- **Configurable Log Levels**: Environment-based log level configuration
- **Error Context Preservation**: Proper error wrapping that maintains error chains
- **Performance Optimized**: Built on Go's efficient slog package

## Usage

### Basic Logging

```go
import "github.com/opd-ai/go-netrek/pkg/logging"

logger := logging.NewLogger()
ctx := context.Background()

// Different log levels
logger.Info(ctx, "Server started", "port", 8080)
logger.Warn(ctx, "Connection limit approaching", "current", 95, "max", 100)
logger.Error(ctx, "Failed to process request", err, "request_id", "123")
logger.Debug(ctx, "Processing message", "type", "player_input")
```

### Correlation IDs

```go
// Add correlation ID to context
ctx = logging.WithCorrelationID(ctx, "request-123")

// Or auto-generate one
ctx = logging.WithCorrelationID(ctx, "")

// All subsequent logs will include the correlation ID
logger.Info(ctx, "Processing request") // Will include "correlation_id": "request-123"
```

### Error Handling

```go
// Wrap errors with context
err := doSomething()
if err != nil {
    wrappedErr := logging.WrapError(err, "failed to process player input for player %s", playerName)
    logger.Error(ctx, "Request processing failed", wrappedErr)
    return wrappedErr
}
```

## Configuration

The logging package can be configured via environment variables:

- `NETREK_LOG_LEVEL`: Sets the minimum log level (DEBUG, INFO, WARN, ERROR). Defaults to INFO.

### Example

```bash
export NETREK_LOG_LEVEL=DEBUG
./go-netrek-server
```

## Security Features

The logging package automatically redacts sensitive information from log entries:

- Password fields (password, passwd, pwd)
- Authentication tokens (token, auth, authorization)
- Secrets and keys (secret, key, private)
- Session data (cookie, session)

```go
logger.Info(ctx, "User authenticated", 
    "username", "john",
    "password", "secret123",  // Will be logged as "[REDACTED]"
    "token", "bearer-token",  // Will be logged as "[REDACTED]"
)
```

## Log Format

All logs are output as JSON with the following structure:

```json
{
    "time": "2025-07-20T10:30:00Z",
    "level": "INFO",
    "msg": "Server started",
    "correlation_id": "abc123def456",
    "port": 8080,
    "component": "server"
}
```

## Best Practices

1. **Always pass context**: Use the context-aware logging methods to enable correlation ID tracking
2. **Include relevant context**: Add key-value pairs that help with debugging
3. **Use appropriate log levels**: 
   - DEBUG: Detailed information for development
   - INFO: General operational information
   - WARN: Something unexpected but not necessarily an error
   - ERROR: Error conditions that need attention
4. **Don't log sensitive data**: The package automatically redacts common sensitive fields, but be cautious
5. **Wrap errors with context**: Use `WrapError` to add context while preserving error chains

## Testing

The package includes comprehensive tests covering:

- Log level configuration
- Correlation ID management
- Sensitive data redaction
- Error wrapping
- All logging methods

Run tests with:

```bash
go test ./pkg/logging/...
```

## Integration with go-netrek

This logging package is designed specifically for the go-netrek application and integrates with:

- Network layer for request tracking
- Game engine for performance monitoring  
- Configuration system for environment-based setup
- Error handling throughout the application
