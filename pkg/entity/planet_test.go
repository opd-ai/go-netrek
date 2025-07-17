// planet_test.go
package entity

import (
	"testing"

	"github.com/opd-ai/go-netrek/pkg/physics"
)

func TestNewPlanet_DefaultsAndTypes(t *testing.T) {
	tests := []struct {
		name           string
		ptype          PlanetType
		wantArmies     int
		wantMaxArmies  int
		wantProduction int
		wantResources  int
	}{
		{"Agricultural", Agricultural, 10, 80, 8, 100},
		{"Industrial", Industrial, 10, 100, 10, 200},
		{"Military", Military, 20, 120, 5, 100},
		{"Homeworld", Homeworld, 30, 200, 12, 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPlanet(1, "Earth", physics.Vector2D{X: 0, Y: 0}, tt.ptype)
			if p.Armies != tt.wantArmies {
				t.Errorf("Armies: got %d, want %d", p.Armies, tt.wantArmies)
			}
			if p.MaxArmies != tt.wantMaxArmies {
				t.Errorf("MaxArmies: got %d, want %d", p.MaxArmies, tt.wantMaxArmies)
			}
			if p.Production != tt.wantProduction {
				t.Errorf("Production: got %d, want %d", p.Production, tt.wantProduction)
			}
			if tt.ptype == Industrial && p.Resources != tt.wantResources {
				t.Errorf("Resources: got %d, want %d", p.Resources, tt.wantResources)
			}
		})
	}
}

func TestPlanet_Update_Production(t *testing.T) {
	p := NewPlanet(1, "Mars", physics.Vector2D{X: 0, Y: 0}, Agricultural)
	p.TeamID = 2
	p.Armies = 10
	p.MaxArmies = 15
	p.Production = 10
	p.Update(1.0)
	if p.Armies <= 10 {
		t.Errorf("Update did not increase armies as expected, got %d", p.Armies)
	}
	// Should not exceed MaxArmies
	p.Armies = p.MaxArmies - 1
	p.Update(10.0)
	if p.Armies > p.MaxArmies {
		t.Errorf("Update exceeded MaxArmies, got %d", p.Armies)
	}
}

func TestPlanet_Update_NoProductionWhenNeutral(t *testing.T) {
	p := NewPlanet(2, "Pluto", physics.Vector2D{X: 0, Y: 0}, Agricultural)
	p.TeamID = -1 // Neutral
	p.Armies = 5
	p.MaxArmies = 10
	p.Production = 10
	p.Update(1.0)
	if p.Armies != 5 {
		t.Errorf("Neutral planet should not produce armies")
	}
}

func TestPlanet_Update_MaxArmiesBoundary(t *testing.T) {
	p := NewPlanet(3, "Mercury", physics.Vector2D{X: 0, Y: 0}, Agricultural)
	p.TeamID = 1
	p.Armies = p.MaxArmies
	p.Update(1.0)
	if p.Armies != p.MaxArmies {
		t.Errorf("Armies should not exceed MaxArmies")
	}
}

func TestPlanet_Bomb_Scenarios(t *testing.T) {
	p := NewPlanet(1, "Venus", physics.Vector2D{X: 0, Y: 0}, Military)
	p.TeamID = 1
	p.Armies = 10
	p.Atmosphere = true
	p.Temperature = 50
	// Bomb with damage
	damage := p.Bomb(10)
	if damage <= 0 || damage > 10 {
		t.Errorf("Bomb damage out of expected range: %d", damage)
	}
	if p.Armies >= 10 {
		t.Errorf("Bomb did not reduce armies")
	}
	// Bomb until neutral
	p.Armies = 1
	p.TeamID = 2
	_ = p.Bomb(5)
	if p.TeamID != -1 {
		t.Errorf("Planet should become neutral when armies reach 0")
	}
	// Bomb with zero armies
	p.Armies = 0
	if p.Bomb(5) != 0 {
		t.Errorf("Bomb should do zero damage when armies are zero")
	}
}

func TestPlanet_Bomb_AtmosphereAndTemperature(t *testing.T) {
	p := NewPlanet(4, "Neptune", physics.Vector2D{X: 0, Y: 0}, Military)
	p.TeamID = 1
	p.Armies = 10
	p.Atmosphere = true
	p.Temperature = 100 // Max temp
	damage := p.Bomb(10)
	if damage < 1 {
		t.Errorf("Bomb should always do at least 1 damage")
	}
}

func TestPlanet_Bomb_DamageGreaterThanArmies(t *testing.T) {
	p := NewPlanet(5, "Uranus", physics.Vector2D{X: 0, Y: 0}, Military)
	p.TeamID = 1
	p.Armies = 3
	damage := p.Bomb(100)
	if damage != 3 {
		t.Errorf("Bomb should not do more damage than available armies")
	}
}

func TestPlanet_BeamDownArmies_CaptureAndFriendly(t *testing.T) {
	p := NewPlanet(1, "Jupiter", physics.Vector2D{X: 0, Y: 0}, Industrial)
	p.TeamID = -1
	p.Armies = 0
	p.MaxArmies = 10
	// Capture neutral planet
	transferred, captured := p.BeamDownArmies(1, 5)
	if !captured || transferred != 5 || p.TeamID != 1 {
		t.Errorf("BeamDownArmies failed to capture neutral planet")
	}
	// Beam to friendly planet
	p.TeamID = 1
	p.Armies = 5
	transferred, _ = p.BeamDownArmies(1, 10)
	if transferred != 5 {
		t.Errorf("BeamDownArmies should transfer only available space")
	}
	// Beam to enemy planet
	p.TeamID = 2
	transferred, captured = p.BeamDownArmies(1, 5)
	if transferred != 0 || captured {
		t.Errorf("BeamDownArmies should not transfer to enemy planet")
	}
}

func TestPlanet_BeamDownArmies_ZeroAmount(t *testing.T) {
	p := NewPlanet(6, "Ceres", physics.Vector2D{X: 0, Y: 0}, Industrial)
	p.TeamID = -1
	p.Armies = 0
	transferred, captured := p.BeamDownArmies(1, 0)
	if transferred != 0 || captured {
		t.Errorf("BeamDownArmies with zero amount should not capture or transfer")
	}
}

func TestPlanet_BeamDownArmies_Overfill(t *testing.T) {
	p := NewPlanet(7, "Vesta", physics.Vector2D{X: 0, Y: 0}, Industrial)
	p.TeamID = 1
	p.Armies = p.MaxArmies - 1
	transferred, _ := p.BeamDownArmies(1, 10)
	if transferred != 1 {
		t.Errorf("BeamDownArmies should not overfill armies")
	}
}

func TestPlanet_BeamUpArmies_FriendlyAndLimits(t *testing.T) {
	p := NewPlanet(1, "Saturn", physics.Vector2D{X: 0, Y: 0}, Homeworld)
	p.TeamID = 1
	p.Armies = 10
	// Beam up less than available
	amount := p.BeamUpArmies(1, 5)
	if amount != 5 || p.Armies != 5 {
		t.Errorf("BeamUpArmies failed to transfer correct amount")
	}
	// Beam up more than available
	amount = p.BeamUpArmies(1, 10)
	if amount != 5 || p.Armies != 0 {
		t.Errorf("BeamUpArmies should transfer only available armies")
	}
	// Beam up from enemy planet
	p.TeamID = 2
	p.Armies = 10
	amount = p.BeamUpArmies(1, 5)
	if amount != 0 {
		t.Errorf("BeamUpArmies should not transfer from enemy planet")
	}
	// Beam up when no armies
	p.TeamID = 1
	p.Armies = 0
	amount = p.BeamUpArmies(1, 5)
	if amount != 0 {
		t.Errorf("BeamUpArmies should not transfer when no armies")
	}
}

func TestPlanet_BeamUpArmies_ZeroMaxAmount(t *testing.T) {
	p := NewPlanet(8, "Pallas", physics.Vector2D{X: 0, Y: 0}, Homeworld)
	p.TeamID = 1
	p.Armies = 5
	amount := p.BeamUpArmies(1, 0)
	if amount != 0 {
		t.Errorf("BeamUpArmies with zero maxAmount should transfer nothing")
	}
}
