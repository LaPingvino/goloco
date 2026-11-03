package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// clearRoad's helpers (GameCommands/Road/CreateRoad.cpp 201–254, and
// GameCommands.cpp's checkCompanyCompatibility), item 5.9f3. The error texts
// are not kept (as elsewhere in goloco's game commands).

// roadTraitJunction is RoadTraitFlags::junction.
const roadTraitJunction = 1 << 6

// roadObject is the road object in a slot, or nil.
func (c *Cmd) roadObject(id uint8) *objects.RoadObject {
	if c.Objects != nil && int(id) < len(c.Objects.RoadObjects) {
		return c.Objects.RoadObjects[id]
	}
	return nil
}

// companyCompatible is checkCompanyCompatibility(Common): a neutral owner, or
// the updating company owns it, or the updating company is neutral (a town).
func (c *Cmd) companyCompatible(owner uint8) bool {
	return owner == CompanyNeutral || c.Company == owner || c.Company == CompanyNeutral
}

// aiUnkCheck is CreateRoad.cpp's aiUnkCheck.
func (c *Cmd) aiUnkCheck(e *scenario.RoadElement, a roadClearArgs) bool {
	if a.UnkFlags&(1<<4|1<<5) == 0 || a.Flags&CmdAiAllocated == 0 || e.AiAllocated {
		return false
	}
	return a.UnkFlags&(1<<4) != 0 || e.Owner == c.Company
}

// compatibleRoads is a road object's compatibleRoads (its compatible names
// resolved to slots).
func (c *Cmd) compatibleRoads(r *objects.RoadObject) uint16 {
	if r == nil || c.Objects == nil {
		return 0
	}
	_, roads := c.Objects.CompatibleSets(r.CompatibleNames)
	return roads
}

// roadTraitAndCompatibilityCheck is CreateRoad.cpp's
// roadTraitAndCompatbilityCheck: the same road object needs the junction
// trait; different ones must be compatible one way or the other.
func (c *Cmd) roadTraitAndCompatibilityCheck(e *scenario.RoadElement, a roadClearArgs, target, placed *objects.RoadObject) bool {
	if e.RoadObjectID == a.RoadObjectID {
		return target != nil && target.RoadPieces&roadTraitJunction != 0
	}
	return c.compatibleRoads(target)&(1<<a.RoadObjectID) != 0 || c.compatibleRoads(placed)&(1<<e.RoadObjectID) != 0
}

// finalJunctionChecks is CreateRoad.cpp's: a junction only between
// single-tile pieces, then the trait and compatibility check.
func (c *Cmd) finalJunctionChecks(e *scenario.RoadElement, a roadClearArgs, target, placed *objects.RoadObject) bool {
	if len(RoadPieces[a.RoadID]) > 1 || len(RoadPieces[e.RoadID]) > 1 {
		return false
	}
	return c.roadTraitAndCompatibilityCheck(e, a, target, placed)
}
