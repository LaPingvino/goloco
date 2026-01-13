package math

import "math"

// Bounded addition and subtraction helpers for small integer types.
// These provide behavior similar to the C++ templates in Bound.hpp: additions
// saturate at the type limits to avoid overflow/underflow.

func AddUint8(lhs uint8, rhs int64) uint8 {
	s := int64(lhs) + rhs
	if s > int64(math.MaxUint8) {
		return uint8(math.MaxUint8)
	}
	if s < 0 {
		return 0
	}
	return uint8(s)
}

func SubUint8(lhs uint8, rhs int64) uint8 {
	s := int64(lhs) - rhs
	if s > int64(math.MaxUint8) {
		return uint8(math.MaxUint8)
	}
	if s < 0 {
		return 0
	}
	return uint8(s)
}

func AddInt8(lhs int8, rhs int64) int8 {
	s := int64(lhs) + rhs
	if s > int64(math.MaxInt8) {
		return int8(math.MaxInt8)
	}
	if s < int64(math.MinInt8) {
		return int8(math.MinInt8)
	}
	return int8(s)
}

func SubInt8(lhs int8, rhs int64) int8 {
	s := int64(lhs) - rhs
	if s > int64(math.MaxInt8) {
		return int8(math.MaxInt8)
	}
	if s < int64(math.MinInt8) {
		return int8(math.MinInt8)
	}
	return int8(s)
}

func AddUint16(lhs uint16, rhs int64) uint16 {
	s := int64(lhs) + rhs
	if s > int64(math.MaxUint16) {
		return uint16(math.MaxUint16)
	}
	if s < 0 {
		return 0
	}
	return uint16(s)
}

func SubUint16(lhs uint16, rhs int64) uint16 {
	s := int64(lhs) - rhs
	if s > int64(math.MaxUint16) {
		return uint16(math.MaxUint16)
	}
	if s < 0 {
		return 0
	}
	return uint16(s)
}

func AddInt16(lhs int16, rhs int64) int16 {
	s := int64(lhs) + rhs
	if s > int64(math.MaxInt16) {
		return int16(math.MaxInt16)
	}
	if s < int64(math.MinInt16) {
		return int16(math.MinInt16)
	}
	return int16(s)
}

func SubInt16(lhs int16, rhs int64) int16 {
	s := int64(lhs) - rhs
	if s > int64(math.MaxInt16) {
		return int16(math.MaxInt16)
	}
	if s < int64(math.MinInt16) {
		return int16(math.MinInt16)
	}
	return int16(s)
}

func AddUint32(lhs uint32, rhs int64) uint32 {
	s := int64(lhs) + rhs
	if s > int64(math.MaxUint32) {
		return uint32(math.MaxUint32)
	}
	if s < 0 {
		return 0
	}
	return uint32(s)
}

func SubUint32(lhs uint32, rhs int64) uint32 {
	s := int64(lhs) - rhs
	if s > int64(math.MaxUint32) {
		return uint32(math.MaxUint32)
	}
	if s < 0 {
		return 0
	}
	return uint32(s)
}

func AddInt32(lhs int32, rhs int64) int32 {
	s := int64(lhs) + rhs
	if s > int64(math.MaxInt32) {
		return int32(math.MaxInt32)
	}
	if s < int64(math.MinInt32) {
		return int32(math.MinInt32)
	}
	return int32(s)
}

func SubInt32(lhs int32, rhs int64) int32 {
	s := int64(lhs) - rhs
	if s > int64(math.MaxInt32) {
		return int32(math.MaxInt32)
	}
	if s < int64(math.MinInt32) {
		return int32(math.MinInt32)
	}
	return int32(s)
}
