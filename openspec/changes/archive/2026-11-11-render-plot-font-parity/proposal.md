# Proposal

## Why

Rendering is usable for core entities but incomplete for production drawings: dimension blocks are lowered without full transforms/text placement, mesh and arrow rendering are partial, plot styles are not modeled, SHX fonts are absent, and ISO hatch patterns are limited.

## What Changes

- Render DIMENSION block content fully, including transforms, measured text, and paper-scale placement.
- Complete mesh and arrow-head rendering.
- Parse and apply STB text and binary CTB plot styles.
- Add SHX font loading plus text metrics for SHX and TTF.
- Add the ISO hatch pattern library.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `path-render`: full dimension, mesh, and arrow rendering.
- `addons`: plot styles, SHX/TTF text metrics, ISO hatch patterns.

## Impact

Affected packages: `render`, `addons/acadctb`, `addons/text2path`, `patterns`, `text2path`. Rendering output may change for dimension-heavy drawings; existing scene APIs remain.
