// Package main provides a quickstart example for the DXF library.
//
// This example demonstrates:
// - Creating a new DXF drawing
// - Adding basic entities (points, lines, circles, arcs)
// - Loading and saving DXF files
// - Querying entities by type and layer
//
// Run with: go run examples/quickstart/main.go
package main

import (
	"fmt"
	"log"

	"github.com/edanko/dxf"
	"github.com/edanko/dxf/color"
)

func main() {
	fmt.Println("=== DXF Library Quickstart Demo ===")

	// Create a new drawing
	d, err := dxf.NewDrawing()
	if err != nil {
		log.Fatal("Failed to create drawing:", err)
	}
	fmt.Println("✓ Created new drawing")

	// Add some layers
	d.AddLayer("LayerA", color.Red, d.LtContinuous())
	d.AddLayer("LayerB", color.Blue, d.LtContinuous())
	fmt.Println("✓ Added layers: LayerA (Red), LayerB (Blue)")

	// Add entities
	d.Point(0, 0, 0)
	d.Point(100, 100, 0)
	d.Point(200, 0, 0)
	fmt.Println("✓ Added 3 points")

	// Add a line
	d.Line(0, 0, 0, 100, 100, 0)
	fmt.Println("✓ Added line")

	// Add circles
	d.Circle(0, 0, 0, 50)
	d.Circle(200, 0, 0, 75)
	fmt.Println("✓ Added 2 circles")

	// Add an arc
	d.Arc(100, 100, 0, 50, 0, 180)
	fmt.Println("✓ Added arc")

	// Set active layer
	d.SetActiveLayer("LayerA")

	// Add more entities on LayerA
	d.Circle(50, 200, 0, 25)
	fmt.Println("✓ Added circle on LayerA")

	// Save to file
	err = d.SaveAs("quickstart.dxf")
	if err != nil {
		log.Fatal("Failed to save:", err)
	}
	fmt.Println("✓ Saved to quickstart.dxf")

	// Load the file back
	d2, err := dxf.FromFile("quickstart.dxf")
	if err != nil {
		log.Fatal("Failed to load:", err)
	}
	fmt.Println("✓ Loaded quickstart.dxf back")

	// Query entities
	entities := d2.GetAllEntities()
	fmt.Printf("✓ Total entities: %d\n", len(entities))

	// Group by type
	typeCounts := make(map[string]int)
	for _, e := range entities {
		typeCounts[e.DXFType()]++
	}
	fmt.Println("\nEntity breakdown:")
	for etype, count := range typeCounts {
		fmt.Printf("  - %s: %d\n", etype, count)
	}

	fmt.Println("\n=== Demo Complete ===")
}
