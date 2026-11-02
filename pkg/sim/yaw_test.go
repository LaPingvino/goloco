package sim

import (
	"bufio"
	"fmt"
	"os"
	"testing"
)

// The yaw functions against upstream's own, compiled from VehicleBody.cpp
// and run over a grid of vectors (testdata/yaw.txt: x y yaw0 yaw1Plane yaw1
// yaw2 yaw3 yaw4).
func TestYawFunctionsMatchUpstream(t *testing.T) {
	f, err := os.Open("testdata/yaw.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	n := 0
	for sc.Scan() {
		var x, y int16
		var w [6]uint8
		if _, err := fmt.Sscan(sc.Text(), &x, &y, &w[0], &w[1], &w[2], &w[3], &w[4], &w[5]); err != nil {
			t.Fatal(err)
		}
		got := [6]uint8{Yaw0(x, y), Yaw1Plane(x, y), Yaw1(x, y), Yaw2(x, y), Yaw3(x, y), Yaw4(x, y)}
		if got != w {
			t.Errorf("(%d, %d): got %v, upstream %v", x, y, got, w)
		}
		n++
	}
	if n != 961 {
		t.Errorf("%d vectors", n)
	}
}
