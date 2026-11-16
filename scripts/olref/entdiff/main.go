// entdiff compares two saves' entity tables byte by byte: per differing
// entity its id, its type (base type and vehicle subtype bytes) and each
// differing offset (a vs b) — for finding which fields of which vehicles
// part between OpenLoco and goloco (LOOP priority 1).
//
//	go run ./scripts/olref/entdiff [-n max] [-base a0.SV5,b0.SV5] a.SV5 b.SV5
//
// -base: the same two games re-saved at tick 0; a byte that differed there
// and hasn't changed on either side since is left out (OpenLoco's load drops
// padding and stale bytes its structs don't keep, goloco's carries them).
package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

func main() {
	max := flag.Int("n", 20, "entities to list")
	base := flag.String("base", "", "a0.SV5,b0.SV5: the tick-0 re-saves, to leave out what differed from the start")
	flag.Parse()
	if flag.NArg() != 2 {
		log.Fatal("usage: entdiff [-n max] a.SV5 b.SV5")
	}
	a, b := load(flag.Arg(0)), load(flag.Arg(1))
	var a0, b0 []byte
	if p := strings.SplitN(*base, ",", 2); len(p) == 2 {
		a0, b0 = load(p[0]), load(p[1])
	}
	shown, total := 0, 0
	for id := range scenario.MaxEntities {
		o := scenario.EntityTableOffset + id*scenario.EntitySize
		ea, eb := a[o:o+scenario.EntitySize], b[o:o+scenario.EntitySize]
		diff := ""
		for k := range ea {
			if a0 != nil && a0[o+k] != b0[o+k] && ea[k] == a0[o+k] && eb[k] == b0[o+k] {
				continue
			}
			if ea[k] != eb[k] {
				diff += fmt.Sprintf(" %#x:%02x/%02x", k, ea[k], eb[k])
			}
		}
		if diff == "" {
			continue
		}
		total++
		if shown < *max {
			shown++
			fmt.Printf("entity %d (base %d sub %d):%s\n", id, ea[0], ea[1], diff)
		}
	}
	fmt.Printf("%d entities differ\n", total)
}

func load(path string) []byte {
	sc, err := scenario.LoadScenarioData(path)
	if err != nil || sc.Raw == nil {
		log.Fatalf("%s: %v", path, err)
	}
	return sc.Raw.GameState
}
