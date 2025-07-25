// pkg/render/engo/assets.go
package engo

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/EngoEngine/engo/common"

	"github.com/opd-ai/go-netrek/pkg/entity"
)

// AssetManager handles loading and managing game assets
type AssetManager struct {
	// Ship sprites by class
	shipSprites map[entity.ShipClass]common.Drawable

	// Planet sprites by type
	planetSprites map[entity.PlanetType]common.Drawable

	// Projectile sprites by type
	projectileSprites map[string]common.Drawable

	// UI textures
	backgroundTexture common.Drawable
}

// NewAssetManager creates a new asset manager
func NewAssetManager() *AssetManager {
	return &AssetManager{
		shipSprites:       make(map[entity.ShipClass]common.Drawable),
		planetSprites:     make(map[entity.PlanetType]common.Drawable),
		projectileSprites: make(map[string]common.Drawable),
	}
}

// LoadAssets loads all game assets
func (am *AssetManager) LoadAssets() error {
	// Load ship sprites
	if err := am.loadShipSprites(); err != nil {
		return err
	}

	// Load planet sprites
	if err := am.loadPlanetSprites(); err != nil {
		return err
	}

	// Load projectile sprites
	if err := am.loadProjectileSprites(); err != nil {
		return err
	}

	// Load UI assets
	if err := am.loadUIAssets(); err != nil {
		return err
	}

	return nil
}

// loadShipSprites creates sprites for different ship classes
func (am *AssetManager) loadShipSprites() error {
	// Since we don't have image files, we'll create simple geometric shapes

	// Scout: Small triangle
	am.shipSprites[entity.Scout] = am.createShipSprite(16, 16, [][]int{
		{0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0},
		{0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0},
		{0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0},
		{0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0},
		{0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0},
		{0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0},
		{0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	})

	// Destroyer: Medium triangle with more angular design
	am.shipSprites[entity.Destroyer] = am.createShipSprite(20, 20, [][]int{
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0},
		{0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0},
		{0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0},
		{0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0},
		{0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0},
		{0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0},
		{0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	})

	// Copy for other ship classes with variations
	am.shipSprites[entity.Cruiser] = am.shipSprites[entity.Destroyer]
	am.shipSprites[entity.Battleship] = am.shipSprites[entity.Destroyer]
	am.shipSprites[entity.Assault] = am.shipSprites[entity.Destroyer]

	return nil
}

// loadPlanetSprites creates sprites for different planet types
func (am *AssetManager) loadPlanetSprites() error {
	// Create a simple circle for planets
	planetPattern := [][]int{
		{0, 0, 0, 1, 1, 1, 1, 1, 1, 0, 0, 0},
		{0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0},
		{0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0},
		{0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0},
		{0, 0, 0, 1, 1, 1, 1, 1, 1, 0, 0, 0},
	}

	am.planetSprites[entity.Homeworld] = am.createShipSprite(12, 12, planetPattern)
	am.planetSprites[entity.Industrial] = am.createShipSprite(12, 12, planetPattern)
	am.planetSprites[entity.Agricultural] = am.createShipSprite(12, 12, planetPattern)
	am.planetSprites[entity.Military] = am.createShipSprite(12, 12, planetPattern)

	return nil
}

// loadProjectileSprites creates sprites for different projectile types
func (am *AssetManager) loadProjectileSprites() error {
	// Torpedo: Small dot
	torpedoPattern := [][]int{
		{0, 1, 1, 0},
		{1, 1, 1, 1},
		{1, 1, 1, 1},
		{0, 1, 1, 0},
	}

	// Phaser: Line
	phaserPattern := [][]int{
		{1, 1},
		{1, 1},
	}

	am.projectileSprites["torpedo"] = am.createShipSprite(4, 4, torpedoPattern)
	am.projectileSprites["phaser"] = am.createShipSprite(2, 2, phaserPattern)

	return nil
}

// loadUIAssets loads UI-related assets
func (am *AssetManager) loadUIAssets() error {
	// Create a simple starfield background
	backgroundPattern := make([][]int, 64)
	for i := range backgroundPattern {
		backgroundPattern[i] = make([]int, 64)
		// Add some random stars
		if i%8 == 0 && (i/8)%3 == 0 {
			backgroundPattern[i][i%64] = 1
		}
	}

	am.backgroundTexture = am.createShipSprite(64, 64, backgroundPattern)

	return nil
}

// createShipSprite creates a sprite from a 2D pattern
func (am *AssetManager) createShipSprite(width, height int, pattern [][]int) common.Drawable {
	// Create base RGBA image
	img := am.createBaseImage(width, height)

	// Draw pattern onto the image
	am.drawPatternOnImage(img, pattern, width, height)

	// Convert to Engo-compatible texture
	return am.convertToEngoTexture(img)
}

// createBaseImage creates a transparent RGBA image with the specified dimensions.
func (am *AssetManager) createBaseImage(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{0, 0, 0, 0}}, image.Point{}, draw.Src)
	return img
}

// drawPatternOnImage draws a 2D pixel pattern onto the provided RGBA image.
func (am *AssetManager) drawPatternOnImage(img *image.RGBA, pattern [][]int, width, height int) {
	for y, row := range pattern {
		if y >= height {
			break
		}
		for x, pixel := range row {
			if x >= width {
				break
			}
			if pixel == 1 {
				img.Set(x, y, color.RGBA{255, 255, 255, 255})
			}
		}
	}
}

// convertToEngoTexture converts an RGBA image to an Engo-compatible texture.
func (am *AssetManager) convertToEngoTexture(img *image.RGBA) common.Drawable {
	bounds := img.Bounds()
	nrgbaImg := image.NewNRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			nrgbaImg.Set(x, y, img.At(x, y))
		}
	}

	texture := common.NewImageObject(nrgbaImg)
	return common.NewTextureSingle(texture)
}

// GetShipSprite returns the sprite for a ship class
func (am *AssetManager) GetShipSprite(class entity.ShipClass) common.Drawable {
	if sprite, exists := am.shipSprites[class]; exists {
		return sprite
	}
	return am.shipSprites[entity.Scout] // Default fallback
}

// GetPlanetSprite returns the sprite for a planet type
func (am *AssetManager) GetPlanetSprite(planetType entity.PlanetType) common.Drawable {
	if sprite, exists := am.planetSprites[planetType]; exists {
		return sprite
	}
	return am.planetSprites[entity.Homeworld] // Default fallback
}

// GetProjectileSprite returns the sprite for a projectile type
func (am *AssetManager) GetProjectileSprite(projectileType string) common.Drawable {
	if sprite, exists := am.projectileSprites[projectileType]; exists {
		return sprite
	}
	return am.projectileSprites["torpedo"] // Default fallback
}

// GetBackgroundTexture returns the background texture
func (am *AssetManager) GetBackgroundTexture() common.Drawable {
	return am.backgroundTexture
}
