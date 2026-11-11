package handle

import (
	"sync"
	"testing"
)

func TestHandleManager(t *testing.T) {
	t.Run("Basic registration", func(t *testing.T) {
		hm := NewHandleManager()

		obj := &testObject{name: "test1"}
		handle := hm.Register(obj)

		if handle == "" {
			t.Error("Expected non-empty handle")
		}

		if hm.GetObject(handle) != obj {
			t.Error("GetObject should return registered object")
		}

		if hm.GetHandle(obj) != handle {
			t.Error("GetHandle should return object's handle")
		}
	})

	t.Run("Multiple objects", func(t *testing.T) {
		hm := NewHandleManager()

		obj1 := &testObject{name: "test1"}
		obj2 := &testObject{name: "test2"}
		obj3 := &testObject{name: "test3"}

		handle1 := hm.Register(obj1)
		handle2 := hm.Register(obj2)
		handle3 := hm.Register(obj3)

		if handle1 == handle2 || handle2 == handle3 || handle1 == handle3 {
			t.Error("Each object should get a unique handle")
		}

		if hm.Count() != 3 {
			t.Errorf("Expected count 3, got %d", hm.Count())
		}
	})

	t.Run("RegisterWithHandle", func(t *testing.T) {
		hm := NewHandleManager()

		obj := &testObject{name: "test"}
		err := hm.RegisterWithHandle(obj, "1F")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if hm.GetHandle(obj) != "1F" {
			t.Error("Handle should be 1F")
		}

		// Try to register same handle again
		obj2 := &testObject{name: "test2"}
		err = hm.RegisterWithHandle(obj2, "1F")
		if err == nil {
			t.Error("Expected error for duplicate handle")
		}
	})

	t.Run("Unregister", func(t *testing.T) {
		hm := NewHandleManager()

		obj := &testObject{name: "test"}
		hm.Register(obj)

		if !hm.Unregister(obj) {
			t.Error("Unregister should return true for registered object")
		}

		if hm.Unregister(obj) {
			t.Error("Unregister should return false for unregistered object")
		}

		if hm.Count() != 0 {
			t.Errorf("Expected count 0, got %d", hm.Count())
		}
	})

	t.Run("Reserved handles", func(t *testing.T) {
		hm := NewHandleManager()

		hm.ReserveHandle("FF")
		if !hm.IsReserved("FF") {
			t.Error("FF should be reserved")
		}

		hm.UnreserveHandle("FF")
		if hm.IsReserved("FF") {
			t.Error("FF should not be reserved after UnreserveHandle")
		}
	})

	t.Run("Concurrency safety", func(t *testing.T) {
		hm := NewHandleManager()
		var wg sync.WaitGroup
		numGoroutines := 100

		for i := range numGoroutines {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				obj := &testObject{name: "test"}
				hm.Register(obj)
			}(i)
		}

		wg.Wait()

		if hm.Count() != numGoroutines {
			t.Errorf("Expected count %d, got %d", numGoroutines, hm.Count())
		}
	})

	t.Run("Handle generation", func(t *testing.T) {
		hm := NewHandleManager()

		// First handle should be 1
		obj := &testObject{name: "test"}
		hm.Register(obj)
		if hm.NextHandle() != "2" {
			t.Errorf("Expected next handle 2, got %s", hm.NextHandle())
		}
	})

	t.Run("Clear", func(t *testing.T) {
		hm := NewHandleManager()

		obj := &testObject{name: "test"}
		hm.Register(obj)
		hm.ReserveHandle("AA")

		hm.Clear()

		if hm.Count() != 0 {
			t.Errorf("Expected count 0 after Clear, got %d", hm.Count())
		}

		if hm.IsReserved("AA") {
			t.Error("AA should not be reserved after Clear")
		}
	})

	t.Run("Clone", func(t *testing.T) {
		hm := NewHandleManager()

		obj := &testObject{name: "test"}
		hm.Register(obj)
		hm.ReserveHandle("BB")

		clone := hm.Clone()

		if clone.Count() != 1 {
			t.Errorf("Clone should have count 1, got %d", clone.Count())
		}

		if !clone.IsReserved("BB") {
			t.Error("Clone should have reserved BB")
		}
	})

	t.Run("Handles and Objects", func(t *testing.T) {
		hm := NewHandleManager()

		obj1 := &testObject{name: "test1"}
		obj2 := &testObject{name: "test2"}
		hm.Register(obj1)
		hm.Register(obj2)

		handles := hm.Handles()
		if len(handles) != 2 {
			t.Errorf("Expected 2 handles, got %d", len(handles))
		}

		objects := hm.Objects()
		if len(objects) != 2 {
			t.Errorf("Expected 2 objects, got %d", len(objects))
		}
	})

	t.Run("Duplicate registration returns same handle", func(t *testing.T) {
		hm := NewHandleManager()

		obj := &testObject{name: "test"}
		handle1 := hm.Register(obj)
		handle2 := hm.Register(obj)

		if handle1 != handle2 {
			t.Error("Registering same object twice should return same handle")
		}

		if hm.Count() != 1 {
			t.Errorf("Expected count 1, got %d", hm.Count())
		}
	})
}

func TestHandleManagerWithSeed(t *testing.T) {
	t.Run("Start from specific seed", func(t *testing.T) {
		hm := NewHandleManagerWithSeed(100)

		obj := &testObject{name: "test"}
		handle := hm.Register(obj)

		// 100 in hex is 64
		if handle != "64" {
			t.Errorf("Expected handle 64, got %s", handle)
		}

		if hm.NextHandle() != "65" {
			t.Errorf("Expected next handle 65, got %s", hm.NextHandle())
		}
	})
}

func TestHandleExists(t *testing.T) {
	hm := NewHandleManager()

	// Should not exist initially
	if hm.HandleExists("1") {
		t.Error("Handle 1 should not exist")
	}

	// Register an object
	obj := &testObject{name: "test"}
	hm.RegisterWithHandle(obj, "1")

	if !hm.HandleExists("1") {
		t.Error("Handle 1 should exist after registration")
	}

	// Reserve a handle
	hm.ReserveHandle("2")
	if !hm.HandleExists("2") {
		t.Error("Handle 2 should exist after reservation")
	}
}

// testObject is a simple object for testing
type testObject struct {
	name string
}
