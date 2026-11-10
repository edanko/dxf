// Tests for the add-on framework and the streaming I/O add-ons.
package addons_test

import (
	"bytes"
	"fmt"
	"testing"

	dxf "github.com/edanko/dxf"
	"github.com/edanko/dxf/addons"
	"github.com/edanko/dxf/addons/iterdxf"
	"github.com/edanko/dxf/addons/r12writer"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/lldxf"
)

func TestRegistryDiscovery(t *testing.T) {
	c, ok := addons.Lookup("r12writer")
	if !ok {
		t.Fatal("r12writer capability not registered")
	}
	if c.Name() != "r12writer" {
		t.Fatalf("name: %s", c.Name())
	}
	if c.Description() == "" {
		t.Fatal("description empty")
	}
	if _, ok := addons.Lookup("iterdxf"); !ok {
		t.Fatal("iterdxf capability not registered")
	}
	if all := addons.Capabilities(); len(all) < 2 {
		t.Fatalf("capabilities: %d", len(all))
	}
	if err := addons.Register("r12writer", nil); err == nil {
		t.Fatal("duplicate registration must error")
	}
}

func makeLine(i int) *entity.Line {
	l := entity.NewLine()
	l.SetHandleString(fmt.Sprintf("%X", i+1))
	l.Start = []float64{float64(i), 0, 0}
	l.End = []float64{float64(i) + 1, 1, 0}
	return l
}

func TestR12WriterStreaming(t *testing.T) {
	var buf bytes.Buffer
	w := r12writer.New(&buf)
	if err := w.WriteHeader(); err != nil {
		t.Fatal(err)
	}
	const n = 10000
	for i := range n {
		if err := w.WriteEntity(makeLine(i)); err != nil {
			t.Fatalf("entity %d: %v", i, err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	doc, err := dxf.FromStringData(buf.String())
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	lines := 0
	for _, e := range doc.Entities() {
		if e.DXFType() == "LINE" {
			lines++
		}
	}
	if lines != n {
		t.Fatalf("lines: %d want %d", lines, n)
	}
}

func TestIterDXFStreaming(t *testing.T) {
	var buf bytes.Buffer
	w := r12writer.New(&buf)
	if err := w.WriteHeader(); err != nil {
		t.Fatal(err)
	}
	const n = 500
	for i := range n {
		if err := w.WriteEntity(makeLine(i)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	it := iterdxf.New(bytes.NewReader(buf.Bytes()))
	count := 0
	last := 0
	err := it.EntityTypes(map[string]bool{"LINE": true}, func(tags lldxf.Tags) bool {
		count++
		var v int
		if _, err := fmt.Sscanf(tags.GetHandle(), "%X", &v); err != nil {
			t.Errorf("handle %q: %v", tags.GetHandle(), err)
			return false
		}
		if v != last+1 {
			t.Errorf("order: got %X want %X", v, last+1)
			return false
		}
		last = v
		return true
	})
	if err != nil {
		t.Fatal(err)
	}
	if count != n {
		t.Fatalf("iterated: %d want %d", count, n)
	}
}
