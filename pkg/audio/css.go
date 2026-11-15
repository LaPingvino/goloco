package audio

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/LaPingvino/goloco/pkg/objects"
)

// Sample is one sound: its WAVEFORMATEX and PCM bytes.
type Sample struct {
	Format objects.WaveFormat
	PCM    []byte
}

// ReadCSS is Audio.cpp loadSoundsFromCSS: a uint32 count, that many uint32
// offsets, and at each offset a uint32 PCM length, an 18-byte WAVEFORMATEX and
// the PCM. css1.dat holds the UI sounds (SoundId 0…), indexed as upstream.
func ReadCSS(path string) ([]Sample, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(b) < 4 {
		return nil, fmt.Errorf("%s: too short", path)
	}
	n := int(binary.LittleEndian.Uint32(b))
	if 4+4*n > len(b) {
		return nil, fmt.Errorf("%s: %d offsets past the end", path, n)
	}
	out := make([]Sample, 0, n)
	for i := 0; i < n; i++ {
		off := int(binary.LittleEndian.Uint32(b[4+4*i:]))
		if off+4+18 > len(b) {
			return nil, fmt.Errorf("%s: sound %d past the end", path, i)
		}
		pcmLen := int(binary.LittleEndian.Uint32(b[off:]))
		f, err := objects.ParseWaveFormat(b[off+4:])
		if err != nil {
			return nil, err
		}
		pcm := off + 4 + 18
		if pcm+pcmLen > len(b) {
			return nil, fmt.Errorf("%s: sound %d's PCM past the end", path, i)
		}
		out = append(out, Sample{Format: f, PCM: b[pcm : pcm+pcmLen]})
	}
	return out, nil
}
