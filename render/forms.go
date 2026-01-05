package render

import (
	"math"

	dxfmath "github.com/edanko/dxf/math"
)

// FormType represents different geometric shape types
type FormType int

const (
	FormCircle FormType = iota
	FormEllipse
	FormSquare
	FormRectangle
	FormTriangle
	FormStar
	FormNGon
	FormArrow
	FormHelix
	FormCube
	FormCylinder
	FormExtrude
	FormBox
	FormGear
	FormCone
	FormSphere
	FormTorus
)

// FormGenerator creates various geometric forms and shapes
type FormGenerator struct {
	scale    float64
	segments int
}

// NewFormGenerator creates a new form generator
func NewFormGenerator() *FormGenerator {
	return &FormGenerator{
		scale:    1.0,
		segments: 32, // Default segment count for curves
	}
}

// SetScale sets the scaling factor for forms
func (fg *FormGenerator) SetScale(scale float64) {
	fg.scale = scale
}

// SetSegments sets the number of segments for curved forms
func (fg *FormGenerator) SetSegments(segments int) {
	fg.segments = segments
}

// Circle generates a circle form
func (fg *FormGenerator) Circle(radius float64) []dxfmath.Vec3 {
	points := make([]dxfmath.Vec3, fg.segments)
	for i := 0; i < fg.segments; i++ {
		angle := 2.0 * math.Pi * float64(i) / float64(fg.segments)
		x := radius * math.Cos(angle)
		y := radius * math.Sin(angle)
		z := 0.0
		points[i] = dxfmath.NewVec3(x, y, z)
	}
	return points
}

// Ellipse generates an ellipse form
func (fg *FormGenerator) Ellipse(rx, ry float64) []dxfmath.Vec3 {
	points := make([]dxfmath.Vec3, fg.segments)
	for i := 0; i < fg.segments; i++ {
		angle := 2.0 * math.Pi * float64(i) / float64(fg.segments)
		x := rx * math.Cos(angle)
		y := ry * math.Sin(angle)
		z := 0.0
		points[i] = dxfmath.NewVec3(x, y, z)
	}
	return points
}

// Square generates a square form
func (fg *FormGenerator) Square(size float64, close ...bool) []dxfmath.Vec3 {
	halfSize := size / 2.0
	shouldClose := len(close) > 0 && close[0]

	vertices := []dxfmath.Vec3{
		dxfmath.NewVec3(-halfSize, -halfSize, 0),
		dxfmath.NewVec3(halfSize, -halfSize, 0),
		dxfmath.NewVec3(halfSize, halfSize, 0),
		dxfmath.NewVec3(-halfSize, halfSize, 0),
	}

	if shouldClose {
		vertices = append(vertices, vertices[0]) // Close polygon
	}

	return vertices
}

// Box generates a rectangular box form (alternative to Rectangle)
func (fg *FormGenerator) Box(width, height float64, close ...bool) []dxfmath.Vec3 {
	return fg.Rectangle(width, height, close...)
}

// Rectangle generates a rectangle form
func (fg *FormGenerator) Rectangle(width, height float64, close ...bool) []dxfmath.Vec3 {
	halfWidth := width / 2.0
	halfHeight := height / 2.0
	shouldClose := len(close) > 0 && close[0]

	vertices := []dxfmath.Vec3{
		dxfmath.NewVec3(-halfWidth, -halfHeight, 0),
		dxfmath.NewVec3(halfWidth, -halfHeight, 0),
		dxfmath.NewVec3(halfWidth, halfHeight, 0),
		dxfmath.NewVec3(-halfWidth, halfHeight, 0),
	}

	if shouldClose {
		vertices = append(vertices, vertices[0]) // Close the polygon
	}

	return vertices
}

// Triangle generates an equilateral triangle form
func (fg *FormGenerator) Triangle(size float64, close ...bool) []dxfmath.Vec3 {
	height := size * math.Sqrt(3.0) / 2.0
	halfSize := size / 2.0
	shouldClose := len(close) > 0 && close[0]

	vertices := []dxfmath.Vec3{
		dxfmath.NewVec3(0, 2.0*height/3.0, 0),      // Top vertex
		dxfmath.NewVec3(-halfSize, -height/3.0, 0), // Bottom left
		dxfmath.NewVec3(halfSize, -height/3.0, 0),  // Bottom right
	}

	if shouldClose {
		vertices = append(vertices, vertices[0]) // Close polygon
	}

	return vertices
}

// Star generates a star form
func (fg *FormGenerator) Star(outerRadius, innerRadius float64, points int) []dxfmath.Vec3 {
	vertices := make([]dxfmath.Vec3, points*2+2)

	angle := 2.0 * math.Pi / float64(points)

	for i := 0; i < points*2+2; i++ {
		var r float64
		if i%2 == 0 {
			r = outerRadius // Outer vertex
		} else {
			r = innerRadius // Inner vertex
		}

		theta := float64(i) * angle / 2.0
		x := r * math.Cos(theta)
		y := r * math.Sin(theta)
		vertices[i] = dxfmath.NewVec3(x, y, 0)
	}

	return vertices
}

// NGon generates an n-sided polygon (regular polygon)
func (fg *FormGenerator) NGon(sides int, radius float64, close ...bool) []dxfmath.Vec3 {
	shouldClose := len(close) > 0 && close[0]

	vertices := make([]dxfmath.Vec3, sides)

	angle := 2.0 * math.Pi / float64(sides)

	for i := 0; i < sides; i++ {
		theta := float64(i) * angle
		x := radius * math.Cos(theta)
		y := radius * math.Sin(theta)
		vertices[i] = dxfmath.NewVec3(x, y, 0)
	}

	if shouldClose {
		vertices = append(vertices, vertices[0]) // Close polygon
	}

	return vertices
}

// Arrow generates an arrow form (pointing right)
func (fg *FormGenerator) Arrow(length, headSize float64, close ...bool) []dxfmath.Vec3 {
	shouldClose := len(close) > 0 && close[0]

	vertices := []dxfmath.Vec3{
		dxfmath.NewVec3(-length, -headSize/2, 0),         // Stem start (left)
		dxfmath.NewVec3(length-headSize, -headSize/2, 0), // Stem end (right, before head)
		dxfmath.NewVec3(length, 0, 0),                    // Head tip (right)
		dxfmath.NewVec3(length-headSize, headSize/2, 0),  // Head bottom right
		dxfmath.NewVec3(-length, headSize/2, 0),          // Head bottom left
	}

	if shouldClose {
		vertices = append(vertices, vertices[0]) // Close polygon
	}

	return vertices
}

// OpenArrow generates an open arrow form (no closing line)
func (fg *FormGenerator) OpenArrow(size float64, angle float64) []dxfmath.Vec3 {
	// Similar to Python ezdxf's open_arrow function
	h := size * math.Sin(angle/2.0) / 2.0

	return []dxfmath.Vec3{
		dxfmath.NewVec3(-size, h, 0),  // Left top
		dxfmath.NewVec3(0, 0, 0),      // Center point (tip)
		dxfmath.NewVec3(-size, -h, 0), // Left bottom
	}
}

// Helix generates a helix form
func (fg *FormGenerator) Helix(radius, height, turns float64, pointsPerTurn int) []dxfmath.Vec3 {
	totalPoints := int(turns) * pointsPerTurn
	vertices := make([]dxfmath.Vec3, totalPoints)

	for i := 0; i < totalPoints; i++ {
		t := float64(i) / float64(pointsPerTurn)
		angle := 2.0 * math.Pi * t
		y := height * t / turns

		x := radius * math.Cos(angle)
		z := radius * math.Sin(angle)

		vertices[i] = dxfmath.NewVec3(x, y, z)
	}

	return vertices
}

// Translate translates form vertices by given offset
func (fg *FormGenerator) Translate(vertices []dxfmath.Vec3, offset dxfmath.Vec3) []dxfmath.Vec3 {
	result := make([]dxfmath.Vec3, len(vertices))
	for i, v := range vertices {
		result[i] = v.Add(offset)
	}
	return result
}

// Rotate rotates form vertices around origin
func (fg *FormGenerator) Rotate(vertices []dxfmath.Vec3, angle float64) []dxfmath.Vec3 {
	cosA := math.Cos(angle)
	sinA := math.Sin(angle)

	result := make([]dxfmath.Vec3, len(vertices))
	for i, v := range vertices {
		x := v.X()*cosA - v.Y()*sinA
		y := v.X()*sinA + v.Y()*cosA
		result[i] = dxfmath.NewVec3(x, y, v.Z())
	}

	return result
}

// Scale scales form vertices
func (fg *FormGenerator) Scale(vertices []dxfmath.Vec3, scaleFactor float64) []dxfmath.Vec3 {
	result := make([]dxfmath.Vec3, len(vertices))
	for i, v := range vertices {
		result[i] = v.Mul(scaleFactor)
	}

	return result
}

// Extrude extrudes a 2D form into 3D by adding depth
func (fg *FormGenerator) Extrude(vertices []dxfmath.Vec3, depth float64) [][]dxfmath.Vec3 {
	if len(vertices) < 3 {
		return nil
	}

	// Create front and back faces
	frontFace := vertices
	backFace := make([]dxfmath.Vec3, len(vertices))

	for i, v := range vertices {
		backFace[i] = dxfmath.NewVec3(v.X(), v.Y(), depth)
	}

	// Create side faces
	sideFaces := make([][]dxfmath.Vec3, len(vertices))

	for i := 0; i < len(vertices); i++ {
		next := (i + 1) % len(vertices)

		// Side face (quad)
		face := []dxfmath.Vec3{
			vertices[i],    // Front vertex i
			vertices[next], // Front vertex i+1
			backFace[next], // Back vertex i+1
			backFace[i],    // Back vertex i
		}

		sideFaces[i] = face
	}

	// Combine all faces
	allFaces := [][]dxfmath.Vec3{frontFace, backFace}
	allFaces = append(allFaces, sideFaces...)

	return allFaces
}

// Cube generates a cube form
func (fg *FormGenerator) Cube(size float64) [][]dxfmath.Vec3 {
	// Generate 2D square and extrude
	square := fg.Square(size)
	return fg.Extrude(square, size)
}

// Cylinder generates a cylinder form (using approximation)
func (fg *FormGenerator) Cylinder(radius, height float64, segments int) [][]dxfmath.Vec3 {
	if segments <= 0 {
		segments = fg.segments
	}

	// Generate top and bottom circles
	topCircle := fg.Circle(radius)
	bottomCircle := make([]dxfmath.Vec3, len(topCircle))
	for i, v := range topCircle {
		bottomCircle[i] = dxfmath.NewVec3(v.X(), v.Y(), 0)
	}

	// Create side faces
	faces := [][]dxfmath.Vec3{topCircle, bottomCircle}

	for i := 0; i < segments; i++ {
		next := (i + 1) % segments

		// Side face (quad)
		face := []dxfmath.Vec3{
			topCircle[i],       // Top vertex i
			topCircle[next],    // Top vertex i+1
			bottomCircle[next], // Bottom vertex i+1
			bottomCircle[i],    // Bottom vertex i
		}

		faces = append(faces, face)
	}

	return faces
}

// Cone generates a cone form
func (fg *FormGenerator) Cone(radius, height float64, segments int) [][]dxfmath.Vec3 {
	if segments <= 0 {
		segments = fg.segments
	}

	// Generate base circle
	baseCircle := fg.Circle(radius)
	apex := dxfmath.NewVec3(0, 0, height)

	faces := [][]dxfmath.Vec3{baseCircle}

	for i := 0; i < segments; i++ {
		next := (i + 1) % segments

		// Side face (triangle)
		face := []dxfmath.Vec3{
			baseCircle[i],    // Base vertex i
			baseCircle[next], // Base vertex i+1
			apex,             // Apex
		}

		faces = append(faces, face)
	}

	return faces
}

// Sphere generates a sphere form using latitude/longitude approximation
func (fg *FormGenerator) Sphere(radius float64, segments int, rings int) [][]dxfmath.Vec3 {
	if segments <= 0 {
		segments = fg.segments
	}
	if rings <= 0 {
		rings = segments / 2
	}

	var faces [][]dxfmath.Vec3

	// Generate vertices
	vertices := make([][]dxfmath.Vec3, rings+1)
	for ring := 0; ring <= rings; ring++ {
		theta := math.Pi * float64(ring) / float64(rings)
		ringRadius := radius * math.Sin(theta)
		z := radius * math.Cos(theta)

		vertices[ring] = make([]dxfmath.Vec3, segments)
		for seg := 0; seg < segments; seg++ {
			phi := 2.0 * math.Pi * float64(seg) / float64(segments)
			x := ringRadius * math.Cos(phi)
			y := ringRadius * math.Sin(phi)
			vertices[ring][seg] = dxfmath.NewVec3(x, y, z)
		}
	}

	// Generate faces (quads)
	for ring := 0; ring < rings; ring++ {
		for seg := 0; seg < segments; seg++ {
			nextSeg := (seg + 1) % segments

			if ring == 0 {
				// Top triangle
				face := []dxfmath.Vec3{
					vertices[0][0],       // Top pole
					vertices[1][seg],     // Ring vertex
					vertices[1][nextSeg], // Ring vertex+1
				}
				faces = append(faces, face)
			} else if ring == rings-1 {
				// Bottom triangle
				face := []dxfmath.Vec3{
					vertices[rings][0],      // Bottom pole
					vertices[ring][nextSeg], // Ring vertex+1
					vertices[ring][seg],     // Ring vertex
				}
				faces = append(faces, face)
			} else {
				// Middle quad
				face := []dxfmath.Vec3{
					vertices[ring][seg],       // Ring vertex
					vertices[ring][nextSeg],   // Ring vertex+1
					vertices[ring+1][nextSeg], // Next ring vertex+1
					vertices[ring+1][seg],     // Next ring vertex
				}
				faces = append(faces, face)
			}
		}
	}

	return faces
}

// Torus generates a torus form
func (fg *FormGenerator) Torus(majorRadius, minorRadius float64, majorSegments, minorSegments int) [][]dxfmath.Vec3 {
	if majorSegments <= 0 {
		majorSegments = fg.segments
	}
	if minorSegments <= 0 {
		minorSegments = fg.segments / 4
	}

	var faces [][]dxfmath.Vec3

	// Generate vertices
	vertices := make([][]dxfmath.Vec3, majorSegments)
	for major := 0; major < majorSegments; major++ {
		majorAngle := 2.0 * math.Pi * float64(major) / float64(majorSegments)
		vertices[major] = make([]dxfmath.Vec3, minorSegments)

		for minor := 0; minor < minorSegments; minor++ {
			minorAngle := 2.0 * math.Pi * float64(minor) / float64(minorSegments)

			// Torus parametric equations
			x := (majorRadius + minorRadius*math.Cos(minorAngle)) * math.Cos(majorAngle)
			y := (majorRadius + minorRadius*math.Cos(minorAngle)) * math.Sin(majorAngle)
			z := minorRadius * math.Sin(minorAngle)

			vertices[major][minor] = dxfmath.NewVec3(x, y, z)
		}
	}

	// Generate faces
	for major := 0; major < majorSegments; major++ {
		nextMajor := (major + 1) % majorSegments
		for minor := 0; minor < minorSegments; minor++ {
			nextMinor := (minor + 1) % minorSegments

			// Quad face
			face := []dxfmath.Vec3{
				vertices[major][minor],
				vertices[nextMajor][minor],
				vertices[nextMajor][nextMinor],
				vertices[major][nextMinor],
			}
			faces = append(faces, face)
		}
	}

	return faces
}

// ClosePolygon ensures polygon is closed by adding first vertex as last if needed
func (fg *FormGenerator) ClosePolygon(vertices []dxfmath.Vec3) []dxfmath.Vec3 {
	if len(vertices) < 3 {
		return vertices
	}

	// Check if polygon is already closed
	first := vertices[0]
	last := vertices[len(vertices)-1]

	// Consider polygons closed if first and last vertices are very close
	if first.Distance(last) < 1e-9 {
		return vertices
	}

	// Close polygon by adding first vertex at end
	closed := make([]dxfmath.Vec3, len(vertices)+1)
	copy(closed, vertices)
	closed[len(vertices)] = first

	return closed
}

// EulerSpiral generates an Euler spiral (clothoid) curve
func (fg *FormGenerator) EulerSpiral(length float64, radius1, radius2 float64, segments int) []dxfmath.Vec3 {
	if segments <= 0 {
		segments = fg.segments
	}

	vertices := make([]dxfmath.Vec3, segments+1)

	for i := 0; i <= segments; i++ {
		t := float64(i) / float64(segments) * length

		// Simplified Euler spiral approximation
		// Real implementation would involve Fresnel integrals
		angle := t * t / (2.0 * length)
		x := t * math.Cos(angle)
		y := t * math.Sin(angle)

		vertices[i] = dxfmath.NewVec3(x, y, 0)
	}

	return vertices
}

// Gear generates a gear form
func (fg *FormGenerator) Gear(outerRadius, innerRadius, toothHeight float64, teeth int) []dxfmath.Vec3 {
	vertices := make([]dxfmath.Vec3, teeth*4+2)

	angleStep := 2.0 * math.Pi / float64(teeth)
	toothAngle := angleStep * 0.4 // Tooth width as 40% of pitch

	vertexIndex := 0

	// Start with inner radius point
	vertices[vertexIndex] = dxfmath.NewVec3(innerRadius, 0, 0)
	vertexIndex++

	for i := 0; i < teeth; i++ {
		baseAngle := float64(i) * angleStep

		// Tooth base (inner radius)
		vertices[vertexIndex] = dxfmath.NewVec3(
			innerRadius*math.Cos(baseAngle-toothAngle/2),
			innerRadius*math.Sin(baseAngle-toothAngle/2),
			0,
		)
		vertexIndex++

		// Tooth top (outer radius)
		vertices[vertexIndex] = dxfmath.NewVec3(
			outerRadius*math.Cos(baseAngle),
			outerRadius*math.Sin(baseAngle),
			0,
		)
		vertexIndex++

		// Tooth base (inner radius)
		vertices[vertexIndex] = dxfmath.NewVec3(
			innerRadius*math.Cos(baseAngle+toothAngle/2),
			innerRadius*math.Sin(baseAngle+toothAngle/2),
			0,
		)
		vertexIndex++
	}

	// Close the shape by repeating first vertex
	vertices[vertexIndex] = vertices[1]

	return vertices
}

// Sweep extrudes a 2D profile along a path
func (fg *FormGenerator) Sweep(profile []dxfmath.Vec3, path []dxfmath.Vec3) [][]dxfmath.Vec3 {
	if len(profile) < 3 || len(path) < 2 {
		return nil
	}

	var faces [][]dxfmath.Vec3

	// Generate cross-sections at each path point
	crossSections := make([][]dxfmath.Vec3, len(path))
	for i, pathPoint := range path {
		crossSections[i] = make([]dxfmath.Vec3, len(profile))
		for j, profilePoint := range profile {
			crossSections[i][j] = dxfmath.NewVec3(
				profilePoint.X()+pathPoint.X(),
				profilePoint.Y()+pathPoint.Y(),
				profilePoint.Z()+pathPoint.Z(),
			)
		}
	}

	// Add first cross-section as face
	faces = append(faces, crossSections[0])

	// Generate side faces
	for i := 0; i < len(path)-1; i++ {
		next := (i + 1)
		for j := 0; j < len(profile); j++ {
			nextJ := (j + 1) % len(profile)

			// Quad face
			face := []dxfmath.Vec3{
				crossSections[i][j],
				crossSections[next][j],
				crossSections[next][nextJ],
				crossSections[i][nextJ],
			}
			faces = append(faces, face)
		}
	}

	// Add last cross-section as face
	faces = append(faces, crossSections[len(path)-1])

	return faces
}

// TurtleCommand represents a turtle graphics command
type TurtleCommand struct {
	Command string    // "move", "line", "arc"
	Params  []float64 // Command parameters
}

// Turtle implements turtle graphics for form generation
type Turtle struct {
	position dxfmath.Vec3
	angle    float64 // Current heading angle in radians
	penDown  bool
	commands []TurtleCommand
}

// NewTurtle creates a new turtle at origin
func NewTurtle() *Turtle {
	return &Turtle{
		position: dxfmath.NewVec3(0, 0, 0),
		angle:    0,
		penDown:  true,
		commands: make([]TurtleCommand, 0),
	}
}

// Move moves turtle without drawing
func (t *Turtle) Move(distance float64) {
	dx := distance * math.Cos(t.angle)
	dy := distance * math.Sin(t.angle)
	newPos := dxfmath.NewVec3(t.position.X()+dx, t.position.Y()+dy, t.position.Z())

	if t.penDown {
		t.commands = append(t.commands, TurtleCommand{
			Command: "line",
			Params:  []float64{t.position.X(), t.position.Y(), t.position.Z(), newPos.X(), newPos.Y(), newPos.Z()},
		})
	} else {
		t.commands = append(t.commands, TurtleCommand{
			Command: "move",
			Params:  []float64{newPos.X(), newPos.Y(), newPos.Z()},
		})
	}

	t.position = newPos
}

// Turn rotates turtle by angle in radians
func (t *Turtle) Turn(angle float64) {
	t.angle += angle
}

// Right turns turtle right by angle in radians
func (t *Turtle) Right(angle float64) {
	t.Turn(-angle)
}

// Left turns turtle left by angle in radians
func (t *Turtle) Left(angle float64) {
	t.Turn(angle)
}

// PenUp lifts the pen (no drawing)
func (t *Turtle) PenUp() {
	t.penDown = false
}

// PenDown puts the pen down (start drawing)
func (t *Turtle) PenDown() {
	t.penDown = true
}

// GoTo moves turtle to specific position
func (t *Turtle) GoTo(x, y, z float64) {
	newPos := dxfmath.NewVec3(x, y, z)

	if t.penDown {
		t.commands = append(t.commands, TurtleCommand{
			Command: "line",
			Params:  []float64{t.position.X(), t.position.Y(), t.position.Z(), x, y, z},
		})
	} else {
		t.commands = append(t.commands, TurtleCommand{
			Command: "move",
			Params:  []float64{x, y, z},
		})
	}

	t.position = newPos
}

// GetPath extracts the path from turtle commands
func (t *Turtle) GetPath() []dxfmath.Vec3 {
	var path []dxfmath.Vec3
	currentPos := dxfmath.NewVec3(0, 0, 0)

	for _, cmd := range t.commands {
		switch cmd.Command {
		case "move":
			if len(cmd.Params) >= 3 {
				currentPos = dxfmath.NewVec3(cmd.Params[0], cmd.Params[1], cmd.Params[2])
			}
		case "line":
			if len(cmd.Params) >= 3 {
				currentPos = dxfmath.NewVec3(cmd.Params[0], cmd.Params[1], cmd.Params[2])
				path = append(path, currentPos)
			}
		}
	}

	return path
}

// Turtle generates a form using turtle graphics commands
func (fg *FormGenerator) Turtle(commands func(*Turtle)) []dxfmath.Vec3 {
	turtle := NewTurtle()
	commands(turtle)
	return turtle.GetPath()
}

// RotationForm creates a form by rotating a profile around an axis
func (fg *FormGenerator) RotationForm(profile []dxfmath.Vec3, angle, segments int, axis string) [][]dxfmath.Vec3 {
	if segments <= 0 {
		segments = fg.segments
	}

	angleRad := float64(angle) * math.Pi / 180.0
	angleStep := angleRad / float64(segments)

	var faces [][]dxfmath.Vec3

	// Generate rotated profiles
	profiles := make([][]dxfmath.Vec3, segments+1)
	profiles[0] = profile // Starting profile

	for i := 1; i <= segments; i++ {
		currentAngle := float64(i) * angleStep
		profiles[i] = make([]dxfmath.Vec3, len(profile))

		for j, vertex := range profile {
			var rotated dxfmath.Vec3

			switch axis {
			case "x", "X":
				// Rotate around X axis
				y := vertex.Y()*math.Cos(currentAngle) - vertex.Z()*math.Sin(currentAngle)
				z := vertex.Y()*math.Sin(currentAngle) + vertex.Z()*math.Cos(currentAngle)
				rotated = dxfmath.NewVec3(vertex.X(), y, z)
			case "y", "Y":
				// Rotate around Y axis
				x := vertex.X()*math.Cos(currentAngle) + vertex.Z()*math.Sin(currentAngle)
				z := -vertex.X()*math.Sin(currentAngle) + vertex.Z()*math.Cos(currentAngle)
				rotated = dxfmath.NewVec3(x, vertex.Y(), z)
			default: // Z axis
				// Rotate around Z axis
				x := vertex.X()*math.Cos(currentAngle) - vertex.Y()*math.Sin(currentAngle)
				y := vertex.X()*math.Sin(currentAngle) + vertex.Y()*math.Cos(currentAngle)
				rotated = dxfmath.NewVec3(x, y, vertex.Z())
			}

			profiles[i][j] = rotated
		}
	}

	// Generate faces between consecutive profiles
	for i := 0; i < segments; i++ {
		for j := 0; j < len(profile); j++ {
			nextJ := (j + 1) % len(profile)

			// Quad face
			face := []dxfmath.Vec3{
				profiles[i][j],
				profiles[i][nextJ],
				profiles[i+1][nextJ],
				profiles[i+1][j],
			}
			faces = append(faces, face)
		}
	}

	return faces
}

// ExtrudeTwistScale extrudes a 2D profile with twisting and scaling
func (fg *FormGenerator) ExtrudeTwistScale(profile []dxfmath.Vec3, height float64, twist float64, scale float64, segments int) [][]dxfmath.Vec3 {
	if segments <= 0 {
		segments = fg.segments
	}

	var faces [][]dxfmath.Vec3

	// Generate cross-sections at different heights
	crossSections := make([][]dxfmath.Vec3, segments+1)
	crossSections[0] = profile // Base profile

	for i := 1; i <= segments; i++ {
		heightRatio := float64(i) / float64(segments)
		currentHeight := height * heightRatio
		currentTwist := twist * heightRatio * math.Pi / 180.0 // Convert to radians
		currentScale := 1.0 + (scale-1.0)*heightRatio         // Linear scale interpolation

		crossSections[i] = make([]dxfmath.Vec3, len(profile))

		for j, vertex := range profile {
			// Apply scaling
			x := vertex.X() * currentScale
			y := vertex.Y() * currentScale
			z := vertex.Z() + currentHeight

			// Apply rotation (twist) around Z axis
			rotatedX := x*math.Cos(currentTwist) - y*math.Sin(currentTwist)
			rotatedY := x*math.Sin(currentTwist) + y*math.Cos(currentTwist)

			crossSections[i][j] = dxfmath.NewVec3(rotatedX, rotatedY, z)
		}
	}

	// Generate side faces
	for i := 0; i < segments; i++ {
		for j := 0; j < len(profile); j++ {
			nextJ := (j + 1) % len(profile)

			// Quad face
			face := []dxfmath.Vec3{
				crossSections[i][j],
				crossSections[i][nextJ],
				crossSections[i+1][nextJ],
				crossSections[i+1][j],
			}
			faces = append(faces, face)
		}
	}

	return faces
}

// FromProfilesLinear connects multiple profiles with linear interpolation
func (fg *FormGenerator) FromProfilesLinear(profiles [][]dxfmath.Vec3, close bool) [][]dxfmath.Vec3 {
	if len(profiles) < 2 {
		return nil
	}

	var faces [][]dxfmath.Vec3

	// Generate faces between consecutive profiles
	for i := 0; i < len(profiles)-1; i++ {
		profile1 := profiles[i]
		profile2 := profiles[i+1]

		if len(profile1) != len(profile2) {
			continue // Skip if profiles have different vertex counts
		}

		for j := 0; j < len(profile1); j++ {
			nextJ := (j + 1) % len(profile1)

			// Quad face
			face := []dxfmath.Vec3{
				profile1[j],
				profile1[nextJ],
				profile2[nextJ],
				profile2[j],
			}
			faces = append(faces, face)
		}
	}

	// Close the shape if requested
	if close && len(profiles) > 2 {
		first := profiles[0]
		last := profiles[len(profiles)-1]

		for j := 0; j < len(first); j++ {
			nextJ := (j + 1) % len(first)

			// Quad face to close the shape
			face := []dxfmath.Vec3{
				last[j],
				last[nextJ],
				first[nextJ],
				first[j],
			}
			faces = append(faces, face)
		}
	}

	return faces
}
