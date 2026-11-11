package entity

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	dxfmath "github.com/edanko/dxf/math"
)

// DecodeACISMesh decodes the flat polygonal faces of a Region, Body or Solid3d
// ACIS payload into a MESH entity. Only text SAT payloads with straight edge
// topology are supported; curved faces and binary SAB payloads return an error
// because they require a full ACIS kernel.
func DecodeACISMesh(e Entity) (*Mesh, error) {
	if e == nil {
		return nil, fmt.Errorf("acis mesh: nil entity")
	}
	lines, formatName, err := acisLinesOf(e)
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return nil, fmt.Errorf("acis mesh: %s carries no ACIS payload", e.DXFType())
	}
	if formatName == "SAB" {
		return nil, fmt.Errorf("acis mesh: SAB geometry decoding is not supported")
	}
	return DecodeSATMesh(lines)
}

func acisLinesOf(e Entity) ([]string, string, error) {
	switch value := e.(type) {
	case *Region:
		return value.SATLines(), value.ACISInfo().Format.String(), nil
	case *Body:
		return value.SATLines(), value.ACISInfo().Format.String(), nil
	case *Solid3d:
		return value.SATLines(), value.ACISInfo().Format.String(), nil
	default:
		return nil, "", fmt.Errorf("acis mesh: %s is not an ACIS body", e.DXFType())
	}
}

// DecodeSATMesh decodes flat polygonal SAT topology. Straight edges are
// required; unsupported curve or surface records are skipped.
func DecodeSATMesh(lines []string) (*Mesh, error) {
	records, err := parseSATRecords(lines)
	if err != nil {
		return nil, err
	}
	transform := satBodyTransform(records)
	vertices := make([]dxfmath.Vec3, 0)
	faces := make([][]int, 0)
	index := map[string]int{}

	for faceIndex, record := range records {
		if record.name != "face" {
			continue
		}
		surfaceIndex, ok := satPointer(record, 5)
		if !ok || records[surfaceIndex].name != "plane-surface" {
			continue
		}
		polygon, err := satFacePolygon(records, faceIndex)
		if err != nil {
			return nil, err
		}
		if len(polygon) < 3 {
			continue
		}
		indices := make([]int, 0, len(polygon))
		for _, point := range polygon {
			if transform != nil {
				point = transformPoint44(*transform, point)
			}
			key := satPointKey(point)
			vertexIndex, exists := index[key]
			if !exists {
				vertexIndex = len(vertices)
				index[key] = vertexIndex
				vertices = append(vertices, point)
			}
			indices = append(indices, vertexIndex)
		}
		// Drop repeated closing vertices and degenerate faces.
		if len(indices) > 1 && indices[0] == indices[len(indices)-1] {
			indices = indices[:len(indices)-1]
		}
		if len(indices) < 3 || satDegenerate(indices) {
			continue
		}
		for i := 1; i+1 < len(indices); i++ {
			faces = append(faces, []int{indices[0], indices[i], indices[i+1]})
		}
	}
	if len(vertices) == 0 || len(faces) == 0 {
		return nil, fmt.Errorf("acis mesh: SAT payload has no flat polygonal faces")
	}
	mesh := NewMesh()
	mesh.SetVertices(vertices)
	mesh.SetFaces(faces)
	return mesh, nil
}

type satRecord struct {
	name string
	data []string
}

func parseSATRecords(lines []string) ([]satRecord, error) {
	if len(lines) == 0 {
		return nil, fmt.Errorf("acis mesh: empty SAT payload")
	}
	plain := lines
	if !looksLikeSATHeader(firstNonEmptyLine(lines)) {
		plain = decodeSATLines(lines)
	}
	records := make([]satRecord, 0, len(plain))
	headerLines := 0
	for _, line := range plain {
		text := strings.TrimSpace(line)
		if text == "" {
			continue
		}
		if strings.HasPrefix(text, "End-of-ACIS-data") || strings.HasPrefix(text, "End-of-ASM-data") {
			break
		}
		if strings.HasPrefix(text, "Begin-of-ACIS-History-data") {
			break
		}
		if headerLines < 3 {
			headerLines++
			continue
		}
		fields := strings.Fields(text)
		if len(fields) > 0 && fields[len(fields)-1] == "#" {
			fields = fields[:len(fields)-1]
		}
		if len(fields) < 3 {
			continue
		}
		records = append(records, satRecord{name: fields[0], data: append([]string(nil), fields[3:]...)})
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("acis mesh: SAT payload has no records")
	}
	return records, nil
}

func satPointer(record satRecord, position int) (int, bool) {
	if position < 0 || position >= len(record.data) {
		return 0, false
	}
	text := record.data[position]
	if !strings.HasPrefix(text, "$") {
		return 0, false
	}
	value, err := strconv.Atoi(text[1:])
	if err != nil || value < 0 {
		return 0, false
	}
	return value, true
}

func satFacePolygon(records []satRecord, faceIndex int) ([]dxfmath.Vec3, error) {
	loopIndex, ok := satPointer(records[faceIndex], 2)
	if !ok || loopIndex >= len(records) || records[loopIndex].name != "loop" {
		return nil, nil
	}
	coedgeIndex, ok := satPointer(records[loopIndex], 2)
	if !ok {
		return nil, nil
	}
	polygon := make([]dxfmath.Vec3, 0, 4)
	seen := make(map[int]bool)
	for coedgeIndex >= 0 && coedgeIndex < len(records) && !seen[coedgeIndex] {
		seen[coedgeIndex] = true
		coedge := records[coedgeIndex]
		if coedge.name != "coedge" {
			return nil, fmt.Errorf("acis mesh: loop references %s instead of a coedge", coedge.name)
		}
		edgeIndex, ok := satPointer(coedge, 4)
		if !ok || edgeIndex >= len(records) || records[edgeIndex].name != "edge" {
			return nil, fmt.Errorf("acis mesh: coedge has no edge")
		}
		reversed := len(coedge.data) > 5 && coedge.data[5] == "reversed"
		vertexPosition := 1
		if reversed {
			vertexPosition = 3
		}
		vertexIndex, ok := satPointer(records[edgeIndex], vertexPosition)
		if !ok || vertexIndex >= len(records) || records[vertexIndex].name != "vertex" {
			return nil, fmt.Errorf("acis mesh: edge has no vertex")
		}
		pointIndex, ok := satPointer(records[vertexIndex], 2)
		if !ok || pointIndex >= len(records) || records[pointIndex].name != "point" {
			return nil, fmt.Errorf("acis mesh: vertex has no point")
		}
		point, err := satPoint(records[pointIndex])
		if err != nil {
			return nil, err
		}
		polygon = append(polygon, point)
		nextIndex, ok := satPointer(coedge, 1)
		if !ok {
			break
		}
		coedgeIndex = nextIndex
	}
	return polygon, nil
}

func satPoint(record satRecord) (dxfmath.Vec3, error) {
	if len(record.data) < 4 {
		return dxfmath.Vec3{}, fmt.Errorf("acis mesh: malformed point record")
	}
	values := [3]float64{}
	for i := range 3 {
		value, err := strconv.ParseFloat(record.data[i+1], 64)
		if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
			return dxfmath.Vec3{}, fmt.Errorf("acis mesh: invalid point coordinate %q", record.data[i+1])
		}
		values[i] = value
	}
	return dxfmath.NewVec3(values[0], values[1], values[2]), nil
}

func satBodyTransform(records []satRecord) *[16]float64 {
	for _, record := range records {
		if record.name != "body" {
			continue
		}
		index, ok := satPointer(record, 3)
		if !ok || index >= len(records) || records[index].name != "transform" {
			return nil
		}
		transform := &[16]float64{}
		position := 0
		for _, token := range records[index].data {
			if position >= 12 {
				break
			}
			value, err := strconv.ParseFloat(token, 64)
			if err != nil {
				continue
			}
			row, column := position/3, position%3
			transform[row*4+column] = value
			position++
		}
		transform[3], transform[7], transform[11], transform[15] = 0, 0, 0, 1
		return transform
	}
	return nil
}

func transformPoint44(m [16]float64, p dxfmath.Vec3) dxfmath.Vec3 {
	return dxfmath.NewVec3(
		m[0]*p.X()+m[1]*p.Y()+m[2]*p.Z()+m[3],
		m[4]*p.X()+m[5]*p.Y()+m[6]*p.Z()+m[7],
		m[8]*p.X()+m[9]*p.Y()+m[10]*p.Z()+m[11],
	)
}

func satPointKey(p dxfmath.Vec3) string {
	return strconv.FormatFloat(p.X(), 'g', 12, 64) + "," +
		strconv.FormatFloat(p.Y(), 'g', 12, 64) + "," +
		strconv.FormatFloat(p.Z(), 'g', 12, 64)
}

func satDegenerate(indices []int) bool {
	unique := make(map[int]bool, len(indices))
	for _, index := range indices {
		unique[index] = true
	}
	return len(unique) < 3
}
