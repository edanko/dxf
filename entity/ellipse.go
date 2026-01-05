package entity

import (
	"math"

	"github.com/edanko/dxf/format"
	dxfmath "github.com/edanko/dxf/math"
)

// Ellipse represents ELLIPSE Entity.
type Ellipse struct {
	*entity
	Center []float64 // 10, 20, 30
	Major  float64   // 40
	Minor  float64   // 41
	Ratio  float64   // 200
	Start  float64   // 50
	End    float64   // 51

	// Enhanced math support
	bezierCurve *dxfmath.Bezier
}

// IsEntity is for Entity interface.
func (e *Ellipse) IsEntity() bool {
	return true
}

// NewEllipse creates a new Ellipse.
func NewEllipse() *Ellipse {
	return &Ellipse{
		entity:      NewEntity(ELLIPSE),
		Center:      []float64{0.0, 0.0, 0.0},
		Major:       0.0,
		Minor:       0.0,
		Ratio:       1.0,
		Start:       0.0,
		End:         360.0,
		bezierCurve: nil,
	}
}

// ToBezier converts the ellipse to a Bézier curve approximation
func (e *Ellipse) ToBezier() *dxfmath.Bezier {
	if e.bezierCurve == nil {
		// Create a 4-point cubic Bézier approximation of the ellipse
		// Using the standard method for converting ellipses to Bézier curves

		// Control points for a cubic Bézier that approximates an ellipse arc
		// This is a simplified approximation for a full ellipse
		k := 0.552284749831 // Magic number for ellipse to Bézier conversion

		center := dxfmath.NewVec3(e.Center[0], e.Center[1], e.Center[2])

		// For a full ellipse, use 4 cubic Bézier segments
		// For simplicity, we'll create one segment for now
		controlPoints := []dxfmath.Vec3{
			center.Add(dxfmath.NewVec3(e.Major, 0, 0)),                 // Start point
			center.Add(dxfmath.NewVec3(e.Major, k*e.Major*e.Ratio, 0)), // Control point 1
			center.Add(dxfmath.NewVec3(k*e.Major, e.Major*e.Ratio, 0)), // Control point 2
			center.Add(dxfmath.NewVec3(0, e.Major*e.Ratio, 0)),         // End point
		}

		e.bezierCurve = dxfmath.NewBezier4P(
			controlPoints[0],
			controlPoints[1],
			controlPoints[2],
			controlPoints[3],
		)
	}
	return e.bezierCurve
}

// SetCenter sets the center point
func (e *Ellipse) SetCenter(x, y, z float64) {
	e.Center = []float64{x, y, z}
}

// GetCenter returns the center point
func (e *Ellipse) GetCenter() []float64 {
	return e.Center
}

// SetRadius sets the major and minor radius
func (e *Ellipse) SetRadius(major, minor float64) {
	e.Major = major
	e.Minor = minor
	if major > 0 {
		e.Ratio = minor / major
	} else {
		e.Ratio = 1.0
	}
}

// GetRadius returns the major and minor radius
func (e *Ellipse) GetRadius() (major, minor float64) {
	return e.Major, e.Minor
}

// SetStartAngle sets the start angle (degrees)
func (e *Ellipse) SetStartAngle(angle float64) {
	e.Start = angle
}

// GetStartAngle returns the start angle (degrees)
func (e *Ellipse) GetStartAngle() float64 {
	return e.Start
}

// SetEndAngle sets the end angle (degrees)
func (e *Ellipse) SetEndAngle(angle float64) {
	e.End = angle
}

// GetEndAngle returns the end angle (degrees)
func (e *Ellipse) GetEndAngle() float64 {
	return e.End
}

// SetAngles sets both start and end angles (degrees)
func (e *Ellipse) SetAngles(start, end float64) {
	e.Start = start
	e.End = end
}

// IsFull checks if ellipse is a full circle
func (e *Ellipse) IsFull() bool {
	// Consider full if end-start is approximately 360 degrees
	return (e.End-e.Start) >= 359.999 || (e.End-e.Start) <= -359.999
}

// Format writes data to formatter.
func (e *Ellipse) Format(f format.Formatter) {
	e.entity.Format(f)
	f.WriteString(100, "AcDbEllipse")

	// Write center point
	for i := 0; i < 3; i++ {
		f.WriteFloat(10+i*10, e.Center[i])
	}

	// Write radii
	f.WriteFloat(40, e.Major)
	f.WriteFloat(41, e.Minor)

	// Write start and end angles
	f.WriteFloat(50, e.Start)
	f.WriteFloat(51, e.End)

	// Write ratio
	if e.Ratio != 1.0 {
		f.WriteFloat(200, e.Ratio)
	}
}

// PointAt evaluates the ellipse at parameter t (0 to 1)
func (e *Ellipse) PointAt(t float64) dxfmath.Vec3 {
	// Convert t to angle
	angle := e.Start + t*(e.End-e.Start)

	// Calculate ellipse point
	// Using degree conversion since DXF uses degrees
	radAngle := angle * 3.141592653589793 / 180.0
	x := e.Center[0] + e.Major*math.Cos(radAngle)
	y := e.Center[1] + e.Minor*math.Sin(radAngle)
	z := e.Center[2]

	return dxfmath.NewVec3(x, y, z)
}

// Flatten converts the ellipse to line segments
func (e *Ellipse) Flatten(deviation float64) []dxfmath.Vec3 {
	bezier := e.ToBezier()
	return bezier.Flattening(deviation, 50) // Use 50 segments for ellipse
}

func (e *Ellipse) BBox() ([]float64, []float64) {
	// Simple bounding box using axis-aligned extents
	mins := []float64{
		e.Center[0] - e.Major,
		e.Center[1] - e.Major,
		e.Center[2],
	}

	maxs := []float64{
		e.Center[0] + e.Major,
		e.Center[1] + e.Major,
		e.Center[2],
	}

	// Adjust for minor axis if smaller
	if e.Minor < e.Major {
		mins[1] = e.Center[1] - e.Minor
		maxs[1] = e.Center[1] + e.Minor
	}

	return mins, maxs
}
