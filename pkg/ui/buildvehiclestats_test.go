package ui

import (
	"strings"
	"testing"

	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/sim"
)

// testCargoOM is an object manager with three cargoes: passengers and mail
// in category 1, coal in category 2; the unit names raw ('{' is byte 123,
// int32_grouped).
func testCargoOM() *objects.ObjectManager {
	return &objects.ObjectManager{CargoObjects: []*objects.CargoObject{
		{Name: "Passengers", CargoCategory: 1, UnitNameSingular: "{ passenger", UnitNamePlural: "{ passengers"},
		{Name: "Mail", CargoCategory: 1, UnitNameSingular: "{ bag of mail", UnitNamePlural: "{ bags of mail"},
		{Name: "Coal", CargoCategory: 2, UnitNameSingular: "{ tonne of coal", UnitNamePlural: "{ tonnes of coal"},
	}}
}

// The cargo string as VehicleObject::getCargoString builds it: the first
// cargo's capacity, "(or Mail)", " (refittable)", and the second slot's
// " + …"; two spaces between alternatives (a space before each but the
// first, one after each).
func TestVehicleCargoString(t *testing.T) {
	om := testCargoOM()
	v := &objects.VehicleObject{CargoSection: [2]objects.VehicleCargoSlot{{MaxCargo: 40, Categories: []uint16{1}, SpriteOffsets: []uint8{0}}}}
	if got, want := VehicleCargoString(v, om), "{NEWLINE}{COLOUR WINDOW_2} Capacity: {COLOUR BLACK}40 passengers (or Mail)"; got != want {
		t.Errorf("one slot:\n got %q\nwant %q", got, want)
	}
	v.Flags = objects.VehicleFlagRefittable
	v.CargoSection[1] = objects.VehicleCargoSlot{MaxCargo: 1, Categories: []uint16{2}, SpriteOffsets: []uint8{0}}
	if got, want := VehicleCargoString(v, om), "{NEWLINE}{COLOUR WINDOW_2} Capacity: {COLOUR BLACK}40 passengers (or Mail) (refittable) + 1 tonne of coal"; got != want {
		t.Errorf("two slots:\n got %q\nwant %q", got, want)
	}
	om.CargoObjects = append(om.CargoObjects, &objects.CargoObject{Name: "Goods", CargoCategory: 1})
	v.Flags = 0
	v.CargoSection[1] = objects.VehicleCargoSlot{}
	if got, want := VehicleCargoString(v, om), "{NEWLINE}{COLOUR WINDOW_2} Capacity: {COLOUR BLACK}40 passengers (or Mail  or Goods)"; got != want {
		t.Errorf("two alternatives:\n got %q\nwant %q", got, want)
	}
	if got := VehicleCargoString(&objects.VehicleObject{}, om); got != "" {
		t.Errorf("no cargo: %q", got)
	}
}

// VehicleStats in draw's order; designed vs proposed by the unlocked test;
// obsolete left out at 0xFFFF; power only with some; the rack rail's lines.
func TestVehicleStats(t *testing.T) {
	v := &objects.VehicleObject{Mode: objects.TransportModeAir, Designed: 1955, Obsolete: 0xFFFF, Power: 900,
		Weight: 40, Speed: 300, NumCarComponents: 1}
	v.CarComponents[0].BodySpriteIdx = 0
	v.BodySprites[0].HalfLength = 32 // 64 long: 64 × 100 / 4 / 32 = 50 → 0.50
	got := VehicleStats(v, nil, VehicleStatsArgs{Cost: 1000, RunningCost: 20, Unlocked: true})
	for _, want := range []string{" Cost: {COLOUR BLACK}", " Running Cost: ", "/month", " Designed: {COLOUR BLACK}1955",
		" Length: {COLOUR BLACK}0.50 tile(s)", " Weight: {COLOUR BLACK}40t", " Max. Speed: {COLOUR BLACK}300"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %q", want, got)
		}
	}
	for _, not := range []string{"Obsolete", "Power", "Requires"} { // aircraft: no power, no track
		if strings.Contains(got, not) {
			t.Errorf("unexpected %q in %q", not, got)
		}
	}
	if i, j := strings.Index(got, "Cost:"), strings.Index(got, "Max. Speed"); i < 0 || j < i {
		t.Errorf("order: %q", got)
	}
	got = VehicleStats(v, nil, VehicleStatsArgs{})
	if !strings.Contains(got, " Proposed: {COLOUR BLACK}1955") || strings.Contains(got, "Designed") {
		t.Errorf("locked: %q", got)
	}
	v.Mode, v.Obsolete = objects.TransportModeRoad, 1990
	v.Flags = objects.VehicleFlagRackRail | 1<<9 // anyRoadType: StringIds::road
	v.RackSpeed = 20
	got = VehicleStats(v, &objects.ObjectManager{}, VehicleStatsArgs{Unlocked: true})
	for _, want := range []string{" Obsolete: {COLOUR BLACK}1990", " Requires: {COLOUR BLACK}Road ( + ", "for steep slopes)",
		" Power: {COLOUR BLACK}900", " on ", "{NEWLINE}{COLOUR WINDOW_2} Weight"} {
		if !strings.Contains(got, want) {
			t.Errorf("road: missing %q in %q", want, got)
		}
	}
}

// The coded string's lines carry the colour runes, break at {NEWLINE}, and
// wrap at the width.
func TestWrapCoded(t *testing.T) {
	lines := WrapCoded("{COLOUR WINDOW_2} Cost: {COLOUR BLACK}£1{NEWLINE}{COLOUR WINDOW_2} Weight: {COLOUR BLACK}4t", G1FontMediumBold, 180)
	if len(lines) != 2 || lines[0] != "\x01 Cost: \x02£1" || lines[1] != "\x01 Weight: \x024t" {
		t.Errorf("lines %q", lines)
	}
}

// onScrollMouseOver: rowHover is the hovered row's vehicle id; past the
// last row or off the list it stays; onResize picks the first row when
// none is hovered.
func TestBuildVehicleHover(t *testing.T) {
	w := &BuildVehicleWindow{WinBase: WinBase{X: 10, Y: 20, W: 500, H: 400}, NumTrackTabs: 1, RowHover: -1,
		Rows: []BVRow{{ID: 7}, {ID: 3}, {ID: 12}}}
	w.OnResize()
	if w.RowHover != 7 {
		t.Fatalf("onResize: rowHover %d, want the first row's 7", w.RowHover)
	}
	lw := BuildVehicleLayout(w.W, w.H, w.Disabled, w.NumTrackTabs)[bvList]
	rowH := sim.BuildVehicleRowHeight[0]
	x, y := w.X+lw.L+20, w.Y+lw.T+1
	w.MouseOver(x, y+rowH+2)
	if w.RowHover != 3 {
		t.Errorf("second row: %d", w.RowHover)
	}
	w.MouseOver(x, y+3*rowH+2) // past the last row
	if w.RowHover != 3 {
		t.Errorf("past the rows: %d", w.RowHover)
	}
	w.MouseOver(-1<<20, -1<<20) // another window on top
	if w.RowHover != 3 {
		t.Errorf("off the list: %d", w.RowHover)
	}
	w.MouseOver(x, y+2*rowH)
	if w.RowHover != 12 {
		t.Errorf("third row: %d", w.RowHover)
	}
	w.OnResize()
	if w.RowHover != 12 {
		t.Errorf("onResize keeps a hovered row: %d", w.RowHover)
	}
}

// Refresh (sub_4B92A5) takes the new list and clears rowHover; the place,
// the scroll and what the last frame drew stay.
func TestBuildVehicleRefresh(t *testing.T) {
	w := &BuildVehicleWindow{WinBase: WinBase{X: 10, Y: 20, W: 500, H: 400}, NumTrackTabs: 1, RowHover: 4,
		Scroll: 22, shownHover: 4, Rows: make([]BVRow, 30)}
	w.Refresh(&BuildVehicleWindow{WinBase: WinBase{W: 380, H: 233}, NumTrackTabs: 1, RowHover: 9,
		Rows: make([]BVRow, 30), TargetName: "Train 1"})
	if w.RowHover != -1 || w.Scroll != 22 || w.X != 10 || w.W != 500 || w.ShownHover() != 4 || w.TargetName != "Train 1" {
		t.Errorf("after refresh: hover %d scroll %d at %d,%d w %d shown %d target %q", w.RowHover, w.Scroll, w.X, w.Y, w.W, w.ShownHover(), w.TargetName)
	}
}
