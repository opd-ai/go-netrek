package network

import (
	"testing"

	"github.com/opd-ai/go-netrek/pkg/config"
	"github.com/opd-ai/go-netrek/pkg/engine"
)

func TestNewGameServer_ConfiguresPartialStateFromConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.NetworkConfig.UsePartialState = false
	cfg.NetworkConfig.TicksPerState = 7
	cfg.NetworkConfig.UpdateRate = 10
	game := engine.NewGame(cfg)
	server := NewGameServer(game, 8)
	if server.partialState != false {
		t.Errorf("expected partialState false, got %v", server.partialState)
	}
	if server.ticksPerState != 7 {
		t.Errorf("expected ticksPerState 7, got %d", server.ticksPerState)
	}
	if server.updateRate != (1e9 / 10) {
		t.Errorf("expected updateRate 1e8ns, got %v", server.updateRate)
	}
}
