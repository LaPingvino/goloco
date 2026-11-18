package main

import (
	"log"
	"path/filepath"
	"strconv"

	"github.com/LaPingvino/goloco/pkg/audio"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/hajimehoshi/ebiten/v2"
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

// playSimSounds plays the sounds the simulation asked for since the last
// frame (sim.Map.TakeSounds: a breakdown …) at their world positions.
func (g *Game) playSimSounds() {
	for _, s := range g.w.Map.TakeSounds() {
		if s.Object != "" { // an object sound: not in the mixer yet (PORT-PLANS B step 3)
			continue
		}
		vol, sx, ok := g.w.SoundAt(s.X, s.Y, s.Z)
		if !ok {
			continue
		}
		g.sfx.ScreenW = int32(g.sw)
		g.sfx.PlaySoundAtVF(audio.SoundId(s.ID), audio.ChannelId(s.Channel), vol+s.Volume, int32(sx), s.Frequency)
	}
}

// dropdownPressSound is MouseInput's press while a dropdown is active: on
// the dropdown, clickUp at its middle (window x + its widget's midX).
// PARTIAL: a press on another window's widget while the dropdown is open
// (upstream plays it there too, at that widget).
func (g *Game) dropdownPressSound() {
	d := g.dropdown
	if d == nil || !d.Visible || g.mouseX < d.X || g.mouseX >= d.X+d.Width || g.mouseY < d.Y || g.mouseY >= d.Y+d.Height {
		return
	}
	g.playSound(audio.SoundClickUp, audio.ChannelUI, d.X+d.Width/2)
}

// playClickSounds plays the window stack's clicks (ui clicks.go: MouseInput's
// clickDown, clickUp and clickPress) on the ui channel.
func (g *Game) playClickSounds() {
	for _, c := range g.wins.TakeClickSounds() {
		id := audio.SoundClickDown
		switch c.ID {
		case ui.ClickUp:
			id = audio.SoundClickUp
		case ui.ClickPress:
			id = audio.SoundClickPress
		}
		g.playSound(id, audio.ChannelUI, c.X)
	}
}

// releaseButton is stateWidgetPressed's leftReleased after a widget press
// (ui WindowStack.ReleaseButton): a dropdown the press opened takes it.
func (g *Game) releaseButton() {
	if g.inMouseJustReleased(ebiten.MouseButtonLeft) && g.wins.WidgetPressed() {
		g.wins.ReleaseButton(g.mouseX, g.mouseY, g.dropdown != nil && g.dropdown.Visible || g.colourDropdown != nil && g.colourDropdown.Visible)
	}
}
