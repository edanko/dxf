package drawing

import (
	"testing"

	"github.com/edanko/dxf/entity"
)

func newTestLineEnt(t *testing.T, x1, y1, x2, y2 float64) entity.Entity {
	t.Helper()
	line := entity.NewLine()
	line.Start = []float64{x1, y1, 0}
	line.End = []float64{x2, y2, 0}
	return line
}

func TestGetEntitiesInPolygonUsesRealGeometry(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatal(err)
	}

	inside := newTestLineEnt(t, 1, 1, 3, 3)
	crossing := newTestLineEnt(t, 9, 9, 12, 12)
	outside := newTestLineEnt(t, 20, 20, 25, 25)
	for _, ent := range []entity.Entity{inside, crossing, outside} {
		if err := d.AddEntity(ent); err != nil {
			t.Fatal(err)
		}
	}

	polygon := []struct{ X, Y float64 }{
		{0, 0}, {10, 0}, {10, 10}, {0, 10},
	}

	db := d.GetEntityDatabase()
	got := db.GetEntitiesInPolygon(polygon)

	handles := make(map[string]bool)
	for _, ent := range got {
		handles[ent.Handle()] = true
	}
	if !handles[inside.Handle()] {
		t.Errorf("entity fully inside the polygon missing from result: %v", handles)
	}
	if !handles[crossing.Handle()] {
		t.Errorf("entity crossing the window boundary should be included: %v", handles)
	}
	if handles[outside.Handle()] {
		t.Error("entity outside the polygon should be excluded")
	}
	if len(got) != 2 {
		t.Errorf("got %d entities, want 2: %v", len(got), handles)
	}
}
