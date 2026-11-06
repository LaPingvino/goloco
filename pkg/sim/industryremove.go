package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Industries removed (item CR): the removeIndustry game command
// (GameCommands/Industries/RemoveIndustry.cpp 0x00455943) with
// removeIndustryElement (0x0045579F) and revokeAllSurfaceClaims
// (0x00455A5C) — what a closed-down industry's last month runs. PARTIAL:
// the windows (industry, industry list), the user string,
// MessageManager::removeAllSubjectRefs (goloco has no messages).

// removeIndustryElement is removeIndustryElement: at the tile and height,
// the industry building's element on each of its tiles (one, or four for a
// 2×2 building) removed, and that tile dropped from the industry's list.
func (m *Map) removeIndustryElement(x, y, z int32, om *objects.ObjectManager) {
	t := m.TileAt(int(x)>>5, int(y)>>5)
	if t == nil {
		return
	}
	for i := range t.Industries {
		el := &t.Industries[i]
		if int32(el.BaseZ)*4 != z {
			continue
		}
		ind := m.industryByID(el.IndustryID)
		if ind == nil {
			continue
		}
		obj := om.GetIndustryObjectByIndex(int(ind.ObjectID))
		if obj == nil {
			continue
		}
		large := obj.BuildingSizeFlags&(1<<el.BuildingType) != 0
		for _, off := range buildingTileOffsets(large) {
			bt := m.TileAt(int(x+off[0])>>5, int(y+off[1])>>5)
			if bt == nil {
				continue
			}
			for j := range bt.Industries {
				if int32(bt.Industries[j].BaseZ)*4 != z {
					continue
				}
				bt.Industries = append(bt.Industries[:j], bt.Industries[j+1:]...)
				bt.removeRef(scenario.KindIndustry, j)
				break
			}
		}
		for k, p := range ind.Tiles {
			if int32(p[0]) == x && int32(p[1]) == y && int32(p[2]&0x7FFF) == z {
				ind.Tiles = append(ind.Tiles[:k], ind.Tiles[k+1:]...)
				break
			}
		}
		return
	}
}

// revokeAllSurfaceClaims is revokeAllSurfaceClaims: every industrial surface
// of the industry back to plain ground (not industrial, growth 0, variation
// 0) and the walls above it gone.
func (m *Map) revokeAllSurfaceClaims(id uint8) {
	for ty := 0; ty < m.Height; ty++ {
		for tx := 0; tx < m.Width; tx++ {
			t := m.TileAt(tx, ty)
			if t == nil || !t.IsIndustrial || t.Variation != id { // surface->industryId(): byte 7
				continue
			}
			t.IsIndustrial, t.GrowthStage, t.Variation = false, 0, 0
			t.removeAllWallsAbove(t.BaseZ)
		}
	}
}

// RemoveIndustry is the removeIndustry game command: with CmdApply its
// buildings' elements removed (always the first tile left, as many times
// as it has tiles), its surface claims revoked, the stations' cargo stats
// forgetting it, the industry gone; the clearing cost (the object's
// clearCostFactor, inflation adjusted). ok false: no such industry or
// object.
func (m *Map) RemoveIndustry(c *Cmd, flags uint8, id uint8) (int32, bool) {
	ind := m.industryByID(id)
	if ind == nil || c.Objects == nil {
		return 0, false
	}
	obj := c.Objects.GetIndustryObjectByIndex(int(ind.ObjectID))
	if obj == nil {
		return 0, false
	}
	if flags&CmdApply != 0 {
		for n := len(ind.Tiles); n != 0; n-- {
			if len(ind.Tiles) == 0 {
				break
			}
			p := ind.Tiles[0]
			m.removeIndustryElement(int32(p[0]), int32(p[1]), int32(p[2]&0x7FFF), c.Objects)
			if ind = m.industryByID(id); ind == nil {
				break
			}
		}
		m.revokeAllSurfaceClaims(id)
		for s := range m.StationNames {
			for k := range m.StationNames[s].Cargo {
				if m.StationNames[s].Cargo[k].IndustryID == id {
					m.StationNames[s].Cargo[k].IndustryID = 0xFF // IndustryId::null
				}
			}
		}
		for i := range m.Industries {
			if m.Industries[i].ID == id {
				m.Industries = append(m.Industries[:i], m.Industries[i+1:]...)
				break
			}
		}
	}
	return m.InflationAdjustedCost(obj.ClearCostFactor, obj.CostIndex, 3), true
}
