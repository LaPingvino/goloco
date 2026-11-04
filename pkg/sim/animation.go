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
	case 0:
		return m.updateSignalAnimation(a, om)
	case 1:
		return m.updateLevelCrossingAnimation(a, om)
	case 2:
		return false // upstream: kept, nothing done
	case 7:
		return m.updateStationAnimation(a, om, 2)
	case 8:
		return m.updateStationAnimation(a, om, 3)
	case 3:
		return m.updateIndustryAnimation1(a, om)
	case 4:
		return m.updateIndustryAnimation2(a, om)
	case 5:
		return m.updateBuildingAnimation1(a, om)
	case 6:
		return m.updateBuildingAnimation2(a, om)
	}
	return false // upstream asserts: no other types
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

// industryObjectOf is industry->getObject() for an industry element.
func (m *Map) industryObjectOf(el *scenario.IndustryElement, om *objects.ObjectManager) *objects.IndustryObject {
	if int(el.IndustryID) >= len(m.IndustryObjectIDs) {
		return nil
	}
	return om.GetIndustryObjectByIndex(int(m.IndustryObjectIDs[el.IndustryID]))
}

// updateIndustryAnimation1 (IndustryElement.cpp): the first industry element
// at the animation's base height; with no animated part (numFrames > 1) the
// animation ends, else the tile is invalidated (quarter zoom) whenever the
// fastest such part's speed mask comes round.
func (m *Map) updateIndustryAnimation1(a scenario.Animation, om *objects.ObjectManager) bool {
	t := m.TileAt(int(a.X)/tileSize, int(a.Y)/tileSize)
	if t == nil {
		return true
	}
	for _, r := range t.Order {
		if r.Kind != scenario.KindIndustry {
			continue
		}
		el := &t.Industries[r.Index]
		if el.BaseZ != a.BaseZ {
			continue
		}
		obj := m.industryObjectOf(el, om)
		if obj == nil || int(el.BuildingType) >= len(obj.Variations) {
			return true
		}
		has, speed := false, uint8(0xFF)
		for _, part := range obj.Variations[el.BuildingType] {
			if int(part) < len(obj.PartAnimations) && obj.PartAnimations[part].NumFrames > 1 {
				has = true
				speed = min(speed, obj.PartAnimations[part].AnimationSpeed&^(1<<7))
			}
		}
		if !has {
			return true
		}
		if m.ScenarioTicks&(1<<speed-1) == 0 {
			m.invalidateTile(int32(a.X), int32(a.Y), int32(el.BaseZ)*4, int32(el.ClearZ)*4, ZoomQuarter, 0)
		}
		return false
	}
	return true
}

// industryMultiTile is IndustryElement.cpp's applyToMultiTile: the element and
// those at the same base height on the other tiles of a 2x2 building.
func (m *Map) industryMultiTile(el0 *scenario.IndustryElement, x, y int32, large bool, f func(el *scenario.IndustryElement, x, y int32)) {
	for i, off := range buildingTileOffsets(large) {
		px, py := x+off[0], y+off[1]
		el := el0
		if i != 0 {
			el = nil
			if t := m.TileAt(int(px)/tileSize, int(py)/tileSize); t != nil {
				for _, r := range t.Order {
					if r.Kind == scenario.KindIndustry && t.Industries[r.Index].BaseZ == el0.BaseZ {
						el = &t.Industries[r.Index]
						break
					}
				}
			}
		}
		if el != nil {
			f(el, px, py)
		}
	}
}

// updateIndustryAnimation2 (0x00456EEB): an industry element at the base
// height with var_6_003F bit 5 (a sequence running); at the first part with
// no frames of its own, the sequence's length (a power of 2) and the part's
// speed decide: with bit 4 set, back at step 0 → invalidate and clear bit 5
// (the animation ends), else invalidate on the speed mask; without bit 4, at
// step 1 → invalidate and set bit 4.
func (m *Map) updateIndustryAnimation2(a scenario.Animation, om *objects.ObjectManager) bool {
	t := m.TileAt(int(a.X)/tileSize, int(a.Y)/tileSize)
	if t == nil {
		return true
	}
	inval := func(el *scenario.IndustryElement, x, y int32) {
		m.invalidateTile(x, y, int32(el.BaseZ)*4, int32(el.ClearZ)*4, ZoomQuarter, 0)
	}
	for _, r := range t.Order {
		if r.Kind != scenario.KindIndustry {
			continue
		}
		el := &t.Industries[r.Index]
		if el.BaseZ != a.BaseZ || el.Var6Low&(1<<5) == 0 {
			continue
		}
		obj := m.industryObjectOf(el, om)
		if obj == nil || int(el.BuildingType) >= len(obj.Variations) {
			continue
		}
		animLength := uint32(len(obj.AnimationSequences[el.Var6Low&3]))
		large := obj.BuildingSizeFlags&(1<<el.BuildingType) != 0
		for _, part := range obj.Variations[el.BuildingType] {
			if int(part) >= len(obj.PartAnimations) || obj.PartAnimations[part].NumFrames != 0 {
				continue
			}
			speed := obj.PartAnimations[part].AnimationSpeed &^ (1 << 7)
			mask := animLength - 1
			if el.Var6Low&(1<<4) != 0 {
				if mask&(m.ScenarioTicks>>speed) == 0 {
					m.industryMultiTile(el, int32(a.X), int32(a.Y), large, func(e *scenario.IndustryElement, x, y int32) {
						inval(e, x, y)
						e.Var6Low &^= 1 << 5
					})
					return true
				}
				if m.ScenarioTicks&(1<<speed-1) == 0 {
					m.industryMultiTile(el, int32(a.X), int32(a.Y), large, inval)
				}
				return false
			}
			if mask&(m.ScenarioTicks>>speed) == 1 {
				m.industryMultiTile(el, int32(a.X), int32(a.Y), large, func(e *scenario.IndustryElement, x, y int32) {
					inval(e, x, y)
					e.Var6Low |= 1 << 4
				})
			}
			return false
		}
	}
	return true
}

// updateSignalAnimation (SignalElement.cpp): each signal element at the base
// height steps both sides' frames toward unk4 × 3 (at most the object's last
// frame) on the object's speed mask; the tile invalidated when one stepped;
// done when no side has a frame left to reach.
func (m *Map) updateSignalAnimation(a scenario.Animation, om *objects.ObjectManager) bool {
	t := m.TileAt(int(a.X)/tileSize, int(a.Y)/tileSize)
	if t == nil {
		return true
	}
	has, inval := false, false
	for _, r := range t.Order {
		if r.Kind != scenario.KindSignal {
			continue
		}
		el := &t.Signals[r.Index]
		if el.BaseZ != a.BaseZ {
			continue
		}
		for i := range el.Sides {
			side := &el.Sides[i]
			if !side.HasSignal() {
				continue
			}
			obj := om.GetTrainSignalObjectByIndex(int(side.ObjectID()))
			if obj == nil || obj.NumFrames == 0 {
				continue
			}
			target := min(side.Unk4()*3, obj.NumFrames-1)
			if side.Frame() == target {
				continue
			}
			has = true
			if m.ScenarioTicks&uint32(obj.AnimSpeed) != 0 {
				continue
			}
			f := side.Frame() + 1
			if side.Frame() >= target {
				f = side.Frame() - 1
				if obj.Flags&(1<<2) != 0 { // TrainSignalObjectFlags::unk2
					f = min(f, 1)
				}
			}
			side.SetFrame(f)
			inval = true
		}
	}
	if inval {
		m.invalidateTile(int32(a.X), int32(a.Y), int32(a.BaseZ)*4, int32(a.BaseZ)*4+32, ZoomHalf, 0)
	}
	return !has
}

// updateLevelCrossingAnimation (RoadElement.cpp): each road element with a
// level crossing at the base height, off its object's tick mask (var_0A),
// steps its frame — closing (unk7_10) up to closedFrames then 15, opening
// down from 15 through closedFrames to 0 — and invalidates the tile; done when
// none is still moving (upstream's quirk kept: an open crossing at frame 0
// doesn't count as animating).
func (m *Map) updateLevelCrossingAnimation(a scenario.Animation, om *objects.ObjectManager) bool {
	t := m.TileAt(int(a.X)/tileSize, int(a.Y)/tileSize)
	if t == nil {
		return true
	}
	has, inval := false, false
	for _, r := range t.Order {
		if r.Kind != scenario.KindRoad {
			continue
		}
		el := &t.Roads[r.Index]
		if el.BaseZ != a.BaseZ || !el.HasLevelCrossing {
			continue
		}
		obj := om.GetLevelCrossingObjectByIndex(int(el.LevelCrossingObjectID))
		if obj == nil {
			continue
		}
		if m.ScenarioTicks&uint32(obj.Var0A) != 0 {
			has = true
			continue
		}
		inval = true
		f := el.AnimFrame
		if el.Unk7_10 {
			if f != 15 {
				f++
				if f > obj.ClosedFrames {
					f = 15
				}
			}
			has = true
		} else if f != 0 {
			f--
			if f == 14 {
				f = obj.ClosedFrames
			}
			has = true
		}
		el.AnimFrame = f
	}
	if inval {
		m.invalidateTile(int32(a.X), int32(a.Y), int32(a.BaseZ)*4, int32(a.BaseZ)*4+32, ZoomHalf, 0)
	}
	return !has
}

// updateStationAnimation (StationElement.cpp, airports and docks): the first
// station element of the type at the base height; without a part of > 1
// frames the animation ends, else the tile invalidated on the fastest such
// part's speed mask (quarter zoom).
func (m *Map) updateStationAnimation(a scenario.Animation, om *objects.ObjectManager, stationType uint8) bool {
	t := m.TileAt(int(a.X)/tileSize, int(a.Y)/tileSize)
	if t == nil {
		return true
	}
	for _, r := range t.Order {
		if r.Kind != scenario.KindStation {
			continue
		}
		el := &t.Stations[r.Index]
		if el.BaseZ != a.BaseZ || el.StationType != stationType {
			continue
		}
		var parts []uint8
		var anims []objects.BuildingPartAnimation
		if stationType == 2 {
			if o := om.GetAirportObjectByIndex(int(el.ObjectID)); o != nil && int(el.BuildingType) < len(o.Variations) {
				parts, anims = o.Variations[el.BuildingType], o.PartAnimations
			}
		} else if o := om.GetDockObjectByIndex(int(el.ObjectID)); o != nil && int(el.BuildingType) < len(o.Variations) {
			parts, anims = o.Variations[el.BuildingType], o.PartAnimations
		}
		has, speed := false, uint8(0xFF)
		for _, p := range parts {
			if int(p) < len(anims) && anims[p].NumFrames > 1 {
				has = true
				speed = min(speed, anims[p].AnimationSpeed&^(1<<7))
			}
		}
		if !has {
			return true
		}
		if m.ScenarioTicks&(1<<speed-1) == 0 {
			m.invalidateTile(int32(a.X), int32(a.Y), int32(el.BaseZ)*4, int32(el.ClearZ)*4, ZoomQuarter, 0)
		}
		return false
	}
	return true
}
