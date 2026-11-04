package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// AnimationManager (Map/AnimationManager.cpp): the map's animations — a tile
// position, a base height and a type — each updated every tick by its type's
// function, which invalidates what it animates and says when to drop it.

const maxAnimations = 8192 // Limits::kMaxAnimations

// CreateAnimation is createAnimation: added unless the list is full or the same
// one is there already.
func (m *Map) CreateAnimation(typ uint8, x, y int32, baseZ uint8) {
	if len(m.Animations) >= maxAnimations {
		return
	}
	for _, a := range m.Animations {
		if a.Type == typ && int32(a.X) == x && int32(a.Y) == y && a.BaseZ == baseZ {
			return
		}
	}
	m.Animations = append(m.Animations, scenario.Animation{BaseZ: baseZ, Type: typ, X: int16(x), Y: int16(y)})
}

// UpdateAnimations is AnimationManager::update: every animation updated, the
// finished ones removed (in order).
func (m *Map) UpdateAnimations(om *objects.ObjectManager) {
	keep := m.Animations[:0]
	for _, a := range m.Animations {
		if !m.updateAnimation(a, om) {
			keep = append(keep, a)
		}
	}
	m.Animations = keep
}

// updateAnimation is callUpdateFunction: true = remove it.
func (m *Map) updateAnimation(a scenario.Animation, om *objects.ObjectManager) bool {
	switch a.Type {
	case 5:
		return m.updateBuildingAnimation1(a, om)
	case 6:
		return m.updateBuildingAnimation2(a, om)
	}
	// PARTIAL: 0 signal, 1 level crossing, 3 / 4 industry, 7 airport, 8 dock
	// (and 2, which upstream keeps without doing anything) are kept unstepped.
	return false
}

// animationBuilding is the building search of updateBuildingAnimation1 / 2,
// quirk included: each element in turn (nil when not a building) until one
// at the animation's base height — with none, the tile's last element's.
func (m *Map) animationBuilding(a scenario.Animation) *scenario.BuildingElement {
	t := m.TileAt(int(a.X)/tileSize, int(a.Y)/tileSize)
	if t == nil {
		return nil
	}
	var b *scenario.BuildingElement
	for _, r := range t.Order {
		b = nil
		if r.Kind != scenario.KindBuilding {
			continue
		}
		b = &t.Buildings[r.Index]
		if b.BaseZ == a.BaseZ {
			break
		}
	}
	return b
}

// updateBuildingAnimation1 (elevators): every other tick, the building's tiles
// invalidated (at half zoom).
func (m *Map) updateBuildingAnimation1(a scenario.Animation, om *objects.ObjectManager) bool {
	b := m.animationBuilding(a)
	if b == nil || b.SequenceIndex != 0 {
		return true
	}
	obj := om.GetBuildingObjectByIndex(int(b.ObjectID))
	if obj == nil || obj.NumElevatorSequences == 0 {
		return true
	}
	if m.ScenarioTicks&1 != 0 {
		return false
	}
	base := b.BaseZ
	for i, off := range buildingTileOffsets(obj.HasFlags(objects.BuildingFlagLargeTile)) { // applyToMultiTile
		x, y := int32(a.X)+off[0], int32(a.Y)+off[1]
		el := b
		if i != 0 {
			el = nil
			if t := m.TileAt(int(x)/tileSize, int(y)/tileSize); t != nil {
				for _, r := range t.Order {
					if r.Kind == scenario.KindBuilding && t.Buildings[r.Index].BaseZ == base {
						el = &t.Buildings[r.Index]
						break
					}
				}
			}
		}
		if el != nil {
			m.invalidateTile(x, y, int32(el.BaseZ)*4, int32(el.ClearZ)*4, ZoomHalf, 0)
		}
	}
	return false
}

// updateBuildingAnimation2 (animated parts): whenever the slowest animated
// part's speed mask comes round, the tile invalidated (at quarter zoom).
func (m *Map) updateBuildingAnimation2(a scenario.Animation, om *objects.ObjectManager) bool {
	b := m.animationBuilding(a)
	if b == nil {
		return true
	}
	obj := om.GetBuildingObjectByIndex(int(b.ObjectID))
	if obj == nil || int(b.Variation) >= len(obj.VariationParts) {
		return true
	}
	slowest, frames := uint8(0xFF), uint8(0)
	for _, part := range obj.VariationParts[b.Variation] {
		if int(part) >= len(obj.PartAnimations) {
			continue
		}
		pa := obj.PartAnimations[part]
		if pa.NumFrames == 0 {
			continue
		}
		frames |= pa.NumFrames
		if pa.NumFrames == 1 {
			continue
		}
		slowest = min(slowest, pa.AnimationSpeed&^(1<<7))
	}
	if frames <= 1 {
		return true
	}
	if m.ScenarioTicks&(1<<slowest-1) == 0 {
		m.invalidateTile(int32(a.X), int32(a.Y), int32(b.BaseZ)*4, int32(b.ClearZ)*4, ZoomQuarter, 0)
	}
	return false
}
