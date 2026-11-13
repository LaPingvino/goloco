package world

import "time"

// Now is the game's clock for the tick pacing (GameTicks): real time, or the
// game's virtual clock while a script plays (cmd/goloco clock.go).
var Now = time.Now
