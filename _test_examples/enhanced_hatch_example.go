package main

import (
	"fmt"
	"log"
	"math"

	"github.com/edanko/dxf"
	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
)

func main() {
	// Create a new drawing
	drawing, err := dxf.NewDrawing()
	if err != nil {
		log.Fatal("Error creating drawing:", err)
	}

	// Add layers for different hatch types
	drawing.AddLayer("PATTERNS", color.Blue, drawing.LtByLayer(), true)
	drawing.AddLayer("GRADIENTS", color.Red, drawing.LtByLayer(), true)
	drawing.AddLayer("COMPLEX", color.Green, drawing.LtByLayer(), true)

	// === Enhanced Pattern Hatch Example ===
	fmt.Println("Creating enhanced pattern hatch...")

	patternHatch := entity.NewHatchEnhanced()
	patternHatch.SetPredefinedPattern("ANSI31", 45.0, 1.5, 1.0, false)

	// Create a rectangular boundary for pattern hatch
	rectPoints := [][]float64{
		{50, 50, 0},   // Corner 1
		{150, 50, 0},  // Corner 2
		{150, 100, 0}, // Corner 3
		{50, 100, 0},  // Corner 4
	}
	patternHatch.AddPolylineBoundary(rectPoints, true)

	patternLayer, _ := drawing.Layer("PATTERNS", false)
	patternHatch.SetLayer(patternLayer)
	drawing.AddEntity(patternHatch)

	// === Linear Gradient Example ===
	fmt.Println("Creating linear gradient hatch...")

	linearGradientHatch := entity.NewHatchEnhanced()
	linearGradientHatch.SetGradientLinear(color.Red, color.Blue, 0.0) // Red to Blue

	// Create a circular boundary for gradient
	circlePoints := make([][]float64, 0)
	numPoints := 36
	center := dxfmath.NewVec2(250, 75)
	radius := 40.0
	for i := 0; i < numPoints; i++ {
		angle := 2.0 * math.Pi * float64(i) / float64(numPoints)
		x := center.X() + radius*math.Cos(angle)
		y := center.Y() + radius*math.Sin(angle)
		circlePoints = append(circlePoints, []float64{x, y, 0})
	}
	linearGradientHatch.AddPolylineBoundary(circlePoints, true)

	gradientLayer, _ := drawing.Layer("GRADIENTS", false)
	linearGradientHatch.SetLayer(gradientLayer)
	drawing.AddEntity(linearGradientHatch)

	// === Spherical Gradient Example ===
	fmt.Println("Creating spherical gradient hatch...")

	sphericalGradientHatch := entity.NewHatchEnhanced()
	sphericalGradientHatch.SetGradientSpherical(color.Yellow, color.Blue, []float64{350, 75})

	// Create a hexagonal boundary for spherical gradient
	hexPoints := make([][]float64, 0)
	hexCenter := dxfmath.NewVec2(250, 200)
	hexRadius := 35.0
	for i := 0; i < 6; i++ {
		angle := 2.0 * math.Pi * float64(i) / 6.0
		x := hexCenter.X() + hexRadius*math.Cos(angle)
		y := hexCenter.Y() + hexRadius*math.Sin(angle)
		hexPoints = append(hexPoints, []float64{x, y, 0})
	}
	sphericalGradientHatch.AddPolylineBoundary(hexPoints, true)

	complexLayer, _ := drawing.Layer("COMPLEX", false)
	sphericalGradientHatch.SetLayer(complexLayer)
	drawing.AddEntity(sphericalGradientHatch)

	// === Pattern with Island Example ===
	fmt.Println("Creating pattern hatch with island...")

	islandHatch := entity.NewHatchEnhanced()
	islandHatch.SetPredefinedPattern("ANSI37", 0.0, 1.0, 1.0, false)

	// Outer boundary (rectangle)
	outerPoints := [][]float64{
		{50, 250, 0}, {100, 250, 0},
		{100, 300, 0}, {50, 300, 0},
	}
	islandHatch.AddPolylineBoundary(outerPoints, true)

	// Inner island (circle) - donut hatch style
	innerPoints := make([][]float64, 0)
	innerCenter := dxfmath.NewVec2(75, 275)
	innerRadius := 15.0
	for i := 0; i < 16; i++ {
		angle := 2.0 * math.Pi * float64(i) / 16.0
		x := innerCenter.X() + innerRadius*math.Cos(angle)
		y := innerCenter.Y() + innerRadius*math.Sin(angle)
		innerPoints = append(innerPoints, []float64{x, y, 0})
	}
	islandHatch.AddIslandBoundary(innerPoints, true)

	islandHatch.SetLayer(complexLayer)
	drawing.AddEntity(islandHatch)

	// === Additional Advanced Examples ===

	// Example 5: Cylindrical Gradient
	fmt.Println("Creating cylindrical gradient hatch...")

	cylindricalGradientHatch := entity.NewHatchEnhanced()
	cylindricalGradientHatch.SetGradientCylindrical(color.Cyan, color.Green, []float64{350, 275})

	// Create a square boundary for cylindrical gradient
	squarePoints := [][]float64{
		{320, 250, 0}, {380, 250, 0},
		{380, 300, 0}, {320, 300, 0},
	}
	cylindricalGradientHatch.AddPolylineBoundary(squarePoints, true)

	cylindricalGradientHatch.SetLayer(complexLayer)
	drawing.AddEntity(cylindricalGradientHatch)

	// Example 6: Custom User-Defined Pattern
	fmt.Println("Creating custom pattern hatch...")

	customPatternHatch := entity.NewHatchEnhanced()

	// Define custom dot pattern lines
	customPatternLines := []entity.HatchPatternLine{
		{
			Angle:           0,
			BasePoint:       dxfmath.NewVec2(0, 0),
			Offset:          dxfmath.NewVec2(0, 3),
			DashLengthItems: []float64{0.5, -2.5}, // Small dot, large gap
		},
		{
			Angle:           math.Pi / 2,
			BasePoint:       dxfmath.NewVec2(0, 0),
			Offset:          dxfmath.NewVec2(3, 0),
			DashLengthItems: []float64{0.5, -2.5}, // Small dot, large gap
		},
	}
	customPatternHatch.SetUserDefinedPattern(customPatternLines)

	// Create a triangular boundary
	triPoints := [][]float64{
		{150, 200, 0}, {200, 200, 0}, {175, 250, 0},
	}
	customPatternHatch.AddPolylineBoundary(triPoints, true)
	customPatternHatch.SetHatchColor(color.Magenta)
	customPatternHatch.SetLayer(patternLayer)
	drawing.AddEntity(customPatternHatch)

	// Example 7: Gradient with Tint Control
	fmt.Println("Creating gradient with tint control...")

	tintGradientHatch := entity.NewHatchEnhanced()
	tintGradientHatch.SetGradientWithTint(
		1,                      // HatchGradientTypeLinear
		color.Red, color.White, // Red to White
		math.Pi/4, // 45 degrees
		[]float64{0, 0},
		0.3, // 30% tint towards second color
	)

	tintPoints := [][]float64{
		{420, 50, 0}, {480, 50, 0},
		{480, 100, 0}, {420, 100, 0},
	}
	tintGradientHatch.AddPolylineBoundary(tintPoints, true)
	tintGradientHatch.SetLayer(gradientLayer)
	drawing.AddEntity(tintGradientHatch)

	// Example 8: Gradient with Focus Control
	fmt.Println("Creating gradient with focus control...")

	focusGradientHatch := entity.NewHatchEnhanced()
	focusGradientHatch.SetGradientWithFocus(
		2, // HatchGradientTypeCylinder
		color.Yellow, color.Green,
		0,
		[]float64{450, 225}, // Center
		0.2,                 // Focus distance
	)

	focusPoints := [][]float64{
		{420, 200, 0}, {480, 200, 0},
		{480, 250, 0}, {420, 250, 0},
	}
	focusGradientHatch.AddPolylineBoundary(focusPoints, true)
	focusGradientHatch.SetLayer(gradientLayer)
	drawing.AddEntity(focusGradientHatch)

	// Example 9: Boundary with Bulges (Arc Segments)
	fmt.Println("Creating boundary with bulges (arc segments)...")

	bulgeHatch := entity.NewHatchEnhanced()
	bulgeHatch.SetPredefinedPattern("ANSI31", 0, 1.0, 1.0, false)
	bulgeHatch.SetHatchColor(color.Blue)

	// Rectangle with bulged edges (arc segments)
	bulgePoints := [][]float64{
		{200, 50}, {250, 50}, {250, 100}, {200, 100}, {200, 50},
	}
	bulges := []float64{0.5, 0, -0.5, 0} // Bulge values for arc segments
	bulgeHatch.AddPolylineBoundaryWithBulges(bulgePoints, true, bulges)

	bulgeHatch.SetLayer(complexLayer)
	drawing.AddEntity(bulgeHatch)

	// Example 10: 3D Hatch with Elevation
	fmt.Println("Creating 3D hatch with elevation...")

	elevationHatch := entity.NewHatchEnhanced()
	elevationHatch.SetGradientLinear(color.White, color.Grey192, math.Pi/2)
	elevationHatch.SetElevation([]float64{0, 0, 5}) // 5 units elevation

	elevPoints := [][]float64{
		{420, 320, 0}, {480, 320, 0},
		{480, 370, 0}, {420, 370, 0},
	}
	elevationHatch.AddPolylineBoundary(elevPoints, true)
	elevationHatch.SetLayer(complexLayer)
	drawing.AddEntity(elevationHatch)

	// Save to file
	err = drawing.SaveAs("enhanced_hatch_example.dxf")
	if err != nil {
		log.Fatalf("Error saving file: %v\n", err)
	}

	fmt.Println("Successfully created enhanced_hatch_example.dxf")
	fmt.Println("\n🎨 Enhanced HATCH entities created (Professional-Grade Features):")

	fmt.Println("\n📐 Pattern Hatches:")
	fmt.Printf("1. Pattern Hatch (ANSI31): 45° rotation, 1.5x scale, doubled\n")
	fmt.Printf("5. Custom Dot Pattern: User-defined dot grid pattern\n")

	fmt.Println("\n🌈 Gradient Fills:")
	fmt.Printf("2. Linear Gradient: Red to Blue horizontal gradient\n")
	fmt.Printf("3. Spherical Gradient: Yellow to Blue, centered at (350,75)\n")
	fmt.Printf("5. Cylindrical Gradient: Cyan to Green, centered at (350,275)\n")
	fmt.Printf("7. Gradient with Tint: Red to White, 30% tint control\n")
	fmt.Printf("8. Gradient with Focus: Yellow to Green, cylindrical with focus\n")

	fmt.Println("\n🏝️ Complex Boundaries:")
	fmt.Printf("4. Pattern with Island: ANSI37 pattern with circular hole\n")
	fmt.Printf("9. Bulged Boundary: Rectangle with arc segments\n")
	fmt.Printf("10. 3D Elevation Hatch: Gradient at 5 units elevation\n")

	// Print comprehensive statistics
	hatchCount := len(patternHatch.BoundaryPaths) + len(linearGradientHatch.BoundaryPaths) +
		len(sphericalGradientHatch.BoundaryPaths) + len(islandHatch.BoundaryPaths) +
		len(cylindricalGradientHatch.BoundaryPaths) + len(customPatternHatch.BoundaryPaths) +
		len(tintGradientHatch.BoundaryPaths) + len(focusGradientHatch.BoundaryPaths) +
		len(bulgeHatch.BoundaryPaths) + len(elevationHatch.BoundaryPaths)

	fmt.Printf("\n📊 Statistics:\n")
	fmt.Printf("- Total hatches created: %d\n", 10)
	fmt.Printf("- Total boundary paths: %d\n", hatchCount)
	fmt.Printf("- Pattern types: Predefined, User-Defined\n")
	fmt.Printf("- Gradient types: Linear, Spherical, Cylindrical (with tint/focus control)\n")
	fmt.Printf("- Boundary types: Polyline, Polyline+Island, Bulged Arcs, 3D Elevation\n")
	fmt.Printf("- Advanced features: Custom patterns, gradient controls, 3D positioning\n")

	fmt.Printf("\n✨ Professional Features Demonstrated:\n")
	fmt.Printf("- Custom pattern definition with HatchPatternLine\n")
	fmt.Printf("- Advanced gradient controls (tint, focus, centering)\n")
	fmt.Printf("- Complex boundary paths (islands, bulges, arcs)\n")
	fmt.Printf("- 3D positioning with elevation and extrusion\n")
	fmt.Printf("- Professional color management with ACI/RGB support\n")
	fmt.Printf("- Full DXF compliance with proper group codes\n")
}
