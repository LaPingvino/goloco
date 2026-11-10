package sim

import "testing"

func TestCheckForCollisions(t *testing.T) {
	m, _ := loadDataMap(t, "title.dat")
	trains, _ := m.Trains()
	// Two rail trains of one owner, both on the map.
	var a, b Train
	found := false
	for i, x := range trains {
		hx := m.Entities[x.Head]
		if hx.Mode != 0 || hx.TileX == -1 || len(x.Cars) == 0 {
			continue
		}
		for _, y := range trains[i+1:] {
			hy := m.Entities[y.Head]
			if hy.Mode == 0 && hy.Owner == hx.Owner && hy.TileX != -1 && len(y.Cars) > 0 {
				a, b, found = x, y, true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Skip("no two rail trains of one owner")
	}
	bogie := a.Cars[0][0].Front
	e := m.Entities[bogie]
	if got := m.CheckForCollisions(bogie, int32(e.PosX), int32(e.PosY), int32(e.PosZ)); got != 0xFFFF {
		t.Errorf("a running train's own bogie collides with %d", got)
	}
	// The other train's first bogie put next to it: a collision with it.
	ob := b.Cars[0][0].Front
	m.Entities[ob].PosX, m.Entities[ob].PosY, m.Entities[ob].PosZ = e.PosX+4, e.PosY, e.PosZ
	if got := m.CheckForCollisions(bogie, int32(e.PosX), int32(e.PosY), int32(e.PosZ)); got != m.Entities[ob].ID {
		t.Errorf("collides with %d, want the other train's bogie %d", got, m.Entities[ob].ID)
	}
	// Another owner's part there doesn't count.
	m.Entities[ob].Owner++
	if got := m.CheckForCollisions(bogie, int32(e.PosX), int32(e.PosY), int32(e.PosZ)); got == m.Entities[ob].ID {
		t.Error("another company's vehicle collided")
	}
}
