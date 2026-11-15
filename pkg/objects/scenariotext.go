package objects

// ScenarioTextObject is Objects/ScenarioTextObject.h (0x06 bytes: name,
// details, padding) and its two string tables: a scenario's name and
// description in the player's language (ScenarioTextObject::load).
type ScenarioTextObject struct {
	Header  *ObjectHeader
	Name    string
	Details string
}

// ParseScenarioTextObject is ScenarioTextObject::load.
func ParseScenarioTextObject(header *ObjectHeader, data []byte) *ScenarioTextObject {
	off, name := walkStringTableLang(data, 0x06)
	_, details := walkStringTableLang(data, off)
	return &ScenarioTextObject{Header: header, Name: name, Details: details}
}
