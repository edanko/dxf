package dxf

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/edanko/dxf/entity"
)

func TestFromFileAllExamples(t *testing.T) {
	testDir := filepath.Join("testdata")
	files, err := os.ReadDir(testDir)
	if err != nil {
		t.Fatalf("Failed to read test directory: %v", err)
	}

	var successCount, failCount int
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filename := file.Name()
		if filepath.Ext(filename) != ".dxf" {
			continue
		}

		t.Run(filename, func(t *testing.T) {
			testFile := filepath.Join(testDir, filename)
			d, err := FromFile(testFile)
			if err != nil {
				t.Errorf("Failed to load %s: %v", filename, err)
				failCount++
				return
			}
			if d == nil {
				t.Error("Drawing is nil after loading")
				failCount++
				return
			}

			entityCount := len(d.Entities())
			t.Logf("Loaded %s: %d entities", filename, entityCount)
			successCount++
		})
	}

	t.Logf("Successfully loaded %d/%d DXF files", successCount, successCount+failCount)
}

func TestFromFileSpecificExamples(t *testing.T) {
	tests := []struct {
		filename string
	}{
		{"3dface.dxf"},
		{"circle_radius_le_0.dxf"},
		{"arc.dxf"},
		{"point.dxf"},
	}

	for _, tc := range tests {
		t.Run(tc.filename, func(t *testing.T) {
			testFile := filepath.Join("testdata", tc.filename)
			d, err := FromFile(testFile)
			if err != nil {
				t.Fatalf("Failed to load %s: %v", tc.filename, err)
			}
			if d == nil {
				t.Error("Drawing should not be nil")
			}
		})
	}
}

func TestDrawingAddPoint(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	_, err = d.Point(0, 0, 0)
	if err != nil {
		t.Errorf("Point failed: %v", err)
	}
	_, err = d.Point(10, 20, 0)
	if err != nil {
		t.Errorf("Point failed: %v", err)
	}

	if len(d.Entities()) != 2 {
		t.Errorf("Expected 2 points, got %d", len(d.Entities()))
	}
}

func TestDrawingAddLine(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	_, err = d.Line(0, 0, 0, 100, 100, 0)
	if err != nil {
		t.Errorf("Line failed: %v", err)
	}

	if len(d.Entities()) != 1 {
		t.Errorf("Expected 1 line, got %d", len(d.Entities()))
	}
}

func TestDrawingAddCircle(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	_, err = d.Circle(50, 50, 0, 25)
	if err != nil {
		t.Errorf("Circle failed: %v", err)
	}

	if len(d.Entities()) != 1 {
		t.Errorf("Expected 1 circle, got %d", len(d.Entities()))
	}
}

func TestDrawingAddArc(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	_, err = d.Arc(0, 0, 0, 100, 0, 90)
	if err != nil {
		t.Errorf("Arc failed: %v", err)
	}

	if len(d.Entities()) != 1 {
		t.Errorf("Expected 1 arc, got %d", len(d.Entities()))
	}
}

func TestDrawingAddPolyline(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	_, err = d.Polyline(false, []float64{0, 0, 0}, []float64{10, 0, 0}, []float64{10, 10, 0}, []float64{0, 10, 0})
	if err != nil {
		t.Errorf("Polyline failed: %v", err)
	}

	if len(d.Entities()) != 1 {
		t.Errorf("Expected 1 polyline, got %d", len(d.Entities()))
	}
}

func TestDrawingAddLWPolyline(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	_, err = d.LwPolyline(false, []float64{0, 0}, []float64{10, 0}, []float64{10, 10}, []float64{0, 10})
	if err != nil {
		t.Errorf("LwPolyline failed: %v", err)
	}

	if len(d.Entities()) != 1 {
		t.Errorf("Expected 1 lwpolyline, got %d", len(d.Entities()))
	}
}

func TestDrawingAdd3DFace(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	_, err = d.ThreeDFace([][]float64{
		{0, 0, 0},
		{10, 0, 0},
		{10, 10, 10},
		{0, 10, 10},
	})
	if err != nil {
		t.Errorf("ThreeDFace failed: %v", err)
	}

	if len(d.Entities()) != 1 {
		t.Errorf("Expected 1 3dface, got %d", len(d.Entities()))
	}
}

func TestDrawingAddText(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	_, err = d.Text("Hello", 0, 0, 0, 2.5)
	if err != nil {
		t.Errorf("Text failed: %v", err)
	}

	if len(d.Entities()) != 1 {
		t.Errorf("Expected 1 text, got %d", len(d.Entities()))
	}
}

func TestDrawingAddMText(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	_, err = d.MText("Multi-line\ntext", 0, 0, 0, 2.5)
	if err != nil {
		t.Errorf("MText failed: %v", err)
	}

	if len(d.Entities()) != 1 {
		t.Errorf("Expected 1 mtext, got %d", len(d.Entities()))
	}
}

func TestDrawingAddBlock(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	_, err = d.NewBlock("TEST_BLOCK", "Test block description")
	if err != nil {
		t.Errorf("NewBlock failed: %v", err)
	}

	blocks := d.GetBlocks()
	if len(blocks) < 1 {
		t.Errorf("Expected at least 1 block, got %d", len(blocks))
	}
}

func TestDrawingGetEntity(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	p, err := d.Point(5, 5, 0)
	if err != nil {
		t.Fatalf("Point failed: %v", err)
	}

	handle := p.Handle()
	entity, ok := d.GetEntity(handle)
	if !ok {
		t.Error("GetEntity should return the entity by handle")
	}
	if entity == nil {
		t.Error("Entity should not be nil")
	}
}

func TestDrawingGetEntitiesByType(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	d.Point(0, 0, 0)
	d.Point(10, 10, 0)
	d.Line(0, 0, 0, 100, 100, 0)

	points := d.GetEntitiesByType("POINT")
	if len(points) != 2 {
		t.Errorf("Expected 2 points, got %d", len(points))
	}

	lines := d.GetEntitiesByType("LINE")
	if len(lines) != 1 {
		t.Errorf("Expected 1 line, got %d", len(lines))
	}
}

func TestDrawingGroups(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	d.Point(0, 0, 0)
	d.Point(10, 10, 0)

	_, err = d.Group("TestGroup", "A test group")
	if err != nil {
		t.Errorf("Group failed: %v", err)
	}

	if d.Groups["TestGroup"] == nil {
		t.Error("Group 'TestGroup' should exist")
	}
}

func TestDrawingLineType(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	lt, err := d.LineType("Continuous")
	if err != nil {
		t.Errorf("LineType failed: %v", err)
	}
	if lt == nil {
		t.Error("LineType should not return nil")
	}
}

func TestDrawingAddLayer(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	lt, _ := d.LineType("Continuous")
	_, err = d.AddLayer("MyLayer", 5, lt)
	if err != nil {
		t.Errorf("AddLayer failed: %v", err)
	}

	if d.Layers["MyLayer"] == nil {
		t.Error("Layer 'MyLayer' should exist")
	}
}

func TestDrawingChangeLayer(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	lt, _ := d.LineType("Continuous")
	d.AddLayer("TestLayer", 3, lt)
	err = d.ChangeLayer("TestLayer")
	if err != nil {
		t.Errorf("ChangeLayer failed: %v", err)
	}

	if d.CurrentLayer.Name() != "TestLayer" {
		t.Errorf("Expected current layer 'TestLayer', got '%s'", d.CurrentLayer.Name())
	}
}

func TestDrawingGetLayerByName(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	layer, err := d.GetLayerByName("0")
	if err != nil {
		t.Errorf("GetLayerByName failed: %v", err)
	}
	if layer == nil {
		t.Error("Layer '0' should exist")
	}
}

func TestDrawingEntitiesCount(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	d.Point(0, 0, 0)
	d.Point(10, 10, 0)
	d.Line(0, 0, 0, 100, 100, 0)

	count := len(d.Entities())
	if count != 3 {
		t.Errorf("Expected 3 entities, got %d", count)
	}
}

func TestDrawingModelSpace(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	d.Point(0, 0, 0)
	d.Line(0, 0, 0, 100, 100, 0)

	ms := d.ModelSpace()
	if len(ms) < 2 {
		t.Errorf("Expected at least 2 entities in model space, got %d", len(ms))
	}
}

func TestDrawingPaperSpace(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	ps := d.PaperSpace()
	t.Logf("Paper space has %d entities", len(ps))
}

func TestDrawingHeader(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	header := d.Header()
	if header == nil {
		t.Error("Header should not be nil")
	}
}

func TestDrawingTables(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	tables := d.Tables()
	if tables == nil {
		t.Error("Tables should not be nil")
	}
}

func TestDrawingLineTypes(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	ltByLayer := d.LtByLayer()
	if ltByLayer == nil {
		t.Error("LtByLayer should not be nil")
	}

	ltByBlock := d.LtByBlock()
	if ltByBlock == nil {
		t.Error("LtByBlock should not be nil")
	}

	ltContinuous := d.LtContinuous()
	if ltContinuous == nil {
		t.Error("LtContinuous should not be nil")
	}

	ltHidden := d.LtHidden()
	if ltHidden == nil {
		t.Error("LtHidden should not be nil")
	}

	ltDashDot := d.LtDashDot()
	if ltDashDot == nil {
		t.Error("LtDashDot should not be nil")
	}
	_ = ltByLayer
	_ = ltByBlock
	_ = ltContinuous
	_ = ltHidden
	_ = ltDashDot
}

func TestDrawingQuery(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	d.Point(0, 0, 0)
	d.Point(10, 10, 0)
	d.Line(0, 0, 0, 100, 100, 0)

	query := d.Query()
	if query == nil {
		t.Error("Query should not be nil")
	}
}

func TestDrawingSelectEntities(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	d.Point(0, 0, 0)
	d.Point(10, 10, 0)
	d.Line(0, 0, 0, 100, 100, 0)

	query := d.SelectEntities("POINT")
	if query == nil {
		t.Error("SelectEntities should return a query")
	}
}

func TestDrawingGetEntityCount(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	d.Point(0, 0, 0)
	d.Point(10, 10, 0)

	count := d.GetEntityCount()
	if count != 2 {
		t.Errorf("Expected 2 entities, got %d", count)
	}
}

func TestDrawingGetEntityCountByType(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	d.Point(0, 0, 0)
	d.Point(10, 10, 0)
	d.Line(0, 0, 0, 100, 100, 0)

	pointCount := d.GetEntityCountByType("POINT")
	if pointCount != 2 {
		t.Errorf("Expected 2 points, got %d", pointCount)
	}

	lineCount := d.GetEntityCountByType("LINE")
	if lineCount != 1 {
		t.Errorf("Expected 1 line, got %d", lineCount)
	}
}

func TestDrawingBlockNames(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	d.Block("BLOCK1")
	d.Block("BLOCK2")

	names := d.GetBlockNames()
	if len(names) < 2 {
		t.Errorf("Expected at least 2 block names, got %d", len(names))
	}
}

func TestDrawingLayoutNames(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	names := d.LayoutNames()
	t.Logf("Layout names: %v", names)
}

func TestDrawingCurrentLayout(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	layout := d.CurrentLayout()
	t.Logf("Current layout: %s", layout)
}

func TestDrawingFormatter(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	formatter := d.Formatter()
	if formatter == nil {
		t.Error("Formatter should not be nil")
	}
}

func TestDrawingAddLineType(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	lt, err := d.AddLineType("DASHED", "Dashed line", 0.5, 0.25, -0.25)
	if err != nil {
		t.Errorf("AddLineType failed: %v", err)
	}
	if lt == nil {
		t.Error("LineType should not be nil")
	}
}

func TestDrawingDeleteEntity(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	p1, err := d.Point(0, 0, 0)
	if err != nil {
		t.Fatalf("Point failed: %v", err)
	}
	_, err = d.Point(10, 10, 0)
	if err != nil {
		t.Fatalf("Point failed: %v", err)
	}

	if len(d.Entities()) != 2 {
		t.Errorf("Expected 2 entities, got %d", len(d.Entities()))
	}

	err = d.DeleteEntity(p1.Handle())
	if err != nil {
		t.Errorf("DeleteEntity failed: %v", err)
	}
}

func TestDrawingFilterEntities(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	d.Point(0, 0, 0)
	d.Point(10, 10, 0)
	d.Line(0, 0, 0, 100, 100, 0)

	query := d.FilterEntities(func(e entity.Entity) bool {
		return true
	})
	if query == nil {
		t.Error("FilterEntities should return a query")
	}
}

func TestDrawingGroupEntitiesBy(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	d.Point(0, 0, 0)
	d.Point(10, 10, 0)
	d.Line(0, 0, 0, 100, 100, 0)

	groups := d.GroupEntitiesBy(func(e entity.Entity) string {
		return "all"
	})
	if groups == nil {
		t.Error("GroupEntitiesBy should return a map")
	}
}
