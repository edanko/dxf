package audit

import (
	"sort"
	"strconv"
	"strings"
)

func init() {
	RegisterFixer(CodeDuplicateHandle, fixDuplicateHandle)
}

type handleStringSetter interface {
	SetHandleString(string)
}

// fixDuplicateHandle keeps the first object registered for a handle and
// assigns deterministic fresh handles to later duplicates. References are only
// rewritten when their target is unambiguous; ambiguous references are left
// untouched and remain visible through the unresolved-owner/reactor checks.
func fixDuplicateHandle(ctx *Context, issue Issue) bool {
	if issue.Handle == "" {
		return false
	}
	objs := ctx.Objects(issue.Handle)
	if len(objs) < 2 {
		return false
	}
	sort.SliceStable(objs, func(i, j int) bool { return Describe(objs[i]) < Describe(objs[j]) })

	next := maxNumericHandle(ctx.Handles()) + 1
	fresh := func() string {
		for {
			h := strings.ToUpper(strconv.FormatInt(int64(next), 16))
			next++
			if !ctx.hasHandle(h) {
				return h
			}
		}
	}

	changed := false
	for _, obj := range objs[1:] {
		setter, ok := obj.(handleStringSetter)
		if !ok {
			continue
		}
		setter.SetHandleString(fresh())
		changed = true
	}
	return changed
}

func maxNumericHandle(handles []string) int {
	max := 0
	for _, h := range handles {
		v, err := strconv.ParseUint(h, 16, 64)
		if err == nil && int(v) > max {
			max = int(v)
		}
	}
	return max
}
