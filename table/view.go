package table

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
)

// View represents VIEW SymbolTable.
type View struct {
	handle string
	owner  handle.Handler
	name   string // 2
}

// NewView creates a new View.
func NewView(name string) *View {
	v := new(View)
	v.name = name
	return v
}

// IsSymbolTable is for SymbolTable interface.
func (v *View) IsSymbolTable() bool {
	return true
}

// Format writes data to formatter.
func (v *View) Format(f format.Formatter) {
	f.WriteString(0, "VIEW")
	f.WriteString(5, v.handle)
	if v.owner != nil {
		f.WriteString(330, v.owner.Handle())
	}
	f.WriteString(100, "AcDbSymbolTableRecord")
	f.WriteString(100, "AcDbViewTableRecord")
	f.WriteString(2, v.name)
}

// Handle returns a handle value.
func (v *View) Handle() string {
	return v.handle
}

// SetHandle sets a handle.
func (v *View) SetHandle(hg *handle.HandleGenerator) {
	v.handle = hg.Next()
}

// SetOwner sets an owner.
func (v *View) SetOwner(h handle.Handler) {
	v.owner = h
}

// Name returns a name of VIEW (code 2).
func (v *View) Name() string {
	return v.name
}
