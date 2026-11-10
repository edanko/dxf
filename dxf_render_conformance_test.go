package dxf

import (
	"bytes"
	"image/png"
	"path/filepath"
	"strings"
	"testing"

	"github.com/edanko/dxf/render"
)

// TestRenderConformance renders every corpus drawing through the scene
// builder and all three registered backends, asserting each backend
// produces structurally valid, non-empty output.
func TestRenderConformance(t *testing.T) {
	files, err := filepath.Glob("testdata/*.dxf")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no testdata fixtures")
	}
	for _, fn := range files {
		t.Run(filepath.Base(fn), func(t *testing.T) {
			d, err := FromFile(fn)
			if err != nil {
				t.Fatalf("load: %v", err)
			}
			sc, err := render.BuildScene(d, 800, 600, nil)
			if err != nil {
				t.Fatalf("BuildScene: %v", err)
			}
			for _, f := range sc.Fallbacks {
				t.Log("fallback:", f)
			}
			for _, name := range []string{"svg", "pdf", "png"} {
				var buf bytes.Buffer
				r, err := render.NewBackend(name, &buf, 800, 600)
				if err != nil {
					t.Fatal(err)
				}
				if err := r.Begin(800, 600); err != nil {
					t.Fatal(err)
				}
				for _, sp := range sc.Paths {
					if err := r.SetStyle(sp.Style); err != nil {
						t.Fatalf("%s style: %v", name, err)
					}
					if sp.FillHatch != nil && sp.FillHatch.Gradient != nil {
						if err := r.FillGradient(sp.Path, sp.FillHatch.Gradient); err != nil {
							t.Fatalf("%s gradient: %v", name, err)
						}
						continue
					}
					if err := r.DrawPath(sp.Path); err != nil {
						t.Fatalf("%s draw: %v", name, err)
					}
				}
				if err := r.End(); err != nil {
					t.Fatalf("%s end: %v", name, err)
				}
				out := buf.String()
				if len(out) == 0 {
					t.Fatalf("%s: empty output", name)
				}
				switch name {
				case "svg":
					if !strings.Contains(out, "<svg") {
						t.Error("svg: missing root element")
					}
				case "pdf":
					if !strings.HasPrefix(out, "%PDF") || !strings.Contains(out, "%%EOF") {
						t.Error("pdf: bad structure")
					}
				case "png":
					img, err := png.Decode(bytes.NewReader(buf.Bytes()))
					if err != nil {
						t.Fatalf("png: %v", err)
					}
					if b := img.Bounds(); b.Dx() != 800 || b.Dy() != 600 {
						t.Errorf("png dims %dx%d, want 800x600", b.Dx(), b.Dy())
					}
				}
			}
		})
	}
}
