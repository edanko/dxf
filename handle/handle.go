// Handler interface
package handle

import (
	"fmt"
	"maps"
	"sync"
)

// Handler is interface for handle (code 5, 105, etc).
//
//	5: Entity handle; text string of up to 16 hexadecimal digits (fixed)
//	105: Object handle for DIMVAR symbol table entry
//	320-329: Arbitrary object handles; handle values that are taken "as is." They are not translated during INSERT and XREF operations
//	330-339: Soft-pointer handle; arbitrary soft pointers to other objects within same DXF file or drawing. Translated during INSERT and XREF operations
//	340-349: Hard-pointer handle; arbitrary hard pointers to other objects within same DXF file or drawing. Translated during INSERT and XREF operations
//	350-359: Soft-owner handle; arbitrary soft ownership links to other objects within same DXF file or drawing. Translated during INSERT and XREF operations
//	360-369: Hard-owner handle; arbitrary hard ownership links to other objects within same DXF file or drawing. Translated during INSERT and XREF operations
//	390-399: String representing handle value of the PlotStyleName object, basically a hard pointer, but has a different range to make backward compatibility easier to deal with.
//	Stored and moved around as an Object ID (a handle in DXF files) and a special type in AutoLISP. Custom non-entity objects may use the full range,
//	but entity classes only use 391-399 DXF group codes in their representation, for the same reason as the Lineweight range above.
type Handler interface {
	Handle() string
	SetHandle(*HandleGenerator)
}

// HandleGenerator is a type that generates unique handles for DXF entities.
type HandleGenerator struct {
	handle int
}

func NewHandleGenerator() *HandleGenerator {
	return &HandleGenerator{
		handle: 1,
	}
}

// String returns a string representation of the current handle value.
func (hg *HandleGenerator) String() string {
	return fmt.Sprintf("%X", hg.handle)
}

// Value returns the current handle counter value.
func (hg *HandleGenerator) Value() int {
	return hg.handle
}

// ObserveValue advances the generator past an existing integer handle value so
// new handles cannot collide with it.
func (hg *HandleGenerator) ObserveValue(v int) {
	if v >= hg.handle {
		hg.handle = v + 1
	}
}

// Next generates the next handle in the sequence, and increments the handle counter.
func (hg *HandleGenerator) Next() string {
	nextHandle := hg.handle
	hg.handle++
	return fmt.Sprintf("%X", nextHandle)
}

// Observe advances the generator past an existing handle so new handles cannot
// collide with handles loaded from a file.
func (hg *HandleGenerator) Observe(h string) {
	var value int
	if _, err := fmt.Sscanf(h, "%X", &value); err != nil {
		return
	}
	if value >= hg.handle {
		hg.handle = value + 1
	}
}

// HandleManager provides centralized handle management for a DXF document.
// It tracks all allocated handles, provides lookup by handle, and manages
// handle persistence across save/load operations.
type HandleManager struct {
	mu              sync.RWMutex
	generator       *HandleGenerator
	handleToObj     map[string]any
	objToHandle     map[any]string
	maxHandle       int
	reservedHandles map[string]bool
}

// NewHandleManager creates a new HandleManager with default settings.
func NewHandleManager() *HandleManager {
	return &HandleManager{
		generator:       NewHandleGenerator(),
		handleToObj:     make(map[string]any),
		objToHandle:     make(map[any]string),
		maxHandle:       0,
		reservedHandles: make(map[string]bool),
	}
}

// NewHandleManagerWithSeed creates a HandleManager starting from a specific handle value.
func NewHandleManagerWithSeed(seed int) *HandleManager {
	hm := &HandleManager{
		generator:       &HandleGenerator{handle: seed},
		handleToObj:     make(map[string]any),
		objToHandle:     make(map[any]string),
		maxHandle:       seed - 1,
		reservedHandles: make(map[string]bool),
	}
	return hm
}

// Register registers an object with the handle manager and assigns a handle if needed.
// Returns the assigned handle.
func (hm *HandleManager) Register(obj any) string {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	// Check if object already has a handle
	if existing, ok := hm.objToHandle[obj]; ok {
		return existing
	}

	// Generate and assign new handle
	handle := hm.generator.Next()
	hm.handleToObj[handle] = obj
	hm.objToHandle[obj] = handle

	// Track max handle for optimization
	if hm.maxHandle < hm.generator.handle-1 {
		hm.maxHandle = hm.generator.handle - 1
	}

	return handle
}

// RegisterWithHandle registers an object with a specific handle.
// Returns an error if the handle is already in use.
func (hm *HandleManager) RegisterWithHandle(obj any, handle string) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	// Check if object already has a handle
	if existing, ok := hm.objToHandle[obj]; ok {
		if existing == handle {
			return nil // Same handle, no-op
		}
		return fmt.Errorf("object already registered with handle %s", existing)
	}

	// Check if handle is already in use
	if _, exists := hm.handleToObj[handle]; exists {
		return fmt.Errorf("handle %s is already in use", handle)
	}

	// Check if handle is reserved
	if hm.reservedHandles[handle] {
		return fmt.Errorf("handle %s is reserved", handle)
	}

	// Parse handle value to update generator if necessary
	var handleNum int
	if _, err := fmt.Sscanf(handle, "%X", &handleNum); err == nil {
		if handleNum >= hm.generator.handle {
			hm.generator.handle = handleNum + 1
		}
		if handleNum > hm.maxHandle {
			hm.maxHandle = handleNum
		}
	}

	hm.handleToObj[handle] = obj
	hm.objToHandle[obj] = handle

	return nil
}

// GetObject returns the object associated with a handle, or nil if not found.
func (hm *HandleManager) GetObject(handle string) any {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return hm.handleToObj[handle]
}

// GetHandle returns the handle associated with an object, or empty string if not found.
func (hm *HandleManager) GetHandle(obj any) string {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return hm.objToHandle[obj]
}

// Unregister removes an object from the handle manager.
func (hm *HandleManager) Unregister(obj any) bool {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	handle, ok := hm.objToHandle[obj]
	if !ok {
		return false
	}

	delete(hm.objToHandle, obj)
	delete(hm.handleToObj, handle)
	return true
}

// UnregisterByHandle removes an object by its handle.
func (hm *HandleManager) UnregisterByHandle(handle string) bool {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	obj, ok := hm.handleToObj[handle]
	if !ok {
		return false
	}

	delete(hm.handleToObj, handle)
	delete(hm.objToHandle, obj)
	return true
}

// ReserveHandle reserves a handle so it won't be automatically assigned.
// This is useful for handling hard-coded handles in DXF files.
func (hm *HandleManager) ReserveHandle(handle string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.reservedHandles[handle] = true
}

// UnreserveHandle removes a handle reservation.
func (hm *HandleManager) UnreserveHandle(handle string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	delete(hm.reservedHandles, handle)
}

// IsReserved returns true if a handle is reserved.
func (hm *HandleManager) IsReserved(handle string) bool {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return hm.reservedHandles[handle]
}

// IsInUse returns true if a handle is currently assigned to an object.
func (hm *HandleManager) IsInUse(handle string) bool {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	_, exists := hm.handleToObj[handle]
	return exists
}

// Count returns the number of registered objects.
func (hm *HandleManager) Count() int {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return len(hm.handleToObj)
}

// MaxHandle returns the maximum handle value that has been allocated.
func (hm *HandleManager) MaxHandle() int {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return hm.maxHandle
}

// NextHandle returns the next handle that would be generated.
func (hm *HandleManager) NextHandle() string {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return fmt.Sprintf("%X", hm.generator.handle)
}

// HandleExists checks if a handle exists in the manager (either assigned or reserved).
func (hm *HandleManager) HandleExists(handle string) bool {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	_, exists := hm.handleToObj[handle]
	return exists || hm.reservedHandles[handle]
}

// Clear removes all registered objects and resets the manager.
func (hm *HandleManager) Clear() {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.generator = NewHandleGenerator()
	hm.handleToObj = make(map[string]any)
	hm.objToHandle = make(map[any]string)
	hm.maxHandle = 0
	hm.reservedHandles = make(map[string]bool)
}

// ClearObjects removes all registered objects but keeps reservations.
func (hm *HandleManager) ClearObjects() {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.handleToObj = make(map[string]any)
	hm.objToHandle = make(map[any]string)
}

// Clone creates a copy of the handle manager with the same state.
// Note: Objects themselves are not cloned, only the handle mappings.
func (hm *HandleManager) Clone() *HandleManager {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	clone := &HandleManager{
		generator:       &HandleGenerator{handle: hm.generator.handle},
		handleToObj:     make(map[string]any),
		objToHandle:     make(map[any]string),
		maxHandle:       hm.maxHandle,
		reservedHandles: make(map[string]bool),
	}

	// Copy mappings
	maps.Copy(clone.handleToObj, hm.handleToObj)
	maps.Copy(clone.objToHandle, hm.objToHandle)
	maps.Copy(clone.reservedHandles, hm.reservedHandles)

	return clone
}

// Handles returns a slice of all currently allocated handles.
func (hm *HandleManager) Handles() []string {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	handles := make([]string, 0, len(hm.handleToObj))
	for handle := range hm.handleToObj {
		handles = append(handles, handle)
	}
	return handles
}

// Objects returns a slice of all registered objects.
func (hm *HandleManager) Objects() []any {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	objects := make([]any, 0, len(hm.objToHandle))
	for obj := range hm.objToHandle {
		objects = append(objects, obj)
	}
	return objects
}
