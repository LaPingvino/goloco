package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// A passenger stop on a free straight road of the title save, built by the
// player's company: the dry run changes nothing and costs what applying does;
// applied, the station element sits right after the road, the road says it
// has one and is taller by the stop's height, and the stop's tile is
// registered in its station.
func TestCreateRoadStation(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	g1, _ := om.G1File.(G1Raw)
	typ := -1
	for i, s := range om.RoadStationObjects {
		if s != nil && s.Name == "Passenger Stop" {
			typ = i
			break
		}
	}
	if typ < 0 {
		t.Skip("no Passenger Stop object in the title save")
	}
	company := m.ControllingCompany
	tried := 0
	for y := 4; y < m.Height-4; y++ {
		for x := 4; x < m.Width-4; x++ {
			tl := m.TileAt(x, y)
			for k, r := range tl.Order {
				if r.Kind != scenario.KindRoad {
					continue
				}
				rd := tl.Roads[r.Index]
				if rd.RoadID != 0 || rd.HasStationElement || rd.HasSignalElement || rd.HasLevelCrossing || rd.Ghost || rd.HasBridge {
					continue
				}
				tried++
				c := &Cmd{Company: company, Objects: om}
				a := RoadStationPlacement{X: int32(x) * 32, Y: int32(y) * 32, Z: int32(rd.BaseZ) * 4, Rotation: rd.Rotation, RoadObjectID: rd.RoadObjectID, Type: uint8(typ)}
				nOrder, clear0 := len(tl.Order), rd.ClearZ
				dry, ok := m.CreateRoadStation(c, 0, a, g1)
				if !ok {
					continue
				}
				if len(tl.Order) != nOrder {
					t.Fatal("the dry run changed the tile")
				}
				cost, ok := m.CreateRoadStation(c, CmdApply, a, g1)
				if !ok || cost != dry || cost <= 0 {
					t.Fatalf("applying: ok %v cost %d (dry run %d)", ok, cost, dry)
				}
				road := tl.Roads[r.Index]
				if k+1 >= len(tl.Order) || tl.Order[k+1].Kind != scenario.KindStation || !road.HasStationElement {
					t.Fatalf("no station element after the road at %d,%d", x, y)
				}
				el := &tl.Stations[tl.Order[k+1].Index]
				if el.StationType != 1 || int(el.ObjectID) != typ || el.Owner != company || road.ClearZ != clear0+om.RoadStationObjects[typ].Height/4 || el.ClearZ != road.ClearZ {
					t.Errorf("element %+v, road clearZ %d (was %d)", *el, road.ClearZ, clear0)
				}
				st := m.StationNames[el.StationID]
				found := false
				for i := 0; i < int(st.TileSize); i++ {
					found = found || st.Tiles[i][0] == int16(a.X) && st.Tiles[i][1] == int16(a.Y)
				}
				if !found || st.Owner != company {
					t.Errorf("station %d %q: tile registered %v, owner %d", el.StationID, m.StationName(int(el.StationID)), found, st.Owner)
				}
				t.Logf("built on %d,%d after %d tries: %q, cost %d", x, y, tried, m.StationName(int(el.StationID)), cost)
				return
			}
		}
	}
	t.Fatalf("no free straight road (of %d) took a stop", tried)
}
