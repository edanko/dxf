// Package importer imports supported external drawing formats into DXF
// entities.
package importer

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"github.com/edanko/dxf/addons"
	"github.com/edanko/dxf/drawing"
)

type capability struct{}

func (capability) Name() string { return "importer" }
func (capability) Description() string {
	return "import external drawing formats (SVG) into DXF entities"
}

func init() { _ = addons.Register("importer", capability{}) }

type svgLine struct {
	X1 float64 `xml:"x1,attr"`
	Y1 float64 `xml:"y1,attr"`
	X2 float64 `xml:"x2,attr"`
	Y2 float64 `xml:"y2,attr"`
}

type svgRect struct {
	X      float64 `xml:"x,attr"`
	Y      float64 `xml:"y,attr"`
	Width  float64 `xml:"width,attr"`
	Height float64 `xml:"height,attr"`
}

type svgCircle struct {
	Cx float64 `xml:"cx,attr"`
	Cy float64 `xml:"cy,attr"`
	R  float64 `xml:"r,attr"`
}

type svgPoly struct {
	Points string `xml:"points,attr"`
}

type svgPath struct {
	D string `xml:"d,attr"`
}

type svgDoc struct {
	XMLName   xml.Name    `xml:"svg"`
	Lines     []svgLine   `xml:"line"`
	Rects     []svgRect   `xml:"rect"`
	Circles   []svgCircle `xml:"circle"`
	Polylines []svgPoly   `xml:"polyline"`
	Polygons  []svgPoly   `xml:"polygon"`
	Paths     []svgPath   `xml:"path"`
}

// ImportSVG converts a simple SVG document (lines, rectangles, circles,
// polylines, polygons and basic path data) into DXF entities on layer 0.
func ImportSVG(data []byte) (*drawing.Drawing, error) {
	var svg svgDoc
	if err := xml.Unmarshal(data, &svg); err != nil {
		return nil, fmt.Errorf("importer: parse SVG: %w", err)
	}
	d, err := drawing.New()
	if err != nil {
		return nil, err
	}
	ms := d.Model()
	for _, l := range svg.Lines {
		if _, err := ms.AddLine(l.X1, l.Y1, 0, l.X2, l.Y2, 0); err != nil {
			return nil, err
		}
	}
	for _, r := range svg.Rects {
		if r.Width == 0 || r.Height == 0 {
			continue
		}
		x, y := r.X, r.Y
		if _, err := ms.AddLwPolyline(true, []float64{
			x, y, x + r.Width, y, x + r.Width, y + r.Height, x, y + r.Height,
		}); err != nil {
			return nil, err
		}
	}
	for _, c := range svg.Circles {
		if _, err := ms.AddCircle(c.Cx, c.Cy, 0, c.R); err != nil {
			return nil, err
		}
	}
	for _, p := range svg.Polylines {
		if err := addPolyline(ms, p.Points, false); err != nil {
			return nil, err
		}
	}
	for _, p := range svg.Polygons {
		if err := addPolyline(ms, p.Points, true); err != nil {
			return nil, err
		}
	}
	for _, p := range svg.Paths {
		if err := addPath(ms, p.D); err != nil {
			return nil, err
		}
	}
	return d, nil
}

func addPolyline(ms *drawing.Space, points string, closed bool) error {
	fields := strings.FieldsFunc(points, func(r rune) bool { return r == ',' || r == ' ' || r == '\t' || r == '\n' })
	if len(fields) < 4 || len(fields)%2 != 0 {
		return fmt.Errorf("importer: polyline needs an even number of coordinates, got %d", len(fields))
	}
	verts := make([]float64, 0, len(fields))
	for _, f := range fields {
		v, err := strconv.ParseFloat(f, 64)
		if err != nil {
			return fmt.Errorf("importer: polyline coordinate %q: %w", f, err)
		}
		verts = append(verts, v)
	}
	_, err := ms.AddLwPolyline(closed, verts)
	return err
}

// addPath appends a path entity (basic M/L/H/V/Z commands).
func addPath(ms *drawing.Space, d string) error {
	fields := strings.FieldsFunc(d, func(r rune) bool { return r == ',' || r == ' ' || r == '\t' || r == '\n' })
	var verts []float64
	var curX, curY, subX, subY float64
	var started bool
	flush := func() error {
		if len(verts) >= 4 {
			if _, err := ms.AddLwPolyline(true, verts); err != nil {
				return err
			}
		}
		verts = nil
		return nil
	}
	i := 0
	for i < len(fields) {
		cmd := fields[i]
		i++
		switch cmd {
		case "M", "m":
			if i+1 >= len(fields) {
				return fmt.Errorf("importer: M needs coordinates")
			}
			x, err1 := strconv.ParseFloat(fields[i], 64)
			y, err2 := strconv.ParseFloat(fields[i+1], 64)
			i += 2
			if err1 != nil || err2 != nil {
				return fmt.Errorf("importer: M coordinate")
			}
			if cmd == "m" {
				x += curX
				y += curY
			}
			if started {
				if err := flush(); err != nil {
					return err
				}
			}
			curX, curY = x, y
			subX, subY = x, y
			verts = append(verts, x, y)
			started = true
		case "L", "l":
			if i+1 >= len(fields) {
				return fmt.Errorf("importer: L needs coordinates")
			}
			x, err1 := strconv.ParseFloat(fields[i], 64)
			y, err2 := strconv.ParseFloat(fields[i+1], 64)
			i += 2
			if err1 != nil || err2 != nil {
				return fmt.Errorf("importer: L coordinate")
			}
			if cmd == "l" {
				x += curX
				y += curY
			}
			curX, curY = x, y
			verts = append(verts, x, y)
		case "H", "h":
			if i >= len(fields) {
				return fmt.Errorf("importer: H needs a value")
			}
			x, err := strconv.ParseFloat(fields[i], 64)
			i++
			if err != nil {
				return fmt.Errorf("importer: H value")
			}
			if cmd == "h" {
				x += curX
			}
			curX = x
			verts = append(verts, x, curY)
		case "V", "v":
			if i >= len(fields) {
				return fmt.Errorf("importer: V needs a value")
			}
			y, err := strconv.ParseFloat(fields[i], 64)
			i++
			if err != nil {
				return fmt.Errorf("importer: V value")
			}
			if cmd == "v" {
				y += curY
			}
			curY = y
			verts = append(verts, curX, y)
		case "Z", "z":
			verts = append(verts, subX, subY)
			if err := flush(); err != nil {
				return err
			}
			started = false
			curX, curY = subX, subY
		default:
			return fmt.Errorf("importer: unsupported path command %q", cmd)
		}
	}
	if len(verts) >= 4 {
		if _, err := ms.AddLwPolyline(true, verts); err != nil {
			return err
		}
	}
	return nil
}
