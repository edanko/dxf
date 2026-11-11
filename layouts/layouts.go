package layouts

import (
	"fmt"
	"strings"

	"github.com/edanko/dxf/drawing"
)

// Layouts manages all layouts in a DXF drawing
type Layouts struct {
	drawing     *drawing.Drawing
	layouts     map[string]*Layout
	modelSpace  *ModelSpace
	paperSpaces []*PaperSpace
}

// NewLayouts creates a new layouts manager
func NewLayouts(d *drawing.Drawing) *Layouts {
	return &Layouts{
		drawing:     d,
		layouts:     make(map[string]*Layout),
		modelSpace:  nil,
		paperSpaces: make([]*PaperSpace, 0),
	}
}

// SetupLayouts initializes default layouts (model space and first paper space)
func (ls *Layouts) SetupLayouts() error {
	// Setup model space
	if err := ls.setupModelSpace(); err != nil {
		return fmt.Errorf("failed to setup model space: %w", err)
	}

	// Setup first paper space
	if err := ls.setupFirstPaperSpace(); err != nil {
		return fmt.Errorf("failed to setup first paper space: %w", err)
	}

	return nil
}

// setupModelSpace creates the model space layout
func (ls *Layouts) setupModelSpace() error {
	ms := NewModelSpace(ls.drawing)
	ls.modelSpace = ms
	ls.layouts[strings.ToUpper(ms.Name())] = ms.Layout
	return nil
}

// setupFirstPaperSpace creates the first paper space layout
func (ls *Layouts) setupFirstPaperSpace() error {
	ps := NewPaperSpace("Layout1", ls.drawing)
	ps.SetTabOrder(1)
	ls.paperSpaces = append(ls.paperSpaces, ps)
	ls.layouts[strings.ToUpper(ps.Name())] = ps.Layout
	return nil
}

// ModelSpace returns the model space layout
func (ls *Layouts) ModelSpace() *ModelSpace {
	return ls.modelSpace
}

// PaperSpaces returns all paper space layouts
func (ls *Layouts) PaperSpaces() []*PaperSpace {
	result := make([]*PaperSpace, len(ls.paperSpaces))
	copy(result, ls.paperSpaces)
	return result
}

// GetLayout returns a layout by name (case-insensitive)
func (ls *Layouts) GetLayout(name string) *Layout {
	normalized := strings.ToUpper(name)
	return ls.layouts[normalized]
}

// AddPaperSpace adds a new paper space layout
func (ls *Layouts) AddPaperSpace(name string) (*PaperSpace, error) {
	// Check for duplicate names
	if ls.GetLayout(name) != nil {
		return nil, fmt.Errorf("layout '%s' already exists", name)
	}

	// Find unique name
	uniqueName := name
	index := 1
	for ls.GetLayout(uniqueName) != nil {
		uniqueName = fmt.Sprintf("%s (%d)", name, index)
		index++
	}

	ps := NewPaperSpace(uniqueName, ls.drawing)
	ps.SetTabOrder(len(ls.layouts) + 1)

	ls.paperSpaces = append(ls.paperSpaces, ps)
	ls.layouts[strings.ToUpper(uniqueName)] = ps.Layout

	return ps, nil
}

// RemoveLayout removes a layout by name
func (ls *Layouts) RemoveLayout(name string) error {
	layout := ls.GetLayout(name)
	if layout == nil {
		return fmt.Errorf("layout '%s' not found", name)
	}

	// Cannot remove model space
	if layout.IsModelSpace() {
		return fmt.Errorf("cannot remove model space layout")
	}

	// Remove from layouts map
	delete(ls.layouts, strings.ToUpper(name))

	// Remove from paper spaces slice
	for i, ps := range ls.paperSpaces {
		if ps.Layout == layout {
			ls.paperSpaces = append(ls.paperSpaces[:i], ls.paperSpaces[i+1:]...)
			break
		}
	}

	return nil
}

// RenameLayout renames a layout
func (ls *Layouts) RenameLayout(oldName, newName string) error {
	layout := ls.GetLayout(oldName)
	if layout == nil {
		return fmt.Errorf("layout '%s' not found", oldName)
	}

	// Check for duplicate names
	if ls.GetLayout(newName) != nil {
		return fmt.Errorf("layout '%s' already exists", newName)
	}

	// Cannot rename model space
	if layout.IsModelSpace() {
		return fmt.Errorf("cannot rename model space layout")
	}

	// Update layouts map
	delete(ls.layouts, strings.ToUpper(oldName))
	layout.name = newName
	ls.layouts[strings.ToUpper(newName)] = layout

	return nil
}

// GetTabOrder returns the tab order of all layouts
func (ls *Layouts) GetTabOrder() []*Layout {
	// Create slice of all layouts
	allLayouts := make([]*Layout, 0, len(ls.layouts))
	allLayouts = append(allLayouts, ls.modelSpace.Layout)
	allLayouts = append(allLayouts, ls.layoutsArrayFromPaperSpaces()...)

	// Sort by tab order
	// Simple bubble sort for now
	for i := 0; i < len(allLayouts)-1; i++ {
		for j := 0; j < len(allLayouts)-i-1; j++ {
			if allLayouts[j].TabOrder() > allLayouts[j+1].TabOrder() {
				allLayouts[j], allLayouts[j+1] = allLayouts[j+1], allLayouts[j]
			}
		}
	}

	return allLayouts
}

// SetTabOrder sets the tab order of a layout
func (ls *Layouts) SetTabOrder(name string, order int) error {
	layout := ls.GetLayout(name)
	if layout == nil {
		return fmt.Errorf("layout '%s' not found", name)
	}

	layout.SetTabOrder(order)
	return nil
}

// LayoutCount returns the total number of layouts
func (ls *Layouts) LayoutCount() int {
	return len(ls.layouts)
}

// PaperSpaceCount returns the number of paper space layouts
func (ls *Layouts) PaperSpaceCount() int {
	return len(ls.paperSpaces)
}

// ForEach iterates over all layouts
func (ls *Layouts) ForEach(fn func(*Layout)) {
	// Visit model space first
	if ls.modelSpace != nil {
		fn(ls.modelSpace.Layout)
	}

	// Visit paper spaces in tab order
	tabOrder := ls.GetTabOrder()
	for _, layout := range tabOrder {
		if !layout.IsModelSpace() {
			fn(layout)
		}
	}
}

// String returns a string representation of the layouts manager
func (ls *Layouts) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Layouts{total=%d, model_space=1, paper_spaces=%d",
		ls.LayoutCount(), ls.PaperSpaceCount()))

	if len(ls.layouts) > 0 {
		sb.WriteString(", layouts=[")
		first := true
		for _, layout := range ls.layouts {
			if !first {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("'%s'", layout.Name()))
			first = false
		}
		sb.WriteString("]")
	}
	sb.WriteString("}")
	return sb.String()
}

// layoutsArrayFromPaperSpaces returns paper space layouts as array
func (ls *Layouts) layoutsArrayFromPaperSpaces() []*Layout {
	result := make([]*Layout, len(ls.paperSpaces))
	for i, ps := range ls.paperSpaces {
		result[i] = ps.Layout
	}
	return result
}

// GetLayoutByHandle returns a layout by handle
func (ls *Layouts) GetLayoutByHandle(handle string) *Layout {
	for _, layout := range ls.layouts {
		if layout.Handle() == handle {
			return layout
		}
	}
	return nil
}

// ValidateLayoutName checks if a layout name is valid
func ValidateLayoutName(name string) error {
	if name == "" {
		return fmt.Errorf("layout name cannot be empty")
	}

	// Check for invalid characters
	if strings.ContainsAny(name, "/\\:;*?\"<>|") {
		return fmt.Errorf("layout name contains invalid characters")
	}

	return nil
}

// CreateLayoutFromTemplate creates a new layout based on an existing one
func (ls *Layouts) CreateLayoutFromTemplate(templateName, newName string) (*Layout, error) {
	template := ls.GetLayout(templateName)
	if template == nil {
		return nil, fmt.Errorf("template layout '%s' not found", templateName)
	}

	if err := ValidateLayoutName(newName); err != nil {
		return nil, fmt.Errorf("invalid layout name: %w", err)
	}

	var newLayout *Layout
	if template.IsModelSpace() {
		return nil, fmt.Errorf("cannot create layout from model space template")
	} else {
		// Create new paper space based on template
		ps, err := ls.AddPaperSpace(newName)
		if err != nil {
			return nil, err
		}
		newLayout = ps.Layout

		// Copy properties from template
		if vp := template.Viewport(); vp != nil {
			// Create a copy of the viewport
			newVp := NewViewport()
			newVp.SetCenter(vp.Center()[0], vp.Center()[1], vp.Center()[2])
			newVp.SetWidth(vp.Width())
			newVp.SetHeight(vp.Height())
			newVp.SetLayer(vp.Layer())
			newLayout.SetViewport(newVp)
		}
	}

	return newLayout, nil
}
