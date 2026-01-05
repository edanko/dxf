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

	// Test basic entity creation
	_, err = d.Point(10, 20, 30)
	if err != nil {
		panic(err)
	}

	_, err = d.Line(0, 0, 0, 10, 10, 0)
	if err != nil {
		panic(err)
	}

	_, err = d.Circle(5, 5, 0, 2.5)
	if err != nil {
		panic(err)
	}

	// Test enhanced entity database
	db := d.GetEntityDatabase()
	fmt.Printf("Entity Count: %d\n", db.GetEntityCount())
	fmt.Printf("Total Drawing Entities: %d\n", d.GetEntityCount())

	// Test entity type queries
	entities := d.GetAllEntities()
	fmt.Printf("All entities: %d\n", len(entities))

	lineEntities := d.GetEntitiesByType("LINE")
	fmt.Printf("LINE entities: %d\n", len(lineEntities))

	pointEntities := d.GetEntitiesByType("POINT")
	fmt.Printf("POINT entities: %d\n", len(pointEntities))

	// Test layer management
	layers := d.GetLayers()
	fmt.Printf("Total layers: %d\n", len(layers))

	// Add a new layer with continuous line type
	continuousLineType := table.NewLineType("Continuous", "Solid Line")
	newLayer, err := d.AddLayer("TestLayer", 1, continuousLineType)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created layer: %s\n", newLayer.Name())

	// Test layer filtering
	testLayers := d.FilterLayers(func(l *table.Layer) bool {
		return l.Name() == "0" || l.Name() == "TestLayer"
	})
	fmt.Printf("Filtered layers: %d\n", len(testLayers))

	// Test query system
	query := d.Query()
	fmt.Printf("Query results: %d entities\n", query.Len())

	// Test select entities
	lineQuery := d.SelectEntities("LINE", "POINT")
	fmt.Printf("LINE+POINT query: %d entities\n", lineQuery.Len())

	// Test statistics
	stats := d.GetDrawingStatistics()
	fmt.Printf("Drawing Statistics:\n")
	for key, value := range stats {
		fmt.Printf("  %s: %v\n", key, value)
	}

	// Test validation
	if err := d.ValidateEntityDatabase(); err != nil {
		fmt.Printf("Validation error: %v\n", err)
	} else {
		fmt.Println("Entity database validation passed")
	}

	// Save drawing
	err = d.SaveAs("enhanced_drawing.dxf")
	if err != nil {
		panic(err)
	}

	fmt.Println("Enhanced document management test completed successfully!")
}
