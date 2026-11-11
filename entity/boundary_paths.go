package entity

import (
	"math"

	"github.com/edanko/dxf/format"
	dxfmath "github.com/edanko/dxf/math"
)

// BoundaryPaths manages multiple boundary paths for MPOLYGON
type BoundaryPaths struct {
	paths []*MPolygonBoundaryPath
}

// NewBoundaryPaths creates a new boundary paths container
func NewBoundaryPaths() *BoundaryPaths {
	return &BoundaryPaths{
		paths: make([]*MPolygonBoundaryPath, 0),
	}
}

// AddPath adds a new boundary path
func (bp *BoundaryPaths) AddPath(pathType int, isClosed bool) *MPolygonBoundaryPath {
	path := &MPolygonBoundaryPath{
		PathType: pathType,
		IsClosed: isClosed,
		Vertices: make([]BoundaryVertex, 0),
	}
	bp.paths = append(bp.paths, path)
	return path
}

// GetPath returns a boundary path by index
func (bp *BoundaryPaths) GetPath(index int) *MPolygonBoundaryPath {
	if index >= 0 && index < len(bp.paths) {
		return bp.paths[index]
	}
	return nil
}

// PathCount returns the number of boundary paths
func (bp *BoundaryPaths) PathCount() int {
	return len(bp.paths)
}

// Clone creates a deep copy of boundary paths
func (bp *BoundaryPaths) Clone() *BoundaryPaths {
	clone := NewBoundaryPaths()

	for _, path := range bp.paths {
		pathClone := path.clone()
		clone.paths = append(clone.paths, pathClone)
	}

	return clone
}

// FormatMPolygon formats boundary paths for MPOLYGON (different order than HATCH)
func (bp *BoundaryPaths) FormatMPolygon(f format.Formatter) {
	f.WriteInt(92, len(bp.paths))

	for _, path := range bp.paths {
		// MPOLYGON specific order: 93, 72, 73 (different from HATCH)
		f.WriteInt(93, len(path.Vertices))

		if len(path.Vertices) > 0 {
			f.WriteInt(72, path.hasBulge())
			f.WriteInt(73, func() int {
				if path.IsClosed {
					return 1
				}
				return 0
			}())
		}

		// Write vertices
		for _, vertex := range path.Vertices {
			vertex.Format(f)
		}
	}
}

// MPolygonBoundaryPath represents a single boundary path for MPOLYGON
type MPolygonBoundaryPath struct {
	PathType int
	IsClosed bool
	Vertices []BoundaryVertex
}

// AddVertex adds a vertex to boundary path
func (bp *MPolygonBoundaryPath) AddVertex(x, y, bulge float64) {
	vertex := BoundaryVertex{
		X:     x,
		Y:     y,
		Bulge: bulge,
	}
	bp.Vertices = append(bp.Vertices, vertex)
}

// AddVertex2D adds a 2D vertex (with zero bulge)
func (bp *MPolygonBoundaryPath) AddVertex2D(x, y float64) {
	bp.AddVertex(x, y, 0.0)
}

// AddArc adds a single arc edge using DXF bulge semantics: the start vertex
// carries the bulge that bends the edge toward the end vertex, so an arc is
// represented by exactly one polyline edge regardless of its sweep. Angles
// are in radians; an end angle below the start angle sweeps clockwise and
// produces a negative bulge.
func (bp *MPolygonBoundaryPath) AddArc(centerX, centerY, radius, startAngle, endAngle float64) {
	span := endAngle - startAngle
	bulge := math.Tan(span / 4)
	sx := centerX + radius*math.Cos(startAngle)
	sy := centerY + radius*math.Sin(startAngle)
	ex := centerX + radius*math.Cos(endAngle)
	ey := centerY + radius*math.Sin(endAngle)
	bp.AddVertex(sx, sy, bulge)
	bp.AddVertex(ex, ey, 0)
}

// GetVertex returns a vertex by index
func (bp *MPolygonBoundaryPath) GetVertex(index int) BoundaryVertex {
	if index >= 0 && index < len(bp.Vertices) {
		return bp.Vertices[index]
	}
	return BoundaryVertex{}
}

// VertexCount returns the number of vertices
func (bp *MPolygonBoundaryPath) VertexCount() int {
	return len(bp.Vertices)
}

// hasBulge returns 1 if any vertex has bulge, 0 otherwise
func (bp *MPolygonBoundaryPath) hasBulge() int {
	for _, vertex := range bp.Vertices {
		if vertex.Bulge != 0 {
			return 1
		}
	}
	return 0
}

// hasAnyBulge reports whether any edge of any path carries a bulge.
func (bp *BoundaryPaths) hasAnyBulge() bool {
	for i := 0; i < bp.PathCount(); i++ {
		if bp.GetPath(i).hasBulge() != 0 {
			return true
		}
	}
	return false
}

// clone creates a deep copy of boundary path
func (bp *MPolygonBoundaryPath) clone() *MPolygonBoundaryPath {
	clone := &MPolygonBoundaryPath{
		PathType: bp.PathType,
		IsClosed: bp.IsClosed,
		Vertices: make([]BoundaryVertex, len(bp.Vertices)),
	}

	for i, vertex := range bp.Vertices {
		clone.Vertices[i] = vertex
	}
	return clone
}

// BoundaryVertex represents a vertex in a boundary path
type BoundaryVertex struct {
	X     float64
	Y     float64
	Bulge float64 // 0 = straight line, non-zero = arc
}

// Format writes vertex to formatter
func (bv *BoundaryVertex) Format(f format.Formatter) {
	f.WriteFloat(10, bv.X)
	f.WriteFloat(20, bv.Y)
	f.WriteFloat(42, bv.Bulge)
}

// ToVec2 converts vertex to dxfmath.Vec2
func (bv *BoundaryVertex) ToVec2() dxfmath.Vec2 {
	return dxfmath.NewVec2(bv.X, bv.Y)
}

// IsLine returns true if vertex represents a straight line (bulge = 0)
func (bv *BoundaryVertex) IsLine() bool {
	return bv.Bulge == 0
}

// IsArc returns true if vertex represents an arc (bulge != 0)
func (bv *BoundaryVertex) IsArc() bool {
	return bv.Bulge != 0
}

// ArcInfo returns the arc geometry of the edge that starts at vertex index:
// its center, radius, start and end angles (radians, consistent with the
// sweep direction), and whether the sweep is counter-clockwise. The arc is
// defined by the bulge of vertex index bending the edge toward the following
// vertex, which wraps around to the first vertex when the path is closed.
// ok is false when the index is out of range, the edge is a straight line
// (bulge 0), or the chord is degenerate.
func (bp *MPolygonBoundaryPath) ArcInfo(index int) (center dxfmath.Vec2, radius, startAngle, endAngle float64, ccw, ok bool) {
	if index < 0 || index >= len(bp.Vertices) {
		return dxfmath.Vec2{}, 0, 0, 0, false, false
	}
	next := index + 1
	if next >= len(bp.Vertices) {
		if !bp.IsClosed {
			return dxfmath.Vec2{}, 0, 0, 0, false, false
		}
		next = 0
	}
	return bp.Vertices[index].GetArcInfo(dxfmath.NewVec2(bp.Vertices[next].X, bp.Vertices[next].Y))
}

// GetArcInfo returns the arc geometry of the edge defined by this vertex's
// bulge bending toward the given edge end point: center, radius, start and
// end angles (radians, consistent with the sweep direction), and whether the
// sweep is counter-clockwise. ok is false when the vertex is a straight line
// (bulge 0) or the chord is degenerate.
func (bv *BoundaryVertex) GetArcInfo(end dxfmath.Vec2) (center dxfmath.Vec2, radius, startAngle, endAngle float64, ccw, ok bool) {
	bulge := dxfmath.Bulge(bv.Bulge)
	if !bulge.IsArc() {
		return dxfmath.Vec2{}, 0, 0, 0, false, false
	}
	start := dxfmath.NewVec2(bv.X, bv.Y)
	center = bulge.Center(start, end)
	radius = bulge.Radius(start.Distance(end))
	if math.IsInf(radius, 0) {
		return dxfmath.Vec2{}, 0, 0, 0, false, false
	}
	startAngle = math.Atan2(start.Y()-center.Y(), start.X()-center.X())
	endAngle = math.Atan2(end.Y()-center.Y(), end.X()-center.X())
	ccw = !bulge.IsClockwise()
	if ccw && endAngle < startAngle {
		endAngle += 2 * math.Pi
	}
	if !ccw && endAngle > startAngle {
		endAngle -= 2 * math.Pi
	}
	return center, radius, startAngle, endAngle, ccw, true
}
