// pkg/config/config.go
package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/opd-ai/go-netrek/pkg/entity"
)

// GameConfig contains configuration for a Netrek game
type GameConfig struct {
	WorldSize     float64        `json:"worldSize"`
	MaxPlayers    int            `json:"maxPlayers"`
	Teams         []TeamConfig   `json:"teams"`
	Planets       []PlanetConfig `json:"planets"`
	PhysicsConfig PhysicsConfig  `json:"physics"`
	NetworkConfig NetworkConfig  `json:"network"`
	GameRules     GameRules      `json:"gameRules"`
}

// TeamConfig contains configuration for a team
type TeamConfig struct {
	Name         string `json:"name"`
	Color        string `json:"color"`
	MaxShips     int    `json:"maxShips"`
	StartingShip string `json:"startingShip"`
}

// PlanetConfig contains configuration for a planet
type PlanetConfig struct {
	Name          string            `json:"name"`
	X             float64           `json:"x"`
	Y             float64           `json:"y"`
	Type          entity.PlanetType `json:"type"`
	HomeWorld     bool              `json:"homeWorld"`
	TeamID        int               `json:"teamID"`
	InitialArmies int               `json:"initialArmies"`
}

// PhysicsConfig contains physics-related configuration
type PhysicsConfig struct {
	Gravity         float64 `json:"gravity"`
	Friction        float64 `json:"friction"`
	CollisionDamage float64 `json:"collisionDamage"`
}

// NetworkConfig contains network-related configuration
type NetworkConfig struct {
	UpdateRate      int    `json:"updateRate"`
	TicksPerState   int    `json:"ticksPerState"`
	UsePartialState bool   `json:"usePartialState"`
	ServerPort      int    `json:"serverPort"`
	ServerAddress   string `json:"serverAddress"`
}

// GameRules contains game rules configuration
type GameRules struct {
	WinCondition   string `json:"winCondition"`
	TimeLimit      int    `json:"timeLimit"`
	MaxScore       int    `json:"maxScore"`
	RespawnDelay   int    `json:"respawnDelay"`
	FriendlyFire   bool   `json:"friendlyFire"`
	StartingArmies int    `json:"startingArmies"`
}

// LoadConfig loads a configuration from a file
func LoadConfig(path string) (*GameConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config GameConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveConfig saves a configuration to a file
func SaveConfig(config *GameConfig, path string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := ioutil.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// DefaultConfig returns a default game configuration
func DefaultConfig() *GameConfig {
	return &GameConfig{
		WorldSize:  10000,
		MaxPlayers: 16,
		Teams: []TeamConfig{
			{
				Name:         "Federation",
				Color:        "#0000FF",
				MaxShips:     8,
				StartingShip: "Scout",
			},
			{
				Name:         "Romulans",
				Color:        "#00FF00",
				MaxShips:     8,
				StartingShip: "Scout",
			},
		},
		Planets: []PlanetConfig{
			{
				Name:          "Earth",
				X:             -4000,
				Y:             0,
				Type:          entity.Homeworld,
				HomeWorld:     true,
				TeamID:        0,
				InitialArmies: 30,
			},
			{
				Name:          "Romulus",
				X:             4000,
				Y:             0,
				Type:          entity.Homeworld,
				HomeWorld:     true,
				TeamID:        1,
				InitialArmies: 30,
			},
			// Additional planets would be defined here
		},
		PhysicsConfig: PhysicsConfig{
			Gravity:         0,
			Friction:        0.1,
			CollisionDamage: 20,
		},
		NetworkConfig: NetworkConfig{
			UpdateRate:      20,
			TicksPerState:   3,
			UsePartialState: true,
			ServerPort:      4566,
			ServerAddress:   "localhost:4566",
		},
		GameRules: GameRules{
			WinCondition:   "conquest",
			TimeLimit:      1800,
			MaxScore:       100,
			RespawnDelay:   5,
			FriendlyFire:   false,
			StartingArmies: 0,
		},
	}
}
