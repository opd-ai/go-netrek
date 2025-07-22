// pkg/network/client.go
package network

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/opd-ai/go-netrek/pkg/config"
	"github.com/opd-ai/go-netrek/pkg/engine"
	"github.com/opd-ai/go-netrek/pkg/entity"
	"github.com/opd-ai/go-netrek/pkg/event"
)

// GameClient handles network communication with the server
type GameClient struct {
	conn                 net.Conn
	clientID             entity.ID
	playerID             entity.ID
	serverAddress        string
	connected            bool
	receivedStates       chan *engine.GameState
	eventBus             *event.Bus
	mu                   sync.Mutex
	latency              time.Duration
	lastPingTime         time.Time
	pingInterval         time.Duration
	reconnectDelay       time.Duration
	reconnectAttempts    int
	maxReconnectAttempts int
	DesiredShipClass     entity.ShipClass

	// Context and timeout support
	ctx               context.Context
	cancel            context.CancelFunc
	connectionTimeout time.Duration
	readTimeout       time.Duration
	writeTimeout      time.Duration
}

// NewGameClient creates a new game client
func NewGameClient(eventBus *event.Bus) *GameClient {
	// Load environment configuration for timeouts
	envConfig, err := config.LoadConfigFromEnv()
	if err != nil {
		// Use defaults if config loading fails
		envConfig = &config.EnvironmentConfig{
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		}
	}

	return &GameClient{
		receivedStates:       make(chan *engine.GameState, 10),
		eventBus:             eventBus,
		pingInterval:         time.Second * 5,
		reconnectDelay:       time.Second * 3,
		maxReconnectAttempts: 5,
		connectionTimeout:    30 * time.Second,
		readTimeout:          envConfig.ReadTimeout,
		writeTimeout:         envConfig.WriteTimeout,
	}
}

func (c *GameClient) RequestShipClass(class entity.ShipClass) error {
	c.DesiredShipClass = class

	// Send request to server
	request := struct {
		ShipClass entity.ShipClass `json:"shipClass"`
	}{
		ShipClass: class,
	}

	return c.sendMessage(RequestShipClass, request)
}

// Connect connects to the game server
func (c *GameClient) Connect(address, playerName string, teamID int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create context for connection operations
	c.ctx, c.cancel = context.WithCancel(context.Background())

	c.prepareConnection(address)

	if err := c.establishTCPConnection(address); err != nil {
		return err
	}

	if err := c.performHandshake(playerName, teamID); err != nil {
		return err
	}

	c.startBackgroundProcesses()
	return nil
}

// prepareConnection closes any existing connection and prepares for a new one.
func (c *GameClient) prepareConnection(address string) {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.connected = false
	c.serverAddress = address
}

// establishTCPConnection creates a TCP connection to the server.
func (c *GameClient) establishTCPConnection(address string) error {
	// Create context with connection timeout
	ctx, cancel := context.WithTimeout(c.ctx, c.connectionTimeout)
	defer cancel()

	// Use DialContext for timeout support
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	c.conn = conn
	return nil
}

// performHandshake sends a connect request and processes the server's response.
func (c *GameClient) performHandshake(playerName string, teamID int) error {
	if err := c.sendConnectRequest(playerName, teamID); err != nil {
		return err
	}

	if err := c.processConnectResponse(); err != nil {
		return err
	}

	return nil
}

// sendConnectRequest creates and sends the initial connection request to the server.
func (c *GameClient) sendConnectRequest(playerName string, teamID int) error {
	connectReq := struct {
		PlayerName string `json:"playerName"`
		TeamID     int    `json:"teamID"`
	}{
		PlayerName: playerName,
		TeamID:     teamID,
	}

	if err := c.sendMessage(ConnectRequest, connectReq); err != nil {
		c.cleanupConnection()
		return fmt.Errorf("failed to send connect request: %w", err)
	}

	return nil
}

// processConnectResponse reads and validates the server's connection response.
func (c *GameClient) processConnectResponse() error {
	// Use connection timeout for reading response
	ctx, cancel := context.WithTimeout(c.ctx, c.connectionTimeout)
	defer cancel()

	msgType, data, err := c.readMessage(ctx)
	if err != nil {
		c.cleanupConnection()
		return fmt.Errorf("failed to read connect response: %w", err)
	}

	if msgType != ConnectResponse {
		c.cleanupConnection()
		return fmt.Errorf("unexpected response type: %d", msgType)
	}

	if err := c.parseAndValidateResponse(data); err != nil {
		return err
	}

	return nil
}

// parseAndValidateResponse parses the connection response and updates client state.
func (c *GameClient) parseAndValidateResponse(data []byte) error {
	var connectResp struct {
		Success  bool      `json:"success"`
		Error    string    `json:"error"`
		PlayerID entity.ID `json:"playerID"`
		ClientID entity.ID `json:"clientID"`
	}

	if err := json.Unmarshal(data, &connectResp); err != nil {
		c.cleanupConnection()
		return fmt.Errorf("failed to parse connect response: %w", err)
	}

	if !connectResp.Success {
		c.cleanupConnection()
		return fmt.Errorf("server rejected connection: %s", connectResp.Error)
	}

	c.playerID = connectResp.PlayerID
	c.clientID = connectResp.ClientID
	c.connected = true

	return nil
}

// startBackgroundProcesses initiates the message and ping handling goroutines.
func (c *GameClient) startBackgroundProcesses() {
	go c.messageLoop()
	go c.pingLoop()
}

// cleanupConnection safely closes the connection and resets state (must be called with lock held)
func (c *GameClient) cleanupConnection() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.connected = false

	// Cancel context to stop any ongoing operations
	if c.cancel != nil {
		c.cancel()
		c.cancel = nil
	}
}

// Disconnect disconnects from the game server
func (c *GameClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	// Send disconnect notification with short timeout
	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	c.sendMessageWithContext(ctx, DisconnectNotification, nil)
	cancel()

	// Clean up connection
	c.cleanupConnection()
	return nil
}

// SendInput sends player input to the server
func (c *GameClient) SendInput(thrust, turnLeft, turnRight bool, fireWeapon int,
	beamDown, beamUp bool, beamAmount int, targetID entity.ID,
) error {
	if !c.connected {
		return errors.New("not connected")
	}

	input := struct {
		Thrust     bool      `json:"thrust"`
		TurnLeft   bool      `json:"turnLeft"`
		TurnRight  bool      `json:"turnRight"`
		FireWeapon int       `json:"fireWeapon"`
		BeamDown   bool      `json:"beamDown"`
		BeamUp     bool      `json:"beamUp"`
		BeamAmount int       `json:"beamAmount"`
		TargetID   entity.ID `json:"targetID"`
	}{
		Thrust:     thrust,
		TurnLeft:   turnLeft,
		TurnRight:  turnRight,
		FireWeapon: fireWeapon,
		BeamDown:   beamDown,
		BeamUp:     beamUp,
		BeamAmount: beamAmount,
		TargetID:   targetID,
	}

	return c.sendMessage(PlayerInput, input)
}

// SendChatMessage sends a chat message to the server
func (c *GameClient) SendChatMessage(message string) error {
	if !c.connected {
		return errors.New("not connected")
	}

	chatMsg := struct {
		Message string `json:"message"`
	}{
		Message: message,
	}

	return c.sendMessage(ChatMessage, chatMsg)
}

// GetLatency returns the current latency to the server
func (c *GameClient) GetLatency() time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.latency
}

// GetGameStateChannel returns the channel for receiving game states
func (c *GameClient) GetGameStateChannel() <-chan *engine.GameState {
	return c.receivedStates
}

// messageLoop handles incoming messages from the server
func (c *GameClient) messageLoop() {
	for c.connected {
		// Create context with read timeout for each message
		ctx, cancel := context.WithTimeout(c.ctx, c.readTimeout)

		msgType, data, err := c.readMessage(ctx)
		cancel() // Clean up timeout context

		if err != nil {
			if c.connected && err != context.DeadlineExceeded && err != context.Canceled {
				c.handleDisconnect(err)
			}
			return
		}

		// Process message based on type
		switch msgType {
		case GameStateUpdate:
			c.handleGameStateUpdate(data)

		case ChatMessage:
			c.handleChatMessage(data)

		case PingResponse:
			c.handlePingResponse(data)

		default:
			// Ignore unknown message types
		}
	}
}

// handleGameStateUpdate processes a game state update
func (c *GameClient) handleGameStateUpdate(data []byte) {
	var gameState engine.GameState
	if err := json.Unmarshal(data, &gameState); err != nil {
		return
	}

	// Send game state to channel, non-blocking
	select {
	case c.receivedStates <- &gameState:
		// State sent successfully
	default:
		// Channel full, drop the state
	}
}

// handleChatMessage processes a chat message
func (c *GameClient) handleChatMessage(data []byte) {
	var chatMsg struct {
		SenderID   entity.ID `json:"senderID"`
		SenderName string    `json:"senderName"`
		TeamID     int       `json:"teamID"`
		Message    string    `json:"message"`
	}

	if err := json.Unmarshal(data, &chatMsg); err != nil {
		return
	}

	// Create and publish chat event
	chatEvent := &ChatEvent{
		BaseEvent: event.BaseEvent{
			EventType: ChatMessageReceived,
			Source:    c,
		},
		SenderID:   chatMsg.SenderID,
		SenderName: chatMsg.SenderName,
		TeamID:     chatMsg.TeamID,
		Message:    chatMsg.Message,
	}

	c.eventBus.Publish(chatEvent)
}

// handlePingResponse processes a ping response
func (c *GameClient) handlePingResponse(data []byte) {
	var pingTime time.Time
	if err := json.Unmarshal(data, &pingTime); err != nil {
		return
	}

	// Calculate latency
	c.mu.Lock()
	c.latency = time.Since(pingTime)
	c.mu.Unlock()
}

// pingLoop periodically sends ping requests to the server
func (c *GameClient) pingLoop() {
	ticker := time.NewTicker(c.pingInterval)
	defer ticker.Stop()

	for c.connected {
		<-ticker.C

		// Send ping request with current time
		c.mu.Lock()
		c.lastPingTime = time.Now()
		c.mu.Unlock()

		c.sendMessage(PingRequest, c.lastPingTime)
	}
}

// handleDisconnect handles an unexpected disconnection
func (c *GameClient) handleDisconnect(err error) {
	c.mu.Lock()
	wasConnected := c.connected
	c.connected = false
	c.mu.Unlock()

	if !wasConnected {
		return
	}

	// Publish disconnect event
	disconnectEvent := &event.BaseEvent{
		EventType: ClientDisconnected,
		Source:    c,
	}
	c.eventBus.Publish(disconnectEvent)

	// Attempt to reconnect
	go c.attemptReconnect()
}

// attemptReconnect tries to reconnect to the server
func (c *GameClient) attemptReconnect() {
	c.reconnectAttempts = 0

	for c.reconnectAttempts < c.maxReconnectAttempts {
		c.reconnectAttempts++

		// Wait before attempting reconnect
		time.Sleep(c.reconnectDelay)

		// Try to reconnect
		err := c.Connect(c.serverAddress, "", 0) // Would need to store initial connection info
		if err == nil {
			// Reconnected successfully
			reconnectEvent := &event.BaseEvent{
				EventType: ClientReconnected,
				Source:    c,
			}
			c.eventBus.Publish(reconnectEvent)
			return
		}
	}

	// Failed to reconnect after max attempts
	reconnectFailedEvent := &event.BaseEvent{
		EventType: ClientReconnectFailed,
		Source:    c,
	}
	c.eventBus.Publish(reconnectFailedEvent)
}

// readMessage reads a message from the server
// readMessage reads a message from the server with context timeout support
func (c *GameClient) readMessage(ctx context.Context) (MessageType, []byte, error) {
	c.setReadDeadline(ctx)
	defer c.conn.SetReadDeadline(time.Time{}) // Clear deadline

	resultChan := make(chan readResult, 1)

	// Start read operation in separate goroutine
	go c.executeRead(resultChan)

	// Wait for completion or context cancellation
	return c.waitForReadCompletion(ctx, resultChan)
}

// readResult contains the result of a read operation
type readResult struct {
	msgType MessageType
	data    []byte
	err     error
}

// setReadDeadline configures the read timeout based on context or fallback
func (c *GameClient) setReadDeadline(ctx context.Context) {
	if deadline, ok := ctx.Deadline(); ok {
		c.conn.SetReadDeadline(deadline)
	} else {
		// Fallback to configured timeout
		c.conn.SetReadDeadline(time.Now().Add(c.readTimeout))
	}
}

// executeRead performs the actual read operation with panic recovery
func (c *GameClient) executeRead(resultChan chan readResult) {
	defer func() {
		if r := recover(); r != nil {
			resultChan <- readResult{err: fmt.Errorf("panic during read: %v", r)}
		}
	}()

	msgType, data, err := c.readMessageData()
	resultChan <- readResult{msgType: msgType, data: data, err: err}
}

// readMessageData reads the message type, length, and data from the connection
func (c *GameClient) readMessageData() (MessageType, []byte, error) {
	// Read message type
	var msgType MessageType
	if err := binary.Read(c.conn, binary.BigEndian, &msgType); err != nil {
		return 0, nil, err
	}

	// Read message length
	var msgLen uint16
	if err := binary.Read(c.conn, binary.BigEndian, &msgLen); err != nil {
		return 0, nil, err
	}

	// Read message data
	data := make([]byte, msgLen)
	if _, err := io.ReadFull(c.conn, data); err != nil {
		return 0, nil, err
	}

	return msgType, data, nil
}

// waitForReadCompletion waits for read completion or handles context cancellation
func (c *GameClient) waitForReadCompletion(ctx context.Context, resultChan chan readResult) (MessageType, []byte, error) {
	select {
	case result := <-resultChan:
		return result.msgType, result.data, result.err
	case <-ctx.Done():
		// Force connection close on timeout
		c.conn.Close()
		return 0, nil, ctx.Err()
	}
}

// sendMessage sends a message to the server with context timeout support
func (c *GameClient) sendMessage(msgType MessageType, msg interface{}) error {
	ctx := c.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return c.sendMessageWithContext(ctx, msgType, msg)
}

// sendMessageWithContext sends a message to the server with explicit context
func (c *GameClient) sendMessageWithContext(ctx context.Context, msgType MessageType, msg interface{}) error {
	data, err := c.prepareMessageData(msg)
	if err != nil {
		return err
	}

	return c.sendPreparedMessage(ctx, msgType, data)
}

// prepareMessageData serializes and validates the message data
func (c *GameClient) prepareMessageData(msg interface{}) ([]byte, error) {
	data, err := c.serializeMessage(msg)
	if err != nil {
		return nil, err
	}

	if err := c.validateMessageSize(data); err != nil {
		return nil, err
	}

	return data, nil
}

// sendPreparedMessage sends already serialized data to the server with proper synchronization
func (c *GameClient) sendPreparedMessage(ctx context.Context, msgType MessageType, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.validateConnection(); err != nil {
		return err
	}

	c.setWriteDeadline(ctx)
	defer c.conn.SetWriteDeadline(time.Time{}) // Clear deadline

	return c.performAsyncWrite(ctx, msgType, data)
}

// serializeMessage serializes the message payload to JSON bytes
func (c *GameClient) serializeMessage(msg interface{}) ([]byte, error) {
	if msg == nil {
		return []byte{}, nil
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}

// validateMessageSize checks if the message size is within allowed limits
func (c *GameClient) validateMessageSize(data []byte) error {
	if len(data) > 65535 {
		return errors.New("message too large")
	}
	return nil
}

// validateConnection ensures the client is connected before sending
func (c *GameClient) validateConnection() error {
	if !c.connected {
		return errors.New("not connected")
	}
	return nil
}

// setWriteDeadline configures the write timeout based on context or fallback
func (c *GameClient) setWriteDeadline(ctx context.Context) {
	if deadline, ok := ctx.Deadline(); ok {
		c.conn.SetWriteDeadline(deadline)
	} else {
		// Fallback to configured timeout
		c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
	}
}

// performAsyncWrite executes the write operation in a goroutine with context cancellation
func (c *GameClient) performAsyncWrite(ctx context.Context, msgType MessageType, data []byte) error {
	resultChan := make(chan error, 1)

	// Start write operation in separate goroutine
	go c.executeWrite(resultChan, msgType, data)

	// Wait for completion or context cancellation
	return c.waitForWriteCompletion(ctx, resultChan)
}

// executeWrite performs the actual write operation with panic recovery
func (c *GameClient) executeWrite(resultChan chan error, msgType MessageType, data []byte) {
	defer func() {
		if r := recover(); r != nil {
			resultChan <- fmt.Errorf("panic during write: %v", r)
		}
	}()

	err := c.writeMessageData(msgType, data)
	resultChan <- err
}

// waitForWriteCompletion waits for write completion or handles context cancellation
func (c *GameClient) waitForWriteCompletion(ctx context.Context, resultChan chan error) error {
	select {
	case err := <-resultChan:
		return err
	case <-ctx.Done():
		// Force connection close on timeout
		c.conn.Close()
		return ctx.Err()
	}
}

// writeMessageData writes the message type, length, and data to the connection
func (c *GameClient) writeMessageData(msgType MessageType, data []byte) error {
	// Write message type
	if err := binary.Write(c.conn, binary.BigEndian, msgType); err != nil {
		return err
	}

	// Write message length
	msgLen := uint16(len(data))
	if err := binary.Write(c.conn, binary.BigEndian, msgLen); err != nil {
		return err
	}

	// Write message data
	if _, err := c.conn.Write(data); err != nil {
		return err
	}

	return nil
}

// Client event types
const (
	ChatMessageReceived   event.Type = "chat_message_received"
	ClientDisconnected    event.Type = "client_disconnected"
	ClientReconnected     event.Type = "client_reconnected"
	ClientReconnectFailed event.Type = "client_reconnect_failed"
)

// ChatEvent contains information about a received chat message
type ChatEvent struct {
	event.BaseEvent
	SenderID   entity.ID
	SenderName string
	TeamID     int
	Message    string
}
