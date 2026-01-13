package worldmap

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Types.hpp"
// #include <OpenLoco/Engine/World.hpp>
// #include <cassert>
// #include <span>
// namespace OpenLoco::World
// forward: struct TileElement;
type ElementType int

const (
	Surface ElementType = iota
	Track
	Station
	Signal
	Building
	Tree
	Wall
	Road
	Industry
)
const TileElementSize int = 8
// namespace ElementFlags
const Ghost uint8 = 1 << 4
const AiAllocated uint8 = 1 << 5
const Flag_6 uint8 = 1 << 6
const Last uint8 = 1 << 7
type TileElementBase struct {
	Type uint8
	Flags uint8
	BaseZ uint8
	ClearZ uint8
// // Temporary, use this to get fields easily before they are defined
// const uint8_t* data() const;
	// method: ElementType type() const;
	// method: void setType(ElementType t)
// // Purposely clobbers any other data in _type
// _type = enumValue(t) << 2;
}
func Flags() uint8 {
	// SmallZ baseZ() const { return _baseZ; }
	// int16_t baseHeight() const { return _baseZ * kSmallZStep; }
	// SmallZ clearZ() const { return _clearZ; }
	// int16_t clearHeight() const { return _clearZ * kSmallZStep; }
	// uint8_t occupiedQuarter() const { return _flags & 0xF; }
	// void setOccupiedQuarter(uint8_t val)
	// _flags &= ~0xF;
	// _flags |= val & 0xF;
	}
	// bool isGhost() const { return _flags & ElementFlags::ghost; }
	// void setGhost(bool state)
	// _flags &= ~ElementFlags::ghost;
	// _flags |= state == true ? ElementFlags::ghost : 0;
	}
	// bool isAiAllocated() const { return _flags & ElementFlags::aiAllocated; }
	// void setAiAllocated(bool state)
	// _flags &= ~ElementFlags::aiAllocated;
	// _flags |= state == true ? ElementFlags::aiAllocated : 0;
	}
	// bool isFlag6() const { return _flags & ElementFlags::flag_6; } // in tracks/roads indicates is last tile of multi tile
	// void setFlag6(bool state)
	// _flags &= ~ElementFlags::flag_6;
	// _flags |= state == true ? ElementFlags::flag_6 : 0;
	}
	// void setBaseZ(uint8_t baseZ) { _baseZ = baseZ; }
	// void setClearZ(uint8_t value) { _clearZ = value; }
	// bool isLast() const;
	// void setLastFlag(bool state)
	// _flags &= ~ElementFlags::last;
	// _flags |= state == true ? ElementFlags::last : 0;
	}
	// std::span<uint8_t> rawData()
	return std.span{ /* unsafe cast */, kTileElementSize }
	}
	// std::span<const uint8_t> rawData() const
	return std.span{ /* unsafe cast */, kTileElementSize }
	}
// template<typename TType>
	// const TType* as() const
	return type() == TType.kElementType ? /* unsafe cast */ : nil
	}
// template<typename TType>
	// TType* as()
	return type() == TType.kElementType ? /* unsafe cast */ : nil
	}
// template<typename TType>
	// const TType& get() const
	// assert(type() == TType::kElementType);
	return */* unsafe cast */
	}
// template<typename TType>
	// TType& get()
	// assert(type() == TType::kElementType);
	return */* unsafe cast */
	}
	// const TileElement* prev() const
	return /* unsafe cast */ - kTileElementSize)
	}
	// TileElement* prev()
	return /* unsafe cast */ - kTileElementSize)
	}
	// const TileElement* next() const
	return /* unsafe cast */ + kTileElementSize)
	}
	// TileElement* next()
	return /* unsafe cast */ + kTileElementSize)
	}
}
