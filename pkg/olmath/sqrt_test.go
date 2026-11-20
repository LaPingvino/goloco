package olmath

import (
	"bufio"
	"fmt"
	"os"
	"testing"
)

// FastSquareRoot against 2000 values from upstream's own function
// (tools/gen-paintdata/gen_sqrt.cpp).
func TestFastSquareRootVectors(t *testing.T) {
	f, err := os.Open("testdata/sqrt.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	n := 0
	for sc.Scan() {
		var v uint32
		var want uint16
		fmt.Sscan(sc.Text(), &v, &want)
		if got := FastSquareRoot(v); got != want {
			t.Fatalf("fastSquareRoot(%d) = %d, want %d", v, got, want)
		}
		n++
	}
	if n != 2000 {
		t.Fatalf("%d vectors", n)
	}
}

// Distance2D goes through the table as upstream's distance2D does, not an
// exact root: Boulder Breakers' Train 1 between its two platforms (Vehicle2 at
// 6576,5823 and 6080,5392; 431777 squared) is 656 units in OpenLoco — the
// exact root's 657 made its journey's average speed 27 instead of OpenLoco's 26.
func TestDistance2DIsFastSquareRoot(t *testing.T) {
	a, b := NewVector2i(6576, 5823), NewVector2i(6080, 5392)
	if got := Distance2D(a, b); got != 656 {
		t.Fatalf("Distance2D = %d, want 656 (fastSquareRoot(431777))", got)
	}
	if got := Distance2D(NewVector2i(0, 0), NewVector2i(3, 4)); got != FastSquareRoot(25) {
		t.Fatalf("Distance2D(3,4) = %d, want fastSquareRoot(25) = %d", got, FastSquareRoot(25))
	}
}
