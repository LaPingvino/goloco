//go:build js && wasm

package audio

import (
	"log"
	"os"
	"syscall/js"
)

func newBrowserMusic() browserMusic { return &htmlAudio{} }

// htmlAudio plays a track through an HTMLAudioElement: the file's bytes (the
// fs shim fetches them on first open) become a Blob URL the element plays.
type htmlAudio struct {
	el  js.Value // the <audio> element, or undefined
	url js.Value // its object URL, revoked on stop
}

func (h *htmlAudio) play(path string, loop bool, volume float64) error {
	h.stop()
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	arr := js.Global().Get("Uint8Array").New(len(b))
	js.CopyBytesToJS(arr, b)
	typ := "audio/wav"
	if len(b) > 4 && string(b[:4]) == "OggS" { // the web page's Opus music (cmd/webpack)
		typ = "audio/ogg; codecs=opus"
	}
	blob := js.Global().Get("Blob").New([]any{arr}, map[string]any{"type": typ})
	h.url = js.Global().Get("URL").Call("createObjectURL", blob)
	h.el = js.Global().Get("Audio").New(h.url)
	h.el.Set("loop", loop)
	h.el.Set("volume", volume)
	p := h.el.Call("play")
	// play() returns a promise; a rejection (autoplay policy) is only logged.
	p.Call("catch", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			log.Println("[Audio] browser refused playback:", args[0].Get("message").String())
		}
		return nil
	}))
	log.Printf("[Audio] %s: playing in the browser's <audio> element (%d bytes)", path, len(b))
	return nil
}

func (h *htmlAudio) stop() {
	if h.el.Truthy() {
		h.el.Call("pause")
		h.el.Set("src", "")
		h.el = js.Undefined()
	}
	if h.url.Truthy() {
		js.Global().Get("URL").Call("revokeObjectURL", h.url)
		h.url = js.Undefined()
	}
}

func (h *htmlAudio) setVolume(v float64) {
	if h.el.Truthy() {
		h.el.Set("volume", v)
	}
}

// playing: the element has a track that has not reached its end. A track the
// browser holds back (autoplay policy, or pause) counts: the jukebox must not
// take silence for a finished track and fetch the next one each tick.
func (h *htmlAudio) playing() bool {
	return h.el.Truthy() && !h.el.Get("ended").Bool()
}

func (h *htmlAudio) pause() {
	if h.el.Truthy() {
		h.el.Call("pause")
	}
}

func (h *htmlAudio) resume() {
	if h.el.Truthy() {
		h.el.Call("play")
	}
}
