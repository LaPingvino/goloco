package vehicles

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Entities/Entity.h"
// #include "Map/Track/TrackModSection.h"
// #include "Objects/VehicleObject.h"
// #include "Routing.h"
// #include "Speed.hpp"
// #include "Types.hpp"
// #include "Ui/Window.h"
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <OpenLoco/Core/Exception.hpp>
// namespace OpenLoco::Vehicles
var MaxRoadVehicleLength = 176 // auto
const WheelSlippingDuration uint8 = 64
type Flags38 int

const (
	None Flags38 = 0
	Unk_0 Flags38 = 1 << 0
	IsReversed Flags38 = 1 << 1
	Unk_2 Flags38 = 1 << 2
	JacobsBogieAvailable Flags38 = 1 << 3
	IsGhost Flags38 = 1 << 4
	FasterAroundCurves Flags38 = 1 << 5
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(Flags38);
type SoundFlags int

const (
	None SoundFlags = 0
	Flag0 SoundFlags = 1 << 0
	Flag1 SoundFlags = 1 << 1
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(SoundFlags);
type UpdateVar1136114Flags int

const (
	None UpdateVar1136114Flags = 0
	Unk_m00 UpdateVar1136114Flags = (1 << 0)
	NoRouteFound UpdateVar1136114Flags = (1 << 1)
	Crashed UpdateVar1136114Flags = (1 << 2)
	Unk_m03 UpdateVar1136114Flags = (1 << 3)
	ApproachingGradeCrossing UpdateVar1136114Flags = (1 << 4)
	Unk_m15 UpdateVar1136114Flags = (1 << 15)
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(UpdateVar1136114Flags);
// forward: struct OrderRingView;
// forward: struct VehicleHead;
// forward: struct Vehicle1;
// forward: struct Vehicle2;
// forward: struct VehicleBogie;
// forward: struct VehicleBody;
// forward: struct VehicleTail;
// forward: struct Vehicle;
// forward: struct VehicleSound;
type BreakdownFlags int

const (
	None BreakdownFlags = 0
	Unk_0 BreakdownFlags = 1 << 0
	BreakdownPending BreakdownFlags = 1 << 1
	BrokenDown BreakdownFlags = 1 << 2
	JourneyStarted BreakdownFlags = 1 << 3
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(BreakdownFlags);
type VehicleFlags int

const (
	None VehicleFlags = 0
	Unk_0 VehicleFlags = 1 << 0
	CommandStop VehicleFlags = 1 << 1
	Unk_2 VehicleFlags = 1 << 2
	Sorted VehicleFlags = 1 << 3
	Unk_5 VehicleFlags = 1 << 5
	ManualControl VehicleFlags = 1 << 6
	ShuntCheat VehicleFlags = 1 << 7
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(VehicleFlags);
type VehicleEntityType int

const (
	Head VehicleEntityType = 0
	Vehicle_1
	Vehicle_2
	Bogie
	Body_start
	Body_continued
	Tail
)
const AirportMovementNodeNull uint8 = 0xFF
const AirportMovementNoValidEdge uint8 = 0xFE
type TrackAndDirection struct {
}

type TrackAndDirectionTrackAndDirection struct {
	Data uint16
	// method: constexpr _TrackAndDirection(uint8_t id, uint8_t direction)
// : _data((id << 3) | direction)
}
func (t *TrackAndDirection) Id() uint8 {
	// constexpr uint8_t cardinalDirection() const { return _data & 0x3; }
	// constexpr bool isReversed() const { return _data & (1 << 2); }
	// constexpr void setReversed(bool state)
	// _data &= ~(1 << 2);
	// _data |= state ? (1 << 2) : 0;
	}
	// constexpr bool operator==(const _TrackAndDirection other) const { return _data == other._data; }
}
}

type TrackAndDirectionRoadAndDirection struct {
	Data uint16
	// method: constexpr _RoadAndDirection(uint8_t id, uint8_t direction)
// : _data((id << 3) | direction)
}
func (t *TrackAndDirection) Id() uint8 {
	// constexpr uint8_t cardinalDirection() const { return _data & 0x3; }
	// // Used by road and tram vehicles to indicate side
	// constexpr bool isReversed() const { return _data & (1 << 2); }
	// constexpr void setReversed(bool state)
	// _data &= ~(1 << 2);
	// _data |= state ? (1 << 2) : 0;
	}
	// constexpr uint8_t basicRad() const { return _data & 0x7F; }
	// // Vehicles can be in overtaking lane (trams can stay in that lane)
	// constexpr bool isOvertaking() const { return _data & (1 << 7); }
	// // Related to road vehicles turning around
	// constexpr bool isChangingLane() const { return _data & (1 << 8); }
	// constexpr bool operator==(const _RoadAndDirection other) const { return _data == other._data; }
}
// union
	Track _TrackAndDirection
	Road _RoadAndDirection
}
// func TrackAndDirection(id uint8, direction uint8) constexpr
// : track(id, direction)
// static_assert(sizeof(TrackAndDirection) == 2);
// // TODO move to a different header
type SignalStateFlags int

const (
	None SignalStateFlags = 0
	Occupied SignalStateFlags = 1 << 0
	BlockedNoRoute SignalStateFlags = 1 << 1
	OccupiedOneWay SignalStateFlags = 1 << 2
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(SignalStateFlags);
// func GetMovementNibble(pos1 World::Pos3, pos2 World::Pos3) uint8
// uint8_t nibble = 0;
// if (pos1.x != pos2.x)
// nibble |= (1U << 0);
// if (pos1.y != pos2.y)
// nibble |= (1U << 1);
// if (pos1.z != pos2.z)
// nibble |= (1U << 2);
// orphan member: return nibble;
// // 0x00500120
var MovementNibbleToDistance = [8]uint32{
// 0,
// 0x220C,
// 0x220C,
// 0x3027,
// 0x199A,
// 0x2A99,
// 0x2A99,
// 0x3689,
// // 0x00500244
// constexpr std::array<World::TilePos2, 9> kMooreNeighbourhood = {
// World::TilePos2{ 0, 0 },
// World::TilePos2{ 0, 1 },
// World::TilePos2{ 1, 1 },
// World::TilePos2{ 1, 0 },
// World::TilePos2{ 1, -1 },
// World::TilePos2{ 0, -1 },
// World::TilePos2{ -1, -1 },
// World::TilePos2{ -1, 0 },
// World::TilePos2{ -1, 1 },
// func SetSignalState(loc World::Pos3, trackAndDirection TrackAndDirection::_TrackAndDirection, trackType uint8, flags uint32) 
// func GetSignalState(loc World::Pos3, trackAndDirection TrackAndDirection::_TrackAndDirection, trackType uint8, flags uint32) SignalStateFlags
// func Sub_4A2AD7(loc World::Pos3, trackAndDirection TrackAndDirection::_TrackAndDirection, company CompanyId, trackType uint8) 
// func SetReverseSignalOccupiedInBlock(loc World::Pos3, trackAndDirection TrackAndDirection::_TrackAndDirection, company CompanyId, trackType uint8) 
// func IsBlockOccupied(loc World::Pos3, trackAndDirection TrackAndDirection::_TrackAndDirection, company CompanyId, trackType uint8) bool
// func Sub_4A2A58(loc World::Pos3, trackAndDirection TrackAndDirection::_TrackAndDirection, company CompanyId, trackType uint8) uint8
// func Sub_4A2A77(loc World::Pos3, trackAndDirection TrackAndDirection::_TrackAndDirection, company CompanyId, trackType uint8) uint8
type ApplyTrackModsResult struct {
	Cost currency32_t
	NetworkTooComplex bool
	AllPlacementsFailed bool
}
// func ApplyTrackModsToTrackNetwork(pos World::Pos3, trackAndDirection Vehicles::TrackAndDirection::_TrackAndDirection, company CompanyId, trackType uint8, flags uint8, modSelection World::Track::ModSection, trackModObjIds uint8) ApplyTrackModsResult
// func RemoveTrackModsToTrackNetwork(pos World::Pos3, trackAndDirection Vehicles::TrackAndDirection::_TrackAndDirection, company CompanyId, trackType uint8, flags uint8, modSelection World::Track::ModSection, trackModObjIds uint8) currency32_t
// func ApplyRoadModsToTrackNetwork(pos World::Pos3, roadAndDirection Vehicles::TrackAndDirection::_RoadAndDirection, company CompanyId, roadType uint8, flags uint8, modSelection World::Track::ModSection, roadModObjIds uint8) ApplyTrackModsResult
// func RemoveRoadModsToTrackNetwork(pos World::Pos3, roadAndDirection Vehicles::TrackAndDirection::_RoadAndDirection, company CompanyId, roadType uint8, flags uint8, modSelection World::Track::ModSection, roadModObjIds uint8) currency32_t
// func LeaveLevelCrossing(loc World::Pos3, trackAndDirection TrackAndDirection::_TrackAndDirection, unk uint16) 
type RoadOccupationFlags int

const (
	None RoadOccupationFlags = 0
	IsLaneOccupied RoadOccupationFlags = 1 << 0
	IsLevelCrossingClosed RoadOccupationFlags = 1 << 1
	HasLevelCrossing RoadOccupationFlags = 1 << 2
	HasStation RoadOccupationFlags = 1 << 3
	IsOneWay RoadOccupationFlags = 1 << 4
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(RoadOccupationFlags);
// func GetRoadOccupation(pos World::Pos3, tad TrackAndDirection::_RoadAndDirection) RoadOccupationFlags
// func CheckForCollisions(bogie VehicleBogie, loc World::Pos3) EntityId
// func PlayPickupSound(veh2 Vehicles::Vehicle2) 
// func PlayPlacedownSound(pos World::Pos3) 
type UpdateMotionResult struct {
	RemainingDistance int32
	Flags UpdateVar1136114Flags
	CollidedEntityId EntityId
func (u *UpdateMotionResult) HasFlags(f UpdateVar1136114Flags) bool {
}
}

type UpdateMotionResultVehicleBase struct {
	EntityBase // embedded
	SubType VehicleEntityType
	Mode TransportMode
	VehicleFlags VehicleFlags
	Head EntityId
	TrackAndDirection TrackAndDirection
	SubPosition uint16
	TileX int16
	TileY int16
// World::SmallZ tileBaseZ;
	TrackType uint8
	RemainingDistance int32
	RoutingHandle RoutingHandle
	NextCarId EntityId
	Var_38 Flags38
// template<VehicleEntityType SubType>
	// method: bool is() const
// return subType == SubType;
}
// template<typename TType, VehicleEntityType TClass>
// TType* as() const
// // This can not use reinterpret_cast due to being a const member without considerable more code
// if (!is<TClass>())
// throw Exception::RuntimeError("Malformed vehicle. Incorrect subType!");
}
const UpdateMotionResultBaseType any = EntityBaseType.vehicle
// return (TType*)this;
// template<typename TType>
// TType* as() const
// return as<TType, TType::kVehicleThingType>();
// public:
func GetSubType() VehicleEntityType {
	// void setSubType(const VehicleEntityType newType) { subType = newType; }
	// bool isVehicleHead() const;
	// VehicleHead* asVehicleHead() const;
	// bool isVehicle1() const;
	// Vehicle1* asVehicle1() const;
	// bool isVehicle2() const;
	// Vehicle2* asVehicle2() const;
	// bool isVehicleBogie() const;
	// VehicleBogie* asVehicleBogie() const;
	// bool isVehicleBody() const;
	// VehicleBody* asVehicleBody() const;
	// bool hasSoundPlayer();
	// VehicleSound* getVehicleSound();
	// bool isVehicleTail() const;
	// VehicleTail* asVehicleTail() const;
	// TransportMode getTransportMode() const;
	// Flags38 getFlags38() const;
	// uint8_t getTrackType() const;
	// World::Pos3 getTrackLoc() const;
	// TrackAndDirection getTrackAndDirection() const;
	// RoutingHandle getRoutingHandle() const;
	// EntityId getHead() const;
	// int32_t getRemainingDistance() const;
	// void setNextCar(const EntityId newNextCar);
	// EntityId getNextCar() const;
	// bool has38Flags(Flags38 flagsToTest) const;
	// bool hasVehicleFlags(VehicleFlags flagsToTest) const;
	// VehicleBase* nextVehicle();
	// VehicleBase* nextVehicleComponent();
	// VehicleBase* previousVehicleComponent();
	// void explodeComponent();
	// void destroyTrain();
	// uint8_t updateRoadTileOccupancy(const World::Pos3& loc, const TrackAndDirection::_RoadAndDirection rad, const bool setOccupied);
	// UpdateMotionResult updateTrackMotion(int32_t unk1, bool isVeh2UnkM15);
}
// static_assert(sizeof(VehicleBase) <= sizeof(Entity));
type VehicleSound struct {
	DrivingSoundId SoundObjectId_t
	DrivingSoundVolume uint8
	DrivingSoundFrequency uint16
	ObjectId uint16
	SoundFlags SoundFlags
// Ui::WindowNumber_t soundWindowNumber;
// Ui::WindowType soundWindowType;
}
type VehicleCargo struct {
	AcceptedTypes uint32
	Type uint8
	MaxQty uint8
	TownFrom StationId
	NumDays uint8
	Qty uint8
}
type Sub4ACEE7Result struct {
	Status uint8
	Flags uint8
	StationId StationId
}
type AirplaneApproachTargetParams struct {
// uint16_t targetZ = 0U;                    // 0x01136168
// uint32_t manhattanDistanceToStation = 0U; // 0x011360D0
// uint8_t targetYaw = 0U;                   // 0x0113646D
// bool isHeliTakeOffEnd = false;            // 0x00525BB0
}
type CarUpdateState struct {
// VehicleBogie* frontBogie; // 0x01136124
// VehicleBogie* backBogie;  // 0x01136128
	HasBogieMoved bool
}
type VehicleUpdateDistances struct {
	UnkDistance1 int32
	UnkDistance2 int32
}
// // Don't use outside of vehicle files
// VehicleUpdateDistances& getVehicleUpdateDistances();
// func CalculateYaw0FromVector(xDiff int16, yDiff int16) uint8
// func CalculateYaw1FromVectorPlane(xDiff int16, yDiff int16) uint8
// func CalculateYaw1FromVector(xDiff int16, yDiff int16) uint8
// func CalculateYaw4FromVector(xOffset int16, yOffset int16) uint8
// func Sub_4BA873(vehBogie VehicleBogie) 
// func LiftUpTail(tail VehicleTail) 
type CarComponent struct {
// VehicleBogie* front = nullptr;
// VehicleBogie* back = nullptr;
// VehicleBody* body = nullptr;
// CarComponent(VehicleBase*& component);
// CarComponent() = default;
// template<typename TFunc>
	// method: void applyToComponents(TFunc&& func) const
// func(*front);
// func(*back);
// func(*body);
}
type Car struct {
	Public // embedded
}

type CarCarComponentIter struct {
	Current CarComponent
// VehicleBase* nextVehicleComponent = nullptr;
// CarComponentIter(const CarComponent* carComponent);
	// CarComponentIter& operator++();
	// CarComponentIter operator++(int)
// CarComponentIter retval = *this;
// ++(*this);
	Retval return
}
	// bool operator==(CarComponentIter other) const
// return nextVehicleComponent == other.nextVehicleComponent;
}
// constexpr CarComponent& operator*()
// orphan member: return current;
// // iterator traits
type Difference_type = std::ptrdiff_t
type Value_type = CarComponent
type Pointer = CarComponent
type Reference = CarComponent
type Iterator_category = std::forward_iterator_tag
// func Begin() CarComponentIter
// func CarComponentIter(this) return
// func End() CarComponentIter
// func CarComponentIter(nullptr) return
// Car(VehicleBase*& component)
// : CarComponent(component)
// Car() = default;
// template<typename TFunc>
// func ApplyToComponents(func TFunc) 
// for (auto& carComponent : *this)
// carComponent.applyToComponents(func);
type Vehicle struct {
}

type VehicleCars struct {
	FirstCar Car
}

type VehicleCarIter struct {
	Current Car
// VehicleBase* nextVehicleComponent = nullptr;
// CarIter(const Car* carComponent);
	// CarIter& operator++();
	// CarIter operator++(int)
// CarIter retval = *this;
// ++(*this);
	Retval return
}
	// bool operator==(CarIter other) const
// return nextVehicleComponent == other.nextVehicleComponent;
}
	// constexpr Car& operator*()
	Current return
}
// // iterator traits
type Difference_type = std::ptrdiff_t
type Value_type = Car
type Pointer = Car
type Reference = Car
type Iterator_category = std::forward_iterator_tag
// func Begin() CarIter
// func CarIter(firstCar) return
// func End() CarIter
// func CarIter(nullptr) return
// std::size_t size() const
// if (firstCar.body == nullptr)
// orphan member: return 0;
// return std::distance(begin(), end());
// func Empty() bool
// if (firstCar.body == nullptr)
// orphan member: return true;
// orphan member: return false;
// Cars(Car&& _firstCar)
// : firstCar(_firstCar)
// Cars() = default;
// template<typename TFunc>
// func ApplyToComponents(func TFunc) 
// for (auto& car : *this)
// car.applyToComponents(func);
// VehicleHead* head;
// Vehicle1* veh1;
// Vehicle2* veh2;
// VehicleTail* tail;
// orphan member: Cars cars;
// Vehicle(const VehicleHead& _head);
// Vehicle(EntityId _head);
// // Call if the cars order may have changed
// func RefreshCars() 
// template<typename TFunc>
// func ApplyToComponents(func TFunc) 
// func(*head);
// func(*veh1);
// func(*veh2);
// cars.applyToComponents(func);
// func(*tail);
// // TODO: move this?
// func GetNumUnitsForCargo(maxPrimaryCargo uint32, primaryCargoId uint8, newCargoId uint8) uint32
// func RemoveAllCargo(carComponent CarComponent) 
// /* flipCar
// * Reverses a Car in-place and returns the new front bogie
// * frontBogie: front bogie of the Car
// * returns new front bogie of the Car
// */
// VehicleBogie* flipCar(VehicleBogie& frontBogie);
// /* insertCarBefore
// * Takes source vehicle out of its train and puts it in front of the destination vehicle in the destination train.
// * Source and destination trains can be the same.
// * source: front bogie of the Car to move
// * dest: VehicleBogie or VehicleTail to place Car before
// * returns nothing
// */
// func InsertCarBefore(source VehicleBogie, dest VehicleBase) 
// func CanVehiclesCouple(newVehicleTypeId uint16, sourceVehicleTypeId uint16) bool
// func ConnectJacobsBogies(head VehicleHead) 
// func ApplyVehicleObjectLength(train Vehicle) 
// func PositionVehicleOnTrack(head VehicleHead, isPlaceDown bool) bool
