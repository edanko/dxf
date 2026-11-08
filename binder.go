package dxf

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/edanko/dxf/block"
	"github.com/edanko/dxf/class"
	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/header"
	"github.com/edanko/dxf/lldxf"
	"github.com/edanko/dxf/object"
	"github.com/edanko/dxf/table"
)

// LoadOptions controls document loading.
type LoadOptions struct {
	// Recover loads as much valid content as possible from damaged input and
	// records the repairs in the drawing's recovery report.
	Recover bool
	// StripComments drops code 999 comments instead of retaining them.
	StripComments bool
	// PreserveUnknown retains unhandled entities, objects, and group codes.
	// It defaults to true when options are nil.
	PreserveUnknown bool
}

func defaultLoadOptions() *LoadOptions {
	return &LoadOptions{PreserveUnknown: true}
}

type documentBinder struct {
	d        *drawing.Drawing
	loader   *lldxf.Loader
	options  *LoadOptions
	report   *drawing.RecoveryReport
	handles  *handle.HandleGenerator
	entities int
}

func newDocumentBinder(d *drawing.Drawing, loader *lldxf.Loader, options *LoadOptions) *documentBinder {
	if options == nil {
		options = defaultLoadOptions()
	}
	return &documentBinder{
		d:       d,
		loader:  loader,
		options: options,
		report:  &drawing.RecoveryReport{},
		handles: handle.NewHandleGenerator(),
	}
}

// FromReaderWithOptions parses a DXF document with explicit load options.
func FromReaderWithOptions(r io.Reader, options *LoadOptions) (*drawing.Drawing, error) {
	if options == nil {
		options = defaultLoadOptions()
	}
	d, err := NewDrawing()
	if err != nil {
		return nil, err
	}

	tagger := lldxf.NewASCIITagger(r)
	var allTags lldxf.Tags
	for tagger.Next() {
		tag := tagger.Tag()
		if tag != nil {
			allTags = append(allTags, tag)
		}
	}
	report := &drawing.RecoveryReport{}
	if err := tagger.Err(); err != nil {
		if !options.Recover {
			return nil, err
		}
		report.Issues = append(report.Issues, err.Error())
	}

	loader := lldxf.NewLoader(&lldxf.LoaderOptions{
		LoadAllSections: true,
		LoadEntities:    true,
		LoadTables:      true,
		LoadBlocks:      true,
		LoadObjects:     true,
		SkipComments:    options.StripComments,
		PreserveUnknown: options.PreserveUnknown,
		BindEntities:    true,
		StrictMode:      false,
		Recover:         options.Recover,
	})
	if err := loader.Load(allTags); err != nil {
		if !options.Recover {
			return nil, fmt.Errorf("failed to load DXF: %w", err)
		}
		report.Issues = append(report.Issues, err.Error())
	}
	report.Issues = append(report.Issues, loader.GetIssues()...)

	binder := newDocumentBinder(d, loader, options)
	binder.report = report
	if err := binder.bind(); err != nil {
		if !options.Recover {
			return nil, fmt.Errorf("failed to build drawing: %w", err)
		}
		report.Issues = append(report.Issues, err.Error())
	}
	d.SetRecoveryReport(report)
	return d, nil
}

// FromBinaryReaderWithOptions parses a binary DXF document with explicit load options.
func FromBinaryReaderWithOptions(r io.Reader, options *LoadOptions) (*drawing.Drawing, error) {
	if options == nil {
		options = defaultLoadOptions()
	}
	d, err := NewDrawing()
	if err != nil {
		return nil, err
	}

	tagger := lldxf.NewBinaryTagger(r)
	var allTags lldxf.Tags
	for tagger.Next() {
		tag := tagger.Tag()
		if tag != nil {
			allTags = append(allTags, tag)
		}
	}
	report := &drawing.RecoveryReport{}
	if err := tagger.Err(); err != nil {
		if !options.Recover {
			return nil, err
		}
		report.Issues = append(report.Issues, err.Error())
	}

	loader := lldxf.NewLoader(&lldxf.LoaderOptions{
		LoadAllSections: true,
		LoadEntities:    true,
		LoadTables:      true,
		LoadBlocks:      true,
		LoadObjects:     true,
		SkipComments:    options.StripComments,
		PreserveUnknown: options.PreserveUnknown,
		BindEntities:    true,
		Recover:         options.Recover,
	})
	if err := loader.Load(allTags); err != nil {
		if !options.Recover {
			return nil, fmt.Errorf("failed to load DXF: %w", err)
		}
		report.Issues = append(report.Issues, err.Error())
	}
	report.Issues = append(report.Issues, loader.GetIssues()...)

	binder := newDocumentBinder(d, loader, options)
	binder.report = report
	if err := binder.bind(); err != nil {
		if !options.Recover {
			return nil, fmt.Errorf("failed to build drawing: %w", err)
		}
		report.Issues = append(report.Issues, err.Error())
	}
	d.SetRecoveryReport(report)
	return d, nil
}

func (b *documentBinder) bind() error {
	b.observeAllHandles()
	b.d.SetSectionOrder(b.loader.GetSectionOrder())
	b.d.SetComments(b.loader.GetComments())

	b.bindHeader()
	b.bindClasses()
	b.bindTables()
	b.bindBlocks()
	b.bindEntities()
	b.bindObjects()
	b.bindOpaqueSections()
	return nil
}

func (b *documentBinder) observeAllHandles() {
	for _, tag := range b.loader.GetSections()[""] {
		_ = tag
	}
	for _, section := range b.loader.GetSections() {
		for _, tag := range section {
			if tag.Code() == 5 || tag.Code() == 105 {
				b.handles.Observe(tagStringValue(tag))
			}
		}
	}
}

func (b *documentBinder) bindHeader() {
	tags := b.loader.GetSection(lldxf.SectionHeader)
	if len(tags) == 0 {
		return
	}
	h := header.ParseTyped(tags)
	b.d.SetHeader(h)
}

func (b *documentBinder) bindClasses() {
	tags := b.loader.GetSection(lldxf.SectionClasses)
	if len(tags) == 0 {
		return
	}
	b.d.SetClasses(class.Parse(tags))
}

func (b *documentBinder) bindTables() {
	tags := b.loader.GetSection(lldxf.SectionTables)
	if len(tags) == 0 {
		return
	}
	b.d.SetTables(table.ParseTables(tags))
}

func (b *documentBinder) bindBlocks() {
	tags := b.loader.GetSection(lldxf.SectionBlocks)
	if len(tags) == 0 {
		return
	}
	blocks := make(block.Blocks, 0)
	i := 0
	for i < len(tags) {
		if !isDXFTag(tags[i], 0, "BLOCK") {
			i++
			continue
		}
		blockTags := lldxf.Tags{tags[i]}
		i++
		for i < len(tags) && tags[i].Code() != 0 {
			blockTags = append(blockTags, tags[i])
			i++
		}
		blk := b.parseBlockHeader(blockTags)
		for i < len(tags) {
			if isDXFTag(tags[i], 0, "ENDBLK") {
				endTags := lldxf.Tags{tags[i]}
				i++
				for i < len(tags) && tags[i].Code() != 0 {
					endTags = append(endTags, tags[i])
					i++
				}
				blk.SetEndHandleString(endTags.GetHandle())
				break
			}
			if tags[i].Code() == 0 {
				// Entity inside the block.
				entityTags := lldxf.Tags{tags[i]}
				i++
				for i < len(tags) && tags[i].Code() != 0 {
					entityTags = append(entityTags, tags[i])
					i++
				}
				e := b.buildEntity(entityTags)
				if e != nil {
					blk.AddEntity(e)
				}
				continue
			}
			i++
		}
		blocks = append(blocks, blk)
	}
	b.d.SetBlocks(blocks)
}

func (b *documentBinder) parseBlockHeader(tags lldxf.Tags) *block.Block {
	name := tagValueString(tags, 2)
	desc := tagValueString(tags, 1)
	blk := block.NewBlock(name, desc)
	blk.SetHandleString(tags.GetHandle())
	blk.SetLayerName(tagValueString(tags, 8))
	if v, ok := tagValueInt(tags, 70); ok {
		blk.Flag = v
	}
	for i := 0; i < 3; i++ {
		if v, ok := tagValueFloat(tags, 10+i*10); ok {
			blk.Coord[i] = v
		}
	}
	return blk
}

func (b *documentBinder) bindEntities() {
	tags := b.loader.GetSection(lldxf.SectionEntities)
	objects := splitObjects(tags)
	for i := 0; i < len(objects); i++ {
		entityTags := objects[i]
		e := b.buildEntity(entityTags)
		if e == nil {
			continue
		}
		if p, ok := e.(*entity.Polyline); ok {
			// Consume the VERTEX records and the trailing SEQEND that belong
			// to this POLYLINE so they are not emitted twice.
			for i+1 < len(objects) && strings.EqualFold(objects[i+1].DXFType(), "VERTEX") {
				i++
				if v := loadVertex(objects[i]); v != nil {
					if h := objects[i].GetHandle(); h != "" {
						v.SetHandleString(h)
					}
					v.SetLayer(p.Layer())
					p.Vertices = append(p.Vertices, v)
				}
			}
			if i+1 < len(objects) && strings.EqualFold(objects[i+1].DXFType(), "SEQEND") {
				i++
				p.SetEndHandleString(objects[i].GetHandle())
			}
		}
		if err := b.d.AddEntity(e); err != nil {
			b.report.SkippedEntities++
			b.report.Issues = append(b.report.Issues, err.Error())
		}
		b.entities++
	}
}

func (b *documentBinder) buildEntity(tags lldxf.Tags) entity.Entity {
	dxfType := tags.DXFType()
	if dxfType == "" {
		return nil
	}
	e, err := loadEntity(dxfType, tags)
	if err != nil || e == nil {
		b.report.SkippedEntities++
		return nil
	}
	if h := tags.GetHandle(); h != "" {
		if setter, ok := e.(interface{ SetHandleString(string) }); ok {
			setter.SetHandleString(h)
		}
	}
	if e.Handle() == "" {
		e.SetHandle(b.handles)
	}
	layerName := tagValueString(tags, 8)
	b.d.SetEntityLayer(e, layerName)
	attachRawTags(e, tags)
	return e
}

func (b *documentBinder) bindObjects() {
	tags := b.loader.GetSection(lldxf.SectionObjects)
	objects := make(object.Objects, 0)
	for _, objectTags := range splitObjects(tags) {
		o := object.ParseObject(objectTags)
		if o == nil {
			continue
		}
		if o.Handle() == "" {
			o.SetHandle(b.handles)
		}
		objects = objects.Add(o)
	}
	b.d.SetObjects(objects)
}

func (b *documentBinder) bindOpaqueSections() {
	opaque := make(map[string][]drawing.RawTag)
	for name, tags := range b.loader.GetSections() {
		switch name {
		case "", lldxf.SectionHeader, lldxf.SectionClasses, lldxf.SectionTables,
			lldxf.SectionBlocks, lldxf.SectionEntities, lldxf.SectionObjects:
			continue
		}
		raw := make([]drawing.RawTag, 0, len(tags))
		for _, tag := range tags {
			raw = append(raw, drawing.RawTag{Code: int(tag.Code()), Value: tagStringValue(tag)})
		}
		opaque[name] = raw
	}
	b.d.SetOpaqueSections(opaque)
}

// attachRawTags records group codes that the typed entity model did not emit,
// so unknown codes survive a load/save round trip.
func attachRawTags(e entity.Entity, tags lldxf.Tags) {
	if e == nil {
		return
	}
	if _, ok := e.(*GenericEntity); ok {
		return
	}
	setter, ok := e.(interface{ SetRawTags([]entity.TagPair) })
	if !ok {
		return
	}
	f := format.NewASCII()
	e.Format(f)
	emitted := make(map[int]bool)
	for _, line := range strings.Split(f.Output(), "\n") {
		var code int
		if _, err := fmt.Sscanf(strings.TrimSpace(line), "%d", &code); err == nil {
			emitted[code] = true
		}
	}
	raw := make([]entity.TagPair, 0)
	for _, tag := range tags {
		code := int(tag.Code())
		if code == 0 || code == 5 || code == 105 || emitted[code] {
			continue
		}
		raw = append(raw, entity.TagPair{Code: code, Value: tagStringValue(tag)})
	}
	setter.SetRawTags(raw)
}

func splitObjects(tags lldxf.Tags) []lldxf.Tags {
	result := make([]lldxf.Tags, 0)
	var current lldxf.Tags
	for _, tag := range tags {
		if tag.Code() == 0 {
			if len(current) > 0 {
				result = append(result, current)
			}
			current = lldxf.Tags{tag}
			continue
		}
		if len(current) > 0 {
			current = append(current, tag)
		}
	}
	if len(current) > 0 {
		result = append(result, current)
	}
	return result
}

func isDXFTag(tag lldxf.Tag, code int, value string) bool {
	return int(tag.Code()) == code && strings.EqualFold(tagStringValue(tag), value)
}

func tagValueString(tags lldxf.Tags, code int) string {
	for _, tag := range tags {
		if int(tag.Code()) == code {
			return tagStringValue(tag)
		}
	}
	return ""
}

func tagValueInt(tags lldxf.Tags, code int) (int, bool) {
	for _, tag := range tags {
		if int(tag.Code()) != code {
			continue
		}
		switch v := tag.Value().(type) {
		case int:
			return v, true
		case int64:
			return int(v), true
		case float64:
			return int(v), true
		default:
			var n int
			if _, err := fmt.Sscanf(tagStringValue(tag), "%d", &n); err == nil {
				return n, true
			}
		}
	}
	return 0, false
}

func tagValueFloat(tags lldxf.Tags, code int) (float64, bool) {
	for _, tag := range tags {
		if int(tag.Code()) != code {
			continue
		}
		switch v := tag.Value().(type) {
		case float64:
			return v, true
		case int:
			return float64(v), true
		case int64:
			return float64(v), true
		default:
			var f float64
			if _, err := fmt.Sscanf(tagStringValue(tag), "%f", &f); err == nil {
				return f, true
			}
		}
	}
	return 0, false
}

// sortedHandles is used by tests and diagnostics that need stable ordering.
func sortedHandles(handles []string) []string {
	sort.Strings(handles)
	return handles
}

func tagStringValue(tag lldxf.Tag) string {
	if b, ok := tag.Value().([]byte); ok {
		return string(b)
	}
	return fmt.Sprintf("%v", tag.Value())
}
