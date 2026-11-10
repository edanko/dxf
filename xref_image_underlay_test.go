// Parity tests for change 10: IMAGE/IMAGEDEF round trip and rendering,
// UNDERLAY entities and definitions, XCLIP boundaries on INSERT, raster
// image drawing.
package dxf_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	dxf "github.com/edanko/dxf"
	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
	"github.com/edanko/dxf/object"
	"github.com/edanko/dxf/render"
)

func writeTestPNG(t *testing.T, path string) {
	t.Helper()
	im := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := range 8 {
		for x := range 8 {
			im.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, im); err != nil {
		t.Fatal(err)
	}
}

func newImageDrawing(t *testing.T, defPath string) *drawing.Drawing {
	t.Helper()
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	def := object.NewImageDef("FA", defPath, 8, 8, 0.001, 0.001)
	d.AddObject(def)
	img := entity.NewImage()
	img.SetHandleString("42")
	img.SetImageDefHandle("FA")
	img.SetInsertPoint(dxfmath.NewVec3(1, 2, 0))
	img.SetUVectors(dxfmath.NewVec3(1, 0, 0), dxfmath.NewVec3(0, 1, 0))
	img.SetImageSize(8, 8)
	img.SetDisplayProperties(50, 50, 30)
	img.SetClipping(true)
	img.SetRectangularClipping(0, 0, 4, 4)
	if err := d.AddEntity(img); err != nil {
		t.Fatal(err)
	}
	return d
}

func TestImageEntityRoundTrip(t *testing.T) {
	d := newImageDrawing(t, "/tmp/pic.png")
	var buf bytes.Buffer
	if _, err := d.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	doc, err := dxf.FromStringData(buf.String())
	if err != nil {
		t.Fatal(err)
	}

	var got *entity.Image
	for _, e := range doc.Entities() {
		if im, ok := e.(*entity.Image); ok {
			got = im
			break
		}
	}
	if got == nil {
		t.Fatal("IMAGE entity not reloaded")
	}
	if got.Fade() != 30 {
		t.Fatalf("fade: %v", got.Fade())
	}
	if got.Brightness() != 50 || got.Contrast() != 50 {
		t.Fatalf("display props: %v %v", got.Brightness(), got.Contrast())
	}
	if got.Clipping() != 1 {
		t.Fatalf("clipping: %v", got.Clipping())
	}
	if got.ClippingBoundaryType() != entity.ImageClippingRectangular {
		t.Fatalf("clip type: %v", got.ClippingBoundaryType())
	}
	if len(got.BoundaryPoints()) != 2 {
		t.Fatalf("clip points: %v", got.BoundaryPoints())
	}
	if got.ImageDefHandle() != "FA" {
		t.Fatalf("def handle: %s", got.ImageDefHandle())
	}
	if got.InsertPoint().X() != 1 || got.InsertPoint().Y() != 2 {
		t.Fatalf("insert: %v", got.InsertPoint())
	}

	o, ok := doc.GetObject("FA")
	if !ok {
		t.Fatal("IMAGEDEF object missing")
	}
	id, ok := o.(*object.ImageDef)
	if !ok {
		t.Fatalf("def type %T", o)
	}
	if id.FilePath() != "/tmp/pic.png" {
		t.Fatalf("def path: %s", id.FilePath())
	}
}

func TestImageRenderingScene(t *testing.T) {
	dir := t.TempDir()
	writeTestPNG(t, filepath.Join(dir, "pic.png"))

	d := newImageDrawing(t, "pic.png")

	sc, err := render.BuildScene(d, 800, 600, &render.SceneOptions{BasePath: dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(sc.Rasters) != 1 {
		t.Fatalf("rasters: %d fallbacks %v", len(sc.Rasters), sc.Fallbacks)
	}

	sc2, err := render.BuildScene(d, 800, 600, &render.SceneOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(sc2.Rasters) != 0 {
		t.Fatalf("rasters without base: %d", len(sc2.Rasters))
	}
	found := false
	for _, f := range sc2.Fallbacks {
		if strings.Contains(f, "image file not found") {
			found = true
		}
	}
	if !found {
		t.Fatalf("fallback note missing: %v", sc2.Fallbacks)
	}
}

func TestUnderlayEntityRoundTrip(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	def := object.NewUnderlayDef("FB", "PDFDEFINITION", "Sheet1", "doc.pdf", 1, 2, 50, 20)
	d.AddObject(def)
	u := entity.NewUnderlay()
	u.SetHandleString("43")
	u.UnderlayDefHandle = "FB"
	u.SetInsertPoint(0, 0, 0)
	u.SetScale(1, 1, 1)
	u.SetFade(20)
	if err := d.AddEntity(u); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if _, err := d.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	doc, err := dxf.FromStringData(buf.String())
	if err != nil {
		t.Fatal(err)
	}

	var got *entity.Underlay
	for _, e := range doc.Entities() {
		if un, ok := e.(*entity.Underlay); ok {
			got = un
			break
		}
	}
	if got == nil {
		t.Fatal("UNDERLAY entity not reloaded")
	}
	if got.Fade != 20 {
		t.Fatalf("fade: %v", got.Fade)
	}
	if got.UnderlayDefHandle != "FB" {
		t.Fatalf("def handle: %s", got.UnderlayDefHandle)
	}

	o, ok := doc.GetObject("FB")
	if !ok {
		t.Fatal("PDFDEFINITION object missing")
	}
	pd, ok := o.(*object.PdfDefinition)
	if !ok {
		t.Fatalf("def type %T", o)
	}
	if pd.FilePath() != "doc.pdf" {
		t.Fatalf("def path: %s", pd.FilePath())
	}
	if pd.Name() != "Sheet1" {
		t.Fatalf("def name: %s", pd.Name())
	}
	if pd.Page() != 1 {
		t.Fatalf("def page: %d", pd.Page())
	}
}

func TestXClipRoundTrip(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	blk, err := d.NewBlock("B", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.BlockSpace(blk).AddLine(0, 0, 0, 10, 10, 0); err != nil {
		t.Fatal(err)
	}
	ins := entity.NewInsert()
	ins.SetHandleString("44")
	ins.SetBlockName("B")
	ins.SetBlockDef(blk)
	ins.SetXClipRect(0, 0, 5, 5)
	if err := d.AddEntity(ins); err != nil {
		t.Fatal(err)
	}
	ins2 := entity.NewInsert()
	ins2.SetHandleString("45")
	ins2.SetBlockName("B")
	ins2.SetBlockDef(blk)
	ins2.SetInsertPoint(dxfmath.NewVec3(20, 0, 0))
	ins2.SetXClipPolyline([][2]float64{{0, 0}, {5, 0}, {5, 5}, {0, 5}})
	if err := d.AddEntity(ins2); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if _, err := d.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	doc, err := dxf.FromStringData(buf.String())
	if err != nil {
		t.Fatal(err)
	}

	var rects, polys int
	for _, e := range doc.Entities() {
		in, ok := e.(*entity.Insert)
		if !ok || !in.IsXClipped() {
			continue
		}
		pts := in.XClipBoundaryPath()
		switch in.XClipType() {
		case entity.XClipRectangular:
			if len(pts) != 2 || pts[0] != [2]float64{0, 0} || pts[1] != [2]float64{5, 5} {
				t.Fatalf("rect clip: %v", pts)
			}
			rects++
		case entity.XClipPolyline:
			if len(pts) != 4 || pts[2] != [2]float64{5, 5} {
				t.Fatalf("poly clip: %v", pts)
			}
			polys++
		}
	}
	if rects != 1 || polys != 1 {
		t.Fatalf("xclip counts: rect %d poly %d", rects, polys)
	}

	sc, err := render.BuildScene(doc, 800, 600, nil)
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	r, err := render.NewBackend("svg", &out, 800, 600)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Begin(800, 600); err != nil {
		t.Fatal(err)
	}
	for _, sp := range sc.Paths {
		if err := r.Clip(sp.Clip); err != nil {
			t.Fatal(err)
		}
		if err := r.SetStyle(sp.Style); err != nil {
			t.Fatal(err)
		}
		if err := r.DrawPath(sp.Path); err != nil {
			t.Fatal(err)
		}
		if err := r.Clip(nil); err != nil {
			t.Fatal(err)
		}
	}
	if err := r.End(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "<clipPath") {
		t.Fatalf("svg output has no clipPath:\n%s", out.String())
	}
}

func TestRasterImageDraw(t *testing.T) {
	dir := t.TempDir()
	writeTestPNG(t, filepath.Join(dir, "pic.png"))

	d := newImageDrawing(t, "pic.png")
	sc, err := render.BuildScene(d, 800, 600, &render.SceneOptions{BasePath: dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(sc.Rasters) != 1 {
		t.Fatalf("rasters: %d", len(sc.Rasters))
	}

	var buf bytes.Buffer
	r, err := render.NewBackend("png", &buf, 800, 600)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Begin(800, 600); err != nil {
		t.Fatal(err)
	}
	for _, ri := range sc.Rasters {
		if err := r.Clip(ri.Clip); err != nil {
			t.Fatal(err)
		}
		if err := r.DrawImage(ri.Image, ri.X, ri.Y, ri.W, ri.H, ri.Rotation, ri.Opacity); err != nil {
			t.Fatal(err)
		}
		if err := r.Clip(nil); err != nil {
			t.Fatal(err)
		}
	}
	if err := r.End(); err != nil {
		t.Fatal(err)
	}
	im, err := png.Decode(&buf)
	if err != nil {
		t.Fatal(err)
	}
	red := 0
	b := im.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			cr, cg, _, _ := im.At(x, y).RGBA()
			if cr>>8 > 200 && cg>>8 < 100 {
				red++
			}
		}
	}
	if red < 100 {
		t.Fatalf("raster image not drawn: %d red pixels", red)
	}
}
