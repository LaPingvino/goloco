package audio

import (
	"io"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

const (
	// SampleRate is the audio sample rate used by Ebitengine
	SampleRate = 44100
)

// Manager handles all game audio playback
type Manager struct {
	context     *audio.Context
	musicPlayer *audio.Player
	musicVolume float64
	sfxVolume   float64
}

// NewManager creates a new audio manager with an initialized audio context
func NewManager() *Manager {
	log.Println("[Audio] Creating new audio manager with sample rate:", SampleRate)
	ctx := audio.NewContext(SampleRate)
	log.Println("[Audio] Audio context created:", ctx != nil)
	return &Manager{
		context:     ctx,
		musicVolume: 1.0,
		sfxVolume:   1.0,
	}
}

// LoadAndPlayMusic loads a WAV file and starts playing it
// The loop parameter controls whether the music loops continuously
func (m *Manager) LoadAndPlayMusic(path string, loop bool) error {
	log.Println("[Audio] LoadAndPlayMusic called with path:", path, "loop:", loop)

	// Stop any currently playing music
	m.StopMusic()

	// Open the WAV file
	log.Println("[Audio] Opening file:", path)
	f, err := os.Open(path)
	if err != nil {
		log.Println("[Audio] ERROR opening file:", err)
		return err
	}
	log.Println("[Audio] File opened successfully")

	// Get file info
	fi, _ := f.Stat()
	if fi != nil {
		log.Println("[Audio] File size:", fi.Size(), "bytes")
	}

	// Decode the WAV file
	// Ebitengine's wav decoder handles sample rate conversion automatically
	log.Println("[Audio] Decoding WAV with sample rate:", SampleRate)
	stream, err := wav.DecodeWithSampleRate(SampleRate, f)
	if err != nil {
		log.Println("[Audio] ERROR decoding WAV:", err)
		f.Close()
		return err
	}
	log.Println("[Audio] WAV decoded, length:", stream.Length())

	var audioStream io.Reader
	if loop {
		// Create an infinite loop stream
		log.Println("[Audio] Creating infinite loop stream")
		audioStream = audio.NewInfiniteLoop(stream, stream.Length())
	} else {
		audioStream = stream
	}

	// Create the player
	log.Println("[Audio] Creating player...")
	player, err := m.context.NewPlayer(audioStream)
	if err != nil {
		log.Println("[Audio] ERROR creating player:", err)
		f.Close()
		return err
	}
	log.Println("[Audio] Player created successfully")

	m.musicPlayer = player
	m.musicPlayer.SetVolume(m.musicVolume)
	log.Println("[Audio] Volume set to:", m.musicVolume)

	log.Println("[Audio] Starting playback...")
	m.musicPlayer.Play()
	log.Println("[Audio] Play() called, IsPlaying:", m.musicPlayer.IsPlaying())

	return nil
}

// StopMusic stops the currently playing music
func (m *Manager) StopMusic() {
	if m.musicPlayer != nil {
		m.musicPlayer.Close()
		m.musicPlayer = nil
	}
}

// SetMusicVolume sets the music volume (0.0 to 1.0)
func (m *Manager) SetMusicVolume(vol float64) {
	if vol < 0 {
		vol = 0
	}
	if vol > 1 {
		vol = 1
	}
	m.musicVolume = vol
	if m.musicPlayer != nil {
		m.musicPlayer.SetVolume(vol)
	}
}

// SetSFXVolume sets the sound effects volume (0.0 to 1.0)
func (m *Manager) SetSFXVolume(vol float64) {
	if vol < 0 {
		vol = 0
	}
	if vol > 1 {
		vol = 1
	}
	m.sfxVolume = vol
}

// IsMusicPlaying returns true if music is currently playing
func (m *Manager) IsMusicPlaying() bool {
	return m.musicPlayer != nil && m.musicPlayer.IsPlaying()
}

// Update should be called each frame to update audio state.
// Stub — currently a no-op.
//
// OpenLoco reference: src/OpenLoco/src/Audio/Audio.cpp
//
//	updateSounds()          — per-frame sound-effect channel processing
//	updateVehicleNoise()    — position-based vehicle sound updates
//	updateAmbientNoise()    — ambient/environmental sound logic
//	playBackgroundMusic()   — background-music state machine
//
// When implemented this should tick any streaming audio channels, check
// whether background music needs to transition, and spatially attenuate
// vehicle/ambient sounds based on viewport position.
func (m *Manager) Update() {
	// stub — no per-frame audio logic yet
}

// Context returns the underlying audio context
func (m *Manager) Context() *audio.Context {
	return m.context
}
