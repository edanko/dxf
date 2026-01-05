package entity

import (
	"math"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/format"
	dxfmath "github.com/edanko/dxf/math"
)

// MPolygonEntity represents a DXF MPOLYGON entity
type MPolygonEntity struct {
	*entity
	dxf      *MPolygonDXF
	paths    *BoundaryPaths
	pattern  *Pattern
	gradient *Gradient
	seeds    []dxfmath.Vec2
}

// MPolygonDXF contains the DXF-specific MPOLYGON data
type MPolygonDXF struct {
	Version           int
	Elevation         dxfmath.Vec3
	Extrusion         dxfmath.Vec3
	PatternName       string
	FillColor         color.ColorNumber
	SolidFill         int
	PatternType       int
	PatternAngle      float64
	PatternScale      float64
	AnnotatedBoundary int
	PatternDouble     int
	PixelSize         float64
	OffsetVector      dxfmath.Vec2
	DegeneratedLoops  int
}

// MPolygon constants
const (
	// Hatch types
	HATCH_TYPE_USER_DEFINED = 0
	HATCH_TYPE_PREDEFINED   = 1
	HATCH_TYPE_CUSTOM       = 2

	// Boundary path types
	BOUNDARY_PATH_DEFAULT   = 0
	BOUNDARY_PATH_EXTERNAL  = 1
	BOUNDARY_PATH_POLYLINE  = 2
	BOUNDARY_PATH_DERIVED   = 4
	BOUNDARY_PATH_TEXTBOX   = 8
	BOUNDARY_PATH_OUTERMOST = 16
)

// NewMPolygon creates a new MPOLYGON entity
func NewMPolygon() *MPolygonEntity {
	m := &MPolygonEntity{
		entity: NewEntity(MPOLYGON_TYPE),
		dxf: &MPolygonDXF{
			Version:           1,
			Elevation:         dxfmath.NewVec3(0, 0, 0),
			Extrusion:         dxfmath.NewVec3(0, 0, 1),
			PatternName:       "",
			FillColor:         color.ByLayer,
			SolidFill:         0,
			PatternType:       HATCH_TYPE_PREDEFINED,
			PatternAngle:      0,
			PatternScale:      1.0,
			AnnotatedBoundary: 0,
			PatternDouble:     0,
			PixelSize:         0,
			OffsetVector:      dxfmath.NewVec2(0, 0),
			DegeneratedLoops:  0,
		},
		paths: NewBoundaryPaths(),
		seeds: make([]dxfmath.Vec2, 0),
	}
	return m
}

// SetElevation sets the elevation of the MPOLYGON
func (m *MPolygonEntity) SetElevation(elevation dxfmath.Vec3) {
	m.dxf.Elevation = elevation
}

// GetElevation returns the elevation of the MPOLYGON
func (m *MPolygonEntity) GetElevation() dxfmath.Vec3 {
	return m.dxf.Elevation
}

// SetExtrusion sets the extrusion direction
func (m *MPolygonEntity) SetExtrusion(extrusion dxfmath.Vec3) {
	m.dxf.Extrusion = extrusion
}

// GetExtrusion returns the extrusion direction
func (m *MPolygonEntity) GetExtrusion() dxfmath.Vec3 {
	return m.dxf.Extrusion
}

// SetSolidFill sets solid fill with optional RGB color
func (m *MPolygonEntity) SetSolidFill(fillColor color.ColorNumber, rgb *RGB) {
	m.gradient = nil
	m.pattern = nil
	m.dxf.SolidFill = 1
	m.dxf.FillColor = fillColor
	m.dxf.PatternName = "SOLID"
	m.dxf.PatternType = HATCH_TYPE_PREDEFINED

	if rgb != nil {
		m.setSolidRGBGradient(*rgb)
	}
}

// SetPatternFill sets pattern fill
func (m *MPolygonEntity) SetPatternFill(patternName string, fillColor color.ColorNumber, scale float64) {
	m.gradient = nil
	m.dxf.SolidFill = 0
	m.dxf.PatternName = patternName
	m.dxf.FillColor = fillColor
	m.dxf.PatternScale = scale
	m.dxf.PatternAngle = 0
	m.dxf.PatternDouble = 0
}

// SetGradientFill sets gradient fill
func (m *MPolygonEntity) SetGradientFill(color1, color2 RGB, rotation float64, gradientType string) {
	m.pattern = nil
	m.dxf.SolidFill = 1
	m.dxf.PatternName = "SOLID"

	m.gradient = &Gradient{
		Kind:           1,
		NumberOfColors: 2,
		Color1:         color1,
		Color2:         color2,
		Rotation:       rotation,
		Name:           gradientType,
	}
}

// AddBoundaryPath adds a boundary path
func (m *MPolygonEntity) AddBoundaryPath(pathType int, isClosed bool) *MPolygonBoundaryPath {
	return m.paths.AddPath(pathType, isClosed)
}

// GetBoundaryPaths returns the boundary paths
func (m *MPolygonEntity) GetBoundaryPaths() *BoundaryPaths {
	return m.paths
}

// SetPatternScale sets the pattern scale
func (m *MPolygonEntity) SetPatternScale(scale float64) {
	m.dxf.PatternScale = scale
}

// SetPatternAngle sets the pattern rotation angle
func (m *MPolygonEntity) SetPatternAngle(angle float64) {
	m.dxf.PatternAngle = angle
}

// SetPatternDouble sets whether to double the pattern
func (m *MPolygonEntity) SetPatternDouble(double bool) {
	if double {
		m.dxf.PatternDouble = 1
	} else {
		m.dxf.PatternDouble = 0
	}
}

// SetOffsetVector sets the offset vector
func (m *MPolygonEntity) SetOffsetVector(offset dxfmath.Vec2) {
	m.dxf.OffsetVector = offset
}

// AddVertex adds a vertex to the current boundary path
func (m *MPolygonEntity) AddVertex(x, y, bulge float64) {
	if m.paths.PathCount() == 0 {
		// Create default path if none exists
		m.AddBoundaryPath(BOUNDARY_PATH_DEFAULT, true)
	}

	currentPath := m.paths.GetPath(m.paths.PathCount() - 1)
	currentPath.AddVertex(x, y, bulge)
}

// AddRectangularBoundary adds a rectangular boundary
func (m *MPolygonEntity) AddRectangularBoundary(x, y, width, height float64, isExternal bool) {
	pathType := BOUNDARY_PATH_DEFAULT
	if isExternal {
		pathType = BOUNDARY_PATH_EXTERNAL
	}

	path := m.AddBoundaryPath(pathType, true)

	// Add rectangle vertices
	path.AddVertex(x, y, 0)              // Bottom-left
	path.AddVertex(x+width, y, 0)        // Bottom-right
	path.AddVertex(x+width, y+height, 0) // Top-right
	path.AddVertex(x, y+height, 0)       // Top-left
}

// AddCircularBoundary adds a circular boundary
func (m *MPolygonEntity) AddCircularBoundary(cx, cy, radius float64, segments int, isExternal bool) {
	if segments < 3 {
		segments = 32 // Default segments for smooth circle
	}

	pathType := BOUNDARY_PATH_DEFAULT
	if isExternal {
		pathType = BOUNDARY_PATH_EXTERNAL
	}

	path := m.AddBoundaryPath(pathType, true)

	// Add circle vertices
	for i := 0; i <= segments; i++ {
		angle := 2 * math.Pi * float64(i) / float64(segments)
		x := cx + radius*math.Cos(angle)
		y := cy + radius*math.Sin(angle)
		path.AddVertex(x, y, 0) // Bulge 0 for straight line segments
	}
}

// Format writes the MPOLYGON entity to the formatter
func (m *MPolygonEntity) Format(f format.Formatter) {
	// Base entity format
	m.entity.Format(f)

	// MPOLYGON specific tags
	f.WriteString(100, "AcDbMPolygon")

	// Basic properties
	f.WriteInt(70, m.dxf.Version)
	f.WriteFloat(10, m.dxf.Elevation.X())
	f.WriteFloat(20, m.dxf.Elevation.Y())
	f.WriteFloat(30, m.dxf.Elevation.Z())
	f.WriteFloat(210, m.dxf.Extrusion.X())
	f.WriteFloat(220, m.dxf.Extrusion.Y())
	f.WriteFloat(230, m.dxf.Extrusion.Z())
	f.WriteString(2, m.dxf.PatternName)
	f.WriteInt(71, m.dxf.SolidFill)

	// Boundary paths - CRITICAL: Must use MPOLYGON order
	m.paths.FormatMPolygon(f)

	f.WriteInt(76, m.dxf.PatternType)

	// Pattern properties
	if m.dxf.SolidFill == 0 {
		f.WriteFloat(52, m.dxf.PatternAngle)
		f.WriteFloat(41, m.dxf.PatternScale)
		f.WriteInt(77, m.dxf.PatternDouble)
	}

	f.WriteInt(73, m.dxf.AnnotatedBoundary)
	if m.dxf.PixelSize != 0 {
		f.WriteFloat(47, m.dxf.PixelSize)
	}

	// Pattern definition for non-solid fills
	if m.dxf.SolidFill == 0 {
		if m.pattern != nil {
			m.pattern.Format(f, true)
		} else {
			f.WriteInt(78, 0) // Required pattern length
		}
	}

	// Fill color (R2004+)
	// TODO: Add version checking when format supports it
	f.WriteInt(63, int(m.dxf.FillColor))

	// Offset and other properties
	f.WriteFloat(11, m.dxf.OffsetVector.X())
	f.WriteFloat(21, m.dxf.OffsetVector.Y())
	if m.dxf.DegeneratedLoops != 0 {
		f.WriteInt(99, m.dxf.DegeneratedLoops)
	}

	// Gradient data (if present)
	if m.gradient != nil {
		m.gradient.Format(f)
	}
}

// BBox returns the bounding box of the MPOLYGON
func (m *MPolygonEntity) BBox() ([]float64, []float64) {
	if m.paths.PathCount() == 0 {
		return []float64{0, 0, 0}, []float64{0, 0, 0}
	}

	// Calculate bounding box from all vertices in all paths
	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := -math.MaxFloat64, -math.MaxFloat64

	for i := 0; i < m.paths.PathCount(); i++ {
		path := m.paths.GetPath(i)
		for j := 0; j < path.VertexCount(); j++ {
			vertex := path.GetVertex(j)
			x, y := vertex.X, vertex.Y

			minX = math.Min(minX, x)
			minY = math.Min(minY, y)
			maxX = math.Max(maxX, x)
			maxY = math.Max(maxY, y)
		}
	}

	// Include elevation in Z coordinates
	z := m.dxf.Elevation.Z()

	return []float64{minX, minY, z}, []float64{maxX, maxY, z}
}

// IsEntity satisfies the Entity interface
func (m *MPolygonEntity) IsEntity() bool {
	return true
}

// setSolidRGBGradient sets up a gradient from solid RGB color
func (m *MPolygonEntity) setSolidRGBGradient(rgb RGB) {
	m.gradient = &Gradient{
		Kind:           1,
		NumberOfColors: 1,
		Color1:         rgb,
		Color2:         rgb, // Same color for single-color gradient
		OneColor:       1,
		Rotation:       0,
		Centered:       0,
		Tint:           0,
		Name:           "LINEAR",
	}
}

// Clone creates a copy of the MPOLYGON entity
func (m *MPolygonEntity) Clone() *MPolygonEntity {
	// Deep copy of DXF data
	dxfCopy := *m.dxf

	// Clone boundary paths
	pathsCopy := m.paths.Clone()

	// Clone gradient if present
	var gradientCopy *Gradient
	if m.gradient != nil {
		gradientCopy = m.gradient.Clone()
	}

	// Clone pattern if present
	var patternCopy *Pattern
	if m.pattern != nil {
		patternCopy = m.pattern.Clone()
	}

	// Copy seeds
	seedsCopy := make([]dxfmath.Vec2, len(m.seeds))
	for i, seed := range m.seeds {
		seedsCopy[i] = seed
	}

	return &MPolygonEntity{
		entity:   m.entity, // TODO: Implement Clone method for entity
		dxf:      &dxfCopy,
		paths:    pathsCopy,
		pattern:  patternCopy,
		gradient: gradientCopy,
		seeds:    seedsCopy,
	}
}
