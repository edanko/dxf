package math

import (
	"math"
	"testing"
)

var unitSquareCCW = []Vec2{{0, 0}, {1, 0}, {1, 1}, {0, 1}}

func TestPolygonAreaAndCentroid(t *testing.T) {
	if got := PolygonSignedArea(unitSquareCCW); math.Abs(got-1) > 1e-12 {
		t.Errorf("signed area (CCW) = %v, want 1", got)
	}
	reversed := []Vec2{{0, 1}, {1, 1}, {1, 0}, {0, 0}}
	if got := PolygonSignedArea(reversed); math.Abs(got+1) > 1e-12 {
		t.Errorf("signed area (CW) = %v, want -1", got)
	}
	if got := PolygonArea(reversed); math.Abs(got-1) > 1e-12 {
		t.Errorf("absolute area = %v, want 1", got)
	}
	centroid := PolygonCentroid(unitSquareCCW)
	if math.Abs(centroid.X()-0.5) > 1e-12 || math.Abs(centroid.Y()-0.5) > 1e-12 {
		t.Errorf("centroid = %v, want (0.5, 0.5)", centroid)
	}
	// A 2x3 rectangle: area 6, centroid (1, 1.5).
	rect := []Vec2{{0, 0}, {2, 0}, {2, 3}, {0, 3}}
	if got := PolygonArea(rect); math.Abs(got-6) > 1e-12 {
		t.Errorf("rectangle area = %v, want 6", got)
	}
	rc := PolygonCentroid(rect)
	if math.Abs(rc.X()-1) > 1e-12 || math.Abs(rc.Y()-1.5) > 1e-12 {
		t.Errorf("rectangle centroid = %v, want (1, 1.5)", rc)
	}
	// A triangle of known area.
	tri := []Vec2{{0, 0}, {4, 0}, {0, 3}}
	if got := PolygonArea(tri); math.Abs(got-6) > 1e-12 {
		t.Errorf("triangle area = %v, want 6", got)
	}
	if got := PolygonCentroid(tri); math.Abs(got.X()-4.0/3) > 1e-12 || math.Abs(got.Y()-1) > 1e-12 {
		t.Errorf("triangle centroid = %v, want (4/3, 1)", got)
	}
}

func TestPolygonWinding(t *testing.T) {
	if got := PolygonWinding(unitSquareCCW); got != PolygonCounterClockwise {
		t.Errorf("CCW winding = %v, want %v", got, PolygonCounterClockwise)
	}
	cw := []Vec2{{0, 0}, {0, 1}, {1, 1}, {1, 0}}
	if got := PolygonWinding(cw); got != PolygonClockwise {
		t.Errorf("CW winding = %v, want %v", got, PolygonClockwise)
	}
	collinear := []Vec2{{0, 0}, {1, 0}, {2, 0}}
	if got := PolygonWinding(collinear); got != PolygonDegenerate {
		t.Errorf("collinear winding = %v, want %v", got, PolygonDegenerate)
	}
}

func TestPointInPolygon(t *testing.T) {
	// Concave arrow pointing right: notch on the left edge.
	arrow := []Vec2{{0, 0}, {4, 0}, {4, 2}, {0, 2}, {1, 1}}
	inside := []Vec2{{3, 1}, {2, 0.5}, {2, 1.5}}
	for _, point := range inside {
		if !PointInPolygon(point, arrow) {
			t.Errorf("point %v expected inside the arrow", point)
		}
	}
	outside := []Vec2{{5, 1}, {-1, 1}, {2, 3}, {0.9, 1}}
	for _, point := range outside {
		if PointInPolygon(point, arrow) {
			t.Errorf("point %v expected outside the arrow", point)
		}
	}
	// Points on an edge count as inside.
	for _, point := range []Vec2{{0, 0}, {1, 0}, {1, 0.5}, {0.5, 0}} {
		if !PointInPolygon(point, unitSquareCCW) {
			t.Errorf("edge point %v expected inside the unit square", point)
		}
	}
	if !PointInPolygon(Vec2{4, 1}, arrow) {
		t.Error("edge point (4, 1) expected inside the arrow")
	}
	if PointInPolygon(Vec2{0.5, 0.5}, []Vec2{{0, 0}, {1, 0}}) {
		t.Error("degenerate polygon must not contain points")
	}
}

func TestConvexHull2D(t *testing.T) {
	points := []Vec2{{0, 0}, {2, 0}, {2, 2}, {0, 2}, {1, 1}, {1.5, 0.5}, {0.2, 0.3}}
	hull := ConvexHull2D(points)
	if len(hull) != 4 {
		t.Fatalf("hull vertex count = %d, want 4: %v", len(hull), hull)
	}
	if got := PolygonArea(hull); math.Abs(got-4) > 1e-12 {
		t.Errorf("hull area = %v, want 4", got)
	}
	if got := PolygonWinding(hull); got != PolygonCounterClockwise {
		t.Errorf("hull winding = %v, want %v", got, PolygonCounterClockwise)
	}
	for _, interior := range []Vec2{{1, 1}, {1.5, 0.5}, {0.2, 0.3}} {
		for _, vertex := range hull {
			if vertex.Distance(interior) < 1e-9 {
				t.Errorf("interior point %v present in hull", interior)
			}
		}
	}
	if got := ConvexHull2D([]Vec2{{0, 0}, {1, 1}}); len(got) != 2 {
		t.Errorf("hull of two points = %v, want both", got)
	}
}

func TestClipPolygonToRect(t *testing.T) {
	rect := NewBoundingBoxFromPoints([]Vec2{{0, 0}, {0.5, 1}})
	clipped := ClipPolygonToRect(unitSquareCCW, rect)
	if len(clipped) < 3 {
		t.Fatalf("clipped polygon has %d vertices, want at least 3", len(clipped))
	}
	if got := PolygonArea(clipped); math.Abs(got-0.5) > 1e-9 {
		t.Errorf("clipped area = %v, want 0.5", got)
	}

	// Fully inside keeps the polygon.
	inside := NewBoundingBoxFromPoints([]Vec2{{-1, -1}, {2, 2}})
	if got := PolygonArea(ClipPolygonToRect(unitSquareCCW, inside)); math.Abs(got-1) > 1e-9 {
		t.Errorf("inside clip area = %v, want 1", got)
	}

	// Fully outside yields an empty result.
	outside := NewBoundingBoxFromPoints([]Vec2{{5, 5}, {6, 6}})
	if got := ClipPolygonToRect(unitSquareCCW, outside); len(got) != 0 {
		t.Errorf("outside clip = %v, want empty", got)
	}

	// A clockwise subject clips the same way as a counter-clockwise one.
	flipped := []Vec2{{0, 1}, {1, 1}, {1, 0}, {0, 0}}
	if got := PolygonArea(ClipPolygonToRect(flipped, rect)); math.Abs(got-0.5) > 1e-9 {
		t.Errorf("clockwise subject clip area = %v, want 0.5", got)
	}
}

func TestClipPolygonToConvex(t *testing.T) {
	clip := []Vec2{{0, 0}, {2, 0}, {0, 2}}
	subject := []Vec2{{-1, -1}, {3, -1}, {3, 3}, {-1, 3}}
	clipped := ClipPolygonToConvex(subject, clip)
	if got := PolygonArea(clipped); math.Abs(got-2) > 1e-9 {
		t.Errorf("clipped area = %v, want 2 (the clip triangle)", got)
	}
	// Clockwise clip polygon is normalized automatically.
	clockwiseClip := []Vec2{{0, 0}, {0, 2}, {2, 0}}
	if got := PolygonArea(ClipPolygonToConvex(subject, clockwiseClip)); math.Abs(got-2) > 1e-9 {
		t.Errorf("clockwise clip area = %v, want 2", got)
	}
	if got := ClipPolygonToConvex([]Vec2{{0, 0}, {1, 1}}, clip); got != nil {
		t.Errorf("degenerate subject = %v, want nil", got)
	}
}
