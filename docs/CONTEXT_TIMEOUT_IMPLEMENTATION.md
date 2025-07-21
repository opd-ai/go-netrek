# Context-Based Timeout Management Implementation

This document describes the implementation of **Task 1.3: Context-Based Timeout Management** from the go-netrek production readiness roadmap.

## Overview

The implementation adds comprehensive context-based timeout management to all network operations in both the server and client components. This addresses critical security and reliability concerns by preventing resource leaks, providing proper cancellation mechanisms, and ensuring graceful handling of timeout scenarios.

## Key Features Implemented

### 1. Server-Side Context Management

- **Connection Timeout**: Each new connection has a configurable timeout for the complete handshake process
- **Read Timeout**: All message read operations use context with configured timeouts
- **Write Timeout**: All message write operations use context with configured timeouts
- **Client Context**: Each connected client has its own context that can be cancelled for cleanup

### 2. Client-Side Context Management

- **Connection Timeout**: TCP connection establishment uses `DialContext` with timeout
- **Message Timeouts**: All read/write operations have configurable timeouts
- **Graceful Cancellation**: Context cancellation properly cleans up resources

### 3. Configuration Integration

The implementation integrates with the existing configuration system:

```go
type EnvironmentConfig struct {
    ReadTimeout     time.Duration `env:"NETREK_READ_TIMEOUT"`
    WriteTimeout    time.Duration `env:"NETREK_WRITE_TIMEOUT"`
    // ... other fields
}
```

Default values:
- **Read Timeout**: 30 seconds
- **Write Timeout**: 30 seconds  
- **Connection Timeout**: 30 seconds

## Implementation Details

### Server Implementation

#### GameServer Structure Updates
```go
type GameServer struct {
    // ... existing fields
    config            *config.EnvironmentConfig
    connectionTimeout time.Duration
    readTimeout       time.Duration
    writeTimeout      time.Duration
}
```

#### Client Structure Updates
```go
type Client struct {
    // ... existing fields
    ctx        context.Context
    cancel     context.CancelFunc
}
```

#### Context-Based Operations

**Read Message with Timeout:**
```go
func (s *GameServer) readMessage(ctx context.Context, conn net.Conn) (MessageType, []byte, error) {
    // Set read deadline based on context
    if deadline, ok := ctx.Deadline(); ok {
        conn.SetReadDeadline(deadline)
    } else {
        conn.SetReadDeadline(time.Now().Add(s.readTimeout))
    }
    defer conn.SetReadDeadline(time.Time{})
    
    // Perform read in goroutine for cancellation support
    // ... implementation details
}
```

**Write Message with Timeout:**
```go
func (s *GameServer) sendMessage(ctx context.Context, conn net.Conn, msgType MessageType, msg interface{}) error {
    // Similar pattern with write deadline and context cancellation
    // ... implementation details
}
```

### Client Implementation

#### GameClient Structure Updates
```go
type GameClient struct {
    // ... existing fields
    ctx               context.Context
    cancel            context.CancelFunc
    connectionTimeout time.Duration
    readTimeout       time.Duration
    writeTimeout      time.Duration
}
```

#### Context-Based Connection
```go
func (c *GameClient) establishTCPConnection(address string) error {
    ctx, cancel := context.WithTimeout(c.ctx, c.connectionTimeout)
    defer cancel()
    
    dialer := &net.Dialer{}
    conn, err := dialer.DialContext(ctx, "tcp", address)
    // ... error handling
}
```

## Benefits Achieved

### 1. Security Improvements
- **DoS Prevention**: Network operations cannot hang indefinitely
- **Resource Protection**: Timeouts prevent resource exhaustion
- **Attack Mitigation**: Malicious clients cannot consume server resources indefinitely

### 2. Reliability Improvements
- **Graceful Degradation**: Operations fail fast with clear error messages
- **Resource Cleanup**: Context cancellation ensures proper cleanup
- **Predictable Behavior**: All operations have defined maximum duration

### 3. Operational Excellence
- **Monitoring**: Timeout errors can be monitored and alerted
- **Debugging**: Context provides clear error context for troubleshooting
- **Configuration**: Timeouts are configurable per environment

## Error Handling

The implementation provides comprehensive error handling:

```go
// Context deadline exceeded
if err == context.DeadlineExceeded {
    log.Printf("Operation timed out after %v", timeout)
    return ctx.Err()
}

// Context cancelled
if err == context.Canceled {
    log.Printf("Operation was cancelled")
    return ctx.Err()
}
```

## Testing

Comprehensive tests verify:
- Timeout behavior under normal and stress conditions
- Context cancellation propagation
- Resource cleanup after timeouts
- Error message accuracy

Example test scenarios:
- Read timeout exceeded
- Write timeout exceeded  
- Context cancellation during operation
- Message size validation with timeouts

## Configuration

Environment variables for timeout configuration:

```bash
# Network timeouts
NETREK_READ_TIMEOUT=30s      # Maximum time for read operations
NETREK_WRITE_TIMEOUT=30s     # Maximum time for write operations

# Server configuration
NETREK_SERVER_ADDR=localhost
NETREK_SERVER_PORT=4566
```

## Backward Compatibility

The implementation maintains full backward compatibility:
- All existing APIs continue to work
- Default timeout values are conservative (30 seconds)
- Configuration is optional with sensible defaults

## Performance Impact

The implementation has minimal performance impact:
- Context overhead is negligible
- Goroutine usage is controlled and short-lived
- Network operations are unchanged except for timeout handling

## Next Steps

This implementation satisfies **Task 1.3** requirements and enables:
- **Task 1.4**: Structured Logging (can use context for correlation IDs)
- **Phase 2 Tasks**: Circuit breakers and connection pooling
- **Production Deployment**: With proper timeout handling

## Acceptance Criteria Met

✅ All operations have configurable timeouts  
✅ Proper context cancellation prevents resource leaks  
✅ Graceful handling of timeout scenarios  
✅ No blocking operations without timeouts  
✅ Connection cleanup on context cancellation

The implementation successfully transforms the network layer from basic blocking operations to production-ready, context-aware communication with comprehensive timeout management.
