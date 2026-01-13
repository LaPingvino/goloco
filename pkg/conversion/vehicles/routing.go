package vehicles

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Engine/Limits.h"
// #include <cstdint>
// namespace OpenLoco::Vehicles
type RoutingHandle struct {
	Data uint16
	// method: constexpr RoutingHandle(const uint16_t vehicleRef, const uint8_t index)
// : _data((vehicleRef * Limits::kMaxRoutingsPerVehicle) | index)
}
func GetVehicleRef() uint16 {
	// constexpr uint8_t getIndex() const { return _data % Limits::kMaxRoutingsPerVehicle; }
	// constexpr void setIndex(uint8_t newIndex)
	// _data &= ~(Limits::kMaxRoutingsPerVehicle - 1);
	// _data |= newIndex & (Limits::kMaxRoutingsPerVehicle - 1);
	}
	var operator bool = =(const RoutingHandle other) const { return _data == other._data
}
// static_assert(sizeof(RoutingHandle) == 2);
