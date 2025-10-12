package engo

import (
	"image/color"

	"github.com/EngoEngine/engo/common"
)

// Typography styles for different types of content
// Following the typography hierarchy defined in the design system

// TextStyle defines a complete text styling configuration
type TextStyle struct {
	Font   *common.Font
	Color  color.Color
	Size   float64
	Weight string // "normal", "bold", "light"
}

// Typography system provides structured text styling
type Typography struct {
	assetManager *AssetManager

	// Primary hierarchy styles
	heading1 *TextStyle // Large headings, critical status
	heading2 *TextStyle // Section headings, important info
	heading3 *TextStyle // Subsection headings
	body     *TextStyle // Default body text, descriptions
	caption  *TextStyle // Small text, secondary info

	// Semantic styles for specific content
	statusCritical *TextStyle // Critical warnings, errors
	statusWarning  *TextStyle // Warnings, cautions
	statusSuccess  *TextStyle // Success messages, good status
	statusInfo     *TextStyle // Informational messages

	// Interactive styles
	interactive      *TextStyle // Clickable text, links
	interactiveHover *TextStyle // Hover state

	// Chat specific styles
	chatMessage   *TextStyle // Regular chat messages
	chatTimestamp *TextStyle // Message timestamps
	chatNickname  *TextStyle // Player names

	// Game specific styles
	teamFriendly     *TextStyle // Friendly team indicators
	teamEnemy        *TextStyle // Enemy team indicators
	teamNeutral      *TextStyle // Neutral elements
	shipStatus       *TextStyle // Ship status information
	connectionStatus *TextStyle // Connection status
}

// NewTypography creates a new typography system with the asset manager
func NewTypography(assetManager *AssetManager) *Typography {
	typ := &Typography{
		assetManager: assetManager,
	}

	// Initialize all styles with default fonts and colors
	typ.initializeStyles()

	return typ
}

// initializeStyles sets up all text styles using the design system constants
func (t *Typography) initializeStyles() {
	// Get fonts from asset manager (fallback to nil if not available)
	var fontSmall, fontMedium, fontLarge, fontXLarge *common.Font
	if t.assetManager != nil {
		fontSmall = t.assetManager.GetSmallFont()
		fontMedium = t.assetManager.GetMediumFont()
		fontLarge = t.assetManager.GetLargeFont()
		fontXLarge = t.assetManager.GetXLargeFont()
	}

	// Primary hierarchy styles
	t.heading1 = &TextStyle{
		Font:   fontXLarge,
		Color:  ColorTextPrimary,
		Size:   FontSizeXLarge,
		Weight: "bold",
	}

	t.heading2 = &TextStyle{
		Font:   fontLarge,
		Color:  ColorTextPrimary,
		Size:   FontSizeLarge,
		Weight: "bold",
	}

	t.heading3 = &TextStyle{
		Font:   fontMedium,
		Color:  ColorTextPrimary,
		Size:   FontSizeMedium,
		Weight: "bold",
	}

	t.body = &TextStyle{
		Font:   fontMedium,
		Color:  ColorTextPrimary,
		Size:   FontSizeMedium,
		Weight: "normal",
	}

	t.caption = &TextStyle{
		Font:   fontSmall,
		Color:  ColorTextSecondary,
		Size:   FontSizeSmall,
		Weight: "normal",
	}

	// Semantic styles
	t.statusCritical = &TextStyle{
		Font:   fontMedium,
		Color:  ColorTeamRed,
		Size:   FontSizeMedium,
		Weight: "bold",
	}

	t.statusWarning = &TextStyle{
		Font:   fontMedium,
		Color:  ColorTeamYellow,
		Size:   FontSizeMedium,
		Weight: "normal",
	}

	t.statusSuccess = &TextStyle{
		Font:   fontMedium,
		Color:  ColorTeamGreen,
		Size:   FontSizeMedium,
		Weight: "normal",
	}

	t.statusInfo = &TextStyle{
		Font:   fontMedium,
		Color:  ColorTeamBlue,
		Size:   FontSizeMedium,
		Weight: "normal",
	}

	// Interactive styles
	t.interactive = &TextStyle{
		Font:   fontMedium,
		Color:  ColorTeamBlue,
		Size:   FontSizeMedium,
		Weight: "normal",
	}

	t.interactiveHover = &TextStyle{
		Font:   fontMedium,
		Color:  ColorEntityHighlight,
		Size:   FontSizeMedium,
		Weight: "bold",
	}

	// Chat specific styles
	t.chatMessage = &TextStyle{
		Font:   fontSmall,
		Color:  ColorTextPrimary,
		Size:   FontSizeSmall,
		Weight: "normal",
	}

	t.chatTimestamp = &TextStyle{
		Font:   fontSmall,
		Color:  ColorTextSecondary,
		Size:   FontSizeSmall,
		Weight: "normal",
	}

	t.chatNickname = &TextStyle{
		Font:   fontSmall,
		Color:  ColorEntityHighlight,
		Size:   FontSizeSmall,
		Weight: "bold",
	}

	// Game specific styles
	t.teamFriendly = &TextStyle{
		Font:   fontMedium,
		Color:  ColorTeamGreen,
		Size:   FontSizeMedium,
		Weight: "normal",
	}

	t.teamEnemy = &TextStyle{
		Font:   fontMedium,
		Color:  ColorTeamRed,
		Size:   FontSizeMedium,
		Weight: "normal",
	}

	t.teamNeutral = &TextStyle{
		Font:   fontMedium,
		Color:  ColorNeutral,
		Size:   FontSizeMedium,
		Weight: "normal",
	}

	t.shipStatus = &TextStyle{
		Font:   fontMedium,
		Color:  ColorTextPrimary,
		Size:   FontSizeMedium,
		Weight: "normal",
	}

	t.connectionStatus = &TextStyle{
		Font:   fontSmall,
		Color:  ColorTextSecondary,
		Size:   FontSizeSmall,
		Weight: "normal",
	}
}

// Style getters for different content types

// GetHeading1 returns the style for large headings and critical status
func (t *Typography) GetHeading1() *TextStyle {
	return t.heading1
}

// GetHeading2 returns the style for section headings and important information
func (t *Typography) GetHeading2() *TextStyle {
	return t.heading2
}

// GetHeading3 returns the style for subsection headings
func (t *Typography) GetHeading3() *TextStyle {
	return t.heading3
}

// GetBody returns the default body text style
func (t *Typography) GetBody() *TextStyle {
	return t.body
}

// GetCaption returns the style for small text and secondary information
func (t *Typography) GetCaption() *TextStyle {
	return t.caption
}

// Semantic style getters

// GetStatusCritical returns the style for critical warnings and errors
func (t *Typography) GetStatusCritical() *TextStyle {
	return t.statusCritical
}

// GetStatusWarning returns the style for warnings and cautions
func (t *Typography) GetStatusWarning() *TextStyle {
	return t.statusWarning
}

// GetStatusSuccess returns the style for success messages and good status
func (t *Typography) GetStatusSuccess() *TextStyle {
	return t.statusSuccess
}

// GetStatusInfo returns the style for informational messages
func (t *Typography) GetStatusInfo() *TextStyle {
	return t.statusInfo
}

// Interactive style getters

// GetInteractive returns the style for clickable text and links
func (t *Typography) GetInteractive() *TextStyle {
	return t.interactive
}

// GetInteractiveHover returns the style for hover state
func (t *Typography) GetInteractiveHover() *TextStyle {
	return t.interactiveHover
}

// Chat style getters

// GetChatMessage returns the style for regular chat messages
func (t *Typography) GetChatMessage() *TextStyle {
	return t.chatMessage
}

// GetChatTimestamp returns the style for message timestamps
func (t *Typography) GetChatTimestamp() *TextStyle {
	return t.chatTimestamp
}

// GetChatNickname returns the style for player names
func (t *Typography) GetChatNickname() *TextStyle {
	return t.chatNickname
}

// Game-specific style getters

// GetTeamFriendly returns the style for friendly team indicators
func (t *Typography) GetTeamFriendly() *TextStyle {
	return t.teamFriendly
}

// GetTeamEnemy returns the style for enemy team indicators
func (t *Typography) GetTeamEnemy() *TextStyle {
	return t.teamEnemy
}

// GetTeamNeutral returns the style for neutral elements
func (t *Typography) GetTeamNeutral() *TextStyle {
	return t.teamNeutral
}

// GetShipStatus returns the style for ship status information
func (t *Typography) GetShipStatus() *TextStyle {
	return t.shipStatus
}

// GetConnectionStatus returns the style for connection status
func (t *Typography) GetConnectionStatus() *TextStyle {
	return t.connectionStatus
}

// Utility methods

// ApplyStyle applies a text style to set font and return color for rendering
func (t *Typography) ApplyStyle(style *TextStyle, assetManager *AssetManager) color.Color {
	// Set font if available and asset manager exists
	if style.Font != nil && assetManager != nil {
		// Font is already set in the style, ready for use
	}

	return style.Color
}

// GetStyleFont returns the font for a given style, with fallback
func (t *Typography) GetStyleFont(style *TextStyle) *common.Font {
	if style.Font != nil {
		return style.Font
	}

	// Fallback to medium font if available
	if t.assetManager != nil {
		return t.assetManager.GetMediumFont()
	}

	return nil
}

// UpdateFonts refreshes all styles when fonts are reloaded
func (t *Typography) UpdateFonts() {
	t.initializeStyles()
}
