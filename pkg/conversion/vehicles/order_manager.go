package vehicles

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "LabelFrame.h"
// #include "Types.hpp"
// #include "Vehicles/Orders.h"
// #include <OpenLoco/Engine/World.hpp>
// #include <cstdint>
// #include <span>
// #include <string>
// namespace OpenLoco::Vehicles
// forward: struct VehicleHead;
type OrderRingView struct {
}

type OrderRingViewIterator struct {
// Order* _beginOrderTable;
// Order* _currentOrder;
// bool _hasLooped = false;
// Iterator(Order* beginOrderTable, Order* currentOrder)
// : _beginOrderTable(beginOrderTable)
// , _currentOrder(currentOrder)
// // Prevent empty tables looping
// if (_currentOrder->getType() == OrderType::End)
// _hasLooped = true;
}
}
// Iterator& operator++();
// Iterator operator++(int)
// Iterator res = *this;
// ++(*this);
// orphan member: return res;
// Iterator operator+(int32_t amount) const
// Iterator res = *this;
// while (amount-- != 0)
// res++;
// orphan member: return res;
// bool operator==(Iterator other) const
// return _currentOrder == other._currentOrder && (_hasLooped || other._hasLooped);
// Order& operator*()
// return *_currentOrder;
// const Order& operator*() const
// return *_currentOrder;
// Order* operator->()
// orphan member: return _currentOrder;
// const Order* operator->() const
// orphan member: return _currentOrder;
// // iterator traits
type Difference_type = std::ptrdiff_t
type Value_type = Order
type Pointer = Order
type Reference = Order
type Iterator_category = std::forward_iterator_tag
// private:
// orphan member: uint32_t _beginTableOffset;
// orphan member: uint32_t _currentOrderOffset;
// public:
// // currentOrderOffset is relative to beginTableOffset and is where the ring will begin and end
// OrderRingView(uint32_t beginTableOffset, uint32_t currentOrderOffset = 0)
// : _beginTableOffset(beginTableOffset)
// , _currentOrderOffset(currentOrderOffset)
// OrderRingView::Iterator begin() const;
// OrderRingView::Iterator end() const;
// Order* atIndex(const uint8_t index) const;
// namespace OpenLoco::Vehicles::OrderManager
type NumDisplayFrame struct {
	OrderOffset uint32
	Frame LabelFrame
	LineNumber uint8
}
// Order* orders();
// uint32_t& orderTableLength();
// func ShiftOrdersLeft(offsetToShiftTowards uint32, sizeToShiftBy int16) 
// func ShiftOrdersRight(offsetToShiftFrom uint32, sizeToShiftBy int16) 
// func ReoffsetVehicleOrderTables(removeOrderTableOffset uint32, sizeOfRemovedOrderTable int16) 
// func SpaceLeftInGlobalOrderTableForOrder(order Order) bool
// func SpaceLeftInVehicleOrderTable(head VehicleHead) bool
// func InsertOrder(head VehicleHead, orderOffset uint16, order Order) 
// func DeleteOrder(head VehicleHead, orderOffset uint16) 
// func ZeroUnusedOrderTable() 
// func Reset() 
// func FreeOrders(head VehicleHead) 
// func AllocateOrders(head VehicleHead) 
// std::pair<World::Pos3, std::string> generateOrderUiStringAndLoc(uint32_t orderOffset, uint8_t orderNum);
// func GenerateNumDisplayFrames(head Vehicles::VehicleHead) 
// std::span<const NumDisplayFrame> displayFrames();
// func ReverseVehicleOrderTable(tableOffset uint16, orderOfInterest uint16) uint16
// func SwapAdjacentOrders(a Order, b Order) uint8
// func RemoveOrdersForStation(stationId StationId) 
// func FixCorruptWaypointOrders() 
