package audit

import (
	"fmt"
	"log"
	"strings"
)

// ErrorCode represents different types of audit errors
type ErrorCode int

const (
	// Structure Errors (1-18)
	ErrorMissingSection ErrorCode = iota + 1
	ErrorInvalidHandle
	ErrorDuplicateHandle
	ErrorBrokenStructure
	ErrorMissingHeader
	ErrorInvalidSection
	ErrorCorruptedTags
	ErrorInvalidGroupCode
	ErrorMissingEndSec
	ErrorInvalidEntityStructure
	ErrorCorruptedData
	ErrorInvalidDatabase
	ErrorMissingRootDict
	ErrorInvalidLayerTable
	ErrorInvalidBlockTable
	ErrorInvalidStyleTable
	ErrorInvalidDimStyleTable
	ErrorInvalidAppIDTable
	ErrorInvalidUCSTable
	ErrorInvalidViewTable
	ErrorInvalidViewportTable

	// Reference Errors (100-119)
	ErrorUndefinedLayer
	ErrorUndefinedLinetype
	ErrorUndefinedStyle
	ErrorUndefinedBlock
	ErrorUndefinedText
	ErrorUndefinedDimStyle
	ErrorUndefinedAppID
	ErrorUndefinedUCS
	ErrorUndefinedView
	ErrorUndefinedViewport
	ErrorBrokenLink
	ErrorInvalidReference
	ErrorCircularReference
	ErrorOrphanedEntity
	ErrorMissingReference
	ErrorInvalidBlockReferenceCycle

	// Property Errors (201-207)
	ErrorInvalidColor
	ErrorInvalidLineweight
	ErrorInvalidHandleFormat
	ErrorInvalidLayerName
	ErrorInvalidLinetypeName
	ErrorInvalidStyleName
	ErrorInvalidDimStyleName

	// Geometry Errors (210-228)
	ErrorInvalidPoint
	ErrorInvalidVector
	ErrorInvalidNormal
	ErrorInvalidSpline
	ErrorInvalidMesh
	ErrorInvalidHatch
	ErrorInvalidDimension
	ErrorInvalidInsertPoint
	ErrorInvalidScale
	ErrorInvalidRotation
	ErrorInvalidExtrusion
	ErrorInvalidBoundaryPath
	ErrorInvalidControlPoints
	ErrorInvalidKnotVector
	ErrorInvalidWeights
	ErrorInvalidTolerance
)

// ErrorSeverity represents the severity level of an audit error
type ErrorSeverity int

const (
	SeverityInfo ErrorSeverity = iota
	SeverityWarning
	SeverityError
	SeverityCritical
)

// AuditError represents a single audit error entry
type AuditError struct {
	Code        ErrorCode
	Severity    ErrorSeverity
	Entity      string
	Handle      string
	Message     string
	Context     string
	Line        int    // Line number in DXF file if available
	Column      int    // Column number in DXF file if available
	CanFix      bool   // Whether the error can be automatically fixed
	Fixed       bool   // Whether the error was fixed
	RecoveryLog string // Details of recovery actions taken
}

// String returns a string representation of the error code
func (e ErrorCode) String() string {
	switch e {
	case ErrorMissingSection:
		return "Missing required DXF section"
	case ErrorInvalidHandle:
		return "Invalid entity handle"
	case ErrorDuplicateHandle:
		return "Duplicate entity handle"
	case ErrorBrokenStructure:
		return "Broken entity structure"
	case ErrorMissingHeader:
		return "Missing DXF header"
	case ErrorInvalidSection:
		return "Invalid section structure"
	case ErrorCorruptedTags:
		return "Corrupted DXF tags"
	case ErrorInvalidGroupCode:
		return "Invalid group code"
	case ErrorMissingEndSec:
		return "Missing ENDSEC tag"
	case ErrorInvalidEntityStructure:
		return "Invalid entity structure"
	case ErrorCorruptedData:
		return "Corrupted entity data"
	case ErrorInvalidDatabase:
		return "Invalid entity database"
	case ErrorMissingRootDict:
		return "Missing root dictionary"
	case ErrorInvalidLayerTable:
		return "Invalid layer table"
	case ErrorInvalidBlockTable:
		return "Invalid block table"
	case ErrorInvalidStyleTable:
		return "Invalid style table"
	case ErrorInvalidDimStyleTable:
		return "Invalid dimension style table"
	case ErrorInvalidAppIDTable:
		return "Invalid application ID table"
	case ErrorInvalidUCSTable:
		return "Invalid UCS table"
	case ErrorInvalidViewTable:
		return "Invalid view table"
	case ErrorInvalidViewportTable:
		return "Invalid viewport table"
	case ErrorUndefinedLayer:
		return "Reference to undefined layer"
	case ErrorUndefinedLinetype:
		return "Reference to undefined linetype"
	case ErrorUndefinedStyle:
		return "Reference to undefined text style"
	case ErrorUndefinedBlock:
		return "Reference to undefined block"
	case ErrorUndefinedText:
		return "Reference to undefined text style"
	case ErrorUndefinedDimStyle:
		return "Reference to undefined dimension style"
	case ErrorUndefinedAppID:
		return "Reference to undefined application ID"
	case ErrorUndefinedUCS:
		return "Reference to undefined UCS"
	case ErrorUndefinedView:
		return "Reference to undefined view"
	case ErrorUndefinedViewport:
		return "Reference to undefined viewport"
	case ErrorBrokenLink:
		return "Broken entity link"
	case ErrorInvalidReference:
		return "Invalid entity reference"
	case ErrorCircularReference:
		return "Circular reference detected"
	case ErrorOrphanedEntity:
		return "Orphaned entity without owner"
	case ErrorMissingReference:
		return "Missing required reference"
	case ErrorInvalidColor:
		return "Invalid color value"
	case ErrorInvalidLineweight:
		return "Invalid lineweight value"
	case ErrorInvalidHandleFormat:
		return "Invalid handle format"
	case ErrorInvalidLayerName:
		return "Invalid layer name"
	case ErrorInvalidLinetypeName:
		return "Invalid linetype name"
	case ErrorInvalidStyleName:
		return "Invalid style name"
	case ErrorInvalidDimStyleName:
		return "Invalid dimension style name"
	case ErrorInvalidPoint:
		return "Invalid point coordinates"
	case ErrorInvalidVector:
		return "Invalid vector direction"
	case ErrorInvalidNormal:
		return "Invalid normal vector"
	case ErrorInvalidSpline:
		return "Invalid spline data"
	case ErrorInvalidMesh:
		return "Invalid mesh data"
	case ErrorInvalidHatch:
		return "Invalid hatch boundary"
	case ErrorInvalidDimension:
		return "Invalid dimension geometry"
	case ErrorInvalidInsertPoint:
		return "Invalid insert point"
	case ErrorInvalidScale:
		return "Invalid scale factor"
	case ErrorInvalidRotation:
		return "Invalid rotation angle"
	case ErrorInvalidExtrusion:
		return "Invalid extrusion direction"
	case ErrorInvalidBoundaryPath:
		return "Invalid boundary path"
	case ErrorInvalidControlPoints:
		return "Invalid control points"
	case ErrorInvalidKnotVector:
		return "Invalid knot vector"
	case ErrorInvalidWeights:
		return "Invalid weight values"
	case ErrorInvalidTolerance:
		return "Invalid tolerance values"
	default:
		return "Unknown error"
	}
}

// String returns a string representation of the severity level
func (s ErrorSeverity) String() string {
	switch s {
	case SeverityInfo:
		return "INFO"
	case SeverityWarning:
		return "WARNING"
	case SeverityError:
		return "ERROR"
	case SeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// String returns a formatted error message
func (e *AuditError) String() string {
	var sb strings.Builder

	// Basic error information
	sb.WriteString(fmt.Sprintf("[%s] %s", e.Severity.String(), e.Code.String()))

	// Add entity and handle information if available
	if e.Entity != "" {
		sb.WriteString(fmt.Sprintf(" (Entity: %s", e.Entity))
		if e.Handle != "" {
			sb.WriteString(fmt.Sprintf(", Handle: %s", e.Handle))
		}
		sb.WriteString(")")
	}

	// Add message
	sb.WriteString(fmt.Sprintf(": %s", e.Message))

	// Add context if available
	if e.Context != "" {
		sb.WriteString(fmt.Sprintf(" [%s]", e.Context))
	}

	// Add line/column if available
	if e.Line > 0 {
		sb.WriteString(fmt.Sprintf(" at line %d", e.Line))
		if e.Column > 0 {
			sb.WriteString(fmt.Sprintf(", column %d", e.Column))
		}
	}

	// Add fix status
	if e.CanFix {
		if e.Fixed {
			sb.WriteString(" [FIXED]")
			if e.RecoveryLog != "" {
				sb.WriteString(fmt.Sprintf(" - %s", e.RecoveryLog))
			}
		} else {
			sb.WriteString(" [FIXABLE]")
		}
	}

	return sb.String()
}

// ErrorCategory groups errors by category
type ErrorCategory string

const (
	CategoryStructure ErrorCategory = "STRUCTURE"
	CategoryReference ErrorCategory = "REFERENCE"
	CategoryProperty  ErrorCategory = "PROPERTY"
	CategoryGeometry  ErrorCategory = "GEOMETRY"
	CategoryData      ErrorCategory = "DATA"
	CategoryRecovery  ErrorCategory = "RECOVERY"
)

// GetCategory returns the category for this error code
func (e ErrorCode) GetCategory() ErrorCategory {
	switch e {
	case ErrorMissingSection, ErrorInvalidHandle, ErrorDuplicateHandle,
		ErrorBrokenStructure, ErrorMissingHeader, ErrorInvalidSection,
		ErrorCorruptedTags, ErrorInvalidGroupCode, ErrorMissingEndSec,
		ErrorInvalidEntityStructure, ErrorCorruptedData, ErrorInvalidDatabase,
		ErrorMissingRootDict, ErrorInvalidLayerTable, ErrorInvalidBlockTable,
		ErrorInvalidStyleTable, ErrorInvalidDimStyleTable, ErrorInvalidAppIDTable,
		ErrorInvalidUCSTable, ErrorInvalidViewTable, ErrorInvalidViewportTable:
		return CategoryStructure

	case ErrorUndefinedLayer, ErrorUndefinedLinetype, ErrorUndefinedStyle,
		ErrorUndefinedBlock, ErrorUndefinedText, ErrorUndefinedDimStyle,
		ErrorUndefinedAppID, ErrorUndefinedUCS, ErrorUndefinedView,
		ErrorUndefinedViewport, ErrorBrokenLink, ErrorInvalidReference,
		ErrorCircularReference, ErrorOrphanedEntity, ErrorMissingReference:
		return CategoryReference

	case ErrorInvalidColor, ErrorInvalidLineweight, ErrorInvalidHandleFormat,
		ErrorInvalidLayerName, ErrorInvalidLinetypeName, ErrorInvalidStyleName,
		ErrorInvalidDimStyleName:
		return CategoryProperty

	case ErrorInvalidPoint, ErrorInvalidVector, ErrorInvalidNormal,
		ErrorInvalidSpline, ErrorInvalidMesh, ErrorInvalidHatch,
		ErrorInvalidDimension, ErrorInvalidInsertPoint, ErrorInvalidScale,
		ErrorInvalidRotation, ErrorInvalidExtrusion, ErrorInvalidBoundaryPath,
		ErrorInvalidControlPoints, ErrorInvalidKnotVector, ErrorInvalidWeights,
		ErrorInvalidTolerance:
		return CategoryGeometry

	default:
		return CategoryData
	}
}

// NewError creates a new audit error
func NewError(code ErrorCode, severity ErrorSeverity, entity, handle, message string) *AuditError {
	return &AuditError{
		Code:     code,
		Severity: severity,
		Entity:   entity,
		Handle:   handle,
		Message:  message,
		CanFix:   isFixable(code),
	}
}

// NewErrorWithContext creates a new audit error with additional context
func NewErrorWithContext(code ErrorCode, severity ErrorSeverity, entity, handle, message, context string) *AuditError {
	return &AuditError{
		Code:     code,
		Severity: severity,
		Entity:   entity,
		Handle:   handle,
		Message:  message,
		Context:  context,
		CanFix:   isFixable(code),
	}
}

// NewErrorWithLocation creates a new audit error with file location information
func NewErrorWithLocation(code ErrorCode, severity ErrorSeverity, entity, handle, message string, line, column int) *AuditError {
	return &AuditError{
		Code:     code,
		Severity: severity,
		Entity:   entity,
		Handle:   handle,
		Message:  message,
		Line:     line,
		Column:   column,
		CanFix:   isFixable(code),
	}
}

// MarkFixed marks an error as fixed and adds recovery log
func (e *AuditError) MarkFixed(recoveryLog string) {
	e.Fixed = true
	e.RecoveryLog = recoveryLog
}

// isFixable determines if an error code is automatically fixable
func isFixable(code ErrorCode) bool {
	fixableErrors := map[ErrorCode]bool{
		// Structure errors that can be fixed
		ErrorDuplicateHandle:        true,
		ErrorBrokenStructure:        true,
		ErrorCorruptedTags:          true,
		ErrorInvalidGroupCode:       true,
		ErrorInvalidEntityStructure: true,

		// Reference errors that can be fixed
		ErrorUndefinedLayer:    true,
		ErrorUndefinedLinetype: true,
		ErrorUndefinedStyle:    true,
		ErrorBrokenLink:        true,
		ErrorInvalidReference:  true,
		ErrorOrphanedEntity:    true,

		// Property errors that can be fixed
		ErrorInvalidColor:        true,
		ErrorInvalidLineweight:   true,
		ErrorInvalidHandleFormat: true,

		// Geometry errors that can be fixed
		ErrorInvalidPoint:       true,
		ErrorInvalidVector:      true,
		ErrorInvalidNormal:      true,
		ErrorInvalidInsertPoint: true,
		ErrorInvalidScale:       true,
		ErrorInvalidRotation:    true,
		ErrorInvalidExtrusion:   true,
	}

	return fixableErrors[code]
}

// AuditLogger provides logging functionality for audit operations
type AuditLogger struct {
	Verbose   bool
	LogFile   *log.Logger
	ErrorFile *log.Logger
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(verbose bool) *AuditLogger {
	return &AuditLogger{
		Verbose: verbose,
	}
}

// Log logs an audit error
func (l *AuditLogger) Log(err *AuditError) {
	if l.Verbose {
		fmt.Println(err.String())
	}

	if l.LogFile != nil {
		l.LogFile.Println(err.String())
	}

	if err.Severity >= SeverityError && l.ErrorFile != nil {
		l.ErrorFile.Println(err.String())
	}
}

// LogMessage logs a general audit message
func (l *AuditLogger) LogMessage(message string) {
	if l.Verbose {
		fmt.Printf("[AUDIT] %s\n", message)
	}

	if l.LogFile != nil {
		l.LogFile.Printf("[AUDIT] %s", message)
	}
}

// LogRecovery logs a recovery action
func (l *AuditLogger) LogRecovery(action, details string) {
	message := fmt.Sprintf("[RECOVERY] %s: %s", action, details)
	if l.Verbose {
		fmt.Println(message)
	}

	if l.LogFile != nil {
		l.LogFile.Println(message)
	}
}
