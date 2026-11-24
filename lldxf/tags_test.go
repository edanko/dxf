package lldxf

import (
	"bytes"
	"fmt"
	"testing"
)

// The tables below are the original per-code lists that the classification
// predicates used to build on every call. They are kept verbatim as reference
// data so tags_classcode_test.go can prove the range-based implementations
// accept exactly the same group codes.
var (
	refBinaryDataCodes = []GroupCode{260, 261, 310, 311, 312, 313, 314, 315, 316, 317, 318, 319, 1004}

	refPointCodes = []GroupCode{10, 11, 12, 13, 14, 15, 16, 17, 18, 110, 111, 112, 210, 211, 212, 213, 1010, 1011, 1012, 1013}

	refFloatCodes = []GroupCode{
		10, 11, 12, 13, 14, 15, 16, 17, 18, 19, // X coordinates
		20, 21, 22, 23, 24, 25, 26, 27, 28, 29, // Y coordinates
		30, 31, 32, 33, 34, 35, 36, 37, 38, 39, // Z coordinates
		40, 41, 42, 43, 44, 45, 46, 47, 48, 49, // Radii, angles, etc.
		50, 51, 52, 53, 54, 55, 56, 57, 58, 59,
		110, 111, 112, 113, 114, 115, 116, 117, 118, 119, // More X coords
		120, 121, 122, 123, 124, 125, 126, 127, 128, 129, // More Y coords
		130, 131, 132, 133, 134, 135, 136, 137, 138, 139, // More Z coords
		140, 141, 142, 143, 144, 145, 146, 147, 148, 149, // Dimension floats
		210, 211, 212, 213, 214, 215, 216, 217, 218, 219, // Extrusion vectors
		460, 461, 462, 463, 464, 465, 466, 467, 468, 469, // More floats
		1010, 1011, 1012, 1013, // XDATA floats
		1040, 1041, 1042, // XDATA floats
	}

	refInt16Codes = []GroupCode{
		60, 61, 62, 63, 64, 65, 66, 67, 68, 69, // Status, color, etc.
		70, 71, 72, 73, 74, 75, 76, 77, 78, 79, // Flags, counts
		170, 171, 172, 173, 174, 175, 176, 177, 178, 179, // More flags
		270, 271, 272, 273, 274, 275, 276, 277, 278, 279, // More flags
		280, 281, 282, 283, 284, 285, 286, 287, 288, 289, // Booleans, small ints
		290, 291, 292, 293, 294, 295, 296, 297, 298, 299, // Booleans
		370, 371, 372, 373, 374, 375, 376, 377, 378, 379, // More flags
		380, 381, 382, 383, 384, 385, 386, 387, 388, 389, // More small ints
		400, 401, 402, 403, 404, 405, 406, 407, 408, 409, // More flags
		1060, 1061, 1062, 1063, 1064, 1065, 1066, 1067, 1068, 1069, // XDATA ints
		1070, // XDATA int
	}

	refInt32Codes = []GroupCode{
		90, 91, 92, 93, 94, 95, 96, 97, 98, 99, // Counts, handles
		420, 421, 422, 423, 424, 425, 426, 427, 428, 429, // Annotation related
		430, 431, 432, 433, 434, 435, 436, 437, 438, 439, // Dictionary handles
		440, 441, 442, 443, 444, 445, 446, 447, 448, 449, // More ints
		450, 451, 452, 453, 454, 455, 456, 457, 458, 459, // More ints
		480, 481, 482, 483, 484, 485, 486, 487, 488, 489, // More handles
		1071, // XDATA int32
	}

	refInt64Codes = []GroupCode{
		160, 161, 162, 163, 164, 165, 166, 167, 168, 169, // Lengths, counts
	}
)

// TestCodeClassificationMatchesReferenceTables proves the range-based
// classification predicates accept exactly the codes the original literal
// tables listed, across and beyond the full group-code space.
func TestCodeClassificationMatchesReferenceTables(t *testing.T) {
	cases := []struct {
		name string
		ref  []GroupCode
		fn   func(GroupCode) bool
	}{
		{"binary", refBinaryDataCodes, isBinaryDataCode},
		{"point", refPointCodes, isPointCode},
		{"float", refFloatCodes, isFloatCode},
		{"int16", refInt16Codes, isInt16Code},
		{"int32", refInt32Codes, isInt32Code},
		{"int64", refInt64Codes, isInt64Code},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			want := make(map[GroupCode]bool, len(tc.ref))
			for _, code := range tc.ref {
				if want[code] {
					t.Fatalf("reference table for %q lists code %d twice", tc.name, code)
				}
				want[code] = true
			}
			for code := GroupCode(-300); code <= 1400; code++ {
				if got, expected := tc.fn(code), want[code]; got != expected {
					t.Errorf("%s(%d) = %v, want %v", tc.name, code, got, expected)
				}
			}
		})
	}
}

// TestBinaryBoolCodeRange pins the binary boolean group-code range.
func TestBinaryBoolCodeRange(t *testing.T) {
	for code := GroupCode(-300); code <= 1400; code++ {
		want := code >= 290 && code <= 299
		if got := IsBinaryBoolCode(code); got != want {
			t.Errorf("IsBinaryBoolCode(%d) = %v, want %v", code, got, want)
		}
	}
}

// TestFormatValueMatchesVerbFormatting proves FormatValue renders non-binary
// values exactly as fmt's %v verb does, which is the canonical text form the
// string model and ASCII output depend on. Values cover every type the tag
// casters produce.
func TestFormatValueMatchesVerbFormatting(t *testing.T) {
	values := []any{
		"",
		"LINE",
		"a\tb",
		float64(0),
		float64(-0.5),
		float64(1.5),
		float64(1e300),
		float64(1e-300),
		float64(1234567.0),
		float64(1.0 / 3.0),
		float32(1.25),
		float32(1e30),
		int(0),
		int(-42),
		int16(7),
		int32(-8),
		int64(1 << 40),
		true,
		false,
		[]float64{1, 2, 3},
		nil,
	}
	for _, v := range values {
		tag := NewDXFTag(1, v)
		want := fmt.Sprintf("%v", v)
		if got := FormatValue(tag); got != want {
			t.Errorf("FormatValue(%T %v) = %q, want %q", v, v, got, want)
		}
	}
}

// TestFormatValueBinaryHex pins the uppercase-hex rendering of binary payloads
// and that non-binary []byte values keep their %v form.
func TestFormatValueBinaryHex(t *testing.T) {
	for _, code := range []GroupCode{260, 261, 310, 319, 1004} {
		tag := BinaryTagFromHex(code, "deadbeef")
		if got, want := FormatValue(tag), "DEADBEEF"; got != want {
			t.Errorf("FormatValue(binary %d) = %q, want %q", code, got, want)
		}
	}
	// A non-binary code holding raw bytes keeps %v's rendering; build the tag
	// directly because NewDXFTag casts []byte to string for such codes.
	nonBinary := &DXFTag{code: 1, value: []byte("ab")}
	if got, want := FormatValue(nonBinary), fmt.Sprintf("%v", []byte("ab")); got != want {
		t.Errorf("FormatValue(non-binary []byte) = %q, want %q", got, want)
	}
}

// TestDecodeHexPayload pins ASCII binary payload decoding, including the
// whitespace tolerance and the failure cases that fall back to raw bytes.
func TestDecodeHexPayload(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  []byte
	}{
		{"plain", "DEADBEEF", []byte{0xDE, 0xAD, 0xBE, 0xEF}},
		{"lowercase", "deadbeef", []byte{0xDE, 0xAD, 0xBE, 0xEF}},
		{"spaced", "DE AD\tBE EF", []byte{0xDE, 0xAD, 0xBE, 0xEF}},
		{"single pixel", "00", []byte{0x00}},
		{"empty", "", []byte{}},
		{"odd digits", "ABC", nil},
		{"non-hex", "GG", nil},
		{"non-hex after digits", "ABGG", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := decodeHexPayload(tc.input)
			if tc.want == nil {
				if got != nil {
					t.Fatalf("decodeHexPayload(%q) = %v, want nil", tc.input, got)
				}
				return
			}
			if !bytes.Equal(got, tc.want) {
				t.Fatalf("decodeHexPayload(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}
