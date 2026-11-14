package sim

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// saveAndLoad writes the state as an S5 saved game and loads it back into
// a fresh state, as a player's save and load would.
func saveAndLoad(t *testing.T, s *State) (*State, []byte) {
	t.Helper()
	s.PrepareSave()
	sc := s.ExportScenario()
	var buf bytes.Buffer
	if err := sc.SavedGameFile(make([]byte, scenario.SaveDetailsSize), s.Raw.RequiredObjects).Write(&buf); err != nil {
		t.Fatal(err)
	}
	sc2, err := scenario.ParseScenario(buf.Bytes(), "roundtrip.SV5")
	if err != nil {
		t.Fatal(err)
	}
	s2 := &State{}
	s2.LoadScenario(sc2)
	s2.LoadWaves(sc2.Waves)
	return s2, buf.Bytes()
}

// stateDiff lists where two states differ (the first max), leaving out what
// a save doesn't carry by design: the file bytes kept (Raw), the callbacks,
// the vehicle list's snapshot (re-read from the entities). The tiles are
// compared as their elements' bytes (the slices' order and a ref's copy of
// the last flag are bookkeeping; the bytes are the state).
func stateDiff(a, b *State, max int) []string {
	var out []string
	ta, tb := scenario.EncodeTileElements(a.ExportScenario().Tiles), scenario.EncodeTileElements(b.ExportScenario().Tiles)
	if !bytes.Equal(ta, tb) {
		n := 0
		for i := 0; i+8 <= min(len(ta), len(tb)) && n < 10; i += 8 {
			if !bytes.Equal(ta[i:i+8], tb[i:i+8]) {
				out = append(out, fmt.Sprintf("tile element %d: % x vs % x", i/8, ta[i:i+8], tb[i:i+8]))
				n++
			}
		}
		out = append(out, fmt.Sprintf("tile elements: %d vs %d bytes", len(ta), len(tb)))
	}
	var walk func(path string, x, y reflect.Value)
	walk = func(path string, x, y reflect.Value) {
		if len(out) >= max {
			return
		}
		switch x.Kind() {
		case reflect.Func:
			return
		case reflect.Pointer, reflect.Interface:
			if x.IsNil() || y.IsNil() {
				if x.IsNil() != y.IsNil() {
					out = append(out, path+": nil vs non-nil")
				}
				return
			}
			walk(path, x.Elem(), y.Elem())
		case reflect.Struct:
			for i := 0; i < x.NumField(); i++ {
				f := x.Type().Field(i)
				if f.Name == "Raw" || f.Name == "VehicleListData" || f.Name == "Tiles" || notSaved[f.Name] || f.Name == "Cons" || strings.Contains(f.Name, "Ghost") && f.Type.Kind() == reflect.Pointer {
					continue
				}
				walk(path+"."+f.Name, x.Field(i), y.Field(i))
			}
		case reflect.Slice, reflect.Array:
			if x.Kind() == reflect.Slice && x.Len() != y.Len() {
				out = append(out, fmt.Sprintf("%s: len %d vs %d", path, x.Len(), y.Len()))
				return
			}
			for i := 0; i < x.Len(); i++ {
				walk(fmt.Sprintf("%s[%d]", path, i), x.Index(i), y.Index(i))
			}
		case reflect.Map:
			if x.Len() != y.Len() {
				out = append(out, fmt.Sprintf("%s: %d entries vs %d", path, x.Len(), y.Len()))
				return
			}
			for _, k := range x.MapKeys() {
				yv := y.MapIndex(k)
				if !yv.IsValid() {
					out = append(out, fmt.Sprintf("%s[%v]: missing", path, k))
					continue
				}
				walk(fmt.Sprintf("%s[%v]", path, k), x.MapIndex(k), yv)
			}
		default:
			if !reflect.DeepEqual(valueOf(x), valueOf(y)) {
				out = append(out, fmt.Sprintf("%s: %v vs %v", path, valueOf(x), valueOf(y)))
			}
		}
	}
	ca, cb := *a, *b
	ca.Entities, cb.Entities = canonicalVehicles(a.Entities), canonicalVehicles(b.Entities)
	walk("State", reflect.ValueOf(&ca).Elem(), reflect.ValueOf(&cb).Elem())
	return out
}

// vehicleFields is, per component type, the VehicleEntity fields that are
// its own (S5Entity.h's VehicleHead … VehicleTail: the struct reads every
// field at its offset for every type, so the others are another field's
// bytes).
var vehicleFields = func() map[uint8]map[string]bool {
	set := func(names ...string) map[string]bool {
		m := map[string]bool{}
		for _, n := range names {
			m[n] = true
		}
		return m
	}
	common := []string{"ID", "EntityType", "SpriteYaw", "SpritePitch", "Owner", "PosX", "PosY", "PosZ", "HeadID", "Name",
		"SpriteWidth", "SpriteHeightNegative", "SpriteHeightPositive", "TileX", "TileY", "TileBaseZ", "NextCarID",
		"RemainingDistance", "TrackAndDirection", "SubPosition", "TrackType", "RoutingHandle", "Mode", "VehicleFlags",
		"Flags38", "Var3C", "SpriteLeft", "SpriteTop", "SpriteRight", "SpriteBottom"}
	with := func(extra ...string) map[string]bool { return set(append(append([]string(nil), common...), extra...)...) }
	body := []string{"ColourPrimary", "ColourSecondary", "ObjectSpriteType", "ObjectID", "Var44", "Animation", "CargoFrame", "Cargo",
		"BodyIndex", "CreationDay", "Var5A", "WheelSlipping", "BreakdownFlags", "RefundCost", "BreakdownTimeout"}
	return map[uint8]map[string]bool{
		scenario.VehicleTypeHead: with("OrdinalNumber", "OrderTableOffset", "CurrentOrder", "SizeOfOrderTable", "TrainAcceptedCargoTypes",
			"HeadVar52", "HeadVar53", "StationID", "CargoTransferTimeout", "Var58", "Var5C", "Status", "VehicleType", "BreakdownFlags",
			"AirportMovementEdge", "TotalRefundCost", "CrashedTimeout", "ManualPower", "JourneyStartX", "JourneyStartY", "JourneyStartTicks", "LastAverageSpeed"),
		scenario.VehicleTypeVeh1: with("TargetSpeed", "TimeAtSignal", "Var48", "Var49", "DayCreated", "HeadVar52", "LastIncome",
			"TrainAcceptedCargoTypes"), // var_4E / var_50 read through the head's name (watermotion.go veh1Frac)
		scenario.VehicleTypeVeh2: with("Var4F", "TotalPower", "TotalWeight", "MaxSpeed", "CurrentSpeed", "MotorState", "BrakeLightTimeout",
			"RackRailMaxSpeed", "CurMonthRevenue", "Profit", "Veh2Reliability", "Var73"),
		scenario.VehicleTypeBogie:         with(append(body, "TotalWeight", "Reliability", "TimeoutToBreakdown")...),
		scenario.VehicleTypeBodyStart:     with(append(body, "ChuffSoundIndex")...),
		scenario.VehicleTypeBodyContinued: with(append(body, "ChuffSoundIndex")...),
		scenario.VehicleTypeTail:          with(),
	}
}()

// canonicalVehicles is the entities with every field not of its type zeroed.
func canonicalVehicles(es []scenario.VehicleEntity) []scenario.VehicleEntity {
	out := make([]scenario.VehicleEntity, len(es))
	for i, e := range es {
		keep := vehicleFields[e.EntityType]
		src, dst := reflect.ValueOf(e), reflect.ValueOf(&out[i]).Elem()
		for k := 0; k < src.NumField(); k++ {
			if keep[src.Type().Field(k).Name] {
				dst.Field(k).Set(src.Field(k))
			}
		}
	}
	return out
}

// notSaved: state upstream doesn't save either — the entities' sprite
// rectangles (loadSave nulls them, updateSpatialIndex redraws them),
// VehicleHead.cpp's module static _vehicleMangled_113623B, the legacy
// return state's last placed wall.
var notSaved = map[string]bool{"SpriteLeft": true, "SpriteTop": true, "SpriteRight": true, "SpriteBottom": true,
	"vehicleMangled113623B": true, "LastPlacedWallBaseZ": true, "LastCollision": true}

// valueOf reads unexported fields too (by value).
func valueOf(v reflect.Value) any {
	switch v.Kind() {
	case reflect.Bool:
		return v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint()
	case reflect.String:
		return v.String()
	case reflect.Float32, reflect.Float64:
		return v.Float()
	}
	return fmt.Sprintf("<%s>", v.Kind())
}

func tickN(s *State, om *objects.ObjectManager, n int) {
	for range n {
		s.Tick(nil, om)
	}
}

// TestSaveRoundTrip_title: the title game, played, saved and loaded back,
// is the same state — and plays on the same.
func TestSaveRoundTrip_title(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	s := &State{Map: *m}
	s.ResetWaves()
	tickN(s, om, 3000)
	s2, _ := saveAndLoad(t, s)
	for _, d := range stateDiff(s, s2, 40) {
		t.Error("after the load:", d)
	}
	tickN(s, om, 1000)
	tickN(s2, om, 1000)
	for _, d := range stateDiff(s, s2, 40) {
		t.Error("played on:", d)
	}
}

// TestSaveRoundTrip_scenario: a scenario started as a new game (companies
// and entities made in play, no saved bytes behind them), with vehicles
// bought, played for months, saved and loaded back, is the same state — and
// plays on the same.
func TestSaveRoundTrip_scenario(t *testing.T) {
	const file = "../Scenarios/Mixing Muscle.SC5"
	m, om := loadDataMap(t, file)
	sc, _ := scenario.LoadScenarioData("../../../locomotion/Data/" + file)
	s := &State{Map: *m}
	s.ResetWaves()
	s.StartScenario(sc.Options.StartYear, om)
	c := &Cmd{Company: s.ControllingCompany, Rng2: &s.Rng2, Objects: om}
	bought := 0
	for id := range om.Vehicles {
		if bought < 3 && om.Vehicles[id] != nil && s.isVehicleIndexUnlocked(c.Company, uint16(id)) {
			if _, ok := s.CreateVehicle(c, CmdApply, uint16(id), false, nil); ok {
				bought++
			}
		}
	}
	if bought == 0 {
		t.Fatal("no vehicle bought")
	}
	tickN(s, om, 8000)
	s2, _ := saveAndLoad(t, s)
	for _, d := range stateDiff(s, s2, 40) {
		t.Error("after the load:", d)
	}
	tickN(s, om, 1500)
	tickN(s2, om, 1500)
	for _, d := range stateDiff(s, s2, 40) {
		t.Error("played on:", d)
	}
}
