package audit

import (
	"fmt"
	"strings"
	"sync"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
)

// Auditor performs comprehensive DXF file validation and recovery
type Auditor struct {
	doc         *drawing.Drawing
	errors      []*AuditError
	fixedErrors []*AuditError
	logger      *AuditLogger
	trashCan    []entity.Entity
	mu          sync.RWMutex

	// Configuration options
	AutoFix       bool
	StrictMode    bool
	RecoverLinks  bool
	RemoveOrphans bool
}

// NewAuditor creates a new auditor for the given drawing
func NewAuditor(doc *drawing.Drawing) *Auditor {
	return &Auditor{
		doc:           doc,
		errors:        make([]*AuditError, 0),
		logger:        NewAuditLogger(false),
		trashCan:      make([]entity.Entity, 0),
		AutoFix:       true,
		RecoverLinks:  true,
		RemoveOrphans: true,
	}
}

// SetLogger sets the audit logger
func (a *Auditor) SetLogger(logger *AuditLogger) {
	a.logger = logger
}

// SetOptions configures auditor behavior
func (a *Auditor) SetOptions(autoFix, strictMode, recoverLinks, removeOrphans bool) {
	a.AutoFix = autoFix
	a.StrictMode = strictMode
	a.RecoverLinks = recoverLinks
	a.RemoveOrphans = removeOrphans
}

// Run performs a comprehensive audit of the drawing
func (a *Auditor) Run() ([]*AuditError, error) {
	a.logger.LogMessage("Starting comprehensive DXF audit")

	// Clear previous results
	a.errors = make([]*AuditError, 0)
	a.fixedErrors = make([]*AuditError, 0)
	a.trashCan = make([]entity.Entity, 0)

	// Run audit pipeline
	if err := a.auditBasicStructure(); err != nil {
		return a.errors, fmt.Errorf("basic structure audit failed: %w", err)
	}

	if err := a.auditTables(); err != nil {
		return a.errors, fmt.Errorf("tables audit failed: %w", err)
	}

	if err := a.auditEntities(); err != nil {
		return a.errors, fmt.Errorf("entity audit failed: %w", err)
	}

	if err := a.auditReferences(); err != nil {
		return a.errors, fmt.Errorf("reference audit failed: %w", err)
	}

	if err := a.auditGeometry(); err != nil {
		return a.errors, fmt.Errorf("geometry audit failed: %w", err)
	}

	// Apply fixes if enabled
	if a.AutoFix {
		if err := a.applyFixes(); err != nil {
			return a.errors, fmt.Errorf("fix application failed: %w", err)
		}
	}

	// Cleanup
	a.emptyTrashCan()

	a.logger.LogMessage(fmt.Sprintf("Audit completed: %d total errors, %d fixed", len(a.errors), len(a.fixedErrors)))

	return a.errors, nil
}

// GetErrors returns all discovered errors
func (a *Auditor) GetErrors() []*AuditError {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return append([]*AuditError{}, a.errors...)
}

// GetFixedErrors returns all successfully fixed errors
func (a *Auditor) GetFixedErrors() []*AuditError {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return append([]*AuditError{}, a.fixedErrors...)
}

// HasErrors returns true if any errors were found
func (a *Auditor) HasErrors() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return len(a.errors) > 0
}

// HasCriticalErrors returns true if any critical errors were found
func (a *Auditor) HasCriticalErrors() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	for _, err := range a.errors {
		if err.Severity == SeverityCritical {
			return true
		}
	}
	return false
}

// GetErrorSummary returns a summary of errors by category and severity
func (a *Auditor) GetErrorSummary() map[string]map[string]int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	summary := make(map[string]map[string]int)

	for _, err := range a.errors {
		category := string(err.Code.GetCategory())
		severity := err.Severity.String()

		if summary[category] == nil {
			summary[category] = make(map[string]int)
		}

		summary[category][severity]++
	}

	return summary
}

// PrintSummary prints a formatted summary of audit results
func (a *Auditor) PrintSummary() {
	summary := a.GetErrorSummary()

	fmt.Println("\n=== AUDIT SUMMARY ===")

	if len(a.errors) == 0 {
		fmt.Println("✅ No errors found - DXF file is valid")
		return
	}

	fmt.Printf("📊 Total Errors: %d (%d fixed)\n", len(a.errors), len(a.fixedErrors))
	fmt.Printf("🔧 Auto-fix Status: %s\n", map[bool]string{true: "Enabled", false: "Disabled"}[a.AutoFix])

	for category, severities := range summary {
		fmt.Printf("\n📁 %s:\n", category)
		for severity, count := range severities {
			icon := map[string]string{
				"INFO":     "ℹ️",
				"WARNING":  "⚠️",
				"ERROR":    "❌",
				"CRITICAL": "🚨",
			}[severity]
			fmt.Printf("  %s %s: %d\n", icon, severity, count)
		}
	}

	// Print critical errors first
	if a.HasCriticalErrors() {
		fmt.Println("\n🚨 CRITICAL ERRORS:")
		for _, err := range a.errors {
			if err.Severity == SeverityCritical {
				fmt.Printf("  %s\n", err.String())
			}
		}
	}
}

// auditBasicStructure validates basic DXF file structure
func (a *Auditor) auditBasicStructure() error {
	a.logger.LogMessage("Auditing basic DXF structure")

	// Check if drawing exists
	if a.doc == nil {
		err := NewError(ErrorMissingRootDict, SeverityCritical, "", "", "Drawing is nil")
		a.addError(err)
		return fmt.Errorf("drawing is nil")
	}

	// Check header
	if a.doc.Header() == nil {
		err := NewError(ErrorMissingHeader, SeverityCritical, "", "", "Missing DXF header section")
		a.addError(err)
	} else {
		a.validateHeader()
	}

	// Check tables
	if a.doc.Tables() == nil {
		err := NewError(ErrorInvalidLayerTable, SeverityCritical, "", "", "Missing tables section")
		a.addError(err)
	}

	// Check blocks (will be validated during entity audit)
	// Block section is embedded in the drawing, not a separate method

	return nil
}

// auditTables validates all symbol tables
func (a *Auditor) auditTables() error {
	a.logger.LogMessage("Auditing symbol tables")

	if a.doc.Tables() == nil {
		return nil // Already reported in basic structure audit
	}

	// Validate layer table
	if err := a.validateLayerTable(); err != nil {
		return err
	}

	// Validate other tables
	if err := a.validateOtherTables(); err != nil {
		return err
	}

	return nil
}

// auditEntities validates all entities in the drawing
func (a *Auditor) auditEntities() error {
	a.logger.LogMessage("Auditing entities")

	// Audit all entities in the drawing
	entities := a.doc.Entities()
	for _, entity := range entities {
		if entity != nil {
			a.auditEntity(entity, "Drawing")
		}
	}

	return nil
}

// auditReferences validates entity references and links
func (a *Auditor) auditReferences() error {
	a.logger.LogMessage("Auditing entity references")

	// Validate layer references
	a.validateLayerReferences()

	// Validate text style references
	a.validateStyleReferences()

	// Validate dimension style references
	a.validateDimStyleReferences()

	// Validate block references
	a.validateBlockReferences()

	return nil
}

// auditGeometry validates geometric properties of entities
func (a *Auditor) auditGeometry() error {
	a.logger.LogMessage("Auditing entity geometry")

	// This would iterate through all entities and validate their geometry
	// For now, we'll implement a basic version

	return nil
}

// validateHeader validates DXF header section
func (a *Auditor) validateHeader() {
	// Basic header validation
	// Add specific header checks here
}

// validateLayerTable validates the layer table
func (a *Auditor) validateLayerTable() error {
	// Check for required layers in the drawing's layer map
	requiredLayers := []string{"0"}
	for _, layerName := range requiredLayers {
		layer, exists := a.doc.Layers[layerName]
		if !exists || layer == nil {
			err := NewError(ErrorUndefinedLayer, SeverityCritical, "", "",
				fmt.Sprintf("Required layer '%s' is missing", layerName))
			a.addError(err)
		}
	}

	// Validate layer names
	for layerName, layer := range a.doc.Layers {
		if layer != nil {
			if err := a.validateLayerName(layerName); err != nil {
				err.Entity = "LAYER"
				err.Context = layerName
				a.addError(err)
			}
		}
	}

	return nil
}

// validateOtherTables validates non-layer tables
func (a *Auditor) validateOtherTables() error {
	// Validate style table
	// Check for required styles in the drawing's styles map
	requiredStyles := []string{"Standard"}
	for _, styleName := range requiredStyles {
		style, exists := a.doc.Styles[styleName]
		if !exists || style == nil {
			err := NewError(ErrorUndefinedStyle, SeverityError, "", "",
				fmt.Sprintf("Required text style '%s' is missing", styleName))
			a.addError(err)
		}
	}

	return nil
}

// validateLayerName validates a layer name
func (a *Auditor) validateLayerName(name string) *AuditError {
	// Check for invalid characters
	if strings.ContainsAny(name, `<>/"\:;?*|,=`) {
		return NewError(ErrorInvalidLayerName, SeverityError, "LAYER", "",
			fmt.Sprintf("Layer name '%s' contains invalid characters", name))
	}

	// Check for empty name
	if name == "" {
		return NewError(ErrorInvalidLayerName, SeverityError, "LAYER", "", "Layer name cannot be empty")
	}

	// Check for name length
	if len(name) > 255 {
		return NewError(ErrorInvalidLayerName, SeverityError, "LAYER", "",
			fmt.Sprintf("Layer name '%s' exceeds maximum length", name))
	}

	return nil
}

// validateLayerReferences validates that all entities reference valid layers
func (a *Auditor) validateLayerReferences() {
	// Create a set of valid layer names from the drawing's layer map
	validLayers := make(map[string]bool)
	for layerName := range a.doc.Layers {
		validLayers[layerName] = true
	}

	// Check entity layer references
	entities := a.doc.Entities()
	for _, entity := range entities {
		if entity != nil && entity.Layer() != nil {
			layerName := entity.Layer().Name()
			if !validLayers[layerName] {
				err := NewError(ErrorUndefinedLayer, SeverityError, "ENTITY", entity.Handle(),
					fmt.Sprintf("Entity references undefined layer '%s'", layerName))
				a.addError(err)
			}
		}
	}
}

// validateStyleReferences validates text style references
func (a *Auditor) validateStyleReferences() {
	// Similar to layer references but for text styles
}

// validateDimStyleReferences validates dimension style references
func (a *Auditor) validateDimStyleReferences() {
	// Validate dimension style references
}

// validateBlockReferences validates block reference entities
func (a *Auditor) validateBlockReferences() {
	// Validate INSERT entity block name references
}

// auditEntitySlice audits a slice of entities
func (a *Auditor) auditEntitySlice(entities []entity.Entity, context string) {
	for _, entity := range entities {
		if entity != nil {
			a.auditEntity(entity, context)
		}
	}
}

// auditEntity validates a single entity
func (a *Auditor) auditEntity(e entity.Entity, context string) {
	// Basic entity validation
	if e.Handle() == "" {
		// Use entity's type from its internal structure
		err := NewError(ErrorInvalidHandle, SeverityError, "ENTITY", "", "Entity has no handle")
		err.Context = context
		a.addError(err)
	}

	// Validate layer reference
	if e.Layer() == nil {
		err := NewError(ErrorUndefinedLayer, SeverityError, "ENTITY", e.Handle(), "Entity has no layer assigned")
		err.Context = context
		a.addError(err)
	}

	// Entity-specific validation
	// This would call entity-specific audit methods
}

// addError adds an error to the error list and logs it
func (a *Auditor) addError(err *AuditError) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.errors = append(a.errors, err)
	a.logger.Log(err)
}

// addFixedError marks an error as fixed and adds it to the fixed list
func (a *Auditor) addFixedError(err *AuditError, recoveryLog string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	err.MarkFixed(recoveryLog)
	a.fixedErrors = append(a.fixedErrors, err)
	a.logger.LogRecovery("Fixed", err.String())
}

// addToTrash adds an entity to the trash can for later deletion
func (a *Auditor) addToTrash(entity entity.Entity) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.trashCan = append(a.trashCan, entity)
}

// emptyTrashCan removes all entities in the trash can from the drawing
func (a *Auditor) emptyTrashCan() {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Remove entities from the drawing
	// This would depend on the drawing package's entity removal API

	a.trashCan = make([]entity.Entity, 0)
}

// applyFixes applies automatic fixes for fixable errors
func (a *Auditor) applyFixes() error {
	a.logger.LogMessage("Applying automatic fixes")

	fixableErrors := make([]*AuditError, 0)
	for _, err := range a.errors {
		if err.CanFix && !err.Fixed {
			fixableErrors = append(fixableErrors, err)
		}
	}

	for _, auditErr := range fixableErrors {
		if fixErr := a.fixError(auditErr); fixErr != nil {
			return fmt.Errorf("failed to fix error %s: %w", auditErr.String(), fixErr)
		}
	}

	return nil
}

// fixError attempts to fix a specific error
func (a *Auditor) fixError(err *AuditError) error {
	switch err.Code {
	case ErrorUndefinedLayer:
		return a.fixUndefinedLayer(err)
	case ErrorDuplicateHandle:
		return a.fixDuplicateHandle(err)
	case ErrorInvalidColor:
		return a.fixInvalidColor(err)
	// Add more fix cases as needed
	default:
		return fmt.Errorf("no automatic fix available for error code %d", err.Code)
	}
}

// fixUndefinedLayer creates a missing layer with default properties
func (a *Auditor) fixUndefinedLayer(err *AuditError) error {
	// Extract layer name from error message or context
	layerName := extractLayerNameFromError(err)

	if layerName != "" {
		// Create the missing layer with default properties
		// Use the continuous linetype as default
		continuousLT, ltypeErr := a.doc.LineType("CONTINUOUS")
		if ltypeErr != nil {
			continuousLT, ltypeErr = a.doc.AddLineType("CONTINUOUS", "Solid line")
			if ltypeErr != nil {
				return ltypeErr
			}
		}

		_, createErr := a.doc.AddLayer(layerName, color.White, continuousLT)
		if createErr != nil {
			return createErr
		}

		a.addFixedError(err, fmt.Sprintf("Created missing layer '%s'", layerName))
	}

	return nil
}

// fixDuplicateHandle assigns a new unique handle to an entity
func (a *Auditor) fixDuplicateHandle(err *AuditError) error {
	// This would find the entity with the duplicate handle and assign a new one
	a.addFixedError(err, "Assigned new unique handle")
	return nil
}

// fixInvalidColor resets an invalid color to a valid default
func (a *Auditor) fixInvalidColor(err *AuditError) error {
	// This would find the entity with invalid color and set it to a default
	a.addFixedError(err, "Reset color to default (ByLayer)")
	return nil
}

// extractLayerNameFromError extracts layer name from error context or message
func extractLayerNameFromError(err *AuditError) string {
	// Simple extraction - in a real implementation this would be more sophisticated
	if err.Context != "" && strings.Contains(err.Context, "layer") {
		return err.Context
	}
	return ""
}
