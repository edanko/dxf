package math

import (
	"fmt"
	"math"
)

// Bezier represents a generic Bézier curve of any degree
type Bezier struct {
	controlPoints []Vec3
}

// NewBezier creates a new Bézier curve from control points
func NewBezier(controlPoints []Vec3) *Bezier {
	if len(controlPoints) < 2 {
		panic("at least 2 control points required for Bézier curve")
	}

	return &Bezier{
		controlPoints: controlPoints,
	}
}

// NewBezier4P creates a cubic Bézier curve (degree 3) from 4 control points
func NewBezier4P(p0, p1, p2, p3 Vec3) *Bezier {
	return NewBezier([]Vec3{p0, p1, p2, p3})
}

// NewBezier3P creates a quadratic Bézier curve (degree 2) from 3 control points
func NewBezier3P(p0, p1, p2 Vec3) *Bezier {
	return NewBezier([]Vec3{p0, p1, p2})
}

// Degree returns the degree of the Bézier curve
func (b *Bezier) Degree() int {
	return len(b.controlPoints) - 1
}

// ControlPoints returns a copy of the control points
func (b *Bezier) ControlPoints() []Vec3 {
	points := make([]Vec3, len(b.controlPoints))
	copy(points, b.controlPoints)
	return points
}

// EvaluatePoint evaluates the Bézier curve at parameter t in range [0, 1]
func (b *Bezier) EvaluatePoint(t float64) Vec3 {
	if t < 0.0 || t > 1.0 {
		panic("parameter t must be in range [0, 1]")
	}

	// Handle end cases efficiently
	if (1.0 - t) < 5e-6 {
		t = 1.0
	}

	n := len(b.controlPoints)
	point := Vec3{}

	for i := 0; i < n; i++ {
		basis := bernsteinBasis(n-1, i, t)
		weightedPoint := b.controlPoints[i].Mul(basis)
		point = point.Add(weightedPoint)
	}

	return point
}

// Approximate approximates the Bézier curve with specified number of segments
func (b *Bezier) Approximate(segments int) []Vec3 {
	if segments < 1 {
		return []Vec3{}
	}

	points := make([]Vec3, segments+1)
	for i := 0; i <= segments; i++ {
		t := float64(i) / float64(segments)
		points[i] = b.EvaluatePoint(t)
	}

	return points
}

// Flattening performs adaptive recursive flattening of the Bézier curve
func (b *Bezier) Flattening(distance float64, segments int) []Vec3 {
	if segments < 4 {
		segments = 4
	}

	var result []Vec3
	dt := 1.0 / float64(segments)
	t0 := 0.0
	startPoint := b.controlPoints[0]
	result = append(result, startPoint)

	for t0 < 1.0 {
		t1 := t0 + dt
		if t1 >= 1.0 {
			t1 = 1.0
		}

		endPoint := b.EvaluatePoint(t1)

		// Check if we need to subdivide further
		midT := (t0 + t1) * 0.5
		midPoint := b.EvaluatePoint(midT)

		// Linear interpolation point
		lerpPoint := startPoint.Lerp(endPoint, 0.5)

		if lerpPoint.Distance(midPoint) < distance {
			result = append(result, endPoint)
		} else {
			// Subdivide recursively
			subResult := b.subdivide(t0, midT, t1, startPoint, midPoint, endPoint, distance, segments)
			result = append(result, subResult...)
		}

		t0 = t1
		startPoint = endPoint
	}

	return result
}

// subdivide is a helper for adaptive subdivision
func (b *Bezier) subdivide(t1, midT, t2 float64, start, mid, end Vec3, distance float64, segments int) []Vec3 {
	var result []Vec3

	// Check distance from curve to linear approximation
	lerpPoint := start.Lerp(end, 0.5)
	if lerpPoint.Distance(mid) < distance {
		result = append(result, end)
		return result
	}

	// Subdivide further
	midLeftT := (t1 + midT) * 0.5
	midLeftPoint := b.EvaluatePoint(midLeftT)
	midRightT := (midT + t2) * 0.5
	midRightPoint := b.EvaluatePoint(midRightT)

	// Recursively subdivide left segment
	leftResult := b.subdivide(t1, midLeftT, midT, start, midLeftPoint, mid, distance, segments)
	result = append(result, leftResult...)

	// Add middle point
	result = append(result, mid)

	// Recursively subdivide right segment
	rightResult := b.subdivide(midT, midRightT, t2, mid, midRightPoint, end, distance, segments)
	result = append(result, rightResult...)

	return result
}

// Derivative computes point and derivatives up to order n at parameter t
func (b *Bezier) Derivative(t float64, n int) []Vec3 {
	if n < 0 || n > b.Degree() {
		panic("invalid derivative order")
	}

	derivatives := make([]Vec3, n+1)

	// Zeroth derivative is the point itself
	derivatives[0] = b.EvaluatePoint(t)

	if n >= 1 {
		derivatives[1] = b.firstDerivative(t)
	}

	if n >= 2 {
		derivatives[2] = b.secondDerivative(t)
	}

	return derivatives
}

// firstDerivative computes the first derivative of the Bézier curve
func (b *Bezier) firstDerivative(t float64) Vec3 {
	n := b.Degree()
	if n == 0 {
		return Vec3{} // Linear curve has constant derivative
	}

	derivative := Vec3{}
	for i := 0; i < n; i++ {
		basis := bernsteinDerivative(n, i, t)
		weightedPoint := b.controlPoints[i].Mul(basis)
		derivative = derivative.Add(weightedPoint)
	}

	return derivative.Mul(float64(n))
}

// secondDerivative computes the second derivative of the Bézier curve
func (b *Bezier) secondDerivative(t float64) Vec3 {
	n := b.Degree()
	if n < 2 {
		return Vec3{} // Linear and quadratic curves have zero second derivative
	}

	derivative := Vec3{}
	for i := 0; i < n; i++ {
		basis := bernsteinSecondDerivative(n, i, t)
		weightedPoint := b.controlPoints[i].Mul(basis)
		derivative = derivative.Add(weightedPoint)
	}

	return derivative.Mul(float64(n * (n - 1)))
}

// Tangent computes the tangent vector (first derivative) at parameter t
func (b *Bezier) Tangent(t float64) Vec3 {
	return b.firstDerivative(t).Normalized()
}

// Normal computes the normal vector (second derivative) at parameter t
func (b *Bezier) Normal(t float64) Vec3 {
	return b.secondDerivative(t).Normalized()
}

// Reverse returns a new Bézier curve with reversed control point order
func (b *Bezier) Reverse() *Bezier {
	points := make([]Vec3, len(b.controlPoints))
	for i, point := range b.controlPoints {
		points[len(b.controlPoints)-1-i] = point
	}
	return NewBezier(points)
}

// Split splits the Bézier curve at parameter t into two curves
func (b *Bezier) Split(t float64) (*Bezier, *Bezier) {
	if t <= 0.0 || t >= 1.0 {
		panic("split parameter t must be in range (0, 1)")
	}

	leftPoints := b.leftControlPoints(t)
	rightPoints := b.rightControlPoints(t)

	return NewBezier(leftPoints), NewBezier(rightPoints)
}

// leftControlPoints computes control points for the left part of split curve
func (b *Bezier) leftControlPoints(t float64) []Vec3 {
	n := b.Degree()
	leftPoints := make([]Vec3, n+1)

	for i := 0; i <= n; i++ {
		leftPoints[i] = Vec3{}
		for j := 0; j <= i; j++ {
			blend := deCasteljauCoefficients(i, j, t)
			leftPoints[i] = leftPoints[i].Add(b.controlPoints[j].Mul(blend))
		}
	}

	return leftPoints
}

// rightControlPoints computes control points for the right part of split curve
func (b *Bezier) rightControlPoints(t float64) []Vec3 {
	n := b.Degree()
	rightPoints := make([]Vec3, n+1)

	for i := 0; i <= n; i++ {
		rightPoints[i] = Vec3{}
		for j := i; j <= n; j++ {
			blend := deCasteljauCoefficients(n-i, n-j, t)
			rightPoints[i] = rightPoints[i].Add(b.controlPoints[j].Mul(blend))
		}
	}

	return rightPoints
}

// Length computes the approximate length of the Bézier curve using numerical integration
func (b *Bezier) Length(segments int) float64 {
	if segments < 1 {
		segments = 20
	}

	points := b.Approximate(segments)
	length := 0.0

	for i := 1; i < len(points); i++ {
		length += points[i].Distance(points[i-1])
	}

	return length
}

// PointAtLength finds the point on the curve at a given arc length
func (b *Bezier) PointAtLength(targetLength float64, segments int) Vec3 {
	if segments < 1 {
		segments = 20
	}

	points := b.Approximate(segments)
	accumulatedLength := 0.0

	for i := 1; i < len(points); i++ {
		segmentLength := points[i].Distance(points[i-1])
		if accumulatedLength+segmentLength >= targetLength {
			// Linear interpolation within this segment
			t := float64(i-1) / float64(segments)
			if accumulatedLength > 0 {
				t += (targetLength - accumulatedLength) / segmentLength / float64(segments)
			}
			return b.EvaluatePoint(t)
		}
		accumulatedLength += segmentLength
	}

	// Return the last point if target length exceeds curve length
	return points[len(points)-1]
}

// bernsteinBasis computes the Bernstein basis function B_i,n(t)
func bernsteinBasis(n, i int, t float64) float64 {
	// Handle special cases to avoid domain problems
	if t == 0.0 && i == 0 {
		return 1.0
	}
	if n == i && t == 1.0 {
		return 1.0
	}

	// Use binomial coefficient and powers
	binomial := binomialCoefficient(n, i)
	ti := math.Pow(t, float64(i))
	tni := math.Pow(1.0-t, float64(n-i))

	return float64(binomial) * ti * tni
}

// bernsteinDerivative computes the derivative of Bernstein basis function
func bernsteinDerivative(n, i int, t float64) float64 {
	if n == 0 {
		return 0.0
	}

	return float64(n) * (bernsteinBasis(n-1, i-1, t) - bernsteinBasis(n-1, i, t))
}

// bernsteinSecondDerivative computes the second derivative of Bernstein basis function
func bernsteinSecondDerivative(n, i int, t float64) float64 {
	if n < 2 {
		return 0.0
	}

	return float64(n*(n-1)) * (bernsteinBasis(n-2, i-2, t) - 2*bernsteinBasis(n-2, i-1, t) + bernsteinBasis(n-2, i, t))
}

// binomialCoefficient computes binomial coefficient C(n, k)
func binomialCoefficient(n, k int) int {
	if k < 0 || k > n {
		return 0
	}
	if k == 0 || k == n {
		return 1
	}

	// Compute C(n, k) = C(n, n-k) to minimize computation
	k = min(k, n-k)

	result := 1
	for i := 0; i < k; i++ {
		result = result * (n - i) / (i + 1)
	}

	return result
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// deCasteljauCoefficients computes de Casteljau's algorithm coefficients
func deCasteljauCoefficients(i, j int, t float64) float64 {
	n := i + j
	if j == 0 {
		return math.Pow(1.0-t, float64(n-i))
	}
	if i == n {
		return math.Pow(t, float64(j))
	}

	// General case
	coeff := float64(binomialCoefficient(n, j))
	for l := 0; l < j; l++ {
		coeff *= math.Pow(t, float64(l)) * math.Pow(1.0-t, float64(j-l))
	}
	for l := j; l < i; l++ {
		coeff *= math.Pow(t, float64(l)) * math.Pow(1.0-t, float64(n-l))
	}

	return coeff
}

// String returns string representation
func (b *Bezier) String() string {
	return fmt.Sprintf("Bezier{degree: %d, controlPoints: %d}", b.Degree(), len(b.controlPoints))
}
