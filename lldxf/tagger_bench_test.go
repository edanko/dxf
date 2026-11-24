package lldxf

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// benchFixtures are representative ASCII DXF inputs, from a minimal document
// through tag-dense drawings, so a change cannot look good only on tiny files.
var benchFixtures = []string{
	"Minimal_DXF_AC1009.dxf",
	"arc.dxf",
	"hatches_2.dxf",
	"uncommon.dxf",
}

// Sinks keep benchmark results observable so the compiler cannot eliminate the
// measured work.
var (
	sinkTags  Tags
	sinkAny   any
	sinkFloat float64
)

// benchAccum folds a tag value into a float without boxing it into an
// interface, so allocations reported by the benchmarks belong to the code
// under measurement rather than to the sink itself.
func benchAccum(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case float32:
		return float64(x)
	case int16:
		return float64(x)
	case int32:
		return float64(x)
	case int64:
		return float64(x)
	case int:
		return float64(x)
	case bool:
		if x {
			return 1
		}
		return 0
	case string:
		return float64(len(x))
	case []byte:
		return float64(len(x))
	}
	return 0
}

func readBenchFixture(b *testing.B, name string) []byte {
	b.Helper()
	data, err := os.ReadFile(filepath.Join("..", "testdata", name))
	if err != nil {
		b.Fatalf("read fixture %s: %v", name, err)
	}
	return data
}

// BenchmarkLoadTags scans an in-memory ASCII DXF stream into tags: the shared
// front end of every read path.
func BenchmarkLoadTags(b *testing.B) {
	for _, name := range benchFixtures {
		data := readBenchFixture(b, name)
		b.Run(name, func(b *testing.B) {
			b.SetBytes(int64(len(data)))
			b.ReportAllocs()
			for b.Loop() {
				tags, err := LoadTags(bytes.NewReader(data))
				if err != nil {
					b.Fatal(err)
				}
				sinkTags = tags
			}
		})
	}
}

// BenchmarkTaggerNext isolates line scanning and group-code parsing from tag
// value casting, so a regression in either half is attributable.
func BenchmarkTaggerNext(b *testing.B) {
	data := readBenchFixture(b, "uncommon.dxf")
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	for b.Loop() {
		tagger := NewASCIITagger(bytes.NewReader(data))
		n := 0
		for tagger.Next() {
			_ = tagger.Tag().Code()
			n++
		}
		if err := tagger.Err(); err != nil {
			b.Fatal(err)
		}
		sinkAny = n
	}
}

// benchCastSamples feeds each classification family the value shape a real DXF
// carries for that group code, so parse failures never inflate the measurement.
var benchCastSamples = []struct {
	code  GroupCode
	value string
}{
	{0, "LINE"},
	{2, "ENTITIES"},
	{5, "2F"},
	{10, "1.5"},
	{40, "0.0"},
	{70, "1"},
	{90, "3"},
	{160, "42"},
	{210, "0.0"},
	{310, "DEADBEEF"},
	{1004, "0102ABCD"},
}

// BenchmarkCastTagValue measures group-code classification and value casting,
// which every scanned tag performs once.
func BenchmarkCastTagValue(b *testing.B) {
	b.ReportAllocs()
	total := 0.0
	for b.Loop() {
		for _, s := range benchCastSamples {
			total += benchAccum(castTagValue(s.code, s.value))
		}
	}
	sinkFloat = total
}

// BenchmarkNewDXFTag measures tag construction, which runs for every scanned
// tag in addition to casting.
func BenchmarkNewDXFTag(b *testing.B) {
	b.ReportAllocs()
	total := 0.0
	for b.Loop() {
		for _, s := range benchCastSamples {
			total += benchAccum(NewDXFTag(s.code, s.value).Value())
		}
	}
	sinkFloat = total
}
