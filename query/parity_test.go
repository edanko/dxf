// Parity tests for the query/transform spec: expressions, spatial and xdata
// predicates, error positions, selection sets, grouping, transforms,
// disassembly, traversal, revcloud and upright helpers.
package query

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/edanko/dxf/block"
	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
	"github.com/edanko/dxf/xdata"
)

func line(handle, layer string, x1, y1, x2, y2 float64) *entity.Line {
	l := entity.NewLine()
	l.SetHandleString(handle)
	l.SetLayerName(layer)
	l.Start = []float64{x1, y1, 0}
	l.End = []float64{x2, y2, 0}
	return l
}

func TestQueryFilterByTypeAndLayer(t *testing.T) {
	ents := []entity.Entity{
		line("1", "Walls", 0, 0, 1, 1),
		line("2", "Walls", 2, 2, 3, 3),
		line("3", "Floor", 0, 0, 1, 0),
	}
	sel, err := ExecuteQuery(ents, "type == 'LINE' and layer == 'Walls'")
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(sel.Handles(), ","); got != "1,2" {
		t.Fatalf("type+layer filter: %s", got)
	}
}

func TestQueryAttributeFilter(t *testing.T) {
	ins := entity.NewInsert()
	ins.SetHandleString("10")
	ins.SetBlockName("SOFA")
	a := entity.NewAttrib()
	a.SetTag("SIZE")
	a.SetText("100")
	ins.AddAttribute(a)
	plain := line("11", "0", 0, 0, 1, 1)

	sel, err := ExecuteQuery([]entity.Entity{ins, plain}, "attr('SIZE') == '100'")
	if err != nil {
		t.Fatal(err)
	}
	if len(sel.Handles()) != 1 || sel.Handles()[0] != "10" {
		t.Fatalf("attribute filter: %v", sel.Handles())
	}
	sel2, err := ExecuteQuery([]entity.Entity{ins, plain}, "attr('SIZE') == '200'")
	if err != nil {
		t.Fatal(err)
	}
	if sel2.Len() != 0 {
		t.Fatalf("non-matching attribute must exclude: %v", sel2.Handles())
	}
}

func TestQuerySpatialFilter(t *testing.T) {
	ents := []entity.Entity{
		line("20", "0", 0, 0, 1, 1),
		line("21", "0", 10, 10, 11, 11),
	}
	sel, err := ExecuteQuery(ents, "bbox_intersects(5, 5, 20, 20)")
	if err != nil {
		t.Fatal(err)
	}
	if len(sel.Handles()) != 1 || sel.Handles()[0] != "21" {
		t.Fatalf("bbox_intersects: %v", sel.Handles())
	}
	sel2, err := ExecuteQuery(ents, "bbox_within(0, 0, 2, 2)")
	if err != nil {
		t.Fatal(err)
	}
	if len(sel2.Handles()) != 1 || sel2.Handles()[0] != "20" {
		t.Fatalf("bbox_within: %v", sel2.Handles())
	}
}

func TestQueryXDataFilter(t *testing.T) {
	e := line("30", "0", 0, 0, 1, 1)
	e.SetXData("MYAPP", []xdata.Tag{{Code: 1000, Value: "hello"}})
	sel, err := ExecuteQuery([]entity.Entity{e, line("31", "0", 0, 0, 1, 1)}, "has_xdata('MYAPP')")
	if err != nil {
		t.Fatal(err)
	}
	if len(sel.Handles()) != 1 || sel.Handles()[0] != "30" {
		t.Fatalf("has_xdata: %v", sel.Handles())
	}
	sel2, err := ExecuteQuery([]entity.Entity{e}, "xdata_value('MYAPP', 1000) == 'hello'")
	if err != nil {
		t.Fatal(err)
	}
	if sel2.Len() != 1 {
		t.Fatalf("xdata_value: %v", sel2.Handles())
	}
}

func TestQueryErrorsReportPosition(t *testing.T) {
	if _, err := ExecuteQuery(nil, "layer =="); err == nil {
		t.Fatal("syntax error expected")
	}
	_, err := ExecuteQuery(nil, "nosuchprop == 'x' and layer == '0'")
	if err == nil {
		t.Fatal("unknown property error expected")
	}
	if !strings.Contains(err.Error(), "nosuchprop") {
		t.Fatalf("error must name the property: %v", err)
	}
	if !strings.Contains(err.Error(), "0") {
		t.Fatalf("error must carry a position: %v", err)
	}
	_, err = ExecuteQuery([]entity.Entity{line("1", "0", 0, 0, 1, 1)}, "bbox_intersects(1, 2)")
	if err == nil {
		t.Fatal("arity error expected")
	}
	if !strings.Contains(err.Error(), "bbox_intersects") {
		t.Fatalf("arity error names the function: %v", err)
	}
}

func TestSelectionSetOperations(t *testing.T) {
	a := NewSelection("1", "2", "3")
	b := NewSelection("2", "3", "4")
	if got := strings.Join(a.Union(b).Handles(), ","); got != "1,2,3,4" {
		t.Fatalf("union: %s", got)
	}
	if got := strings.Join(a.Intersect(b).Handles(), ","); got != "2,3" {
		t.Fatalf("intersection: %s", got)
	}
	if got := strings.Join(a.Difference(b).Handles(), ","); got != "1" {
		t.Fatalf("difference: %s", got)
	}
	src := &mapSource{m: map[string]entity.Entity{"1": line("1", "0", 0, 0, 1, 1), "2": line("2", "0", 0, 0, 1, 1)}}
	ents, missing := a.Resolve(src)
	if len(ents) != 2 || len(missing) != 1 || missing[0] != "3" {
		t.Fatalf("resolve: %d ents, missing %v", len(ents), missing)
	}
}

type mapSource struct{ m map[string]entity.Entity }

func (s *mapSource) GetEntity(h string) (entity.Entity, bool) { e, ok := s.m[h]; return e, ok }

func TestGroupingByLayer(t *testing.T) {
	ents := []entity.Entity{
		line("1", "A", 0, 0, 1, 1),
		line("2", "B", 0, 0, 1, 1),
		line("3", "A", 0, 0, 1, 1),
	}
	groups := GroupByAttribute(ents, "layer")
	if len(groups["A"]) != 2 || len(groups["B"]) != 1 {
		t.Fatalf("group by layer: A=%d B=%d", len(groups["A"]), len(groups["B"]))
	}
	for _, e := range groups["A"] {
		if e.Handle() != "1" && e.Handle() != "3" {
			t.Fatalf("group A members: %s", e.Handle())
		}
	}
	sel := SelectionFromEntities(ents...)
	gsel := GroupHandles(sel, &mapSource{m: map[string]entity.Entity{"1": ents[0], "2": ents[1], "3": ents[2]}}, "layer")
	if gsel["A"].Len() != 2 || gsel["B"].Len() != 1 {
		t.Fatalf("group handles: %v", gsel)
	}
}

func TestExecuteSpatialPrefilter(t *testing.T) {
	ents := []entity.Entity{
		line("40", "0", 0, 0, 1, 1),
		line("41", "0", 100, 100, 101, 101),
	}
	sel, err := ExecuteWithFilter(ents, "bbox_intersects(-1, -1, 2, 2) and layer == '0'")
	if err != nil {
		t.Fatal(err)
	}
	if len(sel.Handles()) != 1 || sel.Handles()[0] != "40" {
		t.Fatalf("prefiltered query: %v", sel.Handles())
	}
	// Prefilter keeps entities without a bounding box.
	box := Prefilter(ents, 0, 0, 2, 2, true)
	if len(box) != 1 {
		t.Fatalf("prefilter: %v", box)
	}
}

func TestTransformSelectionRotation(t *testing.T) {
	a := line("50", "0", 1, 0, 2, 0)
	b := line("51", "0", 0, 1, 0, 2) // unselected
	sel := NewSelection("50")
	if errs := sel.Rotate(90, 0, 0, &mapSource{m: map[string]entity.Entity{"50": a, "51": b}}); len(errs) != 0 {
		t.Fatalf("rotate: %v", errs)
	}
	if math.Abs(a.Start[0]) > 1e-9 || math.Abs(a.Start[1]-1) > 1e-9 {
		t.Fatalf("rotated start: %v", a.Start)
	}
	if b.Start[0] != 0 || b.Start[1] != 1 {
		t.Fatalf("unselected entity changed: %v", b.Start)
	}
}

func TestTransformBlockDefinition(t *testing.T) {
	blk := block.NewBlock("SOFA", "sofa")
	blk.Coord = []float64{0, 0, 0}
	inner := line("60", "0", 1, 0, 2, 0)
	blk.AddEntity(&testBlockEntity{Entity: inner})

	// Reference the block through an insert: after scaling the definition by
	// 2 the insert geometry must double.
	ins := entity.NewInsert()
	ins.SetBlockName("SOFA")
	ins.SetBlockDef(blk)

	m := ScaleMatrix(2, 0, 0)
	ents := make([]entity.Entity, 0)
	for _, be := range blk.Entities() {
		if e, ok := be.(entity.Entity); ok {
			ents = append(ents, e)
		}
	}
	Apply(m, ents...)
	blk.Coord = []float64{0, 0, 0}

	parts, err := ins.Explode()
	if err != nil {
		t.Fatal(err)
	}
	ln, ok := parts[0].(*entity.Line)
	if !ok {
		t.Fatalf("exploded type %T", parts[0])
	}
	if ln.Start[0] != 2 || ln.End[0] != 4 {
		t.Fatalf("insert must reflect scaled definition: %v..%v", ln.Start, ln.End)
	}
}

type testBlockEntity struct{ entity.Entity }

func (testBlockEntity) RawTags() []block.TagPair    { return nil }
func (*testBlockEntity) SetRawTags([]block.TagPair) {}
func (t testBlockEntity) Unwrap() entity.Entity     { return t.Entity }

func TestExplodeDimensionEmitsTextAndArrows(t *testing.T) {
	d := entity.NewDimension()
	d.SetHandleString("70")
	d.SetExtensionPoints([]float64{0, 0, 0}, []float64{10, 0, 0})
	d.SetDefinitionPoint([]float64{5, 5, 0})
	d.SetTextMidpoint([]float64{5, 2.5, 0})
	d.SetText("5.0")
	parts, err := d.Explode()
	if err != nil {
		t.Fatal(err)
	}
	lines, texts := 0, 0
	for _, p := range parts {
		switch p.(type) {
		case *entity.Line:
			lines++
		case *entity.Text:
			texts++
		}
	}
	if lines != 5 || texts != 1 {
		t.Fatalf("dimension explode: %d lines, %d texts (want 5/1)", lines, texts)
	}
}

func TestDisassembleNestedInserts(t *testing.T) {
	leaf := block.NewBlock("A", "a")
	leaf.Coord = []float64{0, 0, 0}
	leaf.AddEntity(&testBlockEntity{Entity: line("80", "0", 1, 0, 2, 0)})

	outer := block.NewBlock("B", "b")
	outer.Coord = []float64{0, 0, 0}
	insA := entity.NewInsert()
	insA.SetHandleString("81")
	insA.SetBlockName("A")
	insA.SetInsertPoint(dxfmath.Vec3{5, 0, 0})
	insA.SetNonUniformScale(2, 2, 1)
	insA.SetBlockDef(leaf)
	outer.AddEntity(&testBlockEntity{Entity: insA})

	insB := entity.NewInsert()
	insB.SetHandleString("82")
	insB.SetBlockName("B")
	insB.SetInsertPoint(dxfmath.Vec3{10, 0, 0})
	insB.SetBlockDef(outer)

	blocks := map[string]*block.Block{"A": leaf, "B": outer}
	res, err := Disassemble([]entity.Entity{insB}, func(name string) *block.Block { return blocks[name] })
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Cycles) != 0 {
		t.Fatalf("unexpected cycles: %v", res.Cycles)
	}
	if len(res.Entities) != 1 {
		t.Fatalf("leaf count: %d", len(res.Entities))
	}
	ln, ok := res.Entities[0].(*entity.Line)
	if !ok {
		t.Fatalf("leaf type %T", res.Entities[0])
	}
	// (1,0) scaled x2 -> (2,0), +5 -> (7,0), +10 -> (17,0).
	if ln.Start[0] != 17 || ln.End[0] != 19 {
		t.Fatalf("nested transform: %v .. %v", ln.Start, ln.End)
	}
	if ln.Handle() != "" {
		t.Fatalf("disassembled handles must be cleared: %q", ln.Handle())
	}
}

func TestDisassembleCyclicInserts(t *testing.T) {
	a := block.NewBlock("A", "a")
	a.Coord = []float64{0, 0, 0}
	self := entity.NewInsert()
	self.SetHandleString("84")
	self.SetBlockName("A")
	self.SetBlockDef(a)
	a.AddEntity(&testBlockEntity{Entity: self})

	root := entity.NewInsert()
	root.SetHandleString("85")
	root.SetBlockName("A")
	root.SetBlockDef(a)

	res, err := Disassemble([]entity.Entity{root}, func(name string) *block.Block {
		if name == "A" {
			return a
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Cycles) != 1 || strings.Join(res.Cycles[0], "->") != "A->A" {
		t.Fatalf("cycle not reported: %v", res.Cycles)
	}
}

func TestTraverseBlockRefsNested(t *testing.T) {
	inner := block.NewBlock("B", "b")
	outer := block.NewBlock("A", "a")
	insB := entity.NewInsert()
	insB.SetHandleString("86")
	insB.SetBlockName("B")
	outer.AddEntity(&testBlockEntity{Entity: insB})
	insA := entity.NewInsert()
	insA.SetHandleString("87")
	insA.SetBlockName("A")
	insA.SetBlockDef(outer)

	blocks := map[string]*block.Block{"A": outer, "B": inner}
	var seen []string
	cycles := TraverseBlockRefs([]entity.Entity{insA}, func(name string) *block.Block { return blocks[name] },
		func(depth int, ins *entity.Insert, path []string) bool {
			seen = append(seen, fmt.Sprintf("%d:%s", depth, ins.BlockName()))
			return true
		})
	if strings.Join(seen, ",") != "0:A,1:B" {
		t.Fatalf("traversal order: %v", seen)
	}
	if len(cycles) != 0 {
		t.Fatalf("unexpected cycles: %v", cycles)
	}
}

func TestRevCloudRectangle(t *testing.T) {
	p, err := RevCloudRectangle(0, 0, 10, 4, 2, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !p.Closed {
		t.Fatal("revcloud must be closed")
	}
	if p.Num != 2*(5+2) {
		t.Fatalf("arc count: %d (want 14)", p.Num)
	}
	for i := 0; i < p.Num; i++ {
		if p.Bulges[i] != DefaultRevCloudBulge {
			t.Fatalf("bulge style at %d: %v", i, p.Bulges[i])
		}
		v := p.Vertices[i]
		onEdge := (v[1] == 0 || v[1] == 4) && v[0] >= 0 && v[0] <= 10 ||
			(v[0] == 0 || v[0] == 10) && v[1] >= 0 && v[1] <= 4
		if !onEdge {
			t.Fatalf("vertex %v off the boundary", v)
		}
	}
	arcs, err := RevCloudArcs(0, 0, 10, 4, 2, 0)
	if err != nil || len(arcs) != p.Num {
		t.Fatalf("arcs: %d %v", len(arcs), err)
	}
	if _, err := RevCloudRectangle(1, 1, 1, 2, 0, 0); err == nil {
		t.Fatal("empty rectangle must error")
	}
}

func TestUprightTextNormalization(t *testing.T) {
	tx := entity.NewText()
	tx.Coord1 = []float64{3, 4, 0}
	tx.Rotation = 180
	if !entity.MakeUpright(tx) {
		t.Fatal("upside-down text must change")
	}
	if tx.Rotation != 0 {
		t.Fatalf("rotation after upright: %v", tx.Rotation)
	}
	if tx.Coord1[0] != 3 || tx.Coord1[1] != 4 {
		t.Fatalf("insertion point must be preserved: %v", tx.Coord1)
	}
	tx2 := entity.NewText()
	tx2.Rotation = 45
	if entity.MakeUpright(tx2) {
		t.Fatal("readable text must not change")
	}
	ln := line("90", "0", 0, 0, 1, 1)
	if entity.MakeUpright(ln) {
		t.Fatal("non-text must not change")
	}
}
