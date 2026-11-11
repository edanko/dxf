package dwg

import (
	"encoding/binary"
	"fmt"
	"os"
)

// Metadata describes the parts of a DWG file that can be read without a full
// geometry decoder. It is intentionally read-only: this package never writes
// DWG output, and ReadMetadata has no side effects on disk.
type Metadata struct {
	Version        string
	Release        string
	MaintRelease   uint8
	AppVersion     uint8
	AppMaint       uint8
	Codepage       uint16
	Sections       []Section
	Diagnostics    []Diagnostic
	SectionMapAddr uint32
}

// Section is one entry of the DWG section locator.
type Section struct {
	Number    uint32
	Name      string
	Address   uint32
	Size      uint32
	Supported bool
}

// Diagnostic reports a partially parsed or unsupported part of the file.
type Diagnostic struct {
	Code    string
	Offset  int
	Message string
}

func (d Diagnostic) String() string {
	if d.Offset > 0 {
		return fmt.Sprintf("%s at %#x: %s", d.Code, d.Offset, d.Message)
	}
	return fmt.Sprintf("%s: %s", d.Code, d.Message)
}

var dwgReleases = map[string]string{
	"AC1006": "R10",
	"AC1009": "R11/R12",
	"AC1012": "R13",
	"AC1013": "R13c3",
	"AC1014": "R14",
	"AC1015": "R2000",
	"AC1018": "R2004",
	"AC1021": "R2007",
	"AC1024": "R2010",
	"AC1027": "R2013",
	"AC1032": "R2018",
}

// ReadMetadataFile reads a DWG file and returns its version and section
// inventory. The file is opened read-only.
func ReadMetadataFile(path string) (*Metadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ReadMetadata(data)
}

// ReadMetadata parses the DWG header and, for R13-R2000 files, the section
// locator. R2004 and newer files expose version metadata and report the
// encrypted section map as unsupported instead of guessing addresses.
func ReadMetadata(data []byte) (*Metadata, error) {
	if len(data) < 6 {
		return nil, fmt.Errorf("DWG: input too short for version header (%d bytes)", len(data))
	}
	version := string(data[:6])
	release, ok := dwgReleases[version]
	if !ok {
		if len(version) >= 2 && version[:2] == "AC" {
			return nil, fmt.Errorf("DWG: unsupported version %q", version)
		}
		return nil, fmt.Errorf("DWG: invalid version header %q", version)
	}
	md := &Metadata{Version: version, Release: release}

	if len(data) < 0x19 {
		md.Diagnostics = append(md.Diagnostics, Diagnostic{
			Code:    "short-header",
			Offset:  len(data),
			Message: "header ends before the section locator",
		})
		return md, nil
	}
	md.MaintRelease = data[0x0b]
	if len(data) >= 0x15 {
		md.Codepage = binary.LittleEndian.Uint16(data[0x13:0x15])
	}

	if isR2004OrLater(version) {
		return readR2004Metadata(data, md)
	}
	return readLocatorMetadata(data, md)
}

func isR2004OrLater(version string) bool {
	switch version {
	case "AC1018", "AC1021", "AC1024", "AC1027", "AC1032":
		return true
	default:
		return false
	}
}

func readLocatorMetadata(data []byte, md *Metadata) (*Metadata, error) {
	if len(data) < 0x19 {
		md.Diagnostics = append(md.Diagnostics, Diagnostic{
			Code:    "short-locator",
			Offset:  0x15,
			Message: "section count is truncated",
		})
		return md, nil
	}
	count := binary.LittleEndian.Uint32(data[0x15:0x19])
	if count > 64 {
		return nil, fmt.Errorf("DWG: implausible section count %d", count)
	}
	offset := 0x19
	for i := range count {
		if offset+9 > len(data) {
			md.Diagnostics = append(md.Diagnostics, Diagnostic{
				Code:    "truncated-section-locator",
				Offset:  offset,
				Message: fmt.Sprintf("section locator %d ends past the file", i),
			})
			return md, nil
		}
		number := uint32(data[offset])
		address := binary.LittleEndian.Uint32(data[offset+1 : offset+5])
		size := binary.LittleEndian.Uint32(data[offset+5 : offset+9])
		name, supported := sectionName(number)
		md.Sections = append(md.Sections, Section{
			Number:    number,
			Name:      name,
			Address:   address,
			Size:      size,
			Supported: supported,
		})
		if !supported {
			md.Diagnostics = append(md.Diagnostics, Diagnostic{
				Code:    "unsupported-section",
				Offset:  offset,
				Message: fmt.Sprintf("section %d has no native reader", number),
			})
		}
		offset += 9
	}
	return md, nil
}

func readR2004Metadata(data []byte, md *Metadata) (*Metadata, error) {
	if len(data) < 0x2c {
		return nil, fmt.Errorf("DWG: input too short for %s header (%d bytes)", md.Version, len(data))
	}
	md.AppVersion = data[0x16]
	md.AppMaint = data[0x17]
	md.SectionMapAddr = binary.LittleEndian.Uint32(data[0x28:0x2c])
	md.Diagnostics = append(md.Diagnostics, Diagnostic{
		Code:    "section-map-unsupported",
		Offset:  0x28,
		Message: fmt.Sprintf("%s section map is encrypted or compressed; only version metadata is available", md.Release),
	})
	return md, nil
}

func sectionName(number uint32) (string, bool) {
	switch number {
	case 0:
		return "Header variables", true
	case 1:
		return "Classes", true
	case 2:
		return "Handles (object map)", true
	case 3:
		return "Object free space", true
	case 4:
		return "Template", true
	case 5:
		return "Aux header", true
	default:
		return fmt.Sprintf("section %d", number), false
	}
}
