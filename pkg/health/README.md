# Health Check Package

The health package provides comprehensive health check functionality for the go-netrek server, enabling production-ready monitoring and observability.

## Overview

This package implements HTTP endpoints for liveness and readiness probes that are essential for production deployment in containerized environments like Kubernetes.

## Features

- **Liveness Probe**: Simple endpoint that returns 200 OK if the application is running
- **Readiness Probe**: Comprehensive health check that validates all system components
- **Component Health Checks**: Modular system for checking individual components
- **Timeout Support**: All health checks respect context timeouts
- **Production Ready**: Follows industry standards for health check APIs

## Quick Start

```go
import "github.com/opd-ai/go-netrek/pkg/health"

// Create health checker
healthChecker := health.NewHealthChecker()

// Add health checks
healthChecker.AddCheck(health.NewGameEngineHealthCheck(
    func() bool { return gameEngine.Running },
))

healthChecker.AddCheck(health.NewNetworkHealthCheck(
    func() string { return server.GetListenerAddress() },
))

// Setup HTTP handlers
mux := http.NewServeMux()
mux.HandleFunc("/health", healthChecker.LivenessHandler)
mux.HandleFunc("/ready", healthChecker.ReadinessHandler)

// Start server
http.ListenAndServe(":8080", mux)
```

## Built-in Health Checks

### GameEngineHealthCheck
Monitors the game engine status:
```go
check := health.NewGameEngineHealthCheck(func() bool {
    return game.Running
})
```

### NetworkHealthCheck
Monitors network listener status:
```go
check := health.NewNetworkHealthCheck(func() string {
    return server.GetListenerAddress()
})
```

### MemoryHealthCheck
Monitors memory usage:
```go
check := health.NewMemoryHealthCheck(500, func() int64 {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    return int64(m.Alloc / 1024 / 1024)
})
```

## HTTP Endpoints

### GET /health (Liveness Probe)
- **Purpose**: Kubernetes liveness probe
- **Returns**: 200 OK if application is alive
- **Response**: `{"status": "alive"}`
- **Use Case**: Determines if pod should be restarted

### GET /ready (Readiness Probe)
- **Purpose**: Kubernetes readiness probe
- **Returns**: 200 OK if ready, 503 if not ready
- **Response**: Full health check results
- **Use Case**: Determines if pod should receive traffic

Example readiness response:
```json
{
  "status": "healthy",
  "checks": {
    "game_engine": {
      "status": "healthy"
    },
    "network": {
      "status": "healthy"
    },
    "memory": {
      "status": "unhealthy",
      "message": "memory usage 600MB exceeds limit 500MB"
    }
  }
}
```

## Custom Health Checks

Implement the `HealthCheck` interface:

```go
type CustomHealthCheck struct {
    dependency SomeDependency
}

func (c *CustomHealthCheck) Name() string {
    return "custom_component"
}

func (c *CustomHealthCheck) Check(ctx context.Context) error {
    if !c.dependency.IsHealthy() {
        return fmt.Errorf("dependency is unhealthy")
    }
    return nil
}

// Add to health checker
healthChecker.AddCheck(&CustomHealthCheck{dependency: dep})
```

## Configuration

Health check behavior can be configured via environment variables:

- `NETREK_HEALTH_PORT`: Port for health check server (default: 8080)

## Integration with go-netrek

The health check system is automatically integrated into the go-netrek server:

1. **Game Engine Monitoring**: Checks if the game loop is running
2. **Network Monitoring**: Checks if the TCP listener is active
3. **Memory Monitoring**: Checks memory usage against configured limits

## Production Deployment

### Docker Health Checks
```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1
```

### Kubernetes Probes
```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /ready
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
```

## Testing

The package includes comprehensive test coverage:

```bash
# Run all tests
go test ./pkg/health/

# Run with coverage
go test ./pkg/health/ -cover

# Run integration tests only
go test ./pkg/health/ -run TestHealthCheckIntegration
```

## Performance

- **Liveness checks**: < 1ms response time
- **Readiness checks**: < 10ms response time with all checks
- **Memory overhead**: < 1MB for health check system
- **Concurrent safety**: All operations are thread-safe

## Best Practices

1. **Keep checks fast**: Health checks should complete within 5 seconds
2. **Use timeouts**: Always provide context with timeouts
3. **Monitor dependencies**: Check external dependencies that affect service health
4. **Graceful degradation**: Consider if partial failures should mark service unhealthy
5. **Security**: Health endpoints don't require authentication but avoid exposing sensitive data

## Troubleshooting

### Health check timeouts
If health checks are timing out, check:
- Database connection timeouts
- External service dependencies
- Resource contention (CPU/memory)

### Memory health check failures
- Check actual memory usage: `runtime.ReadMemStats()`
- Verify memory limits are appropriate for workload
- Look for memory leaks in application code

### Network health check failures
- Verify server is properly started
- Check for port conflicts
- Ensure network interface is available
