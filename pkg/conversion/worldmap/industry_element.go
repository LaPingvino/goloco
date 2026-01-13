package worldmap

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "TileElementBase.h"
// namespace OpenLoco
// forward: struct Industry;
// namespace OpenLoco::World
type IndustryElement struct {
	Public // embedded
	IndustryId IndustryId
	5 uint8
	6 uint16
// // _4
func (i *IndustryElement) IndustryId() IndustryId {
	// void setIndustryId(const IndustryId id) { _industryId = id; }
	// Industry* industry() const;
	// // var_6_07C0
	// uint8_t buildingType() const;
	// void setBuildingType(uint8_t type)
	// _6 &= ~0x7C0;
	// _6 |= type << 6;
	}
	// uint8_t rotation() const { return _type & 0x3; }
	// void setRotation(const uint8_t rotation)
	// _type &= ~0x3;
	// _type |= rotation & 0x3;
	}
	// // var_5_03
	// uint8_t sequenceIndex() const;
	// void setSequenceIndex(const uint8_t index)
	// _5 &= ~0x3;
	// _5 |= index & 0x3;
	}
	// // var_5_E0
	// uint8_t sectionProgress() const;
	// void setSectionProgress(uint8_t val);
	// Colour var_6_F800() const;
	// void setColour(Colour c)
	// _6 &= ~0xF800;
	// _6 |= enumValue(c) << 11;
	}
	// // This has two uses. When under construction it is the number of completed sections. Otherwise its animation sequence related
	// uint8_t var_6_003F() const;
	// void setVar_6_003F(uint8_t val);
	// bool isConstructed() const { return _type & 0x80; }
	// void setIsConstructed(bool val);
	// bool update(const World::Pos2& loc);
}
// static_assert(sizeof(IndustryElement) == kTileElementSize);
// forward: struct Animation;
	// method: bool updateIndustryAnimation1(const Animation& anim);
	// method: bool updateIndustryAnimation2(const Animation& anim);
}
const IndustryElementElementType ElementType = ElementType.industry
