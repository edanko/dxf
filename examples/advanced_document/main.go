package main

import (
	"fmt"

	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/table"
)

func main() {
	// Create a new drawing
	d, err := drawing.New()
	if err != nil {
		panic(err)
	}

	// Create graphics factory
	gf := drawing.NewGraphicsFactory(d)

	fmt.Println("=== Advanced Document Management Demo ===")

	// Test enhanced entity database
	fmt.Println("\n1. Entity Database Management:")
	db := d.GetEntityDatabase()
	fmt.Printf("   Total entities: %d\n", db.GetEntityCount())

	// Test all entity types
	fmt.Println("\n2. Entity Creation with Graphics Factory:")

	// Basic shapes
	point := gf.Point(10, 20, 30)
	fmt.Printf("   Point: %v\n", point.Coord)

	line := gf.Line(0, 0, 0, 50, 50, 0)
	fmt.Printf("   Line: %v to %v\n", line.Start, line.End)

	circle := gf.Circle(50, 50, 0, 25)
	fmt.Printf("   Circle: center=%v, radius=%.2f\n", circle.Center, circle.Radius)

	rect := gf.Rectangle(10, 10, 80, 60)
	fmt.Printf("   Rectangle: %d vertices\n", len(rect.Vertices))

	text := gf.Text(10, 80, 0, 2.5, "Sample Text")
	fmt.Printf("   Text: '%s' at height %.1f\n", text.Value, text.Height)

	// LWPolyline - lightweight polyline with multiple vertices
	lwPolyline := gf.LwPolyline(true,
		150, 100, 0, // Point 1
		200, 100, 0, // Point 2
		200, 200, 0, // Point 3
		150, 200, 0, // Point 4
	)
	fmt.Printf("   LWPolyline: %d vertices, closed=%t\n", len(lwPolyline.Vertices), lwPolyline.Closed)

	// MText - multiline text
	mtext := gf.MText(300, 150, 0, 50, 4.0, "Multiline\nText\nwith formatting")
	fmt.Printf("   MText: '%s' (%.1f x %.1f)\n", mtext.Value, mtext.RectWidth, mtext.Height)

	// Insert - block reference
	insert := gf.Insert("TEST_BLOCK", 400, 300, 0)
	fmt.Printf("   Insert: block='%s' at %v\n", insert.BlockName(), insert.InsertPoint())

	// Layer management
	fmt.Println("\n3. Layer Management:")
	layers := d.GetLayers()
	fmt.Printf("   Total layers: %d\n", len(layers))

	// Add new layer
	newLayer, err := d.AddLayer("DemoLayer", 3, nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("   Created layer: %s (color=%d)\n", newLayer.Name(), 3)

	// Test layer filtering
	filteredLayers := d.FilterLayers(func(l *table.Layer) bool {
		return l.Name() == "0" || l.Name() == "DemoLayer"
	})
	fmt.Printf("   Filtered layers: %d\n", len(filteredLayers))

	// Block management
	fmt.Println("\n4. Block Management:")
	blocks := d.GetBlocks()
	fmt.Printf("   Total blocks: %d\n", len(blocks))

	// Query system
	fmt.Println("\n5. Query System:")
	allEntities := d.GetAllEntities()
	fmt.Printf("   Total entities in query: %d\n", len(allEntities))

	// Type-specific queries
	lineEntities := d.GetEntitiesByType("LINE")
	fmt.Printf("   LINE entities: %d\n", len(lineEntities))

	circleEntities := d.GetEntitiesByType("CIRCLE")
	fmt.Printf("   CIRCLE entities: %d\n", len(circleEntities))

	// Attribute-based queries
	textQuery := d.FindEntitiesByAttribute("Value", "Sample Text")
	fmt.Printf("   Text entities with 'Sample Text': %d\n", textQuery.Len())

	// Statistics
	fmt.Println("\n6. Drawing Statistics:")
	stats := d.GetDrawingStatistics()
	fmt.Printf("   Entity Statistics:\n")
	for key, value := range stats {
		if key[0:6] == "entity" {
			fmt.Printf("     %s: %d\n", key, value)
		}
	}

	// Validation
	fmt.Println("\n7. Validation:")
	if err := d.ValidateEntityDatabase(); err != nil {
		fmt.Printf("   Validation error: %v\n", err)
	} else {
		fmt.Println("   ✓ Entity database validation passed")
	}

	// Save drawing
	fmt.Println("\n8. Saving Drawing:")
	err = d.SaveAs("advanced_dxf_demo.dxf")
	if err != nil {
		panic(err)
	}

	fmt.Println("   ✓ Saved as 'advanced_dxf_demo.dxf'")
	fmt.Println("\n=== Demo completed successfully! ===")
}
