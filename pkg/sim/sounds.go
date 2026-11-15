package sim

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
	soundIncome        = 5
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

// TakeSounds returns the sounds queued since the last call and empties the
// queue.
func (m *Map) TakeSounds() []SoundEvent {
	s := m.Sounds
	m.Sounds = nil
	return s
}
