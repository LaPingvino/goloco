package sim

import (
	"fmt"
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"strconv"
	"strings"
)

// lists.go — read-only summaries of the world's towns, stations and vehicles
// for the toolbar list windows (Vehicles / Stations / Towns).  Each entry
// carries a tile coordinate so the caller can centre the camera on it via
// World.SetCamera.

// townSizeNames maps the S5 TownSize enum to a display label.
// OpenLoco reference: include/OpenLoco/World/Town.h enum class TownSize.
var townSizeNames = [...]string{"Hamlet", "Village", "Town", "City", "Metropolis"}

// TownInfo is one row of the Towns list window.
type TownInfo struct {
	Name         string // "Town N"
	TileX, TileY int    // town centre, tile grid
	Population   uint32
	Size         string
	SizeIndex    int // Town::size (hamlet 0 … metropolis 5)
	NumBuildings int
	NumStations  int
}

// TownList returns the loaded scenario's towns.  Town names are stored as
// dynamically-registered StringIds that are non-trivial to resolve, so each
// town is labelled "Town N" and shown with its population and coordinates.
func (w *Map) TownList() []TownInfo {
	out := make([]TownInfo, 0, len(w.Towns))
	for i, t := range w.Towns {
		size := ""
		if int(t.Size) < len(townSizeNames) {
			size = townSizeNames[t.Size]
		}
		out = append(out, TownInfo{
			Name:         w.TownName(t.Name, i),
			TileX:        int(t.X) / 32,
			TileY:        int(t.Y) / 32,
			Population:   t.Population,
			Size:         size,
			SizeIndex:    int(t.Size),
			NumBuildings: int(t.NumBuildings),
			NumStations:  int(t.NumStations),
		})
	}
	return out
}

// stationKindNames maps the StationElement StationType field to a label.
// OpenLoco reference: src/OpenLoco/src/Map/StationElement.h.
var stationKindNames = [...]string{"Train", "Road", "Airport", "Docks"}

// StationInfo is one row of the Stations list window.
type StationInfo struct {
	ID           int
	TileX, TileY int
	Kind         string
	Name         string
	Status       string
	Waiting      int
	Accepts      string
	Owner        int // Station::owner (-1 unknown)
}

// StationList returns every placed/loaded station on the map.  A station spans
// multiple tiles that share one StationID; only the first tile encountered is
// reported.  This covers both scenario-loaded stations and stations placed at
// runtime (both append a StationElement to their tile).
func (w *Map) StationList(cargos []*objects.CargoObject) []StationInfo {
	seen := make(map[int]bool)
	var out []StationInfo
	for y := range w.Tiles {
		for x := range w.Tiles[y] {
			for i := range w.Tiles[y][x].Stations {
				se := &w.Tiles[y][x].Stations[i]
				id := int(se.StationID)
				if seen[id] {
					continue
				}
				seen[id] = true
				kind := "Station"
				if int(se.StationType) < len(stationKindNames) {
					kind = stationKindNames[se.StationType]
				}
				status, waiting, accepts := w.stationCargoText(id, cargos)
				owner := -1
				if o, ok := w.StationOwner(id); ok {
					owner = int(o)
				}
				out = append(out, StationInfo{ID: id, TileX: x, TileY: y, Kind: kind, Name: w.StationName(id), Owner: owner,
					Status: status, Waiting: waiting, Accepts: accepts})
			}
		}
	}
	return out
}

// stationCargoText is Station::getStatusString, the total units waiting and
// the accepted-cargo list of StationList's drawScroll, from the saved cargo
// stats and the cargo objects (scenario slot order).
func (w *Map) stationCargoText(id int, cargos []*objects.CargoObject) (status string, waiting int, accepts string) {
	if id >= len(w.StationNames) {
		return "", 0, ""
	}
	var st, acc []string
	for c, cs := range w.StationNames[id].Cargo {
		var co *objects.CargoObject
		if c < len(cargos) {
			co = cargos[c]
		}
		if cs.Quantity > 0 {
			waiting += int(cs.Quantity)
			if co != nil {
				unit := co.UnitNamePlural
				if cs.Quantity == 1 {
					unit = co.UnitNameSingular
				}
				st = append(st, formatCount(unit, int(cs.Quantity)))
			}
		}
		if cs.Flags&1 != 0 && co != nil {
			acc = append(acc, co.Name)
		}
	}
	if len(st) == 0 {
		status = "Nothing waiting"
	} else {
		status = strings.Join(st, " + ") + " waiting"
	}
	return status, waiting, strings.Join(acc, ", ")
}

// StationAcceptsKey is StationList's orderByAccepts key: the accepted
// cargoes' names run together, in cargo order.
func (w *Map) StationAcceptsKey(id int, cargos []*objects.CargoObject) string {
	if id >= len(w.StationNames) {
		return ""
	}
	var b strings.Builder
	for c, cs := range w.StationNames[id].Cargo {
		if cs.Flags&1 != 0 && c < len(cargos) && cargos[c] != nil {
			b.WriteString(cargos[c].Name)
		}
	}
	return b.String()
}

// formatCount fills a cargo unit string's number argument. Object strings
// carry Locomotion control codes as raw bytes: 123 is int32_grouped (the
// number with thousands separators, Formatting.h ControlCodes).
func formatCount(s string, n int) string {
	if i := strings.IndexRune(s, rune(123)); i >= 0 && !strings.HasPrefix(s[i:], "{COMMA") && !strings.HasPrefix(s[i:], "{INT") && !strings.HasPrefix(s[i:], "{UINT") {
		return s[:i] + grouped(n) + s[i+1:]
	}
	for _, code := range []string{"{COMMA16}", "{INT32}", "{INT16}", "{UINT16}", "{COMMA32}"} {
		if strings.Contains(s, code) {
			return strings.Replace(s, code, strconv.Itoa(n), 1)
		}
	}
	return s
}

// grouped formats n with comma thousands separators (int32_grouped, en-GB).
func grouped(n int) string {
	d := strconv.Itoa(n)
	out := ""
	for i, c := range d {
		if i > 0 && (len(d)-i)%3 == 0 {
			out += ","
		}
		out += string(c)
	}
	return out
}

// vehicleNameTemplates are the en-GB default vehicle names (StringIds 4..9,
// en-GB.yml); {INT16} is the head's ordinalNumber.
var vehicleNameTemplates = map[uint16]string{
	4: "Train {INT16}",
	5: "Bus {INT16}",
	6: "Truck {INT16}",
	7: "Tram {INT16}",
	8: "Aircraft {INT16}",
	9: "Ship {INT16}",
}

// VehicleListRow is one vehicle of VehicleList.cpp's list.
type VehicleListRow struct {
	Type, Owner           int
	Name                  string
	Profit                int64
	AgeYears, Reliability int
	// The sort keys (VehicleList getOrder): totalRecentProfit, dayCreated.
	RecentProfit int64
	DayCreated   uint32
	HeadID       uint16
	// Status is getStatus's text; StatusRed its {COLOUR RED}.
	Status    string
	StatusRed bool
	Cargos    []int // VehicleCargos
}

// VehicleListRows returns the saved vehicles' list data: name (the head's
// name with its ordinal, or a user string), totalRecentProfit/4, age in
// years ((currentDay - dayCreated) / 365) and reliability.
func (w *Map) VehicleListRows() []VehicleListRow {
	var out []VehicleListRow
	for _, v := range w.LiveVehicleList() {
		name := w.VehicleName(v.Name, int(v.Ordinal))
		age := 0
		if w.CurrentDay >= v.DayCreated {
			age = int((w.CurrentDay - v.DayCreated) / 365)
		}
		out = append(out, VehicleListRow{Type: int(v.VehicleType), Owner: int(v.Owner), Name: name, Profit: int64(v.RecentProfit / 4),
			AgeYears: age, Reliability: int(v.Reliability), RecentProfit: int64(v.RecentProfit), DayCreated: v.DayCreated, HeadID: v.HeadID})
		out[len(out)-1].Status, out[len(out)-1].StatusRed = w.VehicleStatus(v)
		out[len(out)-1].Cargos = w.VehicleCargos(v)
	}
	return out
}

// CompanyName is a company's name (a user string, or StringIds 2 "New
// Company"); "" if unknown.
func (w *Map) CompanyName(id int) string {
	if id >= 0 && id < len(w.CompanyNames) {
		if n := w.CompanyNames[id]; n >= 0x8000 && int(n-0x8000) < len(w.UserStrings) {
			return w.UserStrings[n-0x8000]
		} else if n == stringIDNewCompany {
			return strings.ReplaceAll(i18n.Str(int(n)), "{POP16}", "") // new_company: "New Company{POP16}"
		}
	}
	return ""
}

// LiveVehicleList is the vehicle list's data read from the entities now
// (VehicleList.cpp walks EntityManager::VehicleList each time): every head
// with its Vehicle1 (dayCreated) and Vehicle2 (profits, reliability, broken
// down, speed). Before 2026-11-05 the list showed the save's snapshot
// (VehicleListData), so a bought vehicle never appeared and statuses froze.
func (w *Map) LiveVehicleList() []scenario.VehicleListEntry {
	idx := w.entityIndex()
	var out []scenario.VehicleListEntry
	for i := range w.Entities {
		h := &w.Entities[i]
		if h.EntityType != scenario.VehicleTypeHead {
			continue
		}
		tr, err := w.trainOf(h.ID, idx)
		if err != nil {
			continue
		}
		v1, v2 := &w.Entities[tr.Veh1], &w.Entities[tr.Veh2]
		var sum int32
		for _, p := range v2.Profit {
			sum += p
		}
		out = append(out, scenario.VehicleListEntry{
			HeadID: h.ID, Owner: h.Owner, VehicleType: h.VehicleType, Name: h.Name, Ordinal: h.OrdinalNumber, Status: h.Status,
			DayCreated: v1.DayCreated, RecentProfit: sum, Reliability: v2.Veh2Reliability,
			TileX: int16(h.TileX), VehicleFlags: h.VehicleFlags, OrderTableOffset: h.OrderTableOffset, CurrentOrder: h.CurrentOrder,
			SizeOfOrderTable: h.SizeOfOrderTable, StationID: h.StationID, ManualPower: h.ManualPower,
			BrokenDown: v2.Var73&1 != 0, CurrentSpeed: v2.CurrentSpeed,
		})
	}
	return out
}

// VehicleName formats a head's name (formatString(name, ordinalNumber)): a
// user string, or a type's numbered template ("Train 3").
func (w *Map) VehicleName(id uint16, ordinal int) string {
	if id >= 0x8000 && int(id-0x8000) < len(w.UserStrings) {
		return w.UserStrings[id-0x8000]
	}
	if t, ok := vehicleNameTemplates[id]; ok {
		return strings.Replace(t, "{INT16}", strconv.Itoa(ordinal), 1)
	}
	return fmt.Sprintf("Vehicle %d", ordinal)
}

// stringIDNewOwner is StringIds::new_owner.
const stringIDNewOwner = 3

// OwnerName is a company's owner's name (Company::ownerName): a user
// string, new_owner ("Unnamed{POP16}" — POP16 prints nothing), or — an AI
// company's, set to competitorObj->name (CompanyManager.cpp:605) — its
// competitor object's name; "" if unknown.
func (w *Map) OwnerName(id int, om *objects.ObjectManager) string {
	if id < 0 || id >= len(w.CompanyFaces) {
		return ""
	}
	f := w.CompanyFaces[id]
	switch n := f.OwnerName; {
	case n >= 0x8000 && int(n-0x8000) < len(w.UserStrings):
		return w.UserStrings[n-0x8000]
	case n == stringIDNewOwner:
		return strings.ReplaceAll(i18n.Str(i18n.NewOwner), "{POP16}", "")
	case om != nil && int(f.Competitor) < len(om.CompetitorObjects) && om.CompetitorObjects[f.Competitor] != nil:
		// The object's name starts with pop16 (0x88), which prints nothing.
		return strings.ReplaceAll(om.CompetitorObjects[f.Competitor].Name, "\x88", "")
	}
	return ""
}
