package engo

import (
	"fmt"
	"image/color"
	"testing"
)

// TestSpacingConstants verifies that spacing constants follow the 8-point grid system
func TestSpacingConstants(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		expected int
	}{
		{"MarginXSmall", MarginXSmall, 4},
		{"MarginSmall", MarginSmall, 8},
		{"MarginMedium", MarginMedium, 16},
		{"MarginLarge", MarginLarge, 24},
		{"MarginXLarge", MarginXLarge, 32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.expected {
				t.Errorf("%s = %d, want %d", tt.name, tt.value, tt.expected)
			}
		})
	}

	// Verify 8-point grid system progression
	if MarginSmall != MarginXSmall*2 {
		t.Errorf("MarginSmall should be 2x MarginXSmall, got %d", MarginSmall)
	}
	if MarginMedium != MarginSmall*2 {
		t.Errorf("MarginMedium should be 2x MarginSmall, got %d", MarginMedium)
	}
}

// TestTypographyConstants verifies font size constants are appropriate
func TestTypographyConstants(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		expected int
	}{
		{"FontSizeSmall", FontSizeSmall, 12},
		{"FontSizeMedium", FontSizeMedium, 16},
		{"FontSizeLarge", FontSizeLarge, 20},
		{"FontSizeXLarge", FontSizeXLarge, 24},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.expected {
				t.Errorf("%s = %d, want %d", tt.name, tt.value, tt.expected)
			}
		})
	}

	// Verify font sizes are reasonable
	if FontSizeSmall < 10 || FontSizeSmall > 14 {
		t.Errorf("FontSizeSmall (%d) should be between 10-14 for readability", FontSizeSmall)
	}
	if FontSizeMedium < 14 || FontSizeMedium > 18 {
		t.Errorf("FontSizeMedium (%d) should be between 14-18 for readability", FontSizeMedium)
	}
}

// TestLayoutDimensionConstants verifies layout dimensions are reasonable
func TestLayoutDimensionConstants(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		expected int
	}{
		{"ChatWindowHeight", ChatWindowHeight, 150},
		{"ChatWindowWidth", ChatWindowWidth, 400},
		{"ConnectionStatusWidth", ConnectionStatusWidth, 150},
		{"MinimapDefaultSize", MinimapDefaultSize, 200},
		{"MaxChatLines", MaxChatLines, 10},
		{"LineHeight", LineHeight, 20},
		{"ChatLineHeight", ChatLineHeight, 15},
		{"TeamStatusLineHeight", TeamStatusLineHeight, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.expected {
				t.Errorf("%s = %d, want %d", tt.name, tt.value, tt.expected)
			}
		})
	}

	// Verify dimensions are positive and reasonable
	if ChatWindowHeight <= 0 || ChatWindowHeight > 500 {
		t.Errorf("ChatWindowHeight (%d) should be positive and reasonable", ChatWindowHeight)
	}
	if ChatWindowWidth <= 0 || ChatWindowWidth > 800 {
		t.Errorf("ChatWindowWidth (%d) should be positive and reasonable", ChatWindowWidth)
	}
	if MaxChatLines <= 0 || MaxChatLines > 50 {
		t.Errorf("MaxChatLines (%d) should be positive and reasonable", MaxChatLines)
	}
}

// TestAlphaConstants verifies alpha transparency values
func TestAlphaConstants(t *testing.T) {
	tests := []struct {
		name     string
		value    uint8
		expected uint8
	}{
		{"AlphaTransparent", AlphaTransparent, 0},
		{"AlphaSemiTransparent", AlphaSemiTransparent, 64},
		{"AlphaBackground", AlphaBackground, 128},
		{"AlphaSubtle", AlphaSubtle, 192},
		{"AlphaOpaque", AlphaOpaque, 255},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.expected {
				t.Errorf("%s = %d, want %d", tt.name, tt.value, tt.expected)
			}
		})
	}

	// Verify alpha progression makes sense
	alphas := []uint8{AlphaTransparent, AlphaSemiTransparent, AlphaBackground, AlphaSubtle, AlphaOpaque}
	for i := 1; i < len(alphas); i++ {
		if alphas[i] <= alphas[i-1] {
			t.Errorf("Alpha values should be in ascending order, but alphas[%d]=%d <= alphas[%d]=%d",
				i, alphas[i], i-1, alphas[i-1])
		}
	}
}

// TestColorConstants verifies color definitions are valid RGBA values
func TestColorConstants(t *testing.T) {
	colors := map[string]color.Color{
		"ColorTextPrimary":     ColorTextPrimary,
		"ColorTextSecondary":   ColorTextSecondary,
		"ColorTextDisabled":    ColorTextDisabled,
		"ColorBackgroundChat":  ColorBackgroundChat,
		"ColorBackgroundPanel": ColorBackgroundPanel,
		"ColorTeamRed":         ColorTeamRed,
		"ColorTeamGreen":       ColorTeamGreen,
		"ColorTeamBlue":        ColorTeamBlue,
		"ColorTeamYellow":      ColorTeamYellow,
		"ColorNeutral":         ColorNeutral,
		"ColorEntityDefault":   ColorEntityDefault,
		"ColorEntityHighlight": ColorEntityHighlight,
	}

	for name, col := range colors {
		t.Run(name, func(t *testing.T) {
			rgba, ok := col.(color.RGBA)
			if !ok {
				t.Errorf("%s is not a color.RGBA type", name)
				return
			}

			// Basic validation - all RGBA values are valid by type definition (uint8 is 0-255)
			// Just verify we can access the color components
			_ = rgba.R + rgba.G + rgba.B + rgba.A
		})
	}
}

// TestGetTeamColor verifies team color assignment logic
func TestGetTeamColor(t *testing.T) {
	tests := []struct {
		teamID   int
		expected color.Color
	}{
		{0, ColorTeamRed},
		{1, ColorTeamGreen},
		{2, ColorTeamBlue},
		{3, ColorTeamYellow},
		{4, ColorNeutral},   // Invalid team ID
		{-1, ColorNeutral},  // Invalid team ID
		{999, ColorNeutral}, // Invalid team ID
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("teamID_%d", tt.teamID), func(t *testing.T) {
			result := GetTeamColor(tt.teamID)
			if result != tt.expected {
				t.Errorf("GetTeamColor(%d) = %v, want %v", tt.teamID, result, tt.expected)
			}
		})
	}
}

// TestTeamColors verifies the team colors array matches individual color constants
func TestTeamColors(t *testing.T) {
	expectedColors := []color.Color{ColorTeamRed, ColorTeamGreen, ColorTeamBlue, ColorTeamYellow}

	if len(TeamColors) != len(expectedColors) {
		t.Errorf("TeamColors length = %d, want %d", len(TeamColors), len(expectedColors))
		return
	}

	for i, expected := range expectedColors {
		if TeamColors[i] != expected {
			t.Errorf("TeamColors[%d] = %v, want %v", i, TeamColors[i], expected)
		}
	}
}

// TestColorConsistency verifies that team colors used in GetTeamColor match TeamColors array
func TestColorConsistency(t *testing.T) {
	for i, expectedColor := range TeamColors {
		actualColor := GetTeamColor(i)
		if actualColor != expectedColor {
			t.Errorf("Color inconsistency: GetTeamColor(%d) = %v, but TeamColors[%d] = %v",
				i, actualColor, i, expectedColor)
		}
	}
}

// BenchmarkGetTeamColor benchmarks team color lookup performance
func BenchmarkGetTeamColor(b *testing.B) {
	teamIDs := []int{0, 1, 2, 3, 4, -1, 999} // Mix of valid and invalid IDs

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, teamID := range teamIDs {
			_ = GetTeamColor(teamID)
		}
	}
}

// TestDesignSystemCompliance verifies the design system follows best practices
func TestDesignSystemCompliance(t *testing.T) {
	t.Run("8-point_grid_compliance", func(t *testing.T) {
		margins := []int{MarginXSmall, MarginSmall, MarginMedium, MarginLarge, MarginXLarge}
		for _, margin := range margins {
			if margin%4 != 0 {
				t.Errorf("Margin %d does not follow 4-point base grid", margin)
			}
		}
	})

	t.Run("accessibility_font_sizes", func(t *testing.T) {
		// Minimum font size for accessibility should be at least 12px
		if FontSizeSmall < 12 {
			t.Errorf("FontSizeSmall (%d) is below accessibility minimum of 12px", FontSizeSmall)
		}
	})

	t.Run("contrast_considerations", func(t *testing.T) {
		// Ensure text colors have enough contrast against typical dark backgrounds
		// This is a basic check - real accessibility testing would need more sophisticated contrast calculations
		primary := ColorTextPrimary
		if primary.R < 200 && primary.G < 200 && primary.B < 200 {
			t.Errorf("ColorTextPrimary may not have sufficient contrast against dark backgrounds")
		}
	})
}
