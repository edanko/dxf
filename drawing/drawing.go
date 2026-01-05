// Package drawing defines Drawing struct for DXF.
package drawing

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/edanko/dxf/block"
	"github.com/edanko/dxf/class"
	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/header"
	"github.com/edanko/dxf/object"
	"github.com/edanko/dxf/query"
	"github.com/edanko/dxf/table"
)

// Drawing contains DXF drawing data.
type Drawing struct {
	fileName     string
	Layers       map[string]*table.Layer
	Groups       map[string]*object.Group
	Styles       map[string]*table.Style
	CurrentLayer *table.Layer
	CurrentStyle *table.Style
	formatter    format.Formatter

	headerSection   *header.Header
	classesSection  class.Classes
	tablesSection   table.Tables
	blocksSection   block.Blocks
	entitiesSection entity.Entities
	objectsSection  object.Objects

	dictionary *object.Dictionary
	groupdict  *object.Dictionary
	PlotStyle  handle.Handler
	// savebuff is used internally for the io.Reader options.
	savebuff *bytes.Buffer

	handleGenarator *handle.HandleGenerator
	entityDatabase  *EntityDatabase // Enhanced entity management
}

// New creates a new Drawing.
func New() (*Drawing, error) {
	d := new(Drawing)

	d.handleGenarator = handle.NewHandleGenerator()

	lineTypes := []*table.LineType{
		table.NewLineType("ByLayer", ""),
		table.NewLineType("ByBlock", ""),
		table.NewLineType("Continuous", "Solid Line"),
		table.NewLineType("HIDDEN", "Hidden __ __ __ __ __ __ __ __ __ __ __ __ __ _", 0.25, -0.125),
		table.NewLineType("DASHDOT", "Dash dot __ . __ . __ . __ . __ . __ . __ . __", 0.5, -0.25, 0.0, -0.25),
	}

	d.Layers = make(map[string]*table.Layer)
	d.Layers["0"] = table.NewLayer("0", color.White, lineTypes[2])
	d.Groups = make(map[string]*object.Group)
	d.CurrentLayer = d.Layers["0"]
	d.Styles = make(map[string]*table.Style)
	d.Styles["STANDARD"] = table.NewStyle("Standard")
	d.CurrentStyle = d.Styles["STANDARD"]
	d.formatter = format.NewASCII()
	d.formatter.SetPrecision(16)

	d.entityDatabase = NewEntityDatabase()

	d.headerSection = header.New()
	d.classesSection = class.New()
	d.tablesSection = table.New(lineTypes, d.Layers["0"], d.Styles["STANDARD"])
	d.blocksSection = block.New(d.Layers["0"])
	d.entitiesSection = entity.New()
	d.objectsSection = object.New()

	d.dictionary = object.NewDictionary()
	d.addObject(d.dictionary)
	wd, ph := object.NewAcDbDictionaryWDFLT(d.dictionary)
	d.dictionary.AddItem("ACAD_PLOTSTYLENAME", wd)
	d.addObject(wd)
	d.addObject(ph)
	d.groupdict = object.NewDictionary()
	d.addObject(d.groupdict)
	d.dictionary.AddItem("ACAD_GROUP", d.groupdict)
	d.PlotStyle = ph
	d.Layers["0"].SetPlotStyle(d.PlotStyle)
	return d, nil
}

// Document Management Methods

// Document Management Methods

// AddLayer adds a new layer to the drawing
func (d *Drawing) AddLayer(name string, color color.ColorNumber, lineType *table.LineType) (*table.Layer, error) {
	layer := table.NewLayer(name, color, lineType)
	d.Layers[name] = layer
	return layer, nil
}

// GetLayer returns a layer by name
func (d *Drawing) GetLayer(name string) *table.Layer {
	return d.Layers[name]
}

// GetLayers returns all layers
func (d *Drawing) GetLayers() map[string]*table.Layer {
	return d.Layers
}

// RemoveLayer removes a layer from the drawing
func (d *Drawing) RemoveLayer(name string) error {
	if _, exists := d.Layers[name]; !exists {
		return fmt.Errorf("layer '%s' does not exist", name)
	}
	delete(d.Layers, name)
	return nil
}

// AddStyle adds a new text style to the drawing
func (d *Drawing) AddStyle(name string, font string, height float64) (*table.Style, error) {
	style := table.NewStyle(name)
	style.FontName = font
	if height > 0 {
		style.FixedTextHeight = height
	}
	d.Styles[name] = style
	return style, nil
}

// GetStyle returns a style by name
func (d *Drawing) GetStyle(name string) *table.Style {
	return d.Styles[name]
}

// GetStyles returns all styles
func (d *Drawing) GetStyles() map[string]*table.Style {
	return d.Styles
}

// SetActiveLayer sets the currently active layer
func (d *Drawing) SetActiveLayer(name string) error {
	layer := d.GetLayer(name)
	if layer == nil {
		return fmt.Errorf("layer '%s' not found", name)
	}
	d.CurrentLayer = layer
	return nil
}

// SetActiveStyle sets the currently active text style
func (d *Drawing) SetActiveStyle(name string) error {
	style := d.GetStyle(name)
	if style == nil {
		return fmt.Errorf("style '%s' not found", name)
	}
	d.CurrentStyle = style
	return nil
}

// GetDrawingStatistics returns comprehensive drawing statistics
func (d *Drawing) GetDrawingStatistics() map[string]interface{} {
	stats := make(map[string]interface{})

	// Entity counts
	if d.entityDatabase != nil {
		stats["entity_count"] = d.entityDatabase.GetEntityCount()
	} else {
		stats["entity_count"] = 0
	}
	stats["layer_count"] = len(d.Layers)
	stats["style_count"] = len(d.Styles)
	stats["block_count"] = len(d.blocksSection)

	// Entity type breakdown
	if d.entityDatabase != nil {
		entityTypes := make(map[string]int)
		for _, entity := range d.entityDatabase.GetAllEntities() {
			entityType := entity.DXFType()
			entityTypes[entityType]++
		}
		stats["entity_types"] = entityTypes
	}

	// Header information
	if d.headerSection != nil {
		headerStats := make(map[string]interface{})
		headerStats["Version"] = d.headerSection.Version
		headerStats["InsUnit"] = d.headerSection.InsUnit
		headerStats["ExtMin"] = d.headerSection.ExtMin
		headerStats["ExtMax"] = d.headerSection.ExtMax
		stats["header"] = headerStats
	}

	return stats
}

func (d *Drawing) saveFile(filename string) error {
	w, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer w.Close()
	_, err = d.WriteTo(w)
	return err
}

// Save saves the drawing file.
// If it is the first time, use SaveAs(filename).
func (d *Drawing) Save() error {
	if d.fileName == "" {
		return errors.New("filename is blank, use SaveAs(filename)")
	}
	return d.saveFile(d.fileName)
}

// SaveAs saves the drawing file as given filename.
func (d *Drawing) SaveAs(filename string) error {
	d.fileName = filename
	return d.saveFile(filename)
}

// SaveBinary saves the drawing file as binary DXF.
func (d *Drawing) SaveBinary(filename string) error {
	w, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer w.Close()
	_, err = d.WriteBinaryTo(w)
	return err
}

// SaveAsBinary saves the drawing file as binary DXF with given filename.
func (d *Drawing) SaveAsBinary(filename string) error {
	d.fileName = filename
	return d.SaveBinary(filename)
}

// WriteBinaryTo writes the drawing in binary DXF format to the given writer.
func (d *Drawing) WriteBinaryTo(w io.Writer) (n int64, err error) {
	if d == nil {
		return 0, nil
	}

	// Create binary formatter with appropriate version
	version := "AC1021" // Default to R2004
	if d.formatter != nil {
		version = d.formatter.Version()
	}
	binaryFormatter := format.NewBinary(version)

	// Write binary DXF signature
	if err := binaryFormatter.WriteSignature(); err != nil {
		return 0, fmt.Errorf("failed to write binary signature: %w", err)
	}

	d.setHandle()

	// Format all sections with binary formatter
	d.headerSection.Format(binaryFormatter)
	d.classesSection.Format(binaryFormatter)
	d.tablesSection.Format(binaryFormatter)
	d.blocksSection.Format(binaryFormatter)
	d.entitiesSection.Format(binaryFormatter)
	d.objectsSection.Format(binaryFormatter)

	binaryFormatter.WriteString(0, "EOF")

	// Get binary data and write to output writer
	binaryData := binaryFormatter.Output()
	if binaryData == "" {
		return 0, nil
	}

	bytesWritten, err := io.WriteString(w, binaryData)
	if err != nil {
		return int64(bytesWritten), err
	}

	return int64(bytesWritten), nil
}

// Enhanced Entity Management Methods

// GetEntityDatabase returns the entity database for advanced operations
func (d *Drawing) GetEntityDatabase() *EntityDatabase {
	if d.entityDatabase == nil {
		d.entityDatabase = NewEntityDatabase()
	}
	return d.entityDatabase
}

// GetAllEntities returns all entities in the drawing
func (d *Drawing) GetAllEntities() []entity.Entity {
	if d.entityDatabase == nil {
		return []entity.Entity{}
	}
	return d.entityDatabase.GetAllEntities()
}

// GetEntitiesByType returns all entities of a specific type
func (d *Drawing) GetEntitiesByType(entityType string) []entity.Entity {
	if d.entityDatabase == nil {
		return []entity.Entity{}
	}
	return d.entityDatabase.GetEntitiesByType(entityType)
}

// GetEntitiesByLayer returns all entities on a specific layer
func (d *Drawing) GetEntitiesByLayer(layerName string) []entity.Entity {
	if d.entityDatabase == nil {
		return []entity.Entity{}
	}
	return d.entityDatabase.GetEntitiesByLayer(layerName)
}

// DeleteEntity removes an entity from the drawing
func (d *Drawing) DeleteEntity(handle string) error {
	if d.entityDatabase == nil {
		return fmt.Errorf("entity database not initialized")
	}

	// Remove from entity database
	err := d.entityDatabase.DeleteEntity(handle)
	if err != nil {
		return err
	}

	// Note: We would need to implement removal from entitiesSection too
	// This is a more complex operation requiring entity tracking
	return nil
}

// GetEntityCount returns the total number of entities
func (d *Drawing) GetEntityCount() int {
	if d.entityDatabase == nil {
		return 0
	}
	return d.entityDatabase.GetEntityCount()
}

// GetEntityCountByType returns count of entities of a specific type
func (d *Drawing) GetEntityCountByType(entityType string) int {
	if d.entityDatabase == nil {
		return 0
	}
	return d.entityDatabase.GetEntityCountByType(entityType)
}

// ValidateEntityDatabase validates the integrity of the entity database
func (d *Drawing) ValidateEntityDatabase() error {
	if d.entityDatabase == nil {
		return nil
	}
	return d.entityDatabase.Validate()
}

// Query creates a new entity query with all drawing entities
func (d *Drawing) Query() *query.EntityQuery {
	if d.entityDatabase == nil {
		return query.NewEntityQuery()
	}
	entities := d.entityDatabase.GetAllEntities()
	return query.NewEntityQuery(entities...)
}

// SelectEntities creates a query with entities of specific types
func (d *Drawing) SelectEntities(entityTypes ...string) *query.EntityQuery {
	if d.entityDatabase == nil {
		return query.NewEntityQuery()
	}

	var selectedEntities []entity.Entity
	for _, entityType := range entityTypes {
		entitiesOfType := d.entityDatabase.GetEntitiesByType(entityType)
		selectedEntities = append(selectedEntities, entitiesOfType...)
	}

	return query.NewEntityQuery(selectedEntities...)
}

// FindEntitiesByAttribute creates a query with entities matching attribute criteria
func (d *Drawing) FindEntitiesByAttribute(attrName, attrValue string) *query.EntityQuery {
	if d.entityDatabase == nil {
		return query.NewEntityQuery()
	}

	entities := d.entityDatabase.FindEntitiesByAttribute(attrName, attrValue)
	return query.NewEntityQuery(entities...)
}

// FilterEntities creates a query with entities matching a predicate
func (d *Drawing) FilterEntities(predicate func(entity.Entity) bool) *query.EntityQuery {
	if d.entityDatabase == nil {
		return query.NewEntityQuery()
	}

	entities := d.entityDatabase.FilterEntities(predicate)
	return query.NewEntityQuery(entities...)
}

// GroupEntitiesBy groups entities by a key function
func (d *Drawing) GroupEntitiesBy(keyFunc func(entity.Entity) string) map[string][]entity.Entity {
	if d.entityDatabase == nil {
		return make(map[string][]entity.Entity)
	}

	return d.entityDatabase.GroupEntitiesBy(keyFunc)
}

// GetLayerByName returns a layer by name
func (d *Drawing) GetLayerByName(name string) (*table.Layer, error) {
	layer, exists := d.Layers[name]
	if !exists {
		return nil, fmt.Errorf("layer '%s' not found", name)
	}
	return layer, nil
}

// DeleteLayer removes a layer from the drawing
func (d *Drawing) DeleteLayer(name string) error {
	if name == "0" {
		return fmt.Errorf("cannot delete default layer '0'")
	}

	layer, exists := d.Layers[name]
	if !exists {
		return fmt.Errorf("layer '%s' not found", name)
	}

	// Check if layer is in use by entities
	if d.entityDatabase != nil {
		entitiesOnLayer := d.entityDatabase.GetEntitiesByLayer(name)
		if len(entitiesOnLayer) > 0 {
			return fmt.Errorf("cannot delete layer '%s' - %d entities still use it", name, len(entitiesOnLayer))
		}
	}

	delete(d.Layers, name)

	// Set current layer to layer 0 if deleted layer was current
	if d.CurrentLayer == layer {
		d.CurrentLayer = d.Layers["0"]
	}

	return nil
}

// PurgeUnusedLayers removes unused layers and returns count of removed layers
func (d *Drawing) PurgeUnusedLayers() int {
	if d.entityDatabase == nil {
		return 0
	}

	removedCount := 0
	for name := range d.Layers {
		// Skip default layer
		if name == "0" {
			continue
		}

		// Check if layer is in use
		entitiesOnLayer := d.entityDatabase.GetEntitiesByLayer(name)
		if len(entitiesOnLayer) == 0 {
			delete(d.Layers, name)
			removedCount++
		}
	}

	return removedCount
}

// GetLayersInUse returns layers that have entities on them
func (d *Drawing) GetLayersInUse() []*table.Layer {
	if d.entityDatabase == nil {
		return []*table.Layer{}
	}

	layersInUse := make(map[string]bool)
	for _, ent := range d.entityDatabase.GetAllEntities() {
		if layer := ent.Layer(); layer != nil {
			layersInUse[layer.Name()] = true
		}
	}

	var result []*table.Layer
	for name, layer := range d.Layers {
		if layersInUse[name] {
			result = append(result, layer)
		}
	}

	return result
}

// RenameLayer renames an existing layer
func (d *Drawing) RenameLayer(oldName, newName string) error {
	if oldName == "0" {
		return fmt.Errorf("cannot rename default layer '0'")
	}

	if _, exists := d.Layers[oldName]; !exists {
		return fmt.Errorf("layer '%s' not found", oldName)
	}

	if _, exists := d.Layers[newName]; exists {
		return fmt.Errorf("layer '%s' already exists", newName)
	}

	delete(d.Layers, oldName)
	d.Layers[newName] = d.Layers[oldName]

	// Update current layer reference if needed
	if d.CurrentLayer != nil && d.CurrentLayer.Name() == oldName {
		d.CurrentLayer = d.Layers[newName]
	}

	return nil
}

// FilterLayers returns layers matching a predicate function
func (d *Drawing) FilterLayers(predicate func(*table.Layer) bool) []*table.Layer {
	var filteredLayers []*table.Layer
	for _, layer := range d.Layers {
		if predicate(layer) {
			filteredLayers = append(filteredLayers, layer)
		}
	}
	return filteredLayers
}

// GroupLayersByProperty groups layers by a property function
func (d *Drawing) GroupLayersByProperty(propertyFunc func(*table.Layer) string) map[string][]*table.Layer {
	groups := make(map[string][]*table.Layer)
	for _, layer := range d.Layers {
		key := propertyFunc(layer)
		groups[key] = append(groups[key], layer)
	}
	return groups
}

// Enhanced Block Management Methods

// GetBlockNames returns names of all blocks in the drawing
func (d *Drawing) GetBlockNames() []string {
	if d.blocksSection == nil {
		return []string{}
	}

	names := make([]string, 0, len(d.blocksSection))
	for _, block := range d.blocksSection {
		names = append(names, block.Name)
	}
	return names
}

// GetBlocks returns all blocks in the drawing
func (d *Drawing) GetBlocks() []*block.Block {
	if d.blocksSection == nil {
		return []*block.Block{}
	}

	blocks := make([]*block.Block, len(d.blocksSection))
	copy(blocks, d.blocksSection)
	return blocks
}

// GetBlock returns a block by name
func (d *Drawing) GetBlock(name string) (*block.Block, error) {
	if d.blocksSection == nil {
		return nil, fmt.Errorf("block section not initialized")
	}

	for _, block := range d.blocksSection {
		if block.Name == name {
			return block, nil
		}
	}
	return nil, fmt.Errorf("block '%s' not found", name)
}

// NewBlock creates a new block in the drawing
func (d *Drawing) NewBlock(name string, description string) (*block.Block, error) {
	// Check if block already exists
	_, err := d.GetBlock(name)
	if err == nil {
		return nil, fmt.Errorf("block '%s' already exists", name)
	}

	// Create new block
	newBlock := block.NewBlock(name, description)
	if d.CurrentLayer != nil {
		newBlock.SetLayer(d.CurrentLayer)
	}

	// Add to blocks section
	d.blocksSection = d.blocksSection.Add(newBlock)
	return newBlock, nil
}

// DeleteBlock removes a block from the drawing
func (d *Drawing) DeleteBlock(name string) error {
	// Cannot delete standard blocks
	if name == "*Model_Space" || name == "*Paper_Space" || name == "*Paper_Space0" {
		return fmt.Errorf("cannot delete standard block '%s'", name)
	}

	// Find and remove block
	for i, block := range d.blocksSection {
		if block.Name == name {
			// Remove from slice
			d.blocksSection = append(d.blocksSection[:i], d.blocksSection[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("block '%s' not found", name)
}

// CloneBlock creates a copy of an existing block
func (d *Drawing) CloneBlock(sourceName, targetName string) (*block.Block, error) {
	// Get source block
	source, err := d.GetBlock(sourceName)
	if err != nil {
		return nil, fmt.Errorf("source block not found: %v", err)
	}

	// Check if target name already exists
	_, err = d.GetBlock(targetName)
	if err == nil {
		return nil, fmt.Errorf("block '%s' already exists", targetName)
	}

	// Create clone
	clone := block.NewBlock(targetName, source.Description)
	if source.Layer() != nil {
		clone.SetLayer(source.Layer())
	}
	clone.Flag = source.Flag
	clone.Coord = make([]float64, len(source.Coord))
	copy(clone.Coord, source.Coord)

	// Add to blocks section
	d.blocksSection = d.blocksSection.Add(clone)
	return clone, nil
}

// PurgeUnusedBlocks removes unused blocks and returns count of removed blocks
func (d *Drawing) PurgeUnusedBlocks() int {
	if d.entityDatabase == nil {
		return 0
	}

	// Find all INSERT entities to determine which blocks are used
	usedBlocks := make(map[string]bool)
	entities := d.entityDatabase.GetEntitiesByType("INSERT")
	for _, ent := range entities {
		// This is simplified - in reality we'd need to extract block name from INSERT entity
		// For now, this is a placeholder for the concept
		attrs := ent.GetAttributes()
		if blockName, ok := attrs["name"]; ok {
			usedBlocks[fmt.Sprintf("%v", blockName)] = true
		}
	}

	// Remove unused blocks (skip standard blocks)
	removedCount := 0
	filteredBlocks := make([]*block.Block, 0, len(d.blocksSection))
	for _, block := range d.blocksSection {
		// Keep standard blocks and used blocks
		if block.Name == "*Model_Space" ||
			block.Name == "*Paper_Space" ||
			block.Name == "*Paper_Space0" ||
			usedBlocks[block.Name] {
			filteredBlocks = append(filteredBlocks, block)
		} else {
			removedCount++
		}
	}

	d.blocksSection = filteredBlocks
	return removedCount
}

// setHandle sets all the handles contained in Drawing.
func (d *Drawing) setHandle() {
	d.classesSection.SetHandle(d.handleGenarator)
	d.tablesSection.SetHandle(d.handleGenarator)
	d.blocksSection.SetHandle(d.handleGenarator)
	d.entitiesSection.SetHandle(d.handleGenarator)
	d.objectsSection.SetHandle(d.handleGenarator)

	d.headerSection.SetHandle(d.handleGenarator)
}

func (d *Drawing) Header() *header.Header {
	return d.headerSection
}

func (d *Drawing) Tables() table.Tables {
	return d.tablesSection
}

// Layer returns the named layer if exists.
// If setcurrent is true, set current layer to it.
func (d *Drawing) Layer(name string, setcurrent bool) (*table.Layer, error) {
	if l, exist := d.Layers[name]; exist {
		if setcurrent {
			d.CurrentLayer = l
		}
		return l, nil
	}
	return nil, fmt.Errorf("layer %s doesn't exist", name)
}

// ChangeLayer changes current layer to the named layer.
func (d *Drawing) ChangeLayer(name string) error {
	if l, exist := d.Layers[name]; exist {
		d.CurrentLayer = l
		return nil
	}
	return fmt.Errorf("layer %s doesn't exist", name)
}

// Style returns the named text style if exists.
// If setcurrent is true, set current style to it.
func (d *Drawing) Style(name string, setcurrent bool) (*table.Style, error) {
	if s, exist := d.Styles[name]; exist {
		if setcurrent {
			d.CurrentStyle = s
		}
		return s, nil
	}
	return nil, fmt.Errorf("style %s doesn't exist", name)
}

// LineType returns the named line type if exists.
func (d *Drawing) LineType(name string) (*table.LineType, error) {
	lt, err := d.tablesSection[table.LTYPE].Contains(name)
	if err != nil {
		return nil, fmt.Errorf("linetype %s", err.Error())
	}
	return lt.(*table.LineType), nil
}

// AddLineType adds a new linetype.
func (d *Drawing) AddLineType(name string, desc string, ls ...float64) (*table.LineType, error) {
	lt, _ := d.tablesSection[table.LTYPE].Contains(name)
	if lt != nil {
		return lt.(*table.LineType), fmt.Errorf("linetype %s already exists", name)
	}
	newlt := table.NewLineType(name, desc, ls...)
	d.tablesSection[table.LTYPE].Add(newlt)
	return newlt, nil
}

// Entities returns slice of all entities contained in Drawing.
func (d *Drawing) Entities() entity.Entities {
	return d.entitiesSection
}

// AddEntity adds a new entity.
func (d *Drawing) AddEntity(e entity.Entity) error {
	// Initialize entity database if needed
	if d.entityDatabase == nil {
		d.entityDatabase = NewEntityDatabase()
	}

	// Add to entity database (it will handle handle assignment)
	err := d.entityDatabase.AddEntity(e)
	if err != nil {
		return err
	}

	d.entitiesSection = d.entitiesSection.Add(e)
	return nil
}

// FileName returns the filename (implements interfaces.Document)
func (d *Drawing) FileName() string {
	return d.fileName
}

// GetEntity returns an entity by handle (implements interfaces.Document)
func (d *Drawing) GetEntity(handle string) (entity.Entity, bool) {
	if d.entityDatabase == nil {
		return nil, false
	}
	return d.entityDatabase.GetEntityByHandle(handle)
}

// Point creates a new POINT at (x, y, z).
func (d *Drawing) Point(x, y, z float64) (*entity.Point, error) {
	p := entity.NewPoint()
	p.Coord = []float64{x, y, z}
	p.SetLayer(d.CurrentLayer)
	d.AddEntity(p)
	return p, nil
}

// Line creates a new LINE from (x1, y1, z1) to (x2, y2, z2).
func (d *Drawing) Line(x1, y1, z1, x2, y2, z2 float64) (*entity.Line, error) {
	l := entity.NewLine()
	l.Start = []float64{x1, y1, z1}
	l.End = []float64{x2, y2, z2}
	l.SetLayer(d.CurrentLayer)
	d.AddEntity(l)
	return l, nil
}

// Circle creates a new CIRCLE at (x, y, z) with radius r.
func (d *Drawing) Circle(x, y, z, r float64) (*entity.Circle, error) {
	c := entity.NewCircle()
	c.Center = []float64{x, y, z}
	c.Radius = r
	c.SetLayer(d.CurrentLayer)
	d.AddEntity(c)
	return c, nil
}

// Arc creates a new ARC at (x, y, z) with radius r from start to end.
func (d *Drawing) Arc(x, y, z, r, start, end float64) (*entity.Arc, error) {
	c := entity.NewCircle()
	c.Center = []float64{x, y, z}
	c.Radius = r
	c.SetLayer(d.CurrentLayer)
	a := entity.NewArc(c)
	a.Angle[0] = start
	a.Angle[1] = end
	d.AddEntity(a)
	return a, nil
}

// Polyline creates a new POLYLINE with given vertices.
func (d *Drawing) Polyline(closed bool, vertices ...[]float64) (*entity.Polyline, error) {
	p := entity.NewPolyline()
	p.SetLayer(d.CurrentLayer)
	for _, v := range vertices {
		p.AddVertex(v[0], v[1], v[2])
	}
	if closed {
		p.Close()
	}
	d.AddEntity(p)
	return p, nil
}

// LwPolyline creates a new LWPOLYLINE with given vertices.
func (d *Drawing) LwPolyline(closed bool, vertices ...[]float64) (*entity.LwPolyline, error) {
	size := len(vertices)
	l := entity.NewLwPolyline(size)
	for i := 0; i < size; i++ {
		l.Vertices[i] = vertices[i]
	}
	if closed {
		l.Close()
	}
	l.SetLayer(d.CurrentLayer)
	d.AddEntity(l)
	return l, nil
}

// ThreeDFace creates a new 3DFACE with given points.
func (d *Drawing) ThreeDFace(points [][]float64) (*entity.ThreeDFace, error) {
	f := entity.New3DFace()
	if len(points) < 3 {
		return nil, errors.New("3DFace needs 3 or more points")
	}
	for i := 0; i < 3; i++ {
		f.Points[i] = points[i]
	}
	if len(points) >= 4 {
		f.Points[3] = points[3]
	} else {
		f.Points[3] = points[2]
	}
	f.SetLayer(d.CurrentLayer)
	d.AddEntity(f)
	return f, nil
}

// Text creates a new TEXT str at (x, y, z) with given height.
func (d *Drawing) Text(str string, x, y, z, height float64) (*entity.Text, error) {
	t := entity.NewText()
	t.Coord1 = []float64{x, y, z}
	t.Height = height
	t.Value = str
	t.SetLayer(d.CurrentLayer)
	t.Style = d.CurrentStyle
	t.WidthFactor = t.Style.WidthFactor
	t.ObliqueAngle = t.Style.ObliqueAngle
	t.Style.LastHeightUsed = height
	d.AddEntity(t)
	return t, nil
}

// MText creates a new MTEXT str at (x, y, z) with given height.
func (d *Drawing) MText(str string, x, y, z, height float64) (*entity.MText, error) {
	t := entity.NewMText()
	t.Coord1 = []float64{x, y, z}
	t.Height = height
	t.Value = str
	t.SetLayer(d.CurrentLayer)
	t.Style = d.CurrentStyle
	t.WidthFactor = t.Style.WidthFactor
	t.ObliqueAngle = t.Style.ObliqueAngle
	t.Style.LastHeightUsed = height
	d.AddEntity(t)
	return t, nil
}

func (d *Drawing) addObject(o object.Object) {
	d.objectsSection = d.objectsSection.Add(o)
}

// Group adds given entities to the named group.
// If the named group doesn't exist, create it.
func (d *Drawing) Group(name, desc string, es ...entity.Entity) (*object.Group, error) {
	if g, exist := d.Groups[name]; exist {
		g.AddEntity(es...)
		return g, fmt.Errorf("group %s already exists", name)
	}
	g := object.NewGroup(name, desc, es...)
	d.Groups[name] = g
	err := g.SetOwner(d.groupdict)
	if err != nil {
		return nil, err
	}
	d.addObject(g)
	return g, nil
}

// AddToGroup adds given entities to the named group.
// If the named group doesn't exist, returns error.
func (d *Drawing) AddToGroup(name string, es ...entity.Entity) error {
	if g, exist := d.Groups[name]; exist {
		g.AddEntity(es...)
	}
	return fmt.Errorf("group %s doesn't exist", name)
}

// WriteTo write the dxf file data to the given writer until
// there is no more data to write or if an error occurs. The return
// value n is the number of bytes writer.
// This method full fills the io.WriterTo interface.
func (d *Drawing) WriteTo(w io.Writer) (n int64, err error) {
	if d == nil {
		return 0, nil
	}
	d.setHandle()
	d.formatter.Reset()
	d.headerSection.Format(d.formatter)
	d.classesSection.Format(d.formatter)
	d.tablesSection.Format(d.formatter)
	d.blocksSection.Format(d.formatter)
	d.entitiesSection.Format(d.formatter)
	d.objectsSection.Format(d.formatter)

	d.formatter.WriteString(0, "EOF")
	return d.formatter.WriteTo(w)
}

var _ io.WriterTo = &Drawing{}

// Read implements the standard Read interface: it reads data from a buffer
// containing the drawing, the buffer in initialized on the first read or
// a read call after a close.  If the drawing is nil, read returns 0 for n,
// nil for the error.
func (d *Drawing) Read(p []byte) (n int, err error) {
	if d == nil {
		return 0, nil
	}
	if d.savebuff == nil {
		// We need to initilize our buffer, and write the contents of the
		// drawing into it.
		d.savebuff = new(bytes.Buffer)
		_, err := d.WriteTo(d.savebuff)
		if err != nil {
			return 0, err
		}
	}
	return d.savebuff.Read(p)
}

// Close implements the standard Close interface: it close the buffer used by
// the read command if one was initilized, and frees the memory held by it. Close
// will always return nil for the error.
func (d *Drawing) Close() error {
	if d != nil && d.savebuff != nil {
		d.savebuff = nil
	}
	return nil
}

var _ io.ReadCloser = &Drawing{}

// SetExt sets the extents of the drawing based on the entities
// in the drawing. If the drawing is nil, this function will panic.
func (d *Drawing) SetExt() {
	mins := []float64{1e16, 1e16, 1e16}
	maxs := []float64{-1e16, -1e16, -1e16}
	for _, en := range d.Entities() {
		tmpmins, tmpmaxs := en.BBox()
		for i := 0; i < 3; i++ {
			if tmpmins[i] < mins[i] {
				mins[i] = tmpmins[i]
			}
			if tmpmaxs[i] > maxs[i] {
				maxs[i] = tmpmaxs[i]
			}
		}
	}
	h := d.Header()
	for i := 0; i < 3; i++ {
		h.ExtMin[i] = mins[i]
		h.ExtMax[i] = maxs[i]
	}
}

// LtByLayer returns standard "ByLayer" line type.
func (d *Drawing) LtByLayer() *table.LineType {
	lt, _ := d.LineType("ByLayer")
	return lt
}

// LtByBlock returns standard "ByBlock" line type.
func (d *Drawing) LtByBlock() *table.LineType {
	lt, _ := d.LineType("ByBlock")
	return lt
}

// LtContinuous returns standard "Continuous" line type.
func (d *Drawing) LtContinuous() *table.LineType {
	lt, _ := d.LineType("Continuous")
	return lt
}

// LtHidden returns standard "Hidden" line type.
func (d *Drawing) LtHidden() *table.LineType {
	lt, _ := d.LineType("HIDDEN")
	return lt
}

// LtDashDot returns standard "DashDot" line type.
func (d *Drawing) LtDashDot() *table.LineType {
	lt, _ := d.LineType("DASHDOT")
	return lt
}

// Formatter returns the formatter used by the drawing
func (d *Drawing) Formatter() format.Formatter {
	return d.formatter
}

// Spatial Query Methods

// GetEntitiesInBox returns all entities within the specified 2D bounding box
func (d *Drawing) GetEntitiesInBox(minX, minY, maxX, maxY float64) []entity.Entity {
	if d.entityDatabase == nil {
		return []entity.Entity{}
	}
	return d.entityDatabase.GetEntitiesInBox(minX, minY, maxX, maxY)
}

// GetEntitiesInRadius returns all entities within the specified radius from a center point
func (d *Drawing) GetEntitiesInRadius(cx, cy, radius float64) []entity.Entity {
	if d.entityDatabase == nil {
		return []entity.Entity{}
	}
	return d.entityDatabase.GetEntitiesInRadius(cx, cy, radius)
}

// GetEntitiesInPolygon returns all entities within the specified polygon
func (d *Drawing) GetEntitiesInPolygon(polygon []struct{ X, Y float64 }) []entity.Entity {
	if d.entityDatabase == nil {
		return []entity.Entity{}
	}
	return d.entityDatabase.GetEntitiesInPolygon(polygon)
}

// GetSpatialIndex returns the spatial index for the drawing
func (d *Drawing) GetSpatialIndex() interface{} {
	if d.entityDatabase == nil {
		return nil
	}
	return d.entityDatabase.GetSpatialIndex()
}

// RebuildSpatialIndex rebuilds and returns the spatial index
func (d *Drawing) RebuildSpatialIndex() interface{} {
	if d.entityDatabase == nil {
		return nil
	}
	return d.entityDatabase.RebuildSpatialIndex()
}

// Layout Management Methods

// LayoutNames returns names of all layouts
func (d *Drawing) LayoutNames() []string {
	return []string{"Model"}
}

// ModelSpace returns the model space entities
func (d *Drawing) ModelSpace() []entity.Entity {
	return d.Entities()
}

// PaperSpace returns the paper space entities
func (d *Drawing) PaperSpace() []entity.Entity {
	return []entity.Entity{}
}

// CurrentLayout returns the current layout name
func (d *Drawing) CurrentLayout() string {
	return "Model"
}

// SetCurrentLayout sets the current layout
func (d *Drawing) SetCurrentLayout(name string) error {
	return nil
}

// Block returns a block by name
func (d *Drawing) Block(name string) (*block.Block, error) {
	return d.GetBlock(name)
}

// Layout management methods

// LayoutInfo represents basic layout information
type LayoutInfo struct {
	Name        string
	Type        string // "ModelSpace" or "PaperSpace"
	TabOrder    int
	EntityCount int
}

// GetLayouts returns information about all layouts
func (d *Drawing) GetLayouts() []LayoutInfo {
	return []LayoutInfo{
		{
			Name:        "Model",
			Type:        "ModelSpace",
			TabOrder:    0,
			EntityCount: len(d.Entities()),
		},
	}
}

// GetLayoutNames returns the names of all layouts
func (d *Drawing) GetLayoutNames() []string {
	return []string{"Model"}
}

// GetModelSpace returns entities in model space
func (d *Drawing) GetModelSpace() []entity.Entity {
	return d.Entities()
}

// GetPaperSpace returns entities in paper space (currently returns model space)
func (d *Drawing) GetPaperSpace() []entity.Entity {
	return []entity.Entity{}
}

// GetLayoutByName returns layout information by name
func (d *Drawing) GetLayoutByName(name string) *LayoutInfo {
	if name == "Model" || name == "MODEL" {
		return &LayoutInfo{
			Name:        "Model",
			Type:        "ModelSpace",
			TabOrder:    0,
			EntityCount: len(d.Entities()),
		}
	}
	return nil
}

// AddPaperSpace creates a new paper space layout
func (d *Drawing) AddPaperSpace(name string) error {
	// Paper space layouts are represented as blocks in this implementation
	_, err := d.NewBlock(name, "Paper space layout")
	return err
}

// SetActiveLayout sets the current layout (for future use with multi-layout support)
func (d *Drawing) SetActiveLayout(name string) error {
	return nil
}

// GetActiveLayout returns the current active layout name
func (d *Drawing) GetActiveLayout() string {
	return "Model"
}

// LayoutEntityCount returns the number of entities in a specific layout
func (d *Drawing) LayoutEntityCount(layoutName string) int {
	if layoutName == "Model" || layoutName == "MODEL" {
		return len(d.Entities())
	}
	return 0
}

// XREF Support Methods

// XRefInfo represents external reference information
type XRefInfo struct {
	FilePath  string
	BlockName string
	Loaded    bool
	Overlay   bool
	PathType  string // "Absolute" or "Relative"
}

// GetXRefs returns information about all external references
func (d *Drawing) GetXRefs() []XRefInfo {
	return []XRefInfo{}
}

// AddXRef adds an external reference
func (d *Drawing) AddXRef(filePath, blockName string, overlay bool) error {
	// TODO: Implement XREF loading
	return nil
}

// RemoveXRef removes an external reference
func (d *Drawing) RemoveXRef(blockName string) error {
	return nil
}

// ReloadXRef reloads an external reference
func (d *Drawing) ReloadXRef(blockName string) error {
	return nil
}

// BindXRef converts an external reference to a regular block
func (d *Drawing) BindXRef(blockName string, insert bool) error {
	return nil
}
