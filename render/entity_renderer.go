package render

import (
	"fmt"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/entity"
)

// SimpleRenderer provides basic rendering functionality for DXF entities
type SimpleRenderer struct {
	backend RenderBackend
	ctx     *RenderContext
}

// NewSimpleRenderer creates a new simple renderer
func NewSimpleRenderer(backend RenderBackend) *SimpleRenderer {
	ctx := NewRenderContext()
	return &SimpleRenderer{
		backend: backend,
		ctx:     ctx,
	}
}

// RenderDrawing renders all entities from a drawing
func (er *SimpleRenderer) RenderDrawing(entities []entity.Entity) error {
	if err := er.backend.BeginPath(er.ctx); err != nil {
		return err
	}

	// Process each entity
	for _, entity := range entities {
		if err := er.renderEntity(entity); err != nil {
			return fmt.Errorf("error rendering entity: %w", err)
		}
	}

	return er.backend.EndPath(er.ctx)
}

// renderEntity renders a single entity using basic approach
func (er *SimpleRenderer) renderEntity(ent entity.Entity) error {
	// Get entity type and basic properties
	switch e := ent.(type) {
	case *entity.Line:
		return er.renderBasicLine(e)
	case *entity.Circle:
		return er.renderBasicCircle(e)
	case *entity.Arc:
		return er.renderBasicArc(e)
	case *entity.LwPolyline:
		return er.renderBasicLWPolyline(e)
	default:
		// Skip unsupported entities for now
		return nil
	}
}

// renderBasicLine renders a line entity with basic properties
func (er *SimpleRenderer) renderBasicLine(line *entity.Line) error {
	// Extract coordinates from line data
	if len(line.Start) >= 2 && len(line.End) >= 2 {
		if err := er.backend.MoveTo(line.Start[0], line.Start[1]); err != nil {
			return err
		}

		if err := er.backend.LineTo(line.End[0], line.End[1]); err != nil {
			return err
		}

		// Set default styling
		if err := er.backend.SetStrokeColor(color.White); err != nil {
			return err
		}

		if err := er.backend.SetLineWidth(1.0); err != nil {
			return err
		}

		return er.backend.Stroke()
	}
	return fmt.Errorf("invalid line coordinates")
}

// renderBasicCircle renders a circle entity with basic properties
func (er *SimpleRenderer) renderBasicCircle(circle *entity.Circle) error {
	// For now, render as a simple point (center)
	if len(circle.Center) >= 2 {
		if err := er.backend.MoveTo(circle.Center[0], circle.Center[1]); err != nil {
			return err
		}

		if err := er.backend.LineTo(circle.Center[0]+1, circle.Center[1]); err != nil {
			return err
		}

		// Set default styling
		if err := er.backend.SetStrokeColor(color.White); err != nil {
			return err
		}

		if err := er.backend.SetLineWidth(1.0); err != nil {
			return err
		}

		return er.backend.Stroke()
	}
	return fmt.Errorf("invalid circle coordinates")
}

// renderBasicArc renders an arc entity with basic properties
func (er *SimpleRenderer) renderBasicArc(arc *entity.Arc) error {
	// For now, render as a simple point (center)
	if len(arc.Center) >= 2 {
		if err := er.backend.MoveTo(arc.Center[0], arc.Center[1]); err != nil {
			return err
		}

		if err := er.backend.LineTo(arc.Center[0]+1, arc.Center[1]+1); err != nil {
			return err
		}

		// Set default styling
		if err := er.backend.SetStrokeColor(color.White); err != nil {
			return err
		}

		if err := er.backend.SetLineWidth(1.0); err != nil {
			return err
		}

		return er.backend.Stroke()
	}
	return fmt.Errorf("invalid arc coordinates")
}

// renderBasicLWPolyline renders a LWPOLYLINE entity
func (er *SimpleRenderer) renderBasicLWPolyline(lwpolyline *entity.LwPolyline) error {
	if len(lwpolyline.Vertices) < 2 {
		return fmt.Errorf("lightweight polyline has less than 2 vertices")
	}

	if err := er.backend.BeginPath(er.ctx); err != nil {
		return err
	}

	// Move to first vertex
	if len(lwpolyline.Vertices) > 0 && len(lwpolyline.Vertices[0]) >= 2 {
		if err := er.backend.MoveTo(lwpolyline.Vertices[0][0], lwpolyline.Vertices[0][1]); err != nil {
			return err
		}
	}

	// Draw lines to remaining vertices
	for i := 1; i < len(lwpolyline.Vertices); i++ {
		if len(lwpolyline.Vertices[i]) >= 2 {
			v := lwpolyline.Vertices[i]
			if err := er.backend.LineTo(v[0], v[1]); err != nil {
				return err
			}
		}
	}

	// Set default styling
	if err := er.backend.SetStrokeColor(color.White); err != nil {
		return err
	}

	if err := er.backend.SetLineWidth(1.0); err != nil {
		return err
	}

	return er.backend.Stroke()
}

// SetContext updates the render context
func (er *SimpleRenderer) SetContext(ctx *RenderContext) {
	er.ctx = ctx
}

// GetContext returns the current render context
func (er *SimpleRenderer) GetContext() *RenderContext {
	return er.ctx
}

// Flush flushes the backend
func (er *SimpleRenderer) Flush() error {
	return er.backend.Flush()
}

// GetBounds returns the drawing bounds
func (er *SimpleRenderer) GetBounds() (float64, float64, float64, float64) {
	return er.backend.GetBounds()
}
