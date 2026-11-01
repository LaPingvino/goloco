//go:build !(js && wasm)

package audio

func newBrowserMusic() browserMusic { return nil }
