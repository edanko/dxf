package entity

import (
	"errors"
	"math"
	"testing"

	dxfmath "github.com/edanko/dxf/math"
)

func testEllipse() *Ellipse {
	e := NewEllipse()
	e.Center = []float64{1, 2, 0}
	e.SetRadius(5, 3) // Major=5, Minor=3, Ratio=0.6
	return e
}

// TestEllipseToBezierMatchesEllipse verifies that the arc-to-cubic
// decomposition samples the ellipse within tolerance and closes for a full
// ellipse.
func TestEllipseToBezierMatchesEllipse(t *testing.T) {
	e := testEllipse()
	segments, err := e.ToBezier()
	if err != nil {
		t.Fatalf("ToBezier returned error: %v", err)
	}
	if len(segments) != 4 {
		t.Fatalf("full ellipse should decompose into 4 cubic segments, got %d", len(segments))
	}

	tol := 1e-3 * e.Major
	for i, seg := range segments {
		for k := 0; k <= 10; k++ {
			tk := float64(k) / 10.0
			got := seg.EvaluatePoint(tk)
			// Radial deviation from the ellipse: normalize by the semi-axis
			// lengths and measure how far the point strays from the unit
			// circle. This is independent of the Bézier parameterization.
			dx := (got.X() - e.Center[0]) / e.Major
			dy := (got.Y() - e.Center[1]) / e.Minor
			dev := math.Abs(math.Sqrt(dx*dx+dy*dy)-1.0) * e.Major
			if dev > tol {
				t.Errorf("segment %d t=%.2f: radial deviation from ellipse %g > tol %g", i, tk, dev, tol)
			}
		}
	}

	// Consecutive segments share boundary points and the decomposition closes.
	endOf := func(seg *dxfmath.Bezier) dxfmath.Vec3 { return seg.EvaluatePoint(1) }
	for i := 1; i < len(segments); i++ {
		if d := endOf(segments[i-1]).Distance(segments[i].EvaluatePoint(0)); d > 1e-9 {
			t.Errorf("segments %d and %d do not share boundary point (d=%g)", i-1, i, d)
		}
	}
	if d := endOf(segments[len(segments)-1]).Distance(segments[0].EvaluatePoint(0)); d > 1e-9 {
		t.Errorf("decomposition does not close (d=%g)", d)
	}
}

// TestEllipseToBezierArc verifies an open arc decomposition covers exactly
// the requested parameter range.
func TestEllipseToBezierArc(t *testing.T) {
	e := testEllipse()
	e.SetAngles(30, 150)
	segments, err := e.ToBezier()
	if err != nil {
		t.Fatalf("ToBezier returned error: %v", err)
	}
	if len(segments) != 2 {
		t.Fatalf("120-degree arc should decompose into 2 cubic segments, got %d", len(segments))
	}

	tol := 1e-3 * e.Major
	first := segments[0].EvaluatePoint(0)
	theta0 := 30 * math.Pi / 180.0
	expected0 := dxfmath.NewVec3(e.Center[0]+e.Major*math.Cos(theta0), e.Center[1]+e.Minor*math.Sin(theta0), 0)
	if d := first.Distance(expected0); d > 1e-9 {
		t.Errorf("arc start not at parameter start: %g", d)
	}
	last := segments[len(segments)-1].EvaluatePoint(1)
	theta1 := 150 * math.Pi / 180.0
	expected1 := dxfmath.NewVec3(e.Center[0]+e.Major*math.Cos(theta1), e.Center[1]+e.Minor*math.Sin(theta1), 0)
	if d := last.Distance(expected1); d > 1e-9 {
		t.Errorf("arc end not at parameter end: %g", d)
	}

	// Interior samples stay on the ellipse within tolerance (radial measure).
	for _, seg := range segments {
		for k := 1; k < 10; k++ {
			got := seg.EvaluatePoint(float64(k) / 10.0)
			dx := (got.X() - e.Center[0]) / e.Major
			dy := (got.Y() - e.Center[1]) / e.Minor
			dev := math.Abs(math.Sqrt(dx*dx+dy*dy)-1.0) * e.Major
			if dev > tol {
				t.Errorf("arc sample off ellipse by %g > tol %g", dev, tol)
			}
		}
	}
}

// TestEllipseToBezierDegenerate verifies explicit errors for degenerate
// ellipses.
func TestEllipseToBezierDegenerate(t *testing.T) {
	t.Run("zero major axis", func(t *testing.T) {
		e := testEllipse()
		e.Major = 0
		_, err := e.ToBezier()
		if err == nil {
			t.Fatal("expected error for zero major axis")
		}
		if !errors.Is(err, ErrDegenerateEllipse) {
			t.Fatalf("error %v is not ErrDegenerateEllipse", err)
		}
		if msg := err.Error(); !contains(msg, "major") {
			t.Errorf("error should name offending field 'major': %v", msg)
		}
	})

	t.Run("zero minor axis", func(t *testing.T) {
		e := testEllipse()
		e.Minor = 0
		e.Ratio = 0
		_, err := e.ToBezier()
		if err == nil {
			t.Fatal("expected error for zero minor axis")
		}
		if !errors.Is(err, ErrDegenerateEllipse) {
			t.Fatalf("error %v is not ErrDegenerateEllipse", err)
		}
		if msg := err.Error(); !contains(msg, "minor") {
			t.Errorf("error should name offending field 'minor': %v", msg)
		}
	})

	t.Run("empty parameters", func(t *testing.T) {
		e := testEllipse()
		e.SetAngles(45, 45)
		_, err := e.ToBezier()
		if err == nil {
			t.Fatal("expected error for zero-span arc")
		}
		if !errors.Is(err, ErrDegenerateEllipse) {
			t.Fatalf("error %v is not ErrDegenerateEllipse", err)
		}
	})
}

// TestEllipseToBezierEmbeddedSpline verifies the error naming the embedded
// representation when the record retains raw data.
func TestEllipseToBezierEmbeddedSpline(t *testing.T) {
	e := testEllipse()
	e.SetRawTags([]TagPair{{Code: 330, Value: "5"}})
	_, err := e.ToBezier()
	if err == nil {
		t.Fatal("expected error for embedded spline record")
	}
	if msg := err.Error(); !contains(msg, "embedded spline") {
		t.Errorf("error should name embedded spline representation: %v", msg)
	}
}

// TestEllipseFlattenUsesDecomposition verifies Flatten covers the whole
// ellipse and stays within tolerance.
func TestEllipseFlattenUsesDecomposition(t *testing.T) {
	e := testEllipse()
	pts := e.Flatten(0.01)
	if len(pts) < 8 {
		t.Fatalf("flattened ellipse too sparse: %d points", len(pts))
	}
	tol := 1e-3 * e.Major
	for _, p := range pts {
		angle := math.Atan2((p.Y()-e.Center[1])/e.Minor, (p.X()-e.Center[0])/e.Major)
		expected := dxfmath.NewVec3(
			e.Center[0]+e.Major*math.Cos(angle),
			e.Center[1]+e.Minor*math.Sin(angle),
			0,
		)
		if d := p.Distance(expected); d > tol {
			t.Fatalf("flattened point off ellipse: %g", d)
		}
	}
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
