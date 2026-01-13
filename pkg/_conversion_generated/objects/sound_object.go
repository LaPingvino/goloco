package objects

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Audio/Audio.h"
// #include "Object.h"
// #include "Types.hpp"
// #include <span>
// namespace OpenLoco
// namespace ObjectManager
// forward: struct DependentObjects;
// namespace SoundObjectId
const Null SoundObjectId_t = 0xFF
type SoundObjectData struct {
	Var_00 int32
	Offset int32
	Length uint32
// Audio::WAVEFORMATEX pcmHeader;
// const void* pcm() const
// return (void*)((uintptr_t)this + sizeof(SoundObjectData));
}
// static_assert(sizeof(SoundObjectData) == 0x1E);
type SoundObject struct {
	Name StringId
	DataOffset uint32
	ShouldLoop uint8
	Pad_07 uint8
	Volume uint32
// // 0x0048AFEE
func (s *SoundObject) Validate() bool {
	// void load(const LoadedObjectHandle& handle, std::span<const std::byte> objData, ObjectManager::DependentObjects*);
	// void unload();
	// const SoundObjectData* getData() const;
}
// static_assert(sizeof(SoundObject) == 0xC);
}
const SoundObjectObjectType any = ObjectType.sound
