package entity

import (
	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
	dxfmath "github.com/edanko/dxf/math"
	"github.com/edanko/dxf/table"
	"math"
)

// Shape represents a DXF SHAPE entity (shape reference)
// SHAPE entities are used to insert complex shapes defined in shape files
type Shape struct {
	*entity

	// Basic properties
	insertPoint dxfmath.Vec3 // 10,20,30 - Insertion point (in WCS)
	shapeName   string       // 2 - Shape name
	size        float64      // 40 - Shape size
	thickness   float64      // 39 - Thickness (legacy, R11 and prior)
	rotation    float64      // 50 - Rotation angle (degrees)
	xScale      float64      // 41 - Relative X scale factor
	oblique     float64      // 51 - Oblique angle (degrees)
	extrusion   dxfmath.Vec3 // 210,220,230 - Extrusion direction (normal vector)
	elevation   float64      // 38 - Elevation (legacy, R11 and prior)
}

// NewShape creates a new SHAPE entity
func NewShape() *Shape {
	s := &Shape{
		entity:      NewEntity(SHAPE),
		insertPoint: dxfmath.NewVec3(0, 0, 0),
		shapeName:   "",
		size:        1.0,
		thickness:   0.0,
		rotation:    0.0,
		xScale:      1.0,
		oblique:     0.0,
		extrusion:   dxfmath.NewVec3(0, 0, 1), // Default Z-axis
		elevation:   0.0,
	}
	return s
}

// SetInsertPoint sets insertion point in WCS
func (s *Shape) SetInsertPoint(point dxfmath.Vec3) *Shape {
	s.insertPoint = point
	return s
}

// SetShapeName sets the shape name to reference
func (s *Shape) SetShapeName(name string) *Shape {
	s.shapeName = name
	return s
}

// SetSize sets the shape size
func (s *Shape) SetSize(size float64) *Shape {
	s.size = size
	return s
}

// SetThickness sets the shape thickness (legacy R11 and prior)
func (s *Shape) SetThickness(thickness float64) *Shape {
	s.thickness = thickness
	return s
}

// SetRotation sets the rotation angle in degrees
func (s *Shape) SetRotation(rotation float64) *Shape {
	s.rotation = rotation
	return s
}

// SetXScale sets the relative X scale factor
func (s *Shape) SetXScale(scale float64) *Shape {
	s.xScale = scale
	return s
}

// SetOblique sets the oblique angle in degrees
func (s *Shape) SetOblique(oblique float64) *Shape {
	s.oblique = oblique
	return s
}

// SetExtrusion sets the extrusion direction (normal vector)
func (s *Shape) SetExtrusion(vector dxfmath.Vec3) *Shape {
	s.extrusion = vector
	return s
}

// SetElevation sets the elevation (legacy R11 and prior)
func (s *Shape) SetElevation(elevation float64) *Shape {
	s.elevation = elevation
	return s
}

// GetInsertPoint returns the insertion point
func (s *Shape) GetInsertPoint() dxfmath.Vec3 {
	return s.insertPoint
}

// GetShapeName returns the shape name
func (s *Shape) GetShapeName() string {
	return s.shapeName
}

// GetSize returns the shape size
func (s *Shape) GetSize() float64 {
	return s.size
}

// GetThickness returns the thickness
func (s *Shape) GetThickness() float64 {
	return s.thickness
}

// GetRotation returns the rotation angle
func (s *Shape) GetRotation() float64 {
	return s.rotation
}

// GetXScale returns the X scale factor
func (s *Shape) GetXScale() float64 {
	return s.xScale
}

// GetOblique returns the oblique angle
func (s *Shape) GetOblique() float64 {
	return s.oblique
}

// GetExtrusion returns the extrusion direction
func (s *Shape) GetExtrusion() dxfmath.Vec3 {
	return s.extrusion
}

// GetElevation returns the elevation
func (s *Shape) GetElevation() float64 {
	return s.elevation
}

// BBox returns bounding box of the shape
// For simplicity, returns a box based on the shape size
func (s *Shape) BBox() ([]float64, []float64) {
	halfSize := s.size / 2.0
	min := s.insertPoint.Sub(dxfmath.NewVec3(halfSize, halfSize, halfSize))
	max := s.insertPoint.Add(dxfmath.NewVec3(halfSize, halfSize, halfSize))
	return []float64{min[0], min[1], min[2]}, []float64{max[0], max[1], max[2]}
}

// Format writes SHAPE entity to DXF format
func (s *Shape) Format(f format.Formatter) {
	s.entity.Format(f)

	// AcDbShape subclass (DXF2000+)
	f.WriteString(100, "AcDbShape")

	// Insertion point (WCS)
	if s.insertPoint != (dxfmath.Vec3{}) {
		f.WriteFloat(10, s.insertPoint[0])
		f.WriteFloat(20, s.insertPoint[1])
		f.WriteFloat(30, s.insertPoint[2])
	}

	// Shape name
	f.WriteString(2, s.shapeName)

	// Shape size
	f.WriteFloat(40, s.size)

	// Legacy attributes (R11 and prior)
	if s.thickness != 0.0 {
		f.WriteFloat(39, s.thickness)
	}
	if s.elevation != 0.0 {
		f.WriteFloat(38, s.elevation)
	}

	// Rotation angle
	if s.rotation != 0.0 {
		f.WriteFloat(50, s.rotation)
	}

	// X scale factor
	if s.xScale != 1.0 {
		f.WriteFloat(41, s.xScale)
	}

	// Oblique angle
	if s.oblique != 0.0 {
		f.WriteFloat(51, s.oblique)
	}

	// Extrusion direction (normal vector)
	if s.extrusion != (dxfmath.Vec3{}) {
		f.WriteFloat(210, s.extrusion[0])
		f.WriteFloat(220, s.extrusion[1])
		f.WriteFloat(230, s.extrusion[2])
	}
}

// Transform applies a transformation matrix to the shape
func (s *Shape) Transform(matrix dxfmath.Matrix44) {
	// Transform insertion point
	s.insertPoint = matrix.MulVec(s.insertPoint)

	// Transform extrusion direction (rotation only)
	s.extrusion = matrix.TransformVector(s.extrusion)

	// Scale and rotation are handled by the matrix transformation
}

// Clone creates a deep copy of the shape
func (s *Shape) Clone() Entity {
	clone := NewShape()
	clone.entity.Type = s.entity.Type
	clone.entity.handle = s.entity.handle
	clone.entity.blockRecord = s.entity.blockRecord
	clone.entity.owner = s.entity.owner
	clone.entity.layer = s.entity.layer
	clone.entity.ltscale = s.entity.ltscale
	clone.entity.color = s.entity.color

	clone.insertPoint = s.insertPoint
	clone.shapeName = s.shapeName
	clone.size = s.size
	clone.thickness = s.thickness
	clone.rotation = s.rotation
	clone.xScale = s.xScale
	clone.oblique = s.oblique
	clone.extrusion = s.extrusion
	clone.elevation = s.elevation
	return clone
}

// IsEqual checks if two shapes are equal
func (s *Shape) IsEqual(other Entity) bool {
	if o, ok := other.(*Shape); ok {
		if s.entity.Type != o.entity.Type || s.entity.handle != o.entity.handle {
			return false
		}

		// Compare shape-specific properties
		epsilon := 1e-9
		return s.shapeName == o.shapeName &&
			math.Abs(s.size-o.size) < epsilon &&
			math.Abs(s.thickness-o.thickness) < epsilon &&
			math.Abs(s.rotation-o.rotation) < epsilon &&
			math.Abs(s.xScale-o.xScale) < epsilon &&
			math.Abs(s.oblique-o.oblique) < epsilon &&
			math.Abs(s.elevation-o.elevation) < epsilon &&
			s.insertPoint.IsEqual(o.insertPoint, epsilon) &&
			s.extrusion.IsEqual(o.extrusion, epsilon)
	}
	return false
}

// SetColor sets the color for the shape
func (s *Shape) SetColor(cl color.ColorNumber) {
	s.entity.SetColor(cl)
}

// SetLayer sets the layer for the shape
func (s *Shape) SetLayer(layer *table.Layer) {
	s.entity.SetLayer(layer)
}

// SetLtscale sets the linetype scale for the shape
func (s *Shape) SetLtscale(scale float64) {
	s.entity.SetLtscale(scale)
}

// SetBlockRecord sets the block record for the shape
func (s *Shape) SetBlockRecord(handler handle.Handler) {
	s.entity.SetBlockRecord(handler)
}

// Layer returns the layer of the shape
func (s *Shape) Layer() *table.Layer {
	return s.entity.Layer()
}

// IsEntity returns true since this is an entity
func (s *Shape) IsEntity() bool {
	return true
}

// Handle returns the handle of the shape
func (s *Shape) Handle() string {
	return s.entity.Handle()
}

// String returns a string representation of the shape
func (s *Shape) String() string {
	return EntityTypeString(SHAPE) + " '" + s.shapeName + "' at [" +
		dxfmath.NewVec3(s.insertPoint[0], s.insertPoint[1], s.insertPoint[2]).String() + "]"
}
