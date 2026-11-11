// Primitive decomposition of DXF entities.
package query

import (
	"fmt"
	"math"

	"github.com/edanko/dxf/block"
	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
	"github.com/edanko/dxf/path"
	"github.com/edanko/dxf/text2path"
)

// PrimitiveKind names the geometric representation of a primitive.
type PrimitiveKind string

const (
	PrimitiveLine     PrimitiveKind = "line"
	PrimitiveArc      PrimitiveKind = "arc"
	PrimitiveCircle   PrimitiveKind = "circle"
	PrimitiveEllipse  PrimitiveKind = "ellipse"
	PrimitiveSpline   PrimitiveKind = "spline"
	PrimitivePolyline PrimitiveKind = "polyline"
	PrimitiveKindMesh PrimitiveKind = "mesh"
	PrimitiveText     PrimitiveKind = "text"
	PrimitiveMText    PrimitiveKind = "mtext"
	PrimitiveSolid    PrimitiveKind = "solid"
	PrimitivePoint    PrimitiveKind = "point"
	PrimitiveInsert   PrimitiveKind = "insert"
	PrimitiveUnknown  PrimitiveKind = "unknown"
)

// PrimitiveDiagnostic reports a decomposition problem without aborting it.
type PrimitiveDiagnostic struct {
	Handle  string
	Type    string
	Message string
	Offset  int
}

func (d PrimitiveDiagnostic) String() string {
	name := d.Type
	if d.Handle != "" {
		name += " [" + d.Handle + "]"
	}
	return name + ": " + d.Message
}

// PrimitiveKindMesh is a polygon mesh produced from MESH or POLYLINE records.
type PrimitiveMesh struct {
	Vertices []dxfmath.Vec3
	Faces    [][]int
}

// Primitive is a renderer-independent geometric representation of an entity.
type Primitive struct {
	Kind        PrimitiveKind
	Handle      string
	Vertices    []dxfmath.Vec3
	Path        *path.Path
	Mesh        *PrimitiveMesh
	Control     []dxfmath.Vec3
	Diagnostics []PrimitiveDiagnostic
}

// Decomposition is the result of recursive primitive extraction.
type Decomposition struct {
	Primitives  []Primitive
	Cycles      [][]string
	Diagnostics []PrimitiveDiagnostic
}

// MakePrimitive builds one primitive from an entity. Composite entities may
// carry diagnostics when their representation is only a placeholder; use
// RecursiveDecompose to flatten them.
func MakePrimitive(e entity.Entity) Primitive {
	if e == nil {
		return Primitive{Kind: PrimitiveUnknown, Diagnostics: []PrimitiveDiagnostic{{Message: "nil entity"}}}
	}
	if u, ok := e.(entity.Unwrapper); ok {
		e = u.Unwrap()
	}
	p := Primitive{Handle: e.Handle()}
	switch v := e.(type) {
	case *entity.Line:
		p.Kind = PrimitiveLine
		p.Vertices = vecs(v.Start, v.End)
		p.Path = linePath(v.Start, v.End)
	case *entity.Arc:
		p.Kind = PrimitiveArc
		p.Path = arcPath(v.Center, v.Radius, v.Angle[0], v.Angle[1])
		p.Vertices = pathVecs(p.Path)
	case *entity.Circle:
		p.Kind = PrimitiveCircle
		p.Path = arcPath(v.Center, v.Radius, 0, 360)
		p.Vertices = pathVecs(p.Path)
	case *entity.Ellipse:
		p.Kind = PrimitiveEllipse
		p.Vertices = ellipseVertices(v)
		p.Path = polylinePath(p.Vertices, v.IsFull())
	case *entity.Spline:
		p.Kind = PrimitiveSpline
		for _, c := range v.Controls {
			p.Control = append(p.Control, vec3(c))
		}
		pts := v.Flatten(0.01)
		p.Vertices = pts
		p.Path = polylinePath(pts, v.IsClosed())
	case *entity.LwPolyline:
		p.Kind = PrimitivePolyline
		for _, xy := range v.Vertices {
			z := 0.0
			if len(xy) > 2 {
				z = xy[2]
			}
			p.Vertices = append(p.Vertices, dxfmath.NewVec3(xy[0], xy[1], z))
		}
		p.Path = polylinePath(p.Vertices, v.Closed)
	case *entity.Polyline:
		p.Kind = PrimitivePolyline
		if v.IsPolyface() || v.IsPolymesh() {
			p.Kind = PrimitiveKindMesh
			p.Mesh = polylineMesh(v)
			p.Vertices = append([]dxfmath.Vec3(nil), p.Mesh.Vertices...)
		} else {
			for _, vertex := range v.Vertices {
				if vertex == nil || len(vertex.Coord) < 3 {
					continue
				}
				p.Vertices = append(p.Vertices, vec3(vertex.Coord))
			}
			p.Path = polylinePath(p.Vertices, v.Flag&entity.PolylineFlagClosed != 0)
		}
	case *entity.Mesh:
		p.Kind = PrimitiveKindMesh
		p.Mesh = &PrimitiveMesh{Vertices: append([]dxfmath.Vec3(nil), v.Vertices()...)}
		for _, f := range v.Faces() {
			p.Mesh.Faces = append(p.Mesh.Faces, append([]int(nil), f...))
		}
		p.Vertices = append([]dxfmath.Vec3(nil), p.Mesh.Vertices...)
	case *entity.Text:
		p.Kind = PrimitiveText
		p.Vertices = textVertices(v.Value, v.Height, v.Coord1, v.Rotation)
		p.Path = textPath(v.Value, v.Height, v.Coord1[0], v.Coord1[1], v.Rotation)
	case *entity.MText:
		p.Kind = PrimitiveMText
		lines := text2path.LayoutMText(v.GetText(), v.Height, v.Width)
		for i, line := range lines {
			y := v.Coord1[1] - float64(i)*1.6714*v.Height
			p.Vertices = append(p.Vertices, textVertices(line, v.Height, []float64{v.Coord1[0], y, 0}, v.Rotation)...)
			if lp := textPath(line, v.Height, v.Coord1[0], y, v.Rotation); lp != nil {
				if p.Path == nil {
					p.Path = path.NewPath()
				}
				appendPath(p.Path, lp)
			}
		}
	case *entity.Solid:
		p.Kind = PrimitiveSolid
		p.Vertices = vecs(v.FirstPoint, v.SecondPoint, v.ThirdPoint, v.FourthPoint)
		p.Path = polylinePath(p.Vertices, true)
	case *entity.Point:
		p.Kind = PrimitivePoint
		p.Vertices = vecs(v.Coord)
	case *entity.Insert:
		p.Kind = PrimitiveInsert
		p.Vertices = vecs([]float64{v.InsertPoint().X(), v.InsertPoint().Y(), v.InsertPoint().Z()})
		p.Diagnostics = append(p.Diagnostics, PrimitiveDiagnostic{
			Handle: e.Handle(), Type: e.DXFType(), Message: "composite INSERT; use RecursiveDecompose to expand",
		})
	case *entity.Hatch:
		p.Kind = PrimitivePolyline
		for _, bp := range v.BoundaryPaths() {
			pp, ok := bp.(*entity.PolylinePath)
			if !ok {
				continue
			}
			var pts []dxfmath.Vec3
			for _, vertex := range pp.Vertices() {
				pts = append(pts, vertex.Point)
			}
			if p.Path == nil {
				p.Path = path.NewPath()
			}
			appendPath(p.Path, polylinePath(pts, pp.IsClosed()))
			p.Vertices = append(p.Vertices, pts...)
		}
	default:
		p.Kind = PrimitiveUnknown
		p.Diagnostics = append(p.Diagnostics, PrimitiveDiagnostic{
			Handle: e.Handle(), Type: e.DXFType(), Message: "unsupported entity type",
		})
	}
	return p
}

// RecursiveDecompose expands composites into leaf primitives. INSERTs resolve
// through lookup, MTEXT and POLYLINE keep their path/mesh representation, and
// HATCH boundaries are emitted as paths. Cyclic INSERT references terminate
// and are reported in Cycles.
func RecursiveDecompose(ents []entity.Entity, lookup BlockLookup) *Decomposition {
	d := &Decomposition{}
	d.walk(ents, dxfmath.Identity(), nil, lookup, 0)
	return d
}

const maxDecomposeDepth = 32

func (d *Decomposition) walk(ents []entity.Entity, m dxfmath.Matrix44, stack []string, lookup BlockLookup, depth int) {
	if depth > maxDecomposeDepth {
		d.Diagnostics = append(d.Diagnostics, PrimitiveDiagnostic{Message: fmt.Sprintf("decomposition depth limit %d exceeded", maxDecomposeDepth)})
		return
	}
	for _, e := range ents {
		if u, ok := e.(entity.Unwrapper); ok {
			e = u.Unwrap()
		}
		if ins, ok := e.(*entity.Insert); ok {
			name := ins.BlockName()
			if i := pathIndex(stack, name); i >= 0 {
				d.Cycles = append(d.Cycles, append(append([]string{}, stack[i:]...), name))
				continue
			}
			var blk *block.Block
			if lookup != nil {
				blk = lookup(name)
			}
			if blk == nil {
				p := MakePrimitive(e)
				p.Diagnostics = append(p.Diagnostics, PrimitiveDiagnostic{Handle: ins.Handle(), Type: "INSERT", Message: "unresolved block: " + name})
				d.Primitives = append(d.Primitives, p)
				continue
			}
			nm := m.Mul(ins.InsertMatrix())
			d.walk(blockEntities(blk), nm, append(stack, name), lookup, depth+1)
			continue
		}
		if h, ok := e.(*entity.Hatch); ok {
			parts, _ := h.Explode()
			if len(parts) > 0 {
				d.walk(parts, m, stack, lookup, depth+1)
				continue
			}
		}
		if pl, ok := e.(*entity.Polyline); ok && !pl.IsPolyface() && !pl.IsPolymesh() {
			// Polylines stay a single path primitive.
			p := MakePrimitive(e)
			p = transformPrimitive(p, m)
			d.Primitives = append(d.Primitives, p)
			continue
		}
		cp := e
		if c, ok := e.(entity.Copier); ok {
			if c2 := c.Copy(); c2 != nil {
				cp = c2
			}
		}
		if !m.IsIdentity(1e-12) {
			if t, ok := cp.(Transformer); ok {
				if err := t.Transform(m); err != nil {
					d.Diagnostics = append(d.Diagnostics, PrimitiveDiagnostic{Handle: e.Handle(), Type: e.DXFType(), Message: err.Error()})
				}
			} else {
				d.Diagnostics = append(d.Diagnostics, PrimitiveDiagnostic{Handle: e.Handle(), Type: e.DXFType(), Message: "entity does not support transformation"})
			}
		}
		d.Primitives = append(d.Primitives, MakePrimitive(cp))
	}
}

func transformPrimitive(p Primitive, m dxfmath.Matrix44) Primitive {
	if m.IsIdentity(1e-12) {
		return p
	}
	for i, v := range p.Vertices {
		xyz := entity.TransformPoint(m, []float64{v.X(), v.Y(), v.Z()})
		p.Vertices[i] = vec3(xyz)
	}
	for i, v := range p.Control {
		xyz := entity.TransformPoint(m, []float64{v.X(), v.Y(), v.Z()})
		p.Control[i] = vec3(xyz)
	}
	if p.Mesh != nil {
		for i, v := range p.Mesh.Vertices {
			xyz := entity.TransformPoint(m, []float64{v.X(), v.Y(), v.Z()})
			p.Mesh.Vertices[i] = vec3(xyz)
		}
	}
	if p.Path != nil {
		p.Path = p.Path.Transform(&m)
	}
	return p
}

// ToVertices flattens primitive vertices in stream order.
func ToVertices(primitives []Primitive) []dxfmath.Vec3 {
	var out []dxfmath.Vec3
	for _, p := range primitives {
		out = append(out, p.Vertices...)
	}
	return out
}

// ToPaths returns every non-nil primitive path.
func ToPaths(primitives []Primitive) []*path.Path {
	var out []*path.Path
	for _, p := range primitives {
		if p.Path != nil {
			out = append(out, p.Path)
		}
	}
	return out
}

// ToMeshes returns every primitive mesh.
func ToMeshes(primitives []Primitive) []*PrimitiveMesh {
	var out []*PrimitiveMesh
	for _, p := range primitives {
		if p.Mesh != nil {
			out = append(out, p.Mesh)
		}
	}
	return out
}

// ToControlVertices flattens spline control points.
func ToControlVertices(primitives []Primitive) []dxfmath.Vec3 {
	var out []dxfmath.Vec3
	for _, p := range primitives {
		out = append(out, p.Control...)
	}
	return out
}

func vec3(v []float64) dxfmath.Vec3 {
	if len(v) < 3 {
		var x, y float64
		if len(v) > 0 {
			x = v[0]
		}
		if len(v) > 1 {
			y = v[1]
		}
		return dxfmath.NewVec3(x, y, 0)
	}
	return dxfmath.NewVec3(v[0], v[1], v[2])
}

func vecs(points ...[]float64) []dxfmath.Vec3 {
	out := make([]dxfmath.Vec3, 0, len(points))
	for _, p := range points {
		out = append(out, vec3(p))
	}
	return out
}

func linePath(a, b []float64) *path.Path {
	if len(a) < 2 || len(b) < 2 {
		return nil
	}
	p := path.NewPath()
	p.MoveTo(a[0], a[1])
	p.LineTo(b[0], b[1])
	return p
}

func polylinePath(pts []dxfmath.Vec3, closed bool) *path.Path {
	if len(pts) == 0 {
		return nil
	}
	p := path.NewPath()
	p.MoveTo(pts[0].X(), pts[0].Y())
	for _, v := range pts[1:] {
		p.LineTo(v.X(), v.Y())
	}
	if closed && len(pts) > 2 {
		p.Close()
	}
	return p
}

func pathVecs(p *path.Path) []dxfmath.Vec3 {
	if p == nil {
		return nil
	}
	var out []dxfmath.Vec3
	for _, v := range p.Vertices() {
		out = append(out, dxfmath.NewVec3(v.X(), v.Y(), 0))
	}
	return out
}

func appendPath(dst, src *path.Path) {
	if src == nil {
		return
	}
	for _, cmd := range src.Commands() {
		switch c := cmd.(type) {
		case *path.PathMoveTo:
			dst.MoveTo(c.End.X(), c.End.Y())
		case *path.PathLineTo:
			dst.LineTo(c.End.X(), c.End.Y())
		case *path.PathCurve3To:
			dst.Curve3To(c.End.X(), c.End.Y(), c.Ctrl.X(), c.Ctrl.Y())
		case *path.PathCurve4To:
			dst.Curve4To(c.End.X(), c.End.Y(), c.Ctrl1.X(), c.Ctrl1.Y(), c.Ctrl2.X(), c.Ctrl2.Y())
		}
	}
}

func arcPath(center []float64, radius, startDeg, endDeg float64) *path.Path {
	if len(center) < 2 || radius <= 0 {
		return nil
	}
	if endDeg <= startDeg {
		endDeg += 360
	}
	span := (endDeg - startDeg) * math.Pi / 180
	segments := int(math.Ceil(math.Abs(span) / (math.Pi / 2)))
	if segments < 1 {
		segments = 1
	}
	p := path.NewPath()
	step := span / float64(segments)
	start := startDeg * math.Pi / 180
	p.MoveTo(center[0]+radius*math.Cos(start), center[1]+radius*math.Sin(start))
	for i := 0; i < segments; i++ {
		a0 := start + float64(i)*step
		a1 := a0 + step
		k := 4.0 / 3.0 * math.Tan(step/4)
		x0, y0 := center[0]+radius*math.Cos(a0), center[1]+radius*math.Sin(a0)
		x1, y1 := center[0]+radius*math.Cos(a1), center[1]+radius*math.Sin(a1)
		c1x, c1y := x0-radius*math.Sin(a0)*k, y0+radius*math.Cos(a0)*k
		c2x, c2y := x1+radius*math.Sin(a1)*k, y1-radius*math.Cos(a1)*k
		p.Curve4To(x1, y1, c1x, c1y, c2x, c2y)
	}
	return p
}

func ellipseVertices(e *entity.Ellipse) []dxfmath.Vec3 {
	if len(e.Center) < 2 {
		return nil
	}
	n := 64
	if !e.IsFull() {
		n = 32
	}
	out := make([]dxfmath.Vec3, 0, n+1)
	for i := 0; i <= n; i++ {
		t := float64(i) / float64(n)
		if e.IsFull() && i == n {
			break
		}
		out = append(out, e.PointAt(t))
	}
	return out
}

func textVertices(text string, height float64, origin []float64, rotation float64) []dxfmath.Vec3 {
	if len(origin) < 2 {
		return nil
	}
	width := float64(len(text)) * height * 0.6
	corners := [][2]float64{{0, 0}, {width, 0}, {width, height}, {0, height}}
	rad := rotation * math.Pi / 180
	out := make([]dxfmath.Vec3, 0, 4)
	for _, c := range corners {
		x := origin[0] + c[0]*math.Cos(rad) - c[1]*math.Sin(rad)
		y := origin[1] + c[0]*math.Sin(rad) + c[1]*math.Cos(rad)
		out = append(out, dxfmath.NewVec3(x, y, 0))
	}
	return out
}

func textPath(text string, height, x, y, rotation float64) *path.Path {
	if text == "" || height <= 0 {
		return nil
	}
	paths, _, err := text2path.TextToPathsTTF(text, height)
	if err == nil && len(paths) > 0 && paths[0] != nil {
		m := dxfmath.Translation(x, y, 0).Mul(dxfmath.RotationZ(rotation * math.Pi / 180))
		return paths[0].Transform(&m)
	}
	ttp := text2path.NewTextToPath(height / 120.0)
	p := ttp.Convert(text)
	if p == nil {
		return nil
	}
	m := dxfmath.Translation(x, y, 0).Mul(dxfmath.RotationZ(rotation * math.Pi / 180))
	return p.Transform(&m)
}

func polylineMesh(p *entity.Polyline) *PrimitiveMesh {
	mesh := &PrimitiveMesh{}
	index := make(map[*entity.Vertex]int)
	for _, v := range p.Vertices {
		if v == nil || v.IsFaceRecord() || len(v.Coord) < 3 {
			continue
		}
		index[v] = len(mesh.Vertices)
		mesh.Vertices = append(mesh.Vertices, vec3(v.Coord))
	}
	for _, v := range p.Vertices {
		if v == nil || !v.IsFaceRecord() {
			continue
		}
		var face []int
		for _, idx := range v.FaceVertexIndices() {
			if idx < 0 {
				idx = -idx
			}
			if idx > 0 {
				face = append(face, idx-1)
			}
		}
		if len(face) > 0 {
			mesh.Faces = append(mesh.Faces, face)
		}
	}
	return mesh
}
