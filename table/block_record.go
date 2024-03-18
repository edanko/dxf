package table

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
)

// BlockRecord represents BLOCK_RECORD SymbolTable.
type BlockRecord struct {
	handle string
	owner  handle.Handler
	name   string
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
	f.WriteString(0, "BLOCK_RECORD")
	f.WriteString(5, b.handle)
	if b.owner != nil {
		f.WriteString(330, b.owner.Handle())
	}
	f.WriteString(100, "AcDbSymbolTableRecord")
	f.WriteString(100, "AcDbBlockTableRecord")
	f.WriteString(2, b.name)
	f.WriteInt(70, 0)
	f.WriteInt(280, 1)
	f.WriteInt(281, 0)
}

// Handle returns a handle value.
func (b *BlockRecord) Handle() string {
	return b.handle
}

// SetHandle sets a handle.
func (b *BlockRecord) SetHandle(hg *handle.HandleGenerator) {
	b.handle = hg.Next()
}

// SetOwner sets an owner.
func (b *BlockRecord) SetOwner(h handle.Handler) {
	b.owner = h
}

// Name returns a name of BLOCK_RECORD (code 2).
func (b *BlockRecord) Name() string {
	return b.name
}
