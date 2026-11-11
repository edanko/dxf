package drawing

import (
	"testing"

	"github.com/edanko/dxf/entity"
)

func newTestInsert(t *testing.T, _ *Drawing, blockName string) *entity.Insert {
	t.Helper()
	ins := entity.NewInsert()
	ins.SetBlockName(blockName)
	return ins
}

func TestPurgeUnusedBlocksReport(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatal(err)
	}

	// PARENT is used from model space and nests CHILD, which must stay even
	// though nothing inserts CHILD directly.
	parent, err := d.NewBlock("PARENT", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.NewBlock("CHILD", ""); err != nil {
		t.Fatal(err)
	}
	parent.AddEntity(asBlockEntity(newTestInsert(t, d, "CHILD")))

	// ORPHAN is unused and only nests NESTED_ORPHAN, so both are removed.
	orphan, err := d.NewBlock("ORPHAN", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.NewBlock("NESTED_ORPHAN", ""); err != nil {
		t.Fatal(err)
	}
	orphan.AddEntity(asBlockEntity(newTestInsert(t, d, "NESTED_ORPHAN")))

	// XREF-style block: externally dependent flag keeps it.
	xref, err := d.NewBlock("XREF", "")
	if err != nil {
		t.Fatal(err)
	}
	xref.Flag = 4

	// Layout-owned block: created by NewLayout, unused by any INSERT.
	layoutSpace, err := d.NewLayout("Layout1")
	if err != nil {
		t.Fatal(err)
	}

	if err := d.Model().Add(newTestInsert(t, d, "PARENT")); err != nil {
		t.Fatal(err)
	}
	if err := layoutSpace.Add(newTestInsert(t, d, "PARENT")); err != nil {
		t.Fatal(err)
	}

	report := d.PurgeUnusedBlocksReport()
	if report == nil {
		t.Fatal("expected purge report")
	}
	if got, want := report.RemovedCount(), 2; got != want {
		t.Fatalf("removed %d blocks, want %d: %+v", got, want, report.Removed)
	}

	removed := make(map[string]string)
	for _, p := range report.Removed {
		removed[p.Name] = p.Reason
	}
	if _, ok := removed["ORPHAN"]; !ok {
		t.Errorf("ORPHAN not in removed report: %+v", report.Removed)
	}
	if _, ok := removed["NESTED_ORPHAN"]; !ok {
		t.Errorf("NESTED_ORPHAN not in removed report: %+v", report.Removed)
	}
	for name := range removed {
		if name != "ORPHAN" && name != "NESTED_ORPHAN" {
			t.Errorf("unexpected removed block %q", name)
		}
		if removed[name] == "" {
			t.Errorf("block %q removed without a reason", name)
		}
	}

	for _, keep := range []string{"*Model_Space", "*Paper_Space", "*Paper_Space0", "PARENT", "CHILD", "XREF", "Layout1"} {
		if _, err := d.GetBlock(keep); err != nil {
			t.Errorf("block %q should be kept: %v", keep, err)
		}
	}

	// Block records are cleaned up along with the definitions.
	if t2 := d.blockRecordTable(); t2 != nil {
		if _, err := t2.Lookup("ORPHAN"); err == nil {
			t.Error("block record for ORPHAN still present after purge")
		}
		if _, err := t2.Lookup("PARENT"); err != nil {
			t.Error("block record for PARENT missing after purge")
		}
	}
}

func TestPurgeUnusedBlocksIntWrapper(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.NewBlock("UNUSED", ""); err != nil {
		t.Fatal(err)
	}
	if got := d.PurgeUnusedBlocks(); got != 1 {
		t.Fatalf("PurgeUnusedBlocks() = %d, want 1", got)
	}
	if got := d.PurgeUnusedBlocks(); got != 0 {
		t.Fatalf("second PurgeUnusedBlocks() = %d, want 0", got)
	}
}

func TestPurgeUnusedBlocksUsedFromPaperLayoutOnly(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.NewBlock("TITLE", ""); err != nil {
		t.Fatal(err)
	}
	layoutSpace, err := d.NewLayout("Sheet1")
	if err != nil {
		t.Fatal(err)
	}
	if err := layoutSpace.Add(newTestInsert(t, d, "TITLE")); err != nil {
		t.Fatal(err)
	}

	report := d.PurgeUnusedBlocksReport()
	if report.RemovedCount() != 0 {
		t.Fatalf("expected no removals, got %+v", report.Removed)
	}
	if _, err := d.GetBlock("TITLE"); err != nil {
		t.Fatalf("TITLE should be kept: %v", err)
	}
}
