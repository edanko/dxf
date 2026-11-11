// Binary records: arbitrary byte payloads stored in XRECORD tags.
package entity

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// MaxBinaryRecordChunk is the maximum number of payload bytes encoded in one
// code-310 tag. DXF limits the hex string to 254 characters (127 bytes).
const MaxBinaryRecordChunk = 127

const (
	binaryRecordOpen  = "{BINARY"
	binaryRecordClose = "}"
)

// BinaryRecord is a named arbitrary byte payload.
type BinaryRecord struct {
	Name string
	Data []byte
	// Extra holds XRECORD tags outside the binary-record block.
	Extra []XRecordTag
}

// Compile encodes the record as XRECORD tags. Payload bytes are hex encoded
// in code-310 chunks so every byte value round-trips.
func (r *BinaryRecord) Compile() ([]XRecordTag, error) {
	if err := validateUserString("record name", r.Name); err != nil {
		return nil, err
	}
	tags := []XRecordTag{{Code: 302, Value: binaryRecordOpen}, {Code: 1, Value: r.Name}}
	for start := 0; start < len(r.Data); start += MaxBinaryRecordChunk {
		end := start + MaxBinaryRecordChunk
		if end > len(r.Data) {
			end = len(r.Data)
		}
		tags = append(tags, XRecordTag{Code: 310, Value: strings.ToUpper(hex.EncodeToString(r.Data[start:end]))})
	}
	tags = append(tags, XRecordTag{Code: 302, Value: binaryRecordClose})
	return tags, nil
}

// ApplyTo replaces the binary-record block in x, leaving unrelated tags
// intact.
func (r *BinaryRecord) ApplyTo(x *XRecord) error {
	if x == nil {
		return fmt.Errorf("binary record: nil XRECORD")
	}
	compiled, err := r.Compile()
	if err != nil {
		return err
	}
	kept := stripTagBlock(x.Tags, binaryRecordOpen, binaryRecordClose)
	x.Tags = append(kept, compiled...)
	return nil
}

// ParseBinaryRecord decodes the binary-record block from XRECORD tags.
func ParseBinaryRecord(tags []XRecordTag) (*BinaryRecord, error) {
	start := -1
	for i, tag := range tags {
		if tag.Code == 302 && stringValue(tag.Value) == binaryRecordOpen {
			start = i
			break
		}
	}
	if start < 0 {
		return nil, fmt.Errorf("binary record: missing %q marker", binaryRecordOpen)
	}
	if start+1 >= len(tags) || tags[start+1].Code != 1 {
		return nil, fmt.Errorf("binary record: missing record name")
	}
	rec := &BinaryRecord{Name: stringValue(tags[start+1].Value)}
	end := -1
	for i := start + 2; i < len(tags); i++ {
		tag := tags[i]
		if tag.Code == 302 && stringValue(tag.Value) == binaryRecordClose {
			end = i
			break
		}
		if tag.Code != 310 {
			return nil, fmt.Errorf("binary record: unexpected code %d", tag.Code)
		}
		chunk, err := hex.DecodeString(strings.TrimSpace(stringValue(tag.Value)))
		if err != nil {
			return nil, fmt.Errorf("binary record: invalid hex chunk: %w", err)
		}
		rec.Data = append(rec.Data, chunk...)
	}
	if end < 0 {
		return nil, fmt.Errorf("binary record: missing %q terminator", binaryRecordClose)
	}
	rec.Extra = append(append([]XRecordTag(nil), tags[:start]...), tags[end+1:]...)
	return rec, nil
}

func stripTagBlock(tags []XRecordTag, open, closeMarker string) []XRecordTag {
	start := -1
	for i, tag := range tags {
		if tag.Code == 302 && stringValue(tag.Value) == open {
			start = i
			break
		}
	}
	if start < 0 {
		return append([]XRecordTag(nil), tags...)
	}
	depth := 0
	for i := start; i < len(tags); i++ {
		if tags[i].Code != 302 {
			continue
		}
		switch stringValue(tags[i].Value) {
		case open:
			depth++
		case closeMarker:
			depth--
			if depth == 0 {
				return append(append([]XRecordTag(nil), tags[:start]...), tags[i+1:]...)
			}
		}
	}
	return append([]XRecordTag(nil), tags[:start]...)
}
