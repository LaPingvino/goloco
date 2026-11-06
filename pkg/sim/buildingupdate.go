package sim

import (
	"github.com/LaPingvino/goloco/pkg/core"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Town buildings in the tile loop (item CP): BuildingElement::update
// (Map/BuildingElement.cpp 0x0042DF8B) — a building under construction
// grows a section at a time; a built one ages; an old one may be pulled
// down; and each update a built building may produce its two cargoes (a
// town's passengers and mail) and hand them to the stations around it.
// PARTIAL: the viewport invalidations.

const gameStateFlagUnk2 = 1 << 1 // GameStateFlags::unk2: buildings produce half

// buildingAt is the building element at (tx, ty) of this object and base
// height — applyToMultiTile's search on a large building's other tiles.
func (m *Map) buildingAt(tx, ty int, baseZ uint8) *scenario.BuildingElement {
	t := m.TileAt(tx, ty)
	if t == nil {
		return nil
	}
	for i := range t.Buildings {
		if t.Buildings[i].BaseZ == baseZ {
			return &t.Buildings[i]
		}
	}
	return nil
}

// applyToMultiTile calls fn for the building's element on each of its tiles
// (one, or four for a large building): the first is el itself, the others
// the building element at the same base height there.
func (m *Map) applyToMultiTile(el *scenario.BuildingElement, x, y int32, large bool, fn func(*scenario.BuildingElement, int32, int32)) {
	for i, off := range buildingTileOffsets(large) {
		px, py := x+off[0], y+off[1]
		b := el
		if i != 0 {
			if b = m.buildingAt(int(px)>>5, int(py)>>5, el.BaseZ); b == nil {
				continue
			}
		}
		fn(b, px, py)
	}
}

// updateBuilding is BuildingElement::update for the building element el at
// world (x, y); false when it was removed (the tile's elements changed).
func (m *Map) updateBuilding(el *scenario.BuildingElement, ghost bool, x, y int32, om *objects.ObjectManager) bool {
	if el.SequenceIndex != 0 || ghost { // only from a multi-tile building's first tile
		return true
	}
	if om == nil {
		return true
	}
	obj := om.GetBuildingObjectByIndex(int(el.ObjectID))
	if obj == nil {
		return true
	}
	large := obj.Flags&objects.BuildingFlagLargeTile != 0
	if !el.IsConstructed {
		progress, age, built := el.SectionProgress, el.Age, false
		if el.SectionProgress != 7 {
			progress++
		} else {
			parts := obj.GetBuildingParts(el.Variation)
			if int(el.Age)+1 != len(parts) {
				progress = 0
				age++
			} else {
				total := 3
				for _, part := range parts {
					if int(part) < len(obj.PartHeights) {
						total += int(obj.PartHeights[part])
					}
				}
				clearZ := el.BaseZ + uint8(total/4)
				m.applyToMultiTile(el, x, y, large, func(b *scenario.BuildingElement, _, _ int32) { b.ClearZ = clearZ })
				if obj.NumElevatorSequences != 0 {
					m.CreateAnimation(5, x, y, el.BaseZ)
				}
				m.updateTownInfo(x, y, uint32(obj.ProducedQuantity[0]), 0, 0, 0, &Cmd{Company: CompanyNeutral, Objects: om})
				progress, age, built = 0, 0, true
			}
		}
		m.applyToMultiTile(el, x, y, large, func(b *scenario.BuildingElement, _, _ int32) {
			b.IsConstructed, b.SectionProgress, b.Age = built, progress, age
		})
	}
	if el.IsMisc {
		return true
	}
	ti, _, ok := m.ClosestTownAndDensity(x, y)
	if !ok {
		return true
	}
	tw := &m.Towns[ti]
	if el.IsConstructed {
		el.SectionProgress = (el.SectionProgress + 1) & 7 // setUnk5u(unk5u() + 1): three bits
		if el.SectionProgress == 0 && el.Age != 63 {
			age := el.Age + 1
			m.applyToMultiTile(el, x, y, large, func(b *scenario.BuildingElement, _, _ int32) { b.Age = age })
		}
	}
	if el.IsConstructed && el.Age >= 40 {
		p := core.NewPrngSeed(tw.Prng[0], tw.Prng[1])
		roll := p.RandNextBound(0xFFFF)
		tw.Prng = [2]uint32{p.Srand0(), p.Srand1()}
		if roll <= 16 {
			m.RemoveBuilding(&Cmd{Company: CompanyNeutral, Objects: om}, x, y, int32(el.BaseZ)*4, CmdApply)
			return false
		}
	}
	if el.IsConstructed {
		p := core.NewPrngSeed(tw.Prng[0], tw.Prng[1])
		r := uint16(p.RandNextBound(0xFFFF))
		tw.Prng = [2]uint32{p.Srand0(), p.Srand1()}
		rolls := [2]uint8{uint8(r), uint8(r >> 8)}
		size := 1
		if large {
			size = 2
		}
		cargo := om.ProducedCargoIDs(obj)
		for i := 0; i < 2; i++ {
			if rolls[i] >= obj.ProducedQuantity[i] {
				continue
			}
			amount := int(rolls[i])/4 + 1
			if m.GameStateFlags&gameStateFlagUnk2 != 0 {
				amount = (amount + 1) / 2
			}
			tw.Var19C[i][0] += uint16(amount)
			got := m.DeliverCargoToNearbyStations(cargo[i], uint8(amount), x, y, size, size, om)
			tw.Var19C[i][1] += got & 0xFF
		}
	}
	return true
}
