package patterns

import (
	"math"
	"reflect"
	"testing"
)

func TestParsePATMultiplePatternsAndErrors(t *testing.T) {
	src := []byte(`; leading comment
*DASH, Dashed lines
45, 0,0, 1,1, 0.5,-0.5
90, 2,3, 4,5, 1,-1
*DOTS, Dots
0, 0,0, 2,0, 0,-0.1
bad,line
*EMPTY,
0,0,0,1,0
`)
	f := ParsePAT(src, "custom.pat")
	if len(f.Patterns) != 3 {
		t.Fatalf("patterns = %d, want 3: %+v", len(f.Patterns), f.Patterns)
	}
	if len(f.Errors) != 1 || f.Errors[0].File != "custom.pat" || f.Errors[0].Line != 7 {
		t.Fatalf("errors = %+v", f.Errors)
	}
	first := f.Patterns[0]
	if first.Name != "DASH" || first.Description != "Dashed lines" || len(first.Lines) != 2 {
		t.Fatalf("first pattern = %+v", first)
	}
	if !reflect.DeepEqual(first.Lines[0].BasePoint, []float64{0, 0}) ||
		!reflect.DeepEqual(first.Lines[0].Offset, []float64{1, 1}) ||
		!reflect.DeepEqual(first.Lines[0].DashLengthItems, []float64{0.5, -0.5}) {
		t.Fatalf("parsed line = %+v", first.Lines[0])
	}
	if f.Patterns[1].Description != "Dots" || f.Patterns[2].Description != "" {
		t.Fatalf("descriptions = %q, %q", f.Patterns[1].Description, f.Patterns[2].Description)
	}
}

func TestPATLookupCaseInsensitiveAndSourceName(t *testing.T) {
	f := ParsePAT([]byte("*MyPat, Mine\n0,0,0,1,0\n"), "mine.pat")
	if err := f.Register(); err != nil {
		t.Fatal(err)
	}
	defer UnregisterPattern("MyPat")
	def, ok := LookupPattern("mypat")
	if !ok || def.Name != "MyPat" || len(def.Lines) != 1 {
		t.Fatalf("lookup = %+v, %v", def, ok)
	}
	def, ok = GetPatternDefinition("ansi31")
	if !ok || def.Name != "ANSI31" {
		t.Fatalf("predefined case-insensitive lookup = %+v, %v", def, ok)
	}
	if _, ok := LookupPattern("does-not-exist"); ok {
		t.Fatal("unknown pattern resolved")
	}
}

func TestPatternScaleAndMeasurementConversionAreNonMutating(t *testing.T) {
	def := PatternDefinition{
		Name: "CUSTOM",
		Lines: []PatternLine{{
			Angle:           45,
			BasePoint:       []float64{1, 2},
			Offset:          []float64{3, 4},
			DashLengthItems: []float64{0.5, -0.5},
		}},
	}
	scaled := def.Scaled(2)
	if !reflect.DeepEqual(scaled.Lines[0].BasePoint, []float64{2, 4}) ||
		!reflect.DeepEqual(scaled.Lines[0].Offset, []float64{6, 8}) ||
		!reflect.DeepEqual(scaled.Lines[0].DashLengthItems, []float64{1, -1}) {
		t.Fatalf("scaled = %+v", scaled.Lines[0])
	}
	if !reflect.DeepEqual(def.Lines[0].BasePoint, []float64{1, 2}) ||
		!reflect.DeepEqual(def.Lines[0].Offset, []float64{3, 4}) ||
		!reflect.DeepEqual(def.Lines[0].DashLengthItems, []float64{0.5, -0.5}) {
		t.Fatalf("original mutated = %+v", def.Lines[0])
	}

	imperial, _ := GetMeasurementSystemByName("Imperial")
	metric, _ := GetMeasurementSystemByName("Metric")
	converted := def.ConvertMeasurement(imperial, metric)
	if math.Abs(converted.Lines[0].Offset[0]-76.2) > 1e-9 ||
		math.Abs(converted.Lines[0].Offset[1]-101.6) > 1e-9 {
		t.Fatalf("converted offset = %+v", converted.Lines[0].Offset)
	}
	if !reflect.DeepEqual(def.Lines[0].Offset, []float64{3, 4}) {
		t.Fatalf("conversion mutated source: %+v", def.Lines[0])
	}
}
