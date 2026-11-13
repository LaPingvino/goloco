package main

import (
	"os"
	"path/filepath"
	"strings"
)

// Config (Config.cpp, the openloco.yml read at start and written when a
// setting changes): goloco keeps its own copy in its own folder —
// <user config dir>/goloco/openloco.yml — so it never touches a real
// OpenLoco's. PARTIAL: only the settings goloco saves so far (the language,
// the scale, the audio section's music settings …); the file is upstream's
// YAML — "key: value" lines, a section's keys indented under "section:" —
// read and written line by line, other lines kept as they are. A key is
// named by its path: "language", "audio.playlist",
// "audio.customJukebox.solace".
// On the web the config folder is /w/.config, which the page keeps in the
// browser's localStorage (web/fsshim.js).

func configPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "goloco", "openloco.yml")
}

// configLines is the config file's lines (nil: none yet).
func configLines() []string {
	data, err := os.ReadFile(configPath())
	if err != nil {
		return nil
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil
	}
	return lines
}

// configKey reads a line as a mapping entry: its indent, key and value
// (ok false: blank, a comment, a list item).
func configKey(line string) (indent int, key, value string, ok bool) {
	t := strings.TrimLeft(line, " ")
	if t == "" || strings.HasPrefix(t, "#") || strings.HasPrefix(t, "-") {
		return 0, "", "", false
	}
	k, v, found := strings.Cut(t, ":")
	if !found {
		return 0, "", "", false
	}
	return len(line) - len(t), strings.TrimSpace(k), strings.Trim(strings.TrimSpace(v), `"'`), true
}

// configFind walks the lines for a key path: the line of its deepest part
// that exists (-1: none), how many parts that is, and that line's indent.
func configFind(lines []string, path []string) (line, depth, indent int) {
	line = -1
	type level struct {
		indent int
		key    string
	}
	var stack []level
	for i, l := range lines {
		ind, key, _, ok := configKey(l)
		if !ok {
			continue
		}
		for len(stack) > 0 && stack[len(stack)-1].indent >= ind {
			stack = stack[:len(stack)-1]
		}
		stack = append(stack, level{ind, key})
		if len(stack) > len(path) || len(stack) <= depth {
			continue
		}
		match := true
		for k, lv := range stack {
			match = match && lv.key == path[k]
		}
		if match {
			line, depth, indent = i, len(stack), ind
		}
	}
	return line, depth, indent
}

// configValue is a key's value ("" when unset).
func configValue(key string) string {
	lines, path := configLines(), strings.Split(key, ".")
	if i, depth, _ := configFind(lines, path); depth == len(path) {
		_, _, v, _ := configKey(lines[i])
		return v
	}
	return ""
}

// setConfigValue is Config::write for one key: replaced in place, or added.
func setConfigValue(key, value string) error {
	return setConfigValues([][2]string{{key, value}})
}

// setConfigValues is Config::write for some keys, in one write of the file:
// each replaced in place, or added at the end of its section (the section
// too, when it is new).
func setConfigValues(kvs [][2]string) error {
	p := configPath()
	if p == "" {
		return nil
	}
	lines := configLines()
	for _, kv := range kvs {
		path := strings.Split(kv[0], ".")
		i, depth, indent := configFind(lines, path)
		if depth == len(path) {
			lines[i] = strings.Repeat(" ", indent) + path[depth-1] + ": " + kv[1]
			continue
		}
		at, childIndent := len(lines), 0
		if depth > 0 { // after the section's last line, at its keys' indent
			lines[i] = strings.Repeat(" ", indent) + path[depth-1] + ":" // a section now, whatever it held
			at, childIndent = i+1, indent+2
			for j := i + 1; j < len(lines); j++ {
				ind, _, _, ok := configKey(lines[j])
				if ok && ind <= indent {
					break
				}
				if ok {
					at = j + 1
					if j == i+1 || ind < childIndent {
						childIndent = ind
					}
				}
			}
		}
		var add []string
		for k := depth; k < len(path); k++ {
			l := strings.Repeat(" ", childIndent+2*(k-depth)) + path[k] + ":"
			if k == len(path)-1 {
				l += " " + kv[1]
			}
			add = append(add, l)
		}
		lines = append(lines[:at], append(add, lines[at:]...)...)
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(strings.Join(lines, "\n")+"\n"), 0o644)
}
