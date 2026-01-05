package main

import (
	"fmt"
	"log"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/math"
)

func main() {
	fmt.Println("=== Comprehensive Dimension System Demo ===\n")

	// Create a new drawing
	d, err := drawing.New()
	if err != nil {
		log.Fatal(err)
	}

	// Create layers for different dimension types
	dimLayer, _ := d.AddLayer("Dimensions", color.White, d.Layers["0"].LineType, true)
	tolLayer, _ := d.AddLayer("Tolerances", color.Red, d.Layers["0"].LineType, false)
	advLayer, _ := d.AddLayer("Advanced", color.Blue, d.Layers["0"].LineType, false)

	fmt.Println("1. Basic Linear Dimensions")

	// Horizontal dimension
	horizDim := entity.NewDimension()
	horizDim.SetDimensionType(entity.DimensionTypeRotated)
	horizDim.SetDefinitionPoint([]float64{0, 10, 0})
	horizDim.SetBasePoint([]float64{0, 10, 0})
	horizDim.SetFeaturePoint([]float64{50, 10, 0})
	horizDim.SetText("50.0")
	horizDim.SetLayer(dimLayer)
	d.AddEntity(horizDim)

	// Vertical dimension
	vertDim := entity.NewDimension()
	vertDim.SetDimensionType(entity.DimensionTypeRotated)
	vertDim.SetDefinitionPoint([]float64{60, 0, 0})
	vertDim.SetBasePoint([]float64{60, 0, 0})
	vertDim.SetFeaturePoint([]float64{60, 30, 0})
	vertDim.SetRotation(90.0 * math.Pi / 180)
	vertDim.SetText("30.0")
	vertDim.SetLayer(dimLayer)
	d.AddEntity(vertDim)

	// Aligned dimension
	alignedDim := entity.NewDimension()
	alignedDim.SetDimensionType(entity.DimensionTypeAligned)
	alignedDim.SetDefinitionPoint([]float64{80, 10, 0})
	alignedDim.SetBasePoint([]float64{80, 10, 0})
	alignedDim.SetFeaturePoint([]float64{120, 40, 0})
	alignedDim.SetText("ALIGNED")
	alignedDim.SetLayer(dimLayer)
	d.AddEntity(alignedDim)

	// Rotated dimension (30 degrees)
	rotatedDim := entity.NewDimension()
	rotatedDim.SetDimensionType(entity.DimensionTypeRotated)
	rotatedDim.SetDefinitionPoint([]float64{0, 60, 0})
	rotatedDim.SetBasePoint([]float64{0, 60, 0})
	rotatedDim.SetFeaturePoint([]float64{50, 60, 0})
	rotatedDim.SetRotation(30 * math.Pi / 180)
	rotatedDim.SetText("30° ROTATED")
	rotatedDim.SetLayer(dimLayer)
	d.AddEntity(rotatedDim)

	fmt.Println("2. Dimensions with Tolerances")

	// Linear dimension with tolerance
	toleranceDim := entity.NewDimension()
	toleranceDim.SetDimensionType(entity.DimensionTypeRotated)
	toleranceDim.SetDefinitionPoint([]float64{0, 90, 0})
	toleranceDim.SetBasePoint([]float64{0, 90, 0})
	toleranceDim.SetFeaturePoint([]float64{40, 90, 0})
	toleranceDim.SetText("40.0")
	toleranceDim.SetTolerance(entity.ToleranceTypeSymmetric, 0.1)
	toleranceDim.SetLayer(tolLayer)
	d.AddEntity(toleranceDim)

	// Dimension with limits
	limitsDim := entity.NewDimension()
	limitsDim.SetDimensionType(entity.DimensionTypeRotated)
	limitsDim.SetDefinitionPoint([]float64{50, 90, 0})
	limitsDim.SetBasePoint([]float64{50, 90, 0})
	limitsDim.SetFeaturePoint([]float64{90, 90, 0})
	limitsDim.SetLimitsEnabled(true)
	limitsDim.SetUpperLimit(0.05)
	limitsDim.SetLowerLimit(-0.02)
	limitsDim.SetLayer(tolLayer)
	d.AddEntity(limitsDim)

	fmt.Println("3. Angular Dimensions")

	// Angular dimension
	angularDim := entity.NewDimension()
	angularDim.SetDimensionType(entity.DimensionTypeAngular)
	angularDim.SetDefinitionPoint([]float64{150, 50, 0})
	angularDim.SetFeaturePoint([]float64{170, 70, 0})
	angularDim.SetBasePoint([]float64{190, 50, 0})
	angularDim.SetText("45°")
	angularDim.SetLayer(advLayer)
	d.AddEntity(angularDim)

	fmt.Println("4. Radius and Diameter Dimensions")

	// Create a circle for radius/diameter dimensions
	circle := entity.NewCircle([]float64{250, 50, 0}, 20)
	circle.SetLayer(dimLayer)
	d.AddEntity(circle)

	// Radius dimension
	radiusDim := entity.NewDimension()
	radiusDim.SetDimensionType(entity.DimensionTypeRadius)
	radiusDim.SetDefinitionPoint([]float64{250, 50, 0})
	radiusDim.SetFeaturePoint([]float64{270, 50, 0})
	radiusDim.SetText("R20")
	radiusDim.SetLayer(dimLayer)
	d.AddEntity(radiusDim)

	// Diameter dimension
	diameterDim := entity.NewDimension()
	diameterDim.SetDimensionType(entity.DimensionTypeDiameter)
	diameterDim.SetDefinitionPoint([]float64{250, 50, 0})
	diameterDim.SetFeaturePoint([]float64{230, 50, 0})
	diameterDim.SetText("Ø40")
	diameterDim.SetLayer(dimLayer)
	d.AddEntity(diameterDim)

	fmt.Println("5. Ordinate Dimensions")

	// X-axis ordinate dimension
	ordinateX := entity.NewDimension()
	ordinateX.SetDimensionType(entity.DimensionTypeOrdinate)
	ordinateX.SetDefinitionPoint([]float64{0, 0, 0})
	ordinateX.SetFeaturePoint([]float64{50, 150, 0})
	ordinateX.SetText("X50.0")
	ordinateX.SetLayer(advLayer)
	d.AddEntity(ordinateX)

	// Y-axis ordinate dimension
	ordinateY := entity.NewDimension()
	ordinateY.SetDimensionType(entity.DimensionTypeOrdinate)
	ordinateY.SetDefinitionPoint([]float64{0, 0, 0})
	ordinateY.SetFeaturePoint([]float64{80, 170, 0})
	ordinateY.SetText("Y170.0")
	ordinateY.SetOrdinateType(entity.OrdinateTypeY)
	ordinateY.SetLayer(advLayer)
	d.AddEntity(ordinateY)

	// Save the drawing
	err = d.SaveAs("comprehensive_dimension_demo.dxf")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Comprehensive dimension demo saved to 'comprehensive_dimension_demo.dxf'")
	fmt.Println("\nFeatures demonstrated:")
	fmt.Println("- Linear dimensions (horizontal, vertical, aligned, rotated)")
	fmt.Println("- Tolerance dimensions (symmetric, limits)")
	fmt.Println("- Angular dimensions")
	fmt.Println("- Radius and diameter dimensions")
	fmt.Println("- Ordinate dimensions (X and Y)")
	fmt.Println("- Multiple layers with different colors")
}
