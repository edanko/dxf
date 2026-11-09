package fit

import (
	"math"
	"testing"

	dxfmath "github.com/edanko/dxf/math"
)

func nearlyEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}

func TestFitPolynomialLine(t *testing.T) {
	points := make([]dxfmath.Vec2, 0, 5)
	for i := 0; i < 5; i++ {
		x := float64(i)
		points = append(points, dxfmath.Vec2{x, 2*x + 1})
	}
	coeffs, err := FitPolynomial(points, 1)
	if err != nil {
		t.Fatalf("FitPolynomial: %v", err)
	}
	if len(coeffs) != 2 {
		t.Fatalf("coefficients = %v, want [intercept slope]", coeffs)
	}
	if !nearlyEqual(coeffs[0], 1, 1e-9) || !nearlyEqual(coeffs[1], 2, 1e-9) {
		t.Fatalf("coefficients = %v, want [1 2]", coeffs)
	}
	if got := EvaluatePolynomial(coeffs, 10); !nearlyEqual(got, 21, 1e-9) {
		t.Fatalf(" EvaluatePolynomial(10) = %v, want 21", got)
	}
}

func TestFitPolynomialParabola(t *testing.T) {
	points := make([]dxfmath.Vec2, 0, 7)
	for i := -3; i <= 3; i++ {
		x := float64(i)
		points = append(points, dxfmath.Vec2{x, 3 - 2*x + x*x/2})
	}
	coeffs, err := FitPolynomial(points, 2)
	if err != nil {
		t.Fatalf("FitPolynomial: %v", err)
	}
	for i := range points {
		x := points[i].X()
		want := points[i].Y()
		if got := EvaluatePolynomial(coeffs, x); !nearlyEqual(got, want, 1e-6) {
			t.Fatalf("y(%v) = %v, want %v", x, got, want)
		}
	}
}

func TestInterpolatePolynomialPassesThroughPoints(t *testing.T) {
	points := []dxfmath.Vec2{{-1, 4}, {2, 1}, {5, 9}, {10, 2}}
	coeffs, err := InterpolatePolynomial(points)
	if err != nil {
		t.Fatalf("InterpolatePolynomial: %v", err)
	}
	if len(coeffs) != len(points) {
		t.Fatalf("coefficient count = %d, want %d", len(coeffs), len(points))
	}
	for _, p := range points {
		if got := EvaluatePolynomial(coeffs, p.X()); !nearlyEqual(got, p.Y(), 1e-6) {
			t.Fatalf("y(%v) = %v, want %v", p.X(), got, p.Y())
		}
	}
}

func TestInterpolateBSplinePassesThroughPoints(t *testing.T) {
	spline, params, err := InterpolateBSplineWithParams([]dxfmath.Vec2{
		{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0},
	}, 3)
	if err != nil {
		t.Fatalf("InterpolateBSplineWithParams: %v", err)
	}
	if spline == nil {
		t.Fatal("nil spline")
	}
	if len(params) != 5 {
		t.Fatalf("params = %v, want 5 entries", params)
	}
	if !spline.IsClamped() {
		t.Error("interpolating spline is not clamped")
	}
	for i, p := range []dxfmath.Vec2{{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0}} {
		got := spline.EvaluatePoint(params[i])
		if got.XY().Distance(p) > 1e-7 {
			t.Fatalf("P(params[%d]=%v) = %v, want %v", i, params[i], got, p)
		}
	}
	mid := spline.EvaluatePoint((params[1] + params[2]) / 2)
	if mid.Distance(dxfmath.Vec3{1, 0, 0}) <= 1e-9 {
		t.Errorf("curve collapsed at the first segment: P = %v", mid)
	}
}

func TestInterpolateBSplineDegreeOneIsPolyline(t *testing.T) {
	points := []dxfmath.Vec2{{0, 0}, {2, 1}, {4, 0}}
	spline, params, err := InterpolateBSplineWithParams(points, 1)
	if err != nil {
		t.Fatalf("InterpolateBSplineWithParams: %v", err)
	}
	for i, p := range points {
		got := spline.EvaluatePoint(params[i])
		if got.XY().Distance(p) > 1e-9 {
			t.Fatalf("P(%v) = %v, want %v", params[i], got, p)
		}
	}
}

func TestInterpolateCatmullRomPassesThroughPoints(t *testing.T) {
	points := []dxfmath.Vec2{{0, 0}, {1, 2}, {3, 3}, {4, 1}}
	pts, err := InterpolateCatmullRom(points, 8)
	if err != nil {
		t.Fatalf("InterpolateCatmullRom: %v", err)
	}
	if len(pts) < 2 {
		t.Fatalf("too few samples: %d", len(pts))
	}
	for _, p := range points {
		found := false
		for _, q := range pts {
			if nearlyEqual(p.X(), q.X(), 1e-9) && nearlyEqual(p.Y(), q.Y(), 1e-9) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("input point %v missing from samples", p)
		}
	}
}

func TestFitCircleLSQOnSampledCircle(t *testing.T) {
	center := dxfmath.Vec2{1, -2}
	const radius = 3
	points := make([]dxfmath.Vec2, 0, 24)
	for i := 0; i < 24; i++ {
		angle := 2 * math.Pi * float64(i) / 24
		points = append(points, dxfmath.Vec2{center.X() + radius*math.Cos(angle), center.Y() + radius*math.Sin(angle)})
	}
	got, gotR, err := FitCircleLSQ(points)
	if err != nil {
		t.Fatalf("FitCircleLSQ: %v", err)
	}
	if got.Distance(center) > 1e-6 {
		t.Errorf("center = %v, want %v", got, center)
	}
	if !nearlyEqual(gotR, radius, 1e-6) {
		t.Errorf("radius = %v, want %v", gotR, radius)
	}
}

func TestPerimeterUnitSquare(t *testing.T) {
	square := []dxfmath.Vec2{{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0}}
	if got := Perimeter(square); !nearlyEqual(got, 4, 1e-12) {
		t.Fatalf("Perimeter = %v, want 4", got)
	}
	open := []dxfmath.Vec2{{0, 0}, {1, 0}, {1, 1}}
	if got := Perimeter(open); !nearlyEqual(got, 2, 1e-12) {
		t.Fatalf("Perimeter(open) = %v, want 2", got)
	}
	if got := ClosedPerimeter(open); !nearlyEqual(got, 2+math.Sqrt2, 1e-9) {
		t.Fatalf("ClosedPerimeter = %v, want %v", got, 2+math.Sqrt2)
	}
}

func TestParameterizeMonotonic(t *testing.T) {
	points := []dxfmath.Vec2{{0, 0}, {3, 0}, {3, 4}, {6, 4}}
	chordal := ChordalParameterize(points)
	if len(chordal) != len(points) {
		t.Fatalf("length = %d, want %d", len(chordal), len(points))
	}
	if chordal[0] != 0 || !nearlyEqual(chordal[len(chordal)-1], 1, 1e-12) {
		t.Fatalf("not normalized: %v", chordal)
	}
	for i := 1; i < len(chordal); i++ {
		if chordal[i] <= chordal[i-1] {
			t.Fatalf("not increasing at %d: %v", i, chordal)
		}
	}
	centripetal := CentripetalParameterize(points)
	for i := 1; i < len(centripetal); i++ {
		if centripetal[i] <= centripetal[i-1] {
			t.Fatalf("centripetal not increasing at %d: %v", i, centripetal)
		}
	}
	// Chordal spacing is proportional to length, centripetal is not.
	if nearlyEqual(chordal[1]-chordal[0], centripetal[1]-centripetal[0], 1e-12) {
		t.Log("centripetal matches chordal spacing")
	}
}

func TestFitPolynomialDegeneratePoints(t *testing.T) {
	if _, err := FitPolynomial([]dxfmath.Vec2{{1, 1}}, 1); err == nil {
		t.Fatal("expected error for 1 point degree 1")
	}
	collinear := []dxfmath.Vec2{{0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}}
	if _, err := FitPolynomial(collinear, 5); err == nil {
		t.Fatal("expected singular system for coincident points")
	}
	if _, _, err := FitCircleLSQ([]dxfmath.Vec2{{0, 0}, {1, 0}, {2, 0}}); err == nil {
		t.Fatal("expected error for collinear circle fit")
	}
	if _, err := InterpolateBSpline([]dxfmath.Vec2{{0, 0}, {1, 1}}, 3); err == nil {
		t.Fatal("expected error for too few points")
	}
	if _, err := InterpolateBSpline([]dxfmath.Vec2{{0, 0}, {1, 1}, {2, 2}}, 0); err == nil {
		t.Fatal("expected error for degree 0")
	}
}
