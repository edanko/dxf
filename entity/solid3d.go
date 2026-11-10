package entity

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
	dxfmath "github.com/edanko/dxf/math"
)

// Solid3d represents a 3D solid entity (DXF 3DSOLID)
type Solid3d struct {
	*entity

	// Basic properties
	Version       int            // ACIS version
	Flags         int            // Solid flags
	UID           string         // Unique identifier
	HistoryHandle handle.Handler // History record handle (DXF R2007+)

	// Geometry data
	SATData    []byte   // SAT/SAB binary data (retained via rawTags)
	SATVersion int      // ACIS version for SAT format
	satLines   []string // text SAT payload lines
	satCodes   []int    // wire code per SAT line
	acisInfo   ACISMetadata

	// Transformation
	Transform        *ACASTransform
	InverseTransform *ACASTransform

	// Bounding box and properties
	boundingBox *ACISBoundingBox
	center      dxfmath.Vec3
	volume      float64
	surfaceArea float64

	// Sub-entities
	SubSolids []*Solid3d
}

// NewSolid3d creates a new 3D solid entity
func NewSolid3d() *Solid3d {
	return &Solid3d{
		entity:        NewEntity(SOLID3D),
		Version:       ACISVersion22300,
		Flags:         0,
		UID:           "",
		HistoryHandle: nil,
		SATData:       make([]byte, 0),
		SATVersion:    ACISVersion22300,
		Transform:     NewACASTransform(),
		boundingBox:   nil,
		center:        dxfmath.NewVec3(0, 0, 0),
		volume:        0.0,
		surfaceArea:   0.0,
		SubSolids:     make([]*Solid3d, 0),
	}
}

// SetVersion sets the ACIS version
func (s *Solid3d) SetVersion(version int) {
	s.Version = version
	s.SATVersion = version
}

// SetUID sets the unique identifier
func (s *Solid3d) SetUID(uid string) {
	s.UID = uid
}

// SetHistoryHandle sets the history record handle
func (s *Solid3d) SetHistoryHandle(h handle.Handler) {
	s.HistoryHandle = h
}

// SetSATData sets the SAT/SAB binary data
func (s *Solid3d) SetSATData(data []byte) {
	s.SATData = append([]byte(nil), data...)
}

// SetTransform sets the transformation matrix
func (s *Solid3d) SetTransform(transform *ACASTransform) {
	s.Transform = transform
	// Calculate inverse transform (simplified)
	s.InverseTransform = NewACASTransform() // Placeholder
}

// GetBoundingBox calculates and returns the bounding box
func (s *Solid3d) GetBoundingBox() *ACISBoundingBox {
	if s.boundingBox != nil {
		return s.boundingBox
	}

	// Calculate from geometry if not cached
	// This is a simplified calculation
	min := dxfmath.NewVec3(-1e6, -1e6, -1e6)
	max := dxfmath.NewVec3(1e6, 1e6, 1e6)
	s.boundingBox = NewACISBoundingBox(min, max)
	return s.boundingBox
}

// GetCenter returns the geometric center
func (s *Solid3d) GetCenter() dxfmath.Vec3 {
	if !(s.center.X() == 0 && s.center.Y() == 0 && s.center.Z() == 0) {
		return s.center
	}

	// Calculate from bounding box if not set
	bb := s.GetBoundingBox()
	s.center = dxfmath.NewVec3(
		(bb.Min.X()+bb.Max.X())/2,
		(bb.Min.Y()+bb.Max.Y())/2,
		(bb.Min.Z()+bb.Max.Z())/2,
	)
	return s.center
}

// GetVolume returns the calculated volume
func (s *Solid3d) GetVolume() float64 {
	return s.volume
}

// SetVolume sets the volume
func (s *Solid3d) SetVolume(volume float64) {
	s.volume = volume
}

// GetSurfaceArea returns the calculated surface area
func (s *Solid3d) GetSurfaceArea() float64 {
	return s.surfaceArea
}

// SetSurfaceArea sets the surface area
func (s *Solid3d) SetSurfaceArea(area float64) {
	s.surfaceArea = area
}

// SetCenter sets the geometric center
func (s *Solid3d) SetCenter(center dxfmath.Vec3) {
	s.center = center
}

// BBox returns bounding box
func (s *Solid3d) BBox() ([]float64, []float64) {
	bb := s.GetBoundingBox()
	return []float64{bb.Min.X(), bb.Min.Y(), bb.Min.Z()}, []float64{bb.Max.X(), bb.Max.Y(), bb.Max.Z()}
}

// AddSubSolid adds a sub-solid to this solid
func (s *Solid3d) AddSubSolid(subSolid *Solid3d) {
	s.SubSolids = append(s.SubSolids, subSolid)
}

// IsEntity is for Entity interface
func (s *Solid3d) IsEntity() bool {
	return true
}

// Format writes data to formatter
func (s *Solid3d) Format(f format.Formatter) {
	s.entity.Format(f)

	// AcDbEntity common section carries history handle (code 350) when set.
	if s.HistoryHandle != nil && s.HistoryHandle.Handle() != "0" {
		f.WriteString(350, s.HistoryHandle.Handle())
	}

	f.WriteString(100, "AcDbModelerGeometry")
	// Modeler geometry version (code 70) and history GUID string (code 1).
	f.WriteInt(70, 0)
	if s.UID != "" {
		f.WriteString(1, s.UID)
	}

	// SAT (text ACIS) payload, preserving legacy code 10 lines and current
	// code 1/3 SAT lines. Real ACIS records place the payload inside the
	// AcDbModelerGeometry subclass, before AcDb3dSolid.
	writeSATPayload(f, s.satLines, s.satCodes)

	f.WriteString(100, "AcDb3dSolid")
	// ACIS version (code 70).
	f.WriteInt(70, s.Version)
	// Solid flags (code 91); only when no binary SAB payload is retained in
	// rawTags, which already carries code 91.
	if len(s.SATData) == 0 {
		f.WriteInt(91, s.Flags)
	}

	// Geometric properties and center (AcDb3dSolid section).
	if s.volume != 0 {
		f.WriteFloat(40, s.volume)
	}
	if s.surfaceArea != 0 {
		f.WriteFloat(41, s.surfaceArea)
	}
	center := s.GetCenter()
	f.WriteFloat(10, center.X())
	f.WriteFloat(20, center.Y())
	f.WriteFloat(30, center.Z())
}

// RawWholeRecord reports whole-record retention: a 3DSOLID with an opaque
// binary SAT/SAB payload (codes 90/310) re-emits its retained tags verbatim,
// and the typed accessors below are unavailable for such records.
func (s *Solid3d) RawWholeRecord() bool { return len(s.SATData) > 0 || len(s.satLines) > 0 }

// SATLines returns the stored text SAT payload lines.
func (s *Solid3d) SATLines() []string { return s.satLines }

// SetSATLines replaces the text SAT payload.
func (s *Solid3d) SetSATLines(lines []string) {
	s.satLines = append([]string(nil), lines...)
	s.satCodes = normalizeSATCodes(s.satLines, nil)
}

// SetSATPayload replaces the text SAT payload and preserves each line's wire
// group code.
func (s *Solid3d) SetSATPayload(lines []string, codes []int) {
	s.satLines = append([]string(nil), lines...)
	s.satCodes = normalizeSATCodes(s.satLines, codes)
}

// ACISInfo returns the shallow ACIS metadata parsed from the payload.
func (s *Solid3d) ACISInfo() ACISMetadata { return s.acisInfo }

// SetACISInfo replaces the ACIS metadata.
func (s *Solid3d) SetACISInfo(m ACISMetadata) { s.acisInfo = m }

// SetACISBinaryData attaches a SAB payload and updates its metadata.
func (s *Solid3d) SetACISBinaryData(data []byte) {
	s.SATData = append([]byte(nil), data...)
	s.acisInfo = ParseSABMetadata(s.SATData)
}

// ACISBinaryData returns a copy of the retained SAB payload.
func (s *Solid3d) ACISBinaryData() []byte { return append([]byte(nil), s.SATData...) }

// HasBinaryACIS reports whether the record carries SAB data.
func (s *Solid3d) HasBinaryACIS() bool { return s.acisInfo.IsBinary() || len(s.SATData) > 0 }

// String outputs data using default formatter
func (s *Solid3d) String() string {
	f := format.NewASCII()
	return s.FormatString(f)
}

// FormatString outputs data using given formatter
func (s *Solid3d) FormatString(f format.Formatter) string {
	s.Format(f)
	return f.Output()
}
