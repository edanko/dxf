package csg

import (
	"fmt"

	dxfmath "github.com/edanko/dxf/math"
)

const (
	Epsilon           = 1e-10
	CoplanarThreshold = 0.0001
)

type Polygon struct {
	Vertices []dxfmath.Vec3
	Normal   dxfmath.Vec3
	Plane    Plane
}

type Plane struct {
	Normal dxfmath.Vec3
	W      float64
}

func NewCSGPlane(normal dxfmath.Vec3, point dxfmath.Vec3) Plane {
	n := normal.Normalized()
	return Plane{Normal: n, W: n.Dot(point)}
}

func (p Plane) DistanceToPoint(pt dxfmath.Vec3) float64 {
	return p.Normal.Dot(pt) - p.W
}

type CSG struct {
	Polygons []Polygon
}

func NewCSG() *CSG {
	return &CSG{Polygons: make([]Polygon, 0)}
}

func MeshToCSG(mesh dxfmath.Mesh3D) *CSG {
	csg := NewCSG()
	for _, face := range mesh.Faces {
		if len(face) < 3 {
			continue
		}
		verts := make([]dxfmath.Vec3, len(face))
		for i, idx := range face {
			if idx < len(mesh.Vertices) {
				verts[i] = mesh.Vertices[idx]
			}
		}
		if len(verts) < 3 {
			continue
		}
		v0, v1, v2 := verts[0], verts[1], verts[2]
		edge1 := v1.Sub(v0)
		edge2 := v2.Sub(v0)
		normal := edge1.Cross(edge2).Normalized()
		if normal.IsZero(Epsilon) {
			continue
		}
		csg.Polygons = append(csg.Polygons, Polygon{
			Vertices: verts,
			Normal:   normal,
			Plane:    NewCSGPlane(normal, v0),
		})
	}
	return csg
}

func (c *CSG) ToMesh() dxfmath.Mesh3D {
	builder := dxfmath.NewMesh3DBuilder()
	vertexMap := make(map[string]int)
	getVertexIndex := func(v dxfmath.Vec3) int {
		key := fmt.Sprintf("%.6f,%.6f,%.6f", v.X(), v.Y(), v.Z())
		if idx, ok := vertexMap[key]; ok {
			return idx
		}
		idx := builder.AddVertex(v)
		vertexMap[key] = idx
		return idx
	}
	for _, poly := range c.Polygons {
		if len(poly.Vertices) < 3 {
			continue
		}
		indices := make([]int, len(poly.Vertices))
		for i, v := range poly.Vertices {
			indices[i] = getVertexIndex(v)
		}
		if len(indices) == 3 {
			builder.AddTriangle(indices[0], indices[1], indices[2])
		} else if len(indices) == 4 {
			builder.AddQuad(indices[0], indices[1], indices[2], indices[3])
		} else {
			for i := 2; i < len(indices); i++ {
				builder.AddTriangle(indices[0], indices[i-1], indices[i])
			}
		}
	}
	return builder.Build()
}

func (a *CSG) Union(b *CSG) dxfmath.Mesh3D {
	return meshUnion(a.ToMesh(), b.ToMesh())
}

func (a *CSG) Intersect(b *CSG) dxfmath.Mesh3D {
	return meshIntersect(a.ToMesh(), b.ToMesh())
}

func (a *CSG) Difference(b *CSG) dxfmath.Mesh3D {
	return meshDifference(a.ToMesh(), b.ToMesh())
}

func meshBoundingBox(mesh dxfmath.Mesh3D) (min, max dxfmath.Vec3) {
	if len(mesh.Vertices) == 0 {
		return
	}
	min = mesh.Vertices[0]
	max = mesh.Vertices[0]
	for _, v := range mesh.Vertices {
		if v.X() < min.X() {
			min = dxfmath.NewVec3(v.Y(), min.Y(), min.Z())
		}
		if v.X() > max.X() {
			max = dxfmath.NewVec3(v.X(), max.Y(), max.Z())
		}
		if v.Y() < min.Y() {
			min = dxfmath.NewVec3(min.X(), v.Y(), min.Z())
		}
		if v.Y() > max.Y() {
			max = dxfmath.NewVec3(max.X(), v.Y(), max.Z())
		}
		if v.Z() < min.Z() {
			min = dxfmath.NewVec3(min.X(), min.Y(), v.Z())
		}
		if v.Z() > max.Z() {
			max = dxfmath.NewVec3(max.X(), max.Y(), v.Z())
		}
	}
	return
}

func containsPoint(bboxMin, bboxMax dxfmath.Vec3, pt dxfmath.Vec3) bool {
	return pt.X() >= bboxMin.X()-Epsilon && pt.X() <= bboxMax.X()+Epsilon &&
		pt.Y() >= bboxMin.Y()-Epsilon && pt.Y() <= bboxMax.Y()+Epsilon &&
		pt.Z() >= bboxMin.Z()-Epsilon && pt.Z() <= bboxMax.Z()+Epsilon
}
