// pkg/render/engo/renderer.go
package engo

import (
	"image/color"

	"github.com/EngoEngine/ecs"
	"github.com/EngoEngine/engo"
	"github.com/EngoEngine/engo/common"

	"github.com/opd-ai/go-netrek/pkg/entity"
	"github.com/opd-ai/go-netrek/pkg/physics"
)

// EngoRenderer implements entity.Renderer using the Engo game engine
type EngoRenderer struct {
	world        *ecs.World
	renderSystem *common.RenderSystem

	// Entity management
	shipEntities       map[entity.ID]*ecs.BasicEntity
	planetEntities     map[entity.ID]*ecs.BasicEntity
	projectileEntities map[entity.ID]*ecs.BasicEntity

	// Asset management
	assets *AssetManager
}

// NewEngoRenderer creates a new Engo-based renderer
func NewEngoRenderer(world *ecs.World) *EngoRenderer {
	return &EngoRenderer{
		world:              world,
		shipEntities:       make(map[entity.ID]*ecs.BasicEntity),
		planetEntities:     make(map[entity.ID]*ecs.BasicEntity),
		projectileEntities: make(map[entity.ID]*ecs.BasicEntity),
		assets:             NewAssetManager(),
	}
}

// Initialize sets up the renderer's systems
func (r *EngoRenderer) Initialize() error {
	// Initialize render system
	r.renderSystem = &common.RenderSystem{}
	r.world.AddSystem(r.renderSystem)

	// Load assets
	return r.assets.LoadAssets()
}

// RenderShip implements entity.Renderer
func (r *EngoRenderer) RenderShip(ship *entity.Ship) {
	// Get or create entity for this ship
	basicEntity := r.getOrCreateShipEntity(ship.GetID())

	// Update ship rendering components
	r.updateShipComponents(basicEntity, ship)
}

// RenderPlanet implements entity.Renderer
func (r *EngoRenderer) RenderPlanet(planet *entity.Planet) {
	// Get or create entity for this planet
	basicEntity := r.getOrCreatePlanetEntity(planet.GetID())

	// Update planet rendering components
	r.updatePlanetComponents(basicEntity, planet)
}

// RenderProjectile implements entity.Renderer
func (r *EngoRenderer) RenderProjectile(projectile *entity.Projectile) {
	// Get or create entity for this projectile
	basicEntity := r.getOrCreateProjectileEntity(projectile.GetID())

	// Update projectile rendering components
	r.updateProjectileComponents(basicEntity, projectile)
}

// Clear implements entity.Renderer
func (r *EngoRenderer) Clear() {
	// Engo handles clearing automatically, but we can clean up dead entities here
	r.cleanupInactiveEntities()
}

// Present implements entity.Renderer
func (r *EngoRenderer) Present() {
	// Engo handles presentation automatically through its render system
	// This is called after all entities have been rendered
}

// getOrCreateShipEntity gets an existing ship entity or creates a new one
func (r *EngoRenderer) getOrCreateShipEntity(id entity.ID) *ecs.BasicEntity {
	if entity, exists := r.shipEntities[id]; exists {
		return entity
	}

	// Create new entity
	basicEntity := ecs.NewBasic()
	r.shipEntities[id] = &basicEntity

	// Add to render system with initial components
	renderComponent := common.RenderComponent{
		Drawable: r.assets.GetShipSprite(entity.Scout), // Default ship
		Color:    color.RGBA{255, 255, 255, 255},
	}

	spaceComponent := common.SpaceComponent{
		Position: engo.Point{X: 0, Y: 0},
		Width:    32,
		Height:   32,
	}

	r.renderSystem.Add(&basicEntity, &renderComponent, &spaceComponent)

	return &basicEntity
}

// getOrCreatePlanetEntity gets an existing planet entity or creates a new one
func (r *EngoRenderer) getOrCreatePlanetEntity(id entity.ID) *ecs.BasicEntity {
	if entity, exists := r.planetEntities[id]; exists {
		return entity
	}

	// Create new entity
	basicEntity := ecs.NewBasic()
	r.planetEntities[id] = &basicEntity

	// Add to render system with initial components
	renderComponent := common.RenderComponent{
		Drawable: r.assets.GetPlanetSprite(entity.Homeworld), // Default planet
		Color:    color.RGBA{255, 255, 255, 255},
	}

	spaceComponent := common.SpaceComponent{
		Position: engo.Point{X: 0, Y: 0},
		Width:    48,
		Height:   48,
	}

	r.renderSystem.Add(&basicEntity, &renderComponent, &spaceComponent)

	return &basicEntity
}

// getOrCreateProjectileEntity gets an existing projectile entity or creates a new one
func (r *EngoRenderer) getOrCreateProjectileEntity(id entity.ID) *ecs.BasicEntity {
	if entity, exists := r.projectileEntities[id]; exists {
		return entity
	}

	// Create new entity
	basicEntity := ecs.NewBasic()
	r.projectileEntities[id] = &basicEntity

	// Add to render system with initial components
	renderComponent := common.RenderComponent{
		Drawable: r.assets.GetProjectileSprite("torpedo"), // Default projectile
		Color:    color.RGBA{255, 255, 0, 255},
	}

	spaceComponent := common.SpaceComponent{
		Position: engo.Point{X: 0, Y: 0},
		Width:    8,
		Height:   8,
	}

	r.renderSystem.Add(&basicEntity, &renderComponent, &spaceComponent)

	return &basicEntity
}

// updateShipComponents updates the rendering components for a ship
func (r *EngoRenderer) updateShipComponents(basicEntity *ecs.BasicEntity, ship *entity.Ship) {
	// Update position
	if spaceComponent := r.getSpaceComponent(basicEntity); spaceComponent != nil {
		pos := r.worldToScreen(ship.Position)
		spaceComponent.Position = engo.Point{X: pos.X, Y: pos.Y}
		spaceComponent.Rotation = float32(ship.Rotation)
	}

	// Update ship sprite and color based on class and team
	if renderComponent := r.getRenderComponent(basicEntity); renderComponent != nil {
		renderComponent.Drawable = r.assets.GetShipSprite(ship.Class)
		renderComponent.Color = r.getTeamColor(ship.TeamID)
	}
}

// updatePlanetComponents updates the rendering components for a planet
func (r *EngoRenderer) updatePlanetComponents(basicEntity *ecs.BasicEntity, planet *entity.Planet) {
	// Update position
	if spaceComponent := r.getSpaceComponent(basicEntity); spaceComponent != nil {
		pos := r.worldToScreen(planet.Position)
		spaceComponent.Position = engo.Point{X: pos.X, Y: pos.Y}
	}

	// Update planet sprite and color based on type and ownership
	if renderComponent := r.getRenderComponent(basicEntity); renderComponent != nil {
		renderComponent.Drawable = r.assets.GetPlanetSprite(planet.Type)
		if planet.TeamID >= 0 {
			renderComponent.Color = r.getTeamColor(planet.TeamID)
		} else {
			renderComponent.Color = color.RGBA{128, 128, 128, 255} // Neutral
		}
	}
}

// updateProjectileComponents updates the rendering components for a projectile
func (r *EngoRenderer) updateProjectileComponents(basicEntity *ecs.BasicEntity, projectile *entity.Projectile) {
	// Update position
	if spaceComponent := r.getSpaceComponent(basicEntity); spaceComponent != nil {
		pos := r.worldToScreen(projectile.Position)
		spaceComponent.Position = engo.Point{X: pos.X, Y: pos.Y}
	}

	// Update projectile sprite and color based on type and team
	if renderComponent := r.getRenderComponent(basicEntity); renderComponent != nil {
		renderComponent.Drawable = r.assets.GetProjectileSprite(projectile.Type)
		renderComponent.Color = r.getTeamColor(projectile.TeamID)
	}
}

// Helper functions to get components from entities
func (r *EngoRenderer) getSpaceComponent(entity *ecs.BasicEntity) *common.SpaceComponent {
	// In a real implementation, you would properly query the ECS for components
	// This is a simplified version for this example
	return nil
}

func (r *EngoRenderer) getRenderComponent(entity *ecs.BasicEntity) *common.RenderComponent {
	// In a real implementation, you would properly query the ECS for components
	// This is a simplified version for this example
	return nil
}

// worldToScreen converts world coordinates to screen coordinates
func (r *EngoRenderer) worldToScreen(worldPos physics.Vector2D) engo.Point {
	// Convert from game world coordinates to screen coordinates
	// This would take into account camera position and zoom
	return engo.Point{
		X: float32(worldPos.X) + engo.GameWidth()/2,
		Y: float32(worldPos.Y) + engo.GameHeight()/2,
	}
}

// getTeamColor returns the color for a specific team
func (r *EngoRenderer) getTeamColor(teamID int) color.Color {
	teamColors := []color.Color{
		color.RGBA{255, 0, 0, 255},   // Red
		color.RGBA{0, 255, 0, 255},   // Green
		color.RGBA{0, 0, 255, 255},   // Blue
		color.RGBA{255, 255, 0, 255}, // Yellow
	}

	if teamID >= 0 && teamID < len(teamColors) {
		return teamColors[teamID]
	}

	return color.RGBA{255, 255, 255, 255} // White for unknown teams
}

// cleanupInactiveEntities removes entities that are no longer active
func (r *EngoRenderer) cleanupInactiveEntities() {
	// This would be called to remove entities that are no longer in the game state
	// Implementation would depend on how the game signals entity removal
}

// RemoveShip removes a ship entity from rendering
func (r *EngoRenderer) RemoveShip(id entity.ID) {
	if entity, exists := r.shipEntities[id]; exists {
		r.renderSystem.Remove(*entity)
		delete(r.shipEntities, id)
	}
}

// RemovePlanet removes a planet entity from rendering
func (r *EngoRenderer) RemovePlanet(id entity.ID) {
	if entity, exists := r.planetEntities[id]; exists {
		r.renderSystem.Remove(*entity)
		delete(r.planetEntities, id)
	}
}

// RemoveProjectile removes a projectile entity from rendering
func (r *EngoRenderer) RemoveProjectile(id entity.ID) {
	if entity, exists := r.projectileEntities[id]; exists {
		r.renderSystem.Remove(*entity)
		delete(r.projectileEntities, id)
	}
}
