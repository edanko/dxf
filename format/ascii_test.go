package format

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/edanko/dxf/lldxf"
)

// referenceASCII reproduces the fmt-based tag rendering the formatter used
// before it switched to strconv, so output can be compared byte for byte.
type referenceASCII struct {
	float string
}

func (r referenceASCII) String(num int, val string) string {
	return fmt.Sprintf("%d\n%s\n", num, val)
}

func (r referenceASCII) Int(num int, val int) string {
	return fmt.Sprintf("%d\n%d\n", num, val)
}

func (r referenceASCII) Float(num int, val float64) string {
	return fmt.Sprintf(fmt.Sprintf("%d\n%s\n", num, r.float), val)
}

func floatSamples() []float64 {
	return []float64{
		0,
		math.Copysign(0, -1),
		1,
		-1.5,
		0.1,
		1.0 / 3.0,
		123456789.123456789,
		1e-10,
		1e21,
		1e300,
		-1e300,
		2.0 / 3.0,
		math.NaN(),
		math.Inf(1),
		math.Inf(-1),
		math.MaxFloat64,
		math.SmallestNonzeroFloat64,
	}
}

// TestASCIINumericRenderMatchesReference proves float and integer tags render
// exactly as the fmt verbs they replaced, across precisions and edge values.
func TestASCIINumericRenderMatchesReference(t *testing.T) {
	precisions := []int{0, 1, 3, 6, 8, 17}
	for _, p := range precisions {
		f := NewASCII()
		f.SetPrecision(p)
		ref := referenceASCII{float: fmt.Sprintf("%%.%df", p)}

		for _, val := range floatSamples() {
			if got, want := f.Float(10, val), ref.Float(10, val); got != want {
				t.Errorf("precision %d: Float(10, %v) = %q, want %q", p, val, got, want)
			}
		}
		for _, val := range []int{0, -1, 1 << 40, math.MinInt32} {
			if got, want := f.Int(70, val), ref.Int(70, val); got != want {
				t.Errorf("precision %d: Int(70, %d) = %q, want %q", p, val, got, want)
			}
		}
	}
}

// TestASCIIStringRenderMatchesReference covers the string path, including the
// codepage encoding applied for pre-R2007 targets.
func TestASCIIStringRenderMatchesReference(t *testing.T) {
	values := []string{"", "LINE", "a b", "café", "日本語", "x\tY"}
	f := NewASCII()
	ref := referenceASCII{}
	for _, val := range values {
		if got, want := f.String(1, val), ref.String(1, val); got != want {
			t.Errorf("String(1, %q) = %q, want %q", val, got, want)
		}
	}

	cp := NewASCIIWithVersion(AC1015)
	cp.SetCodepage("ANSI_1252")
	for _, val := range values {
		want := fmt.Sprintf("1\n%s\n", string(lldxf.EncodeCodepageBytes(val, "ANSI_1252")))
		if got := cp.String(1, val); got != want {
			t.Errorf("codepage String(1, %q) = %q, want %q", val, got, want)
		}
	}
}

// TestASCIIWriteMatchesRender proves the buffered writers produce exactly the
// bytes the string-returning helpers do, so the allocation-free fast path and
// the public API cannot drift apart.
func TestASCIIWriteMatchesRender(t *testing.T) {
	f := NewASCII()
	ref := referenceASCII{float: "%.6f"}

	var want strings.Builder
	for _, val := range []string{"SECTION", "café", ""} {
		f.WriteString(2, val)
		want.WriteString(ref.String(2, val))
	}
	for _, val := range []int{0, 1, -7, 1 << 30} {
		f.WriteInt(70, val)
		want.WriteString(ref.Int(70, val))
	}
	for _, val := range floatSamples()[:8] {
		f.WriteFloat(40, val)
		want.WriteString(ref.Float(40, val))
	}
	for _, val := range []bool{true, false} {
		f.WriteBool(290, val)
		n := 0
		if val {
			n = 1
		}
		want.WriteString(ref.Int(290, n))
	}

	if got := f.Output(); got != want.String() {
		t.Errorf("Output() = %q, want %q", got, want.String())
	}
	if got, err := f.WriteTo(&strings.Builder{}); err != nil || got != 0 {
		t.Errorf("WriteTo after Output() = (%d, %v), want (0, nil)", got, err)
	}
}

// TestASCIISetPrecisionKeepsFormatString keeps the promoted precision and the
// legacy format-string field in step.
func TestASCIISetPrecisionKeepsFormatString(t *testing.T) {
	f := NewASCII()
	for _, p := range []int{0, 1, 4, 12} {
		f.SetPrecision(p)
		if want := fmt.Sprintf("%%.%df", p); f.float != want {
			t.Errorf("SetPrecision(%d) left float = %q, want %q", p, f.float, want)
		}
		if f.prec != p {
			t.Errorf("SetPrecision(%d) left prec = %d", p, f.prec)
		}
	}
}
