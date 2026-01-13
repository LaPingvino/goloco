package vehicles

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Vehicle.h"
// namespace OpenLoco::Vehicles
type VehicleBody struct {
	VehicleBase // embedded
	ColourScheme ColourScheme
	ObjectSpriteType uint8
	ObjectId uint16
	Var_44 int16
	AnimationFrame uint8
	CargoFrame uint8
	PrimaryCargo VehicleCargo
	BodyIndex uint8
	ChuffSoundIndex int8
	CreationDay uint32
	Var_5A uint32
	WheelSlipping uint8
	BreakdownFlags BreakdownFlags
	RefundCost uint32
	BreakdownTimeout uint8
// const VehicleObject* getObject() const;
	// method: bool update(const CarUpdateState& carState);
	// method: void secondaryAnimationUpdate(const Vehicle& train, const CarUpdateState& carState, const int32_t unkDistance);
	// method: void updateSegmentCrashed(const CarUpdateState& carState);
	// method: void sub_4AAB0B(const CarUpdateState& carState, const int32_t unkDistance);
	// method: void updateCargoSprite();
	// method: constexpr bool hasBreakdownFlags(BreakdownFlags flagsToTest) const
// return (breakdownFlags & flagsToTest) != BreakdownFlags::none;
}
const VehicleBodyVehicleThingType any = VehicleEntityType.body_continued
// func Sub_4AC255(backBogie VehicleBogie, frontBogie VehicleBogie) 
// private:
// func AnimationUpdate(carState CarUpdateState, unkDistance int32) 
// func SteamPuffsAnimationUpdate(train Vehicle, carState CarUpdateState, unkDistance int32, num uint8, var_05 int32) 
// func DieselExhaust1AnimationUpdate(train Vehicle, carState CarUpdateState, num uint8, var_05 int32) 
// func DieselExhaust2AnimationUpdate(train Vehicle, carState CarUpdateState, num uint8, var_05 int32) 
// func ElectricSpark1AnimationUpdate(train Vehicle, carState CarUpdateState, unkDistance int32, num uint8, var_05 int32) 
// func ElectricSpark2AnimationUpdate(train Vehicle, carState CarUpdateState, unkDistance int32, num uint8, var_05 int32) 
// func ShipWakeAnimationUpdate(train Vehicle, num uint8, var_05 int32) 
// func UpdateSpritePitchSteepSlopes(xyOffset uint16, zOffset int16) Pitch
// func UpdateSpritePitch(xyOffset uint16, zOffset int16) Pitch
// static_assert(sizeof(VehicleBody) <= sizeof(Entity));
