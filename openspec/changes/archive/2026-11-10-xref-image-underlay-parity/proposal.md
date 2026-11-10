# Proposal

## Why

External references, raster images, and underlays are common in production
drawings. The Go library has `xref` manager/loader stubs and no image or underlay
model beyond partial entity definitions. It cannot attach, resolve, bind, clip,
or round-trip these references, so such drawings load incompletely and cannot be
rendered or edited.

## What Changes

- Implement external reference attach, detach, reload, unload, and bind.
- Resolve absolute and relative xref paths, including search paths.
- Support nested xrefs with cycle detection.
- Implement xclip boundaries and clipping for xrefs and images.
- Add IMAGEDEF, IMAGEDEF_REACTOR, RASTERVARIABLES, and image entity support.
- Add underlay definitions and entities for DGN, DWF, and PDF.
- Support display options: show/clip, fade, contrast, transparency, and frame.
- Preserve unresolved references on load and save.
- Integrate with rendering and export.
- Add missing-file and path-resolution tests.

## Capabilities

### New Capabilities

- `xref-image-underlay`: external references, raster images, and underlays,
  including path resolution, binding, clipping, and rendering.

### Modified Capabilities

None.

## Impact

Affected code: `xref/*`, `entity/*` (IMAGE, WIPEOUT, underlays), `object/*`
(definitions), `render/*`, `path/*`, `drawing/*`. New public xref/image APIs.
Tests use nested fixtures and missing-file cases. Depends on the table/object
model, path/render, and document I/O changes.
