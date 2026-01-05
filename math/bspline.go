package math

import (
	"fmt"
)

// BSpline represents a B-spline curve
type BSpline struct {
	controlPoints []Vec3
	weights       []float64
	knots         []float64
	degree        int
	order         int
}

// NewBSpline creates a new B-spline curve
func NewBSpline(controlPoints []Vec3, degree int) *BSpline {
	order := degree + 1
	n := len(controlPoints)
	m := n + order

	// Default weights to 1.0
	weights := make([]float64, n)
	for i := range weights {
		weights[i] = 1.0
	}

	// Create uniform knot vector
	knots := make([]float64, m)
	for i := 0; i < m; i++ {
		if i < order {
			knots[i] = 0.0
		} else if i >= n {
			knots[i] = 1.0
		} else {
			knots[i] = float64(i-order+1) / float64(n-degree)
		}
	}

	return &BSpline{
		controlPoints: controlPoints,
		weights:       weights,
		knots:         knots,
		degree:        degree,
		order:         order,
	}
}

// NewBSplineWithKnots creates a B-spline with custom knots and weights
func NewBSplineWithKnots(controlPoints []Vec3, weights []float64, knots []float64, degree int) *BSpline {
	order := degree + 1
	n := len(controlPoints)

	// Validate inputs
	if len(weights) != n {
		weights = make([]float64, n)
		for i := range weights {
			weights[i] = 1.0
		}
	}

	if len(knots) != n+order {
		// Create uniform knot vector
		knots = make([]float64, n+order)
		for i := 0; i < n+order; i++ {
			if i < order {
				knots[i] = 0.0
			} else if i >= n {
				knots[i] = 1.0
			} else {
				knots[i] = float64(i-order+1) / float64(n-degree)
			}
		}
	}

	return &BSpline{
		controlPoints: controlPoints,
		weights:       weights,
		knots:         knots,
		degree:        degree,
		order:         order,
	}
}

// NewEnhancedBSpline creates an enhanced B-spline (alias for NewBSpline)
func NewEnhancedBSpline(controlPoints []Vec3, degree int) *BSpline {
	return NewBSpline(controlPoints, degree)
}

// NewEnhancedBSplineWithKnots creates an enhanced B-spline with custom knots and weights (alias for NewBSplineWithKnots)
func NewEnhancedBSplineWithKnots(controlPoints []Vec3, weights []float64, knots []float64, degree int) *BSpline {
	return NewBSplineWithKnots(controlPoints, weights, knots, degree)
}

// EvaluatePoint evaluates the B-spline at parameter t
func (bs *BSpline) EvaluatePoint(t float64) Vec3 {
	if t <= 0.0 {
		return bs.controlPoints[0]
	}
	if t >= 1.0 {
		return bs.controlPoints[len(bs.controlPoints)-1]
	}

	// Map t to knot parameter range
	knotRange := bs.knots[len(bs.knots)-1] - bs.knots[0]
	u := t * knotRange

	// Find knot span
	span := bs.findKnotSpan(u)

	// Compute basis functions
	basis := bs.basisFunctions(span, u)

	// Compute weighted sum
	point := Vec3{}
	denominator := 0.0

	for i := 0; i <= bs.degree; i++ {
		ctrlIndex := span - bs.degree + i
		if ctrlIndex >= 0 && ctrlIndex < len(bs.controlPoints) {
			weight := bs.weights[ctrlIndex]
			weightedPoint := bs.controlPoints[ctrlIndex].Mul(weight * basis[i])
			point = point.Add(weightedPoint)
			denominator += weight * basis[i]
		}
	}

	if denominator > 0.0 {
		point = point.Div(denominator)
	}

	return point
}

// findKnotSpan finds knot span containing parameter u
func (bs *BSpline) findKnotSpan(u float64) int {
	n := len(bs.controlPoints) - 1
	p := bs.degree

	// Special cases
	if u >= bs.knots[n+1] {
		return n
	}
	if u <= bs.knots[p] {
		return p
	}

	// Binary search
	low := p
	high := n + p
	mid := (low + high) / 2

	for u < bs.knots[mid] || u >= bs.knots[mid+1] {
		if u < bs.knots[mid] {
			high = mid
		} else {
			low = mid
		}
		mid = (low + high) / 2
	}

	return mid
}

// basisFunctions computes non-zero basis functions for span at parameter u
func (bs *BSpline) basisFunctions(span int, u float64) []float64 {
	left := make([]float64, bs.order)
	right := make([]float64, bs.order)

	basis := make([]float64, bs.order)
	basis[0] = 1.0

	for j := 1; j <= bs.degree; j++ {
		left[j] = u - bs.knots[span+1-j]
		right[j] = bs.knots[span+j] - u

		saved := 0.0

		for r := 0; r < j; r++ {
			temp := basis[r] / (right[r+1] + left[j-r])
			basis[r] = saved + right[r+1]*temp
			saved = left[j-r] * temp
		}

		basis[j] = saved
	}

	return basis[1:]
}

// SamplePoints samples B-spline at specified number of points
func (bs *BSpline) SamplePoints(numPoints int) []Vec3 {
	if numPoints < 2 {
		return []Vec3{}
	}

	points := make([]Vec3, numPoints)
	for i := 0; i < numPoints; i++ {
		t := float64(i) / float64(numPoints-1)
		points[i] = bs.EvaluatePoint(t)
	}

	return points
}

// ControlPoints returns the control points
func (bs *BSpline) ControlPoints() []Vec3 {
	return append([]Vec3{}, bs.controlPoints...)
}

// SetControlPoints sets the control points
func (bs *BSpline) SetControlPoints(points []Vec3) {
	bs.controlPoints = points
}

// Degree returns the degree of the B-spline
func (bs *BSpline) Degree() int {
	return bs.degree
}

// String returns string representation
func (bs *BSpline) String() string {
	return fmt.Sprintf("BSpline{Degree: %d, ControlPoints: %d}", bs.degree, len(bs.controlPoints))
}

// Knots returns the knot vector
func (bs *BSpline) Knots() []float64 {
	result := make([]float64, len(bs.knots))
	copy(result, bs.knots)
	return result
}

// Weights returns the weight vector
func (bs *BSpline) Weights() []float64 {
	result := make([]float64, len(bs.weights))
	copy(result, bs.weights)
	return result
}

// Derivative calculates derivatives of the B-spline at parameter t
func (bs *BSpline) Derivative(t float64, order int) []Vec3 {
	if order < 1 {
		return []Vec3{}
	}

	// Simplified derivative calculation
	derivatives := make([]Vec3, order)

	// For now, implement basic first derivative
	if order >= 1 {
		h := 1e-6
		p1 := bs.EvaluatePoint(t - h)
		p2 := bs.EvaluatePoint(t + h)
		derivatives[0] = p2.Sub(p1).Div(2 * h)
	}

	// Higher order derivatives would require more sophisticated algorithms
	for i := 1; i < order; i++ {
		h := 1e-6
		d1 := bs.Derivative(t-h, order-1)
		d2 := bs.Derivative(t+h, order-1)
		if len(d1) > 0 && len(d2) > 0 {
			derivatives[i] = d2[0].Sub(d1[0]).Div(2 * h)
		} else {
			derivatives[i] = NewVec3(0, 0, 0)
		}
	}

	return derivatives
}

// Flattening converts the B-spline to line segments
func (bs *BSpline) Flattening(tolerance float64) []Vec3 {
	// Simple adaptive sampling
	points := []Vec3{}

	// Start and end points
	start := bs.EvaluatePoint(0.0)
	end := bs.EvaluatePoint(1.0)
	points = append(points, start)

	// Adaptive subdivision
	numSegments := 64
	for i := 1; i < numSegments; i++ {
		t := float64(i) / float64(numSegments)
		point := bs.EvaluatePoint(t)
		points = append(points, point)
	}

	points = append(points, end)
	return points
}
