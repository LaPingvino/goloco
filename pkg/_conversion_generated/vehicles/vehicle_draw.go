package vehicles

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Types.hpp"
// #include <OpenLoco/Engine/Ui/Point.hpp>
// #include <optional>
// namespace OpenLoco
// namespace Gfx
// forward: class DrawingContext;
// namespace Vehicles
// forward: struct Car;
// forward: struct Vehicle;
// forward: struct VehicleObject;
// forward: struct VehicleObjectBodySprite;
type Pitch int

const (
// // roll/animationFrame
	Uint32_t Pitch = iota
	Uint32_t
	Void
	Void
	Int16_t
type VehicleInlineMode int

const (
	Basic VehicleInlineMode = iota
	Animated
)
type VehiclePartsToDraw int

const (
	Bogies VehiclePartsToDraw = iota
	Bodies
)
// func DrawVehicleInline(drawingCtx Gfx::DrawingContext, car Vehicles::Car, loc Ui::Point, mode VehicleInlineMode, parts VehiclePartsToDraw, std::nullopt *Colour) int16
// func GetWidthVehicleInline(car Vehicles::Car) int16
// func DrawTrainInline(drawingCtx Gfx::DrawingContext, train Vehicles::Vehicle, loc Ui::Point) int16
