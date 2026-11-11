package table

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/lldxf"
)

// Viewport represents VPORT SymbolTable.
type Viewport struct {
	EntryBase

	name           string // 2
	LowerLeft      []float64
	UpperRight     []float64
	ViewCenter     []float64
	SnapBase       []float64
	SnapSpacing    []float64
	GridSpacing    []float64
	ViewDirection  []float64
	ViewTarget     []float64
	Height         float64
	AspectRatio    float64
	LensLength     float64
	FrontClip      float64
	BackClip       float64
	SnapAngle      float64
	TwistAngle     float64
	ViewportType   int // 71
	UCSIconSetting int // 72
	SnapOn         int // 73
	GridOn         int // 74
	ZoomFlags      int // 75
}

// NewViewport creates a new Viewport.
func NewViewport(name string) *Viewport {
	v := &Viewport{
		name:          name,
		LowerLeft:     []float64{0.0, 0.0},
		UpperRight:    []float64{0.0, 0.0},
		ViewCenter:    []float64{0.0, 0.0},
		SnapBase:      []float64{0.0, 0.0},
		SnapSpacing:   []float64{0.0, 0.0},
		GridSpacing:   []float64{0.0, 0.0},
		ViewDirection: []float64{0.0, 0.0, 0.0},
		ViewTarget:    []float64{0.0, 0.0, 0.0},
		Height:        400.0,
		AspectRatio:   1.0,
		LensLength:    50.0,
		FrontClip:     0.0,
		BackClip:      0.0,
		SnapAngle:     0.0,
		TwistAngle:    0.0,
	}
	return v
}

// IsSymbolTable is for SymbolTable interface.
func (v *Viewport) IsSymbolTable() bool {
	return true
}

// Format writes data to formatter.
func (v *Viewport) Format(f format.Formatter) {
	v.FormatCommon(f, "VPORT", "AcDbViewportTableRecord")
	f.WriteString(100, "AcDbViewportTableRecord")
	f.WriteString(2, v.name)
	f.WriteInt(70, v.flags)
	for i := range 2 {
		f.WriteFloat((i+1)*10, v.LowerLeft[i])
	}
	for i := range 2 {
		f.WriteFloat((i+1)*10+1, v.UpperRight[i])
	}
	for i := range 2 {
		f.WriteFloat((i+1)*10+2, v.ViewCenter[i])
	}
	for i := range 2 {
		f.WriteFloat((i+1)*10+3, v.SnapBase[i])
	}
	for i := range 2 {
		f.WriteFloat((i+1)*10+4, v.SnapSpacing[i])
	}
	for i := range 2 {
		f.WriteFloat((i+1)*10+5, v.GridSpacing[i])
	}
	for i := range 3 {
		f.WriteFloat((i+1)*10+6, v.ViewDirection[i])
	}
	for i := range 3 {
		f.WriteFloat((i+1)*10+7, v.ViewTarget[i])
	}
	f.WriteFloat(40, v.Height)
	f.WriteFloat(41, v.AspectRatio)
	f.WriteFloat(42, v.LensLength)
	f.WriteFloat(43, v.FrontClip)
	f.WriteFloat(44, v.BackClip)
	f.WriteFloat(50, v.SnapAngle)
	f.WriteFloat(51, v.TwistAngle)
	f.WriteInt(71, v.ViewportType)
	f.WriteInt(72, v.UCSIconSetting)
	f.WriteInt(73, v.SnapOn)
	f.WriteInt(74, v.GridOn)
	f.WriteInt(75, v.ZoomFlags)
	v.FormatCommonTail(f)
}

// RetainTags records VPORT codes the typed model does not emit.
func (v *Viewport) RetainTags(tags lldxf.Tags) {
	emitted := map[int]bool{
		0: true, 2: true, 5: true, 70: true, 100: true, 330: true,
		10: true, 20: true, 11: true, 21: true, 12: true, 22: true,
		13: true, 23: true, 14: true, 24: true, 15: true, 25: true,
		16: true, 26: true, 36: true, 17: true, 27: true, 37: true,
		40: true, 41: true, 42: true, 43: true, 44: true, 50: true, 51: true,
		71: true, 72: true, 73: true, 74: true, 75: true,
	}
	v.RetainUnhandled(tags, emitted)
}

// Name returns a name of VPORT (code 2).
func (v *Viewport) Name() string {
	return v.name
}

// SetName sets a name to VPORT (code 2).
func (v *Viewport) SetName(name string) {
	v.name = name
}
