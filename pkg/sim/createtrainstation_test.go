package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// A train station built on a free straight piece of a company's track in the
// title save: the dry run changes nothing; applied, the tile gets a station
// element over the track, the station (joined or new) has the tile, is fully
// created, serves rail and has a label; the cost is what the dry run said.
func TestCreateTrainStation(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	g1, _ := om.G1File.(G1Raw)
	stationType := uint8(0xFF)
	for i, s := range om.TrainStationObjects {
		if s != nil {
			stationType = uint8(i)
			break
		}
	}
	for y := 4; y < m.Height-4; y++ {
		for x := 4; x < m.Width-4; x++ {
			tl := m.TileAt(x, y)
			for k, r := range tl.Order {
				if r.Kind != scenario.KindTrack {
					continue
				}
				tr := tl.Tracks[r.Index]
				if tr.TrackID != 0 || tr.HasStation || tr.HasSignal || tr.HasLevelCrossing || tr.Ghost || tr.AiAllocated || tr.HasBridge {
					continue
				}
				c := &Cmd{Company: tr.Owner, Objects: om}
				a := TrainStationPlacement{X: int32(x) * 32, Y: int32(y) * 32, Z: int32(tr.BaseZ) * 4, Rotation: tr.Rotation, TrackID: 0, TrackObjectID: tr.TrackObjectID, Type: stationType}
				nOrder := len(tl.Order)
				dry, ok := m.CreateTrainStation(c, 0, a, g1)
				if !ok {
					continue // no room above, a junction …: try another
				}
				if len(tl.Order) != nOrder {
					t.Fatal("the dry run changed the tile")
				}
				cost, ok := m.CreateTrainStation(c, CmdApply, a, g1)
				if !ok || cost != dry || cost <= 0 {
					t.Fatalf("applying: ok %v cost %d (dry run %d)", ok, cost, dry)
				}
				el := m.getStationElement(a.X, a.Y, a.Z)
				if el == nil || tl.Order[k+1].Kind != scenario.KindStation || !tl.Tracks[r.Index].HasStation {
					t.Fatalf("no station element after the track at %d,%d", x, y)
				}
				st := m.StationNames[el.StationID]
				found := false
				for i := 0; i < int(st.TileSize); i++ {
					found = found || st.Tiles[i][0] == int16(a.X) && st.Tiles[i][1] == int16(a.Y)
				}
				if !found || st.Flags&stationFlag5 != 0 || st.Flags&1 == 0 || st.LabelFrame.Right[0] == st.LabelFrame.Left[0] {
					t.Errorf("station %d %q: tile registered %v, flags %#x, label %+v", el.StationID, m.StationName(int(el.StationID)), found, st.Flags, st.LabelFrame)
				}
				t.Logf("built on %d,%d (company %d): %q, cost %d", x, y, tr.Owner, m.StationName(int(el.StationID)), cost)
				return
			}
		}
	}
	t.Fatal("no free straight track piece took a station")
}
