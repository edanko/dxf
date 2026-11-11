package drawing

import (
	"testing"

	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/table"
)

func setRecordXData(t *testing.T, d *Drawing, blockName string, tags []table.EntryTag) {
	t.Helper()
	rec, ok := blockRecordByName(d.blockRecordTable(), blockName)
	if !ok {
		t.Fatalf("block record %q not found", blockName)
	}
	rec.SetRawTags(tags)
}

func TestResolveDynamicBlockDirect(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.NewBlock("DYN", ""); err != nil {
		t.Fatal(err)
	}
	setRecordXData(t, d, "DYN", []table.EntryTag{
		{Code: 1001, Value: DynamicBlockGUIDAppID},
		{Code: 1000, Value: "{11111111-2222-3333-4444-555555555555}"},
		{Code: 1001, Value: DynamicBlockTrueNameAppID},
		{Code: 1000, Value: "DYN"},
	})
	ins := entity.NewInsert()
	ins.SetBlockName("DYN")
	info, err := d.ResolveDynamicBlock(ins)
	if err != nil {
		t.Fatal(err)
	}
	if info == nil {
		t.Fatal("expected dynamic block info")
	}
	if info.DefinitionName != "DYN" || info.TrueName != "DYN" || info.Anonymous {
		t.Fatalf("unexpected info: %+v", info)
	}
	if info.GUID != "{11111111-2222-3333-4444-555555555555}" {
		t.Fatalf("GUID = %q", info.GUID)
	}
	if info.BlockRecordHandle == "" {
		t.Fatal("empty block record handle")
	}
}

func TestResolveDynamicBlockAnonymous(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.NewBlock("DYN", ""); err != nil {
		t.Fatal(err)
	}
	setRecordXData(t, d, "DYN", []table.EntryTag{
		{Code: 1001, Value: DynamicBlockGUIDAppID},
		{Code: 1000, Value: "{AAAAAAAA-BBBB-CCCC-DDDD-EEEEEEEEEEEE}"},
	})
	defRec, _ := blockRecordByName(d.blockRecordTable(), "DYN")
	if _, err := d.NewBlock("*U1", ""); err != nil {
		t.Fatal(err)
	}
	setRecordXData(t, d, "*U1", []table.EntryTag{
		{Code: 1001, Value: DynamicBlockRepBTagAppID},
		{Code: 1000, Value: "named"},
		{Code: 1005, Value: defRec.Handle()},
	})
	ins := entity.NewInsert()
	ins.SetBlockName("*U1")
	info, err := d.ResolveDynamicBlock(ins)
	if err != nil {
		t.Fatal(err)
	}
	if info == nil {
		t.Fatal("expected dynamic block info")
	}
	if !info.Anonymous || info.DefinitionName != "DYN" || info.BlockRecordHandle != defRec.Handle() {
		t.Fatalf("unexpected info: %+v", info)
	}
	if info.GUID != "{AAAAAAAA-BBBB-CCCC-DDDD-EEEEEEEEEEEE}" {
		t.Fatalf("GUID = %q", info.GUID)
	}
}

func TestResolveDynamicBlockOrdinary(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.NewBlock("PLAIN", ""); err != nil {
		t.Fatal(err)
	}
	ins := entity.NewInsert()
	ins.SetBlockName("PLAIN")
	info, err := d.ResolveDynamicBlock(ins)
	if err != nil {
		t.Fatal(err)
	}
	if info != nil {
		t.Fatalf("ordinary block resolved as dynamic: %+v", info)
	}
}

func TestResolveDynamicBlockLostReference(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.NewBlock("*U2", ""); err != nil {
		t.Fatal(err)
	}
	setRecordXData(t, d, "*U2", []table.EntryTag{
		{Code: 1001, Value: DynamicBlockRepBTagAppID},
		{Code: 1005, Value: "DEAD"},
	})
	ins := entity.NewInsert()
	ins.SetBlockName("*U2")
	if _, err := d.ResolveDynamicBlock(ins); err == nil {
		t.Fatal("expected lost definition error")
	}
}
