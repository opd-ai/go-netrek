// Package validation provides input validation and sanitization for network messages.
package validation

import (
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// Message size and content limits as defined in the roadmap
const (
	MaxMessageSize    = 64 * 1024 // 64KB max message
	MaxPlayerNameLen  = 32
	MaxChatMessageLen = 256
	MaxMessagesPerMin = 100
)

// Regular expressions for input validation
var (
	// Allow alphanumeric, spaces, hyphens, underscores, and basic punctuation for player names
	// This prevents most special characters that could cause issues while allowing reasonable names
	validPlayerNameChars = regexp.MustCompile(`^[a-zA-Z0-9\s\-_.<>()]+$`)
)

// MessageValidator provides validation for different message types
type MessageValidator struct {
	rateLimiter *RateLimiter
}

// NewMessageValidator creates a new message validator with rate limiting
func NewMessageValidator() *MessageValidator {
	return &MessageValidator{
		rateLimiter: NewRateLimiter(MaxMessagesPerMin, time.Minute),
	}
}

// Close releases resources used by the message validator
func (v *MessageValidator) Close() {
	if v.rateLimiter != nil {
		v.rateLimiter.Close()
	}
}

// ValidateMessage validates a raw message against size and format constraints
func (v *MessageValidator) ValidateMessage(data []byte, clientID string) error {
	// Check message size
	if len(data) > MaxMessageSize {
		return fmt.Errorf("message too large: %d bytes (max %d)", len(data), MaxMessageSize)
	}

	// Check if message is valid JSON
	if !json.Valid(data) {
		return fmt.Errorf("invalid JSON format")
	}

	// Check rate limiting
	if !v.rateLimiter.Allow(clientID) {
		return fmt.Errorf("rate limit exceeded: max %d messages per minute", MaxMessagesPerMin)
	}

	return nil
}

// ValidatePlayerName validates and sanitizes a player name
func ValidatePlayerName(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("player name cannot be empty")
	}

	// Check length
	if len(name) > MaxPlayerNameLen {
		return "", fmt.Errorf("player name too long: %d characters (max %d)", len(name), MaxPlayerNameLen)
	}

	// Check UTF-8 validity
	if !utf8.ValidString(name) {
		return "", fmt.Errorf("player name contains invalid UTF-8 characters")
	}

	// Trim whitespace
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", fmt.Errorf("player name cannot be only whitespace")
	}

	// Check for control characters first
	for _, r := range trimmed {
		if unicode.IsControl(r) {
			return "", fmt.Errorf("player name contains control characters")
		}
	}

	// Check for allowed character set
	if !validPlayerNameChars.MatchString(trimmed) {
		return "", fmt.Errorf("player name contains invalid characters (only alphanumeric, spaces, hyphens, underscores, and basic punctuation allowed)")
	}

	// Escape HTML to prevent XSS
	sanitized := html.EscapeString(trimmed)

	return sanitized, nil
}

// ValidateChatMessage validates and sanitizes a chat message
func ValidateChatMessage(message string) (string, error) {
	if message == "" {
		return "", fmt.Errorf("chat message cannot be empty")
	}

	// Check length
	if len(message) > MaxChatMessageLen {
		return "", fmt.Errorf("chat message too long: %d characters (max %d)", len(message), MaxChatMessageLen)
	}

	// Check UTF-8 validity
	if !utf8.ValidString(message) {
		return "", fmt.Errorf("chat message contains invalid UTF-8 characters")
	}

	// Trim whitespace
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return "", fmt.Errorf("chat message cannot be only whitespace")
	}

	// Remove control characters except newlines and tabs
	filtered := strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			return -1 // Remove character
		}
		return r
	}, trimmed)

	// Escape HTML to prevent XSS
	sanitized := html.EscapeString(filtered)

	return sanitized, nil
}

// ValidateTeamID validates a team ID
func ValidateTeamID(teamID int) error {
	// Team IDs should be 0 or 1 for Federation and Klingon teams
	if teamID < 0 || teamID > 1 {
		return fmt.Errorf("invalid team ID: %d (must be 0 or 1)", teamID)
	}
	return nil
}

// ValidateBeamAmount validates beam down/up amounts
func ValidateBeamAmount(amount int) error {
	if amount < 0 {
		return fmt.Errorf("beam amount cannot be negative: %d", amount)
	}
	if amount > 1000 { // Reasonable upper limit
		return fmt.Errorf("beam amount too large: %d (max 1000)", amount)
	}
	return nil
}

// ValidateWeaponIndex validates weapon firing index
func ValidateWeaponIndex(index int) error {
	if index < -1 || index >= 8 { // -1 means not firing, 0-7 are weapon slots
		return fmt.Errorf("invalid weapon index: %d (must be -1 or 0-7)", index)
	}
	return nil
}
