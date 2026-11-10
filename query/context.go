package query

import (
	"fmt"
	"strings"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/math"
	"github.com/edanko/dxf/xdata"
)

// EntityContext is the compiled view of one entity used by query predicates:
// identity attributes, bounding box, INSERT attributes and xdata apps are
// extracted once so predicates stay cheap and independent of entity models.
type EntityContext struct {
	Entity   entity.Entity
	Type     string
	Layer    string
	Handle   string
	Linetype string
	Text     string
	Color    int
	Min      math.Vec3
	Max      math.Vec3
	Center   math.Vec3
	HasBBox  bool

	attrs  map[string]string
	apps   map[string]xdata.AppData
	hasXD  bool
	builtX bool
}

// ContextOf extracts an evaluation context from an entity.
func ContextOf(e entity.Entity) *EntityContext {
	c := &EntityContext{Entity: e}
	if e == nil {
		return c
	}
	c.Type = e.DXFType()
	c.Handle = e.Handle()
	if ln, ok := e.(interface{ LayerName() string }); ok {
		c.Layer = ln.LayerName()
	} else if l := e.Layer(); l != nil {
		c.Layer = l.Name()
	}
	if lt, ok := e.(interface{ Linetype() string }); ok {
		c.Linetype = lt.Linetype()
	}
	if col, ok := colorValue(e); ok {
		c.Color = col
	}
	if tx, ok := e.(interface{ Text() string }); ok {
		c.Text = tx.Text()
	} else if tx, ok := e.(interface{ Value() string }); ok {
		c.Text = tx.Value()
	}
	c.fillBBox(e)
	return c
}

func colorValue(e entity.Entity) (int, bool) {
	if cl, ok := e.(interface{ Color() color.ColorNumber }); ok {
		return int(cl.Color()), true
	}
	if cl, ok := e.(interface{ Color() int }); ok {
		return cl.Color(), true
	}
	return 0, false
}

// fillBBox reads the entity bounding box, tolerating panics from exotic
// models, and treats the all-zero placeholder box as "no geometry".
func (c *EntityContext) fillBBox(e entity.Entity) {
	min, max := safeBBox(e)
	if len(min) < 2 || len(max) < 2 {
		return
	}
	zero := true
	for i := 0; i < 3 && i < len(min) && i < len(max); i++ {
		if min[i] != 0 || max[i] != 0 {
			zero = false
			break
		}
	}
	if zero {
		return
	}
	c.HasBBox = true
	for i := range 3 {
		if i < len(min) {
			c.Min[i] = min[i]
		}
		if i < len(max) {
			c.Max[i] = max[i]
		}
	}
	c.Center = math.Vec3{
		(c.Min[0] + c.Max[0]) / 2,
		(c.Min[1] + c.Max[1]) / 2,
		(c.Min[2] + c.Max[2]) / 2,
	}
}

func safeBBox(e entity.Entity) (min, max []float64) {
	defer func() {
		if r := recover(); r != nil {
			min, max = nil, nil
		}
	}()
	bb, ok := e.(interface{ BBox() ([]float64, []float64) })
	if !ok {
		return nil, nil
	}
	return bb.BBox()
}

// Attributes returns the INSERT attribute text keyed by (upper-case) tag.
// Non-INSERT entities have no attributes.
func (c *EntityContext) Attributes() map[string]string {
	if !c.builtX {
		c.buildExtras()
	}
	return c.attrs
}

// XDataApps returns the extended-data applications of the entity.
func (c *EntityContext) XDataApps() map[string]xdata.AppData {
	if !c.builtX {
		c.buildExtras()
	}
	return c.apps
}

func (c *EntityContext) buildExtras() {
	c.builtX = true
	c.attrs = map[string]string{}
	c.apps = map[string]xdata.AppData{}
	if ins, ok := c.Entity.(*entity.Insert); ok {
		for _, a := range ins.GetAllAttributes() {
			c.attrs[strings.ToUpper(a.GetTag())] = a.GetText()
		}
	}
	if xd, ok := c.Entity.(interface{ GetXDataObject() *xdata.XData }); ok {
		if x := xd.GetXDataObject(); x != nil && len(x.Apps) > 0 {
			c.hasXD = true
			for k, v := range x.Apps {
				c.apps[strings.ToUpper(k)] = v
			}
		}
	}
}

// Property resolves a comparison property name to its value. Unknown
// properties resolve to nil, which compares unequal to everything.
func (c *EntityContext) Property(name string) any {
	switch strings.ToLower(name) {
	case "layer":
		return c.Layer
	case "type":
		return c.Type
	case "handle":
		return c.Handle
	case "color":
		return c.Color
	case "linetype":
		return c.Linetype
	case "text":
		return c.Text
	}
	return nil
}

func isKnownProperty(name string) bool {
	switch strings.ToLower(name) {
	case "layer", "type", "handle", "color", "linetype", "text":
		return true
	}
	return false
}

// Call evaluates a query function. String results are produced for value
// functions; boolean functions return bool. Unknown functions return nil.
func (c *EntityContext) Call(fn string, args []any) any {
	switch strings.ToLower(fn) {
	case "bbox_intersects", "bboxintersect":
		r, ok := rectArg(args, 4)
		if !ok || !c.HasBBox {
			return false
		}
		return boxesIntersect(c.Min, c.Max, r)
	case "bbox_within":
		r, ok := rectArg(args, 4)
		if !ok || !c.HasBBox {
			return false
		}
		return c.Min[0] >= r[0] && c.Min[1] >= r[1] &&
			c.Max[0] <= r[2] && c.Max[1] <= r[3]
	case "near":
		if len(args) != 3 {
			return false
		}
		x, y, r := toFloat(args[0]), toFloat(args[1]), toFloat(args[2])
		if !c.HasBBox {
			return false
		}
		dx, dy := c.Center[0]-x, c.Center[1]-y
		return dx*dx+dy*dy <= r*r
	case "has_xdata":
		if len(args) != 1 {
			return false
		}
		_, ok := c.XDataApps()[strings.ToUpper(toString(args[0]))]
		return ok
	case "xdata_value":
		if len(args) != 2 {
			return ""
		}
		app, ok := c.XDataApps()[strings.ToUpper(toString(args[0]))]
		if !ok {
			return ""
		}
		code := int(toFloat(args[1]))
		for _, t := range app.Tags {
			if t.Code == code {
				return fmt.Sprint(t.Value)
			}
		}
		return ""
	case "attr", "attribute":
		if len(args) != 1 {
			return ""
		}
		return c.Attributes()[strings.ToUpper(toString(args[0]))]
	}
	return nil
}

func boxesIntersect(min, max math.Vec3, r [4]float64) bool {
	return min[0] <= r[2] && max[0] >= r[0] &&
		min[1] <= r[3] && max[1] >= r[1]
}

func rectArg(args []any, n int) ([4]float64, bool) {
	var r [4]float64
	if len(args) != n {
		return r, false
	}
	for i := range r {
		r[i] = toFloat(args[i])
	}
	if r[0] > r[2] {
		r[0], r[2] = r[2], r[0]
	}
	if r[1] > r[3] {
		r[1], r[3] = r[3], r[1]
	}
	return r, true
}

// queryFunctions describes the callable predicates and their arities.
var queryFunctions = map[string]int{
	"bbox_intersects": 4,
	"bbox_within":     4,
	"near":            3,
	"has_xdata":       1,
	"xdata_value":     2,
	"attr":            1,
	"attribute":       1,
}

// boolFunctions return booleans; the rest produce values comparable with
// the standard operators.
var boolFunctions = map[string]bool{
	"bbox_intersects": true,
	"bbox_within":     true,
	"near":            true,
	"has_xdata":       true,
}
