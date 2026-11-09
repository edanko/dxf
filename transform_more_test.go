package dxf

import (
	"math"
	"testing"

	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
)

func TestTransformDimensionPoints(t *testing.T) {
	body := "0\nDIMENSION\n5\nD1\n100\nAcDbEntity\n8\n0\n100\nAcDbDimension\n10\n" + f(5.0) + "\n20\n" + f(10.0) + "\n30\n" + f(0.0) + "\n11\n" + f(5.0) + "\n21\n" + f(9.0) + "\n31\n" + f(0.0) + "\n70\n" + f(1) + "\n3\nStandard\n71\n" + f(5) + "\n72\n" + f(1) + "\n41\n" + f(1.0) + "\n42\n" + f(10.0) + "\n100\nAcDbAlignedDimension\n13\n" + f(0.0) + "\n23\n" + f(0.0) + "\n33\n" + f(0.0) + "\n14\n" + f(10.0) + "\n24\n" + f(0.0) + "\n34\n" + f(0.0) + "\n"
	e, _ := loadOne(t, body)
	dim := e.(*entity.Dimension)
	trl := dxfmath.Translation(3, 4, 0)
	if err := dim.Transform(trl); err != nil {
		t.Fatal(err)
	}
	parts, err := dim.Explode()
	if err != nil {
		t.Fatal(err)
	}
	l0 := parts[0].(*entity.Line)
	if math.Abs(l0.Start[0]-3) > 1e-9 || math.Abs(l0.Start[1]-4) > 1e-9 {
		t.Errorf("ext line start = %v want (3,4)", l0.Start[:2])
	}
	l1 := parts[1].(*entity.Line)
	if math.Abs(l1.End[0]-8) > 1e-9 || math.Abs(l1.End[1]-14) > 1e-9 {
		t.Errorf("def point = %v want (8,14)", l1.End[:2])
	}
}

func TestTransformMLeaderVertices(t *testing.T) {
	body := "0\nMULTILEADER\n5\nM1\n100\nAcDbEntity\n8\n0\n100\nAcDbMLeader\n170\n" + f(1) + "\n171\n" + f(0) + "\n172\n" + f(9) + "\n41\n" + f(1.0) + "\n271\n" + f(1) + "\n271\n" + f(0) + "\n10\n" + f(0.0) + "\n20\n" + f(0.0) + "\n30\n" + f(0.0) + "\n10\n" + f(5.0) + "\n20\n" + f(5.0) + "\n30\n" + f(0.0) + "\n175\n" + f(1) + "\n304\nHello\n"
	e, _ := loadOne(t, body)
	ml := e.(*entity.MLeader)
	trl := dxfmath.Translation(2, 3, 0)
	if err := ml.Transform(trl); err != nil {
		t.Fatal(err)
	}
	parts, _ := ml.Explode()
	l0 := parts[0].(*entity.Line)
	if math.Abs(l0.Start[0]-2) > 1e-9 || math.Abs(l0.End[0]-7) > 1e-9 {
		t.Errorf("vertices after translate = %v -> %v", l0.Start, l0.End)
	}
}

func TestTransformLeaderPoints(t *testing.T) {
	body := "0\nLEADER\n5\nL1\n100\nAcDbEntity\n8\n0\n100\nAcDbLeader\n71\n" + f(0) + "\n72\n" + f(0) + "\n10\n" + f(0.0) + "\n20\n" + f(0.0) + "\n30\n" + f(0.0) + "\n10\n" + f(5.0) + "\n20\n" + f(5.0) + "\n30\n" + f(0.0) + "\n"
	e, _ := loadOne(t, body)
	ld := e.(*entity.Leader)
	trl := dxfmath.Translation(1, 1, 0)
	if err := ld.Transform(trl); err != nil {
		t.Fatal(err)
	}
	pts := ld.GetPoints()
	if math.Abs(pts[0][0]-1) > 1e-9 || math.Abs(pts[1][1]-6) > 1e-9 {
		t.Errorf("points = %v", pts)
	}
}
