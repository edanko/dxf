package lldxf

import (
	"fmt"
	"strings"
)

// Loader provides high-level DXF structure loading and entity binding
type Loader struct {
	tags     Tags
	sections map[string]Tags
	entities map[string]Tags
	tables   map[string]map[string]Tags
	blocks   map[string]Tags
	objects  map[string]Tags
	options  *LoaderOptions
}

// LoaderOptions configures loader behavior
type LoaderOptions struct {
	LoadAllSections bool
	LoadEntities    bool
	LoadTables      bool
	LoadBlocks      bool
	LoadObjects     bool
	SkipComments    bool
	PreserveUnknown bool
	BindEntities    bool
	StrictMode      bool
}

// NewLoader creates a new DXF loader
func NewLoader(options *LoaderOptions) *Loader {
	if options == nil {
		options = DefaultLoaderOptions()
	}

	return &Loader{
		sections: make(map[string]Tags),
		entities: make(map[string]Tags),
		tables:   make(map[string]map[string]Tags),
		blocks:   make(map[string]Tags),
		objects:  make(map[string]Tags),
		options:  options,
	}
}

// DefaultLoaderOptions returns default loader options
func DefaultLoaderOptions() *LoaderOptions {
	return &LoaderOptions{
		LoadAllSections: true,
		LoadEntities:    true,
		LoadTables:      true,
		LoadBlocks:      true,
		LoadObjects:     true,
		SkipComments:    true,
		PreserveUnknown: true,
		BindEntities:    true,
		StrictMode:      false,
	}
}

// Load processes the DXF tags and builds the structure
func (l *Loader) Load(tags Tags) error {
	l.tags = tags

	// Group tags by sections
	if err := l.groupSections(); err != nil {
		return fmt.Errorf("failed to group sections: %w", err)
	}

	// Process each section
	if l.options.LoadAllSections {
		if err := l.processSections(); err != nil {
			return fmt.Errorf("failed to process sections: %w", err)
		}
	}

	return nil
}

// groupSections groups tags by DXF sections
func (l *Loader) groupSections() error {
	var currentSection string
	var inSection bool
	var sectionTags Tags

	for _, tag := range l.tags {
		code := int(tag.Code())
		value := fmt.Sprintf("%v", tag.Value())

		// Skip comments if requested
		if l.options.SkipComments && code == 999 {
			continue
		}

		switch {
		case code == 0 && value == "SECTION":
			inSection = true
			sectionTags = NewTags()

		case code == 0 && value == "ENDSEC":
			if !inSection {
				return fmt.Errorf("ENDSEC without matching SECTION")
			}

			if currentSection != "" {
				l.sections[currentSection] = sectionTags
			}

			inSection = false
			currentSection = ""

		case code == 2 && inSection:
			currentSection = value

		case inSection:
			sectionTags = append(sectionTags, tag)
		}
	}

	if inSection {
		return fmt.Errorf("unclosed SECTION at end of file")
	}

	return nil
}

// processSections processes each section
func (l *Loader) processSections() error {
	// Process ENTITIES section
	if l.options.LoadEntities {
		if err := l.processEntities(); err != nil {
			return fmt.Errorf("failed to process entities: %w", err)
		}
	}

	// Process TABLES section
	if l.options.LoadTables {
		if err := l.processTables(); err != nil {
			return fmt.Errorf("failed to process tables: %w", err)
		}
	}

	// Process BLOCKS section
	if l.options.LoadBlocks {
		if err := l.processBlocks(); err != nil {
			return fmt.Errorf("failed to process blocks: %w", err)
		}
	}

	// Process OBJECTS section
	if l.options.LoadObjects {
		if err := l.processObjects(); err != nil {
			return fmt.Errorf("failed to process objects: %w", err)
		}
	}

	return nil
}

// processEntities processes the ENTITIES section
func (l *Loader) processEntities() error {
	entitiesTags, ok := l.sections[SectionEntities]
	if !ok {
		return nil // No entities section is valid
	}

	return l.extractEntities(entitiesTags, l.entities)
}

// processTables processes the TABLES section
func (l *Loader) processTables() error {
	tablesTags, ok := l.sections[SectionTables]
	if !ok {
		return nil // No tables section is valid
	}

	var currentTable string
	var inTable bool
	var tableEntries map[string]Tags

	for _, tag := range tablesTags {
		code := int(tag.Code())
		value := fmt.Sprintf("%v", tag.Value())

		switch {
		case code == 0 && strings.HasSuffix(value, "TABLE"):
			inTable = true
			currentTable = strings.TrimSuffix(value, "TABLE")
			tableEntries = make(map[string]Tags)

		case code == 0 && value == "ENDTAB":
			if !inTable {
				return fmt.Errorf("ENDTAB without matching TABLE")
			}

			if currentTable != "" {
				l.tables[currentTable] = tableEntries
			}

			inTable = false
			currentTable = ""

		case code == 0 && inTable:
			// New table entry
			entryName := value
			if tableEntries == nil {
				tableEntries = make(map[string]Tags)
			}
			tableEntries[entryName] = NewTags()

		case inTable && currentTable != "":
			// Add tag to current table entry
			// Find the last entry created
			var lastEntry string
			for name := range tableEntries {
				if len(tableEntries[name]) > 0 {
					lastEntry = name
				}
			}
			if lastEntry != "" {
				tableEntries[lastEntry] = append(tableEntries[lastEntry], tag)
			}
		}
	}

	return nil
}

// processBlocks processes the BLOCKS section
func (l *Loader) processBlocks() error {
	blocksTags, ok := l.sections[SectionBlocks]
	if !ok {
		return nil // No blocks section is valid
	}

	return l.extractEntities(blocksTags, l.blocks)
}

// processObjects processes the OBJECTS section
func (l *Loader) processObjects() error {
	objectsTags, ok := l.sections[SectionObjects]
	if !ok {
		return nil // No objects section is valid
	}

	return l.extractEntities(objectsTags, l.objects)
}

// extractEntities extracts entities from a section
func (l *Loader) extractEntities(sectionTags Tags, target map[string]Tags) error {
	var currentEntity Tags
	var currentHandle string

	for _, tag := range sectionTags {
		code := int(tag.Code())
		value := fmt.Sprintf("%v", tag.Value())

		if code == 0 {
			// Save previous entity
			if len(currentEntity) > 0 {
				if currentHandle != "" {
					target[currentHandle] = currentEntity
				} else {
					// Generate temporary handle if missing
					tempHandle := fmt.Sprintf("TEMP_%d", len(target))
					target[tempHandle] = currentEntity
				}
			}

			// Start new entity
			currentEntity = Tags{tag}
			currentHandle = ""
		} else {
			currentEntity = append(currentEntity, tag)

			// Track handles
			if code == 5 || code == 105 {
				currentHandle = value
			}
		}
	}

	// Add last entity
	if len(currentEntity) > 0 {
		if currentHandle != "" {
			target[currentHandle] = currentEntity
		} else {
			tempHandle := fmt.Sprintf("TEMP_%d", len(target))
			target[tempHandle] = currentEntity
		}
	}

	return nil
}

// GetSections returns all sections
func (l *Loader) GetSections() map[string]Tags {
	return l.sections
}

// GetSection returns a specific section
func (l *Loader) GetSection(name string) Tags {
	return l.sections[name]
}

// GetEntities returns all entities
func (l *Loader) GetEntities() map[string]Tags {
	return l.entities
}

// GetEntity returns a specific entity by handle
func (l *Loader) GetEntity(handle string) Tags {
	return l.entities[handle]
}

// GetTables returns all tables
func (l *Loader) GetTables() map[string]map[string]Tags {
	return l.tables
}

// GetTable returns a specific table
func (l *Loader) GetTable(name string) map[string]Tags {
	return l.tables[name]
}

// GetTableEntry returns a specific table entry
func (l *Loader) GetTableEntry(table, name string) Tags {
	if tableEntries, ok := l.tables[table]; ok {
		return tableEntries[name]
	}
	return nil
}

// GetBlocks returns all blocks
func (l *Loader) GetBlocks() map[string]Tags {
	return l.blocks
}

// GetBlock returns a specific block
func (l *Loader) GetBlock(handle string) Tags {
	return l.blocks[handle]
}

// GetObjects returns all objects
func (l *Loader) GetObjects() map[string]Tags {
	return l.objects
}

// GetObject returns a specific object
func (l *Loader) GetObject(handle string) Tags {
	return l.objects[handle]
}

// FindEntitiesByType finds entities of a specific type
func (l *Loader) FindEntitiesByType(entityType string) []Tags {
	var result []Tags

	for _, entity := range l.entities {
		if etype, ok := entity.GetFirstValue(0); ok {
			if fmt.Sprintf("%v", etype) == entityType {
				result = append(result, entity)
			}
		}
	}

	return result
}

// FindEntitiesByLayer finds entities on a specific layer
func (l *Loader) FindEntitiesByLayer(layer string) []Tags {
	var result []Tags

	for _, entity := range l.entities {
		if elayer, ok := entity.GetFirstValue(8); ok {
			if fmt.Sprintf("%v", elayer) == layer {
				result = append(result, entity)
			}
		}
	}

	return result
}

// GetEntityCount returns the number of entities
func (l *Loader) GetEntityCount() int {
	return len(l.entities)
}

// GetTableCount returns the number of tables
func (l *Loader) GetTableCount() int {
	return len(l.tables)
}

// GetBlockCount returns the number of blocks
func (l *Loader) GetBlockCount() int {
	return len(l.blocks)
}

// GetObjectCount returns the number of objects
func (l *Loader) GetObjectCount() int {
	return len(l.objects)
}

// Info returns loader information
func (l *Loader) Info() map[string]interface{} {
	return map[string]interface{}{
		"entity_count": l.GetEntityCount(),
		"table_count":  l.GetTableCount(),
		"block_count":  l.GetBlockCount(),
		"object_count": l.GetObjectCount(),
		"sections":     l.GetSectionNames(),
		"tables":       l.GetTableNames(),
	}
}

// GetSectionNames returns all section names
func (l *Loader) GetSectionNames() []string {
	names := make([]string, 0, len(l.sections))
	for name := range l.sections {
		names = append(names, name)
	}
	return names
}

// GetTableNames returns all table names
func (l *Loader) GetTableNames() []string {
	names := make([]string, 0, len(l.tables))
	for name := range l.tables {
		names = append(names, name)
	}
	return names
}

// Validate checks the loaded structure for consistency
func (l *Loader) Validate() error {
	// Check for required sections
	requiredSections := []string{SectionHeader, SectionTables, SectionEntities}
	for _, section := range requiredSections {
		if _, ok := l.sections[section]; !ok {
			if l.options.StrictMode {
				return fmt.Errorf("missing required section: %s", section)
			}
		}
	}

	// Validate entities
	for handle, entity := range l.entities {
		if len(entity) == 0 {
			return fmt.Errorf("empty entity with handle: %s", handle)
		}

		// Check for entity type
		if _, ok := entity.GetFirstValue(0); !ok {
			return fmt.Errorf("entity %s missing type", handle)
		}
	}

	// Validate tables
	for tableType, tableEntries := range l.tables {
		for entryName, entry := range tableEntries {
			if len(entry) == 0 {
				return fmt.Errorf("empty table entry %s in table %s", entryName, tableType)
			}
		}
	}

	return nil
}

// LoadBinary loads DXF from binary data
func LoadBinary(data []byte, options *LoaderOptions) (*Loader, error) {
	if options == nil {
		options = DefaultLoaderOptions()
	}

	// Create binary loader
	binaryLoader, err := NewBinaryLoader(data)
	if err != nil {
		return nil, fmt.Errorf("failed to create binary loader: %w", err)
	}

	// Load all binary tags
	binaryTags, err := binaryLoader.LoadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to load binary tags: %w", err)
	}

	// Convert binary tags to DXF tags
	dxfTags := make(Tags, len(binaryTags))
	for i, binaryTag := range binaryTags {
		dxfTags[i] = NewDXFTag(GroupCode(binaryTag.Code), binaryTag.Value)
	}

	// Create and populate loader
	loader := NewLoader(options)
	if err := loader.Load(dxfTags); err != nil {
		return nil, fmt.Errorf("failed to load converted tags: %w", err)
	}

	return loader, nil
}

// Reload reloads from the same tags with new options
func (l *Loader) Reload(options *LoaderOptions) error {
	if options != nil {
		l.options = options
	}

	// Reset all collections
	l.sections = make(map[string]Tags)
	l.entities = make(map[string]Tags)
	l.tables = make(map[string]map[string]Tags)
	l.blocks = make(map[string]Tags)
	l.objects = make(map[string]Tags)

	return l.Load(l.tags)
}
