package main

import (
	"log"
	"path/filepath"

	"github.com/LaPingvino/goloco/pkg/audio"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// audioTick is Audio::tick (reclaimFinishedInstances is the frame's
// sfx.Update): the vehicles' sounds, then the ambient sound. Upstream calls
// it once per frame (OpenLoco.cpp tick) and once per game tick (GameScene
// tick); so does goloco.
func (g *Game) audioTick() {
	g.updateVehicleNoise()
	g.updateAmbientNoise()
}

// updateAmbientNoise is AmbientAudio.cpp updateAmbientNoise: sound off or
// paused it does nothing (pauseSound / toggleSound stopped the handle:
// stopAmbientNoise); else the main view's scan chooses (world AmbientScan),
// and the handle fades (pkg/audio UpdateAmbient).
func (g *Game) updateAmbientNoise() {
	if g.sfx.Disabled || g.isPaused {
		g.sfx.StopAmbient(&g.ambient)
		return
	}
	chosen, maxVol := audio.AmbientNone, int32(-3500)
	if g.w != nil && g.r != nil && g.r.ObjMgr != nil {
		water, wild, trees, zoom := g.w.AmbientScan(g.r.ObjMgr)
		chosen, maxVol = audio.ChooseAmbient(water, wild, trees), audio.AmbientMaxVolume(zoom)
	}
	g.sfx.UpdateAmbient(&g.ambient, chosen, maxVol, g.ambientSample)
}

// ambientSample is loadMusicSample(css2 / css3 / css4), each file read once.
func (g *Game) ambientSample(which int) (audio.Sample, bool) {
	if g.ambientSamples == nil {
		g.ambientSamples = map[int]*audio.Sample{}
	}
	s, seen := g.ambientSamples[which]
	if !seen {
		if g.dataDir != "" {
			if smp, err := audio.ReadMusicSample(filepath.Join(g.dataDir, audio.AmbientFiles[which])); err == nil {
				s = &smp
			} else {
				log.Printf("[Audio] %v", err)
			}
		}
		g.ambientSamples[which] = s
	}
	if s == nil {
		return audio.Sample{}, false
	}
	return *s, true
}

// updateVehicleNoise is VehicleAudio.cpp updateVehicleNoise, from Audio::tick:
// with a map loaded, sound on and the game not paused, every train's
// Vehicle2 and tail sound (updateSingleVehicleSound, pkg/audio). Paused or
// muted, the sounds stop (pauseSound, toggleSound: stopVehicleNoise). A
// handle no vehicle holds any more — its train deleted (VehicleManager
// deleteTrain: stopVehicleNoise(head)), or another game loaded (S5 load:
// stopVehicleNoise) — is destroyed.
func (g *Game) updateVehicleNoise() {
	if g.w == nil || g.r == nil || g.r.ObjMgr == nil {
		return
	}
	trains, err := g.w.Map.Trains()
	if err != nil {
		return
	}
	stop := g.sfx.Disabled || g.isPaused
	held := map[uint32]bool{}
	for _, tr := range trains {
		for _, i := range []int{tr.Veh2, tr.Tail} {
			e := &g.w.Map.Entities[i]
			s := vehicleSound(e)
			if stop {
				g.sfx.StopVehicleSound(s)
				continue
			}
			var view *audio.VehicleView
			if x, y, zoom, under, ok := g.w.VehicleViewAt(int32(e.PosX), int32(e.PosY), int32(e.PosZ)); ok {
				view = &audio.VehicleView{X: x, Y: y, Zoom: zoom, Underground: under}
			}
			g.sfx.UpdateVehicleSound(s, view, int32(g.sw), int32(g.sh), g.vehicleSoundSample)
			if e.SoundHandle != 0 {
				held[e.SoundHandle] = true
			}
		}
	}
	for h := range g.vehicleHandles {
		if !held[h] {
			g.sfx.Destroy(audio.Handle(h))
		}
	}
	g.vehicleHandles = held
}

// vehicleSound is a Vehicle2's or tail's VehicleSound for the audio.
func vehicleSound(e *scenario.VehicleEntity) audio.VehicleSound {
	return audio.VehicleSound{DrivingSoundID: &e.DrivingSoundID, Volume: e.DrivingSoundVolume,
		Frequency: e.DrivingSoundFrequency, Handle: &e.SoundHandle, ActiveSoundID: &e.ActiveSoundID}
}

// vehicleSoundSample is getSoundBuffer and shouldSoundLoop of an object
// SoundId: the sound object in the slot.
func (g *Game) vehicleSoundSample(slot uint8) (audio.Sample, bool, bool) {
	so := g.r.ObjMgr.GetSoundObject(int(slot))
	if so == nil {
		return audio.Sample{}, false, false
	}
	return audio.ObjectSample(so), so.ShouldLoop, true
}
