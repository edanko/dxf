package dxf

import (
	"bytes"
	"strings"
	"testing"

	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
)

func findAcadTable(t *testing.T, d *drawing.Drawing) *entity.AcadTable {
	t.Helper()
	for _, e := range d.Entities() {
		if table, ok := e.(*entity.AcadTable); ok {
			return table
		}
	}
	t.Fatalf("no ACAD_TABLE entity found")
	return nil
}

func reloadDrawing(t *testing.T, d *drawing.Drawing) *entity.AcadTable {
	t.Helper()
	var buf bytes.Buffer
	if _, err := d.WriteTo(&buf); err != nil {
		t.Fatalf("write drawing: %v", err)
	}
	reloaded, err := FromStringData(buf.String())
	if err != nil {
		t.Fatalf("reload drawing: %v", err)
	}
	return findAcadTable(t, reloaded)
}

func TestAcadTableSimpleTyped(t *testing.T) {
	d, err := FromFile("testdata/acad_table_simple.dxf")
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}
	table := findAcadTable(t, d)

	if table.RowCount != 3 || table.ColCount != 3 {
		t.Fatalf("size = %dx%d, want 3x3", table.RowCount, table.ColCount)
	}
	if table.BlockName != "*T1" {
		t.Errorf("block name = %q, want *T1", table.BlockName)
	}
	if table.TableStyleHandle != "6B" || table.BlockRecordHandle != "9D" {
		t.Errorf("handles = %q/%q, want 6B/9D", table.TableStyleHandle, table.BlockRecordHandle)
	}
	if len(table.InsertPoint) != 3 || table.InsertPoint[0] != 105 || table.InsertPoint[1] != 145 {
		t.Errorf("insert point = %v, want [105 145 0]", table.InsertPoint)
	}
	if len(table.RowHeights) != 3 || len(table.ColWidths) != 3 {
		t.Fatalf("row/col sizes = %d/%d, want 3/3", len(table.RowHeights), len(table.ColWidths))
	}

	cell := table.Rows[0].Cells[0]
	if cell.Text != "Title" || cell.Row != 0 || cell.Col != 0 {
		t.Errorf("cell(0,0) = %q at %d,%d", cell.Text, cell.Row, cell.Col)
	}
	if cell.MergedCell == nil || cell.MergedCell.ColSpan != 3 {
		t.Errorf("cell(0,0) merge = %+v, want col span 3", cell.MergedCell)
	}
	if !table.Rows[0].Cells[1].Merged {
		t.Errorf("cell(0,1) should be covered by the merged title")
	}
	if got := table.Rows[1].Cells[0].Text; got != "1" {
		t.Errorf("cell(1,0) = %q, want 1", got)
	}
	if got := table.Rows[2].Cells[2].Text; got != "6" {
		t.Errorf("cell(2,2) = %q, want 6", got)
	}

	reloaded := reloadDrawing(t, d)
	if reloaded.RowCount != 3 || reloaded.ColCount != 3 {
		t.Errorf("reloaded size = %dx%d, want 3x3", reloaded.RowCount, reloaded.ColCount)
	}
	if got := reloaded.Rows[0].Cells[0].Text; got != "Title" {
		t.Errorf("reloaded cell(0,0) = %q, want Title", got)
	}
	if got := reloaded.Rows[2].Cells[2].Text; got != "6" {
		t.Errorf("reloaded cell(2,2) = %q, want 6", got)
	}
}

func TestAcadTableBlockReferenceTyped(t *testing.T) {
	d, err := FromFile("testdata/acad_table_with_blk_ref.dxf")
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}
	table := findAcadTable(t, d)
	if table.RowCount != 10 || table.ColCount != 5 {
		t.Fatalf("size = %dx%d, want 10x5", table.RowCount, table.ColCount)
	}

	title := table.Rows[0].Cells[0]
	if title.Text != "Title" || title.MergedCell == nil || title.MergedCell.ColSpan != 5 {
		t.Errorf("title cell = %q merge %+v, want col span 5", title.Text, title.MergedCell)
	}

	merge := table.Rows[3].Cells[2]
	if merge.Text == "" || merge.MergedCell == nil {
		t.Fatalf("merge master = %q %+v", merge.Text, merge.MergedCell)
	}
	if merge.MergedCell.RowSpan != 3 || merge.MergedCell.ColSpan != 2 {
		t.Errorf("merge span = %dx%d, want 3x2", merge.MergedCell.RowSpan, merge.MergedCell.ColSpan)
	}
	if !table.Rows[4].Cells[3].Merged {
		t.Errorf("cell(4,3) should be covered by the merge")
	}

	block := table.Rows[4].Cells[4]
	if block.CellType != entity.CellTypeBlock {
		t.Errorf("cell(4,4) type = %v, want block", block.CellType)
	}
	if block.BlockRecordHandle != "10B" {
		t.Errorf("block handle = %q, want 10B", block.BlockRecordHandle)
	}
	if block.BlockScale == 0 {
		t.Errorf("block scale should be populated")
	}

	reloaded := reloadDrawing(t, d)
	if reloaded.RowCount != 10 || reloaded.ColCount != 5 {
		t.Errorf("reloaded size = %dx%d, want 10x5", reloaded.RowCount, reloaded.ColCount)
	}
	if got := reloaded.Rows[0].Cells[0].Text; got != "Title" {
		t.Errorf("reloaded title = %q", got)
	}
	if got := reloaded.Rows[4].Cells[4]; got.CellType != entity.CellTypeBlock || got.BlockRecordHandle != "10B" {
		t.Errorf("reloaded block cell = %+v", got)
	}
	if !strings.Contains(reloaded.Rows[3].Cells[2].Text, "merge") {
		t.Errorf("reloaded merge text = %q", reloaded.Rows[3].Cells[2].Text)
	}
}
