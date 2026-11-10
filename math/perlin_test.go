package math

import (
	"math"
	"testing"
)

func TestPerlinDeterministic2D(t *testing.T) {
	a := NewPerlin(42)
	b := NewPerlin(42)
	for _, p := range [][2]float64{{0.5, 0.5}, {1.25, 3.75}, {-2.1, 7.9}, {100.1, -50.3}} {
		av := a.Noise2D(p[0], p[1])
		bv := b.Noise2D(p[0], p[1])
		if av != bv {
			t.Fatalf("seed 42 mismatch at %v: %v != %v", p, av, bv)
		}
	}
}

func TestPerlinDeterministic3D(t *testing.T) {
	a := NewPerlin(7)
	b := NewPerlin(7)
	for _, p := range [][3]float64{{0.5, 0.5, 0.5}, {1.25, 3.75, -2}, {-2.1, 7.9, 11.2}} {
		av := a.Noise3D(p[0], p[1], p[2])
		bv := b.Noise3D(p[0], p[1], p[2])
		if av != bv {
			t.Fatalf("seed 7 mismatch at %v: %v != %v", p, av, bv)
		}
	}
}

func TestPerlinSeedChangesOutput(t *testing.T) {
	a := NewPerlin(1)
	b := NewPerlin(2)
	diff := 0
	for i := 0; i < 50; i++ {
		x := float64(i) * 0.37
		if a.Noise2D(x, x) != b.Noise2D(x, x) {
			diff++
		}
	}
	if diff == 0 {
		t.Fatal("different seeds produced identical noise across all samples")
	}
}

func TestPerlinRange2D(t *testing.T) {
	p := NewPerlin(99)
	for i := 0; i < 2000; i++ {
		x := float64(i) * 0.1113
		y := float64(i) * -0.0719
		v := p.Noise2D(x, y)
		if v < -1.0-1e-9 || v > 1.0+1e-9 {
			t.Fatalf("2D noise out of range at %v,%v: %v", x, y, v)
		}
	}
}

func TestPerlinRange3D(t *testing.T) {
	p := NewPerlin(123)
	for i := 0; i < 2000; i++ {
		x := float64(i) * 0.1113
		y := float64(i) * -0.0719
		z := float64(i) * 0.0331
		v := p.Noise3D(x, y, z)
		if math.IsNaN(v) || v < -1.0-1e-9 || v > 1.0+1e-9 {
			t.Fatalf("3D noise out of range at %v,%v,%v: %v", x, y, z, v)
		}
	}
}

func TestPerlinLatticeValue2D(t *testing.T) {
	// At integer lattice points the gradient dot products are zero.
	p := NewPerlin(5)
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			if got := p.Noise2D(float64(x), float64(y)); math.Abs(got) > 1e-12 {
				t.Fatalf("expected zero at lattice point %d,%d, got %v", x, y, got)
			}
		}
	}
}

func TestNoise2DFunc(t *testing.T) {
	f := Noise2D(11)
	g := Noise2D(11)
	if f(1.3, 2.7) != g(1.3, 2.7) {
		t.Fatal("Noise2D function not deterministic")
	}
}

func TestOctaveNoise2DRange(t *testing.T) {
	f := OctaveNoise2D(3, 4, 0.5)
	for i := 0; i < 500; i++ {
		x := float64(i) * 0.13
		v := f(x, x*0.7)
		if v < -1.0-1e-9 || v > 1.0+1e-9 {
			t.Fatalf("octave noise out of range: %v", v)
		}
	}
}