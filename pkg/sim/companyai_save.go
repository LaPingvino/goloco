package sim

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"

	"github.com/LaPingvino/goloco/pkg/core"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// The company AIs in the saved game (AI-INTERFACE.md G11): goloco's own
// chunk after the tile elements (scenario.S5File.GolocoAI), never OpenLoco's
// AI bytes in the company record, which the original would misread.
// OpenLoco's loadSave stops after the tile elements, so it loads the file
// and plays the companies as AI-less ones; a save it writes back has no
// chunk, and goloco then loads those companies without an AI.
//
// Layout, little-endian: the magic "GOLOCOAI", version (u16), the number of
// companies (u8); per company: its id (u8), the AI's name (u8 length, the
// bytes), the switches (u8: bit 0 AutoBorrow), the PRNG's two words (u32
// each), the dry runs made (u64), the AI's own state (u32 length, the bytes:
// CompanyAI.Save). Not kept: the decision log, which starts empty after a
// load.

const aiChunkVersion = 1

// EncodeAIChunk is the chunk for the companies with a goloco AI, nil when
// there are none (the save is then exactly what it is without AIs).
func (m *Map) EncodeAIChunk() []byte {
	var ids []int
	for id, st := range m.companyAIs {
		if st != nil && !m.CompanyEmpty(id) {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	var b bytes.Buffer
	w := func(v any) { binary.Write(&b, binary.LittleEndian, v) }
	b.WriteString(scenario.GolocoAIMagic)
	w(uint16(aiChunkVersion))
	w(uint8(len(ids)))
	for _, id := range ids {
		st := m.companyAIs[id]
		w(uint8(id))
		w(uint8(len(st.name)))
		b.WriteString(st.name)
		var flags uint8
		if st.opts.AutoBorrow {
			flags |= 1
		}
		w(flags)
		w(st.rng.Srand0())
		w(st.rng.Srand1())
		w(st.queries)
		state := st.ai.Save()
		w(uint32(len(state)))
		b.Write(state)
	}
	return b.Bytes()
}

// DecodeAIChunk gives the companies their AIs back from a save's chunk (nil:
// none). A company the chunk names that is empty or a player's, or an AI
// that isn't registered or refuses its state, is logged and left without
// one; only a damaged chunk is an error.
func (m *Map) DecodeAIChunk(data []byte) error {
	if data == nil {
		return nil
	}
	r := bytes.NewReader(data)
	rd := func(v any) error { return binary.Read(r, binary.LittleEndian, v) }
	magic := make([]byte, len(scenario.GolocoAIMagic))
	var version uint16
	var n uint8
	if _, err := r.Read(magic); err != nil || string(magic) != scenario.GolocoAIMagic {
		return fmt.Errorf("not a goloco AI chunk")
	}
	if err := rd(&version); err != nil || version != aiChunkVersion {
		return fmt.Errorf("AI chunk version %d unknown", version)
	}
	if err := rd(&n); err != nil {
		return err
	}
	for range int(n) {
		var id, nameLen, flags uint8
		var s0, s1 uint32
		var queries uint64
		var stateLen uint32
		if err := rd(&id); err != nil {
			return err
		}
		if err := rd(&nameLen); err != nil {
			return err
		}
		name := make([]byte, nameLen)
		if _, err := r.Read(name); err != nil && nameLen > 0 {
			return err
		}
		for _, v := range []any{&flags, &s0, &s1, &queries, &stateLen} {
			if err := rd(v); err != nil {
				return err
			}
		}
		if int(stateLen) > r.Len() {
			return fmt.Errorf("AI chunk: company %d's state runs past the chunk", id)
		}
		state := make([]byte, stateLen)
		r.Read(state)
		newAI, ok := aiRegistry[string(name)]
		switch {
		case !ok:
			log.Printf("[AI] the save's company %d runs %q, which this goloco doesn't have: it stands still", id, name)
			continue
		case int(id) >= companyMaxCount || m.CompanyEmpty(int(id)) || m.IsPlayerCompany(int(id)):
			log.Printf("[AI] the save's AI for company %d has no company to run: dropped", id)
			continue
		}
		ai := newAI()
		if err := ai.Load(state); err != nil {
			log.Printf("[AI] company %d's %s state: %v: it stands still", id, name, err)
			continue
		}
		m.companyAIs[id] = &companyAI{ai: ai, name: string(name), opts: AIOptions{AutoBorrow: flags&1 != 0},
			rng: *core.NewPrngSeed(s0, s1), queries: queries}
	}
	return nil
}
