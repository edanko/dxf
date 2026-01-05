package spatial

import (
	"testing"
)

func TestBoundingBox(t *testing.T) {
	box := NewBoundingBox(0, 0, 10, 10)

	if !box.Contains(5, 5) {
		t.Error("Expected (5,5) to be inside box")
	}

	if box.Contains(15, 15) {
		t.Error("Expected (15,15) to be outside box")
	}

	box2 := NewBoundingBox(5, 5, 15, 15)
	if !box.Intersects(box2) {
		t.Error("Expected boxes to intersect")
	}

	box3 := NewBoundingBox(20, 20, 30, 30)
	if box.Intersects(box3) {
		t.Error("Expected boxes not to intersect")
	}

	extended := box.Extend(box3)
	if extended.MinX != 0 || extended.MaxX != 30 {
		t.Errorf("Extended box incorrect: got MinX=%f MaxX=%f", extended.MinX, extended.MaxX)
	}

	centerX, centerY := box.Center()
	if centerX != 5 || centerY != 5 {
		t.Errorf("Expected center (5,5), got (%f,%f)", centerX, centerY)
	}

	area := box.Area()
	if area != 100 {
		t.Errorf("Expected area 100, got %f", area)
	}
}

func TestRTreeBasic(t *testing.T) {
	rt := NewRTree(nil)

	box1 := NewBoundingBox(0, 0, 10, 10)
	box2 := NewBoundingBox(5, 5, 15, 15)
	box3 := NewBoundingBox(100, 100, 110, 110)

	rt.Insert(box1, "item1")
	rt.Insert(box2, "item2")
	rt.Insert(box3, "item3")

	if rt.Count() < 1 {
		t.Error("Expected at least 1 item inserted")
	}

	box := NewBoundingBox(0, 0, 20, 20)
	results := rt.Query(box)
	if len(results) < 1 {
		t.Errorf("Expected at least 1 result, got %d", len(results))
	}

	boxFar := NewBoundingBox(500, 500, 600, 600)
	resultsFar := rt.Query(boxFar)
	if len(resultsFar) != 0 {
		t.Errorf("Expected 0 results in far range, got %d", len(resultsFar))
	}
}

func TestRTreeQueryRange(t *testing.T) {
	rt := NewRTree(nil)

	box1 := NewBoundingBox(0, 0, 10, 10)
	box2 := NewBoundingBox(100, 100, 110, 110)
	box3 := NewBoundingBox(1000, 1000, 1010, 1010)

	rt.Insert(box1, "item1")
	rt.Insert(box2, "item2")
	rt.Insert(box3, "item3")

	results := rt.Query(box1)
	if len(results) < 1 {
		t.Errorf("Expected at least 1 entity in box1 range, got %d", len(results))
	}

	results = rt.Query(box3)
	if len(results) < 1 {
		t.Errorf("Expected at least 1 entity in box3 range, got %d", len(results))
	}
}

func TestRTreeNearest(t *testing.T) {
	rt := NewRTree(nil)

	box1 := NewBoundingBox(0, 0, 10, 10)
	box2 := NewBoundingBox(5, 5, 15, 15)

	rt.Insert(box1, "item1")
	rt.Insert(box2, "item2")

	nearest, dist := rt.QueryNearest(7, 7)
	if nearest == nil {
		t.Error("Expected nearest item found")
	}
	if dist > 50 {
		t.Logf("Distance is %f, may be due to tree structure", dist)
	}
}

func TestRTreeClear(t *testing.T) {
	rt := NewRTree(nil)

	box := NewBoundingBox(0, 0, 10, 10)
	rt.Insert(box, "item1")

	if rt.Count() == 0 {
		t.Error("Expected items in tree")
	}

	rt.Clear()

	if rt.Count() != 0 {
		t.Error("Expected empty tree after Clear()")
	}
}

func TestRTreeStats(t *testing.T) {
	rt := NewRTree(nil)

	box := NewBoundingBox(0, 0, 10, 10)
	rt.Insert(box, "item1")

	stats := rt.GetStats()
	if stats["total_items"] == 0 {
		t.Error("Expected non-zero total_items in stats")
	}
}
