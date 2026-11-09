package entity

import (
	"fmt"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/lldxf"
	stdmath "math"

	"github.com/edanko/dxf/math"
	dxfmath "github.com/edanko/dxf/math"
)

// Hatch pattern types
const (
	HatchPatternTypeUserDefined = 0
	HatchPatternTypePredefined  = 1
	HatchPatternTypeCustom      = 2
)

// Boundary path types
const (
	BoundaryPathTypeDefault   = 0
	BoundaryPathTypePolyline  = 2
	BoundaryPathTypeDerived   = 4
	BoundaryPathTypeTextbox   = 8
	BoundaryPathTypeOutermost = 16
)

// Hatch styles
const (
	HatchStyleNormal = 0 // Nested hatching (odd parity)
	HatchStyleOuter  = 1 // Outermost hatching
	HatchStyleIgnore = 2 // Ignore nesting style
)

// Gradient types
const (
	GradientTypeSolid     = 0
	GradientTypeLinear    = 1
	GradientTypeCylinder  = 2
	GradientTypeSpherical = 4
)

// HatchPatternLine represents a line in a hatch pattern
type HatchPatternLine struct {
	Angle           float64   // 53 - Line angle in degrees
	BasePoint       math.Vec2 // 43,44 - Origin point for line
	Offset          math.Vec2 // 45,46 - Offset for next line
	DashLengthItems []float64 // 49 - Dash pattern (positive=dash, negative=gap, 0=dot)
}

// HatchGradientSimple represents gradient fill information
type HatchGradientSimple struct {
	Kind           int       // 450 - Gradient type
	NumberOfColors int       // 453 - Number of colors (0=solid, 2=gradient)
	Color1, Color2 math.Vec3 // Colors for gradient (RGB 0-255 each component)
	ACI1, ACI2     int       // AutoCAD Color Index values
	OneColor       int       // 452 - One/two color (0=two, 1=one)
	Rotation       float64   // 460 - Gradient rotation in radians
	Centered       float64   // 461 - Gradient centering (-1.0 to 1.0)
	Tint           float64   // 462 - Color tint (0.0 to 1.0)
	Name           string    // 470 - Gradient name
}

// HatchColor represents color with ACI and RGB support
type HatchColor struct {
	ACI int       // AutoCAD Color Index (0-255, 256=BYLAYER)
	RGB math.Vec3 // True color (0-255 each component)
}

// NewHatchColor creates a new color
func NewHatchColor(aci int, rgb math.Vec3) HatchColor {
	return HatchColor{
		ACI: aci,
		RGB: rgb,
	}
}

// IsTrueColor returns true if RGB is set (not BYLAYER)
func (c HatchColor) IsTrueColor() bool {
	return c.ACI >= 0 && c.ACI <= 255
}

// IsByLayer returns true if color is BYLAYER
func (c HatchColor) IsByLayer() bool {
	return c.ACI == 256
}

// BoundaryPath represents a hatch boundary path
type BoundaryPath interface {
	Type() int
	PathTypeFlags() int
	SourceBoundaryObjects() []string
	ExportDXF(writer lldxf.TagWriter, ocs math.OCS, elevation float64) error
	Transform(ocs math.OCS, elevation float64) BoundaryPath
	IsValid() bool
}

// PolylinePath represents a polyline boundary path
type PolylinePath struct {
	pathTypeFlags int
	isClosed      bool
	hasBulge      bool
	dashLen       float64
	plotFlags     int
	vertices      []Vertex3D // 2D vertices with optional bulge
	sourceHandles []string
}

// NewPolylinePath creates a polyline boundary path. flags carries the DXF 93
// flags; closed and hasBulge map to codes 73 and 72.
func NewPolylinePath(flags int, closed, hasBulge bool) *PolylinePath {
	return &PolylinePath{
		pathTypeFlags: flags,
		isClosed:      closed,
		hasBulge:      hasBulge,
		plotFlags:     0,
		vertices:      make([]Vertex3D, 0),
	}
}

// AddVertex appends a boundary vertex.
func (p *PolylinePath) AddVertex(v Vertex3D) {
	p.vertices = append(p.vertices, v)
	if v.Bulge != 0 {
		p.hasBulge = true
	}
}

// SetDashLength sets the constant dash length (code 44).
func (p *PolylinePath) SetDashLength(v float64) { p.dashLen = v }

// AddSourceHandle records a source boundary object handle (code 330).
func (p *PolylinePath) AddSourceHandle(h string) { p.sourceHandles = append(p.sourceHandles, h) }

// CompleteVertexY sets the Y of the last appended vertex.
func (p *PolylinePath) CompleteVertexY(y float64) {
	if n := len(p.vertices); n > 0 {
		vt := p.vertices[n-1]
		p.vertices[n-1] = Vertex3D{Point: dxfmath.NewVec3(vt.Point.X(), y, vt.Point.Z()), Bulge: vt.Bulge}
	}
}

// CompleteVertexZ sets the Z of the last appended vertex.
func (p *PolylinePath) CompleteVertexZ(z float64) {
	if n := len(p.vertices); n > 0 {
		vt := p.vertices[n-1]
		p.vertices[n-1] = Vertex3D{Point: dxfmath.NewVec3(vt.Point.X(), vt.Point.Y(), z), Bulge: vt.Bulge}
	}
}

// CompleteVertexBulge sets the bulge of the last appended vertex.
func (p *PolylinePath) CompleteVertexBulge(b float64) {
	if n := len(p.vertices); n > 0 {
		vt := p.vertices[n-1]
		p.vertices[n-1] = Vertex3D{Point: vt.Point, Bulge: b}
		if b != 0 {
			p.hasBulge = true
		}
	}
}

// Vertices returns the boundary vertices.
func (p *PolylinePath) Vertices() []Vertex3D { return p.vertices }

// IsClosed reports whether the path is closed (code 73).
func (p *PolylinePath) IsClosed() bool { return p.isClosed }

// HasBulge reports whether vertices carry bulges (code 72).
func (p *PolylinePath) HasBulge() bool { return p.hasBulge }

// SetClosed sets the closed flag (code 73).
func (p *PolylinePath) SetClosed(c bool) { p.isClosed = c }

// SetHasBulge sets the has-bulge flag (code 72).
func (p *PolylinePath) SetHasBulge(b bool) { p.hasBulge = b }

// Type returns the boundary path type
func (p *PolylinePath) Type() int {
	return BoundaryPathTypePolyline
}

// PathFlags returns the MPOLYGON path-type flags stored at creation.
func (pp *PolylinePath) PathFlags() int { return pp.pathTypeFlags }

// PathTypeFlags returns the boundary path flags
func (p *PolylinePath) PathTypeFlags() int {
	return p.pathTypeFlags
}

// SourceBoundaryObjects returns the source object handles
func (p *PolylinePath) SourceBoundaryObjects() []string {
	return p.sourceHandles
}

// ExportDXF writes the polyline path to DXF
func (p *PolylinePath) ExportDXF(writer lldxf.TagWriter, ocs math.OCS, elevation float64) error {
	tags := lldxf.Tags{}

	// Boundary path header
	tags = tags.Add(lldxf.NewDXFTag(92, fmt.Sprintf("%d", p.Type())))
	tags = tags.Add(lldxf.NewDXFTag(93, p.PathTypeFlags()))
	tags = tags.Add(lldxf.NewDXFTag(72, 0)) // No bulge
	tags = tags.Add(lldxf.NewDXFTag(73, 0)) // Default closed flag
	tags = tags.Add(lldxf.NewDXFTag(97, 0)) // No source objects

	// Polyline vertices
	for _, vertex := range p.vertices {
		tags = tags.Add(lldxf.NewDXFTag(10, vertex.Point.X()))
		tags = tags.Add(lldxf.NewDXFTag(20, vertex.Point.Y()))
		if vertex.Point.Z() != 0 || vertex.Bulge != 0 {
			tags = tags.Add(lldxf.NewDXFTag(30, vertex.Point.Z()))
		}
		if vertex.Bulge != 0 {
			tags = tags.Add(lldxf.NewDXFTag(42, vertex.Bulge))
		}
	}

	// Source objects (if any)
	for _, handle := range p.sourceHandles {
		tags = tags.Add(lldxf.NewDXFTag(330, handle))
	}

	return writer.WriteTags(tags)
}

// Transform applies OCS transformation to the boundary path
func (p *PolylinePath) Transform(ocs math.OCS, elevation float64) BoundaryPath {
	transformed := &PolylinePath{
		pathTypeFlags: p.pathTypeFlags,
		isClosed:      p.isClosed,
		vertices:      make([]Vertex3D, len(p.vertices)),
		sourceHandles: p.sourceHandles,
	}

	// Transform vertices
	for i, vertex := range p.vertices {
		transformed.vertices[i] = Vertex3D{
			Point: ocs.ToWCS(vertex.Point),
			Bulge: vertex.Bulge,
		}
	}

	return transformed
}

// IsValid validates the polyline path
func (p *PolylinePath) IsValid() bool {
	return len(p.vertices) >= 2 // Need at least 2 vertices
}

// Vertex3D represents a 3D point with optional bulge for boundary paths
type Vertex3D struct {
	Point math.Vec3
	Bulge float64
}

// NewVertex3D creates a new 3D vertex
func NewVertex3D(x, y, z, bulge float64) Vertex3D {
	return Vertex3D{
		Point: math.NewVec3(x, y, z),
		Bulge: bulge,
	}
}

// Hatch represents a DXF HATCH entity (comprehensive implementation)
type Hatch struct {
	*entity
	patternType   int                  // 76 - Pattern type
	patternName   string               // 2 - Pattern name
	patternLines  []HatchPatternLine   // Pattern definition
	solidFill     bool                 // 70 - Solid fill flag
	style         int                  // 75 - Hatch style
	color         HatchColor           // 62 - Hatch color
	entityColor   color.ColorNumber    // For Entity interface compatibility
	transparency  float64              // 460 - Transparency (DXF 2004+)
	gradient      *HatchGradientSimple // 471 - Gradient information
	boundaryPaths []BoundaryPath       // 91-98 - Boundary paths
}

// NewHatch creates a new hatch entity
func NewHatch() *Hatch {
	return &Hatch{
		entity:      NewEntity(HATCH),
		patternType: HatchPatternTypePredefined,
		patternName: "ANSI31",
		patternLines: []HatchPatternLine{
			{
				Angle:           0,
				BasePoint:       math.NewVec2(0, 0),
				Offset:          math.NewVec2(0.125, 0),
				DashLengthItems: []float64{0.125, -0.125, 0.125, -0.125, 0.0, -0.125, 0.125, -0.125, 0.0, -0.125},
			},
		},
		solidFill:     true,
		style:         HatchStyleNormal,
		color:         NewHatchColor(1, math.NewVec3(255, 255, 255)),
		boundaryPaths: make([]BoundaryPath, 0),
	}
}

// IsEntity is for Entity interface.
func (h *Hatch) IsEntity() bool {
	return true
}

// Format writes data to formatter.
// SetRecordTypeName overrides the code 0 record name (MPOLYGON loader reuses
// this entity type and must preserve the source record spelling).
func (h *Hatch) SetRecordTypeName(name string) { h.entity.SetRecordTypeName(name) }

// RecordTypeName returns the code 0 record name this hatch emits.
func (h *Hatch) RecordTypeName() string { return h.entity.recordTypeName() }

// RawWholeRecord reports whole-record retention: a HATCH whose boundary paths
// include edge (non-polyline) geometry cannot be re-emitted from the typed
// model, so its retained record is written verbatim.
func (h *Hatch) RawWholeRecord() bool {
	for _, bp := range h.boundaryPaths {
		if _, ok := bp.(*PolylinePath); !ok {
			return true
		}
	}
	return false
}

func (h *Hatch) Format(f format.Formatter) {
	h.entity.Format(f)
	f.WriteString(100, "AcDbHatch")
	// MPOLYGON records emit the path-type flags per boundary path under code
	// 70; the hatch style (codes 70/75) belongs to HATCH records only.
	isMPolygon := h.recordTypeName() == "MPOLYGON"
	if !isMPolygon {
		f.WriteInt(70, h.style)
	}
	if h.solidFill {
		f.WriteInt(71, 1)
	} else {
		f.WriteInt(71, 0)
	}
	if !isMPolygon {
		f.WriteInt(75, h.style)
	}
	f.WriteInt(76, h.patternType)
	f.WriteString(2, h.patternName)
	f.WriteInt(78, len(h.patternLines))
	for _, patternLine := range h.patternLines {
		f.WriteFloat(53, patternLine.Angle)
		f.WriteFloat(43, patternLine.BasePoint.X())
		f.WriteFloat(44, patternLine.BasePoint.Y())
		f.WriteFloat(45, patternLine.Offset.X())
		f.WriteFloat(46, patternLine.Offset.Y())
		f.WriteInt(79, len(patternLine.DashLengthItems))
		for _, dash := range patternLine.DashLengthItems {
			f.WriteFloat(49, dash)
		}
	}

	// Color information
	if h.color.IsTrueColor() {
		f.WriteInt(62, h.color.ACI)
	} else if h.color.IsByLayer() {
		f.WriteInt(62, 256) // BYLAYER
	}

	// Transparency
	if h.transparency != 0 {
		f.WriteFloat(460, h.transparency)
	}

	// Boundary paths
	f.WriteInt(91, len(h.boundaryPaths))
	for _, boundaryPath := range h.boundaryPaths {
		if isMPolygon {
			// MPOLYGON path header: 70 flags, then 93 vertex count.
			flags := 0
			if pp, ok := boundaryPath.(*PolylinePath); ok {
				flags = pp.PathFlags()
			}
			f.WriteInt(70, flags)
			if pp, ok := boundaryPath.(*PolylinePath); ok {
				f.WriteInt(93, len(pp.vertices))
				f.WriteInt(72, boolInt(pp.hasBulge))
				f.WriteInt(73, boolInt(pp.isClosed))
			}
		} else {
			f.WriteInt(92, boundaryPath.Type())
		}

		// Polyline paths emit 73 (closed), 72 (bulge), 44 (dash length), the
		// source-object count with handles, then the vertex count. Edge paths
		// are retained verbatim through rawTags.
		if polylinePath, ok := boundaryPath.(*PolylinePath); ok {
			bulgeFlag := 0
			if polylinePath.hasBulge {
				bulgeFlag = 1
			}
			closedFlag := 0
			if polylinePath.isClosed {
				closedFlag = 1
			}
			if !isMPolygon {
				f.WriteInt(73, closedFlag)
				f.WriteInt(72, bulgeFlag)
			}
			f.WriteFloat(44, polylinePath.dashLen)
			f.WriteInt(97, len(polylinePath.sourceHandles)) // source objects
			for _, handle := range polylinePath.sourceHandles {
				f.WriteString(330, handle)
			}
			f.WriteInt(93, len(polylinePath.vertices))

			for _, vertex := range polylinePath.vertices {
				f.WriteFloat(10, vertex.Point.X())
				f.WriteFloat(20, vertex.Point.Y())
				if vertex.Point.Z() != 0 {
					f.WriteFloat(30, vertex.Point.Z())
				}
				if bulgeFlag != 0 {
					f.WriteFloat(42, vertex.Bulge)
				}
			}
		}
	}
}

// Explode returns the boundary paths as LWPOLYLINE entities (closed flags and
// bulges preserved for polyline paths). Edge paths are not representable and
// are skipped.
func (h *Hatch) Explode() ([]Entity, error) {
	var out []Entity
	for _, bp := range h.boundaryPaths {
		pp, ok := bp.(*PolylinePath)
		if !ok {
			continue
		}
		lw := NewLwPolyline(0)
		for _, v := range pp.vertices {
			lw.Num++
			lw.Vertices = append(lw.Vertices, []float64{v.Point.X(), v.Point.Y()})
			lw.Bulges = append(lw.Bulges, v.Bulge)
		}
		if pp.isClosed {
			lw.Closed = true
		}
		out = append(out, lw)
	}
	return out, nil
}

// Transform applies an affine matrix to every vertex of every polyline
// boundary path. Edge (non-polyline) paths are not modified; they are
// retained verbatim through raw tags.
func (h *Hatch) Transform(m dxfmath.Matrix44) error {
	for _, bp := range h.boundaryPaths {
		pp, ok := bp.(*PolylinePath)
		if !ok {
			continue
		}
		for idx := range pp.vertices {
			v := &pp.vertices[idx]
			p := TransformPoint(m, []float64{v.Point.X(), v.Point.Y(), v.Point.Z()})
			v.Point = dxfmath.NewVec3(p[0], p[1], p[2])
		}
	}
	return nil
}

// BBox returns the axis-aligned bounding box of all polyline boundary paths.
func (h *Hatch) bboxPaths() ([]float64, []float64) {
	mins := []float64{stdmath.Inf(1), stdmath.Inf(1), stdmath.Inf(1)}
	maxs := []float64{stdmath.Inf(-1), stdmath.Inf(-1), stdmath.Inf(-1)}
	found := false
	for _, bp := range h.boundaryPaths {
		pp, ok := bp.(*PolylinePath)
		if !ok {
			continue
		}
		for _, v := range pp.vertices {
			for k, val := range []float64{v.Point.X(), v.Point.Y(), v.Point.Z()} {
				if val < mins[k] {
					mins[k] = val
				}
				if val > maxs[k] {
					maxs[k] = val
				}
			}
			found = true
		}
	}
	if !found {
		return []float64{0, 0, 0}, []float64{0, 0, 0}
	}
	return mins, maxs
}

// BBox returns the bounding box of the hatch boundary paths.
func (h *Hatch) BBox() ([]float64, []float64) {
	return h.bboxPaths()
}

// DefaultPatternLines returns the default ANSI31 pattern definition used for
// non-solid hatches without an explicit pattern section.
func DefaultPatternLines() []HatchPatternLine {
	return NewHatch().patternLines
}

// boolInt maps a bool to the DXF 0/1 integer.
func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

// Style returns the hatch style (code 70).
func (h *Hatch) Style() int { return h.style }

// BoundaryPaths returns the hatch boundary paths.
func (h *Hatch) BoundaryPaths() []BoundaryPath { return h.boundaryPaths }

// SolidFill reports whether the hatch is solid filled (code 70 bit semantics).
func (h *Hatch) SolidFill() bool { return h.solidFill }

// PatternName returns the pattern name (code 2).
func (h *Hatch) PatternName() string { return h.patternName }

// PatternType returns the pattern type (code 76).
func (h *Hatch) PatternType() int { return h.patternType }

// Copy creates a deep copy of the Hatch entity.
func (h *Hatch) Copy() Entity {
	hatch := NewHatch()
	hatch.entity = h.entity
	hatch.patternType = h.patternType
	hatch.patternName = h.patternName
	hatch.patternLines = make([]HatchPatternLine, len(h.patternLines))
	for i, pl := range h.patternLines {
		dashes := append([]float64(nil), pl.DashLengthItems...)
		pl.DashLengthItems = dashes
		hatch.patternLines[i] = pl
	}
	hatch.solidFill = h.solidFill
	hatch.style = h.style
	hatch.color = h.color
	hatch.entityColor = h.entityColor
	hatch.transparency = h.transparency
	hatch.gradient = h.gradient
	hatch.boundaryPaths = make([]BoundaryPath, 0, len(h.boundaryPaths))
	for _, bp := range h.boundaryPaths {
		if pp, ok := bp.(*PolylinePath); ok {
			vs := append([]Vertex3D(nil), pp.vertices...)
			sh := append([]string(nil), pp.sourceHandles...)
			hatch.boundaryPaths = append(hatch.boundaryPaths, &PolylinePath{
				pathTypeFlags: pp.pathTypeFlags,
				isClosed:      pp.isClosed,
				hasBulge:      pp.hasBulge,
				dashLen:       pp.dashLen,
				plotFlags:     pp.plotFlags,
				vertices:      vs,
				sourceHandles: sh,
			})
		} else {
			hatch.boundaryPaths = append(hatch.boundaryPaths, bp)
		}
	}
	return hatch
}

// SetPattern sets the hatch pattern
func (h *Hatch) SetPattern(patternType int, patternName string, patternLines []HatchPatternLine) {
	h.patternType = patternType
	h.patternName = patternName
	h.patternLines = patternLines
}

// SetSolidFill sets the solid fill flag
func (h *Hatch) SetSolidFill(solid bool) {
	h.solidFill = solid
}

// SetStyle sets the hatch style
func (h *Hatch) SetStyle(style int) {
	h.style = style
}

// SetColor sets the hatch color (implements Entity interface)
func (h *Hatch) SetColor(c color.ColorNumber) {
	h.entityColor = c
	h.color = NewHatchColor(int(c), math.NewVec3(float64(c&0xFF), float64((c>>8)&0xFF), float64((c>>16)&0xFF)))
}
func (h *Hatch) SetTransparency(transparency float64) {
	h.transparency = transparency
}

// SetGradient sets the gradient fill information
func (h *Hatch) SetGradient(gradient *HatchGradientSimple) {
	h.gradient = gradient
}

// AddBoundaryPath adds a boundary path
func (h *Hatch) AddBoundaryPath(boundaryPath BoundaryPath) {
	h.boundaryPaths = append(h.boundaryPaths, boundaryPath)
}

// ClearBoundaryPaths removes all boundary paths
func (h *Hatch) ClearBoundaryPaths() {
	h.boundaryPaths = make([]BoundaryPath, 0)
}

// GetBoundaryPaths returns the boundary paths
func (h *Hatch) GetBoundaryPaths() []BoundaryPath {
	return h.boundaryPaths
}

// IsSolid returns true if hatch is a solid fill
func (h *Hatch) IsSolid() bool {
	return h.solidFill
}

// String returns string representation
func (h *Hatch) String() string {
	return fmt.Sprintf("Hatch{Pattern: '%s', Style: %d, Solid: %v}", h.patternName, h.style, h.solidFill)
}
