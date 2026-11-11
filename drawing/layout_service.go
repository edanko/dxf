// Layout authoring service: create, rename, delete, activate layouts with
// BLOCK_RECORD linkage. Layouts are stored as named blocks plus matching
// BLOCK_RECORD table entries; entities assigned to a layout carry the block
// record's handle as owner (code 330) on save.

package drawing

import (
	"fmt"
	"sort"
	"strings"

	"github.com/edanko/dxf/entity"
)

// layoutEntry is the registry record for one paper-space layout.
type layoutEntry struct {
	blockName string
	tabOrder  int
}

// NewLayout creates a new paper-space layout with a backing block and
// BLOCK_RECORD entry, and returns its authoring space.
func (d *Drawing) NewLayout(name string) (*Space, error) {
	if name == "" {
		return nil, fmt.Errorf("layout name is empty")
	}
	if strings.EqualFold(name, "Model") {
		return nil, fmt.Errorf("layout name %q is reserved for model space", name)
	}
	for reserved := range reservedBlockNames {
		if strings.EqualFold(name, reserved) {
			return nil, fmt.Errorf("layout name %q is reserved", name)
		}
	}
	if _, exists := d.layouts[name]; exists {
		return nil, fmt.Errorf("layout %q already exists", name)
	}
	if _, err := d.GetBlock(name); err == nil {
		return nil, fmt.Errorf("block %q already exists", name)
	}
	t := d.blockRecordTable()
	if t == nil {
		return nil, fmt.Errorf("BLOCK_RECORD table missing")
	}
	if _, err := t.Lookup(name); err == nil {
		return nil, fmt.Errorf("block record %q already exists", name)
	}
	if _, err := d.NewBlock(name, "Layout "+name); err != nil {
		return nil, err
	}
	d.layouts[name] = &layoutEntry{blockName: name, tabOrder: d.maxTabOrder() + 1}
	return &Space{drawing: d, spaceName: name, blockName: name}, nil
}

// reservedBlockNames are the special block names that layouts cannot use.
var reservedBlockNames = map[string]bool{
	"*Model_Space": true, "*Paper_Space": true, "*Paper_Space0": true,
}

func (d *Drawing) maxTabOrder() int {
	max := 0
	for _, le := range d.layouts {
		if le.tabOrder > max {
			max = le.tabOrder
		}
	}
	return max
}

// DeleteLayout removes a paper-space layout, its entities, block, and block
// record. Model space cannot be deleted.
func (d *Drawing) DeleteLayout(name string) error {
	if strings.EqualFold(name, "Model") {
		return fmt.Errorf("cannot delete model space")
	}
	le := d.layouts[name]
	if le == nil {
		return fmt.Errorf("layout %q does not exist", name)
	}
	for handleString, layout := range d.entityLayout {
		if layout == name {
			if err := d.DeleteEntity(handleString); err != nil {
				return fmt.Errorf("delete layout %q: %w", name, err)
			}
		}
	}
	if _, err := d.GetBlock(le.blockName); err == nil {
		if err := d.DeleteBlock(le.blockName); err != nil {
			return err
		}
	}
	// DeleteBlock removes the BLOCK_RECORD entry alongside the block.
	delete(d.layouts, name)
	if d.activeLayout == name {
		d.activeLayout = "Model"
	}
	return nil
}

// RenameLayout renames a paper-space layout, its block, and its block record.
func (d *Drawing) RenameLayout(oldName, newName string) error {
	if strings.EqualFold(oldName, "Model") || strings.EqualFold(newName, "Model") {
		return fmt.Errorf("model space cannot be renamed")
	}
	le := d.layouts[oldName]
	if le == nil {
		return fmt.Errorf("layout %q does not exist", oldName)
	}
	if newName == "" {
		return fmt.Errorf("layout name is empty")
	}
	if _, exists := d.layouts[newName]; exists {
		return fmt.Errorf("layout %q already exists", newName)
	}
	if _, err := d.GetBlock(newName); err == nil {
		return fmt.Errorf("block %q already exists", newName)
	}
	t := d.blockRecordTable()
	if t == nil {
		return fmt.Errorf("BLOCK_RECORD table missing")
	}
	if err := t.Rename(le.blockName, newName); err != nil {
		return err
	}
	if b, err := d.GetBlock(le.blockName); err == nil {
		b.Name = newName
	}
	// Re-assign entity layout keys.
	for handleString, layout := range d.entityLayout {
		if layout == oldName {
			d.entityLayout[handleString] = newName
		}
	}
	delete(d.layouts, oldName)
	le.blockName = newName
	d.layouts[newName] = le
	if d.activeLayout == oldName {
		d.activeLayout = newName
	}
	return nil
}

// SetActiveLayout selects the active layout tab ("Model" or an existing
// layout name).
func (d *Drawing) SetActiveLayout(name string) error {
	if strings.EqualFold(name, "Model") {
		d.activeLayout = "Model"
		return nil
	}
	if d.layouts[name] == nil {
		return fmt.Errorf("layout %q does not exist", name)
	}
	d.activeLayout = name
	return nil
}

// ActiveLayout returns the active layout name.
func (d *Drawing) ActiveLayout() string {
	if d.activeLayout == "" {
		return "Model"
	}
	return d.activeLayout
}

// ActiveSpace returns the authoring space of the active layout.
func (d *Drawing) ActiveSpace() *Space {
	s, err := d.SpaceByName(d.ActiveLayout())
	if err != nil {
		return d.Model()
	}
	return s
}

// LayoutNames returns "Model" followed by the paper-space layout names.
func (d *Drawing) LayoutNames() []string {
	names := []string{"Model"}
	for name := range d.layouts {
		names = append(names, name)
	}
	return names
}

// LayoutInfo describes one layout tab.
type LayoutInfo struct {
	Name        string
	Type        string // "ModelSpace" or "PaperSpace"
	TabOrder    int
	EntityCount int
}

// Layouts returns all layout infos, Model first, sorted by tab order.
func (d *Drawing) Layouts() []LayoutInfo {
	infos := []LayoutInfo{{
		Name:        "Model",
		Type:        "ModelSpace",
		TabOrder:    0,
		EntityCount: d.LayoutEntityCount("Model"),
	}}
	for name, le := range d.layouts {
		infos = append(infos, LayoutInfo{
			Name:        name,
			Type:        "PaperSpace",
			TabOrder:    le.tabOrder,
			EntityCount: d.LayoutEntityCount(name),
		})
	}
	sort.Slice(infos, func(i, j int) bool { return infos[i].TabOrder < infos[j].TabOrder })
	return infos
}

// LayoutEntityCount returns the number of entities assigned to a layout.
func (d *Drawing) LayoutEntityCount(layout string) int {
	model := strings.EqualFold(layout, "Model")
	count := 0
	for _, e := range d.entitiesSection {
		assigned := d.entityLayout[e.Handle()]
		if model {
			if assigned == "" {
				count++
			}
			continue
		}
		if assigned == layout {
			count++
		}
	}
	return count
}

// LayoutSpace returns the authoring space for a named layout.
func (d *Drawing) LayoutSpace(name string) (*Space, error) {
	if strings.EqualFold(name, "Model") {
		return d.Model(), nil
	}
	le := d.layouts[name]
	if le == nil {
		return nil, fmt.Errorf("layout %q does not exist", name)
	}
	return &Space{drawing: d, spaceName: name, blockName: le.blockName}, nil
}

// LayoutEntities returns the entities assigned to a layout; "Model" selects
// the unassigned (model-space) entities.
func (d *Drawing) LayoutEntities(layout string) []entity.Entity {
	model := strings.EqualFold(layout, "Model")
	out := []entity.Entity{}
	for _, e := range d.entitiesSection {
		assigned := d.entityLayout[e.Handle()]
		if model {
			if assigned == "" {
				out = append(out, e)
			}
			continue
		}
		if assigned == layout {
			out = append(out, e)
		}
	}
	return out
}
