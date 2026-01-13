package localisation

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include <cstdint>
// #include <span>
// #include <string>
// namespace OpenLoco::Localisation
type LocoLanguageId int

const (
func init() {
	English_uk LocoLanguageId = iota
}

	English_us
	French
	German
	Spanish
	Italian
	Dutch
	Swedish
	Japanese
	Korean
	Chinese_simplified
	Chinese_traditional
	Id_12
	Portuguese
func init() {
	Blank LocoLanguageId = 254
	End   LocoLanguageId = 255

)

type LanguageDescriptor struct {
	// string locale;
	// string englishName;
	// string nativeName;
	LocoOriginalId LocoLanguageId

// func EnumerateLanguages()
// []<const LanguageDescriptor> getLanguageDescriptors();
// const LanguageDescriptor& getDescriptorForLanguage(string_view targetLocale);