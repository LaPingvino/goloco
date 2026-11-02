package sim

import (
	"fmt"
	"github.com/LaPingvino/goloco/pkg/objects"
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
				if id < len(w.StationOwners) {
					owner = int(w.StationOwners[id])
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
	for _, v := range w.VehicleListData {
		name := fmt.Sprintf("Vehicle %d", v.Ordinal)
		if v.Name >= 0x8000 && int(v.Name-0x8000) < len(w.UserStrings) {
			name = w.UserStrings[v.Name-0x8000]
		} else if t, ok := vehicleNameTemplates[v.Name]; ok {
			name = strings.Replace(t, "{INT16}", strconv.Itoa(int(v.Ordinal)), 1)
		}
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
		} else if n == 2 {
			return "New Company"
		}
	}
	return ""
}
