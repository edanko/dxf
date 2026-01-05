package entity

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/math"
)

// Textstyle represents a STYLE (Text Style) table entry
type Textstyle struct {
	*entity
	Name            string  // 2 - style name
	Flags           int     // 70 - flags (1=shape, 4=vertical, 16=xref dependent, etc.)
	Height          float64 // 40 - fixed height (0=not fixed)
	Width           float64 // 41 - width factor (stretch)
	ObliqueAngle    float64 // 50 - oblique angle in degrees
	GenerationFlags int     // 71 - 2=backward, 4=mirrored in Y
	LastHeight      float64 // 42 - last height used
	Font            string  // 3 - primary font file name
	BigFont         string  // 4 - big font name
}

// NewTextstyle creates a new Textstyle entity
func NewTextstyle() *Textstyle {
	t := &Textstyle{
		entity:          NewEntity(TEXTSTYLE),
		Name:            "Standard",
		Flags:           0,
		Height:          0.0,
		Width:           1.0,
		ObliqueAngle:    0.0,
		GenerationFlags: 0,
		LastHeight:      2.5,
		Font:            "txt.shx",
		BigFont:         "",
	}
	return t
}

// NewTextstyleWithFont creates a text style with specified font
func NewTextstyleWithFont(name, font string) *Textstyle {
	t := NewTextstyle()
	t.Name = name
	t.Font = font
	return t
}

// IsEntity is for Entity interface.
func (t *Textstyle) IsEntity() bool {
	return true
}

// Format writes data to formatter.
func (t *Textstyle) Format(f format.Formatter) {
	t.entity.Format(f)
	f.WriteString(100, "AcDbTextStyleTableRecord")
	f.WriteString(100, "AcDbTextStyleTableRecord")
	f.WriteString(2, t.Name)
	f.WriteInt(70, t.Flags)
	f.WriteFloat(40, t.Height)
	f.WriteFloat(41, t.Width)
	f.WriteFloat(50, t.ObliqueAngle)
	f.WriteInt(71, t.GenerationFlags)
	f.WriteFloat(42, t.LastHeight)
	f.WriteString(3, t.Font)
	f.WriteString(4, t.BigFont)
}

// BBox returns the bounding box of the entity
func (t *Textstyle) BBox() ([]float64, []float64) {
	// Textstyle is a table entry, return empty bounding box
	return []float64{0, 0, 0}, []float64{0, 0, 0}
}

// Transform applies a transformation matrix to the Textstyle
func (t *Textstyle) Transform(m *math.Matrix44) error {
	// Textstyle doesn't have geometric data to transform
	return nil
}

// Copy creates a deep copy of the Textstyle entity
func (t *Textstyle) Copy() Entity {
	ts := NewTextstyle()
	ts.entity = t.entity
	ts.Name = t.Name
	ts.Flags = t.Flags
	ts.Height = t.Height
	ts.Width = t.Width
	ts.ObliqueAngle = t.ObliqueAngle
	ts.GenerationFlags = t.GenerationFlags
	ts.LastHeight = t.LastHeight
	ts.Font = t.Font
	ts.BigFont = t.BigFont
	return ts
}

// Validate validates the Textstyle entity
func (t *Textstyle) Validate() error {
	if t.Height < 0 {
		t.Height = 0.0
	}
	if t.Width <= 0 {
		t.Width = 1.0
	}
	if t.LastHeight < 0 {
		t.LastHeight = 2.5
	}
	return nil
}

// IsVertical returns true if vertical text is enabled
func (t *Textstyle) IsVertical() bool {
	return (t.Flags & 4) != 0
}

// SetVertical sets the vertical text flag
func (t *Textstyle) SetVertical(vertical bool) {
	if vertical {
		t.Flags |= 4
	} else {
		t.Flags &^= 4
	}
}

// IsBackward returns true if text is backward
func (t *Textstyle) IsBackward() bool {
	return (t.GenerationFlags & 2) != 0
}

// SetBackward sets the backward flag
func (t *Textstyle) SetBackward(backward bool) {
	if backward {
		t.GenerationFlags |= 2
	} else {
		t.GenerationFlags &^= 2
	}
}

// IsMirroredY returns true if text is mirrored in Y
func (t *Textstyle) IsMirroredY() bool {
	return (t.GenerationFlags & 4) != 0
}

// SetMirroredY sets the Y mirroring flag
func (t *Textstyle) SetMirroredY(mirrored bool) {
	if mirrored {
		t.GenerationFlags |= 4
	} else {
		t.GenerationFlags &^= 4
	}
}

// IsShape returns true if this describes a shape
func (t *Textstyle) IsShape() bool {
	return (t.Flags & 1) != 0
}

// SetShape sets the shape flag
func (t *Textstyle) SetShape(shape bool) {
	if shape {
		t.Flags |= 1
	} else {
		t.Flags &^= 1
	}
}
