package meshex_test

import (
	"strings"
	"testing"

	"github.com/edanko/dxf/addons/meshex"
	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
)

func TestExportOBJAndOpenSCAD(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	mesh := entity.CreateCube(2)
	mesh.SetHandleString("40")
	if err := d.AddEntity(mesh); err != nil {
		t.Fatal(err)
	}

	obj, err := meshex.ExportOBJ(d)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Count("\n"+obj, "\nv "); got != len(mesh.Vertices()) {
		t.Fatalf("obj vertices: %d want %d", got, len(mesh.Vertices()))
	}
	if !strings.Contains(obj, "f ") {
		t.Fatal("obj has no faces")
	}

	scad, err := meshex.ExportOpenSCAD(d)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(scad, "polyhedron") || !strings.Contains(scad, "points=") || !strings.Contains(scad, "faces=") {
		t.Fatalf("openscad output incomplete: %s", scad)
	}
}
