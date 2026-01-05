package entity

import (
	"fmt"
	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/table"
)

// MLine style flags
const (
	MLineStyleJustificationTop     = 0
	MLineStyleJustificationZero    = 1
	MLineStyleJustificationBottom  = 2
	MLineStyleJustificationByStyle = 3
)

// MLine vertex flags
const (
	MLineVertexFlagHasBulge  = 1
	MLineVertexFlagHasMiter  = 2
	MLineVertexFlagHasVertex = 4
)

// MLine vertex data
type MLineVertex struct {
	Point          []float64 // 10,20,30 - Vertex position
	Bulge          float64   // 42 - Bulge value for arc segments
	MiterDirection []float64 // 43 - Miter direction vector
	VertexFlags    int       // 70 - Vertex flags
	StartWidth     float64   // 40 - Start width at this vertex
	EndWidth       float64   // 41 - End width at this vertex
	HasVertex      bool      // Has vertex point
	HasBulge       bool      // Has bulge value
	HasMiter       bool      // Has miter direction
}

// MLine style definition
type MLineStyle struct {
	Name          string            // Style name
	Description   string            // Style description
	Justification int               // 70 - Justification type
	Scale         float64           // 40 - Style scale
	Flags         int               // 70 - Style flags
	FillColor     color.ColorNumber // 62 - Fill color
	StartAngle    float64           // 50 - Start angle
	EndAngle      float64           // 51 - End angle
	Elements      []MLineElement    // Line elements
}

// MLine element definition
type MLineElement struct {
	Offset   float64           // 49 - Element offset from centerline
	Color    color.ColorNumber // 62 - Element color
	LineType string            // 6 - Line type name
}

// MLine represents MLINE entity
type MLine struct {
	*entity

	// Basic properties
	StyleName     string  // 3 - MLINE style name
	Justification int     // 72 - Justification type
	Scale         float64 // 40 - Scale factor
	Flags         int     // 70 - MLine flags

	// Vertex data
	Vertices []MLineVertex // Vertex list
	Closed   bool          // 73 - Closed flag (1=closed)
	HasFill  bool          // 73 - Has fill flag

	// Style reference
	MLineStyle *MLineStyle // Style definition (for reference)

	// Color and visibility
	Color     color.ColorNumber // 62 - Main color number
	LayerName string            // 8 - Layer name (field to avoid conflict with Layer() method)
}

// NewMLine creates a new MLINE entity
func NewMLine() *MLine {
	return &MLine{
		entity:        NewEntity(MLINE),
		StyleName:     "STANDARD",
		Justification: MLineStyleJustificationTop,
		Scale:         1.0,
		Flags:         0,
		Vertices:      make([]MLineVertex, 0),
		Closed:        false,
		HasFill:       false,
		MLineStyle:    nil,
		Color:         0,
		LayerName:     "0",
	}
}

// NewMLineWithStyle creates a new MLINE with specified style
func NewMLineWithStyle(styleName string, justification int, scale float64) *MLine {
	mline := NewMLine()
	mline.StyleName = styleName
	mline.Justification = justification
	mline.Scale = scale
	return mline
}

// IsEntity is for Entity interface
func (m *MLine) IsEntity() bool {
	return true
}

// AddVertex adds a vertex to the MLINE
func (m *MLine) AddVertex(x, y, z float64, options ...func(*MLineVertex)) {
	vertex := MLineVertex{
		Point:          []float64{x, y, z},
		Bulge:          0.0,
		MiterDirection: []float64{0.0, 0.0, 0.0},
		VertexFlags:    0,
		StartWidth:     0.0,
		EndWidth:       0.0,
		HasVertex:      true,
		HasBulge:       false,
		HasMiter:       false,
	}

	// Apply optional configuration
	for _, option := range options {
		option(&vertex)
	}

	m.Vertices = append(m.Vertices, vertex)
}

// AddVertexWithWidth adds a vertex with start and end widths
func (m *MLine) AddVertexWithWidth(x, y, z, startWidth, endWidth float64, options ...func(*MLineVertex)) {
	m.AddVertex(x, y, z, func(v *MLineVertex) {
		v.StartWidth = startWidth
		v.EndWidth = endWidth
		// Apply additional options after setting widths
		for _, option := range options {
			option(v)
		}
	})
}

// AddVertexWithBulge adds a vertex with bulge for arc segments
func (m *MLine) AddVertexWithBulge(x, y, z, bulge float64, options ...func(*MLineVertex)) {
	m.AddVertex(x, y, z, func(v *MLineVertex) {
		v.Bulge = bulge
		v.HasBulge = true
		v.VertexFlags |= MLineVertexFlagHasBulge
		// Apply additional options after setting bulge
		for _, option := range options {
			option(v)
		}
	})
}

// AddVertexWithMiter adds a vertex with miter direction
func (m *MLine) AddVertexWithMiter(x, y, z, miterX, miterY, miterZ float64, options ...func(*MLineVertex)) {
	m.AddVertex(x, y, z, func(v *MLineVertex) {
		v.MiterDirection = []float64{miterX, miterY, miterZ}
		v.HasMiter = true
		v.VertexFlags |= MLineVertexFlagHasMiter
		// Apply additional options after setting miter
		for _, option := range options {
			option(v)
		}
	})
}

// SetJustification sets MLINE justification
func (m *MLine) SetJustification(justification int) {
	m.Justification = justification
}

// SetScale sets MLINE scale factor
func (m *MLine) SetScale(scale float64) {
	m.Scale = scale
}

// SetClosed sets MLINE closed flag
func (m *MLine) SetClosed(closed bool) {
	m.Closed = closed
}

// SetFill sets MLINE fill flag
func (m *MLine) SetFill(hasFill bool) {
	m.HasFill = hasFill
}

// SetStyleName sets the MLINE style name
func (m *MLine) SetStyleName(styleName string) {
	m.StyleName = styleName
}

// SetColor sets MLINE color
func (m *MLine) SetColor(c color.ColorNumber) {
	m.Color = c
}

// Layer returns MLINE layer
func (m *MLine) Layer() *table.Layer {
	// Return nil for now - in full implementation would lookup actual layer
	return nil
}

// SetLayer sets MLINE layer
func (m *MLine) SetLayer(layer *table.Layer) {
	// Store layer name for now
	if layer != nil {
		m.LayerName = layer.Name()
	}
}

// SetStyle sets the complete MLINE style definition
func (m *MLine) SetStyle(style *MLineStyle) {
	m.MLineStyle = style
	if style.Name != "" {
		m.StyleName = style.Name
	}
	if style.Scale != 0.0 {
		m.Scale = style.Scale
	}
	if style.Justification != 0 {
		m.Justification = style.Justification
	}
}

// Format writes data to formatter
func (m *MLine) Format(f format.Formatter) {
	m.entity.Format(f)
	f.WriteString(100, "AcDbMLine")

	// Basic properties
	f.WriteString(2, m.StyleName)
	f.WriteInt(72, m.Justification)
	f.WriteFloat(40, m.Scale)
	f.WriteInt(70, m.Flags)

	// Vertex count
	f.WriteInt(73, len(m.Vertices))

	// Write each vertex
	for _, vertex := range m.Vertices {
		// Vertex position
		if vertex.HasVertex {
			f.WriteFloat(10, vertex.Point[0])
			f.WriteFloat(20, vertex.Point[1])
			f.WriteFloat(30, vertex.Point[2])
		}

		// Vertex direction (miter)
		if vertex.HasMiter {
			f.WriteFloat(43, vertex.MiterDirection[0])
			f.WriteFloat(44, vertex.MiterDirection[1])
			f.WriteFloat(45, vertex.MiterDirection[2])
		}

		// Vertex flags
		if vertex.VertexFlags != 0 {
			f.WriteInt(71, vertex.VertexFlags)
		}

		// Start and end widths
		if vertex.StartWidth != 0.0 {
			f.WriteFloat(40, vertex.StartWidth)
		}
		if vertex.EndWidth != 0.0 {
			f.WriteFloat(41, vertex.EndWidth)
		}

		// Bulge value
		if vertex.HasBulge {
			f.WriteFloat(42, vertex.Bulge)
		}
	}

	// MLINE properties flags
	flags := 0
	if m.Closed {
		flags |= 1
	}
	if m.HasFill {
		flags |= 2
	}
	if flags != 0 {
		f.WriteInt(73, flags)
	}

	// Color and layer
	if int(m.Color) != 0 {
		f.WriteInt(62, int(m.Color))
	}
	if m.LayerName != "0" {
		f.WriteString(8, m.LayerName)
	}

	// Style-specific properties
	if m.MLineStyle != nil {
		if m.MLineStyle.FillColor != 0 {
			f.WriteInt(62, int(m.MLineStyle.FillColor))
		}
		if m.MLineStyle.StartAngle != 0.0 {
			f.WriteFloat(50, m.MLineStyle.StartAngle)
		}
		if m.MLineStyle.EndAngle != 0.0 {
			f.WriteFloat(51, m.MLineStyle.EndAngle)
		}

		// Style elements
		if len(m.MLineStyle.Elements) > 0 {
			f.WriteInt(73, len(m.MLineStyle.Elements))
			for _, element := range m.MLineStyle.Elements {
				if element.Offset != 0.0 {
					f.WriteFloat(49, element.Offset)
				}
				if int(element.Color) != 0 {
					f.WriteInt(62, int(element.Color))
				}
				if element.LineType != "" {
					f.WriteString(6, element.LineType)
				}
			}
		}
	}
}

// BBox returns bounding box
func (m *MLine) BBox() ([]float64, []float64) {
	mins := make([]float64, 3)
	maxs := make([]float64, 3)

	// Initialize with large/small values
	for i := 0; i < 3; i++ {
		mins[i] = 1e100
		maxs[i] = -1e100
	}

	// Calculate bbox from all vertices
	for _, vertex := range m.Vertices {
		if vertex.HasVertex && len(vertex.Point) >= 3 {
			for i := 0; i < 3; i++ {
				if vertex.Point[i] < mins[i] {
					mins[i] = vertex.Point[i]
				}
				if vertex.Point[i] > maxs[i] {
					maxs[i] = vertex.Point[i]
				}
			}
		}
	}

	return mins, maxs
}

// Move translates the MLINE by specified offset
func (m *MLine) Move(dx, dy, dz float64) {
	for i := range m.Vertices {
		if m.Vertices[i].HasVertex && len(m.Vertices[i].Point) >= 3 {
			m.Vertices[i].Point[0] += dx
			m.Vertices[i].Point[1] += dy
			m.Vertices[i].Point[2] += dz
		}
	}
}

// String returns string representation
func (m *MLine) String() string {
	justStr := "Top"
	switch m.Justification {
	case MLineStyleJustificationZero:
		justStr = "Zero"
	case MLineStyleJustificationBottom:
		justStr = "Bottom"
	case MLineStyleJustificationByStyle:
		justStr = "ByStyle"
	}

	var closedStr string
	if m.Closed {
		closedStr = ", Closed"
	}

	var fillStr string
	if m.HasFill {
		fillStr = ", Filled"
	}

	return fmt.Sprintf("MLine{Style: '%s', Vertices: %d, Justification: %s%s%s, Scale: %.2f}",
		m.StyleName, len(m.Vertices), justStr, closedStr, fillStr, m.Scale)
}

// Helper functions for vertex configuration

// WithBulge sets bulge on vertex
func WithBulge(bulge float64) func(*MLineVertex) {
	return func(v *MLineVertex) {
		v.Bulge = bulge
		v.HasBulge = true
		v.VertexFlags |= MLineVertexFlagHasBulge
	}
}

// WithMiter sets miter direction on vertex
func WithMiter(miterX, miterY, miterZ float64) func(*MLineVertex) {
	return func(v *MLineVertex) {
		v.MiterDirection = []float64{miterX, miterY, miterZ}
		v.HasMiter = true
		v.VertexFlags |= MLineVertexFlagHasMiter
	}
}

// WithWidths sets start and end widths on vertex
func WithWidths(startWidth, endWidth float64) func(*MLineVertex) {
	return func(v *MLineVertex) {
		v.StartWidth = startWidth
		v.EndWidth = endWidth
	}
}

// WithNoVertex removes vertex point (for arc-only vertices)
func WithNoVertex() func(*MLineVertex) {
	return func(v *MLineVertex) {
		v.HasVertex = false
		v.VertexFlags &= ^MLineVertexFlagHasVertex
	}
}

// CreateMLineStyle creates a new MLINE style definition
func CreateMLineStyle(name, description string, justification int, scale float64) *MLineStyle {
	return &MLineStyle{
		Name:          name,
		Description:   description,
		Justification: justification,
		Scale:         scale,
		Flags:         0,
		FillColor:     0,
		StartAngle:    0.0,
		EndAngle:      0.0,
		Elements:      make([]MLineElement, 0),
	}
}

// AddElement adds an element to MLINE style
func (s *MLineStyle) AddElement(offset float64, color color.ColorNumber, lineType string) {
	element := MLineElement{
		Offset:   offset,
		Color:    color,
		LineType: lineType,
	}
	s.Elements = append(s.Elements, element)
}

// SetFillProperties sets fill properties for MLINE style
func (s *MLineStyle) SetFillProperties(fillColor color.ColorNumber, startAngle, endAngle float64) {
	s.FillColor = fillColor
	s.StartAngle = startAngle
	s.EndAngle = endAngle
}
