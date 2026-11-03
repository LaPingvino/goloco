package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Config (Config.cpp, the openloco.yml read at start and written when a
// setting changes): goloco keeps its own copy in its own folder —
// <user config dir>/goloco/openloco.yml — so it never touches a real
// OpenLoco's. PARTIAL: only the settings goloco saves so far (language); the
// file is upstream's flat "key: value" shape, other lines kept as they are.
// On the web the config folder is /w/.config, which the page keeps in the
// browser's localStorage (web/fsshim.js).

func configPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "goloco", "openloco.yml")
}

// configValue is a top-level key's value ("" when unset).
func configValue(key string) string {
	f, err := os.Open(configPath())
	if err != nil {
		return ""
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if k, v, ok := strings.Cut(sc.Text(), ":"); ok && strings.TrimSpace(k) == key && !strings.HasPrefix(k, " ") {
			return strings.Trim(strings.TrimSpace(v), `"'`)
		}
	}
	return ""
}

// setConfigValue is Config::write for one key: replaced in place, or added.
func setConfigValue(key, value string) error {
	p := configPath()
	if p == "" {
		return nil
	}
	data, _ := os.ReadFile(p)
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) == 1 && lines[0] == "" {
		lines = nil
	}
	line, done := key+": "+value, false
	for i, l := range lines {
		if k, _, ok := strings.Cut(l, ":"); ok && k == key {
			lines[i], done = line, true
		}
	}
	if !done {
		lines = append(lines, line)
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(strings.Join(lines, "\n")+"\n"), 0o644)
}
