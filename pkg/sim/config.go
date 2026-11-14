package sim

// Config is the Options window's Miscellaneous settings the simulation reads
// (Config::get(), openloco.yml; all off by default, as Config.cpp reads them):
// breakdowns_disabled, townGrowthDisabled, disableVehicleLoadPenaltyCheat,
// disableStationSizeLimit, trainsReverseAtSignals, companyAIDisabled.
var Config struct {
	BreakdownsDisabled     bool
	TownGrowthDisabled     bool
	LoadPenaltyDisabled    bool
	StationSizeLimitOff    bool
	TrainsReverseAtSignals bool
	CompanyAIDisabled      bool // PARTIAL: goloco has no company AI tick to switch off yet
}
