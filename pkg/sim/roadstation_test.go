package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// On each road-station tile over a straight road, a new stop there would
// join that station: validateNearbyRoadStation says okay with its id, and
// getElRoad finds the road the stop sits on.
func TestValidateNearbyRoadStationOnOwnRoad(t *testing.T) {
	m, _ := loadDataMap(t, "title.dat")
	ok, total := 0, 0
	for id := range m.StationNames {
		st := &m.StationNames[id]
		if st.Name == stationNameNull || st.Flags&stationFlag5 != 0 {
			continue
		}
		for i := 0; i < int(st.TileSize); i++ {
			p := st.Tiles[i]
			tl := m.TileAt(int(p[0])>>5, int(p[1])>>5)
			el := m.getStationElement(int32(p[0]), int32(p[1]), int32(p[2]&^3))
			if el == nil || el.StationType != 1 {
				continue
			}
			for k, r := range tl.Order {
				if r.Kind != scenario.KindStation || &tl.Stations[r.Index] != el || k == 0 || tl.Order[k-1].Kind != scenario.KindRoad {
					continue
				}
				rd := tl.Roads[tl.Order[k-1].Index]
				if rd.RoadID != 0 {
					break
				}
				total++
				z := int32(rd.BaseZ) * 4
				if got, _ := m.getElRoad(int32(p[0]), int32(p[1]), z, rd.Rotation, rd.RoadObjectID, rd.RoadID, rd.SequenceIndex); got == nil {
					t.Errorf("station %d tile %v: getElRoad found no road", id, p)
				}
				res, got := m.validateNearbyRoadStation(int32(p[0])&^31, int32(p[1])&^31, z, uint16(rd.RoadID)<<3|uint16(rd.Rotation), rd.RoadObjectID, st.Owner)
				if res == nearbyOkay && got == id {
					ok++
				} else if total-ok <= 5 {
					t.Errorf("station %d %q tile %v: %d / %d", id, m.StationName(id), p, res, got)
				}
			}
		}
	}
	if total == 0 {
		t.Skip("no road stations on straight road in the title save")
	}
	t.Logf("%d of %d straight road-station pieces validate into their station", ok, total)
}
