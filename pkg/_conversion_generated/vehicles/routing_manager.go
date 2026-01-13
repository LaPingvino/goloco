package vehicles

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Routing.h"
// #include <iterator>
// #include <optional>
// #include <string>
// #include <utility>
// #include <vector>
// namespace OpenLoco::Vehicles::RoutingManager
const AllocatedButFreeRouting uint16 = 0xFFFEU
const RoutingNull uint16 = 0xFFFFU
// std::optional<RoutingHandle> getAndAllocateFreeRoutingHandle();
// func FreeRoutingHandle(handle RoutingHandle) 
// // Returns a routing. Each routing represents a track/road piece that a train is on or has reserved
// // See OpenLoco::World::Track::AdditionalTadFlags for bits of the routing
// func GetRouting(handle RoutingHandle) uint16
// func SetRouting(handle RoutingHandle, routing uint16) 
// func FreeRouting(handle RoutingHandle) 
// // Equivalent of calling freeRouting on all routings for a single vehicle
// func ResetRoutings(handle RoutingHandle) 
// func IsEmptyRoutingSlotAvailable() bool
// func ResetRoutingTable() 
type RingView struct {
}

type RingViewIterator struct {
type Direction int

const (
	Forward Direction = iota
	Reverse
)
	Current RoutingHandle
// bool _hasLooped = false;
// bool _isEnd = false;
// Direction _direction = Direction::forward;
// Iterator(const RoutingHandle& begin, bool isEnd, Direction direction);
	// Iterator& operator++();
	// Iterator operator++(int)
// Iterator res = *this;
// ++(*this);
	Res return
}
	// Iterator& operator--();
	// Iterator operator--(int)
// Iterator res = *this;
// --(*this);
	Res return
}
// bool operator==(const Iterator& other) const;
// RoutingHandle& operator*()
// orphan member: return _current;
// const RoutingHandle& operator*() const
// orphan member: return _current;
// RoutingHandle& operator->()
// orphan member: return _current;
// const RoutingHandle& operator->() const
// orphan member: return _current;
// // iterator traits
type Difference_type = std::ptrdiff_t
type Value_type = RoutingHandle
type Pointer = RoutingHandle
type Reference = RoutingHandle
type Iterator_category = std::bidirectional_iterator_tag
// orphan member: RoutingHandle _begin;
// public:
// // currentOrderOffset is relative to beginTableOffset and is where the ring will begin and end
// RingView(const RoutingHandle begin)
// : _begin(begin)
// RingView::Iterator begin() const { return Iterator(_begin, false, Iterator::Direction::forward); }
// RingView::Iterator end() const { return Iterator(_begin, true, Iterator::Direction::forward); }
func Rbegin() any {
	// auto rend() const { return Iterator(_begin, true, Iterator::Direction::reverse); }
}
