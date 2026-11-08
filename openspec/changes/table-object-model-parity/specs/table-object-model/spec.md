# Spec Delta

## Purpose

Defines the document's symbol tables, table entries, and objects, including
their typed fields, ownership relationships, and lookup behavior.

## ADDED Requirements

### Requirement: Complete symbol tables

The document SHALL expose APPID, BLOCK_RECORD, DIMSTYLE, LAYER, LTYPE, STYLE,
UCS, VIEW, VIEWPORT, and VPORT tables with typed entries.

#### Scenario: Load all tables

- **WHEN** a drawing containing every standard table is loaded
- **THEN** each table is present and its entry names match the file

#### Scenario: Layer fields

- **WHEN** a layer entry declares color, linetype, lineweight, transparency, and
  plot flags
- **THEN** the typed layer exposes all of those values

### Requirement: Table lookup and mutation

Each table SHALL support lookup by name, addition, removal, rename, and
iteration, and SHALL reject duplicate names.

#### Scenario: Add and look up a layer

- **WHEN** a layer named `Walls` is added and then looked up by name
- **THEN** the same entry is returned

#### Scenario: Duplicate name

- **WHEN** an entry with an existing name is added
- **THEN** the operation fails with a duplicate-name error

### Requirement: Table entries have handles and owners

Every table entry SHALL have a unique handle and the correct owner handle, and
the writer SHALL emit both.

#### Scenario: Entry ownership

- **WHEN** a layer is added to a document
- **THEN** its owner handle references the LAYER table and its handle is unique

### Requirement: Standard objects are typed

The document SHALL expose typed models for DICTIONARY, DICTIONARYVAR, XRECORD,
GROUP, LAYOUT, PLOTSETTINGS, MLINESTYLE, MLEADERSTYLE, MATERIAL, VISUALSTYLE,
TABLESTYLE, SORTENTSTABLE, IDBUFFER, FIELD, FIELDLIST, GEODATA, SUN,
SPATIAL_FILTER, RASTERVARIABLES, WIPEOUTVARIABLES, IMAGEDEF,
IMAGEDEF_REACTOR, UNDERLAYDEFINITION, and PDFDEFINITION.

#### Scenario: Layout object

- **WHEN** a drawing with a named paper-space layout is loaded
- **THEN** a LAYOUT object exposes the layout name and its associated block
  record

#### Scenario: Image definition

- **WHEN** a drawing with a raster image is loaded
- **THEN** the IMAGEDEF object exposes the image path and size and is linked
  from the IMAGE entity

### Requirement: Dictionaries resolve by key

The root dictionary and nested dictionaries SHALL support key lookup, insertion,
removal, and iteration, and SHALL resolve entry handles to objects.

#### Scenario: Root dictionary lookup

- **WHEN** an object is stored under key `ACAD_GROUP` in the root dictionary
- **THEN** lookup returns the referenced GROUP dictionary object

#### Scenario: Missing key

- **WHEN** a missing key is looked up
- **THEN** the lookup reports absence without creating an entry

### Requirement: Extension dictionaries and reactors

Objects SHALL expose extension dictionary and persistent reactor links, and the
writer SHALL preserve them.

#### Scenario: Extension dictionary round trip

- **WHEN** an entity or object with an extension dictionary is loaded and saved
- **THEN** the extension dictionary and its entries are preserved

### Requirement: Objects round trip by handle

Loading, saving, and reloading SHALL preserve object handles, owners, and
cross-references.

#### Scenario: Object graph stability

- **WHEN** a document with dictionaries, groups, and layouts is round tripped
- **THEN** every object handle and owner link is unchanged
