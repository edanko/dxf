package table

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/lldxf"
)

// View represents VIEW SymbolTable.
type View struct {
	EntryBase

	name string // 2
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
	v.FormatCommon(f, "VIEW", "AcDbViewTableRecord")
	f.WriteString(100, "AcDbViewTableRecord")
	f.WriteString(2, v.name)
	v.FormatCommonTail(f)
}

// RetainTags records VIEW codes the typed model does not emit.
func (v *View) RetainTags(tags lldxf.Tags) {
	emitted := map[int]bool{0: true, 2: true, 5: true, 70: true, 100: true, 330: true}
	v.RetainUnhandled(tags, emitted)
}

// Name returns a name of VIEW (code 2).
func (v *View) Name() string {
	return v.name
}
