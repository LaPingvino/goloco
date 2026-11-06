package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// From each tile of every live title station, as its owner, the nearby
// station is that station, attached (and a rival company finds none of it).
func TestFindNearbyStationOwnTiles(t *testing.T) {
	m, _ := loadDataMap(t, "title.dat")
	hit, total := 0, 0
	for id := range m.StationNames {
		st := &m.StationNames[id]
		if st.Name == stationNameNull || st.Flags&stationFlag5 != 0 {
			continue
		}
		for i := 0; i < int(st.TileSize); i++ {
			p := st.Tiles[i]
			total++
			got, attached := m.FindNearbyStation(int32(p[0]), int32(p[1]), int32(p[2]&^3), st.Owner)
			if got == id && attached {
				hit++
			} else if total-hit <= 5 {
				t.Errorf("station %d %q tile %v: found %d attached %v", id, m.StationName(id), p, got, attached)
			}
		}
	}
	t.Logf("%d of %d station tiles find their own station", hit, total)
}

// On each train-station tile over a straight piece, a new piece there would
// join that station: validateNearbyStation says okay with its id.
func TestValidateNearbyStationOnOwnTrack(t *testing.T) {
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
			if el == nil || el.StationType != 0 {
				continue
			}
			for k, r := range tl.Order {
				if r.Kind != scenario.KindStation || &tl.Stations[r.Index] != el || k == 0 || tl.Order[k-1].Kind != scenario.KindTrack {
					continue
				}
				tr := tl.Tracks[tl.Order[k-1].Index]
				if tr.TrackID != 0 {
					break
				}
				total++
				res, got := m.validateNearbyStation(int32(p[0]), int32(p[1]), int32(tr.BaseZ)*4, uint16(tr.TrackID)<<3|uint16(tr.Rotation), tr.TrackObjectID, st.Owner)
				if res == nearbyOkay && got == id {
					ok++
				} else if total-ok <= 5 {
					t.Errorf("station %d %q tile %v: %d / %d", id, m.StationName(id), p, res, got)
				}
			}
		}
	}
	t.Logf("%d of %d straight station pieces validate into their station", ok, total)
}
