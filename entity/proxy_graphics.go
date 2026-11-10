// Proxy graphics extraction: decode binary tag payloads embedded in
// ACAD_PROXY_ENTITY records into real entities.
package entity

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/edanko/dxf/lldxf"
)

// ExtractProxyGraphics decodes the entity data (code 93/310) of a proxy
// entity into entities via the registered loaders. Handles decoded from the
// payload are cleared so the caller can assign fresh ones. When entity data
// is empty, the graphics data (code 92/310) is decoded as a fallback.
//
// The payload uses the binary DXF tag encoding: 2-byte little-endian group
// codes, NUL-terminated strings for string codes, little-endian doubles for
// real codes, little-endian int16/int32 for integer codes.
func ExtractProxyGraphics(e Entity) ([]Entity, error) {
	p, ok := e.(*AcadProxyEntity)
	if !ok {
		return nil, fmt.Errorf("proxy graphics: %s is not an ACAD_PROXY_ENTITY", e.DXFType())
	}
	payload := p.EntityData
	source := "entity data"
	if len(payload) == 0 {
		payload = p.GraphicsData
		source = "graphics data"
	}
	if len(payload) == 0 {
		return nil, fmt.Errorf("proxy graphics: proxy entity %s carries no proxy graphics", e.Handle())
	}
	tags, err := DecodeBinaryTags(payload)
	if err != nil {
		return nil, fmt.Errorf("proxy graphics: %s of %s: %w", source, e.Handle(), err)
	}
	var out []Entity
	for _, rec := range splitTagRecords(tags) {
		if len(rec) == 0 || rec[0].Code != 0 {
			continue
		}
		typeName, _ := rec[0].Value.(string)
		if typeName == "" {
			continue
		}
		tags := lldxf.NewTags()
		for _, t := range rec {
			tags = tags.Add(lldxf.NewDXFTag(lldxf.GroupCode(t.Code), t.Value))
		}
		ent, err := LoadEntity(typeName, tags)
		if err != nil {
			return nil, fmt.Errorf("proxy graphics: record %s: %w", typeName, err)
		}
		if s, ok := ent.(HandleStringSetter); ok {
			s.SetHandleString("")
		}
		out = append(out, ent)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("proxy graphics: %s of %s contains no entity records", source, e.Handle())
	}
	return out, nil
}

// BinaryTag is one decoded (group code, value) pair from a binary payload.
type BinaryTag struct {
	Code  int
	Value any
}

// DecodeBinaryTags decodes a binary DXF tag stream.
func DecodeBinaryTags(data []byte) ([]BinaryTag, error) {
	var out []BinaryTag
	pos := 0
	for pos < len(data) {
		if pos+2 > len(data) {
			return nil, fmt.Errorf("unexpected end of payload reading group code")
		}
		code := int(int16(binary.LittleEndian.Uint16(data[pos : pos+2])))
		pos += 2
		var value any
		switch {
		case code <= 9:
			s, n, err := readCString(data, pos)
			if err != nil {
				return nil, fmt.Errorf("code %d: %w", code, err)
			}
			value, pos = s, n
		case code <= 59, code >= 110 && code <= 149, code >= 210 && code <= 239:
			if pos+8 > len(data) {
				return nil, fmt.Errorf("code %d: unexpected end of payload reading float", code)
			}
			value = math.Float64frombits(binary.LittleEndian.Uint64(data[pos : pos+8]))
			pos += 8
		case code >= 60 && code <= 79, code >= 170 && code <= 179, code >= 270 && code <= 299,
			code >= 370 && code <= 389, code >= 400 && code <= 409, code >= 460 && code <= 469:
			if pos+2 > len(data) {
				return nil, fmt.Errorf("code %d: unexpected end of payload reading int16", code)
			}
			value = int(int16(binary.LittleEndian.Uint16(data[pos : pos+2])))
			pos += 2
		case code >= 90 && code <= 99, code == 420, code >= 450 && code <= 459, code >= 470 && code <= 479:
			if pos+4 > len(data) {
				return nil, fmt.Errorf("code %d: unexpected end of payload reading int32", code)
			}
			value = int(int32(binary.LittleEndian.Uint32(data[pos : pos+4])))
			pos += 4
		case code >= 310 && code <= 319:
			if pos+1 > len(data) {
				return nil, fmt.Errorf("code %d: unexpected end of payload reading binary length", code)
			}
			n := int(data[pos])
			pos++
			if pos+n > len(data) {
				return nil, fmt.Errorf("code %d: unexpected end of payload reading binary data", code)
			}
			value = append([]byte(nil), data[pos:pos+n]...)
			pos += n
		default:
			// String-typed code ranges (100-109, 300-309, 320-369, 390-419,
			// 430, 440-449 int16 handled above, 480+, 999+).
			if isStringCode(code) {
				s, n, err := readCString(data, pos)
				if err != nil {
					return nil, fmt.Errorf("code %d: %w", code, err)
				}
				value, pos = s, n
			} else {
				return nil, fmt.Errorf("code %d: unsupported group code in binary payload", code)
			}
		}
		out = append(out, BinaryTag{Code: code, Value: value})
	}
	return out, nil
}

func isStringCode(code int) bool {
	switch {
	case code >= 100 && code <= 109,
		code >= 300 && code <= 309,
		code >= 320 && code <= 329,
		code >= 330 && code <= 369,
		code == 390 || code == 391 || code == 392 || code == 393,
		code == 410 || code == 411 || code == 412 || code == 413 || code == 419,
		code == 430,
		code >= 480 && code <= 481,
		code >= 999:
		return true
	}
	return false
}

func readCString(data []byte, pos int) (string, int, error) {
	start := pos
	for pos < len(data) && data[pos] != 0 {
		pos++
	}
	if pos >= len(data) {
		return "", 0, fmt.Errorf("unterminated string value")
	}
	return string(data[start:pos]), pos + 1, nil
}

// splitTagRecords groups tags into records starting at each code 0 tag.
func splitTagRecords(tags []BinaryTag) [][]BinaryTag {
	var out [][]BinaryTag
	var cur []BinaryTag
	for _, t := range tags {
		if t.Code == 0 && len(cur) > 0 {
			out = append(out, cur)
			cur = nil
		}
		cur = append(cur, t)
	}
	if len(cur) > 0 {
		out = append(out, cur)
	}
	return out
}

// EncodeBinaryTags encodes tags with the binary DXF conventions used by
// DecodeBinaryTags (exported for tests and producers of proxy payloads).
func EncodeBinaryTags(tags []BinaryTag) []byte {
	var out []byte
	for _, t := range tags {
		var code [2]byte
		binary.LittleEndian.PutUint16(code[:], uint16(int16(t.Code)))
		out = append(out, code[:]...)
		switch {
		case t.Code <= 9:
			out = append(out, []byte(toString(t.Value))...)
			out = append(out, 0)
		case t.Code <= 59, t.Code >= 110 && t.Code <= 149, t.Code >= 210 && t.Code <= 239:
			var b [8]byte
			binary.LittleEndian.PutUint64(b[:], math.Float64bits(toFloat(t.Value)))
			out = append(out, b[:]...)
		case t.Code >= 60 && t.Code <= 79, t.Code >= 170 && t.Code <= 179,
			t.Code >= 270 && t.Code <= 299, t.Code >= 370 && t.Code <= 389,
			t.Code >= 400 && t.Code <= 409, t.Code >= 460 && t.Code <= 469:
			var b [2]byte
			binary.LittleEndian.PutUint16(b[:], uint16(int16(toInt(t.Value))))
			out = append(out, b[:]...)
		case t.Code >= 90 && t.Code <= 99, t.Code == 420, t.Code >= 450 && t.Code <= 459, t.Code >= 470 && t.Code <= 479:
			var b [4]byte
			binary.LittleEndian.PutUint32(b[:], uint32(int32(toInt(t.Value))))
			out = append(out, b[:]...)
		case t.Code >= 310 && t.Code <= 319:
			raw, _ := t.Value.([]byte)
			out = append(out, byte(len(raw)))
			out = append(out, raw...)
		default:
			out = append(out, []byte(toString(t.Value))...)
			out = append(out, 0)
		}
	}
	return out
}

func toString(v any) string {
	s, _ := v.(string)
	return s
}

func toFloat(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case float32:
		return float64(x)
	case int:
		return float64(x)
	case int32:
		return float64(x)
	case int64:
		return float64(x)
	}
	return 0
}

func toInt(v any) int {
	switch x := v.(type) {
	case int:
		return x
	case int32:
		return int(x)
	case int64:
		return int(x)
	case float64:
		return int(x)
	}
	return 0
}
