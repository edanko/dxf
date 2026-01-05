package entity

import (
	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/table"
)

// Underlay constants
const (
	UnderlayTypePDF = 1
	UnderlayTypeDWF = 2
	UnderlayTypeDGN = 3

	UnderlayFlagClipping   = 1 // 1 = Clipping is on
	UnderlayFlagOn         = 2 // 2 = Underlay is on
	UnderlayFlagMonochrome = 4 // 4 = Monochrome
	UnderlayFlagAdjustBg   = 8 // 8 = Adjust for background
)

// UnderlayDefinition represents the definition for underlay objects
type UnderlayDefinition struct {
	*entity

	// Basic properties
	Type     int    // Underlay type (PDF, DWF, DGN)
	Name     string // Underlay name
	FilePath string // Path to underlay file

	// Display properties
	Flags    int // Display flags (clipping, on, monochrome, adjust bg)
	Contrast int // Contrast value (20-100, default 100)
	Fade     int // Fade value (0-80, default 0)
}

// NewUnderlayDefinition creates a new underlay definition
func NewUnderlayDefinition(underlayType int, name, filePath string) *UnderlayDefinition {
	return &UnderlayDefinition{
		entity:   NewEntity(UNDERLAY_DEFINITION),
		Type:     underlayType,
		Name:     name,
		FilePath: filePath,
		Flags:    UnderlayFlagOn, // Default: underlay is on
		Contrast: 100,            // Default: full contrast
		Fade:     0,              // Default: no fade
	}
}

// IsEntity is for Entity interface
func (u *UnderlayDefinition) IsEntity() bool {
	return true
}

// SetType sets the underlay type
func (u *UnderlayDefinition) SetType(underlayType int) {
	u.Type = underlayType
}

// SetName sets the underlay name
func (u *UnderlayDefinition) SetName(name string) {
	u.Name = name
}

// SetFilePath sets the underlay file path
func (u *UnderlayDefinition) SetFilePath(path string) {
	u.FilePath = path
}

// SetFlags sets the display flags
func (u *UnderlayDefinition) SetFlags(flags int) {
	u.Flags = flags
}

// SetContrast sets the contrast value (20-100)
func (u *UnderlayDefinition) SetContrast(contrast int) {
	if contrast < 20 {
		contrast = 20
	} else if contrast > 100 {
		contrast = 100
	}
	u.Contrast = contrast
}

// SetFade sets the fade value (0-80)
func (u *UnderlayDefinition) SetFade(fade int) {
	if fade < 0 {
		fade = 0
	} else if fade > 80 {
		fade = 80
	}
	u.Fade = fade
}

// SetClipping sets whether clipping is enabled
func (u *UnderlayDefinition) SetClipping(clipping bool) {
	if clipping {
		u.Flags |= UnderlayFlagClipping
	} else {
		u.Flags &= ^UnderlayFlagClipping
	}
}

// SetMonochrome sets whether underlay is displayed in monochrome
func (u *UnderlayDefinition) SetMonochrome(monochrome bool) {
	if monochrome {
		u.Flags |= UnderlayFlagMonochrome
	} else {
		u.Flags &= ^UnderlayFlagMonochrome
	}
}

// SetAdjustBackground sets whether to adjust for background
func (u *UnderlayDefinition) SetAdjustBackground(adjust bool) {
	if adjust {
		u.Flags |= UnderlayFlagAdjustBg
	} else {
		u.Flags &= ^UnderlayFlagAdjustBg
	}
}

// SetLayer implements Entity interface
func (u *UnderlayDefinition) SetLayer(layer *table.Layer) {
	// Underlay definitions don't have layers, but implement for interface
}

// Layer implements Entity interface
func (u *UnderlayDefinition) Layer() *table.Layer {
	return nil
}

// SetLtscale implements Entity interface
func (u *UnderlayDefinition) SetLtscale(scale float64) {
	// Underlay definitions don't have linetype scales, but implement for interface
}

// SetColor implements Entity interface
func (u *UnderlayDefinition) SetColor(cl color.ColorNumber) {
	// Underlay definitions don't have direct color, but implement for interface
	u.entity.SetColor(cl)
}

// SetBlockRecord implements Entity interface
func (u *UnderlayDefinition) SetBlockRecord(h handle.Handler) {
	// Underlay definitions don't have block records, but implement for interface
	u.entity.SetBlockRecord(h)
}

// Format writes underlay definition to formatter
func (u *UnderlayDefinition) Format(f format.Formatter) {
	u.entity.Format(f)

	switch u.Type {
	case UnderlayTypePDF:
		f.WriteString(100, "AcDbPdfDefinition")
	case UnderlayTypeDWF:
		f.WriteString(100, "AcDbDwfDefinition")
	case UnderlayTypeDGN:
		f.WriteString(100, "AcDbDgnDefinition")
	default:
		f.WriteString(100, "AcDbUnderlayDefinition")
	}

	// Write basic properties
	if u.Name != "" {
		f.WriteString(1, u.Name)
	}
	if u.FilePath != "" {
		f.WriteString(2, u.FilePath)
	}

	// Write display properties
	f.WriteInt(70, u.Type)
	f.WriteInt(280, u.Flags)
	f.WriteInt(281, u.Contrast)
	f.WriteInt(282, u.Fade)
}

// BBox returns bounding box of underlay definition
func (u *UnderlayDefinition) BBox() ([]float64, []float64) {
	// Underlay definitions don't have spatial extent until referenced
	return []float64{0, 0, 0}, []float64{0, 0, 0}
}

// GetAcadDictName returns the ACAD dictionary name for this underlay type
func (u *UnderlayDefinition) GetAcadDictName() string {
	switch u.Type {
	case UnderlayTypePDF:
		return "ACAD_PDFDEFINITIONS"
	case UnderlayTypeDWF:
		return "ACAD_DWFDEFINITIONS"
	case UnderlayTypeDGN:
		return "ACAD_DGNDEFINITIONS"
	default:
		return "ACAD_UNDERLAYDEFINITIONS"
	}
}

// Helper functions to create specific underlay definitions

// NewPdfDefinition creates a new PDF underlay definition
func NewPdfDefinition(name, filePath string) *UnderlayDefinition {
	return NewUnderlayDefinition(UnderlayTypePDF, name, filePath)
}

// NewDwfDefinition creates a new DWF underlay definition
func NewDwfDefinition(name, filePath string) *UnderlayDefinition {
	return NewUnderlayDefinition(UnderlayTypeDWF, name, filePath)
}

// NewDgnDefinition creates a new DGN underlay definition
func NewDgnDefinition(name, filePath string) *UnderlayDefinition {
	return NewUnderlayDefinition(UnderlayTypeDGN, name, filePath)
}

// Underlay represents a reference to an underlay definition
type Underlay struct {
	*entity

	// Reference to underlay definition
	UnderlayDefHandle string              // Handle to underlay definition object
	UnderlayDef       *UnderlayDefinition // Reference to definition

	// Transformation properties
	InsertPoint []float64 // Insertion point (10,20,30)
	ScaleX      float64   // Scale X factor (41)
	ScaleY      float64   // Scale Y factor (42)
	ScaleZ      float64   // Scale Z factor (43)
	Rotation    float64   // Rotation angle in degrees (50)
	Extrusion   []float64 // Extrusion direction (210,220,230)

	// Display properties
	Flags    int // Display flags (280)
	Contrast int // Contrast value 20-100 (281)
	Fade     int // Fade value 0-80 (282)

	// Clipping boundary
	BoundaryPath [][]float64 // Boundary path vertices (group code 11)
}

// NewUnderlay creates a new underlay entity
func NewUnderlay() *Underlay {
	return &Underlay{
		entity:            NewEntity(UNDERLAY),
		UnderlayDefHandle: "",
		UnderlayDef:       nil,
		InsertPoint:       []float64{0, 0, 0},
		ScaleX:            1.0,
		ScaleY:            1.0,
		ScaleZ:            1.0,
		Rotation:          0.0,
		Extrusion:         []float64{0, 0, 1},
		Flags:             UnderlayFlagOn, // Default: underlay is on
		Contrast:          100,            // Default: full contrast
		Fade:              0,              // Default: no fade
		BoundaryPath:      make([][]float64, 0),
	}
}

// IsEntity is for Entity interface
func (u *Underlay) IsEntity() bool {
	return true
}

// SetUnderlayDefinition sets the underlay definition
func (u *Underlay) SetUnderlayDefinition(def *UnderlayDefinition) {
	u.UnderlayDef = def
	if def != nil {
		u.UnderlayDefHandle = def.entity.Handle()
	}
}

// SetInsertPoint sets the insertion point
func (u *Underlay) SetInsertPoint(x, y, z float64) {
	u.InsertPoint = []float64{x, y, z}
}

// SetScale sets the scale factors
func (u *Underlay) SetScale(x, y, z float64) {
	if x == 0 {
		x = 1.0
	}
	if y == 0 {
		y = 1.0
	}
	if z == 0 {
		z = 1.0
	}
	u.ScaleX = x
	u.ScaleY = y
	u.ScaleZ = z
}

// SetRotation sets the rotation angle in degrees
func (u *Underlay) SetRotation(angle float64) {
	u.Rotation = angle
}

// SetExtrusion sets the extrusion direction
func (u *Underlay) SetExtrusion(x, y, z float64) {
	u.Extrusion = []float64{x, y, z}
}

// SetFlags sets the display flags
func (u *Underlay) SetFlags(flags int) {
	u.Flags = flags
}

// SetContrast sets the contrast value (20-100)
func (u *Underlay) SetContrast(contrast int) {
	if contrast < 20 {
		contrast = 20
	} else if contrast > 100 {
		contrast = 100
	}
	u.Contrast = contrast
}

// SetFade sets the fade value (0-80)
func (u *Underlay) SetFade(fade int) {
	if fade < 0 {
		fade = 0
	} else if fade > 80 {
		fade = 80
	}
	u.Fade = fade
}

// SetBoundaryPath sets the clipping boundary path
func (u *Underlay) SetBoundaryPath(path [][]float64) {
	u.BoundaryPath = make([][]float64, len(path))
	for i, vertex := range path {
		u.BoundaryPath[i] = make([]float64, len(vertex))
		copy(u.BoundaryPath[i], vertex)
	}
}

// SetClipping sets whether clipping is enabled
func (u *Underlay) SetClipping(clipping bool) {
	if clipping {
		u.Flags |= UnderlayFlagClipping
	} else {
		u.Flags &= ^UnderlayFlagClipping
	}
}

// SetMonochrome sets whether underlay is displayed in monochrome
func (u *Underlay) SetMonochrome(monochrome bool) {
	if monochrome {
		u.Flags |= UnderlayFlagMonochrome
	} else {
		u.Flags &= ^UnderlayFlagMonochrome
	}
}

// SetOn sets whether underlay is displayed
func (u *Underlay) SetOn(on bool) {
	if on {
		u.Flags |= UnderlayFlagOn
	} else {
		u.Flags &= ^UnderlayFlagOn
	}
}

// SetAdjustBackground sets whether to adjust for background
func (u *Underlay) SetAdjustBackground(adjust bool) {
	if adjust {
		u.Flags |= UnderlayFlagAdjustBg
	} else {
		u.Flags &= ^UnderlayFlagAdjustBg
	}
}

// IsClipped returns whether clipping is enabled
func (u *Underlay) IsClipped() bool {
	return u.Flags&UnderlayFlagClipping != 0
}

// IsOn returns whether underlay is displayed
func (u *Underlay) IsOn() bool {
	return u.Flags&UnderlayFlagOn != 0
}

// IsMonochrome returns whether underlay is monochrome
func (u *Underlay) IsMonochrome() bool {
	return u.Flags&UnderlayFlagMonochrome != 0
}

// IsAdjustBackground returns whether to adjust for background
func (u *Underlay) IsAdjustBackground() bool {
	return u.Flags&UnderlayFlagAdjustBg != 0
}

// Format writes underlay to formatter
func (u *Underlay) Format(f format.Formatter) {
	u.entity.Format(f)
	f.WriteString(100, "AcDbUnderlayReference")

	// Write reference to definition
	if u.UnderlayDefHandle != "" {
		f.WriteString(340, u.UnderlayDefHandle)
	}

	// Write transformation properties
	if len(u.InsertPoint) >= 3 {
		f.WriteFloat(10, u.InsertPoint[0])
		f.WriteFloat(20, u.InsertPoint[1])
		f.WriteFloat(30, u.InsertPoint[2])
	}

	f.WriteFloat(41, u.ScaleX)
	f.WriteFloat(42, u.ScaleY)
	f.WriteFloat(43, u.ScaleZ)
	f.WriteFloat(50, u.Rotation)

	// Write extrusion
	if len(u.Extrusion) >= 3 {
		f.WriteFloat(210, u.Extrusion[0])
		f.WriteFloat(220, u.Extrusion[1])
		f.WriteFloat(230, u.Extrusion[2])
	}

	// Write display properties
	f.WriteInt(280, u.Flags)
	f.WriteInt(281, u.Contrast)
	f.WriteInt(282, u.Fade)

	// Write boundary path
	for _, vertex := range u.BoundaryPath {
		if len(vertex) >= 2 {
			f.WriteFloat(11, vertex[0])
			f.WriteFloat(21, vertex[1])
		}
	}
}

// BBox returns bounding box of underlay
func (u *Underlay) BBox() ([]float64, []float64) {
	if len(u.BoundaryPath) == 0 {
		// Return insertion point if no boundary
		return u.InsertPoint, u.InsertPoint
	}

	minX, minY := u.BoundaryPath[0][0], u.BoundaryPath[0][1]
	maxX, maxY := u.BoundaryPath[0][0], u.BoundaryPath[0][1]

	for _, vertex := range u.BoundaryPath {
		if len(vertex) >= 2 {
			if vertex[0] < minX {
				minX = vertex[0]
			}
			if vertex[0] > maxX {
				maxX = vertex[0]
			}
			if vertex[1] < minY {
				minY = vertex[1]
			}
			if vertex[1] > maxY {
				maxY = vertex[1]
			}
		}
	}

	z := 0.0
	if len(u.InsertPoint) >= 3 {
		z = u.InsertPoint[2]
	}

	return []float64{minX, minY, z}, []float64{maxX, maxY, z}
}

// Helper functions to create specific underlay entities

// NewPdfUnderlay creates a new PDF underlay
func NewPdfUnderlay(def *UnderlayDefinition) *Underlay {
	underlay := NewUnderlay()
	underlay.SetUnderlayDefinition(def)
	return underlay
}

// NewDwfUnderlay creates a new DWF underlay
func NewDwfUnderlay(def *UnderlayDefinition) *Underlay {
	underlay := NewUnderlay()
	underlay.SetUnderlayDefinition(def)
	return underlay
}

// NewDgnUnderlay creates a new DGN underlay
func NewDgnUnderlay(def *UnderlayDefinition) *Underlay {
	underlay := NewUnderlay()
	underlay.SetUnderlayDefinition(def)
	return underlay
}
