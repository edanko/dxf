// Package meshex exports mesh entities to OBJ and OpenSCAD.
package meshex

import (
	"fmt"
	"strings"

	"github.com/edanko/dxf/addons"
	"github.com/edanko/dxf/drawing"
	dxfmath "github.com/edanko/dxf/math"
)

func init() { _ = addons.Register("meshex", capability{}) }

type capability struct{}

func (capability) Name() string        { return "meshex" }
func (capability) Description() string { return "mesh export (OBJ, OpenSCAD)" }

// MeshSource is implemented by mesh-like entities.
type MeshSource interface {
	Vertices() []dxfmath.Vec3
	Faces() [][]int
}

func collect(d *drawing.Drawing) []MeshSource {
	var out []MeshSource
	for _, e := range d.Entities() {
		if m, ok := e.(MeshSource); ok {
			out = append(out, m)
		}
	}
	return out
}

// ExportOBJ renders all mesh entities as a Wavefront OBJ file.
func ExportOBJ(d *drawing.Drawing) (string, error) {
	if d == nil {
		return "", fmt.Errorf("meshex: nil drawing")
	}
	var b strings.Builder
	offset := 0
	for _, m := range collect(d) {
		for _, v := range m.Vertices() {
			fmt.Fprintf(&b, "v %.6f %.6f %.6f\n", v.X(), v.Y(), v.Z())
		}
		for _, f := range m.Faces() {
			b.WriteString("f")
			for _, idx := range f {
				fmt.Fprintf(&b, " %d", idx+1+offset)
			}
			b.WriteString("\n")
		}
		offset += len(m.Vertices())
	}
	return b.String(), nil
}

// ExportOpenSCAD renders all mesh entities as an OpenSCAD polyhedron union.
func ExportOpenSCAD(d *drawing.Drawing) (string, error) {
	if d == nil {
		return "", fmt.Errorf("meshex: nil drawing")
	}
	var b strings.Builder
	b.WriteString("union() {\n")
	for _, m := range collect(d) {
		b.WriteString("  polyhedron(points=[\n")
		for _, v := range m.Vertices() {
			fmt.Fprintf(&b, "    [%.6f, %.6f, %.6f],\n", v.X(), v.Y(), v.Z())
		}
		b.WriteString("  ], faces=[\n")
		for _, f := range m.Faces() {
			b.WriteString("    [")
			for i, idx := range f {
				if i > 0 {
					b.WriteString(", ")
				}
				fmt.Fprintf(&b, "%d", idx)
			}
			b.WriteString("],\n")
		}
		b.WriteString("  ]);\n")
	}
	b.WriteString("}\n")
	return b.String(), nil
}
