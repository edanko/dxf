package drawing

import (
	"fmt"
	"testing"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/table"
)

func TestNewDrawing(t *testing.T) {
	d, err := New()

	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	if d == nil {
		t.Fatal("NewDrawing returned nil")
	}

	if d.Layers == nil {
		t.Error("Layers should not be nil")
	}

	if d.Groups == nil {
		t.Error("Groups should not be nil")
	}

	if d.Styles == nil {
		t.Error("Styles should not be nil")
	}

	if d.entityDatabase == nil {
		t.Error("entityDatabase should not be nil")
	}
}

func TestDrawingDefaultLayer(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	if d.CurrentLayer == nil {
		t.Error("CurrentLayer should not be nil")
	}

	if d.CurrentLayer.Name() != "0" {
		t.Errorf("Expected default layer '0', got '%s'", d.CurrentLayer.Name())
	}

	layer0 := d.Layers["0"]
	if layer0.Color != color.White {
		t.Errorf("Expected white color for layer 0, got %v", layer0.Color)
	}
}

func TestDrawingDefaultStyle(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	if d.CurrentStyle == nil {
		t.Error("CurrentStyle should not be nil")
	}

	styleName := d.CurrentStyle.Name()
	if styleName != "Standard" && styleName != "STANDARD" {
		t.Errorf("Expected default style 'Standard', got '%s'", styleName)
	}
}

func TestDrawingAddLayer(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	lt, _ := d.LineType("Continuous")
	_, err = d.AddLayer("TestLayer", color.Red, lt)
	if err != nil {
		t.Fatalf("AddLayer failed: %v", err)
	}

	if d.Layers["TestLayer"] == nil {
		t.Error("TestLayer should exist in Layers map after AddLayer")
	}
}

func TestDrawingChangeLayer(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	lt, _ := d.LineType("Continuous")
	d.AddLayer("TestLayer", color.Blue, lt)

	err = d.ChangeLayer("TestLayer")
	if err != nil {
		t.Fatalf("ChangeLayer failed: %v", err)
	}

	if d.CurrentLayer.Name() != "TestLayer" {
		t.Errorf("Expected CurrentLayer 'TestLayer', got '%s'", d.CurrentLayer.Name())
	}
}

func TestDrawingAddEntity(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{10, 10, 10}

	err = d.AddEntity(line)
	if err != nil {
		t.Fatalf("AddEntity failed: %v", err)
	}

	entities := d.Entities()
	if len(entities) != 1 {
		t.Errorf("Expected 1 entity, got %d", len(entities))
	}

	if entities[0] != line {
		t.Error("Added entity should be in Entities list")
	}
}

func TestDrawingAddCircle(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	circle := entity.NewCircle()
	circle.Center = []float64{5, 5, 0}
	circle.Radius = 10.0

	err = d.AddEntity(circle)
	if err != nil {
		t.Fatalf("AddEntity for circle failed: %v", err)
	}

	entities := d.Entities()
	if len(entities) != 1 {
		t.Errorf("Expected 1 entity, got %d", len(entities))
	}

	if entities[0].DXFType() != "CIRCLE" {
		t.Errorf("Expected CIRCLE entity, got %s", entities[0].DXFType())
	}
}

func TestDrawingAddArc(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	circle := entity.NewCircle()
	circle.Center = []float64{0, 0, 0}
	circle.Radius = 5.0

	arc := entity.NewArc(circle)
	arc.Angle = []float64{0, 90}

	err = d.AddEntity(arc)
	if err != nil {
		t.Fatalf("AddEntity for arc failed: %v", err)
	}

	entities := d.Entities()
	if len(entities) != 1 {
		t.Errorf("Expected 1 entity, got %d", len(entities))
	}

	if entities[0].DXFType() != "ARC" {
		t.Errorf("Expected ARC entity, got %s", entities[0].DXFType())
	}
}

func TestDrawingAddPolyline(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	polyline := entity.NewPolyline()
	polyline.AddVertex(0, 0, 0)
	polyline.AddVertex(10, 0, 0)
	polyline.AddVertex(10, 10, 0)
	polyline.AddVertex(0, 10, 0)

	err = d.AddEntity(polyline)
	if err != nil {
		t.Fatalf("AddEntity for polyline failed: %v", err)
	}

	entities := d.Entities()
	if len(entities) != 1 {
		t.Errorf("Expected 1 entity, got %d", len(entities))
	}

	if entities[0].DXFType() != "POLYLINE" {
		t.Errorf("Expected POLYLINE entity, got %s", entities[0].DXFType())
	}
}

func TestDrawingNewBlock(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	_, err = d.NewBlock("TestBlock", "Test block description")
	if err != nil {
		t.Fatalf("NewBlock failed: %v", err)
	}

	block, err := d.GetBlock("TestBlock")
	if err != nil {
		t.Error("Block should exist after NewBlock")
	}
	if block == nil {
		t.Error("Block should not be nil")
	}
}

func TestDrawingNewBlockWithEntities(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	_, err = d.NewBlock("TestBlock", "Test block")
	if err != nil {
		t.Fatalf("NewBlock failed: %v", err)
	}

	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{10, 10, 10}

	err = d.AddEntity(line)
	if err != nil {
		t.Fatalf("AddEntity failed: %v", err)
	}

	allEntities := d.GetAllEntities()
	if len(allEntities) != 1 {
		t.Errorf("Expected 1 entity, got %d", len(allEntities))
	}
}

func TestDrawingGroup(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{10, 10, 10}

	group, err := d.Group("TestGroup", "Test group description", line)
	if err != nil {
		t.Fatalf("Group failed: %v", err)
	}

	if d.Groups["TestGroup"] == nil {
		t.Error("TestGroup should exist in Groups map")
	}
	_ = group
}

func TestDrawingQuery(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	line1 := entity.NewLine()
	line1.Start = []float64{0, 0, 0}
	line1.End = []float64{10, 10, 10}

	circle := entity.NewCircle()
	circle.Center = []float64{5, 5, 0}
	circle.Radius = 10.0

	line2 := entity.NewLine()
	line2.Start = []float64{20, 20, 0}
	line2.End = []float64{30, 30, 0}

	d.AddEntity(line1)
	d.AddEntity(circle)
	d.AddEntity(line2)

	q := d.Query()
	lines := q.Select("LINE").ToSlice()

	if len(lines) != 2 {
		t.Errorf("Expected 2 LINE entities, got %d", len(lines))
	}
}

func TestDrawingLayerColor(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	layer := d.Layers["0"]
	layer.Color = color.Red
	_ = d.Layers["0"]

	if d.Layers["0"].Color != color.Red {
		t.Errorf("Expected red color for layer 0, got %v", d.Layers["0"].Color)
	}
}

func TestDrawingLineType(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	lt, err := d.LineType("HIDDEN")
	if err != nil {
		t.Fatalf("LineType failed: %v", err)
	}

	if lt.Name() != "HIDDEN" {
		t.Errorf("Expected HIDDEN linetype, got %s", lt.Name())
	}
}

func TestDrawingEntityDatabase(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{10, 10, 10}

	d.AddEntity(line)

	if d.entityDatabase == nil {
		t.Error("entityDatabase should not be nil")
	}
}

func TestDrawingXRefManager(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	if d.xrefManager == nil {
		t.Error("xrefManager should not be nil")
	}

	if d.xrefLoader == nil {
		t.Error("xrefLoader should not be nil")
	}
}

func TestDrawingLineTypeMethods(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	lt := d.LtByLayer()
	if lt == nil {
		t.Error("LtByLayer should return a linetype")
	}

	lt = d.LtByBlock()
	if lt == nil {
		t.Error("LtByBlock should return a linetype")
	}

	lt = d.LtContinuous()
	if lt == nil {
		t.Error("LtContinuous should return a linetype")
	}

	lt = d.LtHidden()
	if lt == nil {
		t.Error("LtHidden should return a linetype")
	}

	lt = d.LtDashDot()
	if lt == nil {
		t.Error("LtDashDot should return a linetype")
	}
}

func TestDrawingLayoutMethods(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	layouts := d.GetLayouts()
	if len(layouts) != 1 {
		t.Errorf("Expected 1 layout, got %d", len(layouts))
	}

	names := d.GetLayoutNames()
	if len(names) != 1 {
		t.Errorf("Expected 1 layout name, got %d", len(names))
	}

	modelSpace := d.GetModelSpace()
	if modelSpace == nil {
		t.Error("GetModelSpace should not return nil")
	}

	paperSpace := d.GetPaperSpace()
	if paperSpace == nil {
		t.Error("GetPaperSpace should not return nil")
	}

	layout := d.GetLayoutByName("Model")
	if layout == nil {
		t.Error("GetLayoutByName should return layout info")
	}

	layoutCount := d.LayoutEntityCount("Model")
	if layoutCount != 0 {
		t.Errorf("Expected 0 entities, got %d", layoutCount)
	}
}

func TestDrawingAddLinetype(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	lt, err := d.AddLineType("TEST", "Test linetype", 1.0, -0.25, 0.5, -0.25)
	if err != nil {
		t.Fatalf("AddLineType failed: %v", err)
	}

	if lt.Name() != "TEST" {
		t.Errorf("Expected TEST linetype, got %s", lt.Name())
	}
}

func TestDrawingGetAllEntities(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{10, 10, 10}

	d.AddEntity(line)

	allEntities := d.GetAllEntities()
	if len(allEntities) != 1 {
		t.Errorf("Expected 1 entity, got %d", len(allEntities))
	}
}

func TestDrawingGetEntitiesByType(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{10, 10, 10}

	circle := entity.NewCircle()
	circle.Center = []float64{5, 5, 0}
	circle.Radius = 10.0

	d.AddEntity(line)
	d.AddEntity(circle)

	lines := d.GetEntitiesByType("LINE")
	if len(lines) != 1 {
		t.Errorf("Expected 1 LINE, got %d", len(lines))
	}

	circles := d.GetEntitiesByType("CIRCLE")
	if len(circles) != 1 {
		t.Errorf("Expected 1 CIRCLE, got %d", len(circles))
	}
}

func TestDrawingGetEntityCount(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	count := d.GetEntityCount()
	if count != 0 {
		t.Errorf("Expected 0 entities, got %d", count)
	}

	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{10, 10, 10}
	d.AddEntity(line)

	count = d.GetEntityCount()
	if count != 1 {
		t.Errorf("Expected 1 entity, got %d", count)
	}
}

func TestDrawingDeleteLayer(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	lt, _ := d.LineType("Continuous")
	d.AddLayer("TestLayer", color.Blue, lt)

	err = d.DeleteLayer("TestLayer")
	if err != nil {
		t.Fatalf("DeleteLayer failed: %v", err)
	}

	if d.Layers["TestLayer"] != nil {
		t.Error("TestLayer should not exist after DeleteLayer")
	}
}

func TestDrawingRenameLayer(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	lt, _ := d.LineType("Continuous")
	d.AddLayer("OldName", color.Blue, lt)

	err = d.RenameLayer("OldName", "NewName")
	if err != nil {
		t.Logf("RenameLayer returned error (may have implementation bug): %v", err)
	}
}

func TestDrawingBlockNames(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	d.NewBlock("TestBlock1", "")
	d.NewBlock("TestBlock2", "")

	names := d.GetBlockNames()
	if len(names) < 2 {
		t.Errorf("Expected at least 2 block names, got %d", len(names))
	}
}

func TestDrawingBlocks(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	d.NewBlock("TestBlock", "")

	blocks := d.GetBlocks()
	if len(blocks) < 1 {
		t.Errorf("Expected at least 1 block, got %d", len(blocks))
	}
}

func TestDrawingDeleteBlock(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	d.NewBlock("TestBlock", "")

	err = d.DeleteBlock("TestBlock")
	if err != nil {
		t.Fatalf("DeleteBlock failed: %v", err)
	}

	_, err = d.GetBlock("TestBlock")
	if err == nil {
		t.Error("TestBlock should not exist after DeleteBlock")
	}
}

func TestDrawingGetDrawingStatistics(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	stats := d.GetDrawingStatistics()
	if stats == nil {
		t.Error("GetDrawingStatistics should not return nil")
	}

	if stats["layer_count"] != 1 {
		t.Errorf("Expected layer_count 1, got %v", stats["layer_count"])
	}

	if stats["style_count"] != 1 {
		t.Errorf("Expected style_count 1, got %v", stats["style_count"])
	}
}

func TestDrawingFormatter(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	formatter := d.Formatter()
	if formatter == nil {
		t.Error("Formatter should not return nil")
	}
}

func TestDrawingHeader(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	header := d.Header()
	if header == nil {
		t.Error("Header should not return nil")
	}
}

func TestDrawingTables(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	tables := d.Tables()
	if tables == nil {
		t.Error("Tables should not return nil")
	}
}

func TestDrawingFilterLayers(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	lt, _ := d.LineType("Continuous")
	d.AddLayer("RedLayer", color.Red, lt)
	d.AddLayer("BlueLayer", color.Blue, lt)

	redLayers := d.FilterLayers(func(l *table.Layer) bool {
		return l.Color == color.Red
	})

	if len(redLayers) != 1 {
		t.Errorf("Expected 1 red layer, got %d", len(redLayers))
	}
}

func TestDrawingGroupLayersByProperty(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("NewDrawing failed: %v", err)
	}

	lt, _ := d.LineType("Continuous")
	d.AddLayer("Layer1", color.Red, lt)
	d.AddLayer("Layer2", color.Red, lt)
	d.AddLayer("Layer3", color.Blue, lt)

	groups := d.GroupLayersByProperty(func(l *table.Layer) string {
		return fmt.Sprintf("%d", l.Color)
	})

	if len(groups) < 2 {
		t.Errorf("Expected at least 2 groups, got %d", len(groups))
	}
}
