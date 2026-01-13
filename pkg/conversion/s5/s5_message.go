package s5

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include <cstdint>
// namespace OpenLoco
// forward: struct Message;
// namespace OpenLoco::S5
type Message struct {
	Type uint8
// char messageString[198];  // 0x01
	CompanyId uint8
	TimeActive uint16
// uint16_t itemSubjects[3]; // 0xCA
	Date uint32
}
// static_assert(sizeof(Message) == 0xD4);
// S5::Message exportMessage(const OpenLoco::Message& src);
// OpenLoco::Message importMessage(const S5::Message& src);
