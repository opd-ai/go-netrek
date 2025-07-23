# Resource Management Package

This package provides resource management capabilities for the go-netrek server, including memory monitoring, goroutine lifecycle management, and graceful shutdown handling.

## Features

- **Memory Monitoring**: Tracks memory usage and enforces configurable limits
- **Goroutine Management**: Safe goroutine creation with leak prevention
- **Graceful Shutdown**: Clean resource cleanup with timeout handling
- **Health Checks**: Integration with health monitoring system
- **Resource Metrics**: Real-time resource usage statistics

## Configuration

Resource management is configured through environment variables:

```bash
# Memory limit in MB (default: 500)
NETREK_MAX_MEMORY_MB=500

# Maximum number of goroutines (default: 1000)
NETREK_MAX_GOROUTINES=1000

# Shutdown timeout (default: 30s)
NETREK_SHUTDOWN_TIMEOUT=30s

# Resource check interval (default: 10s)
NETREK_RESOURCE_CHECK_INTERVAL=10s
```

## Usage

### Basic Usage

```go
package main

import (
    "context"
    "time"
    
    "github.com/opd-ai/go-netrek/pkg/config"
    "github.com/opd-ai/go-netrek/pkg/resource"
)

func main() {
    // Load configuration
    envConfig, err := config.LoadConfigFromEnv()
    if err != nil {
        log.Fatal(err)
    }

    // Create resource manager
    rm := resource.NewResourceManager(envConfig)
    
    // Start monitoring
    if err := rm.Start(); err != nil {
        log.Fatal(err)
    }
    defer rm.Shutdown(context.Background())

    // Use resource manager...
}
```

### Goroutine Management

```go
// Start a managed goroutine
ctx := context.Background()
err := rm.StartGoroutine(ctx, "worker-task", func(ctx context.Context) {
    // Your work here
    time.Sleep(1 * time.Second)
})
if err != nil {
    log.Printf("Failed to start goroutine: %v", err)
}
```

### Health Monitoring

```go
// Create health check
healthCheck := resource.NewResourceHealthCheck(rm)

// Check resource health
if err := healthCheck.Check(context.Background()); err != nil {
    log.Printf("Resource health check failed: %v", err)
}
```

### Resource Statistics

```go
// Get current resource usage
stats := rm.GetResourceStats()
fmt.Printf("Goroutines: %d/%d\n", stats.GoroutineCount, stats.MaxGoroutines)
fmt.Printf("Memory: %dMB/%dMB\n", stats.MemoryUsageMB, stats.MaxMemoryMB)
```

## Integration with Game Engine

The resource manager integrates with the game engine for automatic resource tracking:

```go
// In game initialization
game := engine.NewGame(config)
err := game.InitializeResourceManager()
if err != nil {
    log.Printf("Resource manager initialization failed: %v", err)
}
```

## Health Checks

The resource manager provides health checks that can be integrated with monitoring systems:

- **Memory Check**: Fails if memory usage exceeds configured limit
- **Goroutine Check**: Warns if goroutine count exceeds 80% of limit

## Error Handling

The resource manager provides comprehensive error handling:

```go
// Goroutine limit exceeded
err := rm.StartGoroutine(ctx, "task", func(ctx context.Context) {})
if err != nil {
    // Handle gracefully - maybe queue the task or retry later
    log.Printf("Resource limit reached: %v", err)
}

// Memory limit exceeded
if err := rm.CheckMemoryUsage(); err != nil {
    // Take action - maybe trigger garbage collection or scale down
    log.Printf("Memory limit exceeded: %v", err)
}
```

## Best Practices

### Goroutine Management

1. **Use descriptive names** for goroutines to aid in debugging
2. **Respect context cancellation** in goroutine functions
3. **Handle resource limits gracefully** - queue work or retry later
4. **Avoid blocking operations** without timeout handling

```go
// Good: Respects context and has timeout
err := rm.StartGoroutine(ctx, "api-call", func(ctx context.Context) {
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    // Make API call with context
    makeAPICall(ctx)
})
```

### Memory Management

1. **Monitor memory trends** - increasing memory usage may indicate leaks
2. **Set appropriate limits** based on your deployment environment
3. **Use health checks** for automated monitoring and alerting
4. **Consider garbage collection** when approaching limits

### Shutdown Handling

1. **Always call Shutdown()** in defer or signal handlers
2. **Use appropriate timeouts** for shutdown operations
3. **Handle shutdown errors** appropriately for your use case

```go
// Good: Proper shutdown handling
func main() {
    rm := resource.NewResourceManager(config)
    rm.Start()
    
    // Setup signal handling
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    <-sigChan
    
    // Graceful shutdown with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := rm.Shutdown(ctx); err != nil {
        log.Printf("Shutdown error: %v", err)
    }
}
```

## Monitoring and Alerting

### Metrics

The resource manager provides metrics suitable for monitoring systems:

```go
stats := rm.GetResourceStats()

// Prometheus-style metrics
resourceMemoryUsage.Set(float64(stats.MemoryUsageMB))
resourceGoroutineCount.Set(float64(stats.GoroutineCount))
```

### Alerting Thresholds

Recommended alerting thresholds:

- **Memory**: Alert at 90% of limit
- **Goroutines**: Alert at 80% of limit
- **Health Check**: Alert on any health check failure

## Thread Safety

The resource manager is fully thread-safe:

- All operations can be called concurrently
- Atomic operations are used for counters
- Mutex protection for complex state changes
- Safe for use across multiple goroutines

## Performance

The resource manager is designed for minimal overhead:

- Atomic counters for high-frequency operations
- Periodic monitoring instead of per-operation checks
- Efficient memory statistics collection
- Lock-free read operations where possible

## Troubleshooting

### Common Issues

**Goroutine limit exceeded**
- Check for goroutine leaks in application code
- Increase limit if appropriate for your workload
- Review goroutine naming for identification

**Memory limit exceeded**
- Check for memory leaks
- Consider increasing limit
- Force garbage collection if appropriate

**Shutdown timeouts**
- Long-running goroutines not respecting context
- Increase shutdown timeout
- Review goroutine cancellation logic

### Debug Information

```go
// Get detailed resource statistics
stats := rm.GetResourceStats()
log.Printf("Resource Stats: %+v", stats)

// Check if resource manager is running
if rm.running {
    log.Println("Resource manager is active")
}
```
