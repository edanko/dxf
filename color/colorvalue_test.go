package color

import "testing"

func TestACITableSamples(t *testing.T) {
	if rgb := IndexColor(1); rgb[0] != 255 || rgb[1] != 0 || rgb[2] != 0 {
		t.Fatalf("ACI 1 must be red, got %v", rgb)
	}
	v := NewACI(1)
	if v.KindOf() != KindACI {
		t.Fatalf("kind: %v", v.KindOf())
	}
	r, g, b, ok := v.RGB8()
	if !ok || r != 255 || g != 0 || b != 0 {
		t.Fatalf("ACI 1 RGB: %v %v %v %v", r, g, b, ok)
	}
}

func TestNearestIndex(t *testing.T) {
	if got := NearestIndex(255, 0, 0); got != 1 {
		t.Fatalf("nearest to red: %d", got)
	}
	if got := NearestIndex(0, 0, 255); got != 5 {
		t.Fatalf("nearest to blue: %d", got)
	}
}

func TestColorKinds(t *testing.T) {
	if !NewByLayer().IsByLayer() || NewByLayer().KindOf() != KindByLayer {
		t.Fatal("BYLAYER kind")
	}
	if !NewByBlock().IsByBlock() || NewByBlock().KindOf() != KindByBlock {
		t.Fatal("BYBLOCK kind")
	}
	tc := NewTrueColor(0x112233)
	r, g, b, ok := tc.RGB8()
	if !ok || r != 0x11 || g != 0x22 || b != 0x33 {
		t.Fatalf("true color: %v %v %v %v", r, g, b, ok)
	}
	if _, _, _, ok := NewByLayer().RGB8(); ok {
		t.Fatal("BYLAYER must not carry concrete RGB")
	}
}

func TestValidateACI(t *testing.T) {
	for _, ok := range []ColorNumber{0, 1, 100, 255} {
		if err := ValidateACI(ok); err != nil {
			t.Fatalf("ACI %d must be valid: %v", ok, err)
		}
	}
	for _, bad := range []ColorNumber{-1, 256, 999} {
		if err := ValidateACI(bad); err == nil {
			t.Fatalf("ACI %d must be invalid", bad)
		}
	}
}

func TestTransparencyValidation(t *testing.T) {
	if _, err := NewTransparency(0.5); err != nil {
		t.Fatal(err)
	}
	if _, err := NewTransparency(-0.1); err == nil {
		t.Fatal("negative alpha must error")
	}
	if _, err := NewTransparency(1.1); err == nil {
		t.Fatal("alpha > 1 must error")
	}
}
