package table

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/lldxf"
)

// AppID represents APPID SymbolTable.
type AppID struct {
	EntryBase

	name string
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
	a.FormatCommon(f, "APPID", "AcDbRegAppTableRecord")
	f.WriteString(100, "AcDbRegAppTableRecord")
	f.WriteString(2, a.name)
	f.WriteInt(70, a.flags)
	a.FormatCommonTail(f)
}

// RetainTags records APPID codes the typed model does not emit.
func (a *AppID) RetainTags(tags lldxf.Tags) {
	emitted := map[int]bool{0: true, 2: true, 5: true, 70: true, 100: true, 330: true}
	a.RetainUnhandled(tags, emitted)
}

// Name returns a name of APPID (code 2).
func (a *AppID) Name() string {
	return a.name
}
