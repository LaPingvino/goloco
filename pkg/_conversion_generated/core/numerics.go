package core

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Numerics.hpp"
// #include <cstddef>
// #include <cstdint>
// #include <limits.h>
// #include <type_traits>
// #include <intrin.h>
// namespace OpenLoco::Numerics
// // Finds the first bit set in a 32-bits numeral and returns its index, or -1 if no bit is set.
// func BitScanForward(source uint32) int32
// unsigned long i;
// uint8_t success = _BitScanForward(&i, source);
// return success != 0 ? i : -1;
// #elif defined(__GNUC__)
// auto result = __builtin_ffs(source);
// return result - 1;
// #else
// if (source != 0)
// for (int32_t i = 0; i < 32; i++)
// if (source & (1u << i))
// orphan member: return i;
// return -1;
// func BitScanReverse(source uint32) int32
// unsigned long i;
// uint8_t success = _BitScanReverse(&i, source);
// return success != 0 ? i : -1;
// #elif defined(__GNUC__)
// auto result = source == 0 ? -1 : __builtin_clz(source) ^ 31;
// orphan member: return result;
// #else
// if (source != 0)
// for (int32_t i = 31; i > -1; i--)
// if (source & (1u << i))
// orphan member: return i;
// return -1;
