package lldxf

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// DXFInfo contains information about a DXF file
type DXFInfo struct {
	Version          string
	Release          string
	ACADVersion      string
	MaintVersion     string
	HasThumbnail     bool
	ThumbnailSize    int
	EntityCount      int
	BlockCount       int
	LayerCount       int
	StyleCount       int
	DimStyleCount    int
	LinetypeCount    int
	AppIDCount       int
	UCSCount         int
	ViewCount        int
	VPortCount       int
	HasBinaryData    bool
	HasProxyGraphics bool
	HasXData         bool
	HasAppData       bool
	Units            string
	InsUnits         int
	AngUnits         int
	InitialView      string
	LastSavedBy      string
	CreationDate     string
}

// Validator performs comprehensive DXF validation
type Validator struct {
	info     *DXFInfo
	errors   []ValidationError
	warnings []ValidationWarning
	options  *ValidationOptions
}

// ValidationOptions configures validation behavior
type ValidationOptions struct {
	StrictMode      bool
	CheckHandles    bool
	CheckReferences bool
	ValidateNames   bool
	CheckRanges     bool
	MaxEntities     int
	MaxHandles      int
}

// ValidationError represents a validation error
type ValidationError struct {
	Code     string
	Message  string
	Entity   string
	Handle   string
	Location string
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Code    string
	Message string
	Entity  string
	Handle  string
}

// NewValidator creates a new DXF validator
func NewValidator(options *ValidationOptions) *Validator {
	if options == nil {
		options = DefaultValidationOptions()
	}

	return &Validator{
		info:     &DXFInfo{},
		errors:   make([]ValidationError, 0),
		warnings: make([]ValidationWarning, 0),
		options:  options,
	}
}

// DefaultValidationOptions returns default validation options
func DefaultValidationOptions() *ValidationOptions {
	return &ValidationOptions{
		StrictMode:      false,
		CheckHandles:    true,
		CheckReferences: true,
		ValidateNames:   true,
		CheckRanges:     true,
		MaxEntities:     100000,
		MaxHandles:      1000000,
	}
}

// Validate validates a collection of tags
func (v *Validator) Validate(tags Tags) (*DXFInfo, error) {
	v.errors = v.errors[:0]
	v.warnings = v.warnings[:0]

	// Extract header information
	v.extractHeaderInfo(tags)

	// Validate structure
	v.validateStructure(tags)

	// Validate entities
	v.validateEntities(tags)

	// Validate tables
	v.validateTables(tags)

	if len(v.errors) > 0 && v.options.StrictMode {
		return v.info, fmt.Errorf("validation failed with %d errors", len(v.errors))
	}

	return v.info, nil
}

// extractHeaderInfo extracts information from HEADER section
func (v *Validator) extractHeaderInfo(tags Tags) {
	inHeader := false
	inSection := false

	for _, tag := range tags {
		code := int(tag.Code())
		value := fmt.Sprintf("%v", tag.Value())

		switch code {
		case 0:
			if value == "SECTION" {
				inSection = true
			} else if value == "ENDSEC" {
				inSection = false
				inHeader = false
			}
		case 2:
			if inSection && value == "HEADER" {
				inHeader = true
			}
		default:
			if inHeader {
				v.processHeaderTag(code, value)
			}
		}
	}
}

// processHeaderTag processes individual header tags
func (v *Validator) processHeaderTag(code int, value string) {
	switch code {
	case 1: // ACADVER
		v.info.Version = value
	case 3: // RELEASE
		v.info.Release = value
	case 6: // $DWGCODEPAGE (optional)
	case 9: // $ACADVER (alternative)
		v.info.ACADVersion = strings.TrimPrefix(value, "$")
	case 20: // $MEASUREMENT
		if units, err := strconv.Atoi(value); err == nil {
			switch units {
			case 0:
				v.info.Units = "English"
			case 1:
				v.info.Units = "Metric"
			}
		}
	case 70: // $MEASUREMENT (alternative)
	case 71: // $INSUNITS
		if units, err := strconv.Atoi(value); err == nil {
			v.info.InsUnits = units
		}
	case 72: // $ANGBASE
	case 73: // $ANGDIR
		if units, err := strconv.Atoi(value); err == nil {
			v.info.AngUnits = units
		}
	case 90: // $VERSIONMAINT
		if _, err := strconv.Atoi(value); err == nil {
			v.info.MaintVersion = value
		}
	}
}

// validateStructure validates overall DXF structure
func (v *Validator) validateStructure(tags Tags) {
	sections := make(map[string]bool)
	hasSection := false
	sectionCount := 0

	for i, tag := range tags {
		code := int(tag.Code())
		value := fmt.Sprintf("%v", tag.Value())

		switch code {
		case 0:
			if value == "SECTION" {
				hasSection = true
				sectionCount++
				// Check for section name
				if i+1 < len(tags) && tags[i+1].Code() == 2 {
					sectionName := fmt.Sprintf("%v", tags[i+1].Value())
					if sections[sectionName] {
						v.addError("DUPLICATE_SECTION", fmt.Sprintf("Duplicate section: %s", sectionName), "", fmt.Sprintf("tag %d", i))
					}
					sections[sectionName] = true
				} else {
					v.addError("MISSING_SECTION_NAME", "Missing section name after SECTION", "", fmt.Sprintf("tag %d", i))
				}
			} else if value == "ENDSEC" {
				sectionCount--
				if sectionCount < 0 {
					v.addError("UNEXPECTED_ENDSEC", "Unexpected ENDSEC without matching SECTION", "", fmt.Sprintf("tag %d", i))
				}
			}
		}
	}

	if !hasSection {
		v.addError("NO_SECTIONS", "No SECTION tag found", "", "")
	}

	if sectionCount != 0 {
		v.addError("UNBALANCED_SECTIONS", fmt.Sprintf("Unbalanced sections: %d unmatched", sectionCount), "", "")
	}
}

// validateEntities validates entity sections
func (v *Validator) validateEntities(tags Tags) {
	inEntities := false
	entityCount := 0
	entityHandles := make(map[string]bool)

	var currentEntity Tags
	var currentType string
	var currentHandle string

	for _, tag := range tags {
		code := int(tag.Code())
		value := fmt.Sprintf("%v", tag.Value())

		switch code {
		case 0:
			if value == "SECTION" {
				// Will be handled by section name
				continue
			} else if value == "ENTITIES" {
				inEntities = true
				continue
			} else if value == "ENDSEC" {
				inEntities = false
				continue
			}

			if inEntities {
				// Save previous entity
				if len(currentEntity) > 0 {
					v.validateSingleEntity(currentEntity, currentType, currentHandle)
					entityCount++
				}

				// Start new entity
				currentType = value
				currentHandle = ""
				currentEntity = Tags{tag}
			}
		default:
			if inEntities {
				currentEntity = append(currentEntity, tag)

				// Track handles
				if code == 5 || code == 105 {
					currentHandle = value
					if v.options.CheckHandles {
						if entityHandles[value] {
							v.addError("DUPLICATE_HANDLE", fmt.Sprintf("Duplicate handle: %s", value), currentType, value)
						}
						entityHandles[value] = true
					}
				}
			}
		}
	}

	// Validate last entity
	if len(currentEntity) > 0 {
		v.validateSingleEntity(currentEntity, currentType, currentHandle)
		entityCount++
	}

	v.info.EntityCount = entityCount

	if entityCount > v.options.MaxEntities {
		v.addWarning("MANY_ENTITIES", fmt.Sprintf("Large number of entities: %d", entityCount), "", "")
	}
}

// validateSingleEntity validates a single entity
func (v *Validator) validateSingleEntity(entity Tags, entityType, handle string) {
	// Check for required tags
	if entityType == "" {
		v.addError("MISSING_ENTITY_TYPE", "Entity type not specified", entityType, handle)
		return
	}

	// Validate layer name
	if layer, ok := entity.GetFirstValue(8); ok {
		layerName := fmt.Sprintf("%v", layer)
		if v.options.ValidateNames && !isValidLayerName(layerName) {
			v.addError("INVALID_LAYER_NAME", fmt.Sprintf("Invalid layer name: %s", layerName), entityType, handle)
		}
	}

	// Validate handle format
	if v.options.ValidateNames && handle != "" && !isValidHandle(handle) {
		v.addError("INVALID_HANDLE", fmt.Sprintf("Invalid handle format: %s", handle), entityType, handle)
	}

	// Entity-specific validation
	switch entityType {
	case "LINE":
		v.validateLine(entity, handle)
	case "CIRCLE":
		v.validateCircle(entity, handle)
	case "ARC":
		v.validateArc(entity, handle)
	case "LWPOLYLINE":
		v.validateLWPolyline(entity, handle)
	case "TEXT":
		v.validateText(entity, handle)
	}
}

// validateLine validates LINE entity
func (v *Validator) validateLine(entity Tags, handle string) {
	// Check start point
	if _, ok := entity.GetFirstValue(10); !ok {
		v.addError("MISSING_START_POINT", "LINE missing start point", "LINE", handle)
	}

	// Check end point
	if _, ok := entity.GetFirstValue(11); !ok {
		v.addError("MISSING_END_POINT", "LINE missing end point", "LINE", handle)
	}
}

// validateCircle validates CIRCLE entity
func (v *Validator) validateCircle(entity Tags, handle string) {
	// Check center point
	if _, ok := entity.GetFirstValue(10); !ok {
		v.addError("MISSING_CENTER", "CIRCLE missing center point", "CIRCLE", handle)
	}

	// Check radius
	if radius, ok := entity.GetFirstValue(40); ok {
		r := castToFloat(radius)
		if r <= 0 {
			v.addError("INVALID_RADIUS", "CIRCLE radius must be positive", "CIRCLE", handle)
		}
	} else {
		v.addError("MISSING_RADIUS", "CIRCLE missing radius", "CIRCLE", handle)
	}
}

// validateArc validates ARC entity
func (v *Validator) validateArc(entity Tags, handle string) {
	// Check center point
	if _, ok := entity.GetFirstValue(10); !ok {
		v.addError("MISSING_CENTER", "ARC missing center point", "ARC", handle)
	}

	// Check radius
	if radius, ok := entity.GetFirstValue(40); ok {
		r := castToFloat(radius)
		if r <= 0 {
			v.addError("INVALID_RADIUS", "ARC radius must be positive", "ARC", handle)
		}
	} else {
		v.addError("MISSING_RADIUS", "ARC missing radius", "ARC", handle)
	}

	// Check angles
	startAngle, hasStart := entity.GetFirstValue(50)
	endAngle, hasEnd := entity.GetFirstValue(51)

	if hasStart && hasEnd {
		start := castToFloat(startAngle)
		end := castToFloat(endAngle)

		// Check angle range
		if v.options.CheckRanges {
			if start < 0 || start > 360 {
				v.addWarning("ANGLE_OUT_OF_RANGE", fmt.Sprintf("Start angle %f out of range [0,360]", start), "ARC", handle)
			}
			if end < 0 || end > 360 {
				v.addWarning("ANGLE_OUT_OF_RANGE", fmt.Sprintf("End angle %f out of range [0,360]", end), "ARC", handle)
			}
		}
	}
}

// validateLWPolyline validates LWPOLYLINE entity
func (v *Validator) validateLWPolyline(entity Tags, handle string) {
	// Check vertices count
	if count, ok := entity.GetFirstValue(90); ok {
		vertexCount := castToInt32(count)
		if vertexCount < 2 {
			v.addError("INSUFFICIENT_VERTICES", "LWPOLYLINE must have at least 2 vertices", "LWPOLYLINE", handle)
		}

		// Count actual vertices
		vertexTags := entity.FindAll(10)
		if len(vertexTags) != int(vertexCount) {
			v.addError("VERTEX_COUNT_MISMATCH",
				fmt.Sprintf("Declared vertices: %d, actual vertices: %d", vertexCount, len(vertexTags)),
				"LWPOLYLINE", handle)
		}
	} else {
		v.addError("MISSING_VERTEX_COUNT", "LWPOLYLINE missing vertex count", "LWPOLYLINE", handle)
	}
}

// validateText validates TEXT entity
func (v *Validator) validateText(entity Tags, handle string) {
	// Check text content
	if _, ok := entity.GetFirstValue(1); !ok {
		v.addError("MISSING_TEXT", "TEXT missing text content", "TEXT", handle)
	}

	// Check insertion point
	if _, ok := entity.GetFirstValue(10); !ok {
		v.addError("MISSING_INSERTION_POINT", "TEXT missing insertion point", "TEXT", handle)
	}

	// Check height
	if height, ok := entity.GetFirstValue(40); ok {
		h := castToFloat(height)
		if h <= 0 {
			v.addError("INVALID_TEXT_HEIGHT", "TEXT height must be positive", "TEXT", handle)
		}
	} else {
		v.addError("MISSING_TEXT_HEIGHT", "TEXT missing height", "TEXT", handle)
	}
}

// validateTables validates TABLES section
func (v *Validator) validateTables(tags Tags) {
	inTablesSection := false

	tableCount := make(map[string]int)

	for _, tag := range tags {
		code := int(tag.Code())
		value := fmt.Sprintf("%v", tag.Value())

		switch code {
		case 0:
			if value == "SECTION" {
				// Will be handled by section name
				continue
			} else if value == "TABLES" {
				inTablesSection = true
				continue
			} else if value == "ENDSEC" {
				inTablesSection = false
				continue
			}
		case 2:
			if inTablesSection {
				tableName := fmt.Sprintf("%v", tag.Value())
				tableCount[tableName]++
			}
		}
	}

	// Update info
	v.info.LayerCount = tableCount["LAYER"]
	v.info.StyleCount = tableCount["STYLE"]
	v.info.DimStyleCount = tableCount["DIMSTYLE"]
	v.info.LinetypeCount = tableCount["LTYPE"]
	v.info.AppIDCount = tableCount["APPID"]
	v.info.UCSCount = tableCount["UCS"]
	v.info.ViewCount = tableCount["VIEW"]
	v.info.VPortCount = tableCount["VPORT"]
	v.info.BlockCount = tableCount["BLOCK"]
}

// addError adds a validation error
func (v *Validator) addError(code, message, entity, handle string) {
	v.errors = append(v.errors, ValidationError{
		Code:    code,
		Message: message,
		Entity:  entity,
		Handle:  handle,
	})
}

// addWarning adds a validation warning
func (v *Validator) addWarning(code, message, entity, handle string) {
	v.warnings = append(v.warnings, ValidationWarning{
		Code:    code,
		Message: message,
		Entity:  entity,
		Handle:  handle,
	})
}

// GetErrors returns all validation errors
func (v *Validator) GetErrors() []ValidationError {
	return v.errors
}

// GetWarnings returns all validation warnings
func (v *Validator) GetWarnings() []ValidationWarning {
	return v.warnings
}

// HasErrors returns true if there are validation errors
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// HasWarnings returns true if there are validation warnings
func (v *Validator) HasWarnings() bool {
	return len(v.warnings) > 0
}

// isValidLayerName checks if a layer name is valid
func isValidLayerName(name string) bool {
	if name == "" {
		return false
	}

	// Check for invalid characters
	invalidChars := []string{"/", "\\", "<", ">", "\"", ":", ";", "?", "*", "|", "="}
	for _, char := range invalidChars {
		if strings.Contains(name, char) {
			return false
		}
	}

	// Check for control characters
	for _, r := range name {
		if r < 32 {
			return false
		}
	}

	// Check length
	if len(name) > 255 {
		return false
	}

	return true
}

// isValidHandle checks if a handle is valid
func isValidHandle(handle string) bool {
	if handle == "" {
		return false
	}

	// Handle should be a hexadecimal string
	matched, _ := regexp.MatchString(`^[0-9A-Fa-f]+$`, handle)
	if !matched {
		return false
	}

	// Check length (typically 8-16 characters)
	if len(handle) < 1 || len(handle) > 16 {
		return false
	}

	return true
}
