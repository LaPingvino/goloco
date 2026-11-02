package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// The game date and the daily updates — OpenLoco.cpp dateTick, Date.cpp
// (updateDayCounter, calcDate), VehicleHead::updateDaily (via
// VehicleManager::updateDaily), Company::updateDailyPlayer's challenge part.
//
// PARTIAL: of dateTick only the day itself, VehicleManager::updateDaily and
// the player companies' challenge progress run; StationManager /
// IndustryManager / MessageManager daily updates, the snow line, the monthly,
// quarterly and yearly updates (monthsInChallenge included), and the rest of
// Company::updateDaily (emotions, AI logic, jail) are not ported.
// VehicleHead::updateDaily: the stopped-road-vehicle restart (a game
// command) and selling crashed AI vehicles are not ported.

// Date is Date (year, month 0–11, day of month 0-based, day of year).
type Date struct {
	Year       int
	Month, Day int
	DayOfYear  int
}

var monthLengths = [12]int{31, 29, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31} // getMonthDay's 366-day table

// CalcDate is calcDate: the date of a day number (Julian, from 1800).
func CalcDate(totalDays uint32) Date {
	const daysInYear, olympiad, feb29 = 365, 365*4 + 1, 31 + 28
	years := int((totalDays/olympiad)&0xFFFF) * 4
	day := int(totalDays % olympiad)
	if day > daysInYear {
		day -= daysInYear + 1
		for {
			years++
			day -= daysInYear
			if day < 0 {
				break
			}
		}
		day += daysInYear
		if day >= feb29 {
			day++
		}
	}
	d := Date{Year: 1800 + years, DayOfYear: day}
	for d.Month = 0; d.Month < 11 && day >= monthLengths[d.Month]; d.Month++ {
		day -= monthLengths[d.Month]
	}
	d.Day = day
	return d
}

// updateDayCounter: true when a new day starts (dayCounter wraps).
func (m *Map) updateDayCounter() bool {
	const increment = 682 // ~96 updates a day
	newDay := uint32(m.DayCounter)+increment > 0xFFFF
	if newDay {
		m.CurrentDay++
	}
	m.DayCounter += increment
	return newDay
}

// dateTick is dateTick's ported part (see PARTIAL).
func (m *Map) dateTick(om *objects.ObjectManager) {
	if !m.updateDayCounter() {
		return
	}
	m.vehiclesUpdateDaily(om)
	today := CalcDate(m.CurrentDay)
	m.CurrentYear, m.CurrentMonth, m.CurrentDayOfMonth = uint16(today.Year), uint8(today.Month), uint8(today.Day)
	m.companiesUpdateDaily(om)
}

// companiesUpdateDaily: updateDailyPlayer's challenge progress for the player
// company (not on the title screen).
func (m *Map) companiesUpdateDaily(om *objects.ObjectManager) {
	if m.TitleMode {
		return
	}
	id := int(m.ControllingCompany)
	if id >= len(m.Companies) || m.CompanyEmpty(id) {
		return
	}
	m.Companies[id].ChallengeProgress = m.NewChallengeProgress(id, om)
	// evaluateChallengeProgress (flags, messages, window): PARTIAL
}

const (
	reliabilityLossPerDay         = 4
	reliabilityLossPerDayObsolete = 10
)

// vehiclesUpdateDaily is VehicleManager::updateDaily → VehicleHead::updateDaily.
func (m *Map) vehiclesUpdateDaily(om *objects.ObjectManager) {
	trains, err := m.Trains()
	if err != nil {
		return
	}
	for _, tr := range trains {
		head := &m.Entities[tr.Head]
		if head.Status == StatusCrashed || head.Status == StatusStuck {
			if head.CrashedTimeout != 0xFF {
				head.CrashedTimeout++
			}
			// AI selling a long-crashed vehicle: PARTIAL
		}
		for _, car := range tr.Cars {
			for _, cc := range car {
				for _, i := range []int{cc.Front, cc.Back, cc.Body} {
					if c := &m.Entities[i].Cargo; c.Qty != 0 && c.NumDays != 0xFF {
						c.NumDays++
					}
				}
			}
		}
		for _, car := range tr.Cars {
			front := &m.Entities[car[0].Front]
			if front.Reliability == 0 {
				continue
			}
			if front.BreakdownFlags&BreakdownBrokenDn != 0 {
				front.BreakdownTimeout--
				if front.BreakdownTimeout == 0 {
					front.BreakdownFlags &^= BreakdownBrokenDn
					m.ApplyBreakdownToTrain(tr, om)
				}
			} else if front.TimeoutToBreakdown != 0xFFFF {
				if front.TimeoutToBreakdown > 0 {
					front.TimeoutToBreakdown--
				}
				if front.TimeoutToBreakdown == 0 {
					m.Sub4BA873(front)
					front.BreakdownFlags |= BreakdownPending
				}
			}
			loss := uint16(reliabilityLossPerDay)
			if obj := vehicleObject(om, front.ObjectID); obj != nil && obj.Obsolete <= m.CurrentYear {
				loss = reliabilityLossPerDayObsolete
			}
			front.Reliability = max(front.Reliability-loss, 100)
		}
		// recalculateTrainMinReliability
		minRel := uint16(0xFFFF)
		for _, car := range tr.Cars {
			r := m.Entities[car[0].Front].Reliability
			if r == 0 {
				r = 0xFFFF
			}
			minRel = min(minRel, r)
		}
		if minRel == 0xFFFF {
			m.Entities[tr.Veh2].Veh2Reliability = 0
		} else {
			m.Entities[tr.Veh2].Veh2Reliability = uint8(minRel / 256)
		}
	}
}

// Sub4BA873 is sub_4BA873: the days until the bogie's next breakdown, from
// its reliability (gPrng1).
func (m *Map) Sub4BA873(b *scenario.VehicleEntity) {
	b.TimeoutToBreakdown = 0xFFFF
	if b.Reliability == 0 {
		return
	}
	f := int32(b.Reliability / 256)
	f *= f
	f /= 16
	randVal := int32(m.Rng1.RandNextBound(65535)) * (f / 2) / 65536
	f -= f / 4
	f += randVal
	b.TimeoutToBreakdown = uint16(max(4, f))
}
