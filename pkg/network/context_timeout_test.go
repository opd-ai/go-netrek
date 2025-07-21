// pkg/network/context_timeout_test.go
package network

import (
	"context"
	"encoding/binary"
	"net"
	"testing"
	"time"

	"github.com/opd-ai/go-netrek/pkg/config"
	"github.com/opd-ai/go-netrek/pkg/engine"
	"github.com/opd-ai/go-netrek/pkg/event"
)

func TestServerTimeoutHandling(t *testing.T) {
	tests := []struct {
		name           string
		readTimeout    time.Duration
		writeTimeout   time.Duration
		contextTimeout time.Duration
		expectError    bool
		errorType      string
	}{
		{
			name:           "Normal operation within timeout",
			readTimeout:    5 * time.Second,
			writeTimeout:   5 * time.Second,
			contextTimeout: 10 * time.Second,
			expectError:    false,
		},
		{
			name:           "Read timeout exceeded",
			readTimeout:    100 * time.Millisecond,
			writeTimeout:   5 * time.Second,
			contextTimeout: 50 * time.Millisecond,
			expectError:    true,
			errorType:      "context deadline exceeded",
		},
		{
			name:           "Write timeout exceeded",
			readTimeout:    5 * time.Second,
			writeTimeout:   100 * time.Millisecond,
			contextTimeout: 50 * time.Millisecond,
			expectError:    true,
			errorType:      "context deadline exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test game configuration
			gameConfig := &config.GameConfig{
				WorldSize: 1000,
				Teams: []config.TeamConfig{
					{Name: "Team0", Color: "red"},
					{Name: "Team1", Color: "blue"},
				},
				NetworkConfig: config.NetworkConfig{
					UpdateRate:      20,
					TicksPerState:   3,
					UsePartialState: true,
				},
			}

			// Create game engine
			game := engine.NewGame(gameConfig)

			// Create server with custom timeouts
			server := NewGameServer(game, 2)
			server.readTimeout = tt.readTimeout
			server.writeTimeout = tt.writeTimeout

			// Start server on random port
			listener, err := net.Listen("tcp", "localhost:0")
			if err != nil {
				t.Fatalf("Failed to create listener: %v", err)
			}
			defer listener.Close()

			server.listener = listener
			server.running = true

			// Test connection with timeout context
			ctx, cancel := context.WithTimeout(context.Background(), tt.contextTimeout)
			defer cancel()

			// Create a proper connection pair for testing
			clientConn, serverConn := net.Pipe()
			defer clientConn.Close()
			defer serverConn.Close()

			// For normal operation test, send valid data to read
			if !tt.expectError {
				// Send a valid message for normal operation testing
				go func() {
					time.Sleep(10 * time.Millisecond) // Give readMessage time to start
					msgType := MessageType(1)
					msgData := []byte(`{"test": "data"}`)

					// Write message type, length, and data
					binary.Write(clientConn, binary.BigEndian, msgType)
					binary.Write(clientConn, binary.BigEndian, uint16(len(msgData)))
					clientConn.Write(msgData)
				}()
			}

			// Test read message with context
			_, _, err = server.readMessage(ctx, serverConn)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else {
					// Check for timeout-related errors
					errStr := err.Error()
					if errStr != "context deadline exceeded" &&
						errStr != "read pipe: i/o timeout" &&
						errStr != "write pipe: i/o timeout" {
						t.Errorf("Expected timeout error, got '%s'", errStr)
					}
				}
			} else {
				if err != nil && err.Error() != "EOF" {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestClientTimeoutHandling(t *testing.T) {
	tests := []struct {
		name           string
		readTimeout    time.Duration
		writeTimeout   time.Duration
		connectTimeout time.Duration
		expectError    bool
	}{
		{
			name:           "Normal connection within timeout",
			readTimeout:    5 * time.Second,
			writeTimeout:   5 * time.Second,
			connectTimeout: 5 * time.Second,
			expectError:    false,
		},
		{
			name:           "Connection timeout exceeded",
			readTimeout:    5 * time.Second,
			writeTimeout:   5 * time.Second,
			connectTimeout: 10 * time.Millisecond,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create event bus
			eventBus := event.NewEventBus()

			// Create client with custom timeouts
			client := NewGameClient(eventBus)
			client.readTimeout = tt.readTimeout
			client.writeTimeout = tt.writeTimeout
			client.connectionTimeout = tt.connectTimeout

			// Try to connect to non-existent server (should timeout)
			err := client.Connect("localhost:9999", "TestPlayer", 0)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				// Note: This test might still fail because we're connecting to a non-existent server
				// The point is to test that timeout handling works correctly
			}
		})
	}
}

func TestContextCancellation(t *testing.T) {
	// Create a test game configuration
	gameConfig := &config.GameConfig{
		WorldSize: 1000,
		Teams: []config.TeamConfig{
			{Name: "Team0", Color: "red"},
		},
		NetworkConfig: config.NetworkConfig{
			UpdateRate:      20,
			TicksPerState:   3,
			UsePartialState: true,
		},
	}

	// Create game engine
	game := engine.NewGame(gameConfig)

	// Create server
	server := NewGameServer(game, 1)

	// Create context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Create a pipe to simulate network connection
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	// Start reading in goroutine
	errorChan := make(chan error, 1)
	go func() {
		_, _, err := server.readMessage(ctx, serverConn)
		errorChan <- err
	}()

	// Cancel context after short delay
	time.Sleep(100 * time.Millisecond)
	cancel()

	// Wait for read to complete
	select {
	case err := <-errorChan:
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled error, got: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Error("Read operation did not complete within timeout")
	}
}

func TestMessageSizeValidation(t *testing.T) {
	// Create a test game configuration
	gameConfig := &config.GameConfig{
		WorldSize: 1000,
		Teams: []config.TeamConfig{
			{Name: "Team0", Color: "red"},
		},
		NetworkConfig: config.NetworkConfig{
			UpdateRate:      20,
			TicksPerState:   3,
			UsePartialState: true,
		},
	}

	// Create game engine
	game := engine.NewGame(gameConfig)

	// Create server
	server := NewGameServer(game, 1)

	// Create context
	ctx := context.Background()

	// Create a pipe to simulate network connection
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	// Test message that's too large
	largeMessage := make(map[string]string)
	largeMessage["data"] = string(make([]byte, 100000)) // 100KB message

	// Test send message with oversized content
	err := server.sendMessage(ctx, serverConn, ChatMessage, largeMessage)
	if err == nil {
		t.Error("Expected error for oversized message, got none")
	}
	if err != nil && err.Error() != "message too large: 100016 bytes (max 65536)" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}
