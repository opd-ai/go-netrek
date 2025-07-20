# Validation Package

The validation package provides input validation and rate limiting capabilities for the go-netrek game server. It implements security measures to prevent malformed data attacks, XSS vulnerabilities, and resource exhaustion.

## Features

### Input Validation
- **Message Size Limits**: Enforces maximum message size of 64KB
- **JSON Format Validation**: Ensures all messages are valid JSON
- **Player Name Sanitization**: Validates and sanitizes player names with XSS prevention
- **Chat Message Sanitization**: Validates and sanitizes chat messages with control character filtering
- **Game Data Validation**: Validates weapon indices, beam amounts, and team IDs

### Rate Limiting
- **Per-Client Rate Limiting**: Prevents message spam with configurable limits (default: 100 messages/minute)
- **Token Bucket Algorithm**: Smooth rate limiting with burst capability
- **Automatic Cleanup**: Removes inactive client rate limit states to prevent memory leaks

### Security Features
- **XSS Prevention**: HTML entity escaping for all user-provided text
- **Control Character Filtering**: Removes potentially dangerous control characters
- **Character Set Validation**: Enforces allowed character sets for player names
- **Input Length Limits**: Prevents buffer overflow and DoS attacks

## Usage

### Basic Validation

```go
// Create validator
validator := validation.NewMessageValidator()
defer validator.Close()

// Validate raw message
err := validator.ValidateMessage(data, clientID)
if err != nil {
    // Handle validation error
}

// Validate player name
sanitizedName, err := validation.ValidatePlayerName(rawName)
if err != nil {
    // Handle invalid name
}

// Validate chat message
sanitizedMessage, err := validation.ValidateChatMessage(rawMessage)
if err != nil {
    // Handle invalid message
}
```

### Server Integration

```go
// In server initialization
server := &GameServer{
    validator: validation.NewMessageValidator(),
    // ... other fields
}

// In message handling
if err := s.validator.ValidateMessage(data, clientID); err != nil {
    log.Printf("Validation failed: %v", err)
    continue // Skip invalid message
}
```

## Configuration

### Constants

- `MaxMessageSize`: 64KB maximum message size
- `MaxPlayerNameLen`: 32 character maximum for player names
- `MaxChatMessageLen`: 256 character maximum for chat messages
- `MaxMessagesPerMin`: 100 messages per minute rate limit

### Validation Rules

#### Player Names
- Length: 1-32 characters
- Allowed characters: alphanumeric, spaces, hyphens, underscores, basic punctuation
- No control characters allowed
- HTML entities are escaped for XSS prevention

#### Chat Messages
- Length: 1-256 characters
- Control characters removed (except newlines and tabs)
- HTML entities are escaped for XSS prevention

#### Game Data
- Team IDs: 0 (Federation) or 1 (Klingon)
- Weapon indices: -1 (not firing) or 0-7 (weapon slots)
- Beam amounts: 0-1000 reasonable range

## Rate Limiting

The rate limiter uses a token bucket algorithm:

1. Each client starts with a full bucket of tokens
2. Each message consumes one token
3. Tokens are refilled over time based on the configured rate
4. Messages are rejected when no tokens are available

### Implementation Details

- **Per-client tracking**: Each client connection has its own rate limit state
- **Memory efficiency**: Inactive clients are automatically cleaned up
- **Configurable windows**: Rate limiting window is configurable (default: 1 minute)
- **Thread-safe**: All operations are protected with appropriate synchronization

## Security Considerations

### Prevented Attacks

1. **DoS via Message Spam**: Rate limiting prevents overwhelming the server
2. **DoS via Large Messages**: Size limits prevent memory exhaustion
3. **XSS Attacks**: HTML escaping prevents script injection
4. **Malformed Data**: JSON validation prevents parsing errors
5. **Control Character Injection**: Character filtering prevents terminal manipulation

### Best Practices

1. Always validate input before processing
2. Use rate limiting for all client-initiated actions
3. Log validation failures for monitoring
4. Don't expose detailed error messages to clients
5. Regularly monitor rate limiting metrics

## Testing

The package includes comprehensive test coverage:

```bash
go test ./pkg/validation/...
```

Test categories:
- Input validation edge cases
- Rate limiting behavior
- Character set validation
- HTML escaping verification
- Token bucket refill logic
