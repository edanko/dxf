package drawing

import (
	"path/filepath"
	"testing"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/entity"
)

func facadeDrawing(t *testing.T) *Drawing {
	t.Helper()
	d, err := New()
	if err != nil {
		t.Fatalf("New drawing: %v", err)
	}
	return d
}

func TestFacadeAddEntityTypes(t *testing.T) {
	d := facadeDrawing(t)
	m := d.Model()
	if _, err := d.NewBlock("SYM", "symbol"); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		call func() (entity.Entity, error)
	}{
		{"POINT", func() (entity.Entity, error) { return m.AddPoint(1, 2, 3) }},
		{"LINE", func() (entity.Entity, error) { return m.AddLine(0, 0, 0, 10, 10, 0) }},
		{"CIRCLE", func() (entity.Entity, error) { return m.AddCircle(0, 0, 0, 5) }},
		{"ARC", func() (entity.Entity, error) { return m.AddArc(0, 0, 0, 5, 0, 90) }},
		{"TEXT", func() (entity.Entity, error) { return m.AddText(0, 0, 0, 2.5, "hello") }},
		{"MTEXT", func() (entity.Entity, error) { return m.AddMText(0, 0, 0, 20, 2.5, "wide text") }},
		{"ELLIPSE", func() (entity.Entity, error) { return m.AddEllipse(0, 0, 0, 5, 3, 0) }},
		{"SPLINE", func() (entity.Entity, error) { return m.AddSpline([]float64{0, 0, 0, 5, 5, 0, 10, 0, 0}) }},
		{"FACE3D", func() (entity.Entity, error) {
			return m.AddFace3D(0, 0, 0, 1, 0, 0, 1, 1, 0, 0, 1, 0)
		}},
		{"SOLID", func() (entity.Entity, error) {
			return m.AddSolid(0, 0, 0, 1, 0, 0, 0, 1, 0, 1, 1, 0)
		}},
		{"TRACE", func() (entity.Entity, error) {
			return m.AddTrace(0, 0, 0, 1, 0, 0, 0, 1, 0, 1, 1, 0)
		}},
		{"MESH", func() (entity.Entity, error) { return m.AddMesh() }},
		{"HELIX", func() (entity.Entity, error) { return m.AddHelix(0, 0, 0, 2, 3, 1) }},
		{"RAY", func() (entity.Entity, error) { return m.AddRay(0, 0, 0, 1, 0, 0) }},
		{"XLINE", func() (entity.Entity, error) { return m.AddXLine(0, 0, 0, 1, 0, 0) }},
		{"INSERT", func() (entity.Entity, error) { return m.AddInsert("SYM", 1, 1, 0) }},
		{"POLYLINE2D", func() (entity.Entity, error) { return m.AddPolyline2D(true, []float64{0, 0, 1, 0, 1, 1}) }},
		{"POLYLINE3D", func() (entity.Entity, error) { return m.AddPolyline3D(false, []float64{0, 0, 0, 1, 1, 1}) }},
		{"RECTANGLE", func() (entity.Entity, error) { return m.AddRectangle(0, 0, 10, 5) }},
		{"LWPOLYLINE", func() (entity.Entity, error) { return m.AddLwPolyline(true, []float64{0, 0, 1, 0, 1, 1}) }},
		{"DIMENSION", func() (entity.Entity, error) { return m.AddLinearDimension(0, 0, 10, 0, 5, 2) }},
		{"DIM-ALIGNED", func() (entity.Entity, error) { return m.AddAlignedDimension(0, 0, 10, 0, 5, 2) }},
		{"DIM-ANGULAR", func() (entity.Entity, error) { return m.AddAngularDimension(0, 0, 5, 5) }},
		{"DIM-ANG3P", func() (entity.Entity, error) { return m.AddAngular3PDimension(0, 0, 5, 5) }},
		{"DIM-DIAMETER", func() (entity.Entity, error) { return m.AddDiameterDimension(0, 0, 5) }},
		{"DIM-RADIUS", func() (entity.Entity, error) { return m.AddRadiusDimension(0, 0, 5) }},
		{"DIM-ORDINATE", func() (entity.Entity, error) { return m.AddOrdinateDimension(1, 1, 3, true) }},
		{"DIM-ARC", func() (entity.Entity, error) { return m.AddArcLengthDimension(0, 0) }},
		{"ATTDEF", func() (entity.Entity, error) { return m.AddAttdef("TAG", "prompt", "def", 0, 0, 0, 2.5) }},
		{"ATTRIB", func() (entity.Entity, error) { return m.AddAttrib("TAG", "val", 0, 0, 0, 2.5) }},
		{"LEADER", func() (entity.Entity, error) {
			return m.AddLeader([][]float64{{0, 0, 0}, {5, 5, 0}})
		}},
		{"MLINE", func() (entity.Entity, error) { return m.AddMLine("STANDARD", 0, 1) }},
		{"3DSOLID", func() (entity.Entity, error) { return m.Add3DSolid() }},
		{"SHAPE", func() (entity.Entity, error) { return m.AddShape("CIRCLE", 2, 0, 0, 0) }},
		{"TOLERANCE", func() (entity.Entity, error) { return m.AddTolerance("|_|_", 0, 0, 0) }},
		{"WIPEOUT", func() (entity.Entity, error) { return m.AddWipeout() }},
		{"IMAGE", func() (entity.Entity, error) { return m.AddImage() }},
		{"HATCH", func() (entity.Entity, error) { return m.AddHatch() }},
		{"VIEWPORT", func() (entity.Entity, error) { return m.AddViewport() }},
	}
	for _, tc := range cases {
		e, err := tc.call()
		if err != nil {
			t.Fatalf("Add %s: %v", tc.name, err)
		}
		if e == nil {
			t.Fatalf("Add %s: nil entity", tc.name)
		}
		if e.DXFType() == "" {
			t.Fatalf("Add %s: empty DXF type", tc.name)
		}
	}
	lines, err := m.AddGrid(0, 0, 10, 10, 2, 2)
	if err != nil || len(lines) == 0 {
		t.Fatalf("AddGrid: %v len=%d", err, len(lines))
	}
	def := entity.NewUnderlayDefinition(1, "ref.pdf", "ref.pdf")
	if _, err := m.AddUnderlay(def); err != nil {
		t.Fatalf("AddUnderlay: %v", err)
	}
	// AddGeneric builds entities from a flat attribute map.
	p, err := m.AddGeneric("POINT", map[string]any{"x": 1.0, "y": 2.0, "z": 3.0})
	if err != nil {
		t.Fatalf("AddGeneric POINT: %v", err)
	}
	if p.DXFType() != "POINT" {
		t.Fatalf("AddGeneric type = %q", p.DXFType())
	}
	if _, err := m.AddGeneric("NOPE", nil); err == nil {
		t.Fatal("AddGeneric unknown type must fail")
	}
	if len(d.entitiesSection) < len(cases)+2 {
		t.Fatalf("entity count %d, want >= %d", len(d.entitiesSection), len(cases)+2)
	}
}

func TestFacadeEntityOptions(t *testing.T) {
	d := facadeDrawing(t)
	m := d.Model()
	line, err := m.AddLine(0, 0, 0, 10, 0, 0,
		d.WithLayer("WALLS"),
		WithColor(color.ColorNumber(3)),
		WithLinetype("CONTINUOUS"),
		WithLineweight(25),
		WithTrueColor(0x00FF00),
		WithLineScale(2),
		WithInvisible,
	)
	if err != nil {
		t.Fatalf("AddLine with options: %v", err)
	}
	if _, ok := d.Layers["WALLS"]; !ok {
		t.Fatal("WithLayer did not create the layer")
	}
	attrs := line.GetAttributes()
	if attrs["layer"] != "WALLS" {
		t.Fatalf("layer attr = %v, want WALLS", attrs["layer"])
	}
	if attrs["color"] != 3 {
		t.Fatalf("color attr = %v, want 3", attrs["color"])
	}
	if attrs["ltscale"] != 2.0 {
		t.Fatalf("ltscale attr = %v, want 2", attrs["ltscale"])
	}
	// Option errors must abort the add.
	if _, err := m.AddLine(0, 0, 0, 1, 0, 0, func(e entity.Entity) error {
		return &testError{"option boom"}
	}); err == nil {
		t.Fatal("expected option error to propagate")
	}
}

type testError struct{ s string }

func (e *testError) Error() string { return e.s }

func TestFacadeBlockSpaceAdd(t *testing.T) {
	d := facadeDrawing(t)
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
	if len(b.Entities()) != 3 {
		t.Fatalf("block entities = %d, want 3", len(b.Entities()))
	}
}

func TestFacadeInsertAttribsExplode(t *testing.T) {
	d := facadeDrawing(t)
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
	ins, err := m.AddInsert("SYM", 5, 5, 0, m.WithAttribs(map[string]string{"SIZE": "42"}))
	if err != nil {
		t.Fatalf("AddInsert: %v", err)
	}
	attrib := ins.GetAttribute("SIZE")
	if attrib == nil {
		t.Fatal("insert has no SIZE attribute")
	}
	if attrib.GetText() != "42" {
		t.Fatalf("attribute = %q, want 42", attrib.GetText())
	}
	ip := ins.InsertPoint()
	mn, mx := ins.BBox()
	if mn == nil || mx == nil {
		t.Fatal("insert has no bbox")
	}
	if mn[0] > ip[0] || ip[0] > mx[0] || mn[1] > ip[1] || ip[1] > mx[1] {
		t.Fatalf("insert point %v outside bbox %v..%v", ip, mn, mx)
	}
	parts, err := ins.Explode()
	if err != nil {
		t.Fatalf("Explode: %v", err)
	}
	if len(parts) == 0 {
		t.Fatal("Explode returned no entities")
	}
}

func TestFacadeLayoutLifecycle(t *testing.T) {
	d := facadeDrawing(t)
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
	if n := d.LayoutEntityCount("Sheet1"); n != 2 {
		t.Fatalf("layout entity count = %d, want 2", n)
	}
	if n := d.LayoutEntityCount("Model"); n != 0 {
		t.Fatalf("model entity count = %d, want 0", n)
	}
	names := d.LayoutNames()
	if len(names) != 2 {
		t.Fatalf("layout names = %v", names)
	}
	infos := d.Layouts()
	if len(infos) != 2 || infos[0].Name != "Model" || infos[1].Name != "Sheet1" {
		t.Fatalf("Layouts() = %+v", infos)
	}
	if infos[1].Type != "PaperSpace" {
		t.Fatalf("layout type = %q", infos[1].Type)
	}
	if ents := d.LayoutEntities("Sheet1"); len(ents) != 2 {
		t.Fatalf("LayoutEntities = %d, want 2", len(ents))
	}
	if err := d.RenameLayout("Sheet1", "Sheet2"); err != nil {
		t.Fatal(err)
	}
	if _, err := d.SpaceByName("Sheet1"); err == nil {
		t.Fatal("old layout name still resolves")
	}
	if err := d.SetActiveLayout("Sheet2"); err != nil {
		t.Fatal(err)
	}
	if d.ActiveLayout() != "Sheet2" {
		t.Fatalf("active layout = %q", d.ActiveLayout())
	}
	if d.ActiveSpace().Name() != "Sheet2" {
		t.Fatal("ActiveSpace mismatch")
	}
	if err := d.DeleteLayout("Sheet2"); err != nil {
		t.Fatal(err)
	}
	if _, err := d.GetBlock("Sheet2"); err == nil {
		t.Fatal("layout block survived DeleteLayout")
	}
	if _, ok := d.blockRecordHandle("Sheet2"); ok {
		t.Fatal("block record survived DeleteLayout")
	}
	if len(d.entitiesSection) != 0 {
		t.Fatalf("layout entities survived DeleteLayout: %d", len(d.entitiesSection))
	}
}

func TestFacadeTransactions(t *testing.T) {
	d := facadeDrawing(t)
	m := d.Model()
	if _, err := m.AddPoint(0, 0, 0); err != nil {
		t.Fatal(err)
	}
	base := len(d.entitiesSection)

	tx, err := d.BeginTransaction()
	if err != nil {
		t.Fatalf("BeginTransaction: %v", err)
	}
	if !d.InTransaction() {
		t.Fatal("InTransaction false during transaction")
	}
	if _, err := d.BeginTransaction(); err == nil {
		t.Fatal("nested transaction must be rejected")
	}
	if _, err := m.AddCircle(1, 1, 0, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := m.AddCircle(2, 2, 0, 1); err != nil {
		t.Fatal(err)
	}
	if added, _ := tx.Stats(); added != 2 {
		t.Fatalf("tx stats added = %d, want 2", added)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if d.InTransaction() {
		t.Fatal("InTransaction true after rollback")
	}
	if len(d.entitiesSection) != base {
		t.Fatalf("after rollback entities = %d, want %d", len(d.entitiesSection), base)
	}

	tx2, err := d.BeginTransaction()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.AddCircle(3, 3, 0, 1); err != nil {
		t.Fatal(err)
	}
	if err := tx2.Commit(); err != nil {
		t.Fatal(err)
	}
	committed := len(d.entitiesSection)
	if committed != base+1 {
		t.Fatalf("after commit entities = %d, want %d", committed, base+1)
	}
	if err := tx2.Commit(); err == nil {
		t.Fatal("double commit must fail")
	}

	tx3, _ := d.BeginTransaction()
	if _, err := m.AddCircle(4, 4, 0, 1); err != nil {
		t.Fatal(err)
	}
	if err := d.DeleteEntity(d.entitiesSection[0].Handle()); err != nil {
		t.Fatal(err)
	}
	if _, removed := tx3.Stats(); removed != 1 {
		t.Fatal("delete not counted in transaction")
	}
	if err := tx3.Rollback(); err != nil {
		t.Fatal(err)
	}
	if len(d.entitiesSection) != committed {
		t.Fatalf("rollback after delete: entities = %d, want %d", len(d.entitiesSection), committed)
	}
	if d.LayoutEntityCount("Model") != committed {
		t.Fatal("layout assignment not restored")
	}
}

func TestFacadeSelectionAndExtents(t *testing.T) {
	d := facadeDrawing(t)
	m := d.Model()
	if _, err := m.AddLine(0, 0, 0, 10, 10, 0, d.WithLayer("WALLS")); err != nil {
		t.Fatal(err)
	}
	if _, err := m.AddLine(20, 0, 0, 30, 10, 0, d.WithLayer("WALLS")); err != nil {
		t.Fatal(err)
	}
	if _, err := m.AddCircle(1000, 1000, 0, 5, d.WithLayer("DOORS")); err != nil {
		t.Fatal(err)
	}
	if n := len(d.SelectByType("LINE")); n != 2 {
		t.Fatalf("SelectByType LINE = %d, want 2", n)
	}
	if n := len(d.SelectByType("LINE", "CIRCLE")); n != 3 {
		t.Fatalf("SelectByType LINE+CIRCLE = %d, want 3", n)
	}
	if n := len(d.SelectByLayer("WALLS")); n != 2 {
		t.Fatalf("SelectByLayer WALLS = %d, want 2", n)
	}
	sel := d.SelectInExtents(-5, -5, 35, 15)
	if len(sel) != 2 {
		t.Fatalf("SelectInExtents = %d, want 2", len(sel))
	}
	mn, mx, ok := d.Extents()
	if !ok {
		t.Fatal("Extents not computable")
	}
	if mn.X() > 0 || mn.Y() > 0 || mx.X() < 1000 {
		t.Fatalf("extents %v..%v do not cover geometry", mn, mx)
	}
	rmn, rmx, ok := d.RecalcExtents()
	if !ok || rmn.X() != mn.X() || rmx.X() != mx.X() {
		t.Fatalf("RecalcExtents %v..%v vs Extents %v..%v", rmn, rmx, mn, mx)
	}
	if d.Header().ExtMax[0] < 1000 {
		t.Fatalf("header EXTMAX not updated by RecalcExtents: %v", d.Header().ExtMax)
	}
}

func TestFacadeGroups(t *testing.T) {
	d := facadeDrawing(t)
	m := d.Model()
	l1, err := m.AddLine(0, 0, 0, 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	l2, err := m.AddLine(1, 0, 0, 2, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	g, err := d.Group("frame", "frame group", l1, l2)
	if err != nil {
		t.Fatalf("Group create: %v", err)
	}
	if g == nil {
		t.Fatal("nil group")
	}
	l3, err := m.AddLine(2, 0, 0, 3, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	g2, err := d.Group("frame", "frame group", l3)
	if err != nil {
		t.Fatal(err)
	}
	if g2 != g {
		t.Fatal("second Group call created a new group")
	}
	if len(g2.Entities()) != 3 {
		t.Fatalf("group members = %d, want 3", len(g2.Entities()))
	}
	if err := d.AddToGroup("frame", l1); err != nil {
		t.Fatalf("AddToGroup (existing member): %v", err)
	}
	if err := d.AddToGroup("missing-group", l1); err == nil {
		t.Fatal("AddToGroup on a missing group must fail")
	}
	// Handles materialize at save time; GroupMembers reads them afterwards.
	fn := filepath.Join(t.TempDir(), "groups.dxf")
	if err := d.SaveAs(fn); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}
	members, err := d.GroupMembers("frame")
	if err != nil {
		t.Fatal(err)
	}
	if len(members) != 3 {
		t.Fatalf("GroupMembers = %v, want 3 handles", members)
	}
	seen := map[string]bool{}
	for _, h := range members {
		if h == "" || seen[h] {
			t.Fatalf("bad member handle %q", h)
		}
		seen[h] = true
	}
	if _, err := d.GroupMembers("nope"); err == nil {
		t.Fatal("GroupMembers on missing group must fail")
	}
}
