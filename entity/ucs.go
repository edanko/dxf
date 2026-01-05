package entity

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/math"
)

// UCSTableEntry represents a User Coordinate System table entry
type UCSTableEntry struct {
	*entity
	Origin    math.Vec3
	XAxis     math.Vec3
	YAxis     math.Vec3
	Name      string
	Elevation float64
	UCSIcon   int
	UCSIconAt int
}

// NewUCSTableEntry creates a new UCSTableEntry entity
func NewUCSTableEntry() *UCSTableEntry {
	u := &UCSTableEntry{
		entity:    NewEntity(UCS),
		Origin:    math.Vec3{0, 0, 0},
		XAxis:     math.Vec3{1, 0, 0},
		YAxis:     math.Vec3{0, 1, 0},
		Name:      "",
		Elevation: 0,
		UCSIcon:   1,
		UCSIconAt: 1,
	}
	return u
}

// IsEntity is for Entity interface.
func (u *UCSTableEntry) IsEntity() bool {
	return true
}

// Format writes data to formatter.
func (u *UCSTableEntry) Format(f format.Formatter) {
	u.entity.Format(f)
	f.WriteString(100, "AcDbSymbolTableRecord")
	f.WriteString(100, "AcDbUcsTableRecord")
	f.WriteString(2, u.Name)
	f.WriteFloat(10, u.Origin.X())
	f.WriteFloat(20, u.Origin.Y())
	f.WriteFloat(30, u.Origin.Z())
	f.WriteFloat(11, u.XAxis.X())
	f.WriteFloat(21, u.XAxis.Y())
	f.WriteFloat(31, u.XAxis.Z())
	f.WriteFloat(12, u.YAxis.X())
	f.WriteFloat(22, u.YAxis.Y())
	f.WriteFloat(32, u.YAxis.Z())
	f.WriteInt(79, u.UCSIcon)
	f.WriteFloat(146, u.Elevation)
}

// BBox returns the bounding box of the entity
func (u *UCSTableEntry) BBox() ([]float64, []float64) {
	// UCS is a metadata entity, return empty bounding box
	return []float64{0, 0, 0}, []float64{0, 0, 0}
}

// Transform applies a transformation matrix to the UCSTableEntry
func (u *UCSTableEntry) Transform(m *math.Matrix44) error {
	// Note: Vec3 doesn't have Transform method, implementing manually
	transformVec := func(v math.Vec3) math.Vec3 {
		x := v[0]*m[0] + v[1]*m[1] + v[2]*m[2] + m[3]
		y := v[0]*m[4] + v[1]*m[5] + v[2]*m[6] + m[7]
		z := v[0]*m[8] + v[1]*m[9] + v[2]*m[10] + m[11]
		return math.Vec3{x, y, z}
	}
	u.Origin = transformVec(u.Origin)
	u.XAxis = transformVec(u.XAxis)
	u.YAxis = transformVec(u.YAxis)
	return nil
}

// Copy creates a deep copy of the UCSTableEntry entity
func (u *UCSTableEntry) Copy() Entity {
	ucs := NewUCSTableEntry()
	ucs.entity = u.entity
	ucs.Origin = u.Origin
	ucs.XAxis = u.XAxis
	ucs.YAxis = u.YAxis
	ucs.Name = u.Name
	ucs.Elevation = u.Elevation
	ucs.UCSIcon = u.UCSIcon
	ucs.UCSIconAt = u.UCSIconAt
	return ucs
}

// Validate validates the UCSTableEntry entity
func (u *UCSTableEntry) Validate() error {
	u.XAxis = u.XAxis.Normalized()
	u.YAxis = u.YAxis.Normalized()
	zAxis := u.XAxis.Cross(u.YAxis).Normalized()
	u.YAxis = zAxis.Cross(u.XAxis).Normalized()
	return nil
}
