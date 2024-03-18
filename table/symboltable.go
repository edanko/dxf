package table

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
)

// SymbolTable is interface for AcDbSymbolTableRecord.
type SymbolTable interface {
	IsSymbolTable() bool
	Format(format.Formatter)
	SetOwner(handle.Handler)
	Name() string

	handle.Handler
}
