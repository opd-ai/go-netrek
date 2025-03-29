package entity

// Renderer handles rendering game entities
type Renderer interface {
	RenderShip(ship *Ship)
	RenderPlanet(planet *Planet)
	RenderProjectile(projectile *Projectile)
	Clear()
	Present()
}
