package drawing

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"

	"github.com/edanko/dxf/entity"
)

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

func (d *Drawing) ToJSON(opts *SerializationOptions) ([]byte, error) {
	if opts == nil {
		opts = NewSerializationOptions()
	}

	type DrawingJSON struct {
		FileName    string                   `json:"file_name,omitempty"`
		LayerCount  int                      `json:"layer_count"`
		StyleCount  int                      `json:"style_count"`
		EntityCount int                      `json:"entity_count"`
		Statistics  map[string]interface{}   `json:"statistics,omitempty"`
		Entities    []map[string]interface{} `json:"entities"`
	}

	stats := d.GetDrawingStatistics()
	entities := d.GetAllEntities()

	entitiesJSON := make([]map[string]interface{}, 0, len(entities))
	for _, e := range entities {
		entMap := entityToMap(e, opts)
		entitiesJSON = append(entitiesJSON, entMap)
	}

	dj := DrawingJSON{
		FileName:    d.fileName,
		LayerCount:  len(d.Layers),
		StyleCount:  len(d.Styles),
		EntityCount: len(entities),
		Statistics:  stats,
		Entities:    entitiesJSON,
	}

	if opts.Compact {
		return json.Marshal(dj)
	}
	return json.MarshalIndent(dj, "", "  ")
}

func entityToMap(e entity.Entity, opts *SerializationOptions) map[string]interface{} {
	result := make(map[string]interface{})
	result["type"] = e.DXFType()

	if opts.IncludeHandles && e.Handle() != "" {
		result["handle"] = e.Handle()
	}

	if layer := e.Layer(); layer != nil {
		result["layer"] = layer.Name()
	}

	return result
}

func (d *Drawing) FromJSON(data []byte) error {
	return nil
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
	Type    string                 `json:"type"`
	Handle  string                 `json:"handle,omitempty"`
	Layer   string                 `json:"layer,omitempty"`
	Color   int                    `json:"color,omitempty"`
	RawData map[string]interface{} `json:"raw_data,omitempty"`
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

func HashDrawing(d *Drawing) (string, error) {
	data, err := d.ToBinary()
	if err != nil {
		return "", err
	}
	return hashBytes(data), nil
}

func hashBytes(data []byte) string {
	hash := 0
	for i, b := range data {
		hash = (hash<<8 + int(b)) % 1000000007
		_ = i
	}
	return ""
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

func (d *Drawing) DeepCopy() *Drawing {
	_ = d
	return nil
}
