package table

import (
	"sort"

	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/lldxf"
)

// DimStyle represents DIMSTYLE SymbolTable. Standard dimension-variable
// values are stored generically by group code so every documented setting
// (DIMASSOC, DIMTXT, DIMCLRD, ...) round-trips with its DXF type.
type DimStyle struct {
	EntryBase

	name string
	vars map[int]string // group code -> string value of dimension settings
	tail lldxf.Tags     // DIMSTYLE table header tail (71 count + 340 links), emitted verbatim on the last record
}

// SetTableTail attaches the DIMSTYLE table header tail (code-71 style count
// followed by one code-340 link per style) to this record. AutoCAD writes it
// at the end of the last style record; retaining it verbatim — in source
// order, never as sorted vars — keeps the table round-trip byte-stable.
func (d *DimStyle) SetTableTail(tags lldxf.Tags) { d.tail = tags }

// TableTail returns the attached header tail, if any.
func (d *DimStyle) TableTail() lldxf.Tags { return d.tail }

// NewDimStyle creates a new DimStyle.
func NewDimStyle(name string) *DimStyle {
	d := new(DimStyle)
	d.name = name
	d.vars = map[int]string{}
	return d
}

// SetVar sets a dimension setting by group code (e.g. 40 for DIMSCALE).
func (d *DimStyle) SetVar(code int, value string) {
	if d.vars == nil {
		d.vars = map[int]string{}
	}
	d.vars[code] = value
}

// Var returns a dimension setting by group code.
func (d *DimStyle) Var(code int) (string, bool) {
	v, ok := d.vars[code]
	return v, ok
}

// Vars returns all settings keyed by group code.
func (d *DimStyle) Vars() map[int]string { return d.vars }

// IsSymbolTable is for SymbolTable interface.
func (d *DimStyle) IsSymbolTable() bool {
	return true
}

// Format writes data to formatter.
func (d *DimStyle) Format(f format.Formatter) {
	// Every record begins with a code-0 type marker so the loader can
	// separate records; pre-R2010 AutoCAD uses code 105 as the handle.
	f.WriteString(0, "DIMSTYLE")
	if h := d.Handle(); h != "" {
		f.WriteString(105, h)
	}
	if len(d.ownerHandle()) > 0 {
		f.WriteString(330, d.ownerHandle())
	}
	f.WriteString(100, "AcDbSymbolTableRecord")
	f.WriteString(100, "AcDbDimStyleTableRecord")
	f.WriteString(2, d.name)
	f.WriteInt(70, d.flags)
	for _, code := range sortedCodes(d.vars) {
		f.WriteString(code, d.vars[code])
	}
	if x := d.ExtDictHandle(); x != "" {
		f.WriteString(350, x)
	}
	for _, t := range d.tail {
		f.WriteString(int(t.Code()), tagString(t))
	}
	for _, t := range d.RawTags() {
		f.WriteString(t.Code, t.Value)
	}
}

// RetainTags records DIMSTYLE codes the typed model does not emit.
func (d *DimStyle) RetainTags(tags lldxf.Tags) {
	emitted := map[int]bool{0: true, 2: true, 70: true, 100: true, 105: true, 330: true, 350: true}
	for code := range d.vars {
		emitted[code] = true
	}
	for _, t := range d.tail {
		emitted[int(t.Code())] = true
	}
	d.RetainUnhandled(tags, emitted)
}

// ownerHandle returns the owner handle string (empty when unset).
func (d *DimStyle) ownerHandle() string {
	if d.owner != nil {
		return d.owner.Handle()
	}
	return ""
}

func sortedCodes(m map[int]string) []int {
	codes := make([]int, 0, len(m))
	for c := range m {
		codes = append(codes, c)
	}
	sort.Ints(codes)
	return codes
}

// Name returns a name of DIMSTYLE (code 2).
func (d *DimStyle) Name() string {
	return d.name
}
