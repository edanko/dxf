package xref

import (
	"os"
	"path/filepath"
)

type XRefType int

const (
	XRefTypeOverlay XRefType = iota
	XRefTypeAttachment
)

type PathType int

const (
	PathTypeAbsolute PathType = iota
	PathTypeRelative
)

type XRef struct {
	filePath  string
	blockName string
	xrefType  XRefType
	pathType  PathType
	loaded    bool
}

func NewXRef(filePath string, blockName string, xrefType XRefType) *XRef {
	return &XRef{
		filePath:  filePath,
		blockName: blockName,
		xrefType:  xrefType,
		pathType:  PathTypeAbsolute,
		loaded:    false,
	}
}

func (x *XRef) FilePath() string {
	return x.filePath
}

func (x *XRef) BlockName() string {
	return x.blockName
}

func (x *XRef) XRefType() XRefType {
	return x.xrefType
}

func (x *XRef) PathType() PathType {
	return x.pathType
}

func (x *XRef) Loaded() bool {
	return x.loaded
}

func (x *XRef) SetPathType(pathType PathType) {
	x.pathType = pathType
}

func (x *XRef) IsOverlay() bool {
	return x.xrefType == XRefTypeOverlay
}

func (x *XRef) IsAttachment() bool {
	return x.xrefType == XRefTypeAttachment
}

func (x *XRef) SetLoaded(loaded bool) {
	x.loaded = loaded
}

func (x *XRef) ResolvePath(basePath string) string {
	if x.pathType == PathTypeRelative {
		return filepath.Join(basePath, x.filePath)
	}
	return x.filePath
}

type XRefManager struct {
	xrefs    map[string]*XRef
	basePath string
}

func NewXRefManager(basePath string) *XRefManager {
	return &XRefManager{
		xrefs:    make(map[string]*XRef),
		basePath: basePath,
	}
}

func (m *XRefManager) Add(xref *XRef) error {
	blockName := xref.BlockName()
	if _, exists := m.xrefs[blockName]; exists {
		return os.ErrExist
	}
	m.xrefs[blockName] = xref
	return nil
}

func (m *XRefManager) AddXRef(filePath, blockName string, xrefType XRefType) error {
	pathType := PathTypeAbsolute
	if !filepath.IsAbs(filePath) {
		pathType = PathTypeRelative
	}

	xref := NewXRef(filePath, blockName, xrefType)
	xref.SetPathType(pathType)
	return m.Add(xref)
}

func (m *XRefManager) Get(blockName string) (*XRef, bool) {
	xref, exists := m.xrefs[blockName]
	return xref, exists
}

func (m *XRefManager) Remove(blockName string) error {
	if _, exists := m.xrefs[blockName]; !exists {
		return os.ErrNotExist
	}
	delete(m.xrefs, blockName)
	return nil
}

func (m *XRefManager) Count() int {
	return len(m.xrefs)
}

func (m *XRefManager) LoadedCount() int {
	count := 0
	for _, xref := range m.xrefs {
		if xref.Loaded() {
			count++
		}
	}
	return count
}

func (m *XRefManager) GetAll() []*XRef {
	xrefs := make([]*XRef, 0, len(m.xrefs))
	for _, xref := range m.xrefs {
		xrefs = append(xrefs, xref)
	}
	return xrefs
}

func (m *XRefManager) List() []string {
	names := make([]string, 0, len(m.xrefs))
	for name := range m.xrefs {
		names = append(names, name)
	}
	return names
}

func (m *XRefManager) Exists(blockName string) bool {
	_, exists := m.xrefs[blockName]
	return exists
}

func (m *XRefManager) SetBasePath(path string) {
	m.basePath = path
}

func (m *XRefManager) GetBasePath() string {
	return m.basePath
}

func (m *XRefManager) ConvertToAbsolute(filePath string) string {
	if filepath.IsAbs(filePath) {
		return filePath
	}
	return filepath.Join(m.basePath, filePath)
}

func (m *XRefManager) ConvertToRelative(filePath string) string {
	if !filepath.IsAbs(filePath) {
		return filePath
	}
	relPath, err := filepath.Rel(m.basePath, filePath)
	if err != nil {
		return filePath
	}
	return relPath
}

func XRefExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}
