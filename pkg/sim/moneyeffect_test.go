package sim

import "testing"

// A window-currency money effect: made from the nullMoney list (an id at or
// past maxNormalEntities), not on the title; it rises by the wiggle table
// (3 a frame for the first ten) drifting by the rotation's nudge on even
// frames, and after 160 frames it is freed — its slot back in nullMoney.
func TestMoneyEffectLifetime(t *testing.T) {
	m := &Map{EntityLists: resetEntityLists()}
	m.TitleMode = true
	if _, ok := m.CreateMoneyEffect(0, 0, 0, 0, 100); ok {
		t.Fatal("a money effect on the title")
	}
	m.TitleMode = false
	free := m.EntityLists.Counts[EntityListNullMoney]
	m.SpendMoneyEffect(320, 320, 40, 2, 1234)
	if len(m.MoneyEffects) != 1 || m.EntityLists.Counts[EntityListNullMoney] != free-1 {
		t.Fatalf("effects %d, nullMoney %d (was %d)", len(m.MoneyEffects), m.EntityLists.Counts[EntityListNullMoney], free)
	}
	var id uint16
	var e *MoneyEffect
	for k, v := range m.MoneyEffects {
		id, e = k, v
	}
	if id < maxNormalEntities || e.Z != 50 || e.SubType != effectTypeWindowCurrency || e.Amount != -1234 { // spendMoneyEffect creates it with −amount
		t.Fatalf("id %d effect %+v", id, *e)
	}
	m.UpdateEffects(0)
	if e.X != 321 || e.Y != 319 || e.Z != 53 || e.Frame != 1 { // frame 0: nudge (1,-1), rise 3
		t.Fatalf("after one update %+v", *e)
	}
	for i := 0; i < 159; i++ {
		m.UpdateEffects(0)
	}
	if len(m.MoneyEffects) != 1 || e.Frame != 160 {
		t.Fatalf("frame %d, %d effects", e.Frame, len(m.MoneyEffects))
	}
	m.UpdateEffects(0) // frame 160: freed
	if len(m.MoneyEffects) != 0 || m.EntityLists.Counts[EntityListNullMoney] != free {
		t.Fatalf("not freed: %d effects, nullMoney %d", len(m.MoneyEffects), m.EntityLists.Counts[EntityListNullMoney])
	}
	t.Logf("rose to z %d over its life", e.Z)
}
