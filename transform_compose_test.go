package dxf

import (
	"errors"
	"math"
	"testing"

	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
)

// spec scenario: translate then rotate a line; the composed matrix applies in
// application order.
func TestTransformLineTranslateRotate(t *testing.T) {
	ln := entity.NewLine()
	ln.Start = []float64{1, 0, 0}
	ln.End = []float64{2, 0, 0}

	trl := dxfmath.Translation(10, 0, 0)
	rot := dxfmath.RotationZ(90 * math.Pi / 180)
	m := rot.Mul(trl) // translate first, then rotate
	if err := ln.Transform(m); err != nil {
		t.Fatal(err)
	}
	// (1,0)+(10,0)=(11,0) rotated 90deg -> (0,11); (2,0)->(12,0)->(0,12)
	if math.Abs(ln.Start[0]) > 1e-9 || math.Abs(ln.Start[1]-11) > 1e-9 {
		t.Errorf("start = %v", ln.Start)
	}
	if math.Abs(ln.End[0]) > 1e-9 || math.Abs(ln.End[1]-12) > 1e-9 {
		t.Errorf("end = %v", ln.End)
	}
}

func TestTransformCircleNonUniformScale(t *testing.T) {
	c := entity.NewCircle()
	c.Center = []float64{0, 0, 0}
	c.Radius = 5
	m := dxfmath.Scale(2, 3, 1)
	err := c.Transform(m)
	if err == nil {
		t.Fatal("expected error for non-uniform scale")
	}
	if !errors.Is(err, entity.ErrNonUniformScale) {
		t.Errorf("err = %v, want ErrNonUniformScale", err)
	}
}

func TestTransformMatrixRoundTripInsert(t *testing.T) {
	ins := entity.NewInsert()
	ins.SetBlockName("B")
	ins.SetInsertPoint(dxfmath.NewVec3(1, 2, 0))
	ins.SetScaleX(2)
	ins.SetRotation(90)

	trl := dxfmath.Translation(5, 5, 0)
	if err := ins.Transform(trl); err != nil {
		t.Fatal(err)
	}
	p := ins.InsertPoint()
	if math.Abs(p.X()-6) > 1e-9 || math.Abs(p.Y()-7) > 1e-9 {
		t.Errorf("insert point = %v want (6,7)", []float64{p.X(), p.Y()})
	}
	if math.Abs(ins.Rotation()-90) > 1e-9 {
		t.Errorf("rotation = %v", ins.Rotation())
	}
}
