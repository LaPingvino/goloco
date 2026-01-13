package olmath

import "math"

// Vector2i represents a 2D vector with int32 components.
type Vector2i struct {
	X int32
	Y int32
}

// Vector3i represents a 3D vector with int32 components.
type Vector3i struct {
	X int32
	Y int32
	Z int32
}

func NewVector2i(x, y int32) Vector2i    { return Vector2i{X: x, Y: y} }
func NewVector3i(x, y, z int32) Vector3i { return Vector3i{X: x, Y: y, Z: z} }

func (v Vector2i) Equals(o Vector2i) bool { return v.X == o.X && v.Y == o.Y }

func (v Vector2i) Add(o Vector2i) Vector2i    { return Vector2i{X: v.X + o.X, Y: v.Y + o.Y} }
func (v Vector2i) Sub(o Vector2i) Vector2i    { return Vector2i{X: v.X - o.X, Y: v.Y - o.Y} }
func (v Vector2i) MulScalar(s int32) Vector2i { return Vector2i{X: v.X * s, Y: v.Y * s} }
func (v Vector2i) DivScalar(s int32) Vector2i { return Vector2i{X: v.X / s, Y: v.Y / s} }
func (v Vector2i) ShiftLeft(s uint) Vector2i  { return Vector2i{X: v.X << s, Y: v.Y << s} }
func (v Vector2i) ShiftRight(s uint) Vector2i { return Vector2i{X: v.X >> s, Y: v.Y >> s} }

// Rotate rotates the vector by direction (0..3):
// 0 => (x, y)
// 1 => (y, -x)
// 2 => (-x, -y)
// 3 => (-y, x)
func (v Vector2i) Rotate(direction int32) Vector2i {
	switch direction & 3 {
	case 0:
		return v
	case 1:
		return Vector2i{X: v.Y, Y: -v.X}
	case 2:
		return Vector2i{X: -v.X, Y: -v.Y}
	case 3:
		return Vector2i{X: -v.Y, Y: v.X}
	default:
		return v
	}
}

func abs32(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
}

func max32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

// ManhattanDistance2D returns taxicab distance between two 2D vectors.
func ManhattanDistance2D(a, b Vector2i) int32 {
	return abs32(a.X-b.X) + abs32(a.Y-b.Y)
}

// ManhattanDistance3D returns taxicab distance in 3D.
func ManhattanDistance3D(a, b Vector3i) int32 {
	x := abs32(a.X - b.X)
	y := abs32(a.Y - b.Y)
	z := abs32(a.Z - b.Z)
	return x + y + z
}

// ChebyshevDistance2D returns the chessboard distance between two 2D vectors.
func ChebyshevDistance2D(a, b Vector2i) int32 {
	dx := abs32(a.X - b.X)
	dy := abs32(a.Y - b.Y)
	return max32(dx, dy)
}

// ChebyshevDistance3D returns the maximum component distance in 3D.
func ChebyshevDistance3D(a, b Vector3i) int32 {
	dx := abs32(a.X - b.X)
	dy := abs32(a.Y - b.Y)
	dz := abs32(a.Z - b.Z)
	m := max32(dx, dy)
	return max32(m, dz)
}

// Dot for 2D returns dot product as int64 to reduce overflow risk.
func Dot2D(a, b Vector2i) int64 {
	return int64(a.X)*int64(b.X) + int64(a.Y)*int64(b.Y)
}

// Dot for 3D
func Dot3D(a, b Vector3i) int64 {
	return int64(a.X)*int64(b.X) + int64(a.Y)*int64(b.Y) + int64(a.Z)*int64(b.Z)
}

// Cross for 3D returns the vector cross product.
func Cross3D(a, b Vector3i) Vector3i {
	return Vector3i{
		X: a.Y*b.Z - a.Z*b.Y,
		Y: a.Z*b.X - a.X*b.Z,
		Z: a.X*b.Y - a.Y*b.X,
	}
}

// fastSquareRoot approximates square root and returns uint16 (mirrors OpenLoco behaviour).
func FastSquareRoot(distance uint32) uint16 {
	if distance == 0 {
		return 0
	}
	v := math.Sqrt(float64(distance))
	if v > float64(^uint16(0)) {
		return ^uint16(0)
	}
	return uint16(v)
}

// Distance2D returns the integer-approximated Euclidean distance between two 2D vectors.
func Distance2D(a, b Vector2i) uint16 {
	dx := int64(a.X - b.X)
	dy := int64(a.Y - b.Y)
	d := uint64(dx*dx + dy*dy)
	if d > math.MaxUint32 {
		// fallback to float calculation
		return uint16(math.Sqrt(float64(d)))
	}
	return FastSquareRoot(uint32(d))
}

// Distance squared (helper)
func Distance2DSquared(a, b Vector2i) int64 {
	dx := int64(a.X - b.X)
	dy := int64(a.Y - b.Y)
	return dx*dx + dy*dy
}
