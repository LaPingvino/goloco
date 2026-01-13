package objects

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Engine/Limits.h"
// #include "Object.h"
// #include <OpenLoco/Engine/Ui/Point.hpp>
// #include <optional>
// #include <span>
// #include <vector>
// namespace OpenLoco
// forward: class SawyerStreamWriter;
// namespace Gfx
// forward: class DrawingContext;
// forward: struct Object;
// forward: struct ObjectEntryExtended;
// forward: struct CargoObject;
// forward: struct InterfaceSkinObject;
// forward: struct SoundObject;
// forward: struct CurrencyObject;
// forward: struct SteamObject;
// forward: struct CliffEdgeObject;
// forward: struct WaterObject;
// forward: struct LandObject;
// forward: struct TownNamesObject;
// forward: struct WallObject;
// forward: struct TrainSignalObject;
// forward: struct LevelCrossingObject;
// forward: struct StreetLightObject;
// forward: struct TunnelObject;
// forward: struct BridgeObject;
// forward: struct TrainStationObject;
// forward: struct TrackExtraObject;
// forward: struct TrackObject;
// forward: struct RoadStationObject;
// forward: struct RoadExtraObject;
// forward: struct RoadObject;
// forward: struct AirportObject;
// forward: struct DockObject;
// forward: struct VehicleObject;
// forward: struct TreeObject;
// forward: struct SnowObject;
// forward: struct ClimateObject;
// forward: struct HillShapesObject;
// forward: struct BuildingObject;
// forward: struct ScaffoldingObject;
// forward: struct IndustryObject;
// forward: struct RegionObject;
// forward: struct CompetitorObject;
// forward: struct ScenarioTextObject;
// /**
// * Represents an index into the entire loaded object array. Not an index for
// * a specific object type. DO NOT USE
// */
type LoadedObjectIndex = int
type LandObjectFlags int

const (
)
// namespace OpenLoco::ObjectManager
// func GetMaxObjects(type ObjectType) int
// // 0x004FE250
// constexpr size_t counts[] = {
// Limits::kMaxInterfaceObjects,
// Limits::kMaxSoundObjects,
// Limits::kMaxCurrencyObjects,
// Limits::kMaxSteamObjects,
// Limits::kMaxRockObjects,
// Limits::kMaxWaterObjects,
// Limits::kMaxSurfaceObjects,
// Limits::kMaxTownNamesObjects,
// Limits::kMaxCargoObjects,
// Limits::kMaxWallObjects,
// Limits::kMaxTrainSignalObjects,
// Limits::kMaxLevelCrossingObjects,
// Limits::kMaxStreetLightObjects,
// Limits::kMaxTunnelObjects,
// Limits::kMaxBridgeObjects,
// Limits::kMaxTrainStationObjects,
// Limits::kMaxTrackExtraObjects,
// Limits::kMaxTrackObjects,
// Limits::kMaxRoadStationObjects,
// Limits::kMaxRoadExtraObjects,
// Limits::kMaxRoadObjects,
// Limits::kMaxAirportObjects,
// Limits::kMaxDockObjects,
// Limits::kMaxVehicleObjects,
// Limits::kMaxTreeObjects,
// Limits::kMaxSnowObjects,
// Limits::kMaxClimateObjects,
// Limits::kMaxHillShapesObjects,
// Limits::kMaxBuildingObjects,
// Limits::kMaxScaffoldingObjects,
// Limits::kMaxIndustryObjects,
// Limits::kMaxRegionObjects,
// Limits::kMaxCompetitorObjects,
// Limits::kMaxScenarioTextObjects
// return counts[(size_t)type];
const MaxObjects int = 859
// Object* getAny(const LoadedObjectHandle& handle);
// template<typename T>
// const T* get()
// static_assert(getMaxObjects(T::kObjectType) == 1);
// return reinterpret_cast<T*>(getAny({ T::kObjectType, 0 }));
// template<typename T>
// const T* get(size_t id)
// static_assert(getMaxObjects(T::kObjectType) != 1);
// return reinterpret_cast<T*>(getAny({ T::kObjectType, static_cast<LoadedObjectId>(id) }));
type ObjectHeader2 struct {
	DecodedFileSize uint32
}
// static_assert(sizeof(ObjectHeader2) == 0x4);
type ObjectHeader3 struct {
	NumImages uint32
	Intelligence uint8
	Aggressiveness uint8
	Competitiveness uint8
	VehicleSubType uint8
// uint8_t pad_08[0x4];
}
// static_assert(sizeof(ObjectHeader3) == 0xC);
type LoadObjectsResult struct {
	Success bool
// std::vector<ObjectHeader> problemObjects;
}
type DependentObjects struct {
// std::vector<ObjectHeader> required;
// std::vector<ObjectHeader> willLoad;
}
type TempLoadMetaData struct {
	FileSizeHeader ObjectHeader2
	DisplayData ObjectHeader3
	DependentObjects DependentObjects
}
// func FreeTemporaryObject() 
// std::optional<TempLoadMetaData> loadTemporaryObject(const ObjectHeader& header);
// Object* getTemporaryObject();
// func IsTemporaryObjectLoad() bool
// std::optional<LoadedObjectHandle> findObjectHandle(const ObjectHeader& header);
// // Calls findObjectHandle and if can't find performs a secondary check with slightly looser
// // definitions of what a matching custom header is (no checksum, partial flags)
// std::optional<LoadedObjectHandle> findObjectHandleFuzzy(const ObjectHeader& header);
// func ReloadAll() 
// ObjectHeader& getHeader(const LoadedObjectHandle& handle);
// std::vector<ObjectHeader> getHeaders();
// func LoadAll(objects any /* std::span<ObjectHeader> */ ) LoadObjectsResult
// func WritePackedObjects(fs SawyerStreamWriter, packedObjects []ObjectHeader) 
// func UnloadAll() 
// // Only unloads the entry (clears entry for packing does not free)
// func Unload(handle LoadedObjectHandle) 
// // Unloads and frees the entry
// func Unload(header ObjectHeader) 
// func Load(header ObjectHeader) bool
// func TryInstallObject(object ObjectHeader, data any /* std::span<std::byte> */ ) bool
// func GetByteLength(handle LoadedObjectHandle) int
// func DrawGenericDescription(drawingCtx Gfx::DrawingContext, rowPosition Ui::Point, designed uint16, obsolete uint16) 
// func UpdateRoadObjectIdFlags() 
// func UpdateDefaultLevelCrossingType() 
// func UpdateYearly2() 
// func UpdateTerraformObjects() 
// func UpdateLastTrackTypeOption() 
// std::span<LandObjectFlags> getLandObjectFlagsCache();
// // Calls function with the handle (LoadedObjectHandle) of each loaded object
// template<typename Function>
// func ForEachLoadedObject(func Function) 
// for (uint8_t objectTypeU = 0; objectTypeU < kMaxObjectTypes; ++objectTypeU)
// const auto objectType = static_cast<ObjectType>(objectTypeU);
// for (LoadedObjectId i = 0U; i < getMaxObjects(objectType); ++i)
// orphan member: LoadedObjectHandle handle{ objectType, i };
// if (getAny(handle) != nullptr)
// func(handle);
