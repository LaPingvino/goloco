package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// soundTrain is a head, Vehicle2 and tail with no cars.
func soundTrain(speed int32, motor uint8) (*Map, Train) {
	m := &Map{Entities: []scenario.VehicleEntity{
		{EntityType: scenario.VehicleTypeHead},
		{EntityType: scenario.VehicleTypeVeh2, CurrentSpeed: speed, MotorState: motor},
		{EntityType: scenario.VehicleTypeTail},
	}}
	return m, Train{Head: 0, Veh2: 1, Tail: 2}
}

// updateDrivingSoundFriction: none under minSpeed; above it frequency
// (diff >> speedFreqFactor) + baseFrequency, volume ((diff >> factor) +
// baseVolume) truncated to a byte then capped at maxVolume.
func TestDrivingSoundFriction(t *testing.T) {
	snd := &objects.VehicleFrictionSound{MinSpeed: 10 << 16, SpeedFreqFactor: 8, BaseFrequency: 20000,
		SpeedVolumeFactor: 15, BaseVolume: 100, MaxVolume: 160}
	m, tr := soundTrain(9<<16, MotorCoasting)
	s := &m.Entities[tr.Veh2]
	s.DrivingSoundID = 3
	m.updateDrivingSoundFriction(tr, s, snd, 7)
	if s.DrivingSoundID != soundObjectIDNull {
		t.Errorf("under minSpeed: id %d, want none", s.DrivingSoundID)
	}
	m.Entities[tr.Veh2].CurrentSpeed = 30 << 16 // diff 20 mph = 0x140000
	m.updateDrivingSoundFriction(tr, s, snd, 7)
	if s.DrivingSoundID != 7 || s.DrivingSoundFrequency != 20000+0x1400 || s.DrivingSoundVolume != 100+40 {
		t.Errorf("30 mph: id %d freq %d vol %d, want 7 25120 140", s.DrivingSoundID, s.DrivingSoundFrequency, s.DrivingSoundVolume)
	}
	m.Entities[tr.Veh2].CurrentSpeed = 90 << 16 // volume 100 + 160 = 260: a byte, 4, under the cap
	m.updateDrivingSoundFriction(tr, s, snd, 7)
	if s.DrivingSoundVolume != 4 {
		t.Errorf("90 mph: vol %d, want 4 (std::min<uint8_t> truncates 260 first)", s.DrivingSoundVolume)
	}
}

// stepDrivingSound: a sound that starts takes half the idle volume and a
// quarter of the idle frequency; then steps towards the target, wrapping as
// uint16 / uint8 when a step passes zero.
func TestDrivingSoundSteps(t *testing.T) {
	s := &scenario.VehicleEntity{DrivingSoundID: soundObjectIDNull}
	stepDrivingSound(s, 5, 8000, 90, 12000, 120, 300, 200, 10, 20, nil)
	if s.DrivingSoundID != 5 || s.DrivingSoundFrequency != 2000 || s.DrivingSoundVolume != 45 {
		t.Fatalf("start: %d %d %d, want 5 2000 45", s.DrivingSoundID, s.DrivingSoundFrequency, s.DrivingSoundVolume)
	}
	stepDrivingSound(s, 5, 8000, 90, 12000, 120, 300, 200, 10, 20, nil)
	if s.DrivingSoundFrequency != 2300 || s.DrivingSoundVolume != 55 {
		t.Errorf("up: %d %d, want 2300 55", s.DrivingSoundFrequency, s.DrivingSoundVolume)
	}
	s.DrivingSoundFrequency, s.DrivingSoundVolume = 12100, 125
	stepDrivingSound(s, 5, 8000, 90, 12000, 120, 300, 200, 10, 20, nil)
	if s.DrivingSoundFrequency != 12000 || s.DrivingSoundVolume != 120 {
		t.Errorf("down to target: %d %d, want 12000 120", s.DrivingSoundFrequency, s.DrivingSoundVolume)
	}
	s.DrivingSoundFrequency, s.DrivingSoundVolume = 150, 15
	stepDrivingSound(s, 5, 8000, 90, 100, 10, 300, 200, 10, 20, nil)
	if s.DrivingSoundFrequency != 65486 || s.DrivingSoundVolume != 251 {
		t.Errorf("wrap: %d %d, want 65486 251 (150 − 200, 15 − 20 as uint16 / uint8)", s.DrivingSoundFrequency, s.DrivingSoundVolume)
	}
	coast := uint8(30) // the gearbox: a falling frequency aims at the coasting volume
	s.DrivingSoundFrequency, s.DrivingSoundVolume = 5000, 100
	stepDrivingSound(s, 5, 8000, 90, 4000, 200, 300, 200, 10, 20, &coast)
	if s.DrivingSoundFrequency != 4800 || s.DrivingSoundVolume != 80 {
		t.Errorf("gearbox falling: %d %d, want 4800 80", s.DrivingSoundFrequency, s.DrivingSoundVolume)
	}
}

// The motors' targets: simple — accelerating, base + speed >> factor at the
// acceleration volume; coasting from 12 mph the coasting sound; else idle.
// Gearbox — accelerating in gear: the first gear's frequency less each
// higher gear's offset reached, plus speed >> factor; coasting under 12 mph
// idles at the coasting volume.
func TestDrivingSoundMotors(t *testing.T) {
	sm := &objects.VehicleSimpleMotorSound{IdleFrequency: 8000, IdleVolume: 80, CoastingFrequency: 9000, CoastingVolume: 60,
		AccelerationBaseFreq: 10000, AccelerationVolume: 120, FreqIncreaseStep: 20000, FreqDecreaseStep: 20000,
		VolumeIncreaseStep: 100, VolumeDecreaseStep: 100, SpeedFreqFactor: 12}
	for _, c := range []struct {
		speed int32
		motor uint8
		freq  uint16
		vol   uint8
		name  string
	}{
		{20 << 16, MotorAccelerating, 10000 + 20<<4, 120, "simple accelerating"},
		{20 << 16, MotorCoasting, 9000, 60, "simple coasting"},
		{11 << 16, MotorCoasting, 8000, 80, "simple coasting slow"},
		{20 << 16, MotorStopped, 8000, 80, "simple stopped"},
	} {
		m, tr := soundTrain(c.speed, c.motor)
		s := &m.Entities[tr.Veh2]
		// running, from 20000 at volume 100: one step (20000, 100) reaches
		// any target in 0…40000 and 0…200 without wrapping
		s.DrivingSoundID, s.DrivingSoundFrequency, s.DrivingSoundVolume = 1, 20000, 100
		m.updateSimpleMotorSound(tr, s, true, sm, 1)
		if s.DrivingSoundFrequency != c.freq || s.DrivingSoundVolume != c.vol {
			t.Errorf("%s: %d %d, want %d %d", c.name, s.DrivingSoundFrequency, s.DrivingSoundVolume, c.freq, c.vol)
		}
	}
	gb := &objects.VehicleGearboxMotorSound{IdleFrequency: 7000, IdleVolume: 70, FirstGearFrequency: 25000, FirstGearSpeed: 20,
		SecondGearFreqOffset: 1000, SecondGearSpeed: 40, ThirdGearFreqOffset: 500, ThirdGearSpeed: 60, FourthGearFreqOffset: 250,
		CoastingVolume: 50, AcceleratingVolume: 130, FreqIncreaseStep: 20000, FreqDecreaseStep: 20000,
		VolumeIncreaseStep: 100, VolumeDecreaseStep: 100, SpeedFreqFactor: 14}
	for _, c := range []struct {
		speed int32
		motor uint8
		freq  uint16
		vol   uint8
		name  string
	}{
		{5 << 16, MotorAccelerating, 25000 + (7<<16)>>14, 130, "first gear, at least 7 mph"},
		{45 << 16, MotorAccelerating, 25000 - 1000 - 500 + (45<<16)>>14, 130, "third gear"},
		{10 << 16, MotorCoasting, 7000, 50, "coasting slow"},
		{65 << 16, MotorBraking, 25000 - 1750 + (65<<16)>>14, 50, "braking in fourth gear"},
		{0, MotorStopped, 7000, 50, "idle (the frequency falling: the coasting volume)"},
	} {
		m, tr := soundTrain(c.speed, c.motor)
		s := &m.Entities[tr.Veh2]
		s.DrivingSoundID, s.DrivingSoundFrequency, s.DrivingSoundVolume = 1, 20000, 100
		m.updateGearboxMotorSound(tr, s, true, gb, 1)
		if s.DrivingSoundFrequency != c.freq || s.DrivingSoundVolume != c.vol {
			t.Errorf("%s: %d %d, want %d %d", c.name, s.DrivingSoundFrequency, s.DrivingSoundVolume, c.freq, c.vol)
		}
	}
}

// The title map was saved by the game: its trains' VehicleSound objects are
// updateTrainProperties' sounding cars, and a driving sound's id is the slot
// of the sound object the vehicle names (ReorderSoundObjects) — unchanged
// after goloco's own updates.
func TestDrivingSoundTitleMap(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	trains, err := m.Trains()
	if err != nil {
		t.Fatal(err)
	}
	checked, sounding := 0, 0
	for _, tr := range trains {
		v2, tail := m.Entities[tr.Veh2], m.Entities[tr.Tail]
		cp := *m
		cp.Entities = append([]scenario.VehicleEntity(nil), m.Entities...)
		cp.updateSoundingObjects(tr, om)
		if g, w := cp.Entities[tr.Veh2].SoundObjectID, v2.SoundObjectID; g != w {
			t.Errorf("train %d: Vehicle2 sounds for %d, saved %d", m.Entities[tr.Head].ID, g, w)
		}
		if g, w := cp.Entities[tr.Tail].SoundObjectID, tail.SoundObjectID; g != w {
			t.Errorf("train %d: tail sounds for %d, saved %d", m.Entities[tr.Head].ID, g, w)
		}
		for _, s := range []scenario.VehicleEntity{v2, tail} {
			if s.DrivingSoundID == soundObjectIDNull {
				continue
			}
			sounding++
			obj := vehicleObject(om, s.SoundObjectID)
			if obj == nil {
				t.Fatalf("train %d: no vehicle object %d", m.Entities[tr.Head].ID, s.SoundObjectID)
			}
			if w := om.SoundObjectIndex(obj.DrivingSoundName); int(s.DrivingSoundID) != w {
				t.Errorf("train %d: driving sound %d, saved; %q is slot %d", m.Entities[tr.Head].ID, s.DrivingSoundID, obj.DrivingSoundName, w)
			}
		}
		checked++
	}
	if checked == 0 || sounding == 0 {
		t.Fatalf("%d trains, %d sounding: nothing checked", checked, sounding)
	}
	t.Logf("%d trains, %d driving sounds", checked, sounding)

	// One updateDrivingSound on the saved state: the game saved each sound
	// one tick's ramp or speed change from now, so a right reading of the
	// objects' sound parameters keeps every id and moves the volume and the
	// frequency by little (a misread offset gives wild values).
	moved := 0
	for _, tr := range trains {
		for _, i := range []int{tr.Veh2, tr.Tail} {
			before := m.Entities[i]
			if before.DrivingSoundID == soundObjectIDNull {
				continue
			}
			m.updateDrivingSound(tr, &m.Entities[i], i == tr.Veh2, om)
			after := m.Entities[i]
			dv := int(after.DrivingSoundVolume) - int(before.DrivingSoundVolume)
			df := int(after.DrivingSoundFrequency) - int(before.DrivingSoundFrequency)
			if after.DrivingSoundID != before.DrivingSoundID || dv < -5 || dv > 5 || df < -1000 || df > 1000 {
				t.Errorf("train %d: id %d→%d vol %d→%d freq %d→%d in one update", m.Entities[tr.Head].ID, before.DrivingSoundID,
					after.DrivingSoundID, before.DrivingSoundVolume, after.DrivingSoundVolume, before.DrivingSoundFrequency, after.DrivingSoundFrequency)
			}
			if dv != 0 || df != 0 {
				moved++
			}
		}
	}
	t.Logf("one update: %d of %d sounds moved a step", moved, sounding)
}

// chuffSound (steamPuffsAnimationUpdate's sound, outside a station): none
// above 15 mph; the chuff at −(speed >> 9), never louder than −4 dB; at or
// under the ground the tunnel chuff (the second half, with tunnel sounds)
// 15 dB down.
func TestChuffSound(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	slot := -1
	for i, st := range om.SteamObjects {
		if st != nil && len(st.SoundNames) >= 2 && st.Flags&objects.SteamFlagHasTunnelSounds != 0 {
			slot = i
			break
		}
	}
	if slot < 0 {
		t.Skip("no steam object with tunnel sounds")
	}
	st := om.SteamObjects[slot]
	n := len(st.SoundNames)
	x, y := int32(64*32+16), int32(64*32+16)
	land, _ := m.GetHeight(x, y)
	body := &scenario.VehicleEntity{ChuffSoundIndex: 1}
	front := &scenario.VehicleEntity{TileX: -1} // on no tile: no station
	for _, c := range []struct {
		speed  int32
		z      int32
		name   string
		volume int32
		none   bool
	}{
		{16 << 16, int32(land) + 40, "", 0, true},
		{10 << 16, int32(land) + 40, st.SoundNames[1], -(10 << 16 >> 9), false},
		{1 << 16, int32(land) + 40, st.SoundNames[1], -400, false},
		{10 << 16, int32(land), st.SoundNames[n/2+1], -(10 << 16 >> 9) - 1500, false},
	} {
		m.Sounds = nil
		v2 := &scenario.VehicleEntity{CurrentSpeed: c.speed}
		m.chuffSound(om, body, front, v2, slot, x, y, c.z)
		if c.none {
			if len(m.Sounds) != 0 {
				t.Errorf("%d mph: %v, want none", c.speed>>16, m.Sounds)
			}
			continue
		}
		if len(m.Sounds) != 1 || m.Sounds[0].Object != c.name || m.Sounds[0].Volume != c.volume || m.Sounds[0].Channel != ChannelVehicles {
			t.Errorf("%d mph at z %d: %+v, want %q at %d", c.speed>>16, c.z, m.Sounds, c.name, c.volume)
		}
	}
}

// A train off the map (picked up) has no driving sound: VehicleHead::update
// runs updateDrivingSounds before its tileX check.
func TestDrivingSoundOffMap(t *testing.T) {
	m, tr := soundTrain(0, MotorStopped)
	m.Entities[tr.Head].TileX = -1
	for _, i := range []int{tr.Veh2, tr.Tail} {
		m.Entities[i].DrivingSoundID, m.Entities[i].SoundObjectID = 4, 0
	}
	m.updateDrivingSounds(tr, &objects.ObjectManager{})
	if m.Entities[tr.Veh2].DrivingSoundID != soundObjectIDNull || m.Entities[tr.Tail].DrivingSoundID != soundObjectIDNull {
		t.Errorf("off the map: %d %d, want none", m.Entities[tr.Veh2].DrivingSoundID, m.Entities[tr.Tail].DrivingSoundID)
	}
}

// updateTrainProperties' sounding objects: the first car with a driving
// sound at Vehicle2, the last other one at the tail; turned round (the
// front's object at the tail), the two sounds and their handles swap.
func TestSoundingObjects(t *testing.T) {
	om := &objects.ObjectManager{Vehicles: []*objects.VehicleObject{
		{DrivingSoundType: objects.DrivingSoundGearboxMotor}, {DrivingSoundType: objects.DrivingSoundNone},
		{DrivingSoundType: objects.DrivingSoundSimpleMotor}}}
	m, tr := soundTrain(0, MotorStopped)
	for _, id := range []uint16{1, 2, 0, 1} { // cars: none, simple, gearbox, none
		m.Entities = append(m.Entities, scenario.VehicleEntity{EntityType: scenario.VehicleTypeBodyStart, ObjectID: id})
		tr.Cars = append(tr.Cars, Car{{Body: len(m.Entities) - 1}})
	}
	v2, tail := &m.Entities[tr.Veh2], &m.Entities[tr.Tail]
	v2.SoundObjectID, tail.SoundObjectID = 0, 2 // as before the turn: gearbox in front
	v2.DrivingSoundID, v2.SoundHandle, tail.DrivingSoundID, tail.SoundHandle = 7, 70, 8, 80
	m.updateSoundingObjects(tr, om)
	if v2.SoundObjectID != 2 || tail.SoundObjectID != 0 {
		t.Fatalf("objects: Vehicle2 %d, tail %d, want 2 and 0", v2.SoundObjectID, tail.SoundObjectID)
	}
	if v2.DrivingSoundID != 8 || v2.SoundHandle != 80 || tail.DrivingSoundID != 7 || tail.SoundHandle != 70 {
		t.Errorf("turned round: Vehicle2 %d/%d, tail %d/%d, want the sounds swapped", v2.DrivingSoundID, v2.SoundHandle, tail.DrivingSoundID, tail.SoundHandle)
	}
	m.updateSoundingObjects(tr, om) // again: nothing turned
	if v2.DrivingSoundID != 8 || tail.DrivingSoundID != 7 {
		t.Errorf("unchanged train swapped again: %d %d", v2.DrivingSoundID, tail.DrivingSoundID)
	}
}

// railProduceCrossingWhistle: the first car's last start sound, at
// Vehicle2 + 22, 15 dB down below the ground (the empty map's land: 16).
func TestCrossingWhistle(t *testing.T) {
	om := &objects.ObjectManager{Vehicles: []*objects.VehicleObject{
		{NumStartSounds: 3 | 0x80, StartSoundNames: []string{"a", "b", "whistle"}}}}
	m, tr := soundTrain(0, MotorStopped)
	m.Entities = append(m.Entities, scenario.VehicleEntity{EntityType: scenario.VehicleTypeBogie, ObjectID: 0})
	tr.Cars = []Car{{{Front: len(m.Entities) - 1}}}
	for _, c := range []struct{ z, volume int32 }{{30, 0}, {10, -1500}} {
		m.Sounds = nil
		m.Entities[tr.Veh2].PosZ = int16(c.z)
		rng := m.Rng1
		m.railProduceCrossingWhistle(tr, om)
		if m.Rng1 != rng { // upstream e71abfb removed vanilla's gPrng1 draw here
			t.Error("the whistle drew gPrng1")
		}
		if len(m.Sounds) != 1 || m.Sounds[0].Object != "whistle" || m.Sounds[0].Volume != c.volume || m.Sounds[0].Z != c.z+22 {
			t.Errorf("z %d: %+v, want the whistle at %d", c.z, m.Sounds, c.volume)
		}
	}
}

// chuffSound in a station: the front bogie on its own track piece with a
// station element after it plays the tunnel chuff (the second half) at the
// speed's volume, uncapped; the station a ghost, the plain chuff, capped.
func TestChuffSoundStation(t *testing.T) {
	om := &objects.ObjectManager{SteamObjects: []*objects.SteamObject{
		{Flags: objects.SteamFlagHasTunnelSounds, SoundNames: []string{"c0", "c1", "t0", "t1"}}}}
	for _, ghost := range []bool{false, true} {
		tiles := make([][]Tile, 4)
		for y := range tiles {
			tiles[y] = make([]Tile, 4)
		}
		tl := &tiles[1][2]
		tl.Tracks = []scenario.TrackElement{{TrackID: 5, Rotation: 1, BaseZ: 8, HasStation: true}}
		tl.Stations = []scenario.StationElement{{}}
		st := scenario.ElemRef{Kind: scenario.KindStation}
		if ghost {
			st.Flags = elemFlagGhost
		}
		tl.Order = []scenario.ElemRef{{Kind: scenario.KindSurface}, {Kind: scenario.KindTrack}, st}
		m := &Map{Width: 4, Height: 4, Tiles: tiles}
		front := &scenario.VehicleEntity{TileX: 2 * 32, TileY: 1 * 32, TileBaseZ: 8, TrackAndDirection: 5<<3 | 1}
		body := &scenario.VehicleEntity{ChuffSoundIndex: 1}
		v2 := &scenario.VehicleEntity{CurrentSpeed: 1 << 16}
		m.chuffSound(om, body, front, v2, 0, 2*32+16, 32+16, 100)
		want, vol := "t1", int32(-(1 << 16 >> 9))
		if ghost {
			want, vol = "c1", -400
		}
		if len(m.Sounds) != 1 || m.Sounds[0].Object != want || m.Sounds[0].Volume != vol {
			t.Errorf("ghost %v: %+v, want %q at %d", ghost, m.Sounds, want, vol)
		}
	}
}
