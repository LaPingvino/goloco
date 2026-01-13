package audio

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Location.hpp"
// #include "Types.hpp"
// #include <OpenLoco/Engine/World.hpp>
// #include <optional>
// #include <string>
// #include <vector>
// namespace OpenLoco::Vehicles
// forward: struct VehicleSoundPlayer;
// namespace OpenLoco::Environment
type PathId int

const (
)
// namespace OpenLoco::Audio
// // TODO: This should only be a byte needs to be split off from sound object
type SoundId int

const (
	ClickDown SoundId = 0
	ClickUp SoundId = 1
	ClickPress SoundId = 2
	Construct SoundId = 3
	Demolish SoundId = 4
	Income SoundId = 5
	Crash SoundId = 6
	Water SoundId = 7
	Splash1 SoundId = 8
	Splash2 SoundId = 9
	Waypoint SoundId = 10
	Notification SoundId = 11
	OpenWindow SoundId = 12
	Applause1 SoundId = 13
	Error SoundId = 14
	MultiplayerConnected SoundId = 15
	MultiplayerDisconnected SoundId = 16
	DemolishTree SoundId = 17
	DemolishBuilding SoundId = 18
	VehiclePlace SoundId = 19
	VehiclePickup SoundId = 20
	ConstructShip SoundId = 21
	Ticker SoundId = 22
	Applause2 SoundId = 23
	NewsOooh SoundId = 24
	NewsAwww SoundId = 25
	Breakdown1 SoundId = 26
	Breakdown2 SoundId = 27
	Breakdown3 SoundId = 28
	Breakdown4 SoundId = 29
	Breakdown5 SoundId = 30
	Breakdown6 SoundId = 31
	Null SoundId = 0xFF
)
type ChannelId int

const (
	Music ChannelId = iota
	Unk_1
	Ambient
	Title_deprecated
	Vehicle_0
)
const NumReservedChannels int32 = 4 + 10
type WAVEFORMATEX struct {
	WFormatTag int16
	NChannels int16
	NSamplesPerSec int32
	NAvgBytesPerSec int32
	NBlockAlign int16
	WBitsPerSample int16
	CbSize int16
}
// func InitialiseDSound() 
// func DisposeDSound() 
// func Close() 
// const std::vector<std::string>& getDevices();
// const char* getCurrentDeviceName();
// func GetCurrentDevice() int
// func SetDevice(index int) 
// std::optional<uint32_t> getSoundSample(SoundId id);
// func ShouldSoundLoop(id SoundId) bool
// func ToggleSound() 
// func PauseSound() 
// func UnpauseSound() 
// func PlaySound(id SoundId, loc World::Pos3) 
// // FOR HOOKS ONLY DO NOT USE THIS FUNCTION FOR OPENLOCO CODE
// // INSTEAD USE playSound(SoundId id, const Map::Pos3& loc) OR playSound(SoundId id, int32_t pan)
// func PlaySound(id SoundId, loc World::Pos3, pan int32) 
// func PlaySound(id SoundId, pan int32) 
// func PlaySound(id SoundId, loc World::Pos3, volume int32, frequency int32) 
// func UpdateSounds() 
// func SetBgmVolume(volume int32) 
// func UpdateVehicleNoise() 
// func StopVehicleNoise() 
// func StopVehicleNoise(head EntityId) 
// func UpdateAmbientNoise() 
// func StopAmbientNoise() 
// func RevalidateCurrentTrack() 
// func ResetMusic() 
// func PlayBackgroundMusic() 
// func StopMusic() 
// func PauseMusic() 
// func UnpauseMusic() 
// func PlayMusic(sample Environment::PathId, volume int32, loop bool) 
// func ResetSoundObjects() 
// func IsAudioEnabled() bool
// func IsObjectSoundId(id SoundId) bool
// return static_cast<int32_t>(id) & 0x8000;
// func MakeObjectSoundId(id SoundObjectId_t) SoundId
// return static_cast<SoundId>((static_cast<int32_t>(id) | 0x8000));
// func CalculatePan(coord coord_t, screenSize int32) int32
