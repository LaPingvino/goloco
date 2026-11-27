// Package companyai holds goloco's own company AIs, written against
// pkg/sim's Company AI interface (sim.CompanyAI, sim.AIView). They are not
// ports of OpenLoco's CompanyAi.cpp (PORT-PLANS F: our own AI on a clean
// interface); they see and do only what a player can.
package companyai

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/sim"
)

// BusLinkName is the reference AI's registered name.
const BusLinkName = "buslink"

func init() { sim.RegisterAI(BusLinkName, func() sim.CompanyAI { return &BusLink{} }) }

// BusLink is the reference AI: the interface's end to end path, not a good
// player. It picks two nearby towns, joins their road networks if they
// aren't (a best-first search over the existing roads and new pieces that a
// dry-run createRoad accepts), builds a passenger stop in each on a straight
// piece of road, buys one bus, places it at the first stop, orders it
// between the two and starts it; then it waits. Every command is queried
// before it is applied; money it lacks it borrows with ChangeLoan, as a
// player does. One command per think.
type BusLink struct {
	state      int
	tried      map[[2]int]bool // town pairs given up on
	pairsTried int
	plan       busPlan
}

const (
	busPlanning = iota
	busRoad
	busBuilding
	busRunning
	busGivenUp
	busStandingStill // given up with something built: it stays as it is
)

// busPlan is the service planned: the towns, the road pieces still to build,
// the stop tiles (a straight road piece each), the stop type and the bus.
type busPlan struct {
	towns    [2]sim.AITown
	road0    uint8 // the road object new road is built of (without bit 7)
	road     []sim.RoadPlacement
	stops    [2]stopSite
	stopType uint8
	bus      uint16
	stations [2]int
	head     uint16
	step     int
}

type stopSite struct {
	X, Y         int // tiles
	BaseZ        uint8
	Rotation     uint8
	RoadObjectID uint8
}

func (b *BusLink) Name() string { return BusLinkName }

func (b *BusLink) Think(v *sim.AIView) {
	if b.tried == nil {
		b.tried = map[[2]int]bool{}
	}
	switch b.state {
	case busPlanning:
		b.planStep(v)
	case busRoad:
		b.roadStep(v)
	case busBuilding:
		b.buildStep(v)
	case busGivenUp:
		b.endCompany(v)
	}
}

// endCompany: having given up, the company ends (aiDestroy, as upstream's
// aiThinkState1 ends a company no thought ever started) — refused if it
// built anything (buslink doesn't sell up), and then it stands still.
func (b *BusLink) endCompany(v *sim.AIView) {
	if err := v.EndCompany(); err != nil {
		v.Logf("not ending the company: %v", err)
		b.state = busStandingStill
	}
}

// OnEvent notes the news about its own bus in the log; buslink acts on none.
func (b *BusLink) OnEvent(v *sim.AIView, e sim.AIEvent) {
	switch e.Type {
	case sim.MsgVehicleCrashed, sim.MsgVehicleSlipped, sim.MsgCantWaitForFullLoad,
		sim.EventVehicleBrokenDown, sim.EventVehicleRepaired:
		if e.Subjects[0] == b.plan.head {
			v.Logf("news: message type %d about bus %d", e.Type, b.plan.head)
		}
	}
}

// maxPairs bounds the town pairs one company looks at before giving up.
const maxPairs = 40

// planStep tries one town pair, the closest untried one first (ties broken
// by the AI's own PRNG): stops joined by road, or else the road to join them.
func (b *BusLink) planStep(v *sim.AIView) {
	towns := v.Towns()
	type pair struct {
		a, b int
		d    int
		tie  uint32
	}
	var pairs []pair
	for i := range towns {
		for j := i + 1; j < len(towns); j++ {
			d := abs(towns[i].X-towns[j].X) + abs(towns[i].Y-towns[j].Y)
			if d < 8 || d > 40 || b.tried[[2]int{towns[i].ID, towns[j].ID}] {
				continue
			}
			pairs = append(pairs, pair{i, j, d, v.Rng().RandNext()})
		}
	}
	if len(pairs) == 0 || b.pairsTried >= maxPairs {
		v.Logf("no town pair left to try: giving up")
		b.state = busGivenUp
		return
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].d != pairs[j].d {
			return pairs[i].d < pairs[j].d
		}
		return pairs[i].tie < pairs[j].tie
	})
	ta, tb := towns[pairs[0].a], towns[pairs[0].b]
	b.tried[[2]int{ta.ID, tb.ID}] = true
	b.pairsTried++
	road0, stopType, bus, ok := b.equipment(v)
	if !ok {
		v.Logf("no bus or passenger stop available: giving up")
		b.state = busGivenUp
		return
	}
	b.plan = busPlan{towns: [2]sim.AITown{ta, tb}, road0: road0, stopType: stopType, bus: bus}
	if b.planStops(v) {
		return
	}
	sitesA, sitesB := stopSites(v, ta, stopType), stopSites(v, tb, stopType)
	if len(sitesA) == 0 || len(sitesB) == 0 {
		v.Logf("%s - %s: no free straight road for a stop", ta.Name, tb.Name)
		return
	}
	from, _ := network(v, sitesA[0])
	_, goal := network(v, sitesB[0])
	road, n, ok := joinRoads(v, from, goal, tb, b.plan.road0)
	if !ok {
		v.Logf("%s - %s: no road found to join them (%d nodes searched)", ta.Name, tb.Name, n)
		return
	}
	var cost int64
	for _, r := range road {
		cost += int64(v.Query(sim.BuildRoad{RoadPlacement: r}).Cost)
	}
	stop := int64(v.Query(stopCommand(sitesA[0], stopType)).Cost)
	busCost := int64(v.Query(sim.BuyVehicle{Object: bus}).Cost)
	c := v.Company()
	if budget := c.Cash + int64(c.MaxLoan-c.Loan); cost+2*stop+busCost > budget {
		v.Logf("%s - %s: %d road pieces cost about %d, more than the %d it can spend", ta.Name, tb.Name, len(road), cost, budget)
		return
	}
	v.Logf("plan: join %s and %s with %d road pieces (about %d), %d nodes searched", ta.Name, tb.Name, len(road), cost, n)
	b.plan.road = road
	b.state = busRoad
}

// planStops finds a stop site in each town joined by road; true: the plan's
// stops are set and building starts.
func (b *BusLink) planStops(v *sim.AIView) bool {
	p := &b.plan
	sitesA := stopSites(v, p.towns[0], p.stopType)
	sitesB := map[[2]int]stopSite{}
	for _, s := range stopSites(v, p.towns[1], p.stopType) {
		sitesB[[2]int{s.X, s.Y}] = s
	}
	searched := 0
	for _, a := range sitesA {
		dest, n, ok := roadRoute(v, a, sitesB)
		searched += n
		if ok {
			p.stops = [2]stopSite{a, dest}
			p.step = 0
			b.state = busBuilding
			v.Logf("plan: %s (%d,%d) - %s (%d,%d) by road, %d pieces searched; bus object %d, stop type %d",
				p.towns[0].Name, a.X, a.Y, p.towns[1].Name, dest.X, dest.Y, n, p.bus, p.stopType)
			return true
		}
	}
	v.Logf("%s - %s: not joined by road (%d stop sites, %d pieces searched)", p.towns[0].Name, p.towns[1].Name, len(sitesA), searched)
	return false
}

// roadStep builds the plan's next road piece; with none left, the stops are
// planned on the joined network.
func (b *BusLink) roadStep(v *sim.AIView) {
	p := &b.plan
	if len(p.road) == 0 {
		if !b.planStops(v) {
			v.Logf("the new road doesn't join %s and %s: trying another pair", p.towns[0].Name, p.towns[1].Name)
			b.state = busPlanning
		}
		return
	}
	r := b.doPaid(v, sim.BuildRoad{RoadPlacement: p.road[0]})
	if !r.OK {
		v.Logf("road piece refused (error %d): trying another pair", r.Err)
		b.state = busPlanning
		return
	}
	p.road = p.road[1:]
}

// equipment is a road the company may build, a passenger stop type for it
// and a bus that carries the passengers and runs on it.
func (b *BusLink) equipment(v *sim.AIView) (road0, stopType uint8, bus uint16, ok bool) {
	pax := v.PassengerCargo()
	for _, road := range v.AvailableRoads() {
		var stop *sim.AIRoadStationObject
		for _, s := range v.RoadStations(road &^ (1 << 7)) {
			if s.Cargo == 1<<pax {
				stop = &s
				break
			}
		}
		if stop == nil {
			continue
		}
		for _, veh := range v.BuildableVehicles(1, road) { // 1: VehicleType::bus
			if veh.Cargo&(1<<pax) != 0 && veh.Power > 0 {
				return road &^ (1 << 7), stop.ID, veh.ID, true
			}
		}
	}
	return 0, 0, 0, false
}

// stopSites are the straight road pieces within 5 tiles of a town's centre
// with nothing on them where the stop type can be built (queried; the money
// is the budget's question).
func stopSites(v *sim.AIView, t sim.AITown, stopType uint8) []stopSite {
	var out []stopSite
	for dy := -5; dy <= 5; dy++ {
		for dx := -5; dx <= 5; dx++ {
			x, y := t.X+dx, t.Y+dy
			for _, r := range v.Roads(x, y) {
				if r.RoadID != 0 || r.SequenceIndex != 0 || r.HasStation || r.HasSignal || r.LevelCrossing || r.HasBridge {
					continue
				}
				s := stopSite{x, y, r.BaseZ, r.Rotation, r.RoadObjectID}
				if q := v.Query(stopCommand(s, stopType)); q.OK || q.Err == i18n.NotEnoughCashRequiresCurrency32 { // buildable, money aside
					out = append(out, s)
				}
			}
		}
	}
	return out
}

func stopCommand(s stopSite, stopType uint8) sim.BuildRoadStation {
	return sim.BuildRoadStation{RoadStationPlacement: sim.RoadStationPlacement{X: int32(s.X) * 32, Y: int32(s.Y) * 32,
		Z: int32(s.BaseZ) * 4, Rotation: s.Rotation, RoadID: 0, RoadObjectID: s.RoadObjectID, Type: stopType}}
}

// roadPos is where a road piece may start: a tile's position (world units,
// z included) and the rotation a vehicle arrives with (getRoadConnectionEnd's
// result).
type roadPos struct {
	x, y, z int32
	rot     uint8
}

// ends is where the road piece tad from (x, y, z) leads.
func ends(x, y, z int32, tad uint16) (roadPos, bool) {
	e, ok := sim.RoadConnectionEnd(x, y, z, tad&sim.TaDBasicMask)
	return roadPos{e.X, e.Y, e.Z, e.NextRotation}, ok
}

// network walks the road network from a stop site, both ways, as a vehicle
// could drive (getRoadConnections, the pieces' ends): every position
// reached, and the tiles road runs on from.
func network(v *sim.AIView, from stopSite) ([]roadPos, map[[2]int]bool) {
	x, y, z := int32(from.X)*32, int32(from.Y)*32, int32(from.BaseZ)*4
	var out []roadPos
	seen := map[roadPos]bool{}
	tiles := map[[2]int]bool{{from.X, from.Y}: true}
	var queue []roadPos
	for _, tad := range []uint16{uint16(from.Rotation), uint16(from.Rotation) | 1<<2} {
		if e, ok := ends(x, y, z, tad); ok && !seen[e] {
			seen[e] = true
			queue = append(queue, e)
		}
	}
	for len(queue) > 0 && len(out) < maxSearch {
		cur := queue[0]
		queue = queue[1:]
		out = append(out, cur)
		conns, _ := v.RoadConnections(cur.x, cur.y, cur.z, cur.rot, 0xFF) // any road every road vehicle may use
		if len(conns) > 0 {                                               // road goes on from this tile (a dead end's next tile has none)
			tiles[[2]int{int(cur.x / 32), int(cur.y / 32)}] = true
		}
		for _, c := range conns {
			if e, ok := ends(cur.x, cur.y, cur.z, c); ok && !seen[e] {
				seen[e] = true
				queue = append(queue, e)
			}
		}
	}
	return out, tiles
}

// maxSearch bounds a network walk (positions) and maxJoin a join search
// (positions expanded); maxNewPieces the length of a new road.
const (
	maxSearch    = 20000
	maxJoin      = 4000
	maxNewPieces = 80
)

// roadRoute is network's walk to one of the destination sites (at its
// height).
func roadRoute(v *sim.AIView, from stopSite, dests map[[2]int]stopSite) (stopSite, int, bool) {
	pos, _ := network(v, from)
	for _, p := range pos {
		if d, ok := dests[[2]int{int(p.x / 32), int(p.y / 32)}]; ok && int32(d.BaseZ)*4 == p.z {
			return d, len(pos), true
		}
	}
	return stopSite{}, len(pos), false
}

// joinRoads is a best-first search from the positions of one network to a
// tile of the other: along existing road (cost 1) or a new piece of the
// road object (cost 4) wherever a dry-run createRoad accepts it without
// removing buildings, until new road runs on into the other network; the
// distance to the goal town as the estimate. It returns the new pieces in
// building order.
func joinRoads(v *sim.AIView, from []roadPos, goal map[[2]int]bool, to sim.AITown, roadObj uint8) ([]sim.RoadPlacement, int, bool) {
	type node struct {
		p      roadPos
		g      int
		parent int
		piece  *sim.RoadPlacement
		pieces int
	}
	var nodes []node
	best := map[roadPos]int{}
	h := func(p roadPos) int { return 8 * (abs(int(p.x/32)-to.X) + abs(int(p.y/32)-to.Y)) } // greedy: twice a new piece per tile
	open := &posHeap{}
	push := func(n node) {
		if g, ok := best[n.p]; ok && g <= n.g {
			return
		}
		best[n.p] = n.g
		nodes = append(nodes, n)
		open.push(len(nodes)-1, n.g+h(n.p))
	}
	for _, p := range from {
		push(node{p: p, parent: -1})
	}
	for expanded := 0; open.len() > 0 && expanded < maxJoin; expanded++ {
		i := open.pop()
		n := nodes[i]
		if best[n.p] < n.g {
			continue
		}
		conns, _ := v.RoadConnections(n.p.x, n.p.y, n.p.z, n.p.rot, 0xFF)
		if n.pieces > 0 && len(conns) > 0 && goal[[2]int{int(n.p.x / 32), int(n.p.y / 32)}] { // new road runs on into the other network
			var road []sim.RoadPlacement
			for k := i; k >= 0; k = nodes[k].parent {
				if nodes[k].piece != nil {
					road = append([]sim.RoadPlacement{*nodes[k].piece}, road...)
				}
			}
			return road, expanded, true
		}
		for _, c := range conns {
			if e, ok := ends(n.p.x, n.p.y, n.p.z, c); ok {
				push(node{p: e, g: n.g + 1, parent: i, pieces: n.pieces})
			}
		}
		if n.pieces >= maxNewPieces {
			continue
		}
		for tad := range uint16(len(sim.RoadCoordinates)) {
			if tad&(1<<2) != 0 || sim.RoadCoordinates[tad].RotBegin != n.p.rot || tad>>3 == 9 { // forward pieces only; no turnaround
				continue
			}
			pl := sim.RoadPlacement{X: n.p.x, Y: n.p.y, Z: n.p.z, Rotation: uint8(tad & 3), RoadID: uint8(tad >> 3),
				Bridge: 0xFF, RoadObjectID: roadObj}
			if q := v.Query(sim.BuildRoad{RoadPlacement: pl}); !q.OK && q.Err != i18n.NotEnoughCashRequiresCurrency32 || q.RemovesBuildings { // money aside; no demolishing for the town's sake
				continue
			}
			if e, ok := ends(n.p.x, n.p.y, n.p.z, tad); ok {
				push(node{p: e, g: n.g + 4, parent: i, piece: &pl, pieces: n.pieces + 1})
			}
		}
	}
	return nil, len(nodes), false
}

// posHeap is a min-heap of node indexes by priority.
type posHeap struct{ idx, pri []int }

func (q *posHeap) len() int { return len(q.idx) }
func (q *posHeap) push(i, pri int) {
	q.idx, q.pri = append(q.idx, i), append(q.pri, pri)
	for k := len(q.idx) - 1; k > 0; {
		p := (k - 1) / 2
		if q.pri[p] <= q.pri[k] {
			break
		}
		q.idx[p], q.idx[k], q.pri[p], q.pri[k] = q.idx[k], q.idx[p], q.pri[k], q.pri[p]
		k = p
	}
}
func (q *posHeap) pop() int {
	top := q.idx[0]
	last := len(q.idx) - 1
	q.idx[0], q.pri[0] = q.idx[last], q.pri[last]
	q.idx, q.pri = q.idx[:last], q.pri[:last]
	for k := 0; ; {
		l, r, m := 2*k+1, 2*k+2, k
		if l < len(q.idx) && q.pri[l] < q.pri[m] {
			m = l
		}
		if r < len(q.idx) && q.pri[r] < q.pri[m] {
			m = r
		}
		if m == k {
			break
		}
		q.idx[m], q.idx[k], q.pri[m], q.pri[k] = q.idx[k], q.idx[m], q.pri[k], q.pri[m]
		k = m
	}
	return top
}

// buildStep issues the plan's next command: the two stops, the bus, its
// placement, its two orders, the start.
func (b *BusLink) buildStep(v *sim.AIView) {
	p := &b.plan
	var cmd sim.AICommand
	switch p.step {
	case 0, 1:
		cmd = stopCommand(p.stops[p.step], p.stopType)
	case 2:
		cmd = sim.BuyVehicle{Object: p.bus}
	case 3:
		s := p.stops[0]
		cmd = sim.PlaceVehicle{VehiclePlacement: sim.VehiclePlacement{X: int32(s.X) * 32, Y: int32(s.Y) * 32, Z: int32(s.BaseZ) * 4,
			TrackAndDirection: uint16(s.Rotation), Head: p.head}}
	case 4, 5:
		cmd = sim.AppendOrder{Head: p.head, Raw: sim.OrderStopAtRaw(p.stations[p.step-4])}
	case 6:
		cmd = sim.StartVehicle{Head: p.head}
	default:
		v.Logf("service running: bus %d between stations %d and %d", p.head, p.stations[0], p.stations[1])
		b.state = busRunning
		return
	}
	r := b.doPaid(v, cmd)
	if !r.OK && p.step == 3 { // the stop's other way round
		pv := cmd.(sim.PlaceVehicle)
		pv.TrackAndDirection |= 1 << 2
		r = b.doPaid(v, pv)
	}
	if !r.OK {
		v.Logf("step %d refused (error %d): giving up on this service", p.step, r.Err)
		b.state = busGivenUp
		return
	}
	switch p.step {
	case 0, 1:
		p.stations[p.step] = r.Station
	case 2:
		p.head = r.Vehicle
	}
	p.step++
}

// doPaid queries cmd and applies it; a refusal for want of cash first raises
// the loan (by the shortfall, in 1000s, up to the maximum) as a player would
// in the Company window.
func (b *BusLink) doPaid(v *sim.AIView, cmd sim.AICommand) sim.AIResult {
	q := v.Query(cmd)
	if !q.OK && q.Err == i18n.NotEnoughCashRequiresCurrency32 {
		c := v.Company()
		need := int64(q.ErrCost) - c.Cash
		loan := c.Loan + int32((need+999)/1000*1000)
		if loan <= c.MaxLoan && v.Do(sim.ChangeLoan{To: loan}).OK {
			q = v.Query(cmd)
		}
	}
	if !q.OK {
		return q
	}
	return v.Do(cmd)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// busLinkSave is BusLink's state as the save keeps it (sim.CompanyAI Save /
// Load): JSON, the town pairs tried sorted so one state is one encoding.
type busLinkSave struct {
	State, PairsTried int
	Tried             [][2]int
	Towns             [2]sim.AITown
	Road0             uint8
	Road              []sim.RoadPlacement
	Stops             [2]stopSite
	StopType          uint8
	Bus               uint16
	Stations          [2]int
	Head              uint16
	Step              int
}

func (b *BusLink) Save() []byte {
	p := b.plan
	s := busLinkSave{State: b.state, PairsTried: b.pairsTried, Towns: p.towns, Road0: p.road0, Road: p.road, Stops: p.stops,
		StopType: p.stopType, Bus: p.bus, Stations: p.stations, Head: p.head, Step: p.step}
	for k := range b.tried {
		s.Tried = append(s.Tried, k)
	}
	sort.Slice(s.Tried, func(i, j int) bool {
		if s.Tried[i][0] != s.Tried[j][0] {
			return s.Tried[i][0] < s.Tried[j][0]
		}
		return s.Tried[i][1] < s.Tried[j][1]
	})
	out, _ := json.Marshal(s)
	return out
}

func (b *BusLink) Load(data []byte) error {
	var s busLinkSave
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s.State < busPlanning || s.State > busStandingStill {
		return fmt.Errorf("buslink: no state %d", s.State)
	}
	b.state, b.pairsTried = s.State, s.PairsTried
	b.tried = map[[2]int]bool{}
	for _, k := range s.Tried {
		b.tried[k] = true
	}
	b.plan = busPlan{towns: s.Towns, road0: s.Road0, road: s.Road, stops: s.Stops, stopType: s.StopType, bus: s.Bus,
		stations: s.Stations, head: s.Head, step: s.Step}
	return nil
}
