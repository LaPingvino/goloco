package objects

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Object.h"
// #include "Types.hpp"
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <span>
// #include <utility>
// namespace OpenLoco
// namespace ObjectManager
// forward: struct DependentObjects;
type SteamObjectFlags int

const (
	None SteamObjectFlags = 0
	ApplyWind SteamObjectFlags = 1 << 0
	DisperseOnCollision SteamObjectFlags = 1 << 1
	HasTunnelSounds SteamObjectFlags = 1 << 2
	Unk3 SteamObjectFlags = 1 << 3
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(SteamObjectFlags);
type SteamObject struct {
}

type SteamObjectImageAndHeight struct {
	ImageOffset uint8
	Height uint8
}
	Name StringId
	NumImages uint16
	NumStationaryTicks uint8
	SpriteWidth uint8
	SpriteHeightNegative uint8
	SpriteHeightPositive uint8
	Flags SteamObjectFlags
	Var_0A uint32
	BaseImageId uint32
	TotalNumFramesType0 uint16
	TotalNumFramesType1 uint16
	FrameInfoType0Offset uint32
	FrameInfoType1Offset uint32
	NumSoundEffects uint8
// SoundObjectId_t soundEffects[9]; // 0x1F, size tbc, see https://github.com/OpenLoco/OpenLoco/pull/2405#discussion_r1555206249
// // 0x00440DDE
func (s *SteamObject) Validate() bool {
	// void load(const LoadedObjectHandle& handle, std::span<const std::byte> data, ObjectManager::DependentObjects* dependencies);
	// void unload();
	// constexpr bool hasFlags(SteamObjectFlags flagsToTest) const
	return (flags & flagsToTest) != SteamObjectFlags.none
	}
	// std::pair<uint16_t, const ImageAndHeight*> getFramesInfo(bool isType1) const
	// const auto* base = reinterpret_cast<const std::byte*>(this);
	if !isType1 {
	return { totalNumFramesType0, /* unsafe cast */ }
	}
	} else {
	return { totalNumFramesType1, /* unsafe cast */ }
	}
	}
	// };
	// static_assert(sizeof(SteamObject) == 0x28);
}
