package sim

import "testing"

// TestAnimationsLoadAndRun: the title map's animations load, and updating
// them keeps the building ones (types 5 / 6) and invalidates their tiles.
func TestAnimationsLoadAndRun(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	if len(m.Animations) == 0 {
		t.Fatal("no animations loaded")
	}
	types := map[uint8]int{}
	for _, a := range m.Animations {
		types[a.Type]++
	}
	inval := 0
	m.InvalidateTile = func(x, y, zMin, zMax int32, zoom uint8, radius int32) { inval++ }
	before := len(m.Animations)
	for range 200 {
		m.ScenarioTicks++
		m.UpdateAnimations(om)
	}
	t.Logf("%d animations by type %v; after 200 updates %d, %d tile invalidations", before, types, len(m.Animations), inval)
	if types[5]+types[6] > 0 && inval == 0 {
		t.Error("building animations invalidated nothing")
	}
}
