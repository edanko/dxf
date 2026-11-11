// Real boolean operations on closed triangle meshes, using BSP trees in the
// style of Evan Wallace's csg.js. Splitting, inversion, clipping, and the
// coplanar-polygon handling follow that algorithm so overlapping and
// coplanar inputs are handled consistently.
package csg

import (
	"fmt"
	"math"

	dxfmath "github.com/edanko/dxf/math"
)

// bspPlaneEps is the tolerance used when classifying a vertex against a
// plane (csg.js uses the same value).
const bspPlaneEps = 1e-5

// flip reverses the plane orientation.
func (p *Plane) flip() {
	p.Normal = p.Normal.Mul(-1)
	p.W = -p.W
}

// bspPoly is a convex polygon with its own plane.
type bspPoly struct {
	verts []dxfmath.Vec3
	plane Plane
}

// newBSPPoly computes the polygon plane from its first three vertices.
func newBSPPoly(verts []dxfmath.Vec3) bspPoly {
	p := bspPoly{verts: verts}
	for i := 0; i+2 < len(verts); i++ {
		n := verts[i+1].Sub(verts[i]).Cross(verts[i+2].Sub(verts[i]))
		if n.Length() > 1e-12 {
			p.plane = NewCSGPlane(n, verts[i])
			break
		}
	}
	return p
}

// flip reverses the polygon winding and plane.
func (p *bspPoly) flip() {
	for i, j := 0, len(p.verts)-1; i < j; i, j = i+1, j-1 {
		p.verts[i], p.verts[j] = p.verts[j], p.verts[i]
	}
	p.plane.Normal = p.plane.Normal.Mul(-1)
	p.plane.W = -p.plane.W
}

// Vertex/polygon classification against a splitting plane.
const (
	bspCoplanar  = 0
	bspFrontSide = 1
	bspBackSide  = 2
	bspSpanning  = 3
)

// splitPoly classifies the polygon against the plane and appends it (or its
// fragments) to coplanarFront, coplanarBack, front, and back.
func (pl Plane) splitPoly(p bspPoly, coplanarFront, coplanarBack, front, back *[]bspPoly) {
	ptype := 0
	types := make([]int, len(p.verts))
	for i, v := range p.verts {
		t := bspCoplanar
		switch d := pl.DistanceToPoint(v); {
		case d < -bspPlaneEps:
			t = bspBackSide
		case d > bspPlaneEps:
			t = bspFrontSide
		}
		types[i] = t
		ptype |= t
	}

	switch ptype {
	case bspCoplanar:
		if pl.Normal.Dot(p.plane.Normal) > 0 {
			*coplanarFront = append(*coplanarFront, p)
		} else {
			*coplanarBack = append(*coplanarBack, p)
		}
	case bspFrontSide:
		*front = append(*front, p)
	case bspBackSide:
		*back = append(*back, p)
	default: // bspSpanning
		var f, b []dxfmath.Vec3
		for i := range p.verts {
			j := (i + 1) % len(p.verts)
			ti, tj := types[i], types[j]
			vi, vj := p.verts[i], p.verts[j]
			if ti != bspBackSide {
				f = append(f, vi)
			}
			if ti != bspFrontSide {
				b = append(b, vi)
			}
			if (ti | tj) == bspSpanning {
				t := (pl.W - pl.Normal.Dot(vi)) / pl.Normal.Dot(vj.Sub(vi))
				v := vi.Add(vj.Sub(vi).Mul(t))
				f = append(f, v)
				b = append(b, v)
			}
		}
		if len(f) >= 3 {
			*front = append(*front, newBSPPoly(f))
		}
		if len(b) >= 3 {
			*back = append(*back, newBSPPoly(b))
		}
	}
}

// bspNode is a node in a BSP tree. There is no leaf distinction: every node
// holds polygons and may have front/back subtrees.
type bspNode struct {
	plane *Plane
	front *bspNode
	back  *bspNode
	polys []bspPoly
}

// meshToPolys converts mesh faces to polygons, skipping degenerate faces.
func meshToPolys(mesh dxfmath.Mesh3D) []bspPoly {
	var polys []bspPoly
	for _, f := range mesh.Faces {
		if len(f) < 3 {
			continue
		}
		verts := make([]dxfmath.Vec3, len(f))
		for i, idx := range f {
			verts[i] = mesh.Vertices[idx]
		}
		p := newBSPPoly(verts)
		if p.plane.Normal.Length() < 0.5 {
			continue
		}
		polys = append(polys, p)
	}
	return polys
}

// newBSPNode builds a BSP tree from a mesh.
func newBSPNode(mesh dxfmath.Mesh3D) *bspNode {
	n := &bspNode{}
	n.build(meshToPolys(mesh))
	return n
}

// invert swaps solid and empty space.
func (n *bspNode) invert() {
	for i := range n.polys {
		n.polys[i].flip()
	}
	if n.plane != nil {
		n.plane.flip()
	}
	if n.front != nil {
		n.front.invert()
	}
	if n.back != nil {
		n.back.invert()
	}
	n.front, n.back = n.back, n.front
}

// clipPolygons removes all polygons inside the solid represented by the
// tree (polygons behind every plane with a matching subtree are kept).
func (n *bspNode) clipPolygons(polys []bspPoly) []bspPoly {
	if n.plane == nil {
		return polys
	}
	var front, back []bspPoly
	for _, p := range polys {
		n.plane.splitPoly(p, &front, &back, &front, &back)
	}
	if n.front != nil {
		front = n.front.clipPolygons(front)
	}
	if n.back != nil {
		back = n.back.clipPolygons(back)
	} else {
		back = nil
	}
	return append(front, back...)
}

// clipTo removes all polygons of this tree that are inside the other tree.
func (n *bspNode) clipTo(bsp *bspNode) {
	n.polys = bsp.clipPolygons(n.polys)
	if n.front != nil {
		n.front.clipTo(bsp)
	}
	if n.back != nil {
		n.back.clipTo(bsp)
	}
}

// allPolygons returns every polygon in the tree.
func (n *bspNode) allPolygons() []bspPoly {
	out := append([]bspPoly(nil), n.polys...)
	if n.front != nil {
		out = append(out, n.front.allPolygons()...)
	}
	if n.back != nil {
		out = append(out, n.back.allPolygons()...)
	}
	return out
}

// build partitions polygons using the first polygon's plane; coplanar
// polygons are stored on the node, the rest go to the subtrees.
func (n *bspNode) build(polys []bspPoly) {
	if len(polys) == 0 {
		return
	}
	if n.plane == nil {
		pl := polys[0].plane
		n.plane = &pl
	}
	var front, back []bspPoly
	for _, p := range polys {
		n.plane.splitPoly(p, &n.polys, &n.polys, &front, &back)
	}
	if len(front) > 0 {
		if n.front == nil {
			n.front = &bspNode{}
		}
		n.front.build(front)
	}
	if len(back) > 0 {
		if n.back == nil {
			n.back = &bspNode{}
		}
		n.back.build(back)
	}
}

// polysToMesh converts polygons back to a triangle mesh, deduplicating
// vertices and dropping degenerate faces.
func polysToMesh(polys []bspPoly) dxfmath.Mesh3D {
	var tris [][3]dxfmath.Vec3
	for _, p := range polys {
		for i := 1; i+1 < len(p.verts); i++ {
			tris = append(tris, [3]dxfmath.Vec3{p.verts[0], p.verts[i], p.verts[i+1]})
		}
	}
	return meshFromTris(tris)
}

// triangulate converts mesh faces to triangles (fan per face).
func triangulate(mesh dxfmath.Mesh3D) [][3]dxfmath.Vec3 {
	var out [][3]dxfmath.Vec3
	for _, f := range mesh.Faces {
		if len(f) < 3 {
			continue
		}
		for i := 1; i+1 < len(f); i++ {
			out = append(out, [3]dxfmath.Vec3{
				mesh.Vertices[f[0]],
				mesh.Vertices[f[i]],
				mesh.Vertices[f[i+1]],
			})
		}
	}
	return out
}

// Volume returns the signed volume of a closed mesh (positive for
// outward-facing triangles), computed with the divergence theorem.
func Volume(mesh dxfmath.Mesh3D) float64 {
	vol := 0.0
	for _, t := range triangulate(mesh) {
		vol += t[0].Dot(t[1].Cross(t[2])) / 6
	}
	return vol
}

// meshFromTris builds a closed mesh from triangles, deduplicating vertices
// and dropping degenerate faces.
func meshFromTris(tris [][3]dxfmath.Vec3) dxfmath.Mesh3D {
	seen := map[string]int{}
	var verts []dxfmath.Vec3
	var faces [][]int
	add := func(v dxfmath.Vec3) int {
		key := fmt.Sprintf("%.6f,%.6f,%.6f", v[0], v[1], v[2])
		if i, ok := seen[key]; ok {
			return i
		}
		i := len(verts)
		verts = append(verts, v)
		seen[key] = i
		return i
	}
	for _, t := range tris {
		a, b, c := add(t[0]), add(t[1]), add(t[2])
		if a == b || b == c || a == c {
			continue
		}
		faces = append(faces, []int{a, b, c})
	}
	return dxfmath.Mesh3D{Vertices: verts, Faces: faces, IsClosed: true}
}

// meshUnion returns the union of two closed meshes.
func meshUnion(mesh1, mesh2 dxfmath.Mesh3D) dxfmath.Mesh3D {
	a := newBSPNode(mesh1)
	b := newBSPNode(mesh2)
	a.clipTo(b)
	b.clipTo(a)
	b.invert()
	b.clipTo(a)
	b.invert()
	a.build(b.allPolygons())
	return polysToMesh(a.allPolygons())
}

// meshIntersect returns the intersection of two closed meshes.
func meshIntersect(mesh1, mesh2 dxfmath.Mesh3D) dxfmath.Mesh3D {
	a := newBSPNode(mesh1)
	b := newBSPNode(mesh2)
	a.invert()
	b.clipTo(a)
	b.invert()
	a.clipTo(b)
	b.clipTo(a)
	a.build(b.allPolygons())
	a.invert()
	return polysToMesh(a.allPolygons())
}

// meshDifference returns mesh1 minus mesh2.
func meshDifference(mesh1, mesh2 dxfmath.Mesh3D) dxfmath.Mesh3D {
	a := newBSPNode(mesh1)
	b := newBSPNode(mesh2)
	a.invert()
	a.clipTo(b)
	b.clipTo(a)
	b.invert()
	b.clipTo(a)
	b.invert()
	a.build(b.allPolygons())
	a.invert()
	return polysToMesh(a.allPolygons())
}

// Box returns a closed cube mesh centered at center with the given edge
// size and outward-facing triangles.
func Box(center dxfmath.Vec3, size float64) dxfmath.Mesh3D {
	h := size / 2
	x0, y0, z0 := center[0]-h, center[1]-h, center[2]-h
	x1, y1, z1 := center[0]+h, center[1]+h, center[2]+h
	v := []dxfmath.Vec3{
		{x0, y0, z0}, {x1, y0, z0}, {x1, y1, z0}, {x0, y1, z0},
		{x0, y0, z1}, {x1, y0, z1}, {x1, y1, z1}, {x0, y1, z1},
	}
	faces := [][]int{
		{0, 3, 2, 1}, // bottom, outward -z
		{4, 5, 6, 7}, // top, outward +z
		{0, 1, 5, 4}, // front, outward -y
		{2, 3, 7, 6}, // back, outward +y
		{1, 2, 6, 5}, // right, outward +x
		{3, 0, 4, 7}, // left, outward -x
	}
	return dxfmath.Mesh3D{Vertices: v, Faces: faces, IsClosed: true}
}

// Sphere returns a closed UV-sphere mesh centered at center.
func Sphere(center dxfmath.Vec3, radius float64, segments int) dxfmath.Mesh3D {
	if segments < 3 {
		segments = 16
	}
	var verts []dxfmath.Vec3
	var faces [][]int
	for i := 0; i <= segments; i++ {
		theta := 3.141592653589793 * float64(i) / float64(segments)
		for j := 0; j < segments; j++ {
			phi := 2 * 3.141592653589793 * float64(j) / float64(segments)
			verts = append(verts, dxfmath.Vec3{
				center[0] + radius*math.Sin(theta)*math.Cos(phi),
				center[1] + radius*math.Sin(theta)*math.Sin(phi),
				center[2] + radius*math.Cos(theta),
			})
		}
	}
	idx := func(i, j int) int { return i*segments + j%segments }
	for i := 0; i < segments; i++ {
		for j := 0; j < segments; j++ {
			faces = append(faces, []int{idx(i, j), idx(i+1, j), idx(i+1, j+1), idx(i, j+1)})
		}
	}
	return dxfmath.Mesh3D{Vertices: verts, Faces: faces, IsClosed: true}
}
