// Generates pkg/scene/testdata/trees_hashes.txt: upstream's generateTrees, Tree.cpp (getRandomTreeTypeFromSurface,
// placeRandomTree, placeTreeCluster), CreateTree, TileClearance's canConstructAt checks and TileManager::getHeight
// (all sed-copied as tt_*.inc by the extraction step) on a stub tile store, over a map dumped from goloco's
// generator just before the tree step (treein.txt — game-derived, never committed). Run: ./tt treein.txt
#include <cstdio>
#include <cstdint>
#include <cstdlib>
#include <algorithm>
#include <array>
#include <functional>
#include <limits>
#include <optional>
#include <vector>
#include <type_traits>
#include <initializer_list>
#include <OpenLoco/Core/Prng.h>
#include <OpenLoco/Math/Trigonometry.hpp>
namespace OpenLoco {
  template<class E> constexpr auto enumValue(E e) { return static_cast<std::underlying_type_t<E>>(e); }
  enum class Colour : uint8_t { black = 0 };
  using currency32_t = int32_t; using coord_t = int16_t; using tile_coord_t = int16_t;
  enum class SeasonT : uint8_t {};
  struct GameState { Core::Prng rng; uint16_t seaLevel; SeasonT currentSeason; };
  GameState& getGameState(); inline Core::Prng& gPrng1() { return getGameState().rng; }
  enum class ObjectType { tree, land };
  enum class TreeObjectFlags : uint16_t { none = 0, hasSnowVariation = 1, unk1 = 2, highAltitude = 4, lowAltitude = 8, requiresWater = 16, unk5 = 32, droughtResistant = 64, hasShadow = 128 };
  constexpr TreeObjectFlags operator|(TreeObjectFlags a, TreeObjectFlags b) { return TreeObjectFlags(uint16_t(a) | uint16_t(b)); }
  constexpr TreeObjectFlags operator&(TreeObjectFlags a, TreeObjectFlags b) { return TreeObjectFlags(uint16_t(a) & uint16_t(b)); }
  constexpr TreeObjectFlags operator~(TreeObjectFlags a) { return TreeObjectFlags(~uint16_t(a)); }
  inline TreeObjectFlags& operator|=(TreeObjectFlags& a, TreeObjectFlags b) { return a = a | b; }
  inline TreeObjectFlags& operator&=(TreeObjectFlags& a, TreeObjectFlags b) { return a = a & b; }
  struct TreeObject { TreeObjectFlags flags; uint8_t initialHeight, height, growth, currentSeason, costIndex = 0; int16_t buildCostFactor = 0, rating = 0;
    bool hasFlags(TreeObjectFlags f) const { return (flags & f) != TreeObjectFlags::none; } };
  enum class LandObjectFlags : uint8_t { none = 0, isDesert = 4, noTrees = 8 };
  struct LandObject { uint8_t flags; bool hasFlags(LandObjectFlags f) const { return flags & uint8_t(f); } };
  namespace ObjectManager { template<class T> T* get(size_t i); template<> TreeObject* get<TreeObject>(size_t); template<> LandObject* get<LandObject>(size_t); constexpr size_t getMaxObjects(ObjectType t) { return t == ObjectType::tree ? 64 : 32; } }
  namespace Scenario { struct Options { uint16_t numberOfForests; uint8_t minForestRadius, maxForestRadius, minForestDensity, maxForestDensity; uint16_t numberRandomTrees; uint8_t minAltitudeForTrees, maxAltitudeForTrees; uint8_t madeAnyChanges; };
    Options& getOptions(); uint8_t getCurrentSnowLine(); enum class Season : uint8_t {}; }
  namespace Economy { inline currency32_t getInflationAdjustedCost(int16_t, uint8_t, uint8_t) { return 0; } }
  enum class ExpenditureType { Construction };
  namespace World {
    constexpr int32_t kTileSize = 32, kMapColumns = 384, kMapRows = 384, kMapWidth = kMapColumns * kTileSize, kMapHeight = kMapRows * kTileSize;
    constexpr int16_t kSmallZStep = 4, kMicroToSmallZStep = 4, kMicroZStep = 16;
    using SmallZ = uint8_t; using MicroZ = uint8_t;
    struct Pos2 { int32_t x = 0, y = 0; constexpr Pos2() = default; constexpr Pos2(int32_t a, int32_t b) : x(a), y(b) {}
      Pos2 operator+(const Pos2& o) const { return { x + o.x, y + o.y }; } Pos2 operator-(const Pos2& o) const { return { x - o.x, y - o.y }; } };
    struct Pos3 { int32_t x, y, z; };
    struct TilePos2 { int32_t x = 0, y = 0; TilePos2() = default; TilePos2(int32_t a, int32_t b) : x(a), y(b) {}
      TilePos2 operator+(const TilePos2& o) const { return { x + o.x, y + o.y }; } TilePos2 operator-(const TilePos2& o) const { return { x - o.x, y - o.y }; } };
    inline TilePos2 toTileSpace(const Pos2& p) { return { p.x / kTileSize, p.y / kTileSize }; }
    inline Pos2 toWorldSpace(const TilePos2& p) { return { p.x * kTileSize, p.y * kTileSize }; }
    constexpr bool validCoords(const Pos2& c) { return c.x >= 0 && c.x < kMapWidth && c.y >= 0 && c.y < kMapHeight; }
    constexpr bool validCoords(const TilePos2& c) { return c.x >= 0 && c.x < kMapColumns && c.y >= 0 && c.y < kMapRows; }
    constexpr bool drawableCoord(int32_t c) { return c >= kTileSize && c < kMapWidth - kTileSize - 1; }
    constexpr bool drawableCoords(const Pos2& c) { return drawableCoord(c.x) && drawableCoord(c.y); }
    enum class MapSelectionType : uint8_t { corner0, corner1, corner2, corner3 };
    namespace SurfaceSlope { constexpr uint8_t flat = 0, doubleHeight = 1 << 4;
      namespace CornerUp { constexpr uint8_t all = 0xF, north = 1, east = 2, south = 4, west = 8; }
      namespace CornerDown { constexpr uint8_t west = CornerUp::all & ~CornerUp::west, south = CornerUp::all & ~CornerUp::south, east = CornerUp::all & ~CornerUp::east, north = CornerUp::all & ~CornerUp::north; }
      namespace SideUp { constexpr uint8_t northeast = CornerUp::north | CornerUp::east, southeast = CornerUp::south | CornerUp::east, northwest = CornerUp::north | CornerUp::west, southwest = CornerUp::south | CornerUp::west; }
      namespace Valley { constexpr uint8_t westeast = CornerUp::east | CornerUp::west, northsouth = CornerUp::north | CornerUp::south; } }
    class QuarterTile { uint8_t _val; public: constexpr QuarterTile(uint8_t t, uint8_t z) : _val(t | (z << 4)) {}
      constexpr uint8_t getBaseQuarterOccupied() const { return _val & 0xF; } constexpr uint8_t getZQuarterOccupied() const { return (_val >> 4) & 0xF; } };
    enum class ElementType : uint8_t { surface, track, station, signal, building, tree, wall, road, industry };
    enum class StationType : uint8_t { trainStation, roadStation, airport, docks };
    struct TreeElement; struct SurfaceElement; struct StationElement { StationType stationType() const { return StationType::trainStation; } };
    struct TileElementEntry {
      ElementType _type; uint8_t _baseZ, _clearZ, _flags;
      uint8_t terrain_ = 0, slope_ = 0, water_ = 0;                        // surface
      uint8_t treeType_ = 0, rot_ = 0, quad_ = 0, growth_ = 0, colour_ = 0, season_ = 0, unk7l_ = 0; bool snow_ = false, dying_ = false; // tree
      int tileIndex = 0;
      ElementType type() const { return _type; }
      uint8_t baseZ() const { return _baseZ; } uint8_t clearZ() const { return _clearZ; }
      int16_t baseHeight() const { return _baseZ * kSmallZStep; } int16_t clearHeight() const { return _clearZ * kSmallZStep; }
      void setBaseZ(uint8_t v) { _baseZ = v; } void setClearZ(uint8_t v) { _clearZ = v; }
      bool isGhost() const { return _flags & 0x10; } void setGhost(bool s) { _flags = s ? _flags | 0x10 : _flags & ~0x10; }
      bool isAiAllocated() const { return _flags & 0x20; }
      uint8_t occupiedQuarter() const { return _flags & 0xF; } void setOccupiedQuarter(uint8_t v) { _flags = (_flags & 0xF0) | (v & 0xF); }
      template<class T> T* as() const;
      template<class T> T& get() { return *as<T>(); }
    };
    struct SurfaceElement : TileElementEntry {
      uint8_t terrain() const { return terrain_; } uint8_t slope() const { return slope_ & 0x1F; } uint8_t slopeCorners() const { return slope_ & 0xF; }
      bool isSlopeDoubleHeight() const { return slope_ & 0x10; } uint8_t water() const { return water_; } int16_t waterHeight() const { return water_ * kMicroZStep; } };
    struct TreeElement : TileElementEntry {
      void setRotation(uint8_t v) { rot_ = v; } void setQuadrant(uint8_t v) { quad_ = v; } void setTreeObjectId(uint8_t v) { treeType_ = v; }
      void setGrowth(uint8_t v) { growth_ = v; } void setUnk5h(uint8_t) {} void setColour(Colour c) { colour_ = uint8_t(c); }
      void setIsDying(bool v) { dying_ = v; } bool isDying() const { return dying_; } void setSnow(bool v) { snow_ = v; }
      void setSeason(uint8_t v) { season_ = v & 7; } uint8_t season() const { return season_; } void setUnk7l(uint8_t v) { unk7l_ = v; } };
    template<class T> T* TileElementEntry::as() const {
      if constexpr (std::is_same_v<T, SurfaceElement>) return _type == ElementType::surface ? (T*)this : nullptr;
      else if constexpr (std::is_same_v<T, TreeElement>) return _type == ElementType::tree ? (T*)this : nullptr;
      else return nullptr; }
    struct Tile { std::vector<TileElementEntry>* els;
      SurfaceElement* surface() const { for (auto& e : *els) if (e._type == ElementType::surface) return (SurfaceElement*)&e; return nullptr; }
      auto begin() const { return els->begin(); } auto end() const { return els->end(); } };
    struct TilePosRangeView { TilePos2 a, b; struct It { TilePos2 p; int ax, bx; const TilePos2& operator*() const { return p; } It& operator++() { if (++p.x > bx) { p.x = ax; ++p.y; } return *this; } bool operator!=(const It& o) const { return p.x != o.p.x || p.y != o.p.y; } };
      It begin() const { return { a, a.x, b.x }; } It end() const { return { TilePos2(a.x, b.y + 1), a.x, b.x }; } };
    inline int clampT(int v) { return v < 0 ? 0 : (v > 383 ? 383 : v); }
    inline TilePosRangeView getClampedRange(const TilePos2& a, const TilePos2& b) { return { TilePos2(clampT(a.x), clampT(a.y)), TilePos2(clampT(b.x), clampT(b.y)) }; }
    inline TilePosRangeView getWorldRange() { return { TilePos2(0, 0), TilePos2(383, 383) }; }
    inline TilePosRangeView getDrawableTileRange() { return { TilePos2(1, 1), TilePos2(382, 382) }; }
    namespace TileManager {
      struct TileHeight { int16_t landHeight, waterHeight; };
      enum class ElementPositionFlags : uint8_t { none = 0, aboveGround = 1, underground = 2, partiallyUnderwater = 4, underwater = 8 };
      constexpr ElementPositionFlags operator|(ElementPositionFlags a, ElementPositionFlags b) { return ElementPositionFlags(uint8_t(a) | uint8_t(b)); }
      constexpr ElementPositionFlags operator&(ElementPositionFlags a, ElementPositionFlags b) { return ElementPositionFlags(uint8_t(a) & uint8_t(b)); }
      constexpr ElementPositionFlags operator~(ElementPositionFlags a) { return ElementPositionFlags(~uint8_t(a)); }
      inline ElementPositionFlags& operator|=(ElementPositionFlags& a, ElementPositionFlags b) { return a = a | b; }
      inline ElementPositionFlags& operator&=(ElementPositionFlags& a, ElementPositionFlags b) { return a = a & b; }
      Tile get(const Pos2& p); Tile get(const TilePos2& p); bool checkFreeElementsAndReorganise(); uint32_t numFreeElements();
      template<class T> TileElementEntry* insertElement(const Pos2& pos, uint8_t baseZ, uint8_t quarters);
      void removeElement(TileElementEntry& el);
    }
  }
  namespace GameCommands { constexpr uint32_t kFailure = 0x80000000; namespace Flags { constexpr uint8_t apply = 1, ghost = 8; }
    inline void setExpenditureType(ExpenditureType) {} inline void setPosition(World::Pos3) {} inline void setErrorText(int) {}
    struct TreePlacementArgs { World::Pos2 pos; uint8_t rotation; uint8_t type; World::MapSelectionType quadrant_; struct Q { uint8_t v; Q& operator=(World::MapSelectionType m) { v = uint8_t(m); return *this; } operator uint8_t() const { return v; } } quadrant; Colour colour; bool buildImmediately; bool requiresFullClearance; }; }
  namespace StringIds { constexpr int cant_build_this_underwater = 0, land_type_not_suitable = 0, cannot_build_partly_above_below_water = 0, off_edge_of_map = 0; }
  namespace Ui { namespace Windows { namespace Terraform { inline void setLastPlacedTree(void*) {} } } enum class ZoomLevel { eighth };
    namespace ViewportManager { inline void invalidate(World::Pos2, int16_t, int16_t, ZoomLevel, int) {} } }
  namespace TownManager { inline void updateTownInfo(World::Pos2, int, int, int16_t, int) {} }
}
// ---- upstream code ----
namespace OpenLoco::World::TileManager {
#include "tt_height.inc"
}
namespace OpenLoco::World::TileClearance {
  using OpenLoco::World::TileManager::ElementPositionFlags;
  enum class ClearFuncResult { allCollisionsRemoved, collision, collisionErrorSet, noCollision, collisionRemoved };
  enum class BuildingCollisionType : bool { standard, anyHeight };
  static ElementPositionFlags _constructAtElementPositionFlags;
  inline void setCollisionErrorMessage(const World::TileElementEntry&) {}
  static ClearFuncResult aiCompanyAboutToBuildCheck(const World::TileElementEntry&) { return ClearFuncResult::collisionErrorSet; } // neutral company
#include "tt_clear.inc"
}
namespace OpenLoco::World { MapSelectionType getQuadrantFromPos(const Pos2& loc); }
namespace OpenLoco::GameCommands {
  using namespace OpenLoco::World;
  using Ui::ZoomLevel;
#include "tt_create.inc"
  inline uint32_t doCommand(const TreePlacementArgs& a, uint8_t f) { return createTree(a, f); }
}
namespace OpenLoco::World {
#include "tt_quad.inc"
#include "tt_tree.inc"
}
namespace OpenLoco::World::MapGenerator {
#include "tt_gen.inc"
  void runTrees() { generateTrees(); }
}
// ---- stub storage ----
namespace OpenLoco {
  static GameState gs; GameState& getGameState() { return gs; }
  static TreeObject trees[64]; static bool treeOk[64]; static LandObject lands[32]; static bool landOk[32];
  template<> TreeObject* ObjectManager::get<TreeObject>(size_t i) { return treeOk[i] ? &trees[i] : nullptr; }
  template<> LandObject* ObjectManager::get<LandObject>(size_t i) { return landOk[i] ? &lands[i] : nullptr; }
  static Scenario::Options opts; Scenario::Options& Scenario::getOptions() { return opts; }
  static uint8_t snowLine; uint8_t Scenario::getCurrentSnowLine() { return snowLine; }
  static std::vector<std::vector<World::TileElementEntry>> tiles(384 * 384); static uint32_t entriesEnd;
  World::Tile World::TileManager::get(const TilePos2& p) { return { &tiles[p.y * 384 + p.x] }; }
  World::Tile World::TileManager::get(const Pos2& p) { return get(toTileSpace(p)); }
  uint32_t World::TileManager::numFreeElements() { return 442368 - entriesEnd; }
  bool World::TileManager::checkFreeElementsAndReorganise() { if (numFreeElements() > 1024) return true; uint32_t n = 0; for (auto& t : tiles) n += t.size(); entriesEnd = n; return true; }
  template<class T> World::TileElementEntry* World::TileManager::insertElement(const Pos2& pos, uint8_t baseZ, uint8_t quarters) {
    auto& t = tiles[(pos.y / 32) * 384 + pos.x / 32]; entriesEnd += t.size() + 1;
    size_t at = 0; while (at < t.size() && baseZ >= t[at]._baseZ) at++;
    TileElementEntry e{}; e._type = ElementType::tree; e._baseZ = baseZ; e._clearZ = baseZ; e._flags = quarters & 0xF; e.tileIndex = (pos.y / 32) * 384 + pos.x / 32;
    if (getenv("TTLOG") && entriesEnd < 147456 + 400) fprintf(stderr, "ins %d %d bz %d q %d\n", (pos.x / 32), (pos.y / 32), baseZ, quarters); // TTLOG=1: the first inserts
    t.insert(t.begin() + at, e); return &t[at]; }
  void World::TileManager::removeElement(TileElementEntry& el) { auto& t = tiles[el.tileIndex]; for (size_t i = 0; i < t.size(); i++) if (&t[i] == &el) { t.erase(t.begin() + i); return; } }
}
using namespace OpenLoco;
int main(int argc, char** argv) {
  FILE* f = fopen(argv[1], "r");
  unsigned nf, minR, maxR, minD, maxD, nr, minA, maxA, sl, season, ee, sea;
  fscanf(f, "%u %u %u %u %u %u %u %u %u %u %u %u", &nf, &minR, &maxR, &minD, &maxD, &nr, &minA, &maxA, &sl, &season, &ee, &sea);
  for (int i = 0; i < 64; i++) { unsigned ok, fl, ih, h, g, cs; fscanf(f, "%u %u %u %u %u %u", &ok, &fl, &ih, &h, &g, &cs); treeOk[i] = ok; trees[i] = { TreeObjectFlags(fl), (uint8_t)ih, (uint8_t)h, (uint8_t)g, (uint8_t)cs }; }
  for (int i = 0; i < 32; i++) { unsigned ok, fl; fscanf(f, "%u %u", &ok, &fl); landOk[i] = ok; lands[i] = { (uint8_t)fl }; }
  std::vector<World::TileElementEntry> base(384 * 384);
  for (int i = 0; i < 384 * 384; i++) { unsigned t, bz, s, w; fscanf(f, "%u %u %u %u", &t, &bz, &s, &w); auto& e = base[i]; e = {}; e._type = World::ElementType::surface; e._baseZ = bz; e.slope_ = s; e.water_ = w; e.terrain_ = t;
    int c = (s & 0xF) ? 4 : 0; if (s & 0x10) c += 4; e._clearZ = bz + c; e._flags = 0xF; e.tileIndex = i; }
  for (int c = 0; c < 3; c++) {
    for (int i = 0; i < 384 * 384; i++) tiles[i] = { base[i] };
    entriesEnd = ee; snowLine = sl; gs.seaLevel = sea; gs.currentSeason = SeasonT(season);
    const unsigned forests[3] = { nf, 40, 400 }, randomTrees[3] = { nr, 3000, 200 };
    opts = { (uint16_t)forests[c], (uint8_t)minR, (uint8_t)maxR, (uint8_t)minD, (uint8_t)maxD, (uint16_t)randomTrees[c], (uint8_t)minA, (uint8_t)(c == 1 ? 18 : maxA) };
    gs.rng = Core::Prng(4242u + c * 101u, 1717u + c * 13u);
    // generateTrees' updateTreeSeasons reads getGameState().currentSeason: stubbed through the dump's season
    World::MapGenerator::runTrees();
    uint64_t h = 1469598103934665603ull; unsigned count = 0;
    for (int i = 0; i < 384 * 384; i++) for (auto& e : tiles[i]) if (e._type == World::ElementType::tree) { count++;
      for (int v : { i, (int)e._baseZ, (int)e._clearZ, (int)e._flags, (int)e.treeType_, (int)e.rot_, (int)e.quad_, (int)e.growth_, (int)e.snow_, (int)e.season_, (int)e.unk7l_ }) { h ^= (uint32_t)v; h *= 1099511628211ull; } }
    printf("%d %u %u %u %u %llx\n", c, forests[c], randomTrees[c], c == 1 ? 18u : maxA, count, (unsigned long long)h);
  }
}
