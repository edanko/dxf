package text2path

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"strings"
)

// SHXGlyph is a single shape record from an AutoCAD SHX font.
type SHXGlyph struct {
	Number int
	Name   string
	Codes  []int
}

// SHXFont is a loaded AutoCAD SHX font file.
type SHXFont struct {
	Name       string
	Above      float64
	Below      float64
	Mode       int
	Encoding   int
	Embedding  int
	Glyphs     map[int]*SHXGlyph
	shapeCount int
}

// LoadSHX loads an AutoCAD SHX font from disk.
func LoadSHX(path string) (*SHXFont, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseSHX(data)
}

// ParseSHX parses the compiled SHX font format.
func ParseSHX(data []byte) (*SHXFont, error) {
	switch {
	case strings.HasPrefix(string(data), "AutoCAD-86 unifont 1.0"):
		return parseSHXUnifont(data)
	case strings.HasPrefix(string(data), "AutoCAD-86 shapes 1.0"),
		strings.HasPrefix(string(data), "AutoCAD-86 shapes 1.1"):
		return parseSHXShapes(data)
	case strings.HasPrefix(string(data), "AutoCAD-86 bigfont 1.0"):
		return nil, fmt.Errorf("text2path: BIGFONT SHX files are not supported")
	default:
		return nil, fmt.Errorf("text2path: unrecognized SHX font format")
	}
}

// Glyph returns the shape record for a rune.
func (f *SHXFont) Glyph(r rune) (*SHXGlyph, bool) {
	if f == nil {
		return nil, false
	}
	g, ok := f.Glyphs[int(r)]
	return g, ok
}

// Measure returns the layout metrics of text rendered at the requested
// height. The height is the cap height from the SHX header.
func (f *SHXFont) Measure(text string, height float64) TextMetrics {
	if f == nil {
		return TextMetrics{}
	}
	capHeight := f.Above
	if capHeight <= 0 {
		capHeight = 1
	}
	scale := height / capHeight
	spaceWidth := capHeight * 0.5
	advance := 0.0
	maxY := 0.0
	minY := 0.0
	for _, r := range text {
		g, ok := f.Glyphs[int(r)]
		if !ok {
			if r == ' ' {
				advance += spaceWidth
			}
			continue
		}
		minX, glyphMinY, maxX, glyphMaxY, width := measureSHXCodes(g.Codes)
		_ = minX
		if maxX > 0 || width > 0 {
			advance += math.Max(width, 1)
		} else {
			advance += 1
		}
		maxY = math.Max(maxY, glyphMaxY)
		minY = math.Min(minY, glyphMinY)
	}
	return TextMetrics{
		Width:   advance * scale,
		Height:  height,
		Ascent:  maxY * scale,
		Descent: -minY * scale,
	}
}

// MeasureSHX is a convenience wrapper for a loaded font.
func MeasureSHX(font *SHXFont, text string, height float64) TextMetrics {
	return font.Measure(text, height)
}

func parseSHXShapes(data []byte) (*SHXFont, error) {
	reader := &shxReader{data: data, pos: 0x17}
	if !reader.has(1) || reader.u8() != 0x1A {
		return nil, fmt.Errorf("text2path: SHX signature byte 0x1A not found")
	}
	first := int(reader.u16())
	last := int(reader.u16())
	count := int(reader.u16())
	index := make([]struct {
		number int
		length int
	}, 0, count)
	for i := 0; i < count; i++ {
		index = append(index, struct {
			number int
			length int
		}{int(reader.u16()), int(reader.u16())})
	}
	if len(index) == 0 {
		return nil, fmt.Errorf("text2path: SHX font has no shape records")
	}
	if index[0].number != first || index[len(index)-1].number != last {
		return nil, fmt.Errorf("text2path: invalid SHX index table")
	}
	font := &SHXFont{Glyphs: make(map[int]*SHXGlyph), shapeCount: count}
	for _, entry := range index {
		record := reader.bytes(entry.length)
		if record == nil {
			return nil, fmt.Errorf("text2path: truncated SHX shape record")
		}
		name, codes, err := parseSHXRecord(record, false)
		if err != nil {
			return nil, err
		}
		glyph := &SHXGlyph{Number: entry.number, Name: name, Codes: codes}
		font.Glyphs[entry.number] = glyph
	}
	if definition, ok := font.Glyphs[0]; ok {
		if len(definition.Codes) >= 3 {
			font.Above = float64(definition.Codes[0])
			font.Below = float64(definition.Codes[1])
			font.Mode = definition.Codes[2]
		}
		delete(font.Glyphs, 0)
	}
	return font, nil
}

func parseSHXUnifont(data []byte) (*SHXFont, error) {
	reader := &shxReader{data: data, pos: 0x18}
	if !reader.has(1) || reader.u8() != 0x1A {
		return nil, fmt.Errorf("text2path: SHX signature byte 0x1A not found")
	}
	reader.skip(6)
	name := reader.cstring()
	above := int(reader.u8())
	below := int(reader.u8())
	mode := int(reader.u8())
	encoding := int(reader.u8())
	embedding := int(reader.u8())
	reader.skip(1)
	font := &SHXFont{
		Name:      name,
		Above:     float64(above),
		Below:     float64(below),
		Mode:      mode,
		Encoding:  encoding,
		Embedding: embedding,
		Glyphs:    make(map[int]*SHXGlyph),
	}
	for reader.pos+4 <= len(data) {
		number := int(reader.u16())
		byteCount := int(reader.u16())
		if byteCount == 0 {
			break
		}
		record := reader.bytes(byteCount)
		if record == nil {
			return nil, fmt.Errorf("text2path: truncated SHX unifont record")
		}
		glyphName, codes, err := parseSHXRecord(record, true)
		if err != nil {
			return nil, err
		}
		font.Glyphs[number] = &SHXGlyph{Number: number, Name: glyphName, Codes: codes}
	}
	return font, nil
}

func parseSHXRecord(record []byte, unifont bool) (string, []int, error) {
	reader := &shxReader{data: record}
	name := reader.cstring()
	if reader.pos > len(record) {
		return "", nil, fmt.Errorf("text2path: malformed SHX shape record")
	}
	codes, err := parseSHXCodes(reader, unifont)
	return name, codes, err
}

func parseSHXCodes(reader *shxReader, unifont bool) ([]int, error) {
	var codes []int
	for {
		code := int(reader.u8())
		if reader.pos > len(reader.data) {
			return nil, fmt.Errorf("text2path: truncated SHX shape data")
		}
		codes = append(codes, code)
		switch {
		case code == 0:
			return codes, nil
		case code == 1, code == 2, code == 5, code == 6, code == 14:
			continue
		case code > 14:
			continue
		case code == 3, code == 4:
			codes = append(codes, int(reader.u8()))
		case code == 7:
			if unifont {
				codes = append(codes, int(reader.u16()))
			} else {
				codes = append(codes, int(reader.u8()))
			}
		case code == 8:
			codes = append(codes, int(reader.i8()), int(reader.i8()))
		case code == 9:
			for {
				x, y := int(reader.i8()), int(reader.i8())
				codes = append(codes, x, y)
				if x == 0 && y == 0 {
					break
				}
			}
		case code == 10:
			codes = append(codes, int(reader.u8()), int(reader.octant()))
		case code == 11:
			codes = append(codes, int(reader.u8()), int(reader.u8()), int(reader.u8()), int(reader.u8()), int(reader.octant()))
		case code == 12:
			codes = append(codes, int(reader.i8()), int(reader.i8()), int(reader.i8()))
		case code == 13:
			for {
				x, y := int(reader.i8()), int(reader.i8())
				codes = append(codes, x, y)
				if x == 0 && y == 0 {
					break
				}
				codes = append(codes, int(reader.i8()))
			}
		default:
			return nil, fmt.Errorf("text2path: unsupported SHX code %d", code)
		}
		if reader.pos > len(reader.data) {
			return nil, fmt.Errorf("text2path: truncated SHX shape data")
		}
	}
}

// measureSHXCodes returns the bounding box and advance of a shape.
func measureSHXCodes(codes []int) (minX, minY, maxX, maxY, width float64) {
	pen := &shxMeasurePen{minX: math.Inf(1), minY: math.Inf(1), maxX: math.Inf(-1), maxY: math.Inf(-1)}
	pen.run(codes)
	if math.IsInf(pen.minX, 1) {
		return 0, 0, 0, 0, 0
	}
	return pen.minX, pen.minY, pen.maxX, pen.maxY, pen.maxX
}

var shxVecX = []float64{1, 1, 1, 0.5, 0, -0.5, -1, -1, -1, -1, -1, -0.5, 0, 0.5, 1, 1}
var shxVecY = []float64{0, 0.5, 1, 1, 1, 1, 1, 0.5, 0, -0.5, -1, -1, -1, -1, -1, -0.5}

type shxMeasurePen struct {
	x, y    float64
	vector  float64
	minX    float64
	minY    float64
	maxX    float64
	maxY    float64
	stack   []shxPoint
	visited map[int]bool
}

type shxPoint struct{ x, y float64 }

func (p *shxMeasurePen) run(codes []int) {
	if p.visited == nil {
		p.visited = make(map[int]bool)
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
			continue
		case code == 2:
			continue
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
					p.visited[number] = true
					// Sub-shapes are not available to the metrics pass without
					// the font; their absence is ignored.
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
				i += 2
				if !skip {
					p.displace(radius, 0)
					p.displace(-radius, 0)
				}
			}
		case code == 11:
			if i+4 < len(codes) {
				radius := float64((codes[i+2] << 8) + codes[i+3])
				i += 5
				if !skip {
					p.displace(radius, 0)
					p.displace(-radius, 0)
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
				i += 3
				if !skip {
					p.displace(x, y)
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

func (p *shxMeasurePen) displace(x, y float64) {
	p.x += x * p.vector
	p.y += y * p.vector
	p.include(p.x, p.y)
}

func (p *shxMeasurePen) move(x, y float64) {
	p.x, p.y = x, y
	p.include(x, y)
}

func (p *shxMeasurePen) include(x, y float64) {
	p.minX = math.Min(p.minX, x)
	p.minY = math.Min(p.minY, y)
	p.maxX = math.Max(p.maxX, x)
	p.maxY = math.Max(p.maxY, y)
}

type shxReader struct {
	data []byte
	pos  int
}

func (r *shxReader) has(n int) bool { return r.pos+n <= len(r.data) }

func (r *shxReader) u8() byte {
	if r.pos >= len(r.data) {
		r.pos++
		return 0
	}
	v := r.data[r.pos]
	r.pos++
	return v
}

func (r *shxReader) i8() int8 {
	v := int8(r.u8())
	return v
}

func (r *shxReader) octant() int8 {
	v := r.u8()
	if v&0x80 != 0 {
		return -int8(v & 0x7F)
	}
	return int8(v)
}

func (r *shxReader) u16() uint16 {
	if r.pos+2 > len(r.data) {
		r.pos += 2
		return 0
	}
	v := binary.LittleEndian.Uint16(r.data[r.pos:])
	r.pos += 2
	return v
}

func (r *shxReader) cstring() string {
	start := r.pos
	for r.pos < len(r.data) && r.data[r.pos] != 0 {
		r.pos++
	}
	value := string(r.data[start:r.pos])
	if r.pos < len(r.data) {
		r.pos++
	}
	return value
}

func (r *shxReader) bytes(n int) []byte {
	if n < 0 || r.pos+n > len(r.data) {
		r.pos += n
		return nil
	}
	v := r.data[r.pos : r.pos+n]
	r.pos += n
	return v
}

func (r *shxReader) skip(n int) { r.pos += n }
