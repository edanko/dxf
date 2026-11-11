package meshex

import (
	"fmt"
	"strings"

	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/edges"
	"github.com/edanko/dxf/entity"
)

// Geometry is the decoded export input: mesh sources plus closed linked-edge
// loops. It lets exporters consume geometry that is not represented by a
// direct mesh entity, such as tessellated ACIS bodies.
type Geometry struct {
	Meshes      []MeshSource
	Loops       []edges.Loop
	Diagnostics []string
}

// NewGeometry returns an empty export geometry.
func NewGeometry() *Geometry { return &Geometry{} }

// AddMesh appends a mesh source.
func (g *Geometry) AddMesh(m MeshSource) {
	if g == nil || m == nil {
		return
	}
	g.Meshes = append(g.Meshes, m)
}

// AddLoops appends closed linked-edge loops in source orientation.
func (g *Geometry) AddLoops(loops ...edges.Loop) {
	if g == nil {
		return
	}
	for _, loop := range loops {
		if len(loop.Vertices) >= 3 {
			g.Loops = append(g.Loops, loop)
		}
	}
}

// AddEntity routes a mesh entity or an ACIS body through decoded geometry.
func (g *Geometry) AddEntity(e entity.Entity) error {
	if e == nil {
		return nil
	}
	if mesh, ok := e.(MeshSource); ok {
		g.AddMesh(mesh)
		return nil
	}
	mesh, err := entity.DecodeACISMesh(e)
	if err != nil {
		return err
	}
	g.AddMesh(mesh)
	return nil
}

// CollectGeometry decodes every mesh-like and ACIS-bearing drawing entity.
// ACIS entities without decodable flat SAT geometry are reported in
// Diagnostics and skipped.
func CollectGeometry(d *drawing.Drawing) *Geometry {
	g := NewGeometry()
	if d == nil {
		return g
	}
	for _, e := range d.Entities() {
		if err := g.AddEntity(e); err != nil {
			g.Diagnostics = append(g.Diagnostics, err.Error())
		}
	}
	return g
}

// VertexCount reports mesh and closed-loop vertex counts.
func (g *Geometry) VertexCount() int {
	if g == nil {
		return 0
	}
	total := 0
	for _, mesh := range g.Meshes {
		total += len(mesh.Vertices())
	}
	for _, loop := range g.Loops {
		total += len(loop.Vertices)
	}
	return total
}

// TriangleCount reports the triangulated face count of meshes and loops.
func (g *Geometry) TriangleCount() int {
	if g == nil {
		return 0
	}
	total := 0
	for _, mesh := range g.Meshes {
		for _, face := range mesh.Faces() {
			if len(face) >= 3 {
				total += len(face) - 2
			}
		}
	}
	for _, loop := range g.Loops {
		if len(loop.Vertices) >= 3 {
			total += len(loop.Vertices) - 2
		}
	}
	return total
}

// OBJ renders decoded meshes and closed loops as Wavefront OBJ.
func (g *Geometry) OBJ() (string, error) {
	if g == nil {
		return "", fmt.Errorf("meshex: nil geometry")
	}
	var b strings.Builder
	offset := 0
	for _, mesh := range g.Meshes {
		for _, v := range mesh.Vertices() {
			fmt.Fprintf(&b, "v %.6f %.6f %.6f\n", v.X(), v.Y(), v.Z())
		}
		for _, face := range mesh.Faces() {
			b.WriteString("f")
			for _, index := range face {
				fmt.Fprintf(&b, " %d", index+1+offset)
			}
			b.WriteString("\n")
		}
		offset += len(mesh.Vertices())
	}
	for _, loop := range g.Loops {
		for _, vertex := range loop.Vertices {
			fmt.Fprintf(&b, "v %.6f %.6f 0.000000\n", vertex.X(), vertex.Y())
		}
		b.WriteString("f")
		for i := range loop.Vertices {
			fmt.Fprintf(&b, " %d", offset+i+1)
		}
		b.WriteString("\n")
		offset += len(loop.Vertices)
	}
	return b.String(), nil
}

// OpenSCAD renders decoded meshes as polyhedra and closed loops as polygons.
func (g *Geometry) OpenSCAD() (string, error) {
	if g == nil {
		return "", fmt.Errorf("meshex: nil geometry")
	}
	var b strings.Builder
	b.WriteString("union() {\n")
	for _, mesh := range g.Meshes {
		b.WriteString("  polyhedron(points=[\n")
		for _, v := range mesh.Vertices() {
			fmt.Fprintf(&b, "    [%.6f, %.6f, %.6f],\n", v.X(), v.Y(), v.Z())
		}
		b.WriteString("  ], faces=[\n")
		for _, face := range mesh.Faces() {
			b.WriteString("    [")
			for i, index := range face {
				if i > 0 {
					b.WriteString(", ")
				}
				fmt.Fprintf(&b, "%d", index)
			}
			b.WriteString("],\n")
		}
		b.WriteString("  ]);\n")
	}
	for _, loop := range g.Loops {
		b.WriteString("  polygon(points=[")
		for i, vertex := range loop.Vertices {
			if i > 0 {
				b.WriteString(", ")
			}
			fmt.Fprintf(&b, "[%.6f, %.6f]", vertex.X(), vertex.Y())
		}
		b.WriteString("]);\n")
	}
	b.WriteString("}\n")
	return b.String(), nil
}

// ExportGeometryOBJ exports arbitrary decoded geometry.
func ExportGeometryOBJ(g *Geometry) (string, error) { return g.OBJ() }

// ExportGeometryOpenSCAD exports arbitrary decoded geometry.
func ExportGeometryOpenSCAD(g *Geometry) (string, error) { return g.OpenSCAD() }

// ExportOBJWithLoops exports drawing meshes, decoded ACIS bodies, and the
// supplied closed edge loops.
func ExportOBJWithLoops(d *drawing.Drawing, loops []edges.Loop) (string, error) {
	g := CollectGeometry(d)
	g.AddLoops(loops...)
	return g.OBJ()
}

// ExportOpenSCADWithLoops exports drawing meshes, decoded ACIS bodies, and the
// supplied closed edge loops.
func ExportOpenSCADWithLoops(d *drawing.Drawing, loops []edges.Loop) (string, error) {
	g := CollectGeometry(d)
	g.AddLoops(loops...)
	return g.OpenSCAD()
}
