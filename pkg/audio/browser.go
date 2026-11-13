package audio

// browserMusic plays music outside the Go runtime. In the browser (js/wasm)
// every goroutine shares the page's one thread, so audio fed from Go stutters
// whenever a game frame runs long; the browser's own <audio> element decodes
// and plays on its own audio thread instead (Joop, 2026-11-01: "piggyback on
// browser features for the web version"). nil on other platforms, where
// Ebitengine's player already runs on its own goroutine/thread.
type browserMusic interface {
	play(path string, loop bool, volume float64) error
	stop()
	setVolume(v float64)
	// playing: a track that has not ended (paused or not).
	playing() bool
	pause()
	resume()
}
