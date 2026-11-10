# dxf CLI

`cmd/dxf` builds a small command-line client over the Go DXF library.

## Build and install

```sh
go build -o bin/dxf ./cmd/dxf
```

## Usage

```
dxf <command> [options] [file]
```

Commands: info, audit, convert, draw, query, explode, fix, strip, header,
layers, browse, version, completion. Run `dxf <command> --help` for
command-specific options; `dxf --help` lists all commands.

## Subcommands

| Command | Description |
|---|---|
| `info` | print version, encoding, entity/table/block/object counts (`--format text\|json`) |
| `audit` | run the audit checks (`--fix`, `--output`) |
| `convert` | rewrite a file in another DXF version (`--version`, `--encoding ascii\|binary`, `--output`) |
| `draw` | render a drawing to SVG/PDF/PNG (`--backend`, `--width`, `--height`, `--background`, `--layout`, `--output`) |
| `query` | evaluate a query expression (`--format text\|json`) |
| `explode` | explode composite entities into primitives (`--output`) |
| `fix` | audit with `--fix` |
| `strip` | remove comments and/or proxy entities (`--comments`, `--proxy`, `--output`) |
| `header` | list or set header variables (`--set NAME=VALUE`) |
| `layers` | list layers with color/linetype (`--format text\|json`) |
| `browse` | list the document structure |
| `version` | print the CLI version |
| `completion` | print a shell completion script (`--shell bash\|zsh\|fish\|powershell`) |

## Exit codes

| Code | Meaning |
|---|---|
| 0 | success |
| 1 | data error (unreadable or invalid input) |
| 2 | usage error (unknown command or flag) |
| 3 | missing external resource (unresolved xref/image) |

## Configuration

Settings resolve with the following precedence (highest first):

1. command flags
2. environment variables (`DXF_OUTPUT`, `DXF_BACKEND`, `DXF_WIDTH`, `DXF_HEIGHT`, `DXF_BACKGROUND`)
3. JSON config file at `$DXF_CONFIG` or `~/.config/dxf/config.json`
4. built-in defaults (SVG, 800x600, white background)

## Atomic writes and backups

Modifying commands write through a temporary file in the destination
directory, fsync it, and rename it over the target. With the default
in-place mode the previous file is copied to `<file>.bak` first;
`--output` writes to an explicit path instead.
