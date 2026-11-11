package render

import (
	"math"

	dxfmath "github.com/edanko/dxf/math"
	"github.com/edanko/dxf/path"
)

// HatchFillSegments generates the stroke segments of a pattern fill clipped
// to the boundary path, in canvas coordinates. The pattern lines (base
// point, offset, dash lengths, all in pattern units) are multiplied by
// hf.Scale; the pattern is rotated by hf.Angle degrees around the origin.
// Segments are kept when their midpoint lies inside the boundary under the
// even-odd rule (hatch islands).
func HatchFillSegments(boundary *path.Path, hf *HatchFill, strokeWidth, deviation float64) *path.Path {
	if boundary == nil || hf == nil || hf.Pattern == nil || len(hf.Pattern.Lines) == 0 {
		return nil
	}
	polys := flattenPath(boundary, deviation)
	if len(polys) == 0 {
		return nil
	}
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	for _, poly := range polys {
		for _, v := range poly {
			minX = math.Min(minX, v.X())
			minY = math.Min(minY, v.Y())
			maxX = math.Max(maxX, v.X())
			maxY = math.Max(maxY, v.Y())
		}
	}
	diag := math.Hypot(maxX-minX, maxY-minY)
	scale := hf.Scale
	if scale == 0 {
		scale = 1
	}
	rot := hf.Angle * math.Pi / 180
	cr, sr := math.Cos(rot), math.Sin(rot)
	out := path.NewPath()
	for _, l := range hf.Pattern.Lines {
		a := (l.Angle*math.Pi/180 + rot)
		ux, uy := math.Cos(a), math.Sin(a)
		nx, ny := -uy, ux
		bx := (l.Base[0]) * scale
		by := (l.Base[1]) * scale
		bx, by = cr*bx-sr*by, sr*bx+cr*by
		ox := l.Offset[0] * scale
		oy := l.Offset[1] * scale
		ox, oy = cr*ox-sr*oy, sr*ox+cr*oy
		// spacing between successive pattern lines along the normal
		sp := ox*nx + oy*ny
		// family index range covering the bbox along the normal
		cMin := nx*(minX-diag) + ny*(minY-diag)
		cMax := nx*(maxX+diag) + ny*(maxY+diag)
		var kMin, kMax int
		if math.Abs(sp) > 1e-12 {
			b0 := bx*nx + by*ny
			kMin = int(math.Floor((cMin-b0)/sp)) - 1
			kMax = int(math.Ceil((cMax-b0)/sp)) + 1
			if kMin > kMax {
				kMin, kMax = kMax, kMin
			}
		} else {
			kMin, kMax = 0, 0 // single line (no offset spacing)
		}
		// dash cycle along the line; an empty cycle is one solid run.
		dash := l.Dashes
		period := 0.0
		for _, d := range dash {
			period += math.Abs(d) * scale
		}
		// parameter range along the line covering the bbox
		tMin := ux*(minX-diag) + uy*(minY-diag) - (ux*bx + uy*by)
		tMax := ux*(maxX+diag) + uy*(maxY+diag) - (ux*bx + uy*by)
		dotLen := strokeWidth * 0.5
		for k := kMin; k <= kMax; k++ {
			oxk := bx + float64(k)*ox
			oyk := by + float64(k)*oy
			if len(dash) == 0 || period <= 1e-12 {
				x0 := oxk + tMin*ux
				y0 := oyk + tMin*uy
				x1 := oxk + tMax*ux
				y1 := oyk + tMax*uy
				if insideBoundary(polys, (x0+x1)/2, (y0+y1)/2) {
					out.MoveTo(x0, y0)
					out.LineTo(x1, y1)
				}
				continue
			}
			// phase: dashes start at the pattern origin projection
			start := math.Floor(tMin/period)*period - period
			t := start
			for t < tMax {
				for _, d := range dash {
					ad := math.Abs(d) * scale
					if d >= 0 {
						ln := ad
						if ln < 1e-9 {
							ln = dotLen
						}
						x0 := oxk + t*ux - ux*ln/2
						y0 := oyk + t*uy - uy*ln/2
						x1 := x0 + ux*ln
						y1 := y0 + uy*ln
						mx, my := (x0+x1)/2, (y0+y1)/2
						if insideBoundary(polys, mx, my) {
							out.MoveTo(x0, y0)
							out.LineTo(x1, y1)
						}
					}
					t += ad
				}
			}
		}
	}
	if len(out.Commands()) == 0 {
		return nil
	}
	return out
}

// insideBoundary applies the even-odd rule across all subpath loops.
func insideBoundary(polys [][]dxfmath.Vec2, x, y float64) bool {
	count := 0
	for _, poly := range polys {
		if len(poly) >= 3 && dxfmath.PointInPolygon(dxfmath.Vec2{x, y}, poly) {
			count++
		}
	}
	return count%2 == 1
}
