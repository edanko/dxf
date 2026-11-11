#dxf

DXF(Drawing Exchange Format by Autodesk, Inc) library for golang.

+ version: ACAD2000(AC1015)
+ format : ASCII

is only supported.

## Reference

[http://www.autodesk.com/techpubs/autocad/acad2000/dxf/index.htm](http://www.autodesk.com/techpubs/autocad/acad2000/dxf/index.htm)

## Example
```go
package main

import (
	"github.com/edanko/dxf"
	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/table"
	"log"
	"math"
)

func main() {
	d := dxf.NewDrawing()
	d.Header().LtScale = 100.0
	d.AddLayer("Toroidal", dxf.DefaultColor, dxf.DefaultLineType, true)
	d.AddLayer("Poloidal", color.Red, table.LT_HIDDEN, true)
	z := 0.0
	r1 := 200.0
	r2 := 500.0
	ndiv := 16
	dtheta := 2.0 * math.Pi / float64(ndiv)
	theta := 0.0
	for i := 0; i < ndiv; i++ {
		d.ChangeLayer("Toroidal")
		d.Circle(0.0, 0.0, z+r1*math.Cos(theta), r2-r1*math.Sin(theta))
		d.ChangeLayer("Poloidal")
		c, _ := d.Circle(r2*math.Cos(theta), r2*math.Sin(theta), 0.0, r1)
		dxf.SetExtrusion(c, []float64{-1.0 * math.Sin(theta), math.Cos(theta), 0.0})
		theta += dtheta
	}
	err := d.SaveAs("torus.dxf")
	if err != nil {
		log.Fatal(err)
	}
}
```

![Torus](example/torus/torus.png)

## Installation

```
$ go get github.com/edanko/dxf
```

## Development

### ezdxf cross-check

The optional `TestEzdxfCrossCheck` harness compares the Go parser against
Python `ezdxf` on every `testdata/*.dxf` fixture. It checks the DXF version,
entity and block inventory (including sub-entities), source TABLES entries, and
source OBJECTS records. Table and object expectations come from the raw DXF
records, so ezdxf's automatically created defaults do not cause false diffs.

Requirements:

- Go toolchain
- Python 3 with `ezdxf` installed (the reference version is `1.4.5`)
- Optional: `uv` to create a temporary environment automatically

Run the harness from the repository root:

```sh
make crosscheck
```

Or invoke the Go test directly:

```sh
DXF_EZDXF_CROSSCHECK=1 go test -run TestEzdxfCrossCheck .
```

If `ezdxf` is already installed in a Python environment, reuse it:

```sh
DXF_EZDXF_CROSSCHECK=1 \
DXF_EZDXF_PYTHON=/path/to/venv/bin/python \
go test -run TestEzdxfCrossCheck .
```

To test against a local ezdxf checkout, point `DXF_EZDXF_SOURCE` at it. The
harness creates a throwaway virtual environment with `uv` or
`python3 -m venv` when `DXF_EZDXF_PYTHON` is not set:

```sh
DXF_EZDXF_CROSSCHECK=1 \
DXF_EZDXF_SOURCE=/home/egor/src/ezdxf \
go test -run TestEzdxfCrossCheck .
```

The harness is opt-in and is not part of the normal `go test ./...` run, so the
Go library has no runtime Python dependency. Fixtures that ezdxf itself cannot
load (currently `testdata/tooling_fixture.dxf`, which intentionally keeps a
malformed APPID table entry for recovery coverage) are skipped. See
[`docs/parity/README.md`](docs/parity/README.md) for the full parity roadmap.

# License

MIT

# Author

Yoshihiro FUKUSHIMA
