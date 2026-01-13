package worldmap

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Economy/Currency.h"
// #include "QuarterTile.h"
// #include "Tile.h"
// #include <functional>
// #include <sfl/small_set.hpp>
// namespace OpenLoco::World
// forward: struct BuildingElement;
// forward: struct TreeElement;
// namespace OpenLoco::World::TileManager
type ElementPositionFlags int

const (
)
// namespace OpenLoco::World::TileClearance
type ClearFuncResult int

const (
	AllCollisionsRemoved ClearFuncResult = iota
	Collision
	CollisionErrorSet
	NoCollision
	CollisionRemoved
)
// func SetCollisionErrorMessage(el World::TileElement) 
// func ApplyClearAtAllHeights(pos World::Pos2, baseZ uint8, clearZ uint8, qt QuarterTile, el any /* std::function<ClearFuncResult(TileElement */ ) bool
// func ApplyClearAtStandardHeight(pos World::Pos2, baseZ uint8, clearZ uint8, qt QuarterTile, el any /* std::function<ClearFuncResult(TileElement */ ) bool
// func CanConstructAt(pos World::Pos2, baseZ uint8, clearZ uint8, qt QuarterTile) bool
type RemovedBuildings = any /* sfl::small_set<World::Pos3, 128, LessThanPos3> */ 
// // Removes Buildings and Trees but everything else is a collision
// func ClearWithDefaultCollision(el World::TileElement, pos World::Pos2, removedBuildings RemovedBuildings, flags uint8, cost currency32_t) ClearFuncResult
// // Removes Buildings and Trees but everything else is **NOT** a collision
// func ClearWithoutDefaultCollision(el World::TileElement, pos World::Pos2, removedBuildings RemovedBuildings, flags uint8, cost currency32_t) ClearFuncResult
// // Removes a building as per normal clear function setup
// func ClearBuildingCollision(elBuilding World::BuildingElement, pos World::Pos2, removedBuildings RemovedBuildings, flags uint8, cost currency32_t) ClearFuncResult
// // Removes a tree as per normal clear function setup
// func ClearTreeCollision(elTree World::TreeElement, pos World::Pos2, flags uint8, cost currency32_t) ClearFuncResult
// // These are an additional return variable from the applyClear functions
// TileManager::ElementPositionFlags getPositionFlags();
