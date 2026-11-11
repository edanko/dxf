package drawing

import (
	"strings"
	"testing"

	"github.com/edanko/dxf/entity"
)

// TestHashDrawingDeterministic verifies HashDrawing returns a non-empty,
// stable hash that changes with content.
func TestHashDrawingDeterministic(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	c := entity.NewCircle()
	c.Radius = 5
	if err := d.Model().Add(c); err != nil {
		t.Fatalf("add circle: %v", err)
	}
	h1, err := HashDrawing(d)
	if err != nil {
		t.Fatalf("HashDrawing: %v", err)
	}
	if h1 == "" {
		t.Fatal("hash is empty")
	}
	h2, err := HashDrawing(d)
	if err != nil {
		t.Fatalf("HashDrawing: %v", err)
	}
	if h1 != h2 {
		t.Errorf("hash not stable: %q vs %q", h1, h2)
	}
	if len(h1) != 64 || strings.Trim(h1, "0123456789abcdef") != "" {
		t.Errorf("hash is not a 64-char hex sha256: %q", h1)
	}

	// A different drawing hashes differently.
	d2, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	c2 := entity.NewCircle()
	c2.Radius = 9
	if err := d2.Model().Add(c2); err != nil {
		t.Fatalf("add circle: %v", err)
	}
	h3, err := HashDrawing(d2)
	if err != nil {
		t.Fatalf("HashDrawing: %v", err)
	}
	if h1 == h3 {
		t.Error("different drawings produced the same hash")
	}
}
