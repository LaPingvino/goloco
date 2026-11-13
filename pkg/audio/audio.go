package audio

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

const (
	// SampleRate is the audio context's rate: Locomotion's music (Data/*.dat,
	// RIFF WAVE) is 22050 Hz 16-bit stereo, and OpenLoco plays it at that rate.
	// Running the context at 44100 made Ebitengine resample every buffer while
	// playing: 64 ms of CPU per second of music natively, more in the browser,
	// where it shares the game's single thread — the music stuttered (Joop).
	// At the native rate decoding is a plain copy.
	SampleRate = 22050

	// browserBuffer is the player buffer in js/wasm: the default is sized for a
	// thread of its own; in the browser a long frame must not drain it.
	browserBuffer = 250 * time.Millisecond
)

// Manager handles all game audio playback
type Manager struct {
	context     *audio.Context
	musicPlayer *audio.Player
	musicFile   *os.File // backing file of musicPlayer; Player.Close does not close its source
	musicVolume float64
	sfxVolume   float64
	browser     browserMusic // js/wasm: music through the browser (browser.go); nil elsewhere
	musicPaused bool         // PauseMusic holds the track (Music.cpp pauseMusic)
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
		browser:     newBrowserMusic(),
	}
}

// LoadAndPlayMusic loads a WAV file and starts playing it
// The loop parameter controls whether the music loops continuously
func (m *Manager) LoadAndPlayMusic(path string, loop bool) (err error) {
	if m == nil {
		return nil // muted (GOLOCO_MUTE)
	}
	log.Println("[Audio] LoadAndPlayMusic called with path:", path, "loop:", loop)

	// Stop any currently playing music
	m.StopMusic()
	if path, err = FindFile(path); err != nil {
		return err
	}
	if m.browser != nil {
		return m.browser.play(path, loop, m.musicVolume)
	}

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

	if runtime.GOOS == "js" {
		player.SetBufferSize(browserBuffer)
	}
	m.musicPlayer = player
	m.musicFile = f
	m.musicPlayer.SetVolume(m.musicVolume)
	log.Println("[Audio] Volume set to:", m.musicVolume)

	log.Println("[Audio] Starting playback...")
	m.musicPlayer.Play()
	log.Println("[Audio] Play() called, IsPlaying:", m.musicPlayer.IsPlaying())

	return nil
}

// StopMusic stops the currently playing music
func (m *Manager) StopMusic() {
	if m == nil {
		return
	}
	m.musicPaused = false
	if m.browser != nil {
		m.browser.stop()
	}
	if m.musicPlayer != nil {
		m.musicPlayer.Close()
		m.musicPlayer = nil
	}
	if m.musicFile != nil {
		m.musicFile.Close()
		m.musicFile = nil
	}
}

// SetMusicVolume sets the music volume (0.0 to 1.0)
func (m *Manager) SetMusicVolume(vol float64) {
	if m == nil {
		return
	}
	if vol < 0 {
		vol = 0
	}
	if vol > 1 {
		vol = 1
	}
	m.musicVolume = vol
	if m.browser != nil {
		m.browser.setVolume(vol)
	}
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

// IsMusicPlaying reports a track that has not reached its end — playing, or
// held by PauseMusic (Music.cpp: the music handle's isPlaying, which the
// jukebox asks only while the game runs).
func (m *Manager) IsMusicPlaying() bool {
	if m == nil {
		return false
	}
	if m.browser != nil {
		return m.browser.playing()
	}
	return m.musicPlayer != nil && (m.musicPaused || m.musicPlayer.IsPlaying())
}

// PauseMusic is Music.cpp pauseMusic: a playing track holds its place.
func (m *Manager) PauseMusic() {
	if m == nil || m.musicPaused || !m.IsMusicPlaying() {
		return
	}
	m.musicPaused = true
	if m.browser != nil {
		m.browser.pause()
	} else {
		m.musicPlayer.Pause()
	}
}

// UnpauseMusic is Music.cpp unpauseMusic: a paused track plays on.
func (m *Manager) UnpauseMusic() {
	if m == nil || !m.musicPaused {
		return
	}
	m.musicPaused = false
	if m.browser != nil {
		m.browser.resume()
	} else if m.musicPlayer != nil {
		m.musicPlayer.Play()
	}
}

// FindFile is a Locomotion file's path whatever the case of its name on disk
// (OpenLoco asks for "Data/20s1.DAT" and "Data/Rag2.DAT"; an installation
// has 20s1.dat and rag2.dat): the path itself when it exists, else the entry
// of its folder that matches without case.
func FindFile(path string) (string, error) {
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}
	dir, name := filepath.Split(path)
	ents, err := os.ReadDir(filepath.Clean(dir))
	if err != nil {
		return "", err
	}
	for _, e := range ents {
		if strings.EqualFold(e.Name(), name) {
			return filepath.Join(dir, e.Name()), nil
		}
	}
	return "", &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
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
	if m == nil {
		return
	}
	// stub — no per-frame audio logic yet
}

// Context returns the underlying audio context
func (m *Manager) Context() *audio.Context {
	if m == nil {
		return nil
	}
	return m.context
}

// MusicVolume returns the music volume (0.0 to 1.0).
func (m *Manager) MusicVolume() float64 {
	if m == nil {
		return 0
	}
	return m.musicVolume
}
