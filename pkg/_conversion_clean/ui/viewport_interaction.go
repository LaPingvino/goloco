package ui

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Graphics/Gfx.h"
// #include "Location.hpp"
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <OpenLoco/Engine/World.hpp>
// #include <optional>
// using namespace OpenLoco::World;
// namespace OpenLoco::Paint
// forward: struct PaintStruct;
// namespace OpenLoco::Ui
// forward: struct Viewport;
// forward: struct Window;
// namespace OpenLoco::Ui::ViewportInteraction
type InteractionItem int

const (
func init() {
	NoInteraction InteractionItem = 0
	Surface InteractionItem = 1
	IndustryTree InteractionItem = 2
	Entity InteractionItem = 3
	Track InteractionItem = 4
	TrackExtra InteractionItem = 5
	Signal InteractionItem = 6
	TrainStation InteractionItem = 7
	RoadStation InteractionItem = 8
	Airport InteractionItem = 9
	Dock InteractionItem = 10
	Water InteractionItem = 11
	Tree InteractionItem = 12
	Wall InteractionItem = 13
	TownLabel InteractionItem = 14
	StationLabel InteractionItem = 15
	Road InteractionItem = 16
	RoadExtra InteractionItem = 17
	Bridge InteractionItem = 18
	Building InteractionItem = 19
	Industry InteractionItem = 20
	HeadquarterBuilding InteractionItem = 21
	BuildingInfo InteractionItem = 22
}

)
type InteractionItemFlags int

func init() {
	None InteractionItemFlags = 0
// SKIPPED CONSTRUCTOR: 	Surface InteractionItemFlags = (1 << 0)
// SKIPPED CONSTRUCTOR: 	Entity InteractionItemFlags = (1 << 1)
// SKIPPED CONSTRUCTOR: 	Track InteractionItemFlags = (1 << 2)
// SKIPPED CONSTRUCTOR: 	Water InteractionItemFlags = (1 << 3)
// SKIPPED CONSTRUCTOR: 	Tree InteractionItemFlags = (1 << 4)
// SKIPPED CONSTRUCTOR: 	RoadAndTram InteractionItemFlags = (1 << 5)
// SKIPPED CONSTRUCTOR: 	RoadAndTramExtra InteractionItemFlags = (1 << 6)
// SKIPPED CONSTRUCTOR: 	Signal InteractionItemFlags = (1 << 7)
// SKIPPED CONSTRUCTOR: 	Wall InteractionItemFlags = (1 << 8)
// SKIPPED CONSTRUCTOR: 	HeadquarterBuilding InteractionItemFlags = (1 << 9)
// SKIPPED CONSTRUCTOR: 	Station InteractionItemFlags = (1 << 11)
// SKIPPED CONSTRUCTOR: 	TownLabel InteractionItemFlags = (1 << 12)
// SKIPPED CONSTRUCTOR: 	StationLabel InteractionItemFlags = (1 << 13)
// SKIPPED CONSTRUCTOR: 	TrackExtra InteractionItemFlags = (1 << 14)
// SKIPPED CONSTRUCTOR: 	Building InteractionItemFlags = (1 << 15)
// SKIPPED CONSTRUCTOR: 	Industry InteractionItemFlags = (1 << 16)

// OPENLOCO_ENABLE_ENUM_OPERATORS(InteractionItemFlags);
type InteractionArg struct {
// World::Pos2 pos;
// union
	Value uint32
// void* object;
// orphan member: InteractionItem type;
// orphan member: uint8 modId; // used for track mods and to indicate left/right signals
// InteractionArg() = default;
// func InteractionArg(_pos World::Pos2, _value uint32, _type InteractionItem, _unkBh uint8) constexpr
// : pos(_pos)
// , value(_value)
// , type(_type)
// , modId(_unkBh)
// InteractionArg(const Paint::PaintStruct& ps);
NoInteractionArg = InteractionArg({ 0, 0 }, 0, InteractionItem.noInteraction, 0) // auto
// func GetItemLeft(tempX int16, tempY int16) InteractionArg
// func RightOver(x int16, y int16) InteractionArg
// func HandleRightReleased(window Window, xPos int16, yPos int16) 
// std::pair<ViewportInteraction::InteractionArg, Ui::Viewport*> getMapCoordinatesFromPos(int32 screenX, int32 screenY, InteractionItemFlags flags);
// std::optional<World::Pos2> getSurfaceOrWaterLocFromUi(const Point& screenCoords);
// std::optional<std::pair<World::Pos2, Ui::Viewport*>> getSurfaceLocFromUi(const Point& screenCoords);