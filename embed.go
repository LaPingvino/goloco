// Package goloco holds files built into the binaries: goloco's own data (not
// the game's), so a goloco started from any folder still has them.
package goloco

import "embed"

// LanguageFS is data/language: OpenLoco's language packs (MIT, its text), the
// fallback when no folder on disk has them.
//
//go:embed data/language/*.yml
var LanguageFS embed.FS
