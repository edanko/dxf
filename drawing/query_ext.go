// Selection helpers (type, layer, spatial extent) built on the query
// engine, and model-space / per-layout extents computation.

package drawing

import (
	"fmt"

	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
	"github.com/edanko/dxf/query"
)

// SelectByType returns model-space entities whose DXF type is one of the
// given names.
func (d *Drawing) SelectByType(types ...string) []entity.Entity {
	q := query.NewEntityQuery(d.LayoutEntities("Model")...)
	sel := query.NewEntitySelector(q)
	return sel.Types(types...).ToSlice()
}

// SelectByLayer returns model-space entities on the named layer.
func (d *Drawing) SelectByLayer(layer string) []entity.Entity {
	q := query.NewEntityQuery(d.LayoutEntities("Model")...)
	sel := query.NewEntitySelector(q)
	return sel.Layer(layer).ToSlice()
}

// SelectInExtents returns model-space entities intersecting the 2D extent
// (minX, minY) to (maxX, maxY).
func (d *Drawing) SelectInExtents(minX, minY, maxX, maxY float64) []entity.Entity {
	q := query.NewEntityQuery(d.LayoutEntities("Model")...)
	sel := query.NewEntitySelector(q)
	return sel.Within(minX, minY, 0, maxX, maxY, 0).ToSlice()
}

// Extents returns the 3D bounding box of model-space entities. The bool
// result is false when no entity has a computable bounding box.
func (d *Drawing) Extents() (min, max dxfmath.Vec3, ok bool) {
	return d.LayoutExtents("Model")
}

// LayoutExtents returns the 3D bounding box of a layout's entities.
func (d *Drawing) LayoutExtents(layout string) (min, max dxfmath.Vec3, ok bool) {
	q := query.NewEntityQuery(d.LayoutEntities(layout)...)
	return q.Bounds()
}

// RecalcExtents recomputes the document extents variables (header EXTMIN
// and EXTMAX) from model-space entities and returns them.
func (d *Drawing) RecalcExtents() (min, max dxfmath.Vec3, ok bool) {
	d.SetExt()
	return d.Extents()
}

// GroupMembers returns the entity handles stored in a group.
func (d *Drawing) GroupMembers(name string) ([]string, error) {
	g, exist := d.Groups[name]
	if !exist {
		return nil, fmt.Errorf("group %s doesn't exist", name)
	}
	return g.EntityHandles(), nil
}
