package math

import (
	"math"
	"testing"
)

func TestIntersectLineLine2D(t *testing.T) {
	cross := IntersectLineLine2D(Vec2{0, 0}, Vec2{1, 0}, Vec2{0.5, -1}, Vec2{0.5, 1}, DefaultTolerance)
	if cross.Kind != IntersectionPoint {
		t.Fatalf("crossing lines kind = %v, want %v", cross.Kind, IntersectionPoint)
	}
	if len(cross.Points) != 1 {
		t.Fatalf("crossing lines point count = %d, want 1", len(cross.Points))
	}
	if got := cross.Points[0]; math.Abs(got.X()-0.5) > 1e-12 || math.Abs(got.Y()) > 1e-12 {
		t.Errorf("crossing point = %v, want (0.5, 0)", got)
	}

	parallel := IntersectLineLine2D(Vec2{0, 0}, Vec2{1, 0}, Vec2{0, 1}, Vec2{1, 1}, DefaultTolerance)
	if parallel.Kind != IntersectionNone {
		t.Errorf("parallel lines kind = %v, want %v", parallel.Kind, IntersectionNone)
	}

	coincident := IntersectLineLine2D(Vec2{0, 0}, Vec2{1, 0}, Vec2{2, 0}, Vec2{3, 0}, DefaultTolerance)
	if coincident.Kind != IntersectionCoincident {
		t.Errorf("coincident lines kind = %v, want %v", coincident.Kind, IntersectionCoincident)
	}

	// Lines that only meet outside both segments still intersect as lines.
	far := IntersectLineLine2D(Vec2{0, 0}, Vec2{1, 0}, Vec2{5, 1}, Vec2{5, -1}, DefaultTolerance)
	if far.Kind != IntersectionPoint || math.Abs(far.Points[0].X()-5) > 1e-12 {
		t.Errorf("far lines = %+v, want point at x = 5", far)
	}
}

func TestIntersectSegmentSegment2D(t *testing.T) {
	cross := IntersectSegmentSegment2D(Vec2{0, 0}, Vec2{2, 2}, Vec2{0, 2}, Vec2{2, 0}, DefaultTolerance)
	if cross.Kind != IntersectionPoint || math.Abs(cross.Points[0].X()-1) > 1e-12 {
		t.Errorf("crossing segments = %+v, want point at (1, 1)", cross)
	}

	// Touching at endpoints is the tangent case.
	touch := IntersectSegmentSegment2D(Vec2{0, 0}, Vec2{1, 0}, Vec2{1, 0}, Vec2{1, 1}, DefaultTolerance)
	if touch.Kind != IntersectionPoint || math.Abs(touch.Points[0].X()-1) > 1e-12 {
		t.Errorf("touching segments = %+v, want point at (1, 0)", touch)
	}

	// Parallel segments that do not reach each other miss.
	miss := IntersectSegmentSegment2D(Vec2{0, 0}, Vec2{1, 0}, Vec2{2, 0}, Vec2{3, 0}, DefaultTolerance)
	if miss.Kind != IntersectionNone {
		t.Errorf("collinear disjoint segments kind = %v, want %v", miss.Kind, IntersectionNone)
	}

	overlap := IntersectSegmentSegment2D(Vec2{0, 0}, Vec2{4, 0}, Vec2{2, 0}, Vec2{6, 0}, DefaultTolerance)
	if overlap.Kind != IntersectionOverlap {
		t.Fatalf("collinear overlapping segments kind = %v, want %v", overlap.Kind, IntersectionOverlap)
	}
	if len(overlap.Points) != 2 {
		t.Fatalf("overlap endpoint count = %d, want 2", len(overlap.Points))
	}
	if math.Abs(overlap.Points[0].X()-2) > 1e-12 || math.Abs(overlap.Points[1].X()-4) > 1e-12 {
		t.Errorf("overlap span = %v..%v, want 2..4", overlap.Points[0], overlap.Points[1])
	}

	// Segments whose supporting lines cross outside one segment do not meet.
	behind := IntersectSegmentSegment2D(Vec2{0, 0}, Vec2{1, 0}, Vec2{5, 1}, Vec2{5, -1}, DefaultTolerance)
	if behind.Kind != IntersectionNone {
		t.Errorf("non-overlapping segments kind = %v, want %v", behind.Kind, IntersectionNone)
	}
}

func TestIntersectLineLine3D(t *testing.T) {
	meet := IntersectLineLine3D(Vec3{0, 0, 0}, Vec3{1, 0, 0}, Vec3{0.5, -1, 0}, Vec3{0.5, 1, 0}, DefaultTolerance)
	if meet.Kind != IntersectionPoint || math.Abs(meet.Points[0].Y()) > 1e-9 {
		t.Errorf("crossing 3D lines = %+v, want point on y = 0", meet)
	}

	skew := IntersectLineLine3D(Vec3{0, 0, 0}, Vec3{1, 0, 0}, Vec3{0.5, -1, 1}, Vec3{0.5, 1, 1}, DefaultTolerance)
	if skew.Kind != IntersectionNone {
		t.Errorf("skew lines kind = %v, want %v", skew.Kind, IntersectionNone)
	}

	parallel := IntersectLineLine3D(Vec3{0, 0, 0}, Vec3{1, 0, 0}, Vec3{0, 1, 0}, Vec3{1, 1, 0}, DefaultTolerance)
	if parallel.Kind != IntersectionNone {
		t.Errorf("parallel 3D lines kind = %v, want %v", parallel.Kind, IntersectionNone)
	}

	same := IntersectLineLine3D(Vec3{0, 0, 0}, Vec3{1, 0, 0}, Vec3{2, 0, 0}, Vec3{3, 0, 0}, DefaultTolerance)
	if same.Kind != IntersectionCoincident {
		t.Errorf("collinear 3D lines kind = %v, want %v", same.Kind, IntersectionCoincident)
	}
}

func TestIntersectRayPlane(t *testing.T) {
	plane := NewPlane(Vec3{0, 0, 0}, Vec3{0, 0, 1})

	hit := IntersectRayPlane(Vec3{1, 1, -5}, Vec3{0, 0, 1}, plane, DefaultTolerance)
	if hit.Kind != IntersectionPoint {
		t.Fatalf("ray-plane kind = %v, want %v", hit.Kind, IntersectionPoint)
	}
	if math.Abs(hit.Points[0].Z()) > 1e-12 || math.Abs(hit.Points[0].X()-1) > 1e-12 {
		t.Errorf("ray-plane point = %v, want (1, 1, 0)", hit.Points[0])
	}

	behind := IntersectRayPlane(Vec3{1, 1, 5}, Vec3{0, 0, 1}, plane, DefaultTolerance)
	if behind.Kind != IntersectionNone {
		t.Errorf("ray pointing away kind = %v, want %v", behind.Kind, IntersectionNone)
	}

	parallel := IntersectRayPlane(Vec3{0, 0, 5}, Vec3{1, 0, 0}, plane, DefaultTolerance)
	if parallel.Kind != IntersectionNone {
		t.Errorf("parallel ray kind = %v, want %v", parallel.Kind, IntersectionNone)
	}

	inPlane := IntersectRayPlane(Vec3{0, 0, 0}, Vec3{1, 0, 0}, plane, DefaultTolerance)
	if inPlane.Kind != IntersectionCoincident {
		t.Errorf("ray inside plane kind = %v, want %v", inPlane.Kind, IntersectionCoincident)
	}
}

func TestIntersectLineBezier2D(t *testing.T) {
	curve, err := NewBezier4P(Vec3{0, 0, 0}, Vec3{1, 2, 0}, Vec3{3, 2, 0}, Vec3{4, 0, 0})
	if err != nil {
		t.Fatalf("NewBezier4P: %v", err)
	}
	hit := IntersectLineBezier2D(Vec2{2, -1}, Vec2{2, 1.5}, curve, 1e-6)
	if hit.Empty() {
		t.Fatalf("segment crossing the arc found no intersection: %+v", hit)
	}
	for _, point := range hit.Points {
		if math.Abs(point.X()-2) > 1e-3 {
			t.Errorf("intersection x = %v, want 2", point.X())
		}
	}

	miss := IntersectLineBezier2D(Vec2{2, 3}, Vec2{2, 5}, curve, 1e-6)
	if !miss.Empty() {
		t.Errorf("segment above the arc intersected: %+v", miss)
	}
}

func TestIntersectBezierBezier2D(t *testing.T) {
	a, err := NewBezier4P(Vec3{0, 0, 0}, Vec3{1, 2, 0}, Vec3{3, 2, 0}, Vec3{4, 0, 0})
	if err != nil {
		t.Fatalf("NewBezier4P a: %v", err)
	}
	b, err := NewBezier4P(Vec3{1, -1, 0}, Vec3{1, 3.4, 0}, Vec3{3, 3.4, 0}, Vec3{3, -1, 0})
	if err != nil {
		t.Fatalf("NewBezier4P b: %v", err)
	}
	hit := IntersectBezierBezier2D(a, b, 1e-6)
	if hit.Kind != IntersectionPoints || len(hit.Points) != 2 {
		t.Fatalf("arc-arc intersection = %+v, want two points", hit)
	}
	for _, point := range hit.Points {
		if point.X() < 1 || point.X() > 3 {
			t.Errorf("intersection outside both curves: %v", point)
		}
	}
}
