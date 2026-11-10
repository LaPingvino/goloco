package sim

import "github.com/LaPingvino/goloco/pkg/objects"

// MapGenerator's misc buildings (MapGenerator.cpp generateMiscBuildings and
// its four generator functions): transmitters in mountains, pylon lines from
// power stations to towns, lighthouses by water, ruins on middling ground.

const industryFlagRequiresElectricityPylons = 1 << 20 // IndustryObjectFlags::requiresElectricityPylons

// drawableCoord is World::drawableCoord (world units).
func (m *Map) drawableCoord(v int32, size int) bool {
	return v >= tileSize && v < int32(size)*tileSize-tileSize-1
}

// generateMiscBuilding: averageNumberOnMap × ¾ plus up to half of it at
// random (uint8 arithmetic, as upstream) buildings, each up to 200 random
// tiles (2 .. size − 2) tried where the predicate holds, at the surface's
// highest corner, a random rotation and variation, black, built at once.
func (m *Map) generateMiscBuilding(c *Cmd, obj *objects.BuildingObject, id int, pred func(t *Tile, tx, ty int) bool) {
	random := uint8(m.Rng1.RandNextRange(0, int32(obj.AverageNumberOnMap/2)))
	static := obj.AverageNumberOnMap - obj.AverageNumberOnMap/4
	amount := random + static
	for i := uint8(0); i < amount; i++ {
		for attempts := 200; attempts > 0; attempts-- {
			x := int(m.Rng1.RandNextRange(2, int32(m.Height)-2)) // randNext(2, kMapRows − 2)
			y := int(m.Rng1.RandNextRange(2, int32(m.Width)-2))
			t := m.TileAt(x, y)
			if t == nil || !pred(t, x, y) {
				continue
			}
			z := int32(surfaceCornerHeight(t)) * smallZStep
			rot := uint8(m.Rng1.RandNextRange(0, 3))
			v := uint8(m.Rng1.RandNextRange(0, int32(obj.NumVariations)-1))
			a := BuildingPlacement{X: int32(x) * tileSize, Y: int32(y) * tileSize, Z: z, Rotation: rot, Type: uint8(id), Variation: v, BuildImmediately: true}
			if _, ok := m.CreateBuilding(c, CmdApply, a); ok {
				break
			}
		}
	}
}

// placeBuildingsAlongLine: from origin, a building every numTilesBetween
// tiles in the direction for the distance, stopping off the drawable map
// (failed placements don't matter). The last position.
func (m *Map) placeBuildingsAlongLine(c *Cmd, ox, oy int32, between uint8, id int, remaining int16, rot uint8) (int32, int32) {
	x, y := ox, oy
	for remaining -= int16(between); remaining > 0; remaining -= int16(between) {
		x += int32(RotationOffset[rot][0]) * int32(between)
		y += int32(RotationOffset[rot][1]) * int32(between)
		if !m.drawableCoord(x, m.Width) || !m.drawableCoord(y, m.Height) {
			break
		}
		t := m.TileAt(int(x)>>5, int(y)>>5)
		if t == nil {
			break
		}
		m.CreateBuilding(c, CmdApply, BuildingPlacement{X: x, Y: y, Z: int32(surfaceCornerHeight(t)) * smallZStep, Rotation: rot, Type: uint8(id), BuildImmediately: true})
	}
	return x, y
}

// generatePylons is generator function 1 (0x0042E893): from each industry
// that needs pylons, lines to up to three towns 800–2240 away (Manhattan):
// along the longer axis, then the shorter, a pylon every averageNumberOnMap tiles.
func (m *Map) generatePylons(c *Cmd, obj *objects.BuildingObject, id int, om *objects.ObjectManager) {
	for _, ind := range m.Industries {
		io := om.GetIndustryObjectByIndex(int(ind.ObjectID))
		if io == nil || io.Flags&industryFlagRequiresElectricityPylons == 0 {
			continue
		}
		left := 3
		for _, tw := range m.Towns {
			rotL, xd := uint8(2), int32(tw.X)-int32(ind.X)
			if xd < 0 {
				xd, rotL = -xd, 0
			}
			rotH, yd := uint8(1), int32(tw.Y)-int32(ind.Y)
			if yd < 0 {
				yd, rotH = -yd, 3
			}
			if d := xd + yd; d < 800 || d > 2240 {
				continue
			}
			if yd > xd {
				xd, yd = yd, xd
				rotL, rotH = rotH, rotL
			}
			ex, ey := m.placeBuildingsAlongLine(c, int32(ind.X), int32(ind.Y), obj.AverageNumberOnMap, id, int16(xd/tileSize), rotL)
			m.placeBuildingsAlongLine(c, ex, ey, obj.AverageNumberOnMap, id, int16(yd/tileSize), rotH)
			if left--; left == 0 {
				break
			}
		}
	}
}

// GenerateMiscBuildings is generateMiscBuildings (0x0042E6F2): every misc,
// non-HQ building object by its generator function — 0 at baseZ ≥ 100, 1
// the pylon lines, 2 with ≥ 50 water tiles round, 3 at baseZ 40–92.
func (m *Map) GenerateMiscBuildings(om *objects.ObjectManager) {
	if om == nil {
		return
	}
	c := &Cmd{Company: CompanyNeutral, Objects: om}
	for id, obj := range om.BuildingObjects {
		if obj == nil || obj.Flags&objects.BuildingFlagMiscBuilding == 0 || obj.Flags&objects.BuildingFlagIsHeadquarters != 0 {
			continue
		}
		switch obj.GeneratorFunction {
		case 0:
			m.generateMiscBuilding(c, obj, id, func(t *Tile, _, _ int) bool { return t.BaseZ >= 100 })
		case 1:
			m.generatePylons(c, obj, id, om)
		case 2:
			m.generateMiscBuilding(c, obj, id, func(_ *Tile, x, y int) bool { return m.countSurroundingWaterTiles(x, y) >= 50 })
		case 3:
			m.generateMiscBuilding(c, obj, id, func(t *Tile, _, _ int) bool { return t.BaseZ >= 40 && t.BaseZ <= 92 })
		}
	}
}
