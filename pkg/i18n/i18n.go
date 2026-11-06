// Package i18n provides lightweight UI localisation for goloco, sourced from
// OpenLoco's language packs (data/language/*.yml).
//
// The packs are flat "id: text" tables under a top-level "strings:" key, with a
// small "header:" block. We do NOT pull in a YAML dependency: the loader here is
// a hand parser tailored to that exact shape.
//
// Parser limitations (documented deliberately):
//   - Only two sections are understood: "header" and "strings".
//   - A string entry must be a single physical line "  <int-id>: <value>".
//     Block scalars ("|", ">"), multi-line values, anchors/aliases and merge
//     keys are NOT supported (OpenLoco's packs use none of these).
//   - Values may be double-quoted or bare. Quotes are stripped; the packs
//     contain no backslash escapes, so none are interpreted.
//   - "#" comment lines and blank lines are skipped. Trailing "# comment" on a
//     bare value line is NOT stripped (the packs never do this).
//
// Translation is keyed by goloco's own UI string keys (see englishFallback for
// the full set). Each key optionally maps to an OpenLoco string id via keyToID;
// when the active pack has non-empty text for that id it wins, otherwise the
// English fallback is used, so T never returns empty for a known key.
//
// OpenLoco reference: src/OpenLoco/src/Localisation (StringManager / language
// yml packs under data/language).
package i18n

import (
	"io/fs"

	"bufio"
	"fmt"
	"github.com/LaPingvino/goloco"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/LaPingvino/goloco/pkg/objects"
)

// loaded holds the active pack: OpenLoco string id -> text.
var loaded = map[int]string{}

// activeLocale / activeNativeName describe the currently loaded pack.
var (
	activeLocale     = "en-GB"
	activeNativeName = "English (UK)"
)

// keyToID maps goloco UI string keys to OpenLoco string ids where a clean
// vanilla string (plain text, no {FORMAT} codes) exists. Ids are from
// data/language/en-GB.yml. Keys absent here are English-only (see fallback).
var keyToID = map[string]int{
	"menu.load_game":  103,  // "Load Game"
	"menu.options":    654,  // "Options"
	"newgame.cancel":  186,  // "Cancel"
	"cat.beginner":    1570, // "Beginner"
	"cat.easy":        1571, // "Easy"
	"cat.medium":      1572, // "Medium"
	"cat.challenging": 1573, // "Challenging"
	"cat.expert":      1574, // "Expert"
	"cons.station":    159,  // "Station"
	"cons.signal":     158,  // "Signal"
	// Months: en-GB ids 635..646 = January..December.
	"month.1":  635,
	"month.2":  636,
	"month.3":  637,
	"month.4":  638,
	"month.5":  639,
	"month.6":  640,
	"month.7":  641,
	"month.8":  642,
	"month.9":  643,
	"month.10": 644,
	"month.11": 645,
	"month.12": 646,
}

// englishFallback guarantees T never returns empty for a known key, and is the
// sole source for keys with no clean vanilla id (title-menu verbs, Play,
// COMPLETED, the Objective prefix, Cost).
var englishFallback = map[string]string{
	"menu.new_game":        "New Game",
	"menu.load_game":       "Load Game",
	"menu.tutorial":        "Tutorial",
	"menu.scenario_editor": "Scenario Editor",
	"menu.options":         "Options",
	"menu.exit_game":       "Exit Game",

	"newgame.cancel":    "Cancel",
	"newgame.play":      "Play",
	"newgame.completed": "COMPLETED",

	"game.objective": "Objective:",

	"cat.beginner":    "Beginner",
	"cat.easy":        "Easy",
	"cat.medium":      "Medium",
	"cat.challenging": "Challenging",
	"cat.expert":      "Expert",

	"cons.station": "Station",
	"cons.signal":  "Signal",
	"cons.cost":    "Cost:",

	"month.1":  "January",
	"month.2":  "February",
	"month.3":  "March",
	"month.4":  "April",
	"month.5":  "May",
	"month.6":  "June",
	"month.7":  "July",
	"month.8":  "August",
	"month.9":  "September",
	"month.10": "October",
	"month.11": "November",
	"month.12": "December",
}

// T returns the localised UI string for key. Resolution order: active pack text
// (via keyToID) -> English fallback -> the key itself (never empty).
func T(key string) string {
	if id, ok := keyToID[key]; ok {
		if txt := loaded[id]; !isFormatOnly(txt) {
			return txt
		}
	}
	if txt, ok := englishFallback[key]; ok {
		return txt
	}
	return key
}

// Month returns the localised full month name for a 1..12 month number.
func Month(m int) string {
	if m < 1 || m > 12 {
		return ""
	}
	return T(fmt.Sprintf("month.%d", m))
}

// Locale / NativeName report the active pack (for logging / options UI).
func Locale() string     { return activeLocale }
func NativeName() string { return activeNativeName }

// Load reads the language pack for locale (e.g. "de-DE") from ./data/language
// then ~/OpenLoco/data/language, populating the translation table and setting
// objects.SelectedLocoLanguage from the pack header's loco_original_id so that
// DAT object string tables pick the matching vanilla language too.
//
// An empty locale, a missing file, or "en-GB" all leave the table empty (T then
// serves pure English fallbacks), which is a valid, non-error state.
func Load(locale string) error {
	if locale == "" {
		locale = "en-GB"
	}
	f, err := openPack(locale)
	if err != nil {
		return fmt.Errorf("i18n: no language pack found for %q", locale)
	}
	defer f.Close()

	strs, hdr := parse(f)
	loaded = strs
	activeLocale = locale
	if hdr.nativeName != "" {
		activeNativeName = hdr.nativeName
	}
	objects.SelectedLocoLanguage = uint8(hdr.locoOriginalID)
	return nil
}

// packDirs are where language packs are looked for, in order: ./data/language
// (a source checkout, packs being edited), ~/OpenLoco/data/language, then the
// packs built into the binary — so text never goes missing when goloco is
// started from another folder (it did: every label empty).
func packDirs() []fs.FS {
	dirs := []fs.FS{os.DirFS(filepath.Join("data", "language")), os.DirFS(filepath.Join(os.Getenv("HOME"), "OpenLoco", "data", "language"))}
	if sub, err := fs.Sub(goloco.LanguageFS, "data/language"); err == nil {
		dirs = append(dirs, sub)
	}
	return dirs
}

// openPack opens the first pack for locale found in packDirs.
func openPack(locale string) (fs.File, error) {
	for _, d := range packDirs() {
		if f, err := d.Open(locale + ".yml"); err == nil {
			return f, nil
		}
	}
	return nil, fs.ErrNotExist
}

type header struct {
	locale         string
	nativeName     string
	locoOriginalID int
}

// parse reads the flat OpenLoco language yml. See package doc for limitations.
func parse(r io.Reader) (map[int]string, header) {
	res := map[int]string{}
	var hdr header
	section := ""

	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		// Top-level section header: no leading whitespace, ends with ':'.
		if line[0] != ' ' && line[0] != '\t' {
			section = strings.TrimSuffix(trimmed, ":")
			continue
		}
		key, val, ok := splitKV(trimmed)
		if !ok {
			continue
		}
		switch section {
		case "header":
			switch key {
			case "loco_original_id":
				hdr.locoOriginalID, _ = strconv.Atoi(strings.TrimSpace(val))
			case "locale":
				hdr.locale = strings.TrimSpace(val)
			case "native_name":
				hdr.nativeName = unquote(val)
			}
		case "strings":
			id, err := strconv.Atoi(key)
			if err != nil {
				continue
			}
			res[id] = dropSkipped(unquote(val))
		}
	}
	return res, hdr
}

// splitKV splits "key: value" on the first colon (ids are pure digits, so this
// is unambiguous even when the value itself contains colons).
func splitKV(s string) (key, val string, ok bool) {
	i := strings.IndexByte(s, ':')
	if i < 0 {
		return "", "", false
	}
	key = strings.TrimSpace(s[:i])
	val = strings.TrimPrefix(s[i+1:], " ")
	return key, val, true
}

// unquote strips one layer of surrounding double quotes (the packs use no
// backslash escapes) and trims trailing whitespace from bare values.
func unquote(v string) string {
	v = strings.TrimRight(v, " \t")
	if len(v) >= 2 && v[0] == '"' && v[len(v)-1] == '"' {
		return v[1 : len(v)-1]
	}
	return v
}

// isFormatOnly reports whether s is empty or contains only {FORMAT} codes and
// whitespace once those are removed — such entries carry no display text, so
// T treats them as absent and falls through to the English fallback.
func isFormatOnly(s string) bool {
	if s == "" {
		return true
	}
	var b strings.Builder
	depth := 0
	for _, r := range s {
		switch r {
		case '{':
			depth++
		case '}':
			if depth > 0 {
				depth--
			}
		default:
			if depth == 0 {
				b.WriteRune(r)
			}
		}
	}
	return strings.TrimSpace(b.String()) == ""
}

// Language is an available pack: its locale and native name.
type Language struct{ Locale, NativeName string }

// Available is enumerateLanguages: every pack in the first language folder
// that has any (./data/language, then ~/OpenLoco/data/language), sorted by
// native name converted to Locomotion's characters (ToLoco: Č → C, CJK → ?),
// sorted by that converted name.
func Available() []Language {
	for _, dir := range packDirs() {
		ents, err := fs.ReadDir(dir, ".")
		if err != nil {
			continue
		}
		var out []Language
		for _, e := range ents {
			if e.IsDir() || filepath.Ext(e.Name()) != ".yml" {
				continue
			}
			f, err := dir.Open(e.Name())
			if err != nil {
				continue
			}
			_, hdr := parse(f)
			f.Close()
			if hdr.locale != "" {
				out = append(out, Language{Locale: hdr.locale, NativeName: ToLoco(hdr.nativeName)})
			}
		}
		if len(out) > 0 {
			sort.Slice(out, func(i, j int) bool { return out[i].NativeName < out[j].NativeName })
			return out
		}
	}
	return nil
}

// LocoChar is convertUnicodeToLoco for one rune: goloco's strings carry Loco
// codes 32–255 as runes (e.g. 170 ▼), and text from outside (language packs,
// file and object names) is Unicode — letters without a Loco glyph lose their
// accent, others above U+00FF become their Loco code or '?'
// (replacement_character). The four ASCII entries upstream turns into '?'
// ({ | } ~) are left alone: goloco formats its strings itself.
func LocoChar(r rune) int {
	if l, ok := unicodeToLoco[r]; ok {
		return int(l)
	}
	if l, ok := unicodeToLocoAstral[r]; ok {
		return int(l)
	}
	if r < 32 || r > 255 {
		return '?'
	}
	return int(r)
}

// ToLoco is convertUnicodeToLoco for a string (runes of Loco codes).
func ToLoco(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		out = append(out, rune(LocoChar(r)))
	}
	return string(out)
}

// unicodeToLocoAstral is convertUnicodeToLoco above U+FFFF, which the
// generated table (U+0080..U+FFFF) doesn't reach: the transport-mode emoji
// (Conversion.h UnicodeChar water / road / railway → LocoChar 184 / 182 / 177).
var unicodeToLocoAstral = map[rune]uint8{0x1F30A: 184, 0x1F6E3: 182, 0x1F6E4: 177}

// dropSkipped is LanguageFiles.cpp readString's skip: the superscript minus
// (U+207B) and the variation selector (U+FE0F, as after 🛣) are left out.
func dropSkipped(s string) string {
	return strings.Map(func(r rune) rune {
		if r == 0x207B || r == 0xFE0F {
			return -1
		}
		return r
	}, s)
}
