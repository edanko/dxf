package entity

import (
	"fmt"
	"strings"
	"sync"

	"github.com/edanko/dxf/lldxf"
)

// EntityLoader builds one typed entity from its group-code tags. Loaders are
// registered per DXF type name and consulted by the load dispatch, replacing a
// central switch; adding a type never edits the root dispatcher.
type EntityLoader func(tags lldxf.Tags) (Entity, error)

var (
	registryMu       sync.RWMutex
	entityLoaders    = map[string]EntityLoader{}
	defaultLoader    EntityLoader
	untypedTypeCount int
)

// Register associates a DXF type name with its loader. Registration is
// case-insensitive; re-registering replaces the previous loader.
func Register(dxfType string, loader EntityLoader) {
	registryMu.Lock()
	defer registryMu.Unlock()
	entityLoaders[strings.ToUpper(dxfType)] = loader
}

// SetDefaultLoader registers the fallback used for unregistered types (the
// raw-tag generic entity). It must never lose tags.
func SetDefaultLoader(loader EntityLoader) {
	registryMu.Lock()
	defer registryMu.Unlock()
	defaultLoader = loader
}

// RegisteredTypes returns all registered DXF type names, sorted.
func RegisteredTypes() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	names := make([]string, 0, len(entityLoaders))
	for name := range entityLoaders {
		names = append(names, name)
	}
	return names
}

// IsRegistered reports whether a DXF type has a registered loader.
func IsRegistered(dxfType string) bool {
	registryMu.RLock()
	defer registryMu.RUnlock()
	_, ok := entityLoaders[strings.ToUpper(dxfType)]
	return ok
}

// LoadEntity dispatches to the registered loader for dxfType. Callers that
// have no local loader for the type register entity.NewRetainedLoader as the
// default; without a registration or default it returns an error rather than
// silently dropping the record.
func LoadEntity(dxfType string, tags lldxf.Tags) (Entity, error) {
	registryMu.RLock()
	loader, ok := entityLoaders[strings.ToUpper(dxfType)]
	def := defaultLoader
	registryMu.RUnlock()
	if !ok {
		if def == nil {
			return nil, fmt.Errorf("no loader registered for DXF type %q", dxfType)
		}
		return def(tags)
	}
	return loader(tags)
}
