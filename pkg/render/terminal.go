package render

import (
    "fmt"
    "strings"

    "github.com/opd-ai/go-netrek/pkg/entity"
    "github.com/opd-ai/go-netrek/pkg/physics"
)

// TerminalRenderer provides a simple ASCII-based rendering for terminals
type TerminalRenderer struct {
    width     int
    height    int
    buffer    [][]rune
    scale     float64
    centerPos physics.Vector2D
}

// NewTerminalRenderer creates a new terminal renderer with the specified dimensions
func NewTerminalRenderer(width, height int, scale float64) *TerminalRenderer {
    buffer := make([][]rune, height)
    for i := range buffer {
        buffer[i] = make([]rune, width)
    }

    return &TerminalRenderer{
        width:  width,
        height: height,
        buffer: buffer,
        scale:  scale,
        centerPos: physics.Vector2D{
            X: 0,
            Y: 0,
        },
    }
}

// SetCenter sets the center position of the view
func (r *TerminalRenderer) SetCenter(pos physics.Vector2D) {
    r.centerPos = pos
}

// worldToScreen converts world coordinates to screen coordinates
func (r *TerminalRenderer) worldToScreen(pos physics.Vector2D) (int, int) {
    // Convert world coordinates to screen coordinates
    screenX := int((pos.X - r.centerPos.X) / r.scale + float64(r.width) / 2)
    screenY := int((pos.Y - r.centerPos.Y) / r.scale + float64(r.height) / 2)
    return screenX, screenY
}

// Clear implements entity.Renderer
func (r *TerminalRenderer) Clear() {
    for y := range r.buffer {
        for x := range r.buffer[y] {
            r.buffer[y][x] = ' '
        }
    }
}

// Present implements entity.Renderer
func (r *TerminalRenderer) Present() {
    // Clear terminal
    fmt.Print("\033[H\033[2J")
    
    // Draw border
    fmt.Println("+" + strings.Repeat("-", r.width) + "+")
    
    // Draw buffer
    for y := range r.buffer {
        fmt.Print("|")
        for x := range r.buffer[y] {
            fmt.Print(string(r.buffer[y][x]))
        }
        fmt.Println("|")
    }
    
    // Draw border
    fmt.Println("+" + strings.Repeat("-", r.width) + "+")
}

// RenderShip implements entity.Renderer
func (r *TerminalRenderer) RenderShip(ship *entity.Ship) {
    x, y := r.worldToScreen(ship.Position)
    
    // Check if within bounds
    if x >= 0 && x < r.width && y >= 0 && y < r.height {
        // Use different symbols based on ship class and team
        symbols := []rune{'S', 'D', 'C', 'B', 'A'}
        symbol := symbols[int(ship.Class)]
        
        // Add team-based coloring in a real implementation
        r.buffer[y][x] = symbol
    }
}

// RenderPlanet implements entity.Renderer
func (r *TerminalRenderer) RenderPlanet(planet *entity.Planet) {
    x, y := r.worldToScreen(planet.Position)
    
    // Check if within bounds
    if x >= 0 && x < r.width && y >= 0 && y < r.height {
        // Use O for planets
        r.buffer[y][x] = 'O'
    }
}

// RenderProjectile implements entity.Renderer
func (r *TerminalRenderer) RenderProjectile(projectile *entity.Projectile) {
    x, y := r.worldToScreen(projectile.Position)
    
    // Check if within bounds
    if x >= 0 && x < r.width && y >= 0 && y < r.height {
        // Use . for projectiles
        r.buffer[y][x] = '.'
    }
}