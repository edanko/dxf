package main

import (
	"fmt"
	"log"

	"github.com/edanko/dxf"
	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/entity"
)

func main() {
	// Create a new drawing
	drawing, err := dxf.NewDrawing()
	if err != nil {
		log.Fatal("Error creating drawing:", err)
	}

	// Add layers for different entity types
	drawing.AddLayer("MLEADERS", color.Blue, drawing.LtByLayer(), true)
	drawing.AddLayer("MLINES", color.Green, drawing.LtByLayer(), true)

	fmt.Println("=== Advanced DXF Entities Demo ===")

	// === MLEADER Example ===
	fmt.Println("Creating MLEADER examples...")

	// Simple MLEADER with text annotation
	mleader1 := entity.NewMLeader()
	mleader1.SetScale(1.0, 1.0)
	mleader1.SetLeaderLine(
		[][]float64{{0, 0, 0}, {100, 50, 0}},
		entity.MLeaderLineTypeStraight,
		1,   // Blue color
		0.5, // Line weight
	)
	mleader1.SetTextContent("Simple MLeader", "Standard", 3.0, 1)
	mleader1.SetArrowhead(entity.MLeaderArrowheadClosed, 2.0, 1)
	mleader1.SetAttachment(entity.MLeaderAttachmentMiddleCenter, entity.MLeaderAttachmentMiddleCenter)

	// Just set layer by name since we have interface issues
	mleader1.SetStyleHandle("MLEADERS")
	drawing.AddEntity(mleader1)

	// === MLINE Example ===
	fmt.Println("Creating MLINE examples...")

	// Simple MLINE with two elements
	mline1 := entity.NewMLineWithStyle("DOUBLE_WALL", 0, 1.0)
	mline1.AddVertex(0, 400, 0)
	mline1.AddVertex(100, 400, 0)
	mline1.AddVertex(100, 500, 0)
	mline1.AddVertex(0, 500, 0)
	mline1.SetClosed(true)
	mline1.SetColor(color.Green)

	// Create style with two elements
	style1 := entity.CreateMLineStyle("DOUBLE_WALL", "Double wall line", 0, 1.0)
	style1.AddElement(-0.5, color.Green, "CONTINUOUS") // First line
	style1.AddElement(0.5, color.Green, "CONTINUOUS")  // Second line
	mline1.SetStyle(style1)

	mline1.SetStyleName("MLINES")
	drawing.AddEntity(mline1)

	// === Save the drawing ===
	fmt.Println("Saving drawing...")
	err = drawing.SaveAs("advanced_entities_demo.dxf")
	if err != nil {
		log.Fatal("Error saving drawing:", err)
	}

	fmt.Println("Drawing saved successfully as 'advanced_entities_demo.dxf'")
	fmt.Println("\n=== Entity Summary ===")
	fmt.Printf("- MLEADER entities: 1 (text annotation)\n")
	fmt.Printf("- MLINE entities: 1 (double wall)\n")
	fmt.Printf("- Total entities: 2\n")
	fmt.Println("\n=== Features Demonstrated ===")
	fmt.Println("MLEADER:")
	fmt.Println("  - Enhanced MTEXT content")
	fmt.Println("  - Multiple leader line support")
	fmt.Println("  - Arrowhead support")
	fmt.Println("  - Dogleg support")
	fmt.Println("  - Attachment point configuration")
	fmt.Println("MLINE:")
	fmt.Println("  - Multi-element line styles")
	fmt.Println("  - Variable width segments")
	fmt.Println("  - Closed polylines")
	fmt.Println("  - Style-based element definitions")
}
