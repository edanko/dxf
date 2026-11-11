package meshex_test

import (
	"os"
	"strings"
	"testing"

	"github.com/edanko/dxf/addons/meshex"
	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/edges"
	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
)

func cubeSAT(t *testing.T) []string {
	t.Helper()
	data, err := os.ReadFile("../../testdata/acis_cube.sat")
	if err != nil {
		t.Fatalf("read SAT fixture: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\r\n"), "\n")
	for i := range lines {
		lines[i] = strings.TrimSuffix(lines[i], "\r")
	}
	return lines
}

func squareLoop(t *testing.T) edges.Loop {
	t.Helper()
	edgesList := []edges.Edge{
		edges.NewEdge(dxfmath.NewVec2(0, 0), dxfmath.NewVec2(4, 0)),
		edges.NewEdge(dxfmath.NewVec2(4, 0), dxfmath.NewVec2(4, 3)),
		edges.NewEdge(dxfmath.NewVec2(4, 3), dxfmath.NewVec2(0, 3)),
		edges.NewEdge(dxfmath.NewVec2(0, 3), dxfmath.NewVec2(0, 0)),
	}
	topology := edges.BuildGraph(edgesList, 1e-6).Analyze()
	if len(topology.Loops) != 1 {
		t.Fatalf("loops = %d, want 1", len(topology.Loops))
	}
	return topology.Loops[0]
}

func TestACISAndLinkedEdgeGeometryReachExporters(t *testing.T) {
	doc, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	body := entity.NewBody()
	body.SetHandleString("AB")
	body.SetSATLines(cubeSAT(t))
	if err := doc.AddEntity(body); err != nil {
		t.Fatal(err)
	}

	loop := squareLoop(t)
	geometry := meshex.CollectGeometry(doc)
	if len(geometry.Diagnostics) != 0 {
		t.Fatalf("diagnostics: %v", geometry.Diagnostics)
	}
	if len(geometry.Meshes) != 1 {
		t.Fatalf("decoded meshes = %d, want 1", len(geometry.Meshes))
	}
	mesh := geometry.Meshes[0]
	if got := len(mesh.Vertices()); got != 8 {
		t.Fatalf("cube vertices = %d, want 8", got)
	}
	if got := len(mesh.Faces()); got != 12 {
		t.Fatalf("cube triangles = %d, want 12", got)
	}
	if got := geometry.TriangleCount(); got != 12 {
		t.Fatalf("triangle count = %d, want 12", got)
	}

	geometry.AddLoops(loop)
	if got := geometry.VertexCount(); got != 12 {
		t.Fatalf("combined vertices = %d, want 12", got)
	}
	if got := geometry.TriangleCount(); got != 14 {
		t.Fatalf("combined triangles = %d, want 14", got)
	}

	obj, err := meshex.ExportOBJWithLoops(doc, []edges.Loop{loop})
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Count("\n"+obj, "\nv "); got != 12 {
		t.Fatalf("obj vertices = %d, want 12", got)
	}
	if !strings.Contains(obj, "f 9 10 11 12") {
		t.Fatalf("obj loop face order missing or changed:\n%s", obj)
	}

	scad, err := meshex.ExportOpenSCADWithLoops(doc, []edges.Loop{loop})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(scad, "polyhedron") || !strings.Contains(scad, "polygon(points=[[0.000000, 0.000000]") {
		t.Fatalf("openscad output incomplete:\n%s", scad)
	}
}
