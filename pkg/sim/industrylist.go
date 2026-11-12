package sim

import (
	"strings"

	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/objects"
)

// The industry list's data: Industry::getStatusString (World/Industry.cpp
// 0x0045932D) with IndustryObject::getProducedCargoString /
// getRequiredCargoString, and IndustryList.cpp's getAverageTransportedCargo
// and getProductionLastMonth.

// IndustryRow is one industry as the industry list shows it.
type IndustryRow struct {
	ID          int
	Name        string
	Status      string
	Produces    bool   // canProduceCargo: the production columns are drawn
	Transported uint8  // getAverageTransportedCargo: percent (0xFF: produces nothing)
	LastMonth   uint32 // getProductionLastMonth: the first produced cargo's amount (max: none)
	Unit        string // its unit name, singular or plural (as upstream: plural when > 1)
}

// cargoName is a cargo object's name by slot.
func cargoName(om *objects.ObjectManager, id uint8) string {
	if int(id) < len(om.CargoObjects) && om.CargoObjects[id] != nil {
		return om.CargoObjects[id].Name
	}
	return ""
}

// cargoList is getProducedCargoString / getRequiredCargoString: the names,
// joined by " and " (cargo_and), or " or " (cargo_or) for required cargo an
// industry needs only one of.
func cargoList(om *objects.ObjectManager, ids []uint8, or bool) string {
	sep := i18n.Label(i18n.CargoAnd)
	if or {
		sep = i18n.Label(i18n.CargoOr)
	}
	var b strings.Builder
	n := 0
	for _, id := range ids {
		if id == 0xFF {
			continue
		}
		if n > 0 {
			b.WriteString(sep)
		}
		b.WriteString(cargoName(om, id))
		n++
	}
	return b.String()
}

// IndustryStatusString is Industry::getStatusString: closing down, under
// construction, "Producing …", "Requires …", or "Requires … to produce …".
func (m *Map) IndustryStatusString(i int, om *objects.ObjectManager) string {
	ind := &m.Industries[i]
	if ind.Flags&industryFlagClosingDown != 0 {
		return i18n.Label(i18n.IndustryClosingDown)
	}
	if ind.UnderConstruction != 0xFF {
		return i18n.Label(i18n.IndustryUnderConstruction)
	}
	obj := om.GetIndustryObjectByIndex(int(ind.ObjectID))
	if obj == nil {
		return ""
	}
	produced, required := om.IndustryCargoIDs(obj)
	produces := produced[0] != 0xFF || produced[1] != 0xFF
	receives := required[0] != 0xFF || required[1] != 0xFF || required[2] != 0xFF
	if !receives {
		if !produces {
			return ""
		}
		return i18n.Label(i18n.IndustryProducing) + cargoList(om, produced[:], false)
	}
	s := i18n.Label(i18n.IndustryRequires) + cargoList(om, required[:], obj.Flags&objects.IndustryFlagRequiresAllCargo == 0)
	if !produces {
		return s
	}
	return s + i18n.Label(i18n.CargoToProduce) + cargoList(om, produced[:], false)
}

// IndustryList is the industry list's rows: every industry on the map, in
// the store's order (the window sorts).
func (m *Map) IndustryList(om *objects.ObjectManager) []IndustryRow {
	var out []IndustryRow
	for i := range m.Industries {
		ind := &m.Industries[i]
		if ind.Name == 0xFFFF {
			continue
		}
		name, _ := m.IndustryName(i, om)
		r := IndustryRow{ID: int(ind.ID), Name: name, Status: m.IndustryStatusString(i, om), Transported: 0xFF, LastMonth: ^uint32(0)}
		if obj := om.GetIndustryObjectByIndex(int(ind.ObjectID)); obj != nil {
			produced, _ := om.IndustryCargoIDs(obj)
			r.Produces = produced[0] != 0xFF || produced[1] != 0xFF
			if r.Produces { // getAverageTransportedCargo
				r.Transported = ind.PercentTransported[0]
				if produced[1] != 0xFF {
					r.Transported = ind.PercentTransported[1]
					if produced[0] != 0xFF {
						r.Transported = uint8((uint16(r.Transported) + uint16(ind.PercentTransported[0])) / 2)
					}
				}
				for k, c := range produced { // getProductionLastMonth: the first produced cargo only
					if c != 0xFF {
						r.LastMonth = uint32(ind.ProducedPrevMonth[k])
						if int(c) < len(om.CargoObjects) && om.CargoObjects[c] != nil {
							r.Unit = om.CargoObjects[c].UnitNameSingular
							if r.LastMonth > 1 {
								r.Unit = om.CargoObjects[c].UnitNamePlural
							}
						}
						break
					}
				}
			}
		}
		out = append(out, r)
	}
	return out
}
