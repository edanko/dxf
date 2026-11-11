package dwg

import (
	"encoding/binary"
	"strings"
	"testing"
)

func locatorFixture(version string, sections []Section) []byte {
	data := make([]byte, 0x19+9*len(sections))
	copy(data, version)
	binary.LittleEndian.PutUint16(data[0x13:0x15], 30)
	binary.LittleEndian.PutUint32(data[0x15:0x19], uint32(len(sections)))
	offset := 0x19
	for _, s := range sections {
		data[offset] = byte(s.Number)
		binary.LittleEndian.PutUint32(data[offset+1:offset+5], s.Address)
		binary.LittleEndian.PutUint32(data[offset+5:offset+9], s.Size)
		offset += 9
	}
	return data
}

func TestReadMetadataLocator(t *testing.T) {
	data := locatorFixture("AC1015", []Section{
		{Number: 0, Address: 0x100, Size: 32},
		{Number: 1, Address: 0x120, Size: 64},
		{Number: 2, Address: 0x160, Size: 128},
		{Number: 9, Address: 0x1e0, Size: 16},
	})
	md, err := ReadMetadata(data)
	if err != nil {
		t.Fatalf("read metadata: %v", err)
	}
	if md.Version != "AC1015" || md.Release != "R2000" {
		t.Fatalf("version = %q release = %q", md.Version, md.Release)
	}
	if md.Codepage != 30 {
		t.Fatalf("codepage = %d", md.Codepage)
	}
	if len(md.Sections) != 4 {
		t.Fatalf("sections = %d", len(md.Sections))
	}
	if md.Sections[0].Name != "Header variables" || !md.Sections[0].Supported {
		t.Fatalf("section 0 = %+v", md.Sections[0])
	}
	if md.Sections[2].Name != "Handles (object map)" {
		t.Fatalf("section 2 = %+v", md.Sections[2])
	}
	if md.Sections[3].Supported {
		t.Fatalf("unknown section marked supported: %+v", md.Sections[3])
	}
	if len(md.Diagnostics) != 1 || md.Diagnostics[0].Code != "unsupported-section" {
		t.Fatalf("diagnostics = %+v", md.Diagnostics)
	}
}

func TestReadMetadataR2004ReportsSectionMapUnsupported(t *testing.T) {
	data := make([]byte, 0x2c)
	copy(data, "AC1032")
	data[0x16] = 32
	data[0x17] = 1
	binary.LittleEndian.PutUint32(data[0x28:0x2c], 0x80)
	md, err := ReadMetadata(data)
	if err != nil {
		t.Fatalf("read metadata: %v", err)
	}
	if md.Release != "R2018" || md.SectionMapAddr != 0x80 {
		t.Fatalf("metadata = %+v", md)
	}
	if len(md.Diagnostics) != 1 || md.Diagnostics[0].Code != "section-map-unsupported" {
		t.Fatalf("diagnostics = %+v", md.Diagnostics)
	}
}

func TestReadMetadataRejectsInvalidInput(t *testing.T) {
	if _, err := ReadMetadata([]byte("AC10")); err == nil {
		t.Fatal("expected short input error")
	}
	data := make([]byte, 0x2c)
	copy(data, "AC9999")
	if _, err := ReadMetadata(data); err == nil || !strings.Contains(err.Error(), "unsupported version") {
		t.Fatalf("unexpected error: %v", err)
	}
}
