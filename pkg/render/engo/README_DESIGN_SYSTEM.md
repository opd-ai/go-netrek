# Design System Constants

This file documents the design system constants used throughout the go-netrek UI to ensure visual consistency and maintainability.

## Overview

The design system follows established UI/UX principles:
- **8-point grid system** for harmonious spacing
- **Semantic color naming** for consistent theming
- **Standardized dimensions** for predictable layouts
- **Centralized constants** for easy maintenance

## Usage

Import the constants from the package:

```go
import "github.com/opd-ai/go-netrek/pkg/render/engo"
```

### Spacing

Use the margin constants for consistent spacing:

```go
// Small spacing between related elements
hud.renderText(text, engo.MarginSmall, engo.MarginSmall, color)

// Medium spacing between sections
chatY := baseY + engo.MarginMedium

// Large spacing for major separations
panelX := engo.MarginLarge
```

### Colors

Use semantic color names instead of hardcoded RGBA values:

```go
// Good ✓
background := engo.ColorBackgroundChat
textColor := engo.ColorTextPrimary
teamColor := engo.GetTeamColor(teamID)

// Avoid ✗
background := color.RGBA{0, 0, 0, 128}
textColor := color.RGBA{255, 255, 255, 255}
```

### Dimensions

Use standardized dimensions for consistent component sizing:

```go
// Chat window
width := engo.ChatWindowWidth
height := engo.ChatWindowHeight

// Minimap
size := engo.MinimapDefaultSize

// Line heights
lineSpacing := engo.LineHeight
```

## Constants Reference

### Spacing (8-point grid)

| Constant | Value | Use Case |
|----------|-------|----------|
| `MarginXSmall` | 4px | Minimal spacing for tight layouts |
| `MarginSmall` | 8px | Standard small spacing between related elements |
| `MarginMedium` | 16px | Comfortable spacing between distinct UI sections |
| `MarginLarge` | 24px | Generous spacing for major layout separation |
| `MarginXLarge` | 32px | Maximum spacing for primary layout divisions |

### Typography

| Constant | Value | Use Case |
|----------|-------|----------|
| `FontSizeSmall` | 12px | Secondary text and footnotes |
| `FontSizeMedium` | 16px | Body text and standard UI elements |
| `FontSizeLarge` | 20px | Headings and emphasized content |
| `FontSizeXLarge` | 24px | Major headings and titles |

### Layout Dimensions

| Constant | Value | Description |
|----------|-------|-------------|
| `ChatWindowHeight` | 150px | Standard height for chat window |
| `ChatWindowWidth` | 400px | Standard width for chat window |
| `ConnectionStatusWidth` | 150px | Width allocation for connection status text |
| `MinimapDefaultSize` | 200px | Default size for the minimap |
| `MaxChatLines` | 10 | Maximum number of chat lines to display |
| `LineHeight` | 20px | Standard line height for text elements |
| `ChatLineHeight` | 15px | Line height specifically for chat messages |
| `TeamStatusLineHeight` | 20px | Line height for team status display |

### Alpha Transparency

| Constant | Value | Use Case |
|----------|-------|----------|
| `AlphaTransparent` | 0 | Completely transparent elements |
| `AlphaSemiTransparent` | 64 | Subtle overlay effects (25% opacity) |
| `AlphaBackground` | 128 | Standard background overlays (50% opacity) |
| `AlphaSubtle` | 192 | Light overlay effects (75% opacity) |
| `AlphaOpaque` | 255 | Fully opaque elements |

### Colors

#### Text Colors
- `ColorTextPrimary` - Primary text color for main content (White)
- `ColorTextSecondary` - Secondary text color for less important content (Light Gray)
- `ColorTextDisabled` - Disabled or inactive text (Gray)

#### Background Colors
- `ColorBackgroundChat` - Semi-transparent background for chat window
- `ColorBackgroundPanel` - Background for UI panels

#### Team Colors
- `ColorTeamRed` - Team 1 (red team)
- `ColorTeamGreen` - Team 2 (green team)
- `ColorTeamBlue` - Team 3 (blue team)
- `ColorTeamYellow` - Team 4 (yellow team)
- `ColorNeutral` - Neutral entities or unaffiliated content

#### Entity Colors
- `ColorEntityDefault` - Default color for game entities (White)
- `ColorEntityHighlight` - Highlight selected or important entities (Yellow)

## Team Color Management

Use the `GetTeamColor(teamID int)` function for consistent team color assignment:

```go
// Automatically handles invalid team IDs
teamColor := engo.GetTeamColor(teamID)

// For iterating over all team colors
for i, color := range engo.TeamColors {
    // Process each team color
}
```

## Migration Guide

When replacing magic numbers with design constants:

1. **Identify the purpose** of the magic number (spacing, color, dimension)
2. **Choose the appropriate constant** from the design system
3. **Update the code** to use the semantic constant name
4. **Test the visual result** to ensure it looks correct

### Before (Magic Numbers)
```go
hud.renderText(text, 10, 10, color.RGBA{255, 255, 255, 255})
hud.renderRect(10, y, 400, 150, color.RGBA{0, 0, 0, 128})
```

### After (Design Constants)
```go
hud.renderText(text, engo.MarginSmall, engo.MarginSmall, engo.ColorTextPrimary)
hud.renderRect(engo.MarginSmall, y, engo.ChatWindowWidth, engo.ChatWindowHeight, engo.ColorBackgroundChat)
```

## Design Principles

### Consistency
- All spacing follows the 8-point grid system
- All colors use semantic naming
- All dimensions use standardized constants

### Accessibility
- Minimum font size is 12px for readability
- Color contrasts are designed for visibility
- Transparent overlays maintain text legibility

### Maintainability
- Single source of truth for all design values
- Easy to update the entire UI by changing constants
- Clear semantic names make code self-documenting

## Contributing

When adding new constants:

1. Follow the existing naming conventions
2. Choose values that fit the 8-point grid system
3. Add comprehensive tests for new constants
4. Update this documentation
5. Consider the impact on existing layouts

For spacing constants, ensure they follow the progression:
- XSmall (4px) → Small (8px) → Medium (16px) → Large (24px) → XLarge (32px)

For colors, ensure they:
- Use semantic names (not descriptive like "LightBlue")
- Have sufficient contrast for accessibility
- Fit within the overall color scheme