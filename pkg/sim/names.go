package sim

import (
	"fmt"
	"strings"
)

// TownName is Town::name formatted: a user string (StringIds 0x8000+, where the
// town name generator puts them); otherwise goloco's placeholder "Town N".
func (w *Map) TownName(id uint16, i int) string {
	if id >= 0x8000 && int(id-0x8000) < len(w.UserStrings) && w.UserStrings[id-0x8000] != "" {
		return w.UserStrings[id-0x8000]
	}
	return fmt.Sprintf("Town %d", i+1)
}

// Town name StringIds (StringManager.h): kTownNamesStart + a template id, or
// kTownNamesEnd for the plain town name.
const (
	townNamesStart = 0x9EE7
	townNamesEnd   = townNamesStart + 345
)

// StationName formats Station::name as formatStringImpl does: a town-name
// template ({STRINGID} = the station's town's name), the plain town name, or a
// user string. Falls back to "Station N".
func (w *Map) StationName(id int) string {
	if id >= len(w.StationNames) {
		return fmt.Sprintf("Station %d", id)
	}
	sn := w.StationNames[id]
	town := ""
	if int(sn.Town) < len(w.Towns) {
		town = w.TownName(w.Towns[sn.Town].Name, int(sn.Town))
	}
	switch {
	case sn.Name >= 0x8000 && sn.Name < 0x8000+2048:
		if int(sn.Name-0x8000) < len(w.UserStrings) {
			return w.UserStrings[sn.Name-0x8000]
		}
	case sn.Name == townNamesEnd:
		return town
	case sn.Name >= townNamesStart && sn.Name < townNamesEnd:
		if t, ok := stationNameTemplates[sn.Name-townNamesStart]; ok {
			return strings.Replace(t, "{STRINGID}", town, 1)
		}
	}
	return fmt.Sprintf("Station %d", id)
}
