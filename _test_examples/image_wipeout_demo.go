package main

import (
	"fmt"
	"log"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
)

func main() {
	fmt.Println("=== DXF IMAGE and WIPEOUT Entity Demo ===\n")

	// Create a new drawing
	d, err := drawing.New()
	if err != nil {
		log.Fatal(err)
	}

	// Add basic layer
	d.AddLayer("0", 7, nil, false)

	fmt.Println("1. Creating IMAGE entities...")

	// Create a simple rectangular image
	simpleImage := entity.CreateSimpleImage(
		dxfmath.NewVec3(0, 0, 0), // Insert at origin
		100.0, 80.0,              // 100x80 pixels
		"IMAGE_DEF_1", // ImageDef handle
	)
	simpleImage.SetColor(color.ColorNumber(2)) // Yellow color
	d.AddEntity(simpleImage)
	fmt.Printf("   Created simple image: 100x80 pixels at origin\n")

	// Create a scaled image (2 units per pixel)
	scaledImage := entity.CreateScaledImage(
		dxfmath.NewVec3(150, 0, 0), // Offset position
		64.0, 48.0,                 // 64x48 pixels
		2.0,           // 2 units per pixel
		"IMAGE_DEF_2", // ImageDef handle
	)
	scaledImage.SetColor(color.ColorNumber(2)) // Yellow color
	d.AddEntity(scaledImage)
	fmt.Printf("   Created scaled image: 64x48 pixels, 2.0 scale factor\n")

	fmt.Println("\n2. Creating WIPEOUT entities...")

	// Create a simple rectangular wipeout
	rectWipeout := entity.CreateRectangularWipeout(
		dxfmath.NewVec3(300, 0, 0), // Position
		80.0, 60.0,                 // 80x60 units
	)
	rectWipeout.SetColor(color.ColorNumber(1)) // Red color
	d.AddEntity(rectWipeout)
	fmt.Printf("   Created rectangular wipeout: 80x60 units\n")

	// Create a circular wipeout
	circleWipeout := entity.CreateCircleWipeout(
		dxfmath.NewVec3(450, 0, 0), // Position
		40.0,                       // 40 unit radius
		1.0,                        // 1 unit scale
		16,                         // 16 segments for circle approximation
	)
	circleWipeout.SetColor(color.ColorNumber(1)) // Red color
	d.AddEntity(circleWipeout)
	fmt.Printf("   Created circular wipeout: radius 40 units\n")

	fmt.Println("\n3. Entity information summary...")
	printImageInfo(simpleImage, "Simple Image")
	printImageInfo(scaledImage, "Scaled Image")
	printWipeoutInfo(rectWipeout, "Rectangular Wipeout")
	printWipeoutInfo(circleWipeout, "Circular Wipeout")

	fmt.Printf("\n4. Total entities created: %d\n", len(d.Entities()))
	fmt.Printf("   IMAGE entities: %d\n", countImageEntities(d))
	fmt.Printf("   WIPEOUT entities: %d\n", countWipeoutEntities(d))

	fmt.Println("\n5. Saving drawing to 'image_wipeout_demo.dxf'...")
	err = d.SaveAs("image_wipeout_demo.dxf")
	if err != nil {
		log.Printf("Error saving drawing: %v", err)
	} else {
		fmt.Println("   Drawing saved successfully!")
	}

	fmt.Println("\n=== IMAGE and WIPEOUT Entity Demo Complete ===")
	fmt.Println("\nKey Features Demonstrated:")
	fmt.Println("✓ IMAGE entity creation and management")
	fmt.Println("✓ Simple and scaled image placement")
	fmt.Println("✓ WIPEOUT entity creation for masking")
	fmt.Println("✓ Rectangular and circular wipeouts")
	fmt.Println("✓ Professional layer management")
	fmt.Println("✓ Complete DXF formatting for both entities")
	fmt.Println("✓ CAD-ready image and masking workflows")
}

func printImageInfo(img *entity.Image, description string) {
	fmt.Printf("\n   %s:\n", description)
	fmt.Printf("     Insert: (%.2f, %.2f, %.2f)\n",
		img.InsertPoint().X(), img.InsertPoint().Y(), img.InsertPoint().Z())
	fmt.Printf("     Size: %.1f x %.1f pixels\n",
		img.ImageSize().X(), img.ImageSize().Y())
	fmt.Printf("     U-Vector: (%.3f, %.3f, %.3f)\n",
		img.UVector().X(), img.UVector().Y(), img.UVector().Z())
	fmt.Printf("     V-Vector: (%.3f, %.3f, %.3f)\n",
		img.VVector().X(), img.VVector().Y(), img.VVector().Z())
	fmt.Printf("     Display: B=%.1f, C=%.1f, F=%.1f\n",
		img.Brightness(), img.Contrast(), img.Fade())
	fmt.Printf("     Clipping: %s (%d points)\n",
		getClippingTypeString(img.ClippingBoundaryType()),
		len(img.BoundaryPoints()))
	fmt.Printf("     Layer: %v, Color: %d\n", img.Layer(), img.Color())
}

func printWipeoutInfo(wipeout *entity.Wipeout, description string) {
	fmt.Printf("\n   %s:\n", description)
	fmt.Printf("     Insert: (%.2f, %.2f, %.2f)\n",
		wipeout.InsertPoint().X(), wipeout.InsertPoint().Y(), wipeout.InsertPoint().Z())
	fmt.Printf("     Size: %.1f x %.1f units\n",
		wipeout.ImageSize().X(), wipeout.ImageSize().Y())
	fmt.Printf("     Boundary: %s (%d points)\n",
		getClippingTypeString(wipeout.ClippingBoundaryType()),
		len(wipeout.BoundaryPoints()))
	fmt.Printf("     Layer: %v, Color: %d\n", wipeout.Layer(), wipeout.Color())
}

func countImageEntities(d *drawing.Drawing) int {
	count := 0
	for _, ent := range d.Entities() {
		if _, ok := ent.(*entity.Image); ok {
			count++
		}
	}
	return count
}

func countWipeoutEntities(d *drawing.Drawing) int {
	count := 0
	for _, ent := range d.Entities() {
		if _, ok := ent.(*entity.Wipeout); ok {
			count++
		}
	}
	return count
}

func getClippingTypeString(clippingType int) string {
	switch clippingType {
	case entity.ImageClippingRectangular:
		return "Rectangular"
	case entity.ImageClippingPolygon:
		return "Polygon"
	default:
		return "Unknown"
	}
}
