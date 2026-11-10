// Model-space draw order: file (plot) order manipulation and SORTENTSTABLE
// persistence with a pre-R13 fallback.
package drawing

import (
	"fmt"
	"strconv"

	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/object"
)

// DrawOrder returns the model-space entity handles in plot order: the first
// handle is drawn first (back-most), the last is drawn on top (front-most).
func (d *Drawing) DrawOrder() []string {
	out := make([]string, 0, len(d.entitiesSection))
	for _, e := range d.entitiesSection {
		out = append(out, e.Handle())
	}
	return out
}

// SendToBack moves the entity to the start of the plot order (drawn first,
// behind everything) and regenerates the SORTENTSTABLE for model space.
// Pre-R13 drawings (no OBJECTS section) keep the physical reorder only.
func (d *Drawing) SendToBack(handle string) error {
	return d.reorderEntity(handle, "front")
}

// BringToFront moves the entity to the end of the plot order (drawn last, on
// top) and regenerates the SORTENTSTABLE for model space.
func (d *Drawing) BringToFront(handle string) error {
	return d.reorderEntity(handle, "back")
}

// SetDrawOrder applies an explicit front-to-back handle list (the ezdxf zseq
// convention). Every model-space entity handle must appear exactly once.
func (d *Drawing) SetDrawOrder(frontToBack []string) error {
	index := map[string]int{}
	for i, e := range d.entitiesSection {
		index[e.Handle()] = i
	}
	if len(frontToBack) != len(d.entitiesSection) {
		return fmt.Errorf("draw order: got %d handles, model space has %d entities", len(frontToBack), len(d.entitiesSection))
	}
	seen := map[string]bool{}
	ordered := make(entity.Entities, 0, len(frontToBack))
	for i := len(frontToBack) - 1; i >= 0; i-- {
		h := frontToBack[i]
		idx, ok := index[h]
		if !ok {
			return fmt.Errorf("draw order: entity %s is not in model space", h)
		}
		if seen[h] {
			return fmt.Errorf("draw order: handle %s listed twice", h)
		}
		seen[h] = true
		ordered = append(ordered, d.entitiesSection[idx])
	}
	d.entitiesSection = ordered
	return d.regenerateSortEntsTable()
}

func (d *Drawing) reorderEntity(handle string, where string) error {
	idx := -1
	for i, e := range d.entitiesSection {
		if e.Handle() == handle {
			idx = i
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("draw order: entity %s is not in model space", handle)
	}
	e := d.entitiesSection[idx]
	rest := make(entity.Entities, 0, len(d.entitiesSection)-1)
	rest = append(rest, d.entitiesSection[:idx]...)
	rest = append(rest, d.entitiesSection[idx+1:]...)
	if where == "front" { // drawn first: start of plot order
		d.entitiesSection = append(entity.Entities{e}, rest...)
	} else { // drawn last: end of plot order
		d.entitiesSection = append(rest, e)
	}
	return d.regenerateSortEntsTable()
}

// supportsObjects reports whether the drawing version can persist a
// SORTENTSTABLE (R13+).
func (d *Drawing) supportsObjects() bool {
	v := d.writeVersion
	if v == "" {
		v = d.sourceVersion
	}
	if v == "" && d.headerSection != nil {
		v = d.headerSection.Version
	}
	if v == "" {
		return true
	}
	p := format.Profile(v)
	return p == nil || p.SupportsObjects
}

// regenerateSortEntsTable writes the current model-space plot order into the
// OBJECTS section as a SORTENTSTABLE (replacing an existing model-space
// table). Older versions skip the table; the physical entity order alone
// carries the draw order.
func (d *Drawing) regenerateSortEntsTable() error {
	dropExistingModelTable(d)
	if !d.supportsObjects() {
		return nil
	}
	blockRec, ok := d.BlockRecordHandle("*Model_Space")
	if !ok || blockRec == "" {
		return nil
	}
	frontToBack := make([]string, 0, len(d.entitiesSection))
	for i := len(d.entitiesSection) - 1; i >= 0; i-- {
		frontToBack = append(frontToBack, d.entitiesSection[i].Handle())
	}
	if len(frontToBack) == 0 {
		return nil
	}
	d.objectsSection = d.objectsSection.Add(
		object.NewSortEntsTable(d.nextFreeHandle(), blockRec, frontToBack))
	return nil
}

// dropExistingModelTable removes SORTENTSTABLE records bound to the model
// space block record from the objects section.
func dropExistingModelTable(d *Drawing) {
	modelRec, ok := d.BlockRecordHandle("*Model_Space")
	if !ok {
		return
	}
	kept := make(object.Objects, 0, len(d.objectsSection))
	for _, o := range d.objectsSection {
		if st, is := o.(*object.SortEntsTable); is && st.BlockRecordHandle() == modelRec {
			continue
		}
		kept = append(kept, o)
	}
	d.objectsSection = kept
}

// ApplySortEntsTables reorders model-space entities according to the
// SORTENTSTABLE records of the OBJECTS section (applied after load).
func (d *Drawing) ApplySortEntsTables() {
	modelRec, ok := d.BlockRecordHandle("*Model_Space")
	if !ok || modelRec == "" {
		return
	}
	var table *object.StdObject
	for _, o := range d.objectsSection {
		st, is := o.(*object.SortEntsTable)
		if !is {
			continue
		}
		if st.BlockRecordHandle() == modelRec {
			table = st.StdObject
			break
		}
	}
	if table == nil {
		return
	}
	entries := table.SortEntries()
	if len(entries) == 0 {
		return
	}
	byHandle := map[string]entity.Entity{}
	kept := make(entity.Entities, 0, len(d.entitiesSection))
	for _, e := range d.entitiesSection {
		byHandle[e.Handle()] = e
	}
	managed := map[string]bool{}
	for _, en := range entries {
		managed[en[0]] = true
	}
	for _, e := range d.entitiesSection {
		if !managed[e.Handle()] {
			kept = append(kept, e)
		}
	}
	// Entries are stored front to back; replaying them back to front
	// reconstructs plot order: "-2" back-most, "-1" front-most, a handle X
	// means "drawn immediately after X".
	for i := len(entries) - 1; i >= 0; i-- {
		h, sortHandle := entries[i][0], entries[i][1]
		e, ok := byHandle[h]
		if !ok {
			continue
		}
		switch {
		case sortHandle == "-2":
			kept = append(entity.Entities{e}, kept...)
		case sortHandle == "-1":
			kept = append(kept, e)
		default:
			pos := -1
			for j, k := range kept {
				if k.Handle() == sortHandle {
					pos = j
					break
				}
			}
			if pos < 0 {
				kept = append(kept, e)
				continue
			}
			kept = append(kept, nil)
			copy(kept[pos+2:], kept[pos+1:])
			kept[pos+1] = e
		}
	}
	d.entitiesSection = kept
}

// nextFreeHandle returns a handle that is not used anywhere in the drawing.
func (d *Drawing) nextFreeHandle() string {
	max := 0
	bump := func(h string) {
		v, err := strconv.ParseUint(h, 16, 64)
		if err == nil && v > uint64(max) {
			max = int(v)
		}
	}
	for _, e := range d.entitiesSection {
		bump(e.Handle())
	}
	for _, o := range d.objectsSection {
		bump(o.Handle())
	}
	for _, b := range d.blocksSection {
		if b == nil {
			continue
		}
		bump(b.Handle())
		for _, be := range b.Entities() {
			bump(be.Handle())
		}
	}
	if ts := d.tablesSection; ts != nil {
		for _, t := range ts {
			if t == nil {
				continue
			}
			bump(t.Handle())
			for _, en := range t.Entries() {
				bump(en.Handle())
			}
		}
	}
	return fmt.Sprintf("%X", max+1)
}
