package spatial

import (
	"math"
)

// BoundingBox represents a 2D bounding box for spatial indexing
type BoundingBox struct {
	MinX, MinY float64
	MaxX, MaxY float64
}

// NewBoundingBox creates a new bounding box
func NewBoundingBox(minX, minY, maxX, maxY float64) *BoundingBox {
	return &BoundingBox{
		MinX: minX,
		MinY: minY,
		MaxX: maxX,
		MaxY: maxY,
	}
}

// Contains returns true if point is inside bounding box
func (bb *BoundingBox) Contains(x, y float64) bool {
	return x >= bb.MinX && x <= bb.MaxX &&
		y >= bb.MinY && y <= bb.MaxY
}

// Intersects returns true if two bounding boxes intersect
func (bb *BoundingBox) Intersects(other *BoundingBox) bool {
	return bb.MaxX >= other.MinX && bb.MinX <= other.MaxX &&
		bb.MaxY >= other.MinY && bb.MinY <= other.MaxY
}

// Extend returns a new bounding box that contains both boxes
func (bb *BoundingBox) Extend(other *BoundingBox) *BoundingBox {
	minX := math.Min(bb.MinX, other.MinX)
	minY := math.Min(bb.MinY, other.MinY)
	maxX := math.Max(bb.MaxX, other.MaxX)
	maxY := math.Max(bb.MaxY, other.MaxY)
	return NewBoundingBox(minX, minY, maxX, maxY)
}

// Area returns the area of the bounding box
func (bb *BoundingBox) Area() float64 {
	width := bb.MaxX - bb.MinX
	height := bb.MaxY - bb.MinY
	return width * height
}

// Center returns the center point of the bounding box
func (bb *BoundingBox) Center() (float64, float64) {
	centerX := (bb.MinX + bb.MaxX) / 2.0
	centerY := (bb.MinY + bb.MaxY) / 2.0
	return centerX, centerY
}

// ContainsBox returns true if this box fully contains another box
func (bb *BoundingBox) ContainsBox(other *BoundingBox) bool {
	return bb.MinX <= other.MinX && bb.MaxX >= other.MaxX &&
		bb.MinY <= other.MinY && bb.MaxY >= other.MaxY
}
