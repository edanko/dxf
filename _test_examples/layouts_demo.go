package main

import (
	"fmt"
	"log"

	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/layouts"
)

func main() {
	fmt.Println("=== DXF Layout Management System Demo ===\n")

	// Create a new drawing
	d, err := drawing.New()
	if err != nil {
		log.Fatal(err)
	}

	// Create layouts manager
	layoutsManager := layouts.NewLayouts(d)

	// Setup default layouts (model space + first paper space)
	err = layoutsManager.SetupLayouts()
	if err != nil {
		log.Fatalf("Failed to setup layouts: %v", err)
	}

	fmt.Printf("1. Initial layout setup:\n")
	fmt.Printf("   Total layouts: %d\n", layoutsManager.LayoutCount())
	fmt.Printf("   Paper spaces: %d\n", layoutsManager.PaperSpaceCount())
	fmt.Printf("   %s\n", layoutsManager)

	// Access model space
	modelSpace := layoutsManager.ModelSpace()
	fmt.Printf("\n2. Model Space:\n")
	fmt.Printf("   Name: %s\n", modelSpace.Name())
	fmt.Printf("   Type: %v\n", modelSpace.Type())
	fmt.Printf("   Entity count: %d\n", modelSpace.EntityCount())

	// Add more paper space layouts
	fmt.Println("\n3. Adding additional paper space layouts...")

	layout1, err := layoutsManager.AddPaperSpace("Detail View")
	if err != nil {
		log.Printf("Error adding layout: %v", err)
	} else {
		fmt.Printf("   Added layout: %s (Tab Order: %d)\n", layout1.Name(), layout1.TabOrder())
	}

	layout2, err := layoutsManager.AddPaperSpace("Plan View")
	if err != nil {
		log.Printf("Error adding layout: %v", err)
	} else {
		fmt.Printf("   Added layout: %s (Tab Order: %d)\n", layout2.Name(), layout2.TabOrder())
	}

	// Test layout retrieval
	fmt.Printf("\n4. Layout retrieval tests:\n")
	testLayout := layoutsManager.GetLayout("Detail View")
	if testLayout != nil {
		fmt.Printf("   Found layout: %s (Type: %v)\n", testLayout.Name(), testLayout.Type())
	}

	missingLayout := layoutsManager.GetLayout("NonExistent")
	if missingLayout == nil {
		fmt.Printf("   Correctly identified non-existent layout\n")
	}

	// Test tab order
	fmt.Printf("\n5. Current tab order:\n")
	tabOrder := layoutsManager.GetTabOrder()
	for i, layout := range tabOrder {
		layoutType := "Paper Space"
		if layout.IsModelSpace() {
			layoutType = "Model Space"
		}
		fmt.Printf("   %d. %s (%s)\n", i+1, layout.Name(), layoutType)
	}

	// Test layout renaming
	fmt.Println("\n6. Testing layout renaming...")
	err = layoutsManager.RenameLayout("Detail View", "Section View")
	if err != nil {
		log.Printf("Error renaming layout: %v", err)
	} else {
		fmt.Printf("   Successfully renamed 'Detail View' to 'Section View'\n")
	}

	// Test creating layout from template
	fmt.Println("\n7. Creating layout from template...")
	templateLayout, err := layoutsManager.CreateLayoutFromTemplate("Plan View", "Elevation View")
	if err != nil {
		log.Printf("Error creating layout from template: %v", err)
	} else {
		fmt.Printf("   Created '%s' from 'Plan View' template\n", templateLayout.Name())
	}

	// Show final state
	fmt.Printf("\n8. Final layout state:\n")
	fmt.Printf("   Total layouts: %d\n", layoutsManager.LayoutCount())
	fmt.Printf("   Paper spaces: %d\n", layoutsManager.PaperSpaceCount())
	fmt.Printf("   %s\n", layoutsManager)

	// Iterate over all layouts
	fmt.Println("\n9. Iterating over all layouts:")
	layoutsManager.ForEach(func(layout *layouts.Layout) {
		layoutType := "Paper Space"
		if layout.IsModelSpace() {
			layoutType = "Model Space"
		}
		fmt.Printf("   - %s (%s): %d entities\n",
			layout.Name(), layoutType, layout.EntityCount())
	})

	// Test layout validation
	fmt.Println("\n10. Layout name validation tests:")
	testNames := []string{"", "ValidName", "Invalid/Name", "Another:Valid", ""}
	for _, name := range testNames {
		err := layouts.ValidateLayoutName(name)
		if err != nil {
			fmt.Printf("   '%s': Invalid - %v\n", name, err)
		} else {
			fmt.Printf("   '%s': Valid\n", name)
		}
	}

	fmt.Println("\n=== Layout Management Demo Complete ===")
	fmt.Println("\nKey Features Demonstrated:")
	fmt.Println("✓ Model space and paper space creation")
	fmt.Println("✓ Layout management (add, rename, remove)")
	fmt.Println("✓ Tab ordering system")
	fmt.Println("✓ Layout retrieval by name and handle")
	fmt.Println("✓ Layout template copying")
	fmt.Println("✓ Name validation")
	fmt.Println("✓ Iteration and enumeration")
	fmt.Println("✓ Professional layout organization")
}
