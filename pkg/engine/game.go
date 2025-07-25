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
	"github.com/opd-ai/go-netrek/pkg/resource"
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

	// Resource management
	ResourceManager *resource.ResourceManager
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

	// Initialize game components first
	game.initSpatialIndex()
	game.initTeams()
	game.initPlanets()
	game.registerEventHandlers()

	return game
}

// InitializeResourceManager initializes the resource manager with environment configuration.
// This is called separately to avoid circular dependencies during game creation.
func (g *Game) InitializeResourceManager() error {
	envConfig, err := config.LoadConfigFromEnv()
	if err != nil {
		// Use safe defaults if environment config fails
		envConfig = &config.EnvironmentConfig{
			MaxMemoryMB:           500,
			MaxGoroutines:         1000,
			ShutdownTimeout:       30 * time.Second,
			ResourceCheckInterval: 10 * time.Second,
		}
	}
	g.ResourceManager = resource.NewResourceManager(envConfig)
	return g.ResourceManager.Start()
}

// initSpatialIndex creates the spatial index for collision detection.
func (g *Game) initSpatialIndex() {
	worldSize := g.Config.WorldSize
	g.SpatialIndex = physics.NewQuadTree(
		physics.Rect{
			Center: physics.Vector2D{X: 0, Y: 0},
			Width:  worldSize,
			Height: worldSize,
		},
		10, // Maximum entities per quad before subdivision
	)
}

// initTeams initializes the teams based on the game configuration.
func (g *Game) initTeams() {
	for i, teamConfig := range g.Config.Teams {
		team := &Team{
			ID:      i,
			Name:    teamConfig.Name,
			Color:   teamConfig.Color,
			Players: make(map[entity.ID]*Player),
		}
		g.Teams[i] = team
	}
}

// initPlanets initializes the planets based on the game configuration.
func (g *Game) initPlanets() {
	for _, planetConfig := range g.Config.Planets {
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
			if team, ok := g.Teams[planetConfig.TeamID]; ok {
				team.PlanetCount++
			}
		}

		g.Planets[planet.GetID()] = planet
	}
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
	deltaTime := g.calculateDeltaTime()

	// Lock the entire update to ensure consistency across all entity operations
	g.EntityLock.Lock()
	defer g.EntityLock.Unlock()

	g.checkTimeLimit()
	g.checkWinConditions() // Check for conquest/score-based win conditions
	g.updateGameState(deltaTime)
}

// checkTimeLimit checks if the game has ended due to the time limit.
func (g *Game) checkTimeLimit() {
	if g.Status == GameStatusActive {
		g.ElapsedTime = time.Since(g.StartTime).Seconds()
		if g.Config.GameRules.TimeLimit > 0 &&
			g.ElapsedTime >= float64(g.Config.GameRules.TimeLimit) {
			g.endGameInternal()
		}
	}
}

// checkWinConditions checks if win conditions have been met (conquest, score-based, or custom).
func (g *Game) checkWinConditions() {
	if g.Status != GameStatusActive {
		return
	}

	// Check custom win condition first if present
	if g.CustomWinCondition != nil {
		if _, hasWinner := g.CustomWinCondition.CheckWinner(g); hasWinner {
			g.endGameInternal()
			return
		}
	}

	winCondition := g.Config.GameRules.WinCondition
	switch winCondition {
	case "conquest":
		g.checkConquestWin()
	case "score":
		g.checkScoreWin()
	}
}

// checkConquestWin checks if any team has conquered all planets.
func (g *Game) checkConquestWin() {
	// Count planets per team
	teamPlanetCounts := make(map[int]int)
	totalPlanets := len(g.Planets)

	for _, planet := range g.Planets {
		if planet.TeamID >= 0 { // Only count conquered planets
			teamPlanetCounts[planet.TeamID]++
		}
	}

	// Check if any team has conquered all planets
	for _, planetCount := range teamPlanetCounts {
		if planetCount == totalPlanets && totalPlanets > 0 {
			g.endGameInternal()
			return
		}
	}
}

// checkScoreWin checks if any team has reached the maximum score.
func (g *Game) checkScoreWin() {
	maxScore := g.Config.GameRules.MaxScore
	if maxScore <= 0 {
		return // No score limit configured
	}

	for _, team := range g.Teams {
		if team.Score >= maxScore {
			g.endGameInternal()
			return
		}
	}
}

// calculateDeltaTime calculates the time since the last update and caps it.
func (g *Game) calculateDeltaTime() float64 {
	now := time.Now()
	deltaTime := now.Sub(g.LastUpdate).Seconds()
	g.LastUpdate = now

	// Cap delta time to prevent physics issues
	if deltaTime > 0.1 {
		deltaTime = 0.1
	}
	return deltaTime
}

// updateGameState updates all entities, processes collisions, and cleans up.
func (g *Game) updateGameState(deltaTime float64) {
	g.updateEntities(deltaTime)
	g.processCollisions()
	g.cleanupInactiveEntities()
	g.CurrentTick++
}

// updateEntities updates all entities and the spatial index.
func (g *Game) updateEntities(deltaTime float64) {
	g.prepareSpatialIndex()

	// Update all entities
	g.updateShips(deltaTime)
	g.updateProjectiles(deltaTime)
	g.updatePlanets(deltaTime)

	g.populateSpatialIndex()
}

// prepareSpatialIndex clears the spatial index for the new frame or initializes it.
func (g *Game) prepareSpatialIndex() {
	if g.SpatialIndex == nil {
		g.SpatialIndex = physics.NewQuadTree(
			physics.Rect{
				Center: physics.Vector2D{X: 0, Y: 0},
				Width:  g.Config.WorldSize,
				Height: g.Config.WorldSize,
			},
			10, // Maximum entities per quad before subdivision
		)
	} else {
		g.SpatialIndex.Clear()
	}
}

// populateSpatialIndex adds all active entities to the spatial index.
func (g *Game) populateSpatialIndex() {
	// Note: Called from within locked context in Update()
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
}

// processCollisions checks for and resolves collisions between entities.
func (g *Game) processCollisions() {
	g.detectCollisions()
}

// updateShips updates all ships
func (g *Game) updateShips(deltaTime float64) {
	// Note: Called from within locked context in Update()
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
	// Note: Called from within locked context in Update()
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
	// Note: Called from within locked context in Update()
	for _, planet := range g.Planets {
		planet.Update(deltaTime)
	}
}

// wrapEntityPosition wraps an entity's position around the world boundaries
func (g *Game) wrapEntityPosition(e interface{}) {
	pos, radius, ok := g.extractEntityPositionData(e)
	if !ok {
		return
	}

	g.wrapCoordinatesAroundWorld(pos)
	g.resolvePositionCollisions(e, pos, radius)
}

// extractEntityPositionData extracts position and radius from an entity interface.
// It returns the position vector, radius, and whether the extraction was successful.
func (g *Game) extractEntityPositionData(e interface{}) (*physics.Vector2D, float64, bool) {
	switch entity := e.(type) {
	case *entity.Ship:
		return &entity.Position, entity.Collider.Radius, true
	case *entity.Projectile:
		return &entity.Position, entity.Collider.Radius, true
	default:
		return nil, 0, false
	}
}

// wrapCoordinatesAroundWorld wraps the given position coordinates around world boundaries.
// It modifies the position in-place to ensure it stays within the world bounds.
func (g *Game) wrapCoordinatesAroundWorld(pos *physics.Vector2D) {
	worldSize := g.Config.WorldSize
	halfWorld := worldSize / 2

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
}

// resolvePositionCollisions nudges the entity away from any overlapping ships.
// It prevents entities from overlapping after position wrapping by adjusting their position.
func (g *Game) resolvePositionCollisions(e interface{}, pos *physics.Vector2D, radius float64) {
	// Note: This method is called from within already-locked sections,
	// so we don't need additional locking here
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
}

// detectCollisions checks for and resolves collisions between entities
func (g *Game) detectCollisions() {
	g.processShipProjectileCollisions()
	g.processShipPlanetCollisions()
	g.processProjectilePlanetCollisions()
}

// processShipProjectileCollisions handles collisions between ships and projectiles.
func (g *Game) processShipProjectileCollisions() {
	// Note: Called from within locked context in Update()
	for _, ship := range g.Ships {
		g.checkCollisionsForShip(ship)
	}
}

// checkCollisionsForShip finds and handles collisions for a single ship.
func (g *Game) checkCollisionsForShip(ship *entity.Ship) {
	if !ship.Active {
		return
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

// handleShipProjectileCollision checks and resolves a collision between a ship and a projectile.
func (g *Game) handleShipProjectileCollision(ship *entity.Ship, projectile *entity.Projectile) {
	if !g.canShipAndProjectileCollide(ship, projectile) {
		return
	}

	if ship.GetCollider().Collides(projectile.GetCollider()) {
		g.processShipDamage(ship, projectile)
	}
}

// canShipAndProjectileCollide determines if a collision check is necessary.
func (g *Game) canShipAndProjectileCollide(ship *entity.Ship, projectile *entity.Projectile) bool {
	return projectile.Active && projectile.TeamID != ship.TeamID
}

// processShipDamage handles the consequences of a ship taking damage from a projectile.
func (g *Game) processShipDamage(ship *entity.Ship, projectile *entity.Projectile) {
	destroyed := ship.TakeDamage(projectile.Damage)
	projectile.Active = false

	g.EventBus.Publish(event.NewCollisionEvent(
		g,
		uint64(ship.ID),
		uint64(projectile.ID),
	))

	if destroyed {
		g.handleShipDestruction(ship, projectile)
	}
}

// handleShipDestruction manages the game state changes when a ship is destroyed.
func (g *Game) handleShipDestruction(ship *entity.Ship, projectile *entity.Projectile) {
	ship.Active = false
	g.updatePlayerStatsOnShipDestruction(ship, projectile)
	g.EventBus.Publish(event.NewShipEvent(
		event.ShipDestroyed,
		g,
		uint64(ship.ID),
		ship.TeamID,
	))
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
	// Note: Called from within locked context in Update()
	for _, ship := range g.Ships {
		g.checkInteractionsForShip(ship)
	}
}

// checkInteractionsForShip finds and handles interactions for a single ship.
func (g *Game) checkInteractionsForShip(ship *entity.Ship) {
	if !ship.Active {
		return
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
	// Note: Called from within locked context in Update()
	for _, proj := range g.Projectiles {
		g.checkCollisionsForProjectile(proj)
	}
}

// checkCollisionsForProjectile finds and handles collisions for a single projectile.
func (g *Game) checkCollisionsForProjectile(proj *entity.Projectile) {
	if !proj.Active {
		return
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

// handleProjectilePlanetCollision checks and resolves a collision between a projectile and a planet.
func (g *Game) handleProjectilePlanetCollision(proj *entity.Projectile, planet *entity.Planet) {
	if proj.Position.Distance(planet.Position) < proj.Collider.Radius+planet.Collider.Radius {
		proj.Active = false
		if planet.TeamID >= 0 && planet.TeamID != proj.TeamID {
			g.processPlanetBombing(proj, planet)
		}
	}
}

// processPlanetBombing handles the logic when a projectile bombs a planet.
func (g *Game) processPlanetBombing(proj *entity.Projectile, planet *entity.Planet) {
	armiesKilled := planet.Bomb(proj.Damage / 2) // Reduced damage for bombing
	if player, ok := g.findPlayerByShipID(proj.OwnerID); ok {
		player.Bombs += armiesKilled
		player.Score += armiesKilled // Points for bombing
	}
	if planet.TeamID == -1 { // Planet was just neutralized
		g.handlePlanetNeutralization(planet)
	}
}

// handlePlanetNeutralization updates game state when a planet becomes neutral.
func (g *Game) handlePlanetNeutralization(planet *entity.Planet) {
	if team, ok := g.Teams[planet.TeamID]; ok {
		team.PlanetCount--
	}
	g.EventBus.Publish(event.NewPlanetEvent(
		event.PlanetCaptured, // Or a new "PlanetNeutralized" event type
		g,
		uint64(planet.ID),
		-1,            // New team is neutral
		planet.TeamID, // Old team ID
	))
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
	// Note: This method is called from within already-locked sections (Update method),
	// so we don't need additional locking here

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

	team, err := g.validateTeam(teamID)
	if err != nil {
		return 0, err
	}

	player := g.createAndAddPlayer(name, team)
	ship := g.createAndAddShip(player)
	g.assignShipToPlayer(player, ship, team)

	g.publishPlayerAndShipCreationEvents(player, ship)

	return player.ID, nil
}

// validateTeam checks if the teamID is valid and returns the team.
func (g *Game) validateTeam(teamID int) (*Team, error) {
	team, ok := g.Teams[teamID]
	if !ok {
		return nil, errors.New("invalid team")
	}
	return team, nil
}

// createAndAddPlayer creates a new player and adds it to the team.
func (g *Game) createAndAddPlayer(name string, team *Team) *Player {
	player := g.createPlayer(name, team.ID)
	team.Players[player.ID] = player
	return player
}

// createAndAddShip creates a new ship and adds it to the game.
func (g *Game) createAndAddShip(player *Player) *entity.Ship {
	ship := g.createShipForPlayer(player)
	g.Ships[ship.ID] = ship
	return ship
}

// assignShipToPlayer assigns a ship to a player and updates team stats.
func (g *Game) assignShipToPlayer(player *Player, ship *entity.Ship, team *Team) {
	player.ShipID = ship.ID
	team.ShipCount++
}

// createPlayer creates a new player entity.
func (g *Game) createPlayer(name string, teamID int) *Player {
	return &Player{
		ID:        entity.GenerateID(),
		Name:      name,
		TeamID:    teamID,
		Connected: true,
	}
}

// createShipForPlayer creates a new ship for a given player.
func (g *Game) createShipForPlayer(player *Player) *entity.Ship {
	spawnPoint := g.findSpawnPoint(player.TeamID)
	shipClass := entity.Scout
	if player.TeamID < len(g.Config.Teams) {
		shipClass = entity.ShipClassFromString(g.Config.Teams[player.TeamID].StartingShip)
	}
	return entity.NewShip(
		entity.GenerateID(),
		shipClass,
		player.TeamID,
		spawnPoint,
	)
}

// publishPlayerAndShipCreationEvents publishes events for player joining and ship creation.
func (g *Game) publishPlayerAndShipCreationEvents(player *Player, ship *entity.Ship) {
	g.EventBus.Publish(&event.BaseEvent{
		EventType: event.PlayerJoined,
		Source:    player,
	})
	g.EventBus.Publish(event.NewShipEvent(
		event.ShipCreated,
		g,
		uint64(ship.ID),
		ship.TeamID,
	))
}

// RemovePlayer removes a player from the game
func (g *Game) RemovePlayer(playerID entity.ID) error {
	g.EntityLock.Lock()
	defer g.EntityLock.Unlock()

	player, team, err := g.findPlayerAndTeam(playerID)
	if err != nil {
		return err
	}

	g.deactivatePlayerShip(player)
	g.removePlayerFromTeam(player, team)
	g.publishPlayerLeftEvent(player)

	return nil
}

// findPlayerAndTeam finds a player and their team by player ID.
func (g *Game) findPlayerAndTeam(playerID entity.ID) (*Player, *Team, error) {
	for _, t := range g.Teams {
		if p, ok := t.Players[playerID]; ok {
			return p, t, nil
		}
	}
	return nil, nil, errors.New("player not found")
}

// deactivatePlayerShip marks the player's ship as inactive and publishes an event.
func (g *Game) deactivatePlayerShip(player *Player) {
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
}

// removePlayerFromTeam removes a player from their team's records.
func (g *Game) removePlayerFromTeam(player *Player, team *Team) {
	delete(team.Players, player.ID)
	team.ShipCount--
}

// publishPlayerLeftEvent publishes an event indicating a player has left.
func (g *Game) publishPlayerLeftEvent(player *Player) {
	g.EventBus.Publish(&event.BaseEvent{
		EventType: event.PlayerLeft,
		Source:    player,
	})
}

// RespawnShip respawns a player's ship
func (g *Game) RespawnShip(playerID entity.ID) error {
	g.EntityLock.Lock()
	defer g.EntityLock.Unlock()

	player, err := g.findPlayerByID(playerID)
	if err != nil {
		return err
	}

	// Create a new ship
	newShip := g.createShipForPlayer(player)

	// Update player's ship ID and replace ship in game
	oldShipID := player.ShipID
	player.ShipID = newShip.ID
	delete(g.Ships, oldShipID)
	g.Ships[newShip.ID] = newShip

	// Publish ship created event
	g.EventBus.Publish(event.NewShipEvent(
		event.ShipCreated,
		g,
		uint64(newShip.ID),
		player.TeamID,
	))

	return nil
}

// findPlayerByID finds a player by their ID.
func (g *Game) findPlayerByID(playerID entity.ID) (*Player, error) {
	for _, team := range g.Teams {
		if p, ok := team.Players[playerID]; ok {
			return p, nil
		}
	}
	return nil, errors.New("player not found")
}

// BeamArmies beams armies between a ship and a planet
func (g *Game) BeamArmies(shipID, planetID entity.ID, direction string, amount int) (int, error) {
	g.EntityLock.Lock()
	defer g.EntityLock.Unlock()

	ship, planet, err := g.findShipAndPlanet(shipID, planetID)
	if err != nil {
		return 0, err
	}

	if err := g.validateBeamingPreconditions(ship, planet, direction); err != nil {
		return 0, err
	}

	return g.executeBeaming(ship, planet, direction, amount)
}

// findShipAndPlanet retrieves the ship and planet for beaming.
func (g *Game) findShipAndPlanet(shipID, planetID entity.ID) (*entity.Ship, *entity.Planet, error) {
	ship, ok := g.Ships[shipID]
	if !ok {
		return nil, nil, errors.New("ship not found")
	}
	planet, ok := g.Planets[planetID]
	if !ok {
		return nil, nil, errors.New("planet not found")
	}
	return ship, planet, nil
}

// executeBeaming performs the beaming action based on the direction.
func (g *Game) executeBeaming(ship *entity.Ship, planet *entity.Planet, direction string, amount int) (int, error) {
	switch direction {
	case "down":
		return g.beamArmiesDown(ship, planet, amount)
	case "up":
		return g.beamArmiesUp(ship, planet, amount)
	default:
		return 0, errors.New("invalid beam direction")
	}
}

// validateBeamingPreconditions checks if the ship and planet are in a valid state for beaming.
func (g *Game) validateBeamingPreconditions(ship *entity.Ship, planet *entity.Planet, direction string) error {
	if !ship.Active {
		return errors.New("ship is not active")
	}
	if err := g.validateBeamingDistance(ship, planet); err != nil {
		return err
	}
	if direction == "up" && ship.TeamID != planet.TeamID {
		return errors.New("cannot beam up from an enemy planet")
	}
	return nil
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
	oldTeamID := planet.TeamID
	if oldTeamID >= 0 {
		if team, ok := g.Teams[oldTeamID]; ok {
			team.PlanetCount--
		}
	}

	planet.TeamID = ship.TeamID
	if team, ok := g.Teams[ship.TeamID]; ok {
		team.PlanetCount++
	}

	if player, ok := g.findPlayerByShipID(ship.ID); ok {
		player.Captures++
		player.Score += 50 // Points for capture
	}

	g.EventBus.Publish(event.NewPlanetEvent(
		event.PlanetCaptured,
		g,
		uint64(planet.ID),
		ship.TeamID,
		oldTeamID,
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

	ship, err := g.findActiveShip(shipID)
	if err != nil {
		return err
	}

	projectile := ship.FireWeapon(weaponIndex)
	if projectile == nil {
		return nil // Weapon on cooldown or out of ammo
	}

	g.registerAndPublishProjectile(projectile, ship.ID)
	return nil
}

// findActiveShip finds a ship by ID and checks if it's active.
func (g *Game) findActiveShip(shipID entity.ID) (*entity.Ship, error) {
	ship, ok := g.Ships[shipID]
	if !ok {
		return nil, errors.New("ship not found")
	}
	if !ship.Active {
		return nil, errors.New("ship is not active")
	}
	return ship, nil
}

// registerAndPublishProjectile adds the projectile to the game and publishes an event.
func (g *Game) registerAndPublishProjectile(projectile *entity.Projectile, ownerID entity.ID) {
	projectile.OwnerID = ownerID
	g.Projectiles[projectile.ID] = projectile

	g.EventBus.Publish(&event.BaseEvent{
		EventType: event.ProjectileFired,
		Source:    projectile,
	})
}

// findSpawnPoint finds a suitable spawn point for a ship
func (g *Game) findSpawnPoint(teamID int) physics.Vector2D {
	if pos, ok := g.findSpawnPointNearHomeworld(teamID); ok {
		return pos
	}
	return g.findRandomSpawnPoint()
}

// findSpawnPointNearHomeworld tries to find a spawn point near a team's planet.
func (g *Game) findSpawnPointNearHomeworld(teamID int) (physics.Vector2D, bool) {
	for _, planet := range g.Planets {
		if planet.TeamID == teamID {
			angle := rand.Float64() * 2 * math.Pi
			distance := planet.Collider.Radius + 100 + rand.Float64()*100
			return physics.Vector2D{
				X: planet.Position.X + math.Cos(angle)*distance,
				Y: planet.Position.Y + math.Sin(angle)*distance,
			}, true
		}
	}
	return physics.Vector2D{}, false
}

// findRandomSpawnPoint returns a random position in the world.
func (g *Game) findRandomSpawnPoint() physics.Vector2D {
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

	return g.createGameStateSnapshot()
}

// createGameStateSnapshot builds and returns the complete game state.
func (g *Game) createGameStateSnapshot() *GameState {
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
	g.EventBus.Subscribe(event.ShipDestroyed, g.handleShipDestroyedEvent)
}

// handleShipDestroyedEvent handles the logic when a ship is destroyed.
func (g *Game) handleShipDestroyedEvent(e event.Event) {
	if _, ok := e.(*event.ShipEvent); !ok {
		return
	}

	for _, team := range g.Teams {
		if team.ShipCount == 0 {
			g.endGameWithWinner()
			return
		}
	}
}

// endGameWithWinner determines the winner and ends the game.
// Note: This method should only be called from within a locked context.
func (g *Game) endGameWithWinner() {
	if g.Status == GameStatusEnded {
		return // Game already ended
	}

	g.Status = GameStatusEnded
	g.EndTime = time.Now()

	winnerID := -1
	maxPlanets := 0

	for id, t := range g.Teams {
		if t.PlanetCount > maxPlanets {
			maxPlanets = t.PlanetCount
			winnerID = id
		}
	}

	g.WinningTeam = winnerID

	var winnerSource interface{} = g
	if winner, ok := g.Teams[winnerID]; ok {
		winnerSource = winner
	}

	g.EventBus.Publish(&event.BaseEvent{
		EventType: event.GameEnded,
		Source:    winnerSource,
	})

	g.Running = false
}

// endGame ends the game safely, acquiring lock if needed
func (g *Game) endGame() {
	g.EntityLock.Lock()
	defer g.EntityLock.Unlock()

	g.endGameInternal()
}

// endGameInternal ends the game (must be called with lock held)
func (g *Game) endGameInternal() {
	if g.Status == GameStatusEnded {
		return
	}
	g.Status = GameStatusEnded
	g.EndTime = time.Now()
	g.Running = false

	winnerID := g.determineWinner()
	g.WinningTeam = winnerID

	g.publishGameEndedEvent(winnerID)
}

// determineWinner calculates the winning team based on game rules.
// It returns the winner's team ID, or -1 for a draw/no winner.
func (g *Game) determineWinner() int {
	// Use custom win condition if set
	if g.CustomWinCondition != nil {
		if winnerID, ok := g.CustomWinCondition.CheckWinner(g); ok {
			return winnerID
		}
	}

	// Default logic: score or conquest
	return g.calculateWinnerByDefaultRules()
}

// calculateWinnerByDefaultRules determines the winner based on score or conquest.
func (g *Game) calculateWinnerByDefaultRules() int {
	var winnerID int = -1
	maxScore := -1 // Use -1 to handle zero scores correctly

	for id, team := range g.Teams {
		currentScore := 0
		if g.Config.GameRules.WinCondition == "conquest" {
			currentScore = team.PlanetCount
		} else {
			currentScore = team.Score
		}

		if currentScore > maxScore {
			maxScore = currentScore
			winnerID = id
		} else if currentScore == maxScore {
			winnerID = -1 // Tie
		}
	}
	return winnerID
}

// publishGameEndedEvent sends the game ended event.
func (g *Game) publishGameEndedEvent(winnerID int) {
	var source interface{} = g
	if winnerID >= 0 {
		if winner, ok := g.Teams[winnerID]; ok {
			source = winner
		}
	}
	g.EventBus.Publish(&event.BaseEvent{
		EventType: event.GameEnded,
		Source:    source,
	})
}
