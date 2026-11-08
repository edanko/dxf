package xref

import (
	"os"
	"path/filepath"
	"sync"
	"time"
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
	drawing   interface{}
	timeStamp time.Time
	fileSize  int64
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

func (x *XRef) Drawing() interface{} {
	return x.drawing
}

func (x *XRef) SetDrawing(drawing interface{}) {
	x.drawing = drawing
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
		if xref.loaded {
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

type Loader struct {
	basePath     string
	cache        map[string]interface{}
	cacheMu      sync.RWMutex
	maxCacheSize int
}

func NewLoader(basePath string) *Loader {
	return &Loader{
		basePath:     basePath,
		cache:        make(map[string]interface{}),
		maxCacheSize: 10,
	}
}

func (l *Loader) SetBasePath(path string) {
	l.basePath = path
}

func (l *Loader) GetBasePath() string {
	return l.basePath
}

func (l *Loader) Load(xref *XRef, loadFunc func(string) (interface{}, error)) error {
	if xref.loaded {
		return nil
	}

	fullPath := xref.ResolvePath(l.basePath)

	if cached, ok := l.getCached(fullPath); ok {
		xref.drawing = cached
		xref.loaded = true
		return nil
	}

	drawing, err := loadFunc(fullPath)
	if err != nil {
		return err
	}

	l.cacheMu.Lock()
	if len(l.cache) >= l.maxCacheSize {
		l.evictOldest()
	}
	l.cache[fullPath] = drawing
	l.cacheMu.Unlock()

	xref.drawing = drawing
	xref.loaded = true

	if info, err := os.Stat(fullPath); err == nil {
		xref.timeStamp = info.ModTime()
		xref.fileSize = info.Size()
	}

	return nil
}

func (l *Loader) getCached(path string) (interface{}, bool) {
	l.cacheMu.RLock()
	defer l.cacheMu.RUnlock()
	dwg, ok := l.cache[path]
	return dwg, ok
}

func (l *Loader) evictOldest() {
	if len(l.cache) == 0 {
		return
	}
	var oldestPath string
	var oldestTime time.Time
	for path := range l.cache {
		if oldestPath == "" {
			oldestPath = path
			if info, err := os.Stat(path); err == nil {
				oldestTime = info.ModTime()
			}
		} else {
			if info, err := os.Stat(path); err == nil {
				if info.ModTime().Before(oldestTime) {
					oldestPath = path
					oldestTime = info.ModTime()
				}
			}
		}
	}
	if oldestPath != "" {
		delete(l.cache, oldestPath)
	}
}

func (l *Loader) GetCachedDrawing(path string) (interface{}, bool) {
	return l.getCached(path)
}

func (l *Loader) ClearCache() {
	l.cacheMu.Lock()
	defer l.cacheMu.Unlock()
	l.cache = make(map[string]interface{})
}

func (l *Loader) CacheSize() int {
	l.cacheMu.RLock()
	defer l.cacheMu.RUnlock()
	return len(l.cache)
}

func (l *Loader) SetMaxCacheSize(size int) {
	l.maxCacheSize = size
}

func (l *Loader) GetMaxCacheSize() int {
	return l.maxCacheSize
}
