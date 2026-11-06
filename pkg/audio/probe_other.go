//go:build !(js && wasm)

package audio

import (
	"os"
	"os/exec"
	"time"

	"github.com/ebitengine/oto/v3"
)

// Item AU: with no sound device OpenLoco plays on silently, but an Ebitengine
// audio context that can't open the device ends RunGame with its error. And
// oto allows one context per process, even after a failed one — so the
// device is tried in a child process (this program, started again with
// probeEnv set) before the real context is made.

const probeEnv = "GOLOCO_AUDIO_PROBE"

// ProbeChild: when this process is the probe, try the device and exit (0: it
// opened). Call it first thing in main.
func ProbeChild() {
	if os.Getenv(probeEnv) != "1" {
		return
	}
	ctx, ready, err := oto.NewContext(&oto.NewContextOptions{SampleRate: SampleRate, ChannelCount: 2, Format: oto.FormatSignedInt16LE})
	if err != nil {
		os.Exit(1)
	}
	select {
	case <-ready:
	case <-time.After(3 * time.Second):
		os.Exit(1)
	}
	if ctx.Err() != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

// DeviceAvailable reports whether a sound device opens (in a child process;
// false when it can't be tried).
func DeviceAvailable() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	cmd := exec.Command(exe)
	cmd.Env = append(os.Environ(), probeEnv+"=1")
	done := make(chan error, 1)
	if err := cmd.Start(); err != nil {
		return false
	}
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		return err == nil
	case <-time.After(5 * time.Second):
		cmd.Process.Kill()
		return false
	}
}
