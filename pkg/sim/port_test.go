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
