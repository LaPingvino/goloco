package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/coords"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// updateLastJourneyAverageSpeed: a journey faster than the mode's record
// takes it (speed, date, owner); a slower one leaves it; air counts × 36
// into the air record.
func TestSpeedRecord(t *testing.T) {
	m := &Map{Entities: make([]scenario.VehicleEntity, 2), CurrentDay: 54000, ScenarioTicks: 1000}
	m.CompanyRecords.Company = [3]uint8{0xFF, 0xFF, 0xFF}
	head, v2 := &m.Entities[0], &m.Entities[1]
	tr := Train{Head: 0, Veh2: 1}
	journey := func(mode uint8, x int16, owner uint8) {
		head.Mode, head.Owner = mode, owner
		head.BreakdownFlags |= BreakdownJourneyStarted
		head.JourneyStartX, head.JourneyStartY, head.JourneyStartTicks = 0, 0, 0
		v2.PosX, v2.PosY, v2.MaxSpeed = coords.WorldUnits(x), 0, 200
		m.updateLastJourneyAverageSpeed(tr)
	}
	journey(0, 3000, 2) // 3000 × 21 / 1000 = 63
	if r := m.CompanyRecords; r.Speed[0] != 63 || r.Company[0] != 2 || r.Date[0] != 54000 {
		t.Fatalf("land record %+v", r)
	}
	journey(0, 2000, 3) // 42: slower, kept
	if r := m.CompanyRecords; r.Speed[0] != 63 || r.Company[0] != 2 {
		t.Fatalf("slower journey changed the record %+v", r)
	}
	journey(2, 2000, 1) // air: 2000 × 36 / 1000 = 72
	if r := m.CompanyRecords; r.Speed[1] != 72 || r.Company[1] != 1 || r.Speed[0] != 63 {
		t.Fatalf("air record %+v", r)
	}
}
