package table

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/lldxf"
)

// BLOCK_RECORD flag bits (code 70) for external references.
const (
	BlockFlagXRef           = 4   // record defines an external reference
	BlockFlagXRefDependent  = 8   // dependent (bind) external reference
	BlockFlagCloned         = 16  // cloned from an external reference
	BlockFlagXRefReferenced = 32  // referenced by other blocks
	BlockFlagXRefResolved   = 64  // xref is resolved (content is current)
	BlockFlagXRefOverlay    = 128 // xref is an overlay, not an attachment
)

// BlockRecord represents BLOCK_RECORD SymbolTable.
type BlockRecord struct {
	EntryBase

	name string

	blockScaling int    // 280 insert units (0 none, 1 horizontal, 2 vertical, 3 both)
	exploding    int    // 281 exploding flags (bit 1: explode yields insert)
	xrefPath     string // 305 xref file path
}

// NewBlockRecord creates a new BlockRecord.
func NewBlockRecord(name string) *BlockRecord {
	b := new(BlockRecord)
	b.name = name
	return b
}

// IsSymbolTable is for SymbolTable interface.
func (b *BlockRecord) IsSymbolTable() bool {
	return true
}

// Format writes data to formatter.
func (b *BlockRecord) Format(f format.Formatter) {
	b.FormatCommon(f, "BLOCK_RECORD", "AcDbBlockTableRecord")
	f.WriteString(100, "AcDbBlockTableRecord")
	f.WriteString(2, b.name)
	f.WriteInt(70, b.flags)
	f.WriteInt(280, b.blockScaling)
	f.WriteInt(281, b.exploding)
	if b.xrefPath != "" {
		f.WriteString(305, b.xrefPath)
	}
	b.FormatCommonTail(f)
}

// XRefPath returns the external reference file path (code 305).
func (b *BlockRecord) XRefPath() string { return b.xrefPath }

// SetXRefPath sets the external reference file path (code 305).
func (b *BlockRecord) SetXRefPath(p string) { b.xrefPath = p }

// IsXRef reports whether the record defines an external reference (bit 4).
func (b *BlockRecord) IsXRef() bool { return b.flags&BlockFlagXRef != 0 }

// SetXRef sets the external-reference flag. Clearing it also clears the
// dependent, referenced and overlay bits, keeping the record consistent as a
// local definition.
func (b *BlockRecord) SetXRef(v bool) {
	if v {
		b.flags |= BlockFlagXRef
		return
	}
	b.flags &^= BlockFlagXRef | BlockFlagXRefDependent | BlockFlagXRefReferenced | BlockFlagXRefOverlay
}

// IsXRefOverlay reports whether the xref is an overlay (bit 128).
func (b *BlockRecord) IsXRefOverlay() bool { return b.flags&BlockFlagXRefOverlay != 0 }

// SetXRefOverlay sets the overlay flag (bit 128).
func (b *BlockRecord) SetXRefOverlay(v bool) {
	if v {
		b.flags |= BlockFlagXRefOverlay
	} else {
		b.flags &^= BlockFlagXRefOverlay
	}
}

// IsXRefResolved reports whether the xref content is resolved (bit 64).
func (b *BlockRecord) IsXRefResolved() bool { return b.flags&BlockFlagXRefResolved != 0 }

// SetXRefResolved sets the resolved flag (bit 64).
func (b *BlockRecord) SetXRefResolved(v bool) {
	if v {
		b.flags |= BlockFlagXRefResolved
	} else {
		b.flags &^= BlockFlagXRefResolved
	}
}

// IsXRefDependent reports whether the xref is bind-dependent (bit 8).
func (b *BlockRecord) IsXRefDependent() bool { return b.flags&BlockFlagXRefDependent != 0 }

// SetXRefDependent sets the dependent flag (bit 8).
func (b *BlockRecord) SetXRefDependent(v bool) {
	if v {
		b.flags |= BlockFlagXRefDependent
	} else {
		b.flags &^= BlockFlagXRefDependent
	}
}

// BlockScaling returns the insert-units flag (code 280).
func (b *BlockRecord) BlockScaling() int { return b.blockScaling }

// SetBlockScaling sets the insert-units flag (code 280).
func (b *BlockRecord) SetBlockScaling(v int) { b.blockScaling = v }

// Exploding returns the exploding flags (code 281).
func (b *BlockRecord) Exploding() int { return b.exploding }

// SetExploding sets the exploding flags (code 281).
func (b *BlockRecord) SetExploding(v int) { b.exploding = v }

// RetainTags records BLOCK_RECORD codes the typed model does not emit.
func (b *BlockRecord) RetainTags(tags lldxf.Tags) {
	emitted := map[int]bool{0: true, 2: true, 5: true, 70: true, 100: true, 280: true, 281: true, 305: true, 330: true, 340: true}
	b.RetainUnhandled(tags, emitted)
}

// Name returns a name of BLOCK_RECORD (code 2).
func (b *BlockRecord) Name() string {
	return b.name
}

// SetName renames the BLOCK_RECORD entry (code 2).
func (b *BlockRecord) SetName(name string) {
	b.name = name
}
