package main

import (
	"fmt"
	"log"
	"math"

	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
)

func main() {
	fmt.Println("=== DXF HELIX Entity Demo ===\n")

	// Create a new drawing
	d, err := drawing.New()
	if err != nil {
		log.Fatal(err)
	}

	// Add basic layer
	d.AddLayer("0", 7, nil, false)

	fmt.Println("1. Creating basic HELIX geometries...")

	// Create a simple right-handed spring helix
	springHelix := entity.CreateSpringHelix(
		dxfmath.NewVec3(0, 0, 0),    // Start at origin
		2.0,                         // Radius
		10.0,                        // 10 turns
		1.0,                         // Turn height
		entity.HelixHandednessRight, // Right-handed
	)

	d.AddEntity(springHelix)
	fmt.Printf("   Created spring helix: %d turns, radius=%.1f\n",
		springHelix.Turns(), springHelix.Radius())

	// Create a left-handed helix
	leftHelix := entity.CreateSpringHelix(
		dxfmath.NewVec3(20, 0, 0),  // Offset position
		1.5,                        // Larger radius
		8.0,                        // 8 turns
		0.5,                        // Turn height
		entity.HelixHandednessLeft, // Left-handed
	)

	d.AddEntity(leftHelix)
	fmt.Printf("   Created left helix: %d turns, radius=%.1f\n",
		leftHelix.Turns(), leftHelix.Radius())

	// Create a tapered helix
	taperedHelix := entity.CreateTaperedHelix(
		dxfmath.NewVec3(40, 0, 0),   // Offset position
		1.0,                         // Start radius
		2.0,                         // End radius
		12.0,                        // 12 turns
		2.0,                         // Turn height
		entity.HelixHandednessRight, // Right-handed
	)

	d.AddEntity(taperedHelix)
	fmt.Printf("   Created tapered helix: %d turns, r1=%.1f, r2=%.1f\n",
		taperedHelix.Turns(), taperedHelix.Radius(), taperedHelix.BaseRadius())

	fmt.Println("\n2. Testing HELIX properties...")

	// Test helix properties
	testHelixProperties(springHelix, "Spring Helix")
	testHelixProperties(leftHelix, "Left Handed Helix")
	testHelixProperties(taperedHelix, "Tapered Helix")

	fmt.Println("\n3. Testing HELIX transformations...")

	// Test transformations by modifying existing helix
	testHelix := springHelix
	_ = testHelix.EndPoint() // Calculate but don't use for now

	// Apply a scale transform
	testHelix.SetRadius(testHelix.Radius() * 1.5)
	testHelix.SetTurns(testHelix.Turns() * 1.2)

	_ = testHelix.EndPoint() // Calculate but don't use for now
	fmt.Printf("   Scaled helix: radius x%.1f, turns x%.1f\n",
		testHelix.Radius(), testHelix.Turns())

	// Test axis rotation by 45 degrees
	rotatedHelix := entity.CreateSpringHelix(
		dxfmath.NewVec3(60, 0, 0),
		2.0,
		8.0,
		1.0,
		entity.HelixHandednessRight,
	)

	// Rotate axis vector by 45 degrees around Z axis
	cos45 := math.Cos(math.Pi / 4.0)
	sin45 := math.Sin(math.Pi / 4.0)
	rotatedAxis := dxfmath.NewVec3(sin45, cos45, 0)
	rotatedHelix.SetAxisVector(rotatedAxis)

	rotatedEnd := rotatedHelix.EndPoint()
	fmt.Printf("   Rotated helix: new end point (%.2f, %.2f, %.2f)\n",
		rotatedEnd.X(), rotatedEnd.Y(), rotatedEnd.Z())

	fmt.Println("\n4. Testing HELIX constraints...")

	// Test constrained helix
	constrainedHelix := entity.CreateConstrainedHelix(
		dxfmath.NewVec3(80, 0, 0),
		dxfmath.NewVec3(0, 0, 1), // Axis vector
		1.0,                      // Radius
		8.0,                      // Turns
		2.0,                      // Turn height
		entity.HelixHandednessRight,
	)

	d.AddEntity(constrainedHelix)
	fmt.Printf("   Created constrained helix: base radius=%.1f, base height=%.1f\n",
		constrainedHelix.BaseRadius(), constrainedHelix.BaseHeight())

	fmt.Println("\n5. Entity information summary...")
	printHelixInfo(springHelix, "Basic Spring Helix")
	printHelixInfo(leftHelix, "Left-handed Spring Helix")
	printHelixInfo(taperedHelix, "Tapered Helix")
	printHelixInfo(rotatedHelix, "Rotated Helix")
	printHelixInfo(constrainedHelix, "Constrained Helix")

	fmt.Printf("\n6. Total entities created: %d\n", len(d.Entities()))
	fmt.Printf("   HELIX entities: %d\n", countHelixEntities(d))

	fmt.Println("\n7. Saving drawing to 'helix_demo.dxf'...")
	err = d.SaveAs("helix_demo.dxf")
	if err != nil {
		log.Printf("Error saving drawing: %v", err)
	} else {
		fmt.Println("   Drawing saved successfully!")
	}

	fmt.Println("\n=== HELIX Entity Demo Complete ===")
	fmt.Println("\nKey Features Demonstrated:")
	fmt.Println("✓ HELIX entity creation and management")
	fmt.Println("✓ Spring helix with customizable parameters")
	fmt.Println("✓ Handedness control (left/right)")
	fmt.Println("✓ Turn and turn height configuration")
	fmt.Println("✓ Tapered helix (varying radius)")
	fmt.Println("✓ Constrained helix (fixed base radius/height)")
	fmt.Println("✓ Helix transformations (scaling, rotation)")
	fmt.Println("✓ Axis vector management")
	fmt.Println("✓ Complete HELIX mathematical implementation")
	fmt.Println("✓ DXF formatting with all HELIX properties")
	fmt.Println("✓ Professional 3D spiral/helix workflows")
}

func testHelixProperties(helix *entity.Helix, description string) {
	fmt.Printf("\n   %s:\n", description)
	fmt.Printf("     Start: (%.2f, %.2f, %.2f)\n",
		helix.StartPoint().X(), helix.StartPoint().Y(), helix.StartPoint().Z())
	fmt.Printf("     End: (%.2f, %.2f, %.2f)\n",
		helix.EndPoint().X(), helix.EndPoint().Y(), helix.EndPoint().Z())
	fmt.Printf("     Axis: (%.2f, %.2f, %.2f)\n",
		helix.AxisVector().X(), helix.AxisVector().Y(), helix.AxisVector().Z())
	fmt.Printf("     Radius: %.2f\n", helix.Radius())
	fmt.Printf("     Turns: %.1f\n", helix.Turns())
	fmt.Printf("     Turn Height: %.2f\n", helix.TurnHeight())
	fmt.Printf("     Total Height: %.2f\n", helix.TotalHeight())
	fmt.Printf("     Total Length: %.2f\n", helix.TotalLength())
	fmt.Printf("     Handedness: %s\n", getHandednessString(helix.Handedness()))
}

func testHelixTransformations(helix *entity.Helix) {
	fmt.Println("\n   Testing transformations:")

	// Test modification
	fmt.Printf("     Original radius: %.2f, turns: %.1f\n", helix.Radius(), helix.Turns())
	helix.SetRadius(helix.Radius() * 2.0)
	helix.SetTurns(helix.Turns() * 1.5)
	fmt.Printf("     Modified radius: %.2f, turns: %.1f\n", helix.Radius(), helix.Turns())
}

func printHelixInfo(helix *entity.Helix, description string) {
	fmt.Printf("\n   %s:\n", description)
	fmt.Printf("     Start: (%.2f, %.2f, %.2f)\n",
		helix.StartPoint().X(), helix.StartPoint().Y(), helix.StartPoint().Z())
	fmt.Printf("     End: (%.2f, %.2f, %.2f)\n",
		helix.EndPoint().X(), helix.EndPoint().Y(), helix.EndPoint().Z())
	fmt.Printf("     Axis: (%.2f, %.2f, %.2f)\n",
		helix.AxisVector().X(), helix.AxisVector().Y(), helix.AxisVector().Z())
	fmt.Printf("     Radius: %.2f\n", helix.Radius())
	fmt.Printf("     Turns: %.1f\n", helix.Turns())
	fmt.Printf("     Turn Height: %.2f\n", helix.TurnHeight())
	fmt.Printf("     Total Height: %.2f\n", helix.TotalHeight())
	fmt.Printf("     Total Length: %.2f\n", helix.TotalLength())
	fmt.Printf("     Handedness: %s\n", getHandednessString(helix.Handedness()))
	fmt.Printf("     Constrain Type: %s\n", getConstrainTypeString(helix.ConstrainType()))
	if helix.BaseRadius() != 0.0 || helix.BaseHeight() != 0.0 {
		fmt.Printf("     Base Radius: %.2f\n", helix.BaseRadius())
		fmt.Printf("     Base Height: %.2f\n", helix.BaseHeight())
	}
}

func countHelixEntities(d *drawing.Drawing) int {
	count := 0
	for _, ent := range d.Entities() {
		if _, ok := ent.(*entity.Helix); ok {
			count++
		}
	}
	return count
}

func getHandednessString(handedness int) string {
	switch handedness {
	case entity.HelixHandednessLeft:
		return "Left"
	case entity.HelixHandednessRight:
		return "Right"
	default:
		return "Unknown"
	}
}

func getConstrainTypeString(constrainType int) string {
	switch constrainType {
	case entity.HelixConstrainTurns:
		return "Constrain Turn Height"
	case entity.HelixConstrainTurnHeight:
		return "Constrain Turn Height"
	case entity.HelixConstrainBoth:
		return "Constrain Both"
	default:
		return "None"
	}
}
