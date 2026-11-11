// Document-wide query, selection-set, transform, disassembly and upright
// operations.
package drawing

import (
	"sort"

	"github.com/edanko/dxf/block"
	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
	"github.com/edanko/dxf/query"
)

// queryableEntities returns the model-space entities plus entity-database
// members, deduplicated by handle.
func (d *Drawing) queryableEntities() []entity.Entity {
	seen := map[string]bool{}
	var out []entity.Entity
	add := func(list []entity.Entity) {
		for _, e := range list {
			if e == nil {
				continue
			}
			h := e.Handle()
			if h != "" && seen[h] {
				continue
			}
			if h != "" {
				seen[h] = true
			}
			out = append(out, e)
		}
	}
	add(d.entitiesSection)
	if d.entityDatabase != nil {
		add(d.entityDatabase.GetAllEntities())
	}
	return out
}

// QueryExpression evaluates a query expression ("layer == 'Walls' and
// bbox_intersects(...)") over the document entities with spatial
// pre-filtering, returning the matching selection.
func (d *Drawing) QueryExpression(expr string) (*query.Selection, error) {
	return query.ExecuteWithFilter(d.queryableEntities(), expr)
}

// QueryEntities evaluates a compiled expression over the document entities.
func (d *Drawing) QueryEntities(expr query.Expression) (*query.Selection, error) {
	return query.Execute(d.queryableEntities(), expr)
}

// SelectionSet returns a stored named selection set.
func (d *Drawing) SelectionSet(name string) (*query.Selection, bool) {
	s, ok := d.selectionSets[name]
	return s, ok
}

// SetSelectionSet stores (or replaces) a named selection set.
func (d *Drawing) SetSelectionSet(name string, sel *query.Selection) {
	if d.selectionSets == nil {
		d.selectionSets = map[string]*query.Selection{}
	}
	sel.SetName(name)
	d.selectionSets[name] = sel
}

// SelectionSetNames returns the stored selection set names in lexical order.
func (d *Drawing) SelectionSetNames() []string {
	names := make([]string, 0, len(d.selectionSets))
	for n := range d.selectionSets {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// DeleteSelectionSet removes a named selection set.
func (d *Drawing) DeleteSelectionSet(name string) bool {
	if _, ok := d.selectionSets[name]; !ok {
		return false
	}
	delete(d.selectionSets, name)
	return true
}

// TransformSelection applies a matrix to the entities of a selection.
// Entities that do not support transformation are reported per handle.
func (d *Drawing) TransformSelection(sel *query.Selection, m dxfmath.Matrix44) []error {
	return sel.ApplyM(m, d)
}

// TransformBlock applies a matrix to every entity of a block definition and
// to its base point. All INSERT references of the block reflect the new
// definition geometry.
func (d *Drawing) TransformBlock(name string, m dxfmath.Matrix44) []error {
	blk, err := d.GetBlock(name)
	if err != nil {
		return []error{err}
	}
	ents := make([]entity.Entity, 0, len(blk.Entities()))
	for _, be := range blk.Entities() {
		if e, ok := be.(entity.Entity); ok {
			ents = append(ents, e)
		}
	}
	errs := query.Apply(m, ents...)
	if len(blk.Coord) >= 2 {
		v := dxfmath.Vec3{blk.Coord[0], blk.Coord[1], coordZ(blk.Coord)}
		p := m.MulVec(v)
		blk.Coord = []float64{p[0], p[1], p[2]}
	}
	return errs
}

// TransformDrawing applies a matrix to every model-space entity and to every
// block definition (including its base point).
func (d *Drawing) TransformDrawing(m dxfmath.Matrix44) []error {
	var errs []error
	errs = append(errs, query.Apply(m, d.entitiesSection...)...)
	for _, b := range d.blocksSection {
		if b == nil {
			continue
		}
		errs = append(errs, d.TransformBlock(b.Name, m)...)
	}
	return errs
}

func coordZ(c []float64) float64 {
	if len(c) >= 3 {
		return c[2]
	}
	return 0
}

// blockLookup resolves block names through the document.
func (d *Drawing) blockLookup() query.BlockLookup {
	return func(name string) *block.Block {
		b, err := d.GetBlock(name)
		if err != nil {
			return nil
		}
		return b
	}
}

// DisassembleEntities flattens nested inserts of the given entities.
func (d *Drawing) DisassembleEntities(ents []entity.Entity) (*query.Disassembled, error) {
	return query.Disassemble(ents, d.blockLookup())
}

// DisassembleModel flattens all model-space entities into a primitive list.
func (d *Drawing) DisassembleModel() (*query.Disassembled, error) {
	return query.Disassemble(d.entitiesSection, d.blockLookup())
}

// TraverseBlockRefs walks model-space INSERT references (and nested ones) in
// traversal order and returns any reference cycles.
func (d *Drawing) TraverseBlockRefs(visit func(depth int, ins *entity.Insert, path []string) bool) [][]string {
	return query.TraverseBlockRefs(d.entitiesSection, d.blockLookup(), visit)
}

// MakeUpright normalizes one entity (by handle) to a readable rotation,
// preserving its insertion point. It reports whether the entity changed.
func (d *Drawing) MakeUpright(handle string) (bool, error) {
	e, ok := d.GetEntity(handle)
	if !ok {
		for _, be := range d.entitiesSection {
			if be.Handle() == handle {
				e, ok = be, true
				break
			}
		}
	}
	if !ok {
		return false, entityNotFoundError(handle)
	}
	return entity.MakeUpright(e), nil
}

// MakeUprightAll normalizes every text-like entity in model space and
// returns the number of entities changed.
func (d *Drawing) MakeUprightAll() int {
	n := 0
	for _, e := range d.entitiesSection {
		if entity.MakeUpright(e) {
			n++
		}
	}
	return n
}

type entityNotFoundError string

func (e entityNotFoundError) Error() string { return "entity " + string(e) + " not found" }
