package sim

import (
	"fmt"
	"strings"

	"github.com/LaPingvino/goloco/pkg/objects"
)

// industryNameFormat formats a Loco-encoded object string with the industry's
// town as its argument, as StringManager::formatString does for the codes an
// industry name uses: stringidStr (0x84 + a StringId — kTownNamesEnd: the
// town's name, popping the town) and int16_grouped (0x7F, popping a number).
// Other control codes are skipped; printable bytes are copied. ok is false
// when the string uses a string id this doesn't format.
func (m *Map) industryNameFormat(raw []byte, town uint16, number int16) (string, bool) {
	var b strings.Builder
	ok := true
	for i := 0; i < len(raw); i++ {
		switch c := raw[i]; {
		case c == 0x84 && i+2 < len(raw): // stringidStr
			id := uint16(raw[i+1]) | uint16(raw[i+2])<<8
			i += 2
			if tw := m.TownByID(town); id == townNamesEnd && tw != nil { // the town by id
				b.WriteString(m.TownName(tw.Name, int(town)))
			} else {
				ok = false
			}
		case c == 0x7F: // int16_grouped
			b.WriteString(groupedInt(int32(number)))
		case c >= 0x20 && c < 0x7F:
			b.WriteByte(c)
		}
	}
	return b.String(), ok
}

func groupedInt(v int32) string {
	s := fmt.Sprint(v)
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	if neg {
		s = "-" + s
	}
	return s
}

// IndustryName is industry->name formatted with its town: a user string
// (0x8000+: renamed, or numbered as the second of its kind in a town), else
// its object's var_02 string ("{town} Coal Mine") — the only object string
// IndustryManager gives an industry (createNewIndustry); ok false when it
// can't be formatted.
func (m *Map) IndustryName(i int, om *objects.ObjectManager) (string, bool) {
	if i < 0 || i >= len(m.Industries) {
		return "", false
	}
	ind := m.Industries[i]
	if ind.Name >= kUserStringsStart && int(ind.Name-kUserStringsStart) < len(m.UserStrings) {
		return m.UserStrings[ind.Name-kUserStringsStart], true
	}
	if om == nil || int(ind.ObjectID) >= len(om.IndustryObjects) || om.IndustryObjects[ind.ObjectID] == nil {
		return "", false
	}
	return m.industryNameFormat(om.IndustryObjects[ind.ObjectID].RawStrings[1], ind.Town, 0)
}

// RenameIndustry is GameCommands::renameIndustry (RenameIndustry.cpp) with
// its three 12-character calls taken as one (at most 36 bytes): empty or
// unchanged does nothing; else a unique user string (refused when none is
// free; the query frees it); applied, the industry takes it and the old
// name's string is freed. PARTIAL: the screen invalidation.
func (m *Map) RenameIndustry(c *Cmd, flags uint8, i int, name string, om *objects.ObjectManager) (int32, bool) {
	if len(name) > 36 {
		name = name[:36]
	}
	if name == "" || i < 0 || i >= len(m.Industries) {
		return 0, true
	}
	if cur, _ := m.IndustryName(i, om); cur == name {
		return 0, true
	}
	id := m.userStringAllocate(name, true)
	if id == 0 {
		return 0, false
	}
	if flags&CmdApply == 0 {
		m.emptyUserString(id)
		return 0, true
	}
	old := m.Industries[i].Name
	m.Industries[i].Name = id
	m.emptyUserString(old)
	return 0, true
}

// IndustryFoundedByPlayer is renameIndustryPrompt's test outside the editor:
// founded by a company (flag_04) and that company a player's.
func (m *Map) IndustryFoundedByPlayer(i int) bool {
	ind := &m.Industries[i]
	return ind.Flags&industryFlag04 != 0 && m.IsPlayerCompany(int(ind.Owner))
}
