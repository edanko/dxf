package lldxf

import (
	"encoding/binary"
	"strings"
	"testing"
)

func binaryTag(code int16, value string) []byte {
	out := make([]byte, 2, 2+len(value)+1)
	binary.LittleEndian.PutUint16(out, uint16(code))
	out = append(out, value...)
	return append(out, 0)
}

func TestParseRawSectionsASCII(t *testing.T) {
	data := []byte("0\nSECTION\n2\nHEADER\n9\n$ACADVER\n1\nAC1015\n0\nENDSEC\n0\nSECTION\n2\nENTITIES\n0\nPOINT\n10\n1.0\n20\n2.0\n0\nENDSEC\n0\nEOF\n")
	rs, err := ParseRawSections(data)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(rs.Order) != 2 || rs.Order[0] != "HEADER" || rs.Order[1] != "ENTITIES" {
		t.Fatalf("order = %v", rs.Order)
	}
	ents, ok := rs.Tags("ENTITIES")
	if !ok || len(ents) != 3 {
		t.Fatalf("entities = %v, ok = %v", ents, ok)
	}
	if ents.DXFType() != "POINT" {
		t.Fatalf("entity type = %q", ents.DXFType())
	}
}

func TestParseRawSectionsBinary(t *testing.T) {
	var data []byte
	data = append(data, []byte(BinaryDXFSignature)...)
	data = append(data, binaryTag(0, "SECTION")...)
	data = append(data, binaryTag(2, "ENTITIES")...)
	data = append(data, binaryTag(0, "LINE")...)
	data = append(data, binaryTag(0, "ENDSEC")...)
	data = append(data, binaryTag(0, "EOF")...)
	rs, err := ParseRawSections(data)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	ents, ok := rs.Tags("ENTITIES")
	if !ok || len(ents) != 1 || ents[0].Value() != "LINE" {
		t.Fatalf("entities = %v, ok = %v", ents, ok)
	}
}

func TestParseRawSectionsTruncatedNamesOpenSection(t *testing.T) {
	data := []byte("0\nSECTION\n2\nENTITIES\n0\nPOINT\n10\n1.0\n20\n2.0\n")
	rs, err := ParseRawSections(data)
	if err == nil || !strings.Contains(err.Error(), "ENTITIES") {
		t.Fatalf("error = %v", err)
	}
	ents, ok := rs.Tags("ENTITIES")
	if !ok || len(ents) != 3 {
		t.Fatalf("partial entities = %v, ok = %v", ents, ok)
	}
}
