package sim

// SoundEvent is a sound the simulation asks for (Audio::playSound with a
// world position): the game plays it after the tick, panned and attenuated
// by the main view (world.SoundAt). ID is Audio::SoundId; Channel is
// Audio::ChannelId (ChannelVehicles, ChannelEffects …).
type SoundEvent struct {
	ID      uint16
	Channel uint8
	X, Y, Z int32
}

// Audio::ChannelId values the sim uses.
const (
	ChannelEffects  = 2
	ChannelVehicles = 3
)

// playSound queues a sound for the game to play (Sounds, drained by
// TakeSounds).
func (m *Map) playSound(id uint16, channel uint8, x, y, z int32) {
	m.Sounds = append(m.Sounds, SoundEvent{ID: id, Channel: channel, X: x, Y: y, Z: z})
}

// TakeSounds returns the sounds queued since the last call and empties the
// queue.
func (m *Map) TakeSounds() []SoundEvent {
	s := m.Sounds
	m.Sounds = nil
	return s
}
