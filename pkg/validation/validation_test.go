package validation

import (
	"strings"
	"testing"
	"time"
)

func TestValidatePlayerName(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		want        string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid simple name",
			input:   "Player1",
			want:    "Player1",
			wantErr: false,
		},
		{
			name:    "valid name with spaces",
			input:   "Player One",
			want:    "Player One",
			wantErr: false,
		},
		{
			name:    "valid name with hyphen",
			input:   "Player-One",
			want:    "Player-One",
			wantErr: false,
		},
		{
			name:    "valid name with underscore",
			input:   "Player_One",
			want:    "Player_One",
			wantErr: false,
		},
		{
			name:    "name with leading/trailing spaces",
			input:   "  Player1  ",
			want:    "Player1",
			wantErr: false,
		},
		{
			name:        "empty name",
			input:       "",
			want:        "",
			wantErr:     true,
			errContains: "cannot be empty",
		},
		{
			name:        "only whitespace",
			input:       "   ",
			want:        "",
			wantErr:     true,
			errContains: "cannot be only whitespace",
		},
		{
			name:        "too long name",
			input:       strings.Repeat("a", MaxPlayerNameLen+1),
			want:        "",
			wantErr:     true,
			errContains: "too long",
		},
		{
			name:        "name with special characters",
			input:       "Player@#$",
			want:        "",
			wantErr:     true,
			errContains: "invalid characters",
		},
		{
			name:        "name with control character",
			input:       "Player\x00One",
			want:        "",
			wantErr:     true,
			errContains: "control characters",
		},
		{
			name:    "HTML entities should be escaped",
			input:   "Player<script>",
			want:    "Player&lt;script&gt;",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidatePlayerName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePlayerName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("ValidatePlayerName() error = %v, should contain %q", err, tt.errContains)
			}
			if got != tt.want {
				t.Errorf("ValidatePlayerName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateChatMessage(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		want        string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid message",
			input:   "Hello world!",
			want:    "Hello world!",
			wantErr: false,
		},
		{
			name:    "message with newlines",
			input:   "Line 1\nLine 2",
			want:    "Line 1\nLine 2",
			wantErr: false,
		},
		{
			name:    "message with tabs",
			input:   "Col1\tCol2",
			want:    "Col1\tCol2",
			wantErr: false,
		},
		{
			name:        "empty message",
			input:       "",
			want:        "",
			wantErr:     true,
			errContains: "cannot be empty",
		},
		{
			name:        "only whitespace",
			input:       "   ",
			want:        "",
			wantErr:     true,
			errContains: "cannot be only whitespace",
		},
		{
			name:        "too long message",
			input:       strings.Repeat("a", MaxChatMessageLen+1),
			want:        "",
			wantErr:     true,
			errContains: "too long",
		},
		{
			name:    "control characters removed (except newline/tab)",
			input:   "Hello\x00World\nNext\tLine",
			want:    "HelloWorld\nNext\tLine",
			wantErr: false,
		},
		{
			name:    "HTML entities escaped",
			input:   "<script>alert('xss')</script>",
			want:    "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateChatMessage(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateChatMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("ValidateChatMessage() error = %v, should contain %q", err, tt.errContains)
			}
			if got != tt.want {
				t.Errorf("ValidateChatMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateTeamID(t *testing.T) {
	tests := []struct {
		name    string
		input   int
		wantErr bool
	}{
		{"valid team 0", 0, false},
		{"valid team 1", 1, false},
		{"invalid negative team", -1, true},
		{"invalid team 2", 2, true},
		{"invalid large team", 999, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTeamID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTeamID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateBeamAmount(t *testing.T) {
	tests := []struct {
		name    string
		input   int
		wantErr bool
	}{
		{"valid zero", 0, false},
		{"valid amount", 100, false},
		{"valid max", 1000, false},
		{"invalid negative", -1, true},
		{"invalid too large", 1001, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBeamAmount(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBeamAmount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateWeaponIndex(t *testing.T) {
	tests := []struct {
		name    string
		input   int
		wantErr bool
	}{
		{"not firing", -1, false},
		{"weapon 0", 0, false},
		{"weapon 7", 7, false},
		{"invalid -2", -2, true},
		{"invalid 8", 8, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWeaponIndex(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateWeaponIndex() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMessageValidator_ValidateMessage(t *testing.T) {
	validator := NewMessageValidator()
	defer validator.Close()

	tests := []struct {
		name        string
		data        []byte
		clientID    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid JSON message",
			data:     []byte(`{"type":"test","data":"value"}`),
			clientID: "client1",
			wantErr:  false,
		},
		{
			name:        "too large message",
			data:        make([]byte, MaxMessageSize+1),
			clientID:    "client1",
			wantErr:     true,
			errContains: "too large",
		},
		{
			name:        "invalid JSON",
			data:        []byte(`{"invalid": json`),
			clientID:    "client1",
			wantErr:     true,
			errContains: "invalid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateMessage(tt.data, tt.clientID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("ValidateMessage() error = %v, should contain %q", err, tt.errContains)
			}
		})
	}
}

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(5, time.Minute) // 5 requests per minute
	defer rl.Close()

	clientID := "test-client"

	// Should allow first 5 requests
	for i := 0; i < 5; i++ {
		if !rl.Allow(clientID) {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 6th request should be denied
	if rl.Allow(clientID) {
		t.Error("6th request should be denied")
	}

	// Different client should still be allowed
	if !rl.Allow("other-client") {
		t.Error("Different client should be allowed")
	}
}

func TestRateLimiter_TokenRefill(t *testing.T) {
	// Use a shorter window for testing
	rl := NewRateLimiter(2, 100*time.Millisecond)
	defer rl.Close()

	clientID := "test-client"

	// Consume all tokens
	rl.Allow(clientID)
	rl.Allow(clientID)

	// Should be denied
	if rl.Allow(clientID) {
		t.Error("Request should be denied after consuming all tokens")
	}

	// Wait for refill period
	time.Sleep(150 * time.Millisecond)

	// Should be allowed again after refill
	if !rl.Allow(clientID) {
		t.Error("Request should be allowed after token refill")
	}
}
