package path

import (
	"math"
	"testing"

	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
)

func vertexAt(p *Path, i int) (float64, float64) {
	pts := p.Vertices()
	return pts[i].X(), pts[i].Y()
}

func TestMakePathEllipseFull(t *testing.T) {
	e := entity.NewEllipse()
	e.Center = []float64{5, 5, 0}
	e.Major = 4
	e.Minor = 2
	e.Start = 0
	e.End = 360

	p, err := MakePath(e)
	if err != nil {
		t.Fatalf("MakePath: %v", err)
	}
	if !p.IsClosed() {
		t.Error("full ellipse must be closed")
	}
	min, max := p.BoundingBox()
	if math.Abs(min.X()-1) > 0.2 || math.Abs(max.X()-9) > 0.2 ||
		math.Abs(min.Y()-3) > 0.2 || math.Abs(max.Y()-7) > 0.2 {
		t.Errorf("ellipse bbox (%v)-(%v), want (1,3)-(9,7)", min, max)
	}
}

func TestMakePathEllipseArc(t *testing.T) {
	e := entity.NewEllipse()
	e.Center = []float64{0, 0, 0}
	e.Major = 10
	e.Minor = 5
	e.Start = 0
	e.End = 90

	p, err := MakePath(e)
	if err != nil {
		t.Fatalf("MakePath: %v", err)
	}
	if p.IsClosed() {
		t.Error("elliptical arc must not be closed")
	}
	x0, y0 := vertexAt(p, 0)
	if math.Abs(x0-10) > 1e-9 || math.Abs(y0) > 1e-9 {
		t.Errorf("arc starts at (%v,%v), want (10,0)", x0, y0)
	}
	x1, y1 := vertexAt(p, len(p.Vertices())-1)
	if math.Abs(x1) > 1e-6 || math.Abs(y1-5) > 1e-6 {
		t.Errorf("arc ends at (%v,%v), want (0,5)", x1, y1)
	}
}

func TestMakePathSpline(t *testing.T) {
	s := entity.NewSpline(3)
	for _, c := range [][3]float64{{0, 0, 0}, {5, 10, 0}, {10, 0, 0}} {
		s.AddControlPoint(c[0], c[1], c[2])
	}
	p, err := MakePath(s, WithDistance(0.01))
	if err != nil {
		t.Fatalf("MakePath: %v", err)
	}
	if len(p.Vertices()) < 8 {
		t.Errorf("spline path has %d vertices, want >= 8", len(p.Vertices()))
	}
	// Path must approach the middle control point's x but stay below its y.
	var maxY float64 = -math.MaxFloat64
	for _, v := range p.Vertices() {
		if v.Y() > maxY {
			maxY = v.Y()
		}
	}
	if maxY <= 0 || maxY >= 10 {
		t.Errorf("spline maxY=%v, want 0 < y < 10", maxY)
	}
}

func TestMakePathLwPolylineBulge(t *testing.T) {
	lw := entity.NewLwPolyline(2)
	lw.Vertices[0] = []float64{0, 0}
	lw.Vertices[1] = []float64{10, 0}
	lw.Bulges[0] = 1 // semicircle over (0,0)->(10,0)

	p, err := MakePath(lw)
	if err != nil {
		t.Fatalf("MakePath: %v", err)
	}
	var botY, midX = math.MaxFloat64, 0.0
	for _, v := range p.Vertices() {
		if v.Y() < botY {
			botY = v.Y()
		}
	}
	for _, v := range p.Vertices() {
		if math.Abs(v.Y()-botY) < 1e-9 {
			midX = v.X()
		}
	}
	// bulge > 0 sweeps CCW: from (0,0) to (10,0) the arc bulges below the chord.
	if math.Abs(botY+5) > 0.2 {
		t.Errorf("semicircle apex y=%v, want ~-5", botY)
	}
	if math.Abs(midX-5) > 0.2 {
		t.Errorf("semicircle apex x=%v, want ~5", midX)
	}
}

func TestMakePathPolylineBulge(t *testing.T) {
	poly := entity.NewPolyline()
	v0 := poly.AddVertex(0, 0, 0)
	v0.Bulge = 0.5
	poly.AddVertex(10, 0, 0)
	p, err := MakePath(poly)
	if err != nil {
		t.Fatalf("MakePath: %v", err)
	}
	pts := p.Vertices()
	if len(pts) <= 3 {
		t.Fatalf("bulge segments not expanded: %d vertices", len(pts))
	}
	// bulge=0.5 on the first vertex sweeps CCW: the arc lies below the chord.
	var minY float64 = 1
	for _, v := range pts {
		if v.Y() < minY {
			minY = v.Y()
		}
	}
	if minY >= -0.5 {
		t.Errorf("bulge arc minY=%v, want < -0.5", minY)
	}
}

func TestMakePathHatchBoundary(t *testing.T) {
	h := entity.NewHatch()
	bp := entity.NewPolylinePath(0, true, false)
	for _, pt := range [][2]float64{{0, 0}, {10, 0}, {10, 5}, {0, 5}} {
		bp.AddVertex(entity.NewVertex3D(pt[0], pt[1], 0, 0))
	}
	h.AddBoundaryPath(bp)
	h.SetSolidFill(true)

	p, err := MakePath(h)
	if err != nil {
		t.Fatalf("MakePath: %v", err)
	}
	if !p.IsClosed() {
		t.Error("closed hatch boundary must close")
	}
	min, max := p.BoundingBox()
	if min.X() != 0 || min.Y() != 0 || max.X() != 10 || max.Y() != 5 {
		t.Errorf("hatch bbox (%v)-(%v), want (0,0)-(10,5)", min, max)
	}
}

func TestMakePathSolidQuad(t *testing.T) {
	s := entity.NewSolid()
	s.FirstPoint = []float64{0, 0, 0}
	s.SecondPoint = []float64{10, 0, 0}
	s.ThirdPoint = []float64{0, 10, 0}
	s.FourthPoint = []float64{10, 10, 0}

	p, err := MakePath(s)
	if err != nil {
		t.Fatalf("MakePath: %v", err)
	}
	if !p.IsClosed() {
		t.Error("solid must be closed")
	}
	if n := len(p.Vertices()); n != 5 {
		t.Errorf("solid vertices=%d, want 5 (quad closed)", n)
	}
}

func TestMakePathTraceQuad(t *testing.T) {
	tr := entity.NewTrace([]float64{0, 0, 10, 10}, 7, 0)
	p, err := MakePath(tr)
	if err != nil {
		t.Fatalf("MakePath: %v", err)
	}
	min, max := p.BoundingBox()
	if min.X() != 0 || max.X() != 10 || min.Y() != 0 || max.Y() != 10 {
		t.Errorf("trace bbox (%v)-(%v), want (0,0)-(10,10)", min, max)
	}
}

func TestMakePath3DFace(t *testing.T) {
	f := entity.New3DFace()
	f.Points = [][]float64{{0, 0, 0}, {10, 0, 0}, {10, 10, 0}, {10, 10, 0}}
	p, err := MakePath(f)
	if err != nil {
		t.Fatalf("MakePath: %v", err)
	}
	if n := len(p.Vertices()); n != 4 {
		t.Errorf("degenerate 3dface vertices=%d, want 4 (triangle closed)", n)
	}
}

func TestMakePathImageQuad(t *testing.T) {
	img := entity.CreateSimpleImage(dxfmath.NewVec3(0, 0, 0), 100, 50, "DEF")
	p, err := MakePath(img)
	if err != nil {
		t.Fatalf("MakePath: %v", err)
	}
	min, max := p.BoundingBox()
	if min.X() != 0 || min.Y() != 0 || max.X() != 100 || max.Y() != 50 {
		t.Errorf("image bbox (%v)-(%v), want (0,0)-(100,50)", min, max)
	}
}

func TestMakePathWipeoutQuad(t *testing.T) {
	w := entity.CreateRectangularWipeout(dxfmath.NewVec3(5, 5, 0), 20, 10)
	p, err := MakePath(w)
	if err != nil {
		t.Fatalf("MakePath: %v", err)
	}
	min, max := p.BoundingBox()
	if min.X() != 5 || min.Y() != 5 || max.X() != 25 || max.Y() != 15 {
		t.Errorf("wipeout bbox (%v)-(%v), want (5,5)-(25,15)", min, max)
	}
}

func TestMakePathLeader(t *testing.T) {
	l := entity.NewLeaderWithPoints([][]float64{{0, 0, 0}, {10, 5, 0}, {20, 5, 0}})
	p, err := MakePath(l)
	if err != nil {
		t.Fatalf("MakePath: %v", err)
	}
	if n := len(p.Vertices()); n != 3 {
		t.Errorf("leader vertices=%d, want 3", n)
	}
	if p.IsClosed() {
		t.Error("leader path must not be closed")
	}
}

func TestMakePathMLeader(t *testing.T) {
	ml := entity.NewMLeader()
	ml.AddLeaderPoint(0, 0, 0)
	ml.AddLeaderPoint(10, 10, 0)
	ml.AddLeaderPoint(20, 10, 0)
	p, err := MakePath(ml)
	if err != nil {
		t.Fatalf("MakePath: %v", err)
	}
	if n := len(p.Vertices()); n != 3 {
		t.Errorf("mleader vertices=%d, want 3", n)
	}
}

func TestMakePathMLine(t *testing.T) {
	ml := entity.NewMLineWithStyle("STANDARD", 0, 1)
	ml.AddVertex(0, 0, 0)
	ml.AddVertex(10, 0, 0)
	ml.AddVertex(10, 10, 0)
	p, err := MakePath(ml)
	if err != nil {
		t.Fatalf("MakePath: %v", err)
	}
	if n := len(p.Vertices()); n != 3 {
		t.Errorf("mline vertices=%d, want 3", n)
	}
}

func TestMakePathPointCross(t *testing.T) {
	pt := entity.NewPoint()
	pt.Coord = []float64{50, 50, 0}
	p, err := MakePath(pt)
	if err != nil {
		t.Fatalf("MakePath: %v", err)
	}
	if len(p.Vertices()) != 4 {
		t.Errorf("point cross vertices=%d, want 4", len(p.Vertices()))
	}
	if !p.HasSubPaths() {
		t.Error("point cross must have two subpaths")
	}
}

func TestMakePathRay(t *testing.T) {
	r := entity.NewRayFromPointDirection(dxfmath.NewVec3(0, 0, 0), dxfmath.NewVec3(10, 0, 0))
	p, err := MakePath(r)
	if err != nil {
		t.Fatalf("MakePath: %v", err)
	}
	if n := len(p.Vertices()); n != 2 {
		t.Errorf("ray vertices=%d, want 2", n)
	}
	if len(p.Vertices()) == 2 && p.Vertices()[1].X() <= 10 {
		t.Error("ray must extend beyond direction point")
	}
}

func TestMakePathDimension(t *testing.T) {
	d := entity.NewDimension()
	d.SetDimensionType(entity.DimensionTypeRotated)
	d.SetExtensionPoints([]float64{0, 0, 0}, []float64{30, 0, 0})
	d.SetDefinitionPoint([]float64{0, 10, 0})
	p, err := MakePath(d)
	if err != nil {
		t.Fatalf("MakePath: %v", err)
	}
	if n := len(p.Vertices()); n < 6 {
		t.Errorf("dimension vertices=%d, want >= 6 (3 lines)", n)
	}
}

func TestMakePathTextRejected(t *testing.T) {
	txt := entity.NewText()
	if _, err := MakePath(txt); err == nil {
		t.Error("TEXT must be routed to text2path, not MakePath")
	}
}
