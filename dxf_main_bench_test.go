package dxf

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/edanko/dxf/drawing"
)

// benchFixtures are representative ASCII DXF documents for end-to-end read and
// write measurement, from a minimal document through tag-dense drawings.
var benchFixtures = []string{
	"Minimal_DXF_AC1009.dxf",
	"arc.dxf",
	"hatches_2.dxf",
	"uncommon.dxf",
}

// sinkDrawing keeps the parsed document observable across iterations.
var sinkDrawing *drawing.Drawing

func readBenchFixture(b *testing.B, name string) []byte {
	b.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		b.Fatalf("read fixture %s: %v", name, err)
	}
	return data
}

// BenchmarkFromReader measures a full parse: tag scanning, entity binding and
// document assembly.
func BenchmarkFromReader(b *testing.B) {
	for _, name := range benchFixtures {
		data := readBenchFixture(b, name)
		b.Run(name, func(b *testing.B) {
			b.SetBytes(int64(len(data)))
			b.ReportAllocs()
			for b.Loop() {
				d, err := FromReader(bytes.NewReader(data))
				if err != nil {
					b.Fatal(err)
				}
				sinkDrawing = d
			}
		})
	}
}

// BenchmarkWriteTo measures serialization of an already-parsed document.
func BenchmarkWriteTo(b *testing.B) {
	for _, name := range benchFixtures {
		data := readBenchFixture(b, name)
		d, err := FromReader(bytes.NewReader(data))
		if err != nil {
			b.Fatalf("parse fixture %s: %v", name, err)
		}
		b.Run(name, func(b *testing.B) {
			b.SetBytes(int64(len(data)))
			b.ReportAllocs()
			for b.Loop() {
				n, err := d.WriteTo(io.Discard)
				if err != nil {
					b.Fatal(err)
				}
				if n == 0 {
					b.Fatal("wrote no bytes")
				}
			}
		})
	}
}
