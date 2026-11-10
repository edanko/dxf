// Package addons provides the capability registry shared by the optional
// add-on packages. Add-ons register a Capability in init; callers discover
// and invoke them by name without importing every add-on.
package addons

import (
	"fmt"
	"sort"
	"sync"
)

// Capability describes a discoverable add-on.
type Capability interface {
	Name() string
	Description() string
}

var (
	mu         sync.RWMutex
	registered = map[string]Capability{}
)

// Register adds a capability under its name. Registering the same name
// twice is an error.
func Register(name string, c Capability) error {
	if name == "" {
		return fmt.Errorf("addons: empty capability name")
	}
	if c == nil {
		return fmt.Errorf("addons: nil capability %q", name)
	}
	mu.Lock()
	defer mu.Unlock()
	if _, exists := registered[name]; exists {
		return fmt.Errorf("addons: capability %q already registered", name)
	}
	registered[name] = c
	return nil
}

// Lookup returns a registered capability.
func Lookup(name string) (Capability, bool) {
	mu.RLock()
	defer mu.RUnlock()
	c, ok := registered[name]
	return c, ok
}

// Capabilities lists registered capabilities sorted by name.
func Capabilities() []Capability {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]Capability, 0, len(registered))
	for _, c := range registered {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name() < out[j].Name() })
	return out
}
