package entity

import (
	"errors"
	"fmt"
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

}

// ErrDegenerateEllipse is returned when an ellipse cannot be converted to a
// Bézier decomposition because one of its defining dimensions is zero.
var ErrDegenerateEllipse = errors.New("degenerate ellipse")

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
	}
}

// ToBezier converts the ellipse or ellipse arc into a decomposition of cubic
// Bézier segments using the standard arc-to-cubic conversion: each sub-arc of
// at most 90 degrees is represented by one cubic segment whose inner control
// points are offset along the ellipse tangents by k = 4/3·tan(span/4)
// ("Approximation of circular arcs by cubic Béziers", a.k.a. the Parry/Ris
// formula). A full ellipse decomposes into four quarter-arc segments.
//
// The returned segments start at the ellipse parameter Start and end at End,
// so consecutive segments share their boundary points and a full ellipse
// closes exactly.
//
// It returns ErrDegenerateEllipse (wrapped with the offending field) when the
// major or minor axis length is zero, and an error naming the embedded
// representation when the ELLIPSE record retains raw data (an embedded spline
// definition that cannot be reproduced as a Bézier decomposition).
func (e *Ellipse) ToBezier() ([]*dxfmath.Bezier, error) {
	if len(e.rawTags) > 0 {
		return nil, fmt.Errorf("embedded spline representation is not supported by Bézier conversion")
	}
	if len(e.Center) < 3 {
		return nil, fmt.Errorf("degenerate ellipse: center point is not set")
	}
	a := e.Major
	if a <= 0 {
		return nil, fmt.Errorf("%w: major axis length is %g", ErrDegenerateEllipse, a)
	}
	b := e.Minor
	if b == 0 {
		b = a * e.Ratio
	}
	if b <= 0 {
		return nil, fmt.Errorf("%w: minor axis length is %g", ErrDegenerateEllipse, b)
	}

	start := e.Start
	end := e.End
	if end < start {
		start, end = end, start
	}
	span := end - start
	// A full ellipse is decomposed into four quarter arcs; an open arc is
	// split into sub-arcs of at most 90 degrees.
	const quarter = 90.0
	count := 4
	if span < 359.999 {
		if span <= 0 {
			return nil, fmt.Errorf("%w: start and end parameters are equal (%g)", ErrDegenerateEllipse, span)
		}
		count = int(math.Ceil(span / quarter))
		if count < 1 {
			count = 1
		}
	}

	segments := make([]*dxfmath.Bezier, 0, count)
	for i := 0; i < count; i++ {
		segSpan := span / float64(count) // degrees, <= 90
		theta0 := (start + float64(i)*segSpan) * math.Pi / 180.0
		theta1 := theta0 + segSpan*math.Pi/180.0

		// Standard arc-to-cubic conversion for the sub-arc [theta0, theta1].
		rad := segSpan * math.Pi / 180.0
		k := (4.0 / 3.0) * math.Tan(rad/4.0)

		p0 := e.ellipsePoint(a, b, theta0)
		p3 := e.ellipsePoint(a, b, theta1)
		t0 := ellipseTangent(a, b, theta0)
		t1 := ellipseTangent(a, b, theta1)
		p1 := p0.Add(t0.Mul(k))
		p2 := p3.Sub(t1.Mul(k))

		bezier, err := dxfmath.NewBezier4P(p0, p1, p2, p3)
		if err != nil {
			return nil, fmt.Errorf("ellipse Bézier conversion: %w", err)
		}
		segments = append(segments, bezier)
	}
	return segments, nil
}

// ellipsePoint evaluates the parametric ellipse point E(theta) with the major
// axis of length a along X and the minor axis of length b along Y.
func (e *Ellipse) ellipsePoint(a, b, theta float64) dxfmath.Vec3 {
	return dxfmath.NewVec3(
		e.Center[0]+a*math.Cos(theta),
		e.Center[1]+b*math.Sin(theta),
		e.Center[2],
	)
}

// ellipseTangent returns the ellipse tangent vector T(theta) =
// (-a·sin(theta), b·cos(theta)), the derivative of the parametric point with
// respect to the angle. It is deliberately NOT normalized: the arc-to-cubic
// conversion scales the raw derivative by k.
func ellipseTangent(a, b, theta float64) dxfmath.Vec3 {
	return dxfmath.NewVec3(-a*math.Sin(theta), b*math.Cos(theta), 0)
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
// RawWholeRecord reports whole-record retention: an ELLIPSE carrying embedded
// spline data cannot be re-emitted from the typed model, so its retained
// record is written verbatim.
func (e *Ellipse) RawWholeRecord() bool { return len(e.rawTags) > 0 }

func (e *Ellipse) Format(f format.Formatter) {
	e.entity.Format(f)
	f.WriteString(100, "AcDbEllipse")
	if e.thickness != 0 {
		f.WriteFloat(39, e.thickness)
	}

	// Write center point
	for i := range 3 {
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

	e.entity.FormatBaseExtrusion(f)
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
	segments, err := e.ToBezier()
	if err != nil {
		return nil
	}
	var out []dxfmath.Vec3
	for _, seg := range segments {
		pts := seg.Flattening(deviation, 50)
		// Drop the duplicated boundary point shared with the previous
		// segment.
		if len(out) > 0 && len(pts) > 0 {
			pts = pts[1:]
		}
		out = append(out, pts...)
	}
	return out
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

// Transform applies an affine matrix to the ellipse center; the major and
// minor radii scale by the uniform XY scale factors (non-uniform shear is not
// decomposed). Rotation in the XY plane shifts the start/end parameters.
func (e *Ellipse) Transform(m dxfmath.Matrix44) error {
	sx, sy, _ := ScaleFactors(m)
	e.Center = TransformPoint(m, e.Center)
	e.Major *= sx
	e.Minor *= sy
	rot := math.Atan2(m.Get(1, 0), m.Get(0, 0))
	if rot != 0 {
		e.Start += rot
		e.End += rot
	}
	return nil
}

// Copy creates a deep copy of the Ellipse entity
func (e *Ellipse) Copy() Entity {
	ellipse := NewEllipse()
	ellipse.entity = e.entity.clone()
	ellipse.Center = []float64{e.Center[0], e.Center[1], e.Center[2]}
	ellipse.Major = e.Major
	ellipse.Minor = e.Minor
	ellipse.Ratio = e.Ratio
	ellipse.Start = e.Start
	ellipse.End = e.End
	return ellipse
}
