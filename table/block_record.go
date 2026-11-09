package table

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/lldxf"
)

// BlockRecord represents BLOCK_RECORD SymbolTable.
type BlockRecord struct {
	EntryBase

	name string

	blockScaling int // 280 insert units (0 none, 1 horizontal, 2 vertical, 3 both)
	exploding    int // 281 exploding flags (bit 1: explode yields insert)
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
	b.FormatCommonTail(f)
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
	emitted := map[int]bool{0: true, 2: true, 5: true, 70: true, 100: true, 280: true, 281: true, 330: true, 340: true}
	b.RetainUnhandled(tags, emitted)
}

// Name returns a name of BLOCK_RECORD (code 2).
func (b *BlockRecord) Name() string {
	return b.name
}
