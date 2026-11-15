package main

import (
	"log"
	"path/filepath"
	"strconv"

	"github.com/LaPingvino/goloco/pkg/audio"
	"github.com/LaPingvino/goloco/pkg/sim"
)

// newEffects is the effects half of Audio.cpp initialiseDSound: css1.dat's
// samples and the channel volumes from openloco.yml's audio section
// (masterVolume, effectsVolume, vehiclesVolume, uiVolume, ambientVolume; 100
// when unset, Config.cpp). Without a sound device or data it is silent but
// still there, so the call sites need no checks.
func newEffects(m *audio.Manager, dataDir string) *audio.Effects {
	var samples []audio.Sample
	if dataDir != "" {
		s, err := audio.ReadCSS(filepath.Join(dataDir, "css1.dat"))
		if err != nil {
			log.Printf("[Audio] css1.dat: %v", err)
		}
		samples = s
	}
	e := audio.NewEffects(m.Context(), samples)
	for _, c := range []struct {
		key string
		ch  audio.ChannelId
	}{{"masterVolume", audio.ChannelMaster}, {"effectsVolume", audio.ChannelEffects},
		{"vehiclesVolume", audio.ChannelVehicles}, {"uiVolume", audio.ChannelUI}, {"ambientVolume", audio.ChannelAmbient}} {
		pct := int32(100)
		if v, err := strconv.Atoi(configValue("audio." + c.key)); err == nil {
			pct = int32(v)
		}
		e.SetChannelVolume(c.ch, audio.PercentToDb(pct))
	}
	return e
}

// playSound is Audio::playSound(id, channel, pan): pan a screen x, or
// audio.KPlayAtCentre.
func (g *Game) playSound(id audio.SoundId, ch audio.ChannelId, pan int) {
	g.sfx.ScreenW = int32(g.sw)
	g.sfx.PlaySound(id, ch, int32(pan))
}

// playSoundAt is Audio::playSound(id, channel, loc): a world position, heard
// only where the main view shows it.
func (g *Game) playSoundAt(id audio.SoundId, ch audio.ChannelId, x, y, z int32) {
	vol, sx, ok := g.w.SoundAt(x, y, z)
	if !ok {
		return
	}
	g.sfx.ScreenW = int32(g.sw)
	g.sfx.PlaySoundAt(id, ch, vol, int32(sx))
}

// playCommandSound is playSound(id, effects, GameCommands::getPosition()):
// the last command's position, nothing when it set none.
func (g *Game) playCommandSound(id audio.SoundId, c *sim.Cmd) {
	if c != nil && c.Pos != nil {
		g.playSoundAt(id, audio.ChannelEffects, c.Pos[0], c.Pos[1], c.Pos[2])
	}
}
