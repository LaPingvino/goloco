package main

import (
	"os"
	"testing"

	"github.com/LaPingvino/goloco/pkg/jukebox"
	"github.com/LaPingvino/goloco/pkg/sim"
)

// The config keeps a key across writes, replaces it in place and keeps other
// lines.
func TestConfigValue(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	if v := configValue("language"); v != "" {
		t.Fatalf("unset: %q", v)
	}
	os.MkdirAll(configPath()[:len(configPath())-len("/openloco.yml")], 0o755)
	os.WriteFile(configPath(), []byte("# goloco\nmeasurement_format: metric\n"), 0o644)
	if err := setConfigValue("language", "de-DE"); err != nil {
		t.Fatal(err)
	}
	if err := setConfigValue("language", "fr-FR"); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(configPath())
	if v := configValue("language"); v != "fr-FR" || string(data) != "# goloco\nmeasurement_format: metric\nlanguage: fr-FR\n" {
		t.Errorf("language %q; file:\n%s", v, data)
	}
}

// A section's keys: read by path, replaced in place, added at the section's
// end, a new section (and a section in a section) at the file's end — the
// shape OpenLoco writes its openloco.yml in.
func TestConfigSections(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	os.MkdirAll(configPath()[:len(configPath())-len("/openloco.yml")], 0o755)
	os.WriteFile(configPath(), []byte("display:\n  mode: window\n  playlist: no\naudio:\n  mainVolume: -1100\n  playlist: all\nlanguage: en-GB\n"), 0o644)
	if v := configValue("audio.playlist"); v != "all" {
		t.Errorf("audio.playlist %q", v)
	}
	if v := configValue("playlist"); v != "" {
		t.Errorf("a section's key at the top level: %q", v)
	}
	if v := configValue("audio.customJukebox.solace"); v != "" {
		t.Errorf("unset nested key: %q", v)
	}
	err := setConfigValues([][2]string{{"audio.playlist", "custom"}, {"audio.playJukeboxMusic", "false"},
		{"audio.customJukebox.solace", "true"}, {"audio.customJukebox.eugenia", "false"}, {"network.port", "11753"}, {"language", "nl-NL"}})
	if err != nil {
		t.Fatal(err)
	}
	want := "display:\n  mode: window\n  playlist: no\naudio:\n  mainVolume: -1100\n  playlist: custom\n  playJukeboxMusic: false\n" +
		"  customJukebox:\n    solace: true\n    eugenia: false\nlanguage: nl-NL\nnetwork:\n  port: 11753\n"
	if data, _ := os.ReadFile(configPath()); string(data) != want {
		t.Errorf("file:\n%s\nwant:\n%s", data, want)
	}
	if v := configValue("audio.customJukebox.eugenia"); v != "false" {
		t.Errorf("audio.customJukebox.eugenia %q", v)
	}
}

// The audio section's music settings: Config's defaults without a file, and
// what writeAudioConfig writes read back (the keys OpenLoco's openloco.yml
// has).
func TestAudioConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	cfg, title := loadAudioConfig()
	if !cfg.PlayJukeboxMusic || !title || cfg.Playlist != jukebox.PlaylistCurrentEra || !cfg.CustomJukebox[0] || cfg.CustomJukebox[29] {
		t.Fatalf("defaults: %+v, title music %v", cfg, title)
	}
	g := &Game{audioCfg: cfg}
	g.audioCfg.PlayJukeboxMusic, g.audioCfg.Playlist = false, jukebox.PlaylistCustom
	g.audioCfg.CustomJukebox[11], g.audioCfg.CustomJukebox[29] = false, true
	g.writeAudioConfig()
	cfg, title = loadAudioConfig()
	if cfg != g.audioCfg || title {
		t.Errorf("read back: %+v, title music %v", cfg, title)
	}
	if configValue("audio.playlist") != "custom" || configValue("audio.customJukebox.solace") != "false" || configValue("audio.play_title_music") != "false" {
		data, _ := os.ReadFile(configPath())
		t.Errorf("file:\n%s", data)
	}
}

// The news section and audio.playNewsSounds: Config's defaults without a
// file (every kind a news window, the sounds on), and what writeNewsConfig
// writes read back (ConfigConvert.hpp's names).
func TestNewsConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	settings, sounds := loadNewsConfig()
	for i, v := range settings {
		if v != sim.NewsWindow {
			t.Errorf("default %d: %d", i, v)
		}
	}
	if !sounds {
		t.Error("news sounds off by default")
	}
	g := &Game{newsSettings: settings}
	g.newsSettings[sim.MsgCritGeneral], g.newsSettings[sim.MsgCritAdvice] = sim.NewsTicker, sim.NewsNone
	g.writeNewsConfig()
	settings, sounds = loadNewsConfig()
	if settings != g.newsSettings || sounds {
		t.Errorf("read back: %v, sounds %v", settings, sounds)
	}
	if configValue("news.general") != "ticker" || configValue("news.advice") != "none" || configValue("news.majorCompany") != "newsWindow" || configValue("audio.playNewsSounds") != "false" {
		data, _ := os.ReadFile(configPath())
		t.Errorf("file:\n%s", data)
	}
}
