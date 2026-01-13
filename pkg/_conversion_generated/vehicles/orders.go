package vehicles

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Map/Tile.h"
// #include "Types.hpp"
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <iterator>
// #include <memory>
// namespace OpenLoco::Vehicles
type OrderType int

const (
	End OrderType = iota
	StopAt
	RouteThrough
	RouteWaypoint
	UnloadAll
	WaitFor
)
type OrderFlags int

const (
	None OrderFlags = 0
	IsRoutable OrderFlags = 1 << 0
	HasNumber OrderFlags = 1 << 1
	HasCargo OrderFlags = 1 << 2
	HasStation OrderFlags = 1 << 3
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(OrderFlags)
type Order struct {
// uint8_t _type = 0; // 0x0
// Order() = default;
func (o *Order) GetType() OrderType {
	// void setType(OrderType type)
	// _type &= ~0x7;
	// _type |= static_cast<uint8_t>(type);
	}
	// uint32_t getOffset() const;
	// std::shared_ptr<Order> clone() const;
	// uint64_t getRaw() const;
	// bool hasFlags(const OrderFlags flag) const;
// template<typename T>
	// constexpr bool is() const { return getType() == T::kType; }
// template<typename T>
	// T* as() { return is<T>() ? reinterpret_cast<T*>(this) : nullptr; }
// template<typename T>
	// const T* as() const { return is<T>() ? reinterpret_cast<const T*>(this) : nullptr; }
}
// static_assert(sizeof(Order) == 1, "Size of order must be 1 for pointer arithmetic to work in OrderTableView");
}

type OrderOrderEnd struct {
	Order // embedded
// OrderEnd()
// setType(kType);
}
}
const OrderType OrderType = OrderType.End
type OrderStation struct {
	Order // embedded
// uint8_t _data = 0; // 0x1
	// method: StationId getStation() const
	// method: return StationId(((_type & 0xC0) << 2) | _data);
}
// func SetStation(station StationId) 
// _type &= ~(0xC0);
// _type |= (enumValue(station) >> 2) & 0xC0;
// _data = enumValue(station) & 0xFF;
// func SetFormatArguments(args FormatArguments) 
type OrderStopAt struct {
	OrderStation // embedded
// OrderStopAt(const StationId station)
// setType(kType);
// setStation(station);
}
const OrderStopAtType OrderType = OrderType.StopAt
type OrderRouteThrough struct {
	OrderStation // embedded
// OrderRouteThrough(const StationId station)
// setType(kType);
// setStation(station);
}
const OrderRouteThroughType OrderType = OrderType.RouteThrough
// template<>
// constexpr bool Order::is<OrderStation>() const
// if (is<OrderStopAt>())
// orphan member: return true;
// return is<OrderRouteThrough>();
type OrderRouteWaypoint struct {
	Order // embedded
// uint8_t _data[5] = { 0 }; // 0x1 - 0x6
// OrderRouteWaypoint(const World::TilePos2& pos, const uint8_t baseZ, const uint8_t direction, const uint8_t trackId)
// setType(kType);
// setWaypoint(pos, baseZ);
// setDirection(direction);
// setTrackId(trackId);
}
const OrderRouteWaypointType OrderType = OrderType.RouteWaypoint
// func SetWaypoint(pos World::TilePos2, baseZ uint8) 
// func SetDirection(direction uint8) 
// func SetTrackId(trackId uint8) 
// World::Pos3 getWaypoint() const;
// func GetDirection() uint8
// func GetTrackId() uint8
type OrderCargo struct {
	Order // embedded
func (o *OrderCargo) GetCargo() uint8 {
	// void setCargo(const uint8_t cargo)
	// _type &= ~0xF8;
	// _type |= (cargo & 0x1F) << 3;
	}
	// void setFormatArguments(FormatArguments& args) const;
}
}

type OrderCargoOrderUnloadAll struct {
	OrderCargo // embedded
// OrderUnloadAll(const uint8_t cargo)
// setType(kType);
// setCargo(cargo);
}
}
const OrderCargoType OrderType = OrderType.UnloadAll
type OrderWaitFor struct {
	OrderCargo // embedded
// OrderWaitFor(const uint8_t cargo)
// setType(kType);
// setCargo(cargo);
}
const OrderWaitForType OrderType = OrderType.WaitFor
