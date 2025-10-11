// Package engo provides responsive layout management for consistent UI positioning
// across different screen sizes and aspect ratios.
package engo

import (
	"github.com/EngoEngine/engo"
)

// Layout percentage constants for responsive positioning.
// These define how UI elements scale relative to screen dimensions.
const (
	// ChatWindowHeightPercent defines chat window height as percentage of screen height
	ChatWindowHeightPercent = 0.25 // 25% of screen height

	// MinimapSizePercent defines minimap size as percentage of minimum screen dimension
	MinimapSizePercent = 0.2 // 20% of minimum screen dimension

	// MarginPercent defines standard margin as percentage of screen edges
	MarginPercent = 0.01 // 1% margin from screen edges

	// StatusPanelWidthPercent defines status panel width as percentage of screen width
	StatusPanelWidthPercent = 0.15 // 15% of screen width

	// MinimumChatHeight defines minimum chat window height in pixels
	MinimumChatHeight = 100

	// MinimumMinimapSize defines minimum minimap size in pixels
	MinimumMinimapSize = 150

	// MaximumMinimapSize defines maximum minimap size in pixels
	MaximumMinimapSize = 300
)

// Position represents a 2D coordinate position
type Position struct {
	X, Y float32
}

// Dimensions represents width and height measurements
type Dimensions struct {
	Width, Height float32
}

// LayoutManager provides responsive layout calculations for UI elements.
// It tracks viewport dimensions and calculates percentage-based positions
// that adapt to different screen sizes while maintaining visual hierarchy.
type LayoutManager struct {
	// viewport tracks current screen dimensions
	viewport Dimensions

	// cachedPositions stores calculated positions to avoid repeated calculations
	cachedPositions  map[string]Position
	cachedDimensions map[string]Dimensions

	// lastViewportUpdate tracks when viewport was last updated for cache invalidation
	lastViewportUpdate int64

	// viewportSource allows dependency injection for testing
	viewportSource ViewportSource
}

// ViewportSource provides viewport dimensions - allows testing without Engo
type ViewportSource interface {
	GetWidth() int
	GetHeight() int
}

// EngoViewportSource implements ViewportSource using Engo's game dimensions
type EngoViewportSource struct{}

// GetWidth returns the current game width from Engo
func (evs *EngoViewportSource) GetWidth() int {
	return int(engo.GameWidth())
}

// GetHeight returns the current game height from Engo
func (evs *EngoViewportSource) GetHeight() int {
	return int(engo.GameHeight())
}

// NewLayoutManager creates a new responsive layout manager.
// It initializes with current screen dimensions and prepares caching systems.
//
// Returns:
//   - *LayoutManager: A new layout manager ready for use
func NewLayoutManager() *LayoutManager {
	lm := &LayoutManager{
		cachedPositions:  make(map[string]Position),
		cachedDimensions: make(map[string]Dimensions),
		viewportSource:   &EngoViewportSource{}, // Use Engo by default
	}

	// Initialize with current viewport
	lm.UpdateViewport()

	return lm
}

// NewLayoutManagerWithViewport creates a layout manager with a custom viewport source.
// This is useful for testing or when using a different rendering system.
//
// Parameters:
//   - source: ViewportSource implementation to provide dimensions
//
// Returns:
//   - *LayoutManager: A new layout manager with custom viewport source
func NewLayoutManagerWithViewport(source ViewportSource) *LayoutManager {
	lm := &LayoutManager{
		cachedPositions:  make(map[string]Position),
		cachedDimensions: make(map[string]Dimensions),
		viewportSource:   source,
	}

	// Initialize with current viewport
	lm.UpdateViewport()

	return lm
}

// UpdateViewport refreshes the viewport dimensions from the current game window.
// This should be called whenever the window is resized or when layout calculations
// need to reflect current screen size.
//
// The method invalidates cached positions when viewport changes significantly.
func (lm *LayoutManager) UpdateViewport() {
	newViewport := Dimensions{
		Width:  float32(lm.viewportSource.GetWidth()),
		Height: float32(lm.viewportSource.GetHeight()),
	}

	// Check if viewport changed significantly (more than 1 pixel difference)
	if abs(newViewport.Width-lm.viewport.Width) > 1 || abs(newViewport.Height-lm.viewport.Height) > 1 {
		// Clear caches when viewport changes
		lm.cachedPositions = make(map[string]Position)
		lm.cachedDimensions = make(map[string]Dimensions)
		lm.lastViewportUpdate++
	}

	lm.viewport = newViewport
}

// GetViewport returns the current viewport dimensions.
//
// Returns:
//   - Dimensions: Current screen width and height
func (lm *LayoutManager) GetViewport() Dimensions {
	return lm.viewport
}

// GetStandardMargin calculates the standard margin size based on screen dimensions.
// Uses the smaller of percentage-based margin or design system minimum.
//
// Returns:
//   - float32: Margin size in pixels
func (lm *LayoutManager) GetStandardMargin() float32 {
	percentageMargin := lm.viewport.Width * MarginPercent

	// Use the larger of percentage margin or design system minimum
	if percentageMargin > float32(MarginSmall) {
		return percentageMargin
	}
	return float32(MarginSmall)
}

// GetShipStatusPosition calculates the position for ship status display.
// Positions at top-left with standard margin.
//
// Returns:
//   - Position: X, Y coordinates for ship status panel
func (lm *LayoutManager) GetShipStatusPosition() Position {
	const cacheKey = "ship_status"

	if pos, exists := lm.cachedPositions[cacheKey]; exists {
		return pos
	}

	margin := lm.GetStandardMargin()
	pos := Position{
		X: margin,
		Y: margin,
	}

	lm.cachedPositions[cacheKey] = pos
	return pos
}

// GetChatPosition calculates position and dimensions for the chat window.
// Positions at bottom-left with responsive height based on screen size.
//
// Returns:
//   - Position: X, Y coordinates for chat window
func (lm *LayoutManager) GetChatPosition() Position {
	const cacheKey = "chat_window"

	if pos, exists := lm.cachedPositions[cacheKey]; exists {
		return pos
	}

	margin := lm.GetStandardMargin()
	chatDimensions := lm.GetChatDimensions()

	pos := Position{
		X: margin,
		Y: lm.viewport.Height - chatDimensions.Height - margin,
	}

	lm.cachedPositions[cacheKey] = pos
	return pos
}

// GetChatDimensions calculates responsive dimensions for the chat window.
// Height scales with screen size but respects minimum and maximum bounds.
//
// Returns:
//   - Dimensions: Width and height for chat window
func (lm *LayoutManager) GetChatDimensions() Dimensions {
	const cacheKey = "chat_dimensions"

	if dims, exists := lm.cachedDimensions[cacheKey]; exists {
		return dims
	}

	// Calculate responsive height
	responsiveHeight := lm.viewport.Height * ChatWindowHeightPercent

	// Apply minimum height constraint
	chatHeight := responsiveHeight
	if chatHeight < MinimumChatHeight {
		chatHeight = MinimumChatHeight
	}

	// Use design system width or scale down for very narrow screens
	chatWidth := float32(ChatWindowWidth)
	if lm.viewport.Width < ChatWindowWidth+2*lm.GetStandardMargin() {
		// Scale down width for narrow screens, keeping some margin
		chatWidth = lm.viewport.Width - 2*lm.GetStandardMargin()
		if chatWidth < 200 { // Minimum usable chat width
			chatWidth = 200
		}
	}

	dims := Dimensions{
		Width:  chatWidth,
		Height: chatHeight,
	}

	lm.cachedDimensions[cacheKey] = dims
	return dims
}

// GetTeamStatusPosition calculates the position for team status display.
// Positions below ship status with appropriate spacing.
//
// Returns:
//   - Position: X, Y coordinates for team status
func (lm *LayoutManager) GetTeamStatusPosition() Position {
	const cacheKey = "team_status"

	if pos, exists := lm.cachedPositions[cacheKey]; exists {
		return pos
	}

	margin := lm.GetStandardMargin()

	// Position below ship status panel with spacing
	// Estimate ship status height (approximately 5 lines of text)
	estimatedShipStatusHeight := float32(FontSizeMedium * 5)

	pos := Position{
		X: margin,
		Y: margin + estimatedShipStatusHeight + float32(MarginMedium),
	}

	lm.cachedPositions[cacheKey] = pos
	return pos
}

// GetTeamStatusDimensions calculates the available dimensions for team status display.
// Returns the maximum width and height available for team status rendering.
//
// Returns:
//   - Dimensions: Available width and height for team status container
func (lm *LayoutManager) GetTeamStatusDimensions() Dimensions {
	const cacheKey = "team_status_dims"

	if dims, exists := lm.cachedDimensions[cacheKey]; exists {
		return dims
	}

	margin := lm.GetStandardMargin()
	teamPos := lm.GetTeamStatusPosition()
	
	// Calculate maximum width (left side of screen minus margins)
	maxWidth := lm.viewport.Width/2 - margin*2
	
	// Calculate maximum height (from team status position to chat window)
	chatPos := lm.GetChatPosition()
	maxHeight := chatPos.Y - teamPos.Y - margin
	
	// Ensure minimum dimensions
	if maxWidth < 200 {
		maxWidth = 200
	}
	if maxHeight < TeamStatusLineHeight {
		maxHeight = TeamStatusLineHeight
	}

	dims := Dimensions{
		Width:  maxWidth,
		Height: maxHeight,
	}

	lm.cachedDimensions[cacheKey] = dims
	return dims
}

// GetConnectionStatusPosition calculates position for connection status display.
// Positions at top-right with appropriate spacing from screen edge.
//
// Returns:
//   - Position: X, Y coordinates for connection status
func (lm *LayoutManager) GetConnectionStatusPosition() Position {
	const cacheKey = "connection_status"

	if pos, exists := lm.cachedPositions[cacheKey]; exists {
		return pos
	}

	margin := lm.GetStandardMargin()

	// Estimate text width for "Status: Connected" (rough approximation)
	// This will be improved when font measurement is implemented
	estimatedTextWidth := float32(len("Status: Connected") * 8) // 8px per character approximation

	pos := Position{
		X: lm.viewport.Width - estimatedTextWidth - margin,
		Y: margin,
	}

	lm.cachedPositions[cacheKey] = pos
	return pos
}

// GetMinimapPosition calculates position for minimap display.
// Positions at top-right below connection status.
//
// Returns:
//   - Position: X, Y coordinates for minimap
func (lm *LayoutManager) GetMinimapPosition() Position {
	const cacheKey = "minimap"

	if pos, exists := lm.cachedPositions[cacheKey]; exists {
		return pos
	}

	margin := lm.GetStandardMargin()
	minimapSize := lm.GetMinimapSize()

	// Position at top-right, accounting for connection status above
	connectionStatusHeight := float32(FontSizeMedium + MarginSmall)

	pos := Position{
		X: lm.viewport.Width - minimapSize - margin,
		Y: margin + connectionStatusHeight + float32(MarginSmall),
	}

	lm.cachedPositions[cacheKey] = pos
	return pos
}

// GetMinimapSize calculates responsive size for the minimap.
// Size scales with screen dimensions but respects minimum and maximum bounds.
//
// Returns:
//   - float32: Minimap size (width and height are equal)
func (lm *LayoutManager) GetMinimapSize() float32 {
	const cacheKey = "minimap_size"

	if dims, exists := lm.cachedDimensions[cacheKey]; exists {
		return dims.Width // Square minimap, so width == height
	}

	// Calculate size as percentage of smaller screen dimension
	minDimension := lm.viewport.Width
	if lm.viewport.Height < minDimension {
		minDimension = lm.viewport.Height
	}

	responsiveSize := minDimension * MinimapSizePercent

	// Apply size constraints
	minimapSize := responsiveSize
	if minimapSize < MinimumMinimapSize {
		minimapSize = MinimumMinimapSize
	} else if minimapSize > MaximumMinimapSize {
		minimapSize = MaximumMinimapSize
	}

	// Cache as dimensions for consistency
	dims := Dimensions{Width: minimapSize, Height: minimapSize}
	lm.cachedDimensions[cacheKey] = dims

	return minimapSize
}

// IsElementVisible checks if an element at given position and dimensions
// is visible within the current viewport.
//
// Parameters:
//   - pos: Element position
//   - dims: Element dimensions
//
// Returns:
//   - bool: True if element is at least partially visible
func (lm *LayoutManager) IsElementVisible(pos Position, dims Dimensions) bool {
	// Check if element is completely outside viewport bounds
	if pos.X+dims.Width <= 0 || pos.X >= lm.viewport.Width {
		return false
	}
	if pos.Y+dims.Height <= 0 || pos.Y >= lm.viewport.Height {
		return false
	}
	return true
}

// ClearCache forces a cache clear for all calculated positions and dimensions.
// Useful when major layout changes occur or for testing purposes.
func (lm *LayoutManager) ClearCache() {
	lm.cachedPositions = make(map[string]Position)
	lm.cachedDimensions = make(map[string]Dimensions)
	lm.lastViewportUpdate++
}

// abs returns the absolute value of a float32
func abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}
