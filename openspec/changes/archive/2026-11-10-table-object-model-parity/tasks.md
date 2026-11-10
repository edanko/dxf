# Tasks

## 1. Table foundation

- [x] 1.1 Add a shared table-entry base with handle, owner, flags, xdata, and extension dictionary; verify with LAYER
- [x] 1.2 Add generic table lookup/add/remove/rename/iterate and duplicate-name errors; verify with unit tests
- [x] 1.3 Bind all tables on load and emit them on save; verify entry counts against fixtures

## 2. Table types

- [x] 2.1 Complete LAYER, LTYPE, STYLE, and APPID; verify typed fields and round trips
- [x] 2.2 Complete BLOCK_RECORD, DIMSTYLE, UCS, VIEW, VIEWPORT, and VPORT; verify typed fields and round trips

## 3. Object foundation

- [x] 3.1 Add the object registry and raw-object fallback; verify unknown objects survive
- [x] 3.2 Add owner/handle resolution and extension dictionary/reactor wiring; verify the object graph

## 4. Object types

- [x] 4.1 Complete DICTIONARY, DICTIONARYVAR, XRECORD, GROUP, and ACDBPLACEHOLDER; verify lookup and round trip
- [x] 4.2 Complete LAYOUT and PLOTSETTINGS; verify model/paper layout association
- [x] 4.3 Complete MLINESTYLE, MLEADERSTYLE, MATERIAL, VISUALSTYLE, and TABLESTYLE; verify typed fields
- [x] 4.4 Complete SORTENTSTABLE, IDBUFFER, FIELD, FIELDLIST, GEODATA, SUN, and SPATIAL_FILTER; verify round trips
- [x] 4.5 Complete RASTERVARIABLES, WIPEOUTVARIABLES, IMAGEDEF, IMAGEDEF_REACTOR, UNDERLAYDEFINITION, and PDFDEFINITION; verify image/underlay links

## 5. Verification

- [x] 5.1 Add golden tests for every table and object type and verify handles/owners are unique and consistent
- [x] 5.2 Update `docs/parity/README.md` table/object matrix and verify it matches implemented coverage

## Workflow follow-up

- Archive the change after review.
