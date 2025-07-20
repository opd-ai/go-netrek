# Configuration Security and Management

This document describes the enhanced configuration security system implemented in go-netrek to replace hardcoded values with secure, environment-based configuration.

## Overview

The configuration system now supports environment variable overrides with validation to eliminate hardcoded values and improve security. All configuration values can be set via:

1. Environment variables (highest priority)
2. Configuration files (medium priority) 
3. Secure defaults (lowest priority)

## Environment Variables

### Network Configuration

| Variable | Type | Default | Valid Range | Description |
|----------|------|---------|-------------|-------------|
| `NETREK_SERVER_ADDR` | string | `localhost` | Any valid hostname/IP | Server address for binding or connecting |
| `NETREK_SERVER_PORT` | int | `4566` | 1024-65535 | Server port number |
| `NETREK_MAX_CLIENTS` | int | `32` | 1-1000 | Maximum concurrent client connections |
| `NETREK_READ_TIMEOUT` | duration | `30s` | 1s-1m | Network read timeout |
| `NETREK_WRITE_TIMEOUT` | duration | `30s` | 1s-1m | Network write timeout |

### Game Configuration

| Variable | Type | Default | Valid Range | Description |
|----------|------|---------|-------------|-------------|
| `NETREK_UPDATE_RATE` | int | `20` | 1-100 | Game update frequency (Hz) |
| `NETREK_TICKS_PER_STATE` | int | `3` | 1-10 | Ticks between full state updates |
| `NETREK_USE_PARTIAL_STATE` | bool | `true` | true/false | Enable partial state updates |
| `NETREK_WORLD_SIZE` | float64 | `10000.0` | 1000.0-100000.0 | Game world size |

## Usage Examples

### Development Environment

```bash
export NETREK_SERVER_ADDR=localhost
export NETREK_SERVER_PORT=4566
export NETREK_MAX_CLIENTS=8
export NETREK_READ_TIMEOUT=60s
export NETREK_WRITE_TIMEOUT=60s
```

### Production Environment

```bash
export NETREK_SERVER_ADDR=0.0.0.0
export NETREK_SERVER_PORT=8080
export NETREK_MAX_CLIENTS=100
export NETREK_READ_TIMEOUT=15s
export NETREK_WRITE_TIMEOUT=15s
export NETREK_UPDATE_RATE=30
```

### Using Environment File

Copy `.env.template` to `.env` and modify values:

```bash
cp .env.template .env
# Edit .env with your configuration
source .env
```

## Server Configuration

The server application now requires explicit configuration and will fail with a clear error message if the server address is not configured:

```bash
# Server will fail without configuration
./server

# Server with environment configuration
export NETREK_SERVER_ADDR=localhost
export NETREK_SERVER_PORT=4566
./server

# Server with config file override
./server -config=production.json
```

## Client Configuration

The client application follows this precedence order:

1. Command line argument (`-server`)
2. Environment variables (`NETREK_SERVER_ADDR`, `NETREK_SERVER_PORT`)
3. Configuration file
4. Default fallback (`localhost:4566`)

```bash
# Using command line (highest priority)
./client -server=192.168.1.100:8080

# Using environment variables
export NETREK_SERVER_ADDR=192.168.1.100
export NETREK_SERVER_PORT=8080
./client

# Using config file
./client -config=client.json
```

## Validation

All configuration values are validated at startup with clear error messages:

- **Port validation**: Must be between 1024-65535
- **Timeout validation**: Must be between 1s-1m
- **Client limits**: Must be between 1-1000
- **Update rate**: Must be between 1-100 Hz
- **World size**: Must be between 1000.0-100000.0

Example validation error:
```
Failed to apply environment configuration: invalid configuration: validation failed for field 'ServerPort' with value '70000': server port must be between 1024 and 65535
```

## Security Benefits

1. **No hardcoded credentials**: All sensitive values externalized
2. **Environment-specific configuration**: Different settings per deployment
3. **Validation**: Prevents invalid configuration at startup
4. **Secure defaults**: Production-safe fallback values
5. **Clear error messages**: Easy debugging of configuration issues

## Migration from Old Configuration

Existing JSON configuration files continue to work, but environment variables take precedence:

```go
// Old way (still works)
gameConfig.NetworkConfig.ServerAddress = "localhost:4566"

// New way (recommended)
export NETREK_SERVER_ADDR=localhost
export NETREK_SERVER_PORT=4566
```

## Testing

The configuration system includes comprehensive unit tests covering:

- Default value loading
- Environment variable overrides
- Validation error cases
- Helper function behavior
- Game config integration

Run tests with:
```bash
go test ./pkg/config/...
```
