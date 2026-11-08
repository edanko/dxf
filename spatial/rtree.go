package spatial

import (
	"math"
)

// TreeNode represents a node in the R-tree
type TreeNode struct {
	// Bounding box
	Box *BoundingBox

	// Tree structure
	Children []*TreeNode
	Parent   *TreeNode
	Leaf     bool
	Data     interface{} // For leaf nodes (could be entity reference)
	Removed  bool        // Flag for lazy removal
}

// NewTreeNode creates a new tree node
func NewTreeNode(box *BoundingBox, leaf bool, data interface{}) *TreeNode {
	return &TreeNode{
		Box:      box,
		Children: make([]*TreeNode, 0),
		Parent:   nil,
		Leaf:     leaf,
		Data:     data,
		Removed:  false,
	}
}

// IsLeaf returns true if this is a leaf node
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

// RTree represents an R-tree for spatial indexing
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
func (rt *RTree) Insert(box *BoundingBox, data interface{}) {
	if rt.Root == nil {
		rt.Root = NewTreeNode(box, true, data)
		rt.itemCount++
		return
	}

	rt.insertRecursive(rt.Root, box, data)
	rt.itemCount++
}

func (rt *RTree) insertRecursive(node *TreeNode, box *BoundingBox, data interface{}) {
	if node.IsLeaf() {
		newNode := NewTreeNode(box, true, data)
		newNode.Parent = node
		node.AddChild(newNode)
		node.Box = node.Box.Extend(box)
	} else {
		for _, child := range node.Children {
			if child.Box.ContainsBox(box) {
				rt.insertRecursive(child, box, data)
				child.Box = child.Box.Extend(box)
				return
			}
		}
		newNode := NewTreeNode(box, true, data)
		newNode.Parent = node
		node.AddChild(newNode)
	}
}

// Remove removes an item from the R-tree by data and box
// Returns true if the item was found and removed
func (rt *RTree) Remove(box *BoundingBox, data interface{}) bool {
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

func (rt *RTree) removeRecursive(node *TreeNode, box *BoundingBox, data interface{}) bool {
	if node == nil {
		return false
	}

	if node.IsLeaf() {
		// Search through children for matching item
		for i, child := range node.Children {
			if !child.Removed && child.Data == data && child.Box.ContainsBox(box) {
				child.Removed = true
				node.Children = append(node.Children[:i], node.Children[i+1:]...)

				// Update parent's bounding box
				rt.updateBoundingBoxes(node)

				return true
			}
		}
	} else {
		// Search in children
		for _, child := range node.Children {
			if child.Box.Intersects(box) {
				if rt.removeRecursive(child, box, data) {
					// Update bounding boxes after removal
					rt.updateBoundingBoxes(node)
					return true
				}
			}
		}
	}
	return false
}

// updateBoundingBoxes recalculates bounding boxes from children
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
		data interface{}
	}

	rt.collectItems(rt.Root, &items)

	// Clear the tree
	rt.Root = nil
	rt.removed = 0

	// Rebuild with remaining items
	for _, item := range items {
		rt.Insert(item.box, item.data)
	}
}

// collectItems gathers all non-removed items from the tree
func (rt *RTree) collectItems(node *TreeNode, items *[]struct {
	box  *BoundingBox
	data interface{}
}) {
	if node == nil {
		return
	}

	if node.IsLeaf() {
		for _, child := range node.Children {
			if !child.Removed {
				*items = append(*items, struct {
					box  *BoundingBox
					data interface{}
				}{child.Box, child.Data})
			}
		}
	} else {
		for _, child := range node.Children {
			rt.collectItems(child, items)
		}
	}
}

// RemoveByData removes an item from the R-tree by data only (slower, searches all nodes)
func (rt *RTree) RemoveByData(data interface{}) bool {
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

func (rt *RTree) removeByDataRecursive(node *TreeNode, data interface{}) bool {
	if node == nil {
		return false
	}

	if node.IsLeaf() {
		for i, child := range node.Children {
			if !child.Removed && child.Data == data {
				child.Removed = true
				node.Children = append(node.Children[:i], node.Children[i+1:]...)
				rt.updateBoundingBoxes(node)
				return true
			}
		}
	} else {
		for _, child := range node.Children {
			if rt.removeByDataRecursive(child, data) {
				rt.updateBoundingBoxes(node)
				return true
			}
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
func (rt *RTree) Query(box *BoundingBox) []interface{} {
	var results []interface{}
	rt.queryRecursive(rt.Root, box, &results)
	return results
}

// queryRecursive performs recursive R-tree query
func (rt *RTree) queryRecursive(node *TreeNode, box *BoundingBox, results *[]interface{}) {
	if node == nil {
		return
	}

	if node.Box.Intersects(box) {
		if node.IsLeaf() {
			for _, child := range node.Children {
				if !child.Removed {
					*results = append(*results, child.Data)
				}
			}
		} else {
			for _, child := range node.Children {
				rt.queryRecursive(child, box, results)
			}
		}
	}
}

// queryRange finds all items within the given range
func (rt *RTree) QueryRange(minX, minY, maxX, maxY float64) []interface{} {
	var results []interface{}
	searchBox := NewBoundingBox(minX, minY, maxX, maxY)
	rt.queryRecursive(rt.Root, searchBox, &results)
	return results
}

// queryNearest finds the nearest item to the given point
func (rt *RTree) QueryNearest(x, y float64) (interface{}, float64) {
	var nearest interface{}
	minDist := math.Inf(1)

	rt.queryNearestRecursive(rt.Root, x, y, &nearest, &minDist)
	return nearest, minDist
}

// queryNearestRecursive performs recursive nearest neighbor search
func (rt *RTree) queryNearestRecursive(node *TreeNode, x, y float64, nearest *interface{}, minDist *float64) {
	if node == nil {
		return
	}

	if node.IsLeaf() {
		for _, child := range node.Children {
			if child.Removed {
				continue
			}
			if child.Box != nil {
				centerX, centerY := child.Box.Center()
				dist := math.Sqrt((x-centerX)*(x-centerX) + (y-centerY)*(y-centerY))
				if dist < *minDist {
					*nearest = child.Data
					*minDist = dist
				}
			}
		}
	} else {
		for _, child := range node.Children {
			rt.queryNearestRecursive(child, x, y, nearest, minDist)
		}
	}
}

// splitLeafNode splits a leaf node when it becomes full
func (rt *RTree) splitLeafNode(node *TreeNode, box *BoundingBox, data interface{}) (*TreeNode, *TreeNode) {
	// Create bounding box for all existing items
	allBounds := make([]*BoundingBox, 0, len(node.Children))
	for _, child := range node.Children {
		allBounds = append(allBounds, child.Box)
	}
	allBounds = append(allBounds, node.Box)

	// Find best split axis
	bestAxis, bestSplit := rt.findBestSplit(allBounds)

	var leftNode, rightNode *TreeNode

	// Create left and right nodes
	if bestAxis == 0 {
		// Split along X axis
		leftBox := NewBoundingBox(node.Box.MinX, node.Box.MinY, bestSplit, node.Box.MaxY)
		rightBox := NewBoundingBox(bestSplit, node.Box.MinY, node.Box.MaxX, node.Box.MaxY)

		leftNode = NewTreeNode(leftBox, false, nil)
		rightNode = NewTreeNode(rightBox, false, nil)
	} else {
		// Split along Y axis
		leftBox := NewBoundingBox(node.Box.MinX, node.Box.MinY, node.Box.MaxX, bestSplit)
		rightBox := NewBoundingBox(node.Box.MinX, bestSplit, node.Box.MaxX, node.Box.MaxY)

		leftNode = NewTreeNode(leftBox, false, nil)
		rightNode = NewTreeNode(rightBox, false, nil)
	}

	// Replace node with new children
	node.Children = nil // Clear existing children
	node.AddChild(leftNode)
	if rightNode != nil {
		node.AddChild(rightNode)
	}

	// Add new items
	node.AddChild(NewTreeNode(box, true, data))

	return leftNode, rightNode
}

// findBestSplit finds the best axis and position to split bounding boxes
func (rt *RTree) findBestSplit(boxes []*BoundingBox) (int, float64) {
	if len(boxes) == 0 {
		return 0, 0
	}

	// Calculate overall bounds
	minX, minY := boxes[0].MinX, boxes[0].MinY
	maxX, maxY := boxes[0].MaxX, boxes[0].MaxY

	for _, box := range boxes[1:] {
		minX = math.Min(minX, box.MinX)
		minY = math.Min(minY, box.MinY)
		maxX = math.Max(maxX, box.MaxX)
		maxY = math.Max(maxY, box.MaxY)
	}

	// Choose split axis with largest range
	xRange := maxX - minX
	yRange := maxY - minY

	if xRange > yRange {
		// Split on X axis
		return 0, (minX + maxX) / 2
	} else {
		// Split on Y axis
		return 1, (minY + maxY) / 2
	}
}

// Count returns the number of items in the R-tree
func (rt *RTree) Count() int {
	return rt.itemCount
}

// GetStats returns R-tree statistics
func (rt *RTree) GetStats() map[string]interface{} {
	totalItems := rt.Count()
	treeDepth := rt.calculateDepth(rt.Root)
	avgItemsPerNode := float64(totalItems) / float64(treeDepth)

	return map[string]interface{}{
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

	if node.IsLeaf() {
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
}

// BulkInsert adds multiple items to the R-tree
func (rt *RTree) BulkInsert(items []*BoundingBox, dataList []interface{}) {
	for i, box := range items {
		if i < len(dataList) {
			rt.Insert(box, dataList[i])
		}
	}
}

// BulkQuery performs multiple range queries
func (rt *RTree) BulkQuery(boxes []*BoundingBox) [][]interface{} {
	results := make([][]interface{}, 0, len(boxes))
	for i, box := range boxes {
		results[i] = rt.Query(box)
	}
	return results
}
