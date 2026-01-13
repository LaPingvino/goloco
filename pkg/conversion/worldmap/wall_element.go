package worldmap

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Graphics/Colour.h"
// #include "TileElementBase.h"
// #include <OpenLoco/Core/EnumFlags.hpp>
// namespace OpenLoco::World
type EdgeSlope int

const (
	None EdgeSlope = 0
	Upwards EdgeSlope = 1 << 0
	Downwards EdgeSlope = 1 << 1
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(EdgeSlope);
type WallElement struct {
	Public // embedded
	4 uint8
	5 uint8
	6 uint8
	7 uint8
func (w *WallElement) WallObjectId() uint8 {
	// void setWallObjectId(uint8_t id)
	_4 = id
	}
	// uint8_t rotation() const { return _type & 0x3; } // _0
	// void setRotation(uint8_t rotation)
	// _type &= ~0x3;
	// _type |= (rotation & 0x3);
	}
	// EdgeSlope getSlopeFlags() const
	return EdgeSlope((_type & 0xC0) >> 6)
	}
	// void setSlopeFlags(EdgeSlope slopeFlags)
	// _type &= ~0xC0;
	// _type |= (enumValue(slopeFlags) << 6) & 0xC0;
	}
	// void setPrimaryColour(Colour colour)
	_6 = (_6 & 0xE0) | (enumValue(colour) & 0x1F)
	}
	// Colour getPrimaryColour() const
	return Colour(_6 & 0x1F)
	}
	// void setSecondaryColour(Colour colour)
	// _6 &= 0x1F;
	// _6 |= (enumValue(colour) & 0x7) << 5;
	// _flags &= ~0x60; // Reuse flags 5 and 6 for storing bits of the secondary colour
	// _flags |= (enumValue(colour) & 0x18) << 2;
	}
	// Colour getSecondaryColour() const
	return Colour(((_6 >> 5) & 0x7) | ((_flags >> 2) & 0x18))
	}
	// void setTertiaryColour(Colour colour)
	_5 = enumValue(colour)
	}
	// Colour getTertiaryColour() const
	return Colour(_5)
	}
}
// static_assert(sizeof(WallElement) == kTileElementSize);
}
const WallElementElementType ElementType = ElementType.wall
