// Package dxf is a DXF(Drawing Exchange Format) library for golang.
// Supports AC1009 through AC1024, ASCII and binary DXF formats.
package dxf

import (
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/lldxf"
	dxfmath "github.com/edanko/dxf/math"
)

// NewDrawing creates a new empty DXF drawing with default layers and styles.
func NewDrawing() (*drawing.Drawing, error) {
	return drawing.New()
}

// FromFile reads a DXF file and returns a Drawing.
// Supports both ASCII and binary DXF formats.
// The file path can be absolute or relative to the current working directory.
func FromFile(fn string) (*drawing.Drawing, error) {
	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	d, err := FromReader(f)
	if err != nil {
		return nil, err
	}
	d.SetFileName(fn)
	return d, nil
}

// FromStringData parses DXF data from a string and returns a Drawing.
// This is useful for testing or when DXF data is embedded in other files.
func FromStringData(d string) (*drawing.Drawing, error) {
	return FromReader(strings.NewReader(d))
}

// FromReader parses DXF data from any io.Reader and returns a Drawing.
// This is the core parsing function used by FromFile and FromStringData.
// Supports ASCII DXF format with automatic group code type detection.
func FromReader(r io.Reader) (*drawing.Drawing, error) {
	return FromReaderWithOptions(r, nil)
}

// FromBinaryFile reads a binary DXF file and returns a Drawing.
// Binary DXF files use the "AutoCAD Binary DXF" format with group codes
// encoded as binary values rather than ASCII text.
func FromBinaryFile(fn string) (*drawing.Drawing, error) {
	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return FromBinaryReader(f)
}

// FromBinaryReader parses binary DXF data from an io.Reader.
// This handles the special binary encoding used in DXB/binary DXF files.
func FromBinaryReader(r io.Reader) (*drawing.Drawing, error) {
	return FromBinaryReaderWithOptions(r, nil)
}

// DetectFormat examines DXF data and returns the format type.
// Returns "ASCII" for text DXF files or "Binary" for binary DXF files.
// The input should be the raw file bytes (first 32 bytes are sufficient).
func DetectFormat(data []byte) (string, error) {
	return lldxf.DetectDXFFormat(data)
}

func buildEntitiesFromLoader(d *drawing.Drawing, loader *lldxf.Loader) error {
	entities := loader.GetEntities()
	type entityHandle struct {
		handle string
		tags   lldxf.Tags
	}
	var entityList []entityHandle
	for handle, tags := range entities {
		entityList = append(entityList, entityHandle{handle: handle, tags: tags})
	}
	sort.Slice(entityList, func(i, j int) bool {
		return entityList[i].handle < entityList[j].handle
	})
	for _, eh := range entityList {
		tags := eh.tags
		dxfType, _ := tags.GetFirstValue(0)
		if dxfType == nil {
			continue
		}
		typeStr := fmt.Sprintf("%v", dxfType)

		e, err := loadEntity(typeStr, tags)
		if err != nil {
			continue
		}
		e.SetBlockRecord(nil)
		d.AddEntity(e)
	}
	return nil
}

// loadEntity dispatches to the entity loader registry. All 29 typed loaders
// register themselves from this package's init; unregistered types fall back
// to retained-record entities that keep every tag.
func loadEntity(dxfType string, tags lldxf.Tags) (entity.Entity, error) {
	// Library enum spellings of surface records (EXTRUDEDSURFACE, ...) map to
	// the canonical wire names registered in the loader table.
	switch strings.ToUpper(dxfType) {
	case "EXTRUDEDSURFACE":
		dxfType = "EXTRUDED_SURFACE"
	case "LOFTEDSURFACE":
		dxfType = "LOFTED_SURFACE"
	case "REVOLVEDSURFACE":
		dxfType = "REVOLVED_SURFACE"
	case "SWEPTSURFACE":
		dxfType = "SWEPT_SURFACE"
	}
	e, err := entity.LoadEntity(dxfType, tags)
	if err != nil {
		return nil, err
	}
	// Surface records share the SURFACE type enum; emit their canonical wire
	// names so re-saved documents keep the DXF record spelling.
	switch strings.ToUpper(dxfType) {
	case "EXTRUDED_SURFACE", "LOFTED_SURFACE", "REVOLVED_SURFACE", "SWEPT_SURFACE":
		if setter, ok := e.(interface{ SetRecordTypeName(string) }); ok {
			setter.SetRecordTypeName(strings.ToUpper(dxfType))
		}
	}
	// Variant record names (dimension variants, MPOLYGON reusing the hatch
	// loader) keep their DXF record spelling.
	up := strings.ToUpper(dxfType)

	if up == "ARC_DIMENSION" || up == "LARGE_RADIAL_DIMENSION" || up == "MPOLYGON" {
		if setter, ok := e.(interface{ SetRecordTypeName(string) }); ok {
			setter.SetRecordTypeName(up)
		}
	}
	return e, nil
}

// registerEntityLoaders wires the typed loaders into the registry once.
func init() {
	entity.SetDefaultLoader(entity.NewRetainedLoader)
	drawing.RegisterDocumentLoader(FromReader)
	loaders := map[string]func(lldxf.Tags) (entity.Entity, error){
		"POINT":            func(t lldxf.Tags) (entity.Entity, error) { return loadPoint(t), nil },
		"LINE":             func(t lldxf.Tags) (entity.Entity, error) { return loadLine(t), nil },
		"CIRCLE":           func(t lldxf.Tags) (entity.Entity, error) { return loadCircle(t), nil },
		"ARC":              func(t lldxf.Tags) (entity.Entity, error) { return loadArc(t), nil },
		"TEXT":             func(t lldxf.Tags) (entity.Entity, error) { return loadText(t), nil },
		"MTEXT":            func(t lldxf.Tags) (entity.Entity, error) { return loadMText(t), nil },
		"SOLID":            func(t lldxf.Tags) (entity.Entity, error) { return loadSolid(t), nil },
		"TRACE":            func(t lldxf.Tags) (entity.Entity, error) { return loadTrace(t), nil },
		"3DFACE":           func(t lldxf.Tags) (entity.Entity, error) { return load3DFace(t), nil },
		"LWPOLYLINE":       func(t lldxf.Tags) (entity.Entity, error) { return loadLwPolyline(t), nil },
		"POLYLINE":         func(t lldxf.Tags) (entity.Entity, error) { return loadPolyline(t), nil },
		"VERTEX":           func(t lldxf.Tags) (entity.Entity, error) { return loadVertex(t), nil },
		"INSERT":           func(t lldxf.Tags) (entity.Entity, error) { return loadInsert(t), nil },
		"ELLIPSE":          func(t lldxf.Tags) (entity.Entity, error) { return loadEllipse(t), nil },
		"SPLINE":           func(t lldxf.Tags) (entity.Entity, error) { return loadSpline(t), nil },
		"LEADER":           func(t lldxf.Tags) (entity.Entity, error) { return loadLeader(t), nil },
		"HATCH":            func(t lldxf.Tags) (entity.Entity, error) { return loadHatch(t), nil },
		"MPOLYGON":         func(t lldxf.Tags) (entity.Entity, error) { return loadHatch(t), nil },
		"IMAGE":            func(t lldxf.Tags) (entity.Entity, error) { return loadImage(t), nil },
		"MLEADER":          func(t lldxf.Tags) (entity.Entity, error) { return loadMLeader(t), nil },
		"MULTILEADER":      func(t lldxf.Tags) (entity.Entity, error) { return loadMLeader(t), nil },
		"DIMENSION":        func(t lldxf.Tags) (entity.Entity, error) { return loadDimension(t), nil },
		"SHAPE":            func(t lldxf.Tags) (entity.Entity, error) { return loadShape(t), nil },
		"XLINE":            func(t lldxf.Tags) (entity.Entity, error) { return loadXLine(t), nil },
		"RAY":              func(t lldxf.Tags) (entity.Entity, error) { return loadRay(t), nil },
		"WIPEOUT":          func(t lldxf.Tags) (entity.Entity, error) { return loadWipeout(t), nil },
		"UNDERLAY":         func(t lldxf.Tags) (entity.Entity, error) { return loadUnderlay(t), nil },
		"MLINE":            func(t lldxf.Tags) (entity.Entity, error) { return loadMLine(t), nil },
		"ATTDEF":           func(t lldxf.Tags) (entity.Entity, error) { return loadAttdef(t), nil },
		"ATTRIB":           func(t lldxf.Tags) (entity.Entity, error) { return loadAttrib(t), nil },
		"TOLERANCE":        func(t lldxf.Tags) (entity.Entity, error) { return loadTolerance(t), nil },
		"3DSOLID":          func(t lldxf.Tags) (entity.Entity, error) { return loadSolid3d(t), nil },
		"REGION":           func(t lldxf.Tags) (entity.Entity, error) { return loadRegion(t), nil },
		"BODY":             func(t lldxf.Tags) (entity.Entity, error) { return loadBody(t), nil },
		"MESH":             func(t lldxf.Tags) (entity.Entity, error) { return loadMesh(t), nil },
		"HELIX":            func(t lldxf.Tags) (entity.Entity, error) { return loadHelix(t), nil },
		"EXTRUDED_SURFACE": func(t lldxf.Tags) (entity.Entity, error) { return loadExtrudedSurface(t), nil },
		"LOFTED_SURFACE":   func(t lldxf.Tags) (entity.Entity, error) { return loadLoftedSurface(t), nil },
		"REVOLVED_SURFACE": func(t lldxf.Tags) (entity.Entity, error) { return loadRevolvedSurface(t), nil },
		"SWEPT_SURFACE":    func(t lldxf.Tags) (entity.Entity, error) { return loadSweptSurface(t), nil },
		"PLANESURFACE":     func(t lldxf.Tags) (entity.Entity, error) { return loadPlaneSurface(t), nil },
	}
	for name, fn := range loaders {
		entity.Register(name, fn)
		if name == "DIMENSION" {
			// DXF variant record names share the dimension loader.
			for _, variant := range []string{"ARC_DIMENSION", "LARGE_RADIAL_DIMENSION"} {
				entity.Register(variant, fn)
			}
		}
	}
}

func loadPoint(tags lldxf.Tags) *entity.Point {
	p := entity.NewPoint()
	loadCommonAttribs(p, tags)
	for _, tag := range tags {
		switch tag.Code() {
		case 210:
			if v, ok := tagFloat(tag); ok {
				setExtrusionComponent(p, 0, v)
			}
		case 220:
			if v, ok := tagFloat(tag); ok {
				setExtrusionComponent(p, 1, v)
			}
		case 230:
			if v, ok := tagFloat(tag); ok {
				setExtrusionComponent(p, 2, v)
			}
		case 10:
			if v, ok := tagFloat(tag); ok {
				p.Coord[0] = v
			}
		case 20:
			if v, ok := tagFloat(tag); ok {
				p.Coord[1] = v
			}
		case 30:
			if v, ok := tagFloat(tag); ok {
				p.Coord[2] = v
			}
		}
	}
	return p
}

func loadLine(tags lldxf.Tags) *entity.Line {
	l := entity.NewLine()
	loadCommonAttribs(l, tags)
	for _, tag := range tags {
		switch tag.Code() {
		case 10:
			if v, ok := tagFloat(tag); ok {
				l.Start[0] = v
			}
		case 20:
			if v, ok := tagFloat(tag); ok {
				l.Start[1] = v
			}
		case 30:
			if v, ok := tagFloat(tag); ok {
				l.Start[2] = v
			}
		case 11:
			if v, ok := tagFloat(tag); ok {
				l.End[0] = v
			}
		case 21:
			if v, ok := tagFloat(tag); ok {
				l.End[1] = v
			}
		case 31:
			if v, ok := tagFloat(tag); ok {
				l.End[2] = v
			}
		case 39:
			if v, ok := tagFloat(tag); ok {
				l.Thickness = v
			}
		case 210:
			if v, ok := tagFloat(tag); ok {
				l.StretchingDirection[0] = v
				setExtrusionComponent(l, 0, v)
			}
		case 220:
			if v, ok := tagFloat(tag); ok {
				l.StretchingDirection[1] = v
				setExtrusionComponent(l, 1, v)
			}
		case 230:
			if v, ok := tagFloat(tag); ok {
				l.StretchingDirection[2] = v
				setExtrusionComponent(l, 2, v)
			}
		}
	}
	return l
}

func loadCircle(tags lldxf.Tags) *entity.Circle {
	c := entity.NewCircle()
	loadCommonAttribs(c, tags)
	for _, tag := range tags {
		switch tag.Code() {
		case 10:
			if v, ok := tagFloat(tag); ok {
				c.Center[0] = v
			}
		case 20:
			if v, ok := tagFloat(tag); ok {
				c.Center[1] = v
			}
		case 30:
			if v, ok := tagFloat(tag); ok {
				c.Center[2] = v
			}
		case 40:
			if v, ok := tagFloat(tag); ok {
				c.Radius = v
			}
		case 210:
			if v, ok := tagFloat(tag); ok {
				c.Direction[0] = v
				setExtrusionComponent(c, 0, v)
			}
		case 220:
			if v, ok := tagFloat(tag); ok {
				c.Direction[1] = v
				setExtrusionComponent(c, 1, v)
			}
		case 230:
			if v, ok := tagFloat(tag); ok {
				c.Direction[2] = v
				setExtrusionComponent(c, 2, v)
			}
		}
	}
	return c
}

func loadArc(tags lldxf.Tags) *entity.Arc {
	var center [3]float64
	var radius, startAngle, endAngle float64
	for _, tag := range tags {
		switch tag.Code() {
		case 10:
			if v, ok := tagFloat(tag); ok {
				center[0] = v
			}
		case 20:
			if v, ok := tagFloat(tag); ok {
				center[1] = v
			}
		case 30:
			if v, ok := tagFloat(tag); ok {
				center[2] = v
			}
		case 40:
			if v, ok := tagFloat(tag); ok {
				radius = v
			}
		case 50:
			if v, ok := tagFloat(tag); ok {
				startAngle = v
			}
		case 51:
			if v, ok := tagFloat(tag); ok {
				endAngle = v
			}
		}
	}
	c := entity.NewCircle()
	c.Center = center[:]
	c.Radius = radius
	for _, tag := range tags {
		switch tag.Code() {
		case 210:
			if v, ok := tagFloat(tag); ok {
				c.Direction[0] = v
				setExtrusionComponent(c, 0, v)
			}
		case 220:
			if v, ok := tagFloat(tag); ok {
				c.Direction[1] = v
				setExtrusionComponent(c, 1, v)
			}
		case 230:
			if v, ok := tagFloat(tag); ok {
				c.Direction[2] = v
				setExtrusionComponent(c, 2, v)
			}
		case 39:
			if v, ok := tagFloat(tag); ok {
				c.SetThickness(v)
			}
		}
	}
	a := entity.NewArc(c)
	a.Angle[0] = startAngle
	a.Angle[1] = endAngle
	loadCommonAttribs(a, tags)
	return a
}

func loadText(tags lldxf.Tags) *entity.Text {
	t := entity.NewText()
	loadCommonAttribs(t, tags)
	for _, tag := range tags {
		switch tag.Code() {
		case 10:
			if v, ok := tagFloat(tag); ok {
				t.Coord1[0] = v
			}
		case 20:
			if v, ok := tagFloat(tag); ok {
				t.Coord1[1] = v
			}
		case 30:
			if v, ok := tagFloat(tag); ok {
				t.Coord1[2] = v
			}
		case 11:
			if v, ok := tagFloat(tag); ok {
				t.Coord2[0] = v
			}
		case 21:
			if v, ok := tagFloat(tag); ok {
				t.Coord2[1] = v
			}
		case 31:
			if v, ok := tagFloat(tag); ok {
				t.Coord2[2] = v
			}
		case 40:
			if v, ok := tagFloat(tag); ok {
				t.Height = v
			}
		case 50:
			if v, ok := tagFloat(tag); ok {
				t.Rotation = v
			}
		case 41:
			if v, ok := tagFloat(tag); ok {
				t.WidthFactor = v
			}
		case 51:
			if v, ok := tagFloat(tag); ok {
				t.ObliqueAngle = v
			}
		case 1:
			if v, ok := tag.Value().(string); ok {
				t.Value = v
			}
		case 71:
			if v, ok := tagInt(tag); ok {
				t.GenFlag = v
			}
		case 72:
			if v, ok := tagInt(tag); ok {
				t.HorizontalFlag = v
			}
		case 73:
			if v, ok := tagInt(tag); ok {
				t.VerticalFlag = v
			}
		case 39:
			if v, ok := tagFloat(tag); ok {
				t.SetThickness(v)
			}
		case 210:
			if v, ok := tagFloat(tag); ok {
				setExtrusionComponent(t, 0, v)
			}
		case 220:
			if v, ok := tagFloat(tag); ok {
				setExtrusionComponent(t, 1, v)
			}
		case 230:
			if v, ok := tagFloat(tag); ok {
				setExtrusionComponent(t, 2, v)
			}
		}
	}
	return t
}

func loadMText(tags lldxf.Tags) *entity.MText {
	t := entity.NewMText()
	loadCommonAttribs(t, tags)
	for _, tag := range tags {
		switch tag.Code() {
		case 10:
			if v, ok := tagFloat(tag); ok {
				t.Coord1[0] = v
			}
		case 20:
			if v, ok := tagFloat(tag); ok {
				t.Coord1[1] = v
			}
		case 30:
			if v, ok := tagFloat(tag); ok {
				t.Coord1[2] = v
			}
		case 40:
			if v, ok := tagFloat(tag); ok {
				t.Height = v
			}
		case 50:
			if v, ok := tagFloat(tag); ok {
				t.Rotation = v
			}
		case 1:
			if v, ok := tag.Value().(string); ok {
				t.Value += v
			}
		case 3:
			if v, ok := tag.Value().(string); ok {
				t.AdditionalText = append(t.AdditionalText, v)
			}
		case 41:
			if v, ok := tagFloat(tag); ok {
				t.Width = v
			}
		case 44:
			if v, ok := tagFloat(tag); ok {
				t.LineSpacingFactor = v
			}
		case 46:
			if v, ok := tagFloat(tag); ok {
				t.DefinedHeight = v
			}
		case 71:
			if v, ok := tagInt(tag); ok {
				t.AttachmentPoint = entity.MTextAttachmentPoint(v)
			}
		case 72:
			if v, ok := tagInt(tag); ok {
				t.FlowDirection = entity.MTextFlowDirection(v)
			}
		case 73:
			if v, ok := tagInt(tag); ok {
				t.LineSpacingStyle = entity.MTextLineSpacingStyle(v)
			}
		case 210:
			if v, ok := tagFloat(tag); ok {
				setMTextExtrusion(t, 0, v)
			}
		case 220:
			if v, ok := tagFloat(tag); ok {
				setMTextExtrusion(t, 1, v)
			}
		case 230:
			if v, ok := tagFloat(tag); ok {
				setMTextExtrusion(t, 2, v)
			}
		}
	}
	return t
}

func setMTextExtrusion(t *entity.MText, idx int, v float64) {
	if len(t.Extrusion) < 3 {
		t.Extrusion = []float64{0, 0, 1}
	}
	t.Extrusion[idx] = v
}

func loadSolid(tags lldxf.Tags) *entity.Solid {
	s := entity.NewSolid()
	loadCommonAttribs(s, tags)
	for _, tag := range tags {
		switch tag.Code() {
		case 10:
			if v, ok := tagFloat(tag); ok {
				s.FirstPoint[0] = v
			}
		case 20:
			if v, ok := tagFloat(tag); ok {
				s.FirstPoint[1] = v
			}
		case 30:
			if v, ok := tagFloat(tag); ok {
				s.FirstPoint[2] = v
			}
		case 11:
			if v, ok := tagFloat(tag); ok {
				s.SecondPoint[0] = v
			}
		case 21:
			if v, ok := tagFloat(tag); ok {
				s.SecondPoint[1] = v
			}
		case 31:
			if v, ok := tagFloat(tag); ok {
				s.SecondPoint[2] = v
			}
		case 12:
			if v, ok := tagFloat(tag); ok {
				s.ThirdPoint[0] = v
			}
		case 22:
			if v, ok := tagFloat(tag); ok {
				s.ThirdPoint[1] = v
			}
		case 32:
			if v, ok := tagFloat(tag); ok {
				s.ThirdPoint[2] = v
			}
		case 13:
			if v, ok := tagFloat(tag); ok {
				s.FourthPoint[0] = v
			}
		case 23:
			if v, ok := tagFloat(tag); ok {
				s.FourthPoint[1] = v
			}
		case 33:
			if v, ok := tagFloat(tag); ok {
				s.FourthPoint[2] = v
			}
		}
	}
	return s
}

func load3DFace(tags lldxf.Tags) *entity.ThreeDFace {
	f := entity.New3DFace()
	loadCommonAttribs(f, tags)
	for _, tag := range tags {
		switch tag.Code() {
		case 10:
			if v, ok := tagFloat(tag); ok {
				f.Points[0][0] = v
			}
		case 20:
			if v, ok := tagFloat(tag); ok {
				f.Points[0][1] = v
			}
		case 30:
			if v, ok := tagFloat(tag); ok {
				f.Points[0][2] = v
			}
		case 11:
			if v, ok := tagFloat(tag); ok {
				f.Points[1][0] = v
			}
		case 21:
			if v, ok := tagFloat(tag); ok {
				f.Points[1][1] = v
			}
		case 31:
			if v, ok := tagFloat(tag); ok {
				f.Points[1][2] = v
			}
		case 12:
			if v, ok := tagFloat(tag); ok {
				f.Points[2][0] = v
			}
		case 22:
			if v, ok := tagFloat(tag); ok {
				f.Points[2][1] = v
			}
		case 32:
			if v, ok := tagFloat(tag); ok {
				f.Points[2][2] = v
			}
		case 13:
			if v, ok := tagFloat(tag); ok {
				f.Points[3][0] = v
			}
		case 23:
			if v, ok := tagFloat(tag); ok {
				f.Points[3][1] = v
			}
		case 33:
			if v, ok := tagFloat(tag); ok {
				f.Points[3][2] = v
			}
		}
	}
	return f
}

func loadLwPolyline(tags lldxf.Tags) *entity.LwPolyline {
	var numVertices int
	var closed bool
	var constWidth float64
	vertices := make([][2]float64, 0)
	bulges := make([]float64, 0)
	for _, tag := range tags {
		switch tag.Code() {
		case 90:
			if v, ok := tagInt(tag); ok {
				numVertices = v
			}
		case 70:
			if v, ok := tagInt(tag); ok && v&1 != 0 {
				closed = true
			}
		case 43:
			if v, ok := tagFloat(tag); ok {
				constWidth = v
			}
		case 10:
			if v, ok := tagFloat(tag); ok {
				vertices = append(vertices, [2]float64{v, 0})
			}
		case 20:
			if v, ok := tagFloat(tag); ok {
				if len(vertices) > 0 {
					vertices[len(vertices)-1][1] = v
				}
			}
		case 42:
			if v, ok := tagFloat(tag); ok {
				// A bulge belongs to the vertex that precedes it.
				idx := max(len(vertices)-1, 0)
				for len(bulges) <= idx {
					bulges = append(bulges, 0)
				}
				bulges[idx] = v
			}
		}
	}
	lw := entity.NewLwPolyline(numVertices)
	for i, v := range vertices {
		if i < numVertices {
			lw.Vertices[i][0] = v[0]
			lw.Vertices[i][1] = v[1]
		}
	}
	for i, b := range bulges {
		if i < numVertices {
			lw.Bulges[i] = b
		}
	}
	lw.ConstantWidth = constWidth
	if closed {
		lw.Close()
	}
	loadCommonAttribs(lw, tags)
	return lw
}

func loadEllipse(tags lldxf.Tags) *entity.Ellipse {
	e := entity.NewEllipse()
	loadCommonAttribs(e, tags)
	for _, tag := range tags {
		switch tag.Code() {
		case 10:
			if v, ok := tagFloat(tag); ok {
				e.Center[0] = v
			}
		case 20:
			if v, ok := tagFloat(tag); ok {
				e.Center[1] = v
			}
		case 30:
			if v, ok := tagFloat(tag); ok {
				e.Center[2] = v
			}
		case 40:
			if v, ok := tagFloat(tag); ok {
				e.Major = v
			}
		case 41:
			if v, ok := tagFloat(tag); ok {
				e.Minor = v
			}
		case 200:
			if v, ok := tagFloat(tag); ok {
				e.Ratio = v
			}
		case 50:
			if v, ok := tagFloat(tag); ok {
				e.Start = v
			}
		case 51:
			if v, ok := tagFloat(tag); ok {
				e.End = v
			}
		}
	}
	return e
}

func loadPolyline(tags lldxf.Tags) *entity.Polyline {
	p := entity.NewPolyline()
	loadCommonAttribs(p, tags)
	for _, tag := range tags {
		switch tag.Code() {
		case 70:
			if v, ok := tagInt(tag); ok {
				p.Flag = v
			}
		}
	}
	return p
}

func loadVertex(tags lldxf.Tags) *entity.Vertex {
	var x, y, z float64
	for _, tag := range tags {
		switch tag.Code() {
		case 10:
			if v, ok := tagFloat(tag); ok {
				x = v
			}
		case 20:
			if v, ok := tagFloat(tag); ok {
				y = v
			}
		case 30:
			if v, ok := tagFloat(tag); ok {
				z = v
			}
		}
	}
	v := entity.NewVertex(x, y, z)
	for _, tag := range tags {
		switch tag.Code() {
		case 70:
			if fv, ok := tagInt(tag); ok {
				v.Flag = fv
			}
		case 42:
			if fv, ok := tagFloat(tag); ok {
				v.Bulge = fv
			}
		case 40:
			if fv, ok := tagFloat(tag); ok {
				v.StartWidth = fv
			}
		case 41:
			if fv, ok := tagFloat(tag); ok {
				v.EndWidth = fv
			}
		case 71:
			if fv, ok := tagInt(tag); ok {
				v.Extra = fv
			}
		case 72:
			if fv, ok := tagInt(tag); ok {
				v.EdgeFlags = fv
			}
		}
	}
	loadCommonAttribs(v, tags)
	return v
}

func loadInsert(tags lldxf.Tags) *entity.Insert {
	i := entity.NewInsert()
	loadCommonAttribs(i, tags)
	xclip, rest := extractXClipTags(tags)
	tags = rest
	var x, y, z float64
	for _, tag := range tags {
		switch tag.Code() {
		case 10:
			if v, ok := tagFloat(tag); ok {
				x = v
			}
		case 20:
			if v, ok := tagFloat(tag); ok {
				y = v
			}
		case 30:
			if v, ok := tagFloat(tag); ok {
				z = v
			}
		case 2:
			if v, ok := tag.Value().(string); ok {
				i.SetBlockName(v)
			}
		case 50:
			if v, ok := tagFloat(tag); ok {
				i.SetRotation(v)
			}
		case 41:
			if v, ok := tagFloat(tag); ok {
				i.SetScaleX(v)
			}
		case 42:
			if v, ok := tagFloat(tag); ok {
				i.SetScaleY(v)
			}
		case 43:
			if v, ok := tagFloat(tag); ok {
				i.SetScaleZ(v)
			}
		case 70:
			if v, ok := tagInt(tag); ok {
				i.SetFlags(v)
			}
		case 71:
			if v, ok := tagInt(tag); ok {
				i.SetMultiple(true)
				i.SetGrid(v, i.RowCount(), i.ColumnSpacing(), i.RowSpacing())
			}
		case 44:
			if v, ok := tagFloat(tag); ok {
				i.SetGrid(i.ColumnCount(), i.RowCount(), v, i.RowSpacing())
			}
		case 72:
			if v, ok := tagInt(tag); ok {
				i.SetMultiple(true)
				i.SetGrid(i.ColumnCount(), v, i.ColumnSpacing(), i.RowSpacing())
			}
		case 45:
			if v, ok := tagFloat(tag); ok {
				i.SetGrid(i.ColumnCount(), i.RowCount(), i.ColumnSpacing(), v)
			}
		case 210:
			if v, ok := tagFloat(tag); ok {
				i.SetExtrusionComponent(0, v)
			}
		case 220:
			if v, ok := tagFloat(tag); ok {
				i.SetExtrusionComponent(1, v)
			}
		case 230:
			if v, ok := tagFloat(tag); ok {
				i.SetExtrusionComponent(2, v)
			}
		}
	}
	i.SetInsertPoint(dxfmath.NewVec3(x, y, z))
	applyXClipBoundary(i, xclip)
	return i
}

// extractXClipTags pulls the 102 {XCLIP_BOUNDS} group out of the tag stream,
// returning its inner tags and the remaining tags.
func extractXClipTags(tags lldxf.Tags) (inner, rest lldxf.Tags) {
	inGroup := false
	for _, tag := range tags {
		if tag.Code() == 102 {
			if s, ok := tag.Value().(string); ok {
				if !inGroup && s == "{XCLIP_BOUNDS" {
					inGroup = true
					continue
				}
				if inGroup && s == "}" {
					inGroup = false
					continue
				}
			}
		}
		if inGroup {
			inner = append(inner, tag)
		} else {
			rest = append(rest, tag)
		}
	}
	return inner, rest
}

// applyXClipBoundary builds the clip boundary of an insert from the tags of
// its XCLIP group.
func applyXClipBoundary(i *entity.Insert, tags lldxf.Tags) {
	typ := 0
	var pts [][2]float64
	var px float64
	var hasX bool
	for _, tag := range tags {
		switch tag.Code() {
		case 71:
			if v, ok := tagInt(tag); ok {
				typ = v
			}
		case 10:
			if v, ok := tagFloat(tag); ok {
				px = v
				hasX = true
			}
		case 20:
			if v, ok := tagFloat(tag); ok && hasX {
				pts = append(pts, [2]float64{px, v})
				hasX = false
			}
		}
	}
	switch {
	case typ == entity.XClipRectangular && len(pts) >= 2:
		i.SetXClipRect(pts[0][0], pts[0][1], pts[1][0], pts[1][1])
	case len(pts) >= 2:
		i.SetXClipPolyline(pts)
	}
}

func loadTrace(tags lldxf.Tags) *entity.Trace {
	points := make([]float64, 12)
	for _, tag := range tags {
		switch tag.Code() {
		case 10:
			if v, ok := tagFloat(tag); ok {
				points[0] = v
			}
		case 20:
			if v, ok := tagFloat(tag); ok {
				points[1] = v
			}
		case 11:
			if v, ok := tagFloat(tag); ok {
				points[2] = v
			}
		case 21:
			if v, ok := tagFloat(tag); ok {
				points[3] = v
			}
		case 12:
			if v, ok := tagFloat(tag); ok {
				points[4] = v
			}
		case 22:
			if v, ok := tagFloat(tag); ok {
				points[5] = v
			}
		case 13:
			if v, ok := tagFloat(tag); ok {
				points[6] = v
			}
		case 23:
			if v, ok := tagFloat(tag); ok {
				points[7] = v
			}
		}
	}
	t := entity.NewTrace(points, 0, 0)
	loadCommonAttribs(t, tags)
	return t
}

func loadSpline(tags lldxf.Tags) *entity.Spline {
	s := entity.NewSpline(3)
	loadCommonAttribs(s, tags)
	for _, tag := range tags {
		switch tag.Code() {
		case 70:
			if v, ok := tagInt(tag); ok {
				s.Flag = v
			}
		case 71:
			if v, ok := tagInt(tag); ok {
				s.Degree = v
			}
		case 40:
			if v, ok := tagFloat(tag); ok {
				s.Knots = append(s.Knots, v)
			}
		case 41:
			if v, ok := tagFloat(tag); ok {
				s.Weights = append(s.Weights, v)
			}
		case 42:
			if v, ok := tagFloat(tag); ok {
				appendTolerance(&s.Tolerance, 0, v)
			}
		case 43:
			if v, ok := tagFloat(tag); ok {
				appendTolerance(&s.Tolerance, 1, v)
			}
		case 44:
			if v, ok := tagFloat(tag); ok {
				appendTolerance(&s.Tolerance, 2, v)
			}
		case 10:
			if v, ok := tagFloat(tag); ok {
				s.Controls = append(s.Controls, []float64{v, 0, 0})
			}
		case 20:
			if len(s.Controls) > 0 {
				if v, ok := tagFloat(tag); ok {
					s.Controls[len(s.Controls)-1][1] = v
				}
			}
		case 30:
			if len(s.Controls) > 0 {
				if v, ok := tagFloat(tag); ok {
					s.Controls[len(s.Controls)-1][2] = v
				}
			}
		case 11:
			if v, ok := tagFloat(tag); ok {
				s.Fits = append(s.Fits, []float64{v, 0, 0})
			}
		case 21:
			if len(s.Fits) > 0 {
				if v, ok := tagFloat(tag); ok {
					s.Fits[len(s.Fits)-1][1] = v
				}
			}
		case 31:
			if len(s.Fits) > 0 {
				if v, ok := tagFloat(tag); ok {
					s.Fits[len(s.Fits)-1][2] = v
				}
			}
		case 210:
			if v, ok := tagFloat(tag); ok {
				s.Normal[0] = v
			}
		case 220:
			if v, ok := tagFloat(tag); ok {
				s.Normal[1] = v
			}
		case 230:
			if v, ok := tagFloat(tag); ok {
				s.Normal[2] = v
			}
		}
	}
	return s
}

func appendTolerance(t *[]float64, idx int, v float64) {
	for len(*t) <= idx {
		*t = append(*t, 0)
	}
	(*t)[idx] = v
}

func loadLeader(tags lldxf.Tags) *entity.Leader {
	l := entity.NewLeader()
	loadCommonAttribs(l, tags)

	var currentPoint []float64
	for _, tag := range tags {
		switch tag.Code() {
		case 10:
			currentPoint = []float64{0, 0, 0}
			if v, ok := tagFloat(tag); ok {
				currentPoint[0] = v
			}
		case 20:
			if v, ok := tagFloat(tag); ok {
				if currentPoint == nil {
					currentPoint = []float64{0, 0, 0}
				}
				currentPoint[1] = v
			}
		case 30:
			if v, ok := tagFloat(tag); ok {
				if currentPoint == nil {
					currentPoint = []float64{0, 0, 0}
				}
				currentPoint[2] = v
			}
			if currentPoint != nil {
				l.AddPoint(currentPoint[0], currentPoint[1], currentPoint[2])
			}
		case 71:
			if v, ok := tagInt(tag); ok {
				l.SetAnnotationType(v)
			}
		case 72:
			if v, ok := tagInt(tag); ok {
				l.SetPathType(v)
			}
		case 3:
			if v, ok := tag.Value().(string); ok {
				l.SetDimstyle(v)
			}
		case 40:
			if v, ok := tagFloat(tag); ok {
				l.SetArrowhead(0, v, 0)
			}
		case 73:
			if v, ok := tagInt(tag); ok {
				if l.GetAnnotationType() == 1 {
					l.SetTextProperties(0, 0, "", 0, v)
				}
			}
		case 74:
			if v, ok := tagInt(tag); ok {
				if l.GetAnnotationType() == 1 {
					l.SetTextProperties(0, 0, "", 0, 0)
					l.SetTextProperties(0, 0, "", 0, v)
				} else {
					l.SetHookLine(true, v)
				}
			}
		case 75:
			if v, ok := tagInt(tag); ok {
				l.SetHookLine(v == 1, 0)
			}
		case 210:
			if v, ok := tagFloat(tag); ok {
				l.SetExtrusion(v, 0, 0)
			}
		}
	}
	return l
}

func loadHatch(tags lldxf.Tags) *entity.Hatch {
	h := entity.NewHatch()
	loadCommonAttribs(h, tags)

	var patternType int
	patternName := ""
	var lines []entity.HatchPatternLine
	var solidFill bool
	// boundary path parsing state
	var paths []entity.BoundaryPath
	var curPath *entity.PolylinePath
	inBoundary := false
	isPolylinePath := false
	seenBulgeFlag := false
	seenClosedFlag := false
	closedVal := false
	bulgeVal := false

	inPattern := false
	seenPathHeader := false
	isMPolygon := strings.EqualFold(tags.DXFType(), "MPOLYGON")

	for _, tag := range tags {
		switch tag.Code() {
		case 2:
			if v, ok := tag.Value().(string); ok && !inBoundary {
				patternName = v
			}
		case 70:
			if v, ok := tagInt(tag); ok {
				if isMPolygon {
					// MPOLYGON path-type flags (AcDbMPolygon code 70) start a
					// new boundary path; the vertex count follows as 93.
					if curPath != nil {
						paths = append(paths, curPath)
						curPath = nil
					}
					curPath = entity.NewPolylinePath(v, false, false)
					isPolylinePath = true
					seenPathHeader = true
				} else if !inBoundary {
					h.SetStyle(v)
				}
			}
		case 76:
			if v, ok := tagInt(tag); ok && !inBoundary {
				patternType = v
			}

		case 71:
			if v, ok := tagInt(tag); ok && !inBoundary {
				solidFill = v != 0
			}
		case 78: // pattern line count; enter the pattern definition section
			if curPath != nil {
				paths = append(paths, curPath)
				curPath = nil
			}
			inBoundary = false
			seenPathHeader = false
			inPattern = true
		case 53: // pattern line angle starts a new line definition
			if v, ok := tagFloat(tag); ok && inPattern {
				lines = append(lines, entity.HatchPatternLine{Angle: v})
			}
		case 97: // source object count (boundary section; skip)
		case 43: // base point x of the last pattern line
			if v, ok := tagFloat(tag); ok && len(lines) > 0 && inPattern {
				l := &lines[len(lines)-1]
				l.BasePoint = dxfmath.NewVec2(v, l.BasePoint.Y())
			}
		case 44: // base point y (pattern section) or dash length (boundary section)
			if v, ok := tagFloat(tag); ok {
				if (inBoundary || seenPathHeader) && curPath != nil {
					curPath.SetDashLength(v)
				} else if len(lines) > 0 && inPattern {
					l := &lines[len(lines)-1]
					l.BasePoint = dxfmath.NewVec2(l.BasePoint.X(), v)
				}
			}
		case 45: // offset x of the last pattern line
			if v, ok := tagFloat(tag); ok && len(lines) > 0 && inPattern {
				l := &lines[len(lines)-1]
				l.Offset = dxfmath.NewVec2(v, l.Offset.Y())
			}
		case 46: // offset y of the last pattern line
			if v, ok := tagFloat(tag); ok && len(lines) > 0 && inPattern {
				l := &lines[len(lines)-1]
				l.Offset = dxfmath.NewVec2(l.Offset.X(), v)
			}
		case 75: // hatch style flag (AcDbHatch code 75)
			if v, ok := tagInt(tag); ok && !inBoundary {
				h.SetStyle(v)
			}
		case 79: // dash count marker; dashes follow as 49s
			if len(lines) > 0 {
				lines[len(lines)-1].DashLengthItems = nil
			}
		case 49:
			if v, ok := tagFloat(tag); ok && len(lines) > 0 && inPattern {
				l := &lines[len(lines)-1]
				l.DashLengthItems = append(l.DashLengthItems, v)
			}
		case 91: // boundary path count; enter the boundary section (HATCH only;
			// MPOLYGON uses code 70 per path with a 93 vertex count)
			if isMPolygon {
				break
			}
			inBoundary = true
			inPattern = false
		case 72: // has-bulge flag (code 72) inside a boundary path
			if (inBoundary || seenPathHeader) && isPolylinePath && !seenBulgeFlag {
				seenBulgeFlag = true
				bulge := false
				bulgeVal = false
				if v, ok := tagInt(tag); ok {
					bulge = v != 0
					bulgeVal = bulge
				}
				if curPath != nil {
					curPath.SetHasBulge(bulge)
				} else {
					curPath = entity.NewPolylinePath(0, closedVal, bulge)
				}
			}
		case 92: // MPOLYGON: after the path header this carries the closed flag
			if isMPolygon && seenPathHeader {
				if v, ok := tagInt(tag); ok && curPath != nil {
					curPath.SetClosed(v != 0)
				}
				break
			}
			if inBoundary && curPath != nil {
				paths = append(paths, curPath)
				curPath = nil
			}
			isPolylinePath = false
			if v, ok := tagInt(tag); ok {
				isPolylinePath = v == entity.BoundaryPathTypePolyline
			}
		case 73: // closed flag (code 73) inside a boundary path
			if (inBoundary || seenPathHeader) && isPolylinePath && !seenClosedFlag {
				seenClosedFlag = true
				if v, ok := tagInt(tag); ok {
					closedVal = v != 0
				}
				if curPath != nil {
					curPath.SetClosed(closedVal)
				} else {
					curPath = entity.NewPolylinePath(0, closedVal, seenBulgeFlag && bulgeVal)
				}
			}
		case 93: // vertex count marker (polyline paths only; skip)
			if isMPolygon {
				// MPOLYGON: 93 counts the vertices of the path that follows.
				seenBulgeFlag = false
				seenClosedFlag = false
				seenPathHeader = true
			}
		case 10, 20, 30, 42:
			if curPath != nil {
				v, _ := tagFloat(tag)
				switch int(tag.Code()) {
				case 10:
					curPath.AddVertex(entity.NewVertex3D(v, 0, 0, 0))
				case 20:
					curPath.CompleteVertexY(v)
				case 30:
					curPath.CompleteVertexZ(v)
				case 42:
					curPath.CompleteVertexBulge(v)
				}
			}
		case 330:
			if curPath != nil {
				if v, ok := tag.Value().(string); ok {
					curPath.AddSourceHandle(v)
				}
			}
		}
	}
	if curPath != nil {
		paths = append(paths, curPath)
	}
	if len(lines) == 0 && !solidFill {
		lines = entity.DefaultPatternLines()
	}
	h.SetPattern(patternType, patternName, lines)
	h.SetSolidFill(solidFill)
	for _, p := range paths {
		h.AddBoundaryPath(p)
	}
	return h
}

func loadImage(tags lldxf.Tags) *entity.Image {
	img := entity.NewImage()
	loadCommonAttribs(img, tags)

	var uVec, vVec dxfmath.Vec3
	brightness, contrast, fade := img.Brightness(), img.Contrast(), img.Fade()
	var boundary []entity.ImageBoundaryPoint
	var pendingX float64
	var hasPendingX bool
	for _, tag := range tags {
		switch tag.Code() {
		case 10:
			if v, ok := tagFloat(tag); ok {
				img.SetInsertPoint(dxfmath.NewVec3(v, 0, 0))
			}
		case 20:
			if v, ok := tagFloat(tag); ok {
				pt := img.InsertPoint()
				img.SetInsertPoint(dxfmath.NewVec3(pt.X(), v, pt.Z()))
			}
		case 30:
			if v, ok := tagFloat(tag); ok {
				pt := img.InsertPoint()
				img.SetInsertPoint(dxfmath.NewVec3(pt.X(), pt.Y(), v))
			}
		case 11:
			if v, ok := tagFloat(tag); ok {
				uVec = dxfmath.NewVec3(v, 0, 0)
			}
		case 21:
			if v, ok := tagFloat(tag); ok {
				uVec = dxfmath.NewVec3(uVec.X(), v, 0)
			}
		case 31:
			if v, ok := tagFloat(tag); ok {
				uVec = dxfmath.NewVec3(uVec.X(), uVec.Y(), v)
			}
		case 12:
			if v, ok := tagFloat(tag); ok {
				vVec = dxfmath.NewVec3(v, 0, 0)
			}
		case 22:
			if v, ok := tagFloat(tag); ok {
				vVec = dxfmath.NewVec3(vVec.X(), v, 0)
			}
		case 32:
			if v, ok := tagFloat(tag); ok {
				vVec = dxfmath.NewVec3(vVec.X(), vVec.Y(), v)
			}
		case 13:
			if v, ok := tagFloat(tag); ok {
				img.SetImageSize(v, 0)
			}
		case 23:
			if v, ok := tagFloat(tag); ok {
				size := img.ImageSize()
				img.SetImageSize(size.X(), v)
			}
		case 70:
			if v, ok := tagInt(tag); ok {
				img.SetFlags(v)
			}
		case 280:
			if v, ok := tagInt(tag); ok {
				img.SetClipping(v != 0)
			}
		case 281:
			if v, ok := tagFloat(tag); ok {
				brightness = v
			}
		case 282:
			if v, ok := tagFloat(tag); ok {
				contrast = v
			}
		case 283:
			if v, ok := tagFloat(tag); ok {
				fade = v
			}
		case 71:
			if v, ok := tagInt(tag); ok {
				img.SetClippingBoundaryType(v)
			}
		case 91:
			// boundary point count: pairs follow as 14/24
		case 14:
			if v, ok := tagFloat(tag); ok {
				pendingX = v
				hasPendingX = true
			}
		case 24:
			if v, ok := tagFloat(tag); ok && hasPendingX {
				boundary = append(boundary, entity.ImageBoundaryPoint{X: pendingX, Y: v})
				hasPendingX = false
			}
		case 290:
			if v, ok := tagInt(tag); ok {
				img.SetClipMode(v)
			}
		case 360:
			if v, ok := tag.Value().(string); ok {
				img.SetImageDefReactorHandle(v)
			}
		case 62:
			if v, ok := tagInt(tag); ok {
				img.SetColor(color.ColorNumber(v))
			}
		case 340:
			if v, ok := tag.Value().(string); ok {
				img.SetImageDefHandle(v)
			}
		}
	}
	if uVec.X() != 0 || uVec.Y() != 0 {
		img.SetUVectors(uVec, vVec)
	}
	img.SetDisplayProperties(brightness, contrast, fade)
	if len(boundary) > 0 {
		if img.ClippingBoundaryType() == entity.ImageClippingRectangular && len(boundary) >= 2 {
			img.SetRectangularClipping(boundary[0].X, boundary[0].Y, boundary[1].X, boundary[1].Y)
		} else {
			img.SetPolygonClipping(boundary)
		}
	}
	return img
}

func loadMLeader(tags lldxf.Tags) *entity.MLeader {
	m := entity.NewMLeader()
	loadCommonAttribs(m, tags)

	var pts [][3]float64
	var cur [3]float64
	started := false
	for _, tag := range tags {
		switch tag.Code() {
		case 170:
			if v, ok := tagInt(tag); ok {
				m.Version = v
			}
		case 340:
			if v, ok := tag.Value().(string); ok {
				m.SetStyleHandle(v)
			}
		case 140:
			if v, ok := tagFloat(tag); ok {
				m.AnnotationScale = v
			}
		case 171:
			if v, ok := tagInt(tag); ok {
				m.ContextType = v
			}
		case 172:
			if v, ok := tagInt(tag); ok {
				m.BitFlags = v
			}
		case 41:
			if v, ok := tagFloat(tag); ok {
				m.Scale = v
			}
		case 290:
			if v, ok := tagInt(tag); ok {
				m.ArrowheadEnabled = v != 0
			}
		case 173:
			if v, ok := tagInt(tag); ok {
				m.ArrowheadType = v
			}
		case 40:
			if v, ok := tagFloat(tag); ok {
				m.ArrowheadSize = v
			}
		case 291:
			if v, ok := tagInt(tag); ok {
				m.HasLanding = v != 0
			}
		case 44:
			if v, ok := tagFloat(tag); ok {
				m.LandingDistance = v
			}
		case 45:
			if v, ok := tagFloat(tag); ok {
				m.LandingGap = v
			}
		case 292:
			if v, ok := tagInt(tag); ok {
				m.HasDogleg = v != 0
			}
		case 46:
			if v, ok := tagFloat(tag); ok {
				m.HookLineDirection = v
			}
		case 47:
			if v, ok := tagFloat(tag); ok {
				m.HookLineLength = v
			}
		case 175:
			if v, ok := tagInt(tag); ok {
				m.ContentType = v
			}
		case 304:
			if v, ok := tag.Value().(string); ok {
				m.Text.Content = v
			}
		case 176:
			if v, ok := tagInt(tag); ok {
				m.HorizontalAttachment = v
			}
		case 177:
			if v, ok := tagInt(tag); ok {
				m.VerticalAttachment = v
			}
		case 90:
			if v, ok := tagInt(tag); ok {
				m.Color = color.ColorNumber(v)
			}
		case 10:
			if v, ok := tagFloat(tag); ok {
				if started {
					pts = append(pts, cur)
				}
				cur = [3]float64{v, 0, 0}
				started = true
			}
		case 20:
			if v, ok := tagFloat(tag); ok && started {
				cur[1] = v
			}
		case 30:
			if v, ok := tagFloat(tag); ok && started {
				cur[2] = v
			}
		}
	}
	if started {
		pts = append(pts, cur)
	}
	if len(pts) > 0 {
		line := entity.LeaderLine{
			Type:     entity.MLeaderLineTypeStraight,
			Vertices: make([]entity.LeaderVertex, 0, len(pts)),
		}
		for _, p := range pts {
			line.Vertices = append(line.Vertices, entity.LeaderVertex{Point: []float64{p[0], p[1], p[2]}})
		}
		m.Lines = []entity.LeaderLine{line}
	}
	return m
}

func loadDimension(tags lldxf.Tags) *entity.Dimension {
	d := entity.NewDimension()
	loadCommonAttribs(d, tags)

	// Point accumulators: 10 (definition), 11 (text midpoint), 13/14
	// (extension line points) with y/z following the same slot.
	var dp, tp, ep1, ep2 [3]float64
	setComp := func(acc *[3]float64, comp int, tag lldxf.Tag) {
		if v, ok := tagFloat(tag); ok {
			acc[comp] = v
		}
	}

	for _, tag := range tags {
		switch int(tag.Code()) {
		case 10:
			setComp(&dp, 0, tag)
		case 20:
			setComp(&dp, 1, tag)
		case 30:
			setComp(&dp, 2, tag)
		case 11:
			setComp(&tp, 0, tag)
		case 21:
			setComp(&tp, 1, tag)
		case 31:
			setComp(&tp, 2, tag)
		case 13:
			setComp(&ep1, 0, tag)
		case 23:
			setComp(&ep1, 1, tag)
		case 33:
			setComp(&ep1, 2, tag)
		case 14:
			setComp(&ep2, 0, tag)
		case 24:
			setComp(&ep2, 1, tag)
		case 34:
			setComp(&ep2, 2, tag)
		case 1:
			if v, ok := tag.Value().(string); ok {
				d.SetText(v)
			}
		case 2:
			if v, ok := tag.Value().(string); ok {
				d.SetUserText(v, "")
			}
		case 3:
			if v, ok := tag.Value().(string); ok {
				d.SetStyle(v)
			}
		case 50:
			if v, ok := tagFloat(tag); ok {
				if d.IsArcLength() {
					d.SetArcStartAngle(v)
				} else {
					d.SetRotation(v)
				}
			}
		case 51:
			if v, ok := tagFloat(tag); ok {
				if d.IsArcLength() {
					d.SetArcEndAngle(v)
				} else {
					d.SetObliqueAngle(v)
				}
			}
		case 70:
			if v, ok := tagInt(tag); ok {
				d.SetDimensionType(v & 15)
				if strings.EqualFold(tags.DXFType(), "ARC_DIMENSION") {
					d.SetDimensionType(entity.DimensionTypeArc)
				} else if strings.EqualFold(tags.DXFType(), "LARGE_RADIAL_DIMENSION") {
					d.SetDimensionType(entity.DimensionTypeLargeRadial)
				}
			}
		case 71:
			if v, ok := tagInt(tag); ok {
				d.SetAttachmentPoint(v)
			}
		case 72:
			if v, ok := tagInt(tag); ok {
				d.SetLineSpacingStyle(v)
			}
		case 41:
			if v, ok := tagFloat(tag); ok {
				d.SetLineSpacingFactor(v)
			}
		case 42:
			if v, ok := tagFloat(tag); ok {
				d.SetActualMeasurement(v)
			}
		case 40:
			if v, ok := tagFloat(tag); ok && d.IsArcLength() {
				d.SetArcRadius(v)
			}
		case 210, 220, 230:
			setDimExtrusion(d, int(tag.Code())-210, tag)
		}
	}

	d.SetDefinitionPoint([]float64{dp[0], dp[1], dp[2]})
	d.SetTextMidpoint([]float64{tp[0], tp[1], tp[2]})
	d.SetExtensionPoints([]float64{ep1[0], ep1[1], ep1[2]}, []float64{ep2[0], ep2[1], ep2[2]})
	return d
}

// setDimExtrusion sets one component of a dimension's extrusion direction.
func setDimExtrusion(d *entity.Dimension, idx int, tag lldxf.Tag) {
	v, ok := tagFloat(tag)
	if !ok || idx < 0 || idx > 2 {
		return
	}
	cur := d.Extrusion()
	e := []float64{cur[0], cur[1], cur[2]}
	e[idx] = v
	d.SetExtrusion(e)
}

func loadShape(tags lldxf.Tags) *entity.Shape {
	s := entity.NewShape()
	loadCommonAttribs(s, tags)

	for _, tag := range tags {
		switch tag.Code() {
		case 10:
			if v, ok := tagFloat(tag); ok {
				s.SetInsertPoint(dxfmath.NewVec3(v, 0, 0))
			}
		case 20:
			if v, ok := tagFloat(tag); ok {
				pt := s.GetInsertPoint()
				s.SetInsertPoint(dxfmath.NewVec3(pt.X(), v, pt.Z()))
			}
		case 30:
			if v, ok := tagFloat(tag); ok {
				pt := s.GetInsertPoint()
				s.SetInsertPoint(dxfmath.NewVec3(pt.X(), pt.Y(), v))
			}
		case 2:
			if v, ok := tag.Value().(string); ok {
				s.SetShapeName(v)
			}
		case 40:
			if v, ok := tagFloat(tag); ok {
				s.SetSize(v)
			}
		case 50:
			if v, ok := tagFloat(tag); ok {
				s.SetRotation(v)
			}
		case 41:
			if v, ok := tagFloat(tag); ok {
				s.SetXScale(v)
			}
		case 51:
			if v, ok := tagFloat(tag); ok {
				s.SetOblique(v)
			}
		}
	}
	return s
}

func loadXLine(tags lldxf.Tags) *entity.XLine {
	x := entity.NewXLine()
	loadCommonAttribs(x, tags)

	var sp, uv [3]float64
	for _, tag := range tags {
		switch tag.Code() {
		case 10, 20, 30:
			if v, ok := tagFloat(tag); ok {
				sp[int(tag.Code())/10-1] = v
			}
		case 11, 21, 31:
			if v, ok := tagFloat(tag); ok {
				uv[(int(tag.Code())-1)/10-1] = v
			}
		}
	}
	x.SetStart(dxfmath.NewVec3(sp[0], sp[1], sp[2]))
	x.SetUnitVector(dxfmath.NewVec3(uv[0], uv[1], uv[2]))
	return x
}

func loadRay(tags lldxf.Tags) *entity.Ray {
	r := entity.NewRay()
	loadCommonAttribs(r, tags)

	var sp, uv [3]float64
	for _, tag := range tags {
		switch tag.Code() {
		case 10, 20, 30:
			if v, ok := tagFloat(tag); ok {
				sp[int(tag.Code())/10-1] = v
			}
		case 11, 21, 31:
			if v, ok := tagFloat(tag); ok {
				uv[(int(tag.Code())-1)/10-1] = v
			}
		}
	}
	r.SetStart(dxfmath.NewVec3(sp[0], sp[1], sp[2]))
	r.SetUnitVector(dxfmath.NewVec3(uv[0], uv[1], uv[2]))
	return r
}

func loadWipeout(tags lldxf.Tags) *entity.Wipeout {
	w := entity.NewWipeout()
	loadCommonAttribs(w, tags)

	var pendingX float64
	haveX := false

	for _, tag := range tags {
		switch tag.Code() {
		case 10:
			if v, ok := tagFloat(tag); ok {
				w.SetInsertPoint(dxfmath.NewVec3(v, 0, 0))
			}
		case 20:
			if v, ok := tagFloat(tag); ok {
				pt := w.InsertPoint()
				w.SetInsertPoint(dxfmath.NewVec3(pt.X(), v, pt.Z()))
			}
		case 30:
			if v, ok := tagFloat(tag); ok {
				pt := w.InsertPoint()
				w.SetInsertPoint(dxfmath.NewVec3(pt.X(), pt.Y(), v))
			}
		case 11:
			if v, ok := tagFloat(tag); ok {
				uv := w.UVector()
				vv := w.VVector()
				w.SetUVectors(dxfmath.NewVec3(v, uv.Y(), uv.Z()), vv)
			}
		case 12:
			if v, ok := tagFloat(tag); ok {
				uv := w.UVector()
				vv := w.VVector()
				w.SetUVectors(uv, dxfmath.NewVec3(v, vv.Y(), vv.Z()))
			}
		case 13:
			if v, ok := tagFloat(tag); ok {
				sz := w.ImageSize()
				w.SetSize(v, sz.Y())
			}
		case 23:
			if v, ok := tagFloat(tag); ok {
				sz := w.ImageSize()
				w.SetSize(sz.X(), v)
			}
		case 70: // display flags (bitfield)
			if v, ok := tagInt(tag); ok {
				w.SetFlags(v)
			}
		case 71: // clipping boundary type
			if v, ok := tagInt(tag); ok {
				w.SetClippingBoundaryType(v)
			}
		case 90: // number of clip vertices (count; points follow as 14 pairs)
		case 14:
			if v, ok := tagFloat(tag); ok {
				pendingX = v
				haveX = true
			}
		case 24:
			if v, ok := tagFloat(tag); ok && haveX {
				w.SetPolygonWipeout(append(w.BoundaryPoints(), entity.ImageBoundaryPoint{X: pendingX, Y: v}))
				haveX = false
			}
		}
	}
	return w
}

func loadMLine(tags lldxf.Tags) *entity.MLine {
	m := entity.NewMLine()
	loadCommonAttribs(m, tags)

	var cur *entity.MLineVertex
	flush := func() {
		if cur != nil {
			m.Vertices = append(m.Vertices, *cur)
			cur = nil
		}
	}
	startVertex := func() *entity.MLineVertex {
		if cur == nil {
			cur = &entity.MLineVertex{
				Point:          []float64{0, 0, 0},
				MiterDirection: []float64{0, 0, 0},
				HasVertex:      true,
				HasMiter:       true,
			}
		}
		return cur
	}

	for _, tag := range tags {
		switch tag.Code() {
		case 3: // mline style name (AcDbMLine code 3)
			if v, ok := tag.Value().(string); ok {
				m.SetStyleName(v)
			}
		case 70: // mline flags (AcDbMLine code 70): has-fill and miter bits
			if v, ok := tagInt(tag); ok {
				m.SetFill(v&1 == 0)
				m.Flags = v
			}
		case 72: // justification
			if v, ok := tagInt(tag); ok {
				m.SetJustification(v)
			}
		case 40: // scale (AcDbMLine) — per-vertex start width also uses 40
			if v, ok := tagFloat(tag); ok {
				if cur != nil {
					cur.StartWidth = v
				} else {
					m.SetScale(v)
				}
			}
		case 71: // vertex count (AcDbMLine code 71)
			if _, ok := tagInt(tag); ok {
				flush()
			}
		case 73: // per-vertex mline element count (skip marker)
			_, _ = tagInt(tag)
		case 10:
			if v, ok := tagFloat(tag); ok {
				flush()
				startVertex().Point[0] = v
			}
		case 20:
			if v, ok := tagFloat(tag); ok {
				startVertex().Point[1] = v
			}
		case 30:
			if v, ok := tagFloat(tag); ok {
				startVertex().Point[2] = v
			}
		case 41: // per-vertex end width
			if v, ok := tagFloat(tag); ok && cur != nil {
				cur.EndWidth = v
			}
		case 42: // per-vertex bulge
			if v, ok := tagFloat(tag); ok && cur != nil {
				cur.Bulge = v
				cur.HasBulge = true
			}
		case 43: // miter direction x (starts vector; y/z follow as 44/45)
			if v, ok := tagFloat(tag); ok && cur != nil {
				cur.MiterDirection[0] = v
				cur.HasMiter = true
			}
		case 44:
			if v, ok := tagFloat(tag); ok && cur != nil {
				cur.MiterDirection[1] = v
			}
		case 45:
			if v, ok := tagFloat(tag); ok && cur != nil {
				cur.MiterDirection[2] = v
			}
		case 74: // fill flag (AcDbMLine code 74)
			if v, ok := tagInt(tag); ok {
				m.SetFill(v == 1)
			}
		case 75: // mline style handle (AcDbMLine code 75)
			if v, ok := tag.Value().(string); ok {
				m.StyleHandle = v
			}
		}
	}
	flush()
	return m
}

func loadAttdef(tags lldxf.Tags) *entity.Attdef {
	a := entity.NewAttdef()
	loadCommonAttribs(a, tags)

	var ip [3]float64
	for _, tag := range tags {
		switch tag.Code() {
		case 2:
			if v, ok := tag.Value().(string); ok {
				a.SetTag(v)
			}
		case 3:
			if v, ok := tag.Value().(string); ok {
				a.SetPrompt(v)
			}
		case 1:
			if v, ok := tag.Value().(string); ok {
				a.SetDefault(v)
			}
		case 10:
			if v, ok := tagFloat(tag); ok {
				ip[0] = v
			}
		case 20:
			if v, ok := tagFloat(tag); ok {
				ip[1] = v
			}
		case 30:
			if v, ok := tagFloat(tag); ok {
				ip[2] = v
			}
		case 40:
			if v, ok := tagFloat(tag); ok {
				a.SetHeight(v)
			}
		case 50:
			if v, ok := tagFloat(tag); ok {
				a.SetRotation(v)
			}
		case 7:
			if v, ok := tag.Value().(string); ok {
				a.SetStyle(v)
			}
		}
	}
	a.SetInsertPoint(ip[0], ip[1], ip[2])
	return a
}

func loadAttrib(tags lldxf.Tags) *entity.Attrib {
	a := entity.NewAttrib()
	loadCommonAttribs(a, tags)

	var ip [3]float64
	for _, tag := range tags {
		switch tag.Code() {
		case 2:
			if v, ok := tag.Value().(string); ok {
				a.SetTag(v)
			}
		case 1:
			if v, ok := tag.Value().(string); ok {
				a.SetText(v)
			}
		case 3:
			if v, ok := tag.Value().(string); ok {
				a.SetPrompt(v)
			}
		case 10:
			if v, ok := tagFloat(tag); ok {
				ip[0] = v
			}
		case 20:
			if v, ok := tagFloat(tag); ok {
				ip[1] = v
			}
		case 30:
			if v, ok := tagFloat(tag); ok {
				ip[2] = v
			}
		case 40:
			if v, ok := tagFloat(tag); ok {
				a.SetHeight(v)
			}
		case 41:
			if v, ok := tagFloat(tag); ok {
				a.SetWidthFactor(v)
			}
		case 50:
			if v, ok := tagFloat(tag); ok {
				a.SetRotation(v)
			}
		case 51:
			if v, ok := tagFloat(tag); ok {
				a.SetOblique(v)
			}
		case 7:
			if v, ok := tag.Value().(string); ok {
				a.SetStyle(v)
			}
		case 71:
			if v, ok := tagInt(tag); ok {
				a.SetMirror(v)
			}
		case 72, 74:
			if v, ok := tagInt(tag); ok {
				h, vv := a.Justification()
				if int(tag.Code()) == 72 {
					h = v
				} else {
					vv = v
				}
				a.SetJustification(h, vv)
			}
		case 210:
			if v, ok := tagFloat(tag); ok {
				a.SetExtrusionComponent(0, v)
			}
		case 220:
			if v, ok := tagFloat(tag); ok {
				a.SetExtrusionComponent(1, v)
			}
		case 230:
			if v, ok := tagFloat(tag); ok {
				a.SetExtrusionComponent(2, v)
			}
		}
	}
	a.SetInsertPoint(ip[0], ip[1], ip[2])
	return a
}

func loadTolerance(tags lldxf.Tags) *entity.Tolerance {
	t := entity.NewTolerance()
	loadCommonAttribs(t, tags)

	for _, tag := range tags {
		switch tag.Code() {
		case 10:
			if v, ok := tagFloat(tag); ok {
				t.SetInsertPoint(dxfmath.NewVec3(v, 0, 0))
			}
		case 20:
			if v, ok := tagFloat(tag); ok {
				pt := t.GetInsertPoint()
				t.SetInsertPoint(dxfmath.NewVec3(pt.X(), v, pt.Z()))
			}
		case 3:
			if v, ok := tag.Value().(string); ok {
				t.SetDimStyle(v)
			}
		case 1:
			if v, ok := tag.Value().(string); ok {
				t.SetContent(v)
			}
		}
	}
	return t
}

func loadMesh(tags lldxf.Tags) *entity.Mesh {
	m := entity.NewMesh()
	loadCommonAttribs(m, tags)
	var verts []dxfmath.Vec3
	var comps [3]float64
	flushVertex := func() {
		if len(verts) < capHint(len(verts)) {
			verts = append(verts, dxfmath.NewVec3(comps[0], comps[1], comps[2]))
		}
	}
	_ = flushVertex
	var faces []int  // flat face index groups (code 90)
	faceCount := 0   // code 93: number of faces
	edgeCount := 0   // code 94: number of edges
	inEdges := false // after 94, indices belong to edges

	flatToGroups := func(flat []int, n int) [][]int {
		out := make([][]int, 0, n)
		i := 0
		for g := 0; g < n && i < len(flat); g++ {
			size := flat[i]
			if size <= 0 || i+size >= len(flat)+1 {
				break
			}
			out = append(out, flat[i+1:i+1+size])
			i += size + 1
		}
		return out
	}
	var edges []int
	for _, tag := range tags {
		switch tag.Code() {
		case 71:
			if v, ok := tagInt(tag); ok {
				m.SetVersion(v)
			}
		case 72:
			if v, ok := tagInt(tag); ok {
				m.SetBlendCrease(v)
			}
		case 91:
			if v, ok := tagInt(tag); ok {
				m.SetSubdivisionLevels(v)
			}
		case 92:
			inEdges = false
		case 93:
			if v, ok := tagInt(tag); ok {
				faceCount = v
			} else if v, ok := tagFloat(tag); ok {
				faceCount = int(v)
			}
			inEdges = false
			faces = nil // face indices restart after the count marker
		case 94:
			if v, ok := tagInt(tag); ok {
				edgeCount = v
			} else if v, ok := tagFloat(tag); ok {
				edgeCount = int(v)
			}
			inEdges = true
		case 10, 20, 30:
			if v, ok := tagFloat(tag); ok {
				comps[int(tag.Code())/10-1] = v
				if int(tag.Code()) == 30 {
					verts = append(verts, dxfmath.NewVec3(comps[0], comps[1], comps[2]))
					comps = [3]float64{}
				}
			}
		case 90:
			if v, ok := tagInt(tag); ok {
				if inEdges {
					edges = append(edges, v)
				} else {
					faces = append(faces, v)
				}
			}
		}
	}
	m.SetVertices(verts)
	if faceCount > 0 {
		m.SetFaces(flatToGroups(faces, faceCount))
	}
	if edgeCount > 0 {
		m.SetEdges(flatToGroups(edges, edgeCount))
	}
	return m
}

// capHint guards the vertex flush closure above (unused helper).
func capHint(n int) int { return n }

func loadHelix(tags lldxf.Tags) *entity.Helix {
	h := entity.NewHelix()
	loadCommonAttribs(h, tags)
	var sp, av [3]float64
	setComp := func(dst *[3]float64, idx int, tag lldxf.Tag) {
		if v, ok := tagFloat(tag); ok {
			dst[idx] = v
		}
	}
	for _, tag := range tags {
		switch tag.Code() {
		case 10, 20, 30:
			setComp(&sp, int(tag.Code())/10-1, tag)
		case 11, 21, 31:
			setComp(&av, (int(tag.Code())-1)/10-1, tag)
		case 40:
			if v, ok := tagFloat(tag); ok {
				h.SetRadius(v)
			}
		case 90:
			if v, ok := tagFloat(tag); ok {
				h.SetTurns(v)
			} else if v, ok := tagInt(tag); ok {
				h.SetTurns(float64(v))
			}
		case 70:
			if v, ok := tagFloat(tag); ok {
				h.SetBaseRadius(v)
			}
		case 71:
			if v, ok := tagFloat(tag); ok {
				h.SetBaseHeight(v)
			}
		case 41:
			if v, ok := tagFloat(tag); ok {
				h.SetTurnHeight(v)
			}
		case 290:
			if v, ok := tagInt(tag); ok {
				h.SetHandedness(v)
			}
		case 280:
			if v, ok := tagInt(tag); ok {
				h.SetConstrainType(v)
			}
		}
	}
	h.SetStartPoint(dxfmath.NewVec3(sp[0], sp[1], sp[2]))
	h.SetAxisVector(dxfmath.NewVec3(av[0], av[1], av[2]))
	return h
}

func loadExtrudedSurface(tags lldxf.Tags) *entity.ExtrudedSurface {
	es := entity.NewExtrudedSurface()
	loadCommonAttribs(es, tags)
	var sv [3]float64
	for _, tag := range tags {
		switch tag.Code() {
		case 70:
			if v, ok := tagInt(tag); ok {
				es.Version = v
			}
		case 71:
			if v, ok := tagInt(tag); ok {
				es.Flags = v
			}
		case 10, 20, 30:
			if v, ok := tagFloat(tag); ok {
				sv[int(tag.Code())/10-1] = v
			}
		case 42:
			if v, ok := tagFloat(tag); ok {
				es.DraftAngle = v
			}
		case 45:
			if v, ok := tagFloat(tag); ok {
				es.TwistAngle = v
			}
		case 46:
			if v, ok := tagFloat(tag); ok {
				es.ScaleFactor = v
			}
		case 72:
			if v, ok := tagInt(tag); ok {
				es.Solid = v&1 != 0
				es.AlignStart = v&2 != 0
				es.Bank = v&4 != 0
				es.BasePointSet = v&8 != 0
				es.SweepAlignmentFlags = v >> 4
			}
		case 1:
			if v, ok := tag.Value().(string); ok {
				es.UID = v
			}
		}
	}
	es.SetSweepVector(dxfmath.NewVec3(sv[0], sv[1], sv[2]))
	return es
}

func loadLoftedSurface(tags lldxf.Tags) *entity.LoftedSurface {
	ls := entity.NewLoftedSurface()
	loadCommonAttribs(ls, tags)
	for _, tag := range tags {
		switch tag.Code() {
		case 70:
			if v, ok := tagInt(tag); ok {
				ls.Version = v
			}
		case 71:
			if v, ok := tagInt(tag); ok {
				ls.Flags = v
			}
		case 42:
			if v, ok := tagFloat(tag); ok {
				ls.DraftAngle = v
			}
		case 1:
			if v, ok := tag.Value().(string); ok {
				ls.SetUID(v)
			}
		}
	}
	return ls
}

func loadRevolvedSurface(tags lldxf.Tags) *entity.RevolvedSurface {
	rs := entity.NewRevolvedSurface()
	loadCommonAttribs(rs, tags)
	for _, tag := range tags {
		switch tag.Code() {
		case 70:
			if v, ok := tagInt(tag); ok {
				rs.Version = v
			}
		case 71:
			if v, ok := tagInt(tag); ok {
				rs.Flags = v
			}
		case 42:
			if v, ok := tagFloat(tag); ok {
				rs.SetRevolutionAngle(v)
			}
		case 1:
			if v, ok := tag.Value().(string); ok {
				rs.SetUID(v)
			}
		}
	}
	return rs
}

func loadSweptSurface(tags lldxf.Tags) *entity.SweptSurface {
	ss := entity.NewSweptSurface()
	loadCommonAttribs(ss, tags)
	for _, tag := range tags {
		switch tag.Code() {
		case 70:
			if v, ok := tagInt(tag); ok {
				ss.Version = v
			}
		case 71:
			if v, ok := tagInt(tag); ok {
				ss.Flags = v
			}
		case 42:
			if v, ok := tagFloat(tag); ok {
				ss.DraftAngle = v
			}
		case 1:
			if v, ok := tag.Value().(string); ok {
				ss.UID = v
			}
		}
	}
	return ss
}

func loadPlaneSurface(tags lldxf.Tags) *entity.PlaneSurface {
	ps := entity.NewPlaneSurface()
	loadCommonAttribs(ps, tags)
	for _, tag := range tags {
		switch tag.Code() {
		case 70:
			if v, ok := tagInt(tag); ok {
				ps.Version = v
			}
		case 71:
			if v, ok := tagInt(tag); ok {
				ps.Flags = v
			}
		case 1:
			if v, ok := tag.Value().(string); ok {
				ps.UID = v
			}
		}
	}
	return ps
}

// acisPayload extracts the text SAT payload: code 90 line count, then code 10
// lines. Payload-bearing ACIS records are retained whole (RawWholeRecord).
func acisPayload(tags lldxf.Tags) []string {
	var lines []string
	for _, tag := range tags {
		if int(tag.Code()) != 10 {
			continue
		}
		switch v := tag.Value().(type) {
		case string:
			lines = append(lines, v)
		default:
			_ = v
		}
	}
	return lines
}

func loadSolid3d(tags lldxf.Tags) *entity.Solid3d {
	s := entity.NewSolid3d()
	loadCommonAttribs(s, tags)
	for _, tag := range tags {
		switch tag.Code() {
		case 70:
			if v, ok := tagInt(tag); ok {
				s.SetVersion(v)
			}
		case 10, 20, 30:
			if v, ok := tagFloat(tag); ok {
				c := s.GetCenter()
				idx := int(tag.Code())/10 - 1
				comps := []float64{c.X(), c.Y(), c.Z()}
				comps[idx] = v
				s.SetCenter(dxfmath.NewVec3(comps[0], comps[1], comps[2]))
			}
		case 91:
			if v, ok := tagInt(tag); ok {
				s.Flags = v
			}
		case 40:
			if v, ok := tagFloat(tag); ok {
				s.SetVolume(v)
			}
		case 41:
			if v, ok := tagFloat(tag); ok {
				s.SetSurfaceArea(v)
			}
		}
	}
	if lines := acisStringPayload(tags); len(lines) > 0 {
		s.SetSATLines(lines)
	}
	return s
}

func loadRegion(tags lldxf.Tags) *entity.Region {
	r := entity.NewRegion()
	loadCommonAttribs(r, tags)
	if lines := acisStringPayload(tags); len(lines) > 0 {
		r.SetSATLines(lines)
	}
	return r
}

// acisRawPairs keeps the complete record (minus code 0) for whole-record
// retention of payload-bearing ACIS entities.
func acisRawPairs(tags lldxf.Tags) []entity.TagPair {
	raw := make([]entity.TagPair, 0, len(tags))
	for _, tag := range tags {
		code := int(tag.Code())
		if code == 0 {
			continue
		}
		raw = append(raw, entity.TagPair{Code: code, Value: tagStringValue(tag)})
	}
	return raw
}

func loadBody(tags lldxf.Tags) *entity.Body {
	b := entity.NewBody()
	loadCommonAttribs(b, tags)
	for _, tag := range tags {
		switch tag.Code() {
		case 70:
			if v, ok := tagInt(tag); ok {
				b.Version = v
			}
		case 71:
			if v, ok := tagInt(tag); ok {
				b.Flags = v
			}
		}
	}
	if lines := acisStringPayload(tags); len(lines) > 0 {
		b.SetSATLines(lines)
	}
	return b
}

// acisStringPayload extracts string-valued code-10 tags: the ACIS text payload
// of REGION and BODY records. Float code-10 tags are geometric points, not
// payload lines.
func acisStringPayload(tags lldxf.Tags) []string {
	var lines []string
	for _, tag := range tags {
		if int(tag.Code()) != 10 {
			continue
		}
		if v, ok := tag.Value().(string); ok {
			lines = append(lines, v)
		}
	}
	return lines
}

func loadCommonAttribs(e entity.Entity, tags lldxf.Tags) {
	for _, tag := range tags {
		switch tag.Code() {
		case 62:
			if v, ok := tagInt(tag); ok {
				e.SetColor(color.ColorNumber(v))
			}
		case 48:
			if v, ok := tagFloat(tag); ok {
				e.SetLtscale(v)
			}
		case 6:
			if v, ok := tag.Value().(string); ok {
				if setter, ok := e.(interface{ SetLinetype(string) }); ok {
					setter.SetLinetype(v)
				}
			}
		case 39:
			if v, ok := tagFloat(tag); ok {
				if setter, ok := e.(interface{ SetThickness(float64) }); ok {
					setter.SetThickness(v)
				}
			}
		case 60:
			if v, err := strconv.Atoi(fmtTagValue(tag.Value())); err == nil && v != 0 {
				if setter, ok := e.(interface{ SetInvisible(bool) }); ok {
					setter.SetInvisible(true)
				}
			}
		case 370:
			if v, ok := tagInt(tag); ok {
				if setter, ok := e.(interface{ SetLineWeight(int) }); ok {
					setter.SetLineWeight(v)
				}
			}
		case 420:
			if v, ok := tagInt(tag); ok {
				if setter, ok := e.(interface{ SetColorRGB(int) }); ok {
					setter.SetColorRGB(v)
				}
			}
		case 440:
			if v, ok := tagInt(tag); ok {
				if setter, ok := e.(interface{ SetTransparency(float64) }); ok {
					setter.SetTransparency(float64(v) / 100.0)
				}
			}
		}
	}
}

// setExtrusionComponent stores one extrusion component on the shared base.
func setExtrusionComponent(e interface {
	Extrusion() []float64
	SetExtrusion([]float64)
}, idx int, v float64) {
	if e == nil {
		return
	}
	cur := e.Extrusion()
	if cur == nil || len(cur) < 3 {
		cur = []float64{0, 0, 1}
	}
	cur[idx] = v
	e.SetExtrusion(cur)
}

// tagFloat reads a numeric tag value, coercing the string forms produced by
// the ASCII tagger (float64/int values are stored as strings for some code
// ranges) to float64. Loaders must not type-assert float64 directly: enhanced
// tags keep 210/220/230 and similar codes as raw text.
func tagFloat(tag lldxf.Tag) (float64, bool) {
	switch v := tag.Value().(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case bool:
		if v {
			return 1, true
		}
		return 0, true
	case string:
		var f float64
		if _, err := fmt.Sscanf(strings.TrimSpace(v), "%f", &f); err == nil {
			return f, true
		}
	}
	return 0, false
}

// tagInt reads an integer tag value with the same coercion.
func tagInt(tag lldxf.Tag) (int, bool) {
	if f, ok := tagFloat(tag); ok {
		return int(math.Round(f)), true
	}
	return 0, false
}

// fmtTagValue renders a tag value as text for integer parsing. Binary and
// enhanced readers store numeric codes as float64/int, so those forms must
// parse too; only non-numeric values yield "".
func fmtTagValue(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case int:
		return strconv.Itoa(t)
	case int8:
		return strconv.Itoa(int(t))
	case int16:
		return strconv.Itoa(int(t))
	case int32:
		return strconv.Itoa(int(t))
	case int64:
		return strconv.FormatInt(t, 10)
	case uint16:
		return strconv.FormatUint(uint64(t), 10)
	case uint32:
		return strconv.FormatUint(uint64(t), 10)
	case bool:
		if t {
			return "1"
		}
		return "0"
	}
	return ""
}

// FromFileWithXRefs loads a drawing and resolves its external references
// (recursively) using the optional search paths.
func FromFileWithXRefs(path string, searchPaths []string) (*drawing.Drawing, error) {
	d, err := FromFile(path)
	if err != nil {
		return nil, err
	}
	for _, sp := range searchPaths {
		d.AddXRefSearchPath(sp)
	}
	if err := d.LoadXRefsRecursive(); err != nil {
		return nil, err
	}
	return d, nil
}

// loadUnderlay parses an UNDERLAY entity record.
func loadUnderlay(tags lldxf.Tags) *entity.Underlay {
	u := entity.NewUnderlay()
	loadCommonAttribs(u, tags)
	var x, y, z float64
	var bx float64
	var hasBX bool
	var bnd [][]float64
	for _, tag := range tags {
		switch tag.Code() {
		case 340:
			if v, ok := tag.Value().(string); ok {
				u.UnderlayDefHandle = v
			}
		case 10:
			if v, ok := tagFloat(tag); ok {
				x = v
			}
		case 20:
			if v, ok := tagFloat(tag); ok {
				y = v
			}
		case 30:
			if v, ok := tagFloat(tag); ok {
				z = v
			}
		case 41:
			if v, ok := tagFloat(tag); ok {
				u.ScaleX = v
			}
		case 42:
			if v, ok := tagFloat(tag); ok {
				u.ScaleY = v
			}
		case 43:
			if v, ok := tagFloat(tag); ok {
				u.ScaleZ = v
			}
		case 50:
			if v, ok := tagFloat(tag); ok {
				u.Rotation = v
			}
		case 210, 220, 230:
			if v, ok := tagFloat(tag); ok {
				if u.Extrusion == nil {
					u.Extrusion = []float64{0, 0, 1}
				}
				u.Extrusion[tag.Code()/10-21] = v
			}
		case 280:
			if v, ok := tagInt(tag); ok {
				u.Flags = v
			}
		case 281:
			if v, ok := tagInt(tag); ok {
				u.Contrast = v
			}
		case 282:
			if v, ok := tagInt(tag); ok {
				u.Fade = v
			}
		case 11:
			if v, ok := tagFloat(tag); ok {
				bx = v
				hasBX = true
			}
		case 21:
			if v, ok := tagFloat(tag); ok && hasBX {
				bnd = append(bnd, []float64{bx, v})
				hasBX = false
			}
		}
	}
	u.SetInsertPoint(x, y, z)
	if len(bnd) > 0 {
		u.SetBoundaryPath(bnd)
	}
	return u
}
