package main

import (
	"fmt"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

func main() {
	sc, _ := scenario.LoadScenarioData("../locomotion/Data/title.dat")
	n := map[int]int{}
	first := map[int][2]int{}
	for y := 0; y < sc.MapHeight; y++ {
		for x := 0; x < sc.MapWidth; x++ {
			t := sc.GetTile(x, y)
			if t != nil && t.IsIndustrial {
				k := int(t.Variation)
				if n[k] == 0 {
					first[k] = [2]int{x, y}
				}
				n[k]++
			}
		}
	}
	fmt.Println(n, first)
}
