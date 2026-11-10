package entity

import (
	"math"

	"github.com/edanko/dxf/format"
	dxfmath "github.com/edanko/dxf/math"
)

// Circle represents CIRCLE Entity.
type Circle struct {
	*entity
	Center    []float64 // 10, 20, 30
	Radius    float64   // 40
	Direction []float64 // 210, 220, 230
}

// IsEntity is for Entity interface.
func (c *Circle) IsEntity() bool {
	return true
}

// NewCircle creates a new Circle.
func NewCircle() *Circle {
	c := &Circle{
		entity:    NewEntity(CIRCLE),
		Center:    []float64{0.0, 0.0, 0.0},
		Radius:    0.0,
		Direction: []float64{0.0, 0.0, 1.0},
	}
	return c
}

// Format writes data to formatter.
func (c *Circle) Format(f format.Formatter) {
	c.entity.Format(f)
	f.WriteString(100, "AcDbCircle")
	if c.thickness != 0 {
		f.WriteFloat(39, c.thickness)
	}
	for i := range 3 {
		f.WriteFloat((i+1)*10, c.Center[i])
	}
	f.WriteFloat(40, c.Radius)
	c.writeDirection(f)
}

// writeDirection emits the extrusion direction (210, 220, 230). A negative Z
// component marks a mirrored entity; DXF requires it verbatim.
func (c *Circle) writeDirection(f format.Formatter) {
	d := c.Direction
	if len(d) < 3 {
		d = []float64{0, 0, 1}
	}
	f.WriteFloat(210, d[0])
	f.WriteFloat(220, d[1])
	f.WriteFloat(230, d[2])
}

// CurrentDirection returns extrusion direction.
func (c *Circle) CurrentDirection() []float64 {
	return c.Direction
}

// SetDirection sets new extrusion direction.
func (c *Circle) SetDirection(d []float64) {
	if len(d) >= 3 {
		c.Direction = []float64{d[0], d[1], d[2]}
	}
}

// CurrentCoord returns center point coord.
func (c *Circle) CurrentCoord() []float64 {
	return c.Center
}

// SetCoord sets new center point coord.
func (c *Circle) SetCoord(co []float64) {
	if len(co) >= 3 {
		c.Center = []float64{co[0], co[1], co[2]}
	}
}

func (c *Circle) BBox() ([]float64, []float64) {
	if len(c.Center) < 3 {
		return []float64{0, 0, 0}, []float64{0, 0, 0}
	}

	// Default axis-aligned bounding box (when extrusion is [0, 0, 1])
	if len(c.Direction) < 3 {
		return []float64{c.Center[0] - c.Radius, c.Center[1] - c.Radius, c.Center[2]}, []float64{c.Center[0] + c.Radius, c.Center[1] + c.Radius, c.Center[2]}
	}

	// Check if direction is approximately the default Z axis
	dx, dy, dz := c.Direction[0], c.Direction[1], c.Direction[2]
	if math.Abs(dx) < 1e-9 && math.Abs(dy) < 1e-9 && math.Abs(dz-1.0) < 1e-9 {
		return []float64{c.Center[0] - c.Radius, c.Center[1] - c.Radius, c.Center[2]}, []float64{c.Center[0] + c.Radius, c.Center[1] + c.Radius, c.Center[2]}
	}

	// For non-default extrusion, calculate OCS axes
	// The extrusion direction is the OCS Z axis
	ocsZ := dxfmath.NewVec3(dx, dy, dz)
	ocsZ = ocsZ.Normalized()

	// Find two orthogonal axes in the OCS XY plane
	// Start with an arbitrary vector that's not parallel to ocsZ
	var ocsX, ocsY dxfmath.Vec3
	if math.Abs(ocsZ.X()) < 0.9 {
		ocsX = dxfmath.NewVec3(1, 0, 0)
	} else {
		ocsX = dxfmath.NewVec3(0, 1, 0)
	}
	// Project ocsX onto the plane perpendicular to ocsZ, then normalize
	dot := ocsX.Dot(ocsZ)
	ocsX = ocsX.Sub(ocsZ.Mul(dot)).Normalized()
	ocsY = ocsZ.Cross(ocsX).Normalized()

	// Calculate the bounding box in WCS
	center := dxfmath.NewVec3(c.Center[0], c.Center[1], c.Center[2])

	// Sample points around the circle to find the bounding box
	numSamples := 8
	cornerMins := []float64{math.MaxFloat64, math.MaxFloat64, math.MaxFloat64}
	cornerMaxs := []float64{-math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64}

	for i := range numSamples {
		angle := float64(i) * 2 * math.Pi / float64(numSamples)
		ocsPoint := ocsX.Mul(c.Radius * math.Cos(angle)).Add(ocsY.Mul(c.Radius * math.Sin(angle)))
		wcsPoint := center.Add(ocsPoint)

		for j := range 3 {
			if wcsPoint[j] < cornerMins[j] {
				cornerMins[j] = wcsPoint[j]
			}
			if wcsPoint[j] > cornerMaxs[j] {
				cornerMaxs[j] = wcsPoint[j]
			}
		}
	}

	return cornerMins, cornerMaxs
}

// Transform applies an affine matrix to the circle. Translation and rotation
// are exact; a uniform scale scales the radius. Non-uniform scaling produces
// an ellipse, which CIRCLE cannot represent, so ErrNonUniformScale is
// returned (documented policy; convert to ELLIPSE at the call site).
func (c *Circle) Transform(m dxfmath.Matrix44) error {
	if !IsUniformScale(m, 1e-9) {
		return ErrNonUniformScale
	}
	sx, _, _ := ScaleFactors(m)
	c.Center = TransformPoint(m, c.Center)
	c.Radius = abs64(c.Radius * sx)
	if len(c.Direction) >= 3 {
		d := TransformDirection(m, c.Direction)
		n := dxfmath.NewVec3(d[0], d[1], d[2]).Normalized()
		c.Direction = []float64{n.X(), n.Y(), n.Z()}
	}
	return nil
}

func abs64(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

// Copy creates a deep copy of the Circle entity
func (c *Circle) Copy() Entity {
	circle := NewCircle()
	circle.entity = c.entity.clone()
	if len(c.Center) >= 3 {
		circle.Center = []float64{c.Center[0], c.Center[1], c.Center[2]}
	}
	circle.Radius = c.Radius
	if len(c.Direction) >= 3 {
		circle.Direction = []float64{c.Direction[0], c.Direction[1], c.Direction[2]}
	}
	return circle
}
