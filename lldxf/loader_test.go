package lldxf

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoaderAllTestFiles(t *testing.T) {
	testDir := filepath.Join("..", "testdata")
	files, err := os.ReadDir(testDir)
	if err != nil {
		t.Fatalf("Failed to read test directory: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filename := file.Name()
		if filepath.Ext(filename) != ".dxf" {
			continue
		}

		t.Run(filename, func(t *testing.T) {
			testFile := filepath.Join(testDir, filename)
			f, err := os.Open(testFile)
			if err != nil {
				t.Fatalf("Failed to open file %s: %v", filename, err)
			}
			defer f.Close()

			tagger := NewASCIITagger(f)

			var allTags Tags
			for tagger.Next() {
				tag := tagger.Tag()
				if tag != nil {
					allTags = append(allTags, tag)
				}
			}

			if err := tagger.Err(); err != nil {
				t.Fatalf("Error reading %s: %v", filename, err)
			}

			if len(allTags) == 0 {
				t.Errorf("No tags found in %s", filename)
				return
			}

			loader := NewLoader(nil)
			if err := loader.Load(allTags); err != nil {
				t.Fatalf("Failed to load %s: %v", filename, err)
			}

			sections := loader.GetSections()
			if len(sections) == 0 {
				t.Errorf("No sections found in %s", filename)
				return
			}

			entityCount := loader.GetEntityCount()
			t.Logf("File %s: %d entities, %d sections", filename, entityCount, len(sections))
		})
	}
}

func TestLoaderEntityExtraction(t *testing.T) {
	testFiles := []string{
		"mypoint.dxf",
		"arc.dxf",
		"torus.dxf",
	}

	for _, filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			testFile := filepath.Join("..", "testdata", filename)
			f, err := os.Open(testFile)
			if err != nil {
				t.Fatalf("Failed to open file: %v", err)
			}
			defer f.Close()

			tagger := NewASCIITagger(f)

			var allTags Tags
			for tagger.Next() {
				tag := tagger.Tag()
				if tag != nil {
					allTags = append(allTags, tag)
				}
			}

			loader := NewLoader(DefaultLoaderOptions())
			if err := loader.Load(allTags); err != nil {
				t.Fatalf("Failed to load: %v", err)
			}

			entities := loader.GetEntities()
			t.Logf("Found %d entities", len(entities))

			entityType := ""
			for _, e := range entities {
				if etype, ok := e.GetFirstValue(0); ok {
					entityType = etype.(string)
					break
				}
			}

			foundByType := loader.FindEntitiesByType(entityType)
			if len(foundByType) == 0 {
				t.Errorf("Failed to find entities by type %s", entityType)
			}

			layers := make(map[string]bool)
			for _, e := range entities {
				if layer, ok := e.GetFirstValue(8); ok {
					layers[layer.(string)] = true
				}
			}

			for layer := range layers {
				foundByLayer := loader.FindEntitiesByLayer(layer)
				if len(foundByLayer) == 0 {
					t.Errorf("Failed to find entities on layer %s", layer)
				}
			}
		})
	}
}

func TestLoaderOptions(t *testing.T) {
	testFile := filepath.Join("..", "testdata", "mypoint.dxf")
	f, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer f.Close()

	tagger := NewASCIITagger(f)

	var allTags Tags
	for tagger.Next() {
		tag := tagger.Tag()
		if tag != nil {
			allTags = append(allTags, tag)
		}
	}

	opts := DefaultLoaderOptions()
	loader := NewLoader(opts)
	if err := loader.Load(allTags); err != nil {
		t.Fatalf("Failed to load with default options: %v", err)
	}

	if err := loader.Validate(); err != nil {
		t.Errorf("Validation failed: %v", err)
	}

	info := loader.Info()
	t.Logf("Loader info: %+v", info)

	if err := loader.Reload(nil); err != nil {
		t.Errorf("Reload failed: %v", err)
	}
}

func TestLoaderSections(t *testing.T) {
	testFile := filepath.Join("..", "testdata", "torus.dxf")
	f, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer f.Close()

	tagger := NewASCIITagger(f)

	var allTags Tags
	for tagger.Next() {
		tag := tagger.Tag()
		if tag != nil {
			allTags = append(allTags, tag)
		}
	}

	loader := NewLoader(nil)
	if err := loader.Load(allTags); err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	sectionNames := loader.GetSectionNames()
	t.Logf("Sections: %v", sectionNames)

	tableNames := loader.GetTableNames()
	t.Logf("Tables: %v", tableNames)

	entities := loader.GetSection(SectionEntities)
	if entities == nil {
		t.Error("ENTITIES section not found")
	}

	tables := loader.GetTables()
	if tables == nil {
		t.Error("No tables found")
	}

	blocks := loader.GetBlocks()
	if blocks == nil {
		t.Error("No blocks found")
	}
}

func TestLoaderTables(t *testing.T) {
	testFile := filepath.Join("..", "testdata", "torus.dxf")
	f, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer f.Close()

	tagger := NewASCIITagger(f)

	var allTags Tags
	for tagger.Next() {
		tag := tagger.Tag()
		if tag != nil {
			allTags = append(allTags, tag)
		}
	}

	loader := NewLoader(nil)
	if err := loader.Load(allTags); err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	layerTable := loader.GetTable("LAYER")
	if layerTable == nil {
		t.Log("No LAYER table found (may be valid)")
	} else {
		t.Logf("LAYER table has %d entries", len(layerTable))
		for name := range layerTable {
			t.Logf("  Layer: %s", name)
		}
	}

	ltypeTable := loader.GetTable("LTYPE")
	if ltypeTable == nil {
		t.Log("No LTYPE table found (may be valid)")
	} else {
		t.Logf("LTYPE table has %d entries", len(ltypeTable))
	}

	tableEntries := loader.GetTables()
	for tableName, entries := range tableEntries {
		t.Logf("Table %s: %d entries", tableName, len(entries))
	}
}
