# Tasks

## 1. Primitive decomposition

- [x] 1.1 Add a `Primitive` value type (kind, source handle, vertices, optional path or mesh, diagnostics) and `make_primitive` for LINE, ARC, CIRCLE, ELLIPSE, SPLINE, POLYLINE, MESH, TEXT, MTEXT, SOLID, and INSERT; verify unsupported entities produce a diagnostic and POINT yields vertices only.
- [x] 1.2 Add `recursive_decompose` that expands nested INSERT, MTEXT, POLYLINE, and HATCH with depth and cycle guards; verify a nested INSERT with LINE and TEXT yields transformed leaf primitives and a cycle terminates with a report.
- [x] 1.3 Add `to_vertices`, `to_paths`, `to_meshes`, and `to_control_vertices` conversions; verify an ARC path preserves endpoints, a mesh preserves vertices and face indices, and POINT returns an empty path.
- [x] 1.4 Update `docs/parity/README.md` query/disassembly rows to describe primitive decomposition accurately.

## 2. Dynamic blocks, XRECORD helpers, proxy primitives

- [x] 2.1 Add dynamic-block read helpers resolving an INSERT directly or through an anonymous block to the dynamic definition and metadata; verify direct, anonymous, and ordinary-block cases.
- [x] 2.2 Add `UserRecord` compile/parse for strings, integers, floats, points, nested lists, and dictionaries using XRECORD tags; verify a nested round trip, invalid string rejection, and preservation of unrelated XRECORD tags.
- [x] 2.3 Add `BinaryRecord` compile/parse for arbitrary byte payloads; verify every byte value round-trips.
- [x] 2.4 Extend proxy-graphics decoding with documented line, arc, circle, polyline, text, color, and transform opcodes while retaining embedded entity-record extraction; verify decoded primitives and an unknown-opcode diagnostic with byte offset.
- [x] 2.5 Update `docs/parity/README.md` for dynamic blocks, XRECORD user/binary records, and proxy primitives.

## 3. Text size APIs

- [x] 3.1 Add a `TextSize` API for TEXT reporting width, height, ascent, descent, and transformed bounding box using configured height, width factor, rotation, alignment, and font fallback; verify a known string measures non-zero extents and reports the fallback reason.
- [x] 3.2 Extend the API to MTEXT with wrapping, paragraph spacing, and column width; verify laid-out extents grow with additional wrapped lines.
- [x] 3.3 Use the measured metrics in path text layout and scene extents; verify wrapping and extents tests use measured width rather than a fixed advance.
- [x] 3.4 Update `docs/parity/README.md` text metric and layout rows.

## 4. Custom hatch patterns

- [x] 4.1 Add a `.pat` parser for comments, names, descriptions, angles, base points, offsets, and dash items with file/line errors and partial results; verify a multi-pattern file and a malformed line.
- [x] 4.2 Add case-insensitive pattern lookup that preserves the source name and reports unknown names; verify lookup and unknown-name results.
- [x] 4.3 Add non-mutating scale and measurement-system conversion; verify the original definition is unchanged and scaled copies produce expected offsets.
- [x] 4.4 Expand parsed patterns into clipped hatch line segments and report unknown pattern names while retaining the boundary stroke; verify expansion counts and the unknown fallback.
- [x] 4.5 Update `docs/parity/README.md` hatch pattern rows.

## 5. Linked-edge topology

- [x] 5.1 Add an `edges` package with immutable edges and a graph builder merging endpoints within a tolerance through a spatial hash; verify degree counts, leaves, and components.
- [x] 5.2 Add loop and open-chain analysis over the graph; verify one closed loop, one open chain, and mixed networks.
- [x] 5.3 Add creation of ordered LWPOLYLINE and POLYLINE entities from loops and chains; verify vertex order and closure.
- [x] 5.4 Add hatch boundary creation from closed edge loops with an explicit error for open boundaries; verify one closed boundary and the open-boundary error.
- [x] 5.5 Preserve source edge orientation in created entities and support reversal; verify reversed output has reversed vertex order.
- [x] 5.6 Update `docs/parity/README.md` with the edge tooling coverage.

## 6. Native DWG metadata and small utilities

- [x] 6.1 Add a read-only DWG version/release reader; verify recognized `AC10xx` headers and invalid short input.
- [x] 6.2 Add partial DWG section-locator parsing for header variables, classes, and object summaries with diagnostics for unsupported sections; verify one recognized section and one unsupported section.
- [x] 6.3 Keep the external ODA converter unchanged and document that native DWG support is metadata-only; verify the converter tests still pass and a native parse never writes output.
- [x] 6.4 Add civil date-time to Julian date conversion and back; verify 2000-01-01 12:00 UTC equals 2451545.0 within one second and a round trip.
- [x] 6.5 Add a raw section loader returning ordered tags per section without document binding; verify sections and a truncated-section error.
- [x] 6.6 Update `docs/parity/README.md` DWG and utility rows.

## 7. Integration

- [x] 7.1 Route ACIS and linked-edge export add-ons through decoded geometry; verify a cube body and a closed edge loop reach the exporter with expected triangle/vertex counts.
- [x] 7.2 Run `go test ./...`, the DWG/pattern/proxy fuzz smoke tests, and a fixture exercising decomposition, text size, patterns, edges, and raw loading.
- [x] 7.3 Refresh `docs/parity/README.md` remaining-gap notes for the completed tooling scope.
