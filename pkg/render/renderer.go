// pkg/render/renderer.go
package render

import (
	"context"

	"github.com/opd-ai/go-netrek/pkg/entity"
	"github.com/opd-ai/go-netrek/pkg/logging"
)

// NullRenderer is a simple implementation of entity.Renderer.
type NullRenderer struct {
	logger *logging.Logger
}

// NewNullRenderer creates a new NullRenderer with structured logging.
func NewNullRenderer() *NullRenderer {
	return &NullRenderer{
		logger: logging.NewLogger(),
	}
}

// Clear implements entity.Renderer.
func (d *NullRenderer) Clear() {
	ctx := context.Background()
	d.logger.Debug(ctx, "Clear called")
}

// Present implements entity.Renderer.
func (d *NullRenderer) Present() {
	ctx := context.Background()
	d.logger.Debug(ctx, "Present called")
}

// RenderPlanet implements entity.Renderer.
func (d *NullRenderer) RenderPlanet(planet *entity.Planet) {
	ctx := context.Background()
	if planet == nil {
		d.logger.Debug(ctx, "RenderPlanet called with nil planet")
		return
	}
	d.logger.Debug(ctx, "RenderPlanet called",
		"planet_id", planet.ID,
		"planet_name", planet.Name,
		"team_id", planet.TeamID,
	)
}

// RenderProjectile implements entity.Renderer.
func (d *NullRenderer) RenderProjectile(projectile *entity.Projectile) {
	ctx := context.Background()
	if projectile == nil {
		d.logger.Debug(ctx, "RenderProjectile called with nil projectile")
		return
	}
	d.logger.Debug(ctx, "RenderProjectile called",
		"projectile_id", projectile.ID,
		"projectile_type", projectile.Type,
	)
}

// RenderShip implements entity.Renderer.
func (d *NullRenderer) RenderShip(ship *entity.Ship) {
	ctx := context.Background()
	if ship == nil {
		d.logger.Debug(ctx, "RenderShip called with nil ship")
		return
	}
	d.logger.Debug(ctx, "RenderShip called",
		"ship_id", ship.ID,
		"ship_class", ship.Class,
		"team_id", ship.TeamID,
	)
}

// NullRendererInstance is a global instance of NullRenderer for convenience.
var NullRendererInstance entity.Renderer = NewNullRenderer()
