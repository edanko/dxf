package table

import (
	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/lldxf"
)

// Layer represents LAYER SymbolTable.
type Layer struct {
	EntryBase

	name         string
	Color        color.ColorNumber
	LineType     *LineType
	lineWidth    int
	PlotStyle    handle.Handler
	transparency int // -3 by default, 0-100 percent
	plottable    bool
	trueColor    int // 420 (RGB), -1 = not set
}

// NewLayer creates a new Layer.
func NewLayer(name string, color color.ColorNumber, lt *LineType) *Layer {
	l := new(Layer)
	l.name = name
	l.Color = color
	l.LineType = lt
	l.lineWidth = -3
	l.transparency = -3
	l.plottable = true
	l.trueColor = -1
	return l
}

// IsSymbolTable is for SymbolTable interface.
func (l *Layer) IsSymbolTable() bool {
	return true
}

// Format writes data to formatter.
func (l *Layer) Format(f format.Formatter) {
	l.FormatCommon(f, "LAYER", "AcDbLayerTableRecord")
	f.WriteString(2, l.name)
	f.WriteInt(70, l.flags)
	f.WriteInt(62, int(l.Color))
	if l.LineType != nil {
		f.WriteString(6, l.LineType.Name())
	}
	f.WriteInt(370, l.lineWidth)
	if l.transparency > -3 {
		f.WriteInt(440, l.transparency)
	}
	if !l.plottable {
		f.WriteInt(290, 0)
	}
	if l.trueColor >= 0 {
		f.WriteInt(420, l.trueColor)
	}
	if l.PlotStyle != nil {
		f.WriteString(390, l.PlotStyle.Handle())
	}
	l.FormatCommonTail(f)
}

// RetainTags records LAYER codes the typed model does not emit.
func (l *Layer) RetainTags(tags lldxf.Tags) {
	emitted := map[int]bool{0: true, 2: true, 5: true, 6: true, 62: true, 70: true, 100: true, 290: true, 330: true, 370: true, 390: true, 420: true, 440: true}
	l.RetainUnhandled(tags, emitted)
}

// Name returns a name of LAYER (code 2).
func (l *Layer) Name() string {
	return l.name
}

// SetName sets the layer name.
func (l *Layer) SetName(name string) { l.name = name }

// SetLineWidth sets line width.
// As DXF has limitation in line width,
// it returns the actual value set to Layer.
func (l *Layer) SetLineWidth(w int) int {
	if _, ok := LineWidth[w]; ok {
		l.lineWidth = w
		return w
	}
	if w > 211 {
		l.lineWidth = 211
		return 211
	}
	if w < 0 {
		l.lineWidth = -3
		return -3
	}
	minkey := -3
	minval := 211
	for k := range LineWidth {
		tmp := k - w
		if tmp > 0 && tmp < minval {
			minkey = k
			minval = tmp
		}
	}
	l.lineWidth = minkey
	return minkey
}

// LineWidth returns the layer line width code (negative values are defaults).
func (l *Layer) LineWidth() int { return l.lineWidth }

// Transparency returns the layer transparency (-3 default, 0-100).
func (l *Layer) Transparency() int { return l.transparency }

// SetTransparency sets the layer transparency.
func (l *Layer) SetTransparency(v int) { l.transparency = v }

// TrueColor returns the RGB true color (code 420, -1 when unset).
func (l *Layer) TrueColor() int { return l.trueColor }

// SetTrueColor sets the RGB true color; -1 clears it.
func (l *Layer) SetTrueColor(v int) { l.trueColor = v }

// Plottable reports whether the layer is plottable (code 290).
func (l *Layer) Plottable() bool { return l.plottable }

// SetPlottable sets the plot flag.
func (l *Layer) SetPlottable(v bool) { l.plottable = v }

// SetPlotStyle sets plot style by a handle.
func (l *Layer) SetPlotStyle(ps handle.Handler) {
	l.PlotStyle = ps
}

// SetLineType sets the layer line type.
func (l *Layer) SetLineType(lt *LineType) {
	l.LineType = lt
}

// SetFlag sets standard flags.
//
//	1  = Layer is frozen; otherwise layer is thawed.
//	2  = Layer is frozen by default in new viewports.
//	4  = Layer is locked.
//	16 = If set, table entry is externally dependent on an xref.
//	32 = If this bit and bit 16 are both set, the externally dependent xref has been successfully resolved.
//	64 = If set, the table entry was referenced by at least one entity in the drawing the last time the drawing was edited. (This flag is for the benefit of AutoCAD commands. It can be ignored by most programs that read DXF files and need not be set by programs that write DXF files.)
func (l *Layer) SetFlag(val int) {
	l.flags = val
}

// Freeze freezes Layer.
func (l *Layer) Freeze() {
	l.flags |= 1
}

// UnFreeze unfreezes Layer.
func (l *Layer) UnFreeze() {
	l.flags &= ^1
}

// Lock locks Layer.
func (l *Layer) Lock() {
	l.flags |= 4
}

// UnLock unlocks Layer.
func (l *Layer) UnLock() {
	l.flags &= ^4
}
