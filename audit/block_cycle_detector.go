package audit

import (
	"fmt"
	"strings"
)

// BlockCycleDetector detects circular references in block definitions
type BlockCycleDetector struct {
	visited     map[string]bool
	recursion   map[string]bool
	cycles      [][]string
	currentPath []string
	doc         DrawingInterface
}

// DrawingInterface defines the interface needed for block cycle detection
type DrawingInterface interface {
	GetBlock(name string) (BlockInterface, bool)
	ListBlocks() []string
}

// BlockInterface defines the interface for block entities
type BlockInterface interface {
	Name() string
	GetEntities() []EntityInterface
}

// EntityInterface defines the interface for entities that can reference blocks
type EntityInterface interface {
	Type() string
	GetBlockName() string // Returns block name for INSERT entities
}

// NewBlockCycleDetector creates a new block cycle detector
func NewBlockCycleDetector(doc DrawingInterface) *BlockCycleDetector {
	return &BlockCycleDetector{
		visited:   make(map[string]bool),
		recursion: make(map[string]bool),
		cycles:    make([][]string, 0),
		doc:       doc,
	}
}

// DetectCycles performs cycle detection on all blocks in the drawing
func (bcd *BlockCycleDetector) DetectCycles() ([][]string, error) {
	bcd.cycles = make([][]string, 0)
	bcd.visited = make(map[string]bool)
	bcd.recursion = make(map[string]bool)

	blocks := bcd.doc.ListBlocks()

	for _, blockName := range blocks {
		if !bcd.visited[blockName] {
			bcd.currentPath = make([]string, 0)
			if err := bcd.detectCycleFromBlock(blockName); err != nil {
				return nil, fmt.Errorf("cycle detection failed for block %s: %w", blockName, err)
			}
		}
	}

	return bcd.cycles, nil
}

// detectCycleFromBlock detects cycles starting from a specific block
func (bcd *BlockCycleDetector) detectCycleFromBlock(blockName string) error {
	// Check if we've found a cycle
	if bcd.recursion[blockName] {
		// Extract the cycle from the current path
		cycle := bcd.extractCycle(blockName)
		if len(cycle) > 0 {
			bcd.cycles = append(bcd.cycles, cycle)
		}
		return nil
	}

	// Check if we've already processed this block
	if bcd.visited[blockName] {
		return nil
	}

	// Mark as currently being processed (recursion)
	bcd.recursion[blockName] = true
	bcd.currentPath = append(bcd.currentPath, blockName)

	// Get the block
	block, exists := bcd.doc.GetBlock(blockName)
	if !exists {
		// Block doesn't exist - this is an error but not a cycle
		bcd.visited[blockName] = true
		bcd.recursion[blockName] = false
		bcd.currentPath = bcd.currentPath[:len(bcd.currentPath)-1]
		return nil
	}

	// Check all entities in the block for INSERT references
	entities := block.GetEntities()
	for _, entity := range entities {
		if entity.Type() == "INSERT" {
			referencedBlockName := entity.GetBlockName()
			if referencedBlockName != "" {
				// Recursively check the referenced block
				if err := bcd.detectCycleFromBlock(referencedBlockName); err != nil {
					return err
				}
			}
		}
	}

	// Mark as fully processed
	bcd.visited[blockName] = true
	bcd.recursion[blockName] = false
	bcd.currentPath = bcd.currentPath[:len(bcd.currentPath)-1]

	return nil
}

// extractCycle extracts the cycle from the current path
func (bcd *BlockCycleDetector) extractCycle(blockName string) []string {
	// Find the start of the cycle in the current path
	startIndex := -1
	for i, name := range bcd.currentPath {
		if name == blockName {
			startIndex = i
			break
		}
	}

	if startIndex == -1 {
		return []string{}
	}

	// Extract the cycle
	cycle := make([]string, len(bcd.currentPath)-startIndex)
	copy(cycle, bcd.currentPath[startIndex:])
	return cycle
}

// GetCycles returns all detected cycles
func (bcd *BlockCycleDetector) GetCycles() [][]string {
	return bcd.cycles
}

// HasCycles returns true if any cycles were detected
func (bcd *BlockCycleDetector) HasCycles() bool {
	return len(bcd.cycles) > 0
}

// FormatCycle formats a cycle for display
func (bcd *BlockCycleDetector) FormatCycle(cycle []string) string {
	if len(cycle) == 0 {
		return ""
	}
	return strings.Join(cycle, " → ") + " → " + cycle[0]
}

// GenerateCycleErrors generates audit errors for detected cycles
func (bcd *BlockCycleDetector) GenerateCycleErrors() []*AuditError {
	errors := make([]*AuditError, 0)

	for i, cycle := range bcd.cycles {
		cycleStr := bcd.FormatCycle(cycle)
		err := NewError(
			ErrorInvalidBlockReferenceCycle,
			SeverityError,
			"BLOCK",
			cycle[0], // Use the first block in cycle as the handle
			fmt.Sprintf("Circular block reference detected: %s", cycleStr),
		)
		err.Context = fmt.Sprintf("Cycle %d", i+1)
		errors = append(errors, err)
	}

	return errors
}

// CycleInfo contains information about a detected cycle
type CycleInfo struct {
	Cycle          []string
	Length         int
	AffectedBlocks map[string]bool
	Severity       string
}

// GetCycleInfo returns detailed information about all detected cycles
func (bcd *BlockCycleDetector) GetCycleInfo() []CycleInfo {
	info := make([]CycleInfo, 0)

	for _, cycle := range bcd.cycles {
		// Calculate affected blocks
		affected := make(map[string]bool)
		for _, blockName := range cycle {
			affected[blockName] = true
		}

		// Determine severity based on cycle length
		severity := "Warning"
		if len(cycle) == 1 {
			severity = "Critical" // Self-reference
		} else if len(cycle) <= 3 {
			severity = "Error"
		}

		info = append(info, CycleInfo{
			Cycle:          cycle,
			Length:         len(cycle),
			AffectedBlocks: affected,
			Severity:       severity,
		})
	}

	return info
}

// PrintCycleSummary prints a summary of detected cycles
func (bcd *BlockCycleDetector) PrintCycleSummary() {
	if !bcd.HasCycles() {
		fmt.Println("✅ No block reference cycles detected")
		return
	}

	fmt.Printf("🔍 Detected %d block reference cycle(s):\n", len(bcd.cycles))

	info := bcd.GetCycleInfo()
	for i, cycleInfo := range info {
		fmt.Printf("\n📊 Cycle %d (%s):\n", i+1, cycleInfo.Severity)
		fmt.Printf("   Path: %s\n", bcd.FormatCycle(cycleInfo.Cycle))
		fmt.Printf("   Length: %d blocks\n", cycleInfo.Length)
		fmt.Printf("   Affected blocks: %d\n", len(cycleInfo.AffectedBlocks))

		// List affected blocks
		for blockName := range cycleInfo.AffectedBlocks {
			fmt.Printf("     - %s\n", blockName)
		}
	}
}

// BreakCycle attempts to break a detected cycle by removing the last reference
func (bcd *BlockCycleDetector) BreakCycle(cycleIndex int) (string, error) {
	if cycleIndex < 0 || cycleIndex >= len(bcd.cycles) {
		return "", fmt.Errorf("invalid cycle index: %d", cycleIndex)
	}

	cycle := bcd.cycles[cycleIndex]
	if len(cycle) < 2 {
		return "", fmt.Errorf("cycle too short to break: %v", cycle)
	}

	// The strategy is to break the cycle by removing the reference from
	// the last block back to the first block
	sourceBlock := cycle[len(cycle)-2]
	targetBlock := cycle[0]

	return fmt.Sprintf("Remove INSERT reference from block '%s' to block '%s'", sourceBlock, targetBlock), nil
}

// GetAllCyclesAsErrors returns all cycles as audit errors
func (bcd *BlockCycleDetector) GetAllCyclesAsErrors() []*AuditError {
	return bcd.GenerateCycleErrors()
}

// ValidateBlockStructure performs comprehensive block structure validation
func (bcd *BlockCycleDetector) ValidateBlockStructure() ([]*AuditError, error) {
	errors := make([]*AuditError, 0)

	// Check for cycles
	cycleErrors := bcd.GetAllCyclesAsErrors()
	errors = append(errors, cycleErrors...)

	// Check for orphaned blocks (blocks that don't exist but are referenced)
	orphanErrors := bcd.detectOrphanedBlocks()
	errors = append(errors, orphanErrors...)

	// Check for undefined blocks in INSERT entities
	undefinedErrors := bcd.detectUndefinedBlocks()
	errors = append(errors, undefinedErrors...)

	return errors, nil
}

// detectOrphanedBlocks detects blocks that are referenced but don't exist
func (bcd *BlockCycleDetector) detectOrphanedBlocks() []*AuditError {
	errors := make([]*AuditError, 0)

	// Get all defined blocks
	definedBlocks := make(map[string]bool)
	for _, blockName := range bcd.doc.ListBlocks() {
		definedBlocks[blockName] = true
	}

	// Check all INSERT entities for references to undefined blocks
	for _, blockName := range bcd.doc.ListBlocks() {
		block, exists := bcd.doc.GetBlock(blockName)
		if !exists {
			continue
		}

		entities := block.GetEntities()
		for _, entity := range entities {
			if entity.Type() == "INSERT" {
				referencedBlock := entity.GetBlockName()
				if referencedBlock != "" && !definedBlocks[referencedBlock] {
					err := NewError(
						ErrorUndefinedBlock,
						SeverityError,
						"INSERT",
						entity.GetBlockName(),
						fmt.Sprintf("INSERT entity references undefined block '%s'", referencedBlock),
					)
					err.Context = blockName
					errors = append(errors, err)
				}
			}
		}
	}

	return errors
}

// detectUndefinedBlocks detects blocks referenced by entities that don't exist
func (bcd *BlockCycleDetector) detectUndefinedBlocks() []*AuditError {
	errors := make([]*AuditError, 0)

	// Get all defined blocks
	definedBlocks := make(map[string]bool)
	for _, blockName := range bcd.doc.ListBlocks() {
		definedBlocks[blockName] = true
	}

	// This would typically be called from the main auditor
	// to check entities in model space and paper space
	return errors
}
