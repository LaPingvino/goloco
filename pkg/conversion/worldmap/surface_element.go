package worldmap

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "TileElementBase.h"
// namespace OpenLoco::World
// namespace SurfaceSlope
const Flat uint8 = 0x00
// namespace CornerUp
const All uint8 = 0x0F
const North uint8 = (1 << 0)
const East uint8 = (1 << 1)
const South uint8 = (1 << 2)
const West uint8 = (1 << 3)
const DoubleHeight uint8 = (1 << 4)
const RequiresHeightAdjustment uint8 = (1 << 5)
// namespace CornerDown
const West uint8 = CornerUp.all & ~CornerUp.west
const South uint8 = CornerUp.all & ~CornerUp.south
const East uint8 = CornerUp.all & ~CornerUp.east
const North uint8 = CornerUp.all & ~CornerUp.north
// namespace SideUp
const Northeast uint8 = CornerUp.north | CornerUp.east
const Southeast uint8 = CornerUp.south | CornerUp.east
const Northwest uint8 = CornerUp.north | CornerUp.west
const Southwest uint8 = CornerUp.south | CornerUp.west
// namespace Valley
const Westeast uint8 = CornerUp.east | CornerUp.west
const Northsouth uint8 = CornerUp.north | CornerUp.south
type SurfaceElement struct {
	Public // embedded
	Slope uint8
	Water uint8
	Terrain uint8
	7 uint8
// SurfaceElement() = default;
// SurfaceElement(World::SmallZ baseZ, World::SmallZ clearZ, uint8_t quarterTile, bool highTypeFlag)
// setType(kElementType);
// setBaseZ(baseZ);
// setClearZ(clearZ);
// _flags = quarterTile;
// setIsIndustrialFlag(highTypeFlag);
}
const SurfaceElementElementType ElementType = ElementType.surface
func IsSlopeDoubleHeight() bool {
	// uint8_t slopeCorners() const { return _slope & 0x0F; }
	// uint8_t slope() const { return _slope & 0x1F; }
	// void setSlope(uint8_t slope)
	// _slope &= ~0x1F;
	_slope = slope & 0x1F
	}
	// uint8_t snowCoverage() const { return (_slope & 0xE0) >> 5; }
	// void setSnowCoverage(uint8_t coverage)
	// _slope &= ~0xE0;
	// _slope |= coverage << 5;
	}
	// MicroZ water() const { return _water & 0x1F; }
	// int16_t waterHeight() const { return (_water & 0x1F) * kMicroZStep; }
	// void setWater(MicroZ level) { _water = (_water & 0xE0) | (level & 0x1F); };
	// uint8_t getUpdateTimer() const
	return (_water & 0xE0) >> 5
	}
	// void setUpdateTimer(uint8_t var5) { _water = (_water & 0x1F) | ((var5 << 5) & 0xE0); }
	// uint8_t terrain() const { return _terrain & 0x1F; }
	// void setTerrain(uint8_t terrain)
	// _terrain &= ~0x1F;
	// _terrain |= terrain & 0x1F;
	}
	// uint8_t getGrowthStage() const { return _terrain >> 5; }
	// void setGrowthStage(uint8_t var6)
	// _terrain &= 0x1F;
	// _terrain |= var6 << 5;
	}
	// IndustryId industryId() const { return IndustryId(_7); }
	// uint8_t variation() const { return _7; }
	// void setIndustry(const IndustryId industry) { _7 = enumValue(industry); }
	// void setVariation(const uint8_t variation) { _7 = variation; }
	// void setType6Flag(bool state)
	// _type &= ~0x40;
	// _type |= state ? 0x40 : 0;
	}
	// bool hasType6Flag() const { return _type & 0x40; }
	// // Note: Also used for other means for boats
	// bool isIndustrial() const { return _type & 0x80; }
	// // Note: Also used for other means for boats
	// void setIsIndustrialFlag(bool state)
	// _type &= ~0x80;
	// _type |= state ? 0x80 : 0;
	}
	// void removeIndustry(const World::Pos2& pos);
}
// static_assert(sizeof(SurfaceElement) == kTileElementSize);
// func UpdateSurface(elSurface SurfaceElement, loc World::Pos2) bool
