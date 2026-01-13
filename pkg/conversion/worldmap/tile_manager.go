package worldmap

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Tile.h"
// #include "TileClearance.h"
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <cstdint>
// #include <set>
// #include <span>
// namespace OpenLoco::World
// forward: class QuarterTile;
// forward: struct BuildingElement;
// forward: struct TreeElement;
// forward: struct SurfaceElement;
// forward: struct RoadElement;
type ElementType int

const (
)
// namespace OpenLoco::World::TileManager
const MaxElements int = 3 * kMapColumns * kMapRows
const MaxElementsOnOneTile int = 1024
const MaxUsableElements int = kMaxElements - kMaxElementsOnOneTile
// const TileElement* const kInvalidTile = reinterpret_cast<const TileElement*>(static_cast<intptr_t>(-1));
type ElementPositionFlags int

const (
	None ElementPositionFlags = 0
	AboveGround ElementPositionFlags = 1 << 0
	Underground ElementPositionFlags = 1 << 1
	PartiallyUnderwater ElementPositionFlags = 1 << 2
	Underwater ElementPositionFlags = 1 << 3
	AnyHeightBuildingCollisions ElementPositionFlags = 1 << 7
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(ElementPositionFlags);
// func AllocateMapElements() 
// func Initialise() 
// std::span<TileElement> getElements();
// func NumFreeElements() uint32
// func Get(pos TilePos2) Tile
// func Get(pos Pos2) Tile
// func Get(x coord_t, y coord_t) Tile
// func SetElements(elements any /* std::span<TileElement> */ ) 
// func RemoveElement(element TileElement) 
// // This is used with wasRemoveOnLastElement to indicate that pointer passed to removeElement is now bad
// func SetRemoveElementPointerChecker(element TileElement) 
// // See above. Used to indicate if pointer to removeElement is now bad
// func WasRemoveOnLastElement() bool
// // Note: Any TileElement pointers invalid after this call
// TileElement* insertElement(ElementType type, const Pos2& pos, uint8_t baseZ, uint8_t occupiedQuads);
// // Note: Any TileElement pointers invalid after this call
// template<typename TileT>
// TileT* insertElement(const Pos2& pos, const uint8_t baseZ, const uint8_t occupiedQuads)
// func InsertElement(TileT::kElementType, pos, baseZ, occupiedQuads) return
// // Note: `after` pointer will be invalid after this call
// TileElement* insertElementAfterNoReorg(TileElement* after, ElementType type, const Pos2& pos, uint8_t baseZ, uint8_t occupiedQuads);
// // Note: `after` pointer will be invalid after this call
// template<typename TileT>
// TileT* insertElementAfterNoReorg(TileElement* after, const Pos2& pos, const uint8_t baseZ, const uint8_t occupiedQuads)
// func InsertElementAfterNoReorg(after, TileT::kElementType, pos, baseZ, occupiedQuads) return
// // Special road element insert
// World::RoadElement* insertElementRoad(const Pos2& pos, uint8_t baseZ, uint8_t occupiedQuads);
// func GetHeight(pos Pos2) TileHeight
// func GetSurfaceCornerHeight(surface SurfaceElement) SmallZ
// func GetSurfaceCornerDownHeight(surface SurfaceElement, cornerMask uint8) SmallZ
// func UpdateTilePointers() 
// // Only disables first call to defrag
// func DisablePeriodicDefrag() 
// // Fully defragment the tile element array
// func Reorganise() 
// // Defragments singular tile (chosen tile updates each call)
// func DefragmentTilePeriodic() 
// func CheckFreeElementsAndReorganise() bool
// func GetTileOwner(el World::TileElement) CompanyId
// func MapInvalidateTileFull(pos World::Pos2) 
// func ResetSurfaceClearance() 
// func MountainHeight(loc World::Pos2) int16
// func CountSurroundingWaterTiles(pos Pos2) uint16
// func CountSurroundingDesertTiles(pos Pos2) uint16
// func CountSurroundingTrees(pos Pos2) uint16
// func CountNearbyWaterTiles(pos Pos2) uint16
// func Update() 
// func UpdateYearly() 
// func RemoveSurfaceIndustry(pos Pos2) 
// func RemoveSurfaceIndustryAtHeight(pos Pos3) 
// func CreateDestructExplosion(pos World::Pos3) 
// func RemoveBuildingElement(element BuildingElement, pos World::Pos2) 
// func RemoveTree(element TreeElement, flags uint8, pos World::Pos2) 
// func RemoveAllWallsOnTileAbove(pos World::TilePos2, baseZ SmallZ) 
// func RemoveAllWallsOnTileBelow(pos World::TilePos2, baseZ SmallZ) 
// func SetLevelCrossingFlags(pos World::Pos3) 
// func SetTerrainStyleAsCleared(pos Pos2) 
// func SetTerrainStyleAsClearedAtHeight(pos Pos3) 
// func AdjustSurfaceHeight(pos World::Pos2, targetBaseZ SmallZ, slopeFlags uint8, removedBuildings World::TileClearance::RemovedBuildings, flags uint8) uint32
// func AdjustWaterHeight(pos World::Pos2, targetHeight SmallZ, removedBuildings World::TileClearance::RemovedBuildings, flags uint8) uint32
