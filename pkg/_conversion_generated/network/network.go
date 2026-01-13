package network

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Types.hpp"
// #include <cstdint>
// #include <string_view>
// namespace OpenLoco::GameCommands
// forward: struct registers;
// namespace OpenLoco::Network
type Client_id_t = uint32
type Port_t = uint16
const DefaultPort port_t = 11754
const MaxPacketSize uint16 = 4096
const NetworkVersion uint16 = 1
// func OpenServer() 
// func JoinServer(host string) bool
// func JoinServer(host string, port port_t) bool
// func Close() 
// func Update() 
// func SendChatMessage(message string) 
// func ReceiveChatMessage(client client_id_t, message string) 
// func QueueGameCommand(company CompanyId, regs OpenLoco::GameCommands::registers) 
// func ShouldProcessTick(tick uint32) bool
// func ProcessGameCommands(tick uint32) 
// /**
// * Whether the game state is networked.
// * This will return false if the client is still receiving the map from the server.
// */
// func IsConnected() bool
// /**
// * Gets the current tick the server is on.
// */
// func GetServerTick() uint32
