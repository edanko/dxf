package entity

import (
	"math"
	"strings"
	"testing"

	"github.com/edanko/dxf/format"
	dxfmath "github.com/edanko/dxf/math"
)

func containsStr(s, sub string) bool { return strings.Contains(s, sub) }

// TestBoundaryPathAddArcBulgeEdge verifies AddArc emits a single bulge edge
// whose curve is the requested arc.
func TestBoundaryPathAddArcBulgeEdge(t *testing.T) {
	p := &MPolygonBoundaryPath{IsClosed: false}
	// Quarter circle, center (5,5), radius 2, from angle 0 to pi/2 (ccw).
	p.AddArc(5, 5, 2, 0, math.Pi/2)

	if p.VertexCount() != 2 {
		t.Fatalf("arc should be one edge = 2 vertices, got %d", p.VertexCount())
	}
	start := p.GetVertex(0)
	end := p.GetVertex(1)
	wantBulge := math.Tan(math.Pi / 8) // tan(span/4)
	if math.Abs(start.Bulge-wantBulge) > 1e-9 {
		t.Errorf("bulge = %v, want %v", start.Bulge, wantBulge)
	}
	if end.Bulge != 0 {
		t.Errorf("end vertex bulge = %v, want 0", end.Bulge)
	}
	if d := dxfmath.NewVec2(start.X, start.Y).Distance(dxfmath.NewVec2(7, 5)); d > 1e-9 {
		t.Errorf("start point = %v, want (7,5)", start)
	}
	if d := dxfmath.NewVec2(end.X, end.Y).Distance(dxfmath.NewVec2(5, 7)); d > 1e-9 {
		t.Errorf("end point = %v, want (5,7)", end)
	}

	// The bulge edge reproduces the requested arc geometry.
	center, radius, a0, a1, ccw, ok := p.ArcInfo(0)
	if !ok {
		t.Fatal("ArcInfo should succeed for a bulge edge")
	}
	if d := center.Distance(dxfmath.NewVec2(5, 5)); d > 1e-9 {
		t.Errorf("center = %v, want (5,5)", center)
	}
	if math.Abs(radius-2) > 1e-9 {
		t.Errorf("radius = %v, want 2", radius)
	}
	if math.Abs(a0) > 1e-9 {
		t.Errorf("start angle = %v, want 0", a0)
	}
	if math.Abs(a1-math.Pi/2) > 1e-9 {
		t.Errorf("end angle = %v, want pi/2", a1)
	}
	if !ccw {
		t.Error("arc should sweep counter-clockwise")
	}

	// Clockwise sweep produces a negative bulge and mirrored sweep.
	cw := &MPolygonBoundaryPath{}
	cw.AddArc(5, 5, 2, math.Pi/2, 0)
	if cw.GetVertex(0).Bulge >= 0 {
		t.Errorf("clockwise bulge = %v, want negative", cw.GetVertex(0).Bulge)
	}
	if _, _, _, _, ccw2, ok2 := cw.ArcInfo(0); !ok2 || ccw2 {
		t.Errorf("clockwise arc: ccw=%v ok=%v", ccw2, ok2)
	}
}

// TestBoundaryPathLineEdgeArcInfo verifies a straight edge reports no arc.
func TestBoundaryPathLineEdgeArcInfo(t *testing.T) {
	p := &MPolygonBoundaryPath{}
	p.AddVertex2D(0, 0)
	p.AddVertex2D(10, 0)
	if _, _, _, _, _, ok := p.ArcInfo(0); ok {
		t.Error("line edge should report ok=false")
	}
	// Open path: last vertex has no following edge.
	if _, _, _, _, _, ok := p.ArcInfo(1); ok {
		t.Error("open path last vertex should report ok=false")
	}
	// Closed path: last vertex edge wraps to the first vertex.
	p.IsClosed = true
	p.Vertices[1].Bulge = 0.5
	if _, _, _, _, _, ok := p.ArcInfo(1); !ok {
		t.Error("closed path last vertex edge should wrap and report ok=true")
	}
}

// TestMPolygonEdgeTypesSurviveCopyTransformFormat verifies that line and arc
// edges keep their curves across copy, transform, and format.
func TestMPolygonEdgeTypesSurviveCopyTransformFormat(t *testing.T) {
	m := NewMPolygon()
	path := m.AddBoundaryPath(BOUNDARY_PATH_DEFAULT, true)
	path.AddVertex2D(0, 0)             // start of arc edge
	path.AddArc(5, 5, 2, 0, math.Pi/2) // arc edge (2 vertices)
	path.AddVertex2D(20, 0)            // line edge from arc end

	arcCurve := func(mp *MPolygonEntity) (center dxfmath.Vec2, radius, a0, a1 float64, ccw, ok bool) {
		// Vertices: [0] line-start, [1] arc start (bulge), [2] arc end,
		// [3] line end. The arc edge starts at index 1.
		return mp.paths.GetPath(0).ArcInfo(1)
	}

	// Reference geometry before any operation.
	wantCenter, wantRadius, _, _, wantCCW, ok := arcCurve(m)
	if !ok {
		t.Fatal("arc edge missing")
	}

	// Copy: curve unchanged.
	clone := m.Clone()
	cCenter, cRadius, _, _, cCCW, ok := clone.paths.GetPath(0).ArcInfo(1)
	if !ok || cCenter.Distance(wantCenter) > 1e-9 || math.Abs(cRadius-wantRadius) > 1e-9 || cCCW != wantCCW {
		t.Fatalf("copy changed arc curve: center=%v radius=%v ccw=%v", cCenter, cRadius, cCCW)
	}

	// Transform: rotate 90 deg about Z, translate (10,20), uniform scale 2.
	mat := dxfmath.Matrix44{
		0, -2, 0, 10,
		2, 0, 0, 20,
		0, 0, 2, 0,
		0, 0, 0, 1,
	}
	if err := m.Transform(mat); err != nil {
		t.Fatalf("Transform: %v", err)
	}
	// Radius scales by 2, center transforms by the matrix, bulge unchanged.
	ctr, rad, _, _, ccw, ok := arcCurve(m)
	if !ok {
		t.Fatal("arc edge missing after transform")
	}
	scaledWantCenter := wantCenter
	expected := mat.MulVec(dxfmath.NewVec3(scaledWantCenter.X(), scaledWantCenter.Y(), 0))
	if ctr.Distance(dxfmath.NewVec2(expected.X(), expected.Y())) > 1e-9 {
		t.Errorf("transformed center = %v, want %v", ctr, expected)
	}
	if math.Abs(rad-2*wantRadius) > 1e-9 {
		t.Errorf("transformed radius = %v, want %v", rad, 2*wantRadius)
	}
	if ccw != wantCCW {
		t.Errorf("sweep direction changed under rotation")
	}

	// Format: vertices written with bulges intact.
	f := format.NewASCII()
	m.Format(f)
	out := f.Output()
	if !containsStr(out, "42") || !containsStr(out, "0.4142") {
		t.Errorf("formatted output missing bulge 0.4142:\n%s", out)
	}
}

// TestMPolygonTransformReflectionAndNonUniform verifies bulge sign flip under
// reflection and the non-uniform scale rejection.
func TestMPolygonTransformReflectionAndNonUniform(t *testing.T) {
	newPolygon := func() *MPolygonEntity {
		m := NewMPolygon()
		path := m.AddBoundaryPath(BOUNDARY_PATH_DEFAULT, true)
		path.AddArc(5, 5, 2, 0, math.Pi/2)
		return m
	}

	// Reflection across X axis: determinant -2 (uniform mirror scale).
	m := newPolygon()
	refl := dxfmath.Matrix44{
		2, 0, 0, 0,
		0, -2, 0, 0,
		0, 0, 2, 0,
		0, 0, 0, 1,
	}
	if err := m.Transform(refl); err != nil {
		t.Fatalf("reflection Transform: %v", err)
	}
	if bulge := m.paths.GetPath(0).GetVertex(0).Bulge; bulge >= 0 {
		t.Errorf("bulge after reflection = %v, want negative", bulge)
	}
	if _, _, _, _, ccw, ok := m.paths.GetPath(0).ArcInfo(0); !ok || ccw {
		t.Errorf("reflected arc should sweep clockwise: ccw=%v ok=%v", ccw, ok)
	}

	// Non-uniform scale with bulges: rejected.
	m2 := newPolygon()
	nu := dxfmath.Matrix44{
		3, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
	err := m2.Transform(nu)
	if err == nil {
		t.Fatal("expected error for non-uniform scale with bulge edges")
	}
	if !containsStr(err.Error(), "non-uniform") {
		t.Errorf("error should mention non-uniform scale: %v", err)
	}
	// Geometry unchanged after rejection.
	if x := m2.paths.GetPath(0).GetVertex(0).X; math.Abs(x-7) > 1e-9 {
		t.Errorf("vertex changed after rejected transform: %v", x)
	}

	// Non-uniform scale without bulges: accepted.
	m3 := NewMPolygon()
	p3 := m3.AddBoundaryPath(BOUNDARY_PATH_DEFAULT, false)
	p3.AddVertex2D(0, 0)
	p3.AddVertex2D(10, 0)
	if err := m3.Transform(nu); err != nil {
		t.Fatalf("non-uniform scale without bulges should succeed: %v", err)
	}
	if x := m3.paths.GetPath(0).GetVertex(1).X; math.Abs(x-30) > 1e-9 {
		t.Errorf("scaled vertex = %v, want 30", x)
	}
}
