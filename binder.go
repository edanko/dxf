package dxf

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/edanko/dxf/block"
	"github.com/edanko/dxf/class"
	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/header"
	"github.com/edanko/dxf/internal/blockadapter"
	"github.com/edanko/dxf/lldxf"
	"github.com/edanko/dxf/object"
	"github.com/edanko/dxf/table"
	"github.com/edanko/dxf/xdata"

	"github.com/edanko/dxf/color"
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
	allTags, validationIssues := applyTagValidation(allTags, options.Recover)
	report.Issues = append(report.Issues, validationIssues...)

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
	d.ObserveLoadedHandles()
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
	allTags, validationIssues := applyTagValidation(allTags, options.Recover)
	report.Issues = append(report.Issues, validationIssues...)

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
	d.ObserveLoadedHandles()
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
	b.d.SetPositionedComments(b.loader.GetPositionedComments())

	b.bindHeader()
	b.bindClasses()
	b.bindTables()
	b.d.RestoreXRefRecords()
	b.bindBlocks()
	b.bindEntities()
	b.linkMeshVertexCaches()
	b.linkACISData()
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
				if setter, ok := e.(entity.OwnerSetter); ok {
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
	return blockadapter.Adapt(e)
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
	for i := range 3 {
		if v, ok := tagValueFloat(tags, 10+i*10); ok {
			blk.Coord[i] = v
		}
	}
	return blk
}

// wireXData attaches extended data (group codes 1001-1071) from the record
// tags to entities that support it, so audit and round-trip see the apps.
func wireXData(e entity.Entity, tags lldxf.Tags) {
	setter, ok := e.(entity.XDataObjectSetter)
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
			if setter, ok := e.(entity.OwnerSetter); ok {
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

// linkMeshVertexCaches attaches MESHVERTEXCACHE records to the MESH entity
// named by their owner handle (code 330). Missing owners are left unlinked;
// the cache record still round-trips as an independent entity.
func (b *documentBinder) linkMeshVertexCaches() {
	meshes := make(map[string]*entity.Mesh)
	for _, e := range b.d.Entities() {
		if mesh, ok := e.(*entity.Mesh); ok && mesh.Handle() != "" {
			meshes[strings.ToUpper(mesh.Handle())] = mesh
		}
	}
	for _, e := range b.d.Entities() {
		cache, ok := e.(*entity.MeshVertexCache)
		if !ok {
			continue
		}
		owner := cache.Owner()
		if owner == nil {
			continue
		}
		if mesh := meshes[strings.ToUpper(owner.Handle())]; mesh != nil {
			mesh.SetVertexCache(cache)
		}
	}
}

// linkACISData attaches binary SAB payloads retained in the ACDSDATA section to
// their owning ACIS entities. The section itself remains opaque and is emitted
// byte-for-byte on save.
func (b *documentBinder) linkACISData() {
	tags := b.loader.GetSection(lldxf.SectionAcDsData)
	if len(tags) == 0 {
		return
	}
	payloads := acisDataByHandle(tags)
	if len(payloads) == 0 {
		return
	}
	for _, e := range b.d.Entities() {
		setter, ok := e.(entity.ACISBinaryDataSetter)
		if !ok {
			continue
		}
		data, ok := payloads[strings.ToUpper(e.Handle())]
		if !ok || len(data) == 0 {
			continue
		}
		setter.SetACISBinaryData(data)
	}
}

// acisDataByHandle extracts ASM_Data binary chunks from ACDSDATA records,
// keyed by the associated entity handle (code 320).
func acisDataByHandle(tags lldxf.Tags) map[string][]byte {
	out := make(map[string][]byte)
	var handle string
	var chunks []byte
	flush := func() {
		if handle != "" && len(chunks) > 0 {
			out[strings.ToUpper(handle)] = append([]byte(nil), chunks...)
		}
		handle = ""
		chunks = nil
	}
	for _, tag := range tags {
		switch int(tag.Code()) {
		case 0:
			flush()
		case 320:
			if handle == "" {
				handle = tagStringValue(tag)
			}
		case 310:
			switch v := tag.Value().(type) {
			case []byte:
				chunks = append(chunks, v...)
			case string:
				if decoded, err := hex.DecodeString(strings.TrimSpace(v)); err == nil {
					chunks = append(chunks, decoded...)
				}
			}
		}
	}
	flush()
	return out
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
	rt, ok := e.(entity.RawTagReadWriter)
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
		if setter, ok := e.(entity.HandleStringSetter); ok {
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
	attachEntityLinks(e, tags)
	attachRawTags(e, tags)
	b.validateEntityColor(e)
	return e
}

// attachEntityLinks parses the persistent reactor and extension-dictionary
// groups into the typed entity base when the model exposes those fields.
func attachEntityLinks(e entity.Entity, tags lldxf.Tags) {
	reactors, xdict := entityLinks(tags)
	if setter, ok := e.(entity.ReactorSetter); ok && len(reactors) > 0 {
		setter.SetReactorHandles(reactors)
	}
	if setter, ok := e.(entity.ExtDictSetter); ok && xdict != "" {
		setter.SetExtDictHandle(xdict)
	}
}

// entityLinks extracts handles from the 102 {ACAD_REACTORS and
// 102 {ACAD_XDICTIONARY groups.
func entityLinks(tags lldxf.Tags) ([]string, string) {
	var reactors []string
	var xdict string
	inReactors, inXDict := false, false
	for _, tag := range tags {
		code := int(tag.Code())
		if code == 102 {
			switch strings.ToUpper(strings.TrimSpace(tagStringValue(tag))) {
			case "{ACAD_REACTORS":
				inReactors = true
			case "{ACAD_XDICTIONARY":
				inXDict = true
			case "}":
				inReactors = false
				inXDict = false
			}
			continue
		}
		if inReactors && code == 330 {
			reactors = append(reactors, tagStringValue(tag))
			continue
		}
		if inXDict && code == 360 && xdict == "" {
			xdict = tagStringValue(tag)
		}
	}
	return reactors, xdict
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
	objects = b.linkObjectCollections(objects)
	b.d.SetObjects(objects)
	// Model-space draw order: SORTENTSTABLE records reorder the entities.
	b.d.ApplySortEntsTables()
}

// linkObjectCollections upgrades dictionaries referenced by the root
// dictionary's collection keys (ACAD_MLEADERSTYLE, ACAD_MLINESTYLE) to their
// typed collection models. Handles, owners, entries, and retained tags are
// preserved; only the concrete Go type changes.
func (b *documentBinder) linkObjectCollections(objects object.Objects) object.Objects {
	root := rootDictionary(objects)
	if root == nil {
		return objects
	}
	byHandle := make(map[string]int, len(objects))
	for i, o := range objects {
		byHandle[strings.ToUpper(o.Handle())] = i
	}
	for _, key := range root.Keys() {
		collectionType, _, ok := object.CollectionForDictionaryKey(key)
		if !ok {
			continue
		}
		entry, ok := root.GetEntry(key)
		if !ok || entry.Value == "" {
			continue
		}
		idx, ok := byHandle[strings.ToUpper(entry.Value)]
		if !ok {
			b.report.Issues = append(b.report.Issues,
				fmt.Sprintf("collection %s references missing dictionary %s", key, entry.Value))
			continue
		}
		d, ok := objects[idx].(*object.Dictionary)
		if !ok {
			continue // already typed or not a dictionary
		}
		objects[idx] = object.WrapCollection(d, collectionType)
	}
	return objects
}

// rootDictionary returns the OBJECTS-section dictionary with no owner.
func rootDictionary(objects object.Objects) *object.Dictionary {
	for _, o := range objects {
		if d, ok := o.(*object.Dictionary); ok && (d.Owner == "" || d.Owner == "0") {
			return d
		}
	}
	return nil
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
	setter, ok := e.(entity.RawTagWriter)
	if !ok {
		return
	}
	if getter, ok := e.(entity.RawTagReader); ok && len(getter.RawTags()) > 0 {
		return // raw tags already derived; keep them append-only
	}
	f := format.NewASCII()
	e.Format(f)
	emitted := make(map[int]bool)
	for line := range strings.SplitSeq(f.Output(), "\n") {
		var code int
		if _, err := fmt.Sscanf(strings.TrimSpace(line), "%d", &code); err == nil {
			emitted[code] = true
		}
	}
	raw := make([]entity.TagPair, 0)
	var reactors []string
	var xdict string
	if reader, ok := e.(entity.ReactorReader); ok {
		reactors = reader.ReactorHandles()
	}
	if reader, ok := e.(entity.ExtDictReader); ok {
		xdict = reader.ExtDictHandle()
	}
	inReactors, inXDict := false, false
	for _, tag := range tags {
		code := int(tag.Code())
		if code == 102 {
			switch strings.ToUpper(strings.TrimSpace(tagStringValue(tag))) {
			case "{ACAD_REACTORS":
				inReactors = true
				if len(reactors) == 0 {
					raw = append(raw, entity.TagPair{Code: code, Value: tagStringValue(tag)})
				}
				continue
			case "{ACAD_XDICTIONARY":
				inXDict = true
				if xdict == "" {
					raw = append(raw, entity.TagPair{Code: code, Value: tagStringValue(tag)})
				}
				continue
			case "}":
				if inReactors {
					inReactors = false
					if len(reactors) == 0 {
						raw = append(raw, entity.TagPair{Code: code, Value: tagStringValue(tag)})
					}
					continue
				}
				if inXDict {
					inXDict = false
					if xdict == "" {
						raw = append(raw, entity.TagPair{Code: code, Value: tagStringValue(tag)})
					}
					continue
				}
			}
			raw = append(raw, entity.TagPair{Code: code, Value: tagStringValue(tag)})
			continue
		}
		if inReactors {
			if len(reactors) == 0 {
				raw = append(raw, entity.TagPair{Code: code, Value: tagStringValue(tag)})
			}
			continue
		}
		if inXDict {
			if xdict == "" {
				raw = append(raw, entity.TagPair{Code: code, Value: tagStringValue(tag)})
			}
			continue
		}
		// 66 is a derived "entities follow" flag for INSERT/POLYLINE that the
		// typed model emits once vertices/attributes are attached. Retaining it
		// as an unknown tag duplicates it after attribute binding.
		if code == 0 || code == 5 || code == 66 || code == 105 || emitted[code] {
			continue
		}
		// Owner (330) and extension-dictionary (360) links are re-emitted by
		// the typed base. Standalone occurrences are consumed here.
		if code == 330 || (code == 360 && xdict != "") {
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

func tagStringValue(tag lldxf.Tag) string {
	return lldxf.FormatValue(tag)
}

// validateEntityColor reports invalid ACI color values without mutating the
// entity: the raw value is preserved and the issue is recorded on the binder
// recovery report.
func (b *documentBinder) validateEntityColor(e entity.Entity) {
	c, ok := e.(entity.ColorReader)
	if !ok {
		return
	}
	if err := color.ValidateACI(c.Color()); err != nil {
		b.report.Issues = append(b.report.Issues, fmt.Sprintf("%v on %s [%s]: raw value preserved", err, e.DXFType(), e.Handle()))
	}
}
