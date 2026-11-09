package table

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/lldxf"
)

// Style represents STYLE SymbolTable.
type Style struct {
	EntryBase

	name            string  // 2
	FontName        string  // 3
	BigFontName     string  // 4
	FixedTextHeight float64 // 40
	WidthFactor     float64 // 41
	LastHeightUsed  float64 // 42
	ObliqueAngle    float64 // 50
}

// NewStyle create a new Style.
func NewStyle(name string) *Style {
	return &Style{
		name:            name,
		FontName:        "arial.ttf",
		FixedTextHeight: 0.0,
		WidthFactor:     1.0,
		LastHeightUsed:  100.0,
		ObliqueAngle:    0.0,
	}
}

// IsSymbolTable is for SymbolTable interface.
func (st *Style) IsSymbolTable() bool {
	return true
}

// Format writes data to formatter.
func (st *Style) Format(f format.Formatter) {
	st.FormatCommon(f, "STYLE", "AcDbTextStyleTableRecord")
	f.WriteString(2, st.name)
	f.WriteInt(70, st.flags)
	f.WriteFloat(40, st.FixedTextHeight)
	f.WriteFloat(41, st.WidthFactor)
	f.WriteFloat(50, st.ObliqueAngle)
	f.WriteInt(71, 0)
	f.WriteFloat(42, st.LastHeightUsed)
	f.WriteString(3, st.FontName)
	f.WriteString(4, st.BigFontName)
	st.FormatCommonTail(f)
}

// RetainTags records STYLE codes the typed model does not emit.
func (st *Style) RetainTags(tags lldxf.Tags) {
	emitted := map[int]bool{0: true, 2: true, 3: true, 4: true, 5: true, 40: true, 41: true, 42: true, 50: true, 70: true, 71: true, 100: true, 330: true}
	st.RetainUnhandled(tags, emitted)
}

// Name returns a name of STYLE (code 2).
func (st *Style) Name() string {
	return st.name
}

// SetName sets the style name.
func (st *Style) SetName(name string) { st.name = name }
