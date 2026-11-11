# dxf

A pure-Go DXF library for reading, writing, inspecting, and rendering AutoCAD
drawing files.

- **Broad version support** — reads R12 through R2018 in ASCII, and binary
  from R14 onward, with read-time upgrade for R13/R14 and pre-R12 files.
- **Typed document model** — entities, tables, blocks, layouts, objects,
  dictionaries, groups, xrefs, images, and underlays.
- **Round-trip safety** — unknown tags, comments, opaquely loaded sections, and
  unsupported entities are preserved when saving.
- **Rendering** — SVG, PDF, and PNG backends with gradients, complex
  linetypes, plot styles, and fallback reporting.
- **Tools included** — query language, audit/recovery checks, geometry and
  path utilities, and a `dxf` command-line client.
- **No cgo and no runtime Python dependency.**

## Installation

```sh
go get github.com/edanko/dxf
```

Go 1.27 or newer is required.

## Quick start

```go
package main

import (
	"log"

	"github.com/edanko/dxf"
	"github.com/edanko/dxf/color"
)

func main() {
	d, err := dxf.NewDrawing()
	if err != nil {
		log.Fatal(err)
	}

	if _, err := d.AddLayer("walls", color.Yellow, d.LtContinuous()); err != nil {
		log.Fatal(err)
	}
	if _, err := d.AddLayer("doors", color.Cyan, d.LtDashDot()); err != nil {
		log.Fatal(err)
	}

	if err := d.SetActiveLayer("walls"); err != nil {
		log.Fatal(err)
	}
	if _, err := d.Line(0, 0, 0, 100, 0, 0); err != nil {
		log.Fatal(err)
	}
	if _, err := d.Arc(100, 25, 0, 25, 0, 90); err != nil {
		log.Fatal(err)
	}
	if _, err := d.Text("plan", 0, 10, 0, 5); err != nil {
		log.Fatal(err)
	}

	if err := d.SaveAs("plan.dxf"); err != nil {
		log.Fatal(err)
	}
}
```

Load an existing drawing just as easily:

```go
d, err := dxf.FromFile("plan.dxf")
if err != nil {
	log.Fatal(err)
}
log.Printf("%d entities in model space", len(d.ModelSpace()))
```

`dxf.FromStringData`, `dxf.FromReader`, `dxf.FromBinaryFile`, and
`dxf.FromBinaryReader` cover the other input forms. Use
`dxf.FromReaderWithOptions` with `dxf.LoadOptions{Recover: true}` to retain as
much data as possible from damaged files.

## Rendering

Render a drawing to SVG, PDF, or PNG through the core renderer backends:

```go
import (
	"os"

	renderdrawing "github.com/edanko/dxf/addons/drawing"
	"github.com/edanko/dxf/render"
)

opts := &render.SceneOptions{Tessellation: 0.05}
svg, err := renderdrawing.RenderSVG(d, 1600, 1200, opts)
if err != nil {
	log.Fatal(err)
}
if err := os.WriteFile("plan.svg", svg, 0o644); err != nil {
	log.Fatal(err)
}

pdf, err := renderdrawing.Render(d, 1600, 1200, "pdf", opts)
if err != nil {
	log.Fatal(err)
}
if err := os.WriteFile("plan.pdf", pdf, 0o644); err != nil {
	log.Fatal(err)
}
```

`render.BuildScene` exposes the resolved scene when you need to draw with a
custom backend. `StyledScene.Fallbacks` reports degradations such as missing
SHX shape fonts or malformed geometry.

## Query

```go
import "github.com/edanko/dxf/query"

walls, err := query.NewEntityQuery(d.ModelSpace()...).Where(`layer=="walls"`)
if err != nil {
	log.Fatal(err)
}
for _, e := range walls.ToSlice() {
	log.Printf("%s handle=%s", e.DXFType(), e.Handle())
}
```

Queries can also be compiled and reused with `query.CompileQuery`.

## Audit and recovery

```go
import "github.com/edanko/dxf/audit"

auditor := audit.NewComprehensiveAuditor(nil) // default configuration
result, err := auditor.AuditDrawing(d)
if err != nil {
	log.Fatal(err)
}
auditor.PrintResult(result)
```

The auditor checks references, tables, dictionaries, extents, duplicate
handles, and block cycles, with optional automatic fixes. The same checks are
available for files through `AuditFile`.

## Command-line client

```sh
go install github.com/edanko/dxf/cmd/dxf@latest
# or build from a checkout
go build -o bin/dxf ./cmd/dxf
```

Examples:

```sh
dxf info plan.dxf
dxf audit plan.dxf
dxf query plan.dxf 'LINE layer=="walls"'
dxf draw plan.dxf --backend svg --output plan.svg
dxf convert plan.dxf --version R12 --output plan-r12.dxf
```

See [`docs/cli.md`](docs/cli.md) for every command, flag, configuration
variable, and exit code.

## Supported DXF versions

| DXF version | ACADVER | Read | Write |
| --- | --- | --- | --- |
| R12 | AC1009 | yes | ASCII |
| R13 / R14 | AC1012 / AC1014 | yes, upgraded | ASCII + binary as R2000 |
| R2000 | AC1015 | yes | ASCII + binary |
| R2004 | AC1018 | yes | ASCII + binary |
| R2007 | AC1021 | yes | ASCII + binary |
| R2010 | AC1024 | yes | ASCII + binary |
| R2013 | AC1027 | yes | ASCII + binary |
| R2018 | AC1032 | yes | ASCII + binary |

Unknown or pre-R12 versions are normalized using the same policy as ezdxf
(pre-R12 to R12, R13/R14 to R2000). Use `d.SetTargetVersion("R2018")` before
saving to write a specific version; `d.TargetVersion()` reports the selected
target.

## Feature overview

- **DXF I/O**: ASCII and binary tags, codepage-aware text, comments, opaque
  sections, unknown-tag preservation, recovery mode, target-version saving,
  and R12 strict mode.
- **Entities**: lines, points, circles, arcs, ellipses with embedded splines,
  polylines with bulges, splines, text and MTEXT, dimensions and leaders,
  hatches with typed boundary edges and gradients, MPOLYGON, meshes, 3D faces,
  ACIS bodies and surfaces, images, wipeouts, underlays, blocks, attributes,
  and xrefs.
- **Document model**: header variables, tables, layouts and paper space,
  dictionaries, groups, draw order, and extension dictionaries.
- **Geometry**: vectors, matrices, Bézier and B-spline utilities, clipping,
  offsets, Delaunay triangulation, bounding boxes, and constructive solid
  geometry through BSP trees.
- **Rendering**: SVG, PDF, and PNG output, linear and radial gradients, complex
  linetypes with SHX shapes, CTB/STB plot styles, and deterministic fallback
  diagnostics.
- **Inspection**: typed queries, audit and recovery, drawing standards checks,
  JSON snapshots, content hashes, and deep copies.
- **Add-ons**: R12 streaming export, DXF-to-code generation, dimension lines,
  fractal and genetic generators, DWG metadata inspection, and more.

The project tracks feature and behavior parity with Python
[`ezdxf`](https://ezdxf.mozman.at/). The detailed status, known gaps, and
implementation notes live in [`docs/parity/README.md`](docs/parity/README.md).

## Development

Common tasks are wrapped in the `Makefile`:

```sh
make build          # build all packages
make test           # run the test suite
make lint           # run golangci-lint v2
make check          # tests + lint
make crosscheck     # compare the loader against Python ezdxf
make fmt            # format with goimports
```

Install the pinned linter with `make lint-install`. CI runs build, tests, vet,
lint, and the ezdxf cross-check on every push and pull request.

### ezdxf cross-check

`make crosscheck` runs `TestEzdxfCrossCheck`, which compares the Go loader
against Python `ezdxf` on every `testdata/*.dxf` fixture. It checks the DXF
version, entity and block inventory (including sub-entities), source table
entries, and source object records. Expectations are read from the raw DXF, so
ezdxf's automatically created defaults do not cause false differences.

Requirements: Python 3 with `ezdxf` (reference version `1.4.5`). If `uv` is
available, the harness can create a temporary environment automatically.

```sh
make crosscheck
```

To reuse an existing environment or test a local ezdxf checkout:

```sh
DXF_EZDXF_PYTHON=/path/to/venv/bin/python make crosscheck
DXF_EZDXF_SOURCE=/path/to/ezdxf make crosscheck
```

The harness is opt-in and is not part of `go test ./...`, so the library has no
runtime Python dependency. Fixtures that ezdxf itself cannot load (currently
the intentionally malformed `testdata/tooling_fixture.dxf`, kept for recovery
coverage) are skipped.

## Authors

- Yoshihiro FUKUSHIMA — original author
- Egor Danko — current maintainer

## License

MIT
