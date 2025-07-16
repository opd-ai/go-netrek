package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/opd-ai/go-netrek/pkg/entity"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	// Test basic structure
	if config == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	// Test world size
	if config.WorldSize != 10000 {
		t.Errorf("Expected WorldSize 10000, got %f", config.WorldSize)
	}

	// Test max players
	if config.MaxPlayers != 16 {
		t.Errorf("Expected MaxPlayers 16, got %d", config.MaxPlayers)
	}

	// Test teams
	if len(config.Teams) != 2 {
		t.Errorf("Expected 2 teams, got %d", len(config.Teams))
	}

	// Test team configuration
	if config.Teams[0].Name != "Federation" {
		t.Errorf("Expected first team name 'Federation', got '%s'", config.Teams[0].Name)
	}
	if config.Teams[1].Name != "Romulans" {
		t.Errorf("Expected second team name 'Romulans', got '%s'", config.Teams[1].Name)
	}

	// Test planets
	if len(config.Planets) != 2 {
		t.Errorf("Expected 2 planets, got %d", len(config.Planets))
	}

	// Test specific planet configuration
	earth := config.Planets[0]
	if earth.Name != "Earth" {
		t.Errorf("Expected first planet name 'Earth', got '%s'", earth.Name)
	}
	if earth.Type != entity.Homeworld {
		t.Errorf("Expected Earth to be Homeworld type, got %v", earth.Type)
	}
	if !earth.HomeWorld {
		t.Error("Expected Earth to be marked as HomeWorld")
	}
	if earth.TeamID != 0 {
		t.Errorf("Expected Earth TeamID 0, got %d", earth.TeamID)
	}

	// Test physics config
	if config.PhysicsConfig.Gravity != 0 {
		t.Errorf("Expected Gravity 0, got %f", config.PhysicsConfig.Gravity)
	}
	if config.PhysicsConfig.Friction != 0.1 {
		t.Errorf("Expected Friction 0.1, got %f", config.PhysicsConfig.Friction)
	}

	// Test network config
	if config.NetworkConfig.UpdateRate != 20 {
		t.Errorf("Expected UpdateRate 20, got %d", config.NetworkConfig.UpdateRate)
	}
	if config.NetworkConfig.ServerPort != 4566 {
		t.Errorf("Expected ServerPort 4566, got %d", config.NetworkConfig.ServerPort)
	}

	// Test game rules
	if config.GameRules.WinCondition != "conquest" {
		t.Errorf("Expected WinCondition 'conquest', got '%s'", config.GameRules.WinCondition)
	}
	if config.GameRules.FriendlyFire {
		t.Error("Expected FriendlyFire to be false")
	}
}

func TestLoadConfig_Success(t *testing.T) {
	// Create temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.json")

	// Create test config data
	testConfig := &GameConfig{
		WorldSize:  5000,
		MaxPlayers: 8,
		Teams: []TeamConfig{
			{
				Name:         "TestTeam",
				Color:        "#FF0000",
				MaxShips:     4,
				StartingShip: "Destroyer",
			},
		},
		Planets: []PlanetConfig{
			{
				Name:          "TestPlanet",
				X:             100,
				Y:             200,
				Type:          entity.Industrial,
				HomeWorld:     false,
				TeamID:        1,
				InitialArmies: 15,
			},
		},
		PhysicsConfig: PhysicsConfig{
			Gravity:         9.8,
			Friction:        0.2,
			CollisionDamage: 30,
		},
		NetworkConfig: NetworkConfig{
			UpdateRate:      30,
			TicksPerState:   2,
			UsePartialState: false,
			ServerPort:      8080,
			ServerAddress:   "test.example.com:8080",
		},
		GameRules: GameRules{
			WinCondition:   "deathmatch",
			TimeLimit:      3600,
			MaxScore:       50,
			RespawnDelay:   10,
			FriendlyFire:   true,
			StartingArmies: 5,
		},
	}

	// Write test config to file
	data, err := json.MarshalIndent(testConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test config: %v", err)
	}

	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Test loading config
	loadedConfig, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify loaded config matches original
	if loadedConfig.WorldSize != testConfig.WorldSize {
		t.Errorf("Expected WorldSize %f, got %f", testConfig.WorldSize, loadedConfig.WorldSize)
	}
	if loadedConfig.MaxPlayers != testConfig.MaxPlayers {
		t.Errorf("Expected MaxPlayers %d, got %d", testConfig.MaxPlayers, loadedConfig.MaxPlayers)
	}
	if len(loadedConfig.Teams) != len(testConfig.Teams) {
		t.Errorf("Expected %d teams, got %d", len(testConfig.Teams), len(loadedConfig.Teams))
	}
	if loadedConfig.Teams[0].Name != testConfig.Teams[0].Name {
		t.Errorf("Expected team name '%s', got '%s'", testConfig.Teams[0].Name, loadedConfig.Teams[0].Name)
	}
	if loadedConfig.PhysicsConfig.Gravity != testConfig.PhysicsConfig.Gravity {
		t.Errorf("Expected Gravity %f, got %f", testConfig.PhysicsConfig.Gravity, loadedConfig.PhysicsConfig.Gravity)
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	nonExistentPath := "/path/that/does/not/exist/config.json"

	config, err := LoadConfig(nonExistentPath)

	if err == nil {
		t.Error("Expected error when loading non-existent file, got nil")
	}
	if config != nil {
		t.Error("Expected nil config when file not found, got non-nil")
	}

	// Check error message contains expected information
	expectedSubstring := "failed to open config file"
	if err != nil && len(err.Error()) > 0 {
		if !contains(err.Error(), expectedSubstring) {
			t.Errorf("Expected error to contain '%s', got '%s'", expectedSubstring, err.Error())
		}
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	// Create temporary file with invalid JSON
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid_config.json")

	invalidJSON := `{"worldSize": 5000, "maxPlayers": 8, invalid json}`
	err := os.WriteFile(configPath, []byte(invalidJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid JSON file: %v", err)
	}

	config, err := LoadConfig(configPath)

	if err == nil {
		t.Error("Expected error when loading invalid JSON, got nil")
	}
	if config != nil {
		t.Error("Expected nil config when JSON is invalid, got non-nil")
	}

	// Check error message contains expected information
	expectedSubstring := "failed to parse config file"
	if err != nil {
		if !contains(err.Error(), expectedSubstring) {
			t.Errorf("Expected error to contain '%s', got '%s'", expectedSubstring, err.Error())
		}
	}
}

func TestSaveConfig_Success(t *testing.T) {
	// Create test config
	testConfig := &GameConfig{
		WorldSize:  7500,
		MaxPlayers: 12,
		Teams: []TeamConfig{
			{
				Name:         "SaveTest",
				Color:        "#00FF00",
				MaxShips:     6,
				StartingShip: "Cruiser",
			},
		},
		PhysicsConfig: PhysicsConfig{
			Gravity:         1.5,
			Friction:        0.15,
			CollisionDamage: 25,
		},
	}

	// Create temporary file path
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "save_test_config.json")

	// Test saving config
	err := SaveConfig(testConfig, configPath)
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Load the saved config and verify contents
	loadedConfig, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loadedConfig.WorldSize != testConfig.WorldSize {
		t.Errorf("Expected WorldSize %f, got %f", testConfig.WorldSize, loadedConfig.WorldSize)
	}
	if loadedConfig.MaxPlayers != testConfig.MaxPlayers {
		t.Errorf("Expected MaxPlayers %d, got %d", testConfig.MaxPlayers, loadedConfig.MaxPlayers)
	}
	if len(loadedConfig.Teams) != len(testConfig.Teams) {
		t.Errorf("Expected %d teams, got %d", len(testConfig.Teams), len(loadedConfig.Teams))
	}
	if len(loadedConfig.Teams) > 0 && loadedConfig.Teams[0].Name != testConfig.Teams[0].Name {
		t.Errorf("Expected team name '%s', got '%s'", testConfig.Teams[0].Name, loadedConfig.Teams[0].Name)
	}
}

func TestSaveConfig_InvalidPath(t *testing.T) {
	testConfig := DefaultConfig()

	// Try to save to invalid path (directory that doesn't exist and can't be created)
	invalidPath := "/root/nonexistent/directory/config.json"

	err := SaveConfig(testConfig, invalidPath)

	if err == nil {
		t.Error("Expected error when saving to invalid path, got nil")
	}

	// Check error message contains expected information
	expectedSubstring := "failed to write config file"
	if err != nil {
		if !contains(err.Error(), expectedSubstring) {
			t.Errorf("Expected error to contain '%s', got '%s'", expectedSubstring, err.Error())
		}
	}
}

func TestSaveConfig_NilConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nil_config.json")

	// nil marshals to "null" in JSON, which is valid
	err := SaveConfig(nil, configPath)

	if err != nil {
		t.Errorf("Unexpected error when saving nil config: %v", err)
	}

	// Verify the file was created and contains "null"
	data, readErr := os.ReadFile(configPath)
	if readErr != nil {
		t.Fatalf("Failed to read config file: %v", readErr)
	}

	if string(data) != "null" {
		t.Errorf("Expected file to contain 'null', got '%s'", string(data))
	}
}

// Test table-driven approach for team configurations
func TestDefaultConfig_TeamConfigurations(t *testing.T) {
	config := DefaultConfig()

	tests := []struct {
		name                 string
		teamIndex            int
		expectedName         string
		expectedColor        string
		expectedMaxShips     int
		expectedStartingShip string
	}{
		{
			name:                 "Federation team",
			teamIndex:            0,
			expectedName:         "Federation",
			expectedColor:        "#0000FF",
			expectedMaxShips:     8,
			expectedStartingShip: "Scout",
		},
		{
			name:                 "Romulan team",
			teamIndex:            1,
			expectedName:         "Romulans",
			expectedColor:        "#00FF00",
			expectedMaxShips:     8,
			expectedStartingShip: "Scout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.teamIndex >= len(config.Teams) {
				t.Fatalf("Team index %d out of range, only %d teams available", tt.teamIndex, len(config.Teams))
			}

			team := config.Teams[tt.teamIndex]

			if team.Name != tt.expectedName {
				t.Errorf("Expected name '%s', got '%s'", tt.expectedName, team.Name)
			}
			if team.Color != tt.expectedColor {
				t.Errorf("Expected color '%s', got '%s'", tt.expectedColor, team.Color)
			}
			if team.MaxShips != tt.expectedMaxShips {
				t.Errorf("Expected MaxShips %d, got %d", tt.expectedMaxShips, team.MaxShips)
			}
			if team.StartingShip != tt.expectedStartingShip {
				t.Errorf("Expected StartingShip '%s', got '%s'", tt.expectedStartingShip, team.StartingShip)
			}
		})
	}
}

// Test table-driven approach for planet configurations
func TestDefaultConfig_PlanetConfigurations(t *testing.T) {
	config := DefaultConfig()

	tests := []struct {
		name              string
		planetIndex       int
		expectedName      string
		expectedX         float64
		expectedY         float64
		expectedType      entity.PlanetType
		expectedTeamID    int
		expectedArmies    int
		expectedHomeWorld bool
	}{
		{
			name:              "Earth planet",
			planetIndex:       0,
			expectedName:      "Earth",
			expectedX:         -4000,
			expectedY:         0,
			expectedType:      entity.Homeworld,
			expectedTeamID:    0,
			expectedArmies:    30,
			expectedHomeWorld: true,
		},
		{
			name:              "Romulus planet",
			planetIndex:       1,
			expectedName:      "Romulus",
			expectedX:         4000,
			expectedY:         0,
			expectedType:      entity.Homeworld,
			expectedTeamID:    1,
			expectedArmies:    30,
			expectedHomeWorld: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.planetIndex >= len(config.Planets) {
				t.Fatalf("Planet index %d out of range, only %d planets available", tt.planetIndex, len(config.Planets))
			}

			planet := config.Planets[tt.planetIndex]

			if planet.Name != tt.expectedName {
				t.Errorf("Expected name '%s', got '%s'", tt.expectedName, planet.Name)
			}
			if planet.X != tt.expectedX {
				t.Errorf("Expected X %f, got %f", tt.expectedX, planet.X)
			}
			if planet.Y != tt.expectedY {
				t.Errorf("Expected Y %f, got %f", tt.expectedY, planet.Y)
			}
			if planet.Type != tt.expectedType {
				t.Errorf("Expected Type %v, got %v", tt.expectedType, planet.Type)
			}
			if planet.TeamID != tt.expectedTeamID {
				t.Errorf("Expected TeamID %d, got %d", tt.expectedTeamID, planet.TeamID)
			}
			if planet.InitialArmies != tt.expectedArmies {
				t.Errorf("Expected InitialArmies %d, got %d", tt.expectedArmies, planet.InitialArmies)
			}
			if planet.HomeWorld != tt.expectedHomeWorld {
				t.Errorf("Expected HomeWorld %t, got %t", tt.expectedHomeWorld, planet.HomeWorld)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
