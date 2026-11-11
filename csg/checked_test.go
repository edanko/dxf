package csg

import (
	"math"
	"strings"
	"testing"

	dxfmath "github.com/edanko/dxf/math"
)

// unitCubes returns two unit cubes at center, offset by 0.5 along +x so
// that they overlap in a 0.5-wide slab.
func unitCubes() (dxfmath.Mesh3D, dxfmath.Mesh3D) {
	a := Box(dxfmath.Vec3{0, 0, 0}, 1)
	b := Box(dxfmath.Vec3{0.5, 0, 0}, 1)
	return a, b
}

func TestPerformBooleanCheckedVolumes(t *testing.T) {
	a, b := unitCubes()
	// a: volume 1; b: volume 1; overlap slab: 0.5 x 1 x 1 = 0.5.
	cases := []struct {
		name string
		op   dxfmath.BooleanOperation
		want float64
	}{
		{"union", dxfmath.BooleanUnion, 1.5},
		{"intersection", dxfmath.BooleanIntersection, 0.5},
		{"difference", dxfmath.BooleanDifference, 0.5},
		{"xor", dxfmath.BooleanXOR, 1.0},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := PerformBooleanChecked(a, b, tc.op)
			if err != nil {
				t.Fatalf("PerformBooleanChecked(%v): %v", tc.op, err)
			}
			assertVolume(t, tc.name, Volume(got), tc.want)
		})
	}
}

func TestPerformBooleanCheckedIntersectionExcludesSecond(t *testing.T) {
	a, b := unitCubes()
	got, err := PerformBooleanChecked(a, b, dxfmath.BooleanDifference)
	if err != nil {
		t.Fatal(err)
	}
	// The b-only part (x > 0.5) must not be part of the difference result.
	maxX := math.Inf(-1)
	for _, v := range got.Vertices {
		if v[0] > maxX {
			maxX = v[0]
		}
	}
	if maxX > 0.5+1e-6 {
		t.Errorf("difference result extends to x=%g, want <= 0.5", maxX)
	}
}

func TestPerformBooleanCheckedRejectsInvalidMesh(t *testing.T) {
	a, b := unitCubes()

	openMesh := dxfmath.Mesh3D{Vertices: a.Vertices, Faces: a.Faces[:len(a.Faces)-1], IsClosed: false}
	if _, err := PerformBooleanChecked(openMesh, b, dxfmath.BooleanUnion); err == nil {
		t.Error("open mesh: want error, got nil")
	} else if !strings.Contains(err.Error(), "open") {
		t.Errorf("open mesh: error %q does not mention open", err)
	}

	empty := dxfmath.Mesh3D{}
	if _, err := PerformBooleanChecked(empty, b, dxfmath.BooleanUnion); err == nil {
		t.Error("empty mesh: want error, got nil")
	} else if !strings.Contains(err.Error(), "mesh1") {
		t.Errorf("empty mesh: error %q does not name mesh1", err)
	}

	invalid := dxfmath.Mesh3D{Vertices: a.Vertices, Faces: [][]int{{0, 1, 99}}, IsClosed: true}
	if _, err := PerformBooleanChecked(b, invalid, dxfmath.BooleanUnion); err == nil {
		t.Error("invalid indices: want error, got nil")
	} else if !strings.Contains(err.Error(), "mesh2") {
		t.Errorf("invalid indices: error %q does not name mesh2", err)
	}

	if _, err := PerformBooleanChecked(a, b, dxfmath.BooleanOperation(42)); err == nil {
		t.Error("unknown operation: want error, got nil")
	}
}

func TestPerformBooleanCheckedRejectsNonManifold(t *testing.T) {
	// Tetrahedron with one face duplicated: edges (0,1), (1,2), (0,2) are
	// each shared by three faces, which is non-manifold but not open.
	v := []dxfmath.Vec3{{0, 0, 0}, {1, 0, 0}, {0, 1, 0}, {0, 0, 1}}
	faces := [][]int{{0, 1, 2}, {0, 1, 3}, {0, 2, 3}, {1, 2, 3}, {0, 1, 2}}
	m := dxfmath.Mesh3D{Vertices: v, Faces: faces, IsClosed: true}
	if _, err := PerformBooleanChecked(m, m, dxfmath.BooleanUnion); err == nil {
		t.Fatal("non-manifold mesh: want error, got nil")
	} else if !strings.Contains(err.Error(), "non-manifold") {
		t.Errorf("non-manifold: error %q does not mention non-manifold", err)
	}
}

func TestPerformBooleanCheckedRejectsEmptyResult(t *testing.T) {
	a := Box(dxfmath.Vec3{0, 0, 0}, 1)
	identical := Box(dxfmath.Vec3{0, 0, 0}, 1)
	if _, err := PerformBooleanChecked(a, identical, dxfmath.BooleanXOR); err == nil {
		t.Error("XOR of identical meshes: want empty-result error, got nil")
	} else if !strings.Contains(err.Error(), "empty") {
		t.Errorf("XOR identical: error %q does not mention empty", err)
	}
}
