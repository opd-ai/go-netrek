// pkg/network/server.go
package network

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/opd-ai/go-netrek/pkg/engine"
	"github.com/opd-ai/go-netrek/pkg/entity"
	"github.com/opd-ai/go-netrek/pkg/physics"
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
	listener      net.Listener
	game          *engine.Game
	clients       map[entity.ID]*Client
	clientsLock   sync.RWMutex
	running       bool
	updateRate    time.Duration
	maxClients    int
	ticksPerState int  // How many game ticks between full state updates
	partialState  bool // Whether to send partial updates between full updates
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
}

// NewGameServer creates a new game server
func NewGameServer(game *engine.Game, maxClients int) *GameServer {
	return &GameServer{
		game:          game,
		clients:       make(map[entity.ID]*Client),
		running:       false,
		updateRate:    time.Second / 20, // 20 updates per second
		maxClients:    maxClients,
		ticksPerState: 3, // Send full state every 3 game ticks
		partialState:  true,
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

	log.Printf("Game server started on %s", address)
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

	// Stop game
	s.game.Stop()

	log.Println("Game server stopped")
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

	// Wait for connect request
	msgType, data, err := s.readMessage(conn)
	if err != nil {
		log.Printf("Error reading connect request: %v", err)
		return
	}

	if msgType != ConnectRequest {
		log.Printf("Expected connect request, got %d", msgType)
		return
	}

	// Parse connect request
	var connectReq struct {
		PlayerName string `json:"playerName"`
		TeamID     int    `json:"teamID"`
	}

	if err := json.Unmarshal(data, &connectReq); err != nil {
		log.Printf("Error parsing connect request: %v", err)
		return
	}

	// Add player to game
	playerID, err := s.game.AddPlayer(connectReq.PlayerName, connectReq.TeamID)
	if err != nil {
		log.Printf("Error adding player: %v", err)

		// Send error response
		errorResp := struct {
			Success bool   `json:"success"`
			Error   string `json:"error"`
		}{
			Success: false,
			Error:   err.Error(),
		}

		s.sendMessage(conn, ConnectResponse, errorResp)
		return
	}

	// Create client
	clientID := entity.GenerateID()
	client := &Client{
		ID:         clientID,
		Conn:       conn,
		PlayerID:   playerID,
		PlayerName: connectReq.PlayerName,
		TeamID:     connectReq.TeamID,
		Connected:  true,
		LastInput:  time.Now(),
	}

	// Add client to server
	s.clientsLock.Lock()
	s.clients[clientID] = client
	s.clientsLock.Unlock()

	// Send success response
	successResp := struct {
		Success  bool      `json:"success"`
		PlayerID entity.ID `json:"playerID"`
		ClientID entity.ID `json:"clientID"`
	}{
		Success:  true,
		PlayerID: playerID,
		ClientID: clientID,
	}

	s.sendMessage(conn, ConnectResponse, successResp)

	// Handle client messages
	s.handleClientMessages(client)
}

// handleClientMessages processes messages from a connected client
func (s *GameServer) handleClientMessages(client *Client) {
	for client.Connected && s.running {
		msgType, data, err := s.readMessage(client.Conn)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading message from client %d: %v", client.ID, err)
			}
			break
		}

		// Process message based on type
		switch msgType {
		case PlayerInput:
			s.handlePlayerInput(client, data)

		case PingRequest:
			// Respond to ping request with ping response
			s.sendMessage(client.Conn, PingResponse, data)

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

// handlePlayerInput processes player input messages
func (s *GameServer) handlePlayerInput(client *Client, data []byte) {
	var input struct {
		Thrust     bool      `json:"thrust"`
		TurnLeft   bool      `json:"turnLeft"`
		TurnRight  bool      `json:"turnRight"`
		FireWeapon int       `json:"fireWeapon"` // -1 if not firing, weapon index otherwise
		BeamDown   bool      `json:"beamDown"`
		BeamUp     bool      `json:"beamUp"`
		BeamAmount int       `json:"beamAmount"`
		TargetID   entity.ID `json:"targetID"` // Target planet ID for beaming
	}

	if err := json.Unmarshal(data, &input); err != nil {
		log.Printf("Error parsing player input: %v", err)
		return
	}

	// Update last input time
	client.LastInput = time.Now()

	// Find player's ship
	var ship *entity.Ship
	s.game.EntityLock.RLock()
	for _, player := range s.game.Teams[client.TeamID].Players {
		if player.ID == client.PlayerID {
			if shipObj, ok := s.game.Ships[player.ShipID]; ok {
				ship = shipObj
			}
			break
		}
	}
	s.game.EntityLock.RUnlock()

	if ship == nil {
		return
	}

	// Apply input
	s.game.EntityLock.Lock()

	// Update ship movement
	ship.Thrusting = input.Thrust
	ship.TurningCW = input.TurnRight
	ship.TurningCCW = input.TurnLeft

	// Handle weapon firing
	if input.FireWeapon >= 0 {
		s.game.FireWeapon(ship.ID, input.FireWeapon)
	}

	// Handle beaming
	if (input.BeamDown || input.BeamUp) && input.TargetID != 0 {
		direction := "down"
		if input.BeamUp {
			direction = "up"
		}

		s.game.BeamArmies(ship.ID, input.TargetID, direction, input.BeamAmount)
	}

	s.game.EntityLock.Unlock()
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
		Message:    chatMsg.Message,
	}

	// Broadcast to all clients
	s.clientsLock.RLock()
	for _, client := range s.clients {
		if client.Connected {
			s.sendMessage(client.Conn, ChatMessage, broadcastMsg)
		}
	}
	s.clientsLock.RUnlock()
}

// removeClient removes a client from the server
func (s *GameServer) removeClient(client *Client) {
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
			s.sendMessage(client.Conn, GameStateUpdate, gameState)
		}
	}
	s.clientsLock.RUnlock()
}

// sendPartialStateUpdate sends only changed game state to clients
func (s *GameServer) sendPartialStateUpdate() {
	currentState := s.game.GetGameState()

	// For each client, create a custom partial update
	s.clientsLock.RLock()
	for _, client := range s.clients {
		if !client.Connected {
			continue
		}

		// Create partial state with only nearby entities
		partialState := &engine.GameState{
			Tick:        currentState.Tick,
			Ships:       make(map[entity.ID]engine.ShipState),
			Planets:     make(map[entity.ID]engine.PlanetState),
			Projectiles: make(map[entity.ID]engine.ProjectileState),
			Teams:       currentState.Teams, // Teams always included
		}

		// Get player's ship position for visibility check
		var playerShipPos physics.Vector2D
		var viewRadius float64 = 3000 // Default view radius

		// Find player's ship
		for _, player := range s.game.Teams[client.TeamID].Players {
			if player.ID == client.PlayerID {
				if ship, ok := currentState.Ships[player.ShipID]; ok {
					playerShipPos = ship.Position
					break
				}
			}
		}

		// Add nearby ships
		for id, ship := range currentState.Ships {
			if ship.Position.Distance(playerShipPos) <= viewRadius {
				partialState.Ships[id] = ship
			}
		}

		// Add nearby projectiles
		for id, proj := range currentState.Projectiles {
			if proj.Position.Distance(playerShipPos) <= viewRadius {
				partialState.Projectiles[id] = proj
			}
		}

		// Always include all planets
		partialState.Planets = currentState.Planets

		// Send the partial state
		s.sendMessage(client.Conn, GameStateUpdate, partialState)
	}
	s.clientsLock.RUnlock()
}

// readMessage reads a message from a connection
func (s *GameServer) readMessage(conn net.Conn) (MessageType, []byte, error) {
	// Read message type
	var msgType MessageType
	if err := binary.Read(conn, binary.BigEndian, &msgType); err != nil {
		return 0, nil, err
	}

	// Read message length
	var msgLen uint16
	if err := binary.Read(conn, binary.BigEndian, &msgLen); err != nil {
		return 0, nil, err
	}

	// Read message data
	data := make([]byte, msgLen)
	if _, err := io.ReadFull(conn, data); err != nil {
		return 0, nil, err
	}

	return msgType, data, nil
}

// sendMessage sends a message to a connection
func (s *GameServer) sendMessage(conn net.Conn, msgType MessageType, msg interface{}) error {
	// Serialize message
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Check message size
	if len(data) > 65535 {
		return errors.New("message too large")
	}

	// Write message type
	if err := binary.Write(conn, binary.BigEndian, msgType); err != nil {
		return err
	}

	// Write message length
	msgLen := uint16(len(data))
	if err := binary.Write(conn, binary.BigEndian, msgLen); err != nil {
		return err
	}

	// Write message data
	if _, err := conn.Write(data); err != nil {
		return err
	}

	return nil
}
