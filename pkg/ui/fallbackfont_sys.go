//go:build !js

package ui

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// findFallbackSource looks for GNU Unifont on the computer: GOLOCO_FALLBACK_FONT
// (a .bdf, .hex or .otf/.ttf, optionally .bdf.gz / .hex.gz) first, else the
// system's and the user's font folders (fallbackFontDirs) and Debian's
// /usr/share/unifont. The bitmap formats come before the outline ones (their
// pixels need no rasterising), the plain Unifont before its variants
// (Unifont_jp, _Upper, _CSUR).
func findFallbackSource() glyphSource {
	if p := os.Getenv("GOLOCO_FALLBACK_FONT"); p != "" {
		return loadFontFile(p)
	}
	home, _ := os.UserHomeDir()
	dirs := append(fallbackFontDirs(runtime.GOOS, home, os.Getenv), "/usr/share/unifont")
	var found []string
	for _, d := range dirs {
		filepath.WalkDir(d, func(path string, e fs.DirEntry, err error) error {
			if err == nil && !e.IsDir() && unifontRank(path) >= 0 {
				found = append(found, path)
			}
			return nil
		})
	}
	sort.SliceStable(found, func(i, j int) bool { return unifontRank(found[i]) < unifontRank(found[j]) })
	for _, p := range found {
		if s := loadFontFile(p); s != nil {
			return s
		}
	}
	return nil
}

// unifontRank orders the Unifont files found: -1 for any other file.
func unifontRank(path string) int {
	base := strings.ToLower(filepath.Base(path))
	if !strings.HasPrefix(base, "unifont") {
		return -1
	}
	stem := strings.TrimSuffix(base, ".gz")
	ext := filepath.Ext(stem)
	formats := map[string]int{".bdf": 0, ".hex": 1, ".otf": 2, ".ttf": 3}
	f, ok := formats[ext]
	if !ok || (ext != ".bdf" && ext != ".hex" && stem != base) {
		return -1
	}
	name := strings.TrimSuffix(stem, ext)
	variant := 0
	// unifont.bdf, unifont-17.0.03.otf: the plain one; _jp, _upper, _csur after
	if strings.ContainsAny(strings.TrimPrefix(name, "unifont"), "_") {
		variant = 10
	}
	return variant + f
}

// fallbackFontDirs are the system's and the user's font folders, as
// Lexington finds its fallback fonts (~/Projects/lexington pdf/sysfont.go
// fontDirs).
func fallbackFontDirs(goos, home string, getenv func(string) string) []string {
	switch goos {
	case "darwin":
		dirs := []string{"/System/Library/Fonts", "/Library/Fonts"}
		if home != "" {
			dirs = append(dirs, filepath.Join(home, "Library/Fonts"))
		}
		return dirs
	case "windows":
		var dirs []string
		if w := getenv("WINDIR"); w != "" {
			dirs = append(dirs, filepath.Join(w, "Fonts"))
		} else {
			dirs = append(dirs, `C:\Windows\Fonts`)
		}
		if l := getenv("LOCALAPPDATA"); l != "" {
			dirs = append(dirs, filepath.Join(l, `Microsoft\Windows\Fonts`))
		}
		return dirs
	default:
		dirs := []string{"/usr/share/fonts", "/usr/local/share/fonts"}
		data := getenv("XDG_DATA_HOME")
		if data == "" && home != "" {
			data = filepath.Join(home, ".local/share")
		}
		if data != "" {
			dirs = append(dirs, filepath.Join(data, "fonts"))
		}
		if home != "" {
			dirs = append(dirs, filepath.Join(home, ".fonts"))
		}
		return dirs
	}
}
