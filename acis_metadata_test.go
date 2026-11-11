package dxf

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"math"
	"strings"
	"testing"

	"github.com/edanko/dxf/entity"
)

func TestACISSATMetadataAndPayload(t *testing.T) {
	body := "0\n3DSOLID\n5\nS1\n100\nAcDbEntity\n8\n0\n100\nAcDbModelerGeometry\n70\n1\n" +
		"1\n700 0 2 0\n" +
		"1\n@12 ACIS 32.0 NT\n" +
		"1\n1 0 0 0\n" +
		"1\nbody $1 1 $-1 #\n" +
		"1\nbody $2 1 $-1 #\n"
	e, out := loadOne(t, body)
	s, ok := e.(*entity.Solid3d)
	if !ok {
		t.Fatalf("type %T (out:\n%s)", e, out)
	}
	info := s.ACISInfo()
	if info.Format != entity.ACISFormatSAT {
		t.Fatalf("format = %v (out:\n%s)", info.Format, out)
	}
	if info.Version != 700 {
		t.Errorf("version = %d, want 700", info.Version)
	}
	if info.RecordCount != 2 || info.BodyCount != 2 || !info.HasBody {
		t.Errorf("metadata = %+v", info)
	}
	want := []string{
		"700 0 2 0",
		"@12 ACIS 32.0 NT",
		"1 0 0 0",
		"body $1 1 $-1 #",
		"body $2 1 $-1 #",
	}
	got := s.SATLines()
	if len(got) != len(want) {
		t.Fatalf("sat lines = %q", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("line %d = %q, want %q", i, got[i], want[i])
		}
	}
	for _, line := range want {
		if !strings.Contains(out, "1\n"+line) {
			t.Errorf("missing code-1 payload line %q in\n%s", line, out)
		}
	}
}

func TestACISSABMetadataAndByteEquality(t *testing.T) {
	payload := testSABPayload()
	half := len(payload) / 2
	chunk1 := strings.ToUpper(hex.EncodeToString(payload[:half]))
	chunk2 := strings.ToUpper(hex.EncodeToString(payload[half:]))
	body := "0\n3DSOLID\n5\nS1\n100\nAcDbEntity\n8\n0\n100\nAcDbModelerGeometry\n70\n1\n" +
		"100\nAcDb3dSolid\n70\n21800\n" +
		"310\n" + chunk1 + "\n310\n" + chunk2 + "\n"
	e, out := loadOne(t, body)
	s, ok := e.(*entity.Solid3d)
	if !ok {
		t.Fatalf("type %T (out:\n%s)", e, out)
	}
	if !s.HasBinaryACIS() {
		t.Fatalf("binary ACIS not reported")
	}
	info := s.ACISInfo()
	if info.Format != entity.ACISFormatSAB || info.Version != 21800 {
		t.Errorf("metadata = %+v", info)
	}
	if !bytes.Equal(s.ACISBinaryData(), payload) {
		t.Fatalf("payload changed: got %x want %x", s.ACISBinaryData(), payload)
	}
	for _, chunk := range []string{chunk1, chunk2} {
		if !strings.Contains(out, "310\n"+chunk) {
			t.Errorf("missing payload chunk %q in\n%s", chunk, out)
		}
	}
}

func TestACISSABACDSDataLinkage(t *testing.T) {
	payload := testSABPayload()
	chunk := strings.ToUpper(hex.EncodeToString(payload))
	text := "0\nSECTION\n2\nENTITIES\n" +
		"0\n3DSOLID\n5\nS1\n100\nAcDbEntity\n8\n0\n100\nAcDbModelerGeometry\n70\n1\n" +
		"0\nENDSEC\n" +
		"0\nSECTION\n2\nACDSDATA\n70\n2\n71\n8\n" +
		"0\nACDSRECORD\n90\n1\n2\nAcDbDs::ID\n280\n10\n320\nS1\n" +
		"2\nASM_Data\n280\n15\n94\n" + f(float64(len(payload))) + "\n310\n" + chunk + "\n" +
		"0\nENDSEC\n0\nEOF\n"
	d, err := FromStringData(text)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	var solid *entity.Solid3d
	for _, e := range d.Entities() {
		if s, ok := e.(*entity.Solid3d); ok {
			solid = s
			break
		}
	}
	if solid == nil {
		t.Fatal("3DSOLID not loaded")
	}
	if !solid.HasBinaryACIS() {
		t.Fatal("ACDSDATA payload not linked to 3DSOLID")
	}
	if !bytes.Equal(solid.ACISBinaryData(), payload) {
		t.Fatalf("payload changed: got %x want %x", solid.ACISBinaryData(), payload)
	}
	var buf bytes.Buffer
	if _, err := d.WriteTo(&buf); err != nil {
		t.Fatalf("write: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "ACDSDATA") || !strings.Contains(out, "310\n"+chunk) {
		t.Errorf("ACDSDATA payload not preserved in\n%s", out)
	}
}

func testSABPayload() []byte {
	var b bytes.Buffer
	b.WriteString("ACIS BinaryFile")
	_ = binary.Write(&b, binary.LittleEndian, int32(21800))
	_ = binary.Write(&b, binary.LittleEndian, int32(1))
	_ = binary.Write(&b, binary.LittleEndian, int32(2))
	_ = binary.Write(&b, binary.LittleEndian, int32(0))
	writeSABString(&b, "go-dxf")
	writeSABString(&b, "ACIS 218.00 NT")
	writeSABString(&b, "Mon Nov 10 00:00:00 2026")
	_ = binary.Write(&b, binary.LittleEndian, math.Float64bits(1.0))
	_ = binary.Write(&b, binary.LittleEndian, math.Float64bits(9.9999999999999995e-7))
	_ = binary.Write(&b, binary.LittleEndian, math.Float64bits(1e-10))
	b.WriteString("body-data")
	return b.Bytes()
}

func writeSABString(b *bytes.Buffer, s string) {
	b.WriteByte(0x07)
	b.WriteByte(byte(len(s)))
	b.WriteString(s)
}
