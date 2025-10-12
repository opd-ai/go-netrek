// pkg/render/engo/hud.go
package engo

import (
	"fmt"
	"image/color"
	"time"

	"github.com/EngoEngine/ecs"
	"github.com/EngoEngine/engo"
	"github.com/EngoEngine/engo/common"

	"github.com/opd-ai/go-netrek/pkg/engine"
)

// HUDSystem manages the heads-up display
type HUDSystem struct {
	// HUD entities
	hudEntities []*ecs.BasicEntity

	// Status display
	connectionStatus string

	// Chat system
	chatMessages []ChatMessage
	maxChatLines int

	// Ship status
	currentShip *engine.ShipState

	// Minimap
	minimapEnabled bool
	minimapSize    float32

	// Team status
	teamStates map[int]engine.TeamState

	// Font for text rendering
	font *common.Font

	// Colors
	hudColor      color.Color
	friendlyColor color.Color
	enemyColor    color.Color
	neutralColor  color.Color

	// Layout manager for responsive positioning
	layoutManager *LayoutManager

	// Typography system for consistent text styling
	typography *Typography

	// Asset manager for accessing fonts and textures
	assetManager *AssetManager

	// Render system for adding HUD entities
	renderSystem *common.RenderSystem
}

// ChatMessage represents a chat message in the HUD
type ChatMessage struct {
	Sender    string
	Message   string
	Timestamp time.Time
	Color     color.Color
}

// NewHUDSystem creates a new HUD system
func NewHUDSystem(assetManager *AssetManager, renderSystem *common.RenderSystem) *HUDSystem {
	hud := &HUDSystem{
		connectionStatus: "Connected",
		maxChatLines:     MaxChatLines,
		minimapEnabled:   true,
		minimapSize:      MinimapDefaultSize,
		teamStates:       make(map[int]engine.TeamState),
		hudColor:         ColorTextPrimary,
		friendlyColor:    ColorTeamGreen,
		enemyColor:       ColorTeamRed,
		neutralColor:     ColorNeutral,
		layoutManager:    NewLayoutManager(),
		typography:       NewTypography(assetManager),
		assetManager:     assetManager,
		renderSystem:     renderSystem,
	}

	// Initialize font from asset manager
	if assetManager != nil {
		hud.font = assetManager.GetMediumFont()
	}

	return hud
}

// Add satisfies the ecs.System interface
func (hud *HUDSystem) Add(basic *ecs.BasicEntity, render *common.RenderComponent, space *common.SpaceComponent) {
	// Not used for HUD system
}

// Remove satisfies the ecs.System interface
func (hud *HUDSystem) Remove(basic ecs.BasicEntity) {
	// Not used for HUD system
}

// Update updates the HUD display
func (hud *HUDSystem) Update(dt float32) {
	// Clear previous HUD entities
	hud.clearHUDEntities()

	// Render HUD components
	hud.renderShipStatus()
	hud.renderChatWindow()
	hud.renderTeamStatus()
	hud.renderConnectionStatus()

	if hud.minimapEnabled {
		hud.renderMinimap()
	}
}

// clearHUDEntities removes previous HUD entities from the render system
func (hud *HUDSystem) clearHUDEntities() {
	// Remove all HUD entities from the render system
	if hud.renderSystem != nil {
		for _, entity := range hud.hudEntities {
			hud.renderSystem.Remove(*entity)
		}
	}

	// Clear the entities list
	hud.hudEntities = hud.hudEntities[:0]
}

// renderShipStatus renders the player's ship status panel
func (hud *HUDSystem) renderShipStatus() {
	if hud.currentShip == nil {
		return
	}

	// Create status panel
	statusText := fmt.Sprintf(
		"Hull: %d/%d\nShields: %d\nFuel: %d\nArmies: %d",
		hud.currentShip.Hull,
		100, // Max hull - this should come from ship stats
		hud.currentShip.Shields,
		hud.currentShip.Fuel,
		hud.currentShip.Armies,
	)

	// Render status text using layout manager for responsive positioning
	hud.layoutManager.UpdateViewport() // Ensure viewport is current
	pos := hud.layoutManager.GetShipStatusPosition()
	hud.renderStyledText(statusText, pos.X, pos.Y, hud.typography.GetShipStatus())
}

// renderChatWindow renders the chat message window
func (hud *HUDSystem) renderChatWindow() {
	// Use layout manager for responsive positioning
	chatPos := hud.layoutManager.GetChatPosition()
	chatDims := hud.layoutManager.GetChatDimensions()

	// Render chat background
	hud.renderRect(chatPos.X, chatPos.Y, chatDims.Width, chatDims.Height, ColorBackgroundChat)

	// Render chat messages
	y := chatPos.Y + hud.layoutManager.GetStandardMargin()
	for i := len(hud.chatMessages) - 1; i >= 0 && i >= len(hud.chatMessages)-hud.maxChatLines; i-- {
		msg := hud.chatMessages[i]
		chatLine := fmt.Sprintf("[%s]: %s", msg.Sender, msg.Message)

		// Use chat message typography style but override color with message color
		chatStyle := &TextStyle{
			Font:   hud.typography.GetChatMessage().Font,
			Color:  msg.Color,
			Size:   hud.typography.GetChatMessage().Size,
			Weight: hud.typography.GetChatMessage().Weight,
		}

		hud.renderStyledText(chatLine, chatPos.X+MarginMedium, y, chatStyle)
		y += ChatLineHeight
	}
}

// renderTeamStatus renders team scores and status with overflow handling
func (hud *HUDSystem) renderTeamStatus() {
	pos := hud.layoutManager.GetTeamStatusPosition()
	dims := hud.layoutManager.GetTeamStatusDimensions()
	startY := pos.Y

	// Calculate how many team lines can fit in available space
	maxLines := int(dims.Height / TeamStatusLineHeight)
	if maxLines < 1 {
		maxLines = 1 // Always show at least one line if possible
	}

	// Convert map to sorted slice for consistent ordering
	type teamEntry struct {
		ID    int
		State engine.TeamState
	}

	var teams []teamEntry
	for teamID, teamState := range hud.teamStates {
		teams = append(teams, teamEntry{ID: teamID, State: teamState})
	}

	// Sort teams by ID for consistent display order
	for i := 0; i < len(teams)-1; i++ {
		for j := i + 1; j < len(teams); j++ {
			if teams[i].ID > teams[j].ID {
				teams[i], teams[j] = teams[j], teams[i]
			}
		}
	}

	// Render teams that fit in available space
	renderedLines := 0
	for _, team := range teams {
		if renderedLines >= maxLines {
			break
		}

		teamText := fmt.Sprintf(
			"%s: Score %d, Ships %d, Planets %d",
			team.State.Name,
			team.State.Score,
			team.State.ShipCount,
			team.State.PlanetCount,
		)

		// Use appropriate typography style based on team relationship
		// For now, we'll use the body style with team color
		// In the future, this could be enhanced to use semantic styles
		teamStyle := &TextStyle{
			Font:   hud.typography.GetBody().Font,
			Color:  GetTeamColor(team.ID),
			Size:   hud.typography.GetBody().Size,
			Weight: "normal",
		}

		hud.renderStyledText(teamText, pos.X, startY, teamStyle)
		startY += TeamStatusLineHeight
		renderedLines++
	}

	// Show overflow indicator if there are more teams than we can display
	if len(teams) > maxLines {
		overflowText := fmt.Sprintf("... and %d more teams", len(teams)-maxLines)
		hud.renderStyledText(overflowText, pos.X, startY, hud.typography.GetCaption())
	}
}

// renderConnectionStatus renders the connection status
func (hud *HUDSystem) renderConnectionStatus() {
	statusText := "Status: " + hud.connectionStatus

	// Choose appropriate style based on connection state
	var style *TextStyle
	if hud.connectionStatus != "Connected" {
		style = hud.typography.GetStatusWarning()
	} else {
		style = hud.typography.GetConnectionStatus()
	}

	// Use dynamic text measurement for precise right-alignment
	font := hud.typography.GetStyleFont(style)
	if font != nil {
		textWidth := MeasureText(statusText, font)

		// Get right-aligned position using actual text width
		pos := hud.layoutManager.GetRightAlignedPosition(statusText, textWidth, hud.layoutManager.GetStandardMargin())

		// Render with typography style
		hud.renderStyledText(statusText, pos.X, pos.Y, style)
	} else {
		// Fallback to legacy positioning if font is not available
		pos := hud.layoutManager.GetConnectionStatusPosition()
		hud.renderStyledText(statusText, pos.X, pos.Y, style)
	}
}

// renderMinimap renders a minimap showing the game world
func (hud *HUDSystem) renderMinimap() {
	pos := hud.layoutManager.GetMinimapPosition()
	size := hud.layoutManager.GetMinimapSize()

	// Render minimap background
	hud.renderRect(pos.X, pos.Y, size, size, ColorBackgroundChat)

	// Render minimap border
	hud.renderRectOutline(pos.X, pos.Y, size, size, hud.hudColor)

	// Note: In a real implementation, you would render planets and ships on the minimap
	// based on the current game state
}

// renderText renders text at the specified position
func (hud *HUDSystem) renderText(text string, x, y float32, textColor color.Color) {
	// Calculate approximate text dimensions
	textWidth := float32(len(text) * 8) // Approximate width
	textHeight := float32(16)           // Approximate height

	// Validate bounds before rendering
	pos := Position{X: x, Y: y}
	dims := Dimensions{Width: textWidth, Height: textHeight}
	if !hud.layoutManager.IsElementVisible(pos, dims) {
		// Skip rendering if element would be outside visible bounds
		return
	}

	// Create a text entity
	basic := ecs.NewBasic()

	renderComponent := common.RenderComponent{
		Drawable: common.Text{
			Font: hud.font,
			Text: text,
		},
		Color: textColor,
	}

	spaceComponent := common.SpaceComponent{
		Position: engo.Point{X: x, Y: y},
		Width:    textWidth,
		Height:   textHeight,
	}

	// Add to HUD entities list for cleanup
	hud.hudEntities = append(hud.hudEntities, &basic)

	// Add to render system so it actually appears on screen
	if hud.renderSystem != nil {
		hud.renderSystem.Add(&basic, &renderComponent, &spaceComponent)
	}
}

// renderStyledText renders text using a typography style
func (hud *HUDSystem) renderStyledText(text string, x, y float32, style *TextStyle) {
	// Use the style's font and color
	font := hud.typography.GetStyleFont(style)
	color := style.Color

	// Calculate text dimensions using font-aware measurement
	var textWidth, textHeight float32
	if font != nil {
		textWidth = MeasureText(text, font)
		textHeight = float32(font.Size)
	} else {
		// Fallback to approximate dimensions
		textWidth = float32(len(text) * 8)
		textHeight = float32(16)
	}

	// Validate bounds before rendering
	pos := Position{X: x, Y: y}
	dims := Dimensions{Width: textWidth, Height: textHeight}
	if !hud.layoutManager.IsElementVisible(pos, dims) {
		// Skip rendering if element would be outside visible bounds
		return
	}

	// Create a text entity
	basic := ecs.NewBasic()

	renderComponent := common.RenderComponent{
		Drawable: common.Text{
			Font: font,
			Text: text,
		},
		Color: color,
	}

	spaceComponent := common.SpaceComponent{
		Position: engo.Point{X: x, Y: y},
		Width:    textWidth,
		Height:   textHeight,
	}

	// Add to HUD entities list for cleanup
	hud.hudEntities = append(hud.hudEntities, &basic)

	// Add to render system so it actually appears on screen
	if hud.renderSystem != nil {
		hud.renderSystem.Add(&basic, &renderComponent, &spaceComponent)
	}
}

// renderRect renders a filled rectangle
func (hud *HUDSystem) renderRect(x, y, width, height float32, rectColor color.Color) {
	// Validate bounds before rendering
	pos := Position{X: x, Y: y}
	dims := Dimensions{Width: width, Height: height}
	if !hud.layoutManager.IsElementVisible(pos, dims) {
		// Skip rendering if element would be outside visible bounds
		return
	}

	// Create a rectangle entity
	basic := ecs.NewBasic()

	renderComponent := common.RenderComponent{
		Drawable: common.Rectangle{
			BorderWidth: 0,
			BorderColor: color.Transparent,
		},
		Color: rectColor,
	}

	spaceComponent := common.SpaceComponent{
		Position: engo.Point{X: x, Y: y},
		Width:    width,
		Height:   height,
	}

	// Add to HUD entities list for cleanup
	hud.hudEntities = append(hud.hudEntities, &basic)

	// Add to render system so it actually appears on screen
	if hud.renderSystem != nil {
		hud.renderSystem.Add(&basic, &renderComponent, &spaceComponent)
	}
}

// renderRectOutline renders a rectangle outline
func (hud *HUDSystem) renderRectOutline(x, y, width, height float32, outlineColor color.Color) {
	// Validate bounds before rendering
	pos := Position{X: x, Y: y}
	dims := Dimensions{Width: width, Height: height}
	if !hud.layoutManager.IsElementVisible(pos, dims) {
		// Skip rendering if element would be outside visible bounds
		return
	}

	// Create a rectangle outline entity
	basic := ecs.NewBasic()

	renderComponent := common.RenderComponent{
		Drawable: common.Rectangle{
			BorderWidth: 2,
			BorderColor: outlineColor,
		},
		Color: color.Transparent,
	}

	spaceComponent := common.SpaceComponent{
		Position: engo.Point{X: x, Y: y},
		Width:    width,
		Height:   height,
	}

	// Add to HUD entities list for cleanup
	hud.hudEntities = append(hud.hudEntities, &basic)

	// Add to render system so it actually appears on screen
	if hud.renderSystem != nil {
		hud.renderSystem.Add(&basic, &renderComponent, &spaceComponent)
	}
}

// AddChatMessage adds a new chat message to the display
func (hud *HUDSystem) AddChatMessage(sender, message string) {
	chatMsg := ChatMessage{
		Sender:    sender,
		Message:   message,
		Timestamp: time.Now(),
		Color:     hud.hudColor,
	}

	hud.chatMessages = append(hud.chatMessages, chatMsg)

	// Keep only the most recent messages
	if len(hud.chatMessages) > hud.maxChatLines*2 {
		hud.chatMessages = hud.chatMessages[len(hud.chatMessages)-hud.maxChatLines:]
	}
}

// SetConnectionStatus sets the connection status display
func (hud *HUDSystem) SetConnectionStatus(status string) {
	hud.connectionStatus = status
}

// UpdateGameState updates the HUD with current game state
func (hud *HUDSystem) UpdateGameState(gameState *engine.GameState) {
	hud.teamStates = gameState.Teams

	// Find the player's ship (this would need to be implemented based on player tracking)
	// For now, we'll use the first ship as an example
	for _, shipState := range gameState.Ships {
		hud.currentShip = &shipState
		break
	}
}

// SetMinimapEnabled enables or disables the minimap
func (hud *HUDSystem) SetMinimapEnabled(enabled bool) {
	hud.minimapEnabled = enabled
}

// IsMinimapEnabled returns whether the minimap is enabled
func (hud *HUDSystem) IsMinimapEnabled() bool {
	return hud.minimapEnabled
}

// SetMinimapSize sets the size of the minimap
func (hud *HUDSystem) SetMinimapSize(size float32) {
	hud.minimapSize = size
}

// GetMinimapSize returns the current minimap size
func (hud *HUDSystem) GetMinimapSize() float32 {
	return hud.minimapSize
}

// SetFont sets the font used for HUD text rendering
func (hud *HUDSystem) SetFont(font *common.Font) {
	hud.font = font
}

// SetFontSize sets the font size using the asset manager
func (hud *HUDSystem) SetFontSize(size float64) {
	if hud.assetManager != nil {
		hud.font = hud.assetManager.GetFont(size)
	}
}

// SetSmallFont sets the font to small size
func (hud *HUDSystem) SetSmallFont() {
	hud.SetFontSize(float64(FontSizeSmall))
}

// SetMediumFont sets the font to medium size
func (hud *HUDSystem) SetMediumFont() {
	hud.SetFontSize(float64(FontSizeMedium))
}

// SetLargeFont sets the font to large size
func (hud *HUDSystem) SetLargeFont() {
	hud.SetFontSize(float64(FontSizeLarge))
}

// GetChatMessages returns the current chat messages
func (hud *HUDSystem) GetChatMessages() []ChatMessage {
	return hud.chatMessages
}

// ClearChatMessages clears all chat messages
func (hud *HUDSystem) ClearChatMessages() {
	hud.chatMessages = hud.chatMessages[:0]
}
