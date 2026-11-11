package math

import (
	"math"
	"testing"
)

func testPortal() BoundingBox {
	return NewBoundingBox(Vec2{0, 0}, Vec2{10, 10})
}

func TestClipSegmentOutsideToInside(t *testing.T) {
	seg, ok := ClipSegmentToPortal(Vec2{-5, 5}, Vec2{5, 5}, testPortal())
	if !ok {
		t.Fatal("expected visible segment")
	}
	if math.Abs(seg.Start.X()) > 1e-9 || math.Abs(seg.Start.Y()-5) > 1e-9 {
		t.Fatalf("start = %v, want (0,5)", seg.Start)
	}
	if math.Abs(seg.End.X()-5) > 1e-9 || math.Abs(seg.End.Y()-5) > 1e-9 {
		t.Fatalf("end = %v, want (5,5)", seg.End)
	}
}

func TestClipSegmentEntirelyOutside(t *testing.T) {
	if _, ok := ClipSegmentToPortal(Vec2{20, 20}, Vec2{30, 30}, testPortal()); ok {
		t.Fatal("expected segment entirely outside to be removed")
	}
	if _, ok := ClipSegmentToPortal(Vec2{-5, 5}, Vec2{-1, 5}, testPortal()); ok {
		t.Fatal("expected segment left of portal to be removed")
	}
}

func TestClipSegmentInsideUnchanged(t *testing.T) {
	seg, ok := ClipSegmentToPortal(Vec2{2, 2}, Vec2{8, 8}, testPortal())
	if !ok {
		t.Fatal("expected visible segment")
	}
	if seg.Start != (Vec2{2, 2}) || seg.End != (Vec2{8, 8}) {
		t.Fatalf("unexpected clipped segment: %+v", seg)
	}
}

func TestClipSegmentCrossing(t *testing.T) {
	seg, ok := ClipSegmentToPortal(Vec2{-5, 5}, Vec2{15, 5}, testPortal())
	if !ok {
		t.Fatal("expected visible segment")
	}
	if math.Abs(seg.Start.X()) > 1e-9 || math.Abs(seg.End.X()-10) > 1e-9 {
		t.Fatalf("unexpected clipped segment: %+v", seg)
	}
}

func TestClipSegmentsToPortalFilters(t *testing.T) {
	segs := []Segment{
		{Start: Vec2{-1, 5}, End: Vec2{5, 5}},
		{Start: Vec2{20, 20}, End: Vec2{30, 30}},
		{Start: Vec2{2, 2}, End: Vec2{8, 8}},
	}
	out := ClipSegmentsToPortal(segs, testPortal())
	if len(out) != 2 {
		t.Fatalf("expected 2 visible segments, got %d", len(out))
	}
}

func TestClipQuadraticToPortal(t *testing.T) {
	// Curve dips from outside-left, through the portal, to outside-right.
	q := QuadraticBezier{P0: Vec2{-5, 5}, C: Vec2{5, -5}, P1: Vec2{15, 5}}
	parts := ClipQuadraticToPortal(q, testPortal())
	if len(parts) == 0 {
		t.Fatal("expected visible curve portions")
	}
	for _, part := range parts {
		for i := 0; i <= 10; i++ {
			p := part.Point(float64(i) / 10)
			if p.X() < -DefaultTolerance-1e-9 || p.X() > 10+DefaultTolerance+1e-9 ||
				p.Y() < -DefaultTolerance-1e-9 || p.Y() > 10+DefaultTolerance+1e-9 {
				t.Fatalf("clipped curve point outside portal: %v", p)
			}
		}
	}
}

func TestClipCubicToPortal(t *testing.T) {
	c := CubicBezier{P0: Vec2{-5, 5}, C1: Vec2{0, 8}, C2: Vec2{10, 8}, P1: Vec2{15, 5}}
	parts := ClipCubicToPortal(c, testPortal())
	if len(parts) == 0 {
		t.Fatal("expected visible cubic portions")
	}
}

func TestClipCubicEntirelyOutside(t *testing.T) {
	c := CubicBezier{P0: Vec2{20, 20}, C1: Vec2{21, 21}, C2: Vec2{22, 22}, P1: Vec2{23, 23}}
	if parts := ClipCubicToPortal(c, testPortal()); len(parts) != 0 {
		t.Fatalf("expected no visible parts, got %d", len(parts))
	}
}

func TestClipSegmentEmptyPortal(t *testing.T) {
	if _, ok := ClipSegmentToPortal(Vec2{0, 0}, Vec2{1, 1}, BoundingBox{}); ok {
		t.Fatal("expected empty portal to produce no segment")
	}
}
