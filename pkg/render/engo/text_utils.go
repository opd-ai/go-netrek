package engo

import (
	"strings"

	"github.com/EngoEngine/engo/common"
)

// TextMeasurement provides text dimension calculation capabilities
type TextMeasurement struct {
	font *common.Font
}

// NewTextMeasurement creates a new text measurement utility
func NewTextMeasurement(fontResource *common.Font) *TextMeasurement {
	return &TextMeasurement{
		font: fontResource,
	}
}

// MeasureText calculates the width and height of the given text using the current font.
// Returns approximate dimensions based on character count and font size.
func (tm *TextMeasurement) MeasureText(text string) (width, height float32) {
	if tm.font == nil {
		// Fallback to default size estimation
		return tm.approximateTextSize(text, 16)
	}

	// Use font size for calculation
	fontSize := float32(tm.font.Size)
	if fontSize <= 0 {
		fontSize = 16 // Default fallback
	}

	return tm.approximateTextSize(text, fontSize)
}

// approximateTextSize provides text measurement based on font size
func (tm *TextMeasurement) approximateTextSize(text string, fontSize float32) (width, height float32) {
	// Character width is approximately 0.6 * font size for monospace-like fonts
	charWidth := fontSize * 0.6
	height = fontSize * 1.2 // Line height is typically 1.2 * font size

	// Count actual characters (more accurate than len for unicode)
	charCount := float32(len([]rune(text)))
	width = charCount * charWidth

	return width, height
}

// MeasureMultiLineText calculates dimensions for multi-line text
func (tm *TextMeasurement) MeasureMultiLineText(lines []string) (width, height float32) {
	if len(lines) == 0 {
		return 0, 0
	}

	maxWidth := float32(0)
	lineHeight := float32(16) // Default

	if tm.font != nil && tm.font.Size > 0 {
		lineHeight = float32(tm.font.Size) * 1.2
	}

	for _, line := range lines {
		lineWidth, _ := tm.MeasureText(line)
		if lineWidth > maxWidth {
			maxWidth = lineWidth
		}
	}

	totalHeight := float32(len(lines)) * lineHeight
	return maxWidth, totalHeight
}

// RightAlignPosition calculates the X position for right-aligned text
func (tm *TextMeasurement) RightAlignPosition(text string, rightEdge, margin float32) float32 {
	textWidth, _ := tm.MeasureText(text)
	return rightEdge - textWidth - margin
}

// CenterAlignPosition calculates the X position for center-aligned text
func (tm *TextMeasurement) CenterAlignPosition(text string, containerWidth float32) float32 {
	textWidth, _ := tm.MeasureText(text)
	return (containerWidth - textWidth) / 2
}

// TruncateText truncates text to fit within the specified width
func (tm *TextMeasurement) TruncateText(text string, maxWidth float32) string {
	textWidth, _ := tm.MeasureText(text)
	if textWidth <= maxWidth {
		return text
	}

	// Estimate how many characters will fit
	fontSize := float32(16)
	if tm.font != nil && tm.font.Size > 0 {
		fontSize = float32(tm.font.Size)
	}

	charWidth := fontSize * 0.6
	ellipsisWidth := charWidth * 3 // "..." is 3 characters
	availableWidth := maxWidth - ellipsisWidth

	if availableWidth <= 0 {
		return "..."
	}

	maxChars := int(availableWidth / charWidth)
	if maxChars <= 0 {
		return "..."
	}

	runes := []rune(text)
	if maxChars >= len(runes) {
		return text
	}

	return string(runes[:maxChars]) + "..."
}

// WrapText wraps text to fit within specified width, returning lines
func (tm *TextMeasurement) WrapText(text string, maxWidth float32) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{}
	}

	var lines []string
	currentLine := ""

	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		lineWidth, _ := tm.MeasureText(testLine)

		if lineWidth <= maxWidth {
			currentLine = testLine
		} else {
			// Start new line
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// MeasureText calculates the width of the given text using the specified font.
// This is a convenience function that creates a TextMeasurement and measures the text.
//
// Parameters:
//   - text: The text to measure
//   - font: The font to use for measurement
//
// Returns:
//   - width: The width of the text in pixels
func MeasureText(text string, font *common.Font) float32 {
	if font == nil {
		// Fallback to estimated width if no font is available
		return float32(len(text) * 8) // 8px per character approximation
	}

	tm := &TextMeasurement{font: font}
	width, _ := tm.MeasureText(text)
	return width
}
