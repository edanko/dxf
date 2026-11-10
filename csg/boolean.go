// Real triangle-level boolean operations on closed triangle meshes.
//
// The implementation splits triangles against the unique face planes of the
// other solid and classifies the fragments. It is exact for convex and
// mildly concave solids; self-intersecting inputs are not repaired (the
// documented numerical limits of the add-on).
package csg

import (
	"fmt"
	"math"

	dxfmath "github.com/edanko/dxf/math"
)

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

// facePlanes returns the unique face planes of the mesh with outward
// normals (NewCSGPlane normalizes the normal).
func facePlanes(mesh dxfmath.Mesh3D) []Plane {
	seen := map[string]bool{}
	var out []Plane
	for _, f := range mesh.Faces {
		if len(f) < 3 {
			continue
		}
		v0 := mesh.Vertices[f[0]]
		v1 := mesh.Vertices[f[1]]
		v2 := mesh.Vertices[f[2]]
		n := v1.Sub(v0).Cross(v2.Sub(v0))
		if n.Length() < 1e-12 {
			continue
		}
		pl := NewCSGPlane(n, v0)
		key := fmt.Sprintf("%.5f,%.5f,%.5f,%.5f", pl.Normal[0], pl.Normal[1], pl.Normal[2], pl.W)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, pl)
	}
	return out
}

// clipPolygonByPlane clips a polygon against a plane (Sutherland-Hodgman),
// keeping the positive (d >= 0) or negative (d <= 0) side.
func clipPolygonByPlane(poly []dxfmath.Vec3, pl Plane, keepPositive bool) []dxfmath.Vec3 {
	if len(poly) == 0 {
		return nil
	}
	var out []dxfmath.Vec3
	for i := range poly {
		a := poly[i]
		b := poly[(i+1)%len(poly)]
		da := pl.DistanceToPoint(a)
		db := pl.DistanceToPoint(b)
		inA := (da >= 0) == keepPositive
		inB := (db >= 0) == keepPositive
		if inA {
			out = append(out, a)
		}
		if inA != inB {
			t := da / (da - db)
			out = append(out, a.Add(b.Sub(a).Mul(t)))
		}
	}
	return out
}

// fanTriangulate splits a convex polygon into triangles.
func fanTriangulate(poly []dxfmath.Vec3) [][3]dxfmath.Vec3 {
	var out [][3]dxfmath.Vec3
	for i := 1; i+1 < len(poly); i++ {
		out = append(out, [3]dxfmath.Vec3{poly[0], poly[i], poly[i+1]})
	}
	return out
}

// clipTriByPlane splits a triangle into its positive and negative parts.
func clipTriByPlane(tri [3]dxfmath.Vec3, pl Plane) (pos, neg [][3]dxfmath.Vec3) {
	posPoly := clipPolygonByPlane(tri[:], pl, true)
	negPoly := clipPolygonByPlane(tri[:], pl, false)
	return fanTriangulate(posPoly), fanTriangulate(negPoly)
}

// triFrag is a triangle fragment with its inside flag.
type triFrag struct {
	verts  [3]dxfmath.Vec3
	inside bool
}

// splitTris clips triangles against planes, returning the fragments inside
// (negative side of every plane) and outside the clipped solid.
func splitTris(tris [][3]dxfmath.Vec3, planes []Plane) (inside, outside [][3]dxfmath.Vec3) {
	for _, tri := range tris {
		frags := []triFrag{{verts: tri, inside: true}}
		for _, pl := range planes {
			var next []triFrag
			for _, f := range frags {
				pos, neg := clipTriByPlane(f.verts, pl)
				for _, p := range pos {
					next = append(next, triFrag{verts: p, inside: false})
				}
				for _, n := range neg {
					next = append(next, triFrag{verts: n, inside: f.inside})
				}
			}
			frags = next
		}
		for _, f := range frags {
			if f.inside {
				inside = append(inside, f.verts)
			} else {
				outside = append(outside, f.verts)
			}
		}
	}
	return inside, outside
}

// reverseTri flips the winding of a triangle.
func reverseTri(t [3]dxfmath.Vec3) [3]dxfmath.Vec3 {
	return [3]dxfmath.Vec3{t[0], t[2], t[1]}
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
	_, out1 := splitTris(triangulate(mesh1), facePlanes(mesh2))
	_, out2 := splitTris(triangulate(mesh2), facePlanes(mesh1))
	return meshFromTris(append(out1, out2...))
}

// meshIntersect returns the intersection of two closed meshes.
func meshIntersect(mesh1, mesh2 dxfmath.Mesh3D) dxfmath.Mesh3D {
	in1, _ := splitTris(triangulate(mesh1), facePlanes(mesh2))
	in2, _ := splitTris(triangulate(mesh2), facePlanes(mesh1))
	return meshFromTris(append(in1, in2...))
}

// meshDifference returns mesh1 minus mesh2.
func meshDifference(mesh1, mesh2 dxfmath.Mesh3D) dxfmath.Mesh3D {
	_, out1 := splitTris(triangulate(mesh1), facePlanes(mesh2))
	in2, _ := splitTris(triangulate(mesh2), facePlanes(mesh1))
	for i := range in2 {
		in2[i] = reverseTri(in2[i])
	}
	return meshFromTris(append(out1, in2...))
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
		theta := math.Pi * float64(i) / float64(segments)
		for j := 0; j < segments; j++ {
			phi := 2 * math.Pi * float64(j) / float64(segments)
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
