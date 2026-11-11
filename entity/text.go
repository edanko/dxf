package entity

import (
	"math"

	"github.com/edanko/dxf/format"
	dxfmath "github.com/edanko/dxf/math"
	"github.com/edanko/dxf/table"
)

// Text Anchor
const (
	LEFT_BASE = iota
	CENTER_BASE
	RIGHT_BASE
	LEFT_BOTTOM
	CENTER_BOTTOM
	RIGHT_BOTTOM
	LEFT_CENTER
	CENTER_CENTER
	RIGHT_CENTER
	LEFT_TOP
	CENTER_TOP
	RIGHT_TOP
)

// Text represents TEXT Entity.
type Text struct {
	*entity
	Coord1         []float64    // 10, 20, 30
	Coord2         []float64    // 11, 21, 31
	Height         float64      // 40
	Rotation       float64      // 50
	WidthFactor    float64      // 41
	ObliqueAngle   float64      // 51
	Value          string       // 1
	Style          *table.Style // 7
	GenFlag        int          // 71
	HorizontalFlag int          // 72
	VerticalFlag   int          // 73
}

// IsEntity is for Entity interface.
func (t *Text) IsEntity() bool {
	return true
}

// NewText creates a new Text.
func NewText() *Text {
	t := &Text{
		entity:         NewEntity(TEXT),
		Coord1:         []float64{0.0, 0.0, 0.0},
		Coord2:         []float64{0.0, 0.0, 0.0},
		Height:         1.0,
		WidthFactor:    1.0,
		Rotation:       0.0,
		ObliqueAngle:   0.0,
		Value:          "",
		GenFlag:        0,
		HorizontalFlag: 0,
		VerticalFlag:   0,
	}
	return t
}

// Format writes data to formatter.
func (t *Text) Format(f format.Formatter) {
	t.entity.Format(f)
	f.WriteString(100, "AcDbText")
	for i := range 3 {
		f.WriteFloat((i+1)*10, t.Coord1[i])
	}
	f.WriteFloat(40, t.Height)
	f.WriteFloat(50, t.Rotation)
	f.WriteFloat(41, t.WidthFactor)
	f.WriteFloat(51, t.ObliqueAngle)
	f.WriteString(1, t.Value)
	if t.Style != nil {
		f.WriteString(7, t.Style.Name())
	}
	if t.GenFlag != 0 {
		f.WriteInt(71, t.GenFlag)
	}
	if t.HorizontalFlag != 0 {
		f.WriteInt(72, t.HorizontalFlag)
		if t.VerticalFlag != 0 {
			for i := range 3 {
				f.WriteFloat((i+1)*10+1, t.Coord2[i])
			}
		}
	}
	f.WriteString(100, "AcDbText")
	if t.VerticalFlag != 0 {
		f.WriteInt(73, t.VerticalFlag)
	}
}

func (t *Text) togglegenflag(val int) {
	if t.GenFlag&val != 0 {
		t.GenFlag &= ^val
	} else {
		t.GenFlag |= val
	}
}

// FlipHorizontal flips Text horizontally.
func (t *Text) FlipHorizontal() {
	t.togglegenflag(2)
}

// FlipVertical flips Text vertically.
func (t *Text) FlipVertical() {
	t.togglegenflag(4)
}

// Anchor sets anchor point flags and returns the new anchor.
// Valid anchor values are LEFT_BASE (0) through RIGHT_TOP (11).
func (t *Text) Anchor(pos int) int {
	if pos < LEFT_BASE || pos > RIGHT_TOP {
		return t.GetAnchor()
	}
	t.HorizontalFlag = pos % 3
	t.VerticalFlag = pos / 3
	return pos
}

// GetAnchor returns the current anchor point
func (t *Text) GetAnchor() int {
	return t.HorizontalFlag + t.VerticalFlag*3
}

// TextLength calculates the approximate length of the text.
// This is an approximation based on character count, height, and width factor.
// For accurate measurements, a font renderer would be needed.
func (t *Text) TextLength() float64 {
	if t.Value == "" {
		return 0
	}
	// Approximate: average character width is about 0.6 * height
	// Then apply the width factor
	approxCharWidth := t.Height * 0.6 * t.WidthFactor
	return float64(len(t.Value)) * approxCharWidth
}

// BBox returns the bounding box of the text entity.
// The bounding box considers the text length, height, and anchor point.
func (t *Text) BBox() ([]float64, []float64) {
	if len(t.Coord1) < 3 {
		return []float64{0, 0, 0}, []float64{0, 0, 0}
	}

	textLen := t.TextLength()
	anchor := t.GetAnchor()

	mins := make([]float64, 3)
	maxs := make([]float64, 3)

	// Calculate bounds based on anchor point
	// Horizontal anchor: 0=left, 1=center, 2=right
	// Vertical anchor: 0=base, 1=bottom, 2=center, 3=top
	horizontalAnchor := anchor % 3
	verticalAnchor := anchor / 3

	// X bounds based on horizontal anchor
	switch horizontalAnchor {
	case 0: // Left aligned
		mins[0] = t.Coord1[0]
		maxs[0] = t.Coord1[0] + textLen
	case 1: // Center aligned
		mins[0] = t.Coord1[0] - textLen/2
		maxs[0] = t.Coord1[0] + textLen/2
	case 2: // Right aligned
		mins[0] = t.Coord1[0] - textLen
		maxs[0] = t.Coord1[0]
	}

	// Y bounds based on vertical anchor
	switch verticalAnchor {
	case 0: // Base
		mins[1] = t.Coord1[1]
		maxs[1] = t.Coord1[1] + t.Height
	case 1: // Bottom
		mins[1] = t.Coord1[1]
		maxs[1] = t.Coord1[1] + t.Height
	case 2: // Center
		mins[1] = t.Coord1[1] - t.Height/2
		maxs[1] = t.Coord1[1] + t.Height/2
	case 3: // Top
		mins[1] = t.Coord1[1] - t.Height
		maxs[1] = t.Coord1[1]
	}

	mins[2] = t.Coord1[2]
	maxs[2] = t.Coord1[2]

	// Apply rotation if needed
	if t.Rotation != 0 {
		// For rotated text, expand the bounding box to be safe
		rad := t.Rotation * math.Pi / 180
		cosR := math.Cos(rad)
		sinR := math.Sin(rad)

		// Calculate the rotated corners
		width := maxs[0] - mins[0]
		height := maxs[1] - mins[1]
		centerX := (mins[0] + maxs[0]) / 2
		centerY := (mins[1] + maxs[1]) / 2

		// Rotated half-diagonal
		halfDiagX := width/2*math.Abs(cosR) + height/2*math.Abs(sinR)
		halfDiagY := width/2*math.Abs(sinR) + height/2*math.Abs(cosR)

		mins[0] = centerX - halfDiagX
		maxs[0] = centerX + halfDiagX
		mins[1] = centerY - halfDiagY
		maxs[1] = centerY + halfDiagY
	}

	return mins, maxs
}

// SetStyle sets text entity style.
func (t *Text) SetStyle(style *table.Style) {
	t.Style = style
}

// SetCoord sets the insertion point (first alignment point)
func (t *Text) SetCoord(coord []float64) {
	if len(coord) >= 3 {
		t.Coord1 = []float64{coord[0], coord[1], coord[2]}
	}
}

// SetSecondCoord sets the second alignment point (for fitted text)
func (t *Text) SetSecondCoord(coord []float64) {
	if len(coord) >= 3 {
		t.Coord2 = []float64{coord[0], coord[1], coord[2]}
	}
}

// SecondCoord returns the second alignment point
func (t *Text) SecondCoord() []float64 {
	return t.Coord2
}

// Transform applies an affine matrix to the text insertion point and rotates
// its angle by the XY-plane rotation component.
func (t *Text) Transform(m dxfmath.Matrix44) error {
	t.Coord1 = TransformPoint(m, t.Coord1)
	if len(t.Coord2) >= 2 {
		t.Coord2 = TransformPoint(m, t.Coord2)
	}
	sx, sy, _ := ScaleFactors(m)
	scale := (sx + sy) / 2
	t.Height *= scale
	if rot := math.Atan2(m.Get(1, 0), m.Get(0, 0)); rot != 0 {
		t.Rotation += rot * 180 / math.Pi
	}
	return nil
}

// Copy creates a deep copy of the Text entity
func (t *Text) Copy() Entity {
	text := NewText()
	text.entity = t.entity.clone()
	text.Coord1 = []float64{t.Coord1[0], t.Coord1[1], t.Coord1[2]}
	text.Coord2 = []float64{t.Coord2[0], t.Coord2[1], t.Coord2[2]}
	text.Height = t.Height
	text.Rotation = t.Rotation
	text.WidthFactor = t.WidthFactor
	text.ObliqueAngle = t.ObliqueAngle
	text.Value = t.Value
	text.Style = t.Style
	text.GenFlag = t.GenFlag
	text.HorizontalFlag = t.HorizontalFlag
	text.VerticalFlag = t.VerticalFlag
	return text
}
