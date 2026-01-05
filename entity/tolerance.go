package entity

import (
	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
	dxfmath "github.com/edanko/dxf/math"
	"github.com/edanko/dxf/table"
)

// Tolerance represents a DXF TOLERANCE entity (geometric tolerance)
// TOLERANCE entities are used for geometric dimensioning and tolerancing (GD&T)
type Tolerance struct {
	*entity

	// Basic properties
	dimstyle    string       // 3 - Dimension style name
	insertPoint dxfmath.Vec3 // 10,20,30 - Insertion point (in WCS)
	content     string       // 1 - Tolerance content string
	extrusion   dxfmath.Vec3 // 210,220,230 - Extrusion direction (normal vector)
	xAxisVector dxfmath.Vec3 // 11,21,31 - X-axis direction vector (in WCS)
}

// NewTolerance creates a new TOLERANCE entity
func NewTolerance() *Tolerance {
	t := &Tolerance{
		entity:      NewEntity(TOLERANCE),
		dimstyle:    "Standard",
		insertPoint: dxfmath.NewVec3(0, 0, 0),
		content:     "",
		extrusion:   dxfmath.NewVec3(0, 0, 1), // Default Z-axis
		xAxisVector: dxfmath.NewVec3(1, 0, 0), // Default X-axis
	}
	return t
}

// SetDimStyle sets dimension style name
func (t *Tolerance) SetDimStyle(style string) *Tolerance {
	t.dimstyle = style
	return t
}

// SetInsertPoint sets insertion point in WCS
func (t *Tolerance) SetInsertPoint(point dxfmath.Vec3) *Tolerance {
	t.insertPoint = point
	return t
}

// SetContent sets tolerance content string
// The content should contain GD&T symbols and tolerance values
func (t *Tolerance) SetContent(content string) *Tolerance {
	t.content = content
	return t
}

// SetExtrusion sets extrusion direction (normal vector)
func (t *Tolerance) SetExtrusion(vector dxfmath.Vec3) *Tolerance {
	t.extrusion = vector
	return t
}

// SetXAxisVector sets X-axis direction vector
func (t *Tolerance) SetXAxisVector(vector dxfmath.Vec3) *Tolerance {
	t.xAxisVector = vector
	return t
}

// GetDimStyle returns the dimension style name
func (t *Tolerance) GetDimStyle() string {
	return t.dimstyle
}

// GetInsertPoint returns the insertion point
func (t *Tolerance) GetInsertPoint() dxfmath.Vec3 {
	return t.insertPoint
}

// GetContent returns the tolerance content
func (t *Tolerance) GetContent() string {
	return t.content
}

// GetExtrusion returns the extrusion direction
func (t *Tolerance) GetExtrusion() dxfmath.Vec3 {
	return t.extrusion
}

// GetXAxisVector returns the X-axis direction vector
func (t *Tolerance) GetXAxisVector() dxfmath.Vec3 {
	return t.xAxisVector
}

// BBox returns the bounding box of the tolerance entity
// For simplicity, returns a small box around the insertion point
func (t *Tolerance) BBox() ([]float64, []float64) {
	min := t.insertPoint.Sub(dxfmath.NewVec3(1, 1, 1))
	max := t.insertPoint.Add(dxfmath.NewVec3(1, 1, 1))
	return []float64{min[0], min[1], min[2]}, []float64{max[0], max[1], max[2]}
}

// Format writes the TOLERANCE entity to DXF format
func (t *Tolerance) Format(f format.Formatter) {
	t.entity.Format(f)

	// AcDbFcf subclass
	f.WriteString(100, "AcDbFcf")

	// Dimension style name
	f.WriteString(3, t.dimstyle)

	// Insertion point (WCS)
	if t.insertPoint != (dxfmath.Vec3{}) {
		f.WriteFloat(10, t.insertPoint[0])
		f.WriteFloat(20, t.insertPoint[1])
		f.WriteFloat(30, t.insertPoint[2])
	}

	// Tolerance content
	f.WriteString(1, t.content)

	// Extrusion direction (normal vector)
	if t.extrusion != (dxfmath.Vec3{}) {
		f.WriteFloat(210, t.extrusion[0])
		f.WriteFloat(220, t.extrusion[1])
		f.WriteFloat(230, t.extrusion[2])
	}

	// X-axis direction vector
	if t.xAxisVector != (dxfmath.Vec3{}) {
		f.WriteFloat(11, t.xAxisVector[0])
		f.WriteFloat(21, t.xAxisVector[1])
		f.WriteFloat(31, t.xAxisVector[2])
	}
}

// Transform applies a transformation matrix to the tolerance
func (t *Tolerance) Transform(matrix dxfmath.Matrix44) {
	// Transform insertion point
	t.insertPoint = matrix.MulVec(t.insertPoint)

	// Transform direction vectors (rotation only)
	t.extrusion = matrix.TransformVector(t.extrusion)
	t.xAxisVector = matrix.TransformVector(t.xAxisVector)
}

// Clone creates a deep copy of the tolerance
func (t *Tolerance) Clone() Entity {
	clone := NewTolerance()
	clone.entity.Type = t.entity.Type
	clone.entity.handle = t.entity.handle
	clone.entity.blockRecord = t.entity.blockRecord
	clone.entity.owner = t.entity.owner
	clone.entity.layer = t.entity.layer
	clone.entity.ltscale = t.entity.ltscale
	clone.entity.color = t.entity.color

	clone.dimstyle = t.dimstyle
	clone.insertPoint = t.insertPoint
	clone.content = t.content
	clone.extrusion = t.extrusion
	clone.xAxisVector = t.xAxisVector
	return clone
}

// IsEqual checks if two tolerances are equal
func (t *Tolerance) IsEqual(other Entity) bool {
	if o, ok := other.(*Tolerance); ok {
		if t.entity.Type != o.entity.Type || t.entity.handle != o.entity.handle {
			return false
		}

		// Compare tolerance-specific properties
		if t.dimstyle != o.dimstyle || t.content != o.content {
			return false
		}

		epsilon := 1e-9
		return t.insertPoint.IsEqual(o.insertPoint, epsilon) &&
			t.extrusion.IsEqual(o.extrusion, epsilon) &&
			t.xAxisVector.IsEqual(o.xAxisVector, epsilon)
	}
	return false
}

// SetColor sets the color for the tolerance
func (t *Tolerance) SetColor(cl color.ColorNumber) {
	t.entity.SetColor(cl)
}

// SetLayer sets the layer for the tolerance
func (t *Tolerance) SetLayer(layer *table.Layer) {
	t.entity.SetLayer(layer)
}

// SetLtscale sets the linetype scale for the tolerance
func (t *Tolerance) SetLtscale(scale float64) {
	t.entity.SetLtscale(scale)
}

// SetBlockRecord sets the block record for the tolerance
func (t *Tolerance) SetBlockRecord(handler handle.Handler) {
	t.entity.SetBlockRecord(handler)
}

// Layer returns the layer of the tolerance
func (t *Tolerance) Layer() *table.Layer {
	return t.entity.Layer()
}

// IsEntity returns true since this is an entity
func (t *Tolerance) IsEntity() bool {
	return true
}

// Handle returns the handle of the tolerance
func (t *Tolerance) Handle() string {
	return t.entity.Handle()
}

// String returns a string representation of the tolerance
func (t *Tolerance) String() string {
	return EntityTypeString(TOLERANCE) + " " + t.content + " at [" +
		dxfmath.NewVec3(t.insertPoint[0], t.insertPoint[1], t.insertPoint[2]).String() + "]"
}
