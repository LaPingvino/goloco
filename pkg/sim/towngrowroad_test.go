package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/core"

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

// appendToRoadEnd: on flat ground a piece is added (straight or, 1% each, a
// curve — as the town's prng draws); past the second try with no buildings
// near, beside a parallel town road, and on water nothing is built; on a slope
// rising ahead (south-west corners up, rotation 0) the steep-slope-up piece.
func TestAppendToRoadEnd(t *testing.T) {
	c := townRoadCmd()
	fresh := func() *Map { m := townRoadMap(16, 4); m.RoadAnyTypeCompatible = 1; return m }

	m := fresh()
	p := core.NewPrngSeed(0x1234, 0x5678)
	want := uint8(0)
	if v := p.RandNext() & 0xFFFF; v < 0x028F {
		want = 1
	} else if v < 0x051E {
		want = 2
	}
	m.appendToRoadEnd(c, 0, 5*32, 5*32, 16, 2, 0, 0, false)
	if rs := m.TileAt(5, 5).Roads; len(rs) != 1 || rs[0].RoadID != want || rs[0].Rotation != 2 {
		t.Fatalf("flat: %+v (want road id %d)", rs, want)
	}

	m = fresh()
	m.appendToRoadEnd(c, 0, 5*32, 5*32, 16, 2, 0, 2, false)
	if len(m.TileAt(5, 5).Roads) != 0 {
		t.Error("third try, no buildings near: nothing built")
	}

	m = fresh()
	m.TileAt(5, 6).Roads = []scenario.RoadElement{{BaseZ: 4, ClearZ: 12, Rotation: 2}} // beside, y+32 (kBuggedRotationOffset[2])
	m.TileAt(5, 6).insertRef(scenario.KindRoad, 0, 4)
	m.appendToRoadEnd(c, 0, 5*32, 5*32, 16, 2, 0, 0, false)
	if len(m.TileAt(5, 5).Roads) != 0 {
		t.Error("beside a parallel town road: nothing built")
	}

	m = fresh()
	m.TileAt(5, 5).WaterLevel = 1
	m.appendToRoadEnd(c, 0, 5*32, 5*32, 16, 2, 0, 0, false)
	if len(m.TileAt(5, 5).Roads) != 0 {
		t.Error("water: nothing built")
	}

	m = fresh()
	m.TileAt(5, 5).Slope = slopeCornerSouth | slopeCornerWest
	m.appendToRoadEnd(c, 0, 5*32, 5*32, 16, 0, 0, 0, false)
	if rs := m.TileAt(5, 5).Roads; len(rs) != 1 || rs[0].RoadID != 7 {
		t.Errorf("slope rising ahead: %+v (want the steep-slope-up piece, 7)", rs)
	}
}
