package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Every live dock tile of the title save, from its owner: the nearby dock is
// its own station, attached; validateNearbyDock says okay with it; and some
// tile next to it has water. The dock objects load a build cost.
func TestNearbyDockOwnTiles(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	for i, d := range om.DockObjects {
		if d != nil && (d.BuildCostFactor <= 0 || d.SellCostFactor >= 0) {
			t.Errorf("dock object %d %q: build %d sell %d", i, d.Name, d.BuildCostFactor, d.SellCostFactor)
		}
	}
	ok, total := 0, 0
	for id := range m.StationNames {
		st := &m.StationNames[id]
		if st.Name == stationNameNull || st.Flags&stationFlag5 != 0 {
			continue
		}
		for i := 0; i < int(st.TileSize); i++ {
			p := st.Tiles[i]
			el := m.getStationElement(int32(p[0]), int32(p[1]), int32(p[2]&^3))
			if el == nil || el.StationType != 3 {
				continue
			}
			total++
			x, y, z := int32(p[0]), int32(p[1]), int32(p[2]&^3)
			got, attached := m.findNearbyStationDocks(x, y, z, st.Owner, false)
			res, v := m.validateNearbyDock(x, y, z, st.Owner, false)
			water := false
			for _, o := range portBorderOffsets {
				water = water || m.tileHasWater(x+int32(o[0])*32, y+int32(o[1])*32)
			}
			if got == id && attached && res == nearbyOkay && v == id && water {
				ok++
			} else {
				t.Errorf("dock %d %q tile %v: found %d attached %v, validate %d/%d, water beside %v", id, m.StationName(id), p, got, attached, res, v, water)
			}
		}
	}
	if total == 0 {
		t.Skip("no docks in the title save")
	}
	t.Logf("%d of %d dock tiles find their own dock", ok, total)
	_ = scenario.KindStation
}

// A dock element put on a tile (the title save has none): the search finds it
// from its 4×4 window (-1 … +2), attached, as its owner only; a ghost or an
// AI-allocated one only by the matching variant. The water tests agree with
// the tiles' water levels.
func TestFindNearbyStationDocksWindow(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	id := -1
	for i := range m.StationNames {
		if m.StationNames[i].Name != stationNameNull {
			id = i
			break
		}
	}
	const company = 3
	tx, ty := 40, 40
	tl := m.TileAt(tx, ty)
	el := scenario.StationElement{StationType: 3, Owner: company, StationID: uint16(id), BaseZ: tl.BaseZ}
	tl.Stations = append(tl.Stations, el)
	tl.Order = append(tl.Order, scenario.ElemRef{Kind: scenario.KindStation, Index: uint16(len(tl.Stations) - 1)})
	dock := &tl.Stations[len(tl.Stations)-1]
	from := func(dx, dy int, c uint8, ai bool) (int, bool) {
		return m.findNearbyStationDocks(int32(tx+dx)*32, int32(ty+dy)*32, int32(tl.BaseZ)*4, c, ai)
	}
	for _, d := range [][2]int{{1, 1}, {-2, -2}, {0, 0}} { // dock at search tile +1 / +2 / 0
		if got, att := from(d[0], d[1], company, false); got != id || !att {
			t.Errorf("from %+v: %d attached %v, want %d", d, got, att, id)
		}
	}
	if got, att := from(2, 2, company, false); got == id && att { // the dock at -2: outside the window
		t.Errorf("from +2,+2 the dock (at -2) was found attached")
	}
	if _, att := from(0, 0, company+1, false); att {
		t.Errorf("another company found it attached")
	}
	dock.AiAllocated = true
	if got, att := from(0, 0, company, false); got == id && att {
		t.Errorf("an AI-allocated dock found by the player's search")
	}
	if got, att := from(0, 0, company, true); got != id || !att {
		t.Errorf("the AI search missed its dock")
	}
	dock.AiAllocated, dock.Ghost = false, true
	if got, att := from(0, 0, company, false); got == id && att {
		t.Errorf("a ghost dock was found")
	}
	wet, dry := 0, 0
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			if m.tileHasWater(int32(x)*32, int32(y)*32) != (m.TileAt(x, y).WaterLevel != 0) {
				t.Fatalf("tileHasWater disagrees at %d,%d", x, y)
			}
			if m.TileAt(x, y).WaterLevel != 0 {
				wet++
			} else {
				dry++
			}
		}
	}
	if m.tileHasWater(-32, 0) || m.tileHasWater(int32(m.Width)*32, 0) {
		t.Error("off the map counted as water")
	}
	ports := 0
	for y := 1; y < m.Height-2; y++ {
		for x := 1; x < m.Width-2; x++ {
			if m.isValidWaterIndustryPort(int32(x)*32, int32(y)*32, om) {
				ports++
			}
		}
	}
	t.Logf("water tiles %d, dry %d; tiles next to a water industry %d", wet, dry, ports)
}

// A dock on the shore of the title save, built by the player: the dry run
// changes nothing and costs what applying does; applied, its four tiles carry
// a docks element of the station, which registers the tile. On open water
// (no water industry in the save) a dock is refused.
func TestCreatePort(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	g1, _ := om.G1File.(G1Raw)
	typ := -1
	for i, d := range om.DockObjects {
		if d != nil {
			typ = i
			break
		}
	}
	if typ < 0 {
		t.Skip("no dock object")
	}
	c := &Cmd{Company: m.ControllingCompany, Objects: om}
	refusedOnWater := false
	for y := 2; y < m.Height-3; y++ {
		for x := 2; x < m.Width-3; x++ {
			for rot := uint8(0); rot < 4; rot++ {
				tl := m.TileAt(x, y)
				a := PortPlacement{X: int32(x) * 32, Y: int32(y) * 32, Z: int32(tl.BaseZ) * 4, Rotation: rot, Type: uint8(typ)}
				if tl.WaterLevel != 0 {
					if !refusedOnWater {
						if _, ok := m.CreatePort(c, 0, a, g1); ok {
							t.Errorf("a dock on open water at %d,%d was allowed", x, y)
						}
						refusedOnWater = true
					}
					continue
				}
				n := len(tl.Order)
				dry, ok := m.CreatePort(c, 0, a, g1)
				if !ok {
					continue
				}
				if len(tl.Order) != n {
					t.Fatal("the dry run changed the tile")
				}
				cost, ok := m.CreatePort(c, CmdApply, a, g1)
				if !ok || cost != dry || cost <= 0 {
					t.Fatalf("applying at %d,%d rot %d: ok %v cost %d (dry run %d)", x, y, rot, ok, cost, dry)
				}
				id := -1
				for _, d := range [][2]int{{0, 0}, {1, 0}, {0, 1}, {1, 1}} {
					tt := m.TileAt(x+d[0], y+d[1])
					found := false
					for _, r := range tt.Order {
						if r.Kind == scenario.KindStation && tt.Stations[r.Index].StationType == 3 {
							found = true
							id = int(tt.Stations[r.Index].StationID)
						}
					}
					if !found {
						t.Errorf("no docks element on %d,%d", x+d[0], y+d[1])
					}
				}
				st := m.StationNames[id]
				reg := false
				for i := 0; i < int(st.TileSize); i++ {
					reg = reg || st.Tiles[i][0] == int16(a.X) && st.Tiles[i][1] == int16(a.Y)
				}
				if !reg || st.Owner != c.Company {
					t.Errorf("station %d %q: registered %v owner %d", id, m.StationName(id), reg, st.Owner)
				}
				t.Logf("dock at %d,%d rotation %d: %q, cost %d", x, y, rot, m.StationName(id), cost)
				// and removed again: the four elements go, the station is left
				// without tiles, the removal refunds; a second removal finds nothing
				refund, ok := m.RemovePort(c, CmdApply, PortRemoval{X: a.X, Y: a.Y, Z: a.Z}, g1)
				if !ok || refund >= 0 {
					t.Fatalf("removing: ok %v refund %d", ok, refund)
				}
				for _, d := range [][2]int{{0, 0}, {1, 0}, {0, 1}, {1, 1}} {
					tt := m.TileAt(x+d[0], y+d[1])
					for _, r := range tt.Order {
						if r.Kind == scenario.KindStation && tt.Stations[r.Index].StationType == 3 {
							t.Errorf("a docks element left on %d,%d", x+d[0], y+d[1])
						}
					}
				}
				if n := m.StationNames[id].TileSize; n != 0 {
					t.Errorf("station %d still has %d tiles", id, n)
				}
				if _, ok := m.RemovePort(c, CmdApply, PortRemoval{X: a.X, Y: a.Y, Z: a.Z}, g1); ok {
					t.Error("a second removal succeeded")
				}
				t.Logf("removed: refund %d", refund)
				if !refusedOnWater {
					t.Log("(no open water tested before the dock)")
				}
				return
			}
		}
	}
	t.Fatal("no shore took a dock")
}

// DockPlacementAt over the title save: wherever it offers a placement, the
// rotation's front tiles are water (what CreatePort requires), except where a
// water industry sets the side; dry inland tiles offer none. And the dock
// placements it offers along the shore are buildable by CreatePort somewhere.
func TestDockPlacementAt(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	typ := -1
	for i, d := range om.DockObjects {
		if d != nil {
			typ = i
			break
		}
	}
	g1, _ := om.G1File.(G1Raw)
	c := &Cmd{Company: m.ControllingCompany, Objects: om}
	offered, frontWet, built := 0, 0, 0
	for y := 3; y < m.Height-4; y++ {
		for x := 3; x < m.Width-4; x++ {
			a, ok := m.DockPlacementAt(int32(x)*32, int32(y)*32, uint8(typ), om)
			if !ok {
				continue
			}
			offered++
			wet := true
			for _, f := range rotationToBuildingFront[a.Rotation] {
				wet = wet && m.tileHasWater(a.X+f[0], a.Y+f[1])
			}
			if wet {
				frontWet++
				if _, ok := m.CreatePort(c, 0, a, g1); ok {
					built++
				}
			}
		}
	}
	if offered == 0 || built == 0 {
		t.Fatalf("offered %d, front wet %d, buildable %d", offered, frontWet, built)
	}
	t.Logf("placements offered %d, water in front %d, buildable (dry run) %d", offered, frontWet, built)
}
