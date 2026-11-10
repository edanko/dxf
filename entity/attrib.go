package entity

import (
	"fmt"
	"strconv"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/format"
)

func itoa(v int) string { return strconv.Itoa(v) }

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
	endhandle   string    // SEQEND handle (last attribute of an INSERT list)
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
	if len(a.insertPoint) >= 3 && (a.insertPoint[0] != 0.0 || a.insertPoint[1] != 0.0 || a.insertPoint[2] != 0.0) {
		for i := range 3 {
			f.WriteFloat((i+1)*10, a.insertPoint[i])
		}
	}
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
	f.WriteString(72, itoa(a.hAlign))
	f.WriteString(74, itoa(a.vAlign))
	if a.mirror != 0 {
		f.WriteInt(71, a.mirror)
	}
	f.WriteString(1, a.textValue)
	f.WriteString(2, a.tag)
	if a.prompt != "" {
		f.WriteString(3, a.prompt)
	}
	if a.style != "" {
		f.WriteString(7, a.style)
	}
	layerName := a.layer
	if layerName == "" {
		if ln := a.LayerName(); ln != "" {
			layerName = ln
		} else if l := a.Layer(); l != nil {
			layerName = l.Name()
		}
	}
	if layerName != "" {
		f.WriteString(8, layerName)
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

// SetWidthFactor sets the text width factor (code 41).
func (a *Attrib) SetWidthFactor(w float64) { a.width = w }

// SetOblique sets the obliquity angle (code 51).
func (a *Attrib) SetOblique(o float64) { a.oblique = o }

// SetMirror sets the mirror flag (code 71).
func (a *Attrib) SetMirror(m int) { a.mirror = m }

// Justification returns horizontal and vertical justification (codes 72, 74).
func (a *Attrib) Justification() (int, int) { return a.hAlign, a.vAlign }

// SetEndHandleString records the SEQEND handle that terminates an INSERT's
// attribute list. The last attribute of the list emits it (code 5) after the
// attributes so the record sequence round-trips.
func (a *Attrib) SetEndHandleString(h string) { a.endhandle = h }

// EndHandle returns the SEQEND handle recorded on this attribute.
func (a *Attrib) EndHandle() string { return a.endhandle }

// SetPrompt sets the attribute prompt (code 3).
func (a *Attrib) SetPrompt(p string) { a.prompt = p }

// Prompt returns the attribute prompt.
func (a *Attrib) Prompt() string { return a.prompt }

// Text returns the attribute value (code 1).
func (a *Attrib) Text() string { return a.textValue }

// Tag returns the attribute tag (code 2).
func (a *Attrib) Tag() string { return a.tag }

// SetExtrusionComponent sets one component of the extrusion direction.
func (a *Attrib) SetExtrusionComponent(idx int, v float64) {
	if len(a.extrusion) < 3 {
		a.extrusion = []float64{0, 0, 1}
	}
	a.extrusion[idx] = v
}

// ExtrusionDirection returns the extrusion direction (codes 210, 220, 230).
func (a *Attrib) ExtrusionDirection() []float64 { return a.extrusion }

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

// Copy creates a deep copy of the Attrib entity
func (a *Attrib) Copy() Entity {
	attrib := NewAttrib()
	attrib.entity = a.entity
	attrib.tag = a.tag
	attrib.textValue = a.textValue
	attrib.insertPoint = []float64{a.insertPoint[0], a.insertPoint[1], a.insertPoint[2]}
	attrib.height = a.height
	attrib.rotation = a.rotation
	attrib.width = a.width
	attrib.oblique = a.oblique
	attrib.style = a.style
	attrib.mirror = a.mirror
	attrib.hAlign = a.hAlign
	attrib.vAlign = a.vAlign
	attrib.layer = a.layer
	attrib.color = a.color
	attrib.extrusion = []float64{a.extrusion[0], a.extrusion[1], a.extrusion[2]}
	attrib.prompt = a.prompt
	return attrib
}

// Helper function to copy slice
func copy(dst, src []float64) {
	if len(dst) >= len(src) {
		for i := range src {
			dst[i] = src[i]
		}
	}
}
