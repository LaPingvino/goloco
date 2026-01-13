package core

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Prng.h"
// #include <bit>
// namespace OpenLoco::Core
// uint32 Prng::randNext()
// auto srand0 = _srand_0;
// auto srand1 = _srand_1;
// _srand_0 += std::rotr<uint32>(srand1 ^ 0x1234567F, 7);
// _srand_1 = std::rotr<uint32>(srand0, 3);
// orphan member: return _srand_1;