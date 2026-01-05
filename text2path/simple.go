package text2path

import (
	"fmt"
	"github.com/edanko/dxf/path"
)

// TextToPath converts text strings to DXF path operations
type TextToPath struct {
	Scale float64
}

// NewTextToPath creates a new text-to-path converter
func NewTextToPath(scale float64) *TextToPath {
	return &TextToPath{Scale: scale}
}

// Convert converts text to a path using simple character approximations
func (t *TextToPath) Convert(text string) *path.Path {
	dxfPath := path.NewPath()

	if len(text) == 0 {
		return dxfPath
	}

	// Starting position
	x := 0.0
	y := 0.0
	charWidth := 100.0 * t.Scale
	charHeight := 120.0 * t.Scale
	lineHeight := charHeight + 20.0*t.Scale // Spacing between lines

	for _, char := range text {
		if char == ' ' || char == '\t' {
			// Space or tab - just advance
			x += charWidth
			continue
		}

		if char == '\n' {
			// New line - reset x and advance y
			x = 0
			y -= lineHeight
			continue
		}

		// Simple character path
		charPath := t.createCharPath(char, x, y)
		if charPath != nil {
			// Add character path to main path
			// Since there's no direct append method, we recreate the path operations
			currentPos := dxfPath.CurrentPosition()
			for _, cmd := range charPath.Commands() {
				switch c := cmd.(type) {
				case *path.PathMoveTo:
					dxfPath.MoveTo(currentPos.X()+c.End.X(), currentPos.Y()+c.End.Y())
				case *path.PathLineTo:
					dxfPath.LineTo(currentPos.X()+c.End.X(), currentPos.Y()+c.End.Y())
				case *path.PathCurve3To:
					dxfPath.Curve3To(
						currentPos.X()+c.End.X(), currentPos.Y()+c.End.Y(),
						currentPos.X()+c.Ctrl.X(), currentPos.Y()+c.Ctrl.Y(),
					)
				case *path.PathCurve4To:
					dxfPath.Curve4To(
						currentPos.X()+c.End.X(), currentPos.Y()+c.End.Y(),
						currentPos.X()+c.Ctrl1.X(), currentPos.Y()+c.Ctrl1.Y(),
						currentPos.X()+c.Ctrl2.X(), currentPos.Y()+c.Ctrl2.Y(),
					)
				}
				currentPos = dxfPath.CurrentPosition()
			}
		}

		// Advance to next character position
		x += charWidth
	}

	return dxfPath
}

// createCharPath creates a simple path for a single character
func (t *TextToPath) createCharPath(char rune, x, y float64) *path.Path {
	charPath := path.NewPath()

	// Scale character dimensions
	scale := t.Scale

	switch string(char) {
	case "A", "a":
		charPath.MoveTo(x, y)
		charPath.LineTo(x+80*scale, y)
		charPath.LineTo(x+80*scale, y+100*scale)
		charPath.LineTo(x+40*scale, y+100*scale)
		charPath.LineTo(x, y+100*scale)
		charPath.LineTo(x, y)
	case "B", "b":
		charPath.MoveTo(x, y)
		charPath.LineTo(x+60*scale, y+100*scale)
		charPath.LineTo(x+60*scale, y)
		charPath.LineTo(x+60*scale, y+100*scale)
		charPath.LineTo(x, y+100*scale)
		charPath.LineTo(x, y)
	case "C", "c":
		charPath.MoveTo(x+60*scale, y+100*scale)
		charPath.LineTo(x, y)
		charPath.LineTo(x, y)
	case "D", "d":
		charPath.MoveTo(x, y)
		charPath.LineTo(x+60*scale, y)
		charPath.LineTo(x+60*scale, y+100*scale)
		charPath.LineTo(x, y+100*scale)
		charPath.LineTo(x, y)
	case "E", "e":
		charPath.MoveTo(x+20*scale, y+100*scale)
		charPath.LineTo(x+20*scale, y)
		charPath.LineTo(x+20*scale, y+100*scale)
		charPath.LineTo(x+80*scale, y+100*scale)
		charPath.LineTo(x+80*scale, y+100*scale)
		charPath.LineTo(x+80*scale, y)
		charPath.LineTo(x+80*scale, y)
	case "F", "f":
		charPath.MoveTo(x, y+50*scale)
		charPath.LineTo(x, y+100*scale)
		charPath.LineTo(x+50*scale, y)
		charPath.LineTo(x+50*scale, y+100*scale)
		charPath.LineTo(x+50*scale, y+100*scale)
		charPath.LineTo(x, y)
	case "G", "g":
		charPath.MoveTo(x+20*scale, y+80*scale)
		charPath.LineTo(x+20*scale, y+80*scale)
		charPath.LineTo(x+20*scale, y+20*scale)
		charPath.LineTo(x+60*scale, y+20*scale)
		charPath.LineTo(x+60*scale, y+80*scale)
		charPath.LineTo(x+60*scale, y)
		charPath.LineTo(x+60*scale, y)
	case "H", "h":
		charPath.MoveTo(x, y)
		charPath.LineTo(x, y+80*scale)
		charPath.LineTo(x+60*scale, y+80*scale)
		charPath.LineTo(x+60*scale, y)
		charPath.LineTo(x+60*scale, y+80*scale)
		charPath.LineTo(x, y+80*scale)
		charPath.LineTo(x, y)
	case "I", "i":
		charPath.MoveTo(x, y)
		charPath.LineTo(x+60*scale, y)
		charPath.LineTo(x, y+40*scale)
		charPath.LineTo(x, y+40*scale)
		charPath.LineTo(x+60*scale, y)
		charPath.LineTo(x, y)
	case "J", "j":
		charPath.MoveTo(x, y)
		charPath.LineTo(x, y+80*scale)
		charPath.LineTo(x, y+80*scale)
		charPath.LineTo(x, y+40*scale)
		charPath.LineTo(x, y)
	case "K", "k":
		charPath.MoveTo(x, y+80*scale)
		charPath.LineTo(x, y+40*scale)
		charPath.LineTo(x, y)
		charPath.LineTo(x, y+80*scale)
		charPath.LineTo(x, y+80*scale)
		charPath.LineTo(x, y)
	case "L", "l":
		charPath.MoveTo(x, y)
		charPath.LineTo(x, y)
		charPath.LineTo(x, y+80*scale)
		charPath.LineTo(x, y+80*scale)
		charPath.LineTo(x, y)
	case "M", "m":
		charPath.MoveTo(x+20*scale, y)
		charPath.LineTo(x+40*scale, y)
		charPath.LineTo(x+60*scale, y+40*scale)
		charPath.LineTo(x+60*scale, y+80*scale)
		charPath.LineTo(x+80*scale, y+80*scale)
		charPath.LineTo(x+80*scale, y)
		charPath.LineTo(x+20*scale, y)
	case "N", "n":
		charPath.MoveTo(x+20*scale, y+80*scale)
		charPath.LineTo(x+20*scale, y+40*scale)
		charPath.LineTo(x+60*scale, y+40*scale)
		charPath.LineTo(x+60*scale, y+80*scale)
		charPath.LineTo(x+80*scale, y+80*scale)
		charPath.LineTo(x+80*scale, y)
		charPath.LineTo(x+20*scale, y)
	case "O", "o":
		charPath.MoveTo(x+40*scale, y+40*scale)
		// Octagon using curves
		charPath.LineTo(x+60*scale, y+20*scale)
		charPath.LineTo(x+60*scale, y)
		charPath.LineTo(x+40*scale, y)
	case "P", "p":
		charPath.MoveTo(x+20*scale, y+80*scale)
		charPath.LineTo(x, y+80*scale)
		charPath.LineTo(x, y)
	case "Q", "q":
		charPath.MoveTo(x+20*scale, y+80*scale)
		charPath.LineTo(x, y+80*scale)
		charPath.LineTo(x+40*scale, y+80*scale)
		charPath.LineTo(x, y)
	case "R", "r":
		charPath.MoveTo(x, y)
		charPath.LineTo(x+40*scale, y)
		charPath.LineTo(x+40*scale, y+40*scale)
		charPath.LineTo(x+40*scale, y+80*scale)
		charPath.LineTo(x, y+80*scale)
		charPath.LineTo(x, y)
	case "S", "s":
		charPath.MoveTo(x, y)
		charPath.LineTo(x, y+80*scale)
		charPath.LineTo(x+40*scale, y+80*scale)
		charPath.LineTo(x+40*scale, y+40*scale)
		charPath.LineTo(x+40*scale, y)
		charPath.LineTo(x, y+80*scale)
		charPath.LineTo(x+40*scale, y+80*scale)
		charPath.LineTo(x+40*scale, y+80*scale)
		charPath.LineTo(x, y)
	case "T", "t":
		charPath.MoveTo(x+10*scale, y+80*scale)
		charPath.LineTo(x+10*scale, y+40*scale)
		charPath.LineTo(x+30*scale, y)
		charPath.LineTo(x+50*scale, y)
		charPath.LineTo(x+70*scale, y+40*scale)
		charPath.LineTo(x+70*scale, y)
	case "U", "u":
		charPath.MoveTo(x+20*scale, y+40*scale)
		charPath.LineTo(x, y)
		charPath.LineTo(x+40*scale, y+80*scale)
		charPath.LineTo(x, y+80*scale)
		charPath.LineTo(x+20*scale, y)
	case "V", "v":
		charPath.MoveTo(x+30*scale, y)
		charPath.LineTo(x, y+80*scale)
		charPath.LineTo(x, y+40*scale)
		charPath.LineTo(x+70*scale, y+40*scale)
		charPath.LineTo(x+70*scale, y)
	case "W", "w":
		charPath.MoveTo(x+10*scale, y+40*scale)
		charPath.LineTo(x, y)
		charPath.LineTo(x+80*scale, y+40*scale)
		charPath.LineTo(x+80*scale, y)
		charPath.LineTo(x+90*scale, y+40*scale)
		charPath.LineTo(x+90*scale, y+80*scale)
		charPath.LineTo(x+10*scale, y)
	case "X", "x":
		charPath.MoveTo(x+20*scale, y+40*scale)
		charPath.LineTo(x+80*scale, y+40*scale)
		charPath.LineTo(x+80*scale, y)
		charPath.LineTo(x+20*scale, y+40*scale)
		charPath.LineTo(x+20*scale, y)
	case "Y", "y":
		charPath.MoveTo(x+20*scale, y+80*scale)
		charPath.LineTo(x+20*scale, y)
		charPath.LineTo(x+20*scale, y+40*scale)
		charPath.LineTo(x, y)
	case "Z", "z":
		charPath.MoveTo(x+30*scale, y+80*scale)
		charPath.LineTo(x+30*scale, y+40*scale)
		charPath.LineTo(x, y)
		charPath.LineTo(x+70*scale, y+40*scale)
		charPath.LineTo(x+70*scale, y+80*scale)
		charPath.LineTo(x+30*scale, y+80*scale)
		charPath.LineTo(x+70*scale, y)
	default:
		// Numbers and symbols - simple box
		boxSize := 60 * scale
		charPath.MoveTo(x, y)
		charPath.LineTo(x+boxSize, y)
		charPath.LineTo(x+boxSize, y+boxSize)
		charPath.LineTo(x, y+boxSize)
		charPath.LineTo(x, y)
	}

	return charPath
}

// ConvertToDxfPath converts the path to DXF-compatible format
func (t *TextToPath) ConvertToDxfPath() string {
	// This would generate a DXF path string
	// For now, return basic path representation
	return fmt.Sprintf("Text-to-Path converted with scale %.2f", t.Scale)
}

// GetTextMetrics calculates approximate text dimensions
func (t *TextToPath) GetTextMetrics(text string) (width, height float64) {
	if len(text) == 0 {
		return 0, 0
	}

	// Simple approximation based on character count and scale
	charCount := 0
	for _, char := range text {
		if char != ' ' && char != '\t' && char != '\n' {
			charCount++
		}
	}

	width, height = float64(charCount)*100.0*t.Scale, 120.0*t.Scale

	return width, height
}
