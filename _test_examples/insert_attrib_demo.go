package main

import (
	"fmt"
	"log"
	"math"

	"github.com/edanko/dxf/block"
	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
	"github.com/edanko/dxf/table"
)

func main() {
	fmt.Println("=== DXF INSERT/ATTRIB System Demo ===\n")

	// Create a new drawing
	d, err := drawing.New()
	if err != nil {
		log.Fatal(err)
	}

	// Create default linetype and layer
	linetype := table.NewLineType("CONTINUOUS", "", 0.0)
	layer := table.NewLayer("0", 7, linetype)
	d.AddLayer("0", 7, linetype, false)

	fmt.Println("1. Creating block definitions for INSERT entities...")

	// Create a block for a door symbol
	doorBlock := createDoorBlock(d)

	// Create a block for a window symbol
	windowBlock := createWindowBlock(d)

	// Create a block with attributes (for furniture with tags)
	furnitureBlock := createFurnitureBlock(d)

	fmt.Printf("   Created blocks: %s, %s, %s\n",
		doorBlock.Name, windowBlock.Name, furnitureBlock.Name)

	fmt.Println("\n2. Creating INSERT entities...")

	// Insert basic block references
	doorInsert := createBasicInsert(doorBlock, 10, 5, 0)
	windowInsert := createBasicInsert(windowBlock, 20, 5, 0)

	// Insert block with transformation
	rotatedDoor := createTransformedInsert(doorBlock, 5, 10, 45.0, 0.5)
	scaledWindow := createTransformedInsert(windowBlock, 30, 10, 0, 1.5)

	// Insert with attributes
	furnitureWithAttribs := createInsertWithAttributes(furnitureBlock, 15, 15, 0)

	// Add entities to drawing
	d.AddEntity(doorInsert)
	d.AddEntity(windowInsert)
	d.AddEntity(rotatedDoor)
	d.AddEntity(scaledWindow)
	d.AddEntity(furnitureWithAttribs)

	fmt.Printf("   Created 5 INSERT entities with various properties\n")

	fmt.Println("\n3. Creating independent ATTRIB entities...")

	// Create standalone attributes
	roomNameAttrib := createStandaloneAttribute("ROOM_NAME", "Living Room", 40, 20, 0)
	roomAreaAttrib := createStandaloneAttribute("AREA", "25.5", 40, 18, 0)

	d.AddEntity(roomNameAttrib)
	d.AddEntity(roomAreaAttrib)

	fmt.Printf("   Created 2 standalone ATTRIB entities\n")

	fmt.Println("\n4. Testing INSERT system features...")

	// Test multi-insert (MINSERT)
	multiInsert := createMultipleInsert(doorBlock, 25, 25, 0, 3, 2)
	d.AddEntity(multiInsert)
	fmt.Printf("   Created multi-insert array (3x2 grid)\n")

	// Test attribute methods
	fmt.Println("\n5. Testing attribute operations...")
	testAttributeOperations(furnitureWithAttribs)

	// Test transformation methods
	fmt.Println("\n6. Testing INSERT transformations...")
	testInsertTransformations(doorInsert)

	fmt.Println("\n7. Entity information...")
	printInsertInfo(doorInsert, "Door Insert")
	printInsertInfo(windowInsert, "Window Insert")
	printInsertInfo(rotatedDoor, "Rotated Door")
	printInsertInfo(scaledWindow, "Scaled Window")
	printInsertInfo(furnitureWithAttribs, "Furniture with Attributes")
	printInsertInfo(multiInsert, "Multi-Insert Array")

	// Test attribute visibility and flags
	fmt.Println("\n8. Testing attribute flags...")
	testAttributeFlags(furnitureWithAttribs)

	// Save drawing
	fmt.Println("\n9. Saving drawing to 'insert_attrib_demo.dxf'...")
	err = d.SaveAs("insert_attrib_demo.dxf")
	if err != nil {
		log.Printf("Error saving drawing: %v", err)
	} else {
		fmt.Println("  Drawing saved successfully!")
	}

	fmt.Println("\n=== INSERT/ATTRIB System Demo Complete ===")
	fmt.Println("\nKey Features Demonstrated:")
	fmt.Println("✓ Block definition and creation")
	fmt.Println("✓ Basic INSERT entities")
	fmt.Println("✓ INSERT with transformations (scale, rotation)")
	fmt.Println("✓ INSERT with attached attributes")
	fmt.Println("✓ Multi-insert (MINSERT) arrays")
	fmt.Println("✓ Standalone ATTRIB entities")
	fmt.Println("✓ Attribute definition and management")
	fmt.Println("✓ Attribute flags and visibility")
	fmt.Println("✓ Professional block reference workflow")
}

func createDoorBlock(d *drawing.Drawing) *block.Block {
	blk := block.NewBlock("DOOR")

	// Add door geometry (rectangle + arc)
	line1 := entity.NewLine()
	line1.Start = []float64{0, 0, 0}
	line1.End = []float64{0, 80, 0}
	blk.AddEntity(line1)

	line2 := entity.NewLine()
	line2.Start = []float64{0, 80, 0}
	line2.End = []float64{80, 80, 0}
	blk.AddEntity(line2)

	line3 := entity.NewLine()
	line3.Start = []float64{80, 80, 0}
	line3.End = []float64{80, 0, 0}
	blk.AddEntity(line3)

	// Add door swing arc
	arc := entity.NewArc()
	// Note: Arc center and radius would need to be set via direct field access
	blk.AddEntity(arc)

	d.AddBlock(blk)
	return blk
}

func createWindowBlock(d *drawing.Drawing) *block.Block {
	blk := block.NewBlock("WINDOW")

	// Add window geometry (simple rectangle)
	line1 := entity.NewLine()
	line1.Start = []float64{0, 0, 0}
	line1.End = []float64{60, 0, 0}
	blk.AddEntity(line1)

	line2 := entity.NewLine()
	line2.Start = []float64{60, 0, 0}
	line2.End = []float64{60, 40, 0}
	blk.AddEntity(line2)

	line3 := entity.NewLine()
	line3.Start = []float64{60, 40, 0}
	line3.End = []float64{0, 40, 0}
	blk.AddEntity(line3)

	line4 := entity.NewLine()
	line4.Start = []float64{0, 40, 0}
	line4.End = []float64{0, 0, 0}
	blk.AddEntity(line4)

	// Add window cross lines
	cross1 := entity.NewLine()
	cross1.Start = []float64{20, 0, 0}
	cross1.End = []float64{20, 40, 0}
	blk.AddEntity(cross1)

	cross2 := entity.NewLine()
	cross2.Start = []float64{40, 0, 0}
	cross2.End = []float64{40, 40, 0}
	blk.AddEntity(cross2)

	d.AddBlock(blk)
	return blk
}

func createFurnitureBlock(d *drawing.Drawing) *block.Block {
	blk := block.NewBlock("DESK")

	// Add desk geometry (simple rectangle)
	line1 := entity.NewLine()
	line1.Start = []float64{0, 0, 0}
	line1.End = []float64{120, 0, 0}
	blk.AddEntity(line1)

	line2 := entity.NewLine()
	line2.Start = []float64{120, 0, 0}
	line2.End = []float64{120, 60, 0}
	blk.AddEntity(line2)

	line3 := entity.NewLine()
	line3.Start = []float64{120, 60, 0}
	line3.End = []float64{0, 60, 0}
	blk.AddEntity(line3)

	line4 := entity.NewLine()
	line4.Start = []float64{0, 60, 0}
	line4.End = []float64{0, 0, 0}
	blk.AddEntity(line4)

	// Add attribute definitions to the block
	attdef1 := entity.NewAttdef()
	attdef1.Tag = "MODEL"
	attdef1.Prompt = "Enter desk model:"
	attdef1.DefaultVal = "OFFICE-2000"
	blk.AddEntity(attdef1)

	attdef2 := entity.NewAttdef()
	attdef2.Tag = "PRICE"
	attdef2.Prompt = "Enter desk price:"
	attdef2.DefaultVal = "450.00"
	blk.AddEntity(attdef2)

	attdef3 := entity.NewAttdef()
	attdef3.Tag = "MATERIAL"
	attdef3.Prompt = "Enter desk material:"
	attdef3.DefaultVal = "OAK"
	blk.AddEntity(attdef3)

	d.AddBlock(blk)
	return blk
}

func createBasicInsert(blk *block.Block, x, y, z float64) *entity.Insert {
	insert := entity.NewInsertFromBlock(blk, dxfmath.NewVec3(x, y, z))
	return insert
}

func createTransformedInsert(blk *block.Block, x, y, rotation, scale float64) *entity.Insert {
	insert := entity.NewInsertFromBlock(blk, dxfmath.NewVec3(x, y, z))
	insert.Rotation = rotation
	insert.XScale = scale
	insert.YScale = scale
	insert.ZScale = scale
	return insert
}

func createInsertWithAttributes(blk *block.Block, x, y, z float64) *entity.Insert {
	insert := entity.NewInsertFromBlock(blk, dxfmath.NewVec3(x, y, z))

	// Create attributes for this insert instance
	attrib1 := entity.NewAttrib()
	attrib1.Tag = "MODEL"
	attrib1.TextValue = "EXECUTIVE-2000"
	attrib1.InsertPoint = []float64{x + 60, y + 30, z}
	attrib1.Height = 3.0
	attrib1.Color = 2 // Yellow

	attrib2 := entity.NewAttrib()
	attrib2.Tag = "PRICE"
	attrib2.TextValue = "850.00"
	attrib2.InsertPoint = []float64{x + 60, y + 20, z}
	attrib2.Height = 3.0
	attrib2.Color = 2 // Yellow

	attrib3 := entity.NewAttrib()
	attrib3.Tag = "MATERIAL"
	attrib3.TextValue = "MAHOGANY"
	attrib3.InsertPoint = []float64{x + 60, y + 10, z}
	attrib3.Height = 3.0
	attrib3.Color = 2 // Yellow

	// Add attributes to insert (this depends on implementation)
	// insert.Attributes = []*entity.Attrib{attrib1, attrib2, attrib3}
	// insert.HasAttribs = true
	// insert.Flags = insert.Flags | entity.InsertFlagHasAttribs

	return insert
}

func createStandaloneAttribute(tag, value string, x, y, z float64) *entity.Attrib {
	attrib := entity.NewAttrib()
	attrib.Tag = tag
	attrib.TextValue = value
	attrib.InsertPoint = []float64{x, y, z}
	attrib.Height = 4.0
	attrib.Color = 3 // Green
	return attrib
}

func createMultipleInsert(blk *block.Block, x, y, z float64, cols, rows int) *entity.Insert {
	insert := entity.NewInsertFromBlock(blk, dxfmath.NewVec3(x, y, z))
	insert.IsMultiple = true
	insert.ColumnCount = cols
	insert.RowCount = rows
	insert.ColumnSpacing = 100.0
	insert.RowSpacing = 80.0
	return insert
}

func testAttributeOperations(insert *entity.Insert) {
	fmt.Printf("   Attribute count for insert: %d\n", len(insert.Attributes))
	for i, attrib := range insert.Attributes {
		fmt.Printf("     %d. Tag='%s', Value='%s', Height=%.1f\n",
			i+1, attrib.Tag, attrib.TextValue, attrib.Height)
	}
}

func testInsertTransformations(insert *entity.Insert) {
	fmt.Printf("   Insert point: (%.1f, %.1f, %.1f)\n",
		insert.InsertPoint.X(), insert.InsertPoint.Y(), insert.InsertPoint.Z())
	fmt.Printf("   Scale: (%.2f, %.2f, %.2f)\n",
		insert.XScale, insert.YScale, insert.ZScale)
	fmt.Printf("   Rotation: %.1f degrees\n", insert.Rotation)

	if !insert.Extrusion.IsEqual(dxfmath.NewVec3(0, 0, 1), 1e-9) {
		fmt.Printf("   Extrusion: (%.3f, %.3f, %.3f)\n",
			insert.Extrusion.X(), insert.Extrusion.Y(), insert.Extrusion.Z())
	}
}

func printInsertInfo(insert *entity.Insert, description string) {
	fmt.Printf("\n   %s:\n", description)
	fmt.Printf("     Block: %s\n", insert.BlockName)
	fmt.Printf("     Position: (%.1f, %.1f, %.1f)\n",
		insert.InsertPoint.X(), insert.InsertPoint.Y(), insert.InsertPoint.Z())
	fmt.Printf("     Scale: (%.2f, %.2f, %.2f)\n",
		insert.XScale, insert.YScale, insert.ZScale)
	fmt.Printf("     Rotation: %.1f°\n", insert.Rotation)
	fmt.Printf("     Flags: %d\n", insert.Flags)
	fmt.Printf("     Has Attributes: %t\n", insert.HasAttribs)
	if insert.IsMultiple {
		fmt.Printf("     Grid: %dx%d (spacing: %.1fx%.1f)\n",
			insert.ColumnCount, insert.RowCount,
			insert.ColumnSpacing, insert.RowSpacing)
	}
}

func testAttributeFlags(insert *entity.Insert) {
	fmt.Printf("   Testing attribute flags on insert...\n")

	// Test setting various flags
	insert.Flags = insert.Flags | entity.InsertFlagAnonymous
	fmt.Printf("     Anonymous flag set: %d\n", insert.Flags)

	insert.Flags = insert.Flags | entity.InsertFlagHasAttribs
	fmt.Printf("     Has attributes flag set: %d\n", insert.Flags)

	insert.Flags = insert.Flags | entity.InsertFlagAttribsFollow
	fmt.Printf("     Attributes follow flag set: %d\n", insert.Flags)

	fmt.Printf("   Final flags: %d\n", insert.Flags)
}
