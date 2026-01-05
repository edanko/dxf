package main

import (
	"fmt"
	"log"

	"github.com/edanko/dxf"
	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/math"
)

func ExampleEnhanced() {
	// Create a new drawing
	drawing, err := dxf.NewDrawing()
	if err != nil {
		log.Fatal("Error creating drawing:", err)
	}

	// Add basic layers
	drawing.AddLayer("SPLINES", color.Blue, drawing.LtByLayer(), true)
	drawing.AddLayer("ELLIPSES", color.Red, drawing.LtByLayer(), true)
	drawing.AddLayer("DIMENSIONS", color.Green, drawing.LtByLayer(), true)

	// === Enhanced SPLINE Examples ===
	fmt.Println("Creating enhanced SPLINE entities...")

	// Create a simple quadratic B-spline
	controlPoints := []math.Vec3{
		math.NewVec3(0, 0, 0),
		math.NewVec3(50, 50, 0),
		math.NewVec3(100, 0, 0),
		math.NewVec3(150, 50, 0),
		math.NewVec3(200, 0, 0),
	}

	enhancedBSpline := math.NewEnhancedBSpline(controlPoints, 2) // Quadratic spline

	// Convert to DXF Spline entity
	spline := entity.NewSplineFromEnhanced(enhancedBSpline)
	splineLayer, _ := drawing.Layer("SPLINES", false)
	spline.SetLayer(splineLayer)
	drawing.AddEntity(spline)

	// === Enhanced ELLIPSE Examples ===
	fmt.Println("Creating enhanced ELLIPSE entities...")

	// Create a full ellipse
	ellipse := entity.NewEllipse()
	ellipse.SetCenter(300, 100, 0)
	ellipse.SetRadius(80, 40) // Major=80, Minor=40
	ellipseLayer, _ := drawing.Layer("ELLIPSES", false)
	ellipse.SetLayer(ellipseLayer)
	drawing.AddEntity(ellipse)

	// Create an ellipse arc
	ellipseArc := entity.NewEllipse()
	ellipseArc.SetCenter(300, 250, 0)
	ellipseArc.SetRadius(60, 30)
	ellipseArc.SetAngles(0, 180) // Half ellipse
	ellipseLayer2, _ := drawing.Layer("ELLIPSES", false)
	ellipseArc.SetLayer(ellipseLayer2)
	drawing.AddEntity(ellipseArc)

	// === Enhanced DIMENSION Examples ===
	fmt.Println("Creating enhanced DIMENSION entities...")

	// Create linear dimensions
	dim1 := entity.NewDimension()
	dim1.SetText("100.0")
	dimLayer, _ := drawing.Layer("DIMENSIONS", false)
	dim1.SetLayer(dimLayer)

	// Set basic geometry through methods
	dim1.SetDimensionType(entity.DimensionTypeRotated)
	dim1.SetDefinitionPoint(math.NewVec3(50, -50, 0))
	dim1.SetTextMidpoint(math.NewVec3(100, -30, 0))
	dim1.SetAngle(0) // Horizontal dimension

	drawing.AddEntity(dim1)

	// Create radial dimension
	radialDim := entity.NewDimension()
	radialDim.SetText("R50.0")
	radialLayer, _ := drawing.Layer("DIMENSIONS", false)
	radialDim.SetLayer(radialLayer)
	radialDim.SetDimensionType(entity.DimensionTypeRadius)
	radialDim.SetDefinitionPoint(math.NewVec3(500, 100, 0))
	radialDim.SetTextMidpoint(math.NewVec3(560, 100, 0))

	drawing.AddEntity(radialDim)

	// === Demonstrate Mathematical Operations ===
	fmt.Println("\n=== Mathematical Operations Demo ===")

	// Sample points from B-spline (avoid flattening to prevent recursion issue)
	samplePoints := enhancedBSpline.SamplePoints(10)
	fmt.Printf("Enhanced B-spline sampled at %d points\n", len(samplePoints))

	// Get derivative information
	derivatives := enhancedBSpline.Derivative(0.5, 2) // 2nd derivative at t=0.5
	fmt.Printf("B-spline derivatives at t=0.5: %d derivatives computed\n", len(derivatives))

	// Demonstrate point on curve evaluation
	splinePoint := enhancedBSpline.EvaluatePoint(0.3) // Point at 30% along curve
	fmt.Printf("B-spline point at t=0.3: (%.2f, %.2f, %.2f)\n",
		splinePoint.X(), splinePoint.Y(), splinePoint.Z())

	// Demonstrate tangent calculations
	splineEntity := entity.NewSplineFromEnhanced(enhancedBSpline)
	splineLayer3, _ := drawing.Layer("SPLINES", false)
	splineEntity.SetLayer(splineLayer3)
	tangent := splineEntity.TangentAt(0.25) // Tangent at 25% along curve
	fmt.Printf("B-spline tangent at t=0.25: (%.2f, %.2f, %.2f)\n",
		tangent.X(), tangent.Y(), tangent.Z())

	// Demonstrate ellipse operations
	bezierCurve := ellipse.ToBezier()
	fmt.Printf("Ellipse converted to Bézier with %d control points\n", len(bezierCurve.ControlPoints()))

	ellipsePoint := ellipse.PointAt(0.5) // Midpoint of ellipse
	fmt.Printf("Ellipse midpoint: (%.2f, %.2f, %.2f)\n",
		ellipsePoint.X(), ellipsePoint.Y(), ellipsePoint.Z())

	// Save the drawing
	err = drawing.SaveAs("enhanced_entities.dxf")
	if err != nil {
		log.Fatal("Error saving drawing:", err)
	}

	fmt.Println("\nEnhanced DXF drawing saved as 'enhanced_entities.dxf'")
	fmt.Println("\n=== Summary ===")
	fmt.Printf("Created entities:\n")
	fmt.Printf("- 2 SPLINE entities\n")
	fmt.Printf("- 2 ELLIPSE entities (full ellipse and arc)\n")
	fmt.Printf("- 2 DIMENSION entities (linear and radial)\n")
	fmt.Printf("\nMathematical features demonstrated:\n")
	fmt.Printf("- Enhanced B-spline with sampling and evaluation\n")
	fmt.Printf("- Bézier curve approximation from ellipse\n")
	fmt.Printf("- Curve evaluation and derivative computation\n")
	fmt.Printf("- Tangent calculations\n")
	fmt.Printf("- Point-on-curve evaluation\n")
}
