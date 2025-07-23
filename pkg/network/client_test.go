package network

import (
	"bytes"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/opd-ai/go-netrek/pkg/entity"
	"github.com/opd-ai/go-netrek/pkg/event"
)

// mockConn implements net.Conn for testing
// It allows us to simulate server/client communication
// Only implements methods needed for tests

type mockConn struct {
	readBuf  *bytes.Buffer
	writeBuf *bytes.Buffer
	closed   bool
	mu       sync.Mutex
}

func newMockConn() *mockConn {
	return &mockConn{
		readBuf:  &bytes.Buffer{},
		writeBuf: &bytes.Buffer{},
	}
}

func (m *mockConn) Read(b []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return 0, io.EOF
	}
	return m.readBuf.Read(b)
}

func (m *mockConn) Write(b []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return 0, io.ErrClosedPipe
	}
	return m.writeBuf.Write(b)
}

func (m *mockConn) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}
func (m *mockConn) LocalAddr() net.Addr                { return nil }
func (m *mockConn) RemoteAddr() net.Addr               { return nil }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

// TestNewGameClient_BasicInitialization tests NewGameClient returns a valid client
func TestNewGameClient_BasicInitialization(t *testing.T) {
	eb := event.NewEventBus()
	c := NewGameClient(eb)
	if c == nil {
		t.Fatal("expected non-nil client")
	}
	if c.eventBus != eb {
		t.Error("eventBus not set correctly")
	}
	if c.pingInterval != 5*time.Second {
		t.Error("default pingInterval incorrect")
	}
	if c.maxReconnectAttempts != 5 {
		t.Error("default maxReconnectAttempts incorrect")
	}
}

// TestRequestShipClass_UpdatesDesiredShipClass
func TestRequestShipClass_UpdatesDesiredShipClass(t *testing.T) {
	c := NewGameClient(event.NewEventBus())
	mc := newMockConn()
	c.conn = mc
	c.connected = true
	class := entity.Scout
	err := c.RequestShipClass(class)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.DesiredShipClass != class {
		t.Errorf("expected DesiredShipClass %v, got %v", class, c.DesiredShipClass)
	}
}

// TestConnect_ErrorCases tests Connect error paths (simulate dial error by using invalid address)
func TestConnect_ErrorCases(t *testing.T) {
	c := NewGameClient(event.NewEventBus())
	c.conn = nil
	c.connected = false
	c.serverAddress = "bad:address"
	// This will fail because address is invalid
	err := c.Connect("bad:address", "", 0)
	if err == nil {
		t.Error("expected error on dial failure")
	}
}

// TestSendInput_NotConnected returns error
func TestSendInput_NotConnected(t *testing.T) {
	c := NewGameClient(event.NewEventBus())
	c.connected = false
	err := c.SendInput(true, false, false, 1, false, false, 0, 0)
	if err == nil {
		t.Error("expected error when not connected")
	}
}

// TestSendChatMessage_NotConnected returns error
func TestSendChatMessage_NotConnected(t *testing.T) {
	c := NewGameClient(event.NewEventBus())
	c.connected = false
	err := c.SendChatMessage("hi")
	if err == nil {
		t.Error("expected error when not connected")
	}
}

// Table-driven test for SendInput
func TestSendInput_TableDriven(t *testing.T) {
	c := NewGameClient(event.NewEventBus())
	mc := newMockConn()
	c.conn = mc
	c.connected = true
	cases := []struct {
		name      string
		thrust    bool
		turnLeft  bool
		turnRight bool
		fire      int
		beamDown  bool
		beamUp    bool
		beamAmt   int
		targetID  entity.ID
	}{
		{"all false", false, false, false, -1, false, false, 0, 0},
		{"fire weapon", true, false, false, 1, false, false, 0, 0},
		{"beam up", false, false, false, -1, false, true, 5, 42},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := c.SendInput(tc.thrust, tc.turnLeft, tc.turnRight, tc.fire, tc.beamDown, tc.beamUp, tc.beamAmt, tc.targetID)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestGetLatency returns correct value
func TestGetLatency(t *testing.T) {
	c := NewGameClient(event.NewEventBus())
	c.latency = 123 * time.Millisecond
	if c.GetLatency() != 123*time.Millisecond {
		t.Errorf("expected 123ms, got %v", c.GetLatency())
	}
}

// TestGetGameStateChannel returns the channel
func TestGetGameStateChannel(t *testing.T) {
	c := NewGameClient(event.NewEventBus())
	ch := c.GetGameStateChannel()
	if ch == nil {
		t.Error("expected non-nil channel")
	}
}
