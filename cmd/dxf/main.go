// Command dxf inspects, validates, converts, renders, and queries DXF files.
//
// Exit codes: 0 success, 1 data error, 2 usage error, 3 missing external
// resource. Data goes to stdout, diagnostics to stderr.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	dxf "github.com/edanko/dxf"
	addonsdrawing "github.com/edanko/dxf/addons/drawing"
	"github.com/edanko/dxf/appsettings"
	"github.com/edanko/dxf/audit"
	"github.com/edanko/dxf/comments"
	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/filemanagement"
	"github.com/edanko/dxf/query"
	"github.com/edanko/dxf/render"
	"image/color"
)

const versionString = "dxf 1.0.0"

const usage = `usage: dxf <command> [options] [file]

commands:
  info       show version, entity/table/block/object counts
  audit      audit a file (--fix repairs)
  convert    convert to a target version and encoding
  draw       render to svg, pdf or png
  query      run a query expression (--format json)
  explode    explode composite entities
  fix        alias for audit --fix
  strip      remove comments and/or proxy entities
  header     list or set header variables
  layers     list layers with color and linetype
  browse     list the document structure
  version    print the version
  completion print a shell completion script
`

// exitError carries the process exit code for a failed command.
type exitError struct {
	code int
	msg  string
}

func (e *exitError) Error() string { return e.msg }

func usageError(format string, a ...any) error {
	return &exitError{code: 2, msg: fmt.Sprintf(format, a...)}
}

func dataError(format string, a ...any) error {
	return &exitError{code: 1, msg: fmt.Sprintf(format, a...)}
}

func missingError(format string, a ...any) error {
	return &exitError{code: 3, msg: fmt.Sprintf(format, a...)}
}

func main() {
	os.Exit(run(os.Args[1:]))
}

var commands = map[string]func([]string) error{
	"info":       cmdInfo,
	"audit":      cmdAudit,
	"convert":    cmdConvert,
	"draw":       cmdDraw,
	"query":      cmdQuery,
	"explode":    cmdExplode,
	"fix":        cmdFix,
	"strip":      cmdStrip,
	"header":     cmdHeader,
	"layers":     cmdLayers,
	"browse":     cmdBrowse,
	"completion": cmdCompletion,
}

func run(args []string) int {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, usage)
		return 2
	}
	switch args[0] {
	case "help", "-h", "--help":
		fmt.Print(usage)
		return 0
	case "version", "-v", "--version":
		fmt.Println(versionString)
		return 0
	}
	fn, ok := commands[args[0]]
	if !ok {
		fmt.Fprintf(os.Stderr, "dxf: unknown command %q\n%s", args[0], usage)
		return 2
	}
	if err := fn(args[1:]); err != nil {
		if ee, ok := err.(*exitError); ok {
			fmt.Fprintln(os.Stderr, "dxf:", ee.msg)
			return ee.code
		}
		fmt.Fprintln(os.Stderr, "dxf:", err)
		return 1
	}
	return 0
}

// newFlagSet creates a flag set whose -h prints the command usage and
// succeeds (exit 0).
func newFlagSet(name string, args []string, define func(*flag.FlagSet)) (*flag.FlagSet, error) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: dxf %s [options] [file]\n", name)
		fs.PrintDefaults()
	}
	fs.SetOutput(io.Discard)
	if define != nil {
		define(fs)
	}
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil, nil
		}
		return nil, usageError("%s: %v", name, err)
	}
	return fs, nil
}

// loadFile loads a DXF file, mapping failures to data errors.
func loadFile(path string) (*drawing.Drawing, error) {
	d, err := dxf.FromFile(path)
	if err != nil {
		return nil, dataError("cannot load %s: %v", path, err)
	}
	return d, nil
}

// emit writes data to the output path ("" means stdout), atomically with an
// optional backup.
func emit(path string, data []byte, backup bool) error {
	if path == "" {
		_, err := os.Stdout.Write(data)
		return err
	}
	if err := filemanagement.AtomicWrite(path, data, backup); err != nil {
		return dataError("cannot write %s: %v", path, err)
	}
	return nil
}

// saveDoc serializes a drawing and emits it.
func saveDoc(d *drawing.Drawing, output string, backup bool) error {
	var buf bytes.Buffer
	if _, err := d.WriteTo(&buf); err != nil {
		return dataError("encode: %v", err)
	}
	return emit(output, buf.Bytes(), backup)
}

// inputOutput resolves the input file argument; modifying commands default
// to writing the input path with a backup.
func inputOutput(fs *flag.FlagSet, modify bool) (in, out string, backup bool, err error) {
	rest := fs.Args()
	if len(rest) < 1 {
		return "", "", false, usageError("missing input file")
	}
	in = rest[0]
	out = in
	backup = modify
	return in, out, backup, nil
}

func cmdInfo(args []string) error {
	var format *string
	fs, err := newFlagSet("info", args, func(fs *flag.FlagSet) {
		format = fs.String("format", "text", "output format: text or json")
	})
	if err != nil || fs == nil {
		return err
	}
	in, _, _, err := inputOutput(fs, false)
	if err != nil {
		return err
	}
	d, err := loadFile(in)
	if err != nil {
		return err
	}

	entities := map[string]int{}
	for _, e := range d.Entities() {
		entities[e.DXFType()]++
	}
	tables := map[string]int{}
	for _, t := range d.Tables() {
		if t == nil {
			continue
		}
		tables[t.Name()] = len(t.Entries())
	}

	if *format == "json" {
		out := map[string]any{
			"version":  d.Header().Version,
			"entities": entities,
			"tables":   tables,
			"blocks":   len(d.Blocks()),
			"objects":  len(d.Objects()),
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		return emit("", append(b, '\n'), false)
	}

	var b bytes.Buffer
	fmt.Fprintf(&b, "version: %s\n", d.Header().Version)
	fmt.Fprintf(&b, "entities: %d\n", len(d.Entities()))
	for _, t := range sortedKeys(entities) {
		fmt.Fprintf(&b, "  %s: %d\n", t, entities[t])
	}
	fmt.Fprintf(&b, "tables: %d\n", len(tables))
	for _, name := range sortedKeys(tables) {
		fmt.Fprintf(&b, "  %s: %d\n", name, tables[name])
	}
	fmt.Fprintf(&b, "blocks: %d\n", len(d.Blocks()))
	fmt.Fprintf(&b, "objects: %d\n", len(d.Objects()))
	_, err = os.Stdout.Write(b.Bytes())
	return err
}

func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func cmdAudit(args []string) error {
	var fix *bool
	var output *string
	fs, err := newFlagSet("audit", args, func(fs *flag.FlagSet) {
		fix = fs.Bool("fix", false, "apply safe repairs")
		output = fs.String("output", "", "output path (default: input, with backup)")
	})
	if err != nil || fs == nil {
		return err
	}
	in, out, backup, err := inputOutput(fs, *fix)
	if err != nil {
		return err
	}
	if *output != "" {
		out = *output
		backup = true
	}
	d, err := loadFile(in)
	if err != nil {
		return err
	}
	rep := audit.Run(d, &audit.Options{Fix: *fix})
	fmt.Println(rep.String())
	if *fix {
		if err := saveDoc(d, out, backup); err != nil {
			return err
		}
		fmt.Printf("fixes written to %s\n", out)
	}
	if rep.HasErrors() {
		return &exitError{code: 1, msg: "audit: errors remain"}
	}
	return nil
}

func cmdFix(args []string) error {
	return cmdAudit(append([]string{"--fix"}, args...))
}

func cmdConvert(args []string) error {
	var target *string
	var encoding *string
	var output *string
	fs, err := newFlagSet("convert", args, func(fs *flag.FlagSet) {
		target = fs.String("version", "AC1015", "target DXF version")
		encoding = fs.String("encoding", "ascii", "encoding: ascii or binary")
		output = fs.String("output", "", "output path (default: input, with backup)")
	})
	if err != nil || fs == nil {
		return err
	}
	in, out, backup, err := inputOutput(fs, true)
	if err != nil {
		return err
	}
	if *output != "" {
		out = *output
	}
	if *encoding != "ascii" && *encoding != "binary" {
		return usageError("invalid encoding %q (ascii or binary)", *encoding)
	}
	d, err := loadFile(in)
	if err != nil {
		return err
	}
	d.SetTargetVersion(*target)
	var buf bytes.Buffer
	if *encoding == "binary" {
		if _, err := d.WriteBinaryTo(&buf); err != nil {
			return dataError("encode: %v", err)
		}
	} else {
		if _, err := d.WriteTo(&buf); err != nil {
			return dataError("encode: %v", err)
		}
	}
	if err := emit(out, buf.Bytes(), backup); err != nil {
		return err
	}
	fmt.Printf("converted %s to %s (%s)\n", in, *target, *encoding)
	return nil
}

func parseBackground(s string) color.RGBA {
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}
	return color.RGBA{R: uint8(v >> 16), G: uint8(v >> 8), B: uint8(v), A: 255}
}

func cmdDraw(args []string) error {
	cfg := appsettings.Load()
	var backend *string
	var width *float64
	var height *float64
	var background *string
	var layout *string
	var output *string
	fs, err := newFlagSet("draw", args, func(fs *flag.FlagSet) {
		backend = fs.String("backend", cfg.Backend, "backend: svg, pdf or png")
		width = fs.Float64("width", cfg.Width, "canvas width")
		height = fs.Float64("height", cfg.Height, "canvas height")
		background = fs.String("background", cfg.Background, "background color (#rrggbb)")
		layout = fs.String("layout", "", "layout name (empty: model space)")
		output = fs.String("output", "", "output path (default: stdout)")
	})
	if err != nil || fs == nil {
		return err
	}
	in, _, _, err := inputOutput(fs, false)
	if err != nil {
		return err
	}
	d, err := loadFile(in)
	if err != nil {
		return err
	}
	data, err := addonsdrawing.Render(d, *width, *height, *backend, &render.SceneOptions{
		Layout:     *layout,
		Background: parseBackground(*background),
	})
	if err != nil {
		return dataError("render: %v", err)
	}
	return emit(*output, data, false)
}

func cmdQuery(args []string) error {
	var format *string
	fs, err := newFlagSet("query", args, func(fs *flag.FlagSet) {
		format = fs.String("format", "text", "output format: text or json")
	})
	if err != nil || fs == nil {
		return err
	}
	rest := fs.Args()
	if len(rest) < 2 {
		return usageError("usage: dxf query [options] <file> <expression>")
	}
	in, expr := rest[0], rest[1]
	d, err := loadFile(in)
	if err != nil {
		return err
	}
	sel, err := query.ExecuteWithFilter(d.Entities(), expr)
	if err != nil {
		return dataError("query: %v", err)
	}
	if *format == "json" {
		out := map[string]any{
			"query":   expr,
			"count":   sel.Len(),
			"handles": sel.Handles(),
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		return emit("", append(b, '\n'), false)
	}
	for _, h := range sel.Handles() {
		fmt.Println(h)
	}
	return nil
}

func cmdExplode(args []string) error {
	var output *string
	fs, err := newFlagSet("explode", args, func(fs *flag.FlagSet) {
		output = fs.String("output", "", "output path (default: input, with backup)")
	})
	if err != nil || fs == nil {
		return err
	}
	in, out, backup, err := inputOutput(fs, true)
	if err != nil {
		return err
	}
	if *output != "" {
		out = *output
	}
	d, err := loadFile(in)
	if err != nil {
		return err
	}
	nd, err := drawing.New()
	if err != nil {
		return dataError("create document: %v", err)
	}
	for _, e := range query.ExplodeAll(d.Entities()) {
		if s, ok := e.(entity.HandleStringSetter); ok {
			s.SetHandleString("")
		}
		if err := nd.AddEntity(e); err != nil {
			return dataError("add entity: %v", err)
		}
	}
	return saveDoc(nd, out, backup)
}

func cmdStrip(args []string) error {
	var stripComments *bool
	var stripProxy *bool
	var output *string
	fs, err := newFlagSet("strip", args, func(fs *flag.FlagSet) {
		stripComments = fs.Bool("comments", false, "remove comments (code 999)")
		stripProxy = fs.Bool("proxy", false, "remove proxy entities")
		output = fs.String("output", "", "output path (default: input, with backup)")
	})
	if err != nil || fs == nil {
		return err
	}
	in, out, backup, err := inputOutput(fs, true)
	if err != nil {
		return err
	}
	if *output != "" {
		out = *output
	}
	d, err := loadFile(in)
	if err != nil {
		return err
	}
	if *stripComments {
		comments.StripComments(d)
	}
	if *stripProxy {
		var dead []string
		for _, e := range d.Entities() {
			if e.DXFType() == "ACAD_PROXY_ENTITY" {
				dead = append(dead, e.Handle())
			}
		}
		for _, h := range dead {
			if err := d.DeleteEntity(h); err != nil {
				return dataError("delete proxy: %v", err)
			}
		}
	}
	return saveDoc(d, out, backup)
}

func cmdHeader(args []string) error {
	var set *string
	var output *string
	fs, err := newFlagSet("header", args, func(fs *flag.FlagSet) {
		set = fs.String("set", "", "set variable NAME=VALUE")
		output = fs.String("output", "", "output path (default: input, with backup)")
	})
	if err != nil || fs == nil {
		return err
	}
	in, out, backup, err := inputOutput(fs, true)
	if err != nil {
		return err
	}
	if *output != "" {
		out = *output
	}
	d, err := loadFile(in)
	if err != nil {
		return err
	}
	h := d.Header()
	if *set != "" {
		parts := strings.SplitN(*set, "=", 2)
		if len(parts) != 2 {
			return usageError("invalid --set %q (want NAME=VALUE)", *set)
		}
		name, value := parts[0], parts[1]
		switch strings.ToUpper(name) {
		case "VERSION":
			h.Version = value
		case "SOURCEVERSION":
			h.SourceVersion = value
		case "NORMALIZEDVERSION":
			h.NormalizedVersion = value
		case "TARGETVERSION":
			h.TargetVersion = value
		default:
			return usageError("unknown header variable %q", name)
		}
		fmt.Printf("set %s=%s\n", name, value)
		return saveDoc(d, out, backup)
	}
	fmt.Printf("VERSION=%s\nSOURCEVERSION=%s\nNORMALIZEDVERSION=%s\nTARGETVERSION=%s\n",
		h.Version, h.SourceVersion, h.NormalizedVersion, h.TargetVersion)
	return nil
}

func cmdLayers(args []string) error {
	var format *string
	fs, err := newFlagSet("layers", args, func(fs *flag.FlagSet) {
		format = fs.String("format", "text", "output format: text or json")
	})
	if err != nil || fs == nil {
		return err
	}
	in, _, _, err := inputOutput(fs, false)
	if err != nil {
		return err
	}
	d, err := loadFile(in)
	if err != nil {
		return err
	}
	names := make([]string, 0, len(d.Layers))
	for name := range d.Layers {
		names = append(names, name)
	}
	sort.Strings(names)
	type layerInfo struct {
		Name     string `json:"name"`
		Color    int    `json:"color"`
		Linetype string `json:"linetype"`
	}
	if *format == "json" {
		out := []layerInfo{}
		for _, name := range names {
			l := d.Layers[name]
			lt := ""
			if l.LineType != nil {
				lt = l.LineType.Name()
			}
			out = append(out, layerInfo{Name: l.Name(), Color: int(l.Color), Linetype: lt})
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		return emit("", append(b, '\n'), false)
	}
	for _, name := range names {
		l := d.Layers[name]
		lt := ""
		if l.LineType != nil {
			lt = l.LineType.Name()
		}
		fmt.Printf("%s color=%d linetype=%s\n", l.Name(), int(l.Color), lt)
	}
	return nil
}

func cmdBrowse(args []string) error {
	fs, err := newFlagSet("browse", args, nil)
	if err != nil || fs == nil {
		return err
	}
	in, _, _, err := inputOutput(fs, false)
	if err != nil {
		return err
	}
	d, err := loadFile(in)
	if err != nil {
		return err
	}
	fmt.Printf("HEADER version=%s\n", d.Header().Version)
	for _, t := range d.Tables() {
		if t == nil {
			continue
		}
		fmt.Printf("TABLE %s (%d entries)\n", t.Name(), len(t.Entries()))
	}
	fmt.Printf("BLOCKS (%d)\n", len(d.Blocks()))
	fmt.Printf("ENTITIES (%d)\n", len(d.Entities()))
	fmt.Printf("OBJECTS (%d)\n", len(d.Objects()))
	return nil
}

func cmdCompletion(args []string) error {
	var shell *string
	fs, err := newFlagSet("completion", args, func(fs *flag.FlagSet) {
		shell = fs.String("shell", "bash", "shell: bash, zsh, fish or powershell")
	})
	if err != nil || fs == nil {
		return err
	}
	switch *shell {
	case "bash":
		fmt.Print("complete -W \"info audit convert draw query explode fix strip header layers browse version completion\" dxf\n")
	case "zsh":
		fmt.Print("#compdef dxf\n_arguments '1:command:(info audit convert draw query explode fix strip header layers browse version completion)'\n")
	case "fish":
		fmt.Print("complete -c dxf -f -a \"info audit convert draw query explode fix strip header layers browse version completion\"\n")
	case "powershell":
		fmt.Print("Register-ArgumentCompleter -CommandName dxf -ScriptBlock { param($w) 'info','audit','convert','draw','query','explode','fix','strip','header','layers','browse','version','completion' | Where-Object { $_ -like \"$w*\" } }\n")
	default:
		return usageError("invalid shell %q (bash, zsh, fish, powershell)", *shell)
	}
	return nil
}
