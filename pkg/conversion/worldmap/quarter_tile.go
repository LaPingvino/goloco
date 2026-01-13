package worldmap

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include <cstdint>
// namespace OpenLoco::World
type QuarterTile struct {
	Val uint8
// explicit constexpr QuarterTile(uint8_t tileQuarter, uint8_t zQuarter)
// : _val(tileQuarter | (zQuarter << 4))
}
// explicit constexpr QuarterTile(uint8_t tileAndZQuarter)
// : _val(tileAndZQuarter)
// // Rotate both of the values amount. Returns new RValue QuarterTile
// constexpr const QuarterTile rotate(uint8_t amount) const
// switch (amount)
// default:
// orphan member: return QuarterTile{ *this };
// case 1:
// auto rotVal1 = _val << 1;
// auto rotVal2 = rotVal1 >> 4;
// // Clear the bit from the tileQuarter
// rotVal1 &= 0b11101110;
// // Clear the bit from the zQuarter
// rotVal2 &= 0b00010001;
// orphan member: return QuarterTile{ static_cast<uint8_t>(rotVal1 | rotVal2) };
// case 2:
// auto rotVal1 = _val << 2;
// auto rotVal2 = rotVal1 >> 4;
// // Clear the bit from the tileQuarter
// rotVal1 &= 0b11001100;
// // Clear the bit from the zQuarter
// rotVal2 &= 0b00110011;
// orphan member: return QuarterTile{ static_cast<uint8_t>(rotVal1 | rotVal2) };
// case 3:
// auto rotVal1 = _val << 3;
// auto rotVal2 = rotVal1 >> 4;
// // Clear the bit from the tileQuarter
// rotVal1 &= 0b10001000;
// // Clear the bit from the zQuarter
// rotVal2 &= 0b01110111;
// orphan member: return QuarterTile{ static_cast<uint8_t>(rotVal1 | rotVal2) };
// func GetBaseQuarterOccupied() uint8
// return _val & 0xF;
// func GetZQuarterOccupied() uint8
// return (_val >> 4) & 0xF;
// static_assert(sizeof(QuarterTile) == 1);
