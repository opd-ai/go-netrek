// Package engo provides design system constants for consistent UI styling.
// This file centralizes all spacing, sizing, color, and typography constants
// to maintain visual consistency and improve maintainability.
package engo

import "image/color"

// Spacing constants for consistent margins and padding throughout the UI.
// These follow an 8-point grid system for harmonious spacing.
const (
	// MarginXSmall provides minimal spacing for tight layouts
	MarginXSmall = 4

	// MarginSmall provides standard small spacing between related elements
	MarginSmall = 8

	// MarginMedium provides comfortable spacing between distinct UI sections
	MarginMedium = 16

	// MarginLarge provides generous spacing for major layout separation
	MarginLarge = 24

	// MarginXLarge provides maximum spacing for primary layout divisions
	MarginXLarge = 32
)

// Typography constants for consistent text sizing across the UI.
// Font sizes are chosen to maintain readability at different screen sizes.
const (
	// FontSizeSmall for secondary text and footnotes
	FontSizeSmall = 12

	// FontSizeMedium for body text and standard UI elements
	FontSizeMedium = 16

	// FontSizeLarge for headings and emphasized content
	FontSizeLarge = 20

	// FontSizeXLarge for major headings and titles
	FontSizeXLarge = 24
)

// Layout dimension constants for consistent component sizing.
// These provide standard sizes for common UI elements.
const (
	// ChatWindowHeight defines the standard height for chat window
	ChatWindowHeight = 150

	// ChatWindowWidth defines the standard width for chat window
	ChatWindowWidth = 400

	// ConnectionStatusWidth defines the standard width for connection status text
	ConnectionStatusWidth = 150

	// MinimapDefaultSize defines the default size for the minimap
	MinimapDefaultSize = 200

	// MaxChatLines defines the maximum number of chat lines to display
	MaxChatLines = 10

	// LineHeight defines the standard line height for text elements
	LineHeight = 20

	// ChatLineHeight defines the line height specifically for chat messages
	ChatLineHeight = 15

	// TeamStatusLineHeight defines the line height for team status display
	TeamStatusLineHeight = 20
)

// Alpha transparency constants for consistent opacity levels.
// These provide standard transparency levels for overlays and backgrounds.
const (
	// AlphaTransparent for completely transparent elements
	AlphaTransparent = 0

	// AlphaSemiTransparent for subtle overlay effects
	AlphaSemiTransparent = 64

	// AlphaBackground for standard background overlays (50% opacity)
	AlphaBackground = 128

	// AlphaSubtle for light overlay effects (75% opacity)
	AlphaSubtle = 192

	// AlphaOpaque for fully opaque elements
	AlphaOpaque = 255
)

// Color constants for consistent theming throughout the UI.
// These colors maintain visual hierarchy and team identification.
var (
	// Primary text colors

	// ColorTextPrimary is the primary text color for main content
	ColorTextPrimary = color.RGBA{R: 255, G: 255, B: 255, A: AlphaOpaque} // White

	// ColorTextSecondary is the secondary text color for less important content
	ColorTextSecondary = color.RGBA{R: 192, G: 192, B: 192, A: AlphaOpaque} // Light Gray

	// ColorTextDisabled is used for disabled or inactive text
	ColorTextDisabled = color.RGBA{R: 128, G: 128, B: 128, A: AlphaOpaque} // Gray

	// Background colors

	// ColorBackgroundChat provides semi-transparent background for chat window
	ColorBackgroundChat = color.RGBA{R: 0, G: 0, B: 0, A: AlphaBackground} // Semi-transparent Black

	// ColorBackgroundPanel provides background for UI panels
	ColorBackgroundPanel = color.RGBA{R: 32, G: 32, B: 32, A: AlphaSubtle} // Dark Gray

	// Team identification colors

	// ColorTeamRed represents team 1 (red team)
	ColorTeamRed = color.RGBA{R: 255, G: 0, B: 0, A: AlphaOpaque} // Red

	// ColorTeamGreen represents team 2 (green team)
	ColorTeamGreen = color.RGBA{R: 0, G: 255, B: 0, A: AlphaOpaque} // Green

	// ColorTeamBlue represents team 3 (blue team)
	ColorTeamBlue = color.RGBA{R: 0, G: 0, B: 255, A: AlphaOpaque} // Blue

	// ColorTeamYellow represents team 4 (yellow team)
	ColorTeamYellow = color.RGBA{R: 255, G: 255, B: 0, A: AlphaOpaque} // Yellow

	// ColorNeutral represents neutral entities or unaffiliated content
	ColorNeutral = color.RGBA{R: 128, G: 128, B: 128, A: AlphaOpaque} // Gray

	// Entity colors

	// ColorEntityDefault is the default color for game entities
	ColorEntityDefault = color.RGBA{R: 255, G: 255, B: 255, A: AlphaOpaque} // White

	// ColorEntityHighlight is used to highlight selected or important entities
	ColorEntityHighlight = color.RGBA{R: 255, G: 255, B: 0, A: AlphaOpaque} // Yellow
)

// GetTeamColor returns the appropriate color for a given team ID.
// This centralizes team color logic and ensures consistency across the UI.
//
// Parameters:
//   - teamID: The team identifier (0-based)
//
// Returns:
//   - color.Color: The color associated with the team, or ColorNeutral if invalid
func GetTeamColor(teamID int) color.Color {
	switch teamID {
	case 0:
		return ColorTeamRed
	case 1:
		return ColorTeamGreen
	case 2:
		return ColorTeamBlue
	case 3:
		return ColorTeamYellow
	default:
		return ColorNeutral
	}
}

// TeamColors provides an array of team colors for easy iteration.
// This is useful when rendering multiple teams or creating color legends.
var TeamColors = []color.Color{
	ColorTeamRed,
	ColorTeamGreen,
	ColorTeamBlue,
	ColorTeamYellow,
}
