package main

import (
	"os"
	"testing"
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
