// Package fit provides curve fitting and interpolation: least-squares
// polynomials, exact interpolation, cubic B-spline interpolation through
// measured points, Catmull-Rom interpolation, arc-length parameterization and
// least-squares circle fitting.
//
// Conventions: points are ordered by increasing parameter; parameter
// intervals are normalized to [0,1] so they line up with BSpline.EvaluatePoint
// and Bezier.EvaluatePoint; tolerances are absolute and supplied by callers
// (math.DefaultTolerance is the library default).
package fit

import (
	"errors"
	"fmt"
	"math"
	"slices"

	dxfmath "github.com/edanko/dxf/math"
)

// ErrSingularSystem is returned when the linear system built from the input
// points cannot be solved, which happens for degenerate point sets (repeated
// abscissae, coincident points, fewer distinct points than the requested
// degree).
var ErrSingularSystem = errors.New("fit: singular system, points are degenerate")

// FitPolynomial returns the coefficients of the least-squares polynomial of the
// requested degree for the supplied points, ordered from the constant term to
// the highest power. degree 1 fits a line, degree 2 a parabola.
func FitPolynomial(points []dxfmath.Vec2, degree int) ([]float64, error) {
	if degree < 0 {
		return nil, fmt.Errorf("degree must not be negative, got %d", degree)
	}
	if len(points) < degree+1 {
		return nil, fmt.Errorf("need at least %d points for degree %d, got %d", degree+1, degree, len(points))
	}
	size := degree + 1
	at := make([][]float64, size)
	for i := range at {
		at[i] = make([]float64, len(points))
		rhs := make([]float64, len(points))
		for j, p := range points {
			rhs[j] = pow(p.X(), i)
			at[i][j] = rhs[j]
		}
	}
	normal := make([][]float64, size)
	for i := range normal {
		normal[i] = make([]float64, size)
		for j := range size {
			normal[i][j] = dot(at[i], at[j])
		}
	}
	b := make([]float64, size)
	for i := range size {
		b[i] = dot(at[i], ys(points))
	}
	return solveDense(normal, b)
}

// EvaluatePolynomial evaluates a polynomial whose coefficients are ordered
// from the constant term to the highest power (the order FitPolynomial
// returns).
func EvaluatePolynomial(coeffs []float64, x float64) float64 {
	var result float64
	for _, coeff := range slices.Backward(coeffs) {
		result = result*x + coeff
	}
	return result
}

// InterpolatePolynomial returns the coefficients of the polynomial of degree
// len(points)-1 that passes through every point (Lagrange/Newton form solved
// as a Vandermonde system).
func InterpolatePolynomial(points []dxfmath.Vec2) ([]float64, error) {
	n := len(points)
	if n < 1 {
		return nil, fmt.Errorf("need at least 1 point")
	}
	matrix := make([][]float64, n)
	for i := range matrix {
		matrix[i] = make([]float64, n)
		for j := range n {
			matrix[i][j] = pow(points[i].X(), j)
		}
	}
	b := ys(points)
	return solveDense(matrix, b)
}

// InterpolateBSpline returns a clamped B-spline of the given degree that
// passes through every point. Parameters are chord-length based and
// normalized to [0,1], so EvaluatePoint(t) with t in [0,1] reproduces the
// input. The z component of each point is preserved.
func InterpolateBSpline(points []dxfmath.Vec2, degree int) (*dxfmath.BSpline, error) {
	spline, _, err := InterpolateBSplineWithParams(points, degree)
	return spline, err
}

// InterpolateBSplineWithParams is InterpolateBSpline plus the chord-length
// parameter of every input point. The spline passes through points[i] exactly
// at params[i]; callers that need to sample the curve at the data points use
// those parameters.
func InterpolateBSplineWithParams(points []dxfmath.Vec2, degree int) (*dxfmath.BSpline, []float64, error) {
	if degree < 1 {
		return nil, nil, fmt.Errorf("degree must be at least 1, got %d", degree)
	}
	if len(points) < degree+1 {
		return nil, nil, fmt.Errorf("need at least %d points for degree %d, got %d", degree+1, degree, len(points))
	}
	n := len(points)
	u := ChordalParameterize(points)
	knots := clampedKnots(u, degree)

	basis := make([][]float64, n)
	for i := range n {
		basis[i] = basisFunctionsAll(knots, degree, u[i])
	}
	control := make([][]float64, 3)
	for axis := range 3 {
		rhs := make([]float64, n)
		for i := range n {
			switch axis {
			case 0:
				rhs[i] = points[i].X()
			case 1:
				rhs[i] = points[i].Y()
			case 2:
				rhs[i] = 0
			}
		}
		solution, err := solveDense(copyMatrix(basis), rhs)
		if err != nil {
			return nil, nil, err
		}
		control[axis] = solution
	}
	cps := make([]dxfmath.Vec3, n)
	for i := range n {
		cps[i] = dxfmath.Vec3{control[0][i], control[1][i], control[2][i]}
	}
	spline := dxfmath.NewBSplineWithKnots(cps, nil, knots, degree)
	if spline == nil {
		return nil, nil, fmt.Errorf("failed to build interpolating B-spline")
	}
	return spline, u, nil
}

// InterpolateCatmullRom returns a polyline that passes through every input
// point, subdivided into samplesPerSegment points per span. Endpoints are
// duplicated so the curve starts and ends on the first and last point.
func InterpolateCatmullRom(points []dxfmath.Vec2, samplesPerSegment int) ([]dxfmath.Vec2, error) {
	if len(points) < 2 {
		return nil, fmt.Errorf("need at least 2 points, got %d", len(points))
	}
	if samplesPerSegment < 1 {
		samplesPerSegment = 8
	}
	out := make([]dxfmath.Vec2, 0, (len(points)-1)*samplesPerSegment+1)
	for i := 0; i+1 < len(points); i++ {
		p0 := points[maxInt(i-1, 0)]
		p1 := points[i]
		p2 := points[i+1]
		p3 := points[minInt(i+2, len(points)-1)]
		for s := 0; s < samplesPerSegment; s++ {
			t := float64(s) / float64(samplesPerSegment)
			out = append(out, catmullRom(p0, p1, p2, p3, t))
		}
	}
	return append(out, points[len(points)-1]), nil
}

// FitCircleLSQ returns the center and radius of the circle fitted to the points
// with the algebraic Kasa least-squares method. It returns ErrSingularSystem
// for collinear or coincident points.
func FitCircleLSQ(points []dxfmath.Vec2) (dxfmath.Vec2, float64, error) {
	if len(points) < 3 {
		return dxfmath.Vec2{}, 0, fmt.Errorf("need at least 3 points, got %d", len(points))
	}
	matrix := [][]float64{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}}
	b := make([]float64, 3)
	var sx, sy, sxx, syy, sxy, sxxy, sxyy, sxxx, syyy float64
	for _, p := range points {
		x, y := p.X(), p.Y()
		sx += x
		sy += y
		sxx += x * x
		syy += y * y
		sxy += x * y
		sxxy += x * x * y
		sxyy += x * y * y
		sxxx += x * x * x
		syyy += y * y * y
	}
	n := float64(len(points))
	matrix[0] = []float64{sxx, sxy, sx}
	matrix[1] = []float64{sxy, syy, sy}
	matrix[2] = []float64{sx, sy, n}
	// Normal equations of the linear system [x y 1]·[A1 A2 A3]ᵀ = x²+y², whose
	// solution is A1 = 2cx, A2 = 2cy, A3 = r² - cx² - cy².
	b[0] = sxxx + sxyy
	b[1] = sxxy + syyy
	var sq float64
	for _, p := range points {
		sq += p.X()*p.X() + p.Y()*p.Y()
	}
	b[2] = sq
	solution, err := solveDense(matrix, b)
	if err != nil {
		return dxfmath.Vec2{}, 0, ErrSingularSystem
	}
	cx, cy := solution[0]/2, solution[1]/2
	radius := math.Sqrt(math.Max(solution[2]+cx*cx+cy*cy, 0))
	return dxfmath.Vec2{cx, cy}, radius, nil
}

// ChordalParameterize returns the normalized cumulative chord length of the
// points, starting at 0 and ending at 1. Repeated points receive the same
// parameter as their predecessor.
func ChordalParameterize(points []dxfmath.Vec2) []float64 {
	return parameterize(points, 1)
}

// CentripetalParameterize returns the normalized cumulative square-root chord
// length, which reduces overshoot on point sets with sharp corners.
func CentripetalParameterize(points []dxfmath.Vec2) []float64 {
	return parameterize(points, 0.5)
}

// Perimeter returns the length of the polyline through the points.
func Perimeter(points []dxfmath.Vec2) float64 {
	var sum float64
	for i := 0; i+1 < len(points); i++ {
		sum += points[i].Distance(points[i+1])
	}
	return sum
}

// ClosedPerimeter returns the length of the closed polyline through the points.
func ClosedPerimeter(points []dxfmath.Vec2) float64 {
	if len(points) < 2 {
		return 0
	}
	return Perimeter(points) + points[len(points)-1].Distance(points[0])
}

func parameterize(points []dxfmath.Vec2, power float64) []float64 {
	out := make([]float64, len(points))
	if len(points) == 0 {
		return out
	}
	var total float64
	for i := 0; i+1 < len(points); i++ {
		total += math.Pow(points[i].Distance(points[i+1]), power)
	}
	if total <= 0 {
		for i := range out {
			out[i] = float64(i) / float64(len(points))
		}
		return out
	}
	var acc float64
	for i := 0; i+1 < len(points); i++ {
		acc += math.Pow(points[i].Distance(points[i+1]), power)
		out[i+1] = acc / total
	}
	out[len(out)-1] = 1
	return out
}

// clampedKnots builds a clamped knot vector of length n+degree+1 from
// normalized parameters: the first degree+1 knots are u[0], the last degree+1
// are u[n-1], and interior knots are the average of the preceding degree
// parameters.
func clampedKnots(u []float64, degree int) []float64 {
	n := len(u)
	knots := make([]float64, n+degree+1)
	for i := 0; i <= degree; i++ {
		knots[i] = u[0]
		knots[n+i] = u[n-1]
	}
	for j := 1; j <= n-degree-1; j++ {
		var sum float64
		for k := j; k <= j+degree-1; k++ {
			sum += u[k]
		}
		knots[j+degree] = sum / float64(degree)
	}
	return knots
}

// basisFunctionsAll evaluates every B-spline basis function of the given degree
// at t with the Cox-de Boor recursion over the full knot vector. The returned
// slice has len(knots)-degree-1 entries and sums to 1 inside the parameter
// domain. The right end of the domain is treated as included, so the last
// control point is reached at t = domain end.
func basisFunctionsAll(knots []float64, degree int, t float64) []float64 {
	n := len(knots) - degree - 1 // number of basis functions of this degree
	prev := make([]float64, len(knots)-1)
	for i := 0; i < len(prev); i++ {
		if knots[i] <= t && t < knots[i+1] {
			prev[i] = 1
		}
	}
	if n >= 1 && t >= knots[n] && t <= knots[n]+1e-12 {
		prev[n-1] = 1
	}
	for k := 1; k <= degree; k++ {
		cur := make([]float64, len(prev)-1)
		for i := range cur {
			dLeft := knots[i+k] - knots[i]
			dRight := knots[i+k+1] - knots[i+1]
			var left, right float64
			if dLeft > 0 {
				left = (t - knots[i]) / dLeft * prev[i]
			}
			if dRight > 0 {
				right = (knots[i+k+1] - t) / dRight * prev[i+1]
			}
			cur[i] = left + right
		}
		prev = cur
	}
	return prev
}

func catmullRom(p0, p1, p2, p3 dxfmath.Vec2, t float64) dxfmath.Vec2 {
	t2 := t * t
	t3 := t2 * t
	x := 0.5 * ((2 * p1.X()) + (-p0.X()+p2.X())*t +
		(2*p0.X()-5*p1.X()+4*p2.X()-p3.X())*t2 +
		(-p0.X()+3*p1.X()-3*p2.X()+p3.X())*t3)
	y := 0.5 * ((2 * p1.Y()) + (-p0.Y()+p2.Y())*t +
		(2*p0.Y()-5*p1.Y()+4*p2.Y()-p3.Y())*t2 +
		(-p0.Y()+3*p1.Y()-3*p2.Y()+p3.Y())*t3)
	return dxfmath.Vec2{x, y}
}

// solveDense solves A*x = b with Gaussian elimination and partial pivoting.
func solveDense(matrix [][]float64, b []float64) ([]float64, error) {
	n := len(b)
	if len(matrix) != n {
		return nil, ErrSingularSystem
	}
	for col := range n {
		pivot := col
		for row := col + 1; row < n; row++ {
			if abs(matrix[row][col]) > abs(matrix[pivot][col]) {
				pivot = row
			}
		}
		if abs(matrix[pivot][col]) <= 1e-12 {
			return nil, ErrSingularSystem
		}
		matrix[col], matrix[pivot] = matrix[pivot], matrix[col]
		b[col], b[pivot] = b[pivot], b[col]
		for row := col + 1; row < n; row++ {
			factor := matrix[row][col] / matrix[col][col]
			if factor == 0 {
				continue
			}
			for k := col; k < n; k++ {
				matrix[row][k] -= factor * matrix[col][k]
			}
			b[row] -= factor * b[col]
		}
	}
	solution := make([]float64, n)
	for row := n - 1; row >= 0; row-- {
		sum := b[row]
		for k := row + 1; k < n; k++ {
			sum -= matrix[row][k] * solution[k]
		}
		solution[row] = sum / matrix[row][row]
	}
	return solution, nil
}

func copyMatrix(m [][]float64) [][]float64 {
	out := make([][]float64, len(m))
	for i := range m {
		out[i] = make([]float64, len(m[i]))
		copy(out[i], m[i])
	}
	return out
}

func ys(points []dxfmath.Vec2) []float64 {
	out := make([]float64, len(points))
	for i, p := range points {
		out[i] = p.Y()
	}
	return out
}

func dot(a, b []float64) float64 {
	var sum float64
	for i := range a {
		sum += a[i] * b[i]
	}
	return sum
}

func pow(base float64, exponent int) float64 {
	result := 1.0
	for range exponent {
		result *= base
	}
	return result
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
