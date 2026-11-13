package ui

import "time"

// Now is the game's clock for timed windows (the error window's 7 s): real
// time, or the game's virtual clock while a script plays (cmd/goloco clock.go).
var Now = time.Now
