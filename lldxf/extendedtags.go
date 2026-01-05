package lldxf

import (
	"fmt"
	"strings"
)

// ExtendedTags represents a collection of DXF tags with support for
// subclasses, AppData, XDATA, and embedded objects
type ExtendedTags struct {
	tags            Tags
	subclasses      map[string]Tags // Subclass name -> tags
	appData         map[string]Tags // App name -> tags
	xdata           map[string]Tags // Xdata app name -> tags
	embeddedObjects map[string]Tags // Embedded objects
	handle          string          // Entity handle
	entityType      string          // Entity type
}

// NewExtendedTags creates a new ExtendedTags from a regular Tags collection
func NewExtendedTags(tags Tags) *ExtendedTags {
	ext := &ExtendedTags{
		tags:            make(Tags, len(tags)),
		subclasses:      make(map[string]Tags),
		appData:         make(map[string]Tags),
		xdata:           make(map[string]Tags),
		embeddedObjects: make(map[string]Tags),
	}

	// Copy tags
	copy(ext.tags, tags)

	// Parse the structure
	ext.parseStructure()

	return ext
}

// parseStructure analyzes the tags and organizes them into subclasses,
// AppData sections, XDATA sections, etc.
func (ext *ExtendedTags) parseStructure() {
	var currentSubclass string
	var currentAppData string
	var currentXData string
	var currentEmbeddedObject string
	var inAppData bool
	var inXData bool
	var inEmbeddedObject bool

	for i, tag := range ext.tags {
		code := tag.Code()
		value := fmt.Sprintf("%v", tag.Value())

		// Store handle and entity type first
		if code == 5 || code == 105 {
			ext.handle = value
		}
		if code == 0 && i == 0 {
			ext.entityType = value
		}

		// Handle subclass markers (group code 100)
		if code == 100 {
			currentSubclass = value
			if _, exists := ext.subclasses[currentSubclass]; !exists {
				ext.subclasses[currentSubclass] = make(Tags, 0)
			}
			continue
		}

		// Handle AppData sections (group code 102)
		if code == 102 {
			if strings.HasPrefix(value, "{") {
				// Start of AppData section
				appName := strings.Trim(value, "{}")
				currentAppData = appName
				ext.appData[currentAppData] = make(Tags, 0)
				inAppData = true
			} else if value == "}" {
				// End of AppData section
				inAppData = false
				currentAppData = ""
			}
			continue
		}

		// Handle XDATA sections (group code 1001)
		if code == 1001 {
			currentXData = value
			ext.xdata[currentXData] = make(Tags, 0)
			inXData = true
			continue
		}

		// Handle XDATA continuation (codes 1002-1071)
		if code >= 1002 && code <= 1071 && inXData {
			ext.xdata[currentXData] = append(ext.xdata[currentXData], tag)
			continue
		}

		// Handle embedded object markers (group code 0 with specific entities)
		if code == 0 && (value == "ACDBPLACEHOLDER" || value == "ACDBDICTIONARYWDFLT") {
			currentEmbeddedObject = value
			ext.embeddedObjects[currentEmbeddedObject] = make(Tags, 0)
			inEmbeddedObject = true
			continue
		}

		// Add tags to appropriate collections
		if inAppData && currentAppData != "" {
			ext.appData[currentAppData] = append(ext.appData[currentAppData], tag)
		} else if inEmbeddedObject && currentEmbeddedObject != "" {
			ext.embeddedObjects[currentEmbeddedObject] = append(ext.embeddedObjects[currentEmbeddedObject], tag)
		} else if inXData && currentXData != "" {
			// XDATA handled above
		} else if currentSubclass != "" {
			// Add to subclass
			ext.subclasses[currentSubclass] = append(ext.subclasses[currentSubclass], tag)
		}

		// Reset XDATA state after XDATA section ends
		if code > 1001 && code <= 1071 {
			// Don't reset here, XDATA can have multiple tags
		}
	}
}

// Tags returns all tags as a regular Tags collection
func (ext *ExtendedTags) Tags() Tags {
	return ext.tags
}

// GetHandle returns the entity handle
func (ext *ExtendedTags) GetHandle() string {
	return ext.handle
}

// GetEntityType returns the entity type
func (ext *ExtendedTags) GetEntityType() string {
	return ext.entityType
}

// HasSubclass checks if a subclass exists
func (ext *ExtendedTags) HasSubclass(name string) bool {
	_, exists := ext.subclasses[name]
	return exists
}

// GetSubclass returns tags for a specific subclass
func (ext *ExtendedTags) GetSubclass(name string) Tags {
	if tags, exists := ext.subclasses[name]; exists {
		return tags
	}
	return make(Tags, 0)
}

// GetSubclassNames returns all subclass names
func (ext *ExtendedTags) GetSubclassNames() []string {
	names := make([]string, 0, len(ext.subclasses))
	for name := range ext.subclasses {
		names = append(names, name)
	}
	return names
}

// HasAppData checks if AppData exists for a specific application
func (ext *ExtendedTags) HasAppData(appName string) bool {
	_, exists := ext.appData[appName]
	return exists
}

// GetAppData returns AppData tags for a specific application
func (ext *ExtendedTags) GetAppData(appName string) Tags {
	if tags, exists := ext.appData[appName]; exists {
		return tags
	}
	return make(Tags, 0)
}

// GetAppDataNames returns all AppData application names
func (ext *ExtendedTags) GetAppDataNames() []string {
	names := make([]string, 0, len(ext.appData))
	for name := range ext.appData {
		names = append(names, name)
	}
	return names
}

// HasXData checks if XDATA exists for a specific application
func (ext *ExtendedTags) HasXData(appName string) bool {
	_, exists := ext.xdata[appName]
	return exists
}

// GetXData returns XDATA tags for a specific application
func (ext *ExtendedTags) GetXData(appName string) Tags {
	if tags, exists := ext.xdata[appName]; exists {
		return tags
	}
	return make(Tags, 0)
}

// GetXDataNames returns all XDATA application names
func (ext *ExtendedTags) GetXDataNames() []string {
	names := make([]string, 0, len(ext.xdata))
	for name := range ext.xdata {
		names = append(names, name)
	}
	return names
}

// HasEmbeddedObject checks if an embedded object exists
func (ext *ExtendedTags) HasEmbeddedObject(name string) bool {
	_, exists := ext.embeddedObjects[name]
	return exists
}

// GetEmbeddedObject returns embedded object tags
func (ext *ExtendedTags) GetEmbeddedObject(name string) Tags {
	if tags, exists := ext.embeddedObjects[name]; exists {
		return tags
	}
	return make(Tags, 0)
}

// GetEmbeddedObjectNames returns all embedded object names
func (ext *ExtendedTags) GetEmbeddedObjectNames() []string {
	names := make([]string, 0, len(ext.embeddedObjects))
	for name := range ext.embeddedObjects {
		names = append(names, name)
	}
	return names
}

// GetFirstValue returns the value of the first tag with the specified group code
// searched across all collections
func (ext *ExtendedTags) GetFirstValue(code GroupCode) (interface{}, bool) {
	// Search in main tags first
	if value, found := ext.tags.GetFirstValue(code); found {
		return value, true
	}

	// Search in subclasses
	for _, tags := range ext.subclasses {
		if value, found := tags.GetFirstValue(code); found {
			return value, true
		}
	}

	// Search in AppData
	for _, tags := range ext.appData {
		if value, found := tags.GetFirstValue(code); found {
			return value, true
		}
	}

	return nil, false
}

// FindAll returns all tags with the specified group code across all collections
func (ext *ExtendedTags) FindAll(code GroupCode) Tags {
	result := make(Tags, 0)

	// Search in main tags
	result = append(result, ext.tags.FindAll(code)...)

	// Search in subclasses
	for _, tags := range ext.subclasses {
		result = append(result, tags.FindAll(code)...)
	}

	// Search in AppData
	for _, tags := range ext.appData {
		result = append(result, tags.FindAll(code)...)
	}

	return result
}

// Clone creates a deep copy of the ExtendedTags
func (ext *ExtendedTags) Clone() *ExtendedTags {
	clone := &ExtendedTags{
		tags:            make(Tags, len(ext.tags)),
		subclasses:      make(map[string]Tags),
		appData:         make(map[string]Tags),
		xdata:           make(map[string]Tags),
		embeddedObjects: make(map[string]Tags),
		handle:          ext.handle,
		entityType:      ext.entityType,
	}

	// Copy main tags
	copy(clone.tags, ext.tags)

	// Copy subclasses
	for name, tags := range ext.subclasses {
		clone.subclasses[name] = make(Tags, len(tags))
		copy(clone.subclasses[name], tags)
	}

	// Copy AppData
	for name, tags := range ext.appData {
		clone.appData[name] = make(Tags, len(tags))
		copy(clone.appData[name], tags)
	}

	// Copy XDATA
	for name, tags := range ext.xdata {
		clone.xdata[name] = make(Tags, len(tags))
		copy(clone.xdata[name], tags)
	}

	// Copy embedded objects
	for name, tags := range ext.embeddedObjects {
		clone.embeddedObjects[name] = make(Tags, len(tags))
		copy(clone.embeddedObjects[name], tags)
	}

	return clone
}

// String returns the string representation of all tags
func (ext *ExtendedTags) String() string {
	var builder strings.Builder

	// Write main tags
	for _, tag := range ext.tags {
		builder.WriteString(tag.String())
	}

	return builder.String()
}

// DebugInfo returns detailed information about the ExtendedTags structure
func (ext *ExtendedTags) DebugInfo() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("ExtendedTags Debug Info:\n"))
	builder.WriteString(fmt.Sprintf("  Entity Type: %s\n", ext.entityType))
	builder.WriteString(fmt.Sprintf("  Handle: %s\n", ext.handle))
	builder.WriteString(fmt.Sprintf("  Total Tags: %d\n", len(ext.tags)))
	builder.WriteString(fmt.Sprintf("  Subclasses: %d\n", len(ext.subclasses)))
	builder.WriteString(fmt.Sprintf("  AppData Sections: %d\n", len(ext.appData)))
	builder.WriteString(fmt.Sprintf("  XDATA Sections: %d\n", len(ext.xdata)))
	builder.WriteString(fmt.Sprintf("  Embedded Objects: %d\n", len(ext.embeddedObjects)))

	if len(ext.subclasses) > 0 {
		builder.WriteString("  Subclass Names:\n")
		for name := range ext.subclasses {
			builder.WriteString(fmt.Sprintf("    %s (%d tags)\n", name, len(ext.subclasses[name])))
		}
	}

	if len(ext.appData) > 0 {
		builder.WriteString("  AppData Names:\n")
		for name := range ext.appData {
			builder.WriteString(fmt.Sprintf("    %s (%d tags)\n", name, len(ext.appData[name])))
		}
	}

	if len(ext.xdata) > 0 {
		builder.WriteString("  XDATA Names:\n")
		for name := range ext.xdata {
			builder.WriteString(fmt.Sprintf("    %s (%d tags)\n", name, len(ext.xdata[name])))
		}
	}

	if len(ext.embeddedObjects) > 0 {
		builder.WriteString("  Embedded Objects:\n")
		for name := range ext.embeddedObjects {
			builder.WriteString(fmt.Sprintf("    %s (%d tags)\n", name, len(ext.embeddedObjects[name])))
		}
	}

	return builder.String()
}
