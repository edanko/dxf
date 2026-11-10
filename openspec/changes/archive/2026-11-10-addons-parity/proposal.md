# Proposal

## Why

ezdxf ships a large set of add-ons that make it useful beyond parsing: streaming
readers/writers, drawing backends, CSG, text explosion, geo export, OpenSCAD and
mesh export, plot-style tables, and more. The Go library has only partial `csg`,
`text2path`, `export`, and `patterns` packages. Without equivalent add-ons, the
Go version cannot replace ezdxf for downstream tooling even if the core model is
complete.

## What Changes

- Add an add-on registration framework and a consistent package layout.
- Add a drawing add-on with vector and raster backends.
- Add streaming `r12writer` and `iterdxf` APIs for large files.
- Add an importer for external formats and `dxf2code` code generation.
- Add `acadctb` plot-style table parsing and application.
- Complete `pycsg`-equivalent mesh booleans and primitives.
- Add MTextExplode and real text-to-path.
- Add `geo` geospatial export, `meshex` mesh export, and `openscad` export.
- Add ODA File Converter integration, HPGL/2 import, and xplayer animation.
- Add bin packing, table painting, gerber, and mtxpl helpers.

## Capabilities

### New Capabilities

- `addons`: optional add-on packages built on the core library, their APIs, and
  their output guarantees.

### Modified Capabilities

None.

## Impact

Affected code: new `addons/*` packages plus existing `csg`, `text2path`,
`export`, `patterns`, `render`, and `path`. Adds optional dependencies for font,
image, geo, and mesh formats. Tests are per add-on and use small fixtures.
Depends on the core document, entity, geometry, and render changes.
