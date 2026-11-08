# Design

## Context

`format` defines version constants AC1009 through AC1032 but the writer always
emits AC1015 ASCII, and `WriteBinaryTo` omits OBJECTS and EOF. The header model
stores only `Version`, `InsUnit`, `ExtMin`, and `ExtMax`. The tagger is
ASCII-only. See `proposal.md` for motivation.

## Goals / Non-Goals

**Goals:**

- One version profile table drives load normalization and save output.
- Complete ASCII and binary writers for every supported version.
- Explicit, reported conversion instead of silent loss.

**Non-Goals:**

- Byte-identical output to AutoCAD or ezdxf.
- Support for writing R13/R14.
- Full `$ACADVER`-dependent behavior of every AutoCAD release beyond the
  documented DXF differences.

## Decisions

### Version profile table

Represent each version as a profile with flags: supportsObjects,
supportsClasses, utf8, supportsTrueColor, supportsTransparency, handleFormat,
and table/entity field restrictions. Both loader and writer consult the same
profile. Alternative: conditionals scattered through entity `Format` methods.
Rejected because it is untestable and drifts.

### Normalize on load, convert on save

Loading always yields a document at a writable version (pre-R12→R12,
R13/R14→R2000). Saving to a different version runs a conversion pass that walks
entities and tables, applies the target profile, and records changes. This keeps
in-memory documents in a known shape. Alternative: retain the original version
and convert only at write time. Rejected because entity models would need
per-version variants.

### Encoding layer in the tagger/writer

Keep encoding out of entity code. The ASCII tagger decodes bytes using the
detected codepage; the ASCII writer encodes using the target profile. Binary
uses its own codec with the same profile. Alternative: per-field encoding.
Rejected as error-prone.

### R12 as a strict output mode

R12 output applies a dedicated sanitizer that removes unsupported sections and
fields and reports omissions. This mirrors ezdxf's `r12strict` behavior without
making R12 a separate document type.

## Risks / Trade-offs

- **Conversion loss** → Always return a report; never drop silently.
- **Codepage detection errors** → Fall back to the declared codepage, then to
  the documented default, and record the choice.
- **Binary format bugs** → Golden binary fixtures plus ASCII/binary equivalence
  tests.
- **Profile drift** → Single table with table-driven tests per version.
