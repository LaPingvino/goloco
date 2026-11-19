package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// GameCommands/Company/BuildCompanyHeadquarters.cpp and
// RemoveCompanyHeadquarters.cpp; Company::setHeadquartersVariation (the
// monthly look by performance).

// HeadquarterPlacement is HeadquarterPlacementArgs: where (world units, z
// the base height), rotation, the HQ building object, built at once.
type HeadquarterPlacement struct {
	X, Y, Z          int32
	Rotation         uint8
	Type             uint8
	BuildImmediately bool
}

// headquarterPerformanceVariation is Company::getHeadquarterPerformanceVariation:
// min(performanceIndex / 200, 4).
func (m *Map) headquarterPerformanceVariation(company int) uint8 {
	if company < 0 || company >= len(m.CompanyFaces) {
		return 0
	}
	return uint8(min(int(m.CompanyFaces[company].Performance)/200, 4))
}

// RemoveCompanyHeadquarters is removeCompanyHeadquarters: not a ghost, the
// company whose HQ stands there (its owner must be compatible); then the
// building removed (removeBuilding, nested); applied (not a ghost), that
// company has no HQ. The cost. PARTIAL: the company window invalidation.
func (m *Map) RemoveCompanyHeadquarters(c *Cmd, flags uint8, x, y, z int32) (int32, bool) {
	c.Pos = &[3]int32{x, y, z}
	target := -1
	if flags&CmdGhost == 0 {
		for i := range m.Companies {
			if m.CompanyEmpty(i) {
				continue
			}
			hq := m.Companies[i]
			if int32(hq.HeadquartersX) != x || int32(hq.HeadquartersY) != y || int32(hq.HeadquartersZ) != z/smallZStep {
				continue
			}
			if !checkCompanyCompatibility(c, uint8(i)) {
				return 0, false
			}
			target = i
			break
		}
	}
	cost, ok := nested(flags, func(f uint8) (int32, bool) { return m.RemoveBuilding(c, x, y, z, f) })
	if !ok {
		return 0, false
	}
	if flags&CmdApply != 0 && flags&CmdGhost == 0 && target >= 0 {
		m.Companies[target].HeadquartersX = -1
	}
	if flags&CmdApply != 0 && flags&(CmdGhost|CmdAiAllocated) == 0 {
		m.playSound(soundDemolish, ChannelEffects, x, y, z) // the demolish sound at pos
	}
	return cost, true
}

// BuildCompanyHeadquarters is buildCompanyHeadquarters: the updating
// company's old HQ removed first (not for a ghost); then the building in its
// performance variation and colour (createBuilding, nested); applied (not a
// ghost), the company's HQ is there. The cost. PARTIAL: the company window
// invalidation, the construct sound.
func (m *Map) BuildCompanyHeadquarters(c *Cmd, flags uint8, a HeadquarterPlacement) (int32, bool) {
	c.Pos = &[3]int32{a.X + tileSize/2, a.Y + tileSize/2, a.Z}
	company := int(c.Company)
	if company < 0 || company >= len(m.Companies) {
		return 0, false
	}
	var total int32
	if hq := m.Companies[company]; hq.HeadquartersX != -1 && flags&CmdGhost == 0 {
		cost, ok := nested(flags, func(f uint8) (int32, bool) {
			return m.RemoveCompanyHeadquarters(c, f, int32(hq.HeadquartersX), int32(hq.HeadquartersY), int32(hq.HeadquartersZ)*smallZStep)
		})
		if !ok {
			return 0, false
		}
		total += cost
	}
	colour := uint8(0)
	if company < len(m.CompanyColours) {
		colour = m.CompanyColours[company]
	}
	cost, ok := nested(flags, func(f uint8) (int32, bool) {
		return m.CreateBuilding(c, f, BuildingPlacement{X: a.X, Y: a.Y, Z: a.Z, Rotation: a.Rotation, Type: a.Type,
			Variation: m.headquarterPerformanceVariation(company), Colour: colour, BuildImmediately: a.BuildImmediately})
	})
	if !ok {
		return 0, false
	}
	total += cost
	if flags&CmdApply != 0 && flags&CmdGhost == 0 {
		m.Companies[company].HeadquartersX = int16(a.X)
		m.Companies[company].HeadquartersY = int16(a.Y)
		m.Companies[company].HeadquartersZ = uint8(a.Z / smallZStep)
	}
	if flags&CmdApply != 0 && flags&(CmdAiAllocated|CmdGhost) == 0 {
		m.playSound(soundConstruct, ChannelEffects, a.X, a.Y, a.Z) // args.pos
	}
	return total, true
}

// companiesUpdateMonthlyHeadquarters is CompanyManager::updateMonthlyHeadquarters:
// every company's headquarters shows its performance (Company::
// updateMonthlyHeadquarters).
func (m *Map) companiesUpdateMonthlyHeadquarters(om *objects.ObjectManager) {
	for id := range m.Companies {
		if m.CompanyEmpty(id) || id >= len(m.CompanyFaces) {
			continue
		}
		// getHeadquarterPerformanceVariation: a step per 20%, at most 4.
		v := min(int(m.CompanyFaces[id].Performance)/200, 4)
		m.setHeadquartersVariation(id, uint8(v), om)
	}
}

// setHeadquartersVariation is Company::setHeadquartersVariation (0x0042F142):
// on each of the headquarters' four tiles, the building at its height, if
// it isn't that variation yet, is rebuilt as it — under construction again,
// its clearance the new variation's height.
func (m *Map) setHeadquartersVariation(id int, variation uint8, om *objects.ObjectManager) {
	c := &m.Companies[id]
	if c.HeadquartersX == -1 {
		return
	}
	tx, ty := int(c.HeadquartersX)>>5, int(c.HeadquartersY)>>5
	for dy := 0; dy <= 1; dy++ {
		for dx := 0; dx <= 1; dx++ {
			t := m.TileAt(tx+dx, ty+dy)
			if t == nil {
				continue
			}
			for _, r := range t.Order {
				if r.Kind != scenario.KindBuilding || int(r.Index) >= len(t.Buildings) {
					continue
				}
				b := &t.Buildings[r.Index]
				if b.BaseZ != c.HeadquartersZ {
					continue
				}
				if b.Variation == variation {
					break
				}
				x, y := int32(tx+dx)*32, int32(ty+dy)*32
				m.invalidateTile(x, y, int32(b.BaseZ)*smallZStep, int32(b.ClearZ)*smallZStep, ZoomEighth, defaultRadius)
				b.Variation, b.Age, b.IsConstructed, b.SectionProgress = variation, 0, false, 0
				total := 0
				if obj := om.GetBuildingObjectByIndex(int(b.ObjectID)); obj != nil {
					for _, part := range obj.GetBuildingParts(variation) {
						if int(part) < len(obj.PartHeights) {
							total += int(obj.PartHeights[part])
						}
					}
				}
				b.ClearZ = uint8(total/4) + b.BaseZ
				m.invalidateTile(x, y, int32(b.BaseZ)*smallZStep, int32(b.ClearZ)*smallZStep, ZoomEighth, defaultRadius)
				break
			}
		}
	}
}
