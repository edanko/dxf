package math

import (
	"math"
)

// NURBS represents a Non-Uniform Rational B-Spline curve
type NURBS struct {
	*BSpline
	controlPoints []Vec3
	weights       []float64
	knots         []float64
	degree        int
	isRational    bool
}

// NewNURBS creates a new NURBS curve
func NewNURBS(controlPoints []Vec3, weights []float64, knots []float64, degree int) *NURBS {
	if len(controlPoints) != len(weights) {
		return nil
	}

	nurbs := &NURBS{
		controlPoints: make([]Vec3, len(controlPoints)),
		weights:       make([]float64, len(weights)),
		knots:         make([]float64, len(knots)),
		degree:        degree,
		isRational:    false,
	}

	copy(nurbs.controlPoints, controlPoints)
	copy(nurbs.weights, weights)
	copy(nurbs.knots, knots)

	// Check if curve is rational (has non-uniform weights)
	for _, w := range weights {
		if math.Abs(w-1.0) > 1e-10 {
			nurbs.isRational = true
			break
		}
	}

	// Create underlying B-spline
	nurbs.BSpline = NewBSpline(controlPoints, degree)

	return nurbs
}

// NewNURBSUniform creates a NURBS with uniform knot vector
func NewNURBSUniform(controlPoints []Vec3, weights []float64, degree int) *NURBS {
	n := len(controlPoints)
	m := n + degree + 1

	// Create uniform knot vector
	knots := make([]float64, m)
	for i := 0; i < m; i++ {
		if i <= degree {
			knots[i] = 0.0
		} else if i >= n {
			knots[i] = float64(m - 2*degree - 1)
		} else {
			knots[i] = float64(i - degree)
		}
	}

	return NewNURBS(controlPoints, weights, knots, degree)
}

// Evaluate returns a point on the NURBS curve at parameter t
func (nurbs *NURBS) Evaluate(t float64) Vec3 {
	if !nurbs.isRational {
		// Non-rational case: use standard B-spline evaluation
		return nurbs.BSpline.EvaluatePoint(t)
	}

	// Rational case: weighted B-spline evaluation
	n := len(nurbs.controlPoints)
	degree := nurbs.degree

	// Find knot span
	span := nurbs.findKnotSpan(t)

	// Compute basis functions
	basis := nurbs.computeBasisFunctions(span, degree, t)

	// Compute weighted sum
	var weightedSum Vec3
	var weightSum float64

	for i := 0; i <= degree; i++ {
		controlIndex := span - degree + i
		if controlIndex >= 0 && controlIndex < n {
			weight := nurbs.weights[controlIndex]
			basisValue := basis[i]

			weightedPoint := nurbs.controlPoints[controlIndex].Mul(weight * basisValue)
			weightedSum = weightedSum.Add(weightedPoint)
			weightSum += weight * basisValue
		}
	}

	if weightSum > 1e-10 {
		return weightedSum.Div(weightSum)
	}

	return Vec3{}
}

// EvaluateDerivative returns first derivative at parameter t
func (nurbs *NURBS) EvaluateDerivative(t float64) Vec3 {
	if !nurbs.isRational {
		// Get derivative using BSpline method
		ders := nurbs.BSpline.Derivative(t, 1)
		if len(ders) > 0 {
			return ders[0]
		}
		return Vec3{}
	}

	// Rational derivative using quotient rule
	dt := 1e-6
	p1 := nurbs.Evaluate(t)
	p2 := nurbs.Evaluate(t + dt)

	return p2.Sub(p1).Div(dt)
}

// EvaluateSecondDerivative returns second derivative at parameter t
func (nurbs *NURBS) EvaluateSecondDerivative(t float64) Vec3 {
	if !nurbs.isRational {
		// Get second derivative using BSpline method
		ders := nurbs.BSpline.Derivative(t, 2)
		if len(ders) > 1 {
			return ders[1]
		}
		return Vec3{}
	}

	// Rational second derivative
	dt := 1e-6
	d1 := nurbs.EvaluateDerivative(t)
	d2 := nurbs.EvaluateDerivative(t + dt)

	return d2.Sub(d1).Div(dt)
}

// findKnotSpan finds knot span index for parameter t
func (nurbs *NURBS) findKnotSpan(t float64) int {
	n := len(nurbs.controlPoints)
	degree := nurbs.degree

	// Special cases
	if t >= nurbs.knots[n] {
		return n - 1
	}
	if t <= nurbs.knots[degree] {
		return degree
	}

	// Binary search
	low := degree
	high := n
	mid := (low + high) / 2

	for t < nurbs.knots[mid] || t >= nurbs.knots[mid+1] {
		if t < nurbs.knots[mid] {
			high = mid
		} else {
			low = mid
		}
		mid = (low + high) / 2
	}

	return mid
}

// computeBasisFunctions computes B-spline basis functions
func (nurbs *NURBS) computeBasisFunctions(span, degree int, t float64) []float64 {
	left := make([]float64, degree+1)
	right := make([]float64, degree+1)
	basis := make([]float64, degree+1)

	basis[0] = 1.0

	for j := 1; j <= degree; j++ {
		left[j] = t - nurbs.knots[span+1-j]
		right[j] = nurbs.knots[span+j] - t
		saved := 0.0

		for r := 0; r < j; r++ {
			temp := basis[r] / (right[r+1] + left[j-r])
			basis[r] = saved + right[r+1]*temp
			saved = left[j-r] * temp
		}

		basis[j] = saved
	}

	return basis
}

// SamplePoints samples the NURBS curve at specified number of points
func (nurbs *NURBS) SamplePoints(numPoints int) []Vec3 {
	if numPoints < 2 {
		return []Vec3{}
	}

	points := make([]Vec3, numPoints)

	// Get parameter range
	tMin := nurbs.knots[nurbs.degree]
	tMax := nurbs.knots[len(nurbs.controlPoints)]

	for i := 0; i < numPoints; i++ {
		t := tMin + float64(i)/(float64(numPoints-1))*(tMax-tMin)
		points[i] = nurbs.Evaluate(t)
	}

	return points
}

// GetControlPoints returns the control points
func (nurbs *NURBS) GetControlPoints() []Vec3 {
	return nurbs.controlPoints
}

// GetWeights returns the control point weights
func (nurbs *NURBS) GetWeights() []float64 {
	return nurbs.weights
}

// GetKnots returns the knot vector
func (nurbs *NURBS) GetKnots() []float64 {
	return nurbs.knots
}

// GetDegree returns the curve degree
func (nurbs *NURBS) GetDegree() int {
	return nurbs.degree
}

// IsRational returns true if the curve has rational weights
func (nurbs *NURBS) IsRational() bool {
	return nurbs.isRational
}

// Length calculates the approximate length of the NURBS curve
func (nurbs *NURBS) Length() float64 {
	// Approximate length using sampling
	samples := nurbs.SamplePoints(100)
	length := 0.0

	for i := 0; i < len(samples)-1; i++ {
		length += samples[i].Distance(samples[i+1])
	}

	return length
}

// CreateCircleNURBS creates a circular arc using NURBS
func CreateCircleNURBS(center Vec3, radius float64, startAngle, endAngle float64, numSegments int) *NURBS {
	// Full circle requires 9 control points for exact NURBS representation
	// For arc, we use fewer control points

	angleSpan := endAngle - startAngle
	if angleSpan < 0 {
		angleSpan += 2 * math.Pi
	}

	// Create control points for circular arc
	controlPoints := make([]Vec3, numSegments)
	weights := make([]float64, numSegments)

	for i := 0; i < numSegments; i++ {
		t := float64(i) / float64(numSegments-1)
		angle := startAngle + t*angleSpan

		x := center.X() + radius*math.Cos(angle)
		y := center.Y() + radius*math.Sin(angle)
		z := center.Z()

		controlPoints[i] = NewVec3(x, y, z)

		// Weight for circular arc (cosine weighting)
		weights[i] = math.Cos(angle - startAngle - angleSpan/2)
		if weights[i] < 0 {
			weights[i] = 0.0
		}
	}

	// Create appropriate knot vector
	degree := 2
	m := numSegments + degree + 1
	knots := make([]float64, m)

	for i := 0; i < m; i++ {
		if i <= degree {
			knots[i] = 0.0
		} else if i >= numSegments {
			knots[i] = 1.0
		} else {
			knots[i] = float64(i-degree) / float64(numSegments-degree)
		}
	}

	return NewNURBS(controlPoints, weights, knots, degree)
}

// Validate checks if the NURBS curve is valid
func (nurbs *NURBS) Validate() bool {
	if nurbs == nil {
		return false
	}

	if len(nurbs.controlPoints) != len(nurbs.weights) {
		return false
	}

	if len(nurbs.controlPoints) < nurbs.degree+1 {
		return false
	}

	if len(nurbs.knots) != len(nurbs.controlPoints)+nurbs.degree+1 {
		return false
	}

	// Check knot vector monotonicity
	for i := 1; i < len(nurbs.knots); i++ {
		if nurbs.knots[i] < nurbs.knots[i-1] {
			return false
		}
	}

	// Check for valid weights
	for _, w := range nurbs.weights {
		if w < 0 || math.IsNaN(w) || math.IsInf(w, 0) {
			return false
		}
	}

	return true
}
