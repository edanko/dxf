package audit

import (
	"strings"

	"github.com/edanko/dxf/block"
)

// CodeBlockCycle reports INSERT edges that make a block definition contain
// itself, directly or through other blocks.
const CodeBlockCycle = "BLOCK_CYCLE"

func init() {
	RegisterCheck(Check{Code: CodeBlockCycle, Severity: SeverityError, Safe: true, Fn: checkBlockCycles})
	RegisterFixer(CodeBlockCycle, fixBlockCycles)
}

type blockNameReader interface {
	BlockName() string
}

// checkBlockCycles reports every cycle in the block INSERT graph.
func checkBlockCycles(ctx *Context) []Issue {
	graph := blockGraph(ctx)
	cycles := findBlockCycles(graph)
	out := make([]Issue, 0, len(cycles))
	for _, cycle := range cycles {
		out = append(out, Issue{
			Severity: SeverityError,
			Code:     CodeBlockCycle,
			Handle:   cycle[0],
			Message:  "block INSERT cycle: " + strings.Join(cycle, " -> "),
		})
	}
	return out
}

func blockGraph(ctx *Context) map[string][]string {
	graph := make(map[string][]string)
	for _, blk := range ctx.Blocks() {
		for _, e := range blk.Entities() {
			if !strings.EqualFold(e.DXFType(), "INSERT") {
				continue
			}
			reader, ok := e.(blockNameReader)
			if !ok {
				continue
			}
			name := reader.BlockName()
			if name == "" {
				continue
			}
			graph[blk.Name] = append(graph[blk.Name], name)
		}
	}
	return graph
}

// findBlockCycles returns each distinct cycle as a path whose first and last
// names repeat (A -> B -> A).
func findBlockCycles(graph map[string][]string) [][]string {
	const (
		white = 0
		gray  = 1
		black = 2
	)
	color := make(map[string]int)
	var stack []string
	var cycles [][]string
	seen := make(map[string]bool)

	var visit func(string)
	visit = func(name string) {
		color[name] = gray
		stack = append(stack, name)
		for _, next := range graph[name] {
			switch color[next] {
			case white:
				visit(next)
			case gray:
				for i, n := range stack {
					if n != next {
						continue
					}
					cycle := append([]string(nil), stack[i:]...)
					cycle = append(cycle, next)
					key := strings.Join(cycle, ">")
					if !seen[key] {
						seen[key] = true
						cycles = append(cycles, cycle)
					}
					break
				}
			}
		}
		stack = stack[:len(stack)-1]
		color[name] = black
	}
	for name := range graph {
		if color[name] == white {
			visit(name)
		}
	}
	return cycles
}

// fixBlockCycles removes the INSERT edge that closes the first cycle. Blocks
// themselves are retained, so definitions are never deleted.
func fixBlockCycles(ctx *Context, issue Issue) bool {
	changed := false
	for {
		cycles := findBlockCycles(blockGraph(ctx))
		if len(cycles) == 0 {
			return changed
		}
		cycle := cycles[0]
		if len(cycle) < 2 {
			return changed
		}
		from, to := cycle[0], cycle[1]
		blk := findBlock(ctx, from)
		if blk == nil {
			return changed
		}
		entities := blk.Entities()
		kept := make([]block.BlockEntity, 0, len(entities))
		removed := false
		for _, e := range entities {
			reader, ok := e.(blockNameReader)
			if !removed && ok && strings.EqualFold(e.DXFType(), "INSERT") && strings.EqualFold(reader.BlockName(), to) {
				removed = true
				continue
			}
			kept = append(kept, e)
		}
		if !removed {
			return changed
		}
		blk.SetEntities(kept)
		changed = true
	}
}

func findBlock(ctx *Context, name string) *block.Block {
	for _, blk := range ctx.Blocks() {
		if strings.EqualFold(blk.Name, name) {
			return blk
		}
	}
	return nil
}
