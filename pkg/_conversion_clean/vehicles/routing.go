package vehicles

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Engine/Limits.h"
// #include <cstdint>
// namespace OpenLoco::Vehicles
type RoutingHandle struct {
	Data uint16
	// method: constexpr RoutingHandle(const uint16 vehicleRef, const uint8 index)
// : _data((vehicleRef * Limits::kMaxRoutingsPerVehicle) | index)
}
func GetVehicleRef() uint16 {
	// constexpr uint8 getIndex() const { return _data % Limits::kMaxRoutingsPerVehicle; }
	// constexpr void setIndex(uint8 newIndex)
	// _data &= ~(Limits::kMaxRoutingsPerVehicle - 1);
	// _data |= newIndex & (Limits::kMaxRoutingsPerVehicle - 1);
	operator bool = =(const RoutingHandle other) const { return _data == other._data
// static_assert(sizeof(RoutingHandle) == 2);