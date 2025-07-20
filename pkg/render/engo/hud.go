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
}

// ChatMessage represents a chat message in the HUD
type ChatMessage struct {
	Sender    string
	Message   string
	Timestamp time.Time
	Color     color.Color
}

// NewHUDSystem creates a new HUD system
func NewHUDSystem() *HUDSystem {
	return &HUDSystem{
		connectionStatus: "Connected",
		maxChatLines:     10,
		minimapEnabled:   true,
		minimapSize:      200.0,
		teamStates:       make(map[int]engine.TeamState),
		hudColor:         color.RGBA{255, 255, 255, 255},
		friendlyColor:    color.RGBA{0, 255, 0, 255},
		enemyColor:       color.RGBA{255, 0, 0, 255},
		neutralColor:     color.RGBA{128, 128, 128, 255},
	}
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

// clearHUDEntities removes previous HUD entities
func (hud *HUDSystem) clearHUDEntities() {
	// In a real implementation, you would properly remove entities from ECS
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

	// Render status text at top-left corner
	hud.renderText(statusText, 10, 10, hud.hudColor)
}

// renderChatWindow renders the chat message window
func (hud *HUDSystem) renderChatWindow() {
	chatStartY := float32(engo.GameHeight()) - 200

	// Render chat background
	hud.renderRect(10, chatStartY, 400, 150, color.RGBA{0, 0, 0, 128})

	// Render chat messages
	y := chatStartY + 10
	for i := len(hud.chatMessages) - 1; i >= 0 && i >= len(hud.chatMessages)-hud.maxChatLines; i-- {
		msg := hud.chatMessages[i]
		chatLine := fmt.Sprintf("[%s]: %s", msg.Sender, msg.Message)
		hud.renderText(chatLine, 15, y, msg.Color)
		y += 15
	}
}

// renderTeamStatus renders team scores and status
func (hud *HUDSystem) renderTeamStatus() {
	startY := float32(100)

	for teamID, teamState := range hud.teamStates {
		teamText := fmt.Sprintf(
			"%s: Score %d, Ships %d, Planets %d",
			teamState.Name,
			teamState.Score,
			teamState.ShipCount,
			teamState.PlanetCount,
		)

		teamColor := hud.getTeamColor(teamID)
		hud.renderText(teamText, 10, startY, teamColor)
		startY += 20
	}
}

// renderConnectionStatus renders the connection status
func (hud *HUDSystem) renderConnectionStatus() {
	statusColor := hud.friendlyColor
	if hud.connectionStatus != "Connected" {
		statusColor = hud.enemyColor
	}

	hud.renderText(
		"Status: "+hud.connectionStatus,
		float32(engo.GameWidth())-150,
		10,
		statusColor,
	)
}

// renderMinimap renders a minimap showing the game world
func (hud *HUDSystem) renderMinimap() {
	minimapX := float32(engo.GameWidth()) - hud.minimapSize - 10
	minimapY := float32(10)

	// Render minimap background
	hud.renderRect(minimapX, minimapY, hud.minimapSize, hud.minimapSize, color.RGBA{0, 0, 0, 128})

	// Render minimap border
	hud.renderRectOutline(minimapX, minimapY, hud.minimapSize, hud.minimapSize, hud.hudColor)

	// Note: In a real implementation, you would render planets and ships on the minimap
	// based on the current game state
}

// renderText renders text at the specified position
func (hud *HUDSystem) renderText(text string, x, y float32, textColor color.Color) {
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
		Width:    float32(len(text) * 8), // Approximate width
		Height:   16,                     // Approximate height
	}

	// Add to HUD entities
	hud.hudEntities = append(hud.hudEntities, &basic)

	// Note: In a real implementation, you would add these components to the render system
	// This is a simplified version for demonstration
	_ = renderComponent // Use the component
	_ = spaceComponent  // Use the component
}

// renderRect renders a filled rectangle
func (hud *HUDSystem) renderRect(x, y, width, height float32, rectColor color.Color) {
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

	// Add to HUD entities
	hud.hudEntities = append(hud.hudEntities, &basic)

	// Note: In a real implementation, you would add these components to the render system
	_ = renderComponent // Use the component
	_ = spaceComponent  // Use the component
}

// renderRectOutline renders a rectangle outline
func (hud *HUDSystem) renderRectOutline(x, y, width, height float32, outlineColor color.Color) {
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

	// Add to HUD entities
	hud.hudEntities = append(hud.hudEntities, &basic)

	// Note: In a real implementation, you would add these components to the render system
	_ = renderComponent // Use the component
	_ = spaceComponent  // Use the component
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

// getTeamColor returns the color for a specific team
func (hud *HUDSystem) getTeamColor(teamID int) color.Color {
	teamColors := []color.Color{
		color.RGBA{255, 0, 0, 255},   // Red
		color.RGBA{0, 255, 0, 255},   // Green
		color.RGBA{0, 0, 255, 255},   // Blue
		color.RGBA{255, 255, 0, 255}, // Yellow
	}

	if teamID >= 0 && teamID < len(teamColors) {
		return teamColors[teamID]
	}

	return hud.neutralColor
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

// GetChatMessages returns the current chat messages
func (hud *HUDSystem) GetChatMessages() []ChatMessage {
	return hud.chatMessages
}

// ClearChatMessages clears all chat messages
func (hud *HUDSystem) ClearChatMessages() {
	hud.chatMessages = hud.chatMessages[:0]
}
