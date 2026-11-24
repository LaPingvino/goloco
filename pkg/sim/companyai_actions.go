package sim

import (
	"fmt"

	"github.com/LaPingvino/goloco/pkg/i18n"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// An AI's actions: game commands a player has, with explicit arguments, each
// through DoCommandFor under the AI's company at a player's price
// (AI-INTERFACE.md §1, §6). The set is closed: an AICommand is one of the
// types below. Query runs the query pass alone (the funding check included,
// so it says whether the company could pay); Do applies.

// AICommand is one game command an AI may issue.
type AICommand interface {
	fmt.Stringer
	// expenditure is the command's setExpenditureType.
	expenditure() uint8
	// run is the command function for one pass (flags with or without
	// CmdApply); it may fill in what the command made.
	run(v *AIView, c *Cmd, flags uint8, r *AIResult) (int32, bool)
}

// AIResult is a command's outcome as the AI sees it.
type AIResult struct {
	OK bool
	// Cost is what it costs (negative: a refund); refused for want of cash
	// alone (not_enough_cash_requires_currency32), the price it would have.
	Cost int32
	// Err is the refusal's StringId (pkg/i18n; 0 none) and ErrCost its
	// {CURRENCY32} — the error window's second line a player would see.
	Err     int
	ErrCost int32
	// The dry run's side information (getLegacyReturnState after createTrack
	// / createRoad, AI-INTERFACE.md §5 P2-P4).
	NeedsBridge      bool
	BridgeHeight     uint8
	RemovesBuildings bool
	OverlaysRoad     bool
	AltRoadObject    uint8 // 0xFF none
	// What the command made (applied): a vehicle head, a station (-1 none).
	Vehicle uint16
	Station int
}

// Query runs c's query pass for the AI's company and changes nothing.
func (v *AIView) Query(c AICommand) AIResult { return v.issue(c, 0) }

// Do runs c for the AI's company: checked, funded, applied and paid.
func (v *AIView) Do(c AICommand) AIResult { return v.issue(c, CmdApply) }

func (v *AIView) issue(cmd AICommand, flags uint8) AIResult {
	m := v.m
	c := &Cmd{Company: v.id, Objects: v.om, Players: func(id uint8) bool { return m.IsPlayerCompany(int(id)) }}
	r := AIResult{Vehicle: EntityIDNull, Station: -1, AltRoadObject: 0xFF}
	cost, ok := m.DoCommandFor(c, flags, cmd.expenditure(), func(f uint8) (int32, bool) { return cmd.run(v, c, f, &r) })
	r.OK, r.Cost, r.Err, r.ErrCost = ok, cost, c.ErrorText, c.ErrorCost
	if !ok {
		r.Cost = 0
		if r.Err == i18n.NotEnoughCashRequiresCurrency32 {
			r.Cost = r.ErrCost // refused for the money alone: the price
		}
	}
	r.NeedsBridge = c.Flags73&(1<<0) != 0
	r.RemovesBuildings = c.Flags73&(1<<4) != 0
	r.OverlaysRoad = c.Flags73&(1<<5) != 0
	r.BridgeHeight = c.BridgeHeight
	if _, road := cmd.(BuildRoad); road {
		r.AltRoadObject = c.AltRoadObject // only createRoad sets alternateRoadObjectId
	}
	if flags&CmdApply == 0 {
		v.st.queries++ // dry runs are counted, not logged: a search makes thousands
		return r
	}
	v.st.record(AIDecision{Tick: m.ScenarioTicks, Day: m.CurrentDay, Text: cmd.String(), Command: true,
		OK: r.OK, Cost: r.Cost, Err: r.Err})
	return r
}

func (v *AIView) g1() G1Raw {
	if v.om == nil {
		return nil
	}
	g1, _ := v.om.G1File.(G1Raw)
	return g1
}

// BuildTrack is createTrack (CreateTrack.cpp).
type BuildTrack struct{ TrackPlacement }

func (b BuildTrack) String() string {
	return fmt.Sprintf("build track %d rot %d at %d,%d,%d (object %d)", b.TrackID, b.Rotation, b.X/32, b.Y/32, b.Z, b.TrackObjectID)
}
func (BuildTrack) expenditure() uint8 { return ExpenditureConstruction }
func (b BuildTrack) run(v *AIView, c *Cmd, f uint8, _ *AIResult) (int32, bool) {
	return v.m.CreateTrack(c, f, b.TrackPlacement)
}

// BuildRoad is createRoad (CreateRoad.cpp).
type BuildRoad struct{ RoadPlacement }

func (b BuildRoad) String() string {
	return fmt.Sprintf("build road %d rot %d at %d,%d,%d (object %d)", b.RoadID, b.Rotation, b.X/32, b.Y/32, b.Z, b.RoadObjectID)
}
func (BuildRoad) expenditure() uint8 { return ExpenditureConstruction }
func (b BuildRoad) run(v *AIView, c *Cmd, f uint8, _ *AIResult) (int32, bool) {
	return v.m.CreateRoad(c, f, b.RoadPlacement)
}

// RemoveRoad is removeRoad (RemoveRoad.cpp).
type RemoveRoad struct{ RoadRemoval }

func (b RemoveRoad) String() string   { return fmt.Sprintf("remove road %+v", b.RoadRemoval) }
func (RemoveRoad) expenditure() uint8 { return ExpenditureConstruction }
func (b RemoveRoad) run(v *AIView, c *Cmd, f uint8, _ *AIResult) (int32, bool) {
	return v.m.RemoveRoad(c, f, b.RoadRemoval, v.g1())
}

// BuildRoadStation is createRoadStation (CreateRoadStation.cpp); applied,
// the result names the station the stop belongs to.
type BuildRoadStation struct{ RoadStationPlacement }

func (b BuildRoadStation) String() string {
	return fmt.Sprintf("build road station type %d on road %d rot %d at %d,%d,%d", b.Type, b.RoadID, b.Rotation, b.X/32, b.Y/32, b.Z)
}
func (BuildRoadStation) expenditure() uint8 { return ExpenditureConstruction }
func (b BuildRoadStation) run(v *AIView, c *Cmd, f uint8, r *AIResult) (int32, bool) {
	cost, ok := v.m.CreateRoadStation(c, f, b.RoadStationPlacement, v.g1())
	if ok && f&CmdApply != 0 {
		if t := v.m.TileAt(int(b.X/32), int(b.Y/32)); t != nil {
			for _, ref := range t.Order {
				if ref.Kind == scenario.KindStation {
					if st := t.Stations[ref.Index]; st.Owner == c.Company && st.StationType == 1 && int32(st.BaseZ)*4 == b.Z {
						r.Station = int(st.StationID)
					}
				}
			}
		}
	}
	return cost, ok
}

// BuyVehicle is createVehicle with no head (a new vehicle, CreateVehicle.cpp);
// applied, the result names its head.
type BuyVehicle struct{ Object uint16 }

func (b BuyVehicle) String() string   { return fmt.Sprintf("buy vehicle object %d", b.Object) }
func (BuyVehicle) expenditure() uint8 { return ExpenditureVehiclePurchases }
func (b BuyVehicle) run(v *AIView, c *Cmd, f uint8, r *AIResult) (int32, bool) {
	return v.m.CreateVehicle(c, f, b.Object, false, &r.Vehicle)
}

// AddCar is createVehicle with a head: another car on it.
type AddCar struct {
	Head   uint16
	Object uint16
}

func (b AddCar) String() string {
	return fmt.Sprintf("add car object %d to vehicle %d", b.Object, b.Head)
}
func (AddCar) expenditure() uint8 { return ExpenditureVehiclePurchases }
func (b AddCar) run(v *AIView, c *Cmd, f uint8, _ *AIResult) (int32, bool) {
	return v.m.AddCarToVehicle(c, f, b.Object, b.Head, false)
}

// PlaceVehicle is vehiclePlace (rail or road, VehiclePlace.cpp).
type PlaceVehicle struct{ VehiclePlacement }

func (b PlaceVehicle) String() string {
	return fmt.Sprintf("place vehicle %d at %d,%d,%d tad %#x", b.Head, b.X/32, b.Y/32, b.Z, b.TrackAndDirection)
}
func (PlaceVehicle) expenditure() uint8 { return ExpenditureTrainRunningCosts }
func (b PlaceVehicle) run(v *AIView, c *Cmd, f uint8, _ *AIResult) (int32, bool) {
	return v.m.VehiclePlace(c, f, b.VehiclePlacement)
}

// AppendOrder is vehicleOrderInsert at the end of a vehicle's orders; Raw is
// the order's bytes (OrderStopAtRaw, OrderRouteWaypointRaw …).
type AppendOrder struct {
	Head uint16
	Raw  []byte
}

func (b AppendOrder) String() string {
	return fmt.Sprintf("append order %x to vehicle %d", b.Raw, b.Head)
}
func (AppendOrder) expenditure() uint8 { return ExpenditureTrainRunningCosts }
func (b AppendOrder) run(v *AIView, c *Cmd, f uint8, _ *AIResult) (int32, bool) {
	e := v.m.entity(b.Head)
	if e == nil || e.SizeOfOrderTable == 0 {
		return 0, false
	}
	return v.m.VehicleOrderInsert(c, f, b.Head, e.SizeOfOrderTable-1, b.Raw)
}

// StartVehicle and StopVehicle are vehicleChangeRunningMode.
type StartVehicle struct{ Head uint16 }

func (b StartVehicle) String() string   { return fmt.Sprintf("start vehicle %d", b.Head) }
func (StartVehicle) expenditure() uint8 { return ExpenditureTrainRunningCosts }
func (b StartVehicle) run(v *AIView, c *Cmd, f uint8, _ *AIResult) (int32, bool) {
	return v.m.VehicleChangeRunningMode(c, f, b.Head, RunningModeStart)
}

type StopVehicle struct{ Head uint16 }

func (b StopVehicle) String() string   { return fmt.Sprintf("stop vehicle %d", b.Head) }
func (StopVehicle) expenditure() uint8 { return ExpenditureTrainRunningCosts }
func (b StopVehicle) run(v *AIView, c *Cmd, f uint8, _ *AIResult) (int32, bool) {
	return v.m.VehicleChangeRunningMode(c, f, b.Head, RunningModeStop)
}

// SellVehicle is vehicleSell (VehicleSell.cpp).
type SellVehicle struct{ Head uint16 }

func (b SellVehicle) String() string   { return fmt.Sprintf("sell vehicle %d", b.Head) }
func (SellVehicle) expenditure() uint8 { return ExpenditureVehicleDisposals }
func (b SellVehicle) run(v *AIView, c *Cmd, f uint8, _ *AIResult) (int32, bool) {
	return v.m.VehicleSell(c, f, b.Head)
}

// ChangeLoan is changeLoan (ChangeLoan.cpp): the loan set to To.
type ChangeLoan struct{ To int32 }

func (b ChangeLoan) String() string   { return fmt.Sprintf("change loan to %d", b.To) }
func (ChangeLoan) expenditure() uint8 { return ExpenditureLoanInterest }
func (b ChangeLoan) run(v *AIView, c *Cmd, f uint8, _ *AIResult) (int32, bool) {
	return v.m.ChangeLoan(c, f, b.To)
}
