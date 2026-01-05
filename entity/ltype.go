package entity

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/math"
)

// LinetypePattern represents the dash-gap pattern for a linetype
type LinetypePattern struct {
	TotalLength float64   // 40 - total pattern length
	Elements    []float64 // 49 - dash/gap/point lengths (positive=dash, negative=gap, 0=point)
	NumDashes   int       // 73 - number of dash elements
	Complex     bool      // 340 - handle to style (complex linetype)
	StyleHandle string    // 340
	Scale       float64   // 46 - text scale (complex)
	Rotation    float64   // 50 - text rotation (complex)
	SFlags      int       // 44 - sflags (complex)
	XOffset     float64   // 45 - x offset (complex)
	YOffset     float64   // 9 - y offset (complex)
}

// Linetype represents a LINETYPE table entry
type Linetype struct {
	*entity
	Name        string          // 2 - linetype name
	Description string          // 3 - description
	Flags       int             // 70 - flags
	Pattern     LinetypePattern // linetype pattern
}

// NewLinetype creates a new Linetype entity
func NewLinetype() *Linetype {
	l := &Linetype{
		entity:      NewEntity(LINETYPE),
		Name:        "",
		Description: "",
		Flags:       0,
		Pattern: LinetypePattern{
			TotalLength: 0,
			Elements:    []float64{},
			NumDashes:   0,
			Complex:     false,
			StyleHandle: "",
			Scale:       1.0,
			Rotation:    0.0,
			SFlags:      0,
			XOffset:     0.0,
			YOffset:     0.0,
		},
	}
	return l
}

// NewLinetypeSimple creates a simple dashed linetype
func NewLinetypeSimple(name, desc string, pattern []float64) *Linetype {
	l := NewLinetype()
	l.Name = name
	l.Description = desc
	if len(pattern) > 0 {
		// Calculate total length
		l.Pattern.Elements = pattern
		l.Pattern.TotalLength = 0
		for _, e := range pattern {
			l.Pattern.TotalLength += e
		}
		l.Pattern.NumDashes = len(pattern)
	}
	return l
}

// NewContinuousLinetype creates a continuous linetype
func NewContinuousLinetype(name, desc string) *Linetype {
	l := NewLinetype()
	l.Name = name
	l.Description = desc
	l.Pattern.TotalLength = 0
	l.Pattern.Elements = []float64{}
	l.Pattern.NumDashes = 0
	return l
}

// IsEntity is for Entity interface.
func (l *Linetype) IsEntity() bool {
	return true
}

// Format writes data to formatter.
func (l *Linetype) Format(f format.Formatter) {
	l.entity.Format(f)
	f.WriteString(100, "AcDbLinetypeTableRecord")
	f.WriteString(100, "AcDbLinetypeTableRecord")
	f.WriteString(2, l.Name)
	f.WriteString(3, l.Description)
	f.WriteInt(70, l.Flags)

	if l.Pattern.Complex {
		// Complex linetype with text/shape
		f.WriteString(340, l.Pattern.StyleHandle)
		f.WriteFloat(46, l.Pattern.Scale)
		f.WriteFloat(50, l.Pattern.Rotation)
		f.WriteInt(44, l.Pattern.SFlags)
		f.WriteFloat(45, l.Pattern.XOffset)
		for i, elem := range l.Pattern.Elements {
			f.WriteFloat(49, elem)
			if i < len(l.Pattern.Elements) {
				// Complex elements may have text
			}
		}
	} else if len(l.Pattern.Elements) > 0 {
		// Simple linetype
		f.WriteInt(72, 65) // Alignment code
		f.WriteInt(73, l.Pattern.NumDashes)
		f.WriteFloat(40, l.Pattern.TotalLength)
		for _, elem := range l.Pattern.Elements {
			f.WriteFloat(49, elem)
		}
	}
}

// BBox returns the bounding box of the entity
func (l *Linetype) BBox() ([]float64, []float64) {
	// Linetype is a table entry, return empty bounding box
	return []float64{0, 0, 0}, []float64{0, 0, 0}
}

// Transform applies a transformation matrix to the Linetype
func (l *Linetype) Transform(m *math.Matrix44) error {
	// Linetype doesn't have geometric data to transform
	return nil
}

// Copy creates a deep copy of the Linetype entity
func (l *Linetype) Copy() Entity {
	lt := NewLinetype()
	lt.entity = l.entity
	lt.Name = l.Name
	lt.Description = l.Description
	lt.Flags = l.Flags
	lt.Pattern.TotalLength = l.Pattern.TotalLength
	lt.Pattern.Elements = make([]float64, len(l.Pattern.Elements))
	copy(lt.Pattern.Elements, l.Pattern.Elements)
	lt.Pattern.NumDashes = l.Pattern.NumDashes
	lt.Pattern.Complex = l.Pattern.Complex
	lt.Pattern.StyleHandle = l.Pattern.StyleHandle
	lt.Pattern.Scale = l.Pattern.Scale
	lt.Pattern.Rotation = l.Pattern.Rotation
	lt.Pattern.SFlags = l.Pattern.SFlags
	lt.Pattern.XOffset = l.Pattern.XOffset
	lt.Pattern.YOffset = l.Pattern.YOffset
	return lt
}

// Validate validates the Linetype entity
func (l *Linetype) Validate() error {
	if l.Pattern.TotalLength < 0 {
		l.Pattern.TotalLength = 0
	}
	if l.Pattern.NumDashes < 0 {
		l.Pattern.NumDashes = 0
	}
	if l.Pattern.Scale < 0 {
		l.Pattern.Scale = 1.0
	}
	return nil
}

// IsContinuous returns true if the linetype is continuous (no dash pattern)
func (l *Linetype) IsContinuous() bool {
	return len(l.Pattern.Elements) == 0
}

// DashLengths returns the dash and gap lengths
func (l *Linetype) DashLengths() (dashes, gaps []float64) {
	dashes = []float64{}
	gaps = []float64{}
	for _, elem := range l.Pattern.Elements {
		if elem > 0 {
			dashes = append(dashes, elem)
		} else if elem < 0 {
			gaps = append(gaps, -elem)
		}
		// elem == 0 represents a point
	}
	return
}

// SetDashPattern sets the dash-gap pattern
func (l *Linetype) SetDashPattern(pattern []float64) {
	l.Pattern.Elements = pattern
	l.Pattern.TotalLength = 0
	for _, e := range pattern {
		l.Pattern.TotalLength += e
	}
	l.Pattern.NumDashes = len(pattern)
}
