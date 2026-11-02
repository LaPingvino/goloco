package sim

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
