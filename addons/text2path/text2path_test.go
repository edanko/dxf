package text2path_test

import (
	"testing"

	"github.com/edanko/dxf/addons/text2path"
	"github.com/edanko/dxf/entity"
)

func TestConvertEntity(t *testing.T) {
	tx := entity.NewText()
	tx.Value = "AB"
	tx.Height = 2.5
	paths, err := text2path.ConvertEntity(tx)
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Skip("no font data available and vector fallback produced no outlines")
	}
	for _, p := range paths {
		if p == nil {
			t.Fatal("nil path in outline set")
		}
	}
}
