package dxf

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/handle"
)

func facadeSaveReload(t *testing.T, d *drawing.Drawing) *drawing.Drawing {
	t.Helper()
	fn := filepath.Join(t.TempDir(), "facade.dxf")
	if err := d.SaveAs(fn); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}
	f, err := os.Open(fn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()
	nd, err := FromReader(f)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	return nd
}

type ownerGetter interface{ Owner() handle.Handler }

func TestFacadeHandleUniqueness(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	m := d.Model()
	for i := range 100 {
		if _, err := m.AddPoint(float64(i), 0, 0); err != nil {
			t.Fatalf("AddPoint %d: %v", i, err)
		}
	}
	nd := facadeSaveReload(t, d)
	seen := map[string]bool{}
	for _, e := range nd.Entities() {
		h := e.Handle()
		if h == "" {
			t.Fatal("entity without handle after reload")
		}
		if seen[h] {
			t.Fatalf("duplicate handle %s", h)
		}
		seen[h] = true
	}
	if len(seen) != 100 {
		t.Fatalf("reloaded %d entities, want 100", len(seen))
	}
}

func TestFacadeOwnerAssignedOnSave(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	m := d.Model()
	if _, err := m.AddLine(0, 0, 0, 5, 5, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := m.AddCircle(0, 0, 0, 2); err != nil {
		t.Fatal(err)
	}
	nd := facadeSaveReload(t, d)
	rec, ok := nd.BlockRecordHandle("*Model_Space")
	if !ok {
		t.Fatal("no BLOCK_RECORD for *Model_Space after reload")
	}
	count := 0
	for _, e := range nd.Entities() {
		owner, ok := e.(ownerGetter)
		if !ok {
			continue
		}
		count++
		if owner.Owner() == nil {
			t.Fatalf("entity %s (%s) has no owner after reload", e.DXFType(), e.Handle())
		}
		if owner.Owner().Handle() != rec {
			t.Fatalf("entity %s owner = %s, want block record %s", e.DXFType(), owner.Owner().Handle(), rec)
		}
	}
	if count != 2 {
		t.Fatalf("checked %d entities, want 2", count)
	}
}

func TestFacadeBlockRoundTrip(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	b, err := d.NewBlock("TITLEBLOCK", "title block")
	if err != nil {
		t.Fatal(err)
	}
	bs := d.BlockSpace(b)
	if _, err := bs.AddLine(0, 0, 0, 100, 0, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := bs.AddText(1, 1, 0, 2.5, "TITLE"); err != nil {
		t.Fatal(err)
	}
	if _, err := bs.AddAttdef("SHEET", "Sheet", "1", 90, 1, 0, 2.5); err != nil {
		t.Fatal(err)
	}
	nd := facadeSaveReload(t, d)
	nb, err := nd.GetBlock("TITLEBLOCK")
	if err != nil {
		t.Fatalf("block after reload: %v", err)
	}
	if len(nb.Entities()) != 3 {
		t.Fatalf("block entities = %d, want 3", len(nb.Entities()))
	}
	seen := map[string]bool{}
	for _, e := range nb.Entities() {
		h := e.Handle()
		if h == "" || seen[h] {
			t.Fatalf("block entity bad handle %q", h)
		}
		seen[h] = true
	}
}

func TestFacadeInsertAttribsReload(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	sb, err := d.NewBlock("SYM", "symbol")
	if err != nil {
		t.Fatal(err)
	}
	bs := d.BlockSpace(sb)
	if _, err := bs.AddAttdef("SIZE", "Size", "10", 0, 0, 0, 2.5); err != nil {
		t.Fatal(err)
	}
	if _, err := bs.AddCircle(0, 0, 0, 5); err != nil {
		t.Fatal(err)
	}
	m := d.Model()
	if _, err := m.AddInsert("SYM", 5, 5, 0, m.WithAttribs(map[string]string{"SIZE": "42"})); err != nil {
		t.Fatalf("AddInsert: %v", err)
	}
	nd := facadeSaveReload(t, d)
	inserts := nd.SelectByType("INSERT")
	if len(inserts) != 1 {
		t.Fatalf("inserts after reload = %d, want 1", len(inserts))
	}
	ni, ok := inserts[0].(*entity.Insert)
	if !ok {
		t.Fatal("reloaded INSERT has wrong type")
	}
	na := ni.GetAttribute("SIZE")
	if na == nil {
		t.Fatal("attribute lost after reload (not linked to insert)")
	}
	if na.GetText() != "42" {
		t.Fatalf("reloaded attribute = %q, want 42", na.GetText())
	}
}

func TestFacadeLayoutSaveReload(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	sheet, err := d.NewLayout("Sheet1")
	if err != nil {
		t.Fatalf("NewLayout: %v", err)
	}
	if _, err := sheet.AddViewport(); err != nil {
		t.Fatal(err)
	}
	if _, err := sheet.AddCircle(0, 0, 0, 50); err != nil {
		t.Fatal(err)
	}
	if err := d.RenameLayout("Sheet1", "Sheet2"); err != nil {
		t.Fatal(err)
	}
	nd := facadeSaveReload(t, d)
	if _, err := nd.GetBlock("Sheet2"); err != nil {
		t.Fatalf("layout block after reload: %v", err)
	}
	rec, ok := nd.BlockRecordHandle("Sheet2")
	if !ok {
		t.Fatal("layout BLOCK_RECORD missing after reload")
	}
	ownerCount := 0
	for _, e := range nd.Entities() {
		owner, ok := e.(ownerGetter)
		if ok && owner.Owner() != nil && owner.Owner().Handle() == rec {
			ownerCount++
		}
	}
	if ownerCount != 2 {
		t.Fatalf("%d entities owned by layout record, want 2", ownerCount)
	}
}

// TestFacadeEndToEndAuthoring exercises the full authoring flow in one pass:
// new drawing, typed entities with options, block definition with ATTDEFs,
// INSERT with ATTRIBs, new layout with entities, group membership, save,
// reload, and verification of the reloaded document.
func TestFacadeEndToEndAuthoring(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	m := d.Model()

	line, err := m.AddLine(0, 0, 0, 100, 0, 0, d.WithLayer("WALLS"))
	if err != nil {
		t.Fatalf("AddLine: %v", err)
	}
	circle, err := m.AddCircle(50, 25, 0, 25, d.WithLayer("WALLS"))
	if err != nil {
		t.Fatalf("AddCircle: %v", err)
	}
	if _, err := m.AddText(10, 10, 0, 5, "HELLO", d.WithLayer("TEXT")); err != nil {
		t.Fatalf("AddText: %v", err)
	}

	blk, err := d.NewBlock("TITLEBLOCK", "title block")
	if err != nil {
		t.Fatalf("NewBlock: %v", err)
	}
	bs := d.BlockSpace(blk)
	if _, err := bs.AddLine(0, 0, 0, 120, 0, 0); err != nil {
		t.Fatalf("block AddLine: %v", err)
	}
	if _, err := bs.AddText(2, 2, 0, 3.5, "DRAWING TITLE"); err != nil {
		t.Fatalf("block AddText: %v", err)
	}
	if _, err := bs.AddAttdef("SIZE", "Sheet size", "A3", 100, 2, 0, 3.5); err != nil {
		t.Fatalf("block AddAttdef: %v", err)
	}

	if _, err := m.AddInsert("TITLEBLOCK", 200, 0, 0, m.WithAttribs(map[string]string{"SIZE": "A1"})); err != nil {
		t.Fatalf("AddInsert: %v", err)
	}

	sheet, err := d.NewLayout("Sheet1")
	if err != nil {
		t.Fatalf("NewLayout: %v", err)
	}
	if err := d.SetActiveLayout("Sheet1"); err != nil {
		t.Fatalf("SetActiveLayout: %v", err)
	}
	if _, err := sheet.AddViewport(); err != nil {
		t.Fatalf("AddViewport: %v", err)
	}
	if _, err := sheet.AddCircle(0, 0, 0, 50); err != nil {
		t.Fatalf("layout AddCircle: %v", err)
	}
	if err := d.SetActiveLayout("Model"); err != nil {
		t.Fatalf("SetActiveLayout Model: %v", err)
	}

	if _, err := d.Group("GA", "wall layer set", line, circle); err != nil {
		t.Fatalf("Group: %v", err)
	}

	nd := facadeSaveReload(t, d)

	// Group member handles materialize during save; membership references
	// the two wall entities by handle.
	if members, err := d.GroupMembers("GA"); err != nil {
		t.Fatalf("GroupMembers: %v", err)
	} else if len(members) != 2 {
		t.Fatalf("group members = %d, want 2", len(members))
	}

	// Model space: three plain entities plus INSERT (ATTRIB is linked to it).
	modelCount := 0
	modelRec, ok := nd.BlockRecordHandle("*Model_Space")
	if !ok {
		t.Fatal("model space BLOCK_RECORD missing after reload")
	}
	var inserts []*entity.Insert
	for _, e := range nd.Entities() {
		owner, ok := e.(ownerGetter)
		if !ok || owner.Owner() == nil || owner.Owner().Handle() != modelRec {
			continue
		}
		modelCount++
		if ins, ok := e.(*entity.Insert); ok {
			inserts = append(inserts, ins)
		}
	}
	if modelCount != 4 {
		t.Fatalf("model-space entities after reload = %d, want 4", modelCount)
	}
	if len(inserts) != 1 {
		t.Fatalf("inserts after reload = %d, want 1", len(inserts))
	}
	if attr := inserts[0].GetAttribute("SIZE"); attr == nil || attr.GetText() != "A1" {
		t.Fatalf("insert attribute after reload: %+v", attr)
	}

	// Wall layer selection survives reload.
	walls := nd.SelectByLayer("WALLS")
	if len(walls) != 2 {
		t.Fatalf("WALLS entities after reload = %d, want 2", len(walls))
	}

	// Block definition keeps its content and record ownership.
	nb, err := nd.GetBlock("TITLEBLOCK")
	if err != nil {
		t.Fatalf("block after reload: %v", err)
	}
	if len(nb.Entities()) != 3 {
		t.Fatalf("block entities after reload = %d, want 3", len(nb.Entities()))
	}
	blkRec, ok := nd.BlockRecordHandle("TITLEBLOCK")
	if !ok {
		t.Fatal("TITLEBLOCK record missing after reload")
	}
	for _, be := range nb.Entities() {
		var e entity.Entity
		switch v := be.(type) {
		case interface{ Unwrap() entity.Entity }:
			e = v.Unwrap()
		case entity.Entity:
			e = v
		}
		if e == nil {
			t.Fatal("block entity does not expose entity.Entity")
		}
		owner, ok := e.(ownerGetter)
		if !ok || owner.Owner() == nil || owner.Owner().Handle() != blkRec {
			t.Fatalf("block entity owner = %v, want record %s", owner, blkRec)
		}
	}

	// Layout sheet keeps block, record, and entities.
	if _, err := nd.GetBlock("Sheet1"); err != nil {
		t.Fatalf("layout block after reload: %v", err)
	}
	sheetRec, ok := nd.BlockRecordHandle("Sheet1")
	if !ok {
		t.Fatal("Sheet1 record missing after reload")
	}
	sheetCount := 0
	for _, e := range nd.Entities() {
		owner, ok := e.(ownerGetter)
		if ok && owner.Owner() != nil && owner.Owner().Handle() == sheetRec {
			sheetCount++
		}
	}
	if sheetCount != 2 {
		t.Fatalf("Sheet1 entities after reload = %d, want 2", sheetCount)
	}

	// Handles remain unique across the whole reloaded document.
	seen := map[string]bool{}
	for _, e := range nd.Entities() {
		h := e.Handle()
		if h == "" || seen[h] {
			t.Fatalf("duplicate/empty handle %q after reload", h)
		}
		seen[h] = true
	}
	for _, e := range nb.Entities() {
		h := e.Handle()
		if h == "" || seen[h] {
			t.Fatalf("block entity handle %q collides with entity section", h)
		}
		seen[h] = true
	}
}
