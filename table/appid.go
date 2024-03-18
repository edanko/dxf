package table

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
)

// AppID represents APPID SymbolTable.
type AppID struct {
	handle string
	owner  handle.Handler
	name   string
}

// NewAppID create a new AppID.
func NewAppID(name string) *AppID {
	a := new(AppID)
	a.name = name
	return a
}

// IsSymbolTable is for SymbolTable interface.
func (a *AppID) IsSymbolTable() bool {
	return true
}

// Format writes data to formatter.
func (a *AppID) Format(f format.Formatter) {
	f.WriteString(0, "APPID")
	f.WriteString(5, a.handle)
	if a.owner != nil {
		f.WriteString(330, a.owner.Handle())
	}
	f.WriteString(100, "AcDbSymbolTableRecord")
	f.WriteString(100, "AcDbRegAppTableRecord")
	f.WriteString(2, a.name)
	f.WriteInt(70, 0)
}

// Handle returns a handle value.
func (a *AppID) Handle() string {
	return a.handle
}

// SetHandle sets a handle.
func (a *AppID) SetHandle(hg *handle.HandleGenerator) {
	a.handle = hg.Next()
}

// SetOwner sets an owner.
func (a *AppID) SetOwner(h handle.Handler) {
	a.owner = h
}

// Name returns a name of APPID (code 2).
func (a *AppID) Name() string {
	return a.name
}
