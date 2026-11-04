package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// EntityManager's lists after reset: a new vehicle entity takes the lowest
// free slot, a freed slot is taken again first, the counts follow.
func TestEntityListsReset(t *testing.T) {
	l := resetEntityLists()
	if l.numFreeEntities() != maxNormalEntities || l.Counts[EntityListNullMoney] != maxEntities-maxNormalEntities {
		t.Fatalf("counts %v", l.Counts)
	}
	a, _ := l.createEntityVehicle()
	b, _ := l.createEntityVehicle()
	if a != 0 || b != 1 || l.Heads[EntityListVehicle] != 1 || l.Counts[EntityListVehicle] != 2 || l.numFreeEntities() != maxNormalEntities-2 {
		t.Fatalf("created %d %d; heads %v counts %v", a, b, l.Heads, l.Counts)
	}
	l.freeEntity(0)
	if c, _ := l.createEntityVehicle(); c != 0 {
		t.Errorf("after freeing 0 the next is %d (the freed slot is the free list's head)", c)
	}
	l.moveEntityToList(maxNormalEntities+5, EntityListMisc) // a money effect in use
	l.freeEntity(maxNormalEntities + 5)
	if l.Heads[EntityListNullMoney] != maxNormalEntities+5 {
		t.Error("a money slot goes back to the money list")
	}
}

// The lists of real saves (local game data; skips without): each list walked
// from its head is as long as its count, links both ways, and the free and
// vehicle lists hold exactly the free and vehicle slots.
func TestEntityListsFromSaves(t *testing.T) {
	for _, f := range []string{"Data/title.dat", "Single Player Saved Games/DEBUG_AUTO_DUMP.SV5"} {
		sc, err := scenario.LoadScenarioData("../../../locomotion/" + f)
		if err != nil || sc.EntityLists == nil {
			t.Logf("%s: not present", f)
			continue
		}
		var m Map
		m.LoadScenario(sc)
		l := m.EntityLists
		for list := range 7 {
			n, prev := 0, uint16(entityIDNull)
			for id := l.Heads[list]; id != entityIDNull && n <= maxEntities; id = l.Links[id].Next {
				if l.Links[id].ListOffset != uint8(list*2) || l.Links[id].Prev != prev {
					t.Fatalf("%s list %d: slot %d offset %d prev %d (want %d)", f, list, id, l.Links[id].ListOffset, l.Links[id].Prev, prev)
				}
				prev = id
				n++
			}
			if n != int(l.Counts[list]) {
				t.Errorf("%s list %d: %d walked, count %d", f, list, n, l.Counts[list])
			}
		}
		vehicles := 0
		for _, ln := range l.Links {
			if ln.BaseType == scenario.EntityBaseTypeVehicle {
				vehicles++
			}
		}
		if vehicles != len(m.Entities) || int(l.Counts[EntityListVehicle])+int(l.Counts[EntityListVehicleHead]) != len(m.Entities) {
			t.Errorf("%s: %d vehicle slots, %d entities, lists vehicle %d + heads %d", f, vehicles, len(m.Entities), l.Counts[EntityListVehicle], l.Counts[EntityListVehicleHead])
		}
		t.Logf("%s: counts %v, free head %d, %d vehicle entities", f, l.Counts, l.Heads[EntityListNull], len(m.Entities))
	}
}
