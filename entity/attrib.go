package entity

import (
	"fmt"
	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/format"
)

// Attrib represents an ATTRIB entity (attribute)
type Attrib struct {
	*entity
	tag         string    // 2 - Attribute tag
	textValue   string    // 1,3 - Text value/Default value
	insertPoint []float64 // 10,20,30 - Insertion point (in OCS)
	height      float64   // 40 - Text height
	rotation    float64   // 50 - Rotation angle (degrees)
	width       float64   // 41 - Text width factor
	oblique     float64   // 51 - Obliquity angle
	style       string    // 7 - Text style name
	mirror      int       // 71 - Mirror flag
	hAlign      int       // 72 - Horizontal justification
	vAlign      int       // 74 - Vertical justification
	layer       string    // 8 - Layer name
	color       int       // 62 - Color number
	extrusion   []float64 // 210,220,230 - Extrusion direction
	prompt      string    // 3 - Prompt (for ATTDEF)
}

// NewAttrib creates a new ATTRIB entity
func NewAttrib() *Attrib {
	return &Attrib{
		entity:      NewEntity(ATTRIB),
		tag:         "",
		textValue:   "",
		insertPoint: []float64{0.0, 0.0, 0.0},
		height:      1.0,
		rotation:    0.0,
		width:       1.0,
		oblique:     0.0,
		style:       "",
		mirror:      0,
		hAlign:      0, // Left
		vAlign:      0, // Baseline
		layer:       "0",
		color:       0,
		extrusion:   []float64{0.0, 0.0, 1.0},
		prompt:      "",
	}
}

// NewAttribWithValues creates an ATTRIB with specified values
func NewAttribWithValues(tag, text string, x, y, z float64) *Attrib {
	attrib := NewAttrib()
	attrib.tag = tag
	attrib.textValue = text
	attrib.insertPoint = []float64{x, y, z}
	return attrib
}

// IsEntity is for Entity interface.
func (a *Attrib) IsEntity() bool {
	return true
}

// Format writes data to formatter.
func (a *Attrib) Format(f format.Formatter) {
	a.entity.Format(f)
	f.WriteString(100, "AcDbText")
	f.WriteString(1, a.textValue)
	f.WriteString(2, a.tag)
	f.WriteString(3, a.prompt)
	f.WriteString(7, a.style)
	f.WriteString(8, a.layer)

	// Insertion point
	if len(a.insertPoint) >= 3 {
		f.WriteFloat(10, a.insertPoint[0])
		f.WriteFloat(20, a.insertPoint[1])
		f.WriteFloat(30, a.insertPoint[2])
	} else if len(a.insertPoint) >= 2 {
		f.WriteFloat(10, a.insertPoint[0])
		f.WriteFloat(20, a.insertPoint[1])
		f.WriteFloat(30, 0.0)
	}

	// Text properties
	if a.height != 0.0 {
		f.WriteFloat(40, a.height)
	}
	if a.width != 1.0 {
		f.WriteFloat(41, a.width)
	}
	if a.rotation != 0.0 {
		f.WriteFloat(50, a.rotation)
	}
	if a.oblique != 0.0 {
		f.WriteFloat(51, a.oblique)
	}

	// Justification
	if a.hAlign != 0 || a.vAlign != 0 {
		f.WriteInt(72, a.hAlign)
		f.WriteInt(74, a.vAlign)
	}

	// Mirror flag
	if a.mirror != 0 {
		f.WriteInt(71, a.mirror)
	}

	// Color
	if a.color != 0 {
		f.WriteInt(62, a.color)
	}

	// Extrusion direction
	if len(a.extrusion) >= 3 && (a.extrusion[0] != 0.0 || a.extrusion[1] != 0.0 || a.extrusion[2] != 1.0) {
		f.WriteFloat(210, a.extrusion[0])
		f.WriteFloat(220, a.extrusion[1])
		f.WriteFloat(230, a.extrusion[2])
	}
}

// BBox returns bounding box
func (a *Attrib) BBox() ([]float64, []float64) {
	mins := make([]float64, 3)
	maxs := make([]float64, 3)

	// Simplified bbox based on insertion point
	// In a complete implementation, we'd calculate actual text bounds
	if len(a.insertPoint) >= 3 {
		copy(mins, a.insertPoint)
		copy(maxs, a.insertPoint)
	}
	return mins, maxs
}

// SetTag sets the attribute tag
func (a *Attrib) SetTag(tag string) {
	a.tag = tag
}

// SetText sets the attribute text value
func (a *Attrib) SetText(text string) {
	a.textValue = text
}

// SetInsertPoint sets the insertion point
func (a *Attrib) SetInsertPoint(x, y, z float64) {
	a.insertPoint = []float64{x, y, z}
}

// SetHeight sets the text height
func (a *Attrib) SetHeight(height float64) {
	a.height = height
}

// SetRotation sets the rotation angle in degrees
func (a *Attrib) SetRotation(angle float64) {
	a.rotation = angle
}

// SetStyle sets the text style
func (a *Attrib) SetStyle(style string) {
	a.style = style
}

// SetJustification sets horizontal and vertical justification
func (a *Attrib) SetJustification(hAlign, vAlign int) {
	a.hAlign = hAlign
	a.vAlign = vAlign
}

// SetColor sets the color (implements Entity interface)
func (a *Attrib) SetColor(color color.ColorNumber) {
	a.entity.color = color
}

// GetTag returns the attribute tag
func (a *Attrib) GetTag() string {
	return a.tag
}

// GetText returns the attribute text value
func (a *Attrib) GetText() string {
	return a.textValue
}

// GetInsertPoint returns the insertion point
func (a *Attrib) GetInsertPoint() []float64 {
	return a.insertPoint
}

// GetHeight returns the text height
func (a *Attrib) GetHeight() float64 {
	return a.height
}

// GetRotation returns the rotation angle
func (a *Attrib) GetRotation() float64 {
	return a.rotation
}

// GetStyle returns the text style
func (a *Attrib) GetStyle() string {
	return a.style
}

// Move translates the attribute by specified offset
func (a *Attrib) Move(dx, dy, dz float64) {
	if len(a.insertPoint) >= 3 {
		a.insertPoint[0] += dx
		a.insertPoint[1] += dy
		a.insertPoint[2] += dz
	}
}

// String returns string representation
func (a *Attrib) String() string {
	return fmt.Sprintf("Attrib{Tag: '%s', Value: '%s', Point: [%.3f,%.3f,%.3f], Height: %.3f}",
		a.tag, a.textValue,
		a.insertPoint[0], a.insertPoint[1], a.insertPoint[2],
		a.height)
}

// Helper function to copy slice
func copy(dst, src []float64) {
	if len(dst) >= len(src) {
		for i := 0; i < len(src); i++ {
			dst[i] = src[i]
		}
	}
}
