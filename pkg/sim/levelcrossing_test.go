package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// A level crossing (made on an empty flat tile of the title map, which has
// none): removing its road straight leaves the track without the level
// crossing flag (setLevelCrossingFlags after the removal), and removing the
// track leaves the road without it (and without its crossing object).
func TestLevelCrossingFlagsAfterRemoval(t *testing.T) {
	for _, removeRoad := range []bool{true, false} {
		m, om := loadDataMap(t, "title.dat")
		c := &Cmd{Company: CompanyNeutral, Objects: om}
		x, y := -1, -1
		for ty := 8; ty < m.Height-8 && x < 0; ty++ {
			for tx := 8; tx < m.Width-8; tx++ {
				if tl := m.TileAt(tx, ty); len(tl.Order) == 1 && tl.Slope == 0 && tl.WaterLevel == 0 {
					x, y = tx, ty
					break
				}
			}
		}
		if x < 0 {
			t.Skip("no empty flat tile")
		}
		tl := m.TileAt(x, y)
		z := tl.BaseZ
		m.AddTrack(x, y, scenario.TrackElement{TrackID: 0, Rotation: 0, BaseZ: z, ClearZ: z + 8, HasLevelCrossing: true, Owner: CompanyNeutral, IsLast: true})
		m.AddRoad(x, y, scenario.RoadElement{RoadID: 0, Rotation: 1, BaseZ: z, ClearZ: z + 8, HasLevelCrossing: true, LevelCrossingObjectID: 1, Unk7_10: true, Owner: CompanyNeutral, IsLast: true})
		var ok bool
		if removeRoad {
			_, ok = m.RemoveRoad(c, CmdApply, RoadRemoval{X: int32(x) * 32, Y: int32(y) * 32, Z: int32(z) * 4, Rotation: 1, RoadID: 0}, nil)
		} else {
			_, ok = m.RemoveTrack(c, CmdApply, TrackRemoval{X: int32(x) * 32, Y: int32(y) * 32, Z: int32(z) * 4, Rotation: 0, TrackID: 0}, nil)
		}
		if !ok {
			t.Fatalf("removing the crossing's %s was refused", map[bool]string{true: "road", false: "track"}[removeRoad])
		}
		if removeRoad {
			if len(tl.Tracks) != 1 || tl.Tracks[0].HasLevelCrossing {
				t.Errorf("road removed: track %+v still has the level crossing flag", tl.Tracks)
			}
		} else if len(tl.Roads) != 1 || tl.Roads[0].HasLevelCrossing || tl.Roads[0].Unk7_10 || tl.Roads[0].LevelCrossingObjectID != 0 {
			t.Errorf("track removed: road %+v keeps its crossing", tl.Roads)
		}
	}
}
