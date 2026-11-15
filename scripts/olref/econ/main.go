// econ prints the economy of saved games, one line per company and one per
// save for the world, so that the same game played on in OpenLoco and in
// goloco can be compared month by month (LOOP priority 1: the economy against
// OpenLoco). OpenLoco's monthly autosaves and goloco's own saves (gosave
// -ticks N, or its autosaves) are read alike, through pkg/scenario and the
// sim's load.
//
//	go run ./scripts/olref/econ a.SV5 b.SV5 …
//
// Each save gives a "world" line (date, towns' population, trains on the map)
// and a line per company in play: cash, loan, company value, vehicle profit,
// cargo delivered this month and in total, performance index, challenge
// progress. diff two runs' outputs to find the first month they part.
package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/sim"
)

var showTowns = flag.Bool("towns", false, "also list every town: population, capacity, buildings, build speed")

var company = flag.Int("company", -1, "also list this company's vehicles (status, place, speed, orders, profit)")

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		log.Fatal("usage: econ save.SV5 …")
	}
	for _, path := range flag.Args() {
		sc, err := scenario.LoadScenarioData(path)
		if err != nil {
			log.Fatalf("%s: %v", path, err)
		}
		s := &sim.State{}
		s.LoadScenario(sc)
		fmt.Println(describe(filepath.Base(path), &s.Map))
	}
}

// describe is one save's lines.
func describe(name string, m *sim.Map) string {
	pop := uint32(0)
	for _, t := range m.Towns {
		pop += t.Population
	}
	placed := 0
	if trains, err := m.Trains(); err == nil {
		for _, tr := range trains {
			if m.Entities[tr.Head].TileX != -1 {
				placed++
			}
		}
	}
	out := fmt.Sprintf("%s  %04d-%02d-%02d  world: towns %d pop %d, trains on the map %d",
		name, m.CurrentYear, int(m.CurrentMonth)+1, int(m.CurrentDayOfMonth)+1, len(m.Towns), pop, placed)
	for i := range m.Companies {
		if m.CompanyEmpty(i) {
			continue
		}
		c := &m.Companies[i]
		perf := int16(0)
		if i < len(m.CompanyFaces) {
			perf = m.CompanyFaces[i].Performance
		}
		total := uint64(c.CargoUnitsTotalDelivered)
		for _, n := range c.CargoUnitsDeliveredHist {
			total += uint64(n)
		}
		out += fmt.Sprintf("\n  company %d: cash %d loan %d value %d vehicleProfit %d delivered %d (total %d) perf %d progress %d",
			i, c.Cash, c.CurrentLoan, c.CompanyValueHistory[0], c.VehicleProfit, c.CargoUnitsTotalDelivered, total, perf, c.ChallengeProgress)
	}
	if *company >= 0 {
		out += trains(m, *company)
	}
	if *showTowns {
		for i, t := range m.Towns {
			out += fmt.Sprintf("\n    town %d: pop %d capacity %d buildings %d buildSpeed %d", i, t.Population, t.PopulationCapacity, t.NumBuildings, t.BuildSpeed)
		}
	}
	return out
}

// trains is a company's vehicles: head id, status, tile, station, the
// current order, Vehicle2's speed and motor state, breakdowns, this month's
// revenue and last income.
func trains(m *sim.Map, company int) string {
	out := ""
	list, err := m.Trains()
	if err != nil {
		return "\n  (trains: " + err.Error() + ")"
	}
	for _, tr := range list {
		h := &m.Entities[tr.Head]
		if int(h.Owner) != company {
			continue
		}
		v2 := &m.Entities[tr.Veh2]
		v1 := &m.Entities[tr.Veh1]
		out += fmt.Sprintf("\n    head %d: status %d tile %d,%d station %d order %d/%d speed %d motor %d breakdown 0x%02x revenue %d lastIncome %d",
			tr.Head, h.Status, h.TileX>>5, h.TileY>>5, h.StationID, h.CurrentOrder, h.SizeOfOrderTable, v2.CurrentSpeed, v2.MotorState,
			h.BreakdownFlags, v2.CurMonthRevenue, v1.LastIncome)
	}
	return out
}
