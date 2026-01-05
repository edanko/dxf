package entity

import (
	"github.com/edanko/dxf/format"
)

// RGB color representation
type RGB struct {
	R, G, B uint8
}

// Pattern represents a hatch pattern definition
type Pattern struct {
	Lines []PatternLine
}

// NewPattern creates a new pattern
func NewPattern() *Pattern {
	return &Pattern{
		Lines: make([]PatternLine, 0),
	}
}

// AddLine adds a pattern line
func (p *Pattern) AddLine(angle, baseX, baseY, offsetX, offsetY float64, dashLengths ...float64) {
	line := PatternLine{
		Angle:     angle,
		BaseX:     baseX,
		BaseY:     baseY,
		OffsetX:   offsetX,
		OffsetY:   offsetY,
		DashItems: dashLengths,
	}
	p.Lines = append(p.Lines, line)
}

// Clone creates a deep copy of pattern
func (p *Pattern) Clone() *Pattern {
	clone := NewPattern()
	clone.Lines = make([]PatternLine, len(p.Lines))
	for i, line := range p.Lines {
		clone.Lines[i] = line
	}
	return clone
}

// Format writes pattern to formatter
func (p *Pattern) Format(f format.Formatter, isMPolygon bool) {
	if isMPolygon {
		f.WriteInt(78, len(p.Lines))
	} else {
		f.WriteInt(78, len(p.Lines))
	}

	for _, line := range p.Lines {
		line.Format(f)
	}
}

// PatternLine represents a single line in a hatch pattern
type PatternLine struct {
	Angle     float64   // 53 - Line angle in degrees
	BaseX     float64   // 43 - Base point X
	BaseY     float64   // 44 - Base point Y
	OffsetX   float64   // 45 - Offset vector X
	OffsetY   float64   // 46 - Offset vector Y
	DashItems []float64 // 49 - Dash length items
}

// Format writes pattern line to formatter
func (pl *PatternLine) Format(f format.Formatter) {
	f.WriteFloat(53, pl.Angle)
	f.WriteFloat(43, pl.BaseX)
	f.WriteFloat(44, pl.BaseY)
	f.WriteFloat(45, pl.OffsetX)
	f.WriteFloat(46, pl.OffsetY)

	for _, dash := range pl.DashItems {
		f.WriteFloat(49, dash)
	}

	f.WriteInt(79, 0) // End of pattern line
}

// Gradient represents a gradient fill definition
type Gradient struct {
	Kind           int     // 450 - 0=solid, 1=gradient
	NumberOfColors int     // 453 - Number of colors (usually 2)
	Color1         RGB     // First color
	Color2         RGB     // Second color
	OneColor       int     // 452 - 1=single color with tint
	Rotation       float64 // 460 - Rotation in radians
	Centered       float64 // 461 - Centered flag
	Tint           float64 // 462 - Tint value
	Name           string  // 470 - Gradient type name
}

// NewGradient creates a new gradient
func NewGradient() *Gradient {
	return &Gradient{
		Kind:           0,
		NumberOfColors: 2,
		Color1:         RGB{R: 255, G: 255, B: 255},
		Color2:         RGB{R: 255, G: 255, B: 255},
		OneColor:       0,
		Rotation:       0,
		Centered:       0,
		Tint:           0,
		Name:           "LINEAR",
	}
}

// Clone creates a deep copy of gradient
func (g *Gradient) Clone() *Gradient {
	clone := *g
	return &clone
}

// Format writes gradient to formatter
func (g *Gradient) Format(f format.Formatter) {
	// TODO: Add version checking when format supports it
	// R2004+ features assumed available

	// Gradient definition group codes
	f.WriteInt(450, g.Kind)
	f.WriteInt(451, 0) // Reserved
	f.WriteInt(452, g.OneColor)
	f.WriteInt(453, g.NumberOfColors)

	// Color 1
	f.WriteInt(463, int(g.Color1.R))
	f.WriteInt(464, int(g.Color1.G))
	f.WriteInt(465, int(g.Color1.B))

	// Color 2
	f.WriteInt(466, int(g.Color2.R))
	f.WriteInt(467, int(g.Color2.G))
	f.WriteInt(468, int(g.Color2.B))

	// Additional properties
	f.WriteFloat(460, g.Rotation)
	f.WriteFloat(461, g.Centered)
	f.WriteFloat(462, g.Tint)
	f.WriteString(470, g.Name)
}

// SetLinearGradient sets up a linear gradient
func (g *Gradient) SetLinearGradient(color1, color2 RGB, rotation float64) {
	g.Kind = 1
	g.NumberOfColors = 2
	g.Color1 = color1
	g.Color2 = color2
	g.Rotation = rotation
	g.Name = "LINEAR"
}

// SetRadialGradient sets up a radial gradient
func (g *Gradient) SetRadialGradient(color1, color2 RGB, centered float64) {
	g.Kind = 1
	g.NumberOfColors = 2
	g.Color1 = color1
	g.Color2 = color2
	g.Centered = centered
	g.Name = "SPHERICAL"
}

// SetSolidColor sets up solid color gradient
func (g *Gradient) SetSolidColor(color RGB, tint float64) {
	g.Kind = 1
	g.NumberOfColors = 1
	g.Color1 = color
	g.Color2 = color
	g.OneColor = 1
	g.Tint = tint
	g.Name = "LINEAR"
}

// Common gradient types
const (
	GradientLinear           = "LINEAR"
	GradientCylinder         = "CYLINDER"
	GradientInvCylinder      = "INVCYLINDER"
	GradientSpherical        = "SPHERICAL"
	GradientInvSpherical     = "INVSPHERICAL"
	GradientHemispherical    = "HEMISPHERICAL"
	GradientInvHemispherical = "INVHEMISPHERICAL"
	GradientCurved           = "CURVED"
	GradientInvCurved        = "INVCURVED"
)

// Helper functions for RGB colors
func NewRGB(r, g, b uint8) RGB {
	return RGB{R: r, G: g, B: b}
}

func RGBFromInt(value int) RGB {
	return RGB{
		R: uint8((value >> 16) & 0xFF),
		G: uint8((value >> 8) & 0xFF),
		B: uint8(value & 0xFF),
	}
}

func (rgb RGB) ToInt() int {
	return int(rgb.R)<<16 | int(rgb.G)<<8 | int(rgb.B)
}

// Predefined solid color patterns
var (
	SolidWhite   = RGB{R: 255, G: 255, B: 255}
	SolidBlack   = RGB{R: 0, G: 0, B: 0}
	SolidRed     = RGB{R: 255, G: 0, B: 0}
	SolidGreen   = RGB{R: 0, G: 255, B: 0}
	SolidBlue    = RGB{R: 0, G: 0, B: 255}
	SolidYellow  = RGB{R: 255, G: 255, B: 0}
	SolidMagenta = RGB{R: 255, G: 0, B: 255}
	SolidCyan    = RGB{R: 0, G: 255, B: 255}
)
