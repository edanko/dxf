// Package edges builds topology from unordered line and curve segments.
//
// Edges are immutable values. A Graph merges endpoints that lie within a
// tolerance using a spatial hash, so source data split at coincident points
// becomes one connected network. Analyze then extracts closed loops and open
// chains while preserving the direction of each source edge.
package edges

import (
	"errors"
	"math"
	"sort"

	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
)

// ErrOpenChain reports that an open chain cannot become a hatch boundary.
var ErrOpenChain = errors.New("edges: open chain cannot form a hatch boundary")

// Edge is one undirected segment between two vertices. Constructing a new
// value or calling Reversed never changes the original.
type Edge struct {
	Start dxfmath.Vec2
	End   dxfmath.Vec2
}

// NewEdge creates an edge from start to end.
func NewEdge(start, end dxfmath.Vec2) Edge { return Edge{Start: start, End: end} }

// Reversed returns the edge with its endpoints exchanged.
func (e Edge) Reversed() Edge { return Edge{Start: e.End, End: e.Start} }

// Length returns the straight-line length of the edge.
func (e Edge) Length() float64 { return e.Start.Distance(e.End) }

type adjacency struct {
	to   int
	edge int
}

// Graph is a topology graph built from edges. Vertex and edge order follows
// insertion order.
type Graph struct {
	tolerance float64
	vertices  []dxfmath.Vec2
	edges     []Edge
	adj       [][]adjacency
	buckets   map[gridKey][]int
}

type gridKey struct{ x, y int64 }

// NewGraph creates an empty graph. A non-positive tolerance uses 1e-9.
func NewGraph(tolerance float64) *Graph {
	if tolerance <= 0 {
		tolerance = 1e-9
	}
	return &Graph{tolerance: tolerance, buckets: map[gridKey][]int{}}
}

// BuildGraph creates a graph and adds all edges.
func BuildGraph(edges []Edge, tolerance float64) *Graph {
	g := NewGraph(tolerance)
	g.Add(edges...)
	return g
}

// Add inserts edges and merges endpoints within the graph tolerance. It
// returns the vertex index assigned to each edge endpoint.
func (g *Graph) Add(edges ...Edge) {
	for _, e := range edges {
		a := g.vertex(e.Start)
		b := g.vertex(e.End)
		id := len(g.edges)
		g.edges = append(g.edges, e)
		g.adj[a] = append(g.adj[a], adjacency{to: b, edge: id})
		g.adj[b] = append(g.adj[b], adjacency{to: a, edge: id})
	}
}

// Tolerance returns the endpoint merge tolerance.
func (g *Graph) Tolerance() float64 { return g.tolerance }

// VertexCount returns the number of distinct vertices.
func (g *Graph) VertexCount() int { return len(g.vertices) }

// EdgeCount returns the number of edges.
func (g *Graph) EdgeCount() int { return len(g.edges) }

// Vertices returns a copy of the graph vertices.
func (g *Graph) Vertices() []dxfmath.Vec2 {
	return append([]dxfmath.Vec2(nil), g.vertices...)
}

// Edges returns a copy of the graph edges.
func (g *Graph) Edges() []Edge { return append([]Edge(nil), g.edges...) }

// Vertex returns vertex i.
func (g *Graph) Vertex(i int) dxfmath.Vec2 { return g.vertices[i] }

// Degree returns the number of edge incidences at vertex i.
func (g *Graph) Degree(i int) int {
	if i < 0 || i >= len(g.adj) {
		return 0
	}
	return len(g.adj[i])
}

// Leaves returns vertex indices with degree one, in ascending order.
func (g *Graph) Leaves() []int {
	var out []int
	for i := range g.vertices {
		if g.Degree(i) == 1 {
			out = append(out, i)
		}
	}
	return out
}

// LeafVertices returns the coordinates of degree-one vertices.
func (g *Graph) LeafVertices() []dxfmath.Vec2 {
	out := make([]dxfmath.Vec2, 0)
	for _, i := range g.Leaves() {
		out = append(out, g.vertices[i])
	}
	return out
}

// Components returns connected vertex groups, each in ascending order.
func (g *Graph) Components() [][]int {
	seen := make([]bool, len(g.vertices))
	var out [][]int
	for start := range g.vertices {
		if seen[start] {
			continue
		}
		seen[start] = true
		queue := []int{start}
		var comp []int
		for len(queue) > 0 {
			v := queue[0]
			queue = queue[1:]
			comp = append(comp, v)
			for _, a := range g.adj[v] {
				if !seen[a.to] {
					seen[a.to] = true
					queue = append(queue, a.to)
				}
			}
		}
		sort.Ints(comp)
		out = append(out, comp)
	}
	return out
}

// Step is one edge traversal in a loop or chain.
type Step struct {
	Edge     int
	Reversed bool
}

// Loop is a closed cycle. Vertices has one entry per edge (the closing
// repeat of the first vertex is omitted); the cycle is closed implicitly.
type Loop struct {
	Vertices []dxfmath.Vec2
	Steps    []Step
}

// Closed reports whether the loop is closed. Loops returned by Analyze are
// always closed.
func (l Loop) Closed() bool { return true }

// Reverse returns the loop traversed in the opposite direction.
func (l Loop) Reverse() Loop {
	out := Loop{
		Vertices: make([]dxfmath.Vec2, len(l.Vertices)),
		Steps:    make([]Step, len(l.Steps)),
	}
	for i, v := range l.Vertices {
		out.Vertices[len(l.Vertices)-1-i] = v
	}
	for i, s := range l.Steps {
		out.Steps[len(l.Steps)-1-i] = Step{Edge: s.Edge, Reversed: !s.Reversed}
	}
	return out
}

// Chain is an open path. Vertices has one more entry than Steps.
type Chain struct {
	Vertices []dxfmath.Vec2
	Steps    []Step
	Leaves   [2]int
}

// Closed reports whether the chain is closed (always false).
func (c Chain) Closed() bool { return false }

// Reverse returns the chain traversed in the opposite direction.
func (c Chain) Reverse() Chain {
	out := Chain{
		Vertices: make([]dxfmath.Vec2, len(c.Vertices)),
		Steps:    make([]Step, len(c.Steps)),
		Leaves:   [2]int{c.Leaves[1], c.Leaves[0]},
	}
	for i, v := range c.Vertices {
		out.Vertices[len(c.Vertices)-1-i] = v
	}
	for i, s := range c.Steps {
		out.Steps[len(c.Steps)-1-i] = Step{Edge: s.Edge, Reversed: !s.Reversed}
	}
	return out
}

// Component groups one connected part of the graph.
type Component struct {
	Vertices []int
	Edges    []int
	Loops    []Loop
	Chains   []Chain
}

// Topology is the loop and chain decomposition of a graph.
type Topology struct {
	Components []Component
	Loops      []Loop
	Chains     []Chain
}

// Analyze extracts closed loops and open chains from every connected
// component. Closed cycles become Loops and leaf-to-leaf paths become Chains;
// branched components may contain several of each.
func (g *Graph) Analyze() Topology {
	used := make([]bool, len(g.edges))
	var topo Topology
	for _, ids := range g.Components() {
		comp := Component{Vertices: append([]int(nil), ids...)}
		comp.Edges = g.componentEdges(ids)
		// Closed cycles first, so a square with an attached tail remains a
		// loop plus a chain instead of one long path.
		for _, e := range comp.Edges {
			if used[e] {
				continue
			}
			if loop, consumed, ok := g.walkLoop(e, used); ok {
				for _, id := range consumed {
					used[id] = true
				}
				comp.Loops = append(comp.Loops, loop)
			}
		}
		// Remaining edges are open; walk from each degree-one endpoint.
		for _, v := range ids {
			if g.Degree(v) != 1 {
				continue
			}
			for _, a := range g.adj[v] {
				if used[a.edge] {
					continue
				}
				chain, consumed := g.walkChain(v, a.edge, used)
				for _, e := range consumed {
					used[e] = true
				}
				comp.Chains = append(comp.Chains, chain)
			}
		}
		// Any leftover branched edges still form paths.
		for _, e := range comp.Edges {
			if used[e] {
				continue
			}
			start := g.findVertex(g.edges[e].Start)
			chain, consumed := g.walkChain(start, e, used)
			for _, id := range consumed {
				used[id] = true
			}
			comp.Chains = append(comp.Chains, chain)
		}
		topo.Loops = append(topo.Loops, comp.Loops...)
		topo.Chains = append(topo.Chains, comp.Chains...)
		topo.Components = append(topo.Components, comp)
	}
	return topo
}

func (g *Graph) componentEdges(vertices []int) []int {
	in := make(map[int]bool, len(vertices))
	for _, v := range vertices {
		in[v] = true
	}
	var out []int
	for id, e := range g.edges {
		a := g.findVertex(e.Start)
		if in[a] {
			out = append(out, id)
		}
	}
	return out
}

func (g *Graph) walkChain(start, firstEdge int, used []bool) (Chain, []int) {
	edge := g.edges[firstEdge]
	cur := start
	next := g.other(edge, cur)
	chain := Chain{
		Vertices: []dxfmath.Vec2{g.vertices[cur], g.vertices[next]},
		Steps:    []Step{{Edge: firstEdge, Reversed: edge.End == g.vertices[cur]}},
		Leaves:   [2]int{start, -1},
	}
	consumed := []int{firstEdge}
	used[firstEdge] = true
	cur = next
	for {
		step, ok := g.nextUnused(cur, used)
		if !ok {
			chain.Leaves[1] = cur
			break
		}
		e := g.edges[step.edge]
		next = g.other(e, cur)
		chain.Steps = append(chain.Steps, Step{Edge: step.edge, Reversed: e.End == g.vertices[cur]})
		chain.Vertices = append(chain.Vertices, g.vertices[next])
		used[step.edge] = true
		consumed = append(consumed, step.edge)
		cur = next
	}
	return chain, consumed
}

func (g *Graph) walkLoop(firstEdge int, used []bool) (Loop, []int, bool) {
	trial := append([]bool(nil), used...)
	edge := g.edges[firstEdge]
	start := g.findVertex(edge.Start)
	next := g.findVertex(edge.End)
	loop := Loop{
		Vertices: []dxfmath.Vec2{g.vertices[start], g.vertices[next]},
		Steps:    []Step{{Edge: firstEdge}},
	}
	consumed := []int{firstEdge}
	trial[firstEdge] = true
	cur := next
	for cur != start {
		step, ok := g.nextUnused(cur, trial)
		if !ok {
			return Loop{}, nil, false
		}
		e := g.edges[step.edge]
		next = g.other(e, cur)
		loop.Steps = append(loop.Steps, Step{Edge: step.edge, Reversed: e.End == g.vertices[cur]})
		loop.Vertices = append(loop.Vertices, g.vertices[next])
		trial[step.edge] = true
		consumed = append(consumed, step.edge)
		cur = next
	}
	if len(loop.Vertices) < 2 {
		return Loop{}, nil, false
	}
	// Last vertex repeats the first; the Loop representation omits it.
	loop.Vertices = loop.Vertices[:len(loop.Vertices)-1]
	return loop, consumed, true
}

func (g *Graph) nextUnused(v int, used []bool) (adjacency, bool) {
	for _, a := range g.adj[v] {
		if !used[a.edge] {
			return a, true
		}
	}
	return adjacency{}, false
}

func (g *Graph) other(e Edge, v int) int {
	if g.findVertex(e.Start) == v {
		return g.findVertex(e.End)
	}
	return g.findVertex(e.Start)
}

func (g *Graph) findVertex(p dxfmath.Vec2) int {
	k := g.key(p)
	for dx := int64(-1); dx <= 1; dx++ {
		for dy := int64(-1); dy <= 1; dy++ {
			for _, i := range g.buckets[gridKey{k.x + dx, k.y + dy}] {
				if g.vertices[i].Distance(p) <= g.tolerance {
					return i
				}
			}
		}
	}
	return -1
}

func (g *Graph) vertex(p dxfmath.Vec2) int {
	if i := g.findVertex(p); i >= 0 {
		return i
	}
	i := len(g.vertices)
	g.vertices = append(g.vertices, p)
	g.adj = append(g.adj, nil)
	k := g.key(p)
	g.buckets[k] = append(g.buckets[k], i)
	return i
}

func (g *Graph) key(p dxfmath.Vec2) gridKey {
	return gridKey{
		x: int64(math.Floor(p.X() / g.tolerance)),
		y: int64(math.Floor(p.Y() / g.tolerance)),
	}
}

// Vertices2D converts loop vertices for entity creation.
func vertices2D(vs []dxfmath.Vec2) [][]float64 {
	out := make([][]float64, len(vs))
	for i, v := range vs {
		out[i] = []float64{v.X(), v.Y()}
	}
	return out
}

// ToLwPolyline creates an ordered LWPOLYLINE from the loop.
func (l Loop) ToLwPolyline() *entity.LwPolyline {
	lw := entity.NewLwPolyline(len(l.Vertices))
	lw.Vertices = vertices2D(l.Vertices)
	lw.Closed = true
	return lw
}

// ToPolyline creates an ordered POLYLINE from the loop.
func (l Loop) ToPolyline() *entity.Polyline {
	p := entity.NewPolyline()
	p.Flag = 0
	for _, v := range l.Vertices {
		p.AddVertex(v.X(), v.Y(), 0)
	}
	p.Close()
	return p
}

// ToLwPolyline creates an ordered LWPOLYLINE from the chain.
func (c Chain) ToLwPolyline() *entity.LwPolyline {
	lw := entity.NewLwPolyline(len(c.Vertices))
	lw.Vertices = vertices2D(c.Vertices)
	lw.Closed = false
	return lw
}

// ToPolyline creates an ordered POLYLINE from the chain.
func (c Chain) ToPolyline() *entity.Polyline {
	p := entity.NewPolyline()
	p.Flag = 0
	for _, v := range c.Vertices {
		p.AddVertex(v.X(), v.Y(), 0)
	}
	return p
}

// ToHatchBoundaryPath converts the closed loop to a hatch polyline path.
func (l Loop) ToHatchBoundaryPath() (*entity.PolylinePath, error) {
	p := entity.NewPolylinePath(1, true, false)
	for _, v := range l.Vertices {
		p.AddVertex(entity.NewVertex3D(v.X(), v.Y(), 0, 0))
	}
	return p, nil
}

// ToHatchBoundaryPath always fails for an open chain.
func (c Chain) ToHatchBoundaryPath() (*entity.PolylinePath, error) {
	return nil, ErrOpenChain
}

// ToHatchBoundaryPath returns the hatch boundary of the first loop.
func ToHatchBoundaryPath(loops []Loop) (*entity.PolylinePath, error) {
	if len(loops) == 0 {
		return nil, errors.New("edges: no closed loop")
	}
	return loops[0].ToHatchBoundaryPath()
}
