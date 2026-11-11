package block

import (
	"testing"

	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/table"
)

func TestNewBlock(t *testing.T) {
	block := NewBlock("TestBlock", "Test description")

	if block.Name != "TestBlock" {
		t.Errorf("Expected name 'TestBlock', got %s", block.Name)
	}
	if block.Description != "Test description" {
		t.Errorf("Expected description 'Test description', got %s", block.Description)
	}
	if block.Flag != 0 {
		t.Errorf("Expected flag 0, got %d", block.Flag)
	}
	if len(block.Coord) != 3 {
		t.Errorf("Expected 3 coordinates, got %d", len(block.Coord))
	}
}

func TestBlockHandle(t *testing.T) {
	block := NewBlock("Test", "Test")
	if block.Handle() != "" {
		t.Errorf("Expected empty handle, got %s", block.Handle())
	}
}

func TestBlockSetHandle(t *testing.T) {
	block := NewBlock("Test", "Test")
	block.SetHandle(handle.NewHandleGenerator())

	if block.Handle() == "" {
		t.Error("Handle should be set after SetHandle()")
	}
}

func TestBlockLayer(t *testing.T) {
	block := NewBlock("Test", "Test")

	if block.Layer() != nil {
		t.Error("Expected nil layer initially")
	}

	layer := table.NewLayer("TestLayer", 1, nil)
	block.SetLayer(layer)

	if block.Layer() != layer {
		t.Error("Layer not set correctly")
	}
}

func TestBlockName(t *testing.T) {
	block := NewBlock("MyBlock", "Description")
	if block.Name != "MyBlock" {
		t.Errorf("Expected name 'MyBlock', got %s", block.Name)
	}
}

func TestBlockDescription(t *testing.T) {
	block := NewBlock("Test", "My description")
	if block.Description != "My description" {
		t.Errorf("Expected description 'My description', got %s", block.Description)
	}
}

func TestBlockFlag(t *testing.T) {
	block := NewBlock("Test", "Test")

	if block.Flag != 0 {
		t.Errorf("Expected flag 0, got %d", block.Flag)
	}

	block.Flag = 5
	if block.Flag != 5 {
		t.Errorf("Expected flag 5, got %d", block.Flag)
	}
}

func TestBlockCoord(t *testing.T) {
	block := NewBlock("Test", "Test")

	if block.Coord[0] != 0 || block.Coord[1] != 0 || block.Coord[2] != 0 {
		t.Errorf("Expected coords [0,0,0], got [%f,%f,%f]", block.Coord[0], block.Coord[1], block.Coord[2])
	}

	block.Coord[0] = 10
	if block.Coord[0] != 10 {
		t.Errorf("Expected coord[0] = 10, got %f", block.Coord[0])
	}
}

func TestBlockEndHandle(t *testing.T) {
	block := NewBlock("Test", "Test")

	if block.endhandle != "" {
		t.Errorf("Expected empty endhandle, got %s", block.endhandle)
	}

	hg := handle.NewHandleGenerator()
	block.SetHandle(hg)

	if block.endhandle == "" {
		t.Error("endhandle should be set after SetHandle()")
	}

	if block.endhandle == block.Handle() {
		t.Error("handle and endhandle should be different")
	}
}

func TestBlocksNew(t *testing.T) {
	layer := table.NewLayer("TestLayer", 1, nil)
	blocks := New(layer)

	if len(blocks) != 3 {
		t.Errorf("Expected 3 blocks, got %d", len(blocks))
	}

	if blocks[0].Name != "*Model_Space" {
		t.Errorf("Expected *Model_Space, got %s", blocks[0].Name)
	}

	if blocks[1].Name != "*Paper_Space" {
		t.Errorf("Expected *Paper_Space, got %s", blocks[1].Name)
	}

	if blocks[2].Name != "*Paper_Space0" {
		t.Errorf("Expected *Paper_Space0, got %s", blocks[2].Name)
	}

	for _, b := range blocks {
		if b.Layer() != layer {
			t.Error("Layer should be set on all blocks")
		}
	}
}

func TestBlocksAdd(t *testing.T) {
	layer := table.NewLayer("TestLayer", 1, nil)
	blocks := New(layer)

	if len(blocks) != 3 {
		t.Errorf("Expected 3 blocks initially, got %d", len(blocks))
	}

	newBlock := NewBlock("MyBlock", "My description")
	blocks = blocks.Add(newBlock)

	if len(blocks) != 4 {
		t.Errorf("Expected 4 blocks after Add, got %d", len(blocks))
	}

	if blocks[3] != newBlock {
		t.Error("Added block should be the last element")
	}

	if blocks[3].Name != "MyBlock" {
		t.Errorf("Expected block name 'MyBlock', got %s", blocks[3].Name)
	}
}

func TestBlocksSetHandle(t *testing.T) {
	layer := table.NewLayer("TestLayer", 1, nil)
	blocks := New(layer)

	hg := handle.NewHandleGenerator()
	blocks.SetHandle(hg)

	for _, b := range blocks {
		if b.Handle() == "" {
			t.Error("Handle should be set on all blocks")
		}
		if b.endhandle == "" {
			t.Error("endhandle should be set on all blocks")
		}
	}
}

func TestBlockFormat(t *testing.T) {
	block := NewBlock("TestBlock", "Test description")
	layer := table.NewLayer("TestLayer", 1, nil)
	block.SetLayer(layer)

	hg := handle.NewHandleGenerator()
	block.SetHandle(hg)

	f := format.NewASCII()
	block.Format(f)

	output := f.Output()
	if output == "" {
		t.Error("Format should produce output")
	}
	if !contains(output, "BLOCK") {
		t.Error("Output should contain BLOCK")
	}
	if !contains(output, "ENDBLK") {
		t.Error("Output should contain ENDBLK")
	}
	if !contains(output, "TestBlock") {
		t.Error("Output should contain block name")
	}
}

func TestBlocksFormat(t *testing.T) {
	layer := table.NewLayer("TestLayer", 1, nil)
	blocks := New(layer)

	f := format.NewASCII()
	blocks.Format(f)

	output := f.Output()
	if output == "" {
		t.Error("Format should produce output")
	}
	if !contains(output, "SECTION") {
		t.Error("Output should contain SECTION")
	}
	if !contains(output, "BLOCKS") {
		t.Error("Output should contain BLOCKS")
	}
	if !contains(output, "ENDSEC") {
		t.Error("Output should contain ENDSEC")
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
