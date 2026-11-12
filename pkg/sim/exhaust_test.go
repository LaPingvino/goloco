package sim

import "testing"

// The title map's running trains give off exhaust (VehicleBody::animationUpdate):
// puffs of loaded steam objects, above the ground, which rise and end
// (Exhaust::update) — the count stays bounded and every one's slot comes back.
func TestTrainsGiveOffExhaust(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	if len(om.SteamObjects) == 0 {
		t.Skip("no steam objects")
	}
	invalidated := 0 // EntityBase::invalidateSprite: puffs redraw where they were and go
	m.InvalidateEntity = func(l, t, r, b int16) { invalidated++ }
	made, most := 0, 0
	seen := map[uint16]bool{}
	for range 600 {
		m.UpdateVehicles(om)
		n := 0
		for id, e := range m.Effects {
			if e.SubType != effectTypeExhaust {
				continue
			}
			n++
			if !seen[id] {
				seen[id] = true
				made++
				if steamObject(om, e.ObjectID) == nil {
					t.Fatalf("exhaust %d of no steam object (%#x)", id, e.ObjectID)
				}
				if land, _ := m.GetHeight(e.X, e.Y); e.Z <= int32(land)-64 {
					t.Errorf("exhaust %d far below the ground: z %d, land %d", id, e.Z, land)
				}
			}
		}
		most = max(most, n)
		for id := range seen {
			if m.Effects[id] == nil {
				delete(seen, id) // freed (the slot may be reused)
			}
		}
	}
	if invalidated < made {
		t.Errorf("%d puffs made, only %d sprite invalidations", made, invalidated)
	}
	if made == 0 {
		t.Fatal("600 ticks of the title map's trains made no exhaust")
	}
	if most > 2000 {
		t.Errorf("%d puffs alive at once: they don't end", most)
	}
	t.Logf("%d puffs made, at most %d alive", made, most)
}
