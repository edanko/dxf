# Design

## Approach

Dimension rendering should reuse the existing path converter for block entities, then apply the dimension entity transform and text style. Measure text through the existing dimension formatter.

Plot styles become a small table model shared by STB and CTB parsers. Binary CTB parsing follows the documented AcadColorBook/plot-style layout; unsupported records are retained or reported.

SHX parsing starts with shape/header metadata and glyph metrics; glyph outline expansion follows the existing TTF path pipeline. ISO patterns are data tables consumed by the existing hatch segment expander.

## Testing

Add golden geometry tests for dimension arrows/text, mesh face counts, each arrow style, binary CTB fixtures, SHX metrics, and ISO pattern expansion. Extend render conformance to assert non-empty geometry for dimension-heavy fixtures.
