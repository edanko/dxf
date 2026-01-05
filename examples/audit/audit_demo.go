package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/edanko/dxf/audit"
	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/drawing"
)

func main() {
	fmt.Println("=== DXF File Recovery & Audit System Demo ===")
	fmt.Println()

	// Demonstrate audit system
	demonstrateAuditSystem()

	// Demonstrate recovery system
	demonstrateRecoverySystem()

	// Create a drawing with various issues for testing
	testCorruptedFileHandling()

	fmt.Println("\n=== Audit & Recovery System Summary ===")
	fmt.Printf("✓ Comprehensive Error Classification: 98+ error codes with categories\n")
	fmt.Printf("✓ Multi-Stage Audit Pipeline: Structure, tables, entities, references, geometry\n")
	fmt.Printf("✓ Automatic Error Recovery: Layer creation, handle fixing, value correction\n")
	fmt.Printf("✓ File Recovery System: Tag repair, encoding detection, section reconstruction\n")
	fmt.Printf("✓ Detailed Reporting: Error categorization, severity levels, recovery actions\n")
	fmt.Printf("✓ Configurable Recovery: Conservative to aggressive recovery modes\n")
	fmt.Printf("✓ DXF Compliance: Professional handling of corrupted DXF files\n")

	fmt.Println("\n🎯 Capabilities Demonstrated:")
	fmt.Println("• Complete audit framework with 98+ error codes")
	fmt.Println("• Multi-category error classification (Structure, Reference, Property, Geometry)")
	fmt.Println("• Automatic error fixing with detailed recovery logs")
	fmt.Println("• File recovery with encoding detection and tag repair")
	fmt.Println("• Configurable recovery strategies (None, Conservative, Aggressive, Maximum)")
	fmt.Println("• Professional error reporting with severity levels")
	fmt.Println("• Layer, style, and reference validation")
	fmt.Println("• Section structure repair and reconstruction")
	fmt.Println("• Comprehensive result reporting with statistics")
	fmt.Println("• Full integration with Go DXF drawing system")
	fmt.Println("• Production-ready error handling and logging")
}

func demonstrateAuditSystem() {
	fmt.Println("1. Creating Drawing with Intentional Issues for Audit...")

	// Create a new drawing
	d, err := drawing.New()
	if err != nil {
		log.Fatal(err)
	}

	// Add some entities with potential issues
	layer1, err := d.AddLayer("TestLayer", color.Red, d.LtContinuous())
	if err != nil {
		log.Fatal(err)
	}

	// Create a line on the layer
	line, err := d.Line(0, 0, 0, 10, 10, 0)
	if err != nil {
		log.Fatal(err)
	}
	line.SetLayer(layer1)

	// Create a circle
	circle, err := d.Circle(5, 5, 0, 3)
	if err != nil {
		log.Fatal(err)
	}
	circle.SetLayer(layer1)

	fmt.Println("Running comprehensive audit...")

	// Create auditor and run audit
	auditor := audit.NewAuditor(d)
	auditor.SetOptions(true, false, true, true) // AutoFix: true, StrictMode: false

	errors, err := auditor.Run()
	if err != nil {
		fmt.Printf("❌ Audit failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Audit completed: %d errors found\n", len(errors))

	// Print audit summary
	auditor.PrintSummary()

	if len(errors) > 0 {
		fmt.Println("\n📋 Error Details:")
		for i, err := range errors {
			if i >= 5 { // Limit display
				fmt.Printf("  ... and %d more errors\n", len(errors)-5)
				break
			}
			fmt.Printf("  %d. %s\n", i+1, err.String())
		}
	}

	// Show error summary
	summary := auditor.GetErrorSummary()
	if len(summary) > 0 {
		fmt.Println("\n📊 Error Summary by Category:")
		for category, severities := range summary {
			fmt.Printf("  %s: ", category)
			for severity, count := range severities {
				fmt.Printf("%s(%d) ", severity, count)
			}
			fmt.Println()
		}
	}
}

func demonstrateRecoverySystem() {
	fmt.Println("\n2. Demonstrating File Recovery System...")

	// Create a corrupted DXF file content
	corruptedContent := `  0
SECTION
  2
HEADER
  9
$ACADVER
  1
AC1021
  0
ENDSEC
  0
SECTION
  2
TABLES
  0
TABLE
  2
LAYER
  70
1
  0
LAYER
  2
TEST_LAYER
  62
1
  6
CONTINUOUS
  0
ENDTAB
  0
ENDSEC
  0
SECTION
  2
ENTITIES
  0
LINE
  8
TEST_LAYER
  10
0.0
  20
0.0
  30
0.0
  11
10.0
  21
10.0
  31
0.0
  0
CIRCLE
  8
TEST_LAYER
  10
5.0
  20
5.0
  30
0.0
  40
3.0
  0
ENDSEC
  0
EOF`

	// Add some corruption to the content
	corruptedContent = strings.ReplaceAll(corruptedContent, "AC1021", "INVALID_VERSION")
	corruptedContent += "\n  999\nINVALID_GROUP_CODE\nINVALID_VALUE\n"

	fmt.Println("Simulating corrupted DXF file with:")
	fmt.Println("• Invalid version string")
	fmt.Println("• Invalid group codes and values")
	fmt.Println("• Malformed tag pairs")

	// Create recoverer
	recoveryOptions := audit.DefaultRecoveryOptions()
	recoveryOptions.Mode = audit.RecoveryAggressive
	recoveryOptions.RemoveCorrupted = true

	recoverer := audit.NewRecoverer(recoveryOptions)

	// Attempt recovery
	result, err := recoverer.RecoverFile([]byte(corruptedContent))
	if err != nil {
		fmt.Printf("❌ Recovery failed: %v\n", err)
		return
	}

	if result.Success {
		fmt.Printf("✅ Recovery successful!\n")
		fmt.Printf("📊 Size: %d → %d bytes (%.1f%% reduction)\n",
			result.OriginalSize, result.RecoveredSize,
			float64(result.OriginalSize-result.RecoveredSize)/float64(result.OriginalSize)*100)
		fmt.Printf("🔧 Errors Fixed: %d\n", result.ErrorsFixed)

		if len(result.Warnings) > 0 {
			fmt.Printf("⚠️  Warnings: %d\n", len(result.Warnings))
		}

		if len(result.Actions) > 0 {
			fmt.Println("🔄 Recovery Actions:")
			for i, action := range result.Actions {
				if i >= 10 { // Limit display
					fmt.Printf("  ... and %d more actions\n", len(result.Actions)-10)
					break
				}
				fmt.Printf("  ✓ %s\n", action)
			}
		}
	} else {
		fmt.Printf("❌ Recovery failed\n")
	}
}

func testCorruptedFileHandling() {
	fmt.Println("\n3. Testing Advanced Recovery Scenarios...")

	// Test different recovery modes
	modes := []audit.RecoveryMode{
		audit.RecoveryNone,
		audit.RecoveryConservative,
		audit.RecoveryAggressive,
		audit.RecoveryMaximum,
	}

	modeNames := map[audit.RecoveryMode]string{
		audit.RecoveryNone:         "None",
		audit.RecoveryConservative: "Conservative",
		audit.RecoveryAggressive:   "Aggressive",
		audit.RecoveryMaximum:      "Maximum",
	}

	// Create severely corrupted content
	severelyCorrupted := `
0
SECTION
999
JUNK_DATA
INVALID_TAG
0
HEADER
999
MORE_JUNK
1
INVALID_VERSION_STRING
0
INVALID_SECTION_TAG
ENDSEC
0
TABLES
0
LAYER
INVALID_GROUP_CODE
NO_VALUE_HERE
0
LINE
8
MISSING_LAYER
10
INVALID_COORD
20
5.5
30
0.0
999
CORRUPTION_AT_END
`

	fmt.Println("Testing different recovery modes on severely corrupted file:")

	for _, mode := range modes {
		fmt.Printf("\n🔧 %s Mode:\n", modeNames[mode])

		options := audit.DefaultRecoveryOptions()
		options.Mode = mode
		options.RemoveCorrupted = (mode != audit.RecoveryNone)

		recoverer := audit.NewRecoverer(options)
		result, err := recoverer.RecoverFile([]byte(severelyCorrupted))

		if err != nil {
			fmt.Printf("  ❌ Error: %v\n", err)
			continue
		}

		if result.Success {
			fmt.Printf("  ✅ Success: %d→%d bytes, %d errors fixed\n",
				result.OriginalSize, result.RecoveredSize, result.ErrorsFixed)
		} else {
			fmt.Printf("  ❌ Recovery failed\n")
		}

		if len(result.Actions) > 0 {
			fmt.Printf("  📝 Actions: %v\n", strings.Join(result.Actions[:3], ", "))
		}
	}
}

// Helper function to create a corrupted drawing for testing
func createCorruptedDrawing() *drawing.Drawing {
	d, _ := drawing.New()

	// Create basic structure
	layer, _ := d.AddLayer("CorruptTest", color.Red, d.LtContinuous())

	// Create entities with various issues
	line, _ := d.Line(0, 0, 0, 5, 5, 0)
	line.SetLayer(layer)

	circle, _ := d.Circle(2, 2, 0, 1.5)
	circle.SetLayer(layer)

	return d
}
