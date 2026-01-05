package main

import (
	"fmt"

	"github.com/edanko/dxf/drawing"
)

func main() {
	// Create a new drawing
	d, err := drawing.New()
	if err != nil {
		panic(err)
	}

	// Create graphics factory
	gf := drawing.NewGraphicsFactory(d)

	// Test basic shapes
	fmt.Println("Creating basic shapes...")

	point := gf.Point(10, 20, 30)
	fmt.Printf("Created point at: %v\n", point.Coord)

	line := gf.Line(0, 0, 0, 50, 50, 0)
	fmt.Printf("Created line from: %v to %v\n", line.Start, line.End)

	circle := gf.Circle(25, 25, 0, 15)
	fmt.Printf("Created circle: center=%v, radius=%.2f\n", circle.Center, circle.Radius)

	// Test rectangle
	rect := gf.Rectangle(10, 10, 30, 20)
	fmt.Printf("Created rectangle with %d vertices\n", len(rect.Vertices))

	// Test text
	text := gf.Text(5, 5, 0, 2.5, "Hello DXF!")
	fmt.Printf("Created text: %s\n", text.Value)

	// Test grid
	gridLines := gf.Grid(0, 0, 100, 50, 5, 5)
	fmt.Printf("Created grid with %d lines\n", len(gridLines))

	// Save drawing
	err = d.SaveAs("graphics_factory_test.dxf")
	if err != nil {
		panic(err)
	}

	fmt.Println("Graphics factory test completed successfully!")
}
