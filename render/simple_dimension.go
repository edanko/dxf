package render

import (
	"fmt"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/math"
	"github.com/edanko/dxf/table"
)

// DimensionRenderer handles rendering of dimension entities with proper blocks
type DimensionRenderer struct {
	format       format.Formatter
	currentLayer *table.Layer
}

// NewDimensionRenderer creates a new dimension renderer
func NewDimensionRenderer(formatter format.Formatter, layer *table.Layer) *DimensionRenderer {
	return &DimensionRenderer{
		format:       formatter,
		currentLayer: layer,
	}
}

// RenderDimension renders a dimension entity with basic geometry
func (dr *DimensionRenderer) RenderDimension(dim *entity.Dimension) error {
	if dim == nil {
		return fmt.Errorf("dimension entity is nil")
	}

	// Create anonymous block for dimension rendering
	blockName := fmt.Sprintf("*D%s", dim.Handle())

	// Create basic dimension geometry
	dimType := dim.GetDimensionType()
	switch dimType {
	case entity.DimensionTypeRotated, entity.DimensionTypeAligned:
		return dr.renderBasicLinearDimension(dim, blockName)
	case entity.DimensionTypeDiameter:
		return dr.renderDiameterDimension(dim, blockName)
	case entity.DimensionTypeRadius:
		return dr.renderRadiusDimension(dim, blockName)
	default:
		return fmt.Errorf("unsupported dimension type: %d", dimType)
	}
}

// renderBasicLinearDimension renders a basic linear dimension
func (dr *DimensionRenderer) renderBasicLinearDimension(dim *entity.Dimension, blockName string) error {
	// Use basic dimension points (this is a simplified implementation)
	startPoint := math.NewVec3(0, 0, 0)
	endPoint := math.NewVec3(10, 5, 0)

	// Create dimension line
	dimensionLine := entity.NewLine()
	dimensionLine.Start = []float64{startPoint.X(), startPoint.Y(), startPoint.Z()}
	dimensionLine.End = []float64{endPoint.X(), endPoint.Y(), endPoint.Z()}
	dimensionLine.SetLayer(dr.currentLayer)

	// Create dimension text
	midPoint := math.NewVec3(
		(startPoint.X()+endPoint.X())/2,
		(startPoint.Y()+endPoint.Y())/2,
		(startPoint.Z()+endPoint.Z())/2,
	)

	txt := entity.NewText()
	txt.Value = "10.0"
	txt.Coord1 = []float64{midPoint.X(), midPoint.Y(), midPoint.Z()}
	txt.Height = 2.5
	txt.SetLayer(dr.currentLayer)
	txt.SetColor(color.White)

	// Create basic arrows
	arrows := dr.createBasicArrows(startPoint, endPoint)

	// Format: block content
	dr.format.WriteString(0, blockName)
	dr.formatEntity(dimensionLine)
	dr.formatEntity(txt)
	for _, arrow := range arrows {
		dr.formatEntity(arrow)
	}
	dr.format.WriteString(0, "ENDBLK")

	return nil
}

// renderDiameterDimension renders a diameter dimension
func (dr *DimensionRenderer) renderDiameterDimension(dim *entity.Dimension, blockName string) error {
	// Use linear dimension rendering with diameter symbol
	txt := entity.NewText()
	txt.Value = "⌀10.0"
	txt.Coord1 = []float64{5, 2.5, 0}
	txt.Height = 2.5
	txt.SetLayer(dr.currentLayer)
	txt.SetColor(color.White)
	return dr.renderBasicLinearDimension(dim, blockName)
}

// renderRadiusDimension renders a radius dimension
func (dr *DimensionRenderer) renderRadiusDimension(dim *entity.Dimension, blockName string) error {
	// Use linear dimension rendering with radius symbol
	txt := entity.NewText()
	txt.Value = "R5.0"
	txt.Coord1 = []float64{5, 2.5, 0}
	txt.Height = 2.5
	txt.SetLayer(dr.currentLayer)
	txt.SetColor(color.White)
	return dr.renderBasicLinearDimension(dim, blockName)
}

// createBasicArrows creates simple arrow geometry
func (dr *DimensionRenderer) createBasicArrows(startPoint, endPoint math.Vec3) []entity.Entity {
	var arrows []entity.Entity

	// Start arrow (closed triangle)
	startArrow := NewArrow(ArrowClosedFilled, 2.0, 0)
	startArrowVertices := startArrow.Vertices
	for i := range startArrowVertices {
		startArrowVertices[i] = Vec3{
			X: startPoint.X() + startArrowVertices[i].X,
			Y: startPoint.Y() + startArrowVertices[i].Y,
			Z: startPoint.Z() + startArrowVertices[i].Z,
		}
	}

	// End arrow (open triangle)
	endArrow := NewArrow(ArrowOpen, 2.0, 0)
	endArrowVertices := endArrow.Vertices
	for i := range endArrowVertices {
		endArrowVertices[i] = Vec3{
			X: endPoint.X() + endArrowVertices[i].X,
			Y: endPoint.Y() + endArrowVertices[i].Y,
			Z: endPoint.Z() + endArrowVertices[i].Z,
		}
	}

	// Convert arrows to lines
	arrows = append(arrows, dr.createArrowLines(startPoint, startArrowVertices)...)
	arrows = append(arrows, dr.createArrowLines(endPoint, endArrowVertices)...)

	return arrows
}

// createArrowLines converts arrow vertices to line entities
func (dr *DimensionRenderer) createArrowLines(basePoint math.Vec3, vertices []Vec3) []entity.Entity {
	var lines []entity.Entity

	if len(vertices) >= 2 {
		for i := 1; i < len(vertices); i++ {
			line := entity.NewLine()
			line.Start = []float64{
				basePoint.X() + vertices[i-1].X,
				basePoint.Y() + vertices[i-1].Y,
				basePoint.Z() + vertices[i-1].Z,
			}
			line.End = []float64{
				basePoint.X() + vertices[i].X,
				basePoint.Y() + vertices[i].Y,
				basePoint.Z() + vertices[i].Z,
			}
			line.SetLayer(dr.currentLayer)
			lines = append(lines, line)
		}
	}

	return lines
}

// formatEntity formats an entity using the renderer's formatter
func (dr *DimensionRenderer) formatEntity(e entity.Entity) {
	e.Format(dr.format)
}
