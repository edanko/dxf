package audit

import (
	"testing"
)

// MockBlock implements BlockInterface for testing
type MockBlock struct {
	name     string
	entities []EntityInterface
}

func (m *MockBlock) Name() string {
	return m.name
}

func (m *MockBlock) GetEntities() []EntityInterface {
	return m.entities
}

// MockEntity implements EntityInterface for testing
type MockEntity struct {
	entityType string
	blockName  string
}

func (m *MockEntity) Type() string {
	return m.entityType
}

func (m *MockEntity) GetBlockName() string {
	return m.blockName
}

// MockDrawing implements DrawingInterface for testing
type MockDrawing struct {
	blocks map[string]*MockBlock
}

func (m *MockDrawing) GetBlock(name string) (BlockInterface, bool) {
	block, exists := m.blocks[name]
	return block, exists
}

func (m *MockDrawing) ListBlocks() []string {
	blocks := make([]string, 0, len(m.blocks))
	for name := range m.blocks {
		blocks = append(blocks, name)
	}
	return blocks
}

func TestBlockCycleDetectorCreation(t *testing.T) {
	drawing := &MockDrawing{
		blocks: make(map[string]*MockBlock),
	}
	detector := NewBlockCycleDetector(drawing)

	if detector == nil {
		t.Error("NewBlockCycleDetector should not return nil")
	}

	if detector.visited == nil {
		t.Error("visited map should be initialized")
	}

	if detector.recursion == nil {
		t.Error("recursion map should be initialized")
	}

	if detector.cycles == nil {
		t.Error("cycles slice should be initialized")
	}
}

func TestBlockCycleDetectorNoCycles(t *testing.T) {
	drawing := &MockDrawing{
		blocks: map[string]*MockBlock{
			"BLOCK1": {
				name: "BLOCK1",
				entities: []EntityInterface{
					&MockEntity{entityType: "LINE", blockName: ""},
				},
			},
			"BLOCK2": {
				name: "BLOCK2",
				entities: []EntityInterface{
					&MockEntity{entityType: "CIRCLE", blockName: ""},
				},
			},
		},
	}

	detector := NewBlockCycleDetector(drawing)
	cycles, err := detector.DetectCycles()

	if err != nil {
		t.Errorf("DetectCycles should not return error, got: %v", err)
	}

	if detector.HasCycles() {
		t.Error("Should not detect cycles for independent blocks")
	}

	if len(cycles) != 0 {
		t.Errorf("Expected 0 cycles, got %d", len(cycles))
	}
}

func TestBlockCycleDetectorSimpleCycle(t *testing.T) {
	drawing := &MockDrawing{
		blocks: map[string]*MockBlock{
			"BLOCK1": {
				name: "BLOCK1",
				entities: []EntityInterface{
					&MockEntity{entityType: "INSERT", blockName: "BLOCK2"},
				},
			},
			"BLOCK2": {
				name: "BLOCK2",
				entities: []EntityInterface{
					&MockEntity{entityType: "INSERT", blockName: "BLOCK1"},
				},
			},
		},
	}

	detector := NewBlockCycleDetector(drawing)
	cycles, err := detector.DetectCycles()

	if err != nil {
		t.Errorf("DetectCycles should not return error, got: %v", err)
	}

	if !detector.HasCycles() {
		t.Error("Should detect cycle between BLOCK1 and BLOCK2")
	}

	if len(cycles) != 1 {
		t.Errorf("Expected 1 cycle, got %d", len(cycles))
	}
}

func TestBlockCycleDetectorSelfReference(t *testing.T) {
	drawing := &MockDrawing{
		blocks: map[string]*MockBlock{
			"BLOCK1": {
				name: "BLOCK1",
				entities: []EntityInterface{
					&MockEntity{entityType: "INSERT", blockName: "BLOCK1"},
				},
			},
		},
	}

	detector := NewBlockCycleDetector(drawing)
	cycles, err := detector.DetectCycles()

	if err != nil {
		t.Errorf("DetectCycles should not return error, got: %v", err)
	}

	if !detector.HasCycles() {
		t.Error("Should detect self-reference cycle")
	}

	if len(cycles) != 1 {
		t.Errorf("Expected 1 cycle, got %d", len(cycles))
	}

	cycle := cycles[0]
	if len(cycle) != 1 {
		t.Errorf("Expected cycle length 1 for self-reference, got %d", len(cycle))
	}
}

func TestBlockCycleDetectorLongerCycle(t *testing.T) {
	drawing := &MockDrawing{
		blocks: map[string]*MockBlock{
			"BLOCK1": {
				name: "BLOCK1",
				entities: []EntityInterface{
					&MockEntity{entityType: "INSERT", blockName: "BLOCK2"},
				},
			},
			"BLOCK2": {
				name: "BLOCK2",
				entities: []EntityInterface{
					&MockEntity{entityType: "INSERT", blockName: "BLOCK3"},
				},
			},
			"BLOCK3": {
				name: "BLOCK3",
				entities: []EntityInterface{
					&MockEntity{entityType: "INSERT", blockName: "BLOCK1"},
				},
			},
		},
	}

	detector := NewBlockCycleDetector(drawing)
	cycles, err := detector.DetectCycles()

	if err != nil {
		t.Errorf("DetectCycles should not return error, got: %v", err)
	}

	if !detector.HasCycles() {
		t.Error("Should detect 3-block cycle")
	}

	if len(cycles) != 1 {
		t.Errorf("Expected 1 cycle, got %d", len(cycles))
	}

	cycle := cycles[0]
	if len(cycle) != 3 {
		t.Errorf("Expected cycle length 3, got %d", len(cycle))
	}
}

func TestBlockCycleDetectorGetCycles(t *testing.T) {
	drawing := &MockDrawing{
		blocks: map[string]*MockBlock{
			"BLOCK1": {
				name: "BLOCK1",
				entities: []EntityInterface{
					&MockEntity{entityType: "INSERT", blockName: "BLOCK2"},
				},
			},
			"BLOCK2": {
				name: "BLOCK2",
				entities: []EntityInterface{
					&MockEntity{entityType: "INSERT", blockName: "BLOCK1"},
				},
			},
		},
	}

	detector := NewBlockCycleDetector(drawing)
	detector.DetectCycles()

	cycles := detector.GetCycles()
	if len(cycles) != 1 {
		t.Errorf("Expected 1 cycle from GetCycles, got %d", len(cycles))
	}
}

func TestBlockCycleDetectorFormatCycle(t *testing.T) {
	drawing := &MockDrawing{
		blocks: map[string]*MockBlock{
			"BLOCK1": {
				name: "BLOCK1",
				entities: []EntityInterface{
					&MockEntity{entityType: "INSERT", blockName: "BLOCK2"},
				},
			},
			"BLOCK2": {
				name: "BLOCK2",
				entities: []EntityInterface{
					&MockEntity{entityType: "INSERT", blockName: "BLOCK1"},
				},
			},
		},
	}

	detector := NewBlockCycleDetector(drawing)
	detector.DetectCycles()

	if len(detector.cycles) == 0 {
		t.Fatal("Expected cycles to be detected")
	}

	formatted := detector.FormatCycle(detector.cycles[0])
	if formatted == "" {
		t.Error("FormatCycle should return non-empty string")
	}

	if len(formatted) < len("BLOCK1 → BLOCK2 → BLOCK1") {
		t.Error("FormatCycle should include all blocks in cycle")
	}
}

func TestBlockCycleDetectorGenerateCycleErrors(t *testing.T) {
	drawing := &MockDrawing{
		blocks: map[string]*MockBlock{
			"BLOCK1": {
				name: "BLOCK1",
				entities: []EntityInterface{
					&MockEntity{entityType: "INSERT", blockName: "BLOCK2"},
				},
			},
			"BLOCK2": {
				name: "BLOCK2",
				entities: []EntityInterface{
					&MockEntity{entityType: "INSERT", blockName: "BLOCK1"},
				},
			},
		},
	}

	detector := NewBlockCycleDetector(drawing)
	detector.DetectCycles()

	errors := detector.GenerateCycleErrors()
	if len(errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(errors))
	}

	if len(errors) > 0 {
		err := errors[0]
		if err.Code != ErrorInvalidBlockReferenceCycle {
			t.Errorf("Expected error code %d, got %d", ErrorInvalidBlockReferenceCycle, err.Code)
		}
	}
}

func TestBlockCycleDetectorGetCycleInfo(t *testing.T) {
	drawing := &MockDrawing{
		blocks: map[string]*MockBlock{
			"BLOCK1": {
				name: "BLOCK1",
				entities: []EntityInterface{
					&MockEntity{entityType: "INSERT", blockName: "BLOCK2"},
				},
			},
			"BLOCK2": {
				name: "BLOCK2",
				entities: []EntityInterface{
					&MockEntity{entityType: "INSERT", blockName: "BLOCK1"},
				},
			},
		},
	}

	detector := NewBlockCycleDetector(drawing)
	detector.DetectCycles()

	info := detector.GetCycleInfo()
	if len(info) != 1 {
		t.Errorf("Expected 1 cycle info, got %d", len(info))
	}

	if len(info) > 0 {
		if info[0].Length != 2 {
			t.Errorf("Expected cycle length 2, got %d", info[0].Length)
		}
		if info[0].Severity != "Error" {
			t.Errorf("Expected severity 'Error', got '%s'", info[0].Severity)
		}
	}
}

func TestBlockCycleDetectorGetCycleInfoSelfReference(t *testing.T) {
	drawing := &MockDrawing{
		blocks: map[string]*MockBlock{
			"BLOCK1": {
				name: "BLOCK1",
				entities: []EntityInterface{
					&MockEntity{entityType: "INSERT", blockName: "BLOCK1"},
				},
			},
		},
	}

	detector := NewBlockCycleDetector(drawing)
	detector.DetectCycles()

	info := detector.GetCycleInfo()
	if len(info) != 1 {
		t.Errorf("Expected 1 cycle info, got %d", len(info))
	}

	if len(info) > 0 && info[0].Severity != "Critical" {
		t.Errorf("Expected severity 'Critical' for self-reference, got '%s'", info[0].Severity)
	}
}

func TestBlockCycleDetectorBreakCycle(t *testing.T) {
	drawing := &MockDrawing{
		blocks: map[string]*MockBlock{
			"BLOCK1": {
				name: "BLOCK1",
				entities: []EntityInterface{
					&MockEntity{entityType: "INSERT", blockName: "BLOCK2"},
				},
			},
			"BLOCK2": {
				name: "BLOCK2",
				entities: []EntityInterface{
					&MockEntity{entityType: "INSERT", blockName: "BLOCK1"},
				},
			},
		},
	}

	detector := NewBlockCycleDetector(drawing)
	detector.DetectCycles()

	action, err := detector.BreakCycle(0)
	if err != nil {
		t.Errorf("BreakCycle should not return error, got: %v", err)
	}

	if action == "" {
		t.Error("BreakCycle should return action string")
	}
}

func TestBlockCycleDetectorBreakCycleInvalidIndex(t *testing.T) {
	drawing := &MockDrawing{
		blocks: map[string]*MockBlock{
			"BLOCK1": {
				name:     "BLOCK1",
				entities: []EntityInterface{},
			},
		},
	}

	detector := NewBlockCycleDetector(drawing)
	detector.DetectCycles()

	_, err := detector.BreakCycle(0)
	if err == nil {
		t.Error("BreakCycle should return error for invalid index")
	}
}

func TestBlockCycleDetectorValidateBlockStructure(t *testing.T) {
	drawing := &MockDrawing{
		blocks: map[string]*MockBlock{
			"BLOCK1": {
				name: "BLOCK1",
				entities: []EntityInterface{
					&MockEntity{entityType: "INSERT", blockName: "BLOCK2"},
				},
			},
			"BLOCK2": {
				name: "BLOCK2",
				entities: []EntityInterface{
					&MockEntity{entityType: "INSERT", blockName: "BLOCK1"},
				},
			},
		},
	}

	detector := NewBlockCycleDetector(drawing)
	_, err := detector.DetectCycles()
	if err != nil {
		t.Errorf("DetectCycles should not return error, got: %v", err)
	}

	errors, err := detector.ValidateBlockStructure()

	if err != nil {
		t.Errorf("ValidateBlockStructure should not return error, got: %v", err)
	}

	if len(errors) == 0 {
		t.Error("ValidateBlockStructure should return errors for cycles")
	}
}

func TestBlockCycleDetectorOrphanedBlock(t *testing.T) {
	drawing := &MockDrawing{
		blocks: map[string]*MockBlock{
			"BLOCK1": {
				name: "BLOCK1",
				entities: []EntityInterface{
					&MockEntity{entityType: "INSERT", blockName: "NONEXISTENT"},
				},
			},
		},
	}

	detector := NewBlockCycleDetector(drawing)
	detector.DetectCycles()

	errors := detector.detectOrphanedBlocks()
	if len(errors) == 0 {
		t.Error("Should detect orphaned block reference")
	}

	if len(errors) > 0 && errors[0].Code != ErrorUndefinedBlock {
		t.Errorf("Expected error code %d, got %d", ErrorUndefinedBlock, errors[0].Code)
	}
}

func TestBlockCycleDetectorNoOrphanedBlocks(t *testing.T) {
	drawing := &MockDrawing{
		blocks: map[string]*MockBlock{
			"BLOCK1": {
				name: "BLOCK1",
				entities: []EntityInterface{
					&MockEntity{entityType: "INSERT", blockName: "BLOCK2"},
				},
			},
			"BLOCK2": {
				name:     "BLOCK2",
				entities: []EntityInterface{},
			},
		},
	}

	detector := NewBlockCycleDetector(drawing)
	detector.DetectCycles()

	errors := detector.detectOrphanedBlocks()
	if len(errors) != 0 {
		t.Errorf("Expected 0 orphaned block errors, got %d", len(errors))
	}
}
