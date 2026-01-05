package audit

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/edanko/dxf/drawing"
)

// AuditConfig holds configuration for audit operations
type AuditConfig struct {
	AutoFix        bool
	StrictMode     bool
	RecoverLinks   bool
	RemoveOrphans  bool
	DetectCycles   bool
	Verbose        bool
	MaxErrors      int
	OutputFile     string
	RecoveryMode   RecoveryMode
	GenerateReport bool
}

// DefaultAuditConfig returns default audit configuration
func DefaultAuditConfig() *AuditConfig {
	return &AuditConfig{
		AutoFix:        true,
		StrictMode:     false,
		RecoverLinks:   true,
		RemoveOrphans:  true,
		DetectCycles:   true,
		Verbose:        false,
		MaxErrors:      1000,
		RecoveryMode:   RecoveryConservative,
		GenerateReport: true,
	}
}

// AuditResult contains the results of an audit operation
type AuditResult struct {
	Success         bool
	TotalErrors     int
	CriticalErrors  int
	FixedErrors     int
	CyclesDetected  int
	Duration        time.Duration
	Errors          []*AuditError
	Cycles          [][]string
	Summary         map[string]map[string]int
	ReportFile      string
	Recommendations []string
}

// ComprehensiveAuditor provides high-level audit functionality
type ComprehensiveAuditor struct {
	config *AuditConfig
	logger *AuditLogger
}

// NewComprehensiveAuditor creates a new comprehensive auditor
func NewComprehensiveAuditor(config *AuditConfig) *ComprehensiveAuditor {
	if config == nil {
		config = DefaultAuditConfig()
	}

	logger := NewAuditLogger(config.Verbose)

	return &ComprehensiveAuditor{
		config: config,
		logger: logger,
	}
}

// AuditDrawing performs a comprehensive audit of a DXF drawing
func (ca *ComprehensiveAuditor) AuditDrawing(doc *drawing.Drawing) (*AuditResult, error) {
	startTime := time.Now()

	result := &AuditResult{
		Success:         false,
		TotalErrors:     0,
		CriticalErrors:  0,
		FixedErrors:     0,
		CyclesDetected:  0,
		Duration:        0,
		Errors:          make([]*AuditError, 0),
		Cycles:          make([][]string, 0),
		Summary:         make(map[string]map[string]int),
		Recommendations: make([]string, 0),
	}

	ca.logger.LogMessage("Starting comprehensive DXF audit")

	// Step 1: Basic validation using existing auditor
	auditor := NewAuditor(doc)
	auditor.SetOptions(ca.config.AutoFix, ca.config.StrictMode,
		ca.config.RecoverLinks, ca.config.RemoveOrphans)
	auditor.SetLogger(ca.logger)

	errors, err := auditor.Run()
	if err != nil {
		return result, fmt.Errorf("audit failed: %w", err)
	}

	result.Errors = errors
	result.TotalErrors = len(errors)

	// Count critical errors
	for _, err := range errors {
		if err.Severity == SeverityCritical {
			result.CriticalErrors++
		}
	}

	// Step 2: Block cycle detection (if enabled)
	if ca.config.DetectCycles {
		if err := ca.detectBlockCycles(doc, result); err != nil {
			return result, fmt.Errorf("block cycle detection failed: %w", err)
		}
	}

	// Step 3: Generate summary and recommendations
	ca.generateSummary(result)

	// Step 4: Generate report (if enabled)
	if ca.config.GenerateReport {
		if err := ca.generateReport(result); err != nil {
			ca.logger.LogMessage(fmt.Sprintf("Report generation failed: %v", err))
		}
	}

	result.Duration = time.Since(startTime)
	result.Success = true

	ca.logger.LogMessage(fmt.Sprintf("Audit completed successfully in %v", result.Duration))

	return result, nil
}

// AuditFile performs audit on a DXF file
func (ca *ComprehensiveAuditor) AuditFile(filename string) (*AuditResult, error) {
	ca.logger.LogMessage(fmt.Sprintf("Auditing DXF file: %s", filename))

	// Load drawing using a placeholder function
	doc, err := ca.loadDrawing(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to load DXF file: %w", err)
	}

	result, err := ca.AuditDrawing(doc)
	if err != nil {
		return result, fmt.Errorf("audit failed: %w", err)
	}

	return result, nil
}

// loadDrawing is a placeholder for loading DXF files
func (ca *ComprehensiveAuditor) loadDrawing(filename string) (*drawing.Drawing, error) {
	// This would use the actual DXF loading function
	// For now, return a placeholder
	return nil, fmt.Errorf("DXF file loading not implemented in this context")
}

// detectBlockCycles performs block cycle detection
func (ca *ComprehensiveAuditor) detectBlockCycles(doc *drawing.Drawing, result *AuditResult) error {
	ca.logger.LogMessage("Detecting block reference cycles")

	// For now, we'll add a placeholder implementation
	// In a real implementation, this would use the BlockCycleDetector

	// Placeholder cycle detection
	// result.Cycles = append(result.Cycles, []string{"BLOCK1", "BLOCK2", "BLOCK1"})
	// result.CyclesDetected = len(result.Cycles)

	return nil
}

// generateSummary generates a summary of audit results
func (ca *ComprehensiveAuditor) generateSummary(result *AuditResult) {
	for _, err := range result.Errors {
		// Update summary
		category := string(err.Code.GetCategory())
		severity := err.Severity.String()

		if result.Summary[category] == nil {
			result.Summary[category] = make(map[string]int)
		}
		result.Summary[category][severity]++
	}

	// Generate recommendations
	result.Recommendations = ca.generateRecommendations(result)
}

// generateRecommendations generates recommendations based on audit results
func (ca *ComprehensiveAuditor) generateRecommendations(result *AuditResult) []string {
	recommendations := make([]string, 0)

	if result.CriticalErrors > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("Address %d critical errors immediately as they may prevent file loading", result.CriticalErrors))
	}

	if result.CyclesDetected > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("Resolve %d block reference cycles to avoid infinite loops in CAD software", result.CyclesDetected))
	}

	if result.TotalErrors-result.CriticalErrors > 50 {
		recommendations = append(recommendations,
			"Consider running audit in aggressive mode to fix multiple issues automatically")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "DXF file is in good condition")
	}

	return recommendations
}

// generateReport generates a detailed audit report
func (ca *ComprehensiveAuditor) generateReport(result *AuditResult) error {
	if !ca.config.GenerateReport {
		return nil
	}

	reportFile := ca.config.OutputFile
	if reportFile == "" {
		timestamp := time.Now().Format("20060102_150405")
		reportFile = fmt.Sprintf("audit_report_%s.txt", timestamp)
	}

	file, err := os.Create(reportFile)
	if err != nil {
		return fmt.Errorf("failed to create report file: %w", err)
	}
	defer file.Close()

	// Write report header
	file.WriteString("=== DXF AUDIT REPORT ===\n")
	file.WriteString(fmt.Sprintf("Generated: %s\n", time.Now().Format(time.RFC3339)))
	file.WriteString(fmt.Sprintf("Duration: %v\n", result.Duration))
	file.WriteString(fmt.Sprintf("Success: %v\n\n", result.Success))

	// Write summary
	file.WriteString("=== SUMMARY ===\n")
	file.WriteString(fmt.Sprintf("Total Errors: %d\n", result.TotalErrors))
	file.WriteString(fmt.Sprintf("Critical Errors: %d\n", result.CriticalErrors))
	file.WriteString(fmt.Sprintf("Fixed Errors: %d\n", result.FixedErrors))
	file.WriteString(fmt.Sprintf("Block Cycles: %d\n\n", result.CyclesDetected))

	// Write error details
	if len(result.Errors) > 0 {
		file.WriteString("=== ERROR DETAILS ===\n")
		for i, err := range result.Errors {
			if i >= 100 { // Limit error details in report
				file.WriteString(fmt.Sprintf("... and %d more errors\n", len(result.Errors)-100))
				break
			}
			file.WriteString(fmt.Sprintf("%d. %s\n", i+1, err.String()))
		}
		file.WriteString("\n")
	}

	// Write cycle details
	if len(result.Cycles) > 0 {
		file.WriteString("=== BLOCK CYCLES ===\n")
		for i, cycle := range result.Cycles {
			cycleStr := strings.Join(cycle, " → ")
			file.WriteString(fmt.Sprintf("%d. %s\n", i+1, cycleStr))
		}
		file.WriteString("\n")
	}

	// Write recommendations
	if len(result.Recommendations) > 0 {
		file.WriteString("=== RECOMMENDATIONS ===\n")
		for i, rec := range result.Recommendations {
			file.WriteString(fmt.Sprintf("%d. %s\n", i+1, rec))
		}
	}

	result.ReportFile = reportFile
	ca.logger.LogMessage(fmt.Sprintf("Audit report generated: %s", reportFile))

	return nil
}

// PrintResult prints audit result to console
func (ca *ComprehensiveAuditor) PrintResult(result *AuditResult) {
	fmt.Println("\n=== DXF AUDIT RESULT ===")
	fmt.Printf("Success: %v\n", result.Success)
	fmt.Printf("Duration: %v\n", result.Duration)
	fmt.Printf("Total Errors: %d\n", result.TotalErrors)
	fmt.Printf("Critical Errors: %d\n", result.CriticalErrors)
	fmt.Printf("Fixed Errors: %d\n", result.FixedErrors)
	fmt.Printf("Block Cycles: %d\n", result.CyclesDetected)

	if result.ReportFile != "" {
		fmt.Printf("Report: %s\n", result.ReportFile)
	}

	// Print summary
	for category, severities := range result.Summary {
		fmt.Printf("\n%s:\n", category)
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

	// Print recommendations
	if len(result.Recommendations) > 0 {
		fmt.Println("\nRecommendations:")
		for i, rec := range result.Recommendations {
			fmt.Printf("%d. %s\n", i+1, rec)
		}
	}
}

// Simple audit function for basic usage
func AuditDrawingComprehensive(doc *drawing.Drawing) (*AuditResult, error) {
	config := DefaultAuditConfig()
	auditor := NewComprehensiveAuditor(config)
	return auditor.AuditDrawing(doc)
}

// Simple audit function for file usage
func AuditFileComprehensive(filename string) (*AuditResult, error) {
	config := DefaultAuditConfig()
	auditor := NewComprehensiveAuditor(config)
	return auditor.AuditFile(filename)
}
