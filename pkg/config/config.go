// pkg/config/config.go
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/opd-ai/go-netrek/pkg/entity"
)

// EnvironmentConfig contains configuration loaded from environment variables
type EnvironmentConfig struct {
	ServerAddr      string        `env:"NETREK_SERVER_ADDR"`
	ServerPort      int           `env:"NETREK_SERVER_PORT"`
	MaxClients      int           `env:"NETREK_MAX_CLIENTS"`
	ReadTimeout     time.Duration `env:"NETREK_READ_TIMEOUT"`
	WriteTimeout    time.Duration `env:"NETREK_WRITE_TIMEOUT"`
	UpdateRate      int           `env:"NETREK_UPDATE_RATE"`
	TicksPerState   int           `env:"NETREK_TICKS_PER_STATE"`
	UsePartialState bool          `env:"NETREK_USE_PARTIAL_STATE"`
	WorldSize       float64       `env:"NETREK_WORLD_SIZE"`

	// Circuit Breaker Configuration
	CircuitBreakerMaxRequests         int           `env:"NETREK_CB_MAX_REQUESTS"`
	CircuitBreakerInterval            time.Duration `env:"NETREK_CB_INTERVAL"`
	CircuitBreakerTimeout             time.Duration `env:"NETREK_CB_TIMEOUT"`
	CircuitBreakerMaxConsecutiveFails int           `env:"NETREK_CB_MAX_CONSECUTIVE_FAILS"`

	// Resource Management Configuration
	MaxMemoryMB           int64         `env:"NETREK_MAX_MEMORY_MB"`
	MaxGoroutines         int           `env:"NETREK_MAX_GOROUTINES"`
	ShutdownTimeout       time.Duration `env:"NETREK_SHUTDOWN_TIMEOUT"`
	ResourceCheckInterval time.Duration `env:"NETREK_RESOURCE_CHECK_INTERVAL"`
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s' with value '%v': %s", e.Field, e.Value, e.Message)
}

// LoadConfigFromEnv loads configuration from environment variables with validation
func LoadConfigFromEnv() (*EnvironmentConfig, error) {
	config := &EnvironmentConfig{
		// Secure defaults
		ServerAddr:      getEnvOrDefault("NETREK_SERVER_ADDR", "localhost"),
		ServerPort:      getEnvAsIntOrDefault("NETREK_SERVER_PORT", 4566),
		MaxClients:      getEnvAsIntOrDefault("NETREK_MAX_CLIENTS", 32),
		ReadTimeout:     getEnvAsDurationOrDefault("NETREK_READ_TIMEOUT", 30*time.Second),
		WriteTimeout:    getEnvAsDurationOrDefault("NETREK_WRITE_TIMEOUT", 30*time.Second),
		UpdateRate:      getEnvAsIntOrDefault("NETREK_UPDATE_RATE", 20),
		TicksPerState:   getEnvAsIntOrDefault("NETREK_TICKS_PER_STATE", 3),
		UsePartialState: getEnvAsBoolOrDefault("NETREK_USE_PARTIAL_STATE", true),
		WorldSize:       getEnvAsFloatOrDefault("NETREK_WORLD_SIZE", 10000.0),

		// Circuit Breaker defaults - conservative settings for game stability
		CircuitBreakerMaxRequests:         getEnvAsIntOrDefault("NETREK_CB_MAX_REQUESTS", 3),
		CircuitBreakerInterval:            getEnvAsDurationOrDefault("NETREK_CB_INTERVAL", 60*time.Second),
		CircuitBreakerTimeout:             getEnvAsDurationOrDefault("NETREK_CB_TIMEOUT", 30*time.Second),
		CircuitBreakerMaxConsecutiveFails: getEnvAsIntOrDefault("NETREK_CB_MAX_CONSECUTIVE_FAILS", 5),

		// Resource Management defaults - conservative production settings
		MaxMemoryMB:           getEnvAsInt64OrDefault("NETREK_MAX_MEMORY_MB", 500),
		MaxGoroutines:         getEnvAsIntOrDefault("NETREK_MAX_GOROUTINES", 1000),
		ShutdownTimeout:       getEnvAsDurationOrDefault("NETREK_SHUTDOWN_TIMEOUT", 30*time.Second),
		ResourceCheckInterval: getEnvAsDurationOrDefault("NETREK_RESOURCE_CHECK_INTERVAL", 10*time.Second),
	}

	if err := validateEnvironmentConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// validateEnvironmentConfig validates the environment configuration
func validateEnvironmentConfig(config *EnvironmentConfig) error {
	if err := validateNetworkConfig(config); err != nil {
		return err
	}

	if err := validateTimeoutConfig(config); err != nil {
		return err
	}

	if err := validateGameplayConfig(config); err != nil {
		return err
	}

	if err := validateCircuitBreakerConfig(config); err != nil {
		return err
	}

	if err := validateResourceManagementConfig(config); err != nil {
		return err
	}

	return nil
}

// validateNetworkConfig validates network-related configuration settings
func validateNetworkConfig(config *EnvironmentConfig) error {
	if strings.TrimSpace(config.ServerAddr) == "" {
		return &ValidationError{
			Field:   "ServerAddr",
			Value:   config.ServerAddr,
			Message: "server address cannot be empty",
		}
	}

	if config.ServerPort < 1024 || config.ServerPort > 65535 {
		return &ValidationError{
			Field:   "ServerPort",
			Value:   config.ServerPort,
			Message: "server port must be between 1024 and 65535",
		}
	}

	if config.MaxClients < 1 || config.MaxClients > 1000 {
		return &ValidationError{
			Field:   "MaxClients",
			Value:   config.MaxClients,
			Message: "max clients must be between 1 and 1000",
		}
	}

	return nil
}

// validateTimeoutConfig validates timeout-related configuration settings
func validateTimeoutConfig(config *EnvironmentConfig) error {
	if config.ReadTimeout < time.Second || config.ReadTimeout > time.Minute {
		return &ValidationError{
			Field:   "ReadTimeout",
			Value:   config.ReadTimeout,
			Message: "read timeout must be between 1s and 1m",
		}
	}

	if config.WriteTimeout < time.Second || config.WriteTimeout > time.Minute {
		return &ValidationError{
			Field:   "WriteTimeout",
			Value:   config.WriteTimeout,
			Message: "write timeout must be between 1s and 1m",
		}
	}

	return nil
}

// validateGameplayConfig validates gameplay-related configuration settings
func validateGameplayConfig(config *EnvironmentConfig) error {
	if config.UpdateRate < 1 || config.UpdateRate > 100 {
		return &ValidationError{
			Field:   "UpdateRate",
			Value:   config.UpdateRate,
			Message: "update rate must be between 1 and 100",
		}
	}

	if config.TicksPerState < 1 || config.TicksPerState > 10 {
		return &ValidationError{
			Field:   "TicksPerState",
			Value:   config.TicksPerState,
			Message: "ticks per state must be between 1 and 10",
		}
	}

	if config.WorldSize < 1000.0 || config.WorldSize > 100000.0 {
		return &ValidationError{
			Field:   "WorldSize",
			Value:   config.WorldSize,
			Message: "world size must be between 1000.0 and 100000.0",
		}
	}

	return nil
}

// validateCircuitBreakerConfig validates circuit breaker-related configuration settings
func validateCircuitBreakerConfig(config *EnvironmentConfig) error {
	if config.CircuitBreakerMaxRequests < 1 || config.CircuitBreakerMaxRequests > 100 {
		return &ValidationError{
			Field:   "CircuitBreakerMaxRequests",
			Value:   config.CircuitBreakerMaxRequests,
			Message: "circuit breaker max requests must be between 1 and 100",
		}
	}

	if config.CircuitBreakerInterval < 1*time.Second || config.CircuitBreakerInterval > 300*time.Second {
		return &ValidationError{
			Field:   "CircuitBreakerInterval",
			Value:   config.CircuitBreakerInterval,
			Message: "circuit breaker interval must be between 1s and 300s",
		}
	}

	if config.CircuitBreakerTimeout < 1*time.Second || config.CircuitBreakerTimeout > 120*time.Second {
		return &ValidationError{
			Field:   "CircuitBreakerTimeout",
			Value:   config.CircuitBreakerTimeout,
			Message: "circuit breaker timeout must be between 1s and 120s",
		}
	}

	if config.CircuitBreakerMaxConsecutiveFails < 1 || config.CircuitBreakerMaxConsecutiveFails > 50 {
		return &ValidationError{
			Field:   "CircuitBreakerMaxConsecutiveFails",
			Value:   config.CircuitBreakerMaxConsecutiveFails,
			Message: "circuit breaker max consecutive fails must be between 1 and 50",
		}
	}

	return nil
}

// validateResourceManagementConfig validates resource management-related configuration settings
func validateResourceManagementConfig(config *EnvironmentConfig) error {
	if config.MaxMemoryMB < 128 || config.MaxMemoryMB > 8192 {
		return &ValidationError{
			Field:   "MaxMemoryMB",
			Value:   config.MaxMemoryMB,
			Message: "max memory (MB) must be between 128 and 8192",
		}
	}

	if config.MaxGoroutines < 10 || config.MaxGoroutines > 1000 {
		return &ValidationError{
			Field:   "MaxGoroutines",
			Value:   config.MaxGoroutines,
			Message: "max goroutines must be between 10 and 1000",
		}
	}

	if config.ShutdownTimeout < 1*time.Second || config.ShutdownTimeout > 60*time.Second {
		return &ValidationError{
			Field:   "ShutdownTimeout",
			Value:   config.ShutdownTimeout,
			Message: "shutdown timeout must be between 1s and 60s",
		}
	}

	if config.ResourceCheckInterval < 1*time.Second || config.ResourceCheckInterval > 60*time.Second {
		return &ValidationError{
			Field:   "ResourceCheckInterval",
			Value:   config.ResourceCheckInterval,
			Message: "resource check interval must be between 1s and 60s",
		}
	}

	return nil
}

// getEnvOrDefault returns the environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsIntOrDefault returns the environment variable as int or default if not set or invalid
func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsInt64OrDefault returns the environment variable as int64 or default if not set or invalid
func getEnvAsInt64OrDefault(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if int64Value, err := strconv.ParseInt(value, 10, 64); err == nil {
			return int64Value
		}
	}
	return defaultValue
}

// getEnvAsBoolOrDefault returns the environment variable as bool or default if not set or invalid
func getEnvAsBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getEnvAsFloatOrDefault returns the environment variable as float64 or default if not set or invalid
func getEnvAsFloatOrDefault(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

// getEnvAsDurationOrDefault returns the environment variable as time.Duration or default if not set or invalid
func getEnvAsDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// ApplyEnvironmentOverrides applies environment variable overrides to existing GameConfig
func ApplyEnvironmentOverrides(gameConfig *GameConfig) error {
	envConfig, err := LoadConfigFromEnv()
	if err != nil {
		return fmt.Errorf("failed to load environment configuration: %w", err)
	}

	// Apply environment overrides to NetworkConfig
	gameConfig.NetworkConfig.ServerAddress = fmt.Sprintf("%s:%d", envConfig.ServerAddr, envConfig.ServerPort)
	gameConfig.NetworkConfig.ServerPort = envConfig.ServerPort
	gameConfig.NetworkConfig.UpdateRate = envConfig.UpdateRate
	gameConfig.NetworkConfig.TicksPerState = envConfig.TicksPerState
	gameConfig.NetworkConfig.UsePartialState = envConfig.UsePartialState

	// Apply environment overrides to other configs
	gameConfig.MaxPlayers = envConfig.MaxClients
	gameConfig.WorldSize = envConfig.WorldSize

	return nil
}

// ShipTypeConfig allows defining custom ship types and stats in config
// Keyed by name (e.g. "Scout", "Destroyer")
type ShipTypeConfig struct {
	Name         string  `json:"name"`
	MaxHull      int     `json:"maxHull"`
	MaxShields   int     `json:"maxShields"`
	MaxFuel      int     `json:"maxFuel"`
	Acceleration float64 `json:"acceleration"`
	TurnRate     float64 `json:"turnRate"`
	MaxSpeed     float64 `json:"maxSpeed"`
	WeaponSlots  int     `json:"weaponSlots"`
	MaxArmies    int     `json:"maxArmies"`
}

// GameConfig contains configuration for a Netrek game
type GameConfig struct {
	WorldSize     float64                   `json:"worldSize"`
	MaxPlayers    int                       `json:"maxPlayers"`
	Teams         []TeamConfig              `json:"teams"`
	Planets       []PlanetConfig            `json:"planets"`
	PhysicsConfig PhysicsConfig             `json:"physics"`
	NetworkConfig NetworkConfig             `json:"network"`
	GameRules     GameRules                 `json:"gameRules"`
	ShipTypes     map[string]ShipTypeConfig `json:"shipTypes"`
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
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config GameConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply custom ship configurations to the entity system
	if len(config.ShipTypes) > 0 {
		shipStats := make(map[string]entity.ShipStats)
		for name, shipConfig := range config.ShipTypes {
			shipStats[name] = entity.ShipStats{
				MaxHull:      shipConfig.MaxHull,
				MaxShields:   shipConfig.MaxShields,
				MaxFuel:      shipConfig.MaxFuel,
				Acceleration: shipConfig.Acceleration,
				TurnRate:     shipConfig.TurnRate,
				MaxSpeed:     shipConfig.MaxSpeed,
				WeaponSlots:  shipConfig.WeaponSlots,
				MaxArmies:    shipConfig.MaxArmies,
			}
		}
		entity.SetShipTypeStats(shipStats)
	}

	return &config, nil
}

// SaveConfig saves a configuration to a file
func SaveConfig(config *GameConfig, path string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// DefaultConfig returns a default game configuration
func DefaultConfig() *GameConfig {
	return &GameConfig{
		WorldSize:     10000,
		MaxPlayers:    16,
		Teams:         createDefaultTeams(),
		Planets:       createDefaultPlanets(),
		PhysicsConfig: createDefaultPhysicsConfig(),
		NetworkConfig: createDefaultNetworkConfig(),
		GameRules:     createDefaultGameRules(),
		ShipTypes:     createDefaultShipTypes(),
	}
}

// createDefaultTeams creates the default team configurations for a new game.
func createDefaultTeams() []TeamConfig {
	return []TeamConfig{
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
	}
}

// createDefaultPlanets creates the default planet configurations for a new game.
func createDefaultPlanets() []PlanetConfig {
	return []PlanetConfig{
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
	}
}

// createDefaultPhysicsConfig creates the default physics configuration for a new game.
func createDefaultPhysicsConfig() PhysicsConfig {
	return PhysicsConfig{
		Gravity:         0,
		Friction:        0.1,
		CollisionDamage: 20,
	}
}

// createDefaultNetworkConfig creates the default network configuration for a new game.
// Note: This function now provides base defaults that will be overridden by environment variables
func createDefaultNetworkConfig() NetworkConfig {
	return NetworkConfig{
		UpdateRate:      20,
		TicksPerState:   3,
		UsePartialState: true,
		ServerPort:      4566,
		ServerAddress:   "", // Will be set from environment or secure default
	}
}

// createDefaultGameRules creates the default game rules configuration for a new game.
func createDefaultGameRules() GameRules {
	return GameRules{
		WinCondition:   "conquest",
		TimeLimit:      1800,
		MaxScore:       100,
		RespawnDelay:   5,
		FriendlyFire:   false,
		StartingArmies: 0,
	}
}

// createDefaultShipTypes creates the default ship type configurations for a new game.
func createDefaultShipTypes() map[string]ShipTypeConfig {
	return map[string]ShipTypeConfig{
		"Scout": {
			Name:         "Scout",
			MaxHull:      100,
			MaxShields:   100,
			MaxFuel:      1000,
			Acceleration: 200,
			TurnRate:     3.0,
			MaxSpeed:     300,
			WeaponSlots:  2,
			MaxArmies:    2,
		},
		"Destroyer": {
			Name:         "Destroyer",
			MaxHull:      150,
			MaxShields:   150,
			MaxFuel:      1200,
			Acceleration: 150,
			TurnRate:     2.5,
			MaxSpeed:     250,
			WeaponSlots:  3,
			MaxArmies:    5,
		},
	}
}

// GalaxyTemplate represents a preset galaxy configuration
type GalaxyTemplate struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	WorldSize   float64        `json:"worldSize"`
	Teams       []TeamConfig   `json:"teams"`
	Planets     []PlanetConfig `json:"planets"`
}

// galaxyTemplates contains built-in galaxy map templates
var galaxyTemplates = map[string]GalaxyTemplate{
	"classic_netrek": {
		Name:        "Classic Netrek",
		Description: "Traditional 2-team Netrek galaxy with 4 home worlds",
		WorldSize:   10000,
		Teams: []TeamConfig{
			{Name: "Federation", Color: "#0000FF", MaxShips: 8, StartingShip: "Scout"},
			{Name: "Romulans", Color: "#00FF00", MaxShips: 8, StartingShip: "Scout"},
			{Name: "Klingons", Color: "#FF0000", MaxShips: 8, StartingShip: "Scout"},
			{Name: "Orions", Color: "#FFFF00", MaxShips: 8, StartingShip: "Scout"},
		},
		Planets: []PlanetConfig{
			{Name: "Earth", X: -4000, Y: -4000, Type: entity.Homeworld, HomeWorld: true, TeamID: 0, InitialArmies: 30},
			{Name: "Romulus", X: 4000, Y: -4000, Type: entity.Homeworld, HomeWorld: true, TeamID: 1, InitialArmies: 30},
			{Name: "Qo'noS", X: -4000, Y: 4000, Type: entity.Homeworld, HomeWorld: true, TeamID: 2, InitialArmies: 30},
			{Name: "Orion", X: 4000, Y: 4000, Type: entity.Homeworld, HomeWorld: true, TeamID: 3, InitialArmies: 30},
			{Name: "Centauri", X: 0, Y: 0, Type: entity.Industrial, HomeWorld: false, TeamID: -1, InitialArmies: 15},
			{Name: "Alpha", X: -2000, Y: 0, Type: entity.Military, HomeWorld: false, TeamID: -1, InitialArmies: 20},
			{Name: "Beta", X: 2000, Y: 0, Type: entity.Agricultural, HomeWorld: false, TeamID: -1, InitialArmies: 10},
			{Name: "Gamma", X: 0, Y: -2000, Type: entity.Industrial, HomeWorld: false, TeamID: -1, InitialArmies: 15},
			{Name: "Delta", X: 0, Y: 2000, Type: entity.Agricultural, HomeWorld: false, TeamID: -1, InitialArmies: 10},
		},
	},
	"small_galaxy": {
		Name:        "Small Galaxy",
		Description: "Compact 2-team galaxy for quick games",
		WorldSize:   6000,
		Teams: []TeamConfig{
			{Name: "Federation", Color: "#0000FF", MaxShips: 4, StartingShip: "Scout"},
			{Name: "Klingons", Color: "#FF0000", MaxShips: 4, StartingShip: "Scout"},
		},
		Planets: []PlanetConfig{
			{Name: "Earth", X: -2000, Y: 0, Type: entity.Homeworld, HomeWorld: true, TeamID: 0, InitialArmies: 25},
			{Name: "Qo'noS", X: 2000, Y: 0, Type: entity.Homeworld, HomeWorld: true, TeamID: 1, InitialArmies: 25},
			{Name: "Neutral Station", X: 0, Y: 1000, Type: entity.Industrial, HomeWorld: false, TeamID: -1, InitialArmies: 15},
			{Name: "Mining Colony", X: 0, Y: -1000, Type: entity.Agricultural, HomeWorld: false, TeamID: -1, InitialArmies: 10},
		},
	},
	"balanced_4team": {
		Name:        "Balanced 4-Team",
		Description: "Perfectly balanced 4-team galaxy with symmetric layout",
		WorldSize:   12000,
		Teams: []TeamConfig{
			{Name: "Federation", Color: "#0000FF", MaxShips: 6, StartingShip: "Scout"},
			{Name: "Romulans", Color: "#00FF00", MaxShips: 6, StartingShip: "Scout"},
			{Name: "Klingons", Color: "#FF0000", MaxShips: 6, StartingShip: "Scout"},
			{Name: "Cardassians", Color: "#FF8800", MaxShips: 6, StartingShip: "Scout"},
		},
		Planets: []PlanetConfig{
			// Home worlds in corners
			{Name: "Earth", X: -3000, Y: -3000, Type: entity.Homeworld, HomeWorld: true, TeamID: 0, InitialArmies: 30},
			{Name: "Romulus", X: 3000, Y: -3000, Type: entity.Homeworld, HomeWorld: true, TeamID: 1, InitialArmies: 30},
			{Name: "Qo'noS", X: -3000, Y: 3000, Type: entity.Homeworld, HomeWorld: true, TeamID: 2, InitialArmies: 30},
			{Name: "Cardassia", X: 3000, Y: 3000, Type: entity.Homeworld, HomeWorld: true, TeamID: 3, InitialArmies: 30},
			// Strategic neutral planets
			{Name: "Center Core", X: 0, Y: 0, Type: entity.Industrial, HomeWorld: false, TeamID: -1, InitialArmies: 25},
			{Name: "North Gate", X: 0, Y: -1500, Type: entity.Military, HomeWorld: false, TeamID: -1, InitialArmies: 20},
			{Name: "South Gate", X: 0, Y: 1500, Type: entity.Military, HomeWorld: false, TeamID: -1, InitialArmies: 20},
			{Name: "East Gate", X: 1500, Y: 0, Type: entity.Military, HomeWorld: false, TeamID: -1, InitialArmies: 20},
			{Name: "West Gate", X: -1500, Y: 0, Type: entity.Military, HomeWorld: false, TeamID: -1, InitialArmies: 20},
		},
	},
}

// GetGalaxyTemplate returns a galaxy template by name
func GetGalaxyTemplate(name string) *GalaxyTemplate {
	if template, ok := galaxyTemplates[name]; ok {
		// Return a copy to prevent modification of the original
		templateCopy := template
		return &templateCopy
	}
	return nil
}

// ListGalaxyTemplates returns a list of available galaxy template names and descriptions
func ListGalaxyTemplates() map[string]string {
	templates := make(map[string]string)
	for name, template := range galaxyTemplates {
		templates[name] = template.Description
	}
	return templates
}

// ApplyGalaxyTemplate applies a galaxy template to a GameConfig
func ApplyGalaxyTemplate(config *GameConfig, templateName string) error {
	template := GetGalaxyTemplate(templateName)
	if template == nil {
		return fmt.Errorf("galaxy template '%s' not found", templateName)
	}

	// Apply template settings to config
	config.WorldSize = template.WorldSize
	config.Teams = make([]TeamConfig, len(template.Teams))
	copy(config.Teams, template.Teams)
	config.Planets = make([]PlanetConfig, len(template.Planets))
	copy(config.Planets, template.Planets)

	return nil
}

// LoadConfigWithTemplate loads a base config and applies a galaxy template
func LoadConfigWithTemplate(configPath, templateName string) (*GameConfig, error) {
	// Load base config
	config, err := LoadConfig(configPath)
	if err != nil {
		// If base config fails, start with default
		config = DefaultConfig()
	}

	// Apply template
	if err := ApplyGalaxyTemplate(config, templateName); err != nil {
		return nil, fmt.Errorf("failed to apply galaxy template: %w", err)
	}

	return config, nil
}
