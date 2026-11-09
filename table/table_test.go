package table

import (
	"testing"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
)

func TestNewLayer(t *testing.T) {
	lt := NewLineType("Continuous", "Continuous")
	layer := NewLayer("TestLayer", color.Red, lt)

	if layer.Name() != "TestLayer" {
		t.Errorf("Expected layer name 'TestLayer', got %s", layer.Name())
	}
	if layer.Color != color.Red {
		t.Errorf("Expected color.Red, got %v", layer.Color)
	}
	if layer.LineType != lt {
		t.Error("LineType not set correctly")
	}
}

func TestLayerIsSymbolTable(t *testing.T) {
	layer := NewLayer("Test", color.Blue, nil)
	if !layer.IsSymbolTable() {
		t.Error("Layer should return true for IsSymbolTable()")
	}
}

func TestLayerHandle(t *testing.T) {
	layer := NewLayer("Test", color.Blue, nil)
	if layer.Handle() != "" {
		t.Errorf("Expected empty handle, got %s", layer.Handle())
	}
}

func TestLayerSetHandle(t *testing.T) {
	layer := NewLayer("Test", color.Blue, nil)
	hg := handle.NewHandleGenerator()
	layer.SetHandle(hg)

	if layer.Handle() == "" {
		t.Error("Handle should be set after SetHandle()")
	}
}

func TestLayerSetLineWidth(t *testing.T) {
	layer := NewLayer("Test", color.Blue, nil)

	// Test valid width
	result := layer.SetLineWidth(50)
	if result != 50 {
		t.Errorf("Expected 50, got %d", result)
	}

	// Test width too large
	result = layer.SetLineWidth(300)
	if result != 211 {
		t.Errorf("Expected 211 for too large width, got %d", result)
	}

	// Test negative width
	result = layer.SetLineWidth(-10)
	if result != -3 {
		t.Errorf("Expected -3 for negative width, got %d", result)
	}
}

func TestLayerFlag(t *testing.T) {
	layer := NewLayer("Test", color.Blue, nil)

	if layer.flags != 0 {
		t.Errorf("Expected flag 0, got %d", layer.flags)
	}

	layer.SetFlag(1)
	if layer.flags != 1 {
		t.Errorf("Expected flag 1, got %d", layer.flags)
	}
}

func TestLayerFreezeUnfreeze(t *testing.T) {
	layer := NewLayer("Test", color.Blue, nil)

	layer.Freeze()
	if layer.flags&1 != 1 {
		t.Error("Layer should be frozen")
	}

	layer.UnFreeze()
	if layer.flags&1 != 0 {
		t.Error("Layer should be unfrozen")
	}
}

func TestLayerLockUnlock(t *testing.T) {
	layer := NewLayer("Test", color.Blue, nil)

	layer.Lock()
	if layer.flags&4 != 4 {
		t.Error("Layer should be locked")
	}

	layer.UnLock()
	if layer.flags&4 != 0 {
		t.Error("Layer should be unlocked")
	}
}

func TestNewLineType(t *testing.T) {
	lt := NewLineType("Dashed", "Dashed line", 1.0, -0.5)

	if lt.Name() != "Dashed" {
		t.Errorf("Expected name 'Dashed', got %s", lt.Name())
	}
	if lt.Description != "Dashed line" {
		t.Errorf("Expected description 'Dashed line', got %s", lt.Description)
	}
	if len(lt.lengths) != 2 {
		t.Errorf("Expected 2 lengths, got %d", len(lt.lengths))
	}
}

func TestNewLineTypeEmpty(t *testing.T) {
	lt := NewLineType("Continuous", "Continuous")

	if len(lt.lengths) != 0 {
		t.Errorf("Expected 0 lengths for continuous, got %d", len(lt.lengths))
	}
}

func TestLineTypeIsSymbolTable(t *testing.T) {
	lt := NewLineType("Test", "Test")
	if !lt.IsSymbolTable() {
		t.Error("LineType should return true for IsSymbolTable()")
	}
}

func TestLineTypeHandle(t *testing.T) {
	lt := NewLineType("Test", "Test")
	if lt.Handle() != "" {
		t.Errorf("Expected empty handle, got %s", lt.Handle())
	}
}

func TestLineTypeSetHandle(t *testing.T) {
	lt := NewLineType("Test", "Test")
	hg := handle.NewHandleGenerator()
	lt.SetHandle(hg)

	if lt.Handle() == "" {
		t.Error("Handle should be set after SetHandle()")
	}
}

func TestLineTypeTotalLength(t *testing.T) {
	lt := NewLineType("Dashed", "Dashed", 1.0, -0.5)
	length := lt.TotalLength()

	if length != 1.5 {
		t.Errorf("Expected total length 1.5, got %f", length)
	}

	lt2 := NewLineType("Continuous", "Continuous")
	length2 := lt2.TotalLength()
	if length2 != 0 {
		t.Errorf("Expected total length 0 for continuous, got %f", length2)
	}
}

func TestLineTypeSetLength(t *testing.T) {
	lt := NewLineType("Test", "Test")
	lt.SetLength([]float64{2.0, -1.0, 0.5})

	if len(lt.lengths) != 3 {
		t.Errorf("Expected 3 lengths, got %d", len(lt.lengths))
	}
}

func TestLineWidthMap(t *testing.T) {
	if LineWidth[50] != 0.50 {
		t.Errorf("LineWidth[50] should be 0.50, got %f", LineWidth[50])
	}
	if LineWidth[100] != 1.00 {
		t.Errorf("LineWidth[100] should be 1.00, got %f", LineWidth[100])
	}
	if LineWidth[211] != 2.11 {
		t.Errorf("LineWidth[211] should be 2.11, got %f", LineWidth[211])
	}
}

func TestStyleNew(t *testing.T) {
	style := NewStyle("MyStyle")

	if style.Name() != "MyStyle" {
		t.Errorf("Expected name 'MyStyle', got %s", style.Name())
	}
	if style.FontName != "arial.ttf" {
		t.Errorf("Expected FontName 'arial.ttf', got %s", style.FontName)
	}
	if style.FixedTextHeight != 0.0 {
		t.Errorf("Expected FixedTextHeight 0.0, got %f", style.FixedTextHeight)
	}
	if style.WidthFactor != 1.0 {
		t.Errorf("Expected WidthFactor 1.0, got %f", style.WidthFactor)
	}
}

func TestStyleIsSymbolTable(t *testing.T) {
	style := NewStyle("Test")
	if !style.IsSymbolTable() {
		t.Error("Style should return true for IsSymbolTable()")
	}
}

func TestStyleHandle(t *testing.T) {
	style := NewStyle("Test")
	if style.Handle() != "" {
		t.Errorf("Expected empty handle, got %s", style.Handle())
	}
}

func TestStyleSetHandle(t *testing.T) {
	style := NewStyle("Test")
	hg := handle.NewHandleGenerator()
	style.SetHandle(hg)

	if style.Handle() == "" {
		t.Error("Handle should be set after SetHandle()")
	}
}

func TestStyleFormat(t *testing.T) {
	style := NewStyle("MyStyle")
	hg := handle.NewHandleGenerator()
	style.SetHandle(hg)

	f := format.NewASCII()
	style.Format(f)

	output := f.Output()
	if output == "" {
		t.Error("Format should produce output")
	}
	if !contains(output, "STYLE") {
		t.Error("Output should contain STYLE")
	}
	if !contains(output, "MyStyle") {
		t.Error("Output should contain style name")
	}
}

func TestStyleAttributes(t *testing.T) {
	style := NewStyle("Test")
	style.FontName = "times.ttf"
	style.BigFontName = "big.ttf"
	style.FixedTextHeight = 2.5
	style.WidthFactor = 0.8
	style.LastHeightUsed = 150.0
	style.ObliqueAngle = 15.0

	if style.FontName != "times.ttf" {
		t.Errorf("Expected FontName 'times.ttf', got %s", style.FontName)
	}
	if style.BigFontName != "big.ttf" {
		t.Errorf("Expected BigFontName 'big.ttf', got %s", style.BigFontName)
	}
	if style.FixedTextHeight != 2.5 {
		t.Errorf("Expected FixedTextHeight 2.5, got %f", style.FixedTextHeight)
	}
	if style.WidthFactor != 0.8 {
		t.Errorf("Expected WidthFactor 0.8, got %f", style.WidthFactor)
	}
	if style.LastHeightUsed != 150.0 {
		t.Errorf("Expected LastHeightUsed 150.0, got %f", style.LastHeightUsed)
	}
	if style.ObliqueAngle != 15.0 {
		t.Errorf("Expected ObliqueAngle 15.0, got %f", style.ObliqueAngle)
	}
}

func TestAppIDNew(t *testing.T) {
	appid := NewAppID("ACAD")

	if appid.Name() != "ACAD" {
		t.Errorf("Expected name 'ACAD', got %s", appid.Name())
	}
}

func TestAppIDIsSymbolTable(t *testing.T) {
	appid := NewAppID("Test")
	if !appid.IsSymbolTable() {
		t.Error("AppID should return true for IsSymbolTable()")
	}
}

func TestAppIDHandle(t *testing.T) {
	appid := NewAppID("Test")
	if appid.Handle() != "" {
		t.Errorf("Expected empty handle, got %s", appid.Handle())
	}
}

func TestAppIDSetHandle(t *testing.T) {
	appid := NewAppID("Test")
	hg := handle.NewHandleGenerator()
	appid.SetHandle(hg)

	if appid.Handle() == "" {
		t.Error("Handle should be set after SetHandle()")
	}
}

func TestBlockRecordNew(t *testing.T) {
	br := NewBlockRecord("MyBlock")

	if br.Name() != "MyBlock" {
		t.Errorf("Expected name 'MyBlock', got %s", br.Name())
	}
}

func TestBlockRecordIsSymbolTable(t *testing.T) {
	br := NewBlockRecord("Test")
	if !br.IsSymbolTable() {
		t.Error("BlockRecord should return true for IsSymbolTable()")
	}
}

func TestBlockRecordHandle(t *testing.T) {
	br := NewBlockRecord("Test")
	if br.Handle() != "" {
		t.Errorf("Expected empty handle, got %s", br.Handle())
	}
}

func TestBlockRecordSetHandle(t *testing.T) {
	br := NewBlockRecord("Test")
	hg := handle.NewHandleGenerator()
	br.SetHandle(hg)

	if br.Handle() == "" {
		t.Error("Handle should be set after SetHandle()")
	}
}

func TestLayerFrozen(t *testing.T) {
	layer := NewLayer("Test", color.Blue, nil)

	if layer.flags&1 != 0 {
		t.Error("Layer should not be frozen initially")
	}

	layer.Freeze()
	if layer.flags&1 != 1 {
		t.Error("Layer should be frozen after Freeze()")
	}
}

func TestLayerLocked(t *testing.T) {
	layer := NewLayer("Test", color.Blue, nil)

	if layer.flags&4 != 0 {
		t.Error("Layer should not be locked initially")
	}

	layer.Lock()
	if layer.flags&4 != 4 {
		t.Error("Layer should be locked after Lock()")
	}
}

func TestLineTypeDashed(t *testing.T) {
	lt := NewLineType("Dashed", "Dashed line", 1.0, -0.5)

	lengths := lt.lengths
	if len(lengths) != 2 {
		t.Errorf("Expected 2 lengths, got %d", len(lengths))
	}
	if lengths[0] != 1.0 {
		t.Errorf("Expected first length 1.0, got %f", lengths[0])
	}
	if lengths[1] != -0.5 {
		t.Errorf("Expected second length -0.5, got %f", lengths[1])
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestViewportNew(t *testing.T) {
	vp := NewViewport("Active")

	if vp == nil {
		t.Fatal("NewViewport should not return nil")
	}

	if vp.name != "Active" {
		t.Errorf("Expected name 'Active', got '%s'", vp.name)
	}

	if len(vp.LowerLeft) != 2 {
		t.Errorf("Expected LowerLeft with 2 elements, got %d", len(vp.LowerLeft))
	}

	if vp.Height != 400.0 {
		t.Errorf("Expected Height 400.0, got %f", vp.Height)
	}

	if vp.AspectRatio != 1.0 {
		t.Errorf("Expected AspectRatio 1.0, got %f", vp.AspectRatio)
	}
}

func TestViewportIsSymbolTable(t *testing.T) {
	vp := NewViewport("Test")

	if !vp.IsSymbolTable() {
		t.Error("Viewport should return true for IsSymbolTable()")
	}
}

func TestViewportHandle(t *testing.T) {
	vp := NewViewport("Test")

	if vp.Handle() != "" {
		t.Errorf("Expected empty handle, got %s", vp.Handle())
	}
}

func TestViewportSetHandle(t *testing.T) {
	vp := NewViewport("Test")
	hg := handle.NewHandleGenerator()
	vp.SetHandle(hg)

	if vp.Handle() == "" {
		t.Error("Handle should be set after SetHandle()")
	}
}

func TestViewportSetOwner(t *testing.T) {
	vp := NewViewport("Test")
	owner := NewLayer("Owner", color.Red, nil)
	vp.SetOwner(owner)

	if vp.owner != owner {
		t.Error("Owner should be set after SetOwner()")
	}
}

func TestViewportCenter(t *testing.T) {
	vp := NewViewport("Test")
	vp.ViewCenter = []float64{100, 200}

	if vp.ViewCenter[0] != 100 {
		t.Error("ViewCenter X not set correctly")
	}

	if vp.ViewCenter[1] != 200 {
		t.Error("ViewCenter Y not set correctly")
	}
}

func TestViewportDirection(t *testing.T) {
	vp := NewViewport("Test")
	vp.ViewDirection = []float64{1, 0, 0}
	vp.ViewTarget = []float64{0, 0, 0}

	if vp.ViewDirection[0] != 1 {
		t.Error("ViewDirection not set correctly")
	}
}

func TestViewportFormat(t *testing.T) {
	vp := NewViewport("Test")
	hg := handle.NewHandleGenerator()
	vp.SetHandle(hg)

	f := format.NewASCII()
	vp.Format(f)

	output := f.Output()
	if output == "" {
		t.Error("Format should produce output")
	}
	if !contains(output, "VPORT") {
		t.Error("Output should contain VPORT")
	}
}

func TestDimStyleNew(t *testing.T) {
	ds := NewDimStyle("Standard")

	if ds == nil {
		t.Fatal("NewDimStyle should not return nil")
	}

	if ds.Name() != "Standard" {
		t.Errorf("Expected name 'Standard', got '%s'", ds.Name())
	}
}

func TestDimStyleIsSymbolTable(t *testing.T) {
	ds := NewDimStyle("Test")

	if !ds.IsSymbolTable() {
		t.Error("DimStyle should return true for IsSymbolTable()")
	}
}

func TestDimStyleHandle(t *testing.T) {
	ds := NewDimStyle("Test")

	if ds.Handle() != "" {
		t.Errorf("Expected empty handle, got %s", ds.Handle())
	}
}

func TestDimStyleSetHandle(t *testing.T) {
	ds := NewDimStyle("Test")
	hg := handle.NewHandleGenerator()
	ds.SetHandle(hg)

	if ds.Handle() == "" {
		t.Error("Handle should be set after SetHandle()")
	}
}

func TestDimStyleSetOwner(t *testing.T) {
	ds := NewDimStyle("Test")
	owner := NewLayer("Owner", color.Red, nil)
	ds.SetOwner(owner)

	if ds.owner != owner {
		t.Error("Owner should be set after SetOwner()")
	}
}

func TestDimStyleFormat(t *testing.T) {
	ds := NewDimStyle("Standard")
	hg := handle.NewHandleGenerator()
	ds.SetHandle(hg)

	f := format.NewASCII()
	ds.Format(f)

	output := f.Output()
	if output == "" {
		t.Error("Format should produce output")
	}
	if !contains(output, "DIMSTYLE") {
		t.Error("Output should contain DIMSTYLE")
	}
}

func TestViewNew(t *testing.T) {
	v := NewView("Top")

	if v == nil {
		t.Fatal("NewView should not return nil")
	}

	if v.Name() != "Top" {
		t.Errorf("Expected name 'Top', got '%s'", v.Name())
	}
}

func TestViewIsSymbolTable(t *testing.T) {
	v := NewView("Test")

	if !v.IsSymbolTable() {
		t.Error("View should return true for IsSymbolTable()")
	}
}

func TestViewHandle(t *testing.T) {
	v := NewView("Test")

	if v.Handle() != "" {
		t.Errorf("Expected empty handle, got %s", v.Handle())
	}
}

func TestViewSetHandle(t *testing.T) {
	v := NewView("Test")
	hg := handle.NewHandleGenerator()
	v.SetHandle(hg)

	if v.Handle() == "" {
		t.Error("Handle should be set after SetHandle()")
	}
}

func TestViewSetOwner(t *testing.T) {
	v := NewView("Test")
	owner := NewLayer("Owner", color.Red, nil)
	v.SetOwner(owner)

	if v.owner != owner {
		t.Error("Owner should be set after SetOwner()")
	}
}

func TestViewFormat(t *testing.T) {
	v := NewView("Top")
	hg := handle.NewHandleGenerator()
	v.SetHandle(hg)

	f := format.NewASCII()
	v.Format(f)

	output := f.Output()
	if output == "" {
		t.Error("Format should produce output")
	}
	if !contains(output, "VIEW") {
		t.Error("Output should contain VIEW")
	}
}

func TestUCSNew(t *testing.T) {
	u := NewUCS("Front")

	if u == nil {
		t.Fatal("NewUCS should not return nil")
	}

	if u.Name() != "Front" {
		t.Errorf("Expected name 'Front', got '%s'", u.Name())
	}
}

func TestUCSIsSymbolTable(t *testing.T) {
	u := NewUCS("Test")

	if !u.IsSymbolTable() {
		t.Error("UCS should return true for IsSymbolTable()")
	}
}

func TestUCSHandle(t *testing.T) {
	u := NewUCS("Test")

	if u.Handle() != "" {
		t.Errorf("Expected empty handle, got %s", u.Handle())
	}
}

func TestUCSSetHandle(t *testing.T) {
	u := NewUCS("Test")
	hg := handle.NewHandleGenerator()
	u.SetHandle(hg)

	if u.Handle() == "" {
		t.Error("Handle should be set after SetHandle()")
	}
}

func TestUCSSetOwner(t *testing.T) {
	u := NewUCS("Test")
	owner := NewLayer("Owner", color.Red, nil)
	u.SetOwner(owner)

	if u.owner != owner {
		t.Error("Owner should be set after SetOwner()")
	}
}

func TestUCSFormat(t *testing.T) {
	u := NewUCS("Front")
	hg := handle.NewHandleGenerator()
	u.SetHandle(hg)

	f := format.NewASCII()
	u.Format(f)

	output := f.Output()
	if output == "" {
		t.Error("Format should produce output")
	}
	if !contains(output, "UCS") {
		t.Error("Output should contain UCS")
	}
}
