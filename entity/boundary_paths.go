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

// AddArc adds an arc using bulge calculation
func (bp *MPolygonBoundaryPath) AddArc(centerX, centerY, radius, startAngle, endAngle float64) {
	// Calculate arc vertices
	segments := 16 // Number of segments for smooth arc
	angleStep := (endAngle - startAngle) / float64(segments)

	for i := 0; i <= segments; i++ {
		angle := startAngle + float64(i)*angleStep
		x := centerX + radius*math.Cos(angle)
		y := centerY + radius*math.Sin(angle)

		if i == 0 {
			bp.AddVertex(x, y, 0) // Start vertex
		} else if i == segments {
			// Calculate bulge for arc
			totalAngle := endAngle - startAngle
			bulge := math.Tan(totalAngle / 4)
			bp.Vertices[len(bp.Vertices)-1].Bulge = bulge
		} else {
			bp.AddVertex(x, y, 0)
		}
	}
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

// GetArcInfo returns arc information if bulge represents an arc
func (bv *BoundaryVertex) GetArcInfo() (center dxfmath.Vec2, radius, startAngle, endAngle float64, ok bool) {
	if bv.Bulge == 0 {
		return dxfmath.Vec2{}, 0, 0, 0, false
	}

	// Calculate arc parameters from bulge
	// This is a simplified implementation
	// In practice, you need adjacent vertices to calculate arc properly
	bulge := bv.Bulge
	theta := 4 * math.Atan(math.Abs(bulge))

	// Radius and chord length relationship (simplified)
	radius = 1.0 // This would be calculated from geometry
	startAngle = 0
	endAngle = theta

	if bulge < 0 {
		endAngle = -endAngle
	}

	return dxfmath.NewVec2(bv.X, bv.Y), radius, startAngle, endAngle, true
}
