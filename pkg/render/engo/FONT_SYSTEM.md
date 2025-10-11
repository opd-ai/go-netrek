# Font System Implementation

This document describes the font system implementation in the go-netrek Engo renderer package.

## Overview

The font system provides centralized font management for the HUD and text rendering components. It integrates with the existing asset management system and design constants to ensure consistent typography across the UI.

## Components

### AssetManager Font Support

The `AssetManager` has been extended with font loading and management capabilities:

```go
type AssetManager struct {
    // ... existing fields ...
    fonts map[float64]*common.Font  // Fonts by size
}
```

#### Font Loading Methods

- `loadFonts() error` - Loads all standard font sizes during asset initialization
- `GetFont(size float64) *common.Font` - Returns font by size, with fallback to medium
- `GetSmallFont() *common.Font` - Returns small font (12pt)
- `GetMediumFont() *common.Font` - Returns medium font (16pt) 
- `GetLargeFont() *common.Font` - Returns large font (20pt)
- `GetXLargeFont() *common.Font` - Returns extra large font (24pt)

### HUD System Font Integration

The `HUDSystem` has been updated to work with the font system:

```go
type HUDSystem struct {
    // ... existing fields ...
    assetManager *AssetManager  // Reference to asset manager for font access
}
```

#### Font Management Methods

- `NewHUDSystem(assetManager *AssetManager) *HUDSystem` - Constructor accepting asset manager
- `SetFont(font *common.Font)` - Direct font assignment
- `SetFontSize(size float64)` - Set font by size using asset manager
- `SetSmallFont()` - Convenience method for small font
- `SetMediumFont()` - Convenience method for medium font  
- `SetLargeFont()` - Convenience method for large font

### Design System Integration

Font sizes are defined in the design constants and used consistently throughout:

```go
const (
    FontSizeSmall  = 12  // Secondary text and footnotes
    FontSizeMedium = 16  // Body text and standard UI elements
    FontSizeLarge  = 20  // Headings and emphasized content
    FontSizeXLarge = 24  // Major headings and titles
)
```

## Implementation Details

### Font Loading Process

1. **Initialization**: `AssetManager.loadFonts()` is called during `LoadAssets()`
2. **Font Creation**: Uses `common.LoadedFont()` with white text on transparent background
3. **Fallback**: If LoadedFont fails, creates basic font with `font.Create()`
4. **Storage**: Fonts stored in map by size for quick access

### HUD Font Initialization

1. **Constructor**: `NewHUDSystem()` accepts `AssetManager` parameter
2. **Default Font**: Automatically sets medium font from asset manager
3. **Runtime Changes**: Font can be changed using size methods or direct assignment

### Error Handling

- Font loading gracefully handles OpenGL context requirements
- Missing fonts fall back to medium size
- Nil asset manager handled safely in all HUD methods

## Usage Examples

### Basic Font Usage

```go
// Create asset manager and load fonts
assetManager := NewAssetManager()
err := assetManager.LoadAssets() // Includes font loading

// Create HUD with font support
hud := NewHUDSystem(assetManager)

// Change font sizes
hud.SetSmallFont()  // For secondary text
hud.SetLargeFont()  // For headings
```

### Custom Font Sizes

```go
// Get specific font size
font := assetManager.GetFont(18.0)
hud.SetFont(font)

// Or set directly by size
hud.SetFontSize(18.0)
```

### Renderer Integration

```go
// In GameScene setup
renderer := NewEngoRenderer(world)
err := renderer.Initialize() // Loads assets including fonts

// HUD gets fonts from renderer's asset manager
hud := NewHUDSystem(renderer.GetAssetManager())
```

## Testing

The font system includes comprehensive test coverage:

### AssetManager Tests
- Font map initialization
- Font getter methods
- Custom size handling
- Fallback behavior
- Integration tests

### HUD System Tests
- Constructor with/without asset manager
- Font size setting methods
- Nil safety
- Integration between HUD and asset manager

### Design Constant Tests
- Font size validation
- Logical size progression
- Consistency with design system

## Best Practices

1. **Always use design constants** for font sizes rather than hardcoded values
2. **Handle nil gracefully** - methods work safely even with nil asset manager
3. **Initialize fonts early** - load during asset manager initialization
4. **Use size methods** - prefer `SetMediumFont()` over `SetFontSize(16)`
5. **Test without OpenGL** - font loading tests expect failure in unit test environment

## Future Enhancements

Potential improvements to consider:

1. **Custom Font Files**: Load specific .ttf files for branded typography
2. **Font Metrics**: Add text measurement capabilities for better layout
3. **Font Caching**: Optimize font creation and reuse
4. **Style Variants**: Support bold, italic, and other font styles
5. **Localization**: Multi-language font support

## Dependencies

- **Engo Engine**: `github.com/EngoEngine/engo/common` for Font type
- **Design Constants**: Font size constants from design system
- **Asset Manager**: Central asset loading and management
- **HUD System**: Text rendering and UI display

## Files Modified

- `pkg/render/engo/assets.go` - Added font loading capabilities
- `pkg/render/engo/hud.go` - Added font management methods
- `pkg/render/engo/scene.go` - Updated HUD constructor call
- `pkg/render/engo/renderer.go` - Added GetAssetManager() method

## Files Added

- `pkg/render/engo/assets_font_test.go` - AssetManager font tests
- `pkg/render/engo/hud_font_test.go` - HUD font integration tests