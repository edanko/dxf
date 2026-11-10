package dxf

import (
	"bytes"
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
	"github.com/edanko/dxf/xdata"
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
// Binary DXF input (detected by signature) is routed to the binary parser.
func FromReaderWithOptions(r io.Reader, options *LoadOptions) (*drawing.Drawing, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if lldxf.IsBinaryDXF(data) {
		return FromBinaryReaderWithOptions(bytes.NewReader(data), options)
	}
	if options == nil {
		options = defaultLoadOptions()
	}
	d, err := NewDrawing()
	if err != nil {
		return nil, err
	}

	tagger := lldxf.NewASCIITagger(bytes.NewReader(data))
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
	if cp := d.Codepage(); cp != "" && !format.Profile(d.NormalizedVersion()).UTF8 {
		redecodeStrings(loader, d, cp)
	}
	d.SetRecoveryReport(report)
	return d, nil
}

// redecodeStrings re-decodes string tag contents that the ASCII tagger read
// as raw bytes, using the document's declared codepage. It runs after the
// header is bound so $DWGCODEPAGE is known, and renames loaded layers, styles,
// line types, and blocks whose names contain codepage bytes.
func redecodeStrings(loader *lldxf.Loader, d *drawing.Drawing, codepage string) {
	for _, tags := range loader.GetSections() {
		for _, tag := range tags {
			v, ok := tag.Value().(string)
			if !ok || v == "" {
				continue
			}
			t, ok := tag.(*lldxf.DXFTag)
			if !ok {
				continue
			}
			decoded := lldxf.DecodeCodepageBytes([]byte(v), codepage)
			if decoded != v {
				t.SetValue(decoded)
			}
		}
	}
	d.RenameNamesForCodepage(codepage)
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

	data, rerr := io.ReadAll(r)
	if rerr != nil {
		return nil, fmt.Errorf("failed to read binary DXF: %w", rerr)
	}
	tagger, codepage := newEncodingBinaryTaggerFromBytes(data)
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
	if codepage != "" && d.Codepage() == "" {
		d.SetCodepage(codepage)
	}
	if err := binder.bind(); err != nil {
		if !options.Recover {
			return nil, fmt.Errorf("failed to build drawing: %w", err)
		}
		report.Issues = append(report.Issues, err.Error())
	}
	d.SetRecoveryReport(report)
	return d, nil
}

// newEncodingBinaryTagger creates a binary tagger whose string decoding is
// selected from the file's $ACADVER and $DWGCODEPAGE declarations. It returns
// the tagger and the codepage it was configured with.
func newEncodingBinaryTaggerFromBytes(data []byte) (*lldxf.BinaryTagger, string) {
	version, codepage := scanBinaryHeaderBytes(data)
	utf8 := format.Profile(version).UTF8
	if !utf8 && codepage == "" {
		codepage = "ANSI_1252"
	}
	bt := lldxf.NewBinaryTaggerFromBytes(data)
	bt.SetBinaryEncoding(codepage, utf8)
	return bt, codepage
}

// scanBinaryHeader decodes the leading tags of a binary DXF stream and
// returns the declared $ACADVER and $DWGCODEPAGE values.
func scanBinaryHeaderBytes(data []byte) (version, codepage string) {
	bt := lldxf.NewBinaryTaggerFromBytes(data)
	for i := 0; i < 64 && bt.Next(); i++ {
		tag := bt.Tag()
		name := strings.ToUpper(tagStringValue(tag))
		if name == "$ACADVER" {
			if bt.Next() {
				version = tagStringValue(bt.Tag())
			}
			break
		}
	}
	bt2 := lldxf.NewBinaryTaggerFromBytes(data)
	for i := 0; i < 128 && bt2.Next(); i++ {
		tag := bt2.Tag()
		if strings.ToUpper(tagStringValue(tag)) == "$DWGCODEPAGE" {
			if bt2.Next() {
				codepage = tagStringValue(bt2.Tag())
			}
			break
		}
	}
	return version, codepage
}

// ioTaggerReader lets the binary tagger stop after a bounded number of tags:
// it reports EOF once the budget is exhausted.
type ioTaggerReader struct {
	data []byte
	pos  int
}

func (r *ioTaggerReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
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
	b.d.SetSourceVersion(h.SourceVersion)
	b.d.SetNormalizedVersion(h.NormalizedVersion)
	if h.Codepage != "" {
		b.d.SetCodepage(h.Codepage)
	}
	b.d.SetConversion(h.Conversion)
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

// readBlock parses one BLOCK...ENDBLK record starting at *i, consuming nested
// blocks recursively, and advances *i past the ENDBLK terminator.
func (b *documentBinder) readBlock(tags lldxf.Tags, i *int) *block.Block {
	blockTags := lldxf.Tags{tags[*i]}
	*i++
	for *i < len(tags) && tags[*i].Code() != 0 {
		blockTags = append(blockTags, tags[*i])
		*i++
	}
	blk := b.parseBlockHeader(blockTags)
	var endTags lldxf.Tags
	rawStream := lldxf.Tags{}
	rawStream = append(rawStream, blockTags...)
	for *i < len(tags) {
		if isDXFTag(tags[*i], 0, "ENDBLK") {
			endTags = append(endTags, tags[*i])
			*i++
			for *i < len(tags) && tags[*i].Code() != 0 {
				endTags = append(endTags, tags[*i])
				*i++
			}
			blk.SetEndHandleString(endTags.GetHandle())
			rawStream = append(rawStream, endTags...)
			break
		}
		if isDXFTag(tags[*i], 0, "SECTION") || isDXFTag(tags[*i], 0, "ENDSEC") {
			break
		}
		if isDXFTag(tags[*i], 0, "BLOCK") {
			start := *i
			nested := b.readBlock(tags, i)
			if nested != nil {
				blk.AddEntity(nested)
				rawStream = append(rawStream, tags[start:*i]...)
			}
			continue
		}
		if tags[*i].Code() == 0 {
			start := *i
			entityTags := lldxf.Tags{tags[*i]}
			*i++
			for *i < len(tags) && tags[*i].Code() != 0 {
				entityTags = append(entityTags, tags[*i])
				*i++
			}
			e := b.buildEntity(entityTags)
			if e == nil {
				b.report.SkippedEntities++
				b.report.Issues = append(b.report.Issues,
					fmt.Sprintf("skipped malformed entity record: type %s handle %s",
						tags[start].Value(), tags.GetHandle()))
				rawStream = append(rawStream, tags[start:*i]...)
				continue // Unbuildable record: raw tags retained in stream.
			}
			if owner := entityOwnerHandle(entityTags); owner != "" {
				if setter, ok := e.(interface{ SetOwner(handle.Handler) }); ok {
					setter.SetOwner(handleRef{h: owner})
					stripRawTag330(e)
				}
			}
			blk.AddEntity(asBlockEntity(e))
			if len(entityTags) == *i-start {
				rawStream = append(rawStream, tags[start:*i]...)
			} else {
				rawStream = append(rawStream, entityTags...)
			}
			continue
		}
		*i++
	}
	blk.SetRawTags(toBlockTagPairs(rawStream))
	return blk
}

func pairToTag(code int, value string) lldxf.Tag {
	return lldxf.NewDXFTag(lldxf.GroupCode(code), value)
}

func toBlockTagPairs(tags lldxf.Tags) []block.TagPair {
	out := make([]block.TagPair, 0, len(tags))
	for _, t := range tags {
		out = append(out, block.TagPair{Code: int(t.Code()), Value: tagStringValue(t)})
	}
	return out
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
		blocks = append(blocks, b.readBlock(tags, &i))
	}
	b.d.SetBlocks(blocks)
}

// asBlockEntity adapts an entity.Entity to block.BlockEntity. All library
// entities, including retained records, satisfy the interface directly; external
// implementations that do not are wrapped with a minimal adapter.
func asBlockEntity(e entity.Entity) block.BlockEntity {
	if be, ok := e.(block.BlockEntity); ok {
		return be
	}
	return &blockEntityAdapter{Entity: e}
}

// blockEntityAdapter supplies the raw-tag methods for entities that lack them.
type blockEntityAdapter struct {
	entity.Entity
	raw []block.TagPair
}

// Unwrap exposes the adapted entity for consumers that need the concrete type.
func (a *blockEntityAdapter) Unwrap() entity.Entity { return a.Entity }

func (a *blockEntityAdapter) RawTags() []block.TagPair { return a.raw }

func (a *blockEntityAdapter) SetRawTags(tags []block.TagPair) { a.raw = tags }

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

// wireXData attaches extended data (group codes 1001-1071) from the record
// tags to entities that support it, so audit and round-trip see the apps.
func wireXData(e entity.Entity, tags lldxf.Tags) {
	setter, ok := e.(interface{ SetXDataObject(*xdata.XData) })
	if !ok {
		return
	}
	x := xdata.NewXData()
	var appID string
	var tagsOut []xdata.Tag
	flush := func() {
		if appID != "" {
			x.Apps[appID] = xdata.AppData{AppID: appID, Tags: tagsOut}
			tagsOut = nil
		}
	}
	seen := false
	for _, t := range tags {
		code := int(t.Code())
		if code == 1001 {
			flush()
			appID = fmt.Sprint(t.Value())
			seen = true
			continue
		}
		if appID != "" && code >= 1000 && code <= 1071 {
			tagsOut = append(tagsOut, xdata.Tag{Code: code, Value: t.Value()})
		}
	}
	flush()
	if seen {
		setter.SetXDataObject(x)
	}
}

func (b *documentBinder) bindEntities() {
	tags := b.loader.GetSection(lldxf.SectionEntities)
	objects := splitObjects(tags)
	for i := 0; i < len(objects); i++ {
		entityTags := objects[i]
		e := b.buildEntity(entityTags)
		if e == nil {
			b.report.SkippedEntities++
			b.report.Issues = append(b.report.Issues,
				fmt.Sprintf("skipped malformed entity record: type %s handle %s",
					entityTags[0].Value(), entityTags.GetHandle()))
			continue
		}
		if owner := entityOwnerHandle(entityTags); owner != "" {
			if setter, ok := e.(interface{ SetOwner(handle.Handler) }); ok {
				setter.SetOwner(handleRef{h: owner})
				stripRawTag330(e)
			}
		}
		if ins, ok := e.(*entity.Insert); ok {
			// Consume the ATTRIB records (and trailing SEQEND) that belong to
			// this INSERT so attributes are attached to the block reference.
			for i+1 < len(objects) && strings.EqualFold(objects[i+1].DXFType(), "ATTRIB") {
				i++
				if a := loadAttrib(objects[i]); a != nil {
					if h := objects[i].GetHandle(); h != "" {
						a.SetHandleString(h)
					}
					a.SetLayer(ins.Layer())
					ins.AddAttribute(a)
				}
			}
			if i+1 < len(objects) && strings.EqualFold(objects[i+1].DXFType(), "SEQEND") {
				i++
				// SEQEND terminates the attribute list; keep its handle on the
				// last attribute so it survives round trips.
				attrs := ins.GetAllAttributes()
				if len(attrs) > 0 {
					attrs[len(attrs)-1].SetEndHandleString(objects[i].GetHandle())
				}
			}
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
		if ins, ok := e.(*entity.Insert); ok {
			b.resolveInsertBlock(ins)
		}
		if err := b.d.AddEntity(e); err != nil {
			b.report.SkippedEntities++
			b.report.Issues = append(b.report.Issues, err.Error())
		}
		b.entities++
	}
}

// entityOwnerHandle returns the object-owned owner (code 330) of an entity
// record, skipping the 330 handles inside ACAD_REACTORS sections.
func entityOwnerHandle(tags lldxf.Tags) string {
	inReactors := false
	for _, tag := range tags {
		code := int(tag.Code())
		value := tagStringValue(tag)
		if code == 102 {
			trimmed := strings.Trim(value, "{}")
			if strings.EqualFold(trimmed, "ACAD_REACTORS") {
				inReactors = true
			} else if trimmed == "" {
				inReactors = false
			}
			continue
		}
		if code == 330 && !inReactors {
			return value
		}
	}
	return ""
}

// stripRawTag330 removes code-330 entries from rawTags on an entity so the
// owner handle is not emitted twice (once by SetOwner, once by rawTags).
func stripRawTag330(e entity.Entity) {
	rt, ok := e.(interface {
		RawTags() []entity.TagPair
		SetRawTags([]entity.TagPair)
	})
	if !ok {
		return
	}
	tags := rt.RawTags()
	if len(tags) == 0 {
		return
	}
	var kept []entity.TagPair
	inReactors := false
	for _, t := range tags {
		if t.Code == 102 {
			v := strings.TrimSpace(t.Value)
			inReactors = v == "{ACAD_REACTORS"
			kept = append(kept, t)
			continue
		}
		// Only the standalone owner 330 is removed; reactor links inside
		// {ACAD_REACTORS} groups must survive for audit.
		if t.Code == 330 && !inReactors {
			continue
		}
		kept = append(kept, t)
	}
	rt.SetRawTags(kept)
}

// handleRef pins a handle string captured at load time.
type handleRef struct{ h string }

// Handle returns the captured handle.
func (r handleRef) Handle() string { return r.h }

// SetHandle is a no-op; the handle was captured from the source stream.
func (r handleRef) SetHandle(hg *handle.HandleGenerator) {}

// resolveInsertBlock resolves an insert's block definition from the document
// blocks so bbox and explode operations work on loaded documents.
func (b *documentBinder) resolveInsertBlock(ins *entity.Insert) {
	if ins.BlockDef() != nil {
		return
	}
	name := ins.BlockName()
	if name == "" {
		return
	}
	for _, blk := range b.d.Blocks() {
		if blk.Name == name {
			ins.SetBlockDef(blk)
			return
		}
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
	// Wire xdata before capturing raw tags: the entity then emits its
	// extended data, and those group codes are excluded from the raw tags
	// (no duplication on write).
	wireXData(e, tags)
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
	// Model-space draw order: SORTENTSTABLE records reorder the entities.
	b.d.ApplySortEntsTables()
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
	setter, ok := e.(interface{ SetRawTags([]entity.TagPair) })
	if !ok {
		return
	}
	if getter, ok := e.(interface{ RawTags() []entity.TagPair }); ok && len(getter.RawTags()) > 0 {
		return // raw tags already derived; keep them append-only
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
		// 66 is a derived "entities follow" flag for INSERT/POLYLINE that the
		// typed model emits once vertices/attributes are attached. Retaining it
		// as an unknown tag duplicates it after attribute binding.
		if code == 0 || code == 5 || code == 66 || code == 105 || emitted[code] {
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
	return lldxf.FormatValue(tag)
}
