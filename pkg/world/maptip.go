package world

import (
	"strconv"
	"strings"

	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/ui"
)

// MapTip is a map tooltip's text (ui.MapTip's: newlines as Loco newlines,
// ui.MapTipColour3 where {COLOUR WINDOW_3} switches) and its owner
// (MapToolTip::setOwner; -1: CompanyId::null).
type MapTip struct {
	Text  string
	Owner int
}

const tipNewline = "\x05" // ControlCodes::newline

// tipTwoLines is wcolour3_stringid ("{STRINGID}{NEWLINE}{COLOUR WINDOW_3}{STRINGID}").
func tipTwoLines(top, bottom string) string {
	return top + tipNewline + string(ui.MapTipColour3) + bottom
}

// MapTipLeft is the tooltip getItemLeft's argument functions give an item
// ItemLeftAt found (getTownArguments, getStationArguments,
// getIndustryArguments, getVehicleArguments, getHeadquarterArguments).
// PARTIAL: the hovered station's highlight (hoveringOverStation).
func (w *World) MapTipLeft(it LeftItem) (MapTip, bool) {
	m := &w.Map
	switch it.Kind {
	case paint.ItemTownLabel: // wcolour3_stringid_2: name, then town_size_and_population
		if it.ID >= len(m.Towns) {
			return MapTip{}, false
		}
		t := &m.Towns[it.ID]
		size := ""
		if int(t.Size) < len(sim.TownSizeNames) {
			size = sim.TownSizeNames[t.Size]
		}
		return MapTip{Text: tipTwoLines(m.TownName(t.Name, it.ID), size+" population "+formatInt(int64(t.Population))), Owner: -1}, true
	case paint.ItemStationLabel: // stringid_stringid_wcolour3_stringid: name, icons; status, accepts
		if it.ID >= len(m.StationNames) {
			return MapTip{}, false
		}
		st := &m.StationNames[it.ID]
		status, accepts := m.StationCargoText(it.ID, w.cargoObjects())
		owner := -1
		if st.Owner != 0xFF {
			owner = int(st.Owner)
		}
		top := m.StationName(it.ID) + " " + sim.StationIcons(uint16(st.Flags))
		return MapTip{Text: tipTwoLines(top, status+tipNewline+"Accepts "+accepts), Owner: owner}, true
	case paint.ItemIndustry: // the name; with a status, wcolour3_stringid_2
		name, ok := m.IndustryName(it.ID, w.renderer.ObjMgr)
		if !ok {
			return MapTip{}, false
		}
		if status := m.IndustryStatusString(it.ID, w.renderer.ObjMgr); status != "" {
			return MapTip{Text: tipTwoLines(name, status), Owner: -1}, true
		}
		return MapTip{Text: name, Owner: -1}, true
	case paint.ItemEntity: // company_vehicle / competitor_vehicle, then the head's status
		v, ok := m.LiveVehicle(uint16(it.ID))
		if !ok {
			return MapTip{}, false
		}
		name := m.VehicleName(v.Name, int(v.Ordinal))
		if ctrl, _ := w.ControllingCompany(); int(v.Owner) != ctrl {
			name = m.CompanyName(int(v.Owner)) + " - " + name // competitor_vehicle
		}
		status, _ := m.VehicleStatus(v)
		return MapTip{Text: tipTwoLines(name, status), Owner: int(v.Owner)}, true
	case paint.ItemHeadquarterBuilding:
		return MapTip{Text: tipTwoLines(m.CompanyName(it.ID), "Headquarters"), Owner: it.ID}, true
	}
	return MapTip{}, false
}

// formatInt is {INT32}: thousands separated by commas.
func formatInt(n int64) string {
	s := strconv.FormatInt(n, 10)
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(c)
	}
	if neg {
		return "-" + b.String()
	}
	return b.String()
}
