# Proposal

## Why

The library currently writes a single default format (AC1015), always in ASCII,
and its binary writer omits the OBJECTS section and EOF marker. It cannot detect
or honor the source DXF version, apply version-specific encoding, or convert a
drawing between releases. Real drawings span R12 through R2018 and mix ASCII and
binary, so version and encoding handling must be explicit before broader
compatibility work can succeed.

## What Changes

- Detect the source DXF version from `$ACADVER` and file structure.
- Normalize unsupported versions: pre-R12 to R12, R13/R14 to R2000.
- Add a version profile table covering R12, R2000, R2004, R2007, R2010, R2013,
  and R2018.
- Write any supported version on request, in ASCII or binary.
- Apply version-specific encoding: UTF-8 for R2007+, codepage for older files.
- Apply R12 structural restrictions (no OBJECTS/CLASSES, restricted tables and
  entity fields) when writing R12.
- Complete the binary writer so it emits every section and EOF.
- Add version conversion (upgrade and downgrade) with a report of dropped or
  transformed content.
- Add cross-version round-trip and encoding tests.

## Capabilities

### New Capabilities

- `dxf-version-encoding`: DXF release detection, normalization, conversion,
  ASCII/binary encoding, and version-specific output rules.

### Modified Capabilities

None.

## Impact

Affected code: `format/*`, `lldxf/*`, `drawing/*`, `dxf_main.go`, `header/*`,
`entity/*`, `table/*`. New public options for target version, encoding, and
conversion policy. Existing default output remains AC1015 ASCII. New fixtures
cover R12, R2000, R2007, R2013, and binary inputs.
