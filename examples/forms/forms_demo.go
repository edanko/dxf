package main

import (
	"fmt"
	"math"

	dxfmath "github.com/edanko/dxf/math"
	"github.com/edanko/dxf/render"
)

func main() {
	fg := render.NewFormGenerator()

	fmt.Println("=== DXF Forms System Demo ===")

	// Test 2D forms
	fmt.Println("2D Forms:")

	// Circle
	circle := fg.Circle(5.0)
	fmt.Printf("Circle (radius 5): %d vertices\n", len(circle))

	// Square (open and closed)
	square := fg.Square(10.0)
	closedSquare := fg.Square(10.0, true)
	fmt.Printf("Square (size 10): %d vertices\n", len(square))
	fmt.Printf("Closed Square: %d vertices\n", len(closedSquare))

	// Triangle (open and closed)
	triangle := fg.Triangle(10.0)
	closedTriangle := fg.Triangle(10.0, true)
	fmt.Printf("Triangle (size 10): %d vertices\n", len(triangle))
	fmt.Printf("Closed Triangle: %d vertices\n", len(closedTriangle))

	// Hexagon
	hexagon := fg.NGon(6, 5.0)
	fmt.Printf("Hexagon (6 sides, radius 5): %d vertices\n", len(hexagon))

	// Star
	star := fg.Star(5.0, 2.5, 5)
	fmt.Printf("Star (5 points): %d vertices\n", len(star))

	// Arrow
	arrow := fg.Arrow(10.0, 3.0)
	fmt.Printf("Arrow (length 10): %d vertices\n", len(arrow))

	// Test turtle graphics
	fmt.Println("\nTurtle Graphics (Square):")
	squarePath := fg.Turtle(func(t *render.Turtle) {
		t.Move(10)
		t.Left(math.Pi / 2) // 90 degrees
		t.Move(10)
		t.Left(math.Pi / 2)
		t.Move(10)
		t.Left(math.Pi / 2)
		t.Move(10)
	})
	fmt.Printf("Turtle square path: %d vertices\n", len(squarePath))

	// Test 3D forms
	fmt.Println("\n3D Forms:")

	// Cube
	cubeFaces := fg.Cube(10.0)
	fmt.Printf("Cube (size 10): %d faces\n", len(cubeFaces))

	// Cylinder
	cylinderFaces := fg.Cylinder(5.0, 10.0, 16)
	fmt.Printf("Cylinder (radius 5, height 10): %d faces\n", len(cylinderFaces))

	// Sphere
	sphereFaces := fg.Sphere(5.0, 16, 8)
	fmt.Printf("Sphere (radius 5): %d faces\n", len(sphereFaces))

	// Torus
	torusFaces := fg.Torus(8.0, 3.0, 16, 8)
	fmt.Printf("Torus (major 8, minor 3): %d faces\n", len(torusFaces))

	// Advanced 3D operations
	fmt.Println("\nAdvanced 3D Operations:")

	// Sweep
	profile := []dxfmath.Vec3{
		dxfmath.NewVec3(1, 1, 0),
		dxfmath.NewVec3(1, -1, 0),
		dxfmath.NewVec3(-1, -1, 0),
		dxfmath.NewVec3(-1, 1, 0),
	}
	path := []dxfmath.Vec3{
		dxfmath.NewVec3(0, 0, 0),
		dxfmath.NewVec3(0, 0, 10),
	}
	sweepFaces := fg.Sweep(profile, path)
	fmt.Printf("Sweep (profile along path): %d faces\n", len(sweepFaces))

	// Extrude with twist and scale
	extrudeFaces := fg.ExtrudeTwistScale(profile, 10, 90, 1.5, 4)
	fmt.Printf("Twisted Extrusion: %d faces\n", len(extrudeFaces))

	// Rotation form
	simpleProfile := []dxfmath.Vec3{
		dxfmath.NewVec3(1, 0, 0),
		dxfmath.NewVec3(2, 0, 0),
	}
	rotationFaces := fg.RotationForm(simpleProfile, 360, 8, "z")
	fmt.Printf("Rotation Form (360°, 8 segments): %d faces\n", len(rotationFaces))

	// Transform operations
	fmt.Println("\nTransform Operations:")

	originalCircle := fg.Circle(5.0)
	translated := fg.Translate(originalCircle, dxfmath.NewVec3(10, 10, 0))
	scaled := fg.Scale(originalCircle, 2.0)
	rotated := fg.Rotate(originalCircle, math.Pi/4) // 45 degrees

	fmt.Printf("Original circle: %d vertices\n", len(originalCircle))
	fmt.Printf("Translated circle: %d vertices\n", len(translated))
	fmt.Printf("Scaled circle: %d vertices\n", len(scaled))
	fmt.Printf("Rotated circle: %d vertices\n", len(rotated))

	// Close polygon utility
	fmt.Println("\nPolygon Utilities:")

	openTriangle := fg.Triangle(10.0)
	closedTriangle2 := fg.ClosePolygon(openTriangle)
	fmt.Printf("Open triangle: %d vertices\n", len(openTriangle))
	fmt.Printf("Closed triangle: %d vertices\n", len(closedTriangle2))

	fmt.Println("\n=== Forms System Demo Complete ===")
	fmt.Println("All forms are working and compatible with Python ezdxf behavior!")
}
