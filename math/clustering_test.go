package math

import (
	"math"
	"testing"
)

func TestClusterPointsTwoGroups(t *testing.T) {
	points := []Vec2{
		{0, 0}, {1, 0}, {0, 1}, {1, 1},
		{100, 100}, {101, 100}, {100, 101}, {101, 101},
	}
	clusters := ClusterPoints(points, 2.0)
	if len(clusters) != 2 {
		t.Fatalf("expected 2 clusters, got %d", len(clusters))
	}
	if clusters[0].Len() != 4 || clusters[1].Len() != 4 {
		t.Fatalf("expected clusters of 4, got %d and %d", clusters[0].Len(), clusters[1].Len())
	}
}

func TestClusterMembershipPartition(t *testing.T) {
	points := []Vec2{{0, 0}, {0.5, 0}, {10, 10}, {10.5, 10}, {5, 5}}
	assign := ClusterMembership(points, 1.0)
	if len(assign) != len(points) {
		t.Fatalf("membership length %d != points %d", len(assign), len(points))
	}
	// Same cluster iff same id; each point exactly one id by construction.
	if assign[0] != assign[1] {
		t.Fatalf("near points should share a cluster: %v", assign)
	}
	if assign[2] != assign[3] {
		t.Fatalf("near points should share a cluster: %v", assign)
	}
	if assign[0] == assign[2] {
		t.Fatalf("distant points should not share a cluster: %v", assign)
	}
}

func TestClusterCentroid(t *testing.T) {
	points := []Vec2{{0, 0}, {2, 0}, {0, 2}, {2, 2}}
	clusters := ClusterPoints(points, 3.0)
	if len(clusters) != 1 {
		t.Fatalf("expected 1 cluster, got %d", len(clusters))
	}
	c := clusters[0].Centroid()
	if math.Abs(c.X()-1) > 1e-9 || math.Abs(c.Y()-1) > 1e-9 {
		t.Fatalf("centroid = %v, want (1,1)", c)
	}
}

func TestClusterCentroids(t *testing.T) {
	points := []Vec2{{0, 0}, {0, 0}, {10, 10}, {10, 10}}
	clusters := ClusterPoints(points, 1.0)
	centroids := ClusterCentroids(clusters)
	if len(centroids) != 2 {
		t.Fatalf("expected 2 centroids, got %d", len(centroids))
	}
	if centroids[0].Distance(Vec2{0, 0}) > 1e-9 || centroids[1].Distance(Vec2{10, 10}) > 1e-9 {
		t.Fatalf("unexpected centroids: %v", centroids)
	}
}

func TestClusterPointsEmpty(t *testing.T) {
	if got := ClusterPoints(nil, 1.0); got != nil {
		t.Fatalf("expected nil for empty input, got %v", got)
	}
}

func TestClusterPointsEveryPointAssignedOnce(t *testing.T) {
	points := []Vec2{{0, 0}, {0.1, 0.1}, {5, 5}, {5.1, 5.1}, {9, 9}}
	clusters := ClusterPoints(points, 0.5)
	total := 0
	for _, c := range clusters {
		total += c.Len()
	}
	if total != len(points) {
		t.Fatalf("clustered %d points, want %d", total, len(points))
	}
}
