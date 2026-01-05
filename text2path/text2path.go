package text2path

import (
	"fmt"
	"math"
	"strings"
)

// FontStyle represents text formatting options
type FontStyle struct {
	Family    string
	Size      float64
	Bold      bool
	Italic    bool
	Underline bool
}

// TextMetrics represents text measurement information
type TextMetrics struct {
	Width    float64
	Height   float64
	Ascent   float64
	Descent  float64
	BearingX float64
	BearingY float64
}

// Glyph represents a single character path
type Glyph struct {
	Char     rune
	AdvanceX float64
	AdvanceY float64
	BearingX float64
	BearingY float64
	Path     *Path
}

// Font represents a font with glyph information
type Font struct {
	Name       string
	Style      FontStyle
	Glyphs     map[rune]*Glyph
	UnitsPerEm float64
	Ascent     float64
	Descent    float64
	LineHeight float64
}

// Path represents a 2D vector path
type Path struct {
	Commands []PathCommand
	CurrentX float64
	CurrentY float64
	LastX    float64
	LastY    float64
}

// PathCommand represents a path operation
type PathCommand interface {
	Type() CommandType
	Execute(path *Path)
}

// CommandType represents different path operations
type CommandType int

const (
	CommandMoveTo CommandType = iota
	CommandLineTo
	CommandCurveTo
	CommandClose
)

// MoveTo represents a move-to command
type MoveTo struct {
	X, Y float64
}

func (m *MoveTo) Type() CommandType { return CommandMoveTo }

func (m *MoveTo) Execute(path *Path) {
	path.CurrentX, path.CurrentY = m.X, m.Y
}

// LineTo represents a line-to command
type LineTo struct {
	X, Y float64
}

func (l *LineTo) Type() CommandType { return CommandLineTo }

func (l *LineTo) Execute(path *Path) {
	path.CurrentX, path.CurrentY = l.X, l.Y
	path.LastX, path.LastY = l.X, l.Y
}

// CurveTo represents a quadratic curve command
type CurveTo struct {
	X, Y, CtrlX, CtrlY float64
}

func (c *CurveTo) Type() CommandType { return CommandCurveTo }

func (c *CurveTo) Execute(path *Path) {
	path.CurrentX, path.CurrentY = c.X, c.Y
	path.LastX, path.LastY = c.X, c.Y
}

// CurveTo4 adds a cubic curve command for complex curves
type CurveTo4 struct {
	X, Y, Ctrl1X, Ctrl1Y, Ctrl2X, Ctrl2Y float64
}

func (c *CurveTo4) Type() CommandType { return CommandCurveTo }

func (c *CurveTo4) Execute(path *Path) {
	path.CurrentX, path.CurrentY = c.X, c.Y
	path.LastX, path.LastY = c.X, c.Y
}

// Close represents a close path command
type Close struct{}

func (c *Close) Type() CommandType { return CommandClose }

func (c *Close) Execute(path *Path) {
	if len(path.Commands) > 0 {
		if path.CurrentX != path.LastX || path.CurrentY != path.LastY {
			path.Commands = append(path.Commands, &MoveTo{X: path.LastX, Y: path.LastY})
		}
		path.Commands = append(path.Commands, &LineTo{X: path.LastX, Y: path.LastY})
	}
}

// NewPath creates a new empty path
func NewPath() *Path {
	return &Path{
		Commands: make([]PathCommand, 0),
	}
}

// MoveTo adds a move command
func (p *Path) MoveTo(x, y float64) {
	p.Commands = append(p.Commands, &MoveTo{X: x, Y: y})
	p.CurrentX, p.CurrentY = x, y
}

// LineTo adds a line command
func (p *Path) LineTo(x, y float64) {
	p.Commands = append(p.Commands, &LineTo{X: x, Y: y})
	p.CurrentX, p.CurrentY = x, y
	p.LastX, p.LastY = x, y
}

// CurveTo adds a quadratic curve command
func (p *Path) CurveTo(ctrlX, ctrlY, x, y float64) {
	p.Commands = append(p.Commands, &CurveTo{X: x, Y: y, CtrlX: ctrlX, CtrlY: ctrlY})
	p.CurrentX, p.CurrentY = x, y
	p.LastX, p.LastY = x, y
}

// CurveTo4 adds a cubic curve command
func (p *Path) CurveTo4(ctrl1X, ctrl1Y, ctrl2X, ctrl2Y, x, y float64) {
	p.Commands = append(p.Commands, &CurveTo4{X: x, Y: y, Ctrl1X: ctrl1X, Ctrl1Y: ctrl1Y, Ctrl2X: ctrl2X, Ctrl2Y: ctrl2Y})
	p.CurrentX, p.CurrentY = x, y
	p.LastX, p.LastY = x, y
}

// Close closes the current path
func (p *Path) Close() {
	p.Commands = append(p.Commands, &Close{})
}

// BoundingBox calculates the bounding box of the path
func (p *Path) BoundingBox() (minX, minY, maxX, maxY float64) {
	if len(p.Commands) == 0 {
		return 0, 0, 0, 0
	}

	// Initialize with first point
	minX, minY = math.MaxFloat64, math.MaxFloat64
	maxX, maxY = -math.MaxFloat64, -math.MaxFloat64

	for _, cmd := range p.Commands {
		switch c := cmd.(type) {
		case *MoveTo:
			if c.X < minX {
				minX = c.X
			}
			if c.Y < minY {
				minY = c.Y
			}
			if c.X > maxX {
				maxX = c.X
			}
			if c.Y > maxY {
				maxY = c.Y
			}
		case *LineTo:
			if c.X < minX {
				minX = c.X
			}
			if c.Y < minY {
				minY = c.Y
			}
			if c.X > maxX {
				maxX = c.X
			}
			if c.Y > maxY {
				maxY = c.Y
			}
		case *CurveTo:
			// For curves, consider control points and end point
			if c.CtrlX < minX {
				minX = c.CtrlX
			}
			if c.CtrlY < minY {
				minY = c.CtrlY
			}
			if c.X < minX {
				minX = c.X
			}
			if c.Y < minY {
				minY = c.Y
			}
			if c.CtrlX > maxX {
				maxX = c.CtrlX
			}
			if c.CtrlY > maxY {
				maxY = c.CtrlY
			}
			if c.X > maxX {
				maxX = c.X
			}
			if c.Y > maxY {
				maxY = c.Y
			}
		case *CurveTo4:
			// For cubic curves, consider all control points and end point
			if c.Ctrl1X < minX {
				minX = c.Ctrl1X
			}
			if c.Ctrl1Y < minY {
				minY = c.Ctrl1Y
			}
			if c.Ctrl2X < minX {
				minX = c.Ctrl2X
			}
			if c.Ctrl2Y < minY {
				minY = c.Ctrl2Y
			}
			if c.X < minX {
				minX = c.X
			}
			if c.Y < minY {
				minY = c.Y
			}
			if c.Ctrl1X > maxX {
				maxX = c.Ctrl1X
			}
			if c.Ctrl1Y > maxY {
				maxY = c.Ctrl1Y
			}
			if c.Ctrl2X > maxX {
				maxX = c.Ctrl2X
			}
			if c.Ctrl2Y > maxY {
				maxY = c.Ctrl2Y
			}
			if c.X > maxX {
				maxX = c.X
			}
			if c.Y > maxY {
				maxY = c.Y
			}
		}
	}

	return minX, minY, maxX, maxY
}

// Clone creates a copy of the path
func (p *Path) Clone() *Path {
	newPath := &Path{
		Commands: make([]PathCommand, len(p.Commands)),
		CurrentX: p.CurrentX,
		CurrentY: p.CurrentY,
		LastX:    p.LastX,
		LastY:    p.LastY,
	}

	for i, cmd := range p.Commands {
		switch c := cmd.(type) {
		case *MoveTo:
			newPath.Commands[i] = &MoveTo{X: c.X, Y: c.Y}
		case *LineTo:
			newPath.Commands[i] = &LineTo{X: c.X, Y: c.Y}
		case *CurveTo:
			newPath.Commands[i] = &CurveTo{X: c.X, Y: c.Y, CtrlX: c.CtrlX, CtrlY: c.CtrlY}
		case *CurveTo4:
			newPath.Commands[i] = &CurveTo4{X: c.X, Y: c.Y, Ctrl1X: c.Ctrl1X, Ctrl1Y: c.Ctrl1Y, Ctrl2X: c.Ctrl2X, Ctrl2Y: c.Ctrl2Y}
		case *Close:
			newPath.Commands[i] = &Close{}
		}
	}

	return newPath
}

// PathConverter handles path operations
type PathConverter struct {
	Scale float64
}

// NewPathConverter creates a new path converter
func NewPathConverter(scale float64) *PathConverter {
	return &PathConverter{Scale: scale}
}

// CreatePath creates a path with basic operations
func (pc *PathConverter) CreatePath(operations string) *Path {
	path := NewPath()

	// Simple parser for basic operations
	parts := strings.Fields(operations)
	for i := 0; i < len(parts); i++ {
		switch parts[i] {
		case "M", "move":
			if i+2 < len(parts) {
				path.MoveTo(parseCoord(parts[i+1]), parseCoord(parts[i+2]))
				i += 2
			}
		case "L", "line":
			if i+2 < len(parts) {
				path.LineTo(parseCoord(parts[i+1]), parseCoord(parts[i+2]))
				i += 2
			}
		case "C", "curve":
			if i+4 < len(parts) {
				path.CurveTo(parseCoord(parts[i+1]), parseCoord(parts[i+2]), parseCoord(parts[i+3]), parseCoord(parts[i+4]))
				i += 4
			}
		case "Z", "close":
			path.Close()
		}
	}

	return path
}

// parseCoord parses a coordinate value
func parseCoord(s string) float64 {
	var val float64
	fmt.Sscanf(s, "%f", &val)
	return val
}
