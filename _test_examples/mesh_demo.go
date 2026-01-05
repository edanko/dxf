package main

import (
	"fmt"
	"log"
	"math"

	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
)

func main() {
	fmt.Println("=== DXF MESH Entity Demo ===\n")

	// Create a new drawing
	d, err := drawing.New()
	if err != nil {
		log.Fatal(err)
	}

	// Add basic layer
	d.AddLayer("0", 7, nil, false)

	fmt.Println("1. Creating basic mesh geometries...")

	// Create a cube mesh
	cubeMesh := entity.CreateCube(10.0)
	d.AddEntity(cubeMesh)
	fmt.Printf("   Created cube mesh: %d vertices, %d faces\n",
		cubeMesh.VertexCount(), cubeMesh.FaceCount())

	// Create a pyramid mesh
	pyramidMesh := entity.CreatePyramid(8.0, 12.0)
	d.AddEntity(pyramidMesh)
	fmt.Printf("   Created pyramid mesh: %d vertices, %d faces\n",
		pyramidMesh.VertexCount(), pyramidMesh.FaceCount())

	// Create a sphere mesh
	sphereMesh := entity.CreateSphere(5.0, 16)
	d.AddEntity(sphereMesh)
	fmt.Printf("   Created sphere mesh: %d vertices, %d faces\n",
		sphereMesh.VertexCount(), sphereMesh.FaceCount())

	fmt.Println("\n2. Testing mesh manipulation methods...")

	// Test vertex operations
	fmt.Println("\n2.1 Testing vertex operations:")
	testVertex := dxfmath.NewVec3(1, 2, 3)
	cubeMesh.AddVertex(testVertex)
	fmt.Printf("   Added vertex: (%.1f, %.1f, %.1f)\n", testVertex.X(), testVertex.Y(), testVertex.Z())
	fmt.Printf("   Total vertices: %d\n", cubeMesh.VertexCount())

	// Test face operations
	fmt.Println("\n2.2 Testing face operations:")
	testFace := []int{0, 2, 4} // Triangular face
	cubeMesh.AddFace(testFace)
	fmt.Printf("   Added face: %v\n", testFace)
	fmt.Printf("   Total faces: %d\n", cubeMesh.FaceCount())

	// Test edge operations (wireframe)
	fmt.Println("\n2.3 Testing edge operations:")
	testEdge := []int{0, 1} // Edge between vertices 0 and 1
	cubeMesh.AddEdge(testEdge)
	fmt.Printf("   Added edge: %v\n", testEdge)
	fmt.Printf("   Total edges: %d\n", cubeMesh.EdgeCount())

	fmt.Println("\n3. Testing mesh properties...")
	printMeshInfo(cubeMesh, "Cube Mesh")
	printMeshInfo(pyramidMesh, "Pyramid Mesh")
	printMeshInfo(sphereMesh, "Sphere Mesh")

	fmt.Println("\n4. Testing advanced mesh features...")

	// Test vertex colors
	fmt.Println("\n4.1 Testing vertex colors:")
	coloredMesh := entity.CreateCube(8.0)

	// Set vertex colors for the cube
	colors := []int{1, 2, 3, 4, 5, 6, 7, 8} // Different colors for each vertex
	coloredMesh.SetVertexColors(colors)

	// Enable vertex colors
	coloredMesh.Color = 62 // Use color override to enable vertex colors

	d.AddEntity(coloredMesh)
	fmt.Printf("   Created colored cube mesh with %d vertex colors\n", len(colors))

	// Test edge creases
	fmt.Println("\n4.2 Testing edge creases:")
	creaseMesh := entity.CreateCube(6.0)

	// Set crease values for edges (makes it look faceted)
	creaseValues := []float64{0.8, 0.8, 0.8} // Strong creases
	creaseMesh.SetCreaseValues(creaseValues)

	d.AddEntity(creaseMesh)
	fmt.Printf("   Created mesh with %d edge crease values\n", len(creaseValues))

	// Test smoothing
	fmt.Println("\n4.3 Testing mesh smoothing:")
	smoothMesh := entity.CreateCube(7.0)
	smoothMesh.BlendCrease = entity.MeshFlagCubicSmooth // Use cubic smoothing

	d.AddEntity(smoothMesh)
	fmt.Printf("   Created mesh with cubic smoothing\n")

	// Test subdivision levels
	fmt.Println("\n4.4 Testing subdivision levels:")
	subdivMesh := entity.CreateCube(5.0)
	subdivMesh.SubdivisionLevels = 2 // Apply subdivision
	subdivMesh.BlendCrease = entity.MeshFlagQuadSmooth

	d.AddEntity(subdivMesh)
	fmt.Printf("   Created subdivided mesh with %d subdivision levels\n", subdivMesh.SubdivisionLevels)

	fmt.Println("\n5. Testing mesh transformations...")

	// Test bounding box calculation
	fmt.Println("\n5.1 Testing bounding box:")
	min, max, found := cubeMesh.BBox()
	if found {
		fmt.Printf("   Bounding box: (%.2f, %.2f, %.2f) to (%.2f, %.2f, %.2f)\n",
			min.X(), min.Y(), min.Z(),
			max.X(), max.Y(), max.Z())
	} else {
		fmt.Printf("   Empty mesh - no bounding box\n")
	}

	fmt.Println("\n5.2 Testing mesh validation:")
	if cubeMesh.IsEmpty() {
		fmt.Printf("   Empty mesh: true\n")
	} else {
		fmt.Printf("   Empty mesh: false (%d vertices, %d faces)\n",
			cubeMesh.VertexCount(), cubeMesh.FaceCount())
	}

	fmt.Println("\n6. Entity information summary...")
	fmt.Printf("   Total entities created: %d\n", d.Entities.Len())
	fmt.Printf("   MESH entities: %d\n", countMeshEntities(d))

	// Save drawing
	fmt.Println("\n7. Saving drawing to 'mesh_demo.dxf'...")
	err = d.SaveAs("mesh_demo.dxf")
	if err != nil {
		log.Printf("Error saving drawing: %v", err)
	} else {
		fmt.Println("   Drawing saved successfully!")
	}

	fmt.Println("\n=== MESH Entity Demo Complete ===")
	fmt.Println("\nKey Features Demonstrated:")
	fmt.Println("✓ MESH entity creation and management")
	fmt.Println("✓ Vertex and face operations")
	fmt.Println("✓ Edge operations (wireframe support)")
	fmt.Println("✓ Vertex colors and materials")
	fmt.Println("✓ Edge creases and smoothing")
	fmt.Println("✓ Subdivision levels")
	fmt.Println("✓ Bounding box calculation")
	fmt.Println("✓ Mesh validation and analysis")
	fmt.Println("✓ Professional mesh primitives (cube, pyramid, sphere)")
	fmt.Println("✓ DXF formatting with all MESH features")
	fmt.Println("✓ Complete MESH entity implementation")
}

func printMeshInfo(mesh *entity.Mesh, description string) {
	fmt.Printf("\n   %s:\n", description)
	fmt.Printf("     Vertices: %d\n", mesh.VertexCount())
	fmt.Printf("     Faces: %d\n", mesh.FaceCount())
	fmt.Printf("     Edges: %d\n", mesh.EdgeCount())
	fmt.Printf("     Version: %d\n", mesh.Version)
	fmt.Printf("     Blend Crease: %d\n", mesh.BlendCrease)
	fmt.Printf("     Subdivision Levels: %d\n", mesh.SubdivisionLevels)
	fmt.Printf("     Vertex Colors: %t\n", len(mesh.VertexColors()) > 0)
	fmt.Printf("     Edge Creases: %t\n", len(mesh.CreaseValues()) > 0)
	if mesh.Layer() != "" {
		fmt.Printf("     Layer Override: %s\n", mesh.Layer())
	}
}

func countMeshEntities(d *drawing.Drawing) int {
	count := 0
	for _, ent := range d.Entities {
		if _, ok := ent.(*entity.Mesh); ok {
			count++
		}
	}
	return count
}
