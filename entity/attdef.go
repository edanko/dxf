package entity

import (
	"fmt"
	"github.com/edanko/dxf/format"
)

// Attdef represents an ATTDEF entity (attribute definition)
type Attdef struct {
	*entity
	tag         string    // 2 - Attribute tag
	prompt      string    // 3 - Prompt string
	defaultVal  string    // 1,3 - Default value/Default value
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
	flags       int       // 70 - Attribute flags
	fieldLength int       // 73 - Field length (if present)
}

// AttdefFlags represents attribute definition flags
type AttdefFlags struct {
	Invisible bool // 1 - Attribute is invisible
	Constant  bool // 2 - Attribute is constant
	Verify    bool // 4 - Attribute is verified during input
	Preset    bool // 8 - Attribute is preset (no prompt during insertion)
}

// NewAttdef creates a new ATTDEF entity
func NewAttdef() *Attdef {
	return &Attdef{
		entity:      NewEntity(ATTDEF),
		tag:         "",
		prompt:      "",
		defaultVal:  "",
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
		flags:       0,
		fieldLength: 0,
	}
}

// NewAttdefWithValues creates an ATTDEF with specified values
func NewAttdefWithValues(tag, prompt, defaultVal string, x, y, z float64) *Attdef {
	attdef := NewAttdef()
	attdef.tag = tag
	attdef.prompt = prompt
	attdef.defaultVal = defaultVal
	attdef.insertPoint = []float64{x, y, z}
	return attdef
}

// IsEntity is for Entity interface.
func (a *Attdef) IsEntity() bool {
	return true
}

// Format writes data to formatter.
func (a *Attdef) Format(f format.Formatter) {
	a.entity.Format(f)
	f.WriteString(100, "AcDbAttributeDefinition")
	f.WriteString(1, a.defaultVal)
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

	// Flags
	if a.flags != 0 {
		f.WriteInt(70, a.flags)
	}

	// Field length (extended data)
	if a.fieldLength != 0 {
		f.WriteInt(73, a.fieldLength)
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
func (a *Attdef) BBox() ([]float64, []float64) {
	mins := make([]float64, 3)
	maxs := make([]float64, 3)

	// Simplified bbox based on insertion point
	if len(a.insertPoint) >= 3 {
		copy(mins, a.insertPoint)
		copy(maxs, a.insertPoint)
	}
	return mins, maxs
}

// SetTag sets the attribute tag
func (a *Attdef) SetTag(tag string) {
	a.tag = tag
}

// SetPrompt sets the prompt string
func (a *Attdef) SetPrompt(prompt string) {
	a.prompt = prompt
}

// SetDefault sets the default value
func (a *Attdef) SetDefault(defaultVal string) {
	a.defaultVal = defaultVal
}

// SetInsertPoint sets the insertion point
func (a *Attdef) SetInsertPoint(x, y, z float64) {
	a.insertPoint = []float64{x, y, z}
}

// SetHeight sets the text height
func (a *Attdef) SetHeight(height float64) {
	a.height = height
}

// SetRotation sets the rotation angle in degrees
func (a *Attdef) SetRotation(angle float64) {
	a.rotation = angle
}

// SetStyle sets the text style
func (a *Attdef) SetStyle(style string) {
	a.style = style
}

// SetJustification sets horizontal and vertical justification
func (a *Attdef) SetJustification(hAlign, vAlign int) {
	a.hAlign = hAlign
	a.vAlign = vAlign
}

// SetFlags sets the attribute flags
func (a *Attdef) SetFlags(flags int) {
	a.flags = flags
}

// SetInvisible sets the invisible flag
func (a *Attdef) SetInvisible(invisible bool) {
	if invisible {
		a.flags |= 1
	} else {
		a.flags &^= 1
	}
}

// SetConstant sets the constant flag
func (a *Attdef) SetConstant(constant bool) {
	if constant {
		a.flags |= 2
	} else {
		a.flags &^= 2
	}
}

// SetVerify sets the verify flag
func (a *Attdef) SetVerify(verify bool) {
	if verify {
		a.flags |= 4
	} else {
		a.flags &^= 4
	}
}

// SetPreset sets the preset flag
func (a *Attdef) SetPreset(preset bool) {
	if preset {
		a.flags |= 8
	} else {
		a.flags &^= 8
	}
}

// GetFlags returns the parsed flags
func (a *Attdef) GetFlags() AttdefFlags {
	return AttdefFlags{
		Invisible: (a.flags & 1) != 0,
		Constant:  (a.flags & 2) != 0,
		Verify:    (a.flags & 4) != 0,
		Preset:    (a.flags & 8) != 0,
	}
}

// GetTag returns the attribute tag
func (a *Attdef) GetTag() string {
	return a.tag
}

// GetPrompt returns the prompt string
func (a *Attdef) GetPrompt() string {
	return a.prompt
}

// GetDefault returns the default value
func (a *Attdef) GetDefault() string {
	return a.defaultVal
}

// GetInsertPoint returns the insertion point
func (a *Attdef) GetInsertPoint() []float64 {
	return a.insertPoint
}

// GetHeight returns the text height
func (a *Attdef) GetHeight() float64 {
	return a.height
}

// GetRotation returns the rotation angle
func (a *Attdef) GetRotation() float64 {
	return a.rotation
}

// GetStyle returns the text style
func (a *Attdef) GetStyle() string {
	return a.style
}

// Move translates the attribute definition by specified offset
func (a *Attdef) Move(dx, dy, dz float64) {
	if len(a.insertPoint) >= 3 {
		a.insertPoint[0] += dx
		a.insertPoint[1] += dy
		a.insertPoint[2] += dz
	}
}

// String returns string representation
func (a *Attdef) String() string {
	flags := a.GetFlags()
	return fmt.Sprintf("ATTDEF{Tag: '%s', Prompt: '%s', Default: '%s', Point: [%.3f,%.3f,%.3f], Height: %.3f, Flags: %+v}",
		a.tag, a.prompt, a.defaultVal,
		a.insertPoint[0], a.insertPoint[1], a.insertPoint[2],
		a.height, flags)
}

// Helper function to copy slice
func copySlice(dst, src []float64) {
	if len(dst) >= len(src) {
		for i := 0; i < len(src); i++ {
			dst[i] = src[i]
		}
	}
}
