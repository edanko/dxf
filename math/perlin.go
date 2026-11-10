package math

import "math"

// Perlin generates deterministic Perlin noise from a provided seed.
type Perlin struct {
	perm [256]int
}

// NewPerlin creates a Perlin noise generator seeded with seed. Identical seeds
// produce identical permutation tables and therefore identical noise.
func NewPerlin(seed int64) *Perlin {
	p := &Perlin{}
	for i := range p.perm {
		p.perm[i] = i
	}
	r := newMathRand(seed)
	for i := 255; i > 0; i-- {
		j := r.intn(i + 1)
		p.perm[i], p.perm[j] = p.perm[j], p.perm[i]
	}
	return p
}

// mathRand is a small deterministic linear congruential generator.
type mathRand struct {
	state uint64
}

func newMathRand(seed int64) *mathRand {
	return &mathRand{state: uint64(seed)}
}

func (r *mathRand) intn(n int) int {
	r.state = r.state*6364136223846793005 + 1442695040888963407
	return int(r.state>>33) % n
}

func fade(t float64) float64 { return t * t * (3 - 2*t) }

func lerp(t, a, b float64) float64 { return a + t*(b-a) }

func (p *Perlin) at(x, y int) int { return p.perm[(p.perm[x&255]+y)&255] }

// grad2 is the classic Perlin 2D gradient operator.
func (p *Perlin) grad2(hash int, x, y float64) float64 {
	switch hash & 3 {
	case 0:
		return x + y
	case 1:
		return -x + y
	case 2:
		return x - y
	default:
		return -x - y
	}
}

// grad3 is the classic Perlin improved-noise gradient operator.
func (p *Perlin) grad3(hash int, x, y, z float64) float64 {
	h := hash & 15
	u := x
	if h >= 8 {
		u = y
	}
	v := y
	if h >= 4 {
		if h == 12 || h == 14 {
			v = x
		} else {
			v = z
		}
	}
	if h&1 != 0 {
		u = -u
	}
	if h&2 != 0 {
		v = -v
	}
	return u + v
}

// Noise2D returns Perlin noise at (x, y). Output is in the range [-1, 1].
func (p *Perlin) Noise2D(x, y float64) float64 {
	xi := int(math.Floor(x)) & 255
	yi := int(math.Floor(y)) & 255
	xf := x - math.Floor(x)
	yf := y - math.Floor(y)
	u := fade(xf)
	v := fade(yf)

	aa := p.at(xi, yi)
	ab := p.at(xi, yi+1)
	ba := p.at(xi+1, yi)
	bb := p.at(xi+1, yi+1)

	lower := lerp(u, p.grad2(aa, xf, yf), p.grad2(ba, xf-1, yf))
	upper := lerp(u, p.grad2(ab, xf, yf-1), p.grad2(bb, xf-1, yf-1))
	return lerp(v, lower, upper)
}

// Noise3D returns Perlin noise at (x, y, z). Output is in the range [-1, 1].
func (p *Perlin) Noise3D(x, y, z float64) float64 {
	xi := int(math.Floor(x)) & 255
	yi := int(math.Floor(y)) & 255
	zi := int(math.Floor(z)) & 255
	xf := x - math.Floor(x)
	yf := y - math.Floor(y)
	zf := z - math.Floor(z)
	u := fade(xf)
	v := fade(yf)
	w := fade(zf)

	// Hash all eight cube corners.
	aaa := p.at3(xi, yi, zi)
	aba := p.at3(xi, yi+1, zi)
	aab := p.at3(xi, yi, zi+1)
	abb := p.at3(xi, yi+1, zi+1)
	baa := p.at3(xi+1, yi, zi)
	bba := p.at3(xi+1, yi+1, zi)
	bab := p.at3(xi+1, yi, zi+1)
	bbb := p.at3(xi+1, yi+1, zi+1)

	x0 := lerp(u, p.grad3(aaa, xf, yf, zf), p.grad3(baa, xf-1, yf, zf))
	x1 := lerp(u, p.grad3(aba, xf, yf-1, zf), p.grad3(bba, xf-1, yf-1, zf))
	x2 := lerp(u, p.grad3(aab, xf, yf, zf-1), p.grad3(bab, xf-1, yf, zf-1))
	x3 := lerp(u, p.grad3(abb, xf, yf-1, zf-1), p.grad3(bbb, xf-1, yf-1, zf-1))

	y0 := lerp(v, x0, x1)
	y1 := lerp(v, x2, x3)
	return lerp(w, y0, y1)
}

// at3 hashes a 3D lattice corner.
func (p *Perlin) at3(x, y, z int) int { return p.perm[(p.perm[(p.perm[x&255]+y)&255]+z)&255] }

// Noise2D returns a seeded function computing 2D Perlin noise at (x, y).
func Noise2D(seed int64) func(x, y float64) float64 {
	p := NewPerlin(seed)
	return p.Noise2D
}

// Noise3D returns a seeded function computing 3D Perlin noise at (x, y, z).
func Noise3D(seed int64) func(x, y, z float64) float64 {
	p := NewPerlin(seed)
	return p.Noise3D
}

// OctaveNoise2D layers octaves of 2D Perlin noise and returns a seeded function.
func OctaveNoise2D(seed int64, octaves int, persistence float64) func(x, y float64) float64 {
	p := NewPerlin(seed)
	return func(x, y float64) float64 {
		total, frequency, amplitude, norm := 0.0, 1.0, 1.0, 0.0
		for i := 0; i < octaves; i++ {
			total += p.Noise2D(x*frequency, y*frequency) * amplitude
			norm += amplitude
			amplitude *= persistence
			frequency *= 2
		}
		if norm == 0 {
			return 0
		}
		return total / norm
	}
}

// OctaveNoise3D layers octaves of 3D Perlin noise and returns a seeded function.
func OctaveNoise3D(seed int64, octaves int, persistence float64) func(x, y, z float64) float64 {
	p := NewPerlin(seed)
	return func(x, y, z float64) float64 {
		total, frequency, amplitude, norm := 0.0, 1.0, 1.0, 0.0
		for i := 0; i < octaves; i++ {
			total += p.Noise3D(x*frequency, y*frequency, z*frequency) * amplitude
			norm += amplitude
			amplitude *= persistence
			frequency *= 2
		}
		if norm == 0 {
			return 0
		}
		return total / norm
	}
}