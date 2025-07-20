// pkg/render/engo/scene.go
package engo

import (
	"github.com/EngoEngine/ecs"
	"github.com/EngoEngine/engo"
	"github.com/EngoEngine/engo/common"

	"github.com/opd-ai/go-netrek/pkg/engine"
	"github.com/opd-ai/go-netrek/pkg/entity"
	"github.com/opd-ai/go-netrek/pkg/event"
	"github.com/opd-ai/go-netrek/pkg/network"
)

// GameScene represents the main game scene in Engo
type GameScene struct {
	world *ecs.World

	// Network components
	client   *network.GameClient
	eventBus *event.Bus

	// Rendering components
	renderer *EngoRenderer
	camera   *CameraSystem
	input    *InputSystem
	hud      *HUDSystem

	// Game state
	gameState *engine.GameState
	playerID  uint64
}

// NewGameScene creates a new game scene
func NewGameScene(client *network.GameClient, eventBus *event.Bus, playerID uint64) *GameScene {
	return &GameScene{
		client:   client,
		eventBus: eventBus,
		playerID: playerID,
		world:    &ecs.World{},
	}
}

// Type returns the scene type (required by Engo)
func (scene *GameScene) Type() string {
	return "GameScene"
}

// Preload is called before the scene starts (required by Engo)
func (scene *GameScene) Preload() {
	// Load any assets here
}

// Setup is called when the scene starts (required by Engo)
func (scene *GameScene) Setup(u engo.Updater) {
	// Set up the world
	scene.world = &ecs.World{}

	// Add the common systems (required for Engo)
	scene.world.AddSystem(&common.RenderSystem{})
	scene.world.AddSystem(&common.MouseSystem{})

	// Initialize renderer
	scene.renderer = NewEngoRenderer(scene.world)
	if err := scene.renderer.Initialize(); err != nil {
		// Use panic instead of engo.Log which doesn't exist
		panic("Failed to initialize renderer: " + err.Error())
	}

	// Initialize camera system
	scene.camera = NewCameraSystem()
	scene.world.AddSystem(scene.camera)

	// Initialize input system
	scene.input = NewInputSystem(scene.client)
	scene.world.AddSystem(scene.input)

	// Initialize HUD system
	scene.hud = NewHUDSystem()
	scene.world.AddSystem(scene.hud)

	// Subscribe to game state updates
	go scene.handleGameStateUpdates()

	// Subscribe to events
	scene.subscribeToEvents()
}

// subscribeToEvents sets up event handlers
func (scene *GameScene) subscribeToEvents() {
	// Subscribe to chat events
	scene.eventBus.Subscribe("chat", func(e event.Event) {
		// Handle chat events
		scene.hud.AddChatMessage("System", "Event received")
	})
}

// handleGameStateUpdates processes game state updates from the client
func (scene *GameScene) handleGameStateUpdates() {
	for gameState := range scene.client.GetGameStateChannel() {
		scene.gameState = gameState
		scene.updateGame(gameState)
	}
}

// updateGame updates the game state and renders the current frame
func (scene *GameScene) updateGame(gameState *engine.GameState) {
	// Clear the previous frame
	scene.renderer.Clear()

	// Render ships
	for _, shipState := range gameState.Ships {
		// Convert ShipState to Ship entity for rendering
		ship := scene.convertShipStateToEntity(shipState)
		scene.renderer.RenderShip(ship)

		// Update camera to follow player's ship if this is the player's ship
		if scene.isPlayerShip(uint64(shipState.ID)) {
			scene.camera.SetTarget(shipState.Position)
		}
	}

	// Render planets
	for _, planetState := range gameState.Planets {
		// Convert PlanetState to Planet entity for rendering
		planet := scene.convertPlanetStateToEntity(planetState)
		scene.renderer.RenderPlanet(planet)
	}

	// Render projectiles
	for _, projectileState := range gameState.Projectiles {
		// Convert ProjectileState to Projectile entity for rendering
		projectile := scene.convertProjectileStateToEntity(projectileState)
		scene.renderer.RenderProjectile(projectile)
	}

	// Update HUD with current game state
	scene.hud.UpdateGameState(gameState)

	// Present the rendered frame
	scene.renderer.Present()
}

// convertShipStateToEntity converts a ShipState to a Ship entity for rendering
func (scene *GameScene) convertShipStateToEntity(shipState engine.ShipState) *entity.Ship {
	// Create a minimal ship entity for rendering purposes
	// Note: This is a simplified conversion - in a real implementation
	// you might want to maintain a mapping of entities
	return &entity.Ship{
		BaseEntity: entity.BaseEntity{
			ID:       shipState.ID,
			Position: shipState.Position,
			Rotation: shipState.Rotation,
			Velocity: shipState.Velocity,
			Active:   true,
		},
		Class:   shipState.Class,
		TeamID:  shipState.TeamID,
		Hull:    shipState.Hull,
		Shields: shipState.Shields,
		Fuel:    shipState.Fuel,
		Armies:  shipState.Armies,
	}
}

// convertPlanetStateToEntity converts a PlanetState to a Planet entity for rendering
func (scene *GameScene) convertPlanetStateToEntity(planetState engine.PlanetState) *entity.Planet {
	return &entity.Planet{
		BaseEntity: entity.BaseEntity{
			ID:       planetState.ID,
			Position: planetState.Position,
			Active:   true,
		},
		Name:   planetState.Name,
		TeamID: planetState.TeamID,
		Armies: planetState.Armies,
	}
}

// convertProjectileStateToEntity converts a ProjectileState to a Projectile entity for rendering
func (scene *GameScene) convertProjectileStateToEntity(projectileState engine.ProjectileState) *entity.Projectile {
	return &entity.Projectile{
		BaseEntity: entity.BaseEntity{
			ID:       projectileState.ID,
			Position: projectileState.Position,
			Velocity: projectileState.Velocity,
			Active:   true,
		},
		Type:   projectileState.Type,
		TeamID: projectileState.TeamID,
	}
}

// isPlayerShip checks if the given ship ID belongs to the current player
func (scene *GameScene) isPlayerShip(shipID uint64) bool {
	// This is a placeholder implementation
	// In a real implementation, you would track the player's ship ID
	return shipID == scene.playerID
}

// Exit is called when the scene is exiting (required by Engo)
func (scene *GameScene) Exit() {
	// Clean up resources
}
