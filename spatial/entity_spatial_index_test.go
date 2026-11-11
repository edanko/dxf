// Tests for the entity-backed spatial index.
package spatial

import (
	"math"
	"testing"

	"github.com/edanko/dxf/entity"
)

func newTestLine(x1, y1, x2, y2 float64) *entity.Line {
	l := entity.NewLine()
	l.Start = []float64{x1, y1, 0}
	l.End = []float64{x2, y2, 0}
	return l
}

func TestEntitySpatialIndexRangeExcludesDistant(t *testing.T) {
	idx := NewEntitySpatialIndex()
	near := newTestLine(0, 0, 10, 10)
	far := newTestLine(1000, 1000, 1100, 1100)
	crossing := newTestLine(5, 5, 15, 15)
	idx.Insert(near)
	idx.Insert(far)
	idx.Insert(crossing)

	if idx.Count() != 3 {
		t.Fatalf("expected 3 indexed entities, got %d", idx.Count())
	}

	results := idx.QueryRange(0, 0, 12, 12)
	if len(results) != 2 {
		t.Fatalf("expected 2 entities in window, got %d", len(results))
	}
	for _, e := range results {
		if e == entity.Entity(far) {
			t.Error("distant entity returned by window query")
		}
	}

	// A window touching only the far entity returns only that one.
	results = idx.QueryRange(1005, 1005, 1015, 1015)
	if len(results) != 1 || results[0] != entity.Entity(far) {
		t.Errorf("expected only the far entity, got %d results", len(results))
	}
}

func TestEntitySpatialIndexNearestUsesBoxDistance(t *testing.T) {
	idx := NewEntitySpatialIndex()
	e := newTestLine(0, 0, 10, 10)
	idx.Insert(e)

	// Distance to the box (0,0)-(10,10) from (12,5) is 2, while the
	// distance to the box center (5,5) would be ~7.
	nearest, dist := idx.QueryNearest(12, 5)
	if nearest != entity.Entity(e) {
		t.Fatal("nearest did not find the only entity")
	}
	if math.Abs(dist-2) > 1e-9 {
		t.Errorf("nearest distance = %v, want 2 (box distance)", dist)
	}
}

func TestEntitySpatialIndexRemoveAndReinsert(t *testing.T) {
	idx := NewEntitySpatialIndex()
	e := newTestLine(0, 0, 10, 10)
	idx.Insert(e)

	idx.Remove(e)
	if idx.Count() != 0 {
		t.Errorf("expected empty index after Remove, got %d", idx.Count())
	}
	if results := idx.QueryRange(-1, -1, 11, 11); len(results) != 0 {
		t.Errorf("removed entity still found: %d results", len(results))
	}

	// Reinserting works and the entity is queryable again.
	idx.Insert(e)
	if idx.Count() != 1 {
		t.Fatalf("expected 1 entity after reinsert, got %d", idx.Count())
	}
	results := idx.QueryRange(-1, -1, 11, 11)
	if len(results) != 1 || results[0] != entity.Entity(e) {
		t.Error("reinserted entity not found by query")
	}

	// Reinserting the same entity replaces its entry instead of
	// duplicating it.
	idx.Insert(e)
	if idx.Count() != 1 {
		t.Fatalf("reinsert duplicated the entity: %d entries", idx.Count())
	}
	if results := idx.QueryRange(-1, -1, 11, 11); len(results) != 1 {
		t.Errorf("expected 1 result after duplicate reinsert, got %d", len(results))
	}

	idx.Clear()
	if idx.Count() != 0 || idx.GetBoundingBox() != nil {
		t.Error("Clear did not reset the index")
	}
}

func TestEntitySpatialIndexGetBoundingBox(t *testing.T) {
	idx := NewEntitySpatialIndex()
	idx.Insert(newTestLine(0, 0, 10, 10))
	idx.Insert(newTestLine(20, 30, 40, 50))

	bbox := idx.GetBoundingBox()
	if bbox == nil {
		t.Fatal("expected non-nil bounding box")
	}
	if bbox.MinX != 0 || bbox.MinY != 0 || bbox.MaxX != 40 || bbox.MaxY != 50 {
		t.Errorf("bounding box = %v, want (0,0)-(40,50)", bbox)
	}
}
