// pkg/render/engo/hud_font_test.go
package engo

import (
	"testing"

	"github.com/EngoEngine/engo/common"
)

// TestNewHUDSystem_WithAssetManager tests HUD system creation with asset manager
func TestNewHUDSystem_WithAssetManager(t *testing.T) {
	// Create mock asset manager with fonts
	am := NewAssetManager()
	am.fonts = make(map[float64]*common.Font)
	am.fonts[float64(FontSizeMedium)] = &common.Font{Size: float64(FontSizeMedium)}

	hud := NewHUDSystem(am)

	if hud == nil {
		t.Fatal("NewHUDSystem returned nil")
	}

	if hud.assetManager != am {
		t.Error("HUD assetManager not set correctly")
	}

	if hud.font == nil {
		t.Error("HUD font should be initialized from asset manager")
	}

	if hud.font.Size != float64(FontSizeMedium) {
		t.Errorf("HUD font size should be %d, got %f", FontSizeMedium, hud.font.Size)
	}
}

// TestNewHUDSystem_WithNilAssetManager tests HUD system creation with nil asset manager
func TestNewHUDSystem_WithNilAssetManager(t *testing.T) {
	hud := NewHUDSystem(nil)

	if hud == nil {
		t.Fatal("NewHUDSystem returned nil")
	}

	if hud.assetManager != nil {
		t.Error("HUD assetManager should be nil")
	}

	if hud.font != nil {
		t.Error("HUD font should be nil when asset manager is nil")
	}
}

// TestHUDSystem_SetFont tests the SetFont method
func TestHUDSystem_SetFont(t *testing.T) {
	hud := NewHUDSystem(nil)

	testFont := &common.Font{Size: 18.0}
	hud.SetFont(testFont)

	if hud.font != testFont {
		t.Error("SetFont did not set the font correctly")
	}

	if hud.font.Size != 18.0 {
		t.Errorf("Font size should be 18.0, got %f", hud.font.Size)
	}
}

// TestHUDSystem_FontSizeMethods tests all font size setting methods
func TestHUDSystem_FontSizeMethods(t *testing.T) {
	// Create mock asset manager with all font sizes
	am := NewAssetManager()
	am.fonts = make(map[float64]*common.Font)
	am.fonts[float64(FontSizeSmall)] = &common.Font{Size: float64(FontSizeSmall)}
	am.fonts[float64(FontSizeMedium)] = &common.Font{Size: float64(FontSizeMedium)}
	am.fonts[float64(FontSizeLarge)] = &common.Font{Size: float64(FontSizeLarge)}

	hud := NewHUDSystem(am)

	tests := []struct {
		name         string
		method       func()
		expectedSize float64
	}{
		{
			name:         "SetSmallFont",
			method:       hud.SetSmallFont,
			expectedSize: float64(FontSizeSmall),
		},
		{
			name:         "SetMediumFont",
			method:       hud.SetMediumFont,
			expectedSize: float64(FontSizeMedium),
		},
		{
			name:         "SetLargeFont",
			method:       hud.SetLargeFont,
			expectedSize: float64(FontSizeLarge),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.method()

			if hud.font == nil {
				t.Errorf("%s should set a font", tt.name)
				return
			}

			if hud.font.Size != tt.expectedSize {
				t.Errorf("%s should set font size to %f, got %f",
					tt.name, tt.expectedSize, hud.font.Size)
			}
		})
	}
}

// TestHUDSystem_SetFontSize tests the generic SetFontSize method
func TestHUDSystem_SetFontSize(t *testing.T) {
	// Create mock asset manager
	am := NewAssetManager()
	am.fonts = make(map[float64]*common.Font)
	am.fonts[18.0] = &common.Font{Size: 18.0}
	am.fonts[22.0] = &common.Font{Size: 22.0}

	hud := NewHUDSystem(am)

	// Test setting existing font size
	hud.SetFontSize(18.0)
	if hud.font.Size != 18.0 {
		t.Errorf("SetFontSize(18.0) should set font size to 18.0, got %f", hud.font.Size)
	}

	// Test setting another existing font size
	hud.SetFontSize(22.0)
	if hud.font.Size != 22.0 {
		t.Errorf("SetFontSize(22.0) should set font size to 22.0, got %f", hud.font.Size)
	}
}

// TestHUDSystem_SetFontSize_WithNilAssetManager tests SetFontSize with nil asset manager
func TestHUDSystem_SetFontSize_WithNilAssetManager(t *testing.T) {
	hud := NewHUDSystem(nil)

	// This should not panic even with nil asset manager
	hud.SetFontSize(18.0)

	// Font should remain nil
	if hud.font != nil {
		t.Error("Font should remain nil when asset manager is nil")
	}
}

// TestHUDSystem_FontMethods_WithNilAssetManager tests font size methods with nil asset manager
func TestHUDSystem_FontMethods_WithNilAssetManager(t *testing.T) {
	hud := NewHUDSystem(nil)

	// These should not panic even with nil asset manager
	methods := []struct {
		name   string
		method func()
	}{
		{"SetSmallFont", hud.SetSmallFont},
		{"SetMediumFont", hud.SetMediumFont},
		{"SetLargeFont", hud.SetLargeFont},
	}

	for _, test := range methods {
		t.Run(test.name, func(t *testing.T) {
			// Should not panic
			test.method()

			// Font should remain nil
			if hud.font != nil {
				t.Errorf("%s should not set font when asset manager is nil", test.name)
			}
		})
	}
}

// TestHUDSystem_FontIntegration tests integration between HUD and asset manager fonts
func TestHUDSystem_FontIntegration(t *testing.T) {
	// Create asset manager and load mock fonts
	am := NewAssetManager()
	am.fonts = make(map[float64]*common.Font)

	// Add all font sizes
	fontSizes := []float64{
		float64(FontSizeSmall),
		float64(FontSizeMedium),
		float64(FontSizeLarge),
		float64(FontSizeXLarge),
	}

	for _, size := range fontSizes {
		am.fonts[size] = &common.Font{Size: size}
	}

	// Create HUD
	hud := NewHUDSystem(am)

	// Should start with medium font
	if hud.font.Size != float64(FontSizeMedium) {
		t.Errorf("Initial font should be medium size (%d), got %f",
			FontSizeMedium, hud.font.Size)
	}

	// Test cycling through different font sizes
	hud.SetSmallFont()
	if hud.font.Size != float64(FontSizeSmall) {
		t.Errorf("After SetSmallFont, size should be %d, got %f",
			FontSizeSmall, hud.font.Size)
	}

	hud.SetLargeFont()
	if hud.font.Size != float64(FontSizeLarge) {
		t.Errorf("After SetLargeFont, size should be %d, got %f",
			FontSizeLarge, hud.font.Size)
	}

	hud.SetMediumFont()
	if hud.font.Size != float64(FontSizeMedium) {
		t.Errorf("After SetMediumFont, size should be %d, got %f",
			FontSizeMedium, hud.font.Size)
	}
}

// TestHUDSystem_FontFallback tests font fallback behavior
func TestHUDSystem_FontFallback(t *testing.T) {
	// Create asset manager with only medium font
	am := NewAssetManager()
	am.fonts = make(map[float64]*common.Font)
	am.fonts[float64(FontSizeMedium)] = &common.Font{Size: float64(FontSizeMedium)}

	hud := NewHUDSystem(am)

	// Try to set a font size that doesn't exist
	hud.SetFontSize(99.0)

	// Should fall back to medium font
	if hud.font.Size != float64(FontSizeMedium) {
		t.Errorf("Should fall back to medium font size (%d), got %f",
			FontSizeMedium, hud.font.Size)
	}
}
