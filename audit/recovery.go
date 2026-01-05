package audit

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// RecoveryMode defines the recovery strategy
type RecoveryMode int

const (
	RecoveryNone RecoveryMode = iota
	RecoveryConservative
	RecoveryAggressive
	RecoveryMaximum
)

// RecoveryOptions configures recovery behavior
type RecoveryOptions struct {
	Mode            RecoveryMode
	DetectEncoding  bool
	FixHandles      bool
	FixReferences   bool
	RemoveCorrupted bool
	MaxErrors       int
}

// DefaultRecoveryOptions returns sensible default recovery options
func DefaultRecoveryOptions() *RecoveryOptions {
	return &RecoveryOptions{
		Mode:            RecoveryConservative,
		DetectEncoding:  true,
		FixHandles:      true,
		FixReferences:   true,
		RemoveCorrupted: true,
		MaxErrors:       1000,
	}
}

// RecoveryResult contains the results of a recovery operation
type RecoveryResult struct {
	Success         bool
	OriginalSize    int
	RecoveredSize   int
	ErrorsFixed     int
	ErrorsRemaining int
	Warnings        []string
	Encoding        string
	Actions         []string
}

// String returns a formatted summary of the recovery result
func (r *RecoveryResult) String() string {
	var sb strings.Builder
	sb.WriteString("=== DXF Recovery Result ===\n")
	sb.WriteString(fmt.Sprintf("Success: %v\n", r.Success))
	sb.WriteString(fmt.Sprintf("Size: %d → %d bytes (%.1f%% reduction)\n",
		r.OriginalSize, r.RecoveredSize,
		float64(r.OriginalSize-r.RecoveredSize)/float64(r.OriginalSize)*100))
	sb.WriteString(fmt.Sprintf("Errors Fixed: %d\n", r.ErrorsFixed))
	sb.WriteString(fmt.Sprintf("Errors Remaining: %d\n", r.ErrorsRemaining))

	if r.Encoding != "" {
		sb.WriteString(fmt.Sprintf("Detected Encoding: %s\n", r.Encoding))
	}

	if len(r.Warnings) > 0 {
		sb.WriteString(fmt.Sprintf("Warnings: %d\n", len(r.Warnings)))
		for i, warning := range r.Warnings {
			if i >= 10 { // Limit warnings displayed
				sb.WriteString(fmt.Sprintf("  ... and %d more\n", len(r.Warnings)-10))
				break
			}
			sb.WriteString(fmt.Sprintf("  • %s\n", warning))
		}
	}

	if len(r.Actions) > 0 {
		sb.WriteString("Recovery Actions:\n")
		for _, action := range r.Actions {
			sb.WriteString(fmt.Sprintf("  ✓ %s\n", action))
		}
	}

	return sb.String()
}

// Recoverer handles DXF file recovery operations
type Recoverer struct {
	Options *RecoveryOptions
	Logger  *AuditLogger
}

// NewRecoverer creates a new recoverer with the given options
func NewRecoverer(options *RecoveryOptions) *Recoverer {
	if options == nil {
		options = DefaultRecoveryOptions()
	}

	return &Recoverer{
		Options: options,
		Logger:  NewAuditLogger(true),
	}
}

// RecoverFile attempts to recover a corrupted DXF file from bytes
func (r *Recoverer) RecoverFile(data []byte) (*RecoveryResult, error) {
	result := &RecoveryResult{
		OriginalSize: len(data),
		Actions:      make([]string, 0),
		Warnings:     make([]string, 0),
	}

	r.Logger.LogMessage("Starting DXF file recovery")

	// Step 1: Detect encoding
	encoding := "utf-8"
	if r.Options.DetectEncoding {
		encoding = r.detectEncoding(data)
		result.Encoding = encoding
		result.Actions = append(result.Actions, fmt.Sprintf("Detected encoding: %s", encoding))
	}

	// Step 2: Basic tag structure repair
	repairedData, err := r.repairTagStructure(data, result)
	if err != nil {
		return nil, fmt.Errorf("tag structure repair failed: %w", err)
	}

	// Step 3: Handle and reference repair
	if r.Options.FixHandles || r.Options.FixReferences {
		repairedData, err = r.repairHandlesAndReferences(repairedData, result)
		if err != nil {
			return nil, fmt.Errorf("handle/reference repair failed: %w", err)
		}
	}

	// Step 4: Section structure validation and repair
	repairedData, err = r.repairSectionStructure(repairedData, result)
	if err != nil {
		return nil, fmt.Errorf("section structure repair failed: %w", err)
	}

	// Step 5: Final cleanup
	repairedData = r.finalCleanup(repairedData, result)

	result.RecoveredSize = len(repairedData)
	result.Success = true

	r.Logger.LogMessage(fmt.Sprintf("Recovery completed: %d bytes → %d bytes",
		result.OriginalSize, result.RecoveredSize))

	return result, nil
}

// detectEncoding attempts to detect the text encoding of the DXF file
func (r *Recoverer) detectEncoding(data []byte) string {
	// Simple encoding detection based on BOM and content analysis

	// Check for UTF-8 BOM
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return "utf-8"
	}

	// Check for UTF-16 LE BOM
	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xFE {
		return "utf-16le"
	}

	// Check for UTF-16 BE BOM
	if len(data) >= 2 && data[0] == 0xFE && data[1] == 0xFF {
		return "utf-16be"
	}

	// Analyze content for common encodings
	text := string(data[:min(len(data), 4096)])

	// Look for ASCII range content
	if r.isMostlyASCII(text) {
		return "utf-8"
	}

	// Default to utf-8 for DXF files (most common)
	return "utf-8"
}

// isMostlyASCII checks if text is mostly ASCII characters
func (r *Recoverer) isMostlyASCII(text string) bool {
	asciiCount := 0
	totalCount := 0

	for _, r := range text {
		if r < 128 {
			asciiCount++
		}
		totalCount++
	}

	return totalCount == 0 || float64(asciiCount)/float64(totalCount) > 0.8
}

// repairTagStructure fixes malformed DXF tags
func (r *Recoverer) repairTagStructure(data []byte, result *RecoveryResult) ([]byte, error) {
	r.Logger.LogMessage("Repairing tag structure")

	// Convert to string for processing
	content := string(data)
	lines := strings.Split(content, "\n")

	var repairedLines []string
	errorsFixed := 0

	// Regular expressions for tag validation
	groupCodePattern := regexp.MustCompile(`^\s*(\d+)(?:\s*)$`)

	i := 0
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])

		// Skip empty lines
		if line == "" {
			i++
			continue
		}

		// Check if this looks like a group code
		if groupCodePattern.MatchString(line) {
			// Get the group code
			groupCodeStr := groupCodePattern.FindStringSubmatch(line)[1]
			groupCode, err := strconv.Atoi(groupCodeStr)

			if err != nil {
				// Invalid group code - skip or fix
				if r.Options.RemoveCorrupted {
					i += 2 // Skip value line too
					errorsFixed++
					result.Warnings = append(result.Warnings,
						fmt.Sprintf("Removed invalid group code: %s", line))
					continue
				}
			}

			// Check if there's a corresponding value line
			if i+1 < len(lines) {
				valueLine := strings.TrimSpace(lines[i+1])

				// Validate based on group code type
				if r.isValidGroupCodeValue(groupCode, valueLine) {
					repairedLines = append(repairedLines, lines[i], lines[i+1])
				} else {
					// Fix or remove invalid value
					if r.Options.RemoveCorrupted {
						fixedValue := r.fixGroupCodeValue(groupCode, valueLine)
						if fixedValue != "" {
							repairedLines = append(repairedLines, lines[i], fixedValue)
							errorsFixed++
							result.Actions = append(result.Actions,
								fmt.Sprintf("Fixed value for group code %d", groupCode))
						} else {
							errorsFixed++
							result.Warnings = append(result.Warnings,
								fmt.Sprintf("Removed invalid value for group code %d", groupCode))
						}
					} else {
						repairedLines = append(repairedLines, lines[i], lines[i+1])
					}
				}
			} else {
				// Missing value line - add default or skip
				if r.Options.RemoveCorrupted {
					defaultValue := r.getDefaultValueForGroupCode(groupCode)
					if defaultValue != "" {
						repairedLines = append(repairedLines, lines[i], defaultValue)
						errorsFixed++
						result.Actions = append(result.Actions,
							fmt.Sprintf("Added missing value for group code %d", groupCode))
					}
				}
			}
			i += 2
		} else {
			// Unpaired line - skip or try to fix
			if r.Options.RemoveCorrupted {
				errorsFixed++
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("Removed unpaired line: %s", line))
			} else {
				repairedLines = append(repairedLines, lines[i])
			}
			i++
		}
	}

	result.ErrorsFixed += errorsFixed
	result.Actions = append(result.Actions,
		fmt.Sprintf("Repaired %d tag structure errors", errorsFixed))

	return []byte(strings.Join(repairedLines, "\n")), nil
}

// isValidGroupCodeValue checks if a value is valid for a given group code
func (r *Recoverer) isValidGroupCodeValue(groupCode int, value string) bool {
	switch {
	case groupCode >= 0 && groupCode <= 9:
		// String values
		return true
	case groupCode >= 10 && groupCode <= 39:
		// Point coordinates, primary values
		_, err := strconv.ParseFloat(value, 64)
		return err == nil
	case groupCode >= 40 && groupCode <= 59:
		// Floating point values
		_, err := strconv.ParseFloat(value, 64)
		return err == nil
	case groupCode >= 60 && groupCode <= 79:
		// Integer values
		_, err := strconv.Atoi(value)
		return err == nil
	case groupCode >= 90 && groupCode <= 99:
		// 32-bit integer values
		_, err := strconv.ParseInt(value, 10, 32)
		return err == nil
	case groupCode >= 100 && groupCode <= 102:
		// String values
		return true
	default:
		// Default to valid
		return true
	}
}

// fixGroupCodeValue attempts to fix an invalid value
func (r *Recoverer) fixGroupCodeValue(groupCode int, value string) string {
	switch {
	case groupCode >= 0 && groupCode <= 9:
		// String values - sanitize
		return r.sanitizeString(value)
	case groupCode >= 10 && groupCode <= 59:
		// Numeric values - try to extract number
		return r.extractNumber(value)
	case groupCode >= 60 && groupCode <= 99:
		// Integer values - try to extract integer
		return r.extractInteger(value)
	default:
		// String values - sanitize
		return r.sanitizeString(value)
	}
}

// getDefaultValueForGroupCode returns a default value for a group code
func (r *Recoverer) getDefaultValueForGroupCode(groupCode int) string {
	switch groupCode {
	case 0:
		return "SECTION"
	case 1:
		return ""
	case 2:
		return ""
	case 8:
		return "0" // Default layer
	case 10, 20, 30:
		return "0.0" // Default coordinates
	case 62:
		return "256" // ByLayer color
	case 70:
		return "0" // Default flag
	default:
		return "0"
	}
}

// sanitizeString cleans up string values
func (r *Recoverer) sanitizeString(s string) string {
	// Remove control characters and invalid DXF characters
	var result strings.Builder
	for _, r := range s {
		if r >= 32 && r <= 126 || r == '\n' || r == '\r' || r == '\t' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// extractNumber tries to extract a valid number from a string
func (r *Recoverer) extractNumber(s string) string {
	// Simple number extraction
	re := regexp.MustCompile(`-?\d+\.?\d*`)
	match := re.FindString(s)
	if match != "" {
		return match
	}
	return "0.0"
}

// extractInteger tries to extract a valid integer from a string
func (r *Recoverer) extractInteger(s string) string {
	re := regexp.MustCompile(`-?\d+`)
	match := re.FindString(s)
	if match != "" {
		return match
	}
	return "0"
}

// repairHandlesAndReferences fixes broken handles and references
func (r *Recoverer) repairHandlesAndReferences(data []byte, result *RecoveryResult) ([]byte, error) {
	r.Logger.LogMessage("Repairing handles and references")

	// This would implement handle and reference repair logic
	// For now, return data unchanged
	result.Actions = append(result.Actions, "Handle and reference repair not implemented yet")

	return data, nil
}

// repairSectionStructure fixes broken DXF section structure
func (r *Recoverer) repairSectionStructure(data []byte, result *RecoveryResult) ([]byte, error) {
	r.Logger.LogMessage("Repairing section structure")

	content := string(data)
	lines := strings.Split(content, "\n")

	var repairedLines []string
	sectionStack := make([]string, 0)
	errorsFixed := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check for SECTION
		if line == "0" {
			if len(repairedLines) > 0 && repairedLines[len(repairedLines)-1] == "0" {
				repairedLines = append(repairedLines, line)
				continue
			}
		}

		// Check for SECTION/ENDSEC pairs
		if line == "SECTION" {
			sectionStack = append(sectionStack, "SECTION")
		} else if line == "ENDSEC" {
			if len(sectionStack) > 0 && sectionStack[len(sectionStack)-1] == "SECTION" {
				sectionStack = sectionStack[:len(sectionStack)-1]
			} else {
				// Extra ENDSEC - remove
				errorsFixed++
				result.Warnings = append(result.Warnings, "Removed extra ENDSEC")
				continue
			}
		}

		repairedLines = append(repairedLines, line)
	}

	// Add missing ENDSEC tags
	for len(sectionStack) > 0 {
		repairedLines = append(repairedLines, "0", "ENDSEC")
		sectionStack = sectionStack[:len(sectionStack)-1]
		errorsFixed++
	}

	result.ErrorsFixed += errorsFixed
	result.Actions = append(result.Actions,
		fmt.Sprintf("Fixed %d section structure errors", errorsFixed))

	return []byte(strings.Join(repairedLines, "\n")), nil
}

// finalCleanup performs final cleanup operations
func (r *Recoverer) finalCleanup(data []byte, result *RecoveryResult) []byte {
	r.Logger.LogMessage("Performing final cleanup")

	content := string(data)

	// Remove duplicate empty lines
	re := regexp.MustCompile(`\n\s*\n\s*\n`)
	content = re.ReplaceAllString(content, "\n\n")

	// Remove trailing whitespace
	content = strings.TrimRight(content, " \t\n\r")

	// Ensure file ends with newline
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	result.Actions = append(result.Actions, "Applied final cleanup")

	return []byte(content)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
