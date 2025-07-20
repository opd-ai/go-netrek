# Go Netrek Engo GUI Client

This document describes how to use the new Engo-based graphical client for go-netrek.

## Prerequisites

Make sure you have Go 1.19+ and the following dependencies installed:
- OpenGL drivers for your system
- GLFW library

## Building

Build both the server and client:

```bash
cd /home/user/go/src/github.com/opd-ai/go-netrek
go build ./cmd/server
go build ./cmd/client
```

## Running

### 1. Start the Server

First, start the game server:

```bash
./server
```

The server will start on the default port (8080).

### 2. Start the GUI Client

Start the Engo-based graphical client:

```bash
./client -renderer=engo -server=localhost:8080 -name=YourName -team=0
```

### 3. Start the Terminal Client (Alternative)

You can also use the original terminal-based client:

```bash
./client -renderer=terminal -server=localhost:8080 -name=YourName -team=0
```

## Command Line Options

### Common Options
- `-server=ADDRESS` - Server address (default: config file setting)
- `-name=NAME` - Your player name (default: "Player")
- `-team=ID` - Team ID to join (default: 0)
- `-config=FILE` - Configuration file path (default: "config.json")

### Renderer Selection
- `-renderer=TYPE` - Choose renderer: "engo" for GUI or "terminal" for console (default: "terminal")

### Engo GUI Options
- `-width=WIDTH` - Window width in pixels (default: 1024)
- `-height=HEIGHT` - Window height in pixels (default: 768)
- `-fullscreen` - Run in fullscreen mode (default: false)

## Controls (Engo GUI)

### Movement
- **W** or **↑** - Thrust forward
- **A** or **←** - Turn left
- **D** or **→** - Turn right

### Weapons
- **Space** - Fire current weapon
- **1-3** - Select weapon (simplified, only first 3 weapons)

### Actions
- **B** - Beam down armies to planet
- **Shift+B** - Beam up armies from planet
- **T** - Target nearest enemy

### UI
- **Enter** - Open chat (simplified input)
- **Escape** - Close chat/cancel action
- **R** - Reset camera zoom

### Mouse
- **Left Click** - Target entity (when mouse targeting is implemented)

## HUD Elements

The Engo GUI displays:
- **Ship Status** (top-left): Hull, shields, fuel, armies
- **Chat Window** (bottom-left): Chat messages
- **Minimap** (top-right): Overview of game world
- **Team Status**: Current team scores and statistics

## Architecture

The Engo client integrates with the existing go-netrek architecture:

### File Structure
```
pkg/render/engo/
├── renderer.go    # EngoRenderer implementing entity.Renderer
├── scene.go       # GameScene and Engo integration
├── input.go       # InputSystem for keyboard/mouse handling
├── camera.go      # CameraSystem for viewport management
├── hud.go         # HUDSystem for UI elements
└── assets.go      # AssetManager for sprite generation
```

### Integration Points
- **Network**: Uses existing `network.GameClient` for server communication
- **Game State**: Processes `engine.GameState` updates from server
- **Events**: Subscribes to `event.Bus` for chat and game events
- **Entities**: Converts `engine.*State` types to `entity.*` types for rendering

### Rendering Pipeline
1. Receive `GameState` updates from server via `GameClient.GetGameStateChannel()`
2. Convert state objects (`ShipState`, `PlanetState`, `ProjectileState`) to entity objects
3. Render entities using `EngoRenderer` implementing `entity.Renderer` interface
4. Handle user input via `InputSystem` and send to server via `GameClient.SendInput()`
5. Update camera to follow player's ship
6. Render HUD elements (ship status, chat, minimap)

## Current Limitations

1. **Key Constants**: Some Engo key constants may not be available, resulting in simplified input mapping
2. **Camera Control**: Camera system is basic - advanced zoom/pan controls need refinement
3. **Asset System**: Uses procedural sprite generation rather than loading image files
4. **Text Rendering**: Chat input is simplified due to Engo text input limitations
5. **Mouse Support**: Mouse targeting is partially implemented

## Future Improvements

1. **Enhanced Graphics**: Load proper sprite files and textures
2. **Advanced Input**: Full keyboard text input for chat
3. **Better Camera**: Smooth scrolling, zoom limits, world boundaries
4. **Sound**: Add audio support using Engo's audio system
5. **Effects**: Particle effects for explosions, weapon impacts
6. **UI Polish**: Better HUD design, tooltips, menus

## Troubleshooting

### Build Issues
- Ensure Go 1.19+ is installed
- Run `go mod tidy` to fetch dependencies
- Check that OpenGL drivers are installed

### Runtime Issues
- If the window doesn't open, check OpenGL support
- If input doesn't work, verify Engo key constant availability
- If rendering is slow, check graphics drivers and system specs

### Network Issues
- Ensure server is running before starting client
- Check server address and port are correct
- Verify firewall settings allow connections

## Development

To extend the Engo client:

1. **Add New Rendering Features**: Modify `renderer.go` or add new drawable types
2. **Enhance Input**: Update `input.go` with additional key mappings
3. **Improve HUD**: Extend `hud.go` with new UI elements
4. **Add Graphics**: Update `assets.go` to load actual image files

The implementation follows go-netrek's coding guidelines and maintains compatibility with the existing game architecture.
