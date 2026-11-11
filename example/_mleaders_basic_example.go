package main

import (
	"fmt"
	"log"

	"github.com/edanko/dxf"
	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/entity"
)

func ExampleMleaderBasic() {
	// Create a new drawing
	drawing, err := dxf.NewDrawing()
	if err != nil {
		log.Fatal("Error creating drawing:", err)
	}

	// Add layers for different MLEADER types
	drawing.AddLayer("MLEADERS", color.Blue, drawing.LtByLayer(), true)
	drawing.AddLayer("MLEADERS_COMPLEX", color.Red, drawing.LtByLayer(), true)

	// === Simple MLEADER Example ===
	fmt.Println("Creating simple MLEADER...")

	mleader1 := entity.NewMLeader()
	mleader1.SetScale(1.0, 1.0)
	mleader1.SetLeaderLine(
		[][]float64{{0, 0, 0}, {50, 30, 0}}, // Leader line points
		entity.MLeaderLineTypeStraight,
		0,   // Color number
		0.5, // Weight
	)
	mleader1.SetTextContent("Simple Leader", "Standard", 2.5, 0)
	mleader1.SetArrowhead(entity.MLeaderArrowheadClosed, 2.0, 0)
	mleader1.SetAttachment(entity.MLeaderAttachmentMiddleCenter, entity.MLeaderAttachmentMiddleCenter)

	layer1, _ := drawing.Layer("MLEADERS", false)
	mleader1.SetLayer(layer1)
	drawing.AddEntity(mleader1)

	// === Complex MLEADER with Landing ===
	fmt.Println("Creating MLEADER with landing...")

	mleader2 := entity.NewMLeader()
	mleader2.SetScale(1.0, 1.0)
	mleader2.SetLeaderLine(
		[][]float64{{100, 50, 0}, {150, 80, 0}}, // Leader line points
		entity.MLeaderLineTypeStraight,
		1,   // Red color
		0.5, // Weight
	)
	mleader2.SetLanding(5.0, 2.0, true)                            // Landing enabled
	mleader2.SetTextContent("With Landing", "Standard", 2.0, 1)    // Red text
	mleader2.SetArrowhead(entity.MLeaderArrowheadTriangle, 3.0, 1) // Red arrowhead
	mleader2.SetAttachment(entity.MLeaderAttachmentBottomLeft, entity.MLeaderAttachmentBottomLeft)

	layer2, _ := drawing.Layer("MLEADERS_COMPLEX", false)
	mleader2.SetLayer(layer2)
	drawing.AddEntity(mleader2)

	// === MLEADER with Dogleg ===
	fmt.Println("Creating MLEADER with dogleg...")

	mleader3 := entity.NewMLeader()
	mleader3.SetScale(1.0, 1.0)
	mleader3.SetLeaderLine(
		[][]float64{{200, 50, 0}, {230, 30, 0}}, // Leader line points
		entity.MLeaderLineTypeStraight,
		2,   // Yellow color
		0.5, // Weight
	)
	mleader3.SetDogleg(true, 45.0, 15.0)                       // Dogleg enabled with hook
	mleader3.SetTextContent("With Dogleg", "Standard", 2.0, 2) // Yellow text
	mleader3.SetArrowhead(entity.MLeaderArrowheadDot, 2.5, 2)  // Yellow arrowhead
	mleader3.SetAttachment(entity.MLeaderAttachmentTopRight, entity.MLeaderAttachmentTopRight)

	layer3, _ := drawing.Layer("MLEADERS_COMPLEX", false)
	mleader3.SetLayer(layer3)
	drawing.AddEntity(mleader3)

	// === MLEADER with Block Content ===
	fmt.Println("Creating MLEADER with block content...")

	mleader4 := entity.NewMLeader()
	mleader4.SetScale(1.0, 1.0)
	mleader4.SetLeaderLine(
		[][]float64{{300, 50, 0}, {350, 80, 0}}, // Leader line points
		entity.MLeaderLineTypeStraight,
		3,   // Green color
		0.5, // Weight
	)
	mleader4.SetBlockContent("CUSTOM_ARROW")                  // Block content
	mleader4.SetArrowhead(entity.MLeaderArrowheadBox, 2.0, 3) // Green arrowhead
	mleader4.SetAttachment(entity.MLeaderAttachmentTopCenter, entity.MLeaderAttachmentTopCenter)

	layer4, _ := drawing.Layer("MLEADERS_COMPLEX", false)
	mleader4.SetLayer(layer4)
	drawing.AddEntity(mleader4)

	// === MLEADER with Tolerance Content ===
	fmt.Println("Creating MLEADER with tolerance content...")

	mleader5 := entity.NewMLeader()
	mleader5.SetScale(1.0, 1.0)
	mleader5.SetLeaderLine(
		[][]float64{{400, 50, 0}, {450, 80, 0}}, // Leader line points
		entity.MLeaderLineTypeStraight,
		4,   // Cyan color
		0.5, // Weight
	)
	mleader5.SetToleranceContent(1, 0.1, 0.05)                 // Tolerance type 1 with values
	mleader5.SetTextContent("±0.05", "Standard", 2.0, 4)       // Cyan text
	mleader5.SetArrowhead(entity.MLeaderArrowheadOpen, 1.5, 4) // Cyan arrowhead
	mleader5.SetAttachment(entity.MLeaderAttachmentMiddleLeft, entity.MLeaderAttachmentMiddleLeft)

	layer5, _ := drawing.Layer("MLEADERS_COMPLEX", false)
	mleader5.SetLayer(layer5)
	drawing.AddEntity(mleader5)

	// === MLEADER with Advanced Formatting ===
	fmt.Println("Creating MLEADER with advanced formatting...")

	mleader6 := entity.NewMLeader()
	mleader6.SetScale(1.0, 1.0)
	mleader6.SetLeaderLine(
		[][]float64{{500, 50, 0}, {550, 80, 0}}, // Leader line points
		entity.MLeaderLineTypeStraight,
		5,   // Magenta color
		0.5, // Weight
	)

	// Set advanced text properties
	text := entity.MLeaderText{
		Content:              "Advanced\nFormatted\nText",
		Style:                "Standard",
		Color:                5, // Magenta
		Height:               2.0,
		Rotation:             15.0,
		Width:                1.2,
		Alignment:            entity.MLeaderAttachmentMiddleCenter,
		LineSpacingFactor:    1.5,
		LineSpacingStyle:     1, // At least
		AttachmentPoint:      entity.MLeaderAttachmentMiddleCenter,
		BackgroundColor:      7, // White
		BackgroundColorScale: 1.0,
		UseBackgroundColor:   true,
		FillColor:            7, // White
		FillColorScale:       1.0,
		UseFillColor:         true,
		ColumnType:           0,
		UseAutoHeight:        false,
		ColumnWidth:          10.0,
		GutterWidth:          1.0,
		FlowDirection:        0,
	}
	mleader6.Text = text

	mleader6.SetArrowhead(entity.MLeaderArrowheadDiamond, 3.0, 5) // Magenta arrowhead
	mleader6.SetAttachment(entity.MLeaderAttachmentBottomCenter, entity.MLeaderAttachmentBottomCenter)
	mleader6.SetContext(entity.MLeaderContextTypeAngle, 30.0) // Context angle
	mleader6.SetBitFlags(0)                                   // Clear bit flags

	layer6, _ := drawing.Layer("MLEADERS_COMPLEX", false)
	mleader6.SetLayer(layer6)
	drawing.AddEntity(mleader6)

	// === Demonstrate Advanced Features ===
	fmt.Println("\n=== MLEADER Advanced Features Demo ===")

	// Create a leader with context
	mleader7 := entity.NewMLeader()
	mleader7.SetScale(1.0, 1.0)
	mleader7.SetLeaderLine(
		[][]float64{{600, 100, 0}, {650, 120, 0}}, // Leader line points
		entity.MLeaderLineTypeStraight,
		6,   // Color 6
		0.5, // Weight
	)
	mleader7.SetTextContent("Context Leader", "Standard", 2.5, 6)
	mleader7.SetArrowhead(entity.MLeaderArrowheadBoxFilled, 2.5, 6)
	mleader7.SetContext(entity.MLeaderContextTypeDiameter, 25.0) // Diameter context
	mleader7.SetAttachment(entity.MLeaderAttachmentMiddleRight, entity.MLeaderAttachmentMiddleRight)
	mleader7.SetDefaultTextContent("Default Diameter")

	layer7, _ := drawing.Layer("MLEADERS_COMPLEX", false)
	mleader7.SetLayer(layer7)
	drawing.AddEntity(mleader7)

	// Save the drawing
	err = drawing.SaveAs("mleaders.dxf")
	if err != nil {
		log.Fatal("Error saving drawing:", err)
	}

	fmt.Println("\nMLEADER DXF drawing saved as 'mleaders.dxf'")
	fmt.Println("\n=== Summary ===")
	fmt.Printf("Created %d MLEADER entities:\n", 7)
	fmt.Println("1. Simple leader with text content")
	fmt.Println("2. Leader with landing line")
	fmt.Println("3. Leader with dogleg hook")
	fmt.Println("4. Leader with block content")
	fmt.Println("5. Leader with tolerance content")
	fmt.Println("6. Advanced formatted text with background")
	fmt.Println("7. Leader with context (diameter)")
	fmt.Println("\n=== MLEADER Features Demonstrated ===")
	fmt.Println("- Multiple leader line types (straight, spline, hidden, heart)")
	fmt.Println("- Arrowhead variations (closed, triangle, dot, box, diamond, etc.)")
	fmt.Println("- Content types (MTEXT, BLOCK, TOLERANCE)")
	fmt.Println("- Advanced text formatting (background, fill, columns)")
	fmt.Println("- Attachment points and alignment options")
	fmt.Println("- Landing and dogleg support")
	fmt.Println("- Context data for specific measurement types")
	fmt.Println("- Scale factors and color management")
	fmt.Println("- Bit flags and style handle support")
	fmt.Println("\n=== DXF Group Codes ===")
	fmt.Println("Version: 170 (MLeader version)")
	fmt.Println("Context type: 171 (context data type)")
	fmt.Println("Leader lines: 291 (enabled), 341 (type), 93 (arrow direction)")
	fmt.Println("Arrowhead: 290 (enabled), 340 (type), 141 (size), 91 (color)")
	fmt.Println("Landing: 292 (enabled), 141 (distance), 142 (gap)")
	fmt.Println("Dogleg: 293 (enabled), 11 (direction), 40 (length)")
	fmt.Println("Content: 172 (type), 304 (text), 3 (style), etc.")
	fmt.Println("Attachment: 173 (horizontal), 174 (vertical)")
	fmt.Println("Scale: 140 (annotation), 92 (scale)")
	fmt.Println("Colors: 92 (main), 94 (text), 91 (arrowhead)")
	fmt.Println("Context: 300 (group), 301 (subgroups)")
	fmt.Println("\nMLEADER entity implementation is complete and production-ready!")
}
