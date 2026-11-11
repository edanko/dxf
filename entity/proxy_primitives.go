// Proxy graphics primitive decoding: the opcode stream in an
// ACAD_PROXY_ENTITY graphics payload draws circles, arcs, polylines and text,
// with colour and transform state. Entity-record extraction remains available
// through ExtractProxyGraphics.
package entity

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"unicode/utf16"

	dxfmath "github.com/edanko/dxf/math"
)

// ProxyPrimitiveKind identifies a decoded proxy-graphics primitive.
type ProxyPrimitiveKind string

const (
	ProxyLine        ProxyPrimitiveKind = "LINE"
	ProxyArc         ProxyPrimitiveKind = "ARC"
	ProxyCircle      ProxyPrimitiveKind = "CIRCLE"
	ProxyPolyline    ProxyPrimitiveKind = "POLYLINE"
	ProxyText        ProxyPrimitiveKind = "TEXT"
	ProxyPoint       ProxyPrimitiveKind = "POINT"
	ProxyMesh        ProxyPrimitiveKind = "MESH"
	ProxyEllipticArc ProxyPrimitiveKind = "ELLIPSE"
)

// ProxyDiagnostic reports an unsupported or malformed proxy opcode. Offset is
// the absolute byte offset of the 8-byte chunk header inside the payload.
type ProxyDiagnostic struct {
	Offset  int
	Opcode  uint32
	Message string
}

func (d ProxyDiagnostic) String() string {
	return fmt.Sprintf("proxy opcode %d at byte %d: %s", d.Opcode, d.Offset, d.Message)
}

// ProxyPrimitive is one decoded drawing primitive. Geometry is already
// transformed by the active proxy matrix stack.
type ProxyPrimitive struct {
	Kind       ProxyPrimitiveKind
	Offset     int
	Vertices   []dxfmath.Vec3
	Center     dxfmath.Vec3
	Radius     float64
	StartAngle float64 // degrees
	EndAngle   float64 // degrees
	Closed     bool
	Text       string
	Height     float64
	WidthFact  float64
	Rotation   float64 // degrees
	Oblique    float64 // degrees
	Color      int
	HasColor   bool
	LayerIndex int
	HasLayer   bool
	Transform  dxfmath.Matrix44
}

// ProxyPrimitiveSet is the result of decoding a proxy graphics payload.
type ProxyPrimitiveSet struct {
	Primitives  []ProxyPrimitive
	Diagnostics []ProxyDiagnostic
}

// DecodeProxyPrimitives decodes the graphics payload (code 92/310, falling
// back to entity data) of an ACAD_PROXY_ENTITY. Unknown opcodes are skipped
// and reported with their byte offset; earlier primitives are retained.
func DecodeProxyPrimitives(e Entity) (*ProxyPrimitiveSet, error) {
	p, ok := e.(*AcadProxyEntity)
	if !ok {
		return nil, fmt.Errorf("proxy graphics: %s is not an ACAD_PROXY_ENTITY", e.DXFType())
	}
	payload := p.GraphicsData
	source := "graphics data"
	if len(payload) == 0 {
		payload = p.EntityData
		source = "entity data"
	}
	if len(payload) == 0 {
		return nil, fmt.Errorf("proxy graphics: proxy entity %s carries no graphics payload", e.Handle())
	}
	set, err := decodeProxyPrimitiveData(payload)
	if err != nil {
		return nil, fmt.Errorf("proxy graphics: %s of %s: %w", source, e.Handle(), err)
	}
	return set, nil
}

func decodeProxyPrimitiveData(data []byte) (*ProxyPrimitiveSet, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("payload shorter than proxy header")
	}
	out := &ProxyPrimitiveSet{}
	state := proxyState{
		color:     256,
		layer:     -1,
		transform: dxfmath.NewMatrix44(),
	}
	offset := 8
	for offset+8 <= len(data) {
		size := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
		opcode := binary.LittleEndian.Uint32(data[offset+4 : offset+8])
		if size < 8 || offset+size > len(data) {
			out.Diagnostics = append(out.Diagnostics, ProxyDiagnostic{
				Offset: offset, Opcode: opcode,
				Message: fmt.Sprintf("invalid chunk size %d", size),
			})
			break
		}
		body := data[offset+8 : offset+size]
		switch opcode {
		case 2: // circle
			p, err := proxyCircle(body, 2)
			if err != nil {
				out.Diagnostics = append(out.Diagnostics, proxyChunkError(offset, opcode, err))
				break
			}
			state.apply(&p)
			out.Primitives = append(out.Primitives, p)
		case 3: // circle through three points
			p, err := proxyCircle3P(body)
			if err != nil {
				out.Diagnostics = append(out.Diagnostics, proxyChunkError(offset, opcode, err))
				break
			}
			state.apply(&p)
			out.Primitives = append(out.Primitives, p)
		case 4: // circular arc
			p, err := proxyArc(body, 2)
			if err != nil {
				out.Diagnostics = append(out.Diagnostics, proxyChunkError(offset, opcode, err))
				break
			}
			state.apply(&p)
			out.Primitives = append(out.Primitives, p)
		case 5: // circular arc through three points
			p, err := proxyArc3P(body)
			if err != nil {
				out.Diagnostics = append(out.Diagnostics, proxyChunkError(offset, opcode, err))
				break
			}
			state.apply(&p)
			out.Primitives = append(out.Primitives, p)
		case 6: // polyline
			p, err := proxyPolyline(body, false)
			if err != nil {
				out.Diagnostics = append(out.Diagnostics, proxyChunkError(offset, opcode, err))
				break
			}
			state.apply(&p)
			out.Primitives = append(out.Primitives, p)
		case 7: // polygon
			p, err := proxyPolyline(body, true)
			if err != nil {
				out.Diagnostics = append(out.Diagnostics, proxyChunkError(offset, opcode, err))
				break
			}
			state.apply(&p)
			out.Primitives = append(out.Primitives, p)
		case 8, 9, 32: // mesh / shell / polyline with normals
			p, err := proxyMesh(body, opcode == 32)
			if err != nil {
				out.Diagnostics = append(out.Diagnostics, proxyChunkError(offset, opcode, err))
				break
			}
			state.apply(&p)
			out.Primitives = append(out.Primitives, p)
		case 10, 36: // text and unicode text
			p, err := proxyText(body, opcode == 36)
			if err != nil {
				out.Diagnostics = append(out.Diagnostics, proxyChunkError(offset, opcode, err))
				break
			}
			state.apply(&p)
			out.Primitives = append(out.Primitives, p)
		case 11, 38: // text2 and unicode text2
			p, err := proxyText2(body, opcode == 38)
			if err != nil {
				out.Diagnostics = append(out.Diagnostics, proxyChunkError(offset, opcode, err))
				break
			}
			state.apply(&p)
			out.Primitives = append(out.Primitives, p)
		case 12, 13: // xline / ray
			p, err := proxyRay(body)
			if err != nil {
				out.Diagnostics = append(out.Diagnostics, proxyChunkError(offset, opcode, err))
				break
			}
			state.apply(&p)
			out.Primitives = append(out.Primitives, p)
		case 14: // colour
			if len(body) < 4 {
				out.Diagnostics = append(out.Diagnostics, proxyChunkError(offset, opcode, fmt.Errorf("truncated colour")))
			} else {
				state.color = int(binary.LittleEndian.Uint32(body[:4]))
				state.hasColor = true
			}
		case 16: // layer index
			if len(body) < 4 {
				out.Diagnostics = append(out.Diagnostics, proxyChunkError(offset, opcode, fmt.Errorf("truncated layer index")))
			} else {
				state.layer = int(binary.LittleEndian.Uint32(body[:4]))
			}
		case 29, 30: // push matrix
			if len(body) < 128 {
				out.Diagnostics = append(out.Diagnostics, proxyChunkError(offset, opcode, fmt.Errorf("truncated transform")))
			} else {
				values := make([]float64, 16)
				for i := range values {
					values[i] = float64frombits(binary.LittleEndian.Uint64(body[i*8 : i*8+8]))
				}
				m := dxfmath.NewMatrix44FromSlice(values).Transpose()
				state.transform = m.Mul(state.transform)
			}
		case 31: // pop matrix
			state.transform = dxfmath.NewMatrix44()
		case 44: // elliptic arc
			p, err := proxyEllipticArc(body)
			if err != nil {
				out.Diagnostics = append(out.Diagnostics, proxyChunkError(offset, opcode, err))
				break
			}
			state.apply(&p)
			out.Primitives = append(out.Primitives, p)
		default:
			out.Diagnostics = append(out.Diagnostics, ProxyDiagnostic{
				Offset: offset, Opcode: opcode,
				Message: "unsupported proxy graphics opcode",
			})
		}
		offset += size
	}
	return out, nil
}

type proxyState struct {
	color     int
	hasColor  bool
	layer     int
	transform dxfmath.Matrix44
}

func (s *proxyState) apply(p *ProxyPrimitive) {
	p.Color = s.color
	p.HasColor = s.hasColor
	p.LayerIndex = s.layer
	p.HasLayer = s.layer >= 0
	p.Transform = s.transform
	if !s.transform.IsIdentity(1e-12) {
		p.Center = s.transform.MulVec(p.Center)
		for i := range p.Vertices {
			p.Vertices[i] = s.transform.MulVec(p.Vertices[i])
		}
		scale := 1.0
		if sx, sy, _ := ScaleFactors(s.transform); sx != 0 {
			scale = sx
			if sy != 0 && math.Abs(sy) < math.Abs(scale) {
				scale = sy
			}
		}
		p.Radius *= math.Abs(scale)
		p.Height *= math.Abs(scale)
	}
}

func proxyChunkError(offset int, opcode uint32, err error) ProxyDiagnostic {
	return ProxyDiagnostic{Offset: offset, Opcode: opcode, Message: err.Error()}
}

func proxyCircle(data []byte, _ int) (ProxyPrimitive, error) {
	r := newProxyReader(data)
	center, err := r.vertex()
	if err != nil {
		return ProxyPrimitive{}, err
	}
	radius, err := r.float()
	if err != nil {
		return ProxyPrimitive{}, err
	}
	if _, err := r.vertex(); err != nil { // normal
		return ProxyPrimitive{}, err
	}
	return ProxyPrimitive{Kind: ProxyCircle, Center: center, Radius: radius}, nil
}

func proxyCircle3P(data []byte) (ProxyPrimitive, error) {
	r := newProxyReader(data)
	p1, err := r.vertex()
	if err != nil {
		return ProxyPrimitive{}, err
	}
	p2, err := r.vertex()
	if err != nil {
		return ProxyPrimitive{}, err
	}
	p3, err := r.vertex()
	if err != nil {
		return ProxyPrimitive{}, err
	}
	center, radius, ok := circleFrom3Points(p1, p2, p3)
	if !ok {
		return ProxyPrimitive{}, fmt.Errorf("collinear points")
	}
	return ProxyPrimitive{Kind: ProxyCircle, Center: center, Radius: radius}, nil
}

func proxyArc(data []byte, _ int) (ProxyPrimitive, error) {
	r := newProxyReader(data)
	center, err := r.vertex()
	if err != nil {
		return ProxyPrimitive{}, err
	}
	radius, err := r.float()
	if err != nil {
		return ProxyPrimitive{}, err
	}
	if _, err := r.vertex(); err != nil { // normal
		return ProxyPrimitive{}, err
	}
	start, err := r.vertex()
	if err != nil {
		return ProxyPrimitive{}, err
	}
	sweep, err := r.float()
	if err != nil {
		return ProxyPrimitive{}, err
	}
	startAngle := math.Atan2(start.Y(), start.X()) * 180 / math.Pi
	return ProxyPrimitive{
		Kind: ProxyArc, Center: center, Radius: radius,
		StartAngle: startAngle, EndAngle: startAngle + sweep*180/math.Pi,
	}, nil
}

func proxyArc3P(data []byte) (ProxyPrimitive, error) {
	r := newProxyReader(data)
	p1, err := r.vertex()
	if err != nil {
		return ProxyPrimitive{}, err
	}
	p2, err := r.vertex()
	if err != nil {
		return ProxyPrimitive{}, err
	}
	p3, err := r.vertex()
	if err != nil {
		return ProxyPrimitive{}, err
	}
	center, radius, ok := circleFrom3Points(p1, p2, p3)
	if !ok {
		return ProxyPrimitive{}, fmt.Errorf("collinear points")
	}
	return ProxyPrimitive{
		Kind: ProxyArc, Center: center, Radius: radius,
		StartAngle: math.Atan2(p1.Y()-center.Y(), p1.X()-center.X()) * 180 / math.Pi,
		EndAngle:   math.Atan2(p3.Y()-center.Y(), p3.X()-center.X()) * 180 / math.Pi,
	}, nil
}

func proxyPolyline(data []byte, closed bool) (ProxyPrimitive, error) {
	vertices, err := proxyVertices(data, false)
	if err != nil {
		return ProxyPrimitive{}, err
	}
	kind := ProxyPolyline
	if closed {
		kind = ProxyPolyline
	}
	if len(vertices) == 1 || (len(vertices) == 2 && vertices[0] == vertices[1]) {
		return ProxyPrimitive{Kind: ProxyPoint, Vertices: vertices[:1]}, nil
	}
	if len(vertices) == 2 && !closed {
		kind = ProxyLine
	}
	return ProxyPrimitive{Kind: kind, Vertices: vertices, Closed: closed}, nil
}

func proxyMesh(data []byte, withNormal bool) (ProxyPrimitive, error) {
	vertices, err := proxyVertices(data, withNormal)
	if err != nil {
		return ProxyPrimitive{}, err
	}
	return ProxyPrimitive{Kind: ProxyMesh, Vertices: vertices}, nil
}

func proxyVertices(data []byte, withNormal bool) ([]dxfmath.Vec3, error) {
	r := newProxyReader(data)
	count, err := r.long()
	if err != nil {
		return nil, err
	}
	if count < 0 || count > 1<<20 {
		return nil, fmt.Errorf("invalid vertex count %d", count)
	}
	if withNormal {
		count++
	}
	vertices := make([]dxfmath.Vec3, 0, count)
	for i := 0; i < count; i++ {
		v, err := r.vertex()
		if err != nil {
			return nil, err
		}
		vertices = append(vertices, v)
	}
	if withNormal && len(vertices) > 0 {
		vertices = vertices[:len(vertices)-1]
	}
	return vertices, nil
}

func proxyText(data []byte, unicodeText bool) (ProxyPrimitive, error) {
	r := newProxyReader(data)
	start, err := r.vertex()
	if err != nil {
		return ProxyPrimitive{}, err
	}
	if _, err := r.vertex(); err != nil { // normal
		return ProxyPrimitive{}, err
	}
	direction, err := r.vertex()
	if err != nil {
		return ProxyPrimitive{}, err
	}
	height, err := r.float()
	if err != nil {
		return ProxyPrimitive{}, err
	}
	width, err := r.float()
	if err != nil {
		return ProxyPrimitive{}, err
	}
	oblique, err := r.float()
	if err != nil {
		return ProxyPrimitive{}, err
	}
	text, err := r.paddedString(unicodeText)
	if err != nil {
		return ProxyPrimitive{}, err
	}
	return ProxyPrimitive{
		Kind: ProxyText, Vertices: []dxfmath.Vec3{start}, Text: text,
		Height: height, WidthFact: width, Oblique: oblique * 180 / math.Pi,
		Rotation: math.Atan2(direction.Y(), direction.X()) * 180 / math.Pi,
	}, nil
}

func proxyText2(data []byte, unicodeText bool) (ProxyPrimitive, error) {
	p, err := proxyText(data, unicodeText)
	if err != nil {
		return ProxyPrimitive{}, err
	}
	return p, nil
}

func proxyRay(data []byte) (ProxyPrimitive, error) {
	r := newProxyReader(data)
	p1, err := r.vertex()
	if err != nil {
		return ProxyPrimitive{}, err
	}
	p2, err := r.vertex()
	if err != nil {
		return ProxyPrimitive{}, err
	}
	return ProxyPrimitive{Kind: ProxyLine, Vertices: []dxfmath.Vec3{p1, p2}}, nil
}

func proxyEllipticArc(data []byte) (ProxyPrimitive, error) {
	r := newProxyReader(data)
	center, err := r.vertex()
	if err != nil {
		return ProxyPrimitive{}, err
	}
	if _, err := r.vertex(); err != nil { // extrusion
		return ProxyPrimitive{}, err
	}
	major, err := r.float()
	if err != nil {
		return ProxyPrimitive{}, err
	}
	minor, err := r.float()
	if err != nil {
		return ProxyPrimitive{}, err
	}
	start, err := r.float()
	if err != nil {
		return ProxyPrimitive{}, err
	}
	end, err := r.float()
	if err != nil {
		return ProxyPrimitive{}, err
	}
	if _, err := r.float(); err != nil { // major axis angle
		return ProxyPrimitive{}, err
	}
	ratio := 1.0
	if major != 0 {
		ratio = minor / major
	}
	vertices := make([]dxfmath.Vec3, 0, 33)
	for i := 0; i <= 32; i++ {
		t := start + (end-start)*float64(i)/32
		vertices = append(vertices, dxfmath.NewVec3(
			center.X()+major*math.Cos(t),
			center.Y()+major*ratio*math.Sin(t),
			center.Z(),
		))
	}
	return ProxyPrimitive{Kind: ProxyEllipticArc, Center: center, Radius: major, Vertices: vertices}, nil
}

type proxyReader struct {
	data []byte
	pos  int
}

func newProxyReader(data []byte) *proxyReader { return &proxyReader{data: data} }

func (r *proxyReader) bytes(n int) ([]byte, error) {
	if r.pos+n > len(r.data) {
		return nil, fmt.Errorf("unexpected end of chunk")
	}
	out := r.data[r.pos : r.pos+n]
	r.pos += n
	return out, nil
}

func (r *proxyReader) vertex() (dxfmath.Vec3, error) {
	b, err := r.bytes(24)
	if err != nil {
		return dxfmath.Vec3{}, err
	}
	return dxfmath.NewVec3(
		float64frombits(binary.LittleEndian.Uint64(b[0:8])),
		float64frombits(binary.LittleEndian.Uint64(b[8:16])),
		float64frombits(binary.LittleEndian.Uint64(b[16:24])),
	), nil
}

func (r *proxyReader) float() (float64, error) {
	b, err := r.bytes(8)
	if err != nil {
		return 0, err
	}
	return float64frombits(binary.LittleEndian.Uint64(b)), nil
}

func (r *proxyReader) long() (int, error) {
	b, err := r.bytes(4)
	if err != nil {
		return 0, err
	}
	return int(int32(binary.LittleEndian.Uint32(b))), nil
}

func (r *proxyReader) paddedString(unicodeText bool) (string, error) {
	if r.pos >= len(r.data) {
		return "", nil
	}
	rest := r.data[r.pos:]
	if unicodeText {
		end := len(rest)
		if i := bytes.Index(rest, []byte{0, 0}); i >= 0 {
			end = i
		}
		words := make([]uint16, 0, end/2)
		for i := 0; i+1 < end; i += 2 {
			words = append(words, binary.LittleEndian.Uint16(rest[i:i+2]))
		}
		r.pos = len(r.data)
		return string(utf16.Decode(words)), nil
	}
	if i := bytes.IndexByte(rest, 0); i >= 0 {
		r.pos += i + 1
		return string(rest[:i]), nil
	}
	r.pos = len(r.data)
	return string(rest), nil
}

func float64frombits(v uint64) float64 { return math.Float64frombits(v) }

func circleFrom3Points(p1, p2, p3 dxfmath.Vec3) (dxfmath.Vec3, float64, bool) {
	d := 2 * (p1.X()*(p2.Y()-p3.Y()) + p2.X()*(p3.Y()-p1.Y()) + p3.X()*(p1.Y()-p2.Y()))
	if math.Abs(d) < 1e-12 {
		return dxfmath.Vec3{}, 0, false
	}
	p1s := p1.X()*p1.X() + p1.Y()*p1.Y()
	p2s := p2.X()*p2.X() + p2.Y()*p2.Y()
	p3s := p3.X()*p3.X() + p3.Y()*p3.Y()
	ux := (p1s*(p2.Y()-p3.Y()) + p2s*(p3.Y()-p1.Y()) + p3s*(p1.Y()-p2.Y())) / d
	uy := (p1s*(p3.X()-p2.X()) + p2s*(p1.X()-p3.X()) + p3s*(p2.X()-p1.X())) / d
	center := dxfmath.NewVec3(ux, uy, p1.Z())
	return center, math.Hypot(p1.X()-ux, p1.Y()-uy), true
}
