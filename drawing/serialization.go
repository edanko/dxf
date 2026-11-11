package drawing

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/edanko/dxf/entity"
)

// JSONEnvelopeVersion is the version of the JSON document envelope emitted by
// ToJSON. FromJSON rejects any other version.
const JSONEnvelopeVersion = 1

// JSONEnvelopeFormat identifies the embedded payload encoding.
const JSONEnvelopeFormat = "dxf-ascii"

// jsonEnvelope is the versioned wrapper around a document's ASCII DXF
// payload. The summary fields are for human inspection only; the payload is
// the source of truth for FromJSON.
type jsonEnvelope struct {
	Version     int              `json:"version"`
	Format      string           `json:"format"`
	FileName    string           `json:"file_name,omitempty"`
	LayerCount  int              `json:"layer_count"`
	StyleCount  int              `json:"style_count"`
	EntityCount int              `json:"entity_count"`
	Statistics  map[string]any   `json:"statistics,omitempty"`
	Entities    []map[string]any `json:"entities,omitempty"`
	Payload     string           `json:"payload"`
}

type SerializationOptions struct {
	IncludeHandles     bool
	IncludeXData       bool
	IncludeHandlesAttr bool
	Compact            bool
}

func NewSerializationOptions() *SerializationOptions {
	return &SerializationOptions{
		IncludeHandles:     true,
		IncludeXData:       true,
		IncludeHandlesAttr: true,
		Compact:            false,
	}
}

// ToJSON serializes the drawing into a versioned JSON envelope. The envelope
// embeds the document's ASCII DXF payload plus summary metadata. An error is
// returned when the document cannot be represented as ASCII DXF, rather than
// emitting an envelope that silently loses state.
func (d *Drawing) ToJSON(opts *SerializationOptions) ([]byte, error) {
	if d == nil {
		return nil, fmt.Errorf("ToJSON: nil drawing")
	}
	if opts == nil {
		opts = NewSerializationOptions()
	}

	var payload bytes.Buffer
	if _, err := d.WriteTo(&payload); err != nil {
		return nil, fmt.Errorf("ToJSON: encode DXF payload: %w", err)
	}
	if payload.Len() == 0 {
		return nil, fmt.Errorf("ToJSON: document produced an empty DXF payload")
	}

	stats := d.GetDrawingStatistics()
	entities := d.GetAllEntities()

	entitiesJSON := make([]map[string]any, 0, len(entities))
	for _, e := range entities {
		entitiesJSON = append(entitiesJSON, entityToMap(e, opts))
	}

	env := jsonEnvelope{
		Version:     JSONEnvelopeVersion,
		Format:      JSONEnvelopeFormat,
		FileName:    d.fileName,
		LayerCount:  len(d.Layers),
		StyleCount:  len(d.Styles),
		EntityCount: len(entities),
		Statistics:  stats,
		Entities:    entitiesJSON,
		Payload:     payload.String(),
	}

	if opts.Compact {
		return json.Marshal(env)
	}
	return json.MarshalIndent(env, "", "  ")
}

// FromJSON loads a versioned JSON envelope produced by ToJSON into the
// receiver. The envelope version and payload are validated first; the payload
// is then loaded with the registered document loader. The receiver is replaced
// only after a successful load, so malformed or unsupported input leaves it
// unchanged.
func (d *Drawing) FromJSON(data []byte) error {
	if d == nil {
		return fmt.Errorf("FromJSON: nil drawing")
	}
	var env jsonEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return fmt.Errorf("FromJSON: invalid JSON: %w", err)
	}
	if env.Version != JSONEnvelopeVersion {
		return fmt.Errorf("FromJSON: unsupported envelope version %d (want %d)", env.Version, JSONEnvelopeVersion)
	}
	if env.Format != "" && env.Format != JSONEnvelopeFormat {
		return fmt.Errorf("FromJSON: unsupported payload format %q (want %q)", env.Format, JSONEnvelopeFormat)
	}
	if env.Payload == "" {
		return fmt.Errorf("FromJSON: missing payload")
	}
	if documentLoader == nil {
		return fmt.Errorf("FromJSON: document loader is not registered")
	}
	loaded, err := documentLoader(strings.NewReader(env.Payload))
	if err != nil {
		return fmt.Errorf("FromJSON: load payload: %w", err)
	}
	if loaded == nil {
		return fmt.Errorf("FromJSON: loader returned a nil drawing")
	}
	*d = *loaded
	return nil
}

func entityToMap(e entity.Entity, opts *SerializationOptions) map[string]any {
	result := make(map[string]any)
	result["type"] = e.DXFType()

	if opts.IncludeHandles && e.Handle() != "" {
		result["handle"] = e.Handle()
	}

	if layer := e.Layer(); layer != nil {
		result["layer"] = layer.Name()
	}

	return result
}

func (d *Drawing) ToBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(d)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (d *Drawing) FromBinary(data []byte) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	return dec.Decode(d)
}

func (d *Drawing) ToBase64() (string, error) {
	data, err := d.ToBinary()
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func (d *Drawing) FromBase64(s string) error {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return err
	}
	return d.FromBinary(data)
}

type EntityExport struct {
	Type    string         `json:"type"`
	Handle  string         `json:"handle,omitempty"`
	Layer   string         `json:"layer,omitempty"`
	Color   int            `json:"color,omitempty"`
	RawData map[string]any `json:"raw_data,omitempty"`
}

func ExportEntitiesToJSON(entities []entity.Entity, opts *SerializationOptions) ([]byte, error) {
	if opts == nil {
		opts = NewSerializationOptions()
	}

	result := make([]EntityExport, 0, len(entities))
	for _, e := range entities {
		exp := EntityExport{
			Type:   e.DXFType(),
			Handle: e.Handle(),
		}
		if layer := e.Layer(); layer != nil {
			exp.Layer = layer.Name()
		}
		result = append(result, exp)
	}

	if opts.Compact {
		return json.Marshal(result)
	}
	return json.MarshalIndent(result, "", "  ")
}

func ExportEntitiesToHex(entities []entity.Entity) (string, error) {
	data, err := ExportEntitiesToJSON(entities, NewSerializationOptions())
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(data), nil
}

// HashDrawing returns a stable content hash of the drawing: the ASCII DXF
// payload is serialized and digested with SHA-256.
func HashDrawing(d *Drawing) (string, error) {
	var payload bytes.Buffer
	if _, err := d.WriteTo(&payload); err != nil {
		return "", fmt.Errorf("HashDrawing: encode DXF payload: %w", err)
	}
	return hashBytes(payload.Bytes()), nil
}

func hashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

type DrawingDiff struct {
	Added    []entity.Entity
	Removed  []entity.Entity
	Modified map[string][]entity.Entity
}

func CompareDrawings(original, modified *Drawing) *DrawingDiff {
	origEntities := make(map[string]entity.Entity)
	for _, e := range original.GetAllEntities() {
		origEntities[e.Handle()] = e
	}

	modEntities := make(map[string]entity.Entity)
	for _, e := range modified.GetAllEntities() {
		modEntities[e.Handle()] = e
	}

	diff := &DrawingDiff{
		Added:    make([]entity.Entity, 0),
		Removed:  make([]entity.Entity, 0),
		Modified: make(map[string][]entity.Entity),
	}

	for handle, e := range modEntities {
		if _, ok := origEntities[handle]; !ok {
			diff.Added = append(diff.Added, e)
		}
	}

	for handle, e := range origEntities {
		if _, ok := modEntities[handle]; !ok {
			diff.Removed = append(diff.Removed, e)
		}
	}

	return diff
}

// DeepCopy returns an independent copy of the drawing. The drawing is
// serialized to its ASCII DXF payload and reloaded through the registered
// document loader, so every entity, table entry, and block is re-parsed into
// fresh objects with no shared pointers. It reports an error when
// serialization or the reload fails, or when no document loader is
// registered.
func (d *Drawing) DeepCopy() (*Drawing, error) {
	if d == nil {
		return nil, fmt.Errorf("DeepCopy: nil drawing")
	}
	if documentLoader == nil {
		return nil, fmt.Errorf("DeepCopy: document loader is not registered")
	}
	var payload bytes.Buffer
	if _, err := d.WriteTo(&payload); err != nil {
		return nil, fmt.Errorf("DeepCopy: encode DXF payload: %w", err)
	}
	copied, err := documentLoader(&payload)
	if err != nil {
		return nil, fmt.Errorf("DeepCopy: reload payload: %w", err)
	}
	if copied == nil {
		return nil, fmt.Errorf("DeepCopy: loader returned no drawing")
	}
	return copied, nil
}
