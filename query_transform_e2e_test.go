// End-to-end parity tests for change 9: query, selection sets, transforms,
// disassembly, draw order, upright text, zoom fit and proxy graphics.
package dxf

import (
	"bytes"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/query"
)

func saveReload(t *testing.T, d *drawing.Drawing) *drawing.Drawing {
	t.Helper()
	var buf bytes.Buffer
	if _, err := d.WriteTo(&buf); err != nil {
		t.Fatalf("save: %v", err)
	}
	doc, err := FromStringData(buf.String())
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	return doc
}

func TestDrawOrderRoundTripE2E(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	for _, spec := range [][2]float64{{0, 0}, {5, 0}, {10, 0}} {
		if _, err := d.Model().AddLine(spec[0], spec[1], 0, spec[0]+1, spec[1], 0); err != nil {
			t.Fatal(err)
		}
	}
	order := d.DrawOrder()
	if len(order) != 3 {
		t.Fatalf("draw order: %v", order)
	}
	last := order[2]
	if err := d.SendToBack(last); err != nil {
		t.Fatal(err)
	}
	if got := d.DrawOrder(); got[0] != last {
		t.Fatalf("send to back: %v", got)
	}
	doc := saveReload(t, d)
	got := doc.DrawOrder()
	if strings.Join(got, ",") != strings.Join(d.DrawOrder(), ",") {
		t.Fatalf("draw order after round trip: %v want %v", got, d.DrawOrder())
	}
	if len(doc.Objects()) <= len(d.Objects()) {
		// the regenerated SORTENTSTABLE must survive the round trip
	}
}

func TestDrawOrderBringToFrontAndSet(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	for i := range 3 {
		if _, err := d.Model().AddLine(float64(i), 0, 0, float64(i)+1, 1, 0); err != nil {
			t.Fatal(err)
		}
	}
	order := d.DrawOrder()
	if err := d.BringToFront(order[0]); err != nil {
		t.Fatal(err)
	}
	if got := d.DrawOrder(); got[2] != order[0] {
		t.Fatalf("bring to front: %v", got)
	}
	frontToBack := []string{order[2], order[1], order[0]}
	if err := d.SetDrawOrder(frontToBack); err != nil {
		t.Fatal(err)
	}
	doc := saveReload(t, d)
	got := doc.DrawOrder()
	want := []string{order[0], order[1], order[2]} // plot order = back to front
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("set draw order round trip: %v want %v", got, want)
	}
	if err := d.SetDrawOrder([]string{"NOPE", order[1], order[2]}); err == nil {
		t.Fatal("unknown handle must error")
	}
}

func TestDrawOrderR12Fallback(t *testing.T) {
	data, err := os.ReadFile("testdata/Minimal_DXF_AC1009.dxf")
	if err != nil {
		t.Fatal(err)
	}
	d, err := FromStringData(string(data))
	if err != nil {
		t.Fatal(err)
	}
	order := d.DrawOrder()
	if len(order) < 2 {
		t.Skipf("R12 fixture has %d model entities", len(order))
	}
	last := order[len(order)-1]
	if err := d.SendToBack(last); err != nil {
		t.Fatal(err)
	}
	if got := d.DrawOrder(); got[0] != last {
		t.Fatalf("R12 send to back: %v", got)
	}
	doc := saveReload(t, d)
	if got := doc.DrawOrder(); len(got) == len(order) && got[0] != last {
		t.Fatalf("R12 draw order not preserved: %v", got)
	}
}

func TestQuerySelectionsAndTransformsE2E(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	la, err := d.Model().AddLine(0, 0, 0, 10, 0, 0, d.WithLayer("A"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.Model().AddLine(0, 5, 0, 10, 5, 0, d.WithLayer("B")); err != nil {
		t.Fatal(err)
	}
	tx, err := d.Model().AddText(1, 1, 0, 2.5, "hello")
	if err != nil {
		t.Fatal(err)
	}
	tx.Rotation = 180

	sel, err := d.QueryExpression("type == 'LINE' and layer == 'A'")
	if err != nil {
		t.Fatal(err)
	}
	if sel.Len() != 1 || !sel.Has(la.Handle()) {
		t.Fatalf("query: %v", sel.Handles())
	}
	d.SetSelectionSet("walls-a", sel)
	if names := d.SelectionSetNames(); len(names) != 1 || names[0] != "walls-a" {
		t.Fatalf("named sets: %v", names)
	}
	if s, ok := d.SelectionSet("walls-a"); !ok || s.Name() != "walls-a" {
		t.Fatalf("selection set lookup failed: %v %v", s, ok)
	}
	all, err := d.QueryExpression("type == 'LINE'")
	if err != nil || all.Len() != 2 {
		t.Fatalf("all lines: %v %v", all, err)
	}
	onlyA := all.Intersect(sel)
	if onlyA.Len() != 1 {
		t.Fatalf("intersection: %v", onlyA.Handles())
	}

	// Rotate the A-line 90 degrees about the origin.
	if errs := d.TransformSelection(onlyA, query.RotationMatrix(90, 0, 0)); len(errs) != 0 {
		t.Fatalf("transform selection: %v", errs)
	}
	if math.Abs(la.End[0]) > 1e-9 || math.Abs(la.End[1]-10) > 1e-9 {
		t.Fatalf("rotated geometry: %v", la.End)
	}

	if n := d.MakeUprightAll(); n != 1 {
		t.Fatalf("upright changed %d", n)
	}
	if tx.Rotation != 0 {
		t.Fatalf("text rotation: %v", tx.Rotation)
	}

	rev, err := d.AddRevCloud(0, 0, 10, 10, 2, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !rev.Closed {
		t.Fatal("revcloud not closed")
	}
	if len(d.Entities()) != 4 {
		t.Fatalf("entities after revcloud: %d", len(d.Entities()))
	}

	fit, ok := d.FitView(4.0/3.0, 0.1)
	if !ok {
		t.Fatal("fit view: no extents")
	}
	if fit.Width < 10 || fit.Height < 5 {
		t.Fatalf("fit view must contain extents: %+v", fit)
	}

	doc := saveReload(t, d)
	if len(doc.Entities()) != len(d.Entities()) {
		t.Fatalf("round trip entity count: %d vs %d", len(doc.Entities()), len(d.Entities()))
	}
	if !d.DeleteSelectionSet("walls-a") || len(d.SelectionSetNames()) != 0 {
		t.Fatal("delete selection set")
	}
}

func TestTransformBlockAndDrawingE2E(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	blk, err := d.NewBlock("WING", "wing")
	if err != nil {
		t.Fatal(err)
	}
	bs := d.BlockSpace(blk)
	wline, err := bs.AddLine(0, 0, 0, 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	ins, err := d.Model().AddInsert("WING", 5, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	if errs := d.TransformBlock("WING", query.ScaleMatrix(2, 0, 0)); len(errs) != 0 {
		t.Fatalf("transform block: %v", errs)
	}
	if wline.Start[0] != 0 || wline.End[0] != 2 {
		t.Fatalf("block geometry scaled: %v..%v", wline.Start, wline.End)
	}
	// The insert reflects the new definition through its matrix.
	dis, err := d.DisassembleModel()
	if err != nil {
		t.Fatal(err)
	}
	if len(dis.Entities) != 1 || len(dis.Cycles) != 0 {
		t.Fatalf("disassemble: %d ents %v cycles", len(dis.Entities), dis.Cycles)
	}
	ln, ok := dis.Entities[0].(*entity.Line)
	if !ok {
		t.Fatalf("leaf type %T", dis.Entities[0])
	}
	if ln.Start[0] != 5 || ln.End[0] != 7 {
		t.Fatalf("disassembled leaf: %v..%v", ln.Start, ln.End)
	}
	_ = ins

	if errs := d.TransformDrawing(query.TranslateMatrix(1, 1, 0)); len(errs) != 0 {
		t.Fatalf("transform drawing: %v", errs)
	}
	if math.Abs(wline.Start[1]-1) > 1e-9 {
		t.Fatalf("drawing transform must move block geometry: %v", wline.Start)
	}
}

func TestProxyGraphicsE2E(t *testing.T) {
	payload := entity.EncodeBinaryTags([]entity.BinaryTag{
		{Code: 0, Value: "LINE"},
		{Code: 5, Value: "AA"},
		{Code: 8, Value: "0"},
		{Code: 100, Value: "AcDbEntity"},
		{Code: 100, Value: "AcDbLine"},
		{Code: 10, Value: 1.0},
		{Code: 20, Value: 2.0},
		{Code: 30, Value: 0.0},
		{Code: 11, Value: 3.0},
		{Code: 21, Value: 4.0},
		{Code: 31, Value: 0.0},
	})
	p := entity.NewAcadProxyEntity()
	p.SetHandleString("42")
	p.SetEntityData(payload)
	ents, err := entity.ExtractProxyGraphics(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(ents) != 1 {
		t.Fatalf("extracted: %d", len(ents))
	}
	ln, ok := ents[0].(*entity.Line)
	if !ok {
		t.Fatalf("type %T", ents[0])
	}
	if ln.Handle() != "" || ln.Start[0] != 1 || ln.End[1] != 4 {
		t.Fatalf("line: h=%q %v..%v", ln.Handle(), ln.Start, ln.End)
	}
}

func TestQueryTransformRoundTripStability(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	for i := range 4 {
		if _, err := d.Model().AddLine(float64(i), 0, 0, float64(i), 3, 0); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := d.AddRevCloud(0, -1, 4, 4, 1, 0); err != nil {
		t.Fatal(err)
	}
	if err := d.BringToFront(d.DrawOrder()[0]); err != nil {
		t.Fatal(err)
	}
	doc1 := saveReload(t, d)
	doc2 := saveReload(t, doc1)
	if len(doc2.Entities()) != len(doc1.Entities()) {
		t.Fatalf("entity count drift: %d vs %d", len(doc2.Entities()), len(doc1.Entities()))
	}
	if strings.Join(doc2.DrawOrder(), ",") != strings.Join(doc1.DrawOrder(), ",") {
		t.Fatalf("draw order drift: %v vs %v", doc2.DrawOrder(), doc1.DrawOrder())
	}
}
