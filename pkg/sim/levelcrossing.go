package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// setLevelCrossingFlags is TileManager::setLevelCrossingFlags (OpenLoco
// reference: src/Map/TileManager.cpp): at a tile and height, the straight
// track (piece 0) and straight road (piece 0) elements that aren't AI
// allocated are looked at; when only one kind still says level crossing — the
// other went (a removal) — that kind's flag is cleared (a road's also its
// unk7_10 and level crossing object).
func (m *Map) setLevelCrossingFlags(x, y, z int32) {
	t := m.TileAt(int(x/tileSize), int(y/tileSize))
	if t == nil {
		return
	}
	each := func(track func(*scenario.TrackElement), road func(*scenario.RoadElement)) {
		for _, r := range t.Order {
			if int32(t.baseZOf(r))*smallZStep != z {
				continue
			}
			switch r.Kind {
			case scenario.KindTrack:
				if int(r.Index) < len(t.Tracks) {
					if el := &t.Tracks[r.Index]; !el.AiAllocated && el.TrackID == 0 {
						track(el)
					}
				}
			case scenario.KindRoad:
				if int(r.Index) < len(t.Roads) {
					if el := &t.Roads[r.Index]; !el.AiAllocated && el.RoadID == 0 {
						road(el)
					}
				}
			}
		}
	}
	hasTrack, hasRoad := false, false
	each(func(el *scenario.TrackElement) { hasTrack = hasTrack || el.HasLevelCrossing },
		func(el *scenario.RoadElement) { hasRoad = hasRoad || el.HasLevelCrossing })
	if hasTrack == hasRoad {
		return
	}
	each(func(el *scenario.TrackElement) {
		if hasTrack {
			el.HasLevelCrossing = false
		}
	}, func(el *scenario.RoadElement) {
		if hasRoad {
			el.HasLevelCrossing, el.Unk7_10, el.LevelCrossingObjectID = false, false, 0
		}
	})
}
