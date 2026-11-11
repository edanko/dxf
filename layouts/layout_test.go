package layouts

import (
	"testing"

	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/table"
)

func createTestDrawing() *drawing.Drawing {
	d, _ := drawing.New()
	return d
}

func TestNewLayout(t *testing.T) {
	d := createTestDrawing()
	layout := NewLayout("Test", LayoutTypeModelSpace, d)

	if layout.Name() != "Test" {
		t.Errorf("Expected name 'Test', got %s", layout.Name())
	}
	if layout.Type() != LayoutTypeModelSpace {
		t.Errorf("Expected type ModelSpace, got %v", layout.Type())
	}
	if !layout.IsModelSpace() {
		t.Error("IsModelSpace should return true")
	}
	if layout.IsPaperSpace() {
		t.Error("IsPaperSpace should return false")
	}
}

func TestNewPaperSpace(t *testing.T) {
	d := createTestDrawing()
	ps := NewPaperSpace("Layout1", d)

	if ps.Name() != "Layout1" {
		t.Errorf("Expected name 'Layout1', got %s", ps.Name())
	}
	if !ps.IsPaperSpace() {
		t.Error("IsPaperSpace should return true")
	}
}

func TestNewModelSpace(t *testing.T) {
	d := createTestDrawing()
	ms := NewModelSpace(d)

	if ms.Name() != "Model" {
		t.Errorf("Expected name 'Model', got %s", ms.Name())
	}
	if ms.TabOrder() != 0 {
		t.Errorf("Expected tab order 0, got %d", ms.TabOrder())
	}
}

func TestLayoutHandle(t *testing.T) {
	d := createTestDrawing()
	layout := NewLayout("Test", LayoutTypeModelSpace, d)

	if layout.Handle() != "" {
		t.Errorf("Expected empty handle, got %s", layout.Handle())
	}

	layout.SetHandle("ABC")
	if layout.Handle() != "ABC" {
		t.Errorf("Expected handle 'ABC', got %s", layout.Handle())
	}
}

func TestLayoutBlockRecord(t *testing.T) {
	d := createTestDrawing()
	layout := NewLayout("Test", LayoutTypePaperSpace, d)

	if layout.BlockRecord() != nil {
		t.Error("Expected nil block record initially")
	}

	br := table.NewBlockRecord("TestBlock")
	layout.SetBlockRecord(br)
	if layout.BlockRecord() != br {
		t.Error("BlockRecord not set correctly")
	}
}

func TestLayoutViewport(t *testing.T) {
	d := createTestDrawing()
	layout := NewLayout("Test", LayoutTypePaperSpace, d)

	if layout.Viewport() != nil {
		t.Error("Expected nil viewport initially")
	}

	vp := NewViewport()
	layout.SetViewport(vp)
	if layout.Viewport() != vp {
		t.Error("Viewport not set correctly")
	}
}

func TestLayoutEntityOperations(t *testing.T) {
	d := createTestDrawing()
	layout := NewLayout("Test", LayoutTypeModelSpace, d)

	if layout.EntityCount() != 0 {
		t.Errorf("Expected 0 entities initially, got %d", layout.EntityCount())
	}

	layout.Clear()
	if layout.EntityCount() != 0 {
		t.Errorf("Expected 0 entities after Clear, got %d", layout.EntityCount())
	}
}

func TestNewViewport(t *testing.T) {
	vp := NewViewport()

	if vp.Width() != 1.0 {
		t.Errorf("Expected width 1.0, got %f", vp.Width())
	}
	if vp.Height() != 1.0 {
		t.Errorf("Expected height 1.0, got %f", vp.Height())
	}
	if vp.ViewHeight() != 1.0 {
		t.Errorf("Expected view height 1.0, got %f", vp.ViewHeight())
	}
}

func TestViewportCenter(t *testing.T) {
	vp := NewViewport()
	center := vp.Center()
	if center[0] != 0 || center[1] != 0 || center[2] != 0 {
		t.Errorf("Expected center [0,0,0], got [%f,%f,%f]", center[0], center[1], center[2])
	}

	vp.SetCenter(10, 20, 30)
	center = vp.Center()
	if center[0] != 10 || center[1] != 20 || center[2] != 30 {
		t.Errorf("Expected center [10,20,30], got [%f,%f,%f]", center[0], center[1], center[2])
	}
}

func TestViewportHandle(t *testing.T) {
	vp := NewViewport()

	if vp.Handle() != "" {
		t.Errorf("Expected empty handle, got %s", vp.Handle())
	}

	vp.SetHandle("XYZ")
	if vp.Handle() != "XYZ" {
		t.Errorf("Expected handle 'XYZ', got %s", vp.Handle())
	}
}

func TestViewportLayer(t *testing.T) {
	vp := NewViewport()

	if vp.Layer() != "0" {
		t.Errorf("Expected layer '0', got %s", vp.Layer())
	}

	vp.SetLayer("MyLayer")
	if vp.Layer() != "MyLayer" {
		t.Errorf("Expected layer 'MyLayer', got %s", vp.Layer())
	}
}

func TestLayoutString(t *testing.T) {
	d := createTestDrawing()
	layout := NewLayout("Test", LayoutTypeModelSpace, d)

	s := layout.String()
	if !containsString(s, "Test") {
		t.Errorf("Expected string to contain 'Test', got %s", s)
	}
	if !containsString(s, "Model Space") {
		t.Errorf("Expected string to contain 'Model Space', got %s", s)
	}
}

func TestViewportString(t *testing.T) {
	vp := NewViewport()
	s := vp.String()

	if !containsString(s, "Viewport") {
		t.Errorf("Expected string to contain 'Viewport', got %s", s)
	}
}

func TestLayoutTypeConstants(t *testing.T) {
	if LayoutTypeModelSpace != 0 {
		t.Errorf("LayoutTypeModelSpace should be 0, got %d", LayoutTypeModelSpace)
	}
	if LayoutTypePaperSpace != 1 {
		t.Errorf("LayoutTypePaperSpace should be 1, got %d", LayoutTypePaperSpace)
	}
}

func TestLayoutSetTabOrder(t *testing.T) {
	d := createTestDrawing()
	layout := NewLayout("Test", LayoutTypeModelSpace, d)

	if layout.TabOrder() != 0 {
		t.Errorf("Expected tab order 0, got %d", layout.TabOrder())
	}

	layout.SetTabOrder(5)
	if layout.TabOrder() != 5 {
		t.Errorf("Expected tab order 5, got %d", layout.TabOrder())
	}
}

func TestLayoutEntities(t *testing.T) {
	d := createTestDrawing()
	layout := NewLayout("Test", LayoutTypeModelSpace, d)

	entities := layout.Entities()
	if len(entities) != 0 {
		t.Errorf("Expected 0 entities, got %d", len(entities))
	}
}

func TestLayoutsNewLayouts(t *testing.T) {
	d := createTestDrawing()
	ls := NewLayouts(d)

	if ls.LayoutCount() != 0 {
		t.Errorf("Expected 0 layouts initially, got %d", ls.LayoutCount())
	}
}

func TestValidateLayoutName(t *testing.T) {
	err := ValidateLayoutName("ValidName")
	if err != nil {
		t.Errorf("Expected nil error for valid name, got %v", err)
	}

	err = ValidateLayoutName("")
	if err == nil {
		t.Error("Expected error for empty name")
	}

	err = ValidateLayoutName("Invalid/Name")
	if err == nil {
		t.Error("Expected error for name with invalid characters")
	}
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestLayoutAddEntity(t *testing.T) {
	d := createTestDrawing()
	layout := NewLayout("Test", LayoutTypeModelSpace, d)

	line := entity.NewLine()
	layout.AddEntity(line)

	if layout.EntityCount() != 1 {
		t.Errorf("Expected 1 entity after AddEntity, got %d", layout.EntityCount())
	}
}

func TestLayoutRemoveEntity(t *testing.T) {
	d := createTestDrawing()
	layout := NewLayout("Test", LayoutTypeModelSpace, d)

	line := entity.NewLine()
	layout.AddEntity(line)
	result := layout.RemoveEntity(line)

	if !result {
		t.Error("RemoveEntity should return true for existing entity")
	}

	if layout.EntityCount() != 0 {
		t.Errorf("Expected 0 entities after RemoveEntity, got %d", layout.EntityCount())
	}
}

func TestLayoutRemoveNonExistent(t *testing.T) {
	d := createTestDrawing()
	layout := NewLayout("Test", LayoutTypeModelSpace, d)

	line := entity.NewLine()
	result := layout.RemoveEntity(line)

	if result {
		t.Error("RemoveEntity should return false for non-existent entity")
	}
}

func TestLayoutClear(t *testing.T) {
	d := createTestDrawing()
	layout := NewLayout("Test", LayoutTypeModelSpace, d)

	layout.AddEntity(entity.NewLine())
	layout.AddEntity(entity.NewLine())
	layout.Clear()

	if layout.EntityCount() != 0 {
		t.Errorf("Expected 0 entities after Clear, got %d", layout.EntityCount())
	}
}

func TestViewportSetWidth(t *testing.T) {
	vp := NewViewport()

	vp.SetWidth(11.0)

	if vp.Width() != 11.0 {
		t.Errorf("Expected width 11.0, got %f", vp.Width())
	}
}

func TestViewportSetHeight(t *testing.T) {
	vp := NewViewport()

	vp.SetHeight(8.5)

	if vp.Height() != 8.5 {
		t.Errorf("Expected height 8.5, got %f", vp.Height())
	}
}

func TestViewportViewDirection(t *testing.T) {
	vp := NewViewport()

	dir := vp.ViewDirection()
	if dir[0] != 0 || dir[1] != 0 || dir[2] != 1 {
		t.Errorf("Expected default view direction [0,0,1], got [%f,%f,%f]", dir[0], dir[1], dir[2])
	}
}

func TestViewportSetViewDirection(t *testing.T) {
	vp := NewViewport()

	vp.SetViewDirection(1, 0, 0)

	dir := vp.ViewDirection()
	if dir[0] != 1 || dir[1] != 0 || dir[2] != 0 {
		t.Errorf("Expected view direction [1,0,0], got [%f,%f,%f]", dir[0], dir[1], dir[2])
	}
}

func TestViewportViewTarget(t *testing.T) {
	vp := NewViewport()

	target := vp.ViewTarget()
	if target[0] != 0 || target[1] != 0 || target[2] != 0 {
		t.Errorf("Expected default view target [0,0,0], got [%f,%f,%f]", target[0], target[1], target[2])
	}
}

func TestViewportSetViewTarget(t *testing.T) {
	vp := NewViewport()

	vp.SetViewTarget(10, 20, 30)

	target := vp.ViewTarget()
	if target[0] != 10 || target[1] != 20 || target[2] != 30 {
		t.Errorf("Expected view target [10,20,30], got [%f,%f,%f]", target[0], target[1], target[2])
	}
}

func TestViewportSetViewHeight(t *testing.T) {
	vp := NewViewport()

	vp.SetViewHeight(5.0)

	if vp.ViewHeight() != 5.0 {
		t.Errorf("Expected view height 5.0, got %f", vp.ViewHeight())
	}
}

func TestViewportViewTwistAngle(t *testing.T) {
	vp := NewViewport()

	if vp.ViewTwistAngle() != 0 {
		t.Errorf("Expected default view twist angle 0, got %f", vp.ViewTwistAngle())
	}
}

func TestViewportSetViewTwistAngle(t *testing.T) {
	vp := NewViewport()

	vp.SetViewTwistAngle(45.0)

	if vp.ViewTwistAngle() != 45.0 {
		t.Errorf("Expected view twist angle 45.0, got %f", vp.ViewTwistAngle())
	}
}

func TestLayoutsSetupLayouts(t *testing.T) {
	d := createTestDrawing()
	ls := NewLayouts(d)
	ls.SetupLayouts()

	if ls.LayoutCount() < 2 {
		t.Errorf("Expected at least 2 layouts (Model + Paper), got %d", ls.LayoutCount())
	}
}

func TestLayoutsModelSpace(t *testing.T) {
	d := createTestDrawing()
	ls := NewLayouts(d)
	ls.SetupLayouts()

	ms := ls.ModelSpace()
	if ms == nil {
		t.Error("ModelSpace should not be nil")
	}

	if !ms.IsModelSpace() {
		t.Error("ModelSpace should return model space layout")
	}
}

func TestLayoutsPaperSpaces(t *testing.T) {
	d := createTestDrawing()
	ls := NewLayouts(d)
	ls.SetupLayouts()

	papers := ls.PaperSpaces()
	if len(papers) < 1 {
		t.Error("PaperSpaces should return at least one paper space")
	}

	for _, ps := range papers {
		if !ps.IsPaperSpace() {
			t.Error("PaperSpaces should only contain paper space layouts")
		}
	}
}

func TestLayoutsGetLayout(t *testing.T) {
	d := createTestDrawing()
	ls := NewLayouts(d)
	ls.SetupLayouts()

	layout := ls.GetLayout("Model")
	if layout == nil {
		t.Error("GetLayout should return Model layout")
	}

	if layout.Name() != "Model" {
		t.Errorf("Expected layout name 'Model', got %s", layout.Name())
	}
}

func TestLayoutsGetLayoutNotFound(t *testing.T) {
	d := createTestDrawing()
	ls := NewLayouts(d)
	ls.SetupLayouts()

	layout := ls.GetLayout("NonExistent")
	if layout != nil {
		t.Error("GetLayout should return nil for non-existent layout")
	}
}

func TestLayoutsAddPaperSpace(t *testing.T) {
	d := createTestDrawing()
	ls := NewLayouts(d)
	ls.SetupLayouts()

	initialCount := ls.LayoutCount()
	ls.AddPaperSpace("Layout2")

	if ls.LayoutCount() != initialCount+1 {
		t.Errorf("Expected layout count %d after AddPaperSpace, got %d", initialCount+1, ls.LayoutCount())
	}

	newLayout := ls.GetLayout("Layout2")
	if newLayout == nil {
		t.Error("Newly added paper space should be retrievable")
	}
}

func TestLayoutsRemoveLayout(t *testing.T) {
	d := createTestDrawing()
	ls := NewLayouts(d)
	ls.SetupLayouts()
	ls.AddPaperSpace("ToRemove")

	initialCount := ls.LayoutCount()
	ls.RemoveLayout("ToRemove")

	if ls.LayoutCount() != initialCount-1 {
		t.Errorf("Expected layout count %d after RemoveLayout, got %d", initialCount-1, ls.LayoutCount())
	}
}

func TestLayoutsRemoveNonExistent(t *testing.T) {
	d := createTestDrawing()
	ls := NewLayouts(d)
	ls.SetupLayouts()

	err := ls.RemoveLayout("NonExistent")
	if err == nil {
		t.Error("RemoveLayout should return error for non-existent layout")
	}
}

func TestLayoutsRenameLayout(t *testing.T) {
	d := createTestDrawing()
	ls := NewLayouts(d)
	ls.SetupLayouts()
	ls.AddPaperSpace("OriginalName")

	err := ls.RenameLayout("OriginalName", "NewName")
	if err != nil {
		t.Errorf("RenameLayout should not return error, got %v", err)
	}

	renamed := ls.GetLayout("NewName")
	if renamed == nil {
		t.Error("Renamed layout should be retrievable with new name")
	}

	old := ls.GetLayout("OriginalName")
	if old != nil {
		t.Error("Old name should not be retrievable")
	}
}

func TestLayoutsRenameLayoutInvalid(t *testing.T) {
	d := createTestDrawing()
	ls := NewLayouts(d)
	ls.SetupLayouts()

	err := ls.RenameLayout("NonExistent", "NewName")
	if err == nil {
		t.Error("RenameLayout should return error for non-existent layout")
	}
}

func TestLayoutsRenameLayoutDuplicate(t *testing.T) {
	d := createTestDrawing()
	ls := NewLayouts(d)
	ls.SetupLayouts()

	err := ls.RenameLayout("Model", "Model")
	if err == nil {
		t.Error("RenameLayout should return error for duplicate name")
	}
}

func TestLayoutsGetTabOrder(t *testing.T) {
	d := createTestDrawing()
	ls := NewLayouts(d)
	ls.SetupLayouts()

	order := ls.GetTabOrder()
	if len(order) < 1 {
		t.Error("GetTabOrder should return at least one layout")
	}
}

func TestLayoutsSetTabOrder(t *testing.T) {
	d := createTestDrawing()
	ls := NewLayouts(d)
	ls.SetupLayouts()

	err := ls.SetTabOrder("Model", 5)
	if err != nil {
		t.Errorf("SetTabOrder should not return error, got %v", err)
	}

	order := ls.GetTabOrder()
	if len(order) < 1 {
		t.Error("GetTabOrder should return layouts after SetTabOrder")
	}
}

func TestLayoutsForEach(t *testing.T) {
	d := createTestDrawing()
	ls := NewLayouts(d)
	ls.SetupLayouts()

	count := 0
	ls.ForEach(func(layout *Layout) {
		count++
	})

	if count != ls.LayoutCount() {
		t.Errorf("ForEach should visit all layouts, expected %d, got %d", ls.LayoutCount(), count)
	}
}

func TestLayoutsGetLayoutByHandle(t *testing.T) {
	d := createTestDrawing()
	ls := NewLayouts(d)
	ls.SetupLayouts()
	modelLayout := ls.ModelSpace()
	modelLayout.SetHandle("ABC123")

	layout := ls.GetLayoutByHandle("ABC123")
	if layout == nil {
		t.Error("GetLayoutByHandle should return layout by handle")
	}

	if layout.Handle() != "ABC123" {
		t.Errorf("Expected handle ABC123, got %s", layout.Handle())
	}
}

func TestLayoutsCreateLayoutFromTemplate(t *testing.T) {
	d := createTestDrawing()
	ls := NewLayouts(d)
	ls.SetupLayouts()

	ls.AddPaperSpace("TemplateLayout")
	layout, err := ls.CreateLayoutFromTemplate("Layout1", "CopiedLayout")
	if err != nil {
		t.Errorf("CreateLayoutFromTemplate should not return error, got %v", err)
	}

	if layout == nil {
		t.Error("Created layout should not be nil")
	}

	if layout.Name() != "CopiedLayout" {
		t.Errorf("Expected layout name 'CopiedLayout', got %s", layout.Name())
	}
}

func TestLayoutsCreateLayoutFromTemplateInvalid(t *testing.T) {
	d := createTestDrawing()
	ls := NewLayouts(d)
	ls.SetupLayouts()

	_, err := ls.CreateLayoutFromTemplate("NonExistent", "NewLayout")
	if err == nil {
		t.Error("CreateLayoutFromTemplate should return error for non-existent template")
	}
}

func TestLayoutsString(t *testing.T) {
	d := createTestDrawing()
	ls := NewLayouts(d)
	ls.SetupLayouts()

	s := ls.String()
	if s == "" {
		t.Error("String should not be empty")
	}
}
