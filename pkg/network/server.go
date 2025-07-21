// pkg/network/server.go
package network

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/opd-ai/go-netrek/pkg/config"
	"github.com/opd-ai/go-netrek/pkg/engine"
	"github.com/opd-ai/go-netrek/pkg/entity"
	"github.com/opd-ai/go-netrek/pkg/logging"
	"github.com/opd-ai/go-netrek/pkg/physics"
	"github.com/opd-ai/go-netrek/pkg/validation"
)

// MessageType defines the type of network message
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
	RequestShipClass
)

// GameServer handles network communication and game state
type GameServer struct {
	listener          net.Listener
	game              *engine.Game
	clients           map[entity.ID]*Client
	clientsLock       sync.RWMutex
	running           bool
	updateRate        time.Duration
	maxClients        int
	ticksPerState     int                          // How many game ticks between full state updates
	partialState      bool                         // Whether to send partial updates between full updates
	validator         *validation.MessageValidator // Input validation and rate limiting
	config            *config.EnvironmentConfig    // Configuration for timeouts
	connectionTimeout time.Duration                // Timeout for connection operations
	readTimeout       time.Duration                // Timeout for read operations
	writeTimeout      time.Duration                // Timeout for write operations
	logger            *logging.Logger              // Structured logger
}

// Client represents a connected client
type Client struct {
	ID         entity.ID
	Conn       net.Conn
	PlayerID   entity.ID
	PlayerName string
	TeamID     int
	Connected  bool
	LastInput  time.Time
	Latency    time.Duration
	ctx        context.Context    // Context for client operations
	cancel     context.CancelFunc // Cancel function for client context
}

// NewGameServer creates a new game server
func NewGameServer(game *engine.Game, maxClients int) *GameServer {
	nc := game.Config.NetworkConfig
	logger := logging.NewLogger()

	// Load environment configuration for timeouts
	envConfig, err := config.LoadConfigFromEnv()
	if err != nil {
		logger.Warn(context.Background(), "Failed to load environment config, using defaults",
			"error", err)
		envConfig = &config.EnvironmentConfig{
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		}
	}

	return &GameServer{
		game:              game,
		clients:           make(map[entity.ID]*Client),
		running:           false,
		updateRate:        time.Second / time.Duration(nc.UpdateRate),
		maxClients:        maxClients,
		ticksPerState:     nc.TicksPerState,
		partialState:      nc.UsePartialState,
		validator:         validation.NewMessageValidator(),
		config:            envConfig,
		connectionTimeout: 30 * time.Second, // Default connection timeout
		readTimeout:       envConfig.ReadTimeout,
		writeTimeout:      envConfig.WriteTimeout,
		logger:            logger,
	}
}

// Start starts the game server
func (s *GameServer) Start(address string) error {
	var err error
	s.listener, err = net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	s.running = true

	// Start game
	s.game.Start()

	// Start accepting connections
	go s.acceptConnections()

	// Start game update loop
	go s.gameLoop()

	ctx := context.Background()
	s.logger.Info(ctx, "Game server started",
		"address", address,
		"max_clients", s.maxClients,
		"update_rate", s.updateRate.String(),
	)
	return nil
}

// Stop stops the game server
func (s *GameServer) Stop() {
	s.running = false

	// Close all client connections
	s.clientsLock.Lock()
	for _, client := range s.clients {
		client.Conn.Close()
	}
	s.clientsLock.Unlock()

	// Close listener
	if s.listener != nil {
		s.listener.Close()
	}

	// Stop validator and rate limiter
	if s.validator != nil {
		s.validator.Close()
	}

	// Stop game
	s.game.Stop()

	ctx := context.Background()
	s.logger.Info(ctx, "Game server stopped")
}

// acceptConnections accepts new client connections
func (s *GameServer) acceptConnections() {
	for s.running {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.running {
				log.Printf("Error accepting connection: %v", err)
			}
			continue
		}

		// Check if server is full
		s.clientsLock.RLock()
		clientCount := len(s.clients)
		s.clientsLock.RUnlock()

		if clientCount >= s.maxClients {
			log.Printf("Rejecting connection, server full")
			conn.Close()
			continue
		}

		// Handle new connection
		go s.handleConnection(conn)
	}
}

// handleConnection handles a new client connection
func (s *GameServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Create context with timeout for connection operations
	ctx, cancel := context.WithTimeout(context.Background(), s.connectionTimeout)
	defer cancel()

	connectReq, err := s.readAndValidateConnectRequest(ctx, conn)
	if err != nil {
		log.Printf("Connection failed during connect request: %v", err)
		return
	}

	playerID, err := s.addPlayerToGame(conn, connectReq)
	if err != nil {
		log.Printf("Connection failed during player addition: %v", err)
		return
	}

	client := s.createAndRegisterClient(ctx, conn, playerID, connectReq)
	if client == nil {
		log.Printf("Connection failed during client creation")
		return
	}

	if err := s.sendConnectionSuccessResponse(ctx, conn, playerID, client.ID); err != nil {
		log.Printf("Connection failed during success response: %v", err)
		s.removeClient(client)
		return
	}

	s.handleClientMessages(client)
}

// readAndValidateConnectRequest reads and validates the initial connection request.
func (s *GameServer) readAndValidateConnectRequest(ctx context.Context, conn net.Conn) (*connectRequest, error) {
	msgType, data, err := s.readMessage(ctx, conn)
	if err != nil {
		log.Printf("Error reading connect request: %v", err)
		return nil, err
	}

	if msgType != ConnectRequest {
		log.Printf("Expected connect request, got %d", msgType)
		return nil, errors.New("invalid message type")
	}

	// Use client connection remote address as identifier for rate limiting
	clientID := conn.RemoteAddr().String()

	// Validate message size and format
	if err := s.validator.ValidateMessage(data, clientID); err != nil {
		log.Printf("Message validation failed: %v", err)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	var connectReq connectRequest
	if err := json.Unmarshal(data, &connectReq); err != nil {
		log.Printf("Error parsing connect request: %v", err)
		return nil, err
	}

	// Validate and sanitize player name
	sanitizedName, err := validation.ValidatePlayerName(connectReq.PlayerName)
	if err != nil {
		log.Printf("Invalid player name: %v", err)
		return nil, fmt.Errorf("invalid player name: %w", err)
	}
	connectReq.PlayerName = sanitizedName

	// Validate team ID
	if err := validation.ValidateTeamID(connectReq.TeamID); err != nil {
		log.Printf("Invalid team ID: %v", err)
		return nil, fmt.Errorf("invalid team ID: %w", err)
	}

	return &connectReq, nil
}

// addPlayerToGame adds a new player to the game and handles errors.
func (s *GameServer) addPlayerToGame(conn net.Conn, connectReq *connectRequest) (entity.ID, error) {
	playerID, err := s.game.AddPlayer(connectReq.PlayerName, connectReq.TeamID)
	if err != nil {
		log.Printf("Error adding player: %v", err)
		s.sendConnectionErrorResponse(conn, err)
		return 0, err
	}
	return playerID, nil
}

// createAndRegisterClient creates a new client and registers it with the server.
func (s *GameServer) createAndRegisterClient(ctx context.Context, conn net.Conn, playerID entity.ID, connectReq *connectRequest) *Client {
	clientID := entity.GenerateID()

	// Create context for client operations with connection timeout
	clientCtx, clientCancel := context.WithCancel(ctx)

	client := &Client{
		ID:         clientID,
		Conn:       conn,
		PlayerID:   playerID,
		PlayerName: connectReq.PlayerName,
		TeamID:     connectReq.TeamID,
		Connected:  true,
		LastInput:  time.Now(),
		ctx:        clientCtx,
		cancel:     clientCancel,
	}

	s.clientsLock.Lock()
	s.clients[clientID] = client
	s.clientsLock.Unlock()

	return client
}

// sendConnectionErrorResponse sends an error response for failed connections.
func (s *GameServer) sendConnectionErrorResponse(conn net.Conn, err error) {
	errorResp := struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}{
		Success: false,
		Error:   err.Error(),
	}

	// Use a short timeout for error responses
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if sendErr := s.sendMessage(ctx, conn, ConnectResponse, errorResp); sendErr != nil {
		log.Printf("Failed to send connection error response: %v", sendErr)
	}
}

// sendConnectionSuccessResponse sends a success response for established connections.
func (s *GameServer) sendConnectionSuccessResponse(ctx context.Context, conn net.Conn, playerID, clientID entity.ID) error {
	successResp := struct {
		Success  bool      `json:"success"`
		PlayerID entity.ID `json:"playerID"`
		ClientID entity.ID `json:"clientID"`
	}{
		Success:  true,
		PlayerID: playerID,
		ClientID: clientID,
	}
	return s.sendMessage(ctx, conn, ConnectResponse, successResp)
}

// connectRequest represents the structure of connection request data.
type connectRequest struct {
	PlayerName string `json:"playerName"`
	TeamID     int    `json:"teamID"`
}

// handleClientMessages processes messages from a connected client
func (s *GameServer) handleClientMessages(client *Client) {
	clientID := client.Conn.RemoteAddr().String() + "/" + strconv.FormatUint(uint64(client.ID), 10)

	for client.Connected && s.running {
		// Create context with read timeout for each message
		msgCtx, msgCancel := context.WithTimeout(client.ctx, s.readTimeout)

		msgType, data, err := s.readMessage(msgCtx, client.Conn)
		msgCancel() // Clean up timeout context

		if err != nil {
			if err != io.EOF && err != context.DeadlineExceeded && err != context.Canceled {
				log.Printf("Error reading message from client %d: %v", client.ID, err)
			}
			break
		}

		// Validate message for all types except disconnect
		if msgType != DisconnectNotification {
			if err := s.validator.ValidateMessage(data, clientID); err != nil {
				log.Printf("Message validation failed for client %d: %v", client.ID, err)
				// Don't disconnect client for validation errors, just skip the message
				continue
			}
		}

		// Process message based on type
		switch msgType {
		case PlayerInput:
			s.handlePlayerInput(client, data)

		case PingRequest:
			// Respond to ping request with ping response
			responseCtx, responseCancel := context.WithTimeout(client.ctx, s.writeTimeout)
			if err := s.sendMessage(responseCtx, client.Conn, PingResponse, data); err != nil {
				log.Printf("Failed to send ping response to client %d: %v", client.ID, err)
			}
			responseCancel()

		case ChatMessage:
			// Broadcast chat message to all clients
			s.broadcastChatMessage(client, data)

		case DisconnectNotification:
			// Client is disconnecting gracefully
			log.Printf("Client %d disconnecting", client.ID)
			client.Connected = false

		default:
			log.Printf("Unknown message type %d from client %d", msgType, client.ID)
		}
	}

	// Client disconnected, clean up
	s.removeClient(client)
}

// PlayerInputData represents the structure of player input messages
type PlayerInputData struct {
	Thrust     bool      `json:"thrust"`
	TurnLeft   bool      `json:"turnLeft"`
	TurnRight  bool      `json:"turnRight"`
	FireWeapon int       `json:"fireWeapon"` // -1 if not firing, weapon index otherwise
	BeamDown   bool      `json:"beamDown"`
	BeamUp     bool      `json:"beamUp"`
	BeamAmount int       `json:"beamAmount"`
	TargetID   entity.ID `json:"targetID"` // Target planet ID for beaming
}

// handlePlayerInput processes player input messages
func (s *GameServer) handlePlayerInput(client *Client, data []byte) {
	input, err := s.parsePlayerInput(data)
	if err != nil {
		log.Printf("Error parsing player input: %v", err)
		return
	}

	client.LastInput = time.Now()

	ship := s.findPlayerShip(client)
	if ship == nil {
		return
	}

	s.applyPlayerInput(ship, input)
}

// parsePlayerInput deserializes the player input data from JSON bytes
func (s *GameServer) parsePlayerInput(data []byte) (*PlayerInputData, error) {
	var input PlayerInputData
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, err
	}

	// Validate weapon index
	if err := validation.ValidateWeaponIndex(input.FireWeapon); err != nil {
		return nil, fmt.Errorf("invalid weapon index: %w", err)
	}

	// Validate beam amount
	if err := validation.ValidateBeamAmount(input.BeamAmount); err != nil {
		return nil, fmt.Errorf("invalid beam amount: %w", err)
	}

	return &input, nil
}

// findPlayerShip locates the ship entity for a given client's player
func (s *GameServer) findPlayerShip(client *Client) *entity.Ship {
	s.game.EntityLock.RLock()
	defer s.game.EntityLock.RUnlock()

	for _, player := range s.game.Teams[client.TeamID].Players {
		if player.ID == client.PlayerID {
			if ship, ok := s.game.Ships[player.ShipID]; ok {
				return ship
			}
			break
		}
	}
	return nil
}

// applyPlayerInput applies all validated input commands to the player's ship
func (s *GameServer) applyPlayerInput(ship *entity.Ship, input *PlayerInputData) {
	s.game.EntityLock.Lock()
	defer s.game.EntityLock.Unlock()

	s.applyMovementInput(ship, input)
	s.applyWeaponInput(ship, input)
	s.applyBeamingInput(ship, input)
}

// applyMovementInput updates ship movement controls based on player input
func (s *GameServer) applyMovementInput(ship *entity.Ship, input *PlayerInputData) {
	ship.Thrusting = input.Thrust
	ship.TurningCW = input.TurnRight
	ship.TurningCCW = input.TurnLeft
}

// applyWeaponInput handles weapon firing commands from player input
func (s *GameServer) applyWeaponInput(ship *entity.Ship, input *PlayerInputData) {
	if input.FireWeapon >= 0 {
		s.game.FireWeapon(ship.ID, input.FireWeapon)
	}
}

// applyBeamingInput processes army beaming commands from player input
func (s *GameServer) applyBeamingInput(ship *entity.Ship, input *PlayerInputData) {
	if (input.BeamDown || input.BeamUp) && input.TargetID != 0 {
		direction := "down"
		if input.BeamUp {
			direction = "up"
		}
		s.game.BeamArmies(ship.ID, input.TargetID, direction, input.BeamAmount)
	}
}

// broadcastChatMessage sends a chat message to all connected clients
func (s *GameServer) broadcastChatMessage(sender *Client, data []byte) {
	var chatMsg struct {
		Message string `json:"message"`
		// Can include info about who sent it already
	}

	if err := json.Unmarshal(data, &chatMsg); err != nil {
		log.Printf("Error parsing chat message: %v", err)
		return
	}

	// Validate and sanitize chat message
	sanitizedMessage, err := validation.ValidateChatMessage(chatMsg.Message)
	if err != nil {
		log.Printf("Invalid chat message from client %d: %v", sender.ID, err)
		// Send error back to sender
		errorMsg := struct {
			Error string `json:"error"`
		}{
			Error: "Message rejected: " + err.Error(),
		}

		// Use a short timeout for error response
		ctx, cancel := context.WithTimeout(sender.ctx, 5*time.Second)
		if sendErr := s.sendMessage(ctx, sender.Conn, ChatMessage, errorMsg); sendErr != nil {
			log.Printf("Failed to send chat error to client %d: %v", sender.ID, sendErr)
		}
		cancel()
		return
	}

	// Create message with sender info
	broadcastMsg := struct {
		SenderID   entity.ID `json:"senderID"`
		SenderName string    `json:"senderName"`
		TeamID     int       `json:"teamID"`
		Message    string    `json:"message"`
	}{
		SenderID:   sender.PlayerID,
		SenderName: sender.PlayerName,
		TeamID:     sender.TeamID,
		Message:    sanitizedMessage,
	}

	// Broadcast to all clients
	s.clientsLock.RLock()
	for _, client := range s.clients {
		if client.Connected {
			// Use client context with write timeout for each message
			ctx, cancel := context.WithTimeout(client.ctx, s.writeTimeout)
			if err := s.sendMessage(ctx, client.Conn, ChatMessage, broadcastMsg); err != nil {
				log.Printf("Failed to send chat message to client %d: %v", client.ID, err)
			}
			cancel()
		}
	}
	s.clientsLock.RUnlock()
}

// removeClient removes a client from the server
func (s *GameServer) removeClient(client *Client) {
	// Cancel client context to clean up any ongoing operations
	if client.cancel != nil {
		client.cancel()
	}

	s.clientsLock.Lock()
	delete(s.clients, client.ID)
	s.clientsLock.Unlock()

	// Remove player from game
	s.game.RemovePlayer(client.PlayerID)

	log.Printf("Client %d removed", client.ID)
}

// gameLoop runs the main game loop
func (s *GameServer) gameLoop() {
	ticker := time.NewTicker(s.updateRate)
	defer ticker.Stop()

	for s.running {
		<-ticker.C

		// Update game state
		s.game.Update()

		// Send updates to clients
		if s.game.CurrentTick%uint64(s.ticksPerState) == 0 {
			// Full state update
			s.sendFullStateUpdate()
		} else if s.partialState {
			// Partial state update
			s.sendPartialStateUpdate()
		}
	}
}

// sendFullStateUpdate sends a complete game state to all clients
func (s *GameServer) sendFullStateUpdate() {
	gameState := s.game.GetGameState()

	s.clientsLock.RLock()
	for _, client := range s.clients {
		if client.Connected {
			// Use client context with write timeout
			ctx, cancel := context.WithTimeout(client.ctx, s.writeTimeout)
			if err := s.sendMessage(ctx, client.Conn, GameStateUpdate, gameState); err != nil {
				log.Printf("Failed to send full state update to client %d: %v", client.ID, err)
			}
			cancel()
		}
	}
	s.clientsLock.RUnlock()
}

// sendPartialStateUpdate sends only changed game state to clients
func (s *GameServer) sendPartialStateUpdate() {
	currentState := s.game.GetGameState()

	s.clientsLock.RLock()
	defer s.clientsLock.RUnlock()

	for _, client := range s.clients {
		if !client.Connected {
			continue
		}

		partialState := s.createPartialStateForClient(client, currentState)

		// Use client context with write timeout
		ctx, cancel := context.WithTimeout(client.ctx, s.writeTimeout)
		if err := s.sendMessage(ctx, client.Conn, GameStateUpdate, partialState); err != nil {
			log.Printf("Failed to send partial state update to client %d: %v", client.ID, err)
		}
		cancel()
	}
}

// createPartialStateForClient creates a partial game state containing only entities visible to the client.
func (s *GameServer) createPartialStateForClient(client *Client, currentState *engine.GameState) *engine.GameState {
	partialState := s.initializePartialState(currentState)
	playerShipPos := s.findPlayerShipPosition(client, currentState)

	s.addNearbyEntities(partialState, currentState, playerShipPos)
	s.addAllPlanets(partialState, currentState)

	return partialState
}

// initializePartialState creates an empty partial state with basic information.
func (s *GameServer) initializePartialState(currentState *engine.GameState) *engine.GameState {
	return &engine.GameState{
		Tick:        currentState.Tick,
		Ships:       make(map[entity.ID]engine.ShipState),
		Planets:     make(map[entity.ID]engine.PlanetState),
		Projectiles: make(map[entity.ID]engine.ProjectileState),
		Teams:       currentState.Teams, // Teams always included
	}
}

// findPlayerShipPosition locates the position of the client's ship for visibility calculations.
func (s *GameServer) findPlayerShipPosition(client *Client, currentState *engine.GameState) physics.Vector2D {
	for _, player := range s.game.Teams[client.TeamID].Players {
		if player.ID == client.PlayerID {
			if ship, ok := currentState.Ships[player.ShipID]; ok {
				return ship.Position
			}
			break
		}
	}
	return physics.Vector2D{} // Return zero vector if ship not found
}

// addNearbyEntities adds ships and projectiles within the view radius to the partial state.
func (s *GameServer) addNearbyEntities(partialState *engine.GameState, currentState *engine.GameState, playerPos physics.Vector2D) {
	viewRadius := 3000.0 // Default view radius

	// Add nearby ships
	for id, ship := range currentState.Ships {
		if ship.Position.Distance(playerPos) <= viewRadius {
			partialState.Ships[id] = ship
		}
	}

	// Add nearby projectiles
	for id, proj := range currentState.Projectiles {
		if proj.Position.Distance(playerPos) <= viewRadius {
			partialState.Projectiles[id] = proj
		}
	}
}

// addAllPlanets includes all planets in the partial state as they are always visible.
func (s *GameServer) addAllPlanets(partialState *engine.GameState, currentState *engine.GameState) {
	partialState.Planets = currentState.Planets
}

// readMessage reads a message from a connection
// readMessage reads a message from a connection with context timeout support
func (s *GameServer) readMessage(ctx context.Context, conn net.Conn) (MessageType, []byte, error) {
	// Set read deadline based on context
	if deadline, ok := ctx.Deadline(); ok {
		conn.SetReadDeadline(deadline)
	} else {
		// Fallback to configured timeout
		conn.SetReadDeadline(time.Now().Add(s.readTimeout))
	}
	defer conn.SetReadDeadline(time.Time{}) // Clear deadline

	// Channel to handle the read operation
	type readResult struct {
		msgType MessageType
		data    []byte
		err     error
	}

	resultChan := make(chan readResult, 1)

	// Perform read operation in goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- readResult{err: fmt.Errorf("panic during read: %v", r)}
			}
		}()

		// Read message type
		var msgType MessageType
		if err := binary.Read(conn, binary.BigEndian, &msgType); err != nil {
			resultChan <- readResult{err: err}
			return
		}

		// Read message length
		var msgLen uint16
		if err := binary.Read(conn, binary.BigEndian, &msgLen); err != nil {
			resultChan <- readResult{err: err}
			return
		}

		// Check message size limit
		if int(msgLen) > validation.MaxMessageSize {
			resultChan <- readResult{err: fmt.Errorf("message too large: %d bytes (max %d)", msgLen, validation.MaxMessageSize)}
			return
		}

		// Read message data
		data := make([]byte, msgLen)
		if _, err := io.ReadFull(conn, data); err != nil {
			resultChan <- readResult{err: err}
			return
		}

		resultChan <- readResult{msgType: msgType, data: data, err: nil}
	}()

	// Wait for read completion or context cancellation
	select {
	case result := <-resultChan:
		return result.msgType, result.data, result.err
	case <-ctx.Done():
		// Force connection close on timeout
		conn.Close()
		return 0, nil, ctx.Err()
	}
}

// sendMessage sends a message to a connection with context timeout support
func (s *GameServer) sendMessage(ctx context.Context, conn net.Conn, msgType MessageType, msg interface{}) error {
	// Serialize message
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Check message size
	if len(data) > validation.MaxMessageSize {
		return fmt.Errorf("message too large: %d bytes (max %d)", len(data), validation.MaxMessageSize)
	}

	// Set write deadline based on context
	if deadline, ok := ctx.Deadline(); ok {
		conn.SetWriteDeadline(deadline)
	} else {
		// Fallback to configured timeout
		conn.SetWriteDeadline(time.Now().Add(s.writeTimeout))
	}
	defer conn.SetWriteDeadline(time.Time{}) // Clear deadline

	// Channel to handle the write operation
	type writeResult struct {
		err error
	}

	resultChan := make(chan writeResult, 1)

	// Perform write operation in goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- writeResult{err: fmt.Errorf("panic during write: %v", r)}
			}
		}()

		// Write message type
		if err := binary.Write(conn, binary.BigEndian, msgType); err != nil {
			resultChan <- writeResult{err: err}
			return
		}

		// Write message length
		msgLen := uint16(len(data))
		if err := binary.Write(conn, binary.BigEndian, msgLen); err != nil {
			resultChan <- writeResult{err: err}
			return
		}

		// Write message data
		if _, err := conn.Write(data); err != nil {
			resultChan <- writeResult{err: err}
			return
		}

		resultChan <- writeResult{err: nil}
	}()

	// Wait for write completion or context cancellation
	select {
	case result := <-resultChan:
		return result.err
	case <-ctx.Done():
		// Force connection close on timeout
		conn.Close()
		return ctx.Err()
	}
}
