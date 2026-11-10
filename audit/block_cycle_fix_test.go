package audit

import (
	"testing"

	"github.com/edanko/dxf/block"
	"github.com/edanko/dxf/entity"
)

func blockRef(t *testing.T, target string) *entity.Insert {
	t.Helper()
	ins := entity.NewInsert()
	ins.SetBlockName(target)
	return ins
}

// TestBlockCycleRepair covers a self-cycle and a two-block cycle. Repair must
// remove the closing INSERT edge, report the cycle, and keep both block
// definitions.
func TestBlockCycleRepair(t *testing.T) {
	t.Run("self cycle", func(t *testing.T) {
		doc := mustLoad(t, baseDXF)
		a := block.NewBlock("CYCLE_A", "")
		a.AddEntity(&beAdapter{Entity: blockRef(t, "CYCLE_A")})
		doc.SetBlocks(doc.Blocks().Add(a))

		rep := Run(doc, nil)
		if !hasIssue(rep, CodeBlockCycle, "CYCLE_A") {
			t.Fatalf("cycle not reported: %+v", rep.Issues)
		}
		rep = Run(doc, &Options{Fix: true})
		fixed := false
		for _, issue := range rep.Issues {
			if issue.Code == CodeBlockCycle && issue.Fixed {
				fixed = true
			}
		}
		if !fixed {
			t.Fatalf("cycle not repaired: %s", rep)
		}
		if findBlock(NewContext(doc), "CYCLE_A") == nil {
			t.Fatal("repair deleted the block definition")
		}
		if len(findBlock(NewContext(doc), "CYCLE_A").Entities()) != 0 {
			t.Fatal("self-reference INSERT was not removed")
		}
	})

	t.Run("two block cycle", func(t *testing.T) {
		doc := mustLoad(t, baseDXF)
		a := block.NewBlock("CYCLE_A", "")
		b := block.NewBlock("CYCLE_B", "")
		a.AddEntity(&beAdapter{Entity: blockRef(t, "CYCLE_B")})
		b.AddEntity(&beAdapter{Entity: blockRef(t, "CYCLE_A")})
		doc.SetBlocks(doc.Blocks().Add(a).Add(b))

		Run(doc, &Options{Fix: true})
		ctx := NewContext(doc)
		if findBlock(ctx, "CYCLE_A") == nil || findBlock(ctx, "CYCLE_B") == nil {
			t.Fatal("repair deleted a block definition")
		}
		if cycles := findBlockCycles(blockGraph(ctx)); len(cycles) != 0 {
			t.Fatalf("cycles remain after repair: %v", cycles)
		}
	})
}
