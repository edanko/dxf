package math

import (
	"fmt"
	"math"
)

// Matrix44 represents a 4x4 transformation matrix
type Matrix44 [16]float64

// NewMatrix44 creates a new identity matrix
func NewMatrix44() Matrix44 {
	return Matrix44{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

// NewMatrix44FromSlice creates a matrix from a slice (row-major order)
func NewMatrix44FromSlice(slice []float64) Matrix44 {
	m := NewMatrix44()
	if len(slice) >= 16 {
		copy(m[:], slice[:16])
	}
	return m
}

// NewMatrix44FromColumns creates a matrix from column vectors
func NewMatrix44FromColumns(c1, c2, c3, c4 Vec4) Matrix44 {
	return Matrix44{
		c1[0], c2[0], c3[0], c4[0],
		c1[1], c2[1], c3[1], c4[1],
		c1[2], c2[2], c3[2], c4[2],
		c1[3], c2[3], c3[3], c4[3],
	}
}

// Identity returns an identity matrix
func Identity() Matrix44 {
	return NewMatrix44()
}

// Get returns element at row i, column j
func (m Matrix44) Get(i, j int) float64 {
	if i < 0 || i >= 4 || j < 0 || j >= 4 {
		return 0
	}
	return m[i*4+j]
}

// Set sets element at row i, column j
func (m *Matrix44) Set(i, j int, value float64) {
	if i >= 0 && i < 4 && j >= 0 && j < 4 {
		m[i*4+j] = value
	}
}

// GetRow returns a row as Vec4
func (m Matrix44) GetRow(i int) Vec4 {
	if i < 0 || i >= 4 {
		return Vec4{0, 0, 0, 0}
	}
	base := i * 4
	return Vec4{m[base], m[base+1], m[base+2], m[base+3]}
}

// GetColumn returns a column as Vec4
func (m Matrix44) GetColumn(j int) Vec4 {
	if j < 0 || j >= 4 {
		return Vec4{0, 0, 0, 0}
	}
	return Vec4{m[j], m[j+4], m[j+8], m[j+12]}
}

// SetRow sets a row
func (m *Matrix44) SetRow(i int, row Vec4) {
	if i >= 0 && i < 4 {
		base := i * 4
		m[base] = row[0]
		m[base+1] = row[1]
		m[base+2] = row[2]
		m[base+3] = row[3]
	}
}

// SetColumn sets a column
func (m *Matrix44) SetColumn(j int, col Vec4) {
	if j >= 0 && j < 4 {
		m[j] = col[0]
		m[j+4] = col[1]
		m[j+8] = col[2]
		m[j+12] = col[3]
	}
}

// Add adds another matrix
func (m Matrix44) Add(other Matrix44) Matrix44 {
	result := NewMatrix44()
	for i := 0; i < 16; i++ {
		result[i] = m[i] + other[i]
	}
	return result
}

// Sub subtracts another matrix
func (m Matrix44) Sub(other Matrix44) Matrix44 {
	result := NewMatrix44()
	for i := 0; i < 16; i++ {
		result[i] = m[i] - other[i]
	}
	return result
}

// Mul multiplies by another matrix
func (m Matrix44) Mul(other Matrix44) Matrix44 {
	result := NewMatrix44()
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			sum := 0.0
			for k := 0; k < 4; k++ {
				sum += m.Get(i, k) * other.Get(k, j)
			}
			result.Set(i, j, sum)
		}
	}
	return result
}

// MulVec multiplies a 3D vector (treating as Vec4 with w=1)
func (m Matrix44) MulVec(v Vec3) Vec3 {
	v4 := NewVec4FromVec3(v)
	result := m.MulVec4(v4)
	return result.XYZ()
}

// MulVec4 multiplies a 4D vector
func (m Matrix44) MulVec4(v Vec4) Vec4 {
	result := Vec4{0, 0, 0, 0}
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			result[i] += m.Get(i, j) * v[j]
		}
	}
	return result
}

// MulScalar multiplies by a scalar
func (m Matrix44) MulScalar(scalar float64) Matrix44 {
	result := NewMatrix44()
	for i := 0; i < 16; i++ {
		result[i] = m[i] * scalar
	}
	return result
}

// Transpose returns the transpose
func (m Matrix44) Transpose() Matrix44 {
	result := NewMatrix44()
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			result.Set(i, j, m.Get(j, i))
		}
	}
	return result
}

// Determinant calculates the determinant
func (m Matrix44) Determinant() float64 {
	// Using Laplace expansion for 4x4 determinant
	return m.determinant4x4()
}

func (m Matrix44) determinant4x4() float64 {
	// Compute 3x3 determinants recursively
	var result float64
	for i := 0; i < 4; i++ {
		sign := 1.0
		if i%2 == 1 {
			sign = -1.0
		}
		cofactor := m.cofactor(0, i)
		result += sign * m.Get(0, i) * cofactor
	}
	return result
}

func (m Matrix44) cofactor(row, col int) float64 {
	minor := m.minor(row, col)
	return m.determinant3x3(minor)
}

func (m Matrix44) minor(row, col int) [9]float64 {
	var minor [9]float64
	k := 0
	for i := 0; i < 4; i++ {
		if i == row {
			continue
		}
		for j := 0; j < 4; j++ {
			if j == col {
				continue
			}
			minor[k] = m.Get(i, j)
			k++
		}
	}
	return minor
}

func (m Matrix44) determinant3x3(minor [9]float64) float64 {
	a, b, c := minor[0], minor[1], minor[2]
	d, e, f := minor[3], minor[4], minor[5]
	g, h, i := minor[6], minor[7], minor[8]

	return a*(e*i-f*h) - b*(d*i-f*g) + c*(d*h-e*g)
}

// Inverse returns the inverse matrix
func (m Matrix44) Inverse() (Matrix44, bool) {
	det := m.Determinant()
	if math.Abs(det) < 1e-10 {
		return NewMatrix44(), false // Singular matrix
	}

	adj := m.adjugate()
	return adj.MulScalar(1.0 / det), true
}

func (m Matrix44) adjugate() Matrix44 {
	result := NewMatrix44()
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			// Transpose position for adjugate
			sign := 1.0
			if (i+j)%2 == 1 {
				sign = -1.0
			}
			cofactor := m.cofactor(j, i)
			result.Set(i, j, sign*cofactor)
		}
	}
	return result
}

// IsIdentity checks if matrix is approximately identity
func (m Matrix44) IsIdentity(epsilon float64) bool {
	identity := Identity()
	return m.IsEqual(identity, epsilon)
}

// TransformVector transforms a 3D vector (ignoring translation)
func (m Matrix44) TransformVector(vector Vec3) Vec3 {
	// Convert vector to homogeneous coordinates (w = 0)
	vec4 := Vec4{vector.X(), vector.Y(), vector.Z(), 0.0}

	// Transform and convert back to 3D (ignoring w component)
	result := m.MulVec4(vec4)
	return Vec3{result[0], result[1], result[2]}
}

// IsEqual checks if two matrices are approximately equal
func (m Matrix44) IsEqual(other Matrix44, epsilon float64) bool {
	for i := 0; i < 16; i++ {
		if math.Abs(m[i]-other[i]) > epsilon {
			return false
		}
	}
	return true
}

// ToSlice converts matrix to slice (row-major order)
func (m Matrix44) ToSlice() []float64 {
	result := make([]float64, 16)
	copy(result, m[:])
	return result
}

// String returns string representation
func (m Matrix44) String() string {
	return fmt.Sprintf("Matrix44([\n  %.6f, %.6f, %.6f, %.6f\n  %.6f, %.6f, %.6f, %.6f\n  %.6f, %.6f, %.6f, %.6f\n  %.6f, %.6f, %.6f, %.6f\n])",
		m[0], m[1], m[2], m[3],
		m[4], m[5], m[6], m[7],
		m[8], m[9], m[10], m[11],
		m[12], m[13], m[14], m[15])
}

// Transformation matrix constructors

// Translation creates a translation matrix
func Translation(x, y, z float64) Matrix44 {
	m := NewMatrix44()
	m[0*4+3] = x
	m[1*4+3] = y
	m[2*4+3] = z
	return m
}

// Scale creates a scaling matrix
func Scale(x, y, z float64) Matrix44 {
	m := NewMatrix44()
	m[0*4+0] = x
	m[1*4+1] = y
	m[2*4+2] = z
	return m
}

// ScaleUniform creates a uniform scaling matrix
func ScaleUniform(s float64) Matrix44 {
	return Scale(s, s, s)
}

// RotationX creates a rotation matrix around X axis (radians)
func RotationX(angle float64) Matrix44 {
	cos, sin := math.Cos(angle), math.Sin(angle)
	m := NewMatrix44()
	m[1*4+1] = cos
	m[1*4+2] = -sin
	m[2*4+1] = sin
	m[2*4+2] = cos
	return m
}

// RotationY creates a rotation matrix around Y axis (radians)
func RotationY(angle float64) Matrix44 {
	cos, sin := math.Cos(angle), math.Sin(angle)
	m := NewMatrix44()
	m[0*4+0] = cos
	m[0*4+2] = sin
	m[2*4+0] = -sin
	m[2*4+2] = cos
	return m
}

// RotationZ creates a rotation matrix around Z axis (radians)
func RotationZ(angle float64) Matrix44 {
	cos, sin := math.Cos(angle), math.Sin(angle)
	m := NewMatrix44()
	m[0*4+0] = cos
	m[0*4+1] = -sin
	m[1*4+0] = sin
	m[1*4+1] = cos
	return m
}

// RotationAxis creates a rotation matrix around an arbitrary axis (radians)
func RotationAxis(axis Vec3, angle float64) Matrix44 {
	axis = axis.Normalized()
	cos, sin := math.Cos(angle), math.Sin(angle)
	oneMinusCos := 1.0 - cos
	x, y, z := axis[0], axis[1], axis[2]

	m := NewMatrix44()
	m[0*4+0] = cos + x*x*oneMinusCos
	m[0*4+1] = x*y*oneMinusCos - z*sin
	m[0*4+2] = x*z*oneMinusCos + y*sin

	m[1*4+0] = y*x*oneMinusCos + z*sin
	m[1*4+1] = cos + y*y*oneMinusCos
	m[1*4+2] = y*z*oneMinusCos - x*sin

	m[2*4+0] = z*x*oneMinusCos - y*sin
	m[2*4+1] = z*y*oneMinusCos + x*sin
	m[2*4+2] = cos + z*z*oneMinusCos

	return m
}

// LookAt creates a view matrix looking from eye to center with up vector
func LookAt(eye, center, up Vec3) Matrix44 {
	f := center.Sub(eye).Normalized()
	s := f.Cross(up).Normalized()
	u := s.Cross(f)

	m := NewMatrix44()
	m[0*4+0] = s[0]
	m[1*4+0] = s[1]
	m[2*4+0] = s[2]
	m[0*4+1] = u[0]
	m[1*4+1] = u[1]
	m[2*4+1] = u[2]
	m[0*4+2] = -f[0]
	m[1*4+2] = -f[1]
	m[2*4+2] = -f[2]

	m[0*4+3] = -s.Dot(eye)
	m[1*4+3] = -u.Dot(eye)
	m[2*4+3] = f.Dot(eye)

	return m
}

// Perspective creates a perspective projection matrix
func Perspective(fov, aspect, near, far float64) Matrix44 {
	f := 1.0 / math.Tan(fov/2.0)

	m := NewMatrix44()
	m[0*4+0] = f / aspect
	m[1*4+1] = f
	m[2*4+2] = (far + near) / (near - far)
	m[2*4+3] = (2.0 * far * near) / (near - far)
	m[3*4+2] = -1.0
	m[3*4+3] = 0.0

	return m
}

// Orthographic creates an orthographic projection matrix
func Orthographic(left, right, bottom, top, near, far float64) Matrix44 {
	m := NewMatrix44()
	m[0*4+0] = 2.0 / (right - left)
	m[1*4+1] = 2.0 / (top - bottom)
	m[2*4+2] = -2.0 / (far - near)
	m[0*4+3] = -(right + left) / (right - left)
	m[1*4+3] = -(top + bottom) / (top - bottom)
	m[2*4+3] = -(far + near) / (far - near)

	return m
}
