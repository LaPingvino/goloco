package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// #include "Map/Track/TrackModSection.h"
// namespace OpenLoco::GameCommands
type TrackModsPlacementArgs struct {
// TrackModsPlacementArgs() = default;
	// method: explicit TrackModsPlacementArgs(const registers& regs)
// : pos(regs.ax, regs.cx, regs.di)
// , rotation(regs.bh & 0x3)
// , trackId(regs.dl & 0x3F)
// , index(regs.dh & 0x3)
// , type((regs.edi >> 16) & 0xF)
// , trackObjType(regs.ebp & 0xFF)
// , modSection(static_cast<World::Track::ModSection>((regs.ebp >> 16) & 0xFF))
}
const TrackModsPlacementArgsCommand any = GameCommand.createTrackMod
// World::Pos3 pos;
// orphan member: uint8_t rotation;
// orphan member: uint8_t trackId;
// orphan member: uint8_t index;
// orphan member: uint8_t type;
// orphan member: uint8_t trackObjType;
// World::Track::ModSection modSection;
// explicit operator registers() const
// orphan member: registers regs;
// regs.ax = pos.x;
// regs.cx = pos.y;
// regs.bh = rotation;
// regs.dl = trackId;
// regs.dh = index;
// regs.edi = pos.z | (type << 16);
// regs.ebp = trackObjType | (enumValue(modSection) << 16);
// orphan member: return regs;
// func CreateTrackMod(regs registers) 
