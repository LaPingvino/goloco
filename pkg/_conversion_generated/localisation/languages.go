package localisation

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include <cstdint>
// #include <span>
// #include <string>
// namespace OpenLoco::Localisation
type LocoLanguageId int

const (
	English_uk LocoLanguageId = iota
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
	Blank LocoLanguageId = 254
	End LocoLanguageId = 255
)
type LanguageDescriptor struct {
// std::string locale;
// std::string englishName;
// std::string nativeName;
	LocoOriginalId LocoLanguageId
}
// func EnumerateLanguages() 
// std::span<const LanguageDescriptor> getLanguageDescriptors();
// const LanguageDescriptor& getDescriptorForLanguage(std::string_view targetLocale);
