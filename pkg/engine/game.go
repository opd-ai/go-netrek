// pkg/engine/game.go
package engine

import (
	"errors"
	"math"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/opd-ai/go-netrek/pkg/config"
	"github.com/opd-ai/go-netrek/pkg/entity"
	"github.com/opd-ai/go-netrek/pkg/event"
	"github.com/opd-ai/go-netrek/pkg/physics"
)

// pkg/engine/game.go
type GameStatus int

const (
	GameStatusWaiting GameStatus = iota
	GameStatusActive
	GameStatusEnded
)

// WinCondition defines an interface for custom win condition logic
// Returns (winningTeamID, true) if a winner is found, else (-1, false)
type WinCondition interface {
	CheckWinner(game *Game) (int, bool)
}

// Game represents the core game state and logic
type Game struct {
	Config       *config.GameConfig
	Ships        map[entity.ID]*entity.Ship
	Planets      map[entity.ID]*entity.Planet
	Projectiles  map[entity.ID]*entity.Projectile
	Teams        map[int]*Team
	EntityLock   sync.RWMutex
	Running      bool
	TimeStep     float64 // Seconds per game tick
	CurrentTick  uint64
	LastUpdate   time.Time
	EventBus     *event.Bus
	SpatialIndex *physics.QuadTree
	Status       GameStatus
	WinningTeam  int // Team ID of winner, -1 if no winner
	EndTime      time.Time
	StartTime    time.Time
	ElapsedTime  float64 // seconds

	CustomWinCondition WinCondition // Optional custom win condition
}

// Team represents a player team in the game
type Team struct {
	ID          int
	Name        string
	Color       string
	Score       int
	ShipCount   int
	PlanetCount int
	Players     map[entity.ID]*Player
}

// Player represents a connected player
type Player struct {
	ID        entity.ID
	Name      string
	TeamID    int
	ShipID    entity.ID
	Connected bool
	Score     int
	Kills     int
	Deaths    int
	Bombs     int
	Captures  int
}

// NewGame creates a new game with the specified configuration
func NewGame(config *config.GameConfig) *Game {
	game := &Game{
		Config:      config,
		Ships:       make(map[entity.ID]*entity.Ship),
		Planets:     make(map[entity.ID]*entity.Planet),
		Projectiles: make(map[entity.ID]*entity.Projectile),
		Teams:       make(map[int]*Team),
		TimeStep:    1.0 / 60.0, // 60 FPS
		CurrentTick: 0,
		LastUpdate:  time.Now(),
		EventBus:    event.NewEventBus(),
	}

	// Create spatial index for collision detection
	worldSize := config.WorldSize
	game.SpatialIndex = physics.NewQuadTree(
		physics.Rect{
			Center: physics.Vector2D{X: 0, Y: 0},
			Width:  worldSize,
			Height: worldSize,
		},
		10, // Maximum entities per quad before subdivision
	)

	// Initialize teams
	for i, teamConfig := range config.Teams {
		team := &Team{
			ID:      i,
			Name:    teamConfig.Name,
			Color:   teamConfig.Color,
			Players: make(map[entity.ID]*Player),
		}
		game.Teams[i] = team
	}

	// Initialize planets
	for _, planetConfig := range config.Planets {
		planet := entity.NewPlanet(
			entity.GenerateID(),
			planetConfig.Name,
			physics.Vector2D{X: planetConfig.X, Y: planetConfig.Y},
			planetConfig.Type,
		)

		if planetConfig.HomeWorld {
			planet.TeamID = planetConfig.TeamID
			planet.Armies = planetConfig.InitialArmies

			// Update team planet count
			if team, ok := game.Teams[planetConfig.TeamID]; ok {
				team.PlanetCount++
			}
		}

		game.Planets[planet.GetID()] = planet
	}

	// Register event handlers
	game.registerEventHandlers()

	return game
}

// Start begins the game update loop
func (g *Game) Start() {
	g.Running = true
	g.Status = GameStatusActive
	g.StartTime = time.Now()
	g.LastUpdate = time.Now()
	g.EventBus.Publish(&event.BaseEvent{
		EventType: event.GameStarted,
		Source:    g,
	})
}

// Stop halts the game update loop
func (g *Game) Stop() {
	g.Running = false
	g.EventBus.Publish(&event.BaseEvent{
		EventType: event.GameEnded,
		Source:    g,
	})
}

// Update advances the game state by one tick
func (g *Game) Update() {
	now := time.Now()
	if g.Status == GameStatusActive {
		g.ElapsedTime = now.Sub(g.StartTime).Seconds()

		// Check for time limit
		if g.Config.GameRules.TimeLimit > 0 &&
			g.ElapsedTime >= float64(g.Config.GameRules.TimeLimit) {
			// Game ended due to time limit
			g.endGame()
		}
	}
	deltaTime := now.Sub(g.LastUpdate).Seconds()
	g.LastUpdate = now

	// Cap delta time to prevent physics issues
	if deltaTime > 0.1 {
		deltaTime = 0.1
	}

	g.EntityLock.Lock()
	defer g.EntityLock.Unlock()

	// Clear spatial index for this frame (reuse instead of recreate)
	if g.SpatialIndex == nil {
		g.SpatialIndex = physics.NewQuadTree(
			physics.Rect{
				Center: physics.Vector2D{X: 0, Y: 0},
				Width:  g.Config.WorldSize,
				Height: g.Config.WorldSize,
			},
			10,
		)
	} else {
		g.SpatialIndex.Clear()
	}

	// Update all entities
	g.updateShips(deltaTime)
	g.updateProjectiles(deltaTime)
	g.updatePlanets(deltaTime)

	// Add active entities to spatial index
	for _, ship := range g.Ships {
		if ship.Active {
			g.SpatialIndex.Insert(ship.Position, ship)
		}
	}

	for _, projectile := range g.Projectiles {
		if projectile.Active {
			g.SpatialIndex.Insert(projectile.Position, projectile)
		}
	}

	for _, planet := range g.Planets {
		g.SpatialIndex.Insert(planet.Position, planet)
	}

	// Process collisions
	g.detectCollisions()

	// Cleanup inactive entities
	g.cleanupInactiveEntities()

	// Increment tick counter
	g.CurrentTick++
}

// updateShips updates all ships
func (g *Game) updateShips(deltaTime float64) {
	for _, ship := range g.Ships {
		if ship.Active {
			ship.Update(deltaTime)

			// Wrap ships around the world boundaries
			g.wrapEntityPosition(ship)
		}
	}
}

// updateProjectiles updates all projectiles
func (g *Game) updateProjectiles(deltaTime float64) {
	for _, proj := range g.Projectiles {
		if proj.Active {
			proj.Update(deltaTime)

			// Wrap projectiles around the world boundaries
			g.wrapEntityPosition(proj)
		}
	}
}

// updatePlanets updates all planets
func (g *Game) updatePlanets(deltaTime float64) {
	for _, planet := range g.Planets {
		planet.Update(deltaTime)
	}
}

// wrapEntityPosition wraps an entity's position around the world boundaries
func (g *Game) wrapEntityPosition(e interface{}) {
	worldSize := g.Config.WorldSize
	halfWorld := worldSize / 2

	var pos *physics.Vector2D
	var radius float64

	switch entity := e.(type) {
	case *entity.Ship:
		pos = &entity.Position
		radius = entity.Collider.Radius
	case *entity.Projectile:
		pos = &entity.Position
		radius = entity.Collider.Radius
	default:
		return
	}

	// Wrap X coordinate
	if pos.X > halfWorld {
		pos.X -= worldSize
	} else if pos.X < -halfWorld {
		pos.X += worldSize
	}

	// Wrap Y coordinate
	if pos.Y > halfWorld {
		pos.Y -= worldSize
	} else if pos.Y < -halfWorld {
		pos.Y += worldSize
	}

	// After wrapping, robustly nudge away from any overlapping ship
	for _, other := range g.Ships {
		if other == e || !other.Active {
			continue
		}
		dist := pos.Distance(other.Position)
		minDist := radius + other.Collider.Radius
		if dist < minDist && dist > 0 {
			// Nudge along the vector between centers
			delta := pos.Sub(other.Position).Normalize().Scale(minDist - dist + 0.1)
			*pos = pos.Add(delta)
		}
	}
	// Could add similar logic for projectiles if needed
}

// detectCollisions checks for and resolves collisions between entities
func (g *Game) detectCollisions() {
	g.processShipProjectileCollisions()
	g.processShipPlanetCollisions()
	g.processProjectilePlanetCollisions()
}

// processShipProjectileCollisions handles collisions between ships and projectiles.
func (g *Game) processShipProjectileCollisions() {
	for _, ship := range g.Ships {
		if !ship.Active {
			continue
		}
		shipArea := physics.Rect{
			Center: ship.Position,
			Width:  ship.Collider.Radius * 2,
			Height: ship.Collider.Radius * 2,
		}
		potentialCollisions := g.SpatialIndex.Query(shipArea)
		for _, other := range potentialCollisions {
			if projectile, ok := other.(*entity.Projectile); ok {
				g.handleShipProjectileCollision(ship, projectile)
			}
		}
	}
}

// handleShipProjectileCollision checks and resolves a collision between a ship and a projectile.
func (g *Game) handleShipProjectileCollision(ship *entity.Ship, projectile *entity.Projectile) {
	if !projectile.Active || projectile.TeamID == ship.TeamID {
		return
	}
	if ship.GetCollider().Collides(projectile.GetCollider()) {
		destroyed := ship.TakeDamage(projectile.Damage)
		projectile.Active = false
		g.EventBus.Publish(event.NewCollisionEvent(
			g,
			uint64(ship.ID),
			uint64(projectile.ID),
		))
		if destroyed {
			ship.Active = false
			g.updatePlayerStatsOnShipDestruction(ship, projectile)
			g.EventBus.Publish(event.NewShipEvent(
				event.ShipDestroyed,
				g,
				uint64(ship.ID),
				ship.TeamID,
			))
		}
	}
}

// updatePlayerStatsOnShipDestruction updates player stats when a ship is destroyed by a projectile.
func (g *Game) updatePlayerStatsOnShipDestruction(ship *entity.Ship, projectile *entity.Projectile) {
	if player, ok := g.findPlayerByShipID(projectile.OwnerID); ok {
		player.Kills++
		player.Score += 10 // Points for kill
	}
	if player, ok := g.findPlayerByShipID(ship.ID); ok {
		player.Deaths++
	}
}

// processShipPlanetCollisions handles ship-planet proximity and collision responses.
func (g *Game) processShipPlanetCollisions() {
	for _, ship := range g.Ships {
		if !ship.Active {
			continue
		}
		shipArea := physics.Rect{
			Center: ship.Position,
			Width:  ship.Collider.Radius * 2,
			Height: ship.Collider.Radius * 2,
		}
		potentialCollisions := g.SpatialIndex.Query(shipArea)
		for _, other := range potentialCollisions {
			if planet, ok := other.(*entity.Planet); ok {
				g.handleShipPlanetInteraction(ship, planet)
			}
		}
	}
}

// handleShipPlanetInteraction manages ship-planet proximity and collision response.
func (g *Game) handleShipPlanetInteraction(ship *entity.Ship, planet *entity.Planet) {
	distance := ship.Position.Distance(planet.Position)
	if distance < ship.Collider.Radius+planet.Collider.Radius+10 {
		if distance < ship.Collider.Radius+planet.Collider.Radius {
			pushDir := ship.Position.Sub(planet.Position).Normalize()
			pushDistance := (ship.Collider.Radius + planet.Collider.Radius) - distance
			ship.Position = ship.Position.Add(pushDir.Scale(pushDistance + 1))
			ship.Velocity = ship.Velocity.Scale(0.5)
		}
	}
}

// processProjectilePlanetCollisions handles collisions between projectiles and planets.
func (g *Game) processProjectilePlanetCollisions() {
	for _, proj := range g.Projectiles {
		if !proj.Active {
			continue
		}
		projArea := physics.Rect{
			Center: proj.Position,
			Width:  proj.Collider.Radius * 2,
			Height: proj.Collider.Radius * 2,
		}
		potentialCollisions := g.SpatialIndex.Query(projArea)
		for _, other := range potentialCollisions {
			if planet, ok := other.(*entity.Planet); ok {
				g.handleProjectilePlanetCollision(proj, planet)
			}
		}
	}
}

// handleProjectilePlanetCollision checks and resolves a collision between a projectile and a planet.
func (g *Game) handleProjectilePlanetCollision(proj *entity.Projectile, planet *entity.Planet) {
	if proj.Position.Distance(planet.Position) < proj.Collider.Radius+planet.Collider.Radius {
		proj.Active = false
		if planet.TeamID >= 0 && planet.TeamID != proj.TeamID {
			armiesKilled := planet.Bomb(proj.Damage / 2) // Reduced damage for bombing
			if player, ok := g.findPlayerByShipID(proj.OwnerID); ok {
				player.Bombs += armiesKilled
				player.Score += armiesKilled // Points for bombing
			}
			if planet.TeamID == -1 {
				if team, ok := g.Teams[planet.TeamID]; ok {
					team.PlanetCount--
				}
				g.EventBus.Publish(event.NewPlanetEvent(
					event.PlanetCaptured,
					g,
					uint64(planet.ID),
					-1,
					planet.TeamID,
				))
			}
		}
	}
}

// findPlayerByShipID finds a player by their ship ID
func (g *Game) findPlayerByShipID(shipID entity.ID) (*Player, bool) {
	for _, team := range g.Teams {
		for _, player := range team.Players {
			if player.ShipID == shipID {
				return player, true
			}
		}
	}
	return nil, false
}

// cleanupInactiveEntities removes inactive entities
func (g *Game) cleanupInactiveEntities() {
	// Remove inactive projectiles
	for id, proj := range g.Projectiles {
		if !proj.Active {
			delete(g.Projectiles, id)
		}
	}

	// Ships are respawned, not deleted
}

// AddPlayer adds a new player to the game
func (g *Game) AddPlayer(name string, teamID int) (entity.ID, error) {
	g.EntityLock.Lock()
	defer g.EntityLock.Unlock()

	team, ok := g.Teams[teamID]
	if !ok {
		return 0, errors.New("invalid team")
	}

	playerID := entity.GenerateID()
	player := &Player{
		ID:        playerID,
		Name:      name,
		TeamID:    teamID,
		Connected: true,
	}

	// Add to team
	team.Players[playerID] = player

	// Create a ship for the player
	spawnPoint := g.findSpawnPoint(teamID)

	// Determine ship class from team config
	shipClass := entity.Scout
	if teamID < len(g.Config.Teams) {
		shipClass = entity.ShipClassFromString(g.Config.Teams[teamID].StartingShip)
	}

	shipID := entity.GenerateID()
	ship := entity.NewShip(
		shipID,
		shipClass,
		teamID,
		spawnPoint,
	)

	g.Ships[shipID] = ship
	player.ShipID = shipID

	team.ShipCount++

	// Publish player joined event
	g.EventBus.Publish(&event.BaseEvent{
		EventType: event.PlayerJoined,
		Source:    player,
	})

	// Publish ship created event
	g.EventBus.Publish(event.NewShipEvent(
		event.ShipCreated,
		g,
		uint64(shipID),
		teamID,
	))

	return playerID, nil
}

// RemovePlayer removes a player from the game
func (g *Game) RemovePlayer(playerID entity.ID) error {
	g.EntityLock.Lock()
	defer g.EntityLock.Unlock()

	// Find the player
	var player *Player
	var team *Team

	for _, t := range g.Teams {
		if p, ok := t.Players[playerID]; ok {
			player = p
			team = t
			break
		}
	}

	if player == nil {
		return errors.New("player not found")
	}

	// Mark the player's ship as inactive
	if ship, ok := g.Ships[player.ShipID]; ok {
		ship.Active = false

		// Publish ship destroyed event
		g.EventBus.Publish(event.NewShipEvent(
			event.ShipDestroyed,
			g,
			uint64(ship.ID),
			ship.TeamID,
		))
	}

	// Remove player from team
	delete(team.Players, playerID)
	team.ShipCount--

	// Publish player left event
	g.EventBus.Publish(&event.BaseEvent{
		EventType: event.PlayerLeft,
		Source:    player,
	})

	return nil
}

// RespawnShip respawns a player's ship
func (g *Game) RespawnShip(playerID entity.ID) error {
	g.EntityLock.Lock()
	defer g.EntityLock.Unlock()

	// Find the player
	var player *Player

	for _, team := range g.Teams {
		if p, ok := team.Players[playerID]; ok {
			player = p
			break
		}
	}

	if player == nil {
		return errors.New("player not found")
	}

	// Find spawn point
	spawnPoint := g.findSpawnPoint(player.TeamID)

	// Determine ship class from team config
	shipClass := entity.Scout
	if player.TeamID < len(g.Config.Teams) {
		shipClass = entity.ShipClassFromString(g.Config.Teams[player.TeamID].StartingShip)
	}

	// Create a new ship
	oldShipID := player.ShipID
	shipID := entity.GenerateID()
	ship := entity.NewShip(
		shipID,
		shipClass,
		player.TeamID,
		spawnPoint,
	)

	// Update player's ship ID
	player.ShipID = shipID

	// Remove old ship if it exists
	delete(g.Ships, oldShipID)

	// Add new ship
	g.Ships[shipID] = ship

	// Publish ship created event
	g.EventBus.Publish(event.NewShipEvent(
		event.ShipCreated,
		g,
		uint64(shipID),
		player.TeamID,
	))

	return nil
}

// BeamArmies beams armies between a ship and a planet
func (g *Game) BeamArmies(shipID, planetID entity.ID, direction string, amount int) (int, error) {
	g.EntityLock.Lock()
	defer g.EntityLock.Unlock()

	ship, ok := g.Ships[shipID]
	if !ok {
		return 0, errors.New("ship not found")
	}

	planet, ok := g.Planets[planetID]
	if !ok {
		return 0, errors.New("planet not found")
	}

	if err := g.validateBeamingDistance(ship, planet); err != nil {
		return 0, err
	}

	switch direction {
	case "down":
		return g.beamArmiesDown(ship, planet, amount)
	case "up":
		return g.beamArmiesUp(ship, planet, amount)
	default:
		return 0, errors.New("invalid direction, must be 'up' or 'down'")
	}
}

// validateBeamingDistance checks if a ship is close enough to a planet to beam armies.
func (g *Game) validateBeamingDistance(ship *entity.Ship, planet *entity.Planet) error {
	if ship.Position.Distance(planet.Position) > ship.Collider.Radius+planet.Collider.Radius+50 {
		return errors.New("ship is too far from planet")
	}
	return nil
}

// beamArmiesDown handles beaming armies from a ship to a planet.
func (g *Game) beamArmiesDown(ship *entity.Ship, planet *entity.Planet, amount int) (int, error) {
	if ship.Armies <= 0 {
		return 0, errors.New("ship has no armies")
	}

	if amount > ship.Armies {
		amount = ship.Armies
	}

	transferred, captured := planet.BeamDownArmies(ship.TeamID, amount)
	ship.Armies -= transferred

	if captured {
		g.handlePlanetCapture(ship, planet)
	}

	return transferred, nil
}

// handlePlanetCapture updates game state when a planet is captured.
func (g *Game) handlePlanetCapture(ship *entity.Ship, planet *entity.Planet) {
	// Update team planet counts
	if oldTeam, ok := g.Teams[planet.TeamID]; ok {
		oldTeam.PlanetCount--
	}
	if newTeam, ok := g.Teams[ship.TeamID]; ok {
		newTeam.PlanetCount++
	}

	// Update player stats
	if player, ok := g.findPlayerByShipID(ship.ID); ok {
		player.Captures++
		player.Score += 20 // Points for capture
	}

	// Publish planet captured event
	g.EventBus.Publish(event.NewPlanetEvent(
		event.PlanetCaptured,
		g,
		uint64(planet.ID),
		ship.TeamID,
		-1, // Was neutral
	))
}

// beamArmiesUp handles beaming armies from a planet to a ship.
func (g *Game) beamArmiesUp(ship *entity.Ship, planet *entity.Planet, amount int) (int, error) {
	if planet.TeamID != ship.TeamID {
		return 0, errors.New("cannot beam up armies from enemy planet")
	}

	if ship.Armies >= ship.Stats.MaxArmies {
		return 0, errors.New("ship is at maximum army capacity")
	}

	spaceAvailable := ship.Stats.MaxArmies - ship.Armies
	if amount > spaceAvailable {
		amount = spaceAvailable
	}

	transferred := planet.BeamUpArmies(ship.TeamID, amount)
	ship.Armies += transferred

	return transferred, nil
}

// FireWeapon fires a weapon from a ship
func (g *Game) FireWeapon(shipID entity.ID, weaponIndex int) error {
	g.EntityLock.Lock()
	defer g.EntityLock.Unlock()

	ship, ok := g.Ships[shipID]
	if !ok {
		return errors.New("ship not found")
	}

	if !ship.Active {
		return errors.New("ship is not active")
	}

	// Fire the weapon
	projectile := ship.FireWeapon(weaponIndex)
	if projectile == nil {
		return errors.New("weapon could not be fired")
	}

	// Add projectile to game
	g.Projectiles[projectile.ID] = projectile

	// Publish projectile fired event
	g.EventBus.Publish(&event.BaseEvent{
		EventType: event.ProjectileFired,
		Source:    projectile,
	})

	return nil
}

// findSpawnPoint finds a suitable spawn point for a ship
func (g *Game) findSpawnPoint(teamID int) physics.Vector2D {
	// First try to spawn near a team's homeworld
	for _, planet := range g.Planets {
		if planet.TeamID == teamID {
			// Random position around the planet
			angle := rand.Float64() * 2 * math.Pi
			distance := planet.Collider.Radius + 100 + rand.Float64()*100

			return physics.Vector2D{
				X: planet.Position.X + math.Cos(angle)*distance,
				Y: planet.Position.Y + math.Sin(angle)*distance,
			}
		}
	}

	// Fallback to random position
	worldSize := g.Config.WorldSize
	halfWorld := worldSize / 2

	return physics.Vector2D{
		X: rand.Float64()*worldSize - halfWorld,
		Y: rand.Float64()*worldSize - halfWorld,
	}
}

// GetGameState returns a snapshot of the current game state
func (g *Game) GetGameState() *GameState {
	g.EntityLock.RLock()
	defer g.EntityLock.RUnlock()

	return &GameState{
		Tick:        g.CurrentTick,
		Ships:       g.getShipStates(),
		Planets:     g.getPlanetStates(),
		Projectiles: g.getProjectileStates(),
		Teams:       g.getTeamStates(),
	}
}

// getShipStates creates a snapshot of the current ship states.
func (g *Game) getShipStates() map[entity.ID]ShipState {
	states := make(map[entity.ID]ShipState)
	for id, ship := range g.Ships {
		if ship.Active {
			states[id] = ShipState{
				ID:       id,
				Position: ship.Position,
				Rotation: ship.Rotation,
				Velocity: ship.Velocity,
				Hull:     ship.Hull,
				Shields:  ship.Shields,
				Fuel:     ship.Fuel,
				Armies:   ship.Armies,
				TeamID:   ship.TeamID,
				Class:    ship.Class,
			}
		}
	}
	return states
}

// getPlanetStates creates a snapshot of the current planet states.
func (g *Game) getPlanetStates() map[entity.ID]PlanetState {
	states := make(map[entity.ID]PlanetState)
	for id, planet := range g.Planets {
		states[id] = PlanetState{
			ID:       id,
			Name:     planet.Name,
			Position: planet.Position,
			TeamID:   planet.TeamID,
			Armies:   planet.Armies,
		}
	}
	return states
}

// getProjectileStates creates a snapshot of the current projectile states.
func (g *Game) getProjectileStates() map[entity.ID]ProjectileState {
	states := make(map[entity.ID]ProjectileState)
	for id, proj := range g.Projectiles {
		if proj.Active {
			states[id] = ProjectileState{
				ID:       id,
				Position: proj.Position,
				Velocity: proj.Velocity,
				Type:     proj.Type,
				TeamID:   proj.TeamID,
			}
		}
	}
	return states
}

// getTeamStates creates a snapshot of the current team states.
func (g *Game) getTeamStates() map[int]TeamState {
	states := make(map[int]TeamState)
	for id, team := range g.Teams {
		states[id] = TeamState{
			ID:          id,
			Name:        team.Name,
			Color:       team.Color,
			Score:       team.Score,
			ShipCount:   team.ShipCount,
			PlanetCount: team.PlanetCount,
		}
	}
	return states
}

// GameState represents a snapshot of the game state
type GameState struct {
	Tick        uint64
	Ships       map[entity.ID]ShipState
	Planets     map[entity.ID]PlanetState
	Projectiles map[entity.ID]ProjectileState
	Teams       map[int]TeamState
}

// ShipState represents a snapshot of a ship's state
type ShipState struct {
	ID       entity.ID
	Position physics.Vector2D
	Rotation float64
	Velocity physics.Vector2D
	Hull     int
	Shields  int
	Fuel     int
	Armies   int
	TeamID   int
	Class    entity.ShipClass
}

// PlanetState represents a snapshot of a planet's state
type PlanetState struct {
	ID       entity.ID
	Name     string
	Position physics.Vector2D
	TeamID   int
	Armies   int
}

// ProjectileState represents a snapshot of a projectile's state
type ProjectileState struct {
	ID       entity.ID
	Position physics.Vector2D
	Velocity physics.Vector2D
	Type     string
	TeamID   int
}

// TeamState represents a snapshot of a team's state
type TeamState struct {
	ID          int
	Name        string
	Color       string
	Score       int
	ShipCount   int
	PlanetCount int
}

// registerEventHandlers registers handlers for game events
func (g *Game) registerEventHandlers() {
	// Handle ship destruction
	g.EventBus.Subscribe(event.ShipDestroyed, func(e event.Event) {
		if shipEvent, ok := e.(*event.ShipEvent); ok {
			// shipID
			_ = entity.ID(shipEvent.ShipID)

			// Check if any teams have no ships left
			for _, team := range g.Teams {
				if team.ShipCount == 0 {
					// Set game status to ended with winning team
					g.Status = GameStatusEnded

					// Find the winning team (team with most planets or ships)
					var winnerID int = -1
					maxPlanets := 0

					for id, t := range g.Teams {
						if t.PlanetCount > maxPlanets {
							maxPlanets = t.PlanetCount
							winnerID = id
						}
					}

					g.WinningTeam = winnerID
					g.EndTime = time.Now()

					// Publish game ended event
					g.EventBus.Publish(&event.BaseEvent{
						EventType: event.GameEnded,
						Source:    g.Teams[winnerID],
					})

					// Stop the game
					g.Running = false
				}
			}
		}
	})
}

// Add helper method for ending the game
func (g *Game) endGame() {
	g.Status = GameStatusEnded
	g.EndTime = time.Now()
	g.Running = false

	// Use custom win condition if set
	if g.CustomWinCondition != nil {
		if winnerID, ok := g.CustomWinCondition.CheckWinner(g); ok {
			g.WinningTeam = winnerID
			if winnerID >= 0 {
				g.EventBus.Publish(&event.BaseEvent{
					EventType: event.GameEnded,
					Source:    g.Teams[winnerID],
				})
			} else {
				g.EventBus.Publish(&event.BaseEvent{
					EventType: event.GameEnded,
					Source:    g,
				})
			}
			return
		}
	}

	// Default logic: score or conquest
	var winnerID int = -1
	maxScore := 0
	for id, team := range g.Teams {
		if team.Score > maxScore ||
			(g.Config.GameRules.WinCondition == "conquest" && team.PlanetCount > maxScore) {
			maxScore = team.Score
			if g.Config.GameRules.WinCondition == "conquest" {
				maxScore = team.PlanetCount
			}
			winnerID = id
		}
	}
	g.WinningTeam = winnerID
	if winnerID >= 0 {
		g.EventBus.Publish(&event.BaseEvent{
			EventType: event.GameEnded,
			Source:    g.Teams[winnerID],
		})
	} else {
		g.EventBus.Publish(&event.BaseEvent{
			EventType: event.GameEnded,
			Source:    g,
		})
	}
}
