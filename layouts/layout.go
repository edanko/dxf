package layouts

import (
	"fmt"

	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/table"
)

// LayoutType represents different types of layouts
type LayoutType int

const (
	LayoutTypeModelSpace LayoutType = iota
	LayoutTypePaperSpace
)

// Layout represents a DXF layout (model space or paper space)
type Layout struct {
	name        string
	layoutType  LayoutType
	tabOrder    int
	drawing     *drawing.Drawing
	entities    []entity.Entity
	blockRecord *table.BlockRecord
	handle      string
	viewport    *Viewport
}

// NewLayout creates a new layout
func NewLayout(name string, layoutType LayoutType, d *drawing.Drawing) *Layout {
	return &Layout{
		name:       name,
		layoutType: layoutType,
		tabOrder:   0,
		drawing:    d,
		entities:   make([]entity.Entity, 0),
		handle:     "",
		viewport:   nil,
	}
}

// Name returns the layout name
func (l *Layout) Name() string {
	return l.name
}

// Type returns the layout type
func (l *Layout) Type() LayoutType {
	return l.layoutType
}

// IsModelSpace returns true if this is a model space layout
func (l *Layout) IsModelSpace() bool {
	return l.layoutType == LayoutTypeModelSpace
}

// IsPaperSpace returns true if this is a paper space layout
func (l *Layout) IsPaperSpace() bool {
	return l.layoutType == LayoutTypePaperSpace
}

// TabOrder returns the tab order of the layout
func (l *Layout) TabOrder() int {
	return l.tabOrder
}

// SetTabOrder sets the tab order of the layout
func (l *Layout) SetTabOrder(order int) {
	l.tabOrder = order
}

// Handle returns the layout handle
func (l *Layout) Handle() string {
	return l.handle
}

// SetHandle sets the layout handle
func (l *Layout) SetHandle(h string) {
	l.handle = h
}

// BlockRecord returns the associated block record
func (l *Layout) BlockRecord() *table.BlockRecord {
	return l.blockRecord
}

// SetBlockRecord sets the associated block record
func (l *Layout) SetBlockRecord(br *table.BlockRecord) {
	l.blockRecord = br
}

// Viewport returns the associated viewport (for paper space layouts)
func (l *Layout) Viewport() *Viewport {
	return l.viewport
}

// SetViewport sets the associated viewport (for paper space layouts)
func (l *Layout) SetViewport(vp *Viewport) {
	l.viewport = vp
}

// AddEntity adds an entity to the layout
func (l *Layout) AddEntity(ent entity.Entity) {
	l.entities = append(l.entities, ent)
}

// RemoveEntity removes an entity from the layout
func (l *Layout) RemoveEntity(ent entity.Entity) bool {
	for i, existing := range l.entities {
		if existing.Handle() == ent.Handle() {
			l.entities = append(l.entities[:i], l.entities[i+1:]...)
			return true
		}
	}
	return false
}

// Entities returns all entities in the layout
func (l *Layout) Entities() []entity.Entity {
	result := make([]entity.Entity, len(l.entities))
	copy(result, l.entities)
	return result
}

// EntityCount returns the number of entities in the layout
func (l *Layout) EntityCount() int {
	return len(l.entities)
}

// Clear removes all entities from the layout
func (l *Layout) Clear() {
	l.entities = make([]entity.Entity, 0)
}

// String returns a string representation of the layout
func (l *Layout) String() string {
	layoutTypeStr := "Paper Space"
	if l.IsModelSpace() {
		layoutTypeStr = "Model Space"
	}
	return fmt.Sprintf("Layout{name='%s', type=%s, entities=%d}", l.name, layoutTypeStr, len(l.entities))
}

// ModelSpace represents the model space layout
type ModelSpace struct {
	*Layout
}

// NewModelSpace creates a new model space layout
func NewModelSpace(d *drawing.Drawing) *ModelSpace {
	layout := NewLayout("Model", LayoutTypeModelSpace, d)
	layout.SetTabOrder(0)
	return &ModelSpace{Layout: layout}
}

// PaperSpace represents a paper space layout
type PaperSpace struct {
	*Layout
}

// NewPaperSpace creates a new paper space layout
func NewPaperSpace(name string, d *drawing.Drawing) *PaperSpace {
	layout := NewLayout(name, LayoutTypePaperSpace, d)
	return &PaperSpace{Layout: layout}
}

// Viewport represents a paper space viewport
type Viewport struct {
	handle         string
	center         []float64      // 10,20,30
	width          float64        // 40
	height         float64        // 41
	viewDirection  []float64      // 16,26,36
	viewTarget     []float64      // 17,27,37
	viewHeight     float64        // 45
	viewTwistAngle float64        // 50
	lensLength     float64        // 42
	frontClipZ     float64        // 43
	backClipZ      float64        // 44
	snapAngle      float64        // 50
	viewMode       int            // 71
	viewCircleZoom float64        // 42
	fastZoom       int            // 73
	ucsIcon        int            // 74
	gridOn         int            // 76
	gridMajor      int            // 77
	snapsOn        int            // 75
	snapStyle      int            // 77
	snapIsopair    int            // 78
	viewportID     int            // 69
	status         int            // 68
	layer          string         // 8
	linetype       string         // 6
	color          int            // 62
	lineweight     int            // 370
	plotStyleName  string         // 1
	plotStyleID    handle.Handler // 390
}

// NewViewport creates a new viewport
func NewViewport() *Viewport {
	return &Viewport{
		center:         []float64{0, 0, 0},
		width:          1.0,
		height:         1.0,
		viewDirection:  []float64{0, 0, 1},
		viewTarget:     []float64{0, 0, 0},
		viewHeight:     1.0,
		viewTwistAngle: 0.0,
		lensLength:     50.0,
		frontClipZ:     0.0,
		backClipZ:      0.0,
		snapAngle:      0.0,
		viewMode:       0,
		viewCircleZoom: 1.0,
		fastZoom:       0,
		ucsIcon:        1,
		gridOn:         0,
		gridMajor:      0,
		snapsOn:        0,
		snapStyle:      0,
		snapIsopair:    0,
		viewportID:     1,
		status:         1,
		layer:          "0",
		linetype:       "CONTINUOUS",
		color:          256,
		lineweight:     -1,
		plotStyleName:  "",
		plotStyleID:    nil,
	}
}

// Handle returns the viewport handle
func (vp *Viewport) Handle() string {
	return vp.handle
}

// SetHandle sets the viewport handle
func (vp *Viewport) SetHandle(h string) {
	vp.handle = h
}

// Center returns the viewport center point
func (vp *Viewport) Center() []float64 {
	result := make([]float64, 3)
	copy(result, vp.center)
	return result
}

// SetCenter sets the viewport center point
func (vp *Viewport) SetCenter(x, y, z float64) {
	vp.center = []float64{x, y, z}
}

// Width returns the viewport width
func (vp *Viewport) Width() float64 {
	return vp.width
}

// SetWidth sets the viewport width
func (vp *Viewport) SetWidth(width float64) {
	vp.width = width
}

// Height returns the viewport height
func (vp *Viewport) Height() float64 {
	return vp.height
}

// SetHeight sets the viewport height
func (vp *Viewport) SetHeight(height float64) {
	vp.height = height
}

// ViewDirection returns the view direction vector
func (vp *Viewport) ViewDirection() []float64 {
	result := make([]float64, 3)
	copy(result, vp.viewDirection)
	return result
}

// SetViewDirection sets the view direction vector
func (vp *Viewport) SetViewDirection(x, y, z float64) {
	vp.viewDirection = []float64{x, y, z}
}

// ViewTarget returns the view target point
func (vp *Viewport) ViewTarget() []float64 {
	result := make([]float64, 3)
	copy(result, vp.viewTarget)
	return result
}

// SetViewTarget sets the view target point
func (vp *Viewport) SetViewTarget(x, y, z float64) {
	vp.viewTarget = []float64{x, y, z}
}

// ViewHeight returns the view height
func (vp *Viewport) ViewHeight() float64 {
	return vp.viewHeight
}

// SetViewHeight sets the view height
func (vp *Viewport) SetViewHeight(height float64) {
	vp.viewHeight = height
}

// ViewTwistAngle returns the view twist angle in radians
func (vp *Viewport) ViewTwistAngle() float64 {
	return vp.viewTwistAngle
}

// SetViewTwistAngle sets the view twist angle in radians
func (vp *Viewport) SetViewTwistAngle(angle float64) {
	vp.viewTwistAngle = angle
}

// Layer returns the viewport layer name
func (vp *Viewport) Layer() string {
	return vp.layer
}

// SetLayer sets the viewport layer name
func (vp *Viewport) SetLayer(layer string) {
	vp.layer = layer
}

// String returns a string representation of the viewport
func (vp *Viewport) String() string {
	return fmt.Sprintf("Viewport{center=[%.2f,%.2f,%.2f], size=[%.2fx%.2f], layer='%s'}",
		vp.center[0], vp.center[1], vp.center[2], vp.width, vp.height, vp.layer)
}
