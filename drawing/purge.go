// Block purge: reachability from model and paper layouts through INSERT
// references resolved via the typed entity API.
package drawing

import (
	"github.com/edanko/dxf/block"
	"github.com/edanko/dxf/entity"
)

// Block DXF flags marking blocks that must never be purged.
const (
	blockFlagAnonymous     = 1
	blockFlagXref          = 4
	blockFlagXrefOverlay   = 8
	blockFlagXrefDependent = 16
)

// PurgedBlock describes one block removed by a purge.
type PurgedBlock struct {
	Name   string
	Reason string
}

// PurgeReport summarizes a block purge operation.
type PurgeReport struct {
	Removed []PurgedBlock
}

// RemovedCount returns the number of blocks removed by the purge.
func (r *PurgeReport) RemovedCount() int {
	if r == nil {
		return 0
	}
	return len(r.Removed)
}

// PurgeUnusedBlocksReport removes blocks that are unreachable from any layout
// or from a referenced block definition, and returns a report naming each
// removed block with a reason.
//
// Blocks are kept when they are
//   - one of the standard space blocks (*Model_Space, *Paper_Space,
//     *Paper_Space0),
//   - owned by a paper-space layout,
//   - anonymous or an external reference (xref) per their DXF flags, or
//   - reachable through a chain of INSERT references starting from the
//     entities of the model space and every paper-space layout.
//
// INSERT block names are resolved through the typed entity API
// (entity.Insert.BlockName), so nested block dependencies stay reachable even
// when their only referencing block would itself be removed: reachability is
// computed before anything is deleted.
func (d *Drawing) PurgeUnusedBlocksReport() *PurgeReport {
	report := &PurgeReport{}
	if d == nil {
		return report
	}

	used := d.usedBlockNames()

	var kept []*block.Block
	for _, b := range d.blocksSection {
		name := b.Name
		switch {
		case reservedBlockNames[name]:
			kept = append(kept, b)
		case d.isLayoutOwnedBlock(name):
			kept = append(kept, b)
		case b.Flag&(blockFlagAnonymous|blockFlagXref|blockFlagXrefOverlay|blockFlagXrefDependent) != 0:
			kept = append(kept, b)
		case used[name]:
			kept = append(kept, b)
		default:
			report.Removed = append(report.Removed, PurgedBlock{
				Name:   name,
				Reason: "not referenced by any layout or block definition",
			})
		}
	}

	if len(report.Removed) == 0 {
		return report
	}

	d.blocksSection = kept
	if t := d.blockRecordTable(); t != nil {
		for _, removed := range report.Removed {
			if reservedBlockNames[removed.Name] {
				continue
			}
			_ = t.Remove(removed.Name)
		}
	}
	return report
}

// PurgeUnusedBlocks removes unused blocks and returns the count of removed
// blocks. See PurgeUnusedBlocksReport for the removal rules and a detailed
// report.
func (d *Drawing) PurgeUnusedBlocks() int {
	return d.PurgeUnusedBlocksReport().RemovedCount()
}

// usedBlockNames computes the set of block names reachable from the model
// space and every paper-space layout through INSERT references, recursing into
// block definitions.
func (d *Drawing) usedBlockNames() map[string]bool {
	used := make(map[string]bool)
	visited := make(map[string]bool)

	var visitBlock func(name string)
	visitBlock = func(name string) {
		if name == "" || visited[name] {
			return
		}
		visited[name] = true
		used[name] = true
		blk, err := d.GetBlock(name)
		if err != nil {
			return
		}
		for _, be := range blk.Entities() {
			if ins, ok := insertOf(be); ok {
				visitBlock(ins.BlockName())
			}
		}
	}

	roots := d.LayoutEntities("Model")
	for _, info := range d.Layouts() {
		if info.Type == "PaperSpace" {
			roots = append(roots, d.LayoutEntities(info.Name)...)
		}
	}
	for _, ent := range roots {
		if ins, ok := insertOf(ent); ok {
			visitBlock(ins.BlockName())
		}
	}
	return used
}

// insertOf returns the entity as *entity.Insert when possible. Block
// definitions store entities through the BlockEntity interface, which wraps
// typed entities in an adapter exposing Unwrap when the type does not satisfy
// the interface directly.
func insertOf(e any) (*entity.Insert, bool) {
	switch v := e.(type) {
	case *entity.Insert:
		return v, true
	case interface{ Unwrap() entity.Entity }:
		return insertOf(v.Unwrap())
	}
	return nil, false
}

// isLayoutOwnedBlock reports whether name is the backing block of a
// paper-space layout.
func (d *Drawing) isLayoutOwnedBlock(name string) bool {
	for _, le := range d.layouts {
		if le != nil && le.blockName == name {
			return true
		}
	}
	return false
}
