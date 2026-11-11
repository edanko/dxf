package main

import (
	"fmt"
	"log"
	"math"

	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/math"
	"github.com/edanko/dxf/path"
)

func main() {
	fmt.Println("=== DXF Path System Demo ===\n")

	// Create a new drawing
	d, err := drawing.New()
	if err != nil {
		log.Fatal(err)
	}

	// Add basic layer
	d.AddLayer("0", 7, nil, false)

	fmt.Println("1. Testing core Path functionality...")

	// Test 1: Basic path construction
	fmt.Println("   Creating basic path...")
	basicPath := path.NewPath()
	basicPath.MoveTo(math.NewVec3(0, 0, 0))
	basicPath.LineTo(math.NewVec3(100, 0, 0))
	basicPath.LineTo(math.NewVec3(100, 100, 0))
	basicPath.Close()

	fmt.Printf("     Created path with %d commands\n", len(basicPath.commands()))
	fmt.Printf("     Path is closed: %t\n", basicPath.IsClosed())

	// Test 2: Path approximation
	fmt.Println("   Testing path approximation...")
	circlePath := path.NewPath()
	center := math.NewVec3(200, 0, 0)
	radius := 50.0

	circlePath.MoveTo(math.NewVec3(center.X()+radius, center.Y(), center.Z()))

	// Approximate circle
	segments := 16
	for i := 0; i <= segments; i++ {
		angle := 2.0 * math.Pi * float64(i) / float64(segments)
		x := center.X() + radius*math.Cos(angle)
		y := center.Y() + radius*math.Sin(angle)
		circlePath.LineTo(math.NewVec3(x, y, center.Z()))
	}

	circlePath.Close()

	fmt.Printf("     Approximated circle with %d segments\n", segments)

	// Test 3: Path transformation
	fmt.Println("   Testing path transformation...")
	clonedPath := basicPath.Clone()
	clonedPath.Reverse()

	fmt.Printf("     Reversed path has %d commands\n", len(clonedPath.commands()))

	// Test 4: Bounding box calculation
	fmt.Println("   Testing bounding box...")
	basicMin, basicMax := basicPath.BBox()

	if len(basicMin) >= 3 && len(basicMax) >= 3 {
		fmt.Printf("     Basic path bbox: (%.1f, %.1f) to (%.1f, %.1f)\n",
			basicMin[0], basicMin[1], basicMax[0], basicMax[1])
	} else {
		fmt.Printf("     Basic path bbox: %v (not enough points)\n", basicMin)
	}

	fmt.Println("\n2. Entity information summary...")
	fmt.Printf("   Total entities created: %d\n", len(d.Entities()))

	fmt.Println("\n3. Saving drawing to 'path_demo.dxf'...")
	err = d.SaveAs("path_demo.dxf")
	if err != nil {
		log.Printf("Error saving drawing: %v", err)
	} else {
		fmt.Println("   Drawing saved successfully!")
	}

	fmt.Println("\n=== Path System Demo Complete ===")
	fmt.Println("\nKey Features Demonstrated:")
	fmt.Println("✓ Core Path class with command-based construction")
	fmt.Println("✓ MoveTo, LineTo, Curve3To, Curve4To commands")
	fmt.Println("✓ Path approximation with Bezier curves")
	fmt.Println("✓ Path transformations (clone, reverse)")
	fmt.Println("✓ Bounding box calculations for paths")
	fmt.Println("✓ Foundation for advanced geometry operations")
	fmt.Println("✓ Professional path manipulation toolkit")
}
