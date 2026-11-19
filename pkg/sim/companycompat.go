package sim

// GameCommands::checkCompanyCompatibility (GameCommands.cpp 0x00431E6A) and
// its overloads: whether the updating company may touch something a company
// owns. Refused, the command's error text is kErrorBelongsToOther and the
// owner and the thing are kept (_errorCompanyId, _errorElementData) — the
// error window then names them beside the owner's face (loc_4314EA:
// Error::openWithCompetitor; cmd/goloco openCmdError).

// ErrorBelongsToOther is kErrorBelongsToOther: the error text that is no
// string but asks for the competitor's error window.
const ErrorBelongsToOther = 0xFFFE

// ErrorElementKind is ErrorElementKind: what belongs to the other company.
type ErrorElementKind uint8

const (
	ErrorKindNone ErrorElementKind = iota
	ErrorKindTrack
	ErrorKindRoad
	ErrorKindStation
)

// companyCompatibleCommon is checkCompanyCompatibilityCommon: a neutral
// owner, or the updating company is the owner or neutral (a town, the
// generator); refused, the error text and the owner are set.
func (c *Cmd) companyCompatibleCommon(owner uint8) bool {
	if owner == CompanyNeutral || c.Company == owner || c.Company == CompanyNeutral {
		return true
	}
	c.ErrorText, c.ErrorArg, c.ErrorCompany = ErrorBelongsToOther, "", owner
	return false
}

// checkCompanyCompatibility is checkCompanyCompatibility(company): a thing
// without an element (a vehicle, a company).
func checkCompanyCompatibility(c *Cmd, owner uint8) bool {
	return c.checkCompatibility(owner, ErrorKindNone, 0)
}

// checkTrackCompatibility is checkCompanyCompatibility(company, TrackElement).
func (c *Cmd) checkTrackCompatibility(owner, trackObjectID uint8) bool {
	return c.checkCompatibility(owner, ErrorKindTrack, uint16(trackObjectID))
}

// checkRoadCompatibility is checkCompanyCompatibility(company, RoadElement).
func (c *Cmd) checkRoadCompatibility(owner, roadObjectID uint8) bool {
	return c.checkCompatibility(owner, ErrorKindRoad, uint16(roadObjectID))
}

// checkStationCompatibility is checkCompanyCompatibility(company,
// StationElement).
func (c *Cmd) checkStationCompatibility(owner uint8, stationID uint16) bool {
	return c.checkCompatibility(owner, ErrorKindStation, stationID)
}

func (c *Cmd) checkCompatibility(owner uint8, kind ErrorElementKind, object uint16) bool {
	if c.companyCompatibleCommon(owner) {
		return true
	}
	c.ErrorKind, c.ErrorObject = kind, object
	return false
}
