// pkg/render/engo/assets_font_test.go
package engo

import (
	"testing"

	"github.com/EngoEngine/engo/common"
)

// TestAssetManager_FontSystem tests the font loading and management functionality
func TestAssetManager_FontSystem(t *testing.T) {
	am := NewAssetManager()

	// Test initial state
	if am.fonts == nil {
		t.Error("fonts map should be initialized")
	}

	if len(am.fonts) != 0 {
		t.Errorf("expected empty fonts map, got %d fonts", len(am.fonts))
	}
}

// TestAssetManager_FontGetters tests all font getter methods
func TestAssetManager_FontGetters(t *testing.T) {
	am := NewAssetManager()

	// Mock load fonts for testing (without actual OpenGL context)
	am.fonts = make(map[float64]*common.Font)
	am.fonts[float64(FontSizeSmall)] = &common.Font{Size: float64(FontSizeSmall)}
	am.fonts[float64(FontSizeMedium)] = &common.Font{Size: float64(FontSizeMedium)}
	am.fonts[float64(FontSizeLarge)] = &common.Font{Size: float64(FontSizeLarge)}
	am.fonts[float64(FontSizeXLarge)] = &common.Font{Size: float64(FontSizeXLarge)}

	tests := []struct {
		name         string
		method       func() *common.Font
		expectedSize float64
	}{
		{
			name:         "GetSmallFont",
			method:       am.GetSmallFont,
			expectedSize: float64(FontSizeSmall),
		},
		{
			name:         "GetMediumFont",
			method:       am.GetMediumFont,
			expectedSize: float64(FontSizeMedium),
		},
		{
			name:         "GetLargeFont",
			method:       am.GetLargeFont,
			expectedSize: float64(FontSizeLarge),
		},
		{
			name:         "GetXLargeFont",
			method:       am.GetXLargeFont,
			expectedSize: float64(FontSizeXLarge),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			font := tt.method()
			if font == nil {
				t.Errorf("%s returned nil font", tt.name)
				return
			}

			if font.Size != tt.expectedSize {
				t.Errorf("%s returned font with size %f, expected %f",
					tt.name, font.Size, tt.expectedSize)
			}
		})
	}
}

// TestAssetManager_GetFont_WithCustomSize tests the generic GetFont method
func TestAssetManager_GetFont_WithCustomSize(t *testing.T) {
	am := NewAssetManager()

	// Mock fonts
	am.fonts = make(map[float64]*common.Font)
	am.fonts[float64(FontSizeMedium)] = &common.Font{Size: float64(FontSizeMedium)}
	am.fonts[18.0] = &common.Font{Size: 18.0}

	tests := []struct {
		name           string
		requestedSize  float64
		expectedSize   float64
		shouldFallback bool
	}{
		{
			name:          "existing_medium_font",
			requestedSize: float64(FontSizeMedium),
			expectedSize:  float64(FontSizeMedium),
		},
		{
			name:          "existing_custom_font",
			requestedSize: 18.0,
			expectedSize:  18.0,
		},
		{
			name:           "non_existing_font_fallback",
			requestedSize:  99.0,
			expectedSize:   float64(FontSizeMedium),
			shouldFallback: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			font := am.GetFont(tt.requestedSize)
			if font == nil {
				t.Error("GetFont returned nil")
				return
			}

			if font.Size != tt.expectedSize {
				t.Errorf("GetFont(%f) returned font with size %f, expected %f",
					tt.requestedSize, font.Size, tt.expectedSize)
			}
		})
	}
}

// TestAssetManager_GetFont_EmptyMap tests fallback behavior with empty font map
func TestAssetManager_GetFont_EmptyMap(t *testing.T) {
	am := NewAssetManager()

	// Ensure fonts map is empty
	am.fonts = make(map[float64]*common.Font)

	font := am.GetFont(float64(FontSizeMedium))
	if font != nil {
		t.Error("GetFont should return nil when no fonts are loaded and fallback doesn't exist")
	}
}

// TestAssetManager_LoadFonts_ExpectFailure tests font loading without OpenGL context
func TestAssetManager_LoadFonts_ExpectFailure(t *testing.T) {
	am := NewAssetManager()

	// This should fail gracefully without OpenGL context
	err := am.loadFonts()

	// In unit tests without OpenGL, this will likely fail
	// but the implementation should handle it gracefully
	if err == nil {
		t.Log("loadFonts succeeded unexpectedly (maybe running in OpenGL context)")

		// If it succeeded, verify all expected fonts were created
		expectedSizes := []float64{
			float64(FontSizeSmall),
			float64(FontSizeMedium),
			float64(FontSizeLarge),
			float64(FontSizeXLarge),
		}

		for _, size := range expectedSizes {
			if font, exists := am.fonts[size]; !exists {
				t.Errorf("Expected font size %f not found in fonts map", size)
			} else if font.Size != size {
				t.Errorf("Font size mismatch: expected %f, got %f", size, font.Size)
			}
		}
	} else {
		t.Logf("loadFonts failed as expected in unit test environment: %v", err)
		t.Log("In a real OpenGL environment, loadFonts should populate:")
		t.Log("- fonts map with sizes 12, 16, 20, 24")
		t.Log("- each font with proper size, white foreground, transparent background")
	}
}

// TestAssetManager_FontSystem_Integration tests the integration between AssetManager and font system
func TestAssetManager_FontSystem_Integration(t *testing.T) {
	am := NewAssetManager()

	// Test that fonts map is initialized
	if am.fonts == nil {
		t.Error("fonts map should be initialized in NewAssetManager")
	}

	// Test initial empty state
	if len(am.fonts) != 0 {
		t.Errorf("fonts map should be empty initially, got %d fonts", len(am.fonts))
	}

	// Test font getter methods with empty map (should return nil gracefully)
	testMethods := []struct {
		name   string
		method func() *common.Font
	}{
		{"GetSmallFont", am.GetSmallFont},
		{"GetMediumFont", am.GetMediumFont},
		{"GetLargeFont", am.GetLargeFont},
		{"GetXLargeFont", am.GetXLargeFont},
	}

	for _, test := range testMethods {
		t.Run(test.name+"_empty_map", func(t *testing.T) {
			font := test.method()
			if font != nil {
				t.Errorf("%s should return nil when fonts map is empty", test.name)
			}
		})
	}
}

// TestAssetManager_FontConstants tests that font sizes match design constants
func TestAssetManager_FontConstants(t *testing.T) {
	expectedSizes := map[string]int{
		"FontSizeSmall":  FontSizeSmall,
		"FontSizeMedium": FontSizeMedium,
		"FontSizeLarge":  FontSizeLarge,
		"FontSizeXLarge": FontSizeXLarge,
	}

	for name, expectedSize := range expectedSizes {
		t.Run(name, func(t *testing.T) {
			if expectedSize <= 0 {
				t.Errorf("%s should be positive, got %d", name, expectedSize)
			}

			// Verify sizes follow logical progression
			switch name {
			case "FontSizeSmall":
				if expectedSize >= FontSizeMedium {
					t.Errorf("FontSizeSmall (%d) should be smaller than FontSizeMedium (%d)",
						expectedSize, FontSizeMedium)
				}
			case "FontSizeMedium":
				if expectedSize <= FontSizeSmall || expectedSize >= FontSizeLarge {
					t.Errorf("FontSizeMedium (%d) should be between Small (%d) and Large (%d)",
						expectedSize, FontSizeSmall, FontSizeLarge)
				}
			case "FontSizeLarge":
				if expectedSize <= FontSizeMedium || expectedSize >= FontSizeXLarge {
					t.Errorf("FontSizeLarge (%d) should be between Medium (%d) and XLarge (%d)",
						expectedSize, FontSizeMedium, FontSizeXLarge)
				}
			case "FontSizeXLarge":
				if expectedSize <= FontSizeLarge {
					t.Errorf("FontSizeXLarge (%d) should be larger than FontSizeLarge (%d)",
						expectedSize, FontSizeLarge)
				}
			}
		})
	}
}
