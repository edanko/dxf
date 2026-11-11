package dxf

import (
	"bytes"
	"math"
	"strings"
	"testing"

	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
	"github.com/edanko/dxf/path"
	"github.com/edanko/dxf/render"
)

// hatchEdgeFixture builds a HATCH whose single boundary path contains one edge
// of every supported type.
const hatchEdgeFixture = `0
SECTION
2
ENTITIES
0
HATCH
5
1
100
AcDbEntity
8
0
100
AcDbHatch
70
0
71
1
91
1
92
1
93
4
72
1
10
0
20
0
11
10
21
0
72
2
10
10
20
5
40
5
50
0
51
90
73
1
72
3
10
20
20
5
11
3
21
0
40
0.5
50
0
51
180
73
1
72
4
94
3
73
0
74
0
95
8
40
0
40
0
40
0
40
0
40
1
40
1
40
1
40
1
96
4
10
20
20
10
10
11
20
22
10
12
20
22
10
13
20
14
0
ENDSEC
0
EOF
`

// renderScene builds a raster-ready scene for fallback assertions.
func renderScene(d *drawing.Drawing) (*render.StyledScene, error) {
	return render.BuildScene(d, 200, 200, &render.SceneOptions{Tessellation: 0.05})
}

func hatchEdgePaths(t *testing.T, d *drawing.Drawing) []entity.BoundaryPath {
	t.Helper()
	for _, e := range d.ModelSpace() {
		if h, ok := e.(*entity.Hatch); ok {
			return h.GetBoundaryPaths()
		}
	}
	t.Fatal("no HATCH entity loaded")
	return nil
}

func TestHatchBoundaryEdgeLoadAndRoundTrip(t *testing.T) {
	d, err := FromStringData(hatchEdgeFixture)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	paths := hatchEdgePaths(t, d)
	if len(paths) != 1 {
		t.Fatalf("boundary paths = %d, want 1", len(paths))
	}
	ep, ok := paths[0].(*entity.EdgePath)
	if !ok {
		t.Fatalf("boundary path type = %T, want *entity.EdgePath", paths[0])
	}
	edges := ep.Edges()
	if len(edges) != 4 {
		var kinds []int
		for _, e := range edges {
			kinds = append(kinds, e.EdgeType())
		}
		t.Fatalf("edges = %d (%v), unsupported %v, want 4", len(edges), kinds, ep.Unsupported())
	}
	line, ok := edges[0].(*entity.LineEdge)
	if !ok || line.End != dxfmath.NewVec2(10, 0) {
		t.Fatalf("line edge = %#v", edges[0])
	}
	arc, ok := edges[1].(*entity.ArcEdge)
	if !ok || arc.Radius != 5 || !arc.CCW || arc.Sweep() != 90 {
		t.Fatalf("arc edge = %#v", edges[1])
	}
	ell, ok := edges[2].(*entity.EllipseEdge)
	if !ok || ell.MajorAxis != dxfmath.NewVec2(3, 0) || ell.Ratio != 0.5 {
		t.Fatalf("ellipse edge = %#v", edges[2])
	}
	spl, ok := edges[3].(*entity.SplineEdge)
	if !ok || spl.Degree != 3 || len(spl.Controls) != 4 || len(spl.Knots) != 8 {
		t.Fatalf("spline edge = %#v", edges[3])
	}

	// Save and reload: every typed field must survive unchanged.
	var buf bytes.Buffer
	if _, err := d.WriteTo(&buf); err != nil {
		t.Fatalf("save: %v", err)
	}
	reloaded, err := FromStringData(buf.String())
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	paths = hatchEdgePaths(t, reloaded)
	ep2, ok := paths[0].(*entity.EdgePath)
	if !ok {
		t.Fatalf("reloaded boundary path type = %T", paths[0])
	}
	if len(ep2.Edges()) != 4 {
		t.Fatalf("reloaded edges = %d, want 4", len(ep2.Edges()))
	}
	arc2 := ep2.Edges()[1].(*entity.ArcEdge)
	if arc2.Radius != arc.Radius || arc2.StartAngle != arc.StartAngle || arc2.EndAngle != arc.EndAngle || arc2.CCW != arc.CCW {
		t.Fatalf("arc changed on round trip: %#v vs %#v", arc2, arc)
	}
	spl2 := ep2.Edges()[3].(*entity.SplineEdge)
	if spl2.Degree != spl.Degree || len(spl2.Controls) != len(spl.Controls) {
		t.Fatalf("spline changed on round trip: %#v", spl2)
	}
	for i := range spl.Controls {
		if spl2.Controls[i] != spl.Controls[i] {
			t.Fatalf("spline control %d = %v, want %v", i, spl2.Controls[i], spl.Controls[i])
		}
	}
}

func TestHatchBoundaryEdgeTransform(t *testing.T) {
	d, err := FromStringData(hatchEdgeFixture)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	paths := hatchEdgePaths(t, d)
	ep := paths[0].(*entity.EdgePath)
	m := dxfmath.Translation(100, 50, 0).Mul(dxfmath.Scale(2, 2, 1))
	transformed := ep.TransformMatrix(m)
	edges := transformed.Edges()

	line := edges[0].(*entity.LineEdge)
	if line.Start != dxfmath.NewVec2(100, 50) || line.End != dxfmath.NewVec2(120, 50) {
		t.Errorf("line transform = %v -> %v", line.Start, line.End)
	}
	arc := edges[1].(*entity.ArcEdge)
	if math.Abs(arc.Radius-10) > 1e-9 {
		t.Errorf("arc radius after scale = %g, want 10", arc.Radius)
	}
	if math.Abs(arc.Center.X()-120) > 1e-9 || math.Abs(arc.Center.Y()-60) > 1e-9 {
		t.Errorf("arc center after transform = %v, want (120,60)", arc.Center)
	}
	ell := edges[2].(*entity.EllipseEdge)
	if math.Abs(ell.MajorAxis.Length()-6) > 1e-9 {
		t.Errorf("ellipse major axis after scale = %g, want 6", ell.MajorAxis.Length())
	}
	spl := edges[3].(*entity.SplineEdge)
	if math.Abs(spl.Controls[0].X()-140) > 1e-9 || math.Abs(spl.Controls[0].Y()-70) > 1e-9 {
		t.Errorf("spline control after transform = %v, want (140,70)", spl.Controls[0])
	}
}

func TestHatchBoundaryUnsupportedEdgeReported(t *testing.T) {
	bad := strings.Replace(hatchEdgeFixture, "72\n1\n10\n0\n20\n0\n11\n10\n21\n0", "72\n9\n10\n0\n20\n0", 1)
	d, err := FromStringData(bad)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	paths := hatchEdgePaths(t, d)
	ep := paths[0].(*entity.EdgePath)
	if len(ep.Unsupported()) != 1 || !strings.Contains(ep.Unsupported()[0], "type 9") {
		t.Fatalf("unsupported = %v, want type 9 report", ep.Unsupported())
	}
	if len(ep.Edges()) != 3 {
		t.Fatalf("supported edges = %d, want 3", len(ep.Edges()))
	}
}

// TestHatchMixedBoundaryWithSplineHole converts a hatch whose outer loop is a
// bulge polygon and whose hole is a spline edge path into closed fill loops.
func TestHatchMixedBoundaryWithSplineHole(t *testing.T) {
	h := entity.NewHatch()
	h.SetSolidFill(true)
	outer := entity.NewPolylinePath(0, true, false)
	for _, pt := range [][2]float64{{0, 0}, {20, 0}, {20, 20}, {0, 20}} {
		outer.AddVertex(entity.NewVertex3D(pt[0], pt[1], 0, 0))
	}
	h.AddBoundaryPath(outer)
	hole := entity.NewEdgePath(0)
	hole.AddEdge(entity.NewLineEdge(dxfmath.NewVec2(5, 5), dxfmath.NewVec2(15, 5)))
	hole.AddEdge(entity.NewSplineEdge(3, false, false,
		[]float64{0, 0, 0, 0, 1, 1, 1, 1},
		[]dxfmath.Vec3{
			dxfmath.NewVec3(15, 5, 0), dxfmath.NewVec3(15, 15, 0),
			dxfmath.NewVec3(5, 15, 0), dxfmath.NewVec3(5, 5, 0),
		}, nil))
	h.AddBoundaryPath(hole)

	p, err := path.MakePath(h, path.WithDistance(0.05))
	if err != nil {
		t.Fatalf("MakePath: %v", err)
	}
	subs := p.SubPaths()
	if len(subs) != 2 {
		t.Fatalf("sub-paths = %d, want outer + hole", len(subs))
	}
	for i, sub := range subs {
		if !sub.IsClosed() {
			t.Errorf("sub-path %d is not closed", i)
		}
	}
	// The spline hole must stay inside the outer boundary.
	min, max := subs[1].BoundingBox()
	if min.X() < 4 || max.X() > 16 || min.Y() < 4 || max.Y() > 16 {
		t.Errorf("spline hole bounds %v..%v outside expected region", min, max)
	}

	// Entity transforms must move edge geometry with the hatch.
	if err := h.Transform(dxfmath.Translation(100, 0, 0)); err != nil {
		t.Fatalf("Transform: %v", err)
	}
	p2, err := path.MakePath(h, path.WithDistance(0.05))
	if err != nil {
		t.Fatalf("MakePath after transform: %v", err)
	}
	min2, _ := p2.BoundingBox()
	if math.Abs(min2.X()-100) > 1e-6 {
		t.Errorf("transformed hatch min x = %g, want 100", min2.X())
	}
}

// TestHatchUnsupportedEdgeFallbackStable verifies the scene keeps supported
// edges and reports the unsupported edge with the entity handle, identically
// across repeated builds.
func TestHatchUnsupportedEdgeFallbackStable(t *testing.T) {
	bad := strings.Replace(hatchEdgeFixture, "72\n1\n10\n0\n20\n0\n11\n10\n21\n0", "72\n9\n10\n0\n20\n0", 1)
	d, err := FromStringData(bad)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	first, err := renderScene(d)
	if err != nil {
		t.Fatalf("scene 1: %v", err)
	}
	second, err := renderScene(d)
	if err != nil {
		t.Fatalf("scene 2: %v", err)
	}
	if len(first.Paths) == 0 {
		t.Fatal("supported edges were dropped from the scene")
	}
	if len(first.Fallbacks) == 0 {
		t.Fatal("unsupported edge produced no fallback report")
	}
	want := ""
	for _, f := range first.Fallbacks {
		if strings.Contains(f, "HATCH 1") && strings.Contains(f, "type 9") {
			want = f
		}
	}
	if want == "" {
		t.Fatalf("fallback lacks handle/reason: %v", first.Fallbacks)
	}
	if strings.Join(first.Fallbacks, "|") != strings.Join(second.Fallbacks, "|") {
		t.Fatalf("fallback reports differ: %v vs %v", first.Fallbacks, second.Fallbacks)
	}
}
