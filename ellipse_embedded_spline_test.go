package dxf

import (
	"bytes"
	"strings"
	"testing"

	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/path"
)

// ellipseSplineFixture is an ELLIPSE record whose embedded spline payload
// defines a cubic curve spanning the control polygon (0,0)-(10,0)-(10,5)-(0,5).
// The ellipse axis parameters (Major 1, Minor 0.5) are deliberately smaller
// than the spline so a render that follows them instead of the control data is
// detectable.
const ellipseSplineFixture = `0
SECTION
2
ENTITIES
0
ELLIPSE
5
E1
100
AcDbEntity
8
0
62
7
100
AcDbEllipse
10
0
20
0
30
0
40
1
41
0.5
50
0
51
360
71
3
72
8
73
4
74
0
40
0
40
0
40
0
40
0
40
1
40
1
40
1
40
1
10
0
20
0
30
0
10
10
20
0
30
0
10
10
20
5
30
0
10
0
20
5
30
0
0
ENDSEC
0
EOF
`

func ellipseEntity(t *testing.T, src string) *entity.Ellipse {
	t.Helper()
	d, err := FromStringData(src)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	for _, e := range d.ModelSpace() {
		if el, ok := e.(*entity.Ellipse); ok {
			return el
		}
	}
	t.Fatal("no ELLIPSE entity loaded")
	return nil
}

func TestEllipseEmbeddedSplineLoad(t *testing.T) {
	e := ellipseEntity(t, ellipseSplineFixture)
	if e.Spline == nil {
		t.Fatalf("embedded spline not parsed: %q", e.EmbeddedSplineError)
	}
	if e.EmbeddedSplineError != "" {
		t.Errorf("unexpected embedded spline error: %s", e.EmbeddedSplineError)
	}
	sp := e.Spline
	if sp.Degree != 3 {
		t.Errorf("degree = %d, want 3", sp.Degree)
	}
	if len(sp.Knots) != 8 {
		t.Errorf("knots = %d, want 8", len(sp.Knots))
	}
	if len(sp.Controls) != 4 {
		t.Fatalf("controls = %d, want 4", len(sp.Controls))
	}
	want := [][3]float64{{0, 0, 0}, {10, 0, 0}, {10, 5, 0}, {0, 5, 0}}
	for i, w := range want {
		got := sp.Controls[i]
		if got[0] != w[0] || got[1] != w[1] || got[2] != w[2] {
			t.Errorf("control %d = %v, want %v", i, got, w)
		}
	}
}

func TestEllipseEmbeddedSplineRoundTrip(t *testing.T) {
	d, err := FromStringData(ellipseSplineFixture)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	var buf bytes.Buffer
	if _, err := d.WriteTo(&buf); err != nil {
		t.Fatalf("save: %v", err)
	}
	reloaded, err := FromStringData(buf.String())
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	var got *entity.Ellipse
	for _, e := range reloaded.ModelSpace() {
		if el, ok := e.(*entity.Ellipse); ok {
			got = el
		}
	}
	if got == nil || got.Spline == nil {
		t.Fatal("embedded spline lost on round trip")
	}
	orig := ellipseEntity(t, ellipseSplineFixture).Spline
	if got.Spline.Degree != orig.Degree || len(got.Spline.Knots) != len(orig.Knots) || len(got.Spline.Controls) != len(orig.Controls) {
		t.Fatalf("spline shape changed: %#v", got.Spline)
	}
	for i := range orig.Controls {
		for k := 0; k < 3; k++ {
			if got.Spline.Controls[i][k] != orig.Controls[i][k] {
				t.Errorf("control %d axis %d = %g, want %g", i, k, got.Spline.Controls[i][k], orig.Controls[i][k])
			}
		}
	}
	for i := range orig.Knots {
		if got.Spline.Knots[i] != orig.Knots[i] {
			t.Errorf("knot %d = %g, want %g", i, got.Spline.Knots[i], orig.Knots[i])
		}
	}
}

func TestEllipseEmbeddedSplineRendersThroughSpline(t *testing.T) {
	e := ellipseEntity(t, ellipseSplineFixture)
	p, err := path.MakePath(e)
	if err != nil {
		t.Fatalf("MakePath: %v", err)
	}
	if p == nil {
		t.Fatal("nil path")
	}
	min, max := p.BoundingBox()
	minX, minY, maxX, maxY := min.X(), min.Y(), max.X(), max.Y()
	// The spline spans the control polygon; the ellipse axes (Major 1) do not.
	if maxX-minX < 5 || maxY-minY < 2 {
		t.Fatalf("path bbox (%.2f,%.2f)-(%.2f,%.2f) does not follow spline controls", minX, minY, maxX, maxY)
	}
	if maxX > 12 || maxY > 7 {
		t.Errorf("path bbox (%.2f,%.2f)-(%.2f,%.2f) exceeds spline hull", minX, minY, maxX, maxY)
	}
}

func TestEllipseMalformedEmbeddedSplineRetained(t *testing.T) {
	bad := strings.Replace(ellipseSplineFixture, "73\n4\n74\n0", "73\n5\n74\n0", 1)
	if bad == ellipseSplineFixture {
		t.Fatal("fixture replacement failed")
	}
	d, err := FromStringData(bad)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	var e *entity.Ellipse
	for _, ent := range d.ModelSpace() {
		if el, ok := ent.(*entity.Ellipse); ok {
			e = el
		}
	}
	if e == nil {
		t.Fatal("no ELLIPSE entity loaded")
	}
	if e.Spline != nil {
		t.Fatal("malformed spline should not populate the typed model")
	}
	if !strings.Contains(e.EmbeddedSplineError, "control points") {
		t.Fatalf("error = %q, want control-points mismatch", e.EmbeddedSplineError)
	}
	if len(e.RawTags()) == 0 {
		t.Fatal("malformed payload not retained")
	}

	scene, err := renderScene(d)
	if err != nil {
		t.Fatalf("scene: %v", err)
	}
	want := "ELLIPSE E1: malformed embedded spline: " + e.EmbeddedSplineError
	found := false
	for _, f := range scene.Fallbacks {
		if f == want {
			found = true
		}
	}
	if !found {
		t.Fatalf("fallback %q not reported; got %v", want, scene.Fallbacks)
	}
	scene2, err := renderScene(d)
	if err != nil {
		t.Fatalf("second scene: %v", err)
	}
	if strings.Join(scene.Fallbacks, "\n") != strings.Join(scene2.Fallbacks, "\n") {
		t.Fatalf("fallback report not stable:\n%v\n%v", scene.Fallbacks, scene2.Fallbacks)
	}

	// The original payload must survive a save.
	var buf bytes.Buffer
	if _, err := d.WriteTo(&buf); err != nil {
		t.Fatalf("save: %v", err)
	}
	saved := buf.String()
	for _, coord := range []string{"\n10\n10\n", "\n20\n5\n"} {
		if !strings.Contains(saved, coord) {
			t.Errorf("saved record lost payload coordinate %q", strings.TrimSpace(coord))
		}
	}
}
