package math

import (
	"errors"
	"fmt"
	"math"
)

var (
	ErrInsufficientControlPoints = errors.New("at least 2 control points required for Bézier curve")
	ErrInvalidParameterRange     = errors.New("parameter t must be in range [0, 1]")
	ErrInvalidDerivativeOrder    = errors.New("invalid derivative order: must be between 0 and curve degree")
	ErrInvalidSplitParameter     = errors.New("split parameter t must be in range (0, 1)")
)

// Bezier represents a generic Bézier curve of any degree
type Bezier struct {
	controlPoints []Vec3
}

// NewBezier creates a new Bézier curve from control points.
// Returns an error if fewer than 2 control points are provided.
func NewBezier(controlPoints []Vec3) (*Bezier, error) {
	if len(controlPoints) < 2 {
		return nil, ErrInsufficientControlPoints
	}

	return &Bezier{
		controlPoints: controlPoints,
	}, nil
}

// NewBezier4P creates a cubic Bézier curve (degree 3) from 4 control points.
// Returns an error if the curve cannot be created.
func NewBezier4P(p0, p1, p2, p3 Vec3) (*Bezier, error) {
	return NewBezier([]Vec3{p0, p1, p2, p3})
}

// NewBezier3P creates a quadratic Bézier curve (degree 2) from 3 control points.
// Returns an error if the curve cannot be created.
func NewBezier3P(p0, p1, p2 Vec3) (*Bezier, error) {
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
		panic(fmt.Sprintf("%s: t=%v", ErrInvalidParameterRange.Error(), t))
	}

	if (1.0 - t) < 5e-6 {
		t = 1.0
	}

	n := len(b.controlPoints)
	point := Vec3{}

	for i := range n {
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

		midT := (t0 + t1) * 0.5
		midPoint := b.EvaluatePoint(midT)

		lerpPoint := startPoint.Lerp(endPoint, 0.5)

		if lerpPoint.Distance(midPoint) < distance {
			result = append(result, endPoint)
		} else {
			subResult := b.subdivide(t0, midT, t1, startPoint, midPoint, endPoint, distance, segments)
			result = append(result, subResult...)
		}

		t0 = t1
		startPoint = endPoint
	}

	return result
}

func (b *Bezier) subdivide(t1, midT, t2 float64, start, mid, end Vec3, distance float64, segments int) []Vec3 {
	var result []Vec3

	lerpPoint := start.Lerp(end, 0.5)
	if lerpPoint.Distance(mid) < distance {
		result = append(result, end)
		return result
	}

	midLeftT := (t1 + midT) * 0.5
	midLeftPoint := b.EvaluatePoint(midLeftT)
	midRightT := (midT + t2) * 0.5
	midRightPoint := b.EvaluatePoint(midRightT)

	leftResult := b.subdivide(t1, midLeftT, midT, start, midLeftPoint, mid, distance, segments)
	result = append(result, leftResult...)

	result = append(result, mid)

	rightResult := b.subdivide(midT, midRightT, t2, mid, midRightPoint, end, distance, segments)
	result = append(result, rightResult...)

	return result
}

// Derivative computes point and derivatives up to order n at parameter t
func (b *Bezier) Derivative(t float64, n int) []Vec3 {
	if n < 0 || n > b.Degree() {
		panic(fmt.Sprintf("%s: n=%d, degree=%d", ErrInvalidDerivativeOrder.Error(), n, b.Degree()))
	}

	derivatives := make([]Vec3, n+1)

	derivatives[0] = b.EvaluatePoint(t)

	if n >= 1 {
		derivatives[1] = b.firstDerivative(t)
	}

	if n >= 2 {
		derivatives[2] = b.secondDerivative(t)
	}

	return derivatives
}

func (b *Bezier) firstDerivative(t float64) Vec3 {
	n := b.Degree()
	if n == 0 {
		return Vec3{}
	}

	derivative := Vec3{}
	for i := range n {
		basis := bernsteinDerivative(n, i, t)
		weightedPoint := b.controlPoints[i].Mul(basis)
		derivative = derivative.Add(weightedPoint)
	}

	return derivative.Mul(float64(n))
}

func (b *Bezier) secondDerivative(t float64) Vec3 {
	n := b.Degree()
	if n < 2 {
		return Vec3{}
	}

	derivative := Vec3{}
	for i := range n {
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
func (b *Bezier) Reverse() (*Bezier, error) {
	points := make([]Vec3, len(b.controlPoints))
	for i, point := range b.controlPoints {
		points[len(b.controlPoints)-1-i] = point
	}
	return NewBezier(points)
}

// Split splits the Bézier curve at parameter t into two curves.
// Returns an error if t is not in the range (0, 1).
func (b *Bezier) Split(t float64) (*Bezier, *Bezier, error) {
	if t <= 0.0 || t >= 1.0 {
		return nil, nil, ErrInvalidSplitParameter
	}

	leftPoints := b.leftControlPoints(t)
	rightPoints := b.rightControlPoints(t)

	left, err := NewBezier(leftPoints)
	if err != nil {
		return nil, nil, err
	}
	right, err := NewBezier(rightPoints)
	if err != nil {
		return nil, nil, err
	}

	return left, right, nil
}

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
			t := float64(i-1) / float64(segments)
			if accumulatedLength > 0 {
				t += (targetLength - accumulatedLength) / segmentLength / float64(segments)
			}
			return b.EvaluatePoint(t)
		}
		accumulatedLength += segmentLength
	}

	return points[len(points)-1]
}

func bernsteinBasis(n, i int, t float64) float64 {
	if t == 0.0 && i == 0 {
		return 1.0
	}
	if n == i && t == 1.0 {
		return 1.0
	}

	binomial := binomialCoefficient(n, i)
	ti := math.Pow(t, float64(i))
	tni := math.Pow(1.0-t, float64(n-i))

	return float64(binomial) * ti * tni
}

func bernsteinDerivative(n, i int, t float64) float64 {
	if n == 0 {
		return 0.0
	}

	return float64(n) * (bernsteinBasis(n-1, i-1, t) - bernsteinBasis(n-1, i, t))
}

func bernsteinSecondDerivative(n, i int, t float64) float64 {
	if n < 2 {
		return 0.0
	}

	return float64(n*(n-1)) * (bernsteinBasis(n-2, i-2, t) - 2*bernsteinBasis(n-2, i-1, t) + bernsteinBasis(n-2, i, t))
}

func binomialCoefficient(n, k int) int {
	if k < 0 || k > n {
		return 0
	}
	if k == 0 || k == n {
		return 1
	}

	k = min(k, n-k)

	result := 1
	for i := 0; i < k; i++ {
		result = result * (n - i) / (i + 1)
	}

	return result
}

func deCasteljauCoefficients(i, j int, t float64) float64 {
	n := i + j
	if j == 0 {
		return math.Pow(1.0-t, float64(n-i))
	}
	if i == n {
		return math.Pow(t, float64(j))
	}

	coeff := float64(binomialCoefficient(n, j))
	for l := range j {
		coeff *= math.Pow(t, float64(l)) * math.Pow(1.0-t, float64(j-l))
	}
	for l := j; l < i; l++ {
		coeff *= math.Pow(t, float64(l)) * math.Pow(1.0-t, float64(n-l))
	}

	return coeff
}

func (b *Bezier) String() string {
	return fmt.Sprintf("Bezier{degree: %d, controlPoints: %d}", b.Degree(), len(b.controlPoints))
}
