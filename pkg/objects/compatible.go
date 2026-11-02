package objects

import "strings"

// CompatibleSets is Track/RoadObject::load's compatible loop: a bit per named
// object loaded in a track slot (tracks) or a road slot (roads).
func (m *ObjectManager) CompatibleSets(names []string) (tracks, roads uint16) {
	for _, n := range names {
		for slot, t := range m.TrackObjects {
			if t != nil && t.Header != nil && strings.EqualFold(strings.TrimRight(string(t.Header.Name[:]), " \x00"), n) {
				tracks |= 1 << slot
			}
		}
		for slot, r := range m.RoadObjects {
			if r != nil && r.Header != nil && strings.EqualFold(strings.TrimRight(string(r.Header.Name[:]), " \x00"), n) {
				roads |= 1 << slot
			}
		}
	}
	return
}
