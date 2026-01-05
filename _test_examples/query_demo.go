package main

import (
	"fmt"
	"log"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/query"
)

func main() {
	fmt.Println("=== DXF Entity Query System Demo ===\n")

	// Create a new drawing
	d, err := drawing.New()
	if err != nil {
		log.Fatal(err)
	}

	// Create layers with proper API
	constructionLayer, _ := d.AddLayer("construction", 7, nil, false)
	visibleLayer, _ := d.AddLayer("visible", 1, nil, false)
	hiddenLayer, _ := d.AddLayer("hidden", 2, nil, false)
	centerLayer, _ := d.AddLayer("center", 3, nil, false)

	// Create entities with proper API and layer assignment
	line1 := entity.NewLine()
	line1.Start = []float64{0, 0, 0}
	line1.End = []float64{10, 0, 0}
	line1.SetLayer(constructionLayer)

	line2 := entity.NewLine()
	line2.Start = []float64{10, 0, 0}
	line2.End = []float64{10, 10, 0}
	line2.SetLayer(visibleLayer)

	circle1 := entity.NewCircle()
	// Need to set center and radius through available methods
	circle1.SetLayer(visibleLayer)

	circle2 := entity.NewCircle()
	circle2.SetLayer(hiddenLayer)

	point1 := entity.NewPoint()
	point1.SetLayer(constructionLayer)

	point2 := entity.NewPoint()
	point2.SetLayer(centerLayer)

	// Add more entity types for comprehensive testing
	arc1 := entity.NewArc(entity.NewCircle()) // Arc needs circle entity
	arc1.SetLayer(visibleLayer)

	text1 := entity.NewText()
	text1.SetLayer(visibleLayer)

	polyline1 := entity.NewLWPolyline()
	polyline1.SetLayer(hiddenLayer)

	entities := []entity.Entity{
		line1, line2, circle1, circle2, point1, point2, arc1, text1, polyline1,
	}

	// Add entities to drawing for persistence
	for _, ent := range entities {
		d.AddEntity(ent)
	}

	// Create query with all entities
	q := query.NewEntityQuery(entities...)

	fmt.Printf("Total entities: %d\n", q.Len())
	fmt.Printf("Entity types: %s\n\n", q)

	// === Advanced Query Testing ===

	// Test 1: Select by entity type
	fmt.Println("1. Select all LINE entities:")
	lines := q.Select("LINE")
	fmt.Printf("Found %d LINE entities\n", lines.Len())

	// Test 2: Select multiple entity types
	fmt.Println("\n2. Select LINE and CIRCLE entities:")
	linesCircles := q.Select("LINE", "CIRCLE")
	fmt.Printf("Found %d LINE/CIRCLE entities\n", linesCircles.Len())

	// Test 3: Select all entities except specific type
	fmt.Println("\n3. Select all entities except POINT:")
	notPoints := q.Select("*", "!POINT")
	fmt.Printf("Found %d non-POINT entities\n", notPoints.Len())

	// Test 4: Attribute-based filtering
	fmt.Println("\n4. Filter by layer attribute:")
	visibleEntities := q.Filter(func(ent entity.Entity) bool {
		if layer := ent.Layer(); layer != nil {
			return layer.Name() == "visible"
		}
		return false
	})
	fmt.Printf("Found %d entities on 'visible' layer\n", visibleEntities.Len())

	// Test 5: Group entities by type
	fmt.Println("\n5. Group entities by type:")
	groups := q.GroupBy(getEntityType)
	for entityType, group := range groups {
		fmt.Printf("  %s: %d entities\n", entityType, group.Len())
	}

	// Test 6: Count by layer
	fmt.Println("\n6. Count entities by layer:")
	layerCounts := q.CountBy(func(ent entity.Entity) string {
		if layer := ent.Layer(); layer != nil {
			return layer.Name()
		}
		return "none"
	})
	for layerName, count := range layerCounts {
		fmt.Printf("  %s: %d entities\n", layerName, count)
	}

	// Test 7: First and Last entities
	fmt.Println("\n7. First and Last entities:")
	first := q.First()
	last := q.Last()
	if first != nil {
		fmt.Printf("  First entity: %s (Handle: %s)\n", getEntityType(first), first.Handle())
	}
	if last != nil {
		fmt.Printf("  Last entity: %s (Handle: %s)\n", getEntityType(last), last.Handle())
	}

	// Test 8: Bounding box calculation
	fmt.Println("\n8. Calculate bounding box of all entities:")
	min, max, found := q.Bounds()
	if found {
		fmt.Printf("  Bounding Box: Min(%.2f, %.2f, %.2f) to Max(%.2f, %.2f, %.2f)\n",
			min.X(), min.Y(), min.Z(),
			max.X(), max.Y(), max.Z())
	} else {
		fmt.Println("  No bounding box could be calculated")
	}

	// Test 9: Remove duplicates
	fmt.Println("\n9. Remove duplicates:")
	unique := q.Unique()
	fmt.Printf("Original: %d entities, Unique: %d entities\n", q.Len(), unique.Len())

	// Test 10: Each method for iteration
	fmt.Println("\n10. Iterate with Each method:")
	lineCount := 0
	q.Each(func(ent entity.Entity) {
		if getEntityType(ent) == "LINE" {
			lineCount++
		}
	})
	fmt.Printf("  Found %d LINE entities using Each method\n", lineCount)

	// Test 11: Map operation
	fmt.Println("\n11. Map entity handles:")
	handles := q.Map(func(ent entity.Entity) interface{} {
		return ent.Handle()
	})
	fmt.Printf("  Entity handles: %v...\n", handles[:min(3, len(handles))])

	// Test 12: Complex query chaining
	fmt.Println("\n12. Complex query chaining:")
	visibleLinesAndCircles := q.Filter(func(ent entity.Entity) bool {
		if layer := ent.Layer(); layer != nil {
			return layer.Name() == "visible"
		}
		return false
	}).Filter(func(ent entity.Entity) bool {
		entType := getEntityType(ent)
		return entType == "LINE" || entType == "CIRCLE"
	})
	fmt.Printf("  Visible lines and circles: %d\n", visibleLinesAndCircles.Len())

	// Save drawing to file
	err = d.SaveAs("query_demo.dxf")
	if err != nil {
		log.Printf("Warning: Could not save file: %v\n", err)
		fmt.Println("13. Query system testing complete (file save failed)")
	} else {
		fmt.Println("13. Query system testing complete and drawing saved!")
	}

	fmt.Println("\n=== Query System Demo Complete ===")
	fmt.Println("\n📊 Features Demonstrated:")
	fmt.Println("✓ Entity type selection (single, multiple, wildcard, exclusion)")
	fmt.Println("✓ Attribute-based filtering with custom predicates")
	fmt.Println("✓ Functional operations (Filter, Map, Each, GroupBy, CountBy)")
	fmt.Println("✓ Utility methods (First, Last, Unique, Bounds)")
	fmt.Println("✓ Query composition and chaining")
	fmt.Println("✓ Professional API for complex drawing analysis")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Helper function for query operations

func getEntityType(ent entity.Entity) string {
	switch ent.(type) {
	case *entity.Line:
		return "LINE"
	case *entity.Circle:
		return "CIRCLE"
	case *entity.Point:
		return "POINT"
	case *entity.Arc:
		return "ARC"
	case *entity.Text:
		return "TEXT"
	case *entity.LWPolyline:
		return "LWPOLYLINE"
	default:
		return "UNKNOWN"
	}
}
