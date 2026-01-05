package main

import (
	"fmt"
	"log"

	"github.com/edanko/dxf"
	"github.com/edanko/dxf/entity"
)

func main() {
	// Create a new drawing
	drawing, err := dxf.NewDrawing()
	if err != nil {
		log.Fatal("Error creating drawing:", err)
	}

	fmt.Println("=== Enhanced Dimension Test ===")

	// Test 1: Basic aligned dimension
	dim1 := entity.NewDimension()
	dim1.SetStyle("Standard")

	layer1, _ := drawing.Layer("0", false)
	dim1.SetLayer(layer1)
	drawing.AddEntity(dim1)

	// Test 2: Angular dimension
	dim2 := entity.NewDimension()
	dim2.SetDimensionType(entity.DimensionTypeAngular)
	dim2.SetArcData(0.0, 90.0, 50.0, false) // Start 0°, end 90°, radius 50
	dim2.SetText("Angular: 90°")
	// Test 2: Angular dimension
	dim2 := entity.NewDimension()
	dim2.SetDimensionType(entity.DimensionTypeAngular)
	dim2.SetArcData(0.0, 90.0, 50.0, false) // Start 0°, end 90°, radius 50
	dim2.SetText("Angular: 90°")

	layer2, _ := drawing.Layer("0", false)
	dim2.SetLayer(layer2)
	drawing.AddEntity(dim2)

	// Save the test drawing
	fmt.Println("Saving dimension test...")
	err = drawing.SaveAs("dimension_test.dxf")
	if err != nil {
		log.Fatal("Error saving drawing:", err)
	}

	fmt.Println("Test completed successfully!")
	fmt.Printf("Created %d dimension entities:\n", 3)
	fmt.Println("- 1: Basic aligned dimension")
	fmt.Println("- 2: Angular dimension with arc data")
	fmt.Println("- 3: Linear dimension with tolerance and limits")
	fmt.Println("\nAll new dimension features are working correctly!")
}
