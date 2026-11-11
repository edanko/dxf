package math

import (
	"math"
	"testing"
)

// sampleCurve evaluates points at uniform parameters over [0,1].
func sampleCurve(bs *BSpline, count int) []Vec3 {
	out := make([]Vec3, count)
	for i := range out {
		out[i] = bs.EvaluatePoint(float64(i) / float64(count-1))
	}
	return out
}

func maxDistance(a, b []Vec3) float64 {
	best := 0.0
	for i := range a {
		if d := a[i].Distance(b[i]); d > best {
			best = d
		}
	}
	return best
}

func testControlPoints() []Vec3 {
	return []Vec3{{0, 0, 0}, {1, 2, 0}, {3, 3, 0}, {4, 1, 0}, {6, 2, 0}}
}

func TestClampedBSplineForm(t *testing.T) {
	cps := testControlPoints()
	bs, err := NewClampedBSpline(cps, 3)
	if err != nil {
		t.Fatalf("NewClampedBSpline: %v", err)
	}
	if !bs.IsClamped() {
		t.Fatalf("expected clamped knot vector, got %v", bs.Knots())
	}
	if len(bs.Knots()) != len(cps)+bs.Degree()+1 {
		t.Fatalf("knot count = %d, want %d", len(bs.Knots()), len(cps)+bs.Degree()+1)
	}
	// A clamped curve interpolates its first and last control points.
	if got := bs.EvaluatePoint(0); got.Distance(cps[0]) > DefaultTolerance {
		t.Errorf("start = %v, want %v", got, cps[0])
	}
	if got := bs.EvaluatePoint(1); got.Distance(cps[len(cps)-1]) > DefaultTolerance {
		t.Errorf("end = %v, want %v", got, cps[len(cps)-1])
	}
	if _, err := NewClampedBSpline(cps[:2], 3); err == nil {
		t.Error("expected error for too few control points")
	}
}

func TestPeriodicBSplineForm(t *testing.T) {
	cps := testControlPoints()
	bs, err := NewPeriodicBSpline(cps, 3)
	if err != nil {
		t.Fatalf("NewPeriodicBSpline: %v", err)
	}
	if bs.IsClamped() {
		t.Errorf("periodic form must not repeat the end knots: %v", bs.Knots())
	}
	if len(bs.ControlPoints()) != len(cps) {
		t.Fatalf("control point count = %d, want %d", len(bs.ControlPoints()), len(cps))
	}
	for _, p := range sampleCurve(bs, 25) {
		if math.IsNaN(p.X()) || math.IsNaN(p.Y()) {
			t.Fatalf("periodic evaluation produced NaN: %v", p)
		}
	}
}

func TestInsertKnotPreservesShape(t *testing.T) {
	bs, err := NewClampedBSpline(testControlPoints(), 3)
	if err != nil {
		t.Fatalf("NewClampedBSpline: %v", err)
	}
	before := sampleCurve(bs, 61)

	inserted, err := bs.InsertKnot(0.37)
	if err != nil {
		t.Fatalf("InsertKnot: %v", err)
	}
	if len(inserted.ControlPoints()) != len(bs.ControlPoints())+1 {
		t.Errorf("control points after insertion = %d, want %d",
			len(inserted.ControlPoints()), len(bs.ControlPoints())+1)
	}
	if len(inserted.Knots()) != len(bs.Knots())+1 {
		t.Errorf("knots after insertion = %d, want %d", len(inserted.Knots()), len(bs.Knots())+1)
	}
	after := sampleCurve(inserted, 61)
	if d := maxDistance(before, after); d > 1e-9 {
		t.Errorf("insertion changed the curve, max deviation %v", d)
	}
	// Inserting the same knot a second time refines further without moving the curve.
	twice, err := inserted.InsertKnot(0.37)
	if err != nil {
		t.Fatalf("second InsertKnot: %v", err)
	}
	if d := maxDistance(before, sampleCurve(twice, 61)); d > 1e-9 {
		t.Errorf("repeated insertion changed the curve, max deviation %v", d)
	}
	if d := inserted.EvaluatePoint(0.37).Distance(twice.EvaluatePoint(0.37)); d > 1e-9 {
		t.Errorf("knot insertion at a double knot moved the curve by %v", d)
	}
}

func TestInsertKnotsPreservesShape(t *testing.T) {
	bs, err := NewClampedBSpline(testControlPoints(), 3)
	if err != nil {
		t.Fatalf("NewClampedBSpline: %v", err)
	}
	before := sampleCurve(bs, 81)
	inserted, err := bs.InsertKnots([]float64{0.15, 0.45, 0.85})
	if err != nil {
		t.Fatalf("InsertKnots: %v", err)
	}
	if len(inserted.ControlPoints()) != len(bs.ControlPoints())+3 {
		t.Errorf("control points = %d, want %d", len(inserted.ControlPoints()), len(bs.ControlPoints())+3)
	}
	if d := maxDistance(before, sampleCurve(inserted, 81)); d > 1e-9 {
		t.Errorf("InsertKnots changed the curve, max deviation %v", d)
	}
	if _, err := bs.InsertKnots([]float64{2}); err == nil {
		t.Error("expected out-of-domain knot rejection")
	}
}

func TestRationalCircleIsExact(t *testing.T) {
	center := Vec3{1, 2, 0}
	const radius = 3.0
	circle, err := NewRationalCircle(center, radius)
	if err != nil {
		t.Fatalf("NewRationalCircle: %v", err)
	}
	if !circle.IsRational() {
		t.Fatal("circle must be rational")
	}
	if circle.GetDegree() != 2 {
		t.Errorf("degree = %d, want 2", circle.GetDegree())
	}
	if len(circle.GetControlPoints()) != 9 {
		t.Errorf("control points = %d, want 9", len(circle.GetControlPoints()))
	}
	const tolerance = 1e-12
	for i := 0; i <= 200; i++ {
		p := circle.Evaluate(float64(i) / 200)
		if d := p.Sub(center).Length() - radius; d > tolerance {
			t.Fatalf("t=%v: |P-center| = %v, want %v", float64(i)/200, d+radius, radius)
		}
	}
}

func TestRationalCircleArcIsExact(t *testing.T) {
	const radius = 2.0
	arc, err := NewRationalCircleArc(Vec3{}, radius, 0, math.Pi/2)
	if err != nil {
		t.Fatalf("NewRationalCircleArc: %v", err)
	}
	start := arc.Evaluate(0)
	if start.Distance(Vec3{radius, 0, 0}) > 1e-12 {
		t.Errorf("start = %v, want (%v,0,0)", start, radius)
	}
	end := arc.Evaluate(1)
	if end.Distance(Vec3{0, radius, 0}) > 1e-12 {
		t.Errorf("end = %v, want (0,%v,0)", end, radius)
	}
	for i := 0; i <= 50; i++ {
		p := arc.Evaluate(float64(i) / 50)
		if d := p.Length() - radius; d > 1e-12 {
			t.Fatalf("t=%v: |P| = %v, want %v", float64(i)/50, d+radius, radius)
		}
	}
	if _, err := NewRationalCircleArc(Vec3{}, radius, 0, math.Pi); err == nil {
		t.Error("expected rejection of sweeps larger than a quarter circle")
	}
}

func TestNURBSInsertKnotPreservesShape(t *testing.T) {
	center := Vec3{0, 0, 0}
	const radius = 1.5
	circle, err := NewRationalCircle(center, radius)
	if err != nil {
		t.Fatalf("NewRationalCircle: %v", err)
	}
	const samples = 101
	before := make([]Vec3, samples)
	for i := range before {
		before[i] = circle.Evaluate(float64(i) / (samples - 1))
	}
	inserted, err := circle.InsertKnot(0.3)
	if err != nil {
		t.Fatalf("NURBS.InsertKnot: %v", err)
	}
	if len(inserted.GetControlPoints()) != len(circle.GetControlPoints())+1 {
		t.Errorf("control points = %d, want %d", len(inserted.GetControlPoints()), len(circle.GetControlPoints())+1)
	}
	after := make([]Vec3, samples)
	for i := range after {
		after[i] = inserted.Evaluate(float64(i) / (samples - 1))
	}
	if d := maxDistance(before, after); d > 1e-12 {
		t.Errorf("NURBS knot insertion changed the curve, max deviation %v", d)
	}
	for _, p := range after {
		if d := p.Sub(center).Length() - radius; d > 1e-12 {
			t.Fatalf("refined circle radius error %v at %v", d, p)
		}
	}
}
