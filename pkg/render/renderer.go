// pkg/render/renderer.go
package render

import "github.com/opd-ai/go-netrek/pkg/entity"

// NullRenderer is a simple implementation of entity.Renderer.
type NullRenderer struct {
}

// Clear implements entity.Renderer.
func (d *NullRenderer) Clear() {
	panic("unimplemented")
}

// Present implements entity.Renderer.
func (d *NullRenderer) Present() {
	panic("unimplemented")
}

// RenderPlanet implements entity.Renderer.
func (d *NullRenderer) RenderPlanet(planet *entity.Planet) {
	panic("unimplemented")
}

// RenderProjectile implements entity.Renderer.
func (d *NullRenderer) RenderProjectile(projectile *entity.Projectile) {
	panic("unimplemented")
}

// RenderShip implements entity.Renderer.
func (d *NullRenderer) RenderShip(ship *entity.Ship) {
	panic("unimplemented")
}

var nullRenderer entity.Renderer = &NullRenderer{}
