// Package text2path converts TEXT entities to path outlines using real font
// data (TrueType/OpenType) with a built-in vector-font fallback.
package text2path

import (
	"github.com/edanko/dxf/addons"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/path"
	t2p "github.com/edanko/dxf/text2path"
)

type capability struct{}

func (capability) Name() string        { return "text2path" }
func (capability) Description() string { return "TEXT entity to path outlines via real font data" }

func init() { _ = addons.Register("text2path", capability{}) }

// ConvertEntity converts a TEXT entity to path outlines. It prefers real
// font data (TrueType/OpenType) and falls back to the built-in vector font.
func ConvertEntity(t *entity.Text) ([]*path.Path, error) {
	if t == nil {
		return nil, nil
	}
	if paths, _, err := t2p.TextToPathsTTF(t.Value, t.Height); err == nil && len(paths) > 0 {
		return paths, nil
	}
	if p := t2p.NewTextToPath(t.Height / 120.0).Convert(t.Value); p != nil {
		return []*path.Path{p}, nil
	}
	return nil, nil
}
