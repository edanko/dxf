package text2path

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/edanko/dxf/path"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// TTF text-to-path conversion: TrueType/OpenType fonts are located in the
// standard system font directories, glyph outlines are rasterized through
// the face's anti-aliased glyph masks at a supersampled size, and the
// thresholded mask is traced into closed fill contours (marching squares).
// Callers fall back to the built-in vector font (TextToPath.Convert) when
// this fails.

var (
	fontMu    sync.Mutex
	fontPath  string // resolved system font file, cached
	fontParse = map[string]*opentype.Font{}
)

// preferredFonts are case-insensitive name fragments, best first.
var preferredFonts = []string{"notosans-", "notosans", "liberation-sans", "dejavusans", "freesans"}

// systemFontPath finds a usable TTF/OTF file in the system font directories,
// caching the result.
func systemFontPath() (string, error) {
	fontMu.Lock()
	defer fontMu.Unlock()
	if fontPath != "" {
		return fontPath, nil
	}
	dirs := []string{"/usr/share/fonts", "/usr/local/share/fonts", "/Library/Fonts"}
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, ".fonts"), filepath.Join(home, ".local/share/fonts"))
	}
	var candidates []string
	for _, dir := range dirs {
		_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return nil // skip unreadable subtrees
			}
			if d.IsDir() {
				return nil
			}
			ext := strings.ToLower(d.Name())
			if ext == ".ttf" || ext == ".otf" {
				candidates = append(candidates, p)
			}
			return nil
		})
	}
	if len(candidates) == 0 {
		return "", fmt.Errorf("no TrueType/OpenType fonts found in system directories")
	}
	sort.Strings(candidates)
	for _, pref := range preferredFonts {
		for _, p := range candidates {
			if strings.Contains(strings.ToLower(filepath.Base(p)), pref) {
				fontPath = p
				return p, nil
			}
		}
	}
	fontPath = candidates[0]
	return fontPath, nil
}

// loadFont parses and caches the system font.
func loadFont() (*opentype.Font, string, error) {
	p, err := systemFontPath()
	if err != nil {
		return nil, "", err
	}
	fontMu.Lock()
	f := fontParse[p]
	fontMu.Unlock()
	if f != nil {
		return f, filepath.Base(p), nil
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, p, err
	}
	f, err = opentype.Parse(data)
	if err != nil {
		return nil, p, fmt.Errorf("parsing font %s: %w", p, err)
	}
	fontMu.Lock()
	fontParse[p] = f
	fontMu.Unlock()
	return f, filepath.Base(p), nil
}

// canvas8 is an 8-bit coverage canvas in supersampled pixel space.
type canvas8 struct {
	w, h int
	pix  []uint8
}

func (c *canvas8) blend(x, y int, v uint8) {
	if x < 0 || y < 0 || x >= c.w || y >= c.h || v == 0 {
		return
	}
	i := y*c.w + x
	if v > c.pix[i] {
		c.pix[i] = v
	}
}

// TextToPathsTTF converts text into closed fill outlines using the system
// TTF face, where height is the face size in drawing units. Returned paths
// are positioned with the baseline origin at (0, 0), growing right, y-up.
// The second return value names the font file used.
//
// When no system font is available or parsing fails, an error is returned
// and the caller should fall back to the built-in vector font.
func TextToPathsTTF(text string, height float64) ([]*path.Path, string, error) {
	if text == "" || height <= 0 {
		return nil, "", nil
	}
	f, name, err := loadFont()
	if err != nil {
		return nil, "", err
	}
	// Supersample: rasterize at ss device pixels per drawing unit for
	// smooth contours, then scale contours back down.
	ss := 8.0
	if height*ss > 2048 {
		ss = 2048 / height
	}
	if ss < 1 {
		return nil, name, fmt.Errorf("text size too large for rasterization")
	}
	face, err := opentype.NewFace(f, &opentype.FaceOptions{Size: height * ss, DPI: 72})
	if err != nil {
		return nil, name, err
	}
	m := face.Metrics()
	em := height * ss
	ascent := float64(m.Ascent) / 64
	descent := float64(m.Descent) / 64

	// Measure total advance to size the canvas.
	var pen fixed.Int26_6
	for _, r := range text {
		a, ok := face.GlyphAdvance(r)
		if !ok {
			a, _ = face.GlyphAdvance(' ')
		}
		pen += a
	}
	w := int(float64(pen)/64) + int(3*em/2)
	h := int(ascent+descent) + int(em)
	if w < 4 || h < 4 {
		return nil, name, fmt.Errorf("empty text extent")
	}
	canvas := &canvas8{w: w, h: h, pix: make([]uint8, w*h)}
	yBase := int(ascent) + int(em)/4

	var penX fixed.Int26_6
	for _, r := range text {
		if r == '\n' || r == '\t' {
			r = ' '
		}
		a, ok := face.GlyphAdvance(r)
		if !ok {
			a, _ = face.GlyphAdvance(' ')
		}
		if r != ' ' {
			dot := fixed.Point26_6{X: penX, Y: fixed.Int26_6(yBase << 6)}
			dr, mask, maskp, _, okg := face.Glyph(dot, r)
			if !okg {
				penX += a
				continue
			}
			_ = dr
			if mask != nil {
				mb := mask.Bounds()
				ox := dr.Min.X - maskp.X
				oy := dr.Min.Y - maskp.Y
				for my := mb.Min.Y; my < mb.Max.Y; my++ {
					for mx := mb.Min.X; mx < mb.Max.X; mx++ {
						_, _, _, av := mask.At(mx, my).RGBA()
						canvas.blend(ox+mx, oy+my, uint8(av>>8))
					}
				}
			}
		}
		penX += a
	}

	loops := maskContours(canvas, 128)
	merged := path.NewPath()
	ncontours := 0
	for _, loop := range loops {
		if len(loop) < 3 {
			continue
		}
		ncontours++
		for i, pt := range loop {
			ux := float64(pt[0]) / 2 / ss
			uy := (float64(yBase) - float64(pt[1])/2) / ss // y-up, baseline at 0
			if i == 0 {
				merged.MoveTo(ux, uy)
			} else {
				merged.LineTo(ux, uy)
			}
		}
		ux0 := float64(loop[0][0]) / 2 / ss
		uy0 := (float64(yBase) - float64(loop[0][1])/2) / ss
		merged.LineTo(ux0, uy0) // close the contour
	}
	if ncontours == 0 {
		return nil, name, fmt.Errorf("no glyph contours traced")
	}
	return []*path.Path{merged}, name, nil
}

// maskContours traces closed contours of pixels with coverage >= threshold
// using marching squares on the 2x2 lattice of pixel corners. Adjacent
// cells share edge midpoints, so every midpoint carries exactly two segment
// endpoints and chaining yields closed loops. Nested loops (holes) cancel
// under parity filling. Returned points are in doubled half-pixel coords.
func maskContours(c *canvas8, threshold uint8) [][][2]int {
	set := func(x, y int) bool {
		return x >= 0 && y >= 0 && x < c.w && y < c.h && c.pix[y*c.w+x] >= threshold
	}
	adj := map[[2]int][][2]int{}
	add := func(a, b [2]int) {
		adj[a] = append(adj[a], b)
		adj[b] = append(adj[b], a)
	}
	for y := 1; y < c.h; y++ {
		for x := 1; x < c.w; x++ {
			idx := 0
			if set(x-1, y-1) {
				idx |= 8
			}
			if set(x, y-1) {
				idx |= 4
			}
			if set(x-1, y) {
				idx |= 2
			}
			if set(x, y) {
				idx |= 1
			}
			if idx == 0 || idx == 15 {
				continue
			}
			n := [2]int{2 * x, 2*y - 1}
			e := [2]int{2*x + 1, 2 * y}
			s := [2]int{2 * x, 2*y + 1}
			wm := [2]int{2*x - 1, 2 * y}
			for _, sp := range msSegments(idx, n, e, s, wm) {
				add(sp.a, sp.b)
			}
		}
	}
	visited := map[[2]int]int{} // start marker for loop tracing
	var loops [][][2]int
	for start := range adj {
		if visited[start] == 2 {
			continue
		}
		loop := [][2]int{}
		cur := start
		for {
			visited[cur] = 2
			loop = append(loop, cur)
			next := [2]int{}
			found := false
			for _, cnd := range adj[cur] {
				if visited[cnd] != 2 {
					next = cnd
					found = true
					break
				}
			}
			if !found {
				break
			}
			cur = next
			if cur == start {
				break
			}
		}
		if len(loop) >= 3 {
			loops = append(loops, loop)
		}
	}
	return loops
}

type msSeg struct{ a, b [2]int }

// msSegments returns the marching-squares boundary segments for a cell
// corner configuration (TL*8+TR*4+BL*2+BR).
func msSegments(idx int, n, e, s, w [2]int) []msSeg {
	seg := func(a, b [2]int) msSeg { return msSeg{a, b} }
	switch idx {
	case 1: // BR
		return []msSeg{seg(s, e)}
	case 2: // BL
		return []msSeg{seg(s, w)}
	case 3: // BL+BR
		return []msSeg{seg(w, e)}
	case 4: // TR
		return []msSeg{seg(n, e)}
	case 5: // TR+BR
		return []msSeg{seg(n, s)}
	case 6: // TR+BL saddle
		return []msSeg{seg(n, w), seg(s, e)}
	case 7: // TR+BL+BR
		return []msSeg{seg(n, w)}
	case 8: // TL
		return []msSeg{seg(n, w)}
	case 9: // TL+BR saddle
		return []msSeg{seg(n, e), seg(s, w)}
	case 10: // TL+BL
		return []msSeg{seg(n, s)}
	case 11: // TL+BL+BR
		return []msSeg{seg(n, e)}
	case 12: // TL+TR
		return []msSeg{seg(w, e)}
	case 13: // TL+TR+BR
		return []msSeg{seg(s, e)}
	case 14: // TL+TR+BL
		return []msSeg{seg(s, w)}
	}
	return nil
}
