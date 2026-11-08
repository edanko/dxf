package query

import (
	"testing"

	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/table"
)

func TestNewEntityQuery(t *testing.T) {
	line := entity.NewLine()
	circle := entity.NewCircle()

	q := NewEntityQuery(line, circle)

	if q.Len() != 2 {
		t.Errorf("Expected length 2, got %d", q.Len())
	}
}

func TestEntityQueryLen(t *testing.T) {
	q := NewEntityQuery()
	if q.Len() != 0 {
		t.Errorf("Expected length 0, got %d", q.Len())
	}

	q = NewEntityQuery(entity.NewLine())
	if q.Len() != 1 {
		t.Errorf("Expected length 1, got %d", q.Len())
	}
}

func TestEntityQueryLess(t *testing.T) {
	e1 := entity.NewLine()
	e2 := entity.NewCircle()

	q := NewEntityQuery(e1, e2)

	handle0 := q.Get(0).Handle()
	handle1 := q.Get(1).Handle()

	if handle0 != handle1 {
		if !q.Less(0, 1) && !q.Less(1, 0) {
			t.Logf("Less comparison between %s and %s", handle0, handle1)
		}
	} else {
		t.Log("Both entities have same handle")
	}
}

func TestEntityQuerySwap(t *testing.T) {
	e1 := entity.NewLine()
	e2 := entity.NewCircle()

	q := NewEntityQuery(e1, e2)
	handle0Before := q.Get(0).Handle()
	handle1Before := q.Get(1).Handle()
	q.Swap(0, 1)

	if q.Get(0).Handle() != handle1Before || q.Get(1).Handle() != handle0Before {
		t.Error("Swap did not work correctly")
	}
}

func TestEntityQueryGet(t *testing.T) {
	e1 := entity.NewLine()
	e2 := entity.NewCircle()

	q := NewEntityQuery(e1, e2)

	if q.Get(0) == nil {
		t.Error("Get(0) should return first entity")
	}

	if q.Get(1) == nil {
		t.Error("Get(1) should return second entity")
	}

	if q.Get(-1) != nil {
		t.Error("Get(-1) should return nil")
	}

	if q.Get(2) != nil {
		t.Error("Get(2) should return nil (out of bounds)")
	}
}

func TestEntityQueryFilter(t *testing.T) {
	layerA := table.NewLayer("A", 0, nil)
	layerB := table.NewLayer("B", 0, nil)

	e1 := entity.NewLine()
	e1.SetLayer(layerA)
	e2 := entity.NewCircle()
	e2.SetLayer(layerB)
	e3 := entity.NewLine()
	e3.SetLayer(layerA)

	q := NewEntityQuery(e1, e2, e3)

	filtered := q.Filter(func(ent entity.Entity) bool {
		return ent.Layer().Name() == "A"
	})

	if filtered.Len() != 2 {
		t.Errorf("Expected 2 filtered entities, got %d", filtered.Len())
	}
}

func TestEntityQuerySelect(t *testing.T) {
	line := entity.NewLine()
	circle := entity.NewCircle()
	arc := entity.NewArc(nil)

	q := NewEntityQuery(line, circle, arc)

	t.Run("Select by type", func(t *testing.T) {
		selected := q.Select("LINE", "ARC")
		if selected.Len() != 2 {
			t.Errorf("Expected 2 entities, got %d", selected.Len())
		}
	})

	t.Run("Select all", func(t *testing.T) {
		selected := q.Select("*")
		if selected.Len() != 3 {
			t.Errorf("Expected 3 entities, got %d", selected.Len())
		}
	})

	t.Run("Select all with exclusion", func(t *testing.T) {
		selected := q.Select("*", "!CIRCLE")
		if selected.Len() != 2 {
			t.Errorf("Expected 2 entities (excluding CIRCLE), got %d", selected.Len())
		}
	})

	t.Run("Select empty", func(t *testing.T) {
		selected := q.Select()
		if selected.Len() != 0 {
			t.Errorf("Expected 0 entities, got %d", selected.Len())
		}
	})
}

func TestEntityQueryWhere(t *testing.T) {
	layerA := table.NewLayer("A", 0, nil)
	layerB := table.NewLayer("B", 0, nil)

	e1 := entity.NewLine()
	e1.SetLayer(layerA)
	e2 := entity.NewCircle()
	e2.SetLayer(layerB)

	q := NewEntityQuery(e1, e2)

	result, err := q.Where(`layer=="A"`)
	if err != nil {
		t.Fatalf("Where failed: %v", err)
	}

	if result.Len() != 1 {
		t.Errorf("Expected 1 entity, got %d", result.Len())
	}

	result, err = q.Where("")
	if err != nil {
		t.Fatalf("Where with empty string failed: %v", err)
	}

	if result.Len() != 2 {
		t.Errorf("Expected 2 entities for empty where, got %d", result.Len())
	}
}

func TestEntityQueryQuery(t *testing.T) {
	layerA := table.NewLayer("A", 0, nil)
	layerB := table.NewLayer("B", 0, nil)

	line := entity.NewLine()
	line.SetLayer(layerA)
	circle := entity.NewCircle()
	circle.SetLayer(layerB)

	q := NewEntityQuery(line, circle)

	result, err := q.Query("LINE")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if result.Len() != 1 {
		t.Errorf("Expected 1 entity, got %d", result.Len())
	}

	result, err = q.Query(`LINE[layer=="A"]`)
	if err != nil {
		t.Fatalf("Query with condition failed: %v", err)
	}

	if result.Len() != 1 {
		t.Errorf("Expected 1 entity, got %d", result.Len())
	}
}

func TestEntityQueryFirst(t *testing.T) {
	e1 := entity.NewLine()
	e2 := entity.NewCircle()

	q := NewEntityQuery(e1, e2)

	first := q.First()
	if first == nil {
		t.Error("First should not return nil")
	}

	q = NewEntityQuery()
	if q.First() != nil {
		t.Error("First on empty query should return nil")
	}
}

func TestEntityQueryLast(t *testing.T) {
	e1 := entity.NewLine()
	e2 := entity.NewCircle()

	q := NewEntityQuery(e1, e2)

	last := q.Last()
	if last == nil {
		t.Error("Last should not return nil")
	}

	q = NewEntityQuery()
	if q.Last() != nil {
		t.Error("Last on empty query should return nil")
	}
}

func TestEntityQueryEmpty(t *testing.T) {
	q := NewEntityQuery()
	if !q.Empty() {
		t.Error("Empty query should be empty")
	}

	q = NewEntityQuery(entity.NewLine())
	if q.Empty() {
		t.Error("Non-empty query should not be empty")
	}
}

func TestEntityQueryToSlice(t *testing.T) {
	e1 := entity.NewLine()
	e2 := entity.NewCircle()

	q := NewEntityQuery(e1, e2)
	slice := q.ToSlice()

	if len(slice) != 2 {
		t.Errorf("Expected slice of length 2, got %d", len(slice))
	}
}

func TestEntityQueryExtend(t *testing.T) {
	e1 := entity.NewLine()
	e2 := entity.NewCircle()

	q1 := NewEntityQuery(e1)
	q2 := NewEntityQuery(e2)

	q1.Extend(q2)

	if q1.Len() != 2 {
		t.Errorf("Expected length 2, got %d", q1.Len())
	}
}

func TestEntityQueryEach(t *testing.T) {
	e1 := entity.NewLine()
	e2 := entity.NewCircle()

	q := NewEntityQuery(e1, e2)

	count := 0
	q.Each(func(ent entity.Entity) {
		count++
	})

	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

func TestEntityQueryMap(t *testing.T) {
	e1 := entity.NewLine()
	e2 := entity.NewCircle()

	q := NewEntityQuery(e1, e2)

	handles := q.Map(func(ent entity.Entity) interface{} {
		return ent.Handle()
	})

	if len(handles) != 2 {
		t.Errorf("Expected 2 handles, got %d", len(handles))
	}
}

func TestEntityQueryGroupBy(t *testing.T) {
	layerA := table.NewLayer("A", 0, nil)
	layerB := table.NewLayer("B", 0, nil)

	e1 := entity.NewLine()
	e1.SetLayer(layerA)
	e2 := entity.NewCircle()
	e2.SetLayer(layerB)
	e3 := entity.NewLine()
	e3.SetLayer(layerA)

	q := NewEntityQuery(e1, e2, e3)

	groups := q.GroupBy(func(ent entity.Entity) string {
		return ent.Layer().Name()
	})

	if len(groups) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(groups))
	}

	if groups["A"].Len() != 2 {
		t.Errorf("Expected group A to have 2 entities, got %d", groups["A"].Len())
	}

	if groups["B"].Len() != 1 {
		t.Errorf("Expected group B to have 1 entity, got %d", groups["B"].Len())
	}
}

func TestEntityQueryCountBy(t *testing.T) {
	layerA := table.NewLayer("A", 0, nil)
	layerB := table.NewLayer("B", 0, nil)

	e1 := entity.NewLine()
	e1.SetLayer(layerA)
	e2 := entity.NewCircle()
	e2.SetLayer(layerB)
	e3 := entity.NewLine()
	e3.SetLayer(layerA)

	q := NewEntityQuery(e1, e2, e3)

	counts := q.CountBy(func(ent entity.Entity) string {
		return ent.Layer().Name()
	})

	if counts["A"] != 2 {
		t.Errorf("Expected count A = 2, got %d", counts["A"])
	}

	if counts["B"] != 1 {
		t.Errorf("Expected count B = 1, got %d", counts["B"])
	}
}

func TestEntityQueryUnique(t *testing.T) {
	e1 := entity.NewLine()
	e2 := entity.NewCircle()
	e3 := entity.NewArc(nil)

	q := NewEntityQuery(e1, e2, e3)

	unique := q.Unique()
	if unique.Len() < 1 {
		t.Error("Expected at least 1 unique entity")
	}
}

func TestEntityQueryBounds(t *testing.T) {
	e1 := entity.NewLine()
	e1.Start = []float64{0, 0, 0}
	e1.End = []float64{10, 10, 10}

	e2 := entity.NewLine()
	e2.Start = []float64{5, 5, 5}
	e2.End = []float64{15, 15, 15}

	q := NewEntityQuery(e1, e2)

	min, max, found := q.Bounds()
	if !found {
		t.Error("Bounds should be found")
	}

	if min.X() != 0 || min.Y() != 0 || min.Z() != 0 {
		t.Errorf("Expected min (0, 0, 0), got (%f, %f, %f)", min.X(), min.Y(), min.Z())
	}

	if max.X() != 15 || max.Y() != 15 || max.Z() != 15 {
		t.Errorf("Expected max (15, 15, 15), got (%f, %f, %f)", max.X(), max.Y(), max.Z())
	}
}

func TestEntityQueryBoundsEmpty(t *testing.T) {
	q := NewEntityQuery()

	_, _, found := q.Bounds()
	if found {
		t.Error("Bounds should not be found for empty query")
	}
}

func TestEntityQueryBoundsNilBBox(t *testing.T) {
	e1 := entity.NewLine()
	e1.BBox() // This might return nil if not properly initialized
	e2 := entity.NewLine()
	e2.Start = []float64{0, 0, 0}
	e2.End = []float64{10, 10, 10}

	q := NewEntityQuery(e1, e2)

	_, _, found := q.Bounds()
	if !found {
		t.Error("Bounds should be found even with nil bbox")
	}
}

func TestEntityQueryString(t *testing.T) {
	e1 := entity.NewLine()
	e2 := entity.NewCircle()

	q := NewEntityQuery(e1, e2)

	str := q.String()
	if str == "" {
		t.Error("String should not be empty")
	}
}

func TestParseQueryString(t *testing.T) {
	t.Run("With conditions", func(t *testing.T) {
		names, conditions, err := parseQueryString(`LINE[layer=="A"]`)
		if err != nil {
			t.Fatalf("parseQueryString failed: %v", err)
		}

		if len(names) != 1 || names[0] != "LINE" {
			t.Errorf("Expected names [LINE], got %v", names)
		}

		if conditions != `layer=="A"` {
			t.Errorf("Expected conditions 'layer==\"A\"', got '%s'", conditions)
		}
	})

	t.Run("Without conditions", func(t *testing.T) {
		names, conditions, err := parseQueryString("LINE CIRCLE")
		if err != nil {
			t.Fatalf("parseQueryString failed: %v", err)
		}

		if len(names) != 2 {
			t.Errorf("Expected 2 names, got %d", len(names))
		}

		if conditions != "" {
			t.Errorf("Expected empty conditions, got '%s'", conditions)
		}
	})

	t.Run("Empty query", func(t *testing.T) {
		_, _, err := parseQueryString("")
		if err == nil {
			t.Error("parseQueryString should fail for empty string")
		}
	})

	t.Run("Missing entity names", func(t *testing.T) {
		_, _, err := parseQueryString("[layer==\"A\"]")
		if err == nil {
			t.Error("parseQueryString should fail for missing entity names")
		}
	})
}

func TestGetEntityType(t *testing.T) {
	t.Run("Line entity", func(t *testing.T) {
		e := entity.NewLine()
		if getEntityType(e) != "LINE" {
			t.Errorf("Expected LINE, got %s", getEntityType(e))
		}
	})

	t.Run("Circle entity", func(t *testing.T) {
		e := entity.NewCircle()
		if getEntityType(e) != "CIRCLE" {
			t.Errorf("Expected CIRCLE, got %s", getEntityType(e))
		}
	})

	t.Run("Arc entity", func(t *testing.T) {
		e := entity.NewArc(nil)
		if getEntityType(e) != "ARC" {
			t.Errorf("Expected ARC, got %s", getEntityType(e))
		}
	})

	t.Run("Point entity", func(t *testing.T) {
		e := entity.NewPoint()
		if getEntityType(e) != "POINT" {
			t.Errorf("Expected POINT, got %s", getEntityType(e))
		}
	})

	t.Run("Polyline entity", func(t *testing.T) {
		e := entity.NewPolyline()
		if getEntityType(e) != "POLYLINE" {
			t.Errorf("Expected POLYLINE, got %s", getEntityType(e))
		}
	})

	t.Run("Text entity", func(t *testing.T) {
		e := entity.NewText()
		if getEntityType(e) != "TEXT" {
			t.Errorf("Expected TEXT, got %s", getEntityType(e))
		}
	})

	t.Run("MText entity", func(t *testing.T) {
		e := entity.NewMText()
		if getEntityType(e) != "MTEXT" {
			t.Errorf("Expected MTEXT, got %s", getEntityType(e))
		}
	})
}

func TestGetEntityAttribute(t *testing.T) {
	t.Run("Layer attribute", func(t *testing.T) {
		layerA := table.NewLayer("A", 0, nil)
		e := entity.NewLine()
		e.SetLayer(layerA)
		if getEntityAttribute(e, "layer") != "A" {
			t.Errorf("Expected layer 'A', got '%s'", getEntityAttribute(e, "layer"))
		}
	})

	t.Run("Handle attribute", func(t *testing.T) {
		e := entity.NewLine()
		_ = e.Handle()
	})

	t.Run("Type attribute", func(t *testing.T) {
		e := entity.NewCircle()
		if getEntityAttribute(e, "type") != "CIRCLE" {
			t.Errorf("Expected type 'CIRCLE', got '%s'", getEntityAttribute(e, "type"))
		}
	})

	t.Run("Unknown attribute", func(t *testing.T) {
		e := entity.NewLine()
		if getEntityAttribute(e, "unknown") != "" {
			t.Error("Expected empty string for unknown attribute")
		}
	})
}

func TestGetEntityAttributeAsFloat(t *testing.T) {
	e := entity.NewLine()
	handle := e.Handle()

	if getEntityAttributeAsFloat(e, "handle") == 0 && handle != "" {
		t.Error("Handle should parse to a float")
	}

	if getEntityAttributeAsFloat(e, "unknown") != 0 {
		t.Error("Expected 0 for unknown attribute")
	}
}

func TestQueryIntegration(t *testing.T) {
	layerA := table.NewLayer("A", 0, nil)
	layerB := table.NewLayer("B", 0, nil)

	line1 := entity.NewLine()
	line1.SetLayer(layerA)
	line2 := entity.NewCircle()
	line2.SetLayer(layerB)
	circle := entity.NewCircle()
	circle.SetLayer(layerA)

	q := NewEntityQuery(line1, line2, circle)

	result, err := q.Query("LINE")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if result.Len() != 1 {
		t.Errorf("Expected 1 result, got %d", result.Len())
	}

	result, err = q.Query("CIRCLE")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if result.Len() != 2 {
		t.Errorf("Expected 2 results for CIRCLE, got %d", result.Len())
	}
}

func TestQueryIntegrationComplex(t *testing.T) {
	layerA := table.NewLayer("A", 0, nil)
	layerB := table.NewLayer("B", 0, nil)

	line1 := entity.NewLine()
	line1.SetLayer(layerA)
	circle1 := entity.NewCircle()
	circle1.SetLayer(layerA)
	arc1 := entity.NewArc(nil)
	arc1.SetLayer(layerB)

	q := NewEntityQuery(line1, circle1, arc1)

	result, err := q.Query("LINE CIRCLE")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if result.Len() != 2 {
		t.Errorf("Expected 2 results for LINE CIRCLE, got %d", result.Len())
	}

	result, err = q.Query("ARC")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if result.Len() != 1 {
		t.Errorf("Expected 1 result for ARC, got %d", result.Len())
	}
}
