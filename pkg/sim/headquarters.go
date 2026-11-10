package sim

// GameCommands/Company/BuildCompanyHeadquarters.cpp and
// RemoveCompanyHeadquarters.cpp.

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
// company has no HQ. The cost. PARTIAL: the company window invalidation, the
// demolish sound.
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
			if !c.companyCompatible(uint8(i)) {
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
	return total, true
}
