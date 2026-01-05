package entity

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/math"
)

// Layer represents a LAYER table entry
type Layer struct {
	*entity
	Name            string // 2 - layer name
	Flags           int    // 70 - flags
	Color           int    // 62 - ACI color index (negative = layer is off)
	TrueColor       int    // 420 - true color (DXF2004+)
	LineType        string // 6 - linetype name
	Plot            int    // 290 - plot flag (1=plot, 0=don't plot)
	LineWeight      int    // 370 - line weight (1/100 mm)
	PlotStyleHandle string // 390 - handle to plot style
	MaterialHandle  string // 347 - handle to material (DXF2007+)
	Unknown1        string // 348 - unknown (DXF2007+)
}

// NewLayer creates a new Layer entity
func NewLayer() *Layer {
	l := &Layer{
		entity:          NewEntity(LAYER),
		Name:            "0",
		Flags:           0,
		Color:           7, // white
		TrueColor:       0,
		LineType:        "Continuous",
		Plot:            1,
		LineWeight:      -1, // BYLAYER
		PlotStyleHandle: "",
		MaterialHandle:  "",
		Unknown1:        "",
	}
	return l
}

// NewLayerWithColor creates a layer with specified name and color
func NewLayerWithColor(name string, aci int) *Layer {
	l := NewLayer()
	l.Name = name
	l.Color = aci
	return l
}

// IsEntity is for Entity interface.
func (l *Layer) IsEntity() bool {
	return true
}

// Format writes data to formatter.
func (l *Layer) Format(f format.Formatter) {
	l.entity.Format(f)
	f.WriteString(100, "AcDbSymbolTableRecord")
	f.WriteString(100, "AcDbLayerTableRecord")
	f.WriteString(2, l.Name)
	f.WriteInt(70, l.Flags)
	f.WriteInt(62, l.Color)
	if l.TrueColor > 0 {
		f.WriteInt(420, l.TrueColor)
	}
	f.WriteString(6, l.LineType)
	f.WriteInt(290, l.Plot)
	f.WriteInt(370, l.LineWeight)
	if l.PlotStyleHandle != "" {
		f.WriteString(390, l.PlotStyleHandle)
	}
	if l.MaterialHandle != "" {
		f.WriteString(347, l.MaterialHandle)
	}
	if l.Unknown1 != "" {
		f.WriteString(348, l.Unknown1)
	}
}

// BBox returns the bounding box of the entity
func (l *Layer) BBox() ([]float64, []float64) {
	return []float64{0, 0, 0}, []float64{0, 0, 0}
}

// Transform applies a transformation matrix to the Layer
func (l *Layer) Transform(m *math.Matrix44) error {
	return nil
}

// Copy creates a deep copy of the Layer entity
func (l *Layer) Copy() Entity {
	layer := NewLayer()
	layer.entity = l.entity
	layer.Name = l.Name
	layer.Flags = l.Flags
	layer.Color = l.Color
	layer.TrueColor = l.TrueColor
	layer.LineType = l.LineType
	layer.Plot = l.Plot
	layer.LineWeight = l.LineWeight
	layer.PlotStyleHandle = l.PlotStyleHandle
	layer.MaterialHandle = l.MaterialHandle
	layer.Unknown1 = l.Unknown1
	return layer
}

// Validate validates the Layer entity
func (l *Layer) Validate() error {
	// Validate color (ACI must be between -256 and 256, excluding 0)
	if l.Color == 0 || l.Color < -256 || l.Color > 256 {
		l.Color = 7 // Default to white
	}
	// Validate line weight
	if l.LineWeight < -3 || l.LineWeight > 211 {
		l.LineWeight = -1 // BYLAYER default
	}
	return nil
}

// IsOn returns true if the layer is on (positive color)
func (l *Layer) IsOn() bool {
	return l.Color > 0
}

// SetOn sets the layer on/off status
func (l *Layer) SetOn(on bool) {
	if on {
		l.Color = abs(l.Color)
	} else {
		l.Color = -abs(l.Color)
	}
}

// IsFrozen returns true if the layer is frozen
func (l *Layer) IsFrozen() bool {
	return (l.Flags & 1) != 0
}

// SetFrozen sets the frozen flag
func (l *Layer) SetFrozen(frozen bool) {
	if frozen {
		l.Flags |= 1
	} else {
		l.Flags &^= 1
	}
}

// IsLocked returns true if the layer is locked
func (l *Layer) IsLocked() bool {
	return (l.Flags & 2) != 0
}

// SetLocked sets the locked flag
func (l *Layer) SetLocked(locked bool) {
	if locked {
		l.Flags |= 2
	} else {
		l.Flags &^= 2
	}
}

// PlotEnabled returns true if the layer should be plotted
func (l *Layer) PlotEnabled() bool {
	return l.Plot != 0
}

// SetPlot sets the plot flag
func (l *Layer) SetPlot(plot bool) {
	if plot {
		l.Plot = 1
	} else {
		l.Plot = 0
	}
}

// abs returns absolute value of int
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
