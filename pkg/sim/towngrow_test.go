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

// rotateQuarterTile and rotl4bit over every value and rotation = upstream's
// QuarterTile::rotate and Numerics::rotl4bit (tools/gen-paintdata/gen_quartertile.cpp).
func TestQuarterTileRotateMatchesUpstream(t *testing.T) {
	data, err := os.ReadFile("testdata/quartertile_rotate.txt")
	if err != nil {
		t.Fatal(err)
	}
	h := uint64(1469598103934665603)
	for v := range 256 {
		for r := range 4 {
			h ^= uint64(rotateQuarterTile(uint8(v), uint8(r)))
			h *= 1099511628211
			h ^= uint64(rotl4bit(uint8(v)&0xF, uint8(r)))
			h *= 1099511628211
		}
	}
	if got := fmt.Sprintf("%x", h); got != strings.TrimSpace(string(data)) {
		t.Errorf("hash %s, upstream %s", got, strings.TrimSpace(string(data)))
	}
}

// The generated road preview pieces agree with RoadPieceTiles' positions.
func TestRoadPiecesAgree(t *testing.T) {
	for id, ps := range RoadPieces {
		if len(ps) != len(RoadPieceTiles[id]) {
			t.Fatalf("road %d: %d pieces vs %d", id, len(ps), len(RoadPieceTiles[id]))
		}
		for i, p := range ps {
			if o := RoadPieceTiles[id][i]; p.X != o[0] || p.Y != o[1] || p.Z != o[2] || int(p.Index) != i {
				t.Errorf("road %d piece %d: %+v vs %v", id, i, p, o)
			}
		}
	}
}
