package xplayer_test

import (
	"strings"
	"testing"

	"github.com/edanko/dxf/addons/xplayer"
	"github.com/edanko/dxf/drawing"
)

func TestExportAnimationSVG(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.Model().AddLine(0, 0, 0, 10, 10, 0); err != nil {
		t.Fatal(err)
	}
	frames, err := xplayer.ExportAnimationSVG(d, 4, 200, 200, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != 4 {
		t.Fatalf("frames: %d", len(frames))
	}
	for i, f := range frames {
		if !strings.Contains(f, "<svg") {
			t.Fatalf("frame %d has no svg root", i)
		}
	}
	if _, err := xplayer.ExportAnimationSVG(d, 0, 200, 200, nil); err == nil {
		t.Fatal("zero frames must error")
	}
}
