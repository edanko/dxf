package text2path

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestParseSHXShapesAndMeasure(t *testing.T) {
	data := syntheticSHX()
	font, err := ParseSHX(data)
	if err != nil {
		t.Fatal(err)
	}
	if font.Above != 21 || font.Below != 7 {
		t.Fatalf("metrics header = above %v below %v", font.Above, font.Below)
	}
	glyph, ok := font.Glyph('A')
	if !ok || glyph.Name != "A" {
		t.Fatalf("glyph A = %+v, ok=%v", glyph, ok)
	}
	metrics := font.Measure("A A", 21)
	if metrics.Height != 21 {
		t.Fatalf("height = %v", metrics.Height)
	}
	if metrics.Width <= 0 {
		t.Fatalf("width = %v", metrics.Width)
	}
	if metrics.Ascent < 0 {
		t.Fatalf("ascent = %v", metrics.Ascent)
	}
}

func TestParseSHXRejectsUnknown(t *testing.T) {
	if _, err := ParseSHX([]byte("not a font")); err == nil {
		t.Fatal("expected an error")
	}
}

func syntheticSHX() []byte {
	var records [][]byte
	records = append(records, append([]byte("FONT\x00"), 21, 7, 0, 0))
	records = append(records, append([]byte("SPACE\x00"), 0))
	records = append(records, append([]byte("A\x00"), 160, 0))

	var out bytes.Buffer
	header := []byte("AutoCAD-86 shapes 1.0")
	out.Write(header)
	out.Write(bytes.Repeat([]byte{0}, 0x17-len(header)))
	out.WriteByte(0x1A)
	writeU16(&out, 0)
	writeU16(&out, 65)
	writeU16(&out, uint16(len(records)))
	for i, rec := range records {
		number := []int{0, 32, 65}[i]
		writeU16(&out, uint16(number))
		writeU16(&out, uint16(len(rec)))
	}
	for _, rec := range records {
		out.Write(rec)
	}
	out.WriteString("EOF")
	return out.Bytes()
}

func writeU16(out *bytes.Buffer, v uint16) {
	var buf [2]byte
	binary.LittleEndian.PutUint16(buf[:], v)
	out.Write(buf[:])
}
