package dxf

import (
	"testing"

	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
)

// TestDrawingDeepCopy verifies DeepCopy returns an independent drawing.
// It lives in the root package because the document loader used by DeepCopy
// is registered by the root package's init.
func TestDrawingDeepCopy(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	c := entity.NewCircle()
	c.Radius = 5
	c.SetLayerName("SRC")
	if err := d.Model().Add(c); err != nil {
		t.Fatalf("add circle: %v", err)
	}

	copied, err := d.DeepCopy()
	if err != nil {
		t.Fatalf("DeepCopy: %v", err)
	}
	if copied == d {
		t.Fatal("DeepCopy returned the same drawing")
	}

	// Mutating the copy must not affect the source.
	extra := entity.NewCircle()
	extra.Radius = 9
	if err := copied.Model().Add(extra); err != nil {
		t.Fatalf("add circle to copy: %v", err)
	}
	if n, m := len(d.ModelSpace()), len(copied.ModelSpace()); m != n+1 {
		t.Errorf("copy has %d model entities, source has %d", m, n)
	}

	srcCircle := d.ModelSpace()[0]
	copyCircle := copied.ModelSpace()[0]
	if srcCircle == copyCircle {
		t.Error("copied entities share the same object as the source")
	}
}
