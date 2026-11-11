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
	// Opacity scales the effective alpha of Stroke and Fill (0..1).
	// Zero is the opaque sentinel.
	Opacity float64
}

// applyOpacity bakes the style opacity into the stroke/fill alpha.
func applyOpacity(s Style) Style {
	if s.Opacity > 0 {
		s.Stroke.A = uint8(float64(s.Stroke.A) * s.Opacity)
		s.Fill.A = uint8(float64(s.Fill.A) * s.Opacity)
		s.Opacity = 0
	}
	return s
}

// SolidDash reports whether the style strokes with a solid line.
func (s Style) SolidDash() bool { return len(s.Dash) == 0 }

// Opaque reports whether both colors are fully opaque.
func (s Style) Opaque() bool { return s.Stroke.A == 255 && s.Fill.A == 255 }
