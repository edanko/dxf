package query

import (
	"fmt"
	"github.com/edanko/dxf/entity"
	"math"
)

// Predicate is a compiled query test over an entity context.
type Predicate func(*EntityContext) bool

// CompileQuery parses and compiles an expression into a predicate. The AST is
// validated first: unknown properties, unknown functions, arity mismatches and
// predicate/value confusions are reported with the offending position.
func CompileQuery(input string) (Predicate, error) {
	p, err := NewParser(input)
	if err != nil {
		return nil, err
	}
	expr, err := p.Parse()
	if err != nil {
		return nil, err
	}
	return Compile(expr)
}

// Compile validates an AST and compiles it to a predicate.
func Compile(expr Expression) (Predicate, error) {
	if expr == nil {
		return nil, fmt.Errorf("query: nil expression")
	}
	if err := validate(expr); err != nil {
		return nil, err
	}
	return evalFactory(expr), nil
}

func evalFactory(e Expression) Predicate {
	switch x := e.(type) {
	case *OrExpression:
		l, r := evalFactory(x.Left), evalFactory(x.Right)
		return func(c *EntityContext) bool { return l(c) || r(c) }
	case *AndExpression:
		l, r := evalFactory(x.Left), evalFactory(x.Right)
		return func(c *EntityContext) bool { return l(c) && r(c) }
	case *NotExpression:
		inner := evalFactory(x.Expression)
		return func(c *EntityContext) bool { return !inner(c) }
	case *ComparisonExpression:
		prop := x.Property
		op, val := x.Operator, x.Value
		return func(c *EntityContext) bool {
			return compareValues(c.Property(prop), val, op)
		}
	case *CallExpression:
		fn, args := x.Fn, x.Args
		return func(c *EntityContext) bool {
			r, ok := c.Call(fn, args).(bool)
			return ok && r
		}
	case *FuncComparison:
		fn, args, op, val := x.Fn, x.Args, x.Operator, x.Value
		return func(c *EntityContext) bool {
			return compareValues(c.Call(fn, args), val, op)
		}
	}
	// Fallback for foreign Expression implementations.
	inner := e.Match
	return func(c *EntityContext) bool { return inner(c.Entity) }
}

func validate(e Expression) error {
	switch x := e.(type) {
	case *OrExpression:
		if err := validate(x.Left); err != nil {
			return err
		}
		return validate(x.Right)
	case *AndExpression:
		if err := validate(x.Left); err != nil {
			return err
		}
		return validate(x.Right)
	case *NotExpression:
		return validate(x.Expression)
	case *ComparisonExpression:
		if !isKnownProperty(x.Property) {
			return fmt.Errorf("query: unknown property %q at position %d", x.Property, x.Pos)
		}
		return nil
	case *CallExpression:
		return validateCall(x.Fn, x.Args, x.Pos, true)
	case *FuncComparison:
		if err := validateCall(x.Fn, x.Args, x.Pos, false); err != nil {
			return err
		}
		return nil
	}
	return nil
}

func validateCall(fn string, args []any, pos int, bare bool) error {
	arity, known := queryFunctions[lowerASCII(fn)]
	if !known {
		return fmt.Errorf("query: unknown function %q at position %d", fn, pos)
	}
	if arity >= 0 && len(args) != arity {
		return fmt.Errorf("query: %s expects %d arguments, got %d (position %d)", fn, arity, len(args), pos)
	}
	isBool := boolFunctions[lowerASCII(fn)]
	if bare && !isBool {
		return fmt.Errorf("query: %s produces a value and must be compared with an operator (position %d)", fn, pos)
	}
	if !bare && isBool {
		return fmt.Errorf("query: %s is a predicate and cannot be compared (position %d)", fn, pos)
	}
	return nil
}

func lowerASCII(s string) string {
	out := []rune(s)
	for i, r := range out {
		if r >= 'A' && r <= 'Z' {
			out[i] = r + 32
		}
	}
	return string(out)
}

// SpatialBounds extracts a bounding rectangle forced by the expression's
// positive bbox_intersects/bbox_within calls. Expressions with no spatial
// constraint report ok == false. Used to pre-filter candidates with a
// spatial index before evaluating the full predicate.
func SpatialBounds(e Expression) (minX, minY, maxX, maxY float64, ok bool) {
	minX, minY = -math.MaxFloat64, -math.MaxFloat64
	maxX, maxY = math.MaxFloat64, math.MaxFloat64
	found := false
	var walk func(Expression, bool)
	walk = func(x Expression, positive bool) {
		switch n := x.(type) {
		case *AndExpression:
			walk(n.Left, positive)
			walk(n.Right, positive)
		case *OrExpression:
			if !positive {
				return
			}
			// A disjunction only constrains the box when every branch
			// constrains it, and the union of two boxes is not a box;
			// be conservative and drop OR constraints.
			lx0, ly0, lx1, ly1, lok := SpatialBounds(n.Left)
			rx0, ry0, rx1, ry1, rok := SpatialBounds(n.Right)
			if lok && rok {
				minX = math.Max(minX, math.Min(lx0, rx0))
				minY = math.Max(minY, math.Min(ly0, ry0))
				maxX = math.Min(maxX, math.Max(lx1, rx1))
				maxY = math.Min(maxY, math.Max(ry1, ly1))
				found = true
			}
		case *CallExpression:
			if !positive {
				return
			}
			if r, good := rectArg(n.Args, 4); good {
				applyRect(&minX, &minY, &maxX, &maxY, r)
				found = true
			}
		}
	}
	walk(e, true)
	return minX, minY, maxX, maxY, found
}

// applyRect intersects the accumulated region with r for the positive
// bbox_intersects form (the entity box must intersect r, so the candidate
// search region is exactly r), and widens for bbox_within (entity box inside r
// implies entity geometry inside r as well).
func applyRect(minX, minY, maxX, maxY *float64, r [4]float64) {
	if r[0] > *minX {
		*minX = r[0]
	}
	if r[1] > *minY {
		*minY = r[1]
	}
	if r[2] < *maxX {
		*maxX = r[2]
	}
	if r[3] < *maxY {
		*maxY = r[3]
	}
}

// ExecuteWithFilter parses an expression, applies the spatial pre-filter
// implied by its bounding-box predicates, and evaluates the remainder over
// the surviving candidates.
func ExecuteWithFilter(ents []entity.Entity, input string) (*Selection, error) {
	p, err := NewParser(input)
	if err != nil {
		return nil, err
	}
	expr, err := p.Parse()
	if err != nil {
		return nil, err
	}
	pred, err := Compile(expr)
	if err != nil {
		return nil, err
	}
	minX, minY, maxX, maxY, ok := SpatialBounds(expr)
	return execute(Prefilter(ents, minX, minY, maxX, maxY, ok), pred), nil
}

// Prefilter drops entities whose bounding boxes provably miss the rectangle.
// Entities without bounding boxes are kept conservatively.
func Prefilter(ents []entity.Entity, minX, minY, maxX, maxY float64, hasBounds bool) []entity.Entity {
	if !hasBounds {
		return ents
	}
	out := make([]entity.Entity, 0, len(ents))
	for _, e := range ents {
		c := ContextOf(e)
		if !c.HasBBox {
			out = append(out, e)
			continue
		}
		if c.Max[0] >= minX && c.Min[0] <= maxX && c.Max[1] >= minY && c.Min[1] <= maxY {
			out = append(out, e)
		}
	}
	return out
}
