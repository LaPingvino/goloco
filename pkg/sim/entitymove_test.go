package sim

import "testing"

// TestMoveToSpriteBounds: moveTo's sprite bounds (EntityBase::moveTo) equal the
// ones the original saved, for every entity with a sprite on the title map.
func TestMoveToSpriteBounds(t *testing.T) {
	m, _ := loadDataMap(t, "title.dat")
	n, bad := 0, 0
	for i := range m.Entities {
		e := &m.Entities[i]
		if e.PosX == locationNull || e.SpriteLeft == locationNull || e.SpriteWidth == 0 {
			continue
		}
		n++
		l, tp, r, b := e.SpriteLeft, e.SpriteTop, e.SpriteRight, e.SpriteBottom
		m.moveTo(e, int32(e.PosX), int32(e.PosY), int32(e.PosZ))
		if l != e.SpriteLeft || tp != e.SpriteTop || r != e.SpriteRight || b != e.SpriteBottom {
			bad++
			if bad < 4 {
				t.Logf("type %d: saved %d %d %d %d, moveTo %d %d %d %d", e.EntityType, l, tp, r, b, e.SpriteLeft, e.SpriteTop, e.SpriteRight, e.SpriteBottom)
			}
		}
	}
	if bad > 0 || n == 0 {
		t.Errorf("%d entities with sprites, %d whose saved bounds differ from moveTo's", n, bad)
	}
}
