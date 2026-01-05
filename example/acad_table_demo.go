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
	layer1, _ := d.AddLayer("Tables", color.White, d.Layers["0"].LineType)

	fmt.Println("Creating ACADTABLE example...")

	// Create a simple table
	table := entity.NewAcadTable()
	table.SetTableName("Example Table")
	table.SetTableSize(4, 4) // 4 rows, 4 columns (including header)
	table.SetInsertPoint(10, 10, 0)

	// Use the editor to build the table
	editor := entity.NewAcadTableEditor(table)

	// Set table headers
	editor.SetHeader([]string{"Name", "Age", "City", "Country"})

	// Add data rows
	editor.AddDataRow([]string{"John Doe", "30", "New York", "USA"})
	editor.AddDataRow([]string{"Jane Smith", "25", "London", "UK"})
	editor.AddDataRow([]string{"Bob Johnson", "35", "Tokyo", "Japan"})

	// Set cell formatting
	table.SetCell(1, 2, "New York")
	table.SetCell(2, 3, "London")

	// Set merged cell (spanning 2 columns in first row)
	editor.AddMergedCells(0, []struct {
		StartCol int
		EndCol   int
		Text     string
	}{
		{StartCol: 2, EndCol: 3, Text: "Merged Cell"},
	})

	// Get the final table
	finalTable := editor.GetTable()

	// Assign to layer and add to drawing
	finalTable.SetLayer(layer1)
	d.AddEntity(finalTable)

	// Save drawing
	err = d.SaveAs("acad_table_demo.dxf")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("ACADTABLE example completed!")
	fmt.Println("File saved as: acad_table_demo.dxf")
	fmt.Printf("Table dimensions: %dx%d\n", table.RowCount, table.ColCount)
}
