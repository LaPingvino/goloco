// tilediff compares two saves' tile elements (as each save holds them,
// element by element in the S5 encoding): how many tiles differ, and the
// differing tiles near a spot with their elements in hex — for finding where
// the map has drifted between OpenLoco and goloco (LOOP priority 1).
//
//	go run ./scripts/olref/tilediff [-at X,Y -r R] a.SV5 b.SV5
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

func main() {
	at := flag.String("at", "", "tile X,Y to list the differences around")
	r := flag.Int("r", 3, "radius in tiles around -at")
	flag.Parse()
	if flag.NArg() != 2 {
		log.Fatal("usage: tilediff [-at X,Y -r R] a.SV5 b.SV5")
	}
	a, b := load(flag.Arg(0)), load(flag.Arg(1))
	ax, ay := -1, -1
	if *at != "" {
		fmt.Sscanf(*at, "%d,%d", &ax, &ay)
	}
	n := 0
	for y := range a.Tiles {
		for x := range a.Tiles[y] {
			ea, eb := elements(&a.Tiles[y][x]), elements(&b.Tiles[y][x])
			if ea == eb {
				continue
			}
			n++
			if ax >= 0 && abs(x-ax) <= *r && abs(y-ay) <= *r {
				fmt.Printf("tile %d,%d\n  a %s\n  b %s\n", x, y, ea, eb)
			}
		}
	}
	fmt.Printf("%d tiles differ\n", n)
}

func load(path string) *scenario.Scenario {
	sc, err := scenario.LoadScenarioData(path)
	if err != nil {
		log.Fatalf("%s: %v", path, err)
	}
	return sc
}

// elements is the tile's elements in hex, one 8-byte group each.
func elements(t *scenario.Tile) string {
	enc := scenario.EncodeTileElements([][]scenario.Tile{{*t}})
	s := ""
	for i := 0; i < len(enc); i += 8 {
		s += fmt.Sprintf("%x ", enc[i:i+8])
	}
	return s
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
