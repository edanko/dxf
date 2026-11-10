# Tasks

## 1. Xref core

- [x] 1.1 Model xref block records and metadata; verify load/save round trip
- [x] 1.2 Implement absolute/relative path resolution with search paths; verify resolved and unresolved cases
- [x] 1.3 Implement attach, detach, reload, and bind; verify each operation and a clean audit

## 2. Nested and clipping

- [x] 2.1 Implement nested xref loading with cycle detection; verify nested and cyclic fixtures
- [x] 2.2 Implement rectangular and polyline xclip boundaries; verify rendering clips content

## 3. Images

- [x] 3.1 Model IMAGEDEF, IMAGEDEF_REACTOR, RASTERVARIABLES, and IMAGE; verify round trip and linkage
- [x] 3.2 Implement image display options and rendering placement; verify fade and clip

## 4. Underlays

- [x] 4.1 Model DGN, DWF, and PDF underlay definitions and entities; verify round trip
- [x] 4.2 Implement underlay display options and rendering; verify a PDF underlay fixture

## 5. Verification

- [x] 5.1 Add missing-file and path-resolution tests; verify no load failure and clear reporting
- [x] 5.2 Update `docs/parity/README.md` xref/image/underlay matrix

## Workflow follow-up

- Archive the change after review.
