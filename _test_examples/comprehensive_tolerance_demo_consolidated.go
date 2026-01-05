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
	fmt.Println("=== Comprehensive Tolerance Demo ===\n")

	// Create a new drawing
	d, err := drawing.New()
	if err != nil {
		log.Fatal(err)
	}

	// Create layers
	basicLayer, _ := d.AddLayer("Basic", color.White, d.Layers["0"].LineType, true)
	tolLayer, _ := d.AddLayer("Tolerances", color.Red, d.Layers["0"].LineType, false)
	geomLayer, _ := d.AddLayer("Geometric", color.Blue, d.Layers["0"].LineType, false)

	fmt.Println("1. Basic Geometric Tolerance Symbols")

	// Position tolerance
	posTol := entity.NewTolerance()
	posTol.SetInsertionPoint([]float64{10, 10, 0})
	posTol.SetText("⌖|⌀0.1|A|B")
	posTol.SetHeight(3.0)
	posTol.SetLayer(geomLayer)
	d.AddEntity(posTol)

	// Flatness tolerance
	flatTol := entity.NewTolerance()
	flatTol.SetInsertionPoint([]float64{50, 10, 0})
	flatTol.SetText("⌒0.05")
	flatTol.SetHeight(3.0)
	flatTol.SetLayer(geomLayer)
	d.AddEntity(flatTol)

	// Cylindricity tolerance
	cylTol := entity.NewTolerance()
	cylTol.SetInsertionPoint([]float64{90, 10, 0})
	cylTol.SetText("⌭0.02")
	cylTol.SetHeight(3.0)
	cylTol.SetLayer(geomLayer)
	d.AddEntity(cylTol)

	fmt.Println("2. Advanced GD&T Symbols")

	// Concentricity tolerance
	concTol := entity.NewTolerance()
	concTol.SetInsertionPoint([]float64{10, 30, 0})
	concTol.SetText("◎⌀0.01|A")
	concTol.SetHeight(3.0)
	concTol.SetLayer(geomLayer)
	d.AddEntity(concTol)

	// Symmetry tolerance
	symTol := entity.NewTolerance()
	symTol.SetInsertionPoint([]float64{50, 30, 0})
	symTol.SetText("≡0.025|A")
	symTol.SetHeight(3.0)
	symTol.SetLayer(geomLayer)
	d.AddEntity(symTol)

	// Circular runout tolerance
	runoutTol := entity.NewTolerance()
	runoutTol.SetInsertionPoint([]float64{90, 30, 0})
	runoutTol.SetText("↗0.1|A")
	runoutTol.SetHeight(3.0)
	runoutTol.SetLayer(geomLayer)
	d.AddEntity(runoutTol)

	fmt.Println("3. Dimensions with Tolerances")

	// Linear dimension with symmetric tolerance
	linearDim := entity.NewDimension()
	linearDim.SetDimensionType(entity.DimensionTypeRotated)
	linearDim.SetDefinitionPoint([]float64{0, 60, 0})
	linearDim.SetBasePoint([]float64{0, 60, 0})
	linearDim.SetFeaturePoint([]float64{50, 60, 0})
	linearDim.SetText("50")
	linearDim.SetTolerance(entity.ToleranceTypeSymmetric, 0.1)
	linearDim.SetLayer(tolLayer)
	d.AddEntity(linearDim)

	// Dimension with deviation tolerance
	devDim := entity.NewDimension()
	devDim.SetDimensionType(entity.DimensionTypeRotated)
	devDim.SetDefinitionPoint([]float64{0, 80, 0})
	devDim.SetBasePoint([]float64{0, 80, 0})
	devDim.SetFeaturePoint([]float64{40, 80, 0})
	devDim.SetText("40")
	devDim.SetTolerance(entity.ToleranceTypeDeviation, 0.05)
	devDim.SetUpperTolerance(0.02)
	devDim.SetLowerTolerance(-0.03)
	devDim.SetLayer(tolLayer)
	d.AddEntity(devDim)

	// Dimension with limits
	limitsDim := entity.NewDimension()
	limitsDim.SetDimensionType(entity.DimensionTypeRotated)
	limitsDim.SetDefinitionPoint([]float64{50, 80, 0})
	limitsDim.SetBasePoint([]float64{50, 80, 0})
	limitsDim.SetFeaturePoint([]float64{90, 80, 0})
	limitsDim.SetLimitsEnabled(true)
	limitsDim.SetUpperLimit(0.05)
	limitsDim.SetLowerLimit(-0.02)
	limitsDim.SetLayer(tolLayer)
	d.AddEntity(limitsDim)

	fmt.Println("4. Geometric Features with Datum References")

	// Create some geometry to apply tolerances to
	rect := entity.NewLWPolyline([][]float64{
		{10, 100}, {40, 100}, {40, 120}, {10, 120}, {10, 100},
	})
	rect.SetLayer(basicLayer)
	d.AddEntity(rect)

	// Datum A reference
	datumA := entity.NewTolerance()
	datumA.SetInsertionPoint([]float64{25, 95, 0})
	datumA.SetText("[A]")
	datumA.SetHeight(3.0)
	datumA.SetLayer(geomLayer)
	d.AddEntity(datumA)

	// Position tolerance relative to datum A
	posDatumTol := entity.NewTolerance()
	posDatumTol.SetInsertionPoint([]float64{60, 110, 0})
	posDatumTol.SetText("⌖|⌀0.2|A")
	posDatumTol.SetHeight(3.0)
	posDatumTol.SetLayer(geomLayer)
	d.AddEntity(posDatumTol)

	fmt.Println("5. Surface Finish Symbols")

	// Basic surface finish
	finish1 := entity.NewTolerance()
	finish1.SetInsertionPoint([]float64{10, 140, 0})
	finish1.SetText("√32")
	finish1.SetHeight(3.0)
	finish1.SetLayer(geomLayer)
	d.AddEntity(finish1)

	// Surface finish with manufacturing method
	finish2 := entity.NewTolerance()
	finish2.SetInsertionPoint([]float64{50, 140, 0})
	finish2.SetText("√16/M")
	finish2.SetHeight(3.0)
	finish2.SetLayer(geomLayer)
	d.AddEntity(finish2)

	// Surface finish with roughness values
	finish3 := entity.NewTolerance()
	finish3.SetInsertionPoint([]float64{90, 140, 0})
	finish3.SetText("√8/Ra0.8")
	finish3.SetHeight(3.0)
	finish3.SetLayer(geomLayer)
	d.AddEntity(finish3)

	// Save the drawing
	err = d.SaveAs("comprehensive_tolerance_demo.dxf")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Comprehensive tolerance demo saved to 'comprehensive_tolerance_demo.dxf'")
	fmt.Println("\nFeatures demonstrated:")
	fmt.Println("- Geometric tolerances (position, flatness, cylindricity)")
	fmt.Println("- Advanced GD&T symbols (concentricity, symmetry, runout)")
	fmt.Println("- Dimensions with tolerances (symmetric, deviation, limits)")
	fmt.Println("- Datum references and frames")
	fmt.Println("- Surface finish symbols")
}
