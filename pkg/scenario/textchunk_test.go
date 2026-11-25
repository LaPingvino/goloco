package scenario

import (
	"bytes"
	"testing"
)

// The wide texts (N56) survive a save: "?" in the 8-bit tables, the text in
// GolocoText; one the 8-bit table no longer matches stays as it is.
func TestTextChunkRoundTrip(t *testing.T) {
	sc := &Scenario{UserStrings: make([]string, userStringsCount)}
	sc.UserStrings[3] = "東京駅"
	sc.UserStrings[4] = "Plain"
	sc.UserStrings[5] = "Москва"
	sc.Messages = []MessageRecord{{Text: []byte("At \x02\xe6\x9d\xb1!")}, {Text: []byte("plain")}}
	gs := sc.EncodeGameState()
	if got := gs[userStringsOffset+3*userStringSize : userStringsOffset+3*userStringSize+4]; string(got) != "???\x00" {
		t.Fatalf("8-bit user string %q", got)
	}
	chunk := sc.EncodeTextChunk()
	if !bytes.HasPrefix(chunk, []byte(GolocoTextMagic)) {
		t.Fatal("no chunk")
	}
	back := &Scenario{UserStrings: parseUserStrings(gs)}
	back.Messages, _ = parseMessages(gs)
	if string(back.Messages[0].Text) != "At ?!" {
		t.Fatalf("8-bit message %q", back.Messages[0].Text)
	}
	back.UserStrings[5] = "Edited" // changed by another program
	back.decodeTextChunk(chunk)
	if back.UserStrings[3] != "東京駅" || back.UserStrings[4] != "Plain" || back.UserStrings[5] != "Edited" {
		t.Errorf("user strings %q %q %q", back.UserStrings[3], back.UserStrings[4], back.UserStrings[5])
	}
	if !bytes.Equal(back.Messages[0].Text, sc.Messages[0].Text) || string(back.Messages[1].Text) != "plain" {
		t.Errorf("messages %q %q", back.Messages[0].Text, back.Messages[1].Text)
	}
	if (&Scenario{UserStrings: []string{"x", "Čech"}, Messages: []MessageRecord{{Text: []byte("Caf\xe9")}}}).EncodeTextChunk() != nil {
		t.Error("a chunk without wide text")
	}
}
