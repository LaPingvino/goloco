// Package math provides vector and math utilities for goloco.
// Translated from OpenLoco/src/Math
package math

import (
	gomath "math"
)

// Vector2 represents a 2D vector with int32 components
type Vector2 struct {
	X, Y int32
}

// NewVector2 creates a new Vector2
func NewVector2(x, y int32) Vector2 {
	return Vector2{X: x, Y: y}
}

// Add returns the sum of two vectors
func (v Vector2) Add(other Vector2) Vector2 {
	return Vector2{X: v.X + other.X, Y: v.Y + other.Y}
}

// Sub returns the difference of two vectors
func (v Vector2) Sub(other Vector2) Vector2 {
	return Vector2{X: v.X - other.X, Y: v.Y - other.Y}
}

// Mul returns the vector scaled by a factor
func (v Vector2) Mul(factor int32) Vector2 {
	return Vector2{X: v.X * factor, Y: v.Y * factor}
}

// Div returns the vector divided by a factor
func (v Vector2) Div(factor int32) Vector2 {
	return Vector2{X: v.X / factor, Y: v.Y / factor}
}

// Shl returns the vector with components left-shifted
func (v Vector2) Shl(bits uint8) Vector2 {
	return Vector2{X: v.X << bits, Y: v.Y << bits}
}

// Shr returns the vector with components right-shifted
func (v Vector2) Shr(bits uint8) Vector2 {
	return Vector2{X: v.X >> bits, Y: v.Y >> bits}
}

// Equals checks if two vectors are equal
func (v Vector2) Equals(other Vector2) bool {
	return v.X == other.X && v.Y == other.Y
}

// Vector3 represents a 3D vector with int32 components
type Vector3 struct {
	X, Y, Z int32
}

// NewVector3 creates a new Vector3
func NewVector3(x, y, z int32) Vector3 {
	return Vector3{X: x, Y: y, Z: z}
}

// Add returns the sum of two vectors
func (v Vector3) Add(other Vector3) Vector3 {
	return Vector3{X: v.X + other.X, Y: v.Y + other.Y, Z: v.Z + other.Z}
}

// Sub returns the difference of two vectors
func (v Vector3) Sub(other Vector3) Vector3 {
	return Vector3{X: v.X - other.X, Y: v.Y - other.Y, Z: v.Z - other.Z}
}

// Mul returns the vector scaled by a factor
func (v Vector3) Mul(factor int32) Vector3 {
	return Vector3{X: v.X * factor, Y: v.Y * factor, Z: v.Z * factor}
}

// Div returns the vector divided by a factor
func (v Vector3) Div(factor int32) Vector3 {
	return Vector3{X: v.X / factor, Y: v.Y / factor, Z: v.Z / factor}
}

// Equals checks if two vectors are equal
func (v Vector3) Equals(other Vector3) bool {
	return v.X == other.X && v.Y == other.Y && v.Z == other.Z
}

// XY returns a Vector2 from the X and Y components
func (v Vector3) XY() Vector2 {
	return Vector2{X: v.X, Y: v.Y}
}

// Rotate rotates a Vector2 by 90-degree increments (0-3)
func Rotate(v Vector2, direction int32) Vector2 {
	switch direction & 3 {
	case 1:
		return Vector2{X: v.Y, Y: -v.X}
	case 2:
		return Vector2{X: -v.X, Y: -v.Y}
	case 3:
		return Vector2{X: -v.Y, Y: v.X}
	default:
		return v
	}
}

// abs returns the absolute value of an int32
func abs(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
}

// max returns the larger of two int32 values
func max(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

// max3 returns the largest of three int32 values
func max3(a, b, c int32) int32 {
	return max(max(a, b), c)
}

// ManhattanDistance2D returns the Manhattan (taxicab) distance between two 2D points
func ManhattanDistance2D(a, b Vector2) int32 {
	return abs(a.X-b.X) + abs(a.Y-b.Y)
}

// ManhattanDistance3D returns the Manhattan (taxicab) distance between two 3D points
func ManhattanDistance3D(a, b Vector3) int32 {
	return abs(a.X-b.X) + abs(a.Y-b.Y) + abs(a.Z-b.Z)
}

// ChebyshevDistance2D returns the Chebyshev (chessboard) distance between two 2D points
func ChebyshevDistance2D(a, b Vector2) int32 {
	return max(abs(a.X-b.X), abs(a.Y-b.Y))
}

// ChebyshevDistance3D returns the Chebyshev (chessboard) distance between two 3D points
func ChebyshevDistance3D(a, b Vector3) int32 {
	return max3(abs(a.X-b.X), abs(a.Y-b.Y), abs(a.Z-b.Z))
}

// Dot2D returns the dot product of two 2D vectors
func Dot2D(a, b Vector2) int32 {
	return a.X*b.X + a.Y*b.Y
}

// Dot3D returns the dot product of two 3D vectors
func Dot3D(a, b Vector3) int32 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z
}

// Cross returns the cross product of two 3D vectors
func Cross(a, b Vector3) Vector3 {
	return Vector3{
		X: a.Y*b.Z - a.Z*b.Y,
		Y: a.Z*b.X - a.X*b.Z,
		Z: a.X*b.Y - a.Y*b.X,
	}
}

// Sqrt returns the integer square root using Go's standard library
// Replaces the original lookup table approach for simplicity
func Sqrt(distance uint32) uint32 {
	return uint32(gomath.Sqrt(float64(distance)))
}

// Distance2D returns the Euclidean distance between two 2D points
func Distance2D(a, b Vector2) uint32 {
	dx := int64(a.X - b.X)
	dy := int64(a.Y - b.Y)
	return Sqrt(uint32(dx*dx + dy*dy))
}

// Distance3D returns the Euclidean distance between two 3D points
func Distance3D(a, b Vector3) uint32 {
	dx := int64(a.X - b.X)
	dy := int64(a.Y - b.Y)
	dz := int64(a.Z - b.Z)
	return Sqrt(uint32(dx*dx + dy*dy + dz*dz))
}
