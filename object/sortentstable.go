// SORTENTSTABLE construction (write side) for draw-order persistence.
package object

import (
	"github.com/edanko/dxf/lldxf"
)

// SortEntsTable is a standard OBJECTS record of type SORTENTSTABLE.
// (Typed accessor lives on StdObject.)

// NewSortEntsTable builds a SORTENTSTABLE for a block record from the desired
// draw order. frontToBack lists model entity handles with the front-most
// entity first:
//
//	front-most  -> sort handle "-1"
//	back-most   -> sort handle "-2"
//	others      -> sort handle of the entity drawn immediately before it
//
// The record's retained tags reproduce the standard wire layout (100 marker,
// 330 block record, then (331 entity, 5 sort handle) pairs front to back) so
// the object round-trips through Load and ApplySortEntsTables unchanged.
func NewSortEntsTable(objHandle, blockRec string, frontToBack []string) *SortEntsTable {
	s := &StdObject{
		BaseObject: &BaseObject{},
		typeName:   "SORTENTSTABLE",
		blockRec:   blockRec,
		sortBy:     make(map[string]string),
	}
	s.SetWireType("SORTENTSTABLE")
	s.BaseObject.handle = objHandle

	retained := lldxf.NewTags()
	retained = retained.Add(lldxf.NewDXFTag(100, "AcDbSortentsTable"))
	retained = retained.Add(lldxf.NewDXFTag(330, blockRec))
	n := len(frontToBack)
	for i, h := range frontToBack {
		var sortHandle string
		switch {
		case i == 0:
			sortHandle = "-1"
		case i == n-1:
			sortHandle = "-2"
		default:
			sortHandle = frontToBack[i+1]
		}
		s.sortOrder = append(s.sortOrder, h)
		s.sortBy[h] = sortHandle
		retained = retained.Add(lldxf.NewDXFTag(331, h))
		retained = retained.Add(lldxf.NewDXFTag(5, sortHandle))
	}
	s.SetRetainedTags(retained)
	return &SortEntsTable{StdObject: s}
}

// SortEntries returns the (entity handle, sort handle) pairs of a
// SORTENTSTABLE in record source order.
func (s *StdObject) SortEntries() [][2]string {
	out := make([][2]string, 0, len(s.sortOrder))
	for _, h := range s.sortOrder {
		out = append(out, [2]string{h, s.sortBy[h]})
	}
	return out
}
