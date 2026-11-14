package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// The roads in the tile loop — Map/RoadElement.cpp RoadElement::tick: a
// company that isn't a player's gives up a road of a town road type (any
// road type compatible) that has gone unused, its own road removal run as
// that company. "Unused": neither in-use flag (unk7_40, set by a vehicle on
// it or the road's building; unk7_80, last year's — TileManager::updateYearly
// moves it on) — for these road types the flags share the mods bits.

const (
	roadUnk7_40 = 1 << 0 // Mods bit 0: hasUnk7_40
	roadUnk7_80 = 1 << 1 // Mods bit 1: hasUnk7_80
)

// roadTickBlocked is the tick's checks of one road element: it stays when
// the piece's other tiles, a level crossing (or this road's in-use flags),
// a ghost, an AI allocation or a road stop are there.
func roadTickBlocked(e, self *scenario.RoadElement) bool {
	return e.SequenceIndex != 0 ||
		self.Unk7_10 || self.HasLevelCrossing || self.Mods&(roadUnk7_40|roadUnk7_80) != 0 || // upstream checks `this` here
		e.Ghost || e.AiAllocated || e.HasStationElement
}

// updateRoadElement is RoadElement::tick for the road at t.Order[k] on the
// tile at world (x, y): false when it was removed.
func (s *State) updateRoadElement(t *Tile, k int, x, y int32, om *objects.ObjectManager) bool {
	if int(t.Order[k].Index) >= len(t.Roads) {
		return true
	}
	e := t.Roads[t.Order[k].Index]
	notOurs := func(r *scenario.RoadElement) bool { // neutral, a player's, or not a town road type
		return r.Owner == CompanyNeutral || s.IsPlayerCompany(int(r.Owner)) || s.RoadAnyTypeCompatible&(1<<r.RoadObjectID) == 0
	}
	if notOurs(&e) || roadTickBlocked(&e, &e) {
		return true
	}
	// "Verify there are no other conflicting tile elements on the current tile"
	for _, r := range t.Order {
		if r.Kind != scenario.KindRoad || int(r.Index) >= len(t.Roads) {
			continue
		}
		o := &t.Roads[r.Index]
		if o.BaseZ != e.BaseZ || notOurs(o) {
			continue
		}
		if roadTickBlocked(o, &e) {
			return true
		}
	}
	c := &Cmd{Company: e.Owner, Objects: om}
	a := RoadRemoval{X: x, Y: y, Z: int32(e.BaseZ) * smallZStep, Rotation: e.Rotation, RoadID: e.RoadID, Index: e.SequenceIndex, RoadObjectID: e.RoadObjectID}
	s.DoCommandFor(c, CmdApply, ExpenditureConstruction, func(f uint8) (int32, bool) { return s.RemoveRoad(c, f, a, nil) })
	return false
}

// tilesUpdateYearly is TileManager::updateYearly: each town-type road (not
// a ghost or AI-allocated) has this year's in-use flag moved to last year's.
func (m *Map) tilesUpdateYearly() {
	for y := range m.Tiles {
		for x := range m.Tiles[y] {
			t := &m.Tiles[y][x]
			for i := range t.Roads {
				r := &t.Roads[i]
				if r.AiAllocated || r.Ghost || m.RoadAnyTypeCompatible&(1<<r.RoadObjectID) == 0 {
					continue
				}
				was := r.Mods & roadUnk7_40
				r.Mods &^= roadUnk7_40 | roadUnk7_80 // setUnk7_80(hasUnk7_40), setUnk7_40(false)
				if was != 0 {
					r.Mods |= roadUnk7_80
				}
			}
		}
	}
}
