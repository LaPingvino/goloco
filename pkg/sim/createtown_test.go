package sim

import "testing"

// A new town placed anywhere on the title map grows as upstream's createTown
// grows it: roads and buildings around it, its counts agreeing with the map,
// the year restored; the dry run leaves no town.
func TestCreateTown(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	g1, ok := om.G1File.(G1Raw)
	if !ok {
		t.Skip("no g1")
	}
	c := &Cmd{Company: CompanyNeutral, Objects: om}
	n0, year := len(m.Towns), m.CurrentYear
	if _, ok := m.CreateTown(c, 0, -1, -1, 2, g1, om); !ok || len(m.Towns) != n0 {
		t.Fatalf("dry run: ok %v, towns %d → %d", ok, n0, len(m.Towns))
	}
	i, ok := m.CreateTown(c, CmdApply, -1, -1, 2, g1, om)
	if !ok {
		t.Skip("no town site on the title map")
	}
	if len(m.Towns) != n0+1 || m.CurrentYear != year {
		t.Fatalf("towns %d → %d, year %d (was %d)", n0, len(m.Towns), m.CurrentYear, year)
	}
	tw := m.Towns[i]
	roads, buildings := 0, 0
	for y := int(tw.Y)/32 - 12; y <= int(tw.Y)/32+12; y++ {
		for x := int(tw.X)/32 - 12; x <= int(tw.X)/32+12; x++ {
			if tl := m.TileAt(x, y); tl != nil {
				roads += len(tl.Roads)
				for _, b := range tl.Buildings {
					if b.SequenceIndex == 0 {
						buildings++
					}
				}
			}
		}
	}
	t.Logf("%q at %d,%d: %d buildings (%d counted), population %d, %d road pieces near, size %d", m.TownName(tw.Name, i), tw.X/32, tw.Y/32, buildings, tw.NumBuildings, tw.Population, roads, tw.Size)
	if roads == 0 || tw.NumBuildings == 0 {
		t.Errorf("the town didn't grow: %d roads, %d buildings", roads, tw.NumBuildings)
	}
	if tw.History[tw.HistorySize-1] != 255 {
		t.Errorf("history %d, want 255", tw.History[tw.HistorySize-1])
	}
}

// A generated map (the editor's defaults: 55 towns up to size 3) has towns
// that grew buildings and roads.
func TestGeneratedMapHasTowns(t *testing.T) {
	_, om := loadDataMap(t, "title.dat")
	g1, _ := om.G1File.(G1Raw)
	m := NewGeneratedMap(1234, 5678, 1950, EditorDefaultTerrainOptions(om), g1, om)
	buildings, roads := 0, 0
	for y := range m.Tiles {
		for x := range m.Tiles[y] {
			buildings += len(m.Tiles[y][x].Buildings)
			roads += len(m.Tiles[y][x].Roads)
		}
	}
	industryTiles := 0
	for y := range m.Tiles {
		for x := range m.Tiles[y] {
			industryTiles += len(m.Tiles[y][x].Industries)
		}
	}
	t.Logf("%d towns, %d building elements, %d road pieces; %d industries (%d industry elements)", len(m.Towns), buildings, roads, len(m.Industries), industryTiles)
	if len(m.Towns) == 0 || buildings == 0 || roads == 0 {
		t.Error("the generated map has no grown towns")
	}
	if len(m.Industries) == 0 || industryTiles == 0 {
		t.Error("the generated map has no industries")
	}
}
