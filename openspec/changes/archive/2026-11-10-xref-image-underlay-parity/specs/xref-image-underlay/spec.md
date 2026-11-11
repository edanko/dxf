# Spec Delta

## Purpose

Defines how the library manages external references, raster images, and DGN/DWF/PDF
underlays, including path resolution, binding, clipping, and rendering.

## ADDED Requirements

### Requirement: Attach and detach external references

The library SHALL attach a DXF file as an external reference to a block and SHALL
detach it, removing its definition and inserts.

#### Scenario: Attach an xref

- **WHEN** an xref is attached at a path with an insertion point
- **THEN** a block record and insert exist and the reference resolves to the file

#### Scenario: Detach an xref

- **WHEN** an attached xref is detached
- **THEN** its block record and inserts are removed and the document audits clean

### Requirement: Path resolution

The library SHALL resolve xref and image paths as absolute or relative to the
host document, using configured search paths, and SHALL report unresolved paths.

#### Scenario: Relative path

- **WHEN** a drawing references `../refs/base.dxf`
- **THEN** the loader resolves it relative to the host file and loads the content

#### Scenario: Missing file

- **WHEN** an referenced file does not exist
- **THEN** the reference remains in the document as unresolved and the load
  reports it without failing

### Requirement: Reload and bind

The library SHALL reload an xref from disk and SHALL bind an xref, converting its
content into local block content.

#### Scenario: Bind an xref

- **WHEN** an xref is bound
- **THEN** its geometry becomes local block content and no external path remains

### Requirement: Nested references

The library SHALL load nested xrefs and SHALL detect and report reference cycles.

#### Scenario: Nested xref

- **WHEN** an xref file itself contains an xref
- **THEN** both levels load and are addressable

#### Scenario: Cycle

- **WHEN** xrefs reference each other cyclically
- **THEN** loading terminates and reports the cycle

### Requirement: Clipping

The library SHALL support xclip boundaries for xrefs and images and SHALL apply
the clip during rendering.

#### Scenario: Rectangular clip

- **WHEN** an xref is clipped to a rectangle
- **THEN** rendered content outside the rectangle is not visible

### Requirement: Raster images

The library SHALL model IMAGEDEF, IMAGEDEF_REACTOR, RASTERVARIABLES, and IMAGE
entities, and SHALL link each image to its definition.

#### Scenario: Image round trip

- **WHEN** a drawing with an image is loaded and saved
- **THEN** the image entity, definition, path, size, and display options are
  preserved

### Requirement: Underlays

The library SHALL model DGN, DWF, and PDF underlay definitions and entities,
including path, page, and display options.

#### Scenario: PDF underlay

- **WHEN** a drawing with a PDF underlay is loaded
- **THEN** the underlay entity exposes its definition, page, and frame settings

### Requirement: Display options

Images and underlays SHALL support show, clip, fade, contrast, transparency, and
frame display options, and the writer SHALL preserve them.

#### Scenario: Fade

- **WHEN** an image fade value is set
- **THEN** it is preserved on save and applied during rendering

### Requirement: Rendering integration

The renderer SHALL draw images and underlays at their insertion, scale, rotation,
and clip boundaries when their external file is available.

#### Scenario: Render an image

- **WHEN** a drawing with an available image is rendered
- **THEN** the image pixels appear in the output at the correct placement

#### Scenario: Render an unresolved reference

- **WHEN** an image file is missing during rendering
- **THEN** rendering continues and reports the missing reference
