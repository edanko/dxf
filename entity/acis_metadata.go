package entity

import (
	"bytes"
	"encoding/binary"
	"math"
	"strconv"
	"strings"
)

// ACISFormat identifies the on-disk encoding of an ACIS payload.
type ACISFormat int

const (
	ACISFormatUnknown ACISFormat = iota
	ACISFormatSAT
	ACISFormatSAB
)

// String returns the human-readable ACIS encoding name.
func (f ACISFormat) String() string {
	switch f {
	case ACISFormatSAT:
		return "SAT"
	case ACISFormatSAB:
		return "SAB"
	default:
		return "unknown"
	}
}

// ACISMetadata is the shallow metadata exposed for ACIS payloads. Geometry is
// intentionally not decoded; the original payload remains the source of truth.
type ACISMetadata struct {
	Format       ACISFormat
	Version      int
	RecordCount  int
	EntityCount  int
	BodyCount    int
	HasBody      bool
	HasHistory   bool
	ProductID    string
	ACISVersion  string
	UnitsInMM    float64
	HasEndMarker bool
}

// HasPayload reports whether any ACIS payload metadata was recognized.
func (m ACISMetadata) HasPayload() bool {
	return m.Format != ACISFormatUnknown && (m.Version != 0 || m.RecordCount != 0 || m.EntityCount != 0)
}

// IsBinary reports whether the payload is SAB encoded.
func (m ACISMetadata) IsBinary() bool { return m.Format == ACISFormatSAB }

// ParseSATMetadata extracts the ACIS header and record/body counts from text
// SAT lines. The lines may be plain text or AutoCAD's encoded SAT form; the
// original lines are not modified.
func ParseSATMetadata(lines []string) ACISMetadata {
	m := ACISMetadata{Format: ACISFormatSAT}
	if len(lines) == 0 {
		return m
	}
	plain := lines
	if !looksLikeSATHeader(firstNonEmptyLine(lines)) {
		plain = decodeSATLines(lines)
	}

	var data []string
	for _, line := range plain {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		switch {
		case strings.HasPrefix(line, "End-of-ACIS-data"),
			strings.HasPrefix(line, "End-of-ASM-data"):
			m.HasEndMarker = true
			continue
		case strings.HasPrefix(line, "Begin-of-ACIS-History-data"),
			strings.HasPrefix(line, "End-of-ACIS-History-data"):
			m.HasHistory = true
			continue
		}
		data = append(data, line)
	}
	if len(data) == 0 {
		return m
	}

	header := strings.Fields(data[0])
	if len(header) > 0 {
		m.Version, _ = strconv.Atoi(header[0])
	}
	if len(header) > 1 {
		m.RecordCount, _ = strconv.Atoi(header[1])
	}
	if len(header) > 2 {
		m.EntityCount, _ = strconv.Atoi(header[2])
	}
	if len(data) > 1 {
		tokens := parseSATHeaderString(data[1])
		if len(tokens) > 0 {
			m.ProductID = tokens[0]
		}
		if len(tokens) > 1 {
			m.ACISVersion = tokens[1]
		}
	}
	if len(data) > 2 {
		units := strings.Fields(data[2])
		if len(units) > 0 {
			m.UnitsInMM, _ = strconv.ParseFloat(units[0], 64)
		}
	}

	parsedRecords := 0
	body := ""
	if len(data) > 3 {
		body = strings.Join(data[3:], " ")
	}
	for record := range strings.SplitSeq(body, "#") {
		record = strings.TrimSpace(record)
		if record == "" {
			continue
		}
		parsedRecords++
		fields := strings.Fields(record)
		if len(fields) == 0 {
			continue
		}
		switch fields[0] {
		case "body":
			m.BodyCount++
		case "history":
			m.HasHistory = true
		}
	}
	if m.RecordCount == 0 {
		m.RecordCount = parsedRecords
	}
	if m.EntityCount == 0 {
		m.EntityCount = parsedRecords
	}
	m.HasBody = m.BodyCount > 0
	return m
}

// ParseSABMetadata extracts the ACIS header from a binary SAB payload. The
// payload may include the ACIS/ASM file signature; it is not modified.
func ParseSABMetadata(data []byte) ACISMetadata {
	m := ACISMetadata{Format: ACISFormatSAB}
	pos := 0
	for _, signature := range [][]byte{
		[]byte("ACIS BinaryFile"),
		[]byte("ASM BinaryFile4"),
	} {
		if bytes.HasPrefix(data, signature) {
			pos = len(signature)
			break
		}
	}
	if len(data) < pos+16 {
		return m
	}
	m.Version = int(int32(binary.LittleEndian.Uint32(data[pos:])))
	m.RecordCount = int(int32(binary.LittleEndian.Uint32(data[pos+4:])))
	m.EntityCount = int(int32(binary.LittleEndian.Uint32(data[pos+8:])))
	m.BodyCount = m.EntityCount
	m.HasBody = m.EntityCount > 0

	// The header continues with three length-prefixed strings (product id,
	// ACIS version, creation date) followed by three doubles. Decode only the
	// strings; unknown trailing bytes stay untouched.
	i := pos + 16
	readString := func() string {
		if i+2 > len(data) || data[i] != 0x07 {
			return ""
		}
		n := int(data[i+1])
		i += 2
		if i+n > len(data) {
			return ""
		}
		s := string(data[i : i+n])
		i += n
		return s
	}
	m.ProductID = readString()
	m.ACISVersion = readString()
	_ = readString() // creation date
	if i+8 <= len(data) {
		m.UnitsInMM = math.Float64frombits(binary.LittleEndian.Uint64(data[i:]))
	}
	return m
}

func firstNonEmptyLine(lines []string) string {
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			return line
		}
	}
	return ""
}

func looksLikeSATHeader(line string) bool {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return false
	}
	_, err := strconv.Atoi(fields[0])
	return err == nil
}

func parseSATHeaderString(s string) []string {
	var out []string
	for i := 0; i < len(s); {
		if s[i] == '@' {
			i++
			continue
		}
		if s[i] < '0' || s[i] > '9' {
			i++
			continue
		}
		j := i
		for j < len(s) && s[j] >= '0' && s[j] <= '9' {
			j++
		}
		n, _ := strconv.Atoi(s[i:j])
		k := j
		for k < len(s) && s[k] == ' ' {
			k++
		}
		if n > 0 && k+n <= len(s) {
			out = append(out, s[k:k+n])
			i = k + n
			continue
		}
		i = j
	}
	return out
}

// decodeSATLines applies AutoCAD's reversible SAT text encoding. It mirrors
// the encoding used by ACIS-aware DXF writers.
func decodeSATLines(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, decodeSATLine(line))
	}
	return out
}

func decodeSATLine(text string) string {
	const (
		space = 0x20
		under = 0x40
		at    = 0x5F
	)
	var b strings.Builder
	skip := false
	for i := 0; i < len(text); i++ {
		c := text[i]
		if skip {
			skip = false
			continue
		}
		switch {
		case c == space:
			b.WriteByte(' ')
		case c == under:
			b.WriteByte('_')
		case c == at:
			b.WriteByte('@')
		case c >= 0x41 && c <= 0x5E:
			b.WriteByte(byte(0x41 + (0x5E - c)))
			skip = c == 0x5E
		default:
			b.WriteByte(c ^ 0x5F)
		}
	}
	return b.String()
}
