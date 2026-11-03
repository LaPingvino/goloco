package i18n

import (
	"os"
	"strings"
	"sync"
)

// base is en-GB's pack: the text of every string id when the active pack
// has none (OpenLoco loads en-GB first and the chosen language over it).
var (
	base     map[int]string
	baseOnce sync.Once
)

func loadBase() {
	baseOnce.Do(func() {
		base = map[int]string{}
		if p := findPack("en-GB"); p != "" {
			if f, err := os.Open(p); err == nil {
				base, _ = parse(f)
				f.Close()
			}
		}
	})
}

// Str is StringManager::getString: the active pack's text for an OpenLoco
// string id (StringIds), else en-GB's, with its format codes.
func Str(id int) string {
	if s, ok := loaded[id]; ok && s != "" {
		return s
	}
	loadBase()
	return base[id]
}

// Label is Str without the leading style codes ({COLOUR …}, {SMALLFONT},
// {OUTLINE} …) that the drawing code already applies — for labels drawn in
// the window's colour. Codes inside the text (arguments) stay.
func Label(id int) string {
	s := Str(id)
	for strings.HasPrefix(s, "{") {
		end := strings.Index(s, "}")
		if end < 0 {
			break
		}
		code := s[1:end]
		if !strings.HasPrefix(code, "COLOUR") && code != "SMALLFONT" && code != "BIGFONT" && code != "TINYFONT" && code != "OUTLINE" && code != "OUTLINE_OFF" {
			break
		}
		s = s[end+1:]
	}
	return s
}
