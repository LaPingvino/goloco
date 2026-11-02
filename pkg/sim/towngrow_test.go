package sim

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// squareSearchRange9 = upstream's kSquareSearchRange<9> compiled from
// Town.cpp (tools/gen-paintdata/gen_square_search.cpp).
func TestSquareSearchRange(t *testing.T) {
	data, err := os.ReadFile("testdata/square_search9.txt")
	if err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	for _, p := range squareSearchRange9 {
		fmt.Fprintf(&b, "%d,%d ", p[0], p[1])
	}
	if got, want := strings.TrimSpace(b.String()), strings.TrimSpace(string(data)); got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
}
