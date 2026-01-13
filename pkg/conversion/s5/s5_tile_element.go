package s5

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include <cstdint>
// namespace OpenLoco::S5
type TileElement struct {
	Type uint8
	Flags uint8
	BaseZ uint8
	ClearZ uint8
// uint8_t pad_4[4];
	// method: void setLast(bool value)
// if (value)
// flags |= FLAG_LAST;
}
const TileElementFLAG_GHOST uint8 = 1 << 4
const TileElementFLAG_LAST uint8 = 1 << 7
// else
// flags &= ~FLAG_LAST;
// func IsGhost() bool
// return flags & FLAG_GHOST;
// func IsLast() bool
// return flags & FLAG_LAST;
// static_assert(sizeof(TileElement) == 8);
