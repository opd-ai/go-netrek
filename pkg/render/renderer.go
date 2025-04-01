// pkg/render/renderer.go
package render

import (
	"log"

	"github.com/opd-ai/go-netrek/pkg/entity"
)

// NullRenderer is a simple implementation of entity.Renderer.
type NullRenderer struct{}

// Clear implements entity.Renderer.
func (d *NullRenderer) Clear() {
	log.Println("Clear called")
}

// Present implements entity.Renderer.
func (d *NullRenderer) Present() {
	log.Println("Present called")
}

// RenderPlanet implements entity.Renderer.
func (d *NullRenderer) RenderPlanet(planet *entity.Planet) {
	log.Printf("RenderPlanet called for planet: %v", planet)
}

// RenderProjectile implements entity.Renderer.
func (d *NullRenderer) RenderProjectile(projectile *entity.Projectile) {
	log.Printf("RenderProjectile called for projectile: %v", projectile)
}

// RenderShip implements entity.Renderer.
func (d *NullRenderer) RenderShip(ship *entity.Ship) {
	log.Printf("RenderShip called for ship: %v", ship)
}

var nullRenderer entity.Renderer = &NullRenderer{}
