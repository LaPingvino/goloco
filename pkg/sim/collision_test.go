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

// A collision (VehicleBogie::collision): both trains crashed, the crash news
// for the first player company's own, nothing left to sell, motors stopped,
// the bogie that ran into it and its body exploding (var_5A bit 31, the speed
// / 8), the first train's bogie flagged unk_5; on later ticks a crashed train's
// parts slide to a stop. And the motion's per-step check reports the hit.
func TestTrainCrash(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	trains, _ := m.Trains()
	var a, b Train
	found := false
	for i, x := range trains {
		hx := m.Entities[x.Head]
		if hx.Mode != 0 || hx.TileX == -1 || len(x.Cars) == 0 {
			continue
		}
		for _, y := range trains[i+1:] {
			if hy := m.Entities[y.Head]; hy.Mode == 0 && hy.Owner == hx.Owner && hy.TileX != -1 && len(y.Cars) > 0 {
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
	m.PlayerCompanies[0] = m.Entities[a.Head].Owner
	m.Entities[a.Veh2].CurrentSpeed = 40 << 16
	bogie, hit := a.Cars[0][0].Front, b.Cars[0][0].Front
	// the motion: a step that lands on another vehicle reports it
	c := m.component(bogie)
	c.Collide = func(x, y, z int32) uint16 { return m.Entities[hit].ID }
	if r := c.UpdateMotion(0x368A*3, nil); r.Flags&MotionCrashed == 0 || r.CollidedEntity != m.Entities[hit].ID {
		t.Fatalf("the motion didn't report the collision: %+v", r)
	}
	before := len(m.Messages)
	misc := m.EntityLists.Counts[EntityListMisc]
	m.BogieCollision(a, bogie, m.Entities[hit].ID)
	// explodeComponent: a cloud per exploding part, crash particles in its colours
	clouds, particles := 0, 0
	for _, e := range m.Effects {
		switch e.SubType {
		case effectTypeExplosionCloud:
			clouds++
		case effectTypeVehicleCrashParticle:
			particles++
			if e.SpriteBase > 4 || e.TimeToLive < 140 || e.Frame/256 >= 12 {
				t.Errorf("crash particle out of range: %+v", *e)
			}
		}
	}
	if clouds < 2 || particles == 0 || int(m.EntityLists.Counts[EntityListMisc]) != int(misc)+len(m.Effects) {
		t.Errorf("explosion: %d clouds, %d particles, misc %d → %d", clouds, particles, misc, m.EntityLists.Counts[EntityListMisc])
	}
	ha, hb := m.Entities[a.Head], m.Entities[b.Head]
	if ha.Status != StatusCrashed || hb.Status != StatusCrashed || ha.TotalRefundCost != 0 || hb.TotalRefundCost != 0 {
		t.Errorf("not both crashed and worthless: %d %d", ha.Status, hb.Status)
	}
	if m.Entities[a.Veh2].MotorState != MotorStopped || m.Entities[b.Veh2].MotorState != MotorStopped {
		t.Error("a crashed train's motor runs on")
	}
	be, body := m.Entities[bogie], m.Entities[a.Cars[0][0].Body]
	if be.Var5A&(1<<31) == 0 || be.Var5A&^(1<<31) != (40<<16)>>3 || be.VehicleFlags&vehicleFlagUnk5 == 0 || body.Var5A&(1<<31) == 0 {
		t.Errorf("the colliding bogie / body aren't exploding: bogie %#x flags %#x body %#x", be.Var5A, be.VehicleFlags, body.Var5A)
	}
	if len(m.Messages) == before {
		t.Error("no crash news for the player's train")
	}
	// later ticks (landCrashedUpdate): every bogie slows by 1/64 a tick to a
	// stop, the exploding one slides off its track, and the bodies explode
	// with their bogies
	x0, y0 := m.Entities[bogie].PosX, m.Entities[bogie].PosY
	for range 400 {
		m.updateTrain(a, om)
		m.UpdateEffects(om, 0)
	}
	if len(m.Effects) != 0 || m.EntityLists.Counts[EntityListMisc] != misc {
		t.Errorf("%d effects outlived 400 ticks (misc %d, was %d)", len(m.Effects), m.EntityLists.Counts[EntityListMisc], misc)
	}
	for _, car := range a.Cars {
		for _, cc := range car {
			for _, i := range []int{cc.Front, cc.Back} {
				if v := m.Entities[i].Var5A &^ (1 << 31); v != 0 {
					t.Errorf("bogie %d still at speed %#x after 400 ticks", i, v)
				}
			}
		}
	}
	if e := m.Entities[bogie]; e.PosX == x0 && e.PosY == y0 {
		t.Error("the exploding bogie didn't slide")
	}
	if m.Entities[a.Cars[0][0].Body].VehicleFlags&vehicleFlagUnk5 == 0 {
		t.Error("the body of an exploded bogie didn't explode")
	}
}
