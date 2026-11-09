package dxf

import (
	"testing"

	"github.com/edanko/dxf/entity"
)

func TestInsertBBoxContent(t *testing.T) {
	blockDef := "0\nSECTION\n2\nBLOCKS\n0\nBLOCK\n5\nB1\n100\nAcDbEntity\n8\n0\n100\nAcDbBlockBegin\n2\nMYBLK\n70\n" + f(0) + "\n10\n" + f(0.0) + "\n20\n" + f(0.0) + "\n30\n" + f(0.0) + "\n3\nMYBLK\n0\nLINE\n5\nL1\n100\nAcDbEntity\n8\n0\n100\nAcDbLine\n10\n" + f(0.0) + "\n20\n" + f(0.0) + "\n30\n" + f(0.0) + "\n11\n" + f(10.0) + "\n21\n" + f(0.0) + "\n31\n" + f(0.0) + "\n0\nLINE\n5\nL2\n100\nAcDbEntity\n8\n0\n100\nAcDbLine\n10\n" + f(0.0) + "\n20\n" + f(0.0) + "\n30\n" + f(0.0) + "\n11\n" + f(0.0) + "\n21\n" + f(6.0) + "\n31\n" + f(0.0) + "\n0\nENDBLK\n5\nB2\n100\nAcDbEntity\n8\n0\n100\nAcDbBlockEnd\n0\nENDSEC\n"
	ent := "0\nINSERT\n5\nI1\n100\nAcDbEntity\n8\n0\n100\nAcDbInsert\n2\nMYBLK\n10\n" + f(10.0) + "\n20\n" + f(20.0) + "\n30\n" + f(0.0) + "\n41\n" + f(2.0) + "\n42\n" + f(2.0) + "\n43\n" + f(1.0) + "\n"
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
	mins, maxs := ins.BBox()
	// content x in [0,10] scaled 2 -> [0,20] + insert x 10 -> [10,30]
	// content y in [0,6] scaled 2 -> [0,12] + insert y 20 -> [20,32]
	if mins[0] != 10 || mins[1] != 20 {
		t.Errorf("mins = %v want (10,20)", mins)
	}
	if maxs[0] != 30 || maxs[1] != 32 {
		t.Errorf("maxs = %v want (30,32)", maxs)
	}
}

func TestBBoxFixturesCorpus(t *testing.T) {
	files := []string{"testdata/columns_R2018.dxf", "testdata/clockwise_arcs_hatch.dxf"}
	for _, name := range files {
		d, err := FromFile(name)
		if err != nil {
			t.Fatal(err)
		}
		n := 0
		for _, e := range d.Entities() {
			bb, ok := e.(entity.Bounder)
			if !ok {
				continue
			}
			mins, maxs := bb.BBox()
			if len(mins) < 2 || len(maxs) < 2 {
				t.Fatalf("%s %s: malformed bbox", name, e.DXFType())
			}
			for k := 0; k < 2; k++ {
				if mins[k] > maxs[k] {
					t.Errorf("%s %s: min > max (%v %v)", name, e.DXFType(), mins, maxs)
				}
			}
			n++
		}
		if n == 0 {
			t.Errorf("%s: no bounding boxes", name)
		}
	}
}
