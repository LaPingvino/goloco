package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// Money effects (item ME): upstream's money popups are entities
// (Effects/MoneyEffect.cpp): made by spendMoneyEffect when a company is paid
// or pays, moved by update, freed at the end of their life; painted through
// the paint session's string list (the scene's part). goloco keeps their
// fields by entity id (its Entities slice holds vehicles); the id itself
// comes from EntityManager's nullMoney list, as upstream.

const (
	effectTypeRedGreenCurrency = 1 // EffectType::redGreenCurrency
	effectTypeWindowCurrency   = 2 // EffectType::windowCurrency

	moneyEffectLifetime         = 160 // MoneyEffect::kLifetime (windowCurrency)
	moneyEffectRedGreenLifetime = 55  // kRedGreenLifetime
)

// MoneyEffect is the money effect entity's fields.
type MoneyEffect struct {
	X, Y, Z      int32 // position
	Amount       int32
	Company      uint8 // var_2E
	SubType      uint8 // EffectType
	Frame        uint16
	NumMovements uint16
	MoveDelay    uint8
	Wiggle       uint8
}

// moneyWiggle is _wiggleAmounts (0x004FADD0): the drift per viewport
// rotation.
var moneyWiggle = [4][2]int32{{1, -1}, {1, 1}, {-1, 1}, {-1, -1}}

// moneyWiggleZ is _wiggleZAmounts (0x004FAD21): the rise per frame.
var moneyWiggleZ = func() [moneyEffectLifetime]int8 {
	var z [moneyEffectLifetime]int8
	head := [12]int8{3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 2, 1}
	copy(z[:], head[:])
	for i := 12; i < moneyEffectLifetime; i++ { // then 0, 0, 0, 1 repeating from frame 12
		if (i-12)%4 == 3 {
			z[i] = 1
		}
	}
	return z
}()

// createEntityMoney is EntityManager::createEntityMoney: the head of the
// nullMoney list, moved to misc.
func (l *EntityLists) createEntityMoney() (uint16, bool) {
	if l.Counts[EntityListNullMoney] == 0 {
		return 0, false
	}
	id := l.Heads[EntityListNullMoney]
	l.moveEntityToList(id, EntityListMisc)
	l.Links[id].BaseType = scenario.EntityBaseTypeEffect
	return id, true
}

// CreateMoneyEffect is MoneyEffect::create (0x00440A74): none on the title;
// a windowCurrency effect of the amount for the company at (x, y, z).
// (offsetX — minus half the formatted amount's width — is the painter's.)
func (m *Map) CreateMoneyEffect(x, y, z int32, company uint8, amount int32) (uint16, bool) {
	if m.TitleMode || m.EntityLists == nil {
		return 0, false
	}
	id, ok := m.EntityLists.createEntityMoney()
	if !ok {
		return 0, false
	}
	if m.MoneyEffects == nil {
		m.MoneyEffects = map[uint16]*MoneyEffect{}
	}
	m.MoneyEffects[id] = &MoneyEffect{X: x, Y: y, Z: z, Amount: amount, Company: company, SubType: effectTypeWindowCurrency}
	return id, true
}

// SpendMoneyEffect is CompanyManager::spendMoneyEffect (0x0046DC9F): the
// effect 10 above the location. PARTIAL: an unset location (Location::null:
// the main viewport's centre) is skipped — the sim has no viewport.
func (m *Map) SpendMoneyEffect(x, y, z int32, company uint8, amount int32) {
	if x == -32768 {
		return
	}
	m.CreateMoneyEffect(x, y, z+10, company, amount)
}

// updateMoneyEffect is MoneyEffect::update (0x0044063E) for one effect;
// rotation is the main viewport's (WindowManager::getCurrentRotation).
// false: it ended and was freed.
func (m *Map) updateMoneyEffect(id uint16, e *MoneyEffect, rotation int) bool {
	if e.SubType == effectTypeWindowCurrency {
		if e.Wiggle == 0 {
			e.Wiggle = 21
		} else {
			e.Wiggle--
		}
		if e.Frame >= moneyEffectLifetime {
			m.freeMoneyEffect(id)
			return false
		}
		d := moneyWiggle[rotation&3]
		if e.Frame&1 != 0 {
			d = [2]int32{}
		}
		e.X, e.Y, e.Z = e.X+d[0], e.Y+d[1], e.Z+int32(moneyWiggleZ[e.Frame])
		e.Frame++
		return true
	}
	if e.Wiggle == 22 {
		e.Wiggle = 0
	} else {
		e.Wiggle++
	}
	e.MoveDelay++
	if e.MoveDelay < 2 {
		return true
	}
	e.MoveDelay = 0
	d := moneyWiggle[rotation&3]
	e.X, e.Y, e.Z = e.X+d[0], e.Y+d[1], e.Z+e.Z // moveTo(position + {nudge, position.z}): upstream doubles z
	e.NumMovements++
	if e.NumMovements >= moneyEffectRedGreenLifetime {
		m.freeMoneyEffect(id)
		return false
	}
	return true
}

func (m *Map) freeMoneyEffect(id uint16) {
	delete(m.MoneyEffects, id)
	if m.EntityLists != nil {
		m.EntityLists.freeEntity(id)
	}
}

// UpdateMoneyEffects is EffectsManager::update (0x004402F4) for the money
// effects: the misc list walked from its head, each entity's next taken as it
// is reached (ListIterator), so an effect freed by its update doesn't end the
// walk. rotation is the main viewport's.
func (m *Map) UpdateMoneyEffects(rotation int) {
	if len(m.MoneyEffects) == 0 || m.EntityLists == nil {
		return
	}
	l := m.EntityLists
	for id := l.Heads[EntityListMisc]; id != entityIDNull; {
		next := l.Links[id].Next
		if e := m.MoneyEffects[id]; e != nil {
			m.updateMoneyEffect(id, e, rotation)
		}
		id = next
	}
}
