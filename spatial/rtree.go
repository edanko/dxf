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
}

// NewTreeNode creates a new tree node
func NewTreeNode(box *BoundingBox, leaf bool, data interface{}) *TreeNode {
	return &TreeNode{
		Box:      box,
		Children: make([]*TreeNode, 0),
		Parent:   nil,
		Leaf:     leaf,
		Data:     data,
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

// RTree represents an R-tree for spatial indexing
type RTree struct {
	Root     *TreeNode
	MaxItems int
	Depth    int
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
		return
	}

	rt.insertRecursive(rt.Root, box, data)
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
			*results = append(*results, node.Data)
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
		// Check if this leaf has a nearer item
		if node.Box != nil {
			centerX, centerY := node.Box.Center()
			dist := math.Sqrt((x-centerX)*(x-centerX) + (y-centerY)*(y-centerY))
			if dist < *minDist {
				*nearest = node.Data
				*minDist = dist
			}
		}
	} else {
		// Check children
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
	return rt.countNodes(rt.Root)
}

// countNodes recursively counts nodes in the tree
func (rt *RTree) countNodes(node *TreeNode) int {
	if node == nil {
		return 0
	}

	count := 1 // Count this node

	for _, child := range node.Children {
		count += rt.countNodes(child)
	}

	return count
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
