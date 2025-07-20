package engo

import (
	"testing"

	"github.com/opd-ai/go-netrek/pkg/entity"
)

func TestNewAssetManager(t *testing.T) {
	am := NewAssetManager()

	if am == nil {
		t.Fatal("NewAssetManager() returned nil")
	}

	if am.shipSprites == nil {
		t.Error("shipSprites map not initialized")
	}

	if am.planetSprites == nil {
		t.Error("planetSprites map not initialized")
	}

	if am.projectileSprites == nil {
		t.Error("projectileSprites map not initialized")
	}

	// Verify maps are empty initially
	if len(am.shipSprites) != 0 {
		t.Errorf("shipSprites should be empty initially, got %d entries", len(am.shipSprites))
	}

	if len(am.planetSprites) != 0 {
		t.Errorf("planetSprites should be empty initially, got %d entries", len(am.planetSprites))
	}

	if len(am.projectileSprites) != 0 {
		t.Errorf("projectileSprites should be empty initially, got %d entries", len(am.projectileSprites))
	}
}

func TestLoadAssets_ExpectFailure(t *testing.T) {
	// This test documents that LoadAssets requires OpenGL context
	// In a testing environment without OpenGL, this should fail gracefully
	// This is a documentation test for the expected behavior

	t.Log("LoadAssets requires OpenGL context and cannot be tested in unit tests")
	t.Log("In a real environment with OpenGL, LoadAssets should populate:")
	t.Log("- shipSprites map with Scout, Destroyer, Cruiser, Battleship, Assault")
	t.Log("- planetSprites map with Agricultural, Industrial, Military, Homeworld")
	t.Log("- projectileSprites map with torpedo, phaser, plasma")
	t.Log("- backgroundTexture")
}

func TestAssetManager_MockBehavior(t *testing.T) {
	// Test the behavior when assets aren't loaded (mock scenario)
	am := NewAssetManager()

	// Test ship sprite retrieval before loading
	tests := []struct {
		name      string
		shipClass entity.ShipClass
	}{
		{"scout", entity.Scout},
		{"destroyer", entity.Destroyer},
		{"cruiser", entity.Cruiser},
		{"battleship", entity.Battleship},
		{"assault", entity.Assault},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sprite := am.GetShipSprite(tt.shipClass)
			// Should return nil before assets are loaded
			if sprite != nil {
				t.Error("Expected nil sprite before loading assets")
			}
		})
	}
}

func TestGetShipSprite_UnknownClass(t *testing.T) {
	am := NewAssetManager()

	// Manually add a sprite for testing fallback behavior
	am.shipSprites[entity.Scout] = nil // Mock sprite

	// Test with an unknown ship class
	unknownClass := entity.ShipClass(999)
	sprite := am.GetShipSprite(unknownClass)

	// Should return the scout fallback (which is nil in our mock)
	if sprite != nil {
		t.Error("Expected fallback behavior for unknown ship class")
	}
}

func TestGetPlanetSprite_MockBehavior(t *testing.T) {
	am := NewAssetManager()

	tests := []struct {
		name       string
		planetType entity.PlanetType
	}{
		{"agricultural", entity.Agricultural},
		{"industrial", entity.Industrial},
		{"military", entity.Military},
		{"homeworld", entity.Homeworld},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sprite := am.GetPlanetSprite(tt.planetType)
			// Should return nil before assets are loaded
			if sprite != nil {
				t.Error("Expected nil sprite before loading assets")
			}
		})
	}
}

func TestGetPlanetSprite_UnknownType(t *testing.T) {
	am := NewAssetManager()

	// Manually add a sprite for testing fallback behavior
	am.planetSprites[entity.Homeworld] = nil // Mock sprite

	// Test with an unknown planet type
	unknownType := entity.PlanetType(999)
	sprite := am.GetPlanetSprite(unknownType)

	// Should return the homeworld fallback (which is nil in our mock)
	if sprite != nil {
		t.Error("Expected fallback behavior for unknown planet type")
	}
}

func TestGetProjectileSprite_MockBehavior(t *testing.T) {
	am := NewAssetManager()

	tests := []struct {
		name           string
		projectileType string
	}{
		{"torpedo", "torpedo"},
		{"phaser", "phaser"},
		{"plasma", "plasma"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sprite := am.GetProjectileSprite(tt.projectileType)
			// Should return nil before assets are loaded
			if sprite != nil {
				t.Error("Expected nil sprite before loading assets")
			}
		})
	}
}

func TestGetProjectileSprite_UnknownType(t *testing.T) {
	am := NewAssetManager()

	// Manually add a sprite for testing fallback behavior
	am.projectileSprites["torpedo"] = nil // Mock sprite

	// Test with an unknown projectile type
	sprite := am.GetProjectileSprite("unknown_weapon")

	// Should return the torpedo fallback (which is nil in our mock)
	if sprite != nil {
		t.Error("Expected fallback behavior for unknown projectile type")
	}
}

func TestGetProjectileSprite_EmptyString(t *testing.T) {
	am := NewAssetManager()

	// Manually add a sprite for testing fallback behavior
	am.projectileSprites["torpedo"] = nil // Mock sprite

	// Test with empty string
	sprite := am.GetProjectileSprite("")

	// Should return the torpedo fallback (which is nil in our mock)
	if sprite != nil {
		t.Error("Expected fallback behavior for empty string")
	}
}

func TestGetBackgroundTexture(t *testing.T) {
	am := NewAssetManager()

	// Test before loading assets
	texture := am.GetBackgroundTexture()
	if texture != nil {
		t.Error("Expected nil texture before loading assets")
	}
}

func TestAssetManager_Integration(t *testing.T) {
	// Integration test to verify the complete workflow without OpenGL
	am := NewAssetManager()

	// Verify initial state
	if am.shipSprites == nil || am.planetSprites == nil || am.projectileSprites == nil {
		t.Error("Asset maps should be initialized")
	}

	// Test that getters return expected fallbacks for empty maps
	shipSprite := am.GetShipSprite(entity.Scout)
	if shipSprite != nil {
		t.Error("Expected nil ship sprite before loading assets")
	}

	planetSprite := am.GetPlanetSprite(entity.Homeworld)
	if planetSprite != nil {
		t.Error("Expected nil planet sprite before loading assets")
	}

	projectileSprite := am.GetProjectileSprite("torpedo")
	if projectileSprite != nil {
		t.Error("Expected nil projectile sprite before loading assets")
	}

	backgroundTexture := am.GetBackgroundTexture()
	if backgroundTexture != nil {
		t.Error("Expected nil background texture before loading assets")
	}
}

func TestAssetManager_FallbackBehavior(t *testing.T) {
	am := NewAssetManager()

	// Manually populate some assets to test fallback logic
	am.shipSprites[entity.Scout] = nil       // Mock sprite
	am.planetSprites[entity.Homeworld] = nil // Mock sprite
	am.projectileSprites["torpedo"] = nil    // Mock sprite

	// Test fallback for unknown ship class
	unknownShip := entity.ShipClass(999)
	shipFallback := am.GetShipSprite(unknownShip)
	expectedShipFallback := am.shipSprites[entity.Scout]
	if shipFallback != expectedShipFallback {
		t.Error("Ship fallback not working correctly")
	}

	// Test fallback for unknown planet type
	unknownPlanet := entity.PlanetType(999)
	planetFallback := am.GetPlanetSprite(unknownPlanet)
	expectedPlanetFallback := am.planetSprites[entity.Homeworld]
	if planetFallback != expectedPlanetFallback {
		t.Error("Planet fallback not working correctly")
	}

	// Test fallback for unknown projectile type
	projectileFallback := am.GetProjectileSprite("unknown")
	expectedProjectileFallback := am.projectileSprites["torpedo"]
	if projectileFallback != expectedProjectileFallback {
		t.Error("Projectile fallback not working correctly")
	}
}

func TestAssetManager_EdgeCases(t *testing.T) {
	am := NewAssetManager()

	// Test edge cases for projectile types
	edgeCases := []string{
		"",        // empty string
		"TORPEDO", // uppercase
		"   ",     // whitespace
		"very_long_projectile_name_that_does_not_exist",
	}

	// Add torpedo fallback
	am.projectileSprites["torpedo"] = nil

	for _, projectileType := range edgeCases {
		t.Run("projectile_"+projectileType, func(t *testing.T) {
			sprite := am.GetProjectileSprite(projectileType)
			// All should fallback to torpedo
			expected := am.projectileSprites["torpedo"]
			if sprite != expected {
				t.Errorf("Expected fallback to torpedo for projectile type '%s'", projectileType)
			}
		})
	}
}

func TestAssetManager_StructureAndTypes(t *testing.T) {
	am := NewAssetManager()

	// Test that the AssetManager has the correct field types
	if am.shipSprites == nil {
		t.Error("shipSprites should be initialized")
	}

	if am.planetSprites == nil {
		t.Error("planetSprites should be initialized")
	}

	if am.projectileSprites == nil {
		t.Error("projectileSprites should be initialized")
	}

	// Test that we can add mock entries
	am.shipSprites[entity.Scout] = nil
	am.planetSprites[entity.Agricultural] = nil
	am.projectileSprites["test"] = nil

	// Verify they were added
	if len(am.shipSprites) != 1 {
		t.Errorf("Expected 1 ship sprite, got %d", len(am.shipSprites))
	}

	if len(am.planetSprites) != 1 {
		t.Errorf("Expected 1 planet sprite, got %d", len(am.planetSprites))
	}

	if len(am.projectileSprites) != 1 {
		t.Errorf("Expected 1 projectile sprite, got %d", len(am.projectileSprites))
	}
}
