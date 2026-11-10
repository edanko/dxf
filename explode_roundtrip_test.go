package dxf

import (
	"testing"

	"github.com/edanko/dxf/entity"
)

func TestExplodeInsert(t *testing.T) {
	blockDef := "0\nSECTION\n2\nBLOCKS\n0\nBLOCK\n5\nB1\n100\nAcDbEntity\n8\n0\n100\nAcDbBlockBegin\n2\nMYBLK\n70\n" + f(0) + "\n10\n" + f(0.0) + "\n20\n" + f(0.0) + "\n30\n" + f(0.0) + "\n3\nMYBLK\n0\nLINE\n5\nL1\n100\nAcDbEntity\n8\n0\n100\nAcDbLine\n10\n" + f(0.0) + "\n20\n" + f(0.0) + "\n30\n" + f(0.0) + "\n11\n" + f(1.0) + "\n21\n" + f(0.0) + "\n31\n" + f(0.0) + "\n0\nENDBLK\n5\nB2\n100\nAcDbEntity\n8\n0\n100\nAcDbBlockEnd\n0\nENDSEC\n"
	ent := "0\nINSERT\n5\nI1\n100\nAcDbEntity\n8\n0\n100\nAcDbInsert\n2\nMYBLK\n10\n" + f(10.0) + "\n20\n" + f(20.0) + "\n30\n" + f(0.0) + "\n41\n" + f(2.0) + "\n42\n" + f(2.0) + "\n43\n" + f(1.0) + "\n50\n" + f(90.0) + "\n"
	src := blockDef + "0\nSECTION\n2\nENTITIES\n" + ent + "0\nENDSEC\n0\nEOF\n"
	d, err := FromStringData(src)
	if err != nil {
		t.Fatal(err)
	}
	var ins *entity.Insert
	for _, e := range d.Entities() {
		if i, ok := e.(*entity.Insert); ok {
			ins = i
		}
	}
	if ins == nil {
		t.Fatal("no insert")
	}
	parts, err := ins.Explode()
	if err != nil {
		t.Fatal(err)
	}
	if len(parts) != 1 {
		t.Fatalf("parts = %d", len(parts))
	}
	ln, ok := parts[0].(*entity.Line)
	if !ok {
		t.Fatalf("part type %T", parts[0])
	}
	// (0,0)->(1,0) scaled x2, rotated 90 deg, translated to (10,20):
	// start = (10,20), end = (10, 22)
	if ln.Start[0] != 10.0 || ln.Start[1] != 20.0 {
		t.Errorf("start = %v", ln.Start)
	}
	if ln.End[0] != 10.0 || ln.End[1] != 22.0 {
		t.Errorf("end = %v (want 10,22)", ln.End)
	}
}

func TestExplodePolyline(t *testing.T) {
	body := "0\nPOLYLINE\n5\nP1\n100\nAcDbEntity\n8\n0\n100\nAcDb2dPolyline\n70\n" + f(1) + "\n66\n" + f(1) + "\n0\nVERTEX\n5\nV1\n100\nAcDbVertex\n100\nAcDb2dVertex\n10\n" + f(0.0) + "\n20\n" + f(0.0) + "\n30\n" + f(0.0) + "\n0\nVERTEX\n5\nV2\n100\nAcDbVertex\n100\nAcDb2dVertex\n10\n" + f(4.0) + "\n20\n" + f(0.0) + "\n30\n" + f(0.0) + "\n0\nVERTEX\n5\nV3\n100\nAcDbVertex\n100\nAcDb2dVertex\n10\n" + f(4.0) + "\n20\n" + f(3.0) + "\n30\n" + f(0.0) + "\n0\nSEQEND\n5\nSE1\n"
	e, _ := loadOne(t, body)
	pl, ok := e.(*entity.Polyline)
	if !ok {
		t.Fatalf("type %T", e)
	}
	parts, err := pl.Explode()
	if err != nil {
		t.Fatal(err)
	}
	if len(parts) != 3 { // 2 open segments + closing segment (closed polyline)
		t.Fatalf("parts = %d", len(parts))
	}
	for _, p := range parts {
		if _, ok := p.(*entity.Line); !ok {
			t.Fatalf("part type %T", p)
		}
	}
}

func TestExplodeHatch(t *testing.T) {
	body := "0\nHATCH\n5\nH1\n100\nAcDbEntity\n8\n0\n100\nAcDbHatch\n70\n" + f(0) + "\n71\n" + f(1) + "\n91\n" + f(1) + "\n92\n" + f(2) + "\n72\n" + f(0) + "\n73\n" + f(1) + "\n93\n" + f(4) + "\n10\n" + f(0.0) + "\n20\n" + f(0.0) + "\n10\n" + f(5.0) + "\n20\n" + f(0.0) + "\n10\n" + f(5.0) + "\n20\n" + f(5.0) + "\n10\n" + f(0.0) + "\n20\n" + f(5.0) + "\n78\n" + f(0) + "\n"
	e, _ := loadOne(t, body)
	h := e.(*entity.Hatch)
	parts, err := h.Explode()
	if err != nil {
		t.Fatal(err)
	}
	if len(parts) != 1 {
		t.Fatalf("parts = %d", len(parts))
	}
	lw, ok := parts[0].(*entity.LwPolyline)
	if !ok {
		t.Fatalf("part type %T", parts[0])
	}
	if !lw.Closed || lw.Num != 4 {
		t.Errorf("closed=%v num=%d", lw.Closed, lw.Num)
	}
}

func TestExplodeLeader(t *testing.T) {
	body := "0\nLEADER\n5\nL1\n100\nAcDbEntity\n8\n0\n100\nAcDbLeader\n71\n" + f(0) + "\n72\n" + f(0) + "\n10\n" + f(0.0) + "\n20\n" + f(0.0) + "\n30\n" + f(0.0) + "\n10\n" + f(5.0) + "\n20\n" + f(5.0) + "\n30\n" + f(0.0) + "\n10\n" + f(10.0) + "\n20\n" + f(5.0) + "\n30\n" + f(0.0) + "\n"
	e, _ := loadOne(t, body)
	ld := e.(*entity.Leader)
	parts, err := ld.Explode()
	if err != nil {
		t.Fatal(err)
	}
	if len(parts) != 2 {
		t.Fatalf("parts = %d", len(parts))
	}
	l0 := parts[0].(*entity.Line)
	if l0.Start[0] != 0.0 || l0.End[0] != 5.0 {
		t.Errorf("segment0 = %v -> %v", l0.Start, l0.End)
	}
}

func TestExplodeDimension(t *testing.T) {
	body := "0\nDIMENSION\n5\nD1\n100\nAcDbEntity\n8\n0\n100\nAcDbDimension\n10\n" + f(5.0) + "\n20\n" + f(10.0) + "\n30\n" + f(0.0) + "\n11\n" + f(5.0) + "\n21\n" + f(9.0) + "\n31\n" + f(0.0) + "\n70\n" + f(1) + "\n3\nStandard\n71\n" + f(5) + "\n72\n" + f(1) + "\n41\n" + f(1.0) + "\n42\n" + f(10.0) + "\n100\nAcDbAlignedDimension\n13\n" + f(0.0) + "\n23\n" + f(0.0) + "\n33\n" + f(0.0) + "\n14\n" + f(10.0) + "\n24\n" + f(0.0) + "\n34\n" + f(0.0) + "\n"
	e, _ := loadOne(t, body)
	dim := e.(*entity.Dimension)
	parts, err := dim.Explode()
	if err != nil {
		t.Fatal(err)
	}
	if len(parts) != 6 { // two extension lines + dimension line + two arrow ticks + measured text
		t.Fatalf("parts = %d", len(parts))
	}
	for _, p := range parts[:5] {
		if _, ok := p.(*entity.Line); !ok {
			t.Fatalf("part type %T", p)
		}
	}
	if txt, ok := parts[5].(*entity.Text); !ok || txt.Value != "10.00" {
		t.Fatalf("text part = %#v", parts[5])
	}
	// Arrow ticks sit at the extension points along the dimension line.
	if tick := parts[3].(*entity.Line); tick.End[0] != 0 || tick.End[1] != 0 {
		t.Errorf("arrow 1 end = %v", tick.End)
	}
	if tick := parts[4].(*entity.Line); tick.Start[0] != 10 || tick.Start[1] != 0 {
		t.Errorf("arrow 2 start = %v", tick.Start)
	}
}

func TestExplodeMLeader(t *testing.T) {
	body := "0\nMULTILEADER\n5\nM1\n100\nAcDbEntity\n8\n0\n100\nAcDbMLeader\n170\n" + f(1) + "\n171\n" + f(0) + "\n172\n" + f(9) + "\n41\n" + f(1.0) + "\n271\n" + f(1) + "\n271\n" + f(0) + "\n10\n" + f(0.0) + "\n20\n" + f(0.0) + "\n30\n" + f(0.0) + "\n10\n" + f(5.0) + "\n20\n" + f(5.0) + "\n30\n" + f(0.0) + "\n10\n" + f(10.0) + "\n20\n" + f(5.0) + "\n30\n" + f(0.0) + "\n175\n" + f(1) + "\n304\nHello\n"
	e, _ := loadOne(t, body)
	ml := e.(*entity.MLeader)
	parts, err := ml.Explode()
	if err != nil {
		t.Fatal(err)
	}
	if len(parts) != 2 {
		t.Fatalf("parts = %d", len(parts))
	}
	l1 := parts[1].(*entity.Line)
	if l1.Start[0] != 5.0 || l1.End[0] != 10.0 {
		t.Errorf("segment1 = %v -> %v", l1.Start, l1.End)
	}
}
