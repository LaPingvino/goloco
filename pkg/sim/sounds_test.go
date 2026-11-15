package sim

import "testing"

func TestSoundQueue(t *testing.T) {
	var m Map
	m.playSound(27, ChannelVehicles, 100, 200, 30)
	m.playSound(7, ChannelEffects, 1, 2, 3)
	got := m.TakeSounds()
	if len(got) != 2 || got[0] != (SoundEvent{ID: 27, Channel: ChannelVehicles, X: 100, Y: 200, Z: 30}) || got[1].ID != 7 {
		t.Fatalf("TakeSounds = %+v", got)
	}
	if len(m.TakeSounds()) != 0 {
		t.Error("the queue is not emptied")
	}
}
