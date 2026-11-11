package spatial

import (
	"math"
)

// TreeNode represents a node in the R-tree. Internal nodes (Leaf == false)
// organize the tree and carry a bounding box covering their children; item
// nodes (Leaf == true) hold a single indexed item in Data together with its
// bounding box.
type TreeNode struct {
	// Bounding box
	Box *BoundingBox

	// Tree structure
	Children []*TreeNode
	Parent   *TreeNode
	Leaf     bool
	Data     any  // Item for leaf nodes (could be entity reference)
	Removed  bool // Flag for lazy removal
}

// NewTreeNode creates a new tree node
func NewTreeNode(box *BoundingBox, leaf bool, data any) *TreeNode {
	return &TreeNode{
		Box:      box,
		Children: make([]*TreeNode, 0),
		Parent:   nil,
		Leaf:     leaf,
		Data:     data,
		Removed:  false,
	}
}

// IsLeaf returns true if this is an item node
func (n *TreeNode) IsLeaf() bool {
	return n.Leaf
}

// AddChild adds a child node
func (n *TreeNode) AddChild(child *TreeNode) {
	n.Children = append(n.Children, child)
}

// ReplaceChild replaces a child node with a new one
func (n *TreeNode) ReplaceChild(oldChild, newChild *TreeNode) bool {
	for i, child := range n.Children {
		if child == oldChild {
			n.Children[i] = newChild
			return true
		}
	}
	return false
}

// RemoveChild removes a child node
func (n *TreeNode) RemoveChild(child *TreeNode) bool {
	for i, c := range n.Children {
		if c == child {
			n.Children = append(n.Children[:i], n.Children[i+1:]...)
			return true
		}
	}
	return false
}

// RTree represents an R-tree for spatial indexing. Item nodes are held
// under internal nodes; an internal node's bounding box always covers the
// boxes of its children.
type RTree struct {
	Root      *TreeNode
	MaxItems  int
	Depth     int
	itemCount int // Track number of items
	removed   int // Track number of removed items for lazy cleanup
}

// RTreeOptions contains configuration options for R-tree
type RTreeOptions struct {
	MaxItemsPerNode int // Maximum items per node before splitting
	MaxDepth        int // Maximum tree depth
	MinItems        int // Minimum items before creating a new node
}

// DefaultRTreeOptions returns default R-tree options
func DefaultRTreeOptions() *RTreeOptions {
	return &RTreeOptions{
		MaxItemsPerNode: 10,
		MaxDepth:        8,
		MinItems:        4,
	}
}

// NewRTree creates a new R-tree
func NewRTree(options *RTreeOptions) *RTree {
	if options == nil {
		options = DefaultRTreeOptions()
	}

	return &RTree{
		Root:     nil,
		MaxItems: options.MaxItemsPerNode,
		Depth:    options.MaxDepth,
	}
}

// Insert adds an item to the R-tree
func (rt *RTree) Insert(box *BoundingBox, data any) {
	if rt.Root == nil {
		// The root is an internal node that holds item nodes as children,
		// so queries and removals treat every item the same way.
		rt.Root = NewTreeNode(box, false, nil)
	}

	rt.insertRecursive(rt.Root, box, data)
	rt.itemCount++
}

// insertRecursive adds an item node under the given internal node,
// descending into containing internal children when possible.
func (rt *RTree) insertRecursive(node *TreeNode, box *BoundingBox, data any) {
	for _, child := range node.Children {
		if !child.Leaf && child.Box.ContainsBox(box) {
			rt.insertRecursive(child, box, data)
			return
		}
	}
	newNode := NewTreeNode(box, true, data)
	newNode.Parent = node
	node.AddChild(newNode)
	node.Box = node.Box.Extend(box)
}

// Remove removes an item from the R-tree by data and box
// Returns true if the item was found and removed
func (rt *RTree) Remove(box *BoundingBox, data any) bool {
	if rt.Root == nil {
		return false
	}

	if rt.removeRecursive(rt.Root, box, data) {
		rt.itemCount--
		rt.removed++

		// Lazy cleanup if too many removed items
		if rt.removed > rt.itemCount/2 && rt.itemCount > 10 {
			rt.cleanup()
		}
		return true
	}
	return false
}

func (rt *RTree) removeRecursive(node *TreeNode, box *BoundingBox, data any) bool {
	if node == nil {
		return false
	}

	if node.Leaf {
		if !node.Removed && node.Data == data && (box == nil || node.Box.ContainsBox(box)) {
			node.Removed = true
			if node.Parent != nil {
				node.Parent.RemoveChild(node)
				rt.updateBoundingBoxes(node.Parent)
			}
			return true
		}
		return false
	}

	for _, child := range node.Children {
		if box == nil || child.Box.Intersects(box) {
			if rt.removeRecursive(child, box, data) {
				return true
			}
		}
	}
	return false
}

// updateBoundingBoxes recalculates bounding boxes from children and
// propagates the change up to the root.
func (rt *RTree) updateBoundingBoxes(node *TreeNode) {
	if node == nil || len(node.Children) == 0 {
		return
	}

	// Recalculate bounding box from children
	newBox := node.Children[0].Box
	for _, child := range node.Children[1:] {
		newBox = newBox.Extend(child.Box)
	}
	node.Box = newBox

	// Propagate up if needed
	if node.Parent != nil {
		rt.updateBoundingBoxes(node.Parent)
	}
}

// cleanup removes marked items and rebuilds the tree if needed
func (rt *RTree) cleanup() {
	if rt.Root == nil {
		return
	}

	// Collect all non-removed items
	var items []struct {
		box  *BoundingBox
		data any
	}

	rt.collectItems(rt.Root, &items)

	// Clear the tree
	rt.Root = nil
	rt.removed = 0
	rt.itemCount = 0

	// Rebuild with remaining items
	for _, item := range items {
		rt.Insert(item.box, item.data)
	}
}

// collectItems gathers all non-removed items from the tree
func (rt *RTree) collectItems(node *TreeNode, items *[]struct {
	box  *BoundingBox
	data any
}) {
	if node == nil {
		return
	}

	if node.Leaf {
		if !node.Removed && node.Data != nil {
			*items = append(*items, struct {
				box  *BoundingBox
				data any
			}{node.Box, node.Data})
		}
		return
	}

	for _, child := range node.Children {
		rt.collectItems(child, items)
	}
}

// RemoveByData removes an item from the R-tree by data only (slower, searches all nodes)
func (rt *RTree) RemoveByData(data any) bool {
	if rt.Root == nil {
		return false
	}

	if rt.removeByDataRecursive(rt.Root, data) {
		rt.itemCount--
		rt.removed++

		if rt.removed > rt.itemCount/2 && rt.itemCount > 10 {
			rt.cleanup()
		}
		return true
	}
	return false
}

func (rt *RTree) removeByDataRecursive(node *TreeNode, data any) bool {
	if node == nil {
		return false
	}

	if node.Leaf {
		if !node.Removed && node.Data == data {
			node.Removed = true
			if node.Parent != nil {
				node.Parent.RemoveChild(node)
				rt.updateBoundingBoxes(node.Parent)
			}
			return true
		}
		return false
	}

	for _, child := range node.Children {
		if rt.removeByDataRecursive(child, data) {
			return true
		}
	}
	return false
}

// Compact removes all marked items and rebuilds the tree
func (rt *RTree) Compact() {
	rt.cleanup()
}

// ItemCount returns the number of items in the tree (excluding removed)
func (rt *RTree) ItemCount() int {
	return rt.itemCount
}

// TotalCount returns the total number of items including removed
func (rt *RTree) TotalCount() int {
	return rt.itemCount + rt.removed
}

// Query finds all items that intersect with the given box
func (rt *RTree) Query(box *BoundingBox) []any {
	var results []any
	rt.queryRecursive(rt.Root, box, &results)
	return results
}

// queryRecursive performs recursive R-tree query
func (rt *RTree) queryRecursive(node *TreeNode, box *BoundingBox, results *[]any) {
	if node == nil || node.Removed {
		return
	}

	if node.Leaf {
		// Item node: return the item when its box intersects the query.
		if node.Data != nil && node.Box.Intersects(box) {
			*results = append(*results, node.Data)
		}
		return
	}

	if node.Box != nil && !node.Box.Intersects(box) {
		return
	}
	for _, child := range node.Children {
		rt.queryRecursive(child, box, results)
	}
}

// QueryRange finds all items within the given range
func (rt *RTree) QueryRange(minX, minY, maxX, maxY float64) []any {
	var results []any
	searchBox := NewBoundingBox(minX, minY, maxX, maxY)
	rt.queryRecursive(rt.Root, searchBox, &results)
	return results
}

// QueryNearest finds the nearest item to the given point, using the
// distance from the point to each item's bounding box.
func (rt *RTree) QueryNearest(x, y float64) (any, float64) {
	var nearest any
	minDist := math.Inf(1)

	rt.queryNearestRecursive(rt.Root, x, y, &nearest, &minDist)
	return nearest, minDist
}

// queryNearestRecursive performs recursive nearest neighbor search
func (rt *RTree) queryNearestRecursive(node *TreeNode, x, y float64, nearest *any, minDist *float64) {
	if node == nil || node.Removed {
		return
	}

	if node.Leaf {
		if node.Data != nil && node.Box != nil {
			dist := node.Box.DistanceToPoint(x, y)
			if dist < *minDist {
				*nearest = node.Data
				*minDist = dist
			}
		}
		return
	}

	// Prune subtrees whose bounds are farther than the best match so far.
	if node.Box != nil && node.Box.DistanceToPoint(x, y) > *minDist {
		return
	}
	for _, child := range node.Children {
		rt.queryNearestRecursive(child, x, y, nearest, minDist)
	}
}

// Count returns the number of items in the R-tree
func (rt *RTree) Count() int {
	return rt.itemCount
}

// GetStats returns R-tree statistics
func (rt *RTree) GetStats() map[string]any {
	totalItems := rt.Count()
	treeDepth := rt.calculateDepth(rt.Root)
	avgItemsPerNode := float64(totalItems) / float64(treeDepth)

	return map[string]any{
		"total_items":        totalItems,
		"tree_depth":         treeDepth,
		"avg_items_per_node": avgItemsPerNode,
		"max_items_per_node": rt.MaxItems,
	}
}

// calculateDepth calculates the maximum tree depth
func (rt *RTree) calculateDepth(node *TreeNode) int {
	if node == nil {
		return 0
	}

	if node.Leaf {
		return 1
	}

	maxChildDepth := 0
	for _, child := range node.Children {
		childDepth := rt.calculateDepth(child)
		if childDepth > maxChildDepth {
			maxChildDepth = childDepth
		}
	}

	return maxChildDepth + 1
}

// Clear removes all items from the R-tree
func (rt *RTree) Clear() {
	rt.Root = nil
	rt.itemCount = 0
	rt.removed = 0
}

// BulkInsert adds multiple items to the R-tree
func (rt *RTree) BulkInsert(items []*BoundingBox, dataList []any) {
	for i, box := range items {
		if i < len(dataList) {
			rt.Insert(box, dataList[i])
		}
	}
}

// BulkQuery performs multiple range queries
func (rt *RTree) BulkQuery(boxes []*BoundingBox) [][]any {
	results := make([][]any, 0, len(boxes))
	for _, box := range boxes {
		results = append(results, rt.Query(box))
	}
	return results
}
