//go:build js && wasm

package audio

// ProbeChild does nothing in the browser (there is no child process).
func ProbeChild() {}

// DeviceAvailable: the browser always has its audio context (it may stay
// silent until a click, but it doesn't fail the game).
func DeviceAvailable() bool { return true }
