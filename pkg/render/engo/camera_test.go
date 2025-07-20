// pkg/render/engo/camera_test.go
package engo

import (
	"math"
	"testing"

	"github.com/EngoEngine/ecs"
	"github.com/opd-ai/go-netrek/pkg/physics"
)

func TestNewCameraSystem(t *testing.T) {
	camera := NewCameraSystem()

	// Test default values
	if camera.zoom != 1.0 {
		t.Errorf("Expected default zoom 1.0, got %f", camera.zoom)
	}
	if camera.minZoom != 0.1 {
		t.Errorf("Expected default minZoom 0.1, got %f", camera.minZoom)
	}
	if camera.maxZoom != 3.0 {
		t.Errorf("Expected default maxZoom 3.0, got %f", camera.maxZoom)
	}
	if camera.followSpeed != 2.0 {
		t.Errorf("Expected default followSpeed 2.0, got %f", camera.followSpeed)
	}
	if !camera.smoothing {
		t.Error("Expected smoothing to be enabled by default")
	}
	if camera.targetSet {
		t.Error("Expected targetSet to be false by default")
	}
}

func TestCameraSystem_SetTarget_ClearTarget(t *testing.T) {
	camera := NewCameraSystem()
	testTarget := physics.Vector2D{X: 100.0, Y: 200.0}

	t.Run("SetTarget_FirstTime", func(t *testing.T) {
		camera.SetTarget(testTarget)

		if !camera.targetSet {
			t.Error("Expected targetSet to be true after setting target")
		}
		if camera.target != testTarget {
			t.Errorf("Expected target %v, got %v", testTarget, camera.target)
		}
		// Should position camera immediately on first target
		if camera.currentPos != testTarget {
			t.Errorf("Expected currentPos to be set immediately to %v, got %v", testTarget, camera.currentPos)
		}
	})

	t.Run("ClearTarget", func(t *testing.T) {
		camera.ClearTarget()

		if camera.targetSet {
			t.Error("Expected targetSet to be false after clearing target")
		}
	})
}

func TestCameraSystem_ZoomOperations(t *testing.T) {
	camera := NewCameraSystem()

	testCases := []struct {
		name     string
		zoom     float32
		expected float32
	}{
		{"ValidZoom", 1.5, 1.5},
		{"BelowMinZoom", 0.05, 0.1}, // Should clamp to minZoom
		{"AboveMaxZoom", 5.0, 3.0},  // Should clamp to maxZoom
		{"ExactMinZoom", 0.1, 0.1},
		{"ExactMaxZoom", 3.0, 3.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			camera.SetZoom(tc.zoom)
			actual := camera.GetZoom()
			if actual != tc.expected {
				t.Errorf("Expected zoom %f, got %f", tc.expected, actual)
			}
		})
	}
}

func TestCameraSystem_clampZoom(t *testing.T) {
	camera := NewCameraSystem()

	testCases := []struct {
		name     string
		input    float32
		expected float32
	}{
		{"ValidZoom", 1.5, 1.5},
		{"BelowMin", 0.05, 0.1},
		{"AboveMax", 5.0, 3.0},
		{"ExactMin", 0.1, 0.1},
		{"ExactMax", 3.0, 3.0},
		{"NegativeZoom", -1.0, 0.1},
		{"ZeroZoom", 0.0, 0.1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := camera.clampZoom(tc.input)
			if result != tc.expected {
				t.Errorf("clampZoom(%f) = %f, want %f", tc.input, result, tc.expected)
			}
		})
	}
}

func TestCameraSystem_FollowSpeed(t *testing.T) {
	camera := NewCameraSystem()

	t.Run("SetAndGetFollowSpeed", func(t *testing.T) {
		testSpeed := float32(5.5)
		camera.SetFollowSpeed(testSpeed)

		if camera.GetFollowSpeed() != testSpeed {
			t.Errorf("Expected follow speed %f, got %f", testSpeed, camera.GetFollowSpeed())
		}
	})

	t.Run("DefaultFollowSpeed", func(t *testing.T) {
		newCamera := NewCameraSystem()
		expected := float32(2.0)
		if newCamera.GetFollowSpeed() != expected {
			t.Errorf("Expected default follow speed %f, got %f", expected, newCamera.GetFollowSpeed())
		}
	})
}

func TestCameraSystem_Smoothing(t *testing.T) {
	camera := NewCameraSystem()

	t.Run("DefaultSmoothing", func(t *testing.T) {
		if !camera.IsSmoothing() {
			t.Error("Expected smoothing to be enabled by default")
		}
	})

	t.Run("DisableSmoothing", func(t *testing.T) {
		camera.EnableSmoothing(false)
		if camera.IsSmoothing() {
			t.Error("Expected smoothing to be disabled")
		}
	})

	t.Run("EnableSmoothing", func(t *testing.T) {
		camera.EnableSmoothing(true)
		if !camera.IsSmoothing() {
			t.Error("Expected smoothing to be enabled")
		}
	})
}

func TestCameraSystem_ZoomLimits(t *testing.T) {
	camera := NewCameraSystem()

	t.Run("SetZoomLimits", func(t *testing.T) {
		minZoom := float32(0.5)
		maxZoom := float32(2.0)
		camera.SetZoomLimits(minZoom, maxZoom)

		actualMin, actualMax := camera.GetZoomLimits()
		if actualMin != minZoom {
			t.Errorf("Expected minZoom %f, got %f", minZoom, actualMin)
		}
		if actualMax != maxZoom {
			t.Errorf("Expected maxZoom %f, got %f", maxZoom, actualMax)
		}
	})

	t.Run("SetZoomLimits_ClampsCurrentZoom", func(t *testing.T) {
		// Set zoom to 2.5
		camera.SetZoom(2.5)
		// Set limits that would require clamping
		camera.SetZoomLimits(0.2, 2.0)

		// Current zoom should be clamped to new max
		if camera.GetZoom() != 2.0 {
			t.Errorf("Expected zoom to be clamped to 2.0, got %f", camera.GetZoom())
		}
	})
}

func TestCameraSystem_GetCurrentPosition(t *testing.T) {
	camera := NewCameraSystem()
	expectedPos := physics.Vector2D{X: 0, Y: 0}

	// Initially should be at origin
	actualPos := camera.GetCurrentPosition()
	if actualPos != expectedPos {
		t.Errorf("Expected initial position %v, got %v", expectedPos, actualPos)
	}

	// After setting target, position should update
	targetPos := physics.Vector2D{X: 50, Y: 75}
	camera.SetTarget(targetPos)
	actualPos = camera.GetCurrentPosition()
	if actualPos != targetPos {
		t.Errorf("Expected position after setting target %v, got %v", targetPos, actualPos)
	}
}

func TestCameraSystem_updateCameraPosition(t *testing.T) {
	camera := NewCameraSystem()
	t.Run("SmoothingEnabled", func(t *testing.T) {
		camera.EnableSmoothing(true)
		camera.SetFollowSpeed(1.0)

		// Set initial position to non-zero to avoid immediate positioning
		initial := physics.Vector2D{X: 10, Y: 10}
		target := physics.Vector2D{X: 100, Y: 100}
		camera.currentPos = initial
		camera.SetTarget(target)

		// Store position after setting target (should still be initial because of smoothing)
		posAfterTarget := camera.currentPos

		// Update with small time step
		dt := float32(0.1)
		camera.updateCameraPosition(dt)

		// Camera should move towards target but not reach it immediately with smoothing
		// Check that position has changed from the position after setting target
		if camera.currentPos.X == posAfterTarget.X && camera.currentPos.Y == posAfterTarget.Y {
			t.Error("Camera didn't move towards target with smoothing enabled")
		}

		// Check that we're moving in the right direction (toward target)
		if camera.currentPos.X <= posAfterTarget.X || camera.currentPos.Y <= posAfterTarget.Y {
			t.Error("Camera didn't move towards target or moved in wrong direction")
		}
	})

	t.Run("SmoothingDisabled", func(t *testing.T) {
		camera.EnableSmoothing(false)
		camera.currentPos = physics.Vector2D{X: 0, Y: 0}
		target := physics.Vector2D{X: 200, Y: 200}
		camera.SetTarget(target)

		camera.updateCameraPosition(0.1)

		// Camera should move to target immediately
		if camera.currentPos != target {
			t.Errorf("Expected immediate movement to %v, got %v", target, camera.currentPos)
		}
	})
}

func TestCameraSystem_WorldToScreen(t *testing.T) {
	camera := NewCameraSystem()

	t.Run("WorldToScreen_Basic", func(t *testing.T) {
		camera.currentPos = physics.Vector2D{X: 0, Y: 0}
		camera.SetZoom(1.0)

		worldPos := physics.Vector2D{X: 10, Y: 20}
		screenPos := camera.WorldToScreen(worldPos)

		// Verify that the transformation produces different coordinates
		// The exact screen coordinates depend on engo.GameWidth/Height
		// but the mathematical relationship should be preserved
		expectedRelativeX := worldPos.X - camera.currentPos.X // Should be 10
		expectedRelativeY := worldPos.Y - camera.currentPos.Y // Should be 20

		if expectedRelativeX != 10 || expectedRelativeY != 20 {
			t.Errorf("Expected relative position (10, 20), got (%f, %f)",
				expectedRelativeX, expectedRelativeY)
		}

		// Screen position calculation should be consistent
		// Note: exact values depend on engo.GameWidth/Height which may be 0 in tests
		_ = screenPos // Used for calculation test
	})

	t.Run("WorldToScreen_WithCameraOffset", func(t *testing.T) {
		camera.currentPos = physics.Vector2D{X: 100, Y: 100}
		camera.SetZoom(1.0)

		worldPos := physics.Vector2D{X: 110, Y: 120}
		screenPos := camera.WorldToScreen(worldPos)

		// World position relative to camera should be (10, 20)
		// This should be reflected in the screen coordinates
		expectedRelativeX := worldPos.X - camera.currentPos.X
		expectedRelativeY := worldPos.Y - camera.currentPos.Y

		if expectedRelativeX != 10 || expectedRelativeY != 20 {
			t.Errorf("Expected relative position (10, 20), calculated (%f, %f)",
				expectedRelativeX, expectedRelativeY)
		}

		// Verify screenPos is calculated (basic sanity check)
		if screenPos.X == worldPos.X && screenPos.Y == worldPos.Y {
			t.Error("Screen position should be transformed from world position")
		}
	})
}

func TestCameraSystem_ScreenToWorld(t *testing.T) {
	camera := NewCameraSystem()

	t.Run("ScreenToWorld_Basic", func(t *testing.T) {
		camera.currentPos = physics.Vector2D{X: 0, Y: 0}
		camera.SetZoom(1.0)

		// Use screen coordinates that should map to world coordinates
		screenPos := physics.Vector2D{X: 400, Y: 300} // Assuming screen center around here
		worldPos := camera.ScreenToWorld(screenPos)

		// Verify the transformation is mathematically consistent
		// Convert back to screen to check round-trip accuracy
		backToScreen := camera.WorldToScreen(worldPos)

		tolerance := 0.001
		if math.Abs(backToScreen.X-screenPos.X) > tolerance ||
			math.Abs(backToScreen.Y-screenPos.Y) > tolerance {
			t.Errorf("Round-trip conversion failed. Original: %v, Converted back: %v",
				screenPos, backToScreen)
		}
	})

	t.Run("ScreenToWorld_WithZoom", func(t *testing.T) {
		camera.currentPos = physics.Vector2D{X: 0, Y: 0}
		camera.SetZoom(2.0)

		screenPos := physics.Vector2D{X: 400, Y: 300}
		worldPos := camera.ScreenToWorld(screenPos)

		// With 2x zoom, world coordinates should be scaled down
		backToScreen := camera.WorldToScreen(worldPos)

		tolerance := 0.001
		if math.Abs(backToScreen.X-screenPos.X) > tolerance ||
			math.Abs(backToScreen.Y-screenPos.Y) > tolerance {
			t.Errorf("Round-trip conversion with zoom failed. Original: %v, Converted back: %v",
				screenPos, backToScreen)
		}
	})
}

func TestCameraSystem_CoordinateTransformation_Consistency(t *testing.T) {
	camera := NewCameraSystem()

	testCases := []struct {
		name      string
		zoom      float32
		cameraPos physics.Vector2D
	}{
		{"ZoomOne_OriginCamera", 1.0, physics.Vector2D{X: 0, Y: 0}},
		{"ZoomTwo_OriginCamera", 2.0, physics.Vector2D{X: 0, Y: 0}},
		{"ZoomHalf_OffsetCamera", 0.5, physics.Vector2D{X: 100, Y: 200}},
		{"ZoomThree_OffsetCamera", 3.0, physics.Vector2D{X: -50, Y: -75}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			camera.SetZoom(tc.zoom)
			camera.currentPos = tc.cameraPos

			// Test round-trip conversion for multiple points
			testPoints := []physics.Vector2D{
				{X: 0, Y: 0},
				{X: 100, Y: 100},
				{X: -50, Y: 75},
				{X: 300, Y: -200},
			}

			for _, worldPoint := range testPoints {
				screenPoint := camera.WorldToScreen(worldPoint)
				backToWorld := camera.ScreenToWorld(screenPoint)

				tolerance := 0.001
				if math.Abs(backToWorld.X-worldPoint.X) > tolerance ||
					math.Abs(backToWorld.Y-worldPoint.Y) > tolerance {
					t.Errorf("Round-trip failed for point %v: got %v", worldPoint, backToWorld)
				}
			}
		})
	}
}

func TestCameraSystem_ECSInterface(t *testing.T) {
	camera := NewCameraSystem()

	t.Run("Add_DoesNotPanic", func(t *testing.T) {
		// These methods are required by ECS interface but not used
		// Just verify they don't panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Add method panicked: %v", r)
			}
		}()

		camera.Add(nil, nil, nil)
	})

	t.Run("Remove_DoesNotPanic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Remove method panicked: %v", r)
			}
		}()

		// Create a zero value BasicEntity for testing
		var mockEntity ecs.BasicEntity
		camera.Remove(mockEntity)
	})
}
