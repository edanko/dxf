package table

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
)

// DimStyle represents DIMSTYLE SymbolTable.
type DimStyle struct {
	handle string
	owner  handle.Handler
	name   string
}

// NewDimStyle creates a new DimStyle.
func NewDimStyle(name string) *DimStyle {
	d := new(DimStyle)
	d.name = name
	return d
}

// IsSymbolTable is for SymbolTable interface.
func (d *DimStyle) IsSymbolTable() bool {
	return true
}

// Format writes data to formatter.
func (d *DimStyle) Format(f format.Formatter) {
	f.WriteString(0, "DIMSTYLE")
	f.WriteString(105, d.handle)
	if d.owner != nil {
		f.WriteString(330, d.owner.Handle())
	}
	f.WriteString(100, "AcDbSymbolTableRecord")
	f.WriteString(100, "AcDbDimStyleTableRecord")
	f.WriteString(2, d.name)
	f.WriteInt(70, 0)
}

// Handle returns a handle value.
func (d *DimStyle) Handle() string {
	return d.handle
}

// SetHandle sets a handle.
func (d *DimStyle) SetHandle(hg *handle.HandleGenerator) {
	d.handle = hg.Next()
}

// SetOwner sets an owner.
func (d *DimStyle) SetOwner(h handle.Handler) {
	d.owner = h
}

// Name returns a name of DIMSTYLE (code 2).
func (d *DimStyle) Name() string {
	return d.name
}
