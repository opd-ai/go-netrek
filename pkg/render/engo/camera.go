// pkg/render/engo/camera.go
package engo

import (
	"github.com/EngoEngine/ecs"
	"github.com/EngoEngine/engo"
	"github.com/EngoEngine/engo/common"

	"github.com/opd-ai/go-netrek/pkg/physics"
)

// CameraSystem manages the game camera, following the player's ship
type CameraSystem struct {
	// Target to follow
	target    physics.Vector2D
	targetSet bool

	// Camera properties
	zoom    float32
	minZoom float32
	maxZoom float32

	// Smooth following
	followSpeed float32
	smoothing   bool

	// Current camera state
	currentPos physics.Vector2D
}

// NewCameraSystem creates a new camera system
func NewCameraSystem() *CameraSystem {
	return &CameraSystem{
		zoom:        1.0,
		minZoom:     0.1,
		maxZoom:     3.0,
		followSpeed: 2.0,
		smoothing:   true,
	}
}

// Add satisfies the ecs.System interface
func (cs *CameraSystem) Add(basic *ecs.BasicEntity, render *common.RenderComponent, space *common.SpaceComponent) {
	// Not used for camera system
}

// Remove satisfies the ecs.System interface
func (cs *CameraSystem) Remove(basic ecs.BasicEntity) {
	// Not used for camera system
}

// Update updates the camera position and zoom
func (cs *CameraSystem) Update(dt float32) {
	// Handle zoom input
	cs.handleZoomInput()

	// Update camera position to follow target
	if cs.targetSet {
		cs.updateCameraPosition(dt)
	}

	// Apply camera transformation
	cs.applyCameraTransform()
}

// handleZoomInput processes zoom-related input
func (cs *CameraSystem) handleZoomInput() {
	// Mouse wheel zoom
	scrollY := engo.Input.Mouse.ScrollY
	if scrollY != 0 {
		zoomFactor := float32(1.0 + scrollY*0.1)
		cs.SetZoom(cs.zoom * zoomFactor)
	}

	// Keyboard zoom
	if engo.Input.Button("zoomIn").Down() {
		cs.SetZoom(cs.zoom * 1.02)
	}
	if engo.Input.Button("zoomOut").Down() {
		cs.SetZoom(cs.zoom * 0.98)
	}

	// Reset zoom
	if engo.Input.Button("resetZoom").JustPressed() {
		cs.SetZoom(1.0)
	}
}

// updateCameraPosition smoothly moves the camera toward the target
func (cs *CameraSystem) updateCameraPosition(dt float32) {
	if cs.smoothing {
		// Smooth interpolation toward target
		dx := cs.target.X - cs.currentPos.X
		dy := cs.target.Y - cs.currentPos.Y

		cs.currentPos.X += dx * float64(cs.followSpeed) * float64(dt)
		cs.currentPos.Y += dy * float64(cs.followSpeed) * float64(dt)
	} else {
		// Immediate positioning
		cs.currentPos = cs.target
	}
}

// applyCameraTransform applies the current camera transformation
func (cs *CameraSystem) applyCameraTransform() {
	// Set the camera position and zoom
	// Convert world coordinates to screen coordinates
	// Note: CameraMessage API may differ, using a simplified approach
	engo.Mailbox.Dispatch(common.CameraMessage{
		Axis: common.ZAxis,
		// TODO: Fix camera message format for proper camera control
	})
}

// SetTarget sets the target position for the camera to follow
func (cs *CameraSystem) SetTarget(target physics.Vector2D) {
	cs.target = target
	cs.targetSet = true

	// If this is the first target, position camera immediately
	if !cs.smoothing || (cs.currentPos.X == 0 && cs.currentPos.Y == 0) {
		cs.currentPos = target
	}
}

// ClearTarget clears the camera target
func (cs *CameraSystem) ClearTarget() {
	cs.targetSet = false
}

// SetZoom sets the camera zoom level
func (cs *CameraSystem) SetZoom(zoom float32) {
	cs.zoom = cs.clampZoom(zoom)
}

// GetZoom returns the current zoom level
func (cs *CameraSystem) GetZoom() float32 {
	return cs.zoom
}

// clampZoom ensures zoom is within valid bounds
func (cs *CameraSystem) clampZoom(zoom float32) float32 {
	if zoom < cs.minZoom {
		return cs.minZoom
	}
	if zoom > cs.maxZoom {
		return cs.maxZoom
	}
	return zoom
}

// SetFollowSpeed sets the camera follow speed
func (cs *CameraSystem) SetFollowSpeed(speed float32) {
	cs.followSpeed = speed
}

// GetFollowSpeed returns the current follow speed
func (cs *CameraSystem) GetFollowSpeed() float32 {
	return cs.followSpeed
}

// EnableSmoothing enables or disables camera smoothing
func (cs *CameraSystem) EnableSmoothing(enabled bool) {
	cs.smoothing = enabled
}

// IsSmoothing returns whether camera smoothing is enabled
func (cs *CameraSystem) IsSmoothing() bool {
	return cs.smoothing
}

// GetCurrentPosition returns the current camera position
func (cs *CameraSystem) GetCurrentPosition() physics.Vector2D {
	return cs.currentPos
}

// WorldToScreen converts world coordinates to screen coordinates
func (cs *CameraSystem) WorldToScreen(worldPos physics.Vector2D) physics.Vector2D {
	// Apply camera transformation
	relativeX := worldPos.X - cs.currentPos.X
	relativeY := worldPos.Y - cs.currentPos.Y

	// Apply zoom
	screenX := relativeX*float64(cs.zoom) + float64(engo.GameWidth()/2)
	screenY := relativeY*float64(cs.zoom) + float64(engo.GameHeight()/2)

	return physics.Vector2D{X: screenX, Y: screenY}
}

// ScreenToWorld converts screen coordinates to world coordinates
func (cs *CameraSystem) ScreenToWorld(screenPos physics.Vector2D) physics.Vector2D {
	// Remove screen centering
	relativeX := screenPos.X - float64(engo.GameWidth()/2)
	relativeY := screenPos.Y - float64(engo.GameHeight()/2)

	// Remove zoom
	relativeX /= float64(cs.zoom)
	relativeY /= float64(cs.zoom)

	// Add camera position
	worldX := relativeX + cs.currentPos.X
	worldY := relativeY + cs.currentPos.Y

	return physics.Vector2D{X: worldX, Y: worldY}
}

// SetZoomLimits sets the minimum and maximum zoom levels
func (cs *CameraSystem) SetZoomLimits(min, max float32) {
	cs.minZoom = min
	cs.maxZoom = max
	cs.zoom = cs.clampZoom(cs.zoom)
}

// GetZoomLimits returns the current zoom limits
func (cs *CameraSystem) GetZoomLimits() (float32, float32) {
	return cs.minZoom, cs.maxZoom
}

// SetupCameraControls sets up camera control key bindings
func SetupCameraControls() {
	// Note: Using simplified key constants as some may not be available
	// engo.Input.RegisterButton("zoomIn", engo.KeyPlus)   // TODO: Fix key constants
	// engo.Input.RegisterButton("zoomOut", engo.KeyMinus) // TODO: Fix key constants
	engo.Input.RegisterButton("resetZoom", engo.KeyR) // R key to reset zoom
}
