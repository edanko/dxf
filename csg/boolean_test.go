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

	// Consistent orientation shows up as the exact a-b volume; flipped
	// faces would corrupt it. (BSP results may contain T-junction edges,
	// so per-edge pairing is not required here.)
	if vol := Volume(d); math.Abs(vol-7) > 1e-6 {
		t.Fatalf("difference volume = %v, want 7", vol)
	}
	if !d.IsClosed {
		t.Fatal("difference result must be marked closed")
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
