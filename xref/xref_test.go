package xref

import (
	"os"
	"path/filepath"
	"testing"
)

func TestXRef(t *testing.T) {
	t.Run("Create XRef", func(t *testing.T) {
		xref := NewXRef("/path/to/file.dwg", "block1", XRefTypeOverlay)

		if xref.FilePath() != "/path/to/file.dwg" {
			t.Errorf("Expected filepath /path/to/file.dwg, got %s", xref.FilePath())
		}

		if xref.BlockName() != "block1" {
			t.Errorf("Expected block name block1, got %s", xref.BlockName())
		}

		if xref.XRefType() != XRefTypeOverlay {
			t.Errorf("Expected XRefTypeOverlay, got %v", xref.XRefType())
		}

		if xref.Loaded() {
			t.Error("XRef should not be loaded initially")
		}

		if !xref.IsOverlay() {
			t.Error("IsOverlay should return true for overlay type")
		}

		if xref.IsAttachment() {
			t.Error("IsAttachment should return false for overlay type")
		}
	})

	t.Run("Attachment XRef", func(t *testing.T) {
		xref := NewXRef("/path/to/file.dwg", "block2", XRefTypeAttachment)

		if !xref.IsAttachment() {
			t.Error("IsAttachment should return true for attachment type")
		}

		if xref.IsOverlay() {
			t.Error("IsOverlay should return false for attachment type")
		}
	})

	t.Run("SetLoaded", func(t *testing.T) {
		xref := NewXRef("/path/to/file.dwg", "block1", XRefTypeOverlay)

		xref.SetLoaded(true)
		if !xref.Loaded() {
			t.Error("XRef should be loaded after SetLoaded(true)")
		}

		xref.SetLoaded(false)
		if xref.Loaded() {
			t.Error("XRef should not be loaded after SetLoaded(false)")
		}
	})

	t.Run("SetPathType", func(t *testing.T) {
		xref := NewXRef("file.dwg", "block1", XRefTypeOverlay)

		if xref.PathType() != PathTypeAbsolute {
			// NewXRef sets PathTypeAbsolute by default, but if file is relative it should detect
			t.Logf("Initial path type: %v", xref.PathType())
		}

		xref.SetPathType(PathTypeRelative)
		if xref.PathType() != PathTypeRelative {
			t.Errorf("Expected PathTypeRelative, got %v", xref.PathType())
		}
	})

	t.Run("ResolvePath", func(t *testing.T) {
		xref := NewXRef("subdir/file.dwg", "block1", XRefTypeOverlay)
		xref.SetPathType(PathTypeRelative)

		basePath := "/base"
		resolved := xref.ResolvePath(basePath)

		expected := filepath.Join(basePath, "subdir/file.dwg")
		if resolved != expected {
			t.Errorf("Expected %s, got %s", expected, resolved)
		}
	})

	t.Run("ResolvePath absolute", func(t *testing.T) {
		xref := NewXRef("/absolute/path/file.dwg", "block1", XRefTypeOverlay)
		xref.SetPathType(PathTypeAbsolute)

		basePath := "/base"
		resolved := xref.ResolvePath(basePath)

		if resolved != "/absolute/path/file.dwg" {
			t.Errorf("Expected absolute path to be returned as-is, got %s", resolved)
		}
	})

	t.Run("Drawing accessors", func(t *testing.T) {
		xref := NewXRef("/path/to/file.dwg", "block1", XRefTypeOverlay)

		if xref.Drawing() != nil {
			t.Error("Drawing should be nil initially")
		}

		mockDrawing := &mockDrawing{name: "test"}
		xref.SetDrawing(mockDrawing)

		if xref.Drawing() != mockDrawing {
			t.Error("Drawing should return the set value")
		}
	})
}

func TestXRefManager(t *testing.T) {
	t.Run("Add XRef", func(t *testing.T) {
		manager := NewXRefManager("/base")
		xref := NewXRef("/path/file.dwg", "block1", XRefTypeOverlay)

		err := manager.Add(xref)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if manager.Count() != 1 {
			t.Errorf("Expected count 1, got %d", manager.Count())
		}
	})

	t.Run("Add duplicate XRef", func(t *testing.T) {
		manager := NewXRefManager("/base")
		xref1 := NewXRef("/path/file1.dwg", "block1", XRefTypeOverlay)
		xref2 := NewXRef("/path/file2.dwg", "block1", XRefTypeAttachment)

		err := manager.Add(xref1)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		err = manager.Add(xref2)
		if err == nil {
			t.Error("Expected error for duplicate block name")
		}
	})

	t.Run("Get XRef", func(t *testing.T) {
		manager := NewXRefManager("/base")
		xref := NewXRef("/path/file.dwg", "block1", XRefTypeOverlay)
		manager.Add(xref)

		retrieved, exists := manager.Get("block1")
		if !exists {
			t.Error("XRef should exist")
		}

		if retrieved != xref {
			t.Error("Retrieved XRef should be the same as added")
		}
	})

	t.Run("Get non-existent XRef", func(t *testing.T) {
		manager := NewXRefManager("/base")

		_, exists := manager.Get("nonexistent")
		if exists {
			t.Error("Non-existent XRef should not be found")
		}
	})

	t.Run("Remove XRef", func(t *testing.T) {
		manager := NewXRefManager("/base")
		xref := NewXRef("/path/file.dwg", "block1", XRefTypeOverlay)
		manager.Add(xref)

		err := manager.Remove("block1")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if manager.Count() != 0 {
			t.Errorf("Expected count 0, got %d", manager.Count())
		}
	})

	t.Run("Remove non-existent XRef", func(t *testing.T) {
		manager := NewXRefManager("/base")

		err := manager.Remove("nonexistent")
		if err == nil {
			t.Error("Expected error for non-existent XRef")
		}
	})

	t.Run("Exists", func(t *testing.T) {
		manager := NewXRefManager("/base")
		xref := NewXRef("/path/file.dwg", "block1", XRefTypeOverlay)
		manager.Add(xref)

		if !manager.Exists("block1") {
			t.Error("block1 should exist")
		}

		if manager.Exists("block2") {
			t.Error("block2 should not exist")
		}
	})

	t.Run("List", func(t *testing.T) {
		manager := NewXRefManager("/base")
		manager.Add(NewXRef("/path/file1.dwg", "block1", XRefTypeOverlay))
		manager.Add(NewXRef("/path/file2.dwg", "block2", XRefTypeAttachment))

		list := manager.List()
		if len(list) != 2 {
			t.Errorf("Expected list length 2, got %d", len(list))
		}
	})

	t.Run("GetAll", func(t *testing.T) {
		manager := NewXRefManager("/base")
		xref1 := NewXRef("/path/file1.dwg", "block1", XRefTypeOverlay)
		xref2 := NewXRef("/path/file2.dwg", "block2", XRefTypeAttachment)
		manager.Add(xref1)
		manager.Add(xref2)

		all := manager.GetAll()
		if len(all) != 2 {
			t.Errorf("Expected 2 xrefs, got %d", len(all))
		}
	})

	t.Run("LoadedCount", func(t *testing.T) {
		manager := NewXRefManager("/base")
		xref1 := NewXRef("/path/file1.dwg", "block1", XRefTypeOverlay)
		xref2 := NewXRef("/path/file2.dwg", "block2", XRefTypeAttachment)
		manager.Add(xref1)
		manager.Add(xref2)

		xref1.SetLoaded(true)

		if manager.LoadedCount() != 1 {
			t.Errorf("Expected loaded count 1, got %d", manager.LoadedCount())
		}
	})

	t.Run("AddXRef convenience", func(t *testing.T) {
		manager := NewXRefManager("/base")

		err := manager.AddXRef("/path/file.dwg", "block1", XRefTypeAttachment)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		xref, exists := manager.Get("block1")
		if !exists {
			t.Error("XRef should exist")
		}

		if !xref.IsAttachment() {
			t.Error("XRef should be attachment type")
		}
	})

	t.Run("Base path conversion", func(t *testing.T) {
		manager := NewXRefManager("/base")

		manager.SetBasePath("/new/base")
		if manager.GetBasePath() != "/new/base" {
			t.Errorf("Expected /new/base, got %s", manager.GetBasePath())
		}

		absPath := manager.ConvertToAbsolute("subdir/file.dwg")
		expected := filepath.Join("/new/base", "subdir/file.dwg")
		if absPath != expected {
			t.Errorf("Expected %s, got %s", expected, absPath)
		}

		relPath := manager.ConvertToRelative("/new/base/subdir/file.dwg")
		if relPath != "subdir/file.dwg" {
			t.Errorf("Expected subdir/file.dwg, got %s", relPath)
		}
	})
}

func TestXRefExists(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "test_*.dwg")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	if !XRefExists(tmpFile.Name()) {
		t.Error("XRefExists should return true for existing file")
	}

	if XRefExists("/nonexistent/path/file.dwg") {
		t.Error("XRefExists should return false for non-existing file")
	}
}

func TestLoader(t *testing.T) {
	t.Run("Create Loader", func(t *testing.T) {
		loader := NewLoader("/base")

		if loader.GetBasePath() != "/base" {
			t.Errorf("Expected /base, got %s", loader.GetBasePath())
		}

		if loader.GetMaxCacheSize() != 10 {
			t.Errorf("Expected default max cache size 10, got %d", loader.GetMaxCacheSize())
		}
	})

	t.Run("Cache management", func(t *testing.T) {
		loader := NewLoader("/base")

		loader.SetMaxCacheSize(3)
		if loader.GetMaxCacheSize() != 3 {
			t.Errorf("Expected max cache size 3, got %d", loader.GetMaxCacheSize())
		}

		if loader.CacheSize() != 0 {
			t.Errorf("Expected cache size 0, got %d", loader.CacheSize())
		}

		loader.ClearCache()
		if loader.CacheSize() != 0 {
			t.Errorf("Expected cache size 0 after clear, got %d", loader.CacheSize())
		}
	})

	t.Run("Base path", func(t *testing.T) {
		loader := NewLoader("/initial")

		loader.SetBasePath("/updated")
		if loader.GetBasePath() != "/updated" {
			t.Errorf("Expected /updated, got %s", loader.GetBasePath())
		}
	})

	t.Run("Load already loaded XRef", func(t *testing.T) {
		loader := NewLoader("/base")
		xref := NewXRef("/path/file.dwg", "block1", XRefTypeOverlay)
		xref.SetLoaded(true)

		loadCalled := false
		loadFunc := func(path string) (any, error) {
			loadCalled = true
			return nil, nil
		}

		err := loader.Load(xref, loadFunc)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if loadCalled {
			t.Error("Load function should not be called for already loaded XRef")
		}
	})

	t.Run("Load with cache miss", func(t *testing.T) {
		loader := NewLoader("/base")
		xref := NewXRef("/path/file.dwg", "block1", XRefTypeOverlay)

		mockDrawing := &mockDrawing{name: "loaded"}
		loadFunc := func(path string) (any, error) {
			return mockDrawing, nil
		}

		err := loader.Load(xref, loadFunc)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !xref.Loaded() {
			t.Error("XRef should be loaded after Load")
		}

		if xref.Drawing() != mockDrawing {
			t.Error("XRef should have the loaded drawing")
		}

		if loader.CacheSize() != 1 {
			t.Errorf("Expected cache size 1, got %d", loader.CacheSize())
		}
	})

	t.Run("GetCachedDrawing", func(t *testing.T) {
		loader := NewLoader("/base")
		xref := NewXRef("/path/file.dwg", "block1", XRefTypeOverlay)

		mockDrawing := &mockDrawing{name: "cached"}
		loadFunc := func(path string) (any, error) {
			return mockDrawing, nil
		}

		loader.Load(xref, loadFunc)

		// Try to get from cache
		cached, exists := loader.GetCachedDrawing("/path/file.dwg")
		if !exists {
			t.Error("Drawing should be in cache")
		}

		if cached != mockDrawing {
			t.Error("Cached drawing should match original")
		}
	})
}

func TestLoaderCacheEviction(t *testing.T) {
	t.Run("Cache eviction when full", func(t *testing.T) {
		loader := NewLoader("/base")
		loader.SetMaxCacheSize(2)

		// Load multiple xrefs
		for i := range 3 {
			xref := NewXRef("/path/file"+string(rune('1'+i))+".dwg", "block"+string(rune('1'+i)), XRefTypeOverlay)
			mockDrawing := &mockDrawing{name: "drawing" + string(rune('1'+i))}
			loadFunc := func(path string) (any, error) {
				return mockDrawing, nil
			}
			loader.Load(xref, loadFunc)
		}

		// Cache should be at max size (oldest evicted)
		if loader.CacheSize() != 2 {
			t.Errorf("Expected cache size 2, got %d", loader.CacheSize())
		}
	})
}

// mockDrawing is a simple mock for testing
type mockDrawing struct {
	name string
}
