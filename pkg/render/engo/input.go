// pkg/render/engo/input.go
package engo

import (
	"time"

	"github.com/EngoEngine/ecs"
	"github.com/EngoEngine/engo"
	"github.com/EngoEngine/engo/common"

	"github.com/opd-ai/go-netrek/pkg/entity"
	"github.com/opd-ai/go-netrek/pkg/network"
)

// InputSystem handles keyboard and mouse input for the game
type InputSystem struct {
	client *network.GameClient

	// Input state
	thrustPressed    bool
	turnLeftPressed  bool
	turnRightPressed bool
	currentWeapon    int
	targetID         entity.ID

	// Input timing
	lastInputSent time.Time
	inputDelay    time.Duration

	// Chat state
	chatActive bool
	chatBuffer string
	chatCursor int
}

// NewInputSystem creates a new input system
func NewInputSystem(client *network.GameClient) *InputSystem {
	return &InputSystem{
		client:        client,
		currentWeapon: -1,                    // No weapon selected by default
		inputDelay:    time.Millisecond * 50, // Send input every 50ms
	}
}

// Add satisfies the ecs.System interface
func (is *InputSystem) Add(basic *ecs.BasicEntity, render *common.RenderComponent, space *common.SpaceComponent) {
	// Not used for input system
}

// Remove satisfies the ecs.System interface
func (is *InputSystem) Remove(basic ecs.BasicEntity) {
	// Not used for input system
}

// Update processes input and sends commands to the server
func (is *InputSystem) Update(dt float32) {
	// Handle chat input if chat is active
	if is.chatActive {
		is.handleChatInput()
		return
	}

	// Handle game input
	is.handleGameInput()

	// Send input to server if enough time has passed
	if time.Since(is.lastInputSent) >= is.inputDelay {
		is.sendInputToServer()
		is.lastInputSent = time.Now()
	}
}

// handleGameInput processes game-related input
func (is *InputSystem) handleGameInput() {
	// Movement controls
	is.thrustPressed = engo.Input.Button("thrust").Down()
	is.turnLeftPressed = engo.Input.Button("turnLeft").Down()
	is.turnRightPressed = engo.Input.Button("turnRight").Down()

	// Weapon selection (number keys 1-9)
	for i := 1; i <= 9; i++ {
		if engo.Input.Button(string(rune('0' + i))).JustPressed() {
			is.currentWeapon = i - 1 // Convert to 0-based index
		}
	}

	// Chat activation
	if engo.Input.Button("chat").JustPressed() {
		is.activateChat()
	}

	// Targeting (for now, we'll use a placeholder implementation)
	if engo.Input.Button("target").JustPressed() {
		is.targetID = is.findNearestTarget()
	}
}

// handleChatInput processes chat input when chat is active
func (is *InputSystem) handleChatInput() {
	// Handle text input for chat
	// Note: Engo doesn't have TextInput(), so we'll use a simplified approach
	// for now with just Enter and Backspace keys
	if engo.Input.Button("enter").JustPressed() {
		// Send chat message
		is.sendChatMessage()
		is.deactivateChat()
	} else if engo.Input.Button("backspace").JustPressed() {
		// Backspace
		if len(is.chatBuffer) > 0 && is.chatCursor > 0 {
			is.chatBuffer = is.chatBuffer[:is.chatCursor-1] + is.chatBuffer[is.chatCursor:]
			is.chatCursor--
		}
	}

	// Handle escape to cancel chat
	if engo.Input.Button("escape").JustPressed() {
		is.deactivateChat()
	}
}

// sendInputToServer sends the current input state to the server
func (is *InputSystem) sendInputToServer() {
	fireWeapon := -1
	if engo.Input.Button("fire").JustPressed() && is.currentWeapon >= 0 {
		fireWeapon = is.currentWeapon
	}

	beamDown := engo.Input.Button("beamDown").Down()
	beamUp := engo.Input.Button("beamUp").Down()
	beamAmount := 1 // Default beam amount

	// Adjust beam amount based on modifier keys
	if engo.Input.Button("modifier").Down() {
		beamAmount = 5
	}

	// Send input to server
	err := is.client.SendInput(
		is.thrustPressed,
		is.turnLeftPressed,
		is.turnRightPressed,
		fireWeapon,
		beamDown,
		beamUp,
		beamAmount,
		is.targetID,
	)

	if err != nil {
		// Log error (use fmt.Printf since engo.Log doesn't exist)
		// fmt.Printf("Failed to send input: %v\n", err)
		_ = err // Suppress error for now
	}
}

// activateChat activates chat input mode
func (is *InputSystem) activateChat() {
	is.chatActive = true
	is.chatBuffer = ""
	is.chatCursor = 0
}

// deactivateChat deactivates chat input mode
func (is *InputSystem) deactivateChat() {
	is.chatActive = false
	is.chatBuffer = ""
	is.chatCursor = 0
}

// sendChatMessage sends the current chat buffer as a message
func (is *InputSystem) sendChatMessage() {
	if len(is.chatBuffer) > 0 {
		err := is.client.SendChatMessage(is.chatBuffer)
		if err != nil {
			// Log error (use fmt.Printf since engo.Log doesn't exist)
			// fmt.Printf("Failed to send chat message: %v\n", err)
			_ = err // Suppress error for now
		}
	}
}

// findNearestTarget finds the nearest targetable entity
// This is a placeholder implementation
func (is *InputSystem) findNearestTarget() entity.ID {
	// In a real implementation, this would:
	// 1. Get the current game state
	// 2. Find the player's ship position
	// 3. Find the nearest enemy ship or planet
	// 4. Return its ID
	return 0 // Placeholder
}

// IsChatActive returns whether chat input is currently active
func (is *InputSystem) IsChatActive() bool {
	return is.chatActive
}

// GetChatBuffer returns the current chat input buffer
func (is *InputSystem) GetChatBuffer() string {
	return is.chatBuffer
}

// GetChatCursor returns the current chat cursor position
func (is *InputSystem) GetChatCursor() int {
	return is.chatCursor
}

// SetupInputBindings sets up the key bindings for the game
func SetupInputBindings() {
	// Movement keys
	engo.Input.RegisterButton("thrust", engo.KeyW, engo.KeyArrowUp)
	engo.Input.RegisterButton("turnLeft", engo.KeyA, engo.KeyArrowLeft)
	engo.Input.RegisterButton("turnRight", engo.KeyD, engo.KeyArrowRight)

	// Weapon and action keys
	engo.Input.RegisterButton("fire", engo.KeySpace)
	engo.Input.RegisterButton("beamDown", engo.KeyB)
	engo.Input.RegisterButton("beamUp", engo.KeyB) // With Shift modifier
	engo.Input.RegisterButton("target", engo.KeyT)

	// Number keys for weapon selection (simplified for now)
	// Note: Engo key constants may differ, using a simplified approach
	engo.Input.RegisterButton("weapon1", engo.KeyOne)
	engo.Input.RegisterButton("weapon2", engo.KeyTwo)
	engo.Input.RegisterButton("weapon3", engo.KeyThree)

	// UI keys
	engo.Input.RegisterButton("chat", engo.KeyEnter)
	engo.Input.RegisterButton("escape", engo.KeyEscape)
	engo.Input.RegisterButton("backspace", engo.KeyBackspace)
	engo.Input.RegisterButton("enter", engo.KeyEnter)
	engo.Input.RegisterButton("modifier", engo.KeyLeftShift)

	// Mouse bindings - separate registration for mouse
	// Note: Mouse buttons need to be handled differently in Engo
}
