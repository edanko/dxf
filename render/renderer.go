package render

import (
	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/math"
)

// RenderContext holds rendering state and configuration
type RenderContext struct {
	BackgroundColor color.ColorNumber
	LineWidth       float64
	Scale           float64
	Transform       math.Matrix44
}

// NewRenderContext creates a new render context
func NewRenderContext() *RenderContext {
	identity := math.Identity()
	return &RenderContext{
		BackgroundColor: color.White,
		LineWidth:       1.0,
		Scale:           1.0,
		Transform:       identity,
	}
}
