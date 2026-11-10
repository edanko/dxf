// Package iterdxf iterates the records of an ASCII DXF stream with bounded
// memory: one record is materialized at a time, in file order.
package iterdxf

import (
	"io"

	"github.com/edanko/dxf/addons"
	"github.com/edanko/dxf/lldxf"
)

type capability struct{}

func (capability) Name() string        { return "iterdxf" }
func (capability) Description() string { return "bounded-memory DXF record iterator" }

func init() { _ = addons.Register("iterdxf", capability{}) }

// IterDXF streams records from an ASCII DXF source.
type IterDXF struct {
	r io.Reader
}

// New creates a streaming reader over an ASCII DXF stream.
func New(r io.Reader) *IterDXF {
	return &IterDXF{r: r}
}

// Entities calls yield for every code-0-delimited record in file order.
// Returning false from yield stops the iteration early. Only one record is
// held in memory at a time.
func (it *IterDXF) Entities(yield func(lldxf.Tags) bool) error {
	tagger := lldxf.NewASCIITagger(it.r)
	cur := lldxf.NewTags()
	for tagger.Next() {
		t := tagger.Tag()
		if t.Code() == 0 && len(cur) > 0 {
			if !yield(cur) {
				return nil
			}
			cur = lldxf.NewTags()
		}
		cur = cur.Add(t)
	}
	if len(cur) > 0 {
		if !yield(cur) {
			return nil
		}
	}
	return tagger.Err()
}

// EntityTypes calls yield only for records whose code-0 type matches one of
// the wanted type names.
func (it *IterDXF) EntityTypes(wanted map[string]bool, yield func(lldxf.Tags) bool) error {
	return it.Entities(func(tags lldxf.Tags) bool {
		if len(tags) == 0 || tags[0].Code() != 0 {
			return true
		}
		name, _ := tags[0].Value().(string)
		if wanted[name] {
			return yield(tags)
		}
		return true
	})
}
