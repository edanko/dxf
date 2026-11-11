// Package meshex exports mesh entities to OBJ and OpenSCAD.
package meshex

import (
	"fmt"

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

// ExportOBJ renders all mesh entities and decoded ACIS bodies as a Wavefront
// OBJ file.
func ExportOBJ(d *drawing.Drawing) (string, error) {
	if d == nil {
		return "", fmt.Errorf("meshex: nil drawing")
	}
	return CollectGeometry(d).OBJ()
}

// ExportOpenSCAD renders all mesh entities and decoded ACIS bodies as an
// OpenSCAD polyhedron union.
func ExportOpenSCAD(d *drawing.Drawing) (string, error) {
	if d == nil {
		return "", fmt.Errorf("meshex: nil drawing")
	}
	return CollectGeometry(d).OpenSCAD()
}
