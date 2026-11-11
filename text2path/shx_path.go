package text2path

import (
	"math"

	dxfmath "github.com/edanko/dxf/math"
	"github.com/edanko/dxf/path"
)

// ShapePath converts a shape record of the font into an outline path at the
// requested height. Shape coordinates are scaled so that the font cap height
// (the Above header value) maps to height. Pen-up moves start new subpaths.
//
// Octant arcs (codes 10 and 11) are approximated with line segments. Bulge
// polylines (code 13) are approximated with quadratic curves. Sub-shapes
// (code 7) are expanded recursively up to a small depth.
func (f *SHXFont) ShapePath(number int, height float64) *path.Path {
	if f == nil || height <= 0 {
		return nil
	}
	glyph, ok := f.Glyphs[number]
	if !ok {
		return nil
	}
	scale := height
	if f.Above > 0 {
		scale = height / f.Above
	}
	pen := &shxPathPen{
		font:    f,
		path:    path.NewPath(),
		vector:  1,
		visited: map[int]bool{number: true},
	}
	pen.run(glyph.Codes, 0)
	if !pen.wrote {
		return nil
	}
	m := dxfmath.ScaleUniform(scale)
	return pen.path.Transform(&m)
}

type shxPathPen struct {
	font    *SHXFont
	path    *path.Path
	x, y    float64
	vector  float64
	down    bool
	subpath bool
	wrote   bool
	stack   []shxPoint
	visited map[int]bool
}

func (p *shxPathPen) run(codes []int, depth int) {
	if depth > 8 {
		return
	}
	if p.vector == 0 {
		p.vector = 1
	}
	i := 0
	skip := false
	for i < len(codes) {
		code := codes[i]
		i++
		switch {
		case code > 15:
			if !skip {
				p.displace(float64((code>>4)&0xF)*shxVecX[code&0xF], float64((code>>4)&0xF)*shxVecY[code&0xF])
			}
		case code == 0:
			return
		case code == 1:
			p.down = true
			p.subpath = false
		case code == 2:
			p.down = false
			p.subpath = false
		case code == 3, code == 4:
			if i < len(codes) {
				factor := float64(codes[i])
				i++
				if code == 3 && factor != 0 {
					p.vector /= factor
				} else if code == 4 {
					p.vector *= factor
				}
			}
		case code == 5:
			p.stack = append(p.stack, shxPoint{p.x, p.y})
		case code == 6:
			if n := len(p.stack); n > 0 {
				point := p.stack[n-1]
				p.stack = p.stack[:n-1]
				p.move(point.x, point.y)
			}
		case code == 7:
			if i < len(codes) {
				number := codes[i]
				i++
				if !skip && !p.visited[number] {
					if glyph, ok := p.font.Glyphs[number]; ok {
						p.visited[number] = true
						p.run(glyph.Codes, depth+1)
					}
				}
			}
		case code == 8:
			if i+1 < len(codes) {
				x, y := float64(codes[i]), float64(codes[i+1])
				i += 2
				if !skip {
					p.displace(x, y)
				}
			}
		case code == 9:
			for i+1 < len(codes) {
				x, y := float64(codes[i]), float64(codes[i+1])
				i += 2
				if x == 0 && y == 0 {
					break
				}
				if !skip {
					p.displace(x, y)
				}
			}
		case code == 10:
			if i+1 < len(codes) {
				radius := float64(codes[i])
				spec := codes[i+1]
				i += 2
				if !skip {
					p.octantArc(radius, spec)
				}
			}
		case code == 11:
			if i+4 < len(codes) {
				radius := float64((codes[i+2] << 8) + codes[i+3])
				spec := codes[i+4]
				i += 5
				if !skip {
					p.octantArc(radius, spec)
				}
			}
		case code == 12:
			if i+2 < len(codes) {
				x, y := float64(codes[i]), float64(codes[i+1])
				i += 3
				if !skip {
					p.displace(x, y)
				}
			}
		case code == 13:
			for i+1 < len(codes) {
				x, y := float64(codes[i]), float64(codes[i+1])
				if x == 0 && y == 0 {
					i += 2
					break
				}
				i += 2
				bulge := 0.0
				if i < len(codes) {
					bulge = float64(codes[i])
					i++
				}
				if !skip {
					p.bulgeSegment(x, y, bulge)
				}
			}
		case code == 14:
			skip = !skip
		}
		if code != 14 {
			skip = false
		}
	}
}

// octantArc draws an octant arc starting at the current point. The spec byte
// holds the start octant in the low three bits and the octant count in bits
// 4-6 (zero means eight). Bit 7 selects clockwise direction.
func (p *shxPathPen) octantArc(radius float64, spec int) {
	radius *= p.vector
	if radius <= 0 {
		return
	}
	startOctant := float64(spec & 7)
	count := (spec >> 4) & 7
	if count == 0 {
		count = 8
	}
	startAngle := startOctant * math.Pi / 4
	centerX := p.x - radius*math.Cos(startAngle)
	centerY := p.y - radius*math.Sin(startAngle)
	sweep := float64(count) * math.Pi / 4
	if spec&0x80 != 0 {
		sweep = -sweep
	}
	steps := count * 8
	for k := 1; k <= steps; k++ {
		angle := startAngle + sweep*float64(k)/float64(steps)
		p.lineTo(centerX+radius*math.Cos(angle), centerY+radius*math.Sin(angle))
	}
}

// bulgeSegment draws a straight or bulged segment from the current point.
func (p *shxPathPen) bulgeSegment(dx, dy, bulge float64) {
	startX, startY := p.x, p.y
	endX := startX + dx*p.vector
	endY := startY + dy*p.vector
	if bulge == 0 || !p.down {
		p.lineTo(endX, endY)
		return
	}
	chord := math.Hypot(dx, dy) * p.vector
	midX := (startX + endX) / 2
	midY := (startY + endY) / 2
	nx := -dy
	ny := dx
	if length := math.Hypot(nx, ny); length > 0 {
		nx /= length
		ny /= length
	}
	sagitta := bulge * chord / 2
	apexX := midX + nx*sagitta
	apexY := midY + ny*sagitta
	ctrlX := 2*apexX - midX
	ctrlY := 2*apexY - midY
	if p.down {
		p.path.Curve3To(ctrlX, ctrlY, endX, endY)
		p.wrote = true
	} else {
		p.path.MoveTo(endX, endY)
	}
	p.x, p.y = endX, endY
}

func (p *shxPathPen) displace(x, y float64) {
	p.lineTo(p.x+x*p.vector, p.y+y*p.vector)
}

func (p *shxPathPen) lineTo(x, y float64) {
	if p.down {
		if !p.subpath {
			p.path.MoveTo(p.x, p.y)
			p.subpath = true
		}
		p.path.LineTo(x, y)
		p.wrote = true
	} else {
		p.path.MoveTo(x, y)
	}
	p.x, p.y = x, y
}

func (p *shxPathPen) move(x, y float64) {
	p.path.MoveTo(x, y)
	p.x, p.y = x, y
	p.subpath = false
}
