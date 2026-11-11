package query

import (
	"testing"

	"github.com/edanko/dxf/entity"
)

type sliceSource []entity.Entity

func (s sliceSource) GetEntity(h string) (entity.Entity, bool) {
	for _, e := range s {
		if e.Handle() == h {
			return e, true
		}
	}
	return nil, false
}

// TestEmptySelectionBounds verifies that a selection made only of entities
// without drawable geometry reports no bounds rather than a zero-origin box.
func TestEmptySelectionBounds(t *testing.T) {
	placeholder := entity.NewPlaceholder()
	placeholder.SetHandleString("PH1")
	src := sliceSource{placeholder}
	selection := NewSelection("PH1")
	_, _, ok := selection.Bounds(src)
	if ok {
		t.Fatal("empty geometry selection reported bounds")
	}
}
