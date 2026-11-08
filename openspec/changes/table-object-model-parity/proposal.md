# Proposal

## Why

Tables and objects are the document backbone: layers, linetypes, styles, block
records, dimension styles, layouts, dictionaries, and image/underlay
definitions. The Go library models only a subset, and the load path discards
them. Without complete symbol tables and objects, entities cannot resolve
defaults, layouts cannot be addressed, and files cannot round trip.

## What Changes

- Complete the symbol tables: APPID, BLOCK_RECORD, DIMSTYLE, LAYER, LTYPE,
  STYLE, UCS, VIEW, VIEWPORT, and VPORT, with typed entries and standard fields.
- Add table APIs: lookup, add, remove, rename, iterate, and default-entry
  handling.
- Add the missing object types: DICTIONARY, DICTIONARYVAR, XRECORD, GROUP,
  LAYOUT, PLOTSETTINGS, MLINESTYLE, MLEADERSTYLE, MATERIAL, VISUALSTYLE,
  TABLESTYLE, SORTENTSTABLE, IDBUFFER, FIELD, FIELDLIST, GEODATA, SUN,
  SPATIAL_FILTER, RASTERVARIABLES, WIPEOUTVARIABLES, IMAGEDEF,
  IMAGEDEF_REACTOR, UNDERLAYDEFINITION, PDFDEFINITION, and related definitions.
- Implement owner/handle wiring, extension dictionaries, and reactor links.
- Bind tables and objects during load and emit them during save.

## Capabilities

### New Capabilities

- `table-object-model`: symbol tables, table entries, DXF objects, ownership,
  and dictionary behavior.

### Modified Capabilities

None.

## Impact

Affected code: `table/*`, `object/*`, `drawing/*`, `header/*`, and the document
binder. New public table/object APIs. Tests cover every table and object type,
ownership links, and round trips. Depends on the document I/O change for binding.
