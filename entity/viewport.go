package entity

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/math"
)

// Viewport represents a VPORT (Viewport) entity for layout management
type Viewport struct {
	*entity
	Center                 math.Vec3 // 10, 20, 30 - center in paper space
	Width                  float64   // 40 - width in paper space units
	Height                 float64   // 41 - height in paper space units
	Status                 int       // 68 - viewport status (0=off, >0=on)
	ID                     int       // 69 - viewport ID (1=active)
	ViewCenterPoint        math.Vec2 // 12, 22 - view center in model space
	SnapBasePoint          math.Vec2 // 13, 23
	SnapSpacing            math.Vec2 // 14, 24
	GridSpacing            math.Vec2 // 15, 25
	ViewDirectionVector    math.Vec3 // 16, 26, 36 - view direction
	ViewTargetPoint        math.Vec3 // 17, 27, 37 - view target
	PerspectiveLensLength  float64   // 42
	FrontClipPlaneZValue   float64   // 43
	BackClipPlaneZValue    float64   // 44
	ViewHeight             float64   // 45 - view height in model space
	SnapAngle              float64   // 50
	ViewTwistAngle         float64   // 51
	CircleZoom             int       // 72
	Flags                  int       // 90 - viewport status flags
	ClippingBoundaryHandle string    // 340
	PlotStyleName          string    // 1
	RenderMode             int       // 281
	UCSPerViewport         int       // 71
	UCSIcon                int       // 74
	UCSOrigin              math.Vec3 // 110, 120, 130
	UCSXAxis               math.Vec3 // 111, 121, 131
	UCSYAxis               math.Vec3 // 112, 122, 132
	UCSHandle              string    // 345
	BaseUCSHandle          string    // 346
	UCSOrthoType           int       // 79
	Elevation              float64   // 146
	ShadePlotMode          int       // 170
	PlotStyleSheetHandle   string    // 390
	BackgroundHandle       string    // 331
	ShadePlotHandle        string    // 332
	VisualStyleHandle      string    // 333
}

// NewViewport creates a new Viewport entity
func NewViewport() *Viewport {
	v := &Viewport{
		entity:                 NewEntity(VIEWPORT),
		Center:                 math.Vec3{0, 0, 0},
		Width:                  1.0,
		Height:                 1.0,
		Status:                 1,
		ID:                     2,
		ViewCenterPoint:        math.Vec2{0, 0},
		SnapBasePoint:          math.Vec2{0, 0},
		SnapSpacing:            math.Vec2{10, 10},
		GridSpacing:            math.Vec2{10, 10},
		ViewDirectionVector:    math.Vec3{0, 0, 1},
		ViewTargetPoint:        math.Vec3{0, 0, 0},
		PerspectiveLensLength:  50.0,
		FrontClipPlaneZValue:   0.0,
		BackClipPlaneZValue:    0.0,
		ViewHeight:             1.0,
		SnapAngle:              0.0,
		ViewTwistAngle:         0.0,
		CircleZoom:             100,
		Flags:                  0,
		ClippingBoundaryHandle: "",
		PlotStyleName:          "",
		RenderMode:             0,
		UCSPerViewport:         0,
		UCSIcon:                0,
		UCSOrigin:              math.Vec3{0, 0, 0},
		UCSXAxis:               math.Vec3{1, 0, 0},
		UCSYAxis:               math.Vec3{0, 1, 0},
		UCSHandle:              "",
		BaseUCSHandle:          "",
		UCSOrthoType:           0,
		Elevation:              0.0,
		ShadePlotMode:          0,
		PlotStyleSheetHandle:   "",
		BackgroundHandle:       "",
		ShadePlotHandle:        "",
		VisualStyleHandle:      "",
	}
	return v
}

// IsEntity is for Entity interface.
func (v *Viewport) IsEntity() bool {
	return true
}

// Format writes data to formatter.
func (v *Viewport) Format(f format.Formatter) {
	v.entity.Format(f)
	f.WriteString(100, "AcDbViewport")
	f.WriteFloat(10, v.Center.X())
	f.WriteFloat(20, v.Center.Y())
	f.WriteFloat(30, v.Center.Z())
	f.WriteFloat(40, v.Width)
	f.WriteFloat(41, v.Height)
	f.WriteInt(68, v.Status)
	f.WriteInt(69, v.ID)
	f.WriteFloat(12, v.ViewCenterPoint.X())
	f.WriteFloat(22, v.ViewCenterPoint.Y())
	f.WriteFloat(13, v.SnapBasePoint.X())
	f.WriteFloat(23, v.SnapBasePoint.Y())
	f.WriteFloat(14, v.SnapSpacing.X())
	f.WriteFloat(24, v.SnapSpacing.Y())
	f.WriteFloat(15, v.GridSpacing.X())
	f.WriteFloat(25, v.GridSpacing.Y())
	f.WriteFloat(16, v.ViewDirectionVector.X())
	f.WriteFloat(26, v.ViewDirectionVector.Y())
	f.WriteFloat(36, v.ViewDirectionVector.Z())
	f.WriteFloat(17, v.ViewTargetPoint.X())
	f.WriteFloat(27, v.ViewTargetPoint.Y())
	f.WriteFloat(37, v.ViewTargetPoint.Z())
	f.WriteFloat(42, v.PerspectiveLensLength)
	f.WriteFloat(43, v.FrontClipPlaneZValue)
	f.WriteFloat(44, v.BackClipPlaneZValue)
	f.WriteFloat(45, v.ViewHeight)
	f.WriteFloat(50, v.SnapAngle)
	f.WriteFloat(51, v.ViewTwistAngle)
	f.WriteInt(72, v.CircleZoom)
	f.WriteInt(90, v.Flags)
	if v.ClippingBoundaryHandle != "" {
		f.WriteString(340, v.ClippingBoundaryHandle)
	}
	f.WriteString(1, v.PlotStyleName)
	f.WriteInt(281, v.RenderMode)
	f.WriteInt(71, v.UCSPerViewport)
	f.WriteInt(74, v.UCSIcon)
	f.WriteFloat(110, v.UCSOrigin.X())
	f.WriteFloat(120, v.UCSOrigin.Y())
	f.WriteFloat(130, v.UCSOrigin.Z())
	f.WriteFloat(111, v.UCSXAxis.X())
	f.WriteFloat(121, v.UCSXAxis.Y())
	f.WriteFloat(131, v.UCSXAxis.Z())
	f.WriteFloat(112, v.UCSYAxis.X())
	f.WriteFloat(122, v.UCSYAxis.Y())
	f.WriteFloat(132, v.UCSYAxis.Z())
	if v.UCSHandle != "" {
		f.WriteString(345, v.UCSHandle)
	}
	if v.BaseUCSHandle != "" {
		f.WriteString(346, v.BaseUCSHandle)
	}
	f.WriteInt(79, v.UCSOrthoType)
	f.WriteFloat(146, v.Elevation)
	f.WriteInt(170, v.ShadePlotMode)
	if v.PlotStyleSheetHandle != "" {
		f.WriteString(390, v.PlotStyleSheetHandle)
	}
	if v.BackgroundHandle != "" {
		f.WriteString(331, v.BackgroundHandle)
	}
	if v.ShadePlotHandle != "" {
		f.WriteString(332, v.ShadePlotHandle)
	}
	if v.VisualStyleHandle != "" {
		f.WriteString(333, v.VisualStyleHandle)
	}
}

// BBox returns the bounding box of the entity
func (v *Viewport) BBox() ([]float64, []float64) {
	// Viewport is a layout entity, return bounding box based on center, width, height
	halfW := v.Width / 2
	halfH := v.Height / 2
	mins := []float64{v.Center.X() - halfW, v.Center.Y() - halfH, 0}
	maxs := []float64{v.Center.X() + halfW, v.Center.Y() + halfH, 0}
	return mins, maxs
}

// Transform applies a transformation matrix to the Viewport
func (v *Viewport) Transform(m *math.Matrix44) error {
	// Transform center point
	transformVec3 := func(v math.Vec3) math.Vec3 {
		x := v[0]*m[0] + v[1]*m[1] + v[2]*m[2] + m[3]
		y := v[0]*m[4] + v[1]*m[5] + v[2]*m[6] + m[7]
		z := v[0]*m[8] + v[1]*m[9] + v[2]*m[10] + m[11]
		return math.Vec3{x, y, z}
	}
	v.Center = transformVec3(v.Center)
	v.ViewDirectionVector = transformVec3(v.ViewDirectionVector)
	v.ViewTargetPoint = transformVec3(v.ViewTargetPoint)
	v.UCSOrigin = transformVec3(v.UCSOrigin)
	v.UCSXAxis = transformVec3(v.UCSXAxis)
	v.UCSYAxis = transformVec3(v.UCSYAxis)
	return nil
}

// Copy creates a deep copy of the Viewport entity
func (v *Viewport) Copy() Entity {
	vp := NewViewport()
	vp.entity = v.entity
	vp.Center = v.Center
	vp.Width = v.Width
	vp.Height = v.Height
	vp.Status = v.Status
	vp.ID = v.ID
	vp.ViewCenterPoint = v.ViewCenterPoint
	vp.SnapBasePoint = v.SnapBasePoint
	vp.SnapSpacing = v.SnapSpacing
	vp.GridSpacing = v.GridSpacing
	vp.ViewDirectionVector = v.ViewDirectionVector
	vp.ViewTargetPoint = v.ViewTargetPoint
	vp.PerspectiveLensLength = v.PerspectiveLensLength
	vp.FrontClipPlaneZValue = v.FrontClipPlaneZValue
	vp.BackClipPlaneZValue = v.BackClipPlaneZValue
	vp.ViewHeight = v.ViewHeight
	vp.SnapAngle = v.SnapAngle
	vp.ViewTwistAngle = v.ViewTwistAngle
	vp.CircleZoom = v.CircleZoom
	vp.Flags = v.Flags
	vp.ClippingBoundaryHandle = v.ClippingBoundaryHandle
	vp.PlotStyleName = v.PlotStyleName
	vp.RenderMode = v.RenderMode
	vp.UCSPerViewport = v.UCSPerViewport
	vp.UCSIcon = v.UCSIcon
	vp.UCSOrigin = v.UCSOrigin
	vp.UCSXAxis = v.UCSXAxis
	vp.UCSYAxis = v.UCSYAxis
	vp.UCSHandle = v.UCSHandle
	vp.BaseUCSHandle = v.BaseUCSHandle
	vp.UCSOrthoType = v.UCSOrthoType
	vp.Elevation = v.Elevation
	vp.ShadePlotMode = v.ShadePlotMode
	vp.PlotStyleSheetHandle = v.PlotStyleSheetHandle
	vp.BackgroundHandle = v.BackgroundHandle
	vp.ShadePlotHandle = v.ShadePlotHandle
	vp.VisualStyleHandle = v.VisualStyleHandle
	return vp
}

// Validate validates the Viewport entity
func (v *Viewport) Validate() error {
	if v.Width <= 0 {
		v.Width = 1.0
	}
	if v.Height <= 0 {
		v.Height = 1.0
	}
	if v.ViewHeight <= 0 {
		v.ViewHeight = 1.0
	}
	if v.PerspectiveLensLength < 0 {
		v.PerspectiveLensLength = 50.0
	}
	// Ensure UCS axes are normalized
	v.UCSXAxis = v.UCSXAxis.Normalized()
	v.UCSYAxis = v.UCSYAxis.Normalized()
	return nil
}

// IsOn returns true if the viewport is active
func (v *Viewport) IsOn() bool {
	return v.Status != 0
}

// SetOn sets the viewport status
func (v *Viewport) SetOn(on bool) {
	if on {
		if v.Status == 0 {
			v.Status = v.ID
		}
	} else {
		v.Status = 0
	}
}

// SetRenderMode sets the render mode
func (v *Viewport) SetRenderMode(mode int) {
	if mode < 0 || mode > 6 {
		mode = 0
	}
	v.RenderMode = mode
}
