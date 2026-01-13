package olmath

import "testing"

func TestVectorAddSub(t *testing.T) {
	a := NewVector2i(3, 4)
	b := NewVector2i(-1, 2)
	c := a.Add(b)
	if c.X != 2 || c.Y != 6 {
		t.Fatalf("unexpected add result: %+v", c)
	}
	d := c.Sub(b)
	if d.X != a.X || d.Y != a.Y {
		t.Fatalf("unexpected sub result: %+v vs %+v", d, a)
	}
}

func TestDistances(t *testing.T) {
	a := NewVector2i(0, 0)
	b := NewVector2i(3, 4)
	if ManhattanDistance2D(a, b) != 7 {
		t.Fatalf("unexpected manhattan")
	}
	if ChebyshevDistance2D(a, b) != 4 {
		t.Fatalf("unexpected chebyshev")
	}
	if Distance2D(a, b) == 0 {
		t.Fatalf("unexpected euclidean distance 0")
	}
}

func TestRotate(t *testing.T) {
	a := NewVector2i(1, 2)
	if a.Rotate(0) != a {
		t.Fatalf("rot0")
	}
	if a.Rotate(1) != NewVector2i(2, -1) {
		t.Fatalf("rot1")
	}
	if a.Rotate(2) != NewVector2i(-1, -2) {
		t.Fatalf("rot2")
	}
	if a.Rotate(3) != NewVector2i(-2, 1) {
		t.Fatalf("rot3")
	}
}

func TestDotCross(t *testing.T) {
	a := NewVector3i(1, 2, 3)
	b := NewVector3i(4, -5, 6)
	if Dot3D(a, b) != int64(1*4+2*-5+3*6) {
		t.Fatalf("dot3d")
	}
	c := Cross3D(a, b)
	if c != NewVector3i(2*6-3*-5, 3*4-1*6, 1*-5-2*4) {
		t.Fatalf("cross3d")
	}
}
