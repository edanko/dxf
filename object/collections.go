package object

// ObjectCollection is the typed view of a DICTIONARY that stores a named
// collection of objects, matching ezdxf's ObjectCollection model. The
// underlying record remains a DICTIONARY on the wire; only the Go type and
// the convenience accessors change.
type ObjectCollection struct {
	*Dictionary
	collectionType string
	objectType     string
}

// MLeaderStyleCollection is the ACAD_MLEADERSTYLE dictionary.
type MLeaderStyleCollection struct{ *ObjectCollection }

// MLineStyleCollection is the ACAD_MLINESTYLE dictionary.
type MLineStyleCollection struct{ *ObjectCollection }

// NewObjectCollection wraps an existing dictionary as a typed collection.
// The dictionary handle, owner, entries, and retained tags are unchanged.
func NewObjectCollection(d *Dictionary, collectionType, objectType string) *ObjectCollection {
	return &ObjectCollection{
		Dictionary:     d,
		collectionType: collectionType,
		objectType:     objectType,
	}
}

// WrapCollection returns the concrete typed collection for a collection type.
func WrapCollection(d *Dictionary, collectionType string) Object {
	c := NewObjectCollection(d, collectionType, collectionObjectType(collectionType))
	switch collectionType {
	case "MLEADERSTYLECOLLECTION":
		return &MLeaderStyleCollection{c}
	case "MLINESTYLECOLLECTION":
		return &MLineStyleCollection{c}
	default:
		return c
	}
}

// CollectionForDictionaryKey maps a root-dictionary key to the collection and
// member object types it represents.
func CollectionForDictionaryKey(key string) (collectionType, objectType string, ok bool) {
	switch key {
	case "ACAD_MLEADERSTYLE":
		return "MLEADERSTYLECOLLECTION", "MLEADERSTYLE", true
	case "ACAD_MLINESTYLE":
		return "MLINESTYLECOLLECTION", "MLINESTYLE", true
	default:
		return "", "", false
	}
}

func collectionObjectType(collectionType string) string {
	switch collectionType {
	case "MLEADERSTYLECOLLECTION":
		return "MLEADERSTYLE"
	case "MLINESTYLECOLLECTION":
		return "MLINESTYLE"
	default:
		return ""
	}
}

// DXFType returns the conceptual collection type.
func (c *ObjectCollection) DXFType() string { return c.collectionType }

// CollectionType returns the conceptual collection type.
func (c *ObjectCollection) CollectionType() string { return c.collectionType }

// ObjectType returns the type of the collection members.
func (c *ObjectCollection) ObjectType() string { return c.objectType }

// MemberKeys returns entry keys in source order.
func (c *ObjectCollection) MemberKeys() []string { return c.Dictionary.Keys() }

// MemberHandles returns member object handles in entry order.
func (c *ObjectCollection) MemberHandles() []string {
	keys := c.Dictionary.Keys()
	out := make([]string, 0, len(keys))
	for _, key := range keys {
		if entry, ok := c.Dictionary.GetEntry(key); ok {
			out = append(out, entry.Value)
		}
	}
	return out
}

// MemberHandle returns the member handle stored under key.
func (c *ObjectCollection) MemberHandle(key string) (string, bool) {
	entry, ok := c.Dictionary.GetEntry(key)
	if !ok {
		return "", false
	}
	return entry.Value, true
}

// ResolveMembers resolves every member handle through lookup. Members that do
// not resolve are omitted; callers that need diagnostics can use
// MemberHandles and resolve individually.
func (c *ObjectCollection) ResolveMembers(lookup func(string) (Object, bool)) map[string]Object {
	out := make(map[string]Object)
	for _, key := range c.Dictionary.Keys() {
		entry, ok := c.Dictionary.GetEntry(key)
		if !ok {
			continue
		}
		if obj, ok := lookup(entry.Value); ok {
			out[key] = obj
		}
	}
	return out
}
