package sim

import (
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/objects"
)

// TownGrowAll is TownGrowFlags::all.
const TownGrowAll = TownGrowBuildInitialRoad | TownGrowRoadUpdate | TownGrowNeutralRoadTakeover | TownGrowAllowRoadExpansion |
	TownGrowAllowRoadBranching | TownGrowConstructBuildings | TownGrowBuildImmediately | TownGrowAlwaysUpdate

// CreateTown is the createTown game command (CreateTown.cpp 0x00496C22) for
// a town of this size at (x, y), or anywhere when x is −1 (FindTownSite):
// at a given spot up to 40 tries round it (upstream's first try is
// overwritten by a random one — kept), a [−2, 1] tile jitter for the first
// 11 and [−3, 4] after, inside a 5-tile border, a site whose surroundings
// pass (no water check); the town initialised (too_many_towns when there is
// no slot); applied, grown as upstream — the year moved back 51, eight
// rounds of size² grows with every flag, recalculating the size, the year
// up 7 after each — and its last history entry set (to 255 always, upstream's
// std::max quirk). The dry run frees the town again. Its index, or −1.
// PARTIAL: madeAnyChanges.
func (m *Map) CreateTown(c *Cmd, flags uint8, x, y int32, size uint8, g1 G1Raw, om *objects.ObjectManager) (int, bool) {
	px, py := x, y
	if x == -1 {
		var ok bool
		if px, py, ok = m.FindTownSite(om); !ok {
			return -1, false
		}
	} else {
		found := false
		for attempts := 0; attempts < 40; attempts++ {
			r := m.Rng1.RandNext()
			var ox, oy int32
			if attempts <= 10 {
				ox, oy = int32((r>>9)&3)-2, int32((r>>5)&3)-2
			} else {
				ox, oy = int32((r>>9)&7)-3, int32((r>>5)&7)-3
			}
			ax, ay := x+ox*tileSize, y+oy*tileSize
			if ax < 160 || ay < 160 || ax > 12128 || ay > 12128 {
				continue
			}
			if m.townSiteOK(ax, ay, false, om) {
				px, py, found = ax, ay, true
				break
			}
		}
		if !found {
			return -1, false
		}
	}
	i := m.InitialiseTown(int16(px), int16(py), g1, om)
	if i < 0 {
		c.ErrorText = i18n.TooManyTowns
		return -1, false
	}
	if flags&CmdApply == 0 {
		m.emptyUserString(m.Towns[i].Name)
		m.Towns = append(m.Towns[:i], m.Towns[i+1:]...)
		return -1, true
	}
	backup := m.CurrentYear
	m.CurrentYear = backup - 51
	growth := int(size) * int(size)
	for round := 8; round > 0; round-- {
		for j := growth; j > 0; j-- {
			m.GrowTown(om, i, TownGrowAll)
			m.RecalculateTownSize(i)
		}
		m.CurrentYear += 7
	}
	m.CurrentYear = backup
	tw := &m.Towns[i]
	if tw.HistorySize > 0 {
		tw.History[tw.HistorySize-1] = 255 // std::max<uint8_t>(population / 50, 255)
	}
	land, _ := m.GetHeight(int32(tw.X), int32(tw.Y))
	c.Pos = &[3]int32{int32(tw.X) + tileSize/2, int32(tw.Y) + tileSize/2, int32(land)}
	return i, true
}

// GenerateTowns is MapGenerator's generateTowns: numberOfTowns towns placed
// anywhere (createTown, applied), each of a size drawn from 1..maxTownSize
// (gPrng1's randNext(1, max); 1 when the maximum is 1 or less).
func (m *Map) GenerateTowns(o TerrainOptions, g1 G1Raw, om *objects.ObjectManager) {
	c := &Cmd{Company: CompanyNeutral, Objects: om}
	for i := 0; i < int(o.NumberOfTowns); i++ {
		size := uint8(1)
		if o.MaxTownSize > 1 {
			size = uint8(m.Rng1.RandNextRange(1, int32(o.MaxTownSize)))
		}
		m.CreateTown(c, CmdApply, -1, -1, size, g1, om)
	}
}
