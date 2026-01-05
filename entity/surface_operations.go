package entity

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
	dxfmath "github.com/edanko/dxf/math"
)

// ExtrudedSurface represents an extruded surface entity
type ExtrudedSurface struct {
	*entity

	// Basic properties
	Version       int            // ACIS version
	Flags         int            // Surface flags
	UID           string         // Unique identifier
	HistoryHandle handle.Handler // History record handle

	// Extrusion properties
	SweepVector          dxfmath.Vec3 // Extrusion direction
	TransformationMatrix [16]float64  // 4x4 matrix
	DraftAngle           float64      // Draft angle
	DraftStartDistance   float64      // Draft start distance
	DraftEndDistance     float64      // Draft end distance
	TwistAngle           float64      // Twist angle
	ScaleFactor          float64      // Scale factor
	AlignAngle           float64      // Alignment angle
	Solid                bool         // Solid vs surface
	SweepAlignmentFlags  int          // Sweep alignment flags
	AlignStart           bool         // Align to start
	Bank                 bool         // Bank during sweep
	BasePointSet         bool         // Base point is set

	// Geometry
	ProfileCurve *Curve       // Profile being extruded
	PathCurve    *Curve       // Path for extrusion
	BasePoint    dxfmath.Vec3 // Base point for extrusion
}

// NewExtrudedSurface creates a new extruded surface
func NewExtrudedSurface() *ExtrudedSurface {
	return &ExtrudedSurface{
		entity:        NewEntity(SURFACE),
		Version:       ACISVersion22300,
		Flags:         0,
		UID:           "",
		HistoryHandle: nil,
		SweepVector:   dxfmath.NewVec3(0, 0, 1),
		TransformationMatrix: [16]float64{
			1, 0, 0, 0,
			0, 1, 0, 0,
			0, 0, 1, 0,
			0, 0, 0, 1,
		},
		DraftAngle:          0.0,
		DraftStartDistance:  0.0,
		DraftEndDistance:    0.0,
		TwistAngle:          0.0,
		ScaleFactor:         1.0,
		AlignAngle:          0.0,
		Solid:               false,
		SweepAlignmentFlags: 0,
		AlignStart:          false,
		Bank:                false,
		BasePointSet:        false,
		ProfileCurve:        nil,
		PathCurve:           nil,
		BasePoint:           dxfmath.NewVec3(0, 0, 0),
	}
}

// SetSweepVector sets the extrusion direction
func (es *ExtrudedSurface) SetSweepVector(vector dxfmath.Vec3) {
	es.SweepVector = vector
}

// SetDraftAngle sets the draft angle in radians
func (es *ExtrudedSurface) SetDraftAngle(angle float64) {
	es.DraftAngle = angle
}

// SetTwistAngle sets the twist angle in radians
func (es *ExtrudedSurface) SetTwistAngle(angle float64) {
	es.TwistAngle = angle
}

// SetScaleFactor sets the scale factor
func (es *ExtrudedSurface) SetScaleFactor(factor float64) {
	es.ScaleFactor = factor
}

// SetSolid sets whether this is a solid
func (es *ExtrudedSurface) SetSolid(solid bool) {
	es.Solid = solid
}

// SetProfileCurve sets the profile curve
func (es *ExtrudedSurface) SetProfileCurve(curve *Curve) {
	es.ProfileCurve = curve
}

// SetBasePoint sets the base point for extrusion
func (es *ExtrudedSurface) SetBasePoint(point dxfmath.Vec3) {
	es.BasePoint = point
	es.BasePointSet = true
}

// BBox returns the bounding box of the extruded surface
func (es *ExtrudedSurface) BBox() ([]float64, []float64) {
	// Simple bounding box calculation
	min := []float64{-1e6, -1e6, -1e6}
	max := []float64{1e6, 1e6, 1e6}
	return min, max
}

// SetUID sets the unique identifier
func (es *ExtrudedSurface) SetUID(uid string) {
	es.UID = uid
}

// IsEntity is for Entity interface
func (es *ExtrudedSurface) IsEntity() bool {
	return true
}

// Format writes data to formatter
func (es *ExtrudedSurface) Format(f format.Formatter) {
	es.entity.Format(f)
	f.WriteString(100, "AcDbExtrudedSurface")

	// Write basic properties
	if es.Version != 0 {
		f.WriteInt(70, es.Version)
	}
	if es.UID != "" {
		f.WriteString(1, es.UID)
	}
	if es.HistoryHandle != nil && es.HistoryHandle.Handle() != "0" {
		f.WriteString(350, es.HistoryHandle.Handle())
	}
	if es.Flags != 0 {
		f.WriteInt(71, es.Flags)
	}

	// Write sweep vector
	f.WriteFloat(10, es.SweepVector.X())
	f.WriteFloat(20, es.SweepVector.Y())
	f.WriteFloat(30, es.SweepVector.Z())

	// Write transformation matrix
	for i := 0; i < 16; i++ {
		f.WriteFloat(40+i, es.TransformationMatrix[i])
	}

	// Write extrusion parameters
	f.WriteFloat(42, es.DraftAngle)
	f.WriteFloat(43, es.DraftStartDistance)
	f.WriteFloat(44, es.DraftEndDistance)
	f.WriteFloat(45, es.TwistAngle)
	f.WriteFloat(46, es.ScaleFactor)
	f.WriteFloat(47, es.AlignAngle)

	// Write flags
	var flags int
	if es.Solid {
		flags |= 1
	}
	if es.AlignStart {
		flags |= 2
	}
	if es.Bank {
		flags |= 4
	}
	if es.BasePointSet {
		flags |= 8
	}
	flags |= es.SweepAlignmentFlags << 4
	f.WriteInt(72, flags)

	// Write base point if set
	if es.BasePointSet {
		f.WriteFloat(11, es.BasePoint.X())
		f.WriteFloat(21, es.BasePoint.Y())
		f.WriteFloat(31, es.BasePoint.Z())
	}
}

// String outputs data using default formatter
func (es *ExtrudedSurface) String() string {
	f := format.NewASCII()
	return es.FormatString(f)
}

// FormatString outputs data using given formatter
func (es *ExtrudedSurface) FormatString(f format.Formatter) string {
	es.Format(f)
	return f.Output()
}

// LoftedSurface represents a lofted surface entity
type LoftedSurface struct {
	*entity

	// Basic properties
	Version       int            // ACIS version
	Flags         int            // Surface flags
	UID           string         // Unique identifier
	HistoryHandle handle.Handler // History record handle

	// Lofting properties
	TransformationMatrix [16]float64 // 4x4 matrix
	DraftAngle           float64     // Draft angle
	DraftStartDistance   float64     // Draft start distance
	DraftEndDistance     float64     // Draft end distance
	Solid                bool        // Solid vs surface
	SweepAlignmentFlags  int         // Sweep alignment flags
	AlignStart           bool        // Align to start
	Bank                 bool        // Bank during sweep
	BasePointSet         bool        // Base point is set
	EndPointSet          bool        // End point is set

	// Geometry
	GuideCurves     []*Curve     // Guide curves for lofting
	ProfileCurves   []*Curve     // Profile curves (at least 2 required)
	BasePoint       dxfmath.Vec3 // Base point
	EndPoint        dxfmath.Vec3 // End point
	StartVector     dxfmath.Vec3 // Start tangent vector
	EndVector       dxfmath.Vec3 // End tangent vector
	RaftType        int          // Raft type (0=planar, 1=cylindrical)
	CurveAnchorType int          // Curve anchor type
}

// NewLoftedSurface creates a new lofted surface
func NewLoftedSurface() *LoftedSurface {
	return &LoftedSurface{
		entity:        NewEntity(SURFACE),
		Version:       ACISVersion22300,
		Flags:         0,
		UID:           "",
		HistoryHandle: nil,
		TransformationMatrix: [16]float64{
			1, 0, 0, 0,
			0, 1, 0, 0,
			0, 0, 1, 0,
			0, 0, 0, 1,
		},
		DraftAngle:          0.0,
		DraftStartDistance:  0.0,
		DraftEndDistance:    0.0,
		Solid:               false,
		SweepAlignmentFlags: 0,
		AlignStart:          false,
		Bank:                false,
		BasePointSet:        false,
		EndPointSet:         false,
		GuideCurves:         make([]*Curve, 0),
		ProfileCurves:       make([]*Curve, 0),
		BasePoint:           dxfmath.NewVec3(0, 0, 0),
		EndPoint:            dxfmath.NewVec3(0, 0, 0),
		StartVector:         dxfmath.NewVec3(0, 1, 0),
		EndVector:           dxfmath.NewVec3(0, 1, 0),
		RaftType:            0,
		CurveAnchorType:     0,
	}
}

// BBox returns the bounding box of the lofted surface
func (ls *LoftedSurface) BBox() ([]float64, []float64) {
	min := []float64{-1e6, -1e6, -1e6}
	max := []float64{1e6, 1e6, 1e6}
	return min, max
}

// SetUID sets the unique identifier
func (ls *LoftedSurface) SetUID(uid string) {
	ls.UID = uid
}

// IsEntity is for Entity interface
func (ls *LoftedSurface) IsEntity() bool {
	return true
}

// SetSolid sets whether this is a solid
func (ls *LoftedSurface) SetSolid(solid bool) {
	ls.Solid = solid
}

// SetBasePoint sets the base point for lofting
func (ls *LoftedSurface) SetBasePoint(point dxfmath.Vec3) {
	ls.BasePoint = point
	ls.BasePointSet = true
}

// SetEndPoint sets the end point for lofting
func (ls *LoftedSurface) SetEndPoint(point dxfmath.Vec3) {
	ls.EndPoint = point
	ls.EndPointSet = true
}

// AddGuideCurve adds a guide curve
func (ls *LoftedSurface) AddGuideCurve(curve *Curve) {
	ls.GuideCurves = append(ls.GuideCurves, curve)
}

// AddProfileCurve adds a profile curve
func (ls *LoftedSurface) AddProfileCurve(curve *Curve) {
	ls.ProfileCurves = append(ls.ProfileCurves, curve)
}

// Format writes data to formatter
func (ls *LoftedSurface) Format(f format.Formatter) {
	ls.entity.Format(f)
	f.WriteString(100, "AcDbLoftedSurface")

	// Write basic properties
	if ls.Version != 0 {
		f.WriteInt(70, ls.Version)
	}
	if ls.UID != "" {
		f.WriteString(1, ls.UID)
	}
	if ls.HistoryHandle != nil && ls.HistoryHandle.Handle() != "0" {
		f.WriteString(350, ls.HistoryHandle.Handle())
	}
	if ls.Flags != 0 {
		f.WriteInt(71, ls.Flags)
	}

	// Write transformation matrix
	for i := 0; i < 16; i++ {
		f.WriteFloat(40+i, ls.TransformationMatrix[i])
	}

	// Write lofting parameters
	f.WriteFloat(42, ls.DraftAngle)
	f.WriteFloat(43, ls.DraftStartDistance)
	f.WriteFloat(44, ls.DraftEndDistance)

	// Write flags
	var flags int
	if ls.Solid {
		flags |= 1
	}
	if ls.AlignStart {
		flags |= 2
	}
	if ls.Bank {
		flags |= 4
	}
	if ls.BasePointSet {
		flags |= 8
	}
	if ls.EndPointSet {
		flags |= 16
	}
	flags |= ls.SweepAlignmentFlags << 5
	f.WriteInt(72, flags)

	// Write raft type
	f.WriteInt(73, ls.RaftType)

	// Write curve anchor type
	f.WriteInt(74, ls.CurveAnchorType)

	// Write number of guide curves
	f.WriteInt(95, len(ls.GuideCurves))

	// Write base point if set
	if ls.BasePointSet {
		f.WriteFloat(11, ls.BasePoint.X())
		f.WriteFloat(21, ls.BasePoint.Y())
		f.WriteFloat(31, ls.BasePoint.Z())
	}

	// Write end point if set
	if ls.EndPointSet {
		f.WriteFloat(12, ls.EndPoint.X())
		f.WriteFloat(22, ls.EndPoint.Y())
		f.WriteFloat(32, ls.EndPoint.Z())
	}

	// Write start vector
	f.WriteFloat(13, ls.StartVector.X())
	f.WriteFloat(23, ls.StartVector.Y())
	f.WriteFloat(33, ls.StartVector.Z())

	// Write end vector
	f.WriteFloat(14, ls.EndVector.X())
	f.WriteFloat(24, ls.EndVector.Y())
	f.WriteFloat(34, ls.EndVector.Z())
}

// String outputs data using default formatter
func (ls *LoftedSurface) String() string {
	f := format.NewASCII()
	return ls.FormatString(f)
}

// FormatString outputs data using given formatter
func (ls *LoftedSurface) FormatString(f format.Formatter) string {
	ls.Format(f)
	return f.Output()
}

// RevolvedSurface represents a revolved surface entity
type RevolvedSurface struct {
	*entity

	// Basic properties
	Version       int            // ACIS version
	Flags         int            // Surface flags
	UID           string         // Unique identifier
	HistoryHandle handle.Handler // History record handle

	// Revolve properties
	TransformationMatrix [16]float64  // 4x4 matrix
	AxisVector           dxfmath.Vec3 // Axis of revolution
	RevolutionAngle      float64      // Revolution angle in radians
	StartAngle           float64      // Start angle in radians
	DraftAngle           float64      // Draft angle
	DraftStartDistance   float64      // Draft start distance
	DraftEndDistance     float64      // Draft end distance
	Solid                bool         // Solid vs surface
	SweepAlignmentFlags  int          // Sweep alignment flags
	AlignStart           bool         // Align to start
	Bank                 bool         // Bank during sweep

	// Geometry
	ProfileCurve *Curve       // Profile curve being revolved
	AxisPoint    dxfmath.Vec3 // Point on axis
}

// NewRevolvedSurface creates a new revolved surface
func NewRevolvedSurface() *RevolvedSurface {
	return &RevolvedSurface{
		entity:        NewEntity(SURFACE),
		Version:       ACISVersion22300,
		Flags:         0,
		UID:           "",
		HistoryHandle: nil,
		TransformationMatrix: [16]float64{
			1, 0, 0, 0,
			0, 1, 0, 0,
			0, 0, 1, 0,
			0, 0, 0, 1,
		},
		AxisVector:          dxfmath.NewVec3(0, 0, 1),
		RevolutionAngle:     6.28318530718, // 2*pi radians
		StartAngle:          0,
		DraftAngle:          0,
		DraftStartDistance:  0,
		DraftEndDistance:    0,
		Solid:               false,
		SweepAlignmentFlags: 0,
		AlignStart:          false,
		Bank:                false,
		ProfileCurve:        nil,
		AxisPoint:           dxfmath.NewVec3(0, 0, 0),
	}
}

// BBox returns the bounding box of the revolved surface
func (rs *RevolvedSurface) BBox() ([]float64, []float64) {
	min := []float64{-1e6, -1e6, -1e6}
	max := []float64{1e6, 1e6, 1e6}
	return min, max
}

// SetUID sets the unique identifier
func (rs *RevolvedSurface) SetUID(uid string) {
	rs.UID = uid
}

// IsEntity is for Entity interface
func (rs *RevolvedSurface) IsEntity() bool {
	return true
}

// SetSolid sets whether this is a solid
func (rs *RevolvedSurface) SetSolid(solid bool) {
	rs.Solid = solid
}

// SetRevolutionAngle sets the revolution angle in radians
func (rs *RevolvedSurface) SetRevolutionAngle(angle float64) {
	rs.RevolutionAngle = angle
}

// SetAxisVector sets the axis of revolution
func (rs *RevolvedSurface) SetAxisVector(vector dxfmath.Vec3) {
	rs.AxisVector = vector
}

// SetAxisPoint sets the point on the axis
func (rs *RevolvedSurface) SetAxisPoint(point dxfmath.Vec3) {
	rs.AxisPoint = point
}

// Format writes data to formatter
func (rs *RevolvedSurface) Format(f format.Formatter) {
	rs.entity.Format(f)
	f.WriteString(100, "AcDbRevolvedSurface")

	// Write basic properties
	if rs.Version != 0 {
		f.WriteInt(70, rs.Version)
	}
	if rs.UID != "" {
		f.WriteString(1, rs.UID)
	}
	if rs.HistoryHandle != nil && rs.HistoryHandle.Handle() != "0" {
		f.WriteString(350, rs.HistoryHandle.Handle())
	}
	if rs.Flags != 0 {
		f.WriteInt(71, rs.Flags)
	}

	// Write axis point
	f.WriteFloat(10, rs.AxisPoint.X())
	f.WriteFloat(20, rs.AxisPoint.Y())
	f.WriteFloat(30, rs.AxisPoint.Z())

	// Write axis vector
	f.WriteFloat(11, rs.AxisVector.X())
	f.WriteFloat(21, rs.AxisVector.Y())
	f.WriteFloat(31, rs.AxisVector.Z())

	// Write transformation matrix
	for i := 0; i < 16; i++ {
		f.WriteFloat(40+i, rs.TransformationMatrix[i])
	}

	// Write revolve parameters
	f.WriteFloat(42, rs.RevolutionAngle)
	f.WriteFloat(43, rs.StartAngle)
	f.WriteFloat(44, rs.DraftAngle)
	f.WriteFloat(45, rs.DraftStartDistance)
	f.WriteFloat(46, rs.DraftEndDistance)

	// Write flags
	var flags int
	if rs.Solid {
		flags |= 1
	}
	if rs.AlignStart {
		flags |= 2
	}
	if rs.Bank {
		flags |= 4
	}
	flags |= rs.SweepAlignmentFlags << 3
	f.WriteInt(72, flags)
}

// String outputs data using default formatter
func (rs *RevolvedSurface) String() string {
	f := format.NewASCII()
	return rs.FormatString(f)
}

// FormatString outputs data using given formatter
func (rs *RevolvedSurface) FormatString(f format.Formatter) string {
	rs.Format(f)
	return f.Output()
}

// SweptSurface represents a swept surface entity
type SweptSurface struct {
	*entity

	// Basic properties
	Version       int            // ACIS version
	Flags         int            // Surface flags
	UID           string         // Unique identifier
	HistoryHandle handle.Handler // History record handle

	// Sweep properties
	TransformationMatrix [16]float64 // 4x4 matrix
	DraftAngle           float64     // Draft angle
	DraftStartDistance   float64     // Draft start distance
	DraftEndDistance     float64     // Draft end distance
	TwistAngle           float64     // Twist angle
	ScaleFactor          float64     // Scale factor
	AlignAngle           float64     // Alignment angle
	Solid                bool        // Solid vs surface
	SweepAlignmentFlags  int         // Sweep alignment flags
	AlignStart           bool        // Align to start
	Bank                 bool        // Bank during sweep
	StartDraft           bool        // Start draft enabled
	EndDraft             bool        // End draft enabled
	GlobalSweepAngle     float64     // Global sweep angle
	SectionOffsetS       float64     // Section offset S
	SectionOffsetT       float64     // Section offset T
	GuideDistance        float64     // Guide distance
	AdditionalDraftAngle float64     // Additional draft angle
	AdditionalDraftType  int         // Additional draft type

	// Geometry
	ProfileCurve *Curve // Profile curve being swept
	PathCurve    *Curve // Path curve
}

// NewSweptSurface creates a new swept surface
func NewSweptSurface() *SweptSurface {
	return &SweptSurface{
		entity:        NewEntity(SURFACE),
		Version:       ACISVersion22300,
		Flags:         0,
		UID:           "",
		HistoryHandle: nil,
		TransformationMatrix: [16]float64{
			1, 0, 0, 0,
			0, 1, 0, 0,
			0, 0, 1, 0,
			0, 0, 0, 1,
		},
		DraftAngle:           0,
		DraftStartDistance:   0,
		DraftEndDistance:     0,
		TwistAngle:           0,
		ScaleFactor:          1,
		AlignAngle:           0,
		Solid:                false,
		SweepAlignmentFlags:  0,
		AlignStart:           false,
		Bank:                 false,
		StartDraft:           false,
		EndDraft:             false,
		GlobalSweepAngle:     0,
		SectionOffsetS:       0,
		SectionOffsetT:       0,
		GuideDistance:        0,
		AdditionalDraftAngle: 0,
		AdditionalDraftType:  0,
		ProfileCurve:         nil,
		PathCurve:            nil,
	}
}

// BBox returns the bounding box of the swept surface
func (ss *SweptSurface) BBox() ([]float64, []float64) {
	min := []float64{-1e6, -1e6, -1e6}
	max := []float64{1e6, 1e6, 1e6}
	return min, max
}

// SetUID sets the unique identifier
func (ss *SweptSurface) SetUID(uid string) {
	ss.UID = uid
}

// IsEntity is for Entity interface
func (ss *SweptSurface) IsEntity() bool {
	return true
}

// SetSolid sets whether this is a solid
func (ss *SweptSurface) SetSolid(solid bool) {
	ss.Solid = solid
}

// SetProfileCurve sets the profile curve
func (ss *SweptSurface) SetProfileCurve(curve *Curve) {
	ss.ProfileCurve = curve
}

// SetPathCurve sets the path curve
func (ss *SweptSurface) SetPathCurve(curve *Curve) {
	ss.PathCurve = curve
}

// Format writes data to formatter
func (ss *SweptSurface) Format(f format.Formatter) {
	ss.entity.Format(f)
	f.WriteString(100, "AcDbSweptSurface")

	// Write basic properties
	if ss.Version != 0 {
		f.WriteInt(70, ss.Version)
	}
	if ss.UID != "" {
		f.WriteString(1, ss.UID)
	}
	if ss.HistoryHandle != nil && ss.HistoryHandle.Handle() != "0" {
		f.WriteString(350, ss.HistoryHandle.Handle())
	}
	if ss.Flags != 0 {
		f.WriteInt(71, ss.Flags)
	}

	// Write transformation matrix
	for i := 0; i < 16; i++ {
		f.WriteFloat(40+i, ss.TransformationMatrix[i])
	}

	// Write sweep parameters
	f.WriteFloat(42, ss.DraftAngle)
	f.WriteFloat(43, ss.DraftStartDistance)
	f.WriteFloat(44, ss.DraftEndDistance)
	f.WriteFloat(45, ss.TwistAngle)
	f.WriteFloat(46, ss.ScaleFactor)
	f.WriteFloat(47, ss.AlignAngle)

	// Write flags
	var flags int
	if ss.Solid {
		flags |= 1
	}
	if ss.AlignStart {
		flags |= 2
	}
	if ss.Bank {
		flags |= 4
	}
	if ss.StartDraft {
		flags |= 16
	}
	if ss.EndDraft {
		flags |= 32
	}
	flags |= ss.SweepAlignmentFlags << 6
	f.WriteInt(72, flags)

	// Write additional parameters
	f.WriteFloat(73, ss.GlobalSweepAngle)
	f.WriteFloat(74, ss.SectionOffsetS)
	f.WriteFloat(75, ss.SectionOffsetT)
	f.WriteFloat(76, ss.GuideDistance)
	f.WriteFloat(77, ss.AdditionalDraftAngle)
	f.WriteInt(78, ss.AdditionalDraftType)
}

// String outputs data using default formatter
func (ss *SweptSurface) String() string {
	f := format.NewASCII()
	return ss.FormatString(f)
}

// FormatString outputs data using given formatter
func (ss *SweptSurface) FormatString(f format.Formatter) string {
	ss.Format(f)
	return f.Output()
}
