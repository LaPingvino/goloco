package audio

// The vehicles' running sounds: Audio/VehicleAudio.cpp (OpenLoco reference:
// src/OpenLoco/src/Audio/VehicleAudio.cpp). Each tick every train's
// Vehicle2 and tail sound (VehicleSound: the driving sound the head set,
// pkg/sim drivingsound.go) is kept playing on a handle of its own while a
// view shows the vehicle, its volume from the sound's volume, the view's
// zoom, how far it is off the screen's middle and whether it is below
// ground, its pan from its x on the screen, its pitch the sound's frequency.
//
// PARTIAL: the tunnel reverb (setReverb kTunnelReverb below ground: OpenAL
// EFX, which Ebitengine's mixer has none of).

const (
	volumeModifierZoomIncrement = -35     // kVolumeModifierZoomIncrement
	volumeModifierUnderground   = -28     // kVolumeModifierUnderground
	volumeModifierMax           = 255     // kVolumeModifierMax
	volumeMin                   = -100_00 // kVolumeMin
	vehicleVolumeCalcMin        = -81_91  // kVehicleVolumeCalcMin
	panFalloffStart             = 2048    // kPanFalloffStart
	panFalloffEnd               = 3072    // kPanFalloffEnd
)

// ZoomVolumeModifier is getZoomVolumeModifier: −35 per zoom level, up to 2.
func ZoomVolumeModifier(zoom int) int32 {
	return int32(min(max(zoom, 0), 2)) * volumeModifierZoomIncrement
}

// FalloffModifier is getFalloffModifier: 255 within the middle half of the
// screen's pan, falling to 0 a quarter of a screen past its edge.
func FalloffModifier(pan int32) int32 {
	a := pan
	if a < 0 {
		a = -a
	}
	switch {
	case a <= panFalloffStart:
		return volumeModifierMax
	case a > panFalloffEnd:
		return 0
	}
	return min((panFalloffEnd-a)/4, volumeModifierMax)
}

// VehicleView is where findBestViewportForEntity's view shows a vehicle:
// its point on the screen (viewportToWindow + the window's position), the
// view's zoom, and whether the vehicle is below the surface (isUnderground).
type VehicleView struct {
	X, Y        int32
	Zoom        int
	Underground bool
}

// VehicleAudioAttributes is getVehicleAudioAttributes: the volume from the
// sound's volume (0–255) scaled by the falloff (the nearer screen edge),
// less the zoom's and the underground's modifiers, from −81.91 dB up; the
// pan the screen x's; the pitch the sound's frequency.
func VehicleAudioAttributes(v VehicleView, drivingVolume uint8, frequency uint16, screenW, screenH int32) Attributes {
	panX, panY := CalculatePan(v.X, screenW), CalculatePan(v.Y, screenH)
	under := int32(0)
	if v.Underground {
		under = volumeModifierUnderground
	}
	falloff := min(FalloffModifier(panX), FalloffModifier(panY))
	overall := max(falloff+under+ZoomVolumeModifier(v.Zoom), 0)
	volume := max(int32(drivingVolume)*overall/8+vehicleVolumeCalcMin, volumeMin)
	return Attributes{Volume: volume, Pan: panX, Frequency: int32(frequency)}
}

// VehicleSound is the VehicleSound fields the audio reads and keeps:
// drivingSoundId (0xFF none), its volume and frequency, audioHandle and
// activeSoundId.
type VehicleSound struct {
	DrivingSoundID *uint8
	Volume         uint8
	Frequency      uint16
	Handle         *uint32
	ActiveSoundID  *uint8
}

const soundObjectNull = 0xFF

// StopVehicleSound is stopVehicleSound: the handle destroyed, no active sound.
func (e *Effects) StopVehicleSound(s VehicleSound) {
	if *s.Handle != 0 {
		e.Destroy(Handle(*s.Handle))
		*s.Handle = 0
	}
	*s.ActiveSoundID = soundObjectNull
}

// UpdateVehicleSound is updateSingleVehicleSound: no driving sound, or no
// view showing the vehicle (view nil), stops it; no handle yet, it starts
// (looping as the sound object says: shouldSoundLoop) when the sound object
// has a sample (sample's ok); the driving sound changed, it stops (the next
// update starts the new one); else the handle takes the new volume, pan and
// pitch.
func (e *Effects) UpdateVehicleSound(s VehicleSound, view *VehicleView, screenW, screenH int32,
	sample func(slot uint8) (smp Sample, loop, ok bool)) {
	if *s.DrivingSoundID == soundObjectNull || view == nil {
		if *s.Handle != 0 {
			e.StopVehicleSound(s)
		}
		return
	}
	a := VehicleAudioAttributes(*view, s.Volume, s.Frequency, screenW, screenH)
	if *s.Handle == 0 {
		smp, loop, ok := sample(*s.DrivingSoundID)
		if !ok {
			return
		}
		a.Loop = loop
		*s.ActiveSoundID = *s.DrivingSoundID
		*s.Handle = uint32(e.Create(smp, ChannelVehicles, a))
		return
	}
	if *s.DrivingSoundID != *s.ActiveSoundID {
		e.StopVehicleSound(s)
		return
	}
	e.SetVolume(Handle(*s.Handle), a.Volume)
	e.SetPan(Handle(*s.Handle), a.Pan)
	e.SetPitch(Handle(*s.Handle), a.Frequency)
}
