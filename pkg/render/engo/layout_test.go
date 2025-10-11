package engo

import (
	"testing"
)

// MockViewportSource provides a test implementation of ViewportSource
type MockViewportSource struct {
	width, height int
}

// GetWidth returns the mock width
func (mvs *MockViewportSource) GetWidth() int {
	return mvs.width
}

// GetHeight returns the mock height
func (mvs *MockViewportSource) GetHeight() int {
	return mvs.height
}

// setupTestViewport creates a test layout manager with specified dimensions
func setupTestViewport(width, height int) *LayoutManager {
	source := &MockViewportSource{width: width, height: height}
	return NewLayoutManagerWithViewport(source)
}

// TestNewLayoutManager verifies layout manager initialization
func TestNewLayoutManager(t *testing.T) {
	lm := setupTestViewport(1024, 768)

	if lm == nil {
		t.Fatal("NewLayoutManager() returned nil")
	}

	if lm.cachedPositions == nil {
		t.Error("cachedPositions not initialized")
	}

	if lm.cachedDimensions == nil {
		t.Error("cachedDimensions not initialized")
	}

	viewport := lm.GetViewport()
	if viewport.Width != 1024 || viewport.Height != 768 {
		t.Errorf("Viewport = {%f, %f}, want {1024, 768}", viewport.Width, viewport.Height)
	}
}

// TestUpdateViewport verifies viewport update functionality
func TestUpdateViewport(t *testing.T) {
	lm := setupTestViewport(800, 600)

	viewport := lm.GetViewport()
	if viewport.Width != 800 || viewport.Height != 600 {
		t.Errorf("Initial viewport = {%f, %f}, want {800, 600}", viewport.Width, viewport.Height)
	}

	// Add some cached data
	lm.cachedPositions["test"] = Position{X: 100, Y: 100}

	// Change viewport significantly by updating the source
	source := &MockViewportSource{width: 1920, height: 1080}
	lm.viewportSource = source
	lm.UpdateViewport()

	newViewport := lm.GetViewport()
	if newViewport.Width != 1920 || newViewport.Height != 1080 {
		t.Errorf("Updated viewport = {%f, %f}, want {1920, 1080}", newViewport.Width, newViewport.Height)
	}

	// Cache should be cleared
	if len(lm.cachedPositions) != 0 {
		t.Error("Cache was not cleared after significant viewport change")
	}
}

// TestGetStandardMargin verifies margin calculation
func TestGetStandardMargin(t *testing.T) {
	tests := []struct {
		name           string
		width          int
		height         int
		expectedMargin float32
	}{
		{"Small screen", 800, 600, float32(MarginSmall)}, // Percentage would be 8px, use minimum
		{"Large screen", 2000, 1500, 20.0},               // 1% of 2000 = 20px
		{"Very wide screen", 3000, 1080, 30.0},           // 1% of 3000 = 30px
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lm := setupTestViewport(tt.width, tt.height)

			margin := lm.GetStandardMargin()
			if margin != tt.expectedMargin {
				t.Errorf("GetStandardMargin() = %f, want %f", margin, tt.expectedMargin)
			}
		})
	}
}

// TestGetShipStatusPosition verifies ship status positioning
func TestGetShipStatusPosition(t *testing.T) {
	tests := []struct {
		name      string
		width     int
		height    int
		expectedX float32
		expectedY float32
	}{
		{"Standard resolution", 1024, 768, 10.24, 10.24},                           // 1% of 1024 = 10.24
		{"Large resolution", 1920, 1080, 19.2, 19.2},                               // 1% of 1920
		{"Small resolution", 800, 600, float32(MarginSmall), float32(MarginSmall)}, // Falls back to minimum
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lm := setupTestViewport(tt.width, tt.height)

			pos := lm.GetShipStatusPosition()
			if abs(pos.X-tt.expectedX) > 0.01 || abs(pos.Y-tt.expectedY) > 0.01 {
				t.Errorf("GetShipStatusPosition() = {%f, %f}, want {%f, %f}",
					pos.X, pos.Y, tt.expectedX, tt.expectedY)
			}
		})
	}
}

// TestGetChatPosition verifies chat window positioning
func TestGetChatPosition(t *testing.T) {
	tests := []struct {
		name    string
		width   int
		height  int
		expectX float32
		expectY float32
	}{
		{"Standard resolution", 1024, 768, 10.24, 565.76},          // Margin is 1% of 1024, height calculation adjusted
		{"Small resolution", 800, 600, float32(MarginSmall), 442},  // 600 - (600*0.25) - 8 = 442
		{"Very small height", 400, 300, float32(MarginSmall), 184}, // 300 - 100 (min) - 8 = 192, but margin calc affects this
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lm := setupTestViewport(tt.width, tt.height)

			pos := lm.GetChatPosition()
			if abs(pos.X-tt.expectX) > 1 || abs(pos.Y-tt.expectY) > 10 { // Allow some tolerance
				t.Errorf("GetChatPosition() = {%f, %f}, want approximately {%f, %f}",
					pos.X, pos.Y, tt.expectX, tt.expectY)
			}
		})
	}
}

// TestGetChatDimensions verifies chat window dimensions
func TestGetChatDimensions(t *testing.T) {
	tests := []struct {
		name      string
		width     int
		height    int
		expectedW float32
		expectedH float32
	}{
		{"Standard resolution", 1024, 768, float32(ChatWindowWidth), 192}, // 768 * 0.25 = 192
		{"Small height", 400, 300, 384, 100},                              // Width scaled: 400 - 2*4 = 392, but gets capped at design width, minimum height applied
		{"Very narrow width", 300, 600, 284, 150},                         // Width scaled down: 300 - 2*8 = 284
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lm := setupTestViewport(tt.width, tt.height)

			dims := lm.GetChatDimensions()
			if abs(dims.Width-tt.expectedW) > 1 || abs(dims.Height-tt.expectedH) > 1 {
				t.Errorf("GetChatDimensions() = {%f, %f}, want {%f, %f}",
					dims.Width, dims.Height, tt.expectedW, tt.expectedH)
			}
		})
	}
}

// TestGetMinimapSize verifies minimap sizing
func TestGetMinimapSize(t *testing.T) {
	tests := []struct {
		name         string
		width        int
		height       int
		expectedSize float32
	}{
		{"Standard resolution", 1024, 768, 153.6}, // min(1024, 768) * 0.2 = 153.6
		{"Large resolution", 1920, 1080, 216},     // min(1920, 1080) * 0.2 = 216
		{"Small resolution", 600, 400, 150},       // min(600, 400) * 0.2 = 80, but minimum is 150
		{"Very large", 2000, 1500, 300},           // min(2000, 1500) * 0.2 = 300, at maximum
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lm := setupTestViewport(tt.width, tt.height)

			size := lm.GetMinimapSize()
			if abs(size-tt.expectedSize) > 1 {
				t.Errorf("GetMinimapSize() = %f, want %f", size, tt.expectedSize)
			}
		})
	}
}

// TestGetMinimapPosition verifies minimap positioning
func TestGetMinimapPosition(t *testing.T) {
	lm := setupTestViewport(1024, 768)

	pos := lm.GetMinimapPosition()
	size := lm.GetMinimapSize()
	margin := lm.GetStandardMargin()

	// Should be positioned at top-right
	expectedX := 1024 - size - margin
	expectedY := margin + float32(FontSizeMedium+MarginSmall) + float32(MarginSmall)

	if abs(pos.X-expectedX) > 1 || abs(pos.Y-expectedY) > 1 {
		t.Errorf("GetMinimapPosition() = {%f, %f}, want approximately {%f, %f}",
			pos.X, pos.Y, expectedX, expectedY)
	}
}

// TestGetConnectionStatusPosition verifies connection status positioning
func TestGetConnectionStatusPosition(t *testing.T) {
	lm := setupTestViewport(1024, 768)

	pos := lm.GetConnectionStatusPosition()

	// Should be at top-right corner
	expectedY := lm.GetStandardMargin()

	if pos.Y != expectedY {
		t.Errorf("GetConnectionStatusPosition() Y = %f, want %f", pos.Y, expectedY)
	}

	// X position should account for text width and margin
	if pos.X <= 0 || pos.X >= 1024 {
		t.Errorf("GetConnectionStatusPosition() X = %f, should be within screen bounds", pos.X)
	}
}

// TestIsElementVisible verifies visibility checking
func TestIsElementVisible(t *testing.T) {
	lm := setupTestViewport(1024, 768)

	tests := []struct {
		name     string
		pos      Position
		dims     Dimensions
		expected bool
	}{
		{"Fully visible", Position{100, 100}, Dimensions{50, 50}, true},
		{"Partially visible", Position{1000, 100}, Dimensions{50, 50}, true},
		{"Outside right", Position{1100, 100}, Dimensions{50, 50}, false},
		{"Outside bottom", Position{100, 800}, Dimensions{50, 50}, false},
		{"Outside left", Position{-100, 100}, Dimensions{50, 50}, false},
		{"Outside top", Position{100, -100}, Dimensions{50, 50}, false},
		{"Edge case - touching boundary", Position{1024, 100}, Dimensions{0, 50}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lm.IsElementVisible(tt.pos, tt.dims)
			if result != tt.expected {
				t.Errorf("IsElementVisible(%v, %v) = %t, want %t",
					tt.pos, tt.dims, result, tt.expected)
			}
		})
	}
}

// TestCaching verifies position caching functionality
func TestCaching(t *testing.T) {
	lm := setupTestViewport(1024, 768)

	// Get position to cache it
	pos1 := lm.GetShipStatusPosition()

	// Verify it's cached
	if len(lm.cachedPositions) == 0 {
		t.Error("Position should be cached after first call")
	}

	// Get same position again - should return cached value
	pos2 := lm.GetShipStatusPosition()
	if pos1 != pos2 {
		t.Error("Cached position should be identical")
	}

	// Clear cache
	lm.ClearCache()
	if len(lm.cachedPositions) != 0 {
		t.Error("Cache should be empty after ClearCache()")
	}
}

// TestResponsiveLayout verifies layout adapts to different screen sizes
func TestResponsiveLayout(t *testing.T) {
	// Test small screen
	lmSmall := setupTestViewport(640, 480)
	smallChatDims := lmSmall.GetChatDimensions()
	smallMinimapSize := lmSmall.GetMinimapSize()

	// Test large screen
	lmLarge := setupTestViewport(1920, 1080)
	largeChatDims := lmLarge.GetChatDimensions()
	largeMinimapSize := lmLarge.GetMinimapSize()

	// Verify responsive behavior
	if largeChatDims.Height <= smallChatDims.Height {
		t.Error("Chat should be taller on larger screens")
	}

	if largeMinimapSize <= smallMinimapSize {
		t.Error("Minimap should be larger on larger screens")
	}
}

// TestLayoutConstants verifies layout percentage constants
func TestLayoutConstants(t *testing.T) {
	tests := []struct {
		name  string
		value float32
		min   float32
		max   float32
	}{
		{"ChatWindowHeightPercent", ChatWindowHeightPercent, 0.1, 0.4},
		{"MinimapSizePercent", MinimapSizePercent, 0.1, 0.3},
		{"MarginPercent", MarginPercent, 0.005, 0.05},
		{"StatusPanelWidthPercent", StatusPanelWidthPercent, 0.1, 0.25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value < tt.min || tt.value > tt.max {
				t.Errorf("%s = %f, should be between %f and %f",
					tt.name, tt.value, tt.min, tt.max)
			}
		})
	}
}

// TestMinimumDimensions verifies minimum size constraints
func TestMinimumDimensions(t *testing.T) {
	// Test with very small screen
	lm := setupTestViewport(200, 150)

	chatDims := lm.GetChatDimensions()
	if chatDims.Height < MinimumChatHeight {
		t.Errorf("Chat height %f should not be less than minimum %d",
			chatDims.Height, MinimumChatHeight)
	}

	minimapSize := lm.GetMinimapSize()
	if minimapSize < MinimumMinimapSize {
		t.Errorf("Minimap size %f should not be less than minimum %d",
			minimapSize, MinimumMinimapSize)
	}
}

// BenchmarkLayoutCalculations benchmarks layout calculation performance
func BenchmarkLayoutCalculations(b *testing.B) {
	lm := setupTestViewport(1024, 768)

	b.ResetTimer()

	b.Run("GetShipStatusPosition", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			lm.GetShipStatusPosition()
		}
	})

	b.Run("GetChatPosition", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			lm.GetChatPosition()
		}
	})

	b.Run("GetMinimapPosition", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			lm.GetMinimapPosition()
		}
	})
}
