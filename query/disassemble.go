// Disassembly: explode composites and flatten nested block references.
package query

import (
	"github.com/edanko/dxf/block"
	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
)

// BlockLookup resolves a block name to its definition (nil when missing).
type BlockLookup func(name string) *block.Block

// Disassembled is the flat output of Disassemble.
type Disassembled struct {
	Entities []entity.Entity
	Cycles   [][]string
	Errors   []error
}

// Disassemble flattens nested INSERT references: every leaf entity is
// copied, given an empty handle, and transformed by the accumulated insert
// matrices. Cyclic references terminate and are reported in Cycles.
func Disassemble(ents []entity.Entity, lookup BlockLookup) (*Disassembled, error) {
	d := &Disassembled{}
	d.walk(ents, dxfmath.Identity(), nil, lookup)
	return d, nil
}

func (d *Disassembled) walk(ents []entity.Entity, m dxfmath.Matrix44, path []string, lookup BlockLookup) {
	for _, e := range ents {
		if u, ok := e.(interface{ Unwrap() entity.Entity }); ok {
			e = u.Unwrap()
		}
		ins, ok := e.(*entity.Insert)
		if !ok {
			d.leaf(e, m)
			continue
		}
		name := ins.BlockName()
		if i := pathIndex(path, name); i >= 0 {
			cycle := append(append([]string{}, path[i:]...), name)
			d.Cycles = append(d.Cycles, cycle)
			continue
		}
		var blk *block.Block
		if lookup != nil {
			blk = lookup(name)
		}
		if blk == nil {
			// Unresolved reference: keep the insert itself as a leaf.
			d.leaf(e, m)
			continue
		}
		nm := m.Mul(ins.InsertMatrix())
		d.walk(blockEntities(blk), nm, append(path, name), lookup)
	}
}

func (d *Disassembled) leaf(e entity.Entity, m dxfmath.Matrix44) {
	if u, ok := e.(interface{ Unwrap() entity.Entity }); ok {
		e = u.Unwrap()
	}
	cp := e
	if c, ok := e.(entity.Copier); ok {
		if c2 := c.Copy(); c2 != nil {
			cp = c2
		}
	}
	if s, ok := cp.(interface{ SetHandleString(string) }); ok {
		s.SetHandleString("")
	}
	if m.IsIdentity(1e-12) {
		d.Entities = append(d.Entities, cp)
		return
	}
	t, ok := cp.(Transformer)
	if !ok {
		d.Errors = append(d.Errors, newTransformError(e, "entity does not support transformation"))
		d.Entities = append(d.Entities, cp)
		return
	}
	if err := t.Transform(m); err != nil {
		d.Errors = append(d.Errors, newTransformError(e, err.Error()))
	}
	d.Entities = append(d.Entities, cp)
}

func blockEntities(b *block.Block) []entity.Entity {
	out := make([]entity.Entity, 0, len(b.Entities()))
	for _, be := range b.Entities() {
		if e, ok := be.(entity.Entity); ok {
			out = append(out, e)
		}
	}
	return out
}

func pathIndex(path []string, name string) int {
	for i, p := range path {
		if p == name {
			return i
		}
	}
	return -1
}

// ExplodeAll explodes composite entities (INSERT, POLYLINE, DIMENSION,
// HATCH, LEADER, MLEADER) into their parts, recursing into exploded
// inserts. Entities without an exploder are kept as-is.
func ExplodeAll(ents []entity.Entity) []entity.Entity {
	var out []entity.Entity
	for _, e := range ents {
		out = appendExploded(out, e)
	}
	return out
}

func appendExploded(out []entity.Entity, e entity.Entity) []entity.Entity {
	ex, ok := e.(entity.Exploder)
	if !ok {
		return append(out, e)
	}
	parts, err := ex.Explode()
	if err != nil || len(parts) == 0 {
		return append(out, e)
	}
	for _, p := range parts {
		out = appendExploded(out, p)
	}
	return out
}
