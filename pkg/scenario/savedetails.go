package scenario

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"

	"github.com/LaPingvino/goloco/pkg/assets"
)

// SaveDetails is S5::SaveDetails (S5.h, 0xC618 bytes): the saved game's
// preview — company and owner names, the date, challenge progress and flags,
// and a 250×200 palette-index image.
type SaveDetails struct {
	Company, Owner    string
	Date              uint32
	ChallengeProgress uint8
	Image             []byte // 250×200
	ChallengeFlags    uint32 // CompanyFlags
}

// ReadSaveDetails reads a saved game's details: the header chunk and, when
// it has HeaderFlags::hasSaveDetails, the next chunk (as S5.cpp reads them).
func ReadSaveDetails(path string) (*SaveDetails, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	r := assets.NewS5ChunkReader(data)
	hb, err := r.ReadChunk()
	if err != nil {
		return nil, err
	}
	h, err := assets.ParseS5Header(hb)
	if err != nil {
		return nil, err
	}
	if h.Flags&assets.HeaderFlagHasSaveDetails == 0 {
		return nil, errors.New("no save details")
	}
	b, err := r.ReadChunk()
	if err != nil {
		return nil, err
	}
	if len(b) < 0xC598+4 {
		return nil, errors.New("save details too short")
	}
	cstr := func(p []byte) string {
		if i := bytes.IndexByte(p, 0); i >= 0 {
			p = p[:i]
		}
		return string(p)
	}
	return &SaveDetails{
		Company:           cstr(b[0x000:0x100]),
		Owner:             cstr(b[0x100:0x200]),
		Date:              binary.LittleEndian.Uint32(b[0x200:]),
		ChallengeProgress: b[0x246],
		Image:             append([]byte(nil), b[0x248:0x248+250*200]...),
		ChallengeFlags:    binary.LittleEndian.Uint32(b[0xC598:]),
	}, nil
}
