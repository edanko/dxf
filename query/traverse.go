// Block reference traversal with cycle detection, revision clouds.
package query

import (
	"math"

	"github.com/edanko/dxf/block"
	"github.com/edanko/dxf/entity"
)

// TraverseBlockRefs walks INSERT references depth-first, including nested
// inserts. visit is called for every insert in traversal order with its
// nesting depth and the block path from the root; returning false prunes the
// subtree. Cyclic references terminate and are returned as cycles (the
// repeated path, closed with the repeated name).
func TraverseBlockRefs(ents []entity.Entity, lookup BlockLookup, visit func(depth int, ins *entity.Insert, path []string) bool) [][]string {
	var cycles [][]string
	var walk func(list []entity.Entity, depth int, path []string)
	walk = func(list []entity.Entity, depth int, path []string) {
		for _, e := range list {
			if u, ok := e.(interface{ Unwrap() entity.Entity }); ok {
				e = u.Unwrap()
			}
			ins, ok := e.(*entity.Insert)
			if !ok {
				continue
			}
			name := ins.BlockName()
			if i := pathIndex(path, name); i >= 0 {
				cycles = append(cycles, append(append([]string{}, path[i:]...), name))
				continue
			}
			if visit != nil && !visit(depth, ins, path) {
				continue
			}
			var blk *block.Block
			if lookup != nil {
				blk = lookup(name)
			}
			if blk == nil {
				continue
			}
			walk(blockEntities(blk), depth+1, append(path, name))
		}
	}
	walk(ents, 0, nil)
	return cycles
}

// RevCloudArc is one arc segment of a revision cloud: a chord of fixed length
// bulging outward by the configured bulge factor.
type RevCloudArc struct {
	Start [2]float64
	End   [2]float64
	Bulge float64
}

// DefaultRevCloudArcLength is the standard revision cloud arc chord length.
const DefaultRevCloudArcLength = 1.0

// DefaultRevCloudBulge produces semicircular arcs (bulge = tan(theta/4) with
// theta = 180 degrees).
const DefaultRevCloudBulge = 1.0

// RevCloudRectangle builds a closed LWPOLYLINE revision cloud around the
// rectangle (minX,minY)-(maxX,maxY): each side is divided into arcs of
// approximately arcLength (DefaultRevCloudArcLength when <= 0) with the
// given bulge (DefaultRevCloudBulge when 0).
func RevCloudRectangle(minX, minY, maxX, maxY, arcLength, bulge float64) (*entity.LwPolyline, error) {
	if maxX <= minX || maxY <= minY {
		return nil, errRevCloudBounds
	}
	if arcLength <= 0 {
		arcLength = DefaultRevCloudArcLength
	}
	if bulge == 0 {
		bulge = DefaultRevCloudBulge
	}
	width, height := maxX-minX, maxY-minY
	nx := arcsForSide(width, arcLength)
	ny := arcsForSide(height, arcLength)
	// Corners clockwise from bottom-left: BL -> BR -> TR -> TL -> BL, with
	// segment counts per side; vertices are the division points.
	var verts [][]float64
	addSide := func(x0, y0, x1, y1 float64, n int) {
		stepX, stepY := (x1-x0)/float64(n), (y1-y0)/float64(n)
		for i := 0; i < n; i++ {
			verts = append(verts, []float64{x0 + stepX*float64(i), y0 + stepY*float64(i)})
		}
	}
	addSide(minX, minY, maxX, minY, nx) // bottom, left to right
	addSide(maxX, minY, maxX, maxY, ny) // right, bottom to top
	addSide(maxX, maxY, minX, maxY, nx) // top, right to left
	addSide(minX, maxY, minX, minY, ny) // left, top to bottom

	p := entity.NewLwPolyline(len(verts))
	p.Closed = true
	bulges := make([]float64, len(verts))
	for i, v := range verts {
		p.Vertices[i] = v
		bulges[i] = bulge
	}
	p.Bulges = bulges
	return p, nil
}

// RevCloudArcs returns the individual arcs of a rectangle revision cloud.
func RevCloudArcs(minX, minY, maxX, maxY, arcLength, bulge float64) ([]RevCloudArc, error) {
	p, err := RevCloudRectangle(minX, minY, maxX, maxY, arcLength, bulge)
	if err != nil {
		return nil, err
	}
	n := p.Num
	arcs := make([]RevCloudArc, 0, n)
	for i := 0; i < n; i++ {
		a := p.Vertices[i]
		b := p.Vertices[(i+1)%n]
		arcs = append(arcs, RevCloudArc{
			Start: [2]float64{a[0], a[1]},
			End:   [2]float64{b[0], b[1]},
			Bulge: p.Bulges[i],
		})
	}
	return arcs, nil
}

func arcsForSide(length, arcLength float64) int {
	n := int(math.Round(length / arcLength))
	if n < 1 {
		n = 1
	}
	return n
}

type boundsError string

func (e boundsError) Error() string { return string(e) }

const errRevCloudBounds = boundsError("revcloud: rectangle is empty (max must exceed min)")
