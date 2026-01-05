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

	// Add layers for different entities
	drawing.AddLayer("XLINES", color.Red, drawing.LtByLayer(), true)
	drawing.AddLayer("RAYS", color.Blue, drawing.LtByLayer(), true)
	drawing.AddLayer("INTERSECTIONS", color.Green, drawing.LtByLayer(), true)

	fmt.Println("Creating enhanced XLINE and RAY entities...")

	// === XLINE Examples ===
	fmt.Println("Creating XLINE entities...")

	// Horizontal XLINE
	horizontalXline := entity.NewXLineFromPointDirection(
		dxfmath.NewVec3(0, 10, 0),
		dxfmath.NewVec3(1, 0, 0),
	)
	xlineLayer, _ := drawing.Layer("XLINES", false)
	horizontalXline.SetLayer(xlineLayer)
	drawing.AddEntity(horizontalXline)

	// Vertical XLINE
	verticalXline := entity.NewXLineFromPointDirection(
		dxfmath.NewVec3(15, 0, 0),
		dxfmath.NewVec3(0, 1, 0),
	)
	verticalXline.SetLayer(xlineLayer)
	drawing.AddEntity(verticalXline)

	// Diagonal XLINE (45 degrees)
	diagonalXline := entity.NewXLineFromPointDirection(
		dxfmath.NewVec3(5, 5, 0),
		dxfmath.NewVec3(1, 1, 0),
	)
	diagonalXline.SetLayer(xlineLayer)
	drawing.AddEntity(diagonalXline)

	// Custom angled XLINE (30 degrees)
	angle := 30.0 * math.Pi / 180.0
	customXline := entity.NewXLineFromPointDirection(
		dxfmath.NewVec3(20, 5, 0),
		dxfmath.NewVec3(math.Cos(angle), math.Sin(angle), 0),
	)
	customXline.SetLayer(xlineLayer)
	drawing.AddEntity(customXline)

	// === RAY Examples ===
	fmt.Println("Creating RAY entities...")

	// Ray pointing right
	rightRay := entity.NewRayFromPointDirection(
		dxfmath.NewVec3(0, 20, 0),
		dxfmath.NewVec3(1, 0, 0),
	)
	rayLayer, _ := drawing.Layer("RAYS", false)
	rightRay.SetLayer(rayLayer)
	drawing.AddEntity(rightRay)

	// Ray pointing up-left
	upLeftRay := entity.NewRayFromPointDirection(
		dxfmath.NewVec3(25, 15, 0),
		dxfmath.NewVec3(-1, 1, 0),
	)
	upLeftRay.SetLayer(rayLayer)
	drawing.AddEntity(upLeftRay)

	// Ray at 60 degrees
	rayAngle := 60.0 * math.Pi / 180.0
	angledRay := entity.NewRayFromPointDirection(
		dxfmath.NewVec3(10, 25, 0),
		dxfmath.NewVec3(math.Cos(rayAngle), math.Sin(rayAngle), 0),
	)
	angledRay.SetLayer(rayLayer)
	drawing.AddEntity(angledRay)

	// === Mathematical Operations Demo ===
	fmt.Println("Demonstrating mathematical operations...")

	// Find intersection between horizontal and vertical XLINEs
	intersection, err := horizontalXline.IntersectWith(verticalXline)
	if err != nil {
		fmt.Printf("Intersection error: %v\n", err)
	} else {
		fmt.Printf("Intersection of horizontal and vertical XLINEs: (%.2f, %.2f)\n",
			intersection.X(), intersection.Y())

		// Add a point at intersection
		intersectionPoint := entity.NewPoint()
		intersectionLayer, _ := drawing.Layer("INTERSECTIONS", false)
		intersectionPoint.SetLayer(intersectionLayer)
		intersectionPoint.SetCoord(intersection)
		drawing.AddEntity(intersectionPoint)
	}

	// Get intersection layer for other uses
	intersectionLayer, _ := drawing.Layer("INTERSECTIONS", false)

	// Check parallelism
	isParallel := horizontalXline.IsParallelTo(diagonalXline)
	fmt.Printf("Horizontal and diagonal XLINEs are parallel: %v\n", isParallel)

	// Check perpendicular XLINE
	perpXline := horizontalXline.OrthogonalAt(dxfmath.NewVec3(10, 10, 0))
	perpXline.SetLayer(intersectionLayer)
	drawing.AddEntity(perpXline)

	// Ray and XLINE intersection
	rayIntersection, err := rightRay.IntersectWithXLine(verticalXline)
	if err != nil {
		fmt.Printf("Ray-XLINE intersection error: %v\n", err)
	} else {
		fmt.Printf("Ray-XLINE intersection: (%.2f, %.2f)\n",
			rayIntersection.X(), rayIntersection.Y())

		// Add a point at ray intersection
		rayPoint := entity.NewPoint()
		rayPoint.SetLayer(intersectionLayer)
		rayPoint.SetCoord(rayIntersection)
		drawing.AddEntity(rayPoint)
	}

	// === Distance Calculations Demo ===
	fmt.Println("Demonstrating distance calculations...")

	testPoint := dxfmath.NewVec3(7, 8, 0)

	xlineDist := horizontalXline.DistanceToPoint(testPoint)
	rayDist := rightRay.DistanceToPoint(testPoint)

	fmt.Printf("Distance from point (%.2f, %.2f) to horizontal XLINE: %.4f\n",
		testPoint.X(), testPoint.Y(), xlineDist)
	fmt.Printf("Distance from point (%.2f, %.2f) to right Ray: %.4f\n",
		testPoint.X(), testPoint.Y(), rayDist)

	// === Angle and Slope Demo ===
	fmt.Println("Demonstrating angle and slope calculations...")

	horizontalAngle := horizontalXline.Angle()
	horizontalSlope := horizontalXline.Slope()

	diagonalAngle := diagonalXline.Angle()
	diagonalSlope := diagonalXline.Slope()

	fmt.Printf("Horizontal XLINE - Angle: %.4f rad (%.1f°), Slope: %.4f\n",
		horizontalAngle, horizontalAngle*180/math.Pi, horizontalSlope)
	fmt.Printf("Diagonal XLINE - Angle: %.4f rad (%.1f°), Slope: %.4f\n",
		diagonalAngle, diagonalAngle*180/math.Pi, diagonalSlope)

	// === Validation Demo ===
	fmt.Println("Demonstrating validation...")

	// Valid XLINE creation
	validXline, err := entity.NewXLineFromPointDirectionValid(
		dxfmath.NewVec3(0, 0, 0),
		dxfmath.NewVec3(1, 1, 0),
	)
	if err != nil {
		fmt.Printf("Valid XLINE creation error: %v\n", err)
	} else {
		fmt.Printf("Valid XLINE created successfully: start=(%.2f, %.2f), direction=(%.2f, %.2f)\n",
			validXline.Start().X(), validXline.Start().Y(),
			validXline.UnitVector().X(), validXline.UnitVector().Y())
	}

	// Invalid XLINE creation (zero direction)
	_, err = entity.NewXLineFromPointDirectionValid(
		dxfmath.NewVec3(0, 0, 0),
		dxfmath.NewVec3(0, 0, 0),
	)
	if err != nil {
		fmt.Printf("Expected validation error for zero direction: %v\n", err)
	}

	// Save to file
	err = drawing.SaveAs("xline_ray_example.dxf")
	if err != nil {
		log.Fatalf("Error saving file: %v\n", err)
	}

	fmt.Println("Successfully created xline_ray_example.dxf")
	fmt.Println("\nXLINE and RAY entities created:")
	fmt.Printf("1. Horizontal XLINE at y=10\n")
	fmt.Printf("2. Vertical XLINE at x=15\n")
	fmt.Printf("3. Diagonal XLINE through (5,5)\n")
	fmt.Printf("4. Custom XLINE at 30° angle\n")
	fmt.Printf("5. Right-pointing Ray from (0,20)\n")
	fmt.Printf("6. Up-left pointing Ray from (25,15)\n")
	fmt.Printf("7. Angled Ray at 60° from (10,25)\n")
	fmt.Printf("8. Perpendicular XLINE at (10,10)\n")
	fmt.Printf("9. Intersection points marked in green\n")

	// Print statistics
	fmt.Printf("\nTotal XLINE entities: 4\n")
	fmt.Printf("Total RAY entities: 3\n")
	fmt.Printf("Intersection points: 2\n")
	fmt.Printf("Mathematical operations: Intersections, distances, angles, slopes\n")
	fmt.Printf("Layers: XLINES (red), RAYS (blue), INTERSECTIONS (green)\n")
}
