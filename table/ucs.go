package table

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/lldxf"
)

// UCS represents UCS SymbolTable.
type Ucs struct {
	EntryBase

	name string // 2
}

// NewUCS creates a new Ucs.
func NewUCS(name string) *Ucs {
	u := new(Ucs)
	u.name = name
	return u
}

// IsSymbolTable is for SymbolTable interface.
func (u *Ucs) IsSymbolTable() bool {
	return true
}

// Format writes data to formatter.
func (u *Ucs) Format(f format.Formatter) {
	u.FormatCommon(f, "UCS", "AcDbUCSTableRecord")
	f.WriteString(100, "AcDbUCSTableRecord")
	f.WriteString(2, u.name)
	u.FormatCommonTail(f)
}

// RetainTags records UCS codes the typed model does not emit.
func (u *Ucs) RetainTags(tags lldxf.Tags) {
	emitted := map[int]bool{0: true, 2: true, 5: true, 70: true, 100: true, 330: true}
	u.RetainUnhandled(tags, emitted)
}

// Name returns a name of UCS (code 2).
func (u *Ucs) Name() string {
	return u.name
}
