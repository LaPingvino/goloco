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

// calcDays against 2064 dates computed by upstream's own calcDays
// (tools/gen-paintdata/gen_calcdays.cpp), 1800–2100.
func TestCalcDays(t *testing.T) {
	f, err := os.Open("testdata/calcdays.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	n := 0
	for sc.Scan() {
		var y, mo, d int
		var want uint32
		fmt.Sscan(sc.Text(), &y, &mo, &d, &want)
		if got := CalcDays(y, mo, d); got != want {
			t.Fatalf("calcDays(%d, %d, %d) = %d, want %d", y, mo, d, got, want)
		}
		n++
	}
	if n != 2064 || CalcDays(1925, 0, 1) != 45657 {
		t.Fatal("vectors")
	}
}
