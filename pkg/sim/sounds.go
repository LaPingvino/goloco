package sim

import "github.com/LaPingvino/goloco/pkg/objects"

// SoundEvent is a sound the simulation asks for (Audio::playSound with a
// world position): the game plays it after the tick, panned and attenuated
// by the main view (world.SoundAt). ID is Audio::SoundId; Channel is
// Audio::ChannelId (ChannelVehicles, ChannelEffects …).
type SoundEvent struct {
	ID      uint16
	Channel uint8
	X, Y, Z int32
	// Volume is added to the sound's own (hundredths of a dB, ≤ 0);
	// Frequency is playSound's frequency (0: the usual 22050).
	Volume, Frequency int32
	// Object, when set, is the DAT name of a sound object to play instead
	// of ID (Audio::makeObjectSoundId); the mixer has no object sounds yet
	// (PORT-PLANS B step 3), so the game skips these.
	Object string
}

// Audio::ChannelId values the sim uses.
const (
	ChannelEffects  = 2
	ChannelVehicles = 3
	ChannelUI       = 4
)

// playSound queues a sound for the game to play (Sounds, drained by
// TakeSounds).
func (m *Map) playSound(id uint16, channel uint8, x, y, z int32) {
	m.Sounds = append(m.Sounds, SoundEvent{ID: id, Channel: channel, X: x, Y: y, Z: z})
}

// playSoundVF is playSound(id, channel, pos, volume, frequency).
func (m *Map) playSoundVF(id uint16, channel uint8, x, y, z, volume, frequency int32) {
	m.Sounds = append(m.Sounds, SoundEvent{ID: id, Channel: channel, X: x, Y: y, Z: z, Volume: volume, Frequency: frequency})
}

// Audio::SoundId values the sim plays.
const (
	soundDemolish      = 4
	soundIncome        = 5
	soundCrash         = 6
	soundSplash2       = 9
	soundVehiclePlace  = 19
	soundVehiclePickup = 20
	soundConstructShip = 21
)

// playPickupSound is VehiclePickup.cpp's: the pickup sound at Vehicle2,
// 10 dB down, at a random frequency from the command's gPrng2.
func (m *Map) playPickupSound(c *Cmd, tr Train) {
	freq := int32(22050)
	if c.Rng2 != nil {
		freq = c.Rng2.RandNextRange(20003, 24098)
	}
	v := &m.Entities[tr.Veh2]
	m.playSoundVF(soundVehiclePickup, ChannelUI, int32(v.PosX), int32(v.PosY), int32(v.PosZ), -1000, freq)
}

// playPlacedownSound is VehiclePlaceAir.cpp's (0x0048B11D): the place sound
// at pos, 6 dB down, at a random frequency from gPrng2.
func (m *Map) playPlacedownSound(c *Cmd, x, y, z int32) {
	freq := int32(22050)
	if c.Rng2 != nil {
		freq = c.Rng2.RandNextRange(20003, 24095)
	}
	m.playSoundVF(soundVehiclePlace, ChannelUI, x, y, z, -600, freq)
}

// playWaterPlacedownSound is VehiclePlaceWater.cpp's (0x0048B199): the
// ship's construct sound at the command's position, 6 dB down, at a random
// frequency from gPrng2.
func (m *Map) playWaterPlacedownSound(c *Cmd) {
	freq := int32(22050)
	if c.Rng2 != nil {
		freq = c.Rng2.RandNextRange(20003, 24095)
	}
	if c.Pos != nil {
		m.playSoundVF(soundConstructShip, ChannelEffects, c.Pos[0], c.Pos[1], c.Pos[2], -600, freq)
	}
}

// playConstructionPlacementSound is GameCommands.cpp's (0x0048B013): the
// construct sound at full volume, its frequency from gPrng2.
func (m *Map) playConstructionPlacementSound(c *Cmd, x, y, z int32) {
	freq := int32(22050)
	if c.Rng2 != nil {
		freq = c.Rng2.RandNextRange(17955, 26146)
	}
	m.playSoundVF(soundConstruct, ChannelEffects, x, y, z, 0, freq)
}

// playRemovalSound is RemoveTrack.cpp's playTrackRemovalSound (0x0048B04E)
// and RemoveRoad.cpp's copy: the demolish sound, its frequency from gPrng2.
func (m *Map) playRemovalSound(c *Cmd, x, y, z int32) {
	freq := int32(22050)
	if c.Rng2 != nil {
		freq = c.Rng2.RandNextRange(17955, 26146)
	}
	m.playSoundVF(soundDemolish, ChannelEffects, x, y, z, 0, freq)
}

// TakeSounds returns the sounds queued since the last call and empties the
// queue.
func (m *Map) TakeSounds() []SoundEvent {
	s := m.Sounds
	m.Sounds = nil
	return s
}

// tryCreateInitialMovementSound is VehicleHead::tryCreateInitialMovementSound
// (0x004AC0A3): a train that was stopped or waiting at a signal and now
// travels plays one of its first car's start sounds (a random one from
// gPrng1, never the crossing whistle) at Vehicle2 + 22 up, 15 dB down when
// below ground.
func (m *Map) tryCreateInitialMovementSound(tr Train, initialStatus uint8, om *objects.ObjectManager) {
	if m.Entities[tr.Head].Status != StatusTravelling {
		return
	}
	if initialStatus != StatusStopped && initialStatus != StatusWaitingAtSignal {
		return
	}
	if len(tr.Cars) == 0 || len(tr.Cars[0]) == 0 {
		return
	}
	obj := vehicleObject(om, m.Entities[tr.Cars[0][0].Body].ObjectID)
	if obj == nil || obj.NumStartSounds == 0 {
		return
	}
	numSounds := int32(obj.NumStartSounds & 0x7F)
	if obj.NumStartSounds&0x80 != 0 { // kHasCrossingWhistle
		numSounds = max(numSounds-1, 1)
	}
	i := m.Rng1.RandNextBound(numSounds - 1)
	v2 := &m.Entities[tr.Veh2]
	volume := int32(0)
	if land, _ := m.GetHeight(int32(v2.PosX), int32(v2.PosY)); v2.PosZ < land {
		volume = -1500
	}
	name := ""
	if int(i) < len(obj.StartSoundNames) {
		name = obj.StartSoundNames[i]
	}
	m.playObjectSound(SoundEvent{Object: name, Channel: ChannelVehicles,
		X: int32(v2.PosX), Y: int32(v2.PosY), Z: int32(v2.PosZ) + 22, Volume: volume, Frequency: 22050})
}

// produceStartSound is VehicleHead::produceTouchdownAirportSound (on an
// edge whose target is a touchdown node) and produceLeavingDockSound (a ship
// done loading), the same code: one of the first car's start sounds, picked
// from gPrng1, at Vehicle2 + 22 up.
func (m *Map) produceStartSound(tr Train, om *objects.ObjectManager) {
	if len(tr.Cars) == 0 || len(tr.Cars[0]) == 0 {
		return
	}
	obj := vehicleObject(om, m.Entities[tr.Cars[0][0].Body].ObjectID)
	if obj == nil || obj.NumStartSounds == 0 {
		return
	}
	i := m.Rng1.RandNextBound(int32(obj.NumStartSounds&0x7F) - 1)
	name := ""
	if int(i) < len(obj.StartSoundNames) {
		name = obj.StartSoundNames[i]
	}
	v2 := &m.Entities[tr.Veh2]
	m.playObjectSound(SoundEvent{Object: name, Channel: ChannelVehicles,
		X: int32(v2.PosX), Y: int32(v2.PosY), Z: int32(v2.PosZ) + 22, Frequency: 22050})
}

// railProduceCrossingWhistle is Vehicle1.cpp's (0x004B98DA), when the rail
// move approaches a level crossing: the first car's last start sound (the
// crossing whistle) at Vehicle2 + 22 up, 15 dB down below ground.
func (m *Map) railProduceCrossingWhistle(tr Train, om *objects.ObjectManager) {
	if len(tr.Cars) == 0 || len(tr.Cars[0]) == 0 {
		return
	}
	obj := vehicleObject(om, m.Entities[tr.Cars[0][0].Front].ObjectID)
	if obj == nil || obj.NumStartSounds == 0 {
		return
	}
	i := int(obj.NumStartSounds&0x7F) - 1
	name := ""
	if i >= 0 && i < len(obj.StartSoundNames) {
		name = obj.StartSoundNames[i]
	}
	v2 := &m.Entities[tr.Veh2]
	volume := int32(0)
	if land, _ := m.GetHeight(int32(v2.PosX), int32(v2.PosY)); v2.PosZ < land {
		volume = -1500
	}
	m.playObjectSound(SoundEvent{Object: name, Channel: ChannelVehicles,
		X: int32(v2.PosX), Y: int32(v2.PosY), Z: int32(v2.PosZ) + 22, Volume: volume, Frequency: 22050})
}

// playObjectSound queues a sound object's sound (makeObjectSoundId). One
// the vehicle or steam object doesn't name (an empty name: its index past
// the object's sounds, or the object not loaded) is not queued: upstream
// would read past its array.
func (m *Map) playObjectSound(s SoundEvent) {
	if s.Object != "" {
		m.Sounds = append(m.Sounds, s)
	}
}
