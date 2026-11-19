//go:build js && wasm

package audio

import (
	"log"
	"os"
	"syscall/js"
)

func newBrowserMusic() browserMusic { return &htmlAudio{} }

// htmlAudio plays a track through an HTMLAudioElement from a Blob URL of the
// file's bytes. The web page's fs shim hands the URL over without blocking
// (GolocoFS.audioURL: a track not fetched yet is fetched in the background,
// as the page loader fetches its packs), so a track change doesn't hold the
// game's frame for the download; the element is there from the start, so
// the track counts as playing (not ended) while it loads.
type htmlAudio struct {
	el     js.Value // the <audio> element, or undefined
	url    js.Value // its object URL, revoked on stop
	gen    int      // the play() a resolved URL belongs to: an older one's is dropped
	paused bool
}

func (h *htmlAudio) play(path string, loop bool, volume float64) error {
	h.stop()
	fsys := js.Global().Get("GolocoFS")
	if !fsys.Truthy() || !fsys.Get("audioURL").Truthy() {
		return h.playNow(path, loop, volume)
	}
	h.paused = false
	h.el = js.Global().Get("Audio").New()
	h.el.Set("loop", loop)
	h.el.Set("volume", volume)
	gen := h.gen
	var then, catch js.Func
	release := func() { then.Release(); catch.Release() }
	then = js.FuncOf(func(this js.Value, args []js.Value) any {
		release()
		if gen != h.gen { // stopped or another track since
			js.Global().Get("URL").Call("revokeObjectURL", args[0])
			return nil
		}
		h.url = args[0]
		h.el.Set("src", h.url)
		if !h.paused {
			h.start()
		}
		log.Printf("[Audio] %s: playing in the browser's <audio> element", path)
		return nil
	})
	catch = js.FuncOf(func(this js.Value, args []js.Value) any {
		release()
		// The element stays without a source: the track passes in silence,
		// as a track that can't be read does natively (cmd/goloco silentTrack).
		if len(args) > 0 {
			log.Printf("[Audio] %s: %s (the track passes in silence)", path, args[0].Call("toString").String())
		}
		return nil
	})
	fsys.Call("audioURL", path).Call("then", then, catch)
	return nil
}

// playNow is play without the fs shim's audioURL: the bytes read at once.
func (h *htmlAudio) playNow(path string, loop bool, volume float64) error {
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
	h.paused = false
	h.start()
	log.Printf("[Audio] %s: playing in the browser's <audio> element (%d bytes)", path, len(b))
	return nil
}

// start calls the element's play(); its promise's rejection (autoplay
// policy) is only logged.
func (h *htmlAudio) start() {
	var catch js.Func
	catch = js.FuncOf(func(this js.Value, args []js.Value) any {
		catch.Release()
		if len(args) > 0 {
			log.Println("[Audio] browser refused playback:", args[0].Get("message").String())
		}
		return nil
	})
	h.el.Call("play").Call("catch", catch)
}

func (h *htmlAudio) stop() {
	h.gen++
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

// playing: the element has a track that has not reached its end — still
// loading counts. A track the browser holds back (autoplay policy, or pause)
// counts too: the jukebox must not take silence for a finished track and
// fetch the next one each tick.
func (h *htmlAudio) playing() bool {
	return h.el.Truthy() && !h.el.Get("ended").Bool()
}

func (h *htmlAudio) pause() {
	h.paused = true
	if h.el.Truthy() {
		h.el.Call("pause")
	}
}

func (h *htmlAudio) resume() {
	h.paused = false
	if h.el.Truthy() && h.url.Truthy() { // still loading: it starts when it has its source
		h.start()
	}
}
