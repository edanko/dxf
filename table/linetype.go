package table

import (
	"math"

	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/lldxf"
)

// LineType represents LTYPE SymbolTable.
type LineType struct {
	EntryBase

	name        string // 2
	Description string // 3
	alignment   int    // 72, 65 = AutoCAD text fit
	lengths     []float64
	elements    []LineTypeElement
}

// LineTypeElement is one dash/dot/gap/shape/text element of a complex linetype.
type LineTypeElement struct {
	DashLength float64 // 49
	SegLen     int     // 74 element segment length (0 for simple dashes)
	ShapeNum   int     // 75 shape number (-1 none, -2 text)
	Rotation   float64 // 50
	XScale     float64 // 46
	TextValue  string  // 9
	XOffset    float64 // 44
	YOffset    float64 // 45
}

// NewLineType creates a new LineType.
func NewLineType(name, desc string, ls ...float64) *LineType {
	lt := new(LineType)
	lt.name = name
	lt.Description = desc
	lt.alignment = 65
	if len(ls) > 0 {
		lt.lengths = ls
	} else {
		lt.lengths = make([]float64, 0)
	}
	return lt
}

// IsSymbolTable is for SymbolTable interface.
func (lt *LineType) IsSymbolTable() bool {
	return true
}

// Format writes data to formatter.
func (lt *LineType) Format(f format.Formatter) {
	lt.FormatCommon(f, "LTYPE", "AcDbLinetypeTableRecord")
	f.WriteString(2, lt.name)
	f.WriteInt(70, lt.flags)
	f.WriteString(3, lt.Description)
	if len(lt.elements) > 0 {
		f.WriteInt(72, lt.alignment)
		f.WriteInt(73, len(lt.elements))
		f.WriteFloat(40, lt.TotalLength())
		for _, e := range lt.elements {
			f.WriteFloat(49, e.DashLength)
			if e.SegLen != 0 || e.ShapeNum != -1 || e.TextValue != "" {
				f.WriteInt(74, e.SegLen)
				f.WriteInt(75, e.ShapeNum)
				f.WriteFloat(46, e.XScale)
				f.WriteString(9, e.TextValue)
				f.WriteFloat(50, e.Rotation)
				f.WriteFloat(44, e.XOffset)
				f.WriteFloat(45, e.YOffset)
			}
		}
		return
	}
	f.WriteInt(72, lt.alignment)
	f.WriteInt(73, len(lt.lengths))
	f.WriteFloat(40, lt.TotalLength())
	for _, l := range lt.lengths {
		f.WriteFloat(49, l)
	}
	lt.FormatCommonTail(f)
}

// RetainTags records LTYPE codes the typed model does not emit.
func (lt *LineType) RetainTags(tags lldxf.Tags) {
	emitted := map[int]bool{0: true, 2: true, 3: true, 5: true, 9: true, 40: true, 44: true, 45: true, 46: true, 49: true, 50: true, 70: true, 72: true, 73: true, 74: true, 75: true, 100: true, 330: true}
	lt.RetainUnhandled(tags, emitted)
}

// Alignment returns the text-alignment code (72).
func (lt *LineType) Alignment() int { return lt.alignment }

// SetAlignment sets the text-alignment code (65 = AutoCAD fit).
func (lt *LineType) SetAlignment(v int) { lt.alignment = v }

// Elements returns complex-linetype elements; nil for simple linetypes.
func (lt *LineType) Elements() []LineTypeElement { return lt.elements }

// AddElement appends a complex-linetype element and its dash length.
func (lt *LineType) AddElement(e LineTypeElement) {
	if e.ShapeNum == -1 && e.TextValue != "" {
		e.ShapeNum = -2 // text element marker
	}
	lt.elements = append(lt.elements, e)
}

// Name returns a name of LAYER (code 2).
func (lt *LineType) Name() string {
	return lt.name
}

// SetName sets the line type name.
func (lt *LineType) SetName(name string) { lt.name = name }

// TotalLength returns total pattern length (code 40).
func (lt *LineType) TotalLength() float64 {
	sum := 0.0
	for _, l := range lt.lengths {
		sum += math.Abs(l)
	}
	for _, e := range lt.elements {
		sum += math.Abs(e.DashLength)
	}
	return sum
}

// SetLength sets pattern length (code 49).
//
//	positive value: Dash
//	0.0: Dot
//	negative value: Space
func (lt *LineType) SetLength(ls []float64) {
	lt.lengths = ls
}
