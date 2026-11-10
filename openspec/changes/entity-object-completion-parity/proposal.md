# Proposal

## Why

The shared model is broad but still retains several high-value records as opaque tags. Users cannot inspect table cells, collection objects, polyface/polymesh vertices, or ACIS bodies through typed APIs, so some ezdxf workflows remain impossible even though round-trip preservation works.

## What Changes

- Add typed `ACAD_TABLE` content: rows, columns, cells, styles, merges, and block references.
- Add typed collection objects: `MLEADERSTYLECOLLECTION`, `MLINESTYLECOLLECTION`, `LAYER_FILTER`, and `VBA_PROJECT`.
- Add typed subentity variants for `POLYFACE`/`POLYMESH` vertices and `MESHVERTEXCACHE`.
- Decode ACIS SAT/SAB payloads enough to expose body metadata and preserve unknown history data.
- Load objects that appear in the ENTITIES section (LAYOUT, PLOTSETTINGS, RASTERVARIABLES, WIPEOUTVARIABLES, IDBUFFER, FIELDLIST, SPATIAL_FILTER, GEODATA) as typed entities when applicable.
- Extend extension-dictionary/reactor coverage to all newly typed records.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `entity-model`: typed ACAD_TABLE content, subentity variants, ACIS decoding, objects-as-entities.
- `table-object-model`: typed collection objects and layer filters.

## Impact

Affected packages: `entity`, `object`, `table`, root loader registry, `binder.go`, `dxf_main.go`. Existing retained-record round trips remain valid; new typed paths must preserve every unhandled tag.
