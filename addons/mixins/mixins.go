// Package mixins provides attribute mixins which behave like ezdxf's
// SubscriptAttributes helper.
//
// Go does not support the Python subscript protocol on arbitrary types, so the
// same behavior is exposed through Has, Get and Set methods. The mixin owns a
// private attribute map and never modifies the entity or object it is attached
// to conceptually.
package mixins

import (
	"errors"
	"fmt"
	"sort"

	"github.com/edanko/dxf/addons"
)

// ErrUnknownAttribute is returned when an attribute has not been defined.
var ErrUnknownAttribute = errors.New("mixins: unknown attribute")

// SubscriptAttributes stores named attributes.
type SubscriptAttributes struct {
	values map[string]any
}

// NewSubscriptAttributes creates an empty attribute mixin.
func NewSubscriptAttributes() *SubscriptAttributes {
	return &SubscriptAttributes{values: make(map[string]any)}
}

// NewSubscriptAttributesFrom creates a mixin with a copy of initial.
func NewSubscriptAttributesFrom(initial map[string]any) *SubscriptAttributes {
	mixin := NewSubscriptAttributes()
	for key, value := range initial {
		mixin.values[key] = value
	}
	return mixin
}

// Has reports whether key is defined, matching Python's "__contains__".
func (s *SubscriptAttributes) Has(key string) bool {
	if s == nil {
		return false
	}
	_, ok := s.values[key]
	return ok
}

// Get returns the value for key. It mirrors Python's "__getitem__" behavior:
// an unknown key is an error rather than a zero value.
func (s *SubscriptAttributes) Get(key string) (any, error) {
	if s == nil {
		return nil, fmt.Errorf("%w: %q", ErrUnknownAttribute, key)
	}
	value, ok := s.values[key]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnknownAttribute, key)
	}
	return value, nil
}

// Set changes an existing attribute. It mirrors Python's "__setitem__"
// behavior: setting an unknown attribute is an error.
func (s *SubscriptAttributes) Set(key string, value any) error {
	if s == nil {
		return fmt.Errorf("%w: %q", ErrUnknownAttribute, key)
	}
	if _, ok := s.values[key]; !ok {
		return fmt.Errorf("%w: %q", ErrUnknownAttribute, key)
	}
	s.values[key] = value
	return nil
}

// Define adds or replaces an attribute. Unlike Set it accepts new keys.
func (s *SubscriptAttributes) Define(key string, value any) {
	if s == nil {
		return
	}
	if s.values == nil {
		s.values = make(map[string]any)
	}
	s.values[key] = value
}

// Delete removes an attribute. It returns false when the key did not exist.
func (s *SubscriptAttributes) Delete(key string) bool {
	if s == nil {
		return false
	}
	if _, ok := s.values[key]; !ok {
		return false
	}
	delete(s.values, key)
	return true
}

// Keys returns the defined attribute names in sorted order.
func (s *SubscriptAttributes) Keys() []string {
	if s == nil {
		return nil
	}
	keys := make([]string, 0, len(s.values))
	for key := range s.values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// Len returns the number of defined attributes.
func (s *SubscriptAttributes) Len() int {
	if s == nil {
		return 0
	}
	return len(s.values)
}

// Clone returns an independent copy.
func (s *SubscriptAttributes) Clone() *SubscriptAttributes {
	if s == nil {
		return NewSubscriptAttributes()
	}
	return NewSubscriptAttributesFrom(s.values)
}

// Attribute returns a typed attribute value. It is the generic equivalent of
// retrieving and casting an attribute in Python.
func Attribute[T any](s *SubscriptAttributes, key string) (T, error) {
	var zero T
	value, err := s.Get(key)
	if err != nil {
		return zero, err
	}
	typed, ok := value.(T)
	if !ok {
		return zero, fmt.Errorf("mixins: attribute %q has type %T, not %T", key, value, zero)
	}
	return typed, nil
}

// Capability registers the add-on with the addons registry.
type Capability struct{}

// Name implements addons.Capability.
func (Capability) Name() string { return "mixins" }

// Description implements addons.Capability.
func (Capability) Description() string { return "Subscript-style attribute mixins" }

func init() {
	_ = addons.Register(Capability{}.Name(), Capability{})
}
