package vehicles

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Vehicle.h"
// namespace OpenLoco::Vehicles
// SKIPPED C++ SYNTAX: type CargoTotalArray = any /* []<uint32, Limits::kMaxCargoObjects> */ 
type Status int

const (
func init() {
	Unk_0 Status = 0
	Stopped Status = 1
	Travelling Status = 2
	WaitingAtSignal Status = 3
	Approaching Status = 4
	Unloading Status = 5
	Loading Status = 6
	BrokenDown Status = 7
	Crashed Status = 8
	Stuck Status = 9
	Landing Status = 10
	Taxiing1 Status = 11
	Taxiing2 Status = 12
	TakingOff Status = 13
}

)
type WaterMotionFlags int

func init() {
	None WaterMotionFlags = 0
	IsStopping WaterMotionFlags = 1 << 0
	IsLeavingDock WaterMotionFlags = 1 << 1
	HasReachedDock WaterMotionFlags = 1 << 16
	HasReachedADestination WaterMotionFlags = 1 << 17

// OPENLOCO_ENABLE_ENUM_OPERATORS(WaterMotionFlags);
type SignalTimeoutStatus int

func init() {
	Ok SignalTimeoutStatus = 0
	FirstTimeout SignalTimeoutStatus = 1
	TurnaroundAtSignalTimeout SignalTimeoutStatus = 2

type VehicleStatus struct {
	Status1 StringId
	Status1Args uint32
	Status2 StringId
	Status2Args uint32
type VehicleHead struct {
	VehicleBase // embedded
	Var_3C int32
	OrderTableOffset uint32
	CurrentOrder uint16
	SizeOfOrderTable uint16
	TrainAcceptedCargoTypes uint32
	OrdinalNumber int16
	Var_52 uint8
	Var_53 uint8
	StationId StationId
	CargoTransferTimeout uint16
	Var_58 uint32
	Var_5C uint8
	Status Status
	VehicleType VehicleType
	BreakdownFlags BreakdownFlags
// World::Pos2 aiPlacementPos;
	AiThoughtId uint8
	AirportMovementEdge uint8
	AiPlacementTaD uint16
	AiPlacementBaseZ uint8
	CrashedTimeout uint8
	LastAverageSpeed Speed16
	TotalRefundCost uint32
// World::Pos2 journeyStartPos;       //  journey start position
	JourneyStartTicks uint32
	ManualPower int8
	RestartStoppedCarsTimeout uint8
	// method: bool isVehicleTypeCompatible(const uint16 vehicleTypeId);
	// method: void updateBreakdown();
	// method: void updateVehicle();
	// method: bool update();
	// method: void updateMonthly();
	// method: void updateDaily();
	// method: VehicleStatus getStatus() const;
	// method: OrderRingView getCurrentOrders() const;
func (v *VehicleHead) IsPlaced() bool {
	// bool hasAnyCargo();
	// char* generateCargoTotalString(char* buffer);
	// char* generateCargoCapacityString(char* buffer);
	// char* cargoLUTToString(CargoTotalArray& cargoTotals, char* buffer);
	// bool canBeModified() const;
	// void liftUpVehicle();
	// void updateTrainProperties();
	// currency32_t calculateRunningCost() const;
	// void sub_4AD778();
	// void sub_4AD93A();
	// void sub_4ADB47(bool unk);
	// uint32 getCarCount() const;
	// void applyBreakdownToTrain();
	// void landCrashedUpdate();
	// void autoLayoutTrain();
	// uint32 getVehicleTotalLength() const;
	// constexpr bool hasBreakdownFlags(BreakdownFlags flagsToTest) const
	return (breakdownFlags & flagsToTest) != BreakdownFlags.none
	// void movePlaneTo(const World::Pos3& newLoc, const uint8 newYaw, const Pitch newPitch);
	// void moveBoatTo(const World::Pos3& loc, const uint8 yaw, const Pitch pitch);
	// Sub4ACEE7Result sub_4ACEE7(uint32 unk1, uint32 var_113612C, bool isPlaceDown);
	// private:
	// void updateDrivingSounds();
	// void updateDrivingSound(VehicleSound& sound, const bool isVeh2);
	// void updateDrivingSoundNone(VehicleSound& sound);
	// void updateDrivingSoundFriction(VehicleSound& sound, const VehicleObjectFrictionSound* snd);
	// void updateSimpleMotorSound(VehicleSound& sound, const bool isVeh2, const VehicleSimpleMotorSound* snd);
	// void updateGearboxMotorSound(VehicleSound& sound, const bool isVeh2, const VehicleGearboxMotorSound* snd);
	// bool updateLand();
	// bool sub_4A8DB7();
	// bool tryReverse();
	// bool stoppingUpdate();
	// bool sub_4A8C81();
	// bool landTryBeginUnloading();
	// bool landLoadingUpdate();
	// bool landNormalMovementUpdate();
	// bool trainNormalMovementUpdate(uint8 al, uint8 flags, StationId nextStation);
	// bool roadNormalMovementUpdate(uint8 al, StationId nextStation);
	// bool landReverseFromSignal();
	// bool updateAir();
	// bool airplaneLoadingUpdate();
	// bool sub_4A95CB();
	// bool sub_4A9348(uint8 newMovementEdge, const AirplaneApproachTargetParams& approachParams);
	// bool airplaneApproachTarget(const AirplaneApproachTargetParams& params);
	// std::pair<Status, Speed16> airplaneGetNewStatus();
	// uint8 airportGetNextMovementEdge(uint8 curEdge);
	// AirplaneApproachTargetParams sub_427122();
	// bool updateWater();
	// void tryCreateInitialMovementSound(const Status initialStatus);
	// void setStationVisitedTypes();
	// void checkIfAtOrderStation();
	// void updateLastJourneyAverageSpeed();
	// void beginUnloading();
	// void beginLoading();
	// WaterMotionFlags updateWaterMotion(WaterMotionFlags flags);
	// uint8 getLoadingModifier(const VehicleBogie* bogie);
	// bool updateUnloadCargoComponent(VehicleCargo& cargo, VehicleBogie* bogie);
	// void updateUnloadCargo();
	// bool updateLoadCargoComponent(VehicleCargo& cargo, VehicleBogie* bogie);
	// bool updateLoadCargo();
	// void beginNewJourney();
	// void advanceToNextRoutableOrder();
	// Status sub_427BF2();
	// void produceLeavingDockSound();
	// void produceTouchdownAirportSound();
	// SignalTimeoutStatus categoriseTimeElapsed();
	// bool sub_4AC1C2();
	// bool opposingTrainAtSignal();
	// bool pathingShouldReverse();
	// StationId manualFindTrainStationAtLocation();
	// bool isOnExpectedRoadOrTrack();
	// VehicleStatus getStatusTravelling() const;
	// void getSecondStatus(VehicleStatus& vehStatus) const;
	// void updateLastIncomeStats(uint8 cargoType, uint16 cargoQty, uint16 cargoDist, uint8 cargoAge, currency32_t profit);
	// void calculateRefundCost();
// static_assert(sizeof(VehicleHead) <= sizeof(Entity));
const VehicleHeadVehicleThingType any = VehicleEntityType.head