// Checked boolean operations with explicit input and result validation.
//
// PerformBooleanChecked validates that both input meshes are usable for
// boolean operations (non-empty, well-formed indices, closed and manifold)
// and rejects empty results, returning descriptive errors instead of
// silently producing placeholder geometry.
package csg

import (
	"fmt"
	"math"

	dxfmath "github.com/edanko/dxf/math"
)

// validateMeshForBoolean reports why a mesh cannot participate in a boolean
// operation. name identifies the mesh in error messages.
func validateMeshForBoolean(name string, mesh dxfmath.Mesh3D) error {
	if len(mesh.Vertices) == 0 || len(mesh.Faces) == 0 {
		return fmt.Errorf("boolean %s: mesh is empty", name)
	}
	edgeCount := make(map[[2]int]int)
	for fi, f := range mesh.Faces {
		if len(f) < 3 {
			return fmt.Errorf("boolean %s: face %d has %d vertices, want at least 3", name, fi, len(f))
		}
		for i, v := range f {
			if v < 0 || v >= len(mesh.Vertices) {
				return fmt.Errorf("boolean %s: face %d references vertex %d, but the mesh has %d vertices", name, fi, v, len(mesh.Vertices))
			}
			a, b := v, f[(i+1)%len(f)]
			if a > b {
				a, b = b, a
			}
			edgeCount[[2]int{a, b}]++
		}
	}
	for e, n := range edgeCount {
		if n == 1 {
			return fmt.Errorf("boolean %s: mesh is open: edge (%d, %d) has no matching neighbor face", name, e[0], e[1])
		}
		if n > 2 {
			return fmt.Errorf("boolean %s: mesh is non-manifold: edge (%d, %d) is shared by %d faces", name, e[0], e[1], n)
		}
	}
	return nil
}

// PerformBooleanChecked applies a boolean operation to two closed triangle
// meshes and returns the result together with an explicit error.
//
// Both input meshes must be non-empty, use valid vertex indices, and be
// closed and manifold (every edge shared by exactly two faces). The result
// must be non-empty; an operation that produces no geometry (for example
// XOR of two identical solids) is reported as an error.
//
// Deprecated counterpart: math.PerformBoolean keeps placeholder behavior;
// migrate callers to this function.
func PerformBooleanChecked(mesh1, mesh2 dxfmath.Mesh3D, operation dxfmath.BooleanOperation) (dxfmath.Mesh3D, error) {
	if err := validateMeshForBoolean("mesh1", mesh1); err != nil {
		return dxfmath.Mesh3D{}, err
	}
	if err := validateMeshForBoolean("mesh2", mesh2); err != nil {
		return dxfmath.Mesh3D{}, err
	}

	var out dxfmath.Mesh3D
	switch operation {
	case dxfmath.BooleanUnion:
		out = MeshUnion(mesh1, mesh2)
	case dxfmath.BooleanIntersection:
		out = MeshIntersect(mesh1, mesh2)
	case dxfmath.BooleanDifference:
		out = MeshDifference(mesh1, mesh2)
	case dxfmath.BooleanXOR:
		out = MeshXOR(mesh1, mesh2)
	default:
		return dxfmath.Mesh3D{}, fmt.Errorf("boolean: unsupported operation %d", int(operation))
	}

	if len(out.Vertices) == 0 || len(out.Faces) == 0 {
		return dxfmath.Mesh3D{}, fmt.Errorf("boolean: operation %d produced an empty result", int(operation))
	}
	if vol := math.Abs(Volume(out)); vol <= 1e-9 {
		return dxfmath.Mesh3D{}, fmt.Errorf("boolean: operation %d produced an empty result (volume %g)", int(operation), vol)
	}
	return out, nil
}
