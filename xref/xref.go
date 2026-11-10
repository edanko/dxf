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

// XRefStatus is the resolution state of an external reference.
type XRefStatus int

const (
	// XRefStatusUnresolved: not loaded, file may exist.
	XRefStatusUnresolved XRefStatus = iota
	// XRefStatusLoaded: content is loaded and current.
	XRefStatusLoaded
	// XRefStatusMissing: the resolved file path does not exist.
	XRefStatusMissing
)

// String returns the status name.
func (s XRefStatus) String() string {
	switch s {
	case XRefStatusLoaded:
		return "loaded"
	case XRefStatusMissing:
		return "missing"
	default:
		return "unresolved"
	}
}

// UnresolvedError reports an external reference whose file was not found.
type UnresolvedError struct {
	Path string
	Err  error
}

// Error names the missing path.
func (e *UnresolvedError) Error() string {
	return "external reference not found: " + e.Path
}

// Unwrap exposes the underlying load error when one occurred.
func (e *UnresolvedError) Unwrap() error { return e.Err }

type XRef struct {
	filePath  string
	blockName string
	xrefType  XRefType
	pathType  PathType
	loaded    bool
	status    XRefStatus
	drawing   any
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
	if loaded {
		x.status = XRefStatusLoaded
		return
	}
	if x.status != XRefStatusMissing {
		x.status = XRefStatusUnresolved
	}
}

// Status returns the resolution state of the reference.
func (x *XRef) Status() XRefStatus { return x.status }

// SetStatus records the resolution state, keeping Loaded() in sync.
func (x *XRef) SetStatus(s XRefStatus) {
	x.status = s
	x.loaded = s == XRefStatusLoaded
}

// Timestamp returns the modification time recorded at load, if any.
func (x *XRef) Timestamp() time.Time { return x.timeStamp }

// FileSize returns the byte size recorded at load, if any.
func (x *XRef) FileSize() int64 { return x.fileSize }

func (x *XRef) ResolvePath(basePath string) string {
	if x.pathType == PathTypeRelative {
		return filepath.Join(basePath, x.filePath)
	}
	return x.filePath
}

// ResolveExternalPath resolves target against the host directory basePath
// and the additional searchPaths, returning the first existing candidate.
// When nothing exists it returns the primary candidate (the host-directory
// join, or the target itself) with found=false.
func ResolveExternalPath(target, basePath string, searchPaths []string) (string, bool) {
	var candidates []string
	if filepath.IsAbs(target) {
		candidates = append(candidates, target)
	} else {
		if basePath != "" {
			candidates = append(candidates, filepath.Join(basePath, target))
		}
		for _, sp := range searchPaths {
			if sp == "" {
				continue
			}
			candidates = append(candidates, filepath.Join(sp, target))
		}
		candidates = append(candidates, target)
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, true
		}
	}
	return candidates[0], false
}

func (x *XRef) Drawing() any {
	return x.drawing
}

func (x *XRef) SetDrawing(drawing any) {
	x.drawing = drawing
}

type XRefManager struct {
	xrefs       map[string]*XRef
	basePath    string
	searchPaths []string
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

// AddSearchPath appends a directory searched for relative references.
func (m *XRefManager) AddSearchPath(path string) {
	m.searchPaths = append(m.searchPaths, path)
}

// SearchPaths returns the configured search directories.
func (m *XRefManager) SearchPaths() []string {
	out := make([]string, len(m.searchPaths))
	copy(out, m.searchPaths)
	return out
}

// ResolveFile resolves the reference path against the base path and the
// search paths, reporting whether a candidate exists.
func (m *XRefManager) ResolveFile(x *XRef) (string, bool) {
	if x.PathType() == PathTypeAbsolute {
		return ResolveExternalPath(x.FilePath(), m.basePath, nil)
	}
	return ResolveExternalPath(x.FilePath(), m.basePath, m.searchPaths)
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
	searchPaths  []string
	cache        map[string]any
	cacheMu      sync.RWMutex
	maxCacheSize int
}

func NewLoader(basePath string) *Loader {
	return &Loader{
		basePath:     basePath,
		cache:        make(map[string]any),
		maxCacheSize: 10,
	}
}

func (l *Loader) SetBasePath(path string) {
	l.basePath = path
}

func (l *Loader) GetBasePath() string {
	return l.basePath
}

// AddSearchPath appends a directory searched for relative references.
func (l *Loader) AddSearchPath(path string) {
	l.searchPaths = append(l.searchPaths, path)
}

// SearchPaths returns the configured search directories.
func (l *Loader) SearchPaths() []string {
	out := make([]string, len(l.searchPaths))
	copy(out, l.searchPaths)
	return out
}

// Resolve resolves the reference path against base and search paths,
// reporting whether a candidate exists.
func (l *Loader) Resolve(x *XRef) (string, bool) {
	if x.PathType() == PathTypeAbsolute {
		return ResolveExternalPath(x.FilePath(), l.basePath, nil)
	}
	return ResolveExternalPath(x.FilePath(), l.basePath, l.searchPaths)
}

func (l *Loader) Load(xref *XRef, loadFunc func(string) (any, error)) error {
	if xref.loaded {
		return nil
	}

	fullPath, found := l.Resolve(xref)

	if cached, ok := l.getCached(fullPath); ok {
		xref.drawing = cached
		xref.SetLoaded(true)
		return nil
	}

	drawing, err := loadFunc(fullPath)
	if err != nil {
		if !found {
			xref.status = XRefStatusMissing
			return &UnresolvedError{Path: fullPath, Err: err}
		}
		xref.SetStatus(XRefStatusUnresolved)
		return err
	}

	l.cacheMu.Lock()
	if len(l.cache) >= l.maxCacheSize {
		l.evictOldest()
	}
	l.cache[fullPath] = drawing
	l.cacheMu.Unlock()

	xref.drawing = drawing
	xref.SetLoaded(true)

	if info, err := os.Stat(fullPath); err == nil {
		xref.timeStamp = info.ModTime()
		xref.fileSize = info.Size()
	}

	return nil
}

func (l *Loader) getCached(path string) (any, bool) {
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

func (l *Loader) GetCachedDrawing(path string) (any, bool) {
	return l.getCached(path)
}

func (l *Loader) ClearCache() {
	l.cacheMu.Lock()
	defer l.cacheMu.Unlock()
	l.cache = make(map[string]any)
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
