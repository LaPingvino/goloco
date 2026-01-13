package worldmap

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <OpenLoco/Engine/Types.hpp>
// #include <OpenLoco/Engine/World.hpp>
// #include <span>
// namespace OpenLoco::World
type MapSelectionFlags int

const (
	None MapSelectionFlags = 0
	Enable MapSelectionFlags = 1 << 0
	EnableConstruct MapSelectionFlags = 1 << 1
	EnableConstructionArrow MapSelectionFlags = 1 << 2
	Unk_03 MapSelectionFlags = 1 << 3
	Unk_04 MapSelectionFlags = 1 << 4
	CatchmentArea MapSelectionFlags = 1 << 5
	HoveringOverStation MapSelectionFlags = 1 << 6
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(MapSelectionFlags);
type MapSelectionType int

const (
	Corner0 MapSelectionType = iota
	Corner1
	Corner2
	Corner3
	Full
	FullWater
	Quarter0
	Quarter1
	Quarter2
	Quarter3
	Edge0
	Edge1
	Edge2
	Edge3
)
type ConstructionArrow struct {
// World::Pos3 pos;
	Direction uint8
}
// func SetMapSelectionTiles(loc Pos2, selectionType MapSelectionType, toolSize uint16) uint16
// func SetMapSelectionSingleTile(loc Pos2, false bool setQuadrant =) uint16
// func MapInvalidateSelectionRect() 
// func ResetMapSelectionFreeFormTiles() 
// func AddMapSelectionFreeFormTile(pos Pos2) 
// std::span<const Pos2> getMapSelectionFreeFormTiles();
// func MapInvalidateMapSelectionFreeFormTiles() 
// func IsWithinMapSelectionFreeFormTiles(pos Pos2) bool
// func SetMapSelectionArea(locA Pos2, locB Pos2) 
// std::pair<Pos2, Pos2> getMapSelectionArea();
// func SetMapSelectionCorner(corner MapSelectionType) 
// func GetMapSelectionCorner() MapSelectionType
// func GetQuadrantOrCentreFromPos(loc Pos2) MapSelectionType
// func GetQuadrantFromPos(loc Pos2) MapSelectionType
// func GetSideFromPos(loc Pos2) MapSelectionType
// func GetMapSelectionFlags() MapSelectionFlags
// func HasMapSelectionFlag(flags MapSelectionFlags) bool
// func SetMapSelectionFlags(flags MapSelectionFlags) 
// func ResetMapSelectionFlag(flags MapSelectionFlags) 
// func ResetMapSelectionFlags() 
// const ConstructionArrow& getConstructionArrow();
// func SetConstructionArrow(arrow ConstructionArrow) 
