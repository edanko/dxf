// Package dxf is a DXF(Drawing Exchange Format) library for golang.
// Supports AC1009 through AC1024, ASCII and binary DXF formats.
package dxf

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/lldxf"
	dxfmath "github.com/edanko/dxf/math"
	"github.com/edanko/dxf/table"
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
	return FromReader(f)
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
	d, err := NewDrawing()
	if err != nil {
		return nil, err
	}

	tagger := lldxf.NewASCIITagger(r)
	var allTags lldxf.Tags
	for tagger.Next() {
		tag := tagger.Tag()
		if tag != nil {
			allTags = append(allTags, tag)
		}
	}
	if err := tagger.Err(); err != nil {
		return nil, err
	}

	loader := lldxf.NewLoader(nil)
	if err := loader.Load(allTags); err != nil {
		return nil, fmt.Errorf("failed to load DXF: %w", err)
	}

	if err := buildEntitiesFromLoader(d, loader); err != nil {
		return nil, fmt.Errorf("failed to build drawing: %w", err)
	}

	return d, nil
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
	tagger := lldxf.NewBinaryTagger(r)

	d, err := NewDrawing()
	if err != nil {
		return nil, err
	}

	var allTags lldxf.Tags
	for tagger.Next() {
		tag := tagger.Tag()
		if tag != nil {
			allTags = append(allTags, tag)
		}
	}
	if err := tagger.Err(); err != nil {
		return nil, err
	}

	loader := lldxf.NewLoader(nil)
	if err := loader.Load(allTags); err != nil {
		return nil, fmt.Errorf("failed to load DXF: %w", err)
	}

	if err := buildEntitiesFromLoader(d, loader); err != nil {
		return nil, fmt.Errorf("failed to build drawing: %w", err)
	}

	return d, nil
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

func loadEntity(dxfType string, tags lldxf.Tags) (entity.Entity, error) {
	switch strings.ToUpper(dxfType) {
	case "POINT":
		return loadPoint(tags), nil
	case "LINE":
		return loadLine(tags), nil
	case "CIRCLE":
		return loadCircle(tags), nil
	case "ARC":
		return loadArc(tags), nil
	case "TEXT":
		return loadText(tags), nil
	case "MTEXT":
		return loadMText(tags), nil
	case "SOLID":
		return loadSolid(tags), nil
	case "TRACE":
		return loadTrace(tags), nil
	case "3DFACE":
		return load3DFace(tags), nil
	case "LWPOLYLINE":
		return loadLwPolyline(tags), nil
	case "POLYLINE":
		return loadPolyline(tags), nil
	case "VERTEX":
		return loadVertex(tags), nil
	case "INSERT":
		return loadInsert(tags), nil
	case "ELLIPSE":
		return loadEllipse(tags), nil
	case "SPLINE":
		return loadSpline(tags), nil
	case "LEADER":
		return loadLeader(tags), nil
	case "HATCH":
		return loadHatch(tags), nil
	case "IMAGE":
		return loadImage(tags), nil
	case "MLEADER":
		return loadMLeader(tags), nil
	case "DIMENSION":
		return loadDimension(tags), nil
	case "SHAPE":
		return loadShape(tags), nil
	case "XLINE":
		return loadXLine(tags), nil
	case "RAY":
		return loadRay(tags), nil
	case "WIPEOUT":
		return loadWipeout(tags), nil
	case "MLINE":
		return loadMLine(tags), nil
	case "ATTDEF":
		return loadAttdef(tags), nil
	case "ATTRIB":
		return loadAttrib(tags), nil
	case "TOLERANCE":
		return loadTolerance(tags), nil
	default:
		return loadGeneric(tags), nil
	}
}

func loadPoint(tags lldxf.Tags) *entity.Point {
	p := entity.NewPoint()
	loadCommonAttribs(p, tags)
	for _, tag := range tags {
		switch tag.Code() {
		case 10:
			if v, ok := tag.Value().(float64); ok {
				p.Coord[0] = v
			}
		case 20:
			if v, ok := tag.Value().(float64); ok {
				p.Coord[1] = v
			}
		case 30:
			if v, ok := tag.Value().(float64); ok {
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
			if v, ok := tag.Value().(float64); ok {
				l.Start[0] = v
			}
		case 20:
			if v, ok := tag.Value().(float64); ok {
				l.Start[1] = v
			}
		case 30:
			if v, ok := tag.Value().(float64); ok {
				l.Start[2] = v
			}
		case 11:
			if v, ok := tag.Value().(float64); ok {
				l.End[0] = v
			}
		case 21:
			if v, ok := tag.Value().(float64); ok {
				l.End[1] = v
			}
		case 31:
			if v, ok := tag.Value().(float64); ok {
				l.End[2] = v
			}
		case 39:
			if v, ok := tag.Value().(float64); ok {
				l.Thickness = v
			}
		case 210:
			if v, ok := tag.Value().(float64); ok {
				l.StretchingDirection[0] = v
			}
		case 220:
			if v, ok := tag.Value().(float64); ok {
				l.StretchingDirection[1] = v
			}
		case 230:
			if v, ok := tag.Value().(float64); ok {
				l.StretchingDirection[2] = v
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
			if v, ok := tag.Value().(float64); ok {
				c.Center[0] = v
			}
		case 20:
			if v, ok := tag.Value().(float64); ok {
				c.Center[1] = v
			}
		case 30:
			if v, ok := tag.Value().(float64); ok {
				c.Center[2] = v
			}
		case 40:
			if v, ok := tag.Value().(float64); ok {
				c.Radius = v
			}
		case 210:
			if v, ok := tag.Value().(float64); ok {
				c.Direction[0] = v
			}
		case 220:
			if v, ok := tag.Value().(float64); ok {
				c.Direction[1] = v
			}
		case 230:
			if v, ok := tag.Value().(float64); ok {
				c.Direction[2] = v
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
			if v, ok := tag.Value().(float64); ok {
				center[0] = v
			}
		case 20:
			if v, ok := tag.Value().(float64); ok {
				center[1] = v
			}
		case 30:
			if v, ok := tag.Value().(float64); ok {
				center[2] = v
			}
		case 40:
			if v, ok := tag.Value().(float64); ok {
				radius = v
			}
		case 50:
			if v, ok := tag.Value().(float64); ok {
				startAngle = v
			}
		case 51:
			if v, ok := tag.Value().(float64); ok {
				endAngle = v
			}
		}
	}
	c := entity.NewCircle()
	c.Center = center[:]
	c.Radius = radius
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
			if v, ok := tag.Value().(float64); ok {
				t.Coord1[0] = v
			}
		case 20:
			if v, ok := tag.Value().(float64); ok {
				t.Coord1[1] = v
			}
		case 30:
			if v, ok := tag.Value().(float64); ok {
				t.Coord1[2] = v
			}
		case 11:
			if v, ok := tag.Value().(float64); ok {
				t.Coord2[0] = v
			}
		case 21:
			if v, ok := tag.Value().(float64); ok {
				t.Coord2[1] = v
			}
		case 31:
			if v, ok := tag.Value().(float64); ok {
				t.Coord2[2] = v
			}
		case 40:
			if v, ok := tag.Value().(float64); ok {
				t.Height = v
			}
		case 50:
			if v, ok := tag.Value().(float64); ok {
				t.Rotation = v
			}
		case 41:
			if v, ok := tag.Value().(float64); ok {
				t.WidthFactor = v
			}
		case 51:
			if v, ok := tag.Value().(float64); ok {
				t.ObliqueAngle = v
			}
		case 1:
			if v, ok := tag.Value().(string); ok {
				t.Value = v
			}
		case 71:
			if v, err := strconv.Atoi(fmtTagValue(tag.Value())); err == nil {
				t.GenFlag = v
			}
		case 72:
			if v, err := strconv.Atoi(fmtTagValue(tag.Value())); err == nil {
				t.HorizontalFlag = v
			}
		case 73:
			if v, err := strconv.Atoi(fmtTagValue(tag.Value())); err == nil {
				t.VerticalFlag = v
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
			if v, ok := tag.Value().(float64); ok {
				t.Coord1[0] = v
			}
		case 20:
			if v, ok := tag.Value().(float64); ok {
				t.Coord1[1] = v
			}
		case 30:
			if v, ok := tag.Value().(float64); ok {
				t.Coord1[2] = v
			}
		case 40:
			if v, ok := tag.Value().(float64); ok {
				t.Height = v
			}
		case 50:
			if v, ok := tag.Value().(float64); ok {
				t.Rotation = v
			}
		case 1:
			if v, ok := tag.Value().(string); ok {
				t.Value = v
			}
		}
	}
	return t
}

func loadSolid(tags lldxf.Tags) *entity.Solid {
	s := entity.NewSolid()
	loadCommonAttribs(s, tags)
	for _, tag := range tags {
		switch tag.Code() {
		case 10:
			if v, ok := tag.Value().(float64); ok {
				s.FirstPoint[0] = v
			}
		case 20:
			if v, ok := tag.Value().(float64); ok {
				s.FirstPoint[1] = v
			}
		case 30:
			if v, ok := tag.Value().(float64); ok {
				s.FirstPoint[2] = v
			}
		case 11:
			if v, ok := tag.Value().(float64); ok {
				s.SecondPoint[0] = v
			}
		case 21:
			if v, ok := tag.Value().(float64); ok {
				s.SecondPoint[1] = v
			}
		case 31:
			if v, ok := tag.Value().(float64); ok {
				s.SecondPoint[2] = v
			}
		case 12:
			if v, ok := tag.Value().(float64); ok {
				s.ThirdPoint[0] = v
			}
		case 22:
			if v, ok := tag.Value().(float64); ok {
				s.ThirdPoint[1] = v
			}
		case 32:
			if v, ok := tag.Value().(float64); ok {
				s.ThirdPoint[2] = v
			}
		case 13:
			if v, ok := tag.Value().(float64); ok {
				s.FourthPoint[0] = v
			}
		case 23:
			if v, ok := tag.Value().(float64); ok {
				s.FourthPoint[1] = v
			}
		case 33:
			if v, ok := tag.Value().(float64); ok {
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
			if v, ok := tag.Value().(float64); ok {
				f.Points[0][0] = v
			}
		case 20:
			if v, ok := tag.Value().(float64); ok {
				f.Points[0][1] = v
			}
		case 30:
			if v, ok := tag.Value().(float64); ok {
				f.Points[0][2] = v
			}
		case 11:
			if v, ok := tag.Value().(float64); ok {
				f.Points[1][0] = v
			}
		case 21:
			if v, ok := tag.Value().(float64); ok {
				f.Points[1][1] = v
			}
		case 31:
			if v, ok := tag.Value().(float64); ok {
				f.Points[1][2] = v
			}
		case 12:
			if v, ok := tag.Value().(float64); ok {
				f.Points[2][0] = v
			}
		case 22:
			if v, ok := tag.Value().(float64); ok {
				f.Points[2][1] = v
			}
		case 32:
			if v, ok := tag.Value().(float64); ok {
				f.Points[2][2] = v
			}
		case 13:
			if v, ok := tag.Value().(float64); ok {
				f.Points[3][0] = v
			}
		case 23:
			if v, ok := tag.Value().(float64); ok {
				f.Points[3][1] = v
			}
		case 33:
			if v, ok := tag.Value().(float64); ok {
				f.Points[3][2] = v
			}
		}
	}
	return f
}

func loadLwPolyline(tags lldxf.Tags) *entity.LwPolyline {
	var numVertices int
	var closed bool
	vertices := make([][2]float64, 0)
	bulges := make([]float64, 0)
	for _, tag := range tags {
		switch tag.Code() {
		case 90:
			if v, err := strconv.Atoi(fmtTagValue(tag.Value())); err == nil {
				numVertices = v
			}
		case 70:
			if v, err := strconv.Atoi(fmtTagValue(tag.Value())); err == nil && v == 1 {
				closed = true
			}
		case 10:
			if v, ok := tag.Value().(float64); ok {
				vertices = append(vertices, [2]float64{v, 0})
			}
		case 20:
			if v, ok := tag.Value().(float64); ok {
				if len(vertices) > 0 {
					vertices[len(vertices)-1][1] = v
				}
			}
		case 42:
			if v, ok := tag.Value().(float64); ok {
				bulges = append(bulges, v)
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
			if v, ok := tag.Value().(float64); ok {
				e.Center[0] = v
			}
		case 20:
			if v, ok := tag.Value().(float64); ok {
				e.Center[1] = v
			}
		case 30:
			if v, ok := tag.Value().(float64); ok {
				e.Center[2] = v
			}
		case 40:
			if v, ok := tag.Value().(float64); ok {
				e.Major = v
			}
		case 41:
			if v, ok := tag.Value().(float64); ok {
				e.Minor = v
			}
		case 200:
			if v, ok := tag.Value().(float64); ok {
				e.Ratio = v
			}
		case 50:
			if v, ok := tag.Value().(float64); ok {
				e.Start = v
			}
		case 51:
			if v, ok := tag.Value().(float64); ok {
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
			if v, err := strconv.Atoi(fmtTagValue(tag.Value())); err == nil {
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
			if v, ok := tag.Value().(float64); ok {
				x = v
			}
		case 20:
			if v, ok := tag.Value().(float64); ok {
				y = v
			}
		case 30:
			if v, ok := tag.Value().(float64); ok {
				z = v
			}
		}
	}
	v := entity.NewVertex(x, y, z)
	loadCommonAttribs(v, tags)
	return v
}

func loadInsert(tags lldxf.Tags) *entity.Insert {
	i := entity.NewInsert()
	loadCommonAttribs(i, tags)
	var x, y, z float64
	for _, tag := range tags {
		switch tag.Code() {
		case 10:
			if v, ok := tag.Value().(float64); ok {
				x = v
			}
		case 20:
			if v, ok := tag.Value().(float64); ok {
				y = v
			}
		case 30:
			if v, ok := tag.Value().(float64); ok {
				z = v
			}
		case 2:
			if v, ok := tag.Value().(string); ok {
				i.SetBlockName(v)
			}
		case 50:
			if v, ok := tag.Value().(float64); ok {
				i.SetRotation(v)
			}
		}
	}
	i.SetInsertPoint(dxfmath.NewVec3(x, y, z))
	return i
}

func loadTrace(tags lldxf.Tags) *entity.Trace {
	points := make([]float64, 12)
	for _, tag := range tags {
		switch tag.Code() {
		case 10:
			if v, ok := tag.Value().(float64); ok {
				points[0] = v
			}
		case 20:
			if v, ok := tag.Value().(float64); ok {
				points[1] = v
			}
		case 11:
			if v, ok := tag.Value().(float64); ok {
				points[2] = v
			}
		case 21:
			if v, ok := tag.Value().(float64); ok {
				points[3] = v
			}
		case 12:
			if v, ok := tag.Value().(float64); ok {
				points[4] = v
			}
		case 22:
			if v, ok := tag.Value().(float64); ok {
				points[5] = v
			}
		case 13:
			if v, ok := tag.Value().(float64); ok {
				points[6] = v
			}
		case 23:
			if v, ok := tag.Value().(float64); ok {
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
			if v, err := strconv.Atoi(fmtTagValue(tag.Value())); err == nil {
				s.Flag = v
			}
		case 71:
			if v, err := strconv.Atoi(fmtTagValue(tag.Value())); err == nil {
				s.Degree = v
			}
		case 40:
			if v, ok := tag.Value().(float64); ok {
				s.Knots = append(s.Knots, v)
			}
		case 42:
			if v, ok := tag.Value().(float64); ok {
				s.Weights = append(s.Weights, v)
			}
		case 10:
			if v, ok := tag.Value().(float64); ok {
				s.Controls = append(s.Controls, []float64{v, 0, 0})
			}
		case 20:
			if len(s.Controls) > 0 {
				if v, ok := tag.Value().(float64); ok {
					s.Controls[len(s.Controls)-1][1] = v
				}
			}
		case 30:
			if len(s.Controls) > 0 {
				if v, ok := tag.Value().(float64); ok {
					s.Controls[len(s.Controls)-1][2] = v
				}
			}
		}
	}
	return s
}

func loadLeader(tags lldxf.Tags) *entity.Leader {
	l := entity.NewLeader()
	loadCommonAttribs(l, tags)

	var currentPoint []float64
	for _, tag := range tags {
		switch tag.Code() {
		case 10:
			currentPoint = []float64{0, 0, 0}
			if v, ok := tag.Value().(float64); ok {
				currentPoint[0] = v
			}
		case 20:
			if v, ok := tag.Value().(float64); ok {
				if currentPoint == nil {
					currentPoint = []float64{0, 0, 0}
				}
				currentPoint[1] = v
			}
		case 30:
			if v, ok := tag.Value().(float64); ok {
				if currentPoint == nil {
					currentPoint = []float64{0, 0, 0}
				}
				currentPoint[2] = v
			}
			if currentPoint != nil {
				l.AddPoint(currentPoint[0], currentPoint[1], currentPoint[2])
			}
		case 71:
			if v, err := strconv.Atoi(fmtTagValue(tag.Value())); err == nil {
				l.SetAnnotationType(v)
			}
		case 72:
			if v, err := strconv.Atoi(fmtTagValue(tag.Value())); err == nil {
				l.SetPathType(v)
			}
		case 3:
			if v, ok := tag.Value().(string); ok {
				l.SetDimstyle(v)
			}
		case 40:
			if v, ok := tag.Value().(float64); ok {
				l.SetArrowhead(0, v, 0)
			}
		case 73:
			if v, err := strconv.Atoi(fmtTagValue(tag.Value())); err == nil {
				if l.GetAnnotationType() == 1 {
					l.SetTextProperties(0, 0, "", 0, v)
				}
			}
		case 74:
			if v, err := strconv.Atoi(fmtTagValue(tag.Value())); err == nil {
				if l.GetAnnotationType() == 1 {
					l.SetTextProperties(0, 0, "", 0, 0)
					l.SetTextProperties(0, 0, "", 0, v)
				} else {
					l.SetHookLine(true, v)
				}
			}
		case 75:
			if v, err := strconv.Atoi(fmtTagValue(tag.Value())); err == nil {
				l.SetHookLine(v == 1, 0)
			}
		case 210:
			if v, ok := tag.Value().(float64); ok {
				l.SetExtrusion(v, 0, 0)
			}
		}
	}
	return l
}

func loadHatch(tags lldxf.Tags) *entity.Hatch {
	h := entity.NewHatch()
	loadCommonAttribs(h, tags)

	for _, tag := range tags {
		switch tag.Code() {
		case 2:
			if v, ok := tag.Value().(string); ok {
				h.SetPattern(1, v, nil)
			}
		case 70:
			if v, err := strconv.Atoi(fmtTagValue(tag.Value())); err == nil {
				h.SetStyle(v)
			}
		case 76:
			if v, err := strconv.Atoi(fmtTagValue(tag.Value())); err == nil {
				h.SetPattern(v, "", nil)
			}
		}
	}
	return h
}

func loadImage(tags lldxf.Tags) *entity.Image {
	img := entity.NewImage()
	loadCommonAttribs(img, tags)

	var uVec, vVec dxfmath.Vec3
	for _, tag := range tags {
		switch tag.Code() {
		case 10:
			if v, ok := tag.Value().(float64); ok {
				img.SetInsertPoint(dxfmath.NewVec3(v, 0, 0))
			}
		case 20:
			if v, ok := tag.Value().(float64); ok {
				pt := img.InsertPoint()
				img.SetInsertPoint(dxfmath.NewVec3(pt.X(), v, pt.Z()))
			}
		case 30:
			if v, ok := tag.Value().(float64); ok {
				pt := img.InsertPoint()
				img.SetInsertPoint(dxfmath.NewVec3(pt.X(), pt.Y(), v))
			}
		case 11:
			if v, ok := tag.Value().(float64); ok {
				uVec = dxfmath.NewVec3(v, 0, 0)
			}
		case 21:
			if v, ok := tag.Value().(float64); ok {
				uVec = dxfmath.NewVec3(uVec.X(), v, 0)
			}
		case 31:
			if v, ok := tag.Value().(float64); ok {
				uVec = dxfmath.NewVec3(uVec.X(), uVec.Y(), v)
			}
		case 12:
			if v, ok := tag.Value().(float64); ok {
				vVec = dxfmath.NewVec3(v, 0, 0)
			}
		case 22:
			if v, ok := tag.Value().(float64); ok {
				vVec = dxfmath.NewVec3(vVec.X(), v, 0)
			}
		case 32:
			if v, ok := tag.Value().(float64); ok {
				vVec = dxfmath.NewVec3(vVec.X(), vVec.Y(), v)
			}
		case 13:
			if v, ok := tag.Value().(float64); ok {
				img.SetImageSize(v, 0)
			}
		case 23:
			if v, ok := tag.Value().(float64); ok {
				size := img.ImageSize()
				img.SetImageSize(size.X(), v)
			}
		case 70:
			if v, err := strconv.Atoi(fmtTagValue(tag.Value())); err == nil {
				_ = v
				img.SetDisplayProperties(0, 0, 0)
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
	return img
}

func loadMLeader(tags lldxf.Tags) *entity.MLeader {
	m := entity.NewMLeader()
	loadCommonAttribs(m, tags)
	return m
}

func loadDimension(tags lldxf.Tags) *entity.Dimension {
	d := entity.NewDimension()
	loadCommonAttribs(d, tags)

	for _, tag := range tags {
		switch tag.Code() {
		case 10:
			if v, ok := tag.Value().(float64); ok {
				d.SetDefinitionPoint([]float64{v, 0, 0})
			}
		case 20:
			if v, ok := tag.Value().(float64); ok {
				d.SetDefinitionPoint([]float64{0, v, 0})
			}
		case 30:
			if v, ok := tag.Value().(float64); ok {
				d.SetDefinitionPoint([]float64{0, 0, v})
			}
		case 50:
			if v, ok := tag.Value().(float64); ok {
				d.SetRotation(v)
			}
		case 70:
			if v, err := strconv.Atoi(fmtTagValue(tag.Value())); err == nil {
				d.SetDimensionType(v)
			}
		case 1:
			if v, ok := tag.Value().(string); ok {
				d.SetText(v)
			}
		}
	}
	return d
}

func loadShape(tags lldxf.Tags) *entity.Shape {
	s := entity.NewShape()
	loadCommonAttribs(s, tags)

	for _, tag := range tags {
		switch tag.Code() {
		case 10:
			if v, ok := tag.Value().(float64); ok {
				s.SetInsertPoint(dxfmath.NewVec3(v, 0, 0))
			}
		case 20:
			if v, ok := tag.Value().(float64); ok {
				pt := s.GetInsertPoint()
				s.SetInsertPoint(dxfmath.NewVec3(pt.X(), v, pt.Z()))
			}
		case 30:
			if v, ok := tag.Value().(float64); ok {
				pt := s.GetInsertPoint()
				s.SetInsertPoint(dxfmath.NewVec3(pt.X(), pt.Y(), v))
			}
		case 2:
			if v, ok := tag.Value().(string); ok {
				s.SetShapeName(v)
			}
		case 40:
			if v, ok := tag.Value().(float64); ok {
				s.SetSize(v)
			}
		case 50:
			if v, ok := tag.Value().(float64); ok {
				s.SetRotation(v)
			}
		case 44:
			if v, ok := tag.Value().(float64); ok {
				s.SetXScale(v)
			}
		case 51:
			if v, ok := tag.Value().(float64); ok {
				s.SetOblique(v)
			}
		}
	}
	return s
}

func loadXLine(tags lldxf.Tags) *entity.XLine {
	x := entity.NewXLine()
	loadCommonAttribs(x, tags)

	for _, tag := range tags {
		switch tag.Code() {
		case 10:
			if v, ok := tag.Value().(float64); ok {
				x.SetStart(dxfmath.NewVec3(v, 0, 0))
			}
		case 20:
			if v, ok := tag.Value().(float64); ok {
				pt := x.Start()
				x.SetStart(dxfmath.NewVec3(pt.X(), v, pt.Z()))
			}
		case 30:
			if v, ok := tag.Value().(float64); ok {
				pt := x.Start()
				x.SetStart(dxfmath.NewVec3(pt.X(), pt.Y(), v))
			}
		case 11:
			if v, ok := tag.Value().(float64); ok {
				x.SetUnitVector(dxfmath.NewVec3(v, 0, 0))
			}
		case 21:
			if v, ok := tag.Value().(float64); ok {
				uv := x.UnitVector()
				x.SetUnitVector(dxfmath.NewVec3(uv.X(), v, uv.Z()))
			}
		case 31:
			if v, ok := tag.Value().(float64); ok {
				uv := x.UnitVector()
				x.SetUnitVector(dxfmath.NewVec3(uv.X(), uv.Y(), v))
			}
		}
	}
	return x
}

func loadRay(tags lldxf.Tags) *entity.Ray {
	r := entity.NewRay()
	loadCommonAttribs(r, tags)

	for _, tag := range tags {
		switch tag.Code() {
		case 10:
			if v, ok := tag.Value().(float64); ok {
				r.SetStart(dxfmath.NewVec3(v, 0, 0))
			}
		case 20:
			if v, ok := tag.Value().(float64); ok {
				pt := r.Start()
				r.SetStart(dxfmath.NewVec3(pt.X(), v, pt.Z()))
			}
		case 30:
			if v, ok := tag.Value().(float64); ok {
				pt := r.Start()
				r.SetStart(dxfmath.NewVec3(pt.X(), pt.Y(), v))
			}
		case 11:
			if v, ok := tag.Value().(float64); ok {
				r.SetUnitVector(dxfmath.NewVec3(v, 0, 0))
			}
		case 21:
			if v, ok := tag.Value().(float64); ok {
				uv := r.UnitVector()
				r.SetUnitVector(dxfmath.NewVec3(uv.X(), v, uv.Z()))
			}
		case 31:
			if v, ok := tag.Value().(float64); ok {
				uv := r.UnitVector()
				r.SetUnitVector(dxfmath.NewVec3(uv.X(), uv.Y(), v))
			}
		}
	}
	return r
}

func loadWipeout(tags lldxf.Tags) *entity.Wipeout {
	w := entity.NewWipeout()
	loadCommonAttribs(w, tags)

	for _, tag := range tags {
		switch tag.Code() {
		case 10:
			if v, ok := tag.Value().(float64); ok {
				w.SetInsertPoint(dxfmath.NewVec3(v, 0, 0))
			}
		case 20:
			if v, ok := tag.Value().(float64); ok {
				pt := w.InsertPoint()
				w.SetInsertPoint(dxfmath.NewVec3(pt.X(), v, pt.Z()))
			}
		case 30:
			if v, ok := tag.Value().(float64); ok {
				pt := w.InsertPoint()
				w.SetInsertPoint(dxfmath.NewVec3(pt.X(), pt.Y(), v))
			}
		case 11:
			if v, ok := tag.Value().(float64); ok {
				uv := w.UVector()
				vv := w.VVector()
				w.SetUVectors(dxfmath.NewVec3(v, uv.Y(), uv.Z()), vv)
			}
		case 12:
			if v, ok := tag.Value().(float64); ok {
				uv := w.UVector()
				vv := w.VVector()
				w.SetUVectors(uv, dxfmath.NewVec3(v, vv.Y(), vv.Z()))
			}
		case 70:
			if v, err := strconv.Atoi(fmtTagValue(tag.Value())); err == nil {
				w.SetRectangularWipeout(0, 0, float64(v), float64(v))
			}
		}
	}
	return w
}

func loadMLine(tags lldxf.Tags) *entity.MLine {
	m := entity.NewMLine()
	loadCommonAttribs(m, tags)

	for _, tag := range tags {
		switch tag.Code() {
		case 40:
			if v, ok := tag.Value().(float64); ok {
				m.SetScale(v)
			}
		case 70:
			if v, err := strconv.Atoi(fmtTagValue(tag.Value())); err == nil {
				m.SetJustification(v)
			}
		case 71:
			if v, err := strconv.Atoi(fmtTagValue(tag.Value())); err == nil {
				m.SetJustification(v)
			}
		case 73:
			if v, err := strconv.Atoi(fmtTagValue(tag.Value())); err == nil {
				m.SetClosed(v == 1)
			}
		case 74:
			if v, err := strconv.Atoi(fmtTagValue(tag.Value())); err == nil {
				m.SetFill(v == 1)
			}
		case 75:
			if _, err := strconv.Atoi(fmtTagValue(tag.Value())); err == nil {
				m.SetStyleName("")
			}
		}
	}
	return m
}

func loadAttdef(tags lldxf.Tags) *entity.Attdef {
	a := entity.NewAttdef()
	loadCommonAttribs(a, tags)

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
			if v, ok := tag.Value().(float64); ok {
				a.SetInsertPoint(v, 0, 0)
			}
		case 20:
			if v, ok := tag.Value().(float64); ok {
				a.SetInsertPoint(0, v, 0)
			}
		case 30:
			if v, ok := tag.Value().(float64); ok {
				a.SetInsertPoint(0, 0, v)
			}
		case 40:
			if v, ok := tag.Value().(float64); ok {
				a.SetHeight(v)
			}
		case 50:
			if v, ok := tag.Value().(float64); ok {
				a.SetRotation(v)
			}
		case 7:
			if v, ok := tag.Value().(string); ok {
				a.SetStyle(v)
			}
		}
	}
	return a
}

func loadAttrib(tags lldxf.Tags) *entity.Attrib {
	a := entity.NewAttrib()
	loadCommonAttribs(a, tags)

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
		case 10:
			if v, ok := tag.Value().(float64); ok {
				a.SetInsertPoint(v, 0, 0)
			}
		case 20:
			if v, ok := tag.Value().(float64); ok {
				a.SetInsertPoint(0, v, 0)
			}
		case 30:
			if v, ok := tag.Value().(float64); ok {
				a.SetInsertPoint(0, 0, v)
			}
		case 40:
			if v, ok := tag.Value().(float64); ok {
				a.SetHeight(v)
			}
		case 50:
			if v, ok := tag.Value().(float64); ok {
				a.SetRotation(v)
			}
		case 7:
			if v, ok := tag.Value().(string); ok {
				a.SetStyle(v)
			}
		}
	}
	return a
}

func loadTolerance(tags lldxf.Tags) *entity.Tolerance {
	t := entity.NewTolerance()
	loadCommonAttribs(t, tags)

	for _, tag := range tags {
		switch tag.Code() {
		case 10:
			if v, ok := tag.Value().(float64); ok {
				t.SetInsertPoint(dxfmath.NewVec3(v, 0, 0))
			}
		case 20:
			if v, ok := tag.Value().(float64); ok {
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

type GenericEntity struct {
	typeName string
	tags     []TagPair
	color    color.ColorNumber
	ltscale  float64
}

type TagPair struct {
	Code  int
	Value string
}

func (g *GenericEntity) IsEntity() bool { return true }
func (g *GenericEntity) Format(f format.Formatter) {
	f.WriteString(0, g.typeName)
	for _, tag := range g.tags {
		f.WriteString(tag.Code, tag.Value)
	}
}
func (g *GenericEntity) BBox() ([]float64, []float64) {
	return []float64{0, 0, 0}, []float64{0, 0, 0}
}
func (g *GenericEntity) SetBlockRecord(h handle.Handler) {}
func (g *GenericEntity) Layer() *table.Layer             { return nil }
func (g *GenericEntity) SetLayer(l *table.Layer)         {}
func (g *GenericEntity) SetLtscale(v float64)            { g.ltscale = v }
func (g *GenericEntity) SetColor(c color.ColorNumber)    { g.color = c }
func (g *GenericEntity) Handle() string {
	for _, tag := range g.tags {
		if tag.Code == 5 {
			return tag.Value
		}
	}
	return ""
}
func (g *GenericEntity) SetHandle(hg *handle.HandleGenerator) {}
func (g *GenericEntity) DXFType() string                      { return g.typeName }
func (g *GenericEntity) GetAttributes() map[string]interface{} {
	return map[string]interface{}{"type": g.typeName}
}
func (g *GenericEntity) LoadAttributes(attribs map[string]interface{}) error { return nil }

func loadGeneric(tags lldxf.Tags) entity.Entity {
	dxfType, _ := tags.GetFirstValue(0)
	if dxfType == nil {
		return nil
	}
	typeStr := fmt.Sprintf("%v", dxfType)

	g := &GenericEntity{
		typeName: typeStr,
		tags:     make([]TagPair, 0),
		color:    0,
		ltscale:  1.0,
	}
	for _, tag := range tags {
		code := tag.Code()
		value := fmt.Sprintf("%v", tag.Value())
		if code != 0 {
			g.tags = append(g.tags, TagPair{Code: int(code), Value: value})
		}
	}
	return g
}

func loadCommonAttribs(e entity.Entity, tags lldxf.Tags) {
	for _, tag := range tags {
		switch tag.Code() {
		case 62:
			if v, err := strconv.Atoi(fmtTagValue(tag.Value())); err == nil {
				e.SetColor(color.ColorNumber(v))
			}
		case 48:
			if v, ok := tag.Value().(float64); ok {
				e.SetLtscale(v)
			}
		}
	}
}

func fmtTagValue(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
