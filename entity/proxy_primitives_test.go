package entity

import (
	"encoding/binary"
	"math"
	"testing"
)

func proxyChunk(opcode uint32, body []byte) []byte {
	out := make([]byte, 8+len(body))
	binary.LittleEndian.PutUint32(out[:4], uint32(8+len(body)))
	binary.LittleEndian.PutUint32(out[4:8], opcode)
	for i, b := range body {
		out[8+i] = b
	}
	return out
}

func proxyPayload(chunks ...[]byte) []byte {
	out := make([]byte, 8) // unused proxy header
	for _, chunk := range chunks {
		out = append(out, chunk...)
	}
	return out
}

func proxyVertex(x, y, z float64) []byte {
	out := make([]byte, 24)
	binary.LittleEndian.PutUint64(out[0:8], math.Float64bits(x))
	binary.LittleEndian.PutUint64(out[8:16], math.Float64bits(y))
	binary.LittleEndian.PutUint64(out[16:24], math.Float64bits(z))
	return out
}

func proxyFloat(v float64) []byte {
	out := make([]byte, 8)
	binary.LittleEndian.PutUint64(out, math.Float64bits(v))
	return out
}

func proxyLong(v int32) []byte {
	out := make([]byte, 4)
	binary.LittleEndian.PutUint32(out, uint32(v))
	return out
}

func TestDecodeProxyPrimitives(t *testing.T) {
	circleBody := append(proxyVertex(1, 2, 0), proxyFloat(5)...)
	circleBody = append(circleBody, proxyVertex(0, 0, 1)...)

	arcBody := append(proxyVertex(0, 0, 0), proxyFloat(2)...)
	arcBody = append(arcBody, proxyVertex(0, 0, 1)...)
	arcBody = append(arcBody, proxyVertex(1, 0, 0)...)
	arcBody = append(arcBody, proxyFloat(math.Pi/2)...)
	arcBody = append(arcBody, proxyLong(0)...)

	lineBody := append(proxyLong(2), proxyVertex(0, 0, 0)...)
	lineBody = append(lineBody, proxyVertex(3, 4, 0)...)

	textBody := append(proxyVertex(5, 6, 0), proxyVertex(0, 0, 1)...)
	textBody = append(textBody, proxyVertex(1, 0, 0)...)
	textBody = append(textBody, proxyFloat(2)...)
	textBody = append(textBody, proxyFloat(0.8)...)
	textBody = append(textBody, proxyFloat(0)...)
	textBody = append(textBody, []byte("LABEL\x00")...)

	transform := make([]byte, 128)
	for i := range 16 {
		v := 0.0
		if i%5 == 0 {
			v = 1
		}
		binary.LittleEndian.PutUint64(transform[i*8:i*8+8], math.Float64bits(v))
	}
	binary.LittleEndian.PutUint64(transform[96:104], math.Float64bits(10)) // tx

	payload := proxyPayload(
		proxyChunk(2, circleBody),
		proxyChunk(4, arcBody),
		proxyChunk(14, proxyLong(7)),
		proxyChunk(6, lineBody),
		proxyChunk(29, transform),
		proxyChunk(6, lineBody),
		proxyChunk(31, nil),
		proxyChunk(10, textBody),
		proxyChunk(9999, nil),
	)
	e := NewAcadProxyEntity()
	e.SetGraphicsData(payload)
	set, err := DecodeProxyPrimitives(e)
	if err != nil {
		t.Fatal(err)
	}
	if len(set.Primitives) != 5 {
		t.Fatalf("decoded %d primitives, want 5: %#v", len(set.Primitives), set.Primitives)
	}
	circle := set.Primitives[0]
	if circle.Kind != ProxyCircle || circle.Radius != 5 || circle.Center.X() != 1 || circle.Center.Y() != 2 {
		t.Fatalf("bad circle: %#v", circle)
	}
	arc := set.Primitives[1]
	if arc.Kind != ProxyArc || math.Abs(arc.StartAngle) > 1e-9 || math.Abs(arc.EndAngle-90) > 1e-9 {
		t.Fatalf("bad arc: %#v", arc)
	}
	line := set.Primitives[2]
	if line.Kind != ProxyLine || !line.HasColor || line.Color != 7 {
		t.Fatalf("bad colour line: %#v", line)
	}
	translated := set.Primitives[3]
	if translated.Vertices[0].X() != 10 || translated.Vertices[1].X() != 13 {
		t.Fatalf("transform not applied: %#v", translated.Vertices)
	}
	text := set.Primitives[4]
	if text.Kind != ProxyText || text.Text != "LABEL" || text.Height != 2 || text.WidthFact != 0.8 || text.Rotation != 0 {
		t.Fatalf("bad text: %#v", text)
	}
	if len(set.Diagnostics) != 1 || set.Diagnostics[0].Opcode != 9999 || set.Diagnostics[0].Offset == 0 {
		t.Fatalf("bad diagnostics: %#v", set.Diagnostics)
	}
	lastChunk := len(payload) - 8
	if set.Diagnostics[0].Offset != lastChunk {
		t.Fatalf("diagnostic offset = %d, want %d", set.Diagnostics[0].Offset, lastChunk)
	}
}

func TestDecodeProxyPrimitivesNoPayload(t *testing.T) {
	e := NewAcadProxyEntity()
	if _, err := DecodeProxyPrimitives(e); err == nil {
		t.Fatal("expected no-payload error")
	}
	if _, err := DecodeProxyPrimitives(NewLine()); err == nil {
		t.Fatal("expected wrong-entity error")
	}
}
