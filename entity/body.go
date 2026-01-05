package entity

import (
	"math"

	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
	dxfmath "github.com/edanko/dxf/math"
)

// ACIS version constants for SAT/SAB format
const (
	ACISVersion700   = 700   // DXF R2000-R2010 (basic SAT)
	ACISVersion20800 = 20800 // DXF R2004 (SAT with ASM header)
	ACISVersion21800 = 21800 // DXF R2007 (SAT with history)
	ACISVersion22300 = 22300 // DXF R2010+ (current)
)

// ACIS entity types for SAT parsing
type ACISEntityType int

// ACIS entity types
const (
	ACISEntityTypeModel  ACISEntityType = 0
	ACISEntityTypeHeader ACISEntityType = 1
	ACISEntityTypeBody   ACISEntityType = 2
	ACISEntityTypeLump   ACISEntityType = 3
	ACISEntityTypeShell  ACISEntityType = 4
	ACISEntityTypeFace   ACISEntityType = 5
	ACISEntityTypeLoop   ACISEntityType = 6
	ACISEntityTypeCoedge ACISEntityType = 7
	ACISEntityTypeEdge   ACISEntityType = 8
	ACISEntityTypeVertex ACISEntityType = 9
)

// BodyPattern represents ACIS body pattern structure
type BodyPattern struct {
	Type int
	Name string
}

// NewBodyPattern creates a new body pattern
func NewBodyPattern(patternType int, name string) *BodyPattern {
	return &BodyPattern{
		Type: patternType,
		Name: name,
	}
}

// BodyLump represents ACIS body lump
type BodyLump struct {
	Type int
	Name string
}

// NewBodyLump creates a new body lump
func NewBodyLump(lumpType int, name string) *BodyLump {
	return &BodyLump{
		Type: lumpType,
		Name: name,
	}
}

// Region represents ACIS region (2D planar region)
type Region struct {
	*entity

	// Basic properties
	UID  string
	Name string

	// Geometry
	OuterLoop *Loop
}

// NewRegion creates a new region
func NewRegion() *Region {
	return &Region{
		entity:    NewEntity(REGION),
		UID:       "",
		Name:      "",
		OuterLoop: nil,
	}
}

// Surface represents ACIS surface
type Surface struct {
	*entity

	// Basic properties
	UID  string
	Name string

	// Surface properties
	UCount int // Number of U curves
	VCount int // Number of V curves
}

// NewSurface creates a new surface
func NewSurface() *Surface {
	return &Surface{
		entity: NewEntity(SURFACE),
		UID:    "",
		Name:   "",
		UCount: 0,
		VCount: 0,
	}
}

// Loop represents ACIS boundary loop
type Loop struct {
	*entity

	// Basic properties
	UID  string
	Name string

	// Geometry
	Vertices []*BodyVertex
	Edges    []*Edge
}

// NewLoop creates a new boundary loop
func NewLoop() *Loop {
	return &Loop{
		entity:   NewEntity(LOOP),
		UID:      "",
		Name:     "",
		Vertices: make([]*BodyVertex, 0),
		Edges:    make([]*Edge, 0),
	}
}

// Edge represents ACIS coedge
type Edge struct {
	*entity

	// Basic properties
	UID  string
	Name string

	// Geometry
	StartVertex *BodyVertex
	EndVertex   *BodyVertex
}

// NewEdge creates a new edge
func NewEdge() *Edge {
	return &Edge{
		entity:      NewEntity(COEDGE),
		UID:         "",
		Name:        "",
		StartVertex: nil,
		EndVertex:   nil,
	}
}

// BodyVertex represents ACIS vertex (renamed to avoid conflict with entity.Vertex)
type BodyVertex struct {
	*entity

	// Basic properties
	UID  string
	Name string

	// Geometry
	Point dxfmath.Vec3
}

// NewBodyVertex creates a new vertex
func NewBodyVertex() *BodyVertex {
	return &BodyVertex{
		entity: NewEntity(VERTEX),
		UID:    "",
		Name:   "",
		Point:  dxfmath.NewVec3(0, 0, 0),
	}
}

// ACIS entity represents a single entity in SAT format
type ACISEntity struct {
	Type     ACISEntityType
	ID       int
	Parent   int
	Data     interface{}
	Children []*ACISEntity
}

// Transform represents 3D transformation matrix
type ACASTransform struct {
	Matrix [16]float64 // 4x4 transformation matrix
}

// BoundingBox represents 3D bounding box
type ACISBoundingBox struct {
	Min dxfmath.Vec3
	Max dxfmath.Vec3
}

// NewACASTransform creates a new identity transform
func NewACASTransform() *ACASTransform {
	return &ACASTransform{
		Matrix: [16]float64{
			1, 0, 0, 0,
			0, 1, 0, 0,
			0, 0, 1, 0,
			0, 0, 0, 1,
		},
	}
}

// Apply applies transformation to a 3D point
func (t *ACASTransform) Apply(point dxfmath.Vec3) dxfmath.Vec3 {
	x := t.Matrix[0]*point.X() + t.Matrix[1]*point.Y() + t.Matrix[2]*point.Z() + t.Matrix[3]
	y := t.Matrix[4]*point.X() + t.Matrix[5]*point.Y() + t.Matrix[6]*point.Z() + t.Matrix[7]
	z := t.Matrix[8]*point.X() + t.Matrix[9]*point.Y() + t.Matrix[10]*point.Z() + t.Matrix[11]
	return dxfmath.NewVec3(x, y, z)
}

// NewACISBoundingBox creates a new bounding box
func NewACISBoundingBox(min, max dxfmath.Vec3) *ACISBoundingBox {
	return &ACISBoundingBox{
		Min: min,
		Max: max,
	}
}

// Contains checks if point is inside bounding box
func (bb *ACISBoundingBox) Contains(point dxfmath.Vec3) bool {
	return point.X() >= bb.Min.X() && point.X() <= bb.Max.X() &&
		point.Y() >= bb.Min.Y() && point.Y() <= bb.Max.Y() &&
		point.Z() >= bb.Min.Z() && point.Z() <= bb.Max.Z()
}

// Union merges two bounding boxes
func (bb *ACISBoundingBox) Union(other *ACISBoundingBox) *ACISBoundingBox {
	min := dxfmath.NewVec3(
		math.Min(bb.Min.X(), other.Min.X()),
		math.Min(bb.Min.Y(), other.Min.Y()),
		math.Min(bb.Min.Z(), other.Min.Z()),
	)
	max := dxfmath.NewVec3(
		math.Max(bb.Max.X(), other.Max.X()),
		math.Max(bb.Max.Y(), other.Max.Y()),
		math.Max(bb.Max.Z(), other.Max.Z()),
	)
	return NewACISBoundingBox(min, max)
}

// Body represents an ACIS 3D body entity (DXF BODY)
type Body struct {
	*entity

	// Basic properties
	Version       int            // ACIS version (700, 20800, 21800, etc.)
	Flags         int            // Body flags
	UID           string         // Unique identifier
	HistoryHandle handle.Handler // History record handle (DXF R2007+)

	// Pattern and structure
	Pattern       *BodyPattern
	Lump          *BodyLump
	TransientLump *BodyLump // Temporary/transient data

	// Transformations
	Transform        *ACASTransform
	InverseTransform *ACASTransform

	// Geometry cache
	boundingBox *ACISBoundingBox
	center      dxfmath.Vec3
	volume      float64
	surfaceArea float64

	// SAT data (if loaded from file)
	SATData    []byte
	SATVersion int

	// Sub-entities (for complex bodies)
	SubBodies []*Body
	Regions   []*Region
	Surfaces  []*Surface
}

// NewBody creates a new Body entity
func NewBody() *Body {
	return &Body{
		entity:        NewEntity(BODY),
		Version:       ACISVersion22300, // Default to latest
		Flags:         0,
		UID:           "",
		HistoryHandle: nil,
		Pattern:       NewBodyPattern(0, ""),
		Transform:     NewACASTransform(),
		boundingBox:   nil,
		center:        dxfmath.NewVec3(0, 0, 0),
		volume:        0.0,
		surfaceArea:   0.0,
	}
}

// SetVersion sets the ACIS version
func (b *Body) SetVersion(version int) {
	b.Version = version
}

// SetUID sets the unique identifier
func (b *Body) SetUID(uid string) {
	b.UID = uid
}

// SetHistoryHandle sets the history record handle
func (b *Body) SetHistoryHandle(h handle.Handler) {
	b.HistoryHandle = h
}

// SetPattern sets the body pattern
func (b *Body) SetPattern(pattern *BodyPattern) {
	b.Pattern = pattern
}

// SetLump sets the primary lump
func (b *Body) SetLump(lump *BodyLump) {
	b.Lump = lump
}

// SetTransform sets the transformation matrix
func (b *Body) SetTransform(transform *ACASTransform) {
	b.Transform = transform

	// Calculate inverse transform
	// This is a simplified inverse calculation
	// In practice, would need full 4x4 matrix inversion
	b.InverseTransform = NewACASTransform() // Placeholder
}

// GetBoundingBox calculates and returns the bounding box
func (b *Body) GetBoundingBox() *ACISBoundingBox {
	if b.boundingBox != nil {
		return b.boundingBox
	}

	// Calculate from geometry if not cached
	// This is a simplified calculation
	// In practice, would analyze all geometry
	min := dxfmath.NewVec3(-1e6, -1e6, -1e6)
	max := dxfmath.NewVec3(1e6, 1e6, 1e6)
	b.boundingBox = NewACISBoundingBox(min, max)
	return b.boundingBox
}

// GetCenter returns the geometric center
func (b *Body) GetCenter() dxfmath.Vec3 {
	if !(b.center.X() == 0 && b.center.Y() == 0 && b.center.Z() == 0) {
		// Calculate from bounding box if not set
		bb := b.GetBoundingBox()
		b.center = dxfmath.NewVec3(
			(bb.Min.X()+bb.Max.X())/2,
			(bb.Min.Y()+bb.Max.Y())/2,
			(bb.Min.Z()+bb.Max.Z())/2,
		)
	}
	return b.center
}

// GetVolume returns the calculated volume
func (b *Body) GetVolume() float64 {
	// This would be calculated from the actual geometry
	// For now, return cached value
	return b.volume
}

// SetVolume sets the volume (for manually specified volumes)
func (b *Body) SetVolume(volume float64) {
	b.volume = volume
}

// GetSurfaceArea returns the calculated surface area
func (b *Body) GetSurfaceArea() float64 {
	// This would be calculated from the actual geometry
	// For now, return cached value
	return b.surfaceArea
}

// SetSurfaceArea sets the surface area (for manually specified areas)
func (b *Body) SetSurfaceArea(area float64) {
	b.surfaceArea = area
}

// AddSubBody adds a sub-body to this body
func (b *Body) AddSubBody(subBody *Body) {
	if b.SubBodies == nil {
		b.SubBodies = make([]*Body, 0)
	}
	b.SubBodies = append(b.SubBodies, subBody)
}

// AddRegion adds a region to this body
func (b *Body) AddRegion(region *Region) {
	if b.Regions == nil {
		b.Regions = make([]*Region, 0)
	}
	b.Regions = append(b.Regions, region)
}

// AddSurface adds a surface to this body
func (b *Body) AddSurface(surface *Surface) {
	if b.Surfaces == nil {
		b.Surfaces = make([]*Surface, 0)
	}
	b.Surfaces = append(b.Surfaces, surface)
}

// IsEntity is for Entity interface
func (b *Body) IsEntity() bool {
	return true
}

// Format writes data to formatter
func (b *Body) Format(f format.Formatter) {
	b.entity.Format(f)
	f.WriteString(100, "AcDbModelerGeometry")

	// Write ACIS version
	if b.Version != 0 {
		f.WriteInt(70, b.Version)
	}

	// Write UID
	if b.UID != "" {
		f.WriteString(1, b.UID)
	}

	// Write history handle (DXF R2007+)
	if b.HistoryHandle != nil && b.HistoryHandle.Handle() != "0" {
		f.WriteString(350, b.HistoryHandle.Handle())
	}

	// Write flags
	if b.Flags != 0 {
		f.WriteInt(71, b.Flags)
	}

	// Write SAT data if available
	if len(b.SATData) > 0 {
		// Note: This would need binary encoding support in formatter
		// For now, skip binary data writing
	}

	// Write transformation matrix if present
	if b.Transform != nil {
		// Write 16 float values for 4x4 matrix
		for i := 0; i < 16; i++ {
			f.WriteFloat(40+i, b.Transform.Matrix[i])
		}
	}

	// Write bounding box if calculated
	if b.boundingBox != nil {
		f.WriteFloat(10, b.boundingBox.Min.X())
		f.WriteFloat(20, b.boundingBox.Min.Y())
		f.WriteFloat(30, b.boundingBox.Min.Z())
		f.WriteFloat(11, b.boundingBox.Max.X())
		f.WriteFloat(21, b.boundingBox.Max.Y())
		f.WriteFloat(31, b.boundingBox.Max.Z())
	}

	// Write geometric properties
	if b.volume != 0 {
		f.WriteFloat(42, b.volume)
	}
	if b.surfaceArea != 0 {
		f.WriteFloat(43, b.surfaceArea)
	}

	// Write center point
	center := b.GetCenter()
	f.WriteFloat(10, center.X())
	f.WriteFloat(20, center.Y())
	f.WriteFloat(30, center.Z())

	f.WriteString(100, "AcDbBody")
}

// SetCenter sets the geometric center point
func (b *Body) SetCenter(center dxfmath.Vec3) {
	b.center = center
}

// BBox returns bounding box
func (b *Body) BBox() ([]float64, []float64) {
	bb := b.GetBoundingBox()
	return []float64{bb.Min.X(), bb.Min.Y(), bb.Min.Z()}, []float64{bb.Max.X(), bb.Max.Y(), bb.Max.Z()}
}

// String outputs data using default formatter
func (b *Body) String() string {
	f := format.NewASCII()
	return b.FormatString(f)
}

// FormatString outputs data using given formatter
func (b *Body) FormatString(f format.Formatter) string {
	b.Format(f)
	return f.Output()
}
