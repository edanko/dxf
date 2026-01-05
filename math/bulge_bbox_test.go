package math

import (
	"math"
	"testing"
)

func TestBulge(t *testing.T) {
	// Test straight line (bulge = 0)
	bulge := NewBulgeFromAngle(0)
	if bulge.IsArc() {
		t.Error("Zero bulge should not be an arc")
	}

	if bulge.Angle() != 0 {
		t.Errorf("Expected angle 0, got %v", bulge.Angle())
	}

	// Test 90-degree arc
	bulge = NewBulgeFromAngle(math.Pi / 2) // 90 degrees
	if !bulge.IsArc() {
		t.Error("Non-zero bulge should be an arc")
	}

	// For a 90-degree arc, the bulge is tan(π/8) ≈ 0.4142
	expected := math.Tan(math.Pi / 8)
	if math.Abs(float64(bulge)-expected) > 1e-10 {
		t.Errorf("Expected bulge %v, got %v", expected, bulge)
	}

	// Test clockwise direction
	clockwiseBulge := NewBulgeFromAngle(-math.Pi / 2)
	if !clockwiseBulge.IsClockwise() {
		t.Error("Negative bulge should be clockwise")
	}
}

func TestBulgeCenter(t *testing.T) {
	start := Vec2{0, 0}
	end := Vec2{2, 0}

	// Test straight line (should return midpoint)
	bulge := Bulge(0)
	center := bulge.Center(start, end)
	expected := Vec2{1, 0} // Midpoint

	if !center.IsEqual(expected, 1e-10) {
		t.Errorf("Expected center %v, got %v", expected, center)
	}

	// Test 90-degree arc (should be offset perpendicular)
	bulge = NewBulgeFromAngle(math.Pi / 2) // 90 degrees
	center = bulge.Center(start, end)

	// For a 90-degree arc from (0,0) to (2,0), center should be at (1,-1)
	arcExpected := Vec2{1, -1}

	if !center.IsEqual(arcExpected, 1e-9) {
		t.Errorf("Expected center %v, got %v", arcExpected, center)
	}
}

func TestBoundingBox(t *testing.T) {
	// Test empty bounding box
	bb := BoundingBox{}
	if !bb.IsEmpty() {
		t.Error("Default bounding box should be empty")
	}

	// Test from points
	points := []Vec2{
		{1, 2},
		{4, 6},
		{-1, 3},
	}

	bb = NewBoundingBoxFromPoints(points)

	// Check min and max
	expectedMin := Vec2{-1, 2}
	expectedMax := Vec2{4, 6}

	if !bb.Min.IsEqual(expectedMin, 1e-10) || !bb.Max.IsEqual(expectedMax, 1e-10) {
		t.Errorf("Expected bounds %v-%v, got %v-%v", expectedMin, expectedMax, bb.Min, bb.Max)
	}

	// Test dimensions
	if bb.Width() != 5 {
		t.Errorf("Expected width 5, got %v", bb.Width())
	}
	if bb.Height() != 4 {
		t.Errorf("Expected height 4, got %v", bb.Height())
	}

	// Test center
	expectedCenter := Vec2{1.5, 4}
	if !bb.Center().IsEqual(expectedCenter, 1e-10) {
		t.Errorf("Expected center %v, got %v", expectedCenter, bb.Center())
	}
}

func TestBoundingBoxOperations(t *testing.T) {
	bb1 := NewBoundingBox(Vec2{0, 0}, Vec2{2, 2})
	bb2 := NewBoundingBox(Vec2{1, 1}, Vec2{3, 3})

	// Test intersection
	if !bb1.IntersectsBoundingBox(bb2) {
		t.Error("Bounding boxes should intersect")
	}

	bb3 := NewBoundingBox(Vec2{3, 3}, Vec2{4, 4})
	if bb1.IntersectsBoundingBox(bb3) {
		t.Error("Bounding boxes should not intersect")
	}

	// Test point containment
	if !bb1.ContainsPoint(Vec2{1, 1}) {
		t.Error("Should contain point (1,1)")
	}

	if bb1.ContainsPoint(Vec2{3, 3}) {
		t.Error("Should not contain point (3,3)")
	}

	// Test union
	union := bb1.Union(bb2)
	expectedUnion := NewBoundingBox(Vec2{0, 0}, Vec2{3, 3})
	if !union.Min.IsEqual(expectedUnion.Min, 1e-10) || !union.Max.IsEqual(expectedUnion.Max, 1e-10) {
		t.Errorf("Expected union %v, got %v", expectedUnion, union)
	}
}

func TestBulgeSegment(t *testing.T) {
	start := Vec2{0, 0}
	end := Vec2{2, 0}
	bulge := NewBulgeFromAngle(math.Pi / 2) // 90-degree arc

	segment := BulgeSegment{
		Start: start,
		End:   end,
		Bulge: bulge,
	}

	// Test midpoint
	midpoint := segment.Midpoint()
	expectedMid := Vec2{1, 0} // Straight line midpoint for bulge calculation

	if !midpoint.IsEqual(expectedMid, 1e-10) {
		t.Errorf("Expected midpoint %v, got %v", expectedMid, midpoint)
	}

	// Test approximation
	points := segment.Approximate(4)
	if len(points) != 5 {
		t.Errorf("Expected 5 points (including endpoints), got %d", len(points))
	}

	// First point should be start
	if !points[0].IsEqual(start, 1e-10) {
		t.Errorf("First point should be start, got %v", points[0])
	}

	// Last point should be end
	if !points[4].IsEqual(end, 1e-10) {
		t.Errorf("Last point should be end, got %v", points[4])
	}
}
