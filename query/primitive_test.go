package query

import (
	"math"
	"testing"

	"github.com/edanko/dxf/block"
	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
)

func TestMakePrimitiveSupported(t *testing.T) {
	line := entity.NewLine()
	line.Start = []float64{1, 2, 0}
	line.End = []float64{4, 6, 0}
	p := MakePrimitive(line)
	if p.Kind != PrimitiveLine || p.Path == nil || len(p.Vertices) != 2 {
		t.Fatalf("line primitive = %+v", p)
	}

	arc := entity.NewArc(nil)
	arc.Center = []float64{0, 0, 0}
	arc.Radius = 5
	arc.Angle = []float64{0, 90}
	p = MakePrimitive(arc)
	if p.Kind != PrimitiveArc || p.Path == nil {
		t.Fatalf("arc primitive = %+v", p)
	}
	start, end := p.Path.Start(), p.Path.End()
	if math.Abs(start.X()-5) > 1e-9 || math.Abs(start.Y()) > 1e-9 {
		t.Fatalf("arc start = %v", start)
	}
	if math.Abs(end.X()) > 1e-9 || math.Abs(end.Y()-5) > 1e-9 {
		t.Fatalf("arc end = %v", end)
	}

	point := entity.NewPoint(3, 4, 5)
	p = MakePrimitive(point)
	if p.Kind != PrimitivePoint || len(p.Vertices) != 1 || p.Path != nil {
		t.Fatalf("point primitive = %+v", p)
	}

	unknown := entity.NewRetainedRecord("MYSTERY", nil)
	p = MakePrimitive(unknown)
	if p.Kind != PrimitiveUnknown || len(p.Diagnostics) == 0 {
		t.Fatalf("unknown diagnostic = %+v", p)
	}
}

func TestRecursiveDecomposeNestedInsertAndCycle(t *testing.T) {
	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{1, 0, 0}
	text := entity.NewText()
	text.Value = "ABC"
	text.Height = 1

	inner := block.NewBlock("INNER", "")
	inner.AddEntity(&testBlockEntity{Entity: line})
	inner.AddEntity(&testBlockEntity{Entity: text})

	outer := block.NewBlock("OUTER", "")
	ins := entity.NewInsert()
	ins.SetBlockName("INNER")
	ins.SetInsertPoint(dxfmath.NewVec3(10, 5, 0))
	outer.AddEntity(&testBlockEntity{Entity: ins})

	root := entity.NewInsert()
	root.SetBlockName("OUTER")
	lookup := func(name string) *block.Block {
		switch name {
		case "INNER":
			return inner
		case "OUTER":
			return outer
		}
		return nil
	}
	d := RecursiveDecompose([]entity.Entity{root}, lookup)
	if len(d.Primitives) != 2 {
		t.Fatalf("primitives = %d want 2: %+v", len(d.Primitives), d)
	}
	var linePrim, textPrim *Primitive
	for i := range d.Primitives {
		switch d.Primitives[i].Kind {
		case PrimitiveLine:
			linePrim = &d.Primitives[i]
		case PrimitiveText:
			textPrim = &d.Primitives[i]
		}
	}
	if linePrim == nil || textPrim == nil {
		t.Fatalf("missing leaves: %+v", d.Primitives)
	}
	if linePrim.Vertices[0].X() != 10 || linePrim.Vertices[0].Y() != 5 {
		t.Fatalf("line not transformed: %+v", linePrim.Vertices)
	}
	if textPrim.Path == nil {
		t.Fatal("text has no path")
	}

	cycleBlock := block.NewBlock("SELF", "")
	self := entity.NewInsert()
	self.SetBlockName("SELF")
	cycleBlock.AddEntity(&testBlockEntity{Entity: self})
	d = RecursiveDecompose([]entity.Entity{self}, func(name string) *block.Block { return cycleBlock })
	if len(d.Cycles) == 0 {
		t.Fatalf("cycle not reported: %+v", d)
	}
}

func TestPrimitiveConversions(t *testing.T) {
	mesh := entity.NewMesh()
	mesh.SetVertices([]dxfmath.Vec3{dxfmath.NewVec3(0, 0, 0), dxfmath.NewVec3(1, 0, 0), dxfmath.NewVec3(0, 1, 0)})
	mesh.SetFaces([][]int{{0, 1, 2}})
	prims := []Primitive{MakePrimitive(mesh), MakePrimitive(entity.NewPoint(1, 2, 3))}
	meshes := ToMeshes(prims)
	if len(meshes) != 1 || len(meshes[0].Vertices) != 3 || len(meshes[0].Faces) != 1 {
		t.Fatalf("meshes = %+v", meshes)
	}
	if got := ToVertices(prims); len(got) != 4 {
		t.Fatalf("vertices = %d", len(got))
	}
	pointPrim := prims[1]
	if pointPrim.Path != nil {
		t.Fatal("POINT must not produce a path")
	}
	arc := entity.NewArc(nil)
	arc.Center = []float64{0, 0, 0}
	arc.Radius = 2
	arc.Angle = []float64{0, 180}
	paths := ToPaths([]Primitive{MakePrimitive(arc)})
	if len(paths) != 1 {
		t.Fatalf("arc paths = %d", len(paths))
	}
	s, e := paths[0].Start(), paths[0].End()
	if math.Abs(s.Distance(dxfmath.NewVec2(2, 0))) > 1e-9 || math.Abs(e.Distance(dxfmath.NewVec2(-2, 0))) > 1e-9 {
		t.Fatalf("arc endpoints %v %v", s, e)
	}
}
