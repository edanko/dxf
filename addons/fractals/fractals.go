// Package fractals provides pure-Go geometry generators for the Menger sponge
// and the Sierpinski pyramid.
//
// Both generators expose their subdivision count, per-cell meshes, a merged
// mesh and helpers which add the generated geometry to a drawing. All output
// uses ordinary MESH entities, so layer, handle and ownership behavior is the
// same as for any other drawing entity.
package fractals

import (
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/edanko/dxf/addons"
	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
)

// MengerKind selects one of the supported Menger sponge subdivision rules.
type MengerKind int

const (
	// MengerOriginal is the classic Menger sponge.
	MengerOriginal MengerKind = iota
	// MengerVariantXOX is the XOX variant.
	MengerVariantXOX
	// MengerVariantOXO is the OXO variant.
	MengerVariantOXO
	// JerusalemCube is the Jerusalem cube variant.
	JerusalemCube
)

var mengerSchemas = [][][3]int{
	{
		{0, 0, 0}, {1, 0, 0}, {2, 0, 0}, {0, 1, 0}, {2, 1, 0}, {0, 2, 0}, {1, 2, 0}, {2, 2, 0},
		{0, 0, 1}, {2, 0, 1}, {0, 2, 1}, {2, 2, 1},
		{0, 0, 2}, {1, 0, 2}, {2, 0, 2}, {0, 1, 2}, {2, 1, 2}, {0, 2, 2}, {1, 2, 2}, {2, 2, 2},
	},
	{
		{0, 0, 0}, {2, 0, 0}, {1, 1, 0}, {0, 2, 0}, {2, 2, 0},
		{1, 0, 1}, {0, 1, 1}, {2, 1, 1}, {1, 2, 1},
		{0, 0, 2}, {2, 0, 2}, {1, 1, 2}, {0, 2, 2}, {2, 2, 2},
	},
	{
		{1, 0, 0}, {0, 1, 0}, {2, 1, 0}, {1, 2, 0},
		{0, 0, 1}, {2, 0, 1}, {1, 1, 1}, {0, 2, 1}, {2, 2, 1},
		{1, 0, 2}, {0, 1, 2}, {2, 1, 2}, {1, 2, 2},
	},
	{
		{0, 0, 0}, {1, 0, 0}, {2, 0, 0}, {3, 0, 0}, {4, 0, 0}, {0, 1, 0}, {1, 1, 0}, {3, 1, 0}, {4, 1, 0}, {0, 2, 0},
		{4, 2, 0}, {0, 3, 0}, {1, 3, 0}, {3, 3, 0}, {4, 3, 0}, {0, 4, 0}, {1, 4, 0}, {2, 4, 0}, {3, 4, 0}, {4, 4, 0},
		{0, 0, 1}, {1, 0, 1}, {3, 0, 1}, {4, 0, 1}, {0, 1, 1}, {1, 1, 1}, {3, 1, 1}, {4, 1, 1}, {0, 3, 1}, {1, 3, 1},
		{3, 3, 1}, {4, 3, 1}, {0, 4, 1}, {1, 4, 1}, {3, 4, 1}, {4, 4, 1}, {0, 0, 2}, {4, 0, 2}, {0, 4, 2}, {4, 4, 2},
		{0, 0, 3}, {1, 0, 3}, {3, 0, 3}, {4, 0, 3}, {0, 1, 3}, {1, 1, 3}, {3, 1, 3}, {4, 1, 3}, {0, 3, 3}, {1, 3, 3},
		{3, 3, 3}, {4, 3, 3}, {0, 4, 3}, {1, 4, 3}, {3, 4, 3}, {4, 4, 3}, {0, 0, 4}, {1, 0, 4}, {2, 0, 4}, {3, 0, 4},
		{4, 0, 4}, {0, 1, 4}, {1, 1, 4}, {3, 1, 4}, {4, 1, 4}, {0, 2, 4}, {4, 2, 4}, {0, 3, 4}, {1, 3, 4}, {3, 3, 4},
		{4, 3, 4}, {0, 4, 4}, {1, 4, 4}, {2, 4, 4}, {3, 4, 4}, {4, 4, 4},
	},
}

var mengerCubeSizes = [4]float64{3, 3, 3, 5}

var cubeVertices = [8]dxfmath.Vec3{
	{0, 0, 0}, {1, 0, 0}, {1, 1, 0}, {0, 1, 0},
	{0, 0, 1}, {1, 0, 1}, {1, 1, 1}, {0, 1, 1},
}

var cubeFaces = [][]int{
	{0, 3, 2, 1},
	{4, 5, 6, 7},
	{0, 1, 5, 4},
	{1, 2, 6, 5},
	{3, 7, 6, 2},
	{0, 4, 7, 3},
}

// Cube describes one generated cube.
type Cube struct {
	Location dxfmath.Vec3
	Length   float64
}

// MengerSponge generates a Menger sponge or one of its variants.
type MengerSponge struct {
	Location dxfmath.Vec3
	Length   float64
	Level    int
	Kind     MengerKind
}

// NewMengerSponge validates the parameters and creates a generator.
func NewMengerSponge(location dxfmath.Vec3, length float64, level int, kind MengerKind) (*MengerSponge, error) {
	if level < 1 {
		return nil, errors.New("fractals: Menger sponge level must be at least 1")
	}
	if length <= 0 {
		return nil, errors.New("fractals: Menger sponge length must be positive")
	}
	if kind < MengerOriginal || kind > JerusalemCube {
		return nil, fmt.Errorf("fractals: Menger sponge kind must be 0, 1, 2 or 3, got %d", kind)
	}
	return &MengerSponge{Location: location, Length: length, Level: level, Kind: kind}, nil
}

// CubeCount returns the number of cubes produced at the configured level.
func (m *MengerSponge) CubeCount() int {
	return intPow(len(mengerSchemas[m.Kind]), m.Level)
}

// FaceCountPerCube returns the number of faces on each generated cube.
func (m *MengerSponge) FaceCountPerCube() int { return len(cubeFaces) }

// TotalCubeFaceCount returns CubeCount multiplied by FaceCountPerCube.
func (m *MengerSponge) TotalCubeFaceCount() int {
	return m.CubeCount() * m.FaceCountPerCube()
}

// Cubes returns every generated cube definition.
func (m *MengerSponge) Cubes() []Cube {
	cubes := []Cube{{Location: m.Location, Length: m.Length}}
	for level := 0; level < m.Level; level++ {
		next := make([]Cube, 0, len(cubes)*len(mengerSchemas[m.Kind]))
		for _, cube := range cubes {
			next = append(next, subdivideCube(cube, m.Kind)...)
		}
		cubes = next
	}
	return cubes
}

func subdivideCube(cube Cube, kind MengerKind) []Cube {
	schema := mengerSchemas[kind]
	step := cube.Length / mengerCubeSizes[kind]
	out := make([]Cube, len(schema))
	for i, index := range schema {
		out[i] = Cube{
			Location: dxfmath.Vec3{
				cube.Location[0] + float64(index[0])*step,
				cube.Location[1] + float64(index[1])*step,
				cube.Location[2] + float64(index[2])*step,
			},
			Length: step,
		}
	}
	return out
}

// CubeVertices returns the eight vertices of a cube in the canonical order.
func CubeVertices(cube Cube) []dxfmath.Vec3 {
	vertices := make([]dxfmath.Vec3, len(cubeVertices))
	for i, vertex := range cubeVertices {
		vertices[i] = dxfmath.Vec3{
			cube.Location[0] + vertex[0]*cube.Length,
			cube.Location[1] + vertex[1]*cube.Length,
			cube.Location[2] + vertex[2]*cube.Length,
		}
	}
	return vertices
}

// CubeFaces returns a copy of the canonical cube face template.
func CubeFaces() [][]int { return cloneFaces(cubeFaces) }

// Faces returns the canonical cube face template.
func (m *MengerSponge) Faces() [][]int { return CubeFaces() }

// Meshes returns one mesh per generated cube.
func (m *MengerSponge) Meshes() []*entity.Mesh {
	cubes := m.Cubes()
	meshes := make([]*entity.Mesh, len(cubes))
	for i, cube := range cubes {
		mesh := entity.NewMesh()
		for _, vertex := range CubeVertices(cube) {
			mesh.AddVertex(vertex)
		}
		for _, face := range cubeFaces {
			mesh.AddFace(append([]int(nil), face...))
		}
		meshes[i] = mesh
	}
	return meshes
}

// Mesh returns the whole sponge as one MESH entity. Interior faces shared by
// adjacent cubes are removed so the result is manifold.
func (m *MengerSponge) Mesh() *entity.Mesh {
	builder := newMeshBuilder()
	for _, cube := range m.Cubes() {
		builder.add(CubeVertices(cube), cubeFaces)
	}
	return builder.mesh()
}

// AddTo adds the merged sponge mesh to the drawing.
func (m *MengerSponge) AddTo(d *drawing.Drawing) error {
	if d == nil {
		return errors.New("fractals: drawing is nil")
	}
	return d.AddEntity(m.Mesh())
}

// AddCubesTo adds one MESH entity per generated cube to the drawing.
func (m *MengerSponge) AddCubesTo(d *drawing.Drawing) error {
	if d == nil {
		return errors.New("fractals: drawing is nil")
	}
	for _, mesh := range m.Meshes() {
		if err := d.AddEntity(mesh); err != nil {
			return err
		}
	}
	return nil
}

// SierpinskiPyramid generates a Sierpinski pyramid with triangular or square
// base geometry.
type SierpinskiPyramid struct {
	Location dxfmath.Vec3
	Length   float64
	Level    int
	Sides    int
}

const (
	pyramidHeight4 = 1.0 / 1.4142135623730951
	pyramidHeight3 = 1.632993161855452
	dy1Factor      = 0.28867513459481287
	dy2Factor      = 0.5773502691896258
)

// NewSierpinskiPyramid validates the parameters and creates a generator.
func NewSierpinskiPyramid(location dxfmath.Vec3, length float64, level, sides int) (*SierpinskiPyramid, error) {
	if level < 1 {
		return nil, errors.New("fractals: Sierpinski pyramid level must be at least 1")
	}
	if length <= 0 {
		return nil, errors.New("fractals: Sierpinski pyramid length must be positive")
	}
	if sides != 3 && sides != 4 {
		return nil, fmt.Errorf("fractals: Sierpinski pyramid sides must be 3 or 4, got %d", sides)
	}
	return &SierpinskiPyramid{Location: location, Length: length, Level: level, Sides: sides}, nil
}

// PyramidCount returns the number of generated pyramids.
func (p *SierpinskiPyramid) PyramidCount() int {
	if p.Sides == 3 {
		return intPow(4, p.Level)
	}
	return intPow(5, p.Level)
}

// FaceCountPerPyramid returns the number of faces per generated pyramid.
func (p *SierpinskiPyramid) FaceCountPerPyramid() int {
	if p.Sides == 3 {
		return 4
	}
	return 5
}

// TotalFaceCount returns PyramidCount multiplied by FaceCountPerPyramid.
func (p *SierpinskiPyramid) TotalFaceCount() int {
	return p.PyramidCount() * p.FaceCountPerPyramid()
}

// Pyramid describes one generated pyramid.
type Pyramid struct {
	Location dxfmath.Vec3
	Length   float64
}

// Pyramids returns every generated pyramid definition.
func (p *SierpinskiPyramid) Pyramids() []Pyramid {
	pyramids := []Pyramid{{Location: p.Location, Length: p.Length}}
	for level := 0; level < p.Level; level++ {
		next := make([]Pyramid, 0, len(pyramids)*4)
		for _, pyramid := range pyramids {
			next = append(next, subdividePyramid(pyramid, p.Sides)...)
		}
		pyramids = next
	}
	return pyramids
}

func subdividePyramid(pyramid Pyramid, sides int) []Pyramid {
	length := pyramid.Length
	len2 := length / 2
	len4 := length / 4
	x, y, z := pyramid.Location[0], pyramid.Location[1], pyramid.Location[2]
	if sides == 3 {
		dy1 := length * dy1Factor * 0.5
		dy2 := length * dy2Factor * 0.5
		return []Pyramid{
			{Location: dxfmath.Vec3{x - len4, y - dy1, z}, Length: len2},
			{Location: dxfmath.Vec3{x + len4, y - dy1, z}, Length: len2},
			{Location: dxfmath.Vec3{x, y + dy2, z}, Length: len2},
			{Location: dxfmath.Vec3{x, y, z + len2*pyramidHeight3}, Length: len2},
		}
	}
	return []Pyramid{
		{Location: dxfmath.Vec3{x - len4, y - len4, z}, Length: len2},
		{Location: dxfmath.Vec3{x + len4, y - len4, z}, Length: len2},
		{Location: dxfmath.Vec3{x - len4, y + len4, z}, Length: len2},
		{Location: dxfmath.Vec3{x + len4, y + len4, z}, Length: len2},
		{Location: dxfmath.Vec3{x, y, z + len2*pyramidHeight4}, Length: len2},
	}
}

// PyramidVertices returns the vertices of one pyramid.
func (p *SierpinskiPyramid) PyramidVertices(pyramid Pyramid) []dxfmath.Vec3 {
	length := pyramid.Length
	len2 := length / 2
	x, y, z := pyramid.Location[0], pyramid.Location[1], pyramid.Location[2]
	if p.Sides == 3 {
		dy1 := length * dy1Factor
		dy2 := length * dy2Factor
		return []dxfmath.Vec3{
			{x - len2, y - dy1, z},
			{x + len2, y - dy1, z},
			{x, y + dy2, z},
			{x, y, z + length*pyramidHeight3},
		}
	}
	return []dxfmath.Vec3{
		{x - len2, y - len2, z},
		{x + len2, y - len2, z},
		{x + len2, y + len2, z},
		{x - len2, y + len2, z},
		{x, y, z + length*pyramidHeight4},
	}
}

// Faces returns the canonical pyramid face template.
func (p *SierpinskiPyramid) Faces() [][]int {
	if p.Sides == 3 {
		return [][]int{{2, 1, 0}, {0, 1, 3}, {1, 2, 3}, {2, 0, 3}}
	}
	return [][]int{{3, 2, 1, 0}, {0, 1, 4}, {1, 2, 4}, {2, 3, 4}, {3, 0, 4}}
}

// Meshes returns one mesh per generated pyramid.
func (p *SierpinskiPyramid) Meshes() []*entity.Mesh {
	pyramids := p.Pyramids()
	faces := p.Faces()
	meshes := make([]*entity.Mesh, len(pyramids))
	for i, pyramid := range pyramids {
		mesh := entity.NewMesh()
		for _, vertex := range p.PyramidVertices(pyramid) {
			mesh.AddVertex(vertex)
		}
		for _, face := range faces {
			mesh.AddFace(append([]int(nil), face...))
		}
		meshes[i] = mesh
	}
	return meshes
}

// Mesh returns the whole pyramid as one MESH entity.
func (p *SierpinskiPyramid) Mesh() *entity.Mesh {
	builder := newMeshBuilder()
	faces := p.Faces()
	for _, pyramid := range p.Pyramids() {
		builder.add(p.PyramidVertices(pyramid), faces)
	}
	return builder.mesh()
}

// AddTo adds the merged pyramid mesh to the drawing.
func (p *SierpinskiPyramid) AddTo(d *drawing.Drawing) error {
	if d == nil {
		return errors.New("fractals: drawing is nil")
	}
	return d.AddEntity(p.Mesh())
}

// AddPyramidsTo adds one MESH entity per generated pyramid to the drawing.
func (p *SierpinskiPyramid) AddPyramidsTo(d *drawing.Drawing) error {
	if d == nil {
		return errors.New("fractals: drawing is nil")
	}
	for _, mesh := range p.Meshes() {
		if err := d.AddEntity(mesh); err != nil {
			return err
		}
	}
	return nil
}

func intPow(base, exponent int) int {
	result := 1
	for i := 0; i < exponent; i++ {
		result *= base
	}
	return result
}

func cloneFaces(faces [][]int) [][]int {
	out := make([][]int, len(faces))
	for i, face := range faces {
		out[i] = append([]int(nil), face...)
	}
	return out
}

type meshBuilder struct {
	vertices []dxfmath.Vec3
	index    map[dxfmath.Vec3]int
	faces    [][]int
}

func newMeshBuilder() *meshBuilder {
	return &meshBuilder{index: make(map[dxfmath.Vec3]int)}
}

func (b *meshBuilder) add(vertices []dxfmath.Vec3, faces [][]int) {
	indices := make([]int, len(vertices))
	for i, vertex := range vertices {
		index, ok := b.index[vertex]
		if !ok {
			index = len(b.vertices)
			b.index[vertex] = index
			b.vertices = append(b.vertices, vertex)
		}
		indices[i] = index
	}
	for _, face := range faces {
		mapped := make([]int, len(face))
		for i, index := range face {
			mapped[i] = indices[index]
		}
		b.faces = append(b.faces, mapped)
	}
}

func (b *meshBuilder) mesh() *entity.Mesh {
	mesh := entity.NewMesh()
	for _, vertex := range b.vertices {
		mesh.AddVertex(vertex)
	}
	for _, face := range manifoldFaces(b.faces) {
		mesh.AddFace(face)
	}
	return mesh
}

// manifoldFaces keeps one copy of every face. Face vertex order is ignored, so
// opposite sides of an interior face collapse to the same key.
func manifoldFaces(faces [][]int) [][]int {
	seen := make(map[string]int, len(faces))
	first := make([][]int, 0, len(faces))
	for _, face := range faces {
		sorted := append([]int(nil), face...)
		sort.Ints(sorted)
		key := faceKey(sorted)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = len(first)
		first = append(first, face)
	}
	return first
}

func faceKey(indices []int) string {
	buf := make([]byte, 0, len(indices)*8)
	for _, index := range indices {
		buf = strconv.AppendInt(buf, int64(index), 10)
		buf = append(buf, ',')
	}
	return string(buf)
}

// Capability registers the add-on with the addons registry.
type Capability struct{}

// Name implements addons.Capability.
func (Capability) Name() string { return "fractals" }

// Description implements addons.Capability.
func (Capability) Description() string { return "Menger sponge and Sierpinski pyramid generators" }

func init() {
	_ = addons.Register(Capability{}.Name(), Capability{})
}
