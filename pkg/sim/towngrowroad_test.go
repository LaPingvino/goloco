package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

func townRoadMap(w int, ground uint8) *Map {
	m := &Map{Width: w, Height: w, Tiles: make([][]Tile, w), Towns: []scenario.Town{{}}, CurrentYear: 1950}
	m.Towns[0].Prng = [2]uint32{0x1234, 0x5678}
	for y := range m.Tiles {
		m.Tiles[y] = make([]Tile, w)
		for x := range m.Tiles[y] {
			m.Tiles[y][x].BaseZ = ground
			m.Tiles[y][x].Order = []scenario.ElemRef{{Kind: scenario.KindSurface}}
		}
	}
	return m
}

func townRoadCmd() *Cmd {
	h := func(n string) *objects.ObjectHeader { x := &objects.ObjectHeader{}; copy(x.Name[:], n); return x }
	road := &objects.RoadObject{Header: h("ROADTOWN"), BuildCostFactor: 100, BridgeNames: []string{"BRDGA"}}
	bridge := &objects.BridgeObject{Header: h("BRDGA"), MaxHeight: 10, ClearHeight: 16, DesignedYear: 1900, BaseCostFactor: 50}
	return &Cmd{Company: CompanyNeutral, Objects: &objects.ObjectManager{RoadObjects: []*objects.RoadObject{road}, BridgeObjects: []*objects.BridgeObject{bridge}}}
}

// placeRoadBridge: a valley three tiles wide, 40 below the road, is bridged
// (three pieces, each with the bridge); flat ground gets no bridge (neither
// high nor over water) and nothing is built.
func TestPlaceRoadBridge(t *testing.T) {
	c := townRoadCmd()
	m := townRoadMap(16, 12) // the ground at z 48
	for x := 5; x <= 7; x++ {
		m.TileAt(x, 3).BaseZ = 2 // a valley at z 8
	}
	// rotation 2: RotationOffset (+32, 0), eastwards along x
	if !m.placeRoadBridge(c, 0, 5*32, 3*32, 48, 2, 0) {
		t.Fatal("the valley was not bridged")
	}
	for x := 5; x <= 7; x++ {
		rs := m.TileAt(x, 3).Roads
		if len(rs) != 1 || !rs[0].HasBridge || rs[0].BridgeID != 0 || rs[0].BaseZ != 12 {
			t.Errorf("tile %d: %+v", x, rs)
		}
	}
	if len(m.TileAt(8, 3).Roads) != 0 {
		t.Error("the bridge ran on past the valley")
	}
	if m.Towns[0].Prng == [2]uint32{0x1234, 0x5678} {
		t.Error("the town's prng was not drawn from")
	}
	flat := townRoadMap(16, 4)
	if flat.placeRoadBridge(c, 0, 5*32, 3*32, 16, 2, 0) || len(flat.TileAt(5, 3).Roads) != 0 {
		t.Error("flat ground: no bridge, nothing built")
	}
}
