package csg

import dxfmath "github.com/edanko/dxf/math"

func MeshUnion(mesh1, mesh2 dxfmath.Mesh3D) dxfmath.Mesh3D {
	a := MeshToCSG(mesh1)
	b := MeshToCSG(mesh2)
	return a.Union(b)
}

func MeshDifference(mesh1, mesh2 dxfmath.Mesh3D) dxfmath.Mesh3D {
	a := MeshToCSG(mesh1)
	b := MeshToCSG(mesh2)
	return a.Difference(b)
}

func MeshXOR(mesh1, mesh2 dxfmath.Mesh3D) dxfmath.Mesh3D {
	union := MeshUnion(mesh1, mesh2)
	intersect := MeshIntersect(mesh1, mesh2)
	return MeshDifference(union, intersect)
}

func MeshIntersect(mesh1, mesh2 dxfmath.Mesh3D) dxfmath.Mesh3D {
	a := MeshToCSG(mesh1)
	b := MeshToCSG(mesh2)
	return a.Intersect(b)
}

func (c *CSG) MeshUnion(other dxfmath.Mesh3D) dxfmath.Mesh3D {
	return MeshUnion(c.ToMesh(), other)
}

func (c *CSG) MeshDifference(other dxfmath.Mesh3D) dxfmath.Mesh3D {
	return MeshDifference(c.ToMesh(), other)
}

func (c *CSG) MeshXOR(other dxfmath.Mesh3D) dxfmath.Mesh3D {
	return MeshXOR(c.ToMesh(), other)
}

func (c *CSG) MeshIntersect(other dxfmath.Mesh3D) dxfmath.Mesh3D {
	return MeshIntersect(c.ToMesh(), other)
}
