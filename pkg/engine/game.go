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

	// Clear spatial index for this frame
	g.SpatialIndex = physics.NewQuadTree(
		physics.Rect{
			Center: physics.Vector2D{X: 0, Y: 0},
			Width:  g.Config.WorldSize,
			Height: g.Config.WorldSize,
		},
		10,
	)

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
	// Process ship-projectile collisions
	for _, ship := range g.Ships {
		if !ship.Active {
			continue
		}

		// Find potential collisions
		shipArea := physics.Rect{
			Center: ship.Position,
			Width:  ship.Collider.Radius * 2,
			Height: ship.Collider.Radius * 2,
		}

		potentialCollisions := g.SpatialIndex.Query(shipArea)

		for _, other := range potentialCollisions {
			switch otherEntity := other.(type) {
			case *entity.Projectile:
				if !otherEntity.Active {
					continue
				}

				// Skip projectiles from the same team
				if otherEntity.TeamID == ship.TeamID {
					continue
				}

				// Check for collision
				if ship.GetCollider().Collides(otherEntity.GetCollider()) {
					// Apply damage to ship
					destroyed := ship.TakeDamage(otherEntity.Damage)

					// Deactivate the projectile
					otherEntity.Active = false

					// Publish collision event
					g.EventBus.Publish(event.NewCollisionEvent(
						g,
						uint64(ship.ID),
						uint64(otherEntity.ID),
					))

					// If ship is destroyed
					if destroyed {
						ship.Active = false

						// Find the player that fired the projectile
						if player, ok := g.findPlayerByShipID(otherEntity.OwnerID); ok {
							player.Kills++
							player.Score += 10 // Points for kill
						}

						// Update the killed player's stats
						if player, ok := g.findPlayerByShipID(ship.ID); ok {
							player.Deaths++
						}

						// Publish ship destroyed event
						g.EventBus.Publish(event.NewShipEvent(
							event.ShipDestroyed,
							g,
							uint64(ship.ID),
							ship.TeamID,
						))
					}
				}

			case *entity.Planet:
				// Process ship-planet interactions only if very close
				if ship.Position.Distance(otherEntity.Position) <
					ship.Collider.Radius+otherEntity.Collider.Radius+10 {

					// Prevent ships from flying through planets
					if ship.Position.Distance(otherEntity.Position) <
						ship.Collider.Radius+otherEntity.Collider.Radius {

						// Simple collision response - push ship away
						pushDir := ship.Position.Sub(otherEntity.Position).Normalize()
						pushDistance := (ship.Collider.Radius + otherEntity.Collider.Radius) -
							ship.Position.Distance(otherEntity.Position)

						ship.Position = ship.Position.Add(pushDir.Scale(pushDistance + 1))

						// Reduce velocity
						ship.Velocity = ship.Velocity.Scale(0.5)
					}
				}
			}
		}
	}

	// Process projectile-planet collisions
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
				if proj.Position.Distance(planet.Position) <
					proj.Collider.Radius+planet.Collider.Radius {

					// Deactivate the projectile
					proj.Active = false

					// If this is an enemy planet, apply bombing damage
					if planet.TeamID >= 0 && planet.TeamID != proj.TeamID {
						armiesKilled := planet.Bomb(proj.Damage / 2) // Reduced damage for bombing

						// Update player stats
						if player, ok := g.findPlayerByShipID(proj.OwnerID); ok {
							player.Bombs += armiesKilled
							player.Score += armiesKilled // Points for bombing
						}

						// If planet becomes neutral, update team stats
						if planet.TeamID == -1 {
							if team, ok := g.Teams[planet.TeamID]; ok {
								team.PlanetCount--
							}

							// Publish planet captured event (now neutral)
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

	// Check if ship is close enough to the planet
	if ship.Position.Distance(planet.Position) > ship.Collider.Radius+planet.Collider.Radius+50 {
		return 0, errors.New("ship is too far from planet")
	}

	if direction == "down" {
		// Beam armies from ship to planet
		if ship.Armies <= 0 {
			return 0, errors.New("ship has no armies")
		}

		if amount > ship.Armies {
			amount = ship.Armies
		}

		transferred, captured := planet.BeamDownArmies(ship.TeamID, amount)
		ship.Armies -= transferred

		// If planet was captured
		if captured {
			// Update team planet counts
			if oldTeam, ok := g.Teams[planet.TeamID]; ok {
				oldTeam.PlanetCount--
			}

			if newTeam, ok := g.Teams[ship.TeamID]; ok {
				newTeam.PlanetCount++
			}

			// Update player stats
			if player, ok := g.findPlayerByShipID(shipID); ok {
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

		return transferred, nil
	} else if direction == "up" {
		// Beam armies from planet to ship
		if planet.TeamID != ship.TeamID {
			return 0, errors.New("cannot beam up armies from enemy planet")
		}

		// Check ship capacity
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

	return 0, errors.New("invalid direction, must be 'up' or 'down'")
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

	state := &GameState{
		Tick:        g.CurrentTick,
		Ships:       make(map[entity.ID]ShipState),
		Planets:     make(map[entity.ID]PlanetState),
		Projectiles: make(map[entity.ID]ProjectileState),
		Teams:       make(map[int]TeamState),
	}

	// Copy ship states
	for id, ship := range g.Ships {
		if ship.Active {
			state.Ships[id] = ShipState{
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

	// Copy planet states
	for id, planet := range g.Planets {
		state.Planets[id] = PlanetState{
			ID:       id,
			Name:     planet.Name,
			Position: planet.Position,
			TeamID:   planet.TeamID,
			Armies:   planet.Armies,
		}
	}

	// Copy projectile states
	for id, proj := range g.Projectiles {
		if proj.Active {
			state.Projectiles[id] = ProjectileState{
				ID:       id,
				Position: proj.Position,
				Velocity: proj.Velocity,
				Type:     proj.Type,
				TeamID:   proj.TeamID,
			}
		}
	}

	// Copy team states
	for id, team := range g.Teams {
		state.Teams[id] = TeamState{
			ID:          id,
			Name:        team.Name,
			Color:       team.Color,
			Score:       team.Score,
			ShipCount:   team.ShipCount,
			PlanetCount: team.PlanetCount,
		}
	}

	return state
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
