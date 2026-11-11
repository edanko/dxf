package enums

import "testing"

func TestEnumRoundTrip(t *testing.T) {
	if name, ok := TextAlign.Name(2); !ok || name != "center" {
		t.Fatalf("TextAlign.Name(2): %q %v", name, ok)
	}
	if n, ok := TextAlign.Number("middle"); !ok || n != 5 {
		t.Fatalf("TextAlign.Number(middle): %d %v", n, ok)
	}
	if !TextAlign.Valid(6) || TextAlign.Valid(99) {
		t.Fatal("TextAlign.Valid")
	}
	if got := TextAlign.Unknown(99); got == "" {
		t.Fatal("unknown value must be reported")
	}
	if got := TextAlign.Unknown(2); got != "" {
		t.Fatalf("valid value reported as unknown: %q", got)
	}
}

func TestMTextAttachmentConversion(t *testing.T) {
	if n, ok := MTextAttachment.Number("bottom-right"); !ok || n != 9 {
		t.Fatalf("bottom-right: %d %v", n, ok)
	}
	if name, ok := MTextAttachment.Name(1); !ok || name != "top-left" {
		t.Fatalf("1: %q %v", name, ok)
	}
}

func TestDefaultsAndValidation(t *testing.T) {
	if DefaultLineWeight != -3 || DefaultLinetype != "ByLayer" || DefaultTextStyle != "Standard" {
		t.Fatal("defaults")
	}
	for _, w := range []int{-3, 0, 13, 211} {
		if !ValidateLineWeight(w) {
			t.Fatalf("lineweight %d must be valid", w)
		}
	}
	for _, w := range []int{-4, 212} {
		if ValidateLineWeight(w) {
			t.Fatalf("lineweight %d must be invalid", w)
		}
	}
	if !ValidateLinetypeName("Continuous") || ValidateLinetypeName("") {
		t.Fatal("linetype validation")
	}
}
