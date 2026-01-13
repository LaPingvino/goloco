package world

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "LabelFrame.h"
// #include "Localisation/StringManager.h"
// #include "Map/Tile.h"
// #include "Speed.hpp"
// #include "Types.hpp"
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <OpenLoco/Core/Numerics.hpp>
// #include <cstdint>
// #include <limits>
// #include <optional>
// namespace OpenLoco
// namespace Gfx
// forward: class DrawingContext;
// namespace World
// forward: struct StationElement;
type StationCargoStatsFlags int

const (
	None StationCargoStatsFlags = 0
// // Is there a consumer of this cargo at this station.
	AcceptedForConsumer StationCargoStatsFlags = (1 << 0)
// // Can a producer (town building, industry) distribute cargo to
// // this station.
	AcceptedFromProducer StationCargoStatsFlags = (1 << 1)
	Flag2 StationCargoStatsFlags = (1 << 2)
	Flag3 StationCargoStatsFlags = (1 << 3)
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(StationCargoStatsFlags);
type StationCargoStats struct {
	Quantity uint16
// StationId origin = StationId::null; // 0x30
	Flags StationCargoStatsFlags
	Age uint8
	Rating uint8
	EnrouteAge uint8
	VehicleSpeed Speed16
	VehicleAge uint8
	IndustryId IndustryId
	DensityPerTile uint8
	// method: bool empty() const
// return origin == StationId::null;
}
// func IsAccepted() bool
// return (flags & StationCargoStatsFlags::acceptedForConsumer) != StationCargoStatsFlags::none;
// func IsAccepted(value bool) 
// flags = Numerics::setMask<StationCargoStatsFlags>(flags, StationCargoStatsFlags::acceptedForConsumer, value);
const MaxCargoStats int = 32
type StationType int

const (
	TrainStation StationType = 0
	RoadStation
	Airport
	Docks
)
type StationFlags int

const (
	None StationFlags = 0
	TransportModeRail StationFlags = (1 << 0)
	TransportModeRoad StationFlags = (1 << 1)
	TransportModeAir StationFlags = (1 << 2)
	TransportModeWater StationFlags = (1 << 3)
	Flag_4 StationFlags = (1 << 4)
	Flag_5 StationFlags = (1 << 5)
	Flag_6 StationFlags = (1 << 6)
	Flag_7 StationFlags = (1 << 7)
	Flag_8 StationFlags = (1 << 8)
	AllModes StationFlags = StationFlags.transportModeRail | StationFlags.transportModeRoad | StationFlags.transportModeAir | StationFlags.transportModeWater
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(StationFlags);
// func GetTransportIconsFromStationFlags(flags StationFlags) StringId
// forward: struct CargoSearchState;
type CatchmentFlags int

const (
	Flag_0 CatchmentFlags = 0
	Flag_1 CatchmentFlags = 1
)
// // 0x004FEBD0
// // Where:
// // - Q is the port origin
// // - P is the port
// // - X represents the border offsets
// //
// //    X X
// //  X Q P X
// //  X P P X
// //    X X
// constexpr std::array<World::TilePos2, 8> kPortBorderOffsets = {
// World::TilePos2{ -1, 0 },
// World::TilePos2{ -1, 1 },
// World::TilePos2{ 0, 2 },
// World::TilePos2{ 1, 2 },
// World::TilePos2{ 2, 1 },
// World::TilePos2{ 2, 0 },
// World::TilePos2{ 1, -1 },
// World::TilePos2{ 0, -1 },
type Station struct {
// StringId name = StringIds::null;              // 0x00
	X coord_t
	Y coord_t
	Z coord_t
	LabelFrame LabelFrame
	Owner CompanyId
	NoTilesTimeout uint8
	Flags StationFlags
	Town TownId
// StationCargoStats cargoStats[kMaxCargoStats]; // 0x2E
	StationTileSize uint16
// World::Pos3 stationTiles[80];                 // 0x1D0 Note: z coordinate also contains rotation so always floor
	Var_3B0 uint8
	Var_3B1 uint8
	Var_3B2 uint8
	AirportRotation uint8
// World::Pos3 airportStartPos{};           // 0x3B4
	AirportMovementOccupiedEdges uint32
func (s *Station) Empty() bool {
	// StationId id() const;
	// void update();
	// uint32_t calcAcceptedCargo(CargoSearchState& cargoSearchState) const;
	// char* getStatusString(char* buffer);
	// bool updateCargo();
	// int32_t calculateCargoRating(const StationCargoStats& cargo) const;
	// void updateLabel();
	// void invalidate();
	// void invalidateWindow();
	// void deliverCargoToStation(const uint8_t cargoType, const uint8_t cargoQuantity);
	// void deliverCargoToTown(uint8_t cargoType, uint16_t cargoQuantity);
	// void updateCargoDistribution();
	// private:
	// void updateCargoAcceptance();
	// void alertCargoAcceptanceChange(uint32_t oldCargoAcc, uint32_t newCargoAcc);
}
	// method: void setCatchmentDisplay(const Station* station, const CatchmentFlags flags);
	// method: bool isWithinCatchmentDisplay(const World::Pos2 pos);
}

type StationPotentialCargo struct {
	Accepted uint32
	Produced uint32
}
	// method: PotentialCargo calcAcceptedCargoTrainStationGhost(const Station* ghostStation, const World::Pos2& location, const uint32_t filter);
	// method: PotentialCargo calcAcceptedCargoAirportGhost(const Station* ghostStation, const uint8_t type, const World::Pos2& location, const uint8_t rotation, const uint32_t filter);
	// method: PotentialCargo calcAcceptedCargoDockGhost(const Station* ghostStation, const World::Pos2& location, const uint32_t filter);
	// method: PotentialCargo calcAcceptedCargoAi(const World::TilePos2 minPos, const World::TilePos2 maxPos);
	// method: void sub_491C6F(const uint8_t type, const World::Pos2& pos, const uint8_t rotation, const CatchmentFlags flag);
	// method: void sub_491D20(const World::Pos2& pos, const CatchmentFlags flag);
	// method: void sub_491BF5(const World::Pos2& pos, const CatchmentFlags flag);
	// method: void recalculateStationCenter(const StationId stationId);
	// method: void recalculateStationModes(const StationId stationId);
	// method: void addTileToStation(const StationId stationId, const World::Pos3& pos, uint8_t rotation);
	// method: void removeTileFromStation(const StationId stationId, const World::Pos3& pos, uint8_t rotation);
	// method: void removeTileFromStationAndRecalcCargo(const StationId stationId, const World::Pos3& pos, uint8_t rotation);
// World::StationElement* getStationElement(const World::Pos3& pos);
// std::optional<World::Pos3> getAirportMovementNodeLoc(const StationId stationId, uint8_t node);
	// method: void sub_48D794(const Station& station);
	// method: void drawStationName(Gfx::DrawingContext& drawingCtx, const Station& station, uint8_t zoom, bool isHovered);
}
