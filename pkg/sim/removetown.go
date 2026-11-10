package sim

import (
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// RemoveTown is the removeTown game command (RemoveTown.cpp 0x0049711F) for
// the town with this id: refused while a station belongs to it
// (all_stations_near_this_town_must_be_removed_first); applied, the whole
// map (y outer, x inner) loses the town's non-ghost buildings (first tiles of
// non-misc buildings whose closest town it is) and neutral roads, each tile
// rescanned after a removal; then its name is freed and the town goes, its
// industries are removed, and the buildings' influence recounted.
// PARTIAL: the town window / list, the messages' references, madeAnyChanges.
func (m *Map) RemoveTown(c *Cmd, flags uint8, townID uint16, g1 G1Raw, om *objects.ObjectManager) (int32, bool) {
	for _, st := range m.StationNames {
		if st.Name != stationNameNull && st.Town == townID {
			c.ErrorText = i18n.AllStationsNearThisTownMustBeRemovedFirst
			return 0, false
		}
	}
	tw := m.TownByID(townID)
	if tw == nil {
		return 0, false
	}
	if flags&CmdApply == 0 {
		return 0, true
	}
	ours := func(px, py int32) bool {
		i, _, ok := m.ClosestTownAndDensity(px, py)
		return !ok || m.Towns[i].ID == townID
	}
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			px, py := int32(x)*tileSize, int32(y)*tileSize
			for again := true; again; {
				again = false
				t := m.TileAt(x, y)
				for _, r := range t.Order {
					if r.Flags&elemFlagGhost != 0 {
						continue
					}
					if r.Kind == scenario.KindBuilding {
						b := t.Buildings[r.Index]
						if b.IsMisc || b.SequenceIndex != 0 || !ours(px, py) {
							continue
						}
						if _, ok := m.RemoveBuilding(c, px, py, int32(b.BaseZ)*smallZStep, flags); ok {
							again = true
							break
						}
						continue
					}
					if r.Kind == scenario.KindRoad {
						rd := t.Roads[r.Index]
						if rd.Owner != CompanyNeutral || !ours(px, py) {
							continue
						}
						a := RoadRemoval{X: px, Y: py, Z: int32(rd.BaseZ) * smallZStep, Rotation: rd.Rotation, RoadID: rd.RoadID, Index: rd.SequenceIndex, RoadObjectID: rd.RoadObjectID}
						if _, ok := m.RemoveRoad(c, flags, a, g1); ok {
							again = true
							break
						}
					}
				}
			}
		}
	}
	tw = m.TownByID(townID)
	cx, cy := int32(tw.X), int32(tw.Y)
	m.emptyUserString(tw.Name)
	for i := range m.Towns {
		if m.Towns[i].ID == townID {
			m.Towns = append(m.Towns[:i], m.Towns[i+1:]...) // name = null: an empty slot (Towns skips them)
			break
		}
	}
	for _, ind := range append([]scenario.Industry(nil), m.Industries...) {
		if ind.Town == townID && ind.Name != 0xFFFF {
			m.RemoveIndustry(c, flags, ind.ID)
		}
	}
	land, _ := m.GetHeight(cx, cy)
	c.Pos = &[3]int32{cx, cy, int32(land)}
	m.ResetBuildingsInfluence(om)
	return 0, true
}
