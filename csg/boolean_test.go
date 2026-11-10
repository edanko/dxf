// Tests for the real triangle-level booleans and mesh primitives.
package csg

import (
	"math"
	"testing"

	dxfmath "github.com/edanko/dxf/math"
)

func assertVolume(t *testing.T, name string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-6 {
		t.Errorf("%s volume = %v, want %v", name, got, want)
	}
}

func TestBoxVolume(t *testing.T) {
	a := Box(dxfmath.NewVec3(0, 0, 0), 2)
	assertVolume(t, "box", Volume(a), 8)
}

func TestBoxDifferenceVolume(t *testing.T) {
	a := Box(dxfmath.NewVec3(0, 0, 0), 2)
	b := Box(dxfmath.NewVec3(0.25, 0.25, 0.25), 1)
	d := MeshDifference(a, b)
	assertVolume(t, "a-b", Volume(d), 7)
}

func TestBoxUnionVolume(t *testing.T) {
	a := Box(dxfmath.NewVec3(0, 0, 0), 2)
	b := Box(dxfmath.NewVec3(0.25, 0.25, 0.25), 1)
	u := MeshUnion(a, b)
	assertVolume(t, "a+b", Volume(u), 8)
}

func TestBoxIntersectVolume(t *testing.T) {
	a := Box(dxfmath.NewVec3(0, 0, 0), 2)
	b := Box(dxfmath.NewVec3(0.25, 0.25, 0.25), 1)
	i := MeshIntersect(a, b)
	assertVolume(t, "a*b", Volume(i), 1)
}

func TestDifferenceNoInvertedFaces(t *testing.T) {
	a := Box(dxfmath.NewVec3(0, 0, 0), 2)
	b := Box(dxfmath.NewVec3(0.25, 0.25, 0.25), 1)
	d := MeshDifference(a, b)

	// A consistently oriented closed mesh traverses every undirected edge
	// exactly once in each direction (cavity faces point into the cavity).
	edges := map[[2]int]int{}
	for _, f := range d.Faces {
		for i := 0; i < len(f); i++ {
			x, y := f[i], f[(i+1)%len(f)]
			key := [2]int{x, y}
			dir := 1
			if x > y {
				key = [2]int{y, x}
				dir = -1
			}
			edges[key] += dir
		}
	}
	for e, n := range edges {
		if n != 0 {
			t.Fatalf("inconsistent orientation at edge %v: %d", e, n)
		}
	}
	if vol := Volume(d); vol <= 0 {
		t.Fatalf("difference volume must be positive, got %v", vol)
	}
}

func TestSphereVolume(t *testing.T) {
	s := Sphere(dxfmath.NewVec3(0, 0, 0), 1, 32)
	got := Volume(s)
	want := 4.0 / 3.0 * math.Pi
	if math.Abs(got-want) > 0.05 {
		t.Errorf("sphere volume = %v, want ~%v", got, want)
	}
}
