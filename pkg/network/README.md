# Network Package

The network package provides client-server networking for Go Netrek using TCP and an extensible message protocol.

## Core Types

### MessageType

```go
type MessageType byte

const (
    ConnectRequest MessageType = iota
    ConnectResponse 
    DisconnectNotification
    GameStateUpdate
    PlayerInput
    ChatMessage
    PingRequest 
    PingResponse
)
```

### Server Implementation

The `GameServer` handles multiple client connections and game state synchronization:

```go
server := network.NewGameServer(game, maxClients)
err := server.Start("localhost:4566")
```

### Client Implementation

The `GameClient` manages connection to server and message handling:

```go
client := network.NewGameClient(eventBus)
err := client.Connect("localhost:4566", "Player1", 0)
```

## Custom Transport Implementations

You can implement custom transports by satisfying the `net.Conn` and `net.Listener` interfaces:

### WebSocket Transport

```go
type WSConn struct {
    conn *websocket.Conn
    reader io.Reader
}

func (w *WSConn) Read(b []byte) (n int, err error) {
    if w.reader == nil {
        _, r, err := w.conn.NextReader()
        if err != nil {
            return 0, err
        }
        w.reader = r
    }
    n, err = w.reader.Read(b)
    if err == io.EOF {
        w.reader = nil
        err = nil
    }
    return
}

func (w *WSConn) Write(b []byte) (n int, err error) {
    err = w.conn.WriteMessage(websocket.BinaryMessage, b)
    if err != nil {
        return 0, err
    }
    return len(b), nil
}
```

### UDP Transport

```go
type UDPConn struct {
    conn *net.UDPConn
    addr *net.UDPAddr
}

func (u *UDPConn) Read(b []byte) (n int, err error) {
    return u.conn.Read(b)
}

func (u *UDPConn) Write(b []byte) (n int, err error) {
    return u.conn.WriteToUDP(b, u.addr)
}
```

### In-Memory Transport (for testing)

```go
type PipeConn struct {
    *io.PipeReader
    *io.PipeWriter
}

func NewPipeConnection() (client, server net.Conn) {
    cr, sw := io.Pipe()
    sr, cw := io.Pipe()
    
    return &PipeConn{cr, cw}, &PipeConn{sr, sw}
}
```

## Message Protocol

Messages are framed with:
1. Message type (1 byte)
2. Message length (2 bytes, big endian)
3. JSON payload (variable length)

Example message:
```
[TYPE][LENGTH][JSON PAYLOAD]
 0x01  0x0045  {"playerName": "Player1"}
```

## Usage Examples

### Basic Server

```go
game := engine.NewGame(config)
server := network.NewGameServer(game, 16)

if err := server.Start("localhost:4566"); err != nil {
    log.Fatal(err)
}
defer server.Stop()
```

### Basic Client

```go
bus := event.NewEventBus()
client := network.NewGameClient(bus)

if err := client.Connect("localhost:4566", "Player1", 0); err != nil {
    log.Fatal(err)
}
defer client.Disconnect()

// Handle game states
go func() {
    for state := range client.GetGameStateChannel() {
        // Process game state
    }
}()

// Send input
client.SendInput(true, false, true, 0, false, false, 0, 0)
```

### Chat System

```go
// Server-side broadcast
server.broadcastChatMessage(sender, []byte(`{"message":"Hello!"}`))

// Client-side send
client.SendChatMessage("Hello everyone!")

// Client-side receive
bus.Subscribe(network.ChatMessageReceived, func(e event.Event) {
    if chat, ok := e.(*network.ChatEvent); ok {
        fmt.Printf("[%s]: %s\n", chat.SenderName, chat.Message)
    }
})
```

## Best Practices

- Always check for connection errors
- Handle reconnection attempts gracefully 
- Use channels for async communication
- Implement timeouts for network operations
- Handle message size limits appropriately

## Future Improvements

- Compression for game state updates
- Partial state updates
- UDP transport for position updates 
- WebSocket transport support
- Better connection quality metrics
- Message encryption
- Protocol versioning

For more details, see the client and server implementations.