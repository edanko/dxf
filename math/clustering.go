package math

// Cluster is a group of points produced by ClusterPoints. Points are stored in
// the order they appeared in the input.
type Cluster struct {
	Points []Vec2
}

// Len returns the number of points in the cluster.
func (c Cluster) Len() int { return len(c.Points) }

// Centroid returns the arithmetic mean of the cluster's points. An empty
// cluster returns the zero vector.
func (c Cluster) Centroid() Vec2 {
	if len(c.Points) == 0 {
		return Vec2{}
	}
	var sum Vec2
	for _, p := range c.Points {
		sum = sum.Add(p)
	}
	return sum.Div(float64(len(c.Points)))
}

// ClusterPoints groups points using single-linkage clustering: two points join
// the same cluster when the distance between them is at most maxDistance.
// Clusters are returned in order of first appearance in points.
func ClusterPoints(points []Vec2, maxDistance float64) []Cluster {
	assignments := ClusterMembership(points, maxDistance)
	if len(points) == 0 {
		return nil
	}

	// Cluster ids are compacted by first appearance for deterministic output.
	clusters := make([]Cluster, 0)
	index := make(map[int]int) // raw id -> cluster slice index
	for i, id := range assignments {
		idx, ok := index[id]
		if !ok {
			idx = len(clusters)
			index[id] = idx
			clusters = append(clusters, Cluster{})
		}
		clusters[idx].Points = append(clusters[idx].Points, points[i])
	}
	return clusters
}

// ClusterMembership returns a cluster id for each point, where points that are
// within maxDistance of each other (directly or transitively) share an id.
// Every point belongs to exactly one cluster.
func ClusterMembership(points []Vec2, maxDistance float64) []int {
	n := len(points)
	parent := make([]int, n)
	for i := range parent {
		parent[i] = i
	}
	var find func(int) int
	find = func(x int) int {
		for parent[x] != x {
			parent[x] = parent[parent[x]]
			x = parent[x]
		}
		return x
	}
	union := func(a, b int) {
		ra, rb := find(a), find(b)
		if ra != rb {
			parent[rb] = ra
		}
	}

	threshold := maxDistance
	for i := range n {
		for j := i + 1; j < n; j++ {
			if points[i].DistanceSquared(points[j]) <= threshold*threshold {
				union(i, j)
			}
		}
	}

	assignments := make([]int, n)
	for i := range assignments {
		assignments[i] = find(i)
	}
	return assignments
}

// ClusterCentroids returns one centroid per cluster produced by ClusterPoints.
func ClusterCentroids(clusters []Cluster) []Vec2 {
	centroids := make([]Vec2, len(clusters))
	for i, c := range clusters {
		centroids[i] = c.Centroid()
	}
	return centroids
}
