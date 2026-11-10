package render

import "image/color"

// Style carries the resolved visual attributes of one drawn path.
// Colors are pre-multiplied by the entity and layer transparency:
// a zero-alpha Stroke disables stroking, a zero-alpha Fill disables filling.
type Style struct {
	Stroke   color.RGBA
	Fill     color.RGBA
	Width    float64 // stroke width in canvas units
	Dash     []float64
	FillRule string // "" (nonzero) or "evenodd"
	LineCap  string // "butt" (default), "round", "square"
	LineJoin string // "miter" (default), "round", "bevel"
}

// SolidDash reports whether the style strokes with a solid line.
func (s Style) SolidDash() bool { return len(s.Dash) == 0 }

// Opaque reports whether both colors are fully opaque.
func (s Style) Opaque() bool { return s.Stroke.A == 255 && s.Fill.A == 255 }
