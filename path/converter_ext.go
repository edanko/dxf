package path

import (
	"fmt"
	"math"

	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
)

// bulgeSegments approximates a bulge arc between two points as polyline
// segments. A zero (or near-zero) bulge yields the straight chord endpoints.
func bulgeSegments(x0, y0, x1, y1, bulge float64, options *Options) []dxfmath.Vec2 {
	a0 := dxfmath.NewVec2(x0, y0)
	a1 := dxfmath.NewVec2(x1, y1)
	b := dxfmath.Bulge(bulge)
	if !b.IsArc() {
		return []dxfmath.Vec2{a0, a1}
	}
	segments := 8
	if sweep := math.Abs(float64(b.Angle())); sweep > 0 {
		segments = int(math.Ceil(sweep / (math.Pi / 12)))
	}
	if segments < 4 {
		segments = 4
	}
	if segments > 64 {
		segments = 64
	}
	if options != nil && options.Segments > segments {
		segments = options.Segments
	}
	return b.ApproximateArc(a0, a1, segments)
}

// appendPolyline appends the given 2D points to p, starting a new subpath.
func appendPolyline(p *Path, pts [][2]float64, closeFlag bool) {
	if len(pts) == 0 {
		return
	}
	p.MoveTo(pts[0][0], pts[0][1])
	for _, pt := range pts[1:] {
		p.LineTo(pt[0], pt[1])
	}
	if closeFlag {
		p.Close()
	}
}

// appendQuad appends a closed quadrilateral built from four points. Duplicate
// trailing points collapse to a triangle (SOLID/TRACE convention).
func appendQuad(p *Path, a, b, c, d [2]float64) {
	if d == c {
		appendPolyline(p, [][2]float64{a, b, d}, true)
		return
	}
	// AutoCAD order: first, second, fourth, third.
	appendPolyline(p, [][2]float64{a, b, d, c}, true)
}

func pt3(coords []float64) [2]float64 {
	return [2]float64{getCoord(coords, 0), getCoord(coords, 1)}
}

// appendBulgeChain appends a vertex chain honouring per-segment bulges.
// bulges[i] applies to the segment from vertices[i] to vertices[i+1]
// (wrapping to the first vertex when the chain is closed).
func appendBulgeChain(p *Path, verts [][2]float64, bulges []float64, closed bool, options *Options) {
	if len(verts) == 0 {
		return
	}
	p.MoveTo(verts[0][0], verts[0][1])
	n := len(verts)
	for i := 0; i < n-1; i++ {
		bulge := 0.0
		if i < len(bulges) {
			bulge = bulges[i]
		}
		x0, y0 := verts[i][0], verts[i][1]
		x1, y1 := verts[i+1][0], verts[i+1][1]
		if math.Abs(bulge) < 1e-10 {
			p.LineTo(x1, y1)
			continue
		}
		for _, seg := range bulgeSegments(x0, y0, x1, y1, bulge, options)[1:] {
			p.LineTo(seg.X(), seg.Y())
		}
	}
	if closed {
		bulge := 0.0
		if n-1 < len(bulges) {
			bulge = bulges[n-1]
		}
		x0, y0 := verts[n-1][0], verts[n-1][1]
		x1, y1 := verts[0][0], verts[0][1]
		if math.Abs(bulge) < 1e-10 {
			p.LineTo(x1, y1)
		} else {
			for _, seg := range bulgeSegments(x0, y0, x1, y1, bulge, options)[1:] {
				p.LineTo(seg.X(), seg.Y())
			}
		}
		p.Close()
	}
}

func splineToPath(spline *entity.Spline, options *Options) (*Path, error) {
	deviation := options.Distance
	if deviation <= 0 {
		deviation = 0.1
	}
	cps := make([]dxfmath.Vec3, 0, len(spline.Controls))
	for _, cp := range spline.Controls {
		cps = append(cps, dxfmath.NewVec3(getCoord(cp, 0), getCoord(cp, 1), getCoord(cp, 2)))
	}
	if len(cps) < 2 {
		return nil, fmt.Errorf("spline requires at least 2 control points, got %d", len(cps))
	}
	degree := max(min(spline.Degree, len(cps)-1), 1)
	var bs *dxfmath.BSpline
	if len(spline.Knots) >= degree+2 && len(spline.Weights) == len(cps) {
		bs = dxfmath.NewBSplineWithKnots(cps, spline.Weights, spline.Knots, degree)
	} else {
		weights := make([]float64, len(cps))
		for i := range weights {
			weights[i] = 1.0
		}
		if len(spline.Knots) >= degree+2 {
			bs = dxfmath.NewBSplineWithKnots(cps, weights, spline.Knots, degree)
		} else {
			bs = dxfmath.NewBSpline(cps, degree)
		}
	}
	lo, hi := bs.KnotDomain()
	var pts []dxfmath.Vec3
	if hi > lo && (lo != 0 || hi != 1) {
		const n = 128
		for i := 0; i <= n; i++ {
			pts = append(pts, bs.EvaluatePoint(lo+(hi-lo)*float64(i)/float64(n)))
		}
	} else {
		pts = bs.Flattening(deviation)
	}
	if len(pts) < 2 {
		return nil, fmt.Errorf("spline sampling produced %d points", len(pts))
	}
	p := NewPath()
	p.MoveTo(pts[0].X(), pts[0].Y())
	for _, pt := range pts[1:] {
		p.LineTo(pt.X(), pt.Y())
	}
	if spline.IsClosed() {
		p.Close()
	}
	return p, nil
}

func hatchToPath(h *entity.Hatch, options *Options) (*Path, error) {
	paths := h.GetBoundaryPaths()
	if len(paths) == 0 {
		return nil, fmt.Errorf("hatch has no boundary paths")
	}
	p := NewPath()
	any := false
	for _, bp := range paths {
		pp, ok := bp.(*entity.PolylinePath)
		if !ok {
			continue
		}
		verts := pp.Vertices()
		if len(verts) == 0 {
			continue
		}
		coords := make([][2]float64, 0, len(verts))
		bulges := make([]float64, 0, len(verts))
		for _, v := range verts {
			coords = append(coords, [2]float64{v.Point.X(), v.Point.Y()})
			bulges = append(bulges, v.Bulge)
		}
		appendBulgeChain(p, coords, bulges, pp.IsClosed(), options)
		any = true
	}
	if !any {
		return nil, fmt.Errorf("hatch boundary paths are not polyline-based")
	}
	return p, nil
}

func solidToPath(solid *entity.Solid, _ *Options) (*Path, error) {
	p := NewPath()
	appendQuad(p, pt3(solid.FirstPoint), pt3(solid.SecondPoint),
		pt3(solid.ThirdPoint), pt3(solid.FourthPoint))
	return p, nil
}

func traceToPath(trace *entity.Trace, _ *Options) (*Path, error) {
	var a, b, c, d [2]float64
	switch {
	case len(trace.Points) >= 8:
		// Loader layout: four flat xy corner pairs.
		a = [2]float64{trace.Points[0], trace.Points[1]}
		b = [2]float64{trace.Points[2], trace.Points[3]}
		c = [2]float64{trace.Points[4], trace.Points[5]}
		d = [2]float64{trace.Points[6], trace.Points[7]}
	case len(trace.Points) == 4:
		// Constructor model: two opposite rectangle corners.
		x1, y1 := trace.Points[0], trace.Points[1]
		x2, y2 := trace.Points[2], trace.Points[3]
		a = [2]float64{x1, y1}
		b = [2]float64{x2, y1}
		c = [2]float64{x2, y2}
		d = [2]float64{x1, y2}
	default:
		return nil, fmt.Errorf("trace requires 4 corner floats, got %d", len(trace.Points))
	}
	p := NewPath()
	appendQuad(p, a, b, c, d)
	return p, nil
}

func face3DToPath(face *entity.ThreeDFace, _ *Options) (*Path, error) {
	if len(face.Points) < 3 {
		return nil, fmt.Errorf("3dface requires at least 3 points, got %d", len(face.Points))
	}
	corners := make([][2]float64, 0, 4)
	for _, pt := range face.Points {
		c := pt3(pt)
		if len(corners) > 0 && corners[len(corners)-1] == c {
			continue
		}
		corners = append(corners, c)
	}
	if len(corners) < 3 {
		return nil, fmt.Errorf("3dface has degenerate points")
	}
	p := NewPath()
	appendPolyline(p, corners, true)
	return p, nil
}

func imageQuad(insert dxfmath.Vec3, u, v dxfmath.Vec3, size dxfmath.Vec2) [][2]float64 {
	w, h := size.X(), size.Y()
	ix, iy := insert.X(), insert.Y()
	ux, uy := u.X(), u.Y()
	vx, vy := v.X(), v.Y()
	return [][2]float64{
		{ix, iy},
		{ix + ux*w, iy + uy*w},
		{ix + ux*w + vx*h, iy + uy*w + vy*h},
		{ix + vx*h, iy + vy*h},
	}
}

func imageToPath(img *entity.Image, _ *Options) (*Path, error) {
	p := NewPath()
	appendPolyline(p, imageQuad(img.InsertPoint(), img.UVector(), img.VVector(), img.ImageSize()), true)
	return p, nil
}

func wipeoutToPath(w *entity.Wipeout, _ *Options) (*Path, error) {
	p := NewPath()
	appendPolyline(p, imageQuad(w.InsertPoint(), w.UVector(), w.VVector(), w.ImageSize()), true)
	return p, nil
}

func leaderToPath(leader *entity.Leader, _ *Options) (*Path, error) {
	pts := leader.GetPoints()
	if len(pts) < 2 {
		return nil, fmt.Errorf("leader requires at least 2 points, got %d", len(pts))
	}
	p := NewPath()
	coords := make([][2]float64, 0, len(pts))
	for _, pt := range pts {
		coords = append(coords, pt3(pt))
	}
	appendPolyline(p, coords, false)
	return p, nil
}

func mleaderToPath(ml *entity.MLeader, _ *Options) (*Path, error) {
	pts := ml.GetLeaderPoints()
	if len(pts) < 2 {
		return nil, fmt.Errorf("mleader requires at least 2 points, got %d", len(pts))
	}
	p := NewPath()
	coords := make([][2]float64, 0, len(pts))
	for _, pt := range pts {
		coords = append(coords, pt3(pt))
	}
	appendPolyline(p, coords, false)
	return p, nil
}

func mlineToPath(ml *entity.MLine, options *Options) (*Path, error) {
	if len(ml.Vertices) < 2 {
		return nil, fmt.Errorf("mline requires at least 2 vertices, got %d", len(ml.Vertices))
	}
	p := NewPath()
	coords := make([][2]float64, 0, len(ml.Vertices))
	bulges := make([]float64, 0, len(ml.Vertices))
	for _, v := range ml.Vertices {
		coords = append(coords, pt3(v.Point))
		bulges = append(bulges, v.Bulge)
	}
	appendBulgeChain(p, coords, bulges, ml.Closed, options)
	return p, nil
}

// pointToPath renders a POINT as a small cross centered on its location.
func pointToPath(pt *entity.Point, _ *Options) (*Path, error) {
	c := pt3(pt.Coord)
	x, y := c[0], c[1]
	const arm = 0.5
	p := NewPath()
	p.MoveTo(x-arm, y)
	p.LineTo(x+arm, y)
	p.MoveTo(x, y-arm)
	p.LineTo(x, y+arm)
	return p, nil
}

func dimensionToPath(dim *entity.Dimension, options *Options) (*Path, error) {
	parts, err := dim.Explode()
	if err != nil {
		return nil, fmt.Errorf("dimension explode: %w", err)
	}
	p := NewPath()
	started := false
	for _, part := range parts {
		if _, ok := part.(*entity.Text); ok {
			continue // text is laid out by the text pipeline
		}
		sub, err := MakePath(part, func(o *Options) { *o = *options })
		if err != nil {
			continue
		}
		for _, sp := range sub.SubPaths() {
			p.MoveTo(sp.Start().X(), sp.Start().Y())
			for _, cmd := range sp.Commands()[1:] {
				switch c := cmd.(type) {
				case *PathLineTo:
					p.LineTo(c.EndVertex().X(), c.EndVertex().Y())
				case *PathCurve3To:
					p.Curve3To(c.End.X(), c.End.Y(), c.Ctrl.X(), c.Ctrl.Y())
				case *PathCurve4To:
					p.Curve4To(c.End.X(), c.End.Y(), c.Ctrl1.X(), c.Ctrl1.Y(), c.Ctrl2.X(), c.Ctrl2.Y())
				}
			}
			if sp.IsClosed() {
				p.Close()
			}
			started = true
		}
	}
	if !started {
		return nil, fmt.Errorf("dimension produced no geometry")
	}
	return p, nil
}

// infiniteLine is implemented by RAY and XLINE entities.
type infiniteLine interface {
	GetPointsForDrawing(infiniteLineLength float64) (dxfmath.Vec3, dxfmath.Vec3)
}

func infiniteLineToPath(e infiniteLine, _ *Options) (*Path, error) {
	length := 1e4
	a, b := e.GetPointsForDrawing(length)
	p := NewPath()
	p.MoveTo(a.X(), a.Y())
	p.LineTo(b.X(), b.Y())
	return p, nil
}
