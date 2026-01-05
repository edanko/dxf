package entity

import (
	"fmt"
	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/lldxf"
	"github.com/edanko/dxf/math"
)

// Hatch pattern types
const (
	HatchPatternTypeUserDefined = 0
	HatchPatternTypePredefined  = 1
	HatchPatternTypeCustom      = 2
)

// Boundary path types
const (
	BoundaryPathTypeDefault   = 0
	BoundaryPathTypePolyline  = 2
	BoundaryPathTypeDerived   = 4
	BoundaryPathTypeTextbox   = 8
	BoundaryPathTypeOutermost = 16
)

// Hatch styles
const (
	HatchStyleNormal = 0 // Nested hatching (odd parity)
	HatchStyleOuter  = 1 // Outermost hatching
	HatchStyleIgnore = 2 // Ignore nesting style
)

// Gradient types
const (
	GradientTypeSolid     = 0
	GradientTypeLinear    = 1
	GradientTypeCylinder  = 2
	GradientTypeSpherical = 4
)

// HatchPatternLine represents a line in a hatch pattern
type HatchPatternLine struct {
	Angle           float64   // 53 - Line angle in degrees
	BasePoint       math.Vec2 // 43,44 - Origin point for line
	Offset          math.Vec2 // 45,46 - Offset for next line
	DashLengthItems []float64 // 49 - Dash pattern (positive=dash, negative=gap, 0=dot)
}

// HatchGradientSimple represents gradient fill information
type HatchGradientSimple struct {
	Kind           int       // 450 - Gradient type
	NumberOfColors int       // 453 - Number of colors (0=solid, 2=gradient)
	Color1, Color2 math.Vec3 // Colors for gradient (RGB 0-255 each component)
	ACI1, ACI2     int       // AutoCAD Color Index values
	OneColor       int       // 452 - One/two color (0=two, 1=one)
	Rotation       float64   // 460 - Gradient rotation in radians
	Centered       float64   // 461 - Gradient centering (-1.0 to 1.0)
	Tint           float64   // 462 - Color tint (0.0 to 1.0)
	Name           string    // 470 - Gradient name
}

// HatchColor represents color with ACI and RGB support
type HatchColor struct {
	ACI int       // AutoCAD Color Index (0-255, 256=BYLAYER)
	RGB math.Vec3 // True color (0-255 each component)
}

// NewHatchColor creates a new color
func NewHatchColor(aci int, rgb math.Vec3) HatchColor {
	return HatchColor{
		ACI: aci,
		RGB: rgb,
	}
}

// IsTrueColor returns true if RGB is set (not BYLAYER)
func (c HatchColor) IsTrueColor() bool {
	return c.ACI >= 0 && c.ACI <= 255
}

// IsByLayer returns true if color is BYLAYER
func (c HatchColor) IsByLayer() bool {
	return c.ACI == 256
}

// BoundaryPath represents a hatch boundary path
type BoundaryPath interface {
	Type() int
	PathTypeFlags() int
	SourceBoundaryObjects() []string
	ExportDXF(writer lldxf.TagWriter, ocs math.OCS, elevation float64) error
	Transform(ocs math.OCS, elevation float64) BoundaryPath
	IsValid() bool
}

// PolylinePath represents a polyline boundary path
type PolylinePath struct {
	pathTypeFlags int
	isClosed      bool
	vertices      []Vertex3D // 2D vertices with optional bulge
	sourceHandles []string
}

// Type returns the boundary path type
func (p *PolylinePath) Type() int {
	return BoundaryPathTypePolyline
}

// PathTypeFlags returns the boundary path flags
func (p *PolylinePath) PathTypeFlags() int {
	return p.pathTypeFlags
}

// SourceBoundaryObjects returns the source object handles
func (p *PolylinePath) SourceBoundaryObjects() []string {
	return p.sourceHandles
}

// ExportDXF writes the polyline path to DXF
func (p *PolylinePath) ExportDXF(writer lldxf.TagWriter, ocs math.OCS, elevation float64) error {
	tags := lldxf.Tags{}

	// Boundary path header
	tags = tags.Add(lldxf.NewDXFTag(92, fmt.Sprintf("%d", p.Type())))
	tags = tags.Add(lldxf.NewDXFTag(93, p.PathTypeFlags()))
	tags = tags.Add(lldxf.NewDXFTag(72, 0)) // No bulge
	tags = tags.Add(lldxf.NewDXFTag(73, 0)) // Default closed flag
	tags = tags.Add(lldxf.NewDXFTag(97, 0)) // No source objects

	// Polyline vertices
	for _, vertex := range p.vertices {
		tags = tags.Add(lldxf.NewDXFTag(10, vertex.Point.X()))
		tags = tags.Add(lldxf.NewDXFTag(20, vertex.Point.Y()))
		if vertex.Point.Z() != 0 || vertex.Bulge != 0 {
			tags = tags.Add(lldxf.NewDXFTag(30, vertex.Point.Z()))
		}
		if vertex.Bulge != 0 {
			tags = tags.Add(lldxf.NewDXFTag(42, vertex.Bulge))
		}
	}

	// Source objects (if any)
	for _, handle := range p.sourceHandles {
		tags = tags.Add(lldxf.NewDXFTag(330, handle))
	}

	return writer.WriteTags(tags)
}

// Transform applies OCS transformation to the boundary path
func (p *PolylinePath) Transform(ocs math.OCS, elevation float64) BoundaryPath {
	transformed := &PolylinePath{
		pathTypeFlags: p.pathTypeFlags,
		isClosed:      p.isClosed,
		vertices:      make([]Vertex3D, len(p.vertices)),
		sourceHandles: p.sourceHandles,
	}

	// Transform vertices
	for i, vertex := range p.vertices {
		transformed.vertices[i] = Vertex3D{
			Point: ocs.ToWCS(vertex.Point),
			Bulge: vertex.Bulge,
		}
	}

	return transformed
}

// IsValid validates the polyline path
func (p *PolylinePath) IsValid() bool {
	return len(p.vertices) >= 2 // Need at least 2 vertices
}

// Vertex3D represents a 3D point with optional bulge for boundary paths
type Vertex3D struct {
	Point math.Vec3
	Bulge float64
}

// NewVertex3D creates a new 3D vertex
func NewVertex3D(x, y, z, bulge float64) Vertex3D {
	return Vertex3D{
		Point: math.NewVec3(x, y, z),
		Bulge: bulge,
	}
}

// Hatch represents a DXF HATCH entity (comprehensive implementation)
type Hatch struct {
	*entity
	patternType   int                  // 76 - Pattern type
	patternName   string               // 2 - Pattern name
	patternLines  []HatchPatternLine   // Pattern definition
	solidFill     bool                 // 70 - Solid fill flag
	style         int                  // 75 - Hatch style
	color         HatchColor           // 62 - Hatch color
	entityColor   color.ColorNumber    // For Entity interface compatibility
	transparency  float64              // 460 - Transparency (DXF 2004+)
	gradient      *HatchGradientSimple // 471 - Gradient information
	boundaryPaths []BoundaryPath       // 91-98 - Boundary paths
}

// NewHatch creates a new hatch entity
func NewHatch() *Hatch {
	return &Hatch{
		entity:      NewEntity(HATCH),
		patternType: HatchPatternTypePredefined,
		patternName: "ANSI31",
		patternLines: []HatchPatternLine{
			{
				Angle:           0,
				BasePoint:       math.NewVec2(0, 0),
				Offset:          math.NewVec2(0.125, 0),
				DashLengthItems: []float64{0.125, -0.125, 0.125, -0.125, 0.0, -0.125, 0.125, -0.125, 0.0, -0.125},
			},
		},
		solidFill:     true,
		style:         HatchStyleNormal,
		color:         NewHatchColor(1, math.NewVec3(255, 255, 255)),
		boundaryPaths: make([]BoundaryPath, 0),
	}
}

// IsEntity is for Entity interface.
func (h *Hatch) IsEntity() bool {
	return true
}

// Format writes data to formatter.
func (h *Hatch) Format(f format.Formatter) {
	h.entity.Format(f)
	f.WriteString(100, "AcDbHatch")
	f.WriteInt(70, h.style)
	f.WriteString(2, h.patternName)
	f.WriteInt(76, h.patternType)

	// Pattern lines
	for _, patternLine := range h.patternLines {
		f.WriteFloat(53, patternLine.Angle)
		f.WriteFloat(43, patternLine.BasePoint.X())
		f.WriteFloat(44, patternLine.BasePoint.Y())
		f.WriteFloat(45, patternLine.Offset.X())
		f.WriteFloat(46, patternLine.Offset.Y())
		for j, dash := range patternLine.DashLengthItems {
			f.WriteFloat(49+j, dash)
		}
	}

	// Color information
	if h.color.IsTrueColor() {
		f.WriteInt(62, h.color.ACI)
	} else if h.color.IsByLayer() {
		f.WriteInt(62, 256) // BYLAYER
	}

	// Transparency
	if h.transparency != 0 {
		f.WriteFloat(460, h.transparency)
	}

	// Boundary paths
	f.WriteInt(91, len(h.boundaryPaths))
	for _, boundaryPath := range h.boundaryPaths {
		f.WriteInt(92, boundaryPath.Type())
		f.WriteInt(93, boundaryPath.PathTypeFlags())

		// For now, simplified boundary path export - full implementation would be more complex
		if polylinePath, ok := boundaryPath.(*PolylinePath); ok {
			f.WriteInt(72, 0) // No bulge
			f.WriteInt(73, 0) // Default closed flag
			f.WriteInt(97, 0) // No source objects
			f.WriteInt(93, len(polylinePath.vertices))

			for _, vertex := range polylinePath.vertices {
				f.WriteFloat(10, vertex.Point.X())
				f.WriteFloat(20, vertex.Point.Y())
				if vertex.Point.Z() != 0 {
					f.WriteFloat(30, vertex.Point.Z())
				}
				if vertex.Bulge != 0 {
					f.WriteFloat(42, vertex.Bulge)
				}
			}
		}
	}
}

// BBox returns bounding box
func (h *Hatch) BBox() ([]float64, []float64) {
	mins := make([]float64, 3)
	maxs := make([]float64, 3)
	// Simplified bbox - could be improved by calculating actual bounds from boundary paths
	return mins, maxs
}

// SetPattern sets the hatch pattern
func (h *Hatch) SetPattern(patternType int, patternName string, patternLines []HatchPatternLine) {
	h.patternType = patternType
	h.patternName = patternName
	h.patternLines = patternLines
}

// SetSolidFill sets the solid fill flag
func (h *Hatch) SetSolidFill(solid bool) {
	h.solidFill = solid
}

// SetStyle sets the hatch style
func (h *Hatch) SetStyle(style int) {
	h.style = style
}

// SetColor sets the hatch color (implements Entity interface)
func (h *Hatch) SetColor(c color.ColorNumber) {
	h.entityColor = c
	h.color = NewHatchColor(int(c), math.NewVec3(float64(c&0xFF), float64((c>>8)&0xFF), float64((c>>16)&0xFF)))
}
func (h *Hatch) SetTransparency(transparency float64) {
	h.transparency = transparency
}

// SetGradient sets the gradient fill information
func (h *Hatch) SetGradient(gradient *HatchGradientSimple) {
	h.gradient = gradient
}

// AddBoundaryPath adds a boundary path
func (h *Hatch) AddBoundaryPath(boundaryPath BoundaryPath) {
	h.boundaryPaths = append(h.boundaryPaths, boundaryPath)
}

// ClearBoundaryPaths removes all boundary paths
func (h *Hatch) ClearBoundaryPaths() {
	h.boundaryPaths = make([]BoundaryPath, 0)
}

// GetBoundaryPaths returns the boundary paths
func (h *Hatch) GetBoundaryPaths() []BoundaryPath {
	return h.boundaryPaths
}

// IsSolid returns true if hatch is a solid fill
func (h *Hatch) IsSolid() bool {
	return h.solidFill
}

// String returns string representation
func (h *Hatch) String() string {
	return fmt.Sprintf("Hatch{Pattern: '%s', Style: %d, Solid: %v}", h.patternName, h.style, h.solidFill)
}
