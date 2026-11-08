package dxf

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestCorpusRoundTrip(t *testing.T) {
	files, err := filepath.Glob("testdata/*.dxf")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no testdata fixtures")
	}
	for _, fn := range files {
		fn := fn
		t.Run(filepath.Base(fn), func(t *testing.T) {
			data, err := os.ReadFile(fn)
			if err != nil {
				t.Fatal(err)
			}
			first, err := FromStringData(string(data))
			if err != nil {
				t.Fatalf("load: %v", err)
			}
			var out bytes.Buffer
			if _, err := first.WriteTo(&out); err != nil {
				t.Fatalf("save: %v", err)
			}
			second, err := FromStringData(out.String())
			if err != nil {
				t.Fatalf("reload: %v", err)
			}
			if len(first.Entities()) != len(second.Entities()) {
				t.Fatalf("entity count %d -> %d", len(first.Entities()), len(second.Entities()))
			}
			if len(first.Blocks()) != len(second.Blocks()) {
				t.Fatalf("block count %d -> %d", len(first.Blocks()), len(second.Blocks()))
			}
			if len(first.Objects()) != len(second.Objects()) {
				t.Fatalf("object count %d -> %d", len(first.Objects()), len(second.Objects()))
			}
		})
	}
}
