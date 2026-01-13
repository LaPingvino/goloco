package s5

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include <cstddef>
// namespace OpenLoco::S5::Limits
const MaxMessages int = 200
const MaxCompanies int = 15
const MinTowns int = 1
const MaxTowns int = 80
const MaxIndustries int = 128
const MaxStations int = 1024
const MaxEntities int = 20000
const MaxAnimations int = 8192
const MaxWaves int = 64
const MaxUserStrings int = 2048
const MaxVehicles int = 1000
const MaxRoutingsPerVehicle int = 64
// // The number of orders appears to be the number of routings minus a null byte (OrderEnd)
const MaxOrdersPerVehicle int = kMaxRoutingsPerVehicle - 1
const MaxOrders int = 256000
const NumEntityLists int = 7
// // There is a separate pool of 200 entities dedicated for money
const MaxMoneyEntities int = 200
// // This is the main pool for everything that isn't money
const MaxNormalEntities int = kMaxEntities - kMaxMoneyEntities
// // Money is not counted in this limit
const MaxMiscEntities int = 4000
const MaxStationCargoDensity int = 15
const MaxInterfaceObjects int = 1
const MaxSoundObjects int = 128
const MaxCurrencyObjects int = 1
const MaxSteamObjects int = 32
const MaxRockObjects int = 8
const MaxWaterObjects int = 1
const MaxSurfaceObjects int = 32
const MaxTownNamesObjects int = 1
const MaxCargoObjects int = 32
const MaxWallObjects int = 32
const MaxTrainSignalObjects int = 16
const MaxLevelCrossingObjects int = 4
const MaxStreetLightObjects int = 1
const MaxTunnelObjects int = 16
const MaxBridgeObjects int = 8
const MaxTrainStationObjects int = 16
const MaxTrackExtraObjects int = 8
const MaxTrackObjects int = 8
const MaxRoadStationObjects int = 16
const MaxRoadExtraObjects int = 4
const MaxRoadObjects int = 8
const MaxAirportObjects int = 8
const MaxDockObjects int = 8
const MaxVehicleObjects int = 224
const MaxTreeObjects int = 64
const MaxSnowObjects int = 1
const MaxClimateObjects int = 1
const MaxHillShapesObjects int = 1
const MaxBuildingObjects int = 128
const MaxScaffoldingObjects int = 1
const MaxIndustryObjects int = 16
const MaxRegionObjects int = 1
const MaxCompetitorObjects int = 32
const MaxScenarioTextObjects int = 1
