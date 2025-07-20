package main

import (
	"fmt"
	"testing"

	"github.com/opd-ai/go-netrek/pkg/entity"
	"github.com/opd-ai/go-netrek/pkg/physics"
)

// Debug test to isolate FireWeapon issue
func TestDebugFireWeapon(t *testing.T) {
	ship := entity.NewShip(entity.ID(1), entity.Scout, 0, physics.Vector2D{X: 0, Y: 0})

	fmt.Printf("Ship weapons count: %d\n", len(ship.Weapons))
	fmt.Printf("Ship fuel: %d\n", ship.Fuel)

	if len(ship.Weapons) > 0 {
		weapon := ship.Weapons[0]
		fmt.Printf("Weapon name: %s\n", weapon.GetName())
		fmt.Printf("Weapon fuel cost: %d\n", weapon.GetFuelCost())
		fmt.Printf("Weapon cooldown: %v\n", weapon.GetCooldown())

		projectile := ship.FireWeapon(0)
		fmt.Printf("Projectile result: %v\n", projectile)
		if projectile != nil {
			fmt.Printf("Projectile type: %s\n", projectile.Type)
		}
	}
}
