package entity

import (
	"fmt"
	"math"

	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/lldxf"
	dxfmath "github.com/edanko/dxf/math"
)

// Boundary edge types as stored in a HATCH edge boundary path (code 72).
const (
	BoundaryEdgeLine    = 1
	BoundaryEdgeArc     = 2
	BoundaryEdgeEllipse = 3
	BoundaryEdgeSpline  = 4
)

// BoundaryEdge is one typed edge of a hatch or MPOLYGON edge boundary path.
type BoundaryEdge interface {
	// EdgeType returns the DXF code 72 edge type.
	EdgeType() int
	// ExportEdge writes the edge tags (starting with code 72).
	ExportEdge(tags lldxf.Tags) lldxf.Tags
	// TransformEdge returns a copy of the edge transformed by m.
	TransformEdge(m dxfmath.Matrix44) BoundaryEdge
	// BBox grows the given bounds to include the edge geometry.
	BBox(minX, minY, maxX, maxY float64) (float64, float64, float64, float64)
}

// LineEdge is a straight boundary edge.
type LineEdge struct {
	Start dxfmath.Vec2
	End   dxfmath.Vec2
}

// NewLineEdge creates a line edge.
func NewLineEdge(start, end dxfmath.Vec2) *LineEdge {
	return &LineEdge{Start: start, End: end}
}

// EdgeType implements BoundaryEdge.
func (e *LineEdge) EdgeType() int { return BoundaryEdgeLine }

// ExportEdge implements BoundaryEdge.
func (e *LineEdge) ExportEdge(tags lldxf.Tags) lldxf.Tags {
	tags = tags.Add(lldxf.NewDXFTag(72, e.EdgeType()))
	tags = tags.Add(lldxf.NewDXFTag(10, e.Start.X()))
	tags = tags.Add(lldxf.NewDXFTag(20, e.Start.Y()))
	tags = tags.Add(lldxf.NewDXFTag(11, e.End.X()))
	tags = tags.Add(lldxf.NewDXFTag(21, e.End.Y()))
	return tags
}

// TransformEdge implements BoundaryEdge.
func (e *LineEdge) TransformEdge(m dxfmath.Matrix44) BoundaryEdge {
	return &LineEdge{
		Start: transformVec2(m, e.Start),
		End:   transformVec2(m, e.End),
	}
}

// BBox implements BoundaryEdge.
func (e *LineEdge) BBox(minX, minY, maxX, maxY float64) (float64, float64, float64, float64) {
	minX, minY, maxX, maxY = growPoint(minX, minY, maxX, maxY, e.Start)
	return growPoint(minX, minY, maxX, maxY, e.End)
}

// growPoint expands the bounds to include p.
func growPoint(minX, minY, maxX, maxY float64, p dxfmath.Vec2) (float64, float64, float64, float64) {
	if p.X() < minX {
		minX = p.X()
	}
	if p.Y() < minY {
		minY = p.Y()
	}
	if p.X() > maxX {
		maxX = p.X()
	}
	if p.Y() > maxY {
		maxY = p.Y()
	}
	return minX, minY, maxX, maxY
}

// ArcEdge is a circular arc boundary edge. Angles are in degrees and the
// counter-clockwise flag matches DXF code 73.
type ArcEdge struct {
	Center     dxfmath.Vec2
	Radius     float64
	StartAngle float64
	EndAngle   float64
	CCW        bool
}

// NewArcEdge creates a circular arc edge.
func NewArcEdge(center dxfmath.Vec2, radius, startAngle, endAngle float64, ccw bool) *ArcEdge {
	return &ArcEdge{Center: center, Radius: radius, StartAngle: startAngle, EndAngle: endAngle, CCW: ccw}
}

// EdgeType implements BoundaryEdge.
func (e *ArcEdge) EdgeType() int { return BoundaryEdgeArc }

// ExportEdge implements BoundaryEdge.
func (e *ArcEdge) ExportEdge(tags lldxf.Tags) lldxf.Tags {
	tags = tags.Add(lldxf.NewDXFTag(72, e.EdgeType()))
	tags = tags.Add(lldxf.NewDXFTag(10, e.Center.X()))
	tags = tags.Add(lldxf.NewDXFTag(20, e.Center.Y()))
	tags = tags.Add(lldxf.NewDXFTag(40, e.Radius))
	tags = tags.Add(lldxf.NewDXFTag(50, e.StartAngle))
	tags = tags.Add(lldxf.NewDXFTag(51, e.EndAngle))
	ccw := 0
	if e.CCW {
		ccw = 1
	}
	tags = tags.Add(lldxf.NewDXFTag(73, ccw))
	return tags
}

// TransformEdge implements BoundaryEdge. The transformed arc is fitted through
// the transformed start, midpoint, and end points so uniform scale and rotation
// keep the arc exact and general affine transforms stay well-behaved.
func (e *ArcEdge) TransformEdge(m dxfmath.Matrix44) BoundaryEdge {
	start := e.PointAt(e.StartAngle)
	mid := e.PointAt(e.StartAngle + e.Sweep()/2)
	end := e.PointAt(e.StartAngle + e.Sweep())
	center, radius, ok := circleThrough(transformVec2(m, start), transformVec2(m, mid), transformVec2(m, end))
	if !ok {
		return NewLineEdge(transformVec2(m, start), transformVec2(m, end))
	}
	a0 := angleDegrees(center, transformVec2(m, start))
	a1 := angleDegrees(center, transformVec2(m, end))
	return &ArcEdge{Center: center, Radius: radius, StartAngle: a0, EndAngle: a1, CCW: e.CCW}
}

// BBox implements BoundaryEdge.
func (e *ArcEdge) BBox(minX, minY, maxX, maxY float64) (float64, float64, float64, float64) {
	for i := 0; i <= 16; i++ {
		p := e.PointAt(e.StartAngle + e.Sweep()*float64(i)/16)
		minX, minY, maxX, maxY = growPoint(minX, minY, maxX, maxY, p)
	}
	return minX, minY, maxX, maxY
}

// Sweep returns the signed sweep in degrees according to the direction flag.
func (e *ArcEdge) Sweep() float64 {
	span := e.EndAngle - e.StartAngle
	if e.CCW {
		for span <= 0 {
			span += 360
		}
	} else {
		for span >= 0 {
			span -= 360
		}
	}
	return span
}

// PointAt returns the point at the given angle in degrees.
func (e *ArcEdge) PointAt(angleDeg float64) dxfmath.Vec2 {
	a := angleDeg * math.Pi / 180
	return dxfmath.NewVec2(e.Center.X()+e.Radius*math.Cos(a), e.Center.Y()+e.Radius*math.Sin(a))
}

// EllipseEdge is an elliptical arc boundary edge. MajorAxis is the vector from
// the center to the major-axis endpoint; Ratio is minor/major.
type EllipseEdge struct {
	Center     dxfmath.Vec2
	MajorAxis  dxfmath.Vec2
	Ratio      float64
	StartAngle float64
	EndAngle   float64
	CCW        bool
}

// NewEllipseEdge creates an elliptical arc edge.
func NewEllipseEdge(center, majorAxis dxfmath.Vec2, ratio, startAngle, endAngle float64, ccw bool) *EllipseEdge {
	return &EllipseEdge{Center: center, MajorAxis: majorAxis, Ratio: ratio, StartAngle: startAngle, EndAngle: endAngle, CCW: ccw}
}

// EdgeType implements BoundaryEdge.
func (e *EllipseEdge) EdgeType() int { return BoundaryEdgeEllipse }

// ExportEdge implements BoundaryEdge.
func (e *EllipseEdge) ExportEdge(tags lldxf.Tags) lldxf.Tags {
	tags = tags.Add(lldxf.NewDXFTag(72, e.EdgeType()))
	tags = tags.Add(lldxf.NewDXFTag(10, e.Center.X()))
	tags = tags.Add(lldxf.NewDXFTag(20, e.Center.Y()))
	tags = tags.Add(lldxf.NewDXFTag(11, e.MajorAxis.X()))
	tags = tags.Add(lldxf.NewDXFTag(21, e.MajorAxis.Y()))
	tags = tags.Add(lldxf.NewDXFTag(40, e.Ratio))
	tags = tags.Add(lldxf.NewDXFTag(50, e.StartAngle))
	tags = tags.Add(lldxf.NewDXFTag(51, e.EndAngle))
	ccw := 0
	if e.CCW {
		ccw = 1
	}
	tags = tags.Add(lldxf.NewDXFTag(73, ccw))
	return tags
}

// TransformEdge implements BoundaryEdge.
func (e *EllipseEdge) TransformEdge(m dxfmath.Matrix44) BoundaryEdge {
	center := transformVec2(m, e.Center)
	majorEnd := transformVec2(m, e.Center.Add(e.MajorAxis))
	minorEnd := transformVec2(m, e.Center.Add(dxfmath.NewVec2(-e.MajorAxis.Y()*e.Ratio, e.MajorAxis.X()*e.Ratio)))
	major := majorEnd.Sub(center)
	minor := minorEnd.Sub(center)
	majorLen := major.Length()
	if majorLen == 0 {
		return NewLineEdge(center, center)
	}
	ratio := minor.Length() / majorLen
	a0 := angleDegrees(center, transformVec2(m, e.StartPoint()))
	a1 := angleDegrees(center, transformVec2(m, e.EndPoint()))
	return &EllipseEdge{Center: center, MajorAxis: major, Ratio: ratio, StartAngle: a0, EndAngle: a1, CCW: e.CCW}
}

// BBox implements BoundaryEdge.
func (e *EllipseEdge) BBox(minX, minY, maxX, maxY float64) (float64, float64, float64, float64) {
	for i := 0; i <= 16; i++ {
		p := e.PointAt(e.StartAngle + e.Sweep()*float64(i)/16)
		minX, minY, maxX, maxY = growPoint(minX, minY, maxX, maxY, p)
	}
	return minX, minY, maxX, maxY
}

// Sweep returns the signed sweep in degrees according to the direction flag.
func (e *EllipseEdge) Sweep() float64 {
	span := e.EndAngle - e.StartAngle
	if e.CCW {
		for span <= 0 {
			span += 360
		}
	} else {
		for span >= 0 {
			span -= 360
		}
	}
	return span
}

// StartPoint returns the point at the start parameter.
func (e *EllipseEdge) StartPoint() dxfmath.Vec2 { return e.PointAt(e.StartAngle) }

// EndPoint returns the point at the end parameter.
func (e *EllipseEdge) EndPoint() dxfmath.Vec2 { return e.PointAt(e.EndAngle) }

// PointAt returns the point at the given parameter in degrees.
func (e *EllipseEdge) PointAt(angleDeg float64) dxfmath.Vec2 {
	a := angleDeg * math.Pi / 180
	major := e.MajorAxis
	minor := dxfmath.NewVec2(-major.Y()*e.Ratio, major.X()*e.Ratio)
	return dxfmath.NewVec2(
		e.Center.X()+major.X()*math.Cos(a)+minor.X()*math.Sin(a),
		e.Center.Y()+major.Y()*math.Cos(a)+minor.Y()*math.Sin(a))
}

// SplineEdge is a NURBS boundary edge.
type SplineEdge struct {
	Degree   int
	Rational bool
	Periodic bool
	Knots    []float64
	Controls []dxfmath.Vec3
	Weights  []float64
}

// NewSplineEdge creates a spline edge.
func NewSplineEdge(degree int, rational, periodic bool, knots []float64, controls []dxfmath.Vec3, weights []float64) *SplineEdge {
	return &SplineEdge{Degree: degree, Rational: rational, Periodic: periodic, Knots: knots, Controls: controls, Weights: weights}
}

// EdgeType implements BoundaryEdge.
func (e *SplineEdge) EdgeType() int { return BoundaryEdgeSpline }

// ExportEdge implements BoundaryEdge.
func (e *SplineEdge) ExportEdge(tags lldxf.Tags) lldxf.Tags {
	tags = tags.Add(lldxf.NewDXFTag(72, e.EdgeType()))
	tags = tags.Add(lldxf.NewDXFTag(94, e.Degree))
	rational, periodic := 0, 0
	if e.Rational {
		rational = 1
	}
	if e.Periodic {
		periodic = 1
	}
	tags = tags.Add(lldxf.NewDXFTag(73, rational))
	tags = tags.Add(lldxf.NewDXFTag(74, periodic))
	tags = tags.Add(lldxf.NewDXFTag(95, len(e.Knots)))
	for _, k := range e.Knots {
		tags = tags.Add(lldxf.NewDXFTag(40, k))
	}
	tags = tags.Add(lldxf.NewDXFTag(96, len(e.Controls)))
	for i, cp := range e.Controls {
		tags = tags.Add(lldxf.NewDXFTag(10, cp.X()))
		tags = tags.Add(lldxf.NewDXFTag(20, cp.Y()))
		if cp.Z() != 0 {
			tags = tags.Add(lldxf.NewDXFTag(30, cp.Z()))
		}
		if e.Rational && i < len(e.Weights) {
			tags = tags.Add(lldxf.NewDXFTag(42, e.Weights[i]))
		}
	}
	return tags
}

// TransformEdge implements BoundaryEdge.
func (e *SplineEdge) TransformEdge(m dxfmath.Matrix44) BoundaryEdge {
	controls := make([]dxfmath.Vec3, len(e.Controls))
	for i, cp := range e.Controls {
		controls[i] = m.MulVec(cp)
	}
	return &SplineEdge{
		Degree:   e.Degree,
		Rational: e.Rational,
		Periodic: e.Periodic,
		Knots:    append([]float64(nil), e.Knots...),
		Controls: controls,
		Weights:  append([]float64(nil), e.Weights...),
	}
}

// BBox implements BoundaryEdge.
func (e *SplineEdge) BBox(minX, minY, maxX, maxY float64) (float64, float64, float64, float64) {
	for _, cp := range e.Controls {
		p := dxfmath.NewVec2(cp.X(), cp.Y())
		minX, minY, maxX, maxY = growPoint(minX, minY, maxX, maxY, p)
	}
	return minX, minY, maxX, maxY
}

// EdgePath is a HATCH or MPOLYGON boundary path made of typed edges.
type EdgePath struct {
	pathTypeFlags int
	edges         []BoundaryEdge
	sourceHandles []string
	unsupported   []string
}

// NewEdgePath creates an empty edge boundary path.
func NewEdgePath(flags int) *EdgePath {
	return &EdgePath{pathTypeFlags: flags}
}

// AddEdge appends a typed edge.
func (p *EdgePath) AddEdge(e BoundaryEdge) { p.edges = append(p.edges, e) }

// AddUnsupported records an edge that could not be converted.
func (p *EdgePath) AddUnsupported(reason string) { p.unsupported = append(p.unsupported, reason) }

// Unsupported returns the reasons for unconverted edges.
func (p *EdgePath) Unsupported() []string { return p.unsupported }

// Edges returns the typed edges.
func (p *EdgePath) Edges() []BoundaryEdge { return p.edges }

// Type returns the boundary path type (edge path).
func (p *EdgePath) Type() int { return 1 }

// PathTypeFlags returns the boundary path flags.
func (p *EdgePath) PathTypeFlags() int { return p.pathTypeFlags }

// SourceBoundaryObjects returns the source object handles.
func (p *EdgePath) SourceBoundaryObjects() []string { return p.sourceHandles }

// AddSourceHandle records a source object handle (code 330).
func (p *EdgePath) AddSourceHandle(h string) { p.sourceHandles = append(p.sourceHandles, h) }

// FormatBoundary writes the edge path to a formatter (HATCH boundary order).
func (p *EdgePath) FormatBoundary(f format.Formatter) {
	f.WriteInt(92, p.Type())
	f.WriteInt(93, len(p.edges))
	for _, edge := range p.edges {
		switch e := edge.(type) {
		case *LineEdge:
			f.WriteInt(72, e.EdgeType())
			f.WriteFloat(10, e.Start.X())
			f.WriteFloat(20, e.Start.Y())
			f.WriteFloat(11, e.End.X())
			f.WriteFloat(21, e.End.Y())
		case *ArcEdge:
			f.WriteInt(72, e.EdgeType())
			f.WriteFloat(10, e.Center.X())
			f.WriteFloat(20, e.Center.Y())
			f.WriteFloat(40, e.Radius)
			f.WriteFloat(50, e.StartAngle)
			f.WriteFloat(51, e.EndAngle)
			f.WriteInt(73, boolCode(e.CCW))
		case *EllipseEdge:
			f.WriteInt(72, e.EdgeType())
			f.WriteFloat(10, e.Center.X())
			f.WriteFloat(20, e.Center.Y())
			f.WriteFloat(11, e.MajorAxis.X())
			f.WriteFloat(21, e.MajorAxis.Y())
			f.WriteFloat(40, e.Ratio)
			f.WriteFloat(50, e.StartAngle)
			f.WriteFloat(51, e.EndAngle)
			f.WriteInt(73, boolCode(e.CCW))
		case *SplineEdge:
			f.WriteInt(72, e.EdgeType())
			f.WriteInt(94, e.Degree)
			f.WriteInt(73, boolCode(e.Rational))
			f.WriteInt(74, boolCode(e.Periodic))
			f.WriteInt(95, len(e.Knots))
			for _, k := range e.Knots {
				f.WriteFloat(40, k)
			}
			f.WriteInt(96, len(e.Controls))
			for i, cp := range e.Controls {
				f.WriteFloat(10, cp.X())
				f.WriteFloat(20, cp.Y())
				if cp.Z() != 0 {
					f.WriteFloat(30, cp.Z())
				}
				if e.Rational && i < len(e.Weights) {
					f.WriteFloat(42, e.Weights[i])
				}
			}
		}
	}
	f.WriteInt(97, len(p.sourceHandles))
	for _, h := range p.sourceHandles {
		f.WriteString(330, h)
	}
}

// boolCode maps a bool to the DXF 0/1 code value.
func boolCode(v bool) int {
	if v {
		return 1
	}
	return 0
}

// ExportDXF writes the edge path tags.
func (p *EdgePath) ExportDXF(writer lldxf.TagWriter, _ dxfmath.OCS, _ float64) error {
	tags := lldxf.Tags{}
	tags = tags.Add(lldxf.NewDXFTag(92, p.Type()))
	tags = tags.Add(lldxf.NewDXFTag(93, len(p.edges)))
	for _, edge := range p.edges {
		tags = edge.ExportEdge(tags)
	}
	tags = tags.Add(lldxf.NewDXFTag(97, len(p.sourceHandles)))
	for _, h := range p.sourceHandles {
		tags = tags.Add(lldxf.NewDXFTag(330, h))
	}
	return writer.WriteTags(tags)
}

// Transform applies the OCS to the edge path.
func (p *EdgePath) Transform(ocs dxfmath.OCS, elevation float64) BoundaryPath {
	m := ocsMatrix(ocs)
	out := &EdgePath{pathTypeFlags: p.pathTypeFlags, sourceHandles: p.sourceHandles}
	for _, edge := range p.edges {
		out.edges = append(out.edges, edge.TransformEdge(m))
	}
	return out
}

// TransformMatrix applies an affine matrix to every edge.
func (p *EdgePath) TransformMatrix(m dxfmath.Matrix44) *EdgePath {
	out := &EdgePath{pathTypeFlags: p.pathTypeFlags, sourceHandles: p.sourceHandles}
	for _, edge := range p.edges {
		out.edges = append(out.edges, edge.TransformEdge(m))
	}
	return out
}

// IsValid reports whether the path has at least one edge.
func (p *EdgePath) IsValid() bool { return len(p.edges) > 0 }

// BBox returns the bounds of every edge.
func (p *EdgePath) BBox() (float64, float64, float64, float64, bool) {
	if len(p.edges) == 0 {
		return 0, 0, 0, 0, false
	}
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	for _, edge := range p.edges {
		minX, minY, maxX, maxY = edge.BBox(minX, minY, maxX, maxY)
	}
	return minX, minY, maxX, maxY, true
}

// ocsMatrix returns the matrix that maps OCS coordinates to WCS.
func ocsMatrix(ocs dxfmath.OCS) dxfmath.Matrix44 {
	m := dxfmath.Identity()
	xAxis, yAxis := ocs.AxisVectors()
	extrusion := ocs.Extrusion()
	m.Set(0, 0, xAxis.X())
	m.Set(1, 0, xAxis.Y())
	m.Set(2, 0, xAxis.Z())
	m.Set(0, 1, yAxis.X())
	m.Set(1, 1, yAxis.Y())
	m.Set(2, 1, yAxis.Z())
	m.Set(0, 2, extrusion.X())
	m.Set(1, 2, extrusion.Y())
	m.Set(2, 2, extrusion.Z())
	return m
}

// transformVec2 applies m to a 2D point treated as z=0.
func transformVec2(m dxfmath.Matrix44, v dxfmath.Vec2) dxfmath.Vec2 {
	out := m.MulVec(dxfmath.NewVec3(v.X(), v.Y(), 0))
	return dxfmath.NewVec2(out.X(), out.Y())
}

// angleDegrees returns the angle of p around c in degrees.
func angleDegrees(c, p dxfmath.Vec2) float64 {
	return math.Atan2(p.Y()-c.Y(), p.X()-c.X()) * 180 / math.Pi
}

// circleThrough returns the circumscribed circle of three points.
func circleThrough(a, b, c dxfmath.Vec2) (dxfmath.Vec2, float64, bool) {
	d := 2 * (a.X()*(b.Y()-c.Y()) + b.X()*(c.Y()-a.Y()) + c.X()*(a.Y()-b.Y()))
	if math.Abs(d) < 1e-12 {
		return dxfmath.Vec2{}, 0, false
	}
	a2 := a.X()*a.X() + a.Y()*a.Y()
	b2 := b.X()*b.X() + b.Y()*b.Y()
	c2 := c.X()*c.X() + c.Y()*c.Y()
	ux := (a2*(b.Y()-c.Y()) + b2*(c.Y()-a.Y()) + c2*(a.Y()-b.Y())) / d
	uy := (a2*(c.X()-b.X()) + b2*(a.X()-c.X()) + c2*(b.X()-a.X())) / d
	center := dxfmath.NewVec2(ux, uy)
	return center, center.Sub(a).Length(), true
}

// String describes an edge for diagnostics.
func (p *EdgePath) String() string {
	return fmt.Sprintf("edge path (%d edges)", len(p.edges))
}
