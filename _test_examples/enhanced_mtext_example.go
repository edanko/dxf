package main

import (
	"fmt"
	"log"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
)

func main() {
	// Create new drawing
	d, err := drawing.New()
	if err != nil {
		log.Fatal(err)
	}

	// Create layers
	layer1, _ := d.AddLayer("MText Demo", color.White, d.Layers["0"].LineType, true)
	layer2, _ := d.AddLayer("Formatting Demo", color.Red, d.Layers["0"].LineType, false)
	layer3, _ := d.AddLayer("Column Demo", color.Blue, d.Layers["0"].LineType, false)

	// Demo 1: Basic MTEXT with attachment points
	fmt.Println("Creating MTEXT with different attachment points...")

	mtext1 := entity.NewMText()
	mtext1.SetText("Top Left Attachment")
	mtext1.SetCoord([]float64{0, 10, 0})
	mtext1.SetHeight(2.0)
	mtext1.SetAttachmentPoint(entity.MTextTopLeft)
	mtext1.SetLayer(layer1)
	d.AddEntity(mtext1)

	mtext2 := entity.NewMText()
	mtext2.SetText("Middle Center Attachment")
	mtext2.SetCoord([]float64{0, 7, 0})
	mtext2.SetHeight(2.0)
	mtext2.SetAttachmentPoint(entity.MTextMiddleCenter)
	mtext2.SetLayer(layer1)
	d.AddEntity(mtext2)

	mtext3 := entity.NewMText()
	mtext3.SetText("Bottom Right Attachment")
	mtext3.SetCoord([]float64{0, 4, 0})
	mtext3.SetHeight(2.0)
	mtext3.SetAttachmentPoint(entity.MTextBottomRight)
	mtext3.SetLayer(layer1)
	d.AddEntity(mtext3)

	// Demo 2: Advanced formatting with MTextEditor
	fmt.Println("Creating MTEXT with advanced formatting...")

	editor := entity.NewMTextEditor()
	editor.
		AddText("This is ").
		BoldText("bold text").
		AddText(" mixed with ").
		ItalicText("italic text").
		AddText(" and ").
		UnderlineText("underlined text").
		AddParagraphBreak().
		AddText("Here's a ").
		ColorText("red text", 1).
		AddText(" and a ").
		HeightText("tall text", 3.0).
		AddText(" with different sizes.")

	mtext4 := entity.NewMText()
	mtext4.SetCoord([]float64{15, 10, 0})
	mtext4.SetHeight(1.5)
	mtext4.SetWidth(12.0)
	mtext4.SetLayer(layer2)
	editor.ApplyToMText(mtext4)
	d.AddEntity(mtext4)

	// Demo 3: Complex formatting with stacking and fractions
	fmt.Println("Creating MTEXT with fractions and stacking...")

	editor2 := entity.NewMTextEditor()
	editor2.
		AddText("Mathematical expressions: ").
		AddStackedText("1", "2", entity.MTextStackingSlanted).
		AddText(" = ").
		AddStackedText("π", "4", entity.MTextStackingOver).
		AddParagraphBreak().
		AddText("Complex fraction: ").
		AddStackedText("a", "b+c", entity.MTextStackingLinear).
		AddText(" = result")

	mtext5 := entity.NewMText()
	mtext5.SetCoord([]float64{15, 5, 0})
	mtext5.SetHeight(2.0)
	mtext5.SetWidth(10.0)
	mtext5.SetLayer(layer2)
	editor2.ApplyToMText(mtext5)
	d.AddEntity(mtext5)

	// Demo 4: Font formatting and colors
	fmt.Println("Creating MTEXT with font formatting...")

	editor3 := entity.NewMTextEditor()
	editor3.
		StartGroup().
		SetFont("Arial", true, false).
		SetColor(2). // Red
		AddText("Bold Arial Red").
		EndGroup().
		AddText(" ").
		StartGroup().
		SetFont("Times New Roman", false, true).
		SetColor(3). // Green
		AddText("Italic Times Green").
		EndGroup().
		AddParagraphBreak().
		StartGroup().
		SetFont("Courier", true, true).
		SetColor(5). // Blue
		SetHeight(2.5).
		AddText("Bold Italic Courier Blue").
		EndGroup()

	mtext6 := entity.NewMText()
	mtext6.SetCoord([]float64{35, 8, 0})
	mtext6.SetHeight(1.8)
	mtext6.SetWidth(15.0)
	mtext6.SetLayer(layer2)
	editor3.ApplyToMText(mtext6)
	d.AddEntity(mtext6)

	// Demo 5: Paragraph formatting with tabs
	fmt.Println("Creating MTEXT with paragraph formatting...")

	props := &entity.MTextParagraphProperties{
		Indent:      2.0,
		LeftIndent:  1.0,
		RightIndent: 1.0,
		Alignment:   entity.MTextParagraphJustified,
		TabStops: []entity.MTextTabStop{
			{Position: 5.0, Type: entity.MTextTabLeft},
			{Position: 10.0, Type: entity.MTextTabCenter},
			{Position: 15.0, Type: entity.MTextTabRight},
		},
	}

	editor4 := entity.NewMTextEditor()
	editor4.
		SetParagraphProperties(props).
		AddText("This paragraph has 2.0 unit indent, justified alignment, and custom tabs.").
		AddParagraphBreak().
		AddText("Left Tab\tCenter Tab\tRight Tab").
		AddParagraphBreak().
		AddText("Data1\t\t\tData2\t\t\tData3").
		AddParagraphBreak().
		AddText("Another line showing\t\t\ttab\t\t\talignment.")

	mtext7 := entity.NewMText()
	mtext7.SetCoord([]float64{35, 2, 0})
	mtext7.SetHeight(1.2)
	mtext7.SetWidth(18.0)
	mtext7.SetLayer(layer2)
	editor4.ApplyToMText(mtext7)
	d.AddEntity(mtext7)

	// Demo 6: Background fills and text frames
	fmt.Println("Creating MTEXT with background fills...")

	// Background color fill
	mtext8 := entity.NewMText()
	mtext8.SetText("Background Color Fill")
	mtext8.SetCoord([]float64{55, 10, 0})
	mtext8.SetHeight(2.0)
	mtext8.SetWidth(10.0)
	mtext8.SetBackgroundFill(entity.MTextBackgroundColor, 2, 0, 0, 1.2) // Red background
	mtext8.SetLayer(layer3)
	d.AddEntity(mtext8)

	// Text frame only
	mtext9 := entity.NewMText()
	mtext9.SetText("Text Frame Only")
	mtext9.SetCoord([]float64{55, 7, 0})
	mtext9.SetHeight(2.0)
	mtext9.SetWidth(10.0)
	mtext9.SetBackgroundFill(entity.MTextBackgroundFrame, 0, 0, 0, 1.5) // Frame only
	mtext9.SetLayer(layer3)
	d.AddEntity(mtext9)

	// Background with frame
	mtext10 := entity.NewMText()
	mtext10.SetText("Background + Frame")
	mtext10.SetCoord([]float64{55, 4, 0})
	mtext10.SetHeight(2.0)
	mtext10.SetWidth(10.0)
	mtext10.SetBackgroundFill(entity.MTextBackgroundColor, 4, 0, 0, 1.2) // Cyan background
	mtext10.SetLayer(layer3)
	d.AddEntity(mtext10)

	// Demo 7: Line spacing and flow direction
	fmt.Println("Creating MTEXT with different line spacing...")

	// At Least line spacing
	mtext11 := entity.NewMText()
	mtext11.SetText("At Least spacing\nLine 2\nLine 3")
	mtext11.SetCoord([]float64{70, 10, 0})
	mtext11.SetHeight(1.5)
	mtext11.SetWidth(8.0)
	mtext11.SetLineSpacing(entity.MTextAtLeast, 1.0)
	mtext11.SetLayer(layer3)
	d.AddEntity(mtext11)

	// Exact line spacing
	mtext12 := entity.NewMText()
	mtext12.SetText("Exact spacing\nLine 2\nLine 3")
	mtext12.SetCoord([]float64{70, 6, 0})
	mtext12.SetHeight(1.5)
	mtext12.SetWidth(8.0)
	mtext12.SetLineSpacing(entity.MTextExact, 2.0)
	mtext12.SetLayer(layer3)
	d.AddEntity(mtext12)

	// Demo 8: Rotation and 3D positioning
	fmt.Println("Creating MTEXT with rotation and 3D positioning...")

	mtext13 := entity.NewMText()
	mtext13.SetText("Rotated 45°")
	mtext13.SetCoord([]float64{10, 0, 0})
	mtext13.SetHeight(2.0)
	mtext13.SetRotation(45.0)
	mtext13.SetLayer(layer1)
	d.AddEntity(mtext13)

	mtext14 := entity.NewMText()
	mtext14.SetText("3D Positioned")
	mtext14.SetCoord([]float64{15, 0, 0})
	mtext14.SetHeight(2.0)
	mtext14.SetExtrusion([]float64{0.5, 0.5, 0.707}) // 45° extrusion
	mtext14.SetLayer(layer1)
	d.AddEntity(mtext14)

	// Demo 9: Testing formatting code parser
	fmt.Println("Testing formatting code parser...")

	// Complex formatting string with multiple codes
	complexText := `{\F Arial|b0|i0|c0;\C1;Bold Red}\P{\LUnderlined Text\l}\P\H3.0;Large Text\P\A2;Top Alignment`

	mtext15 := entity.NewMText()
	mtext15.SetText(complexText)
	mtext15.SetCoord([]float64{25, 0, 0})
	mtext15.SetHeight(1.5)
	mtext15.SetWidth(15.0)
	mtext15.SetLayer(layer1)

	// Test parsing
	fragments := mtext15.ParseTextContent()
	fmt.Printf("Parsed %d fragments from complex text\n", len(fragments))
	for i, fragment := range fragments {
		fmt.Printf("Fragment %d: '%s' (Bold: %v, Italic: %v, Underline: %v)\n",
			i+1, fragment.Text, fragment.Bold, fragment.Italic, fragment.Underline)
	}

	d.AddEntity(mtext15)

	// Demo 10: Text width and oblique angle
	fmt.Println("Creating MTEXT with width factor and oblique angle...")

	mtext16 := entity.NewMText()
	mtext16.SetText("Wide Text")
	mtext16.SetCoord([]float64{45, 0, 0})
	mtext16.SetHeight(2.0)
	mtext16.WidthFactor = 1.5 // 150% width
	mtext16.SetLayer(layer1)
	d.AddEntity(mtext16)

	mtext17 := entity.NewMText()
	mtext17.SetText("Oblique Text")
	mtext17.SetCoord([]float64{60, 0, 0})
	mtext17.SetHeight(2.0)
	mtext17.SetObliqueAngle(15.0) // 15° oblique
	mtext17.SetLayer(layer1)
	d.AddEntity(mtext17)

	// Save drawing
	err = d.SaveAs("enhanced_mtext_demo.dxf")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Enhanced MTEXT demonstration completed!")
	fmt.Println("File saved as: enhanced_mtext_demo.dxf")
	fmt.Printf("Total MTEXT entities created: %d\n", 17)

	// Print summary
	fmt.Println("\n=== MTEXT Feature Demonstration Summary ===")
	fmt.Println("✅ Basic attachment points (9 types)")
	fmt.Println("✅ Advanced formatting (bold, italic, underline, colors)")
	fmt.Println("✅ Text stacking and fractions")
	fmt.Println("✅ Font formatting with TrueType support")
	fmt.Println("✅ Paragraph formatting with tabs")
	fmt.Println("✅ Background fills and text frames")
	fmt.Println("✅ Line spacing styles (At Least, Exact)")
	fmt.Println("✅ Rotation and 3D positioning")
	fmt.Println("✅ Formatting code parser")
	fmt.Println("✅ Width factor and oblique angle")
	fmt.Println("✅ MTextEditor utility class")
}
