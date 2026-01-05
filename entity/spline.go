package entity

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/math"
)

// Spline represents SPLINE Entity.
type Spline struct {
	*entity
	Normal    []float64   // 210, 220, 230
	Flag      int         // 70
	Degree    int         // 71
	Knots     []float64   // 72, 40
	Weights   []float64   // 41
	Controls  [][]float64 // 73, 10, 20, 30
	Fits      [][]float64 // 74, 11, 21, 31
	Tolerance []float64   // 42, 43, 44

	// Enhanced math support
	enhancedBSpline *math.BSpline
}

// IsEntity is for Entity interface.
func (s *Spline) IsEntity() bool {
	return true
}

// NewSpline creates a new Spline.
func NewSpline(degree int) *Spline {
	return &Spline{
		entity:          NewEntity(SPLINE),
		Normal:          []float64{0.0, 0.0, 1.0},
		Flag:            0,
		Degree:          degree,
		Knots:           make([]float64, 0),
		Weights:         make([]float64, 0),
		Controls:        make([][]float64, 0),
		Fits:            make([][]float64, 0),
		Tolerance:       []float64{0.000000001, 0.000000001, 0.000000001},
		enhancedBSpline: nil,
	}
}

// NewSplineFromEnhanced creates a new Spline from EnhancedBSpline
func NewSplineFromEnhanced(bspline *math.BSpline) *Spline {
	s := &Spline{
		entity:          NewEntity(SPLINE),
		Normal:          []float64{0.0, 0.0, 1.0},
		Flag:            0,
		Knots:           make([]float64, 0),
		Weights:         make([]float64, 0),
		Controls:        make([][]float64, 0),
		Fits:            make([][]float64, 0),
		Tolerance:       []float64{0.000000001, 0.000000001, 0.000000001},
		enhancedBSpline: bspline,
	}

	// Copy data from enhanced B-spline
	s.Degree = bspline.Degree()
	s.Knots = bspline.Knots()
	s.Weights = bspline.Weights()

	// Convert control points
	controlPoints := bspline.ControlPoints()
	for _, point := range controlPoints {
		s.AddControlPoint(point.X(), point.Y(), point.Z())
	}

	return s
}

// AddKnot adds a knot value
func (s *Spline) AddKnot(knot float64) {
	s.Knots = append(s.Knots, knot)
}

// AddWeight adds a weight value
func (s *Spline) AddWeight(weight float64) {
	s.Weights = append(s.Weights, weight)
}

// AddControlPoint adds a control point (x, y, z)
func (s *Spline) AddControlPoint(x, y, z float64) {
	s.Controls = append(s.Controls, []float64{x, y, z})
	// Invalidate enhanced spline cache
	s.enhancedBSpline = nil
}

// AddFitPoint adds a fit point (x, y, z)
func (s *Spline) AddFitPoint(x, y, z float64) {
	s.Fits = append(s.Fits, []float64{x, y, z})
}

// SetNormal sets the normal vector
func (s *Spline) SetNormal(x, y, z float64) {
	s.Normal = []float64{x, y, z}
}

// SetClosed sets whether spline is closed
func (s *Spline) SetClosed(closed bool) {
	if closed {
		s.Flag |= 1
	} else {
		s.Flag &= ^1
	}
}

// IsClosed returns whether spline is closed
func (s *Spline) IsClosed() bool {
	return s.Flag&1 != 0
}

// SetPeriodic sets whether spline is periodic
func (s *Spline) SetPeriodic(periodic bool) {
	if periodic {
		s.Flag |= 2
	} else {
		s.Flag &= ^2
	}
}

// IsPeriodic returns whether spline is periodic
func (s *Spline) IsPeriodic() bool {
	return s.Flag&2 != 0
}

// SetRational sets whether spline is rational
func (s *Spline) SetRational(rational bool) {
	if rational {
		s.Flag |= 4
	} else {
		s.Flag &= ^4
	}
}

// IsRational returns whether spline is rational
func (s *Spline) IsRational() bool {
	return s.Flag&4 != 0
}

// SetPlanar sets whether spline is planar
func (s *Spline) SetPlanar(planar bool) {
	if planar {
		s.Flag |= 8
	} else {
		s.Flag &= ^8
	}
}

// IsPlanar returns whether spline is planar
func (s *Spline) IsPlanar() bool {
	return s.Flag&8 != 0
}

// SetLinear sets whether spline is linear
func (s *Spline) SetLinear(linear bool) {
	if linear {
		s.Flag |= 16
	} else {
		s.Flag &= ^16
	}
}

// IsLinear returns whether spline is linear
func (s *Spline) IsLinear() bool {
	return s.Flag&16 != 0
}

// Format writes data to formatter.
func (s *Spline) Format(f format.Formatter) {
	s.entity.Format(f)
	f.WriteString(100, "AcDbSpline")

	for i := 0; i < 3; i++ {
		f.WriteFloat(210+i*10, s.Normal[i])
	}

	f.WriteInt(70, s.Flag)
	f.WriteInt(71, s.Degree)
	f.WriteInt(72, len(s.Knots))
	f.WriteInt(73, len(s.Controls))
	f.WriteInt(74, len(s.Fits))

	// Write weights
	if len(s.Weights) > 0 {
		for _, weight := range s.Weights {
			f.WriteFloat(41, weight)
		}
	}

	// Write knots
	for _, knot := range s.Knots {
		f.WriteFloat(40, knot)
	}

	// Write control points
	for _, point := range s.Controls {
		for j, coord := range point {
			f.WriteFloat(10+j*10, coord)
		}
	}

	// Write fit points
	for _, point := range s.Fits {
		for j, coord := range point {
			f.WriteFloat(11+j*10, coord)
		}
	}

	// Write tolerances
	for i := 0; i < 3; i++ {
		if s.Tolerance[i] != 0.000000001 {
			f.WriteFloat(42+i, s.Tolerance[i])
		}
	}
}

// GetEnhancedBSpline returns or creates an EnhancedBSpline from this spline
func (s *Spline) GetEnhancedBSpline() *math.BSpline {
	if s.enhancedBSpline == nil {
		// Convert control points to math.Vec3
		controlPoints := make([]math.Vec3, len(s.Controls))
		for i, cp := range s.Controls {
			if len(cp) >= 3 {
				controlPoints[i] = math.NewVec3(cp[0], cp[1], cp[2])
			} else if len(cp) >= 2 {
				controlPoints[i] = math.NewVec3(cp[0], cp[1], 0)
			} else {
				controlPoints[i] = math.NewVec3(cp[0], 0, 0)
			}
		}

		// Create enhanced B-spline with existing data if available
		if len(s.Knots) > 0 && len(s.Weights) > 0 {
			s.enhancedBSpline = math.NewEnhancedBSplineWithKnots(controlPoints, s.Weights, s.Knots, s.Degree)
		} else {
			s.enhancedBSpline = math.NewEnhancedBSpline(controlPoints, s.Degree)
		}
	}
	return s.enhancedBSpline
}

// PointAt evaluates the spline at parameter t
func (s *Spline) PointAt(t float64) math.Vec3 {
	bspline := s.GetEnhancedBSpline()
	return bspline.EvaluatePoint(t)
}

// TangentAt evaluates the tangent at parameter t
func (s *Spline) TangentAt(t float64) math.Vec3 {
	bspline := s.GetEnhancedBSpline()
	derivatives := bspline.Derivative(t, 1)
	if len(derivatives) > 0 {
		return derivatives[0]
	}
	return math.NewVec3(0, 0, 0)
}

// Flatten converts the spline to line segments
func (s *Spline) Flatten(deviation float64) []math.Vec3 {
	bspline := s.GetEnhancedBSpline()
	return bspline.Flattening(deviation)
}

func (s *Spline) BBox() ([]float64, []float64) {
	allPoints := append(s.Controls, s.Fits...)

	if len(allPoints) == 0 {
		return []float64{0, 0, 0}, []float64{0, 0, 0}
	}

	mins := []float64{allPoints[0][0], allPoints[0][1], allPoints[0][2]}
	maxs := []float64{allPoints[0][0], allPoints[0][1], allPoints[0][2]}

	for _, point := range allPoints {
		if len(point) >= 3 {
			if point[0] < mins[0] {
				mins[0] = point[0]
			}
			if point[1] < mins[1] {
				mins[1] = point[1]
			}
			if point[2] < mins[2] {
				mins[2] = point[2]
			}
			if point[0] > maxs[0] {
				maxs[0] = point[0]
			}
			if point[1] > maxs[1] {
				maxs[1] = point[1]
			}
			if point[2] > maxs[2] {
				maxs[2] = point[2]
			}
		}
	}

	return mins, maxs
}
