// Generates pkg/scene/testdata/terrain_styles_hashes.txt: ./tg3 hs1.txt land2.txt — upstream MapGenerator.cpp generateLand … generateTerrain and generateSurfaceVariation (sed-copied) with stub tiles / land objects; hs1.txt and land2.txt are dumped from the game data by a throwaway test (never committed).
// Generates the expected hash for pkg/sim TestTerrainHeightMapMatchesUpstream: upstream's
// OriginalTerrainGenerator.cpp (includes stripped by sed, stubs below) on HS1's height-map images
// read from a dump (game data, not committed) with a fixed Prng seed.
#include <cstdio>
#include <cstdint>
#include <algorithm>
#include <array>
#include <functional>
#include <vector>
#include <type_traits>
#include <optional>
#include <OpenLoco/Core/Prng.h>
namespace OpenLoco {
  using tile_coord_t = int16_t;
  template<class E> constexpr auto enumValue(E e) { return static_cast<std::underlying_type_t<E>>(e); }
  struct TilePos2 { int32_t x, y; TilePos2(int32_t a, int32_t b) : x(a), y(b) {} TilePos2 operator+(const TilePos2& o) const { return TilePos2(x+o.x, y+o.y); } TilePos2 operator-(const TilePos2& o) const { return TilePos2(x-o.x, y-o.y); } };
  namespace Gfx { struct G1Element { uint8_t* offset; int16_t width, height; }; G1Element* getG1Element(uint32_t); }
  namespace Scenario {
    enum class ScenarioFlags : uint16_t { none = 0, hillsEdgeOfMap = 2 };
    constexpr ScenarioFlags operator&(ScenarioFlags a, ScenarioFlags b) { return ScenarioFlags(uint16_t(a) & uint16_t(b)); }
    enum class TopographyStyle : uint8_t {};
    struct Options { uint8_t minLandHeight; TopographyStyle topographyStyle; uint8_t hillDensity; ScenarioFlags scenarioFlags; uint8_t numRiverbeds, minRiverWidth, maxRiverWidth, riverbankWidth, riverMeanderRate; };
  }
  enum class HillShapeFlags : uint16_t { none = 0, isHeightMap = 1 };
  constexpr HillShapeFlags operator&(HillShapeFlags a, HillShapeFlags b) { return HillShapeFlags(uint16_t(a) & uint16_t(b)); }
  struct HillShapesObject { uint8_t hillHeightMapCount, mountainHeightMapCount; uint32_t image; HillShapeFlags flags; };
  struct LandObject; namespace ObjectManager { template<class T> T* get(); template<> HillShapesObject* get<HillShapesObject>(); template<class T> T* get(size_t); }
  struct GameState { Core::Prng rng; uint16_t seaLevel; uint8_t lastLandOption; };
  GameState& getGameState();
  namespace World {
    constexpr int kMapPitch = 512;
    constexpr bool validCoords(const TilePos2& c) { return int16_t(c.x) >= 0 && int16_t(c.x) < 384 && int16_t(c.y) >= 0 && int16_t(c.y) < 384; }
    namespace MapGenerator {
      struct HeightMap {
        std::vector<uint8_t> h; uint16_t width, height;
        HeightMap(int w, int hh) : h(w * hh), width(w), height(hh) {}
        uint8_t& operator[](TilePos2 p) { return h[p.y * width + p.x]; }
        uint8_t* data() { return h.data(); } size_t size() { return h.size(); }
        uint8_t getHeight(TilePos2 p) { return (*this)[p] & 0x7F; } bool isMarkerSet(TilePos2 p) { return (*this)[p] & 0x80; }
        void setMarker(TilePos2 p) { (*this)[p] |= 0x80; } void resetMarkerFlags() { for (auto& v : h) v &= 0x7F; }
      };
      enum class TopographyFlags : uint8_t { none = 0, hasHills = 1, hasMountains = 2 };
      constexpr TopographyFlags operator|(TopographyFlags a, TopographyFlags b) { return TopographyFlags(uint8_t(a) | uint8_t(b)); }
      constexpr TopographyFlags operator&(TopographyFlags a, TopographyFlags b) { return TopographyFlags(uint8_t(a) & uint8_t(b)); }
    static constexpr std::array<TopographyFlags, 10> kTopographyStyleFlags = {
        TopographyFlags::none,
        TopographyFlags::none,
        TopographyFlags::hasHills,
        TopographyFlags::hasHills,
        TopographyFlags::hasMountains | TopographyFlags::hasHills,
        TopographyFlags::hasMountains | TopographyFlags::hasHills,
        TopographyFlags::hasHills,
        TopographyFlags::hasMountains | TopographyFlags::hasHills,
        TopographyFlags::none,
        TopographyFlags::hasMountains | TopographyFlags::hasHills,
    };
    class OriginalTerrainGenerator
    {
    private:
        // 0x004626B7
        void blitHill(const Scenario::Options& options, HeightMap& heightMap);

        static void blitImageFlippedXFlippedY(Gfx::G1Element* g1Element, uint8_t featureWidth, uint8_t featureHeight, HeightMap& heightMap, tile_coord_t randX, tile_coord_t randY);

        static void blitImageFlippedXNormalY(Gfx::G1Element* g1Element, uint8_t featureWidth, uint8_t featureHeight, HeightMap& heightMap, tile_coord_t randX, tile_coord_t randY);

        static void blitImageNormalXFlippedY(Gfx::G1Element* g1Element, uint8_t featureWidth, uint8_t featureHeight, HeightMap& heightMap, tile_coord_t randX, tile_coord_t randY);

        static void blitImageNormalXNormalY(Gfx::G1Element* g1Element, uint8_t featureWidth, uint8_t featureHeight, HeightMap& heightMap, tile_coord_t randX, tile_coord_t randY);

        // 0x00462518
        void generateHills(const Scenario::Options& options, HeightMap& heightMap);

        // 0x00462556
        void copyHeightMapFromG1(Gfx::G1Element* g1Element, HeightMap& heightMap);

        // 0x00462590
        void capSeaLevels(HeightMap& heightMap);

    public:
        // 0x004624F0
        void generate(const Scenario::Options& options, HeightMap& heightMap);
    };
}}}

namespace OpenLoco::World::MapGenerator
{
    void OriginalTerrainGenerator::blitHill(const Scenario::Options& options, HeightMap& heightMap)
    {
        const uint32_t randomVal = getGameState().rng.randNext();
        const bool flipHillImageLateral = randomVal & 0x100;
        const bool flipHillImageVertical = randomVal & 0x200;

        const auto hillShapesObj = ObjectManager::get<HillShapesObject>();
        const uint8_t hillShapeCount = hillShapesObj->hillHeightMapCount + hillShapesObj->mountainHeightMapCount;

        // The hill index calculation is a minor simplification compared to vanilla
        const uint8_t randomHillIndex = (randomVal & 0xFF) % hillShapeCount;

        const auto hillImage = hillShapesObj->image + randomHillIndex;
        const auto g1Element = Gfx::getG1Element(hillImage);

        auto featureWidth = g1Element->width & 0xFF;
        auto featureHeight = g1Element->height & 0xFF;
        if (flipHillImageLateral)
        {
            std::swap(featureWidth, featureHeight);
        }

        // 0x00462718
        const uint32_t randX = getGameState().rng.randNext(heightMap.width - 1);
        const uint32_t randY = getGameState().rng.randNext(heightMap.height - 1);
        if (randX < 2 || randY < 2)
        {
            return;
        }

        // Vanilla would use an `rcl` for the lower bit, getting the carry flag in.
        // We're substituting it with an extra bit of the random state.
        const bool topographyBit = randX > randY;
        const auto topographyId = (enumValue(options.topographyStyle) << 1) | (topographyBit ? 1U : 0);
        const auto topographyFlags = kTopographyStyleFlags[topographyId];

        const bool generateHills = (topographyFlags & TopographyFlags::hasHills) != TopographyFlags::none;
        const bool generateMountains = (topographyFlags & TopographyFlags::hasMountains) != TopographyFlags::none;

        const bool isValidHillIndex = generateHills && randomHillIndex < hillShapesObj->hillHeightMapCount;
        const bool isValidMountainIndex = generateMountains && randomHillIndex >= hillShapesObj->hillHeightMapCount;
        if (!(isValidHillIndex || isValidMountainIndex))
        {
            return;
        }

        // 0x0046276D

        if ((options.scenarioFlags & Scenario::ScenarioFlags::hillsEdgeOfMap) == Scenario::ScenarioFlags::none)
        {
            if ((randX + featureWidth >= heightMap.width - 2U) || (randY + featureHeight >= heightMap.height - 2U))
            {
                return;
            }
        }

        using BlitFunction = std::function<void(Gfx::G1Element*, uint8_t, uint8_t, HeightMap&, tile_coord_t, tile_coord_t)>;
        static const std::array<std::array<BlitFunction, 2>, 2> blitMethods = { {
            {
                blitImageNormalXNormalY,
                blitImageNormalXFlippedY,
            },
            {
                blitImageFlippedXNormalY,
                blitImageFlippedXFlippedY,
            },
        } };

        auto blitFeature = blitMethods[flipHillImageLateral][flipHillImageVertical];
        blitFeature(g1Element, featureWidth, featureHeight, heightMap, randX, randY);
    }

    void OriginalTerrainGenerator::blitImageFlippedXFlippedY(Gfx::G1Element* g1Element, uint8_t featureWidth, uint8_t featureHeight, HeightMap& heightMap, tile_coord_t randX, tile_coord_t randY)
    {
        auto* src = g1Element->offset;
        int32_t x = randX;
        for (auto j = 0; j < featureWidth; ++j)
        {
            int32_t y = (featureHeight + randY - 1) % (heightMap.height - 1);
            for (auto i = 0; i < featureHeight; ++i)
            {
                const auto data = *src++;
                heightMap[TilePos2(x, y)] = std::max(data, heightMap[TilePos2(x, y)]);
                auto mod = heightMap.height - 1;
                y = (y - 1 + mod) % mod;
            }
            x++;
            x %= heightMap.width - 1;
        }
    }

    void OriginalTerrainGenerator::blitImageFlippedXNormalY(Gfx::G1Element* g1Element, uint8_t featureWidth, uint8_t featureHeight, HeightMap& heightMap, tile_coord_t randX, tile_coord_t randY)
    {
        auto* src = g1Element->offset;
        int32_t x = randX;
        for (auto j = 0; j < featureWidth; ++j)
        {
            int32_t y = randY;
            for (auto i = 0; i < featureHeight; ++i)
            {
                const auto data = *src++;
                heightMap[TilePos2(x, y)] = std::max(data, heightMap[TilePos2(x, y)]);
                y++;
                y %= heightMap.height - 1;
            }
            x++;
            x %= heightMap.width - 1;
        }
    }

    void OriginalTerrainGenerator::blitImageNormalXFlippedY(Gfx::G1Element* g1Element, uint8_t featureWidth, uint8_t featureHeight, HeightMap& heightMap, tile_coord_t randX, tile_coord_t randY)
    {
        auto* src = g1Element->offset;
        int32_t y = randY;
        for (auto j = 0; j < featureHeight; ++j)
        {
            int32_t x = (randX + featureWidth - 1) % (heightMap.width - 1);
            for (auto i = 0; i < featureWidth; ++i)
            {
                const auto data = *src++;
                heightMap[TilePos2(x, y)] = std::max(data, heightMap[TilePos2(x, y)]);
                auto mod = heightMap.width - 1;
                x = (x - 1 + mod) % mod;
            }
            y++;
            y %= heightMap.height - 1;
        }
    }

    void OriginalTerrainGenerator::blitImageNormalXNormalY(Gfx::G1Element* g1Element, uint8_t featureWidth, uint8_t featureHeight, HeightMap& heightMap, tile_coord_t randX, tile_coord_t randY)
    {
        auto* src = g1Element->offset;
        int32_t y = randY;
        for (auto j = 0; j < featureHeight; ++j)
        {
            int32_t x = randX;
            for (auto i = 0; i < featureWidth; ++i)
            {
                const auto data = *src++;

                heightMap[TilePos2(x, y)] = std::max(data, heightMap[TilePos2(x, y)]);

                x++;
                x %= heightMap.width - 1;
            }
            y++;
            y %= heightMap.height - 1;
        }
    }

    // 0x00462518
    void OriginalTerrainGenerator::generateHills(const Scenario::Options& options, HeightMap& heightMap)
    {
        uint8_t baseCount = getGameState().rng.randNext() & 7;
        uint16_t numHills = baseCount + (options.hillDensity * 5) + 1;

        for (uint16_t i = 0; i < numHills; i++)
        {
            blitHill(options, heightMap);
        }
    }

    // 0x00462556
    void OriginalTerrainGenerator::copyHeightMapFromG1(Gfx::G1Element* g1Element, HeightMap& heightMap)
    {
        auto* src = g1Element->offset;

        for (auto y = heightMap.height - 1; y > 0; y--)
        {
            for (auto x = heightMap.width - 1; x > 0; x--)
            {
                auto height = std::max<uint8_t>(*src, heightMap[TilePos2(x, y)]);
                heightMap[TilePos2(x, y)] = height;
                src++;
            }

            src += kMapPitch;
        }
    }

    // 0x00462590
    void OriginalTerrainGenerator::capSeaLevels(HeightMap& heightMap)
    {
        const auto seaLevel = getGameState().seaLevel;

        for (auto y = 0; y < heightMap.height - 2; y++)
        {
            for (auto x = 0; x < heightMap.width - 2; x++)
            {
                if (seaLevel != heightMap[TilePos2(x + 0, y + 0)])
                {
                    continue;
                }

                if (seaLevel != heightMap[TilePos2(x + 1, y + 0)])
                {
                    continue;
                }

                if (seaLevel != heightMap[TilePos2(x + 0, y + 1)])
                {
                    continue;
                }

                if (seaLevel != heightMap[TilePos2(x + 1, y + 1)])
                {
                    continue;
                }

                heightMap[TilePos2(x + 0, y + 0)] += 1;
                heightMap[TilePos2(x + 1, y + 0)] += 1;
                heightMap[TilePos2(x + 0, y + 1)] += 1;
                heightMap[TilePos2(x + 1, y + 1)] += 1;
            }
        }
    }

    void OriginalTerrainGenerator::generate(const Scenario::Options& options, HeightMap& heightMap)
    {
        std::fill_n(heightMap.data(), heightMap.size(), options.minLandHeight);

        auto hillShapesObj = ObjectManager::get<HillShapesObject>();
        if ((hillShapesObj->flags & HillShapeFlags::isHeightMap) != HillShapeFlags::none)
        {
            auto g1Element = Gfx::getG1Element(hillShapesObj->image);
            copyHeightMapFromG1(g1Element, heightMap);
        }
        else
        {
            generateHills(options, heightMap);
        }

        capSeaLevels(heightMap);
    }
}


namespace OpenLoco {
  struct Pos2 { int32_t x, y; Pos2 operator+(const Pos2& o) const { return {x+o.x, y+o.y}; } Pos2& operator+=(const Pos2& o) { x+=o.x; y+=o.y; return *this; } };
  constexpr Pos2 kRotationOffset[4] = { {-32,0}, {0,32}, {32,0}, {0,-32} };
  enum class ObjectType { land, hillShapes };
  enum class LandObjectFlags : uint8_t { none = 0, hasGrowthStages = 1 };
  struct LandObject { uint8_t numVariations, variationLikelihood, flags, numGrowthStages; bool hasFlags(LandObjectFlags f) const { return flags & uint8_t(f); } };
  namespace Scenario { enum class LandDistributionPattern : uint8_t { everywhere, nowhere, farFromWater, nearWater, onMountains, farFromMountains, inSmallRandomAreas, inLargeRandomAreas, aroundCliffs };
    struct GenOptions { LandDistributionPattern landDistributionPatterns[32]; }; GenOptions& getOptions(); uint8_t getCurrentSnowLine(); }
  struct SurfaceElement { uint8_t t=0, var=0, g=0, w=0, sl=0; int bz=0, cz=0;
    uint8_t snow=0; void setSnowCoverage(uint8_t v){snow=v;} void setTerrain(uint8_t v){t=v;} uint8_t terrain() const {return t;} void setVariation(uint8_t v){var=v;} void setGrowthStage(uint8_t v){g=v;}
    uint8_t water() const {return w;} void setWater(uint8_t v){w=v&0x1F;} bool isIndustrial() const {return false;}
    uint8_t slope() const {return sl;} uint8_t slopeCorners() const {return sl&0xF;} bool isSlopeDoubleHeight() const {return sl&0x10;}
    void setSlope(uint8_t v){sl=v;} int baseZ() const {return bz;} void setBaseZ(int v){bz=v;} void setClearZ(int v){cz=v;} };
  struct Tile { SurfaceElement* s; SurfaceElement* surface() const { return s; } };
  namespace World {
    struct TilePosRange { TilePos2 a, b; struct It { TilePos2 p; int ax, bx; const TilePos2& operator*() const { return p; } It& operator++(){ if(++p.x>bx){p.x=ax;++p.y;} return *this; } bool operator!=(const It& o) const { return p.x!=o.p.x||p.y!=o.p.y; } };
      It begin() const { return {a, a.x, b.x}; } It end() const { return {TilePos2(a.x, b.y+1), a.x, b.x}; } };
    inline int clampT(int v){ return v<0?0:(v>383?383:v); }
    inline TilePosRange getDrawableTileRange(){ return {TilePos2(1,1), TilePos2(382,382)}; }
    inline TilePosRange getWorldRange(){ return {TilePos2(0,0), TilePos2(383,383)}; }
    inline TilePosRange getClampedRange(TilePos2 a, TilePos2 b){ return {TilePos2(clampT(a.x),clampT(a.y)), TilePos2(clampT(b.x),clampT(b.y))}; }
    inline Pos2 toWorldSpace(TilePos2 p){ return {p.x*32, p.y*32}; }
    constexpr bool validCoords(const Pos2& c){ return c.x>=0 && c.x<384*32 && c.y>=0 && c.y<384*32; }
    constexpr int kMapColumns = 384, kMapRows = 384, kMicroToSmallZStep = 4, kSmallZStep = 4;
    using MicroZ = uint8_t;
    namespace TileManager { Tile get(TilePos2); Tile get(Pos2); int getSurfaceCornerHeight(const SurfaceElement& s); }
    namespace SurfaceSlope { constexpr uint8_t flat = 0, doubleHeight = 1<<4;
      namespace CornerUp { constexpr uint8_t all=0xF, north=1, east=2, south=4, west=8; }
      namespace CornerDown { constexpr uint8_t west = CornerUp::all & ~CornerUp::west, south = CornerUp::all & ~CornerUp::south, east = CornerUp::all & ~CornerUp::east, north = CornerUp::all & ~CornerUp::north; } }
  }
  namespace ObjectManager { template<> LandObject* get<LandObject>(size_t); constexpr size_t getMaxObjects(ObjectType){ return 32; } }
  inline Core::Prng& gPrng1(){ return getGameState().rng; }
}

namespace OpenLoco::World::MapGenerator {
  using namespace OpenLoco::World; static void updateProgress(int) {}
  static constexpr auto kCliffTerrainHeightDiff = 4;
  static constexpr auto kMountainTerrainHeight = 26;
  using TilePosRangeView = TilePosRange;
  using MicroZ = uint8_t;
    static void generateRivers(const Scenario::Options& options, HeightMap& heightMap)
    {
        for (auto i = 0; i < options.numRiverbeds; i++)
        {
            auto& gs = getGameState();
            const auto riverEastWest = gs.rng.randBool();
            const auto riverWidth = gs.rng.randNext(options.minRiverWidth, options.maxRiverWidth);
            const auto riverbedHeight = std::max<uint8_t>(gs.seaLevel > 0 ? gs.seaLevel - 1 : 1, options.minLandHeight);

            // We'll be varying the bank width as we meander
            auto riverbankWidth = options.riverbankWidth;
            auto totalRiverWidth = riverWidth + 2 * riverbankWidth;
            auto easternBankOffset = riverWidth + riverbankWidth;

            // Pivot: generate a random X position
            auto xStartPos = getGameState().rng.randNext(0.15 * heightMap.width, 0.85 * heightMap.width);
            for (auto yPos = 0; yPos < heightMap.height; yPos++)
            {
                for (auto xOffset = 0; xOffset < totalRiverWidth; xOffset++)
                {
                    auto pos = TilePos2(xStartPos + xOffset, yPos);
                    if (!riverEastWest)
                    {
                        pos = TilePos2(pos.y, pos.x);
                    }

                    if (!validCoords(pos))
                    {
                        // We might meander back to a valid position later,
                        // so we're only breaking out of the inner loop.
                        break;
                    }

                    if (riverbankWidth > 0 && xOffset < riverbankWidth)
                    {
                        // Western riverbank (high to low)
                        auto bankPos = riverbankWidth - xOffset;
                        auto bankHeight = heightMap[pos] * bankPos / riverbankWidth;
                        heightMap[pos] = std::max<uint8_t>(riverbedHeight, bankHeight);
                    }
                    else if (riverbankWidth > 0 && xOffset > easternBankOffset)
                    {
                        // Eastern riverbank (low to high)
                        auto bankPos = xOffset - easternBankOffset;
                        auto bankHeight = heightMap[pos] * bankPos / riverbankWidth;
                        heightMap[pos] = std::max<uint8_t>(riverbedHeight, bankHeight);
                    }
                    else
                    {
                        // Simply carve out the river
                        heightMap[pos] = riverbedHeight;
                    }
                }

                // Let the river meander slightly
                const auto meanderRate = options.riverMeanderRate;
                if (meanderRate > 0 && yPos % 4 == 0)
                {
                    const auto halfMeanderRate = meanderRate / 2;

                    int8_t meanderOffset = getGameState().rng.randNext(0, meanderRate) - halfMeanderRate;
                    xStartPos += meanderOffset;

                    // Adjust bank width slightly as well
                    if (options.riverbankWidth > 0 && halfMeanderRate != 0)
                    {
                        riverbankWidth += meanderOffset / halfMeanderRate;
                        easternBankOffset += meanderOffset / halfMeanderRate;
                        totalRiverWidth += meanderOffset / halfMeanderRate * 2;
                    }
                }
            }
        }
    }


    static void generateLand(HeightMap& heightMap)
    {
        for (auto& pos : World::getDrawableTileRange())
        {
            const MicroZ q00 = heightMap[pos + TilePos2{ -1, -1 }];
            const MicroZ q01 = heightMap[pos + TilePos2{ 0, -1 }];
            const MicroZ q10 = heightMap[pos + TilePos2{ -1, 0 }];
            const MicroZ q11 = heightMap[pos + TilePos2{ 0, 0 }];

            const auto tile = TileManager::get(pos);
            auto* surfaceElement = tile.surface();
            if (surfaceElement == nullptr)
            {
                continue;
            }

            const MicroZ baseHeight = std::min({ q00, q01, q10, q11 });
            surfaceElement->setBaseZ(baseHeight * kMicroToSmallZStep);

            uint8_t currentSlope = SurfaceSlope::flat;

            // First, figure out basic corner style
            if (q00 > baseHeight)
            {
                currentSlope |= SurfaceSlope::CornerUp::south;
            }
            if (q01 > baseHeight)
            {
                currentSlope |= SurfaceSlope::CornerUp::east;
            }
            if (q10 > baseHeight)
            {
                currentSlope |= SurfaceSlope::CornerUp::west;
            }
            if (q11 > baseHeight)
            {
                currentSlope |= SurfaceSlope::CornerUp::north;
            }

            // Now, deduce if we should go for double height
            // clang-format off
            if ((currentSlope == SurfaceSlope::CornerDown::north && q00 - baseHeight >= 2) ||
                (currentSlope == SurfaceSlope::CornerDown::west  && q01 - baseHeight >= 2) ||
                (currentSlope == SurfaceSlope::CornerDown::east  && q10 - baseHeight >= 2) ||
                (currentSlope == SurfaceSlope::CornerDown::south && q11 - baseHeight >= 2))
            {
                currentSlope |= SurfaceSlope::doubleHeight;
            }
            // clang-format on

            surfaceElement->setSlope(currentSlope);

            auto clearZ = surfaceElement->baseZ();
            if (surfaceElement->slopeCorners())
            {
                clearZ += kSmallZStep;
            }
            if (surfaceElement->isSlopeDoubleHeight())
            {
                clearZ += kSmallZStep;
            }
            surfaceElement->setClearZ(clearZ);
        }
    }

    // 0x004C4BD7
    static void generateWater([[maybe_unused]] HeightMap& heightMap)
    {
        auto seaLevel = getGameState().seaLevel;

        for (auto& pos : World::getDrawableTileRange())
        {
            auto tile = TileManager::get(pos);
            auto* surface = tile.surface();

            if (surface != nullptr && surface->baseZ() < (seaLevel << 2))
            {
                surface->setWater(seaLevel);
            }
        }
    }

    static std::optional<uint8_t> getEverywhereSurfaceStyle()
    {
        for (uint8_t landObjectIdx = 0; landObjectIdx < ObjectManager::getMaxObjects(ObjectType::land); ++landObjectIdx)
        {
            auto* landObj = ObjectManager::get<LandObject>(landObjectIdx);
            if (landObj == nullptr)
            {
                continue;
            }
            if (Scenario::getOptions().landDistributionPatterns[landObjectIdx] == Scenario::LandDistributionPattern::everywhere)
            {
                return landObjectIdx;
            }
        }
        if (getGameState().lastLandOption != 0xFF)
        {
            return getGameState().lastLandOption;
        }
        return std::nullopt;
    }

    // 0x00469FC8
    std::optional<uint8_t> getRandomTerrainVariation(const SurfaceElement& surface)
    {
        if (surface.water())
        {
            return std::nullopt;
        }
        if (surface.isIndustrial())
        {
            return std::nullopt;
        }

        auto* landObj = ObjectManager::get<LandObject>(surface.terrain());
        if (landObj == nullptr)
        {
            return std::nullopt;
        }

        if (landObj->numVariations == 0 || surface.slope())
        {
            return 0;
        }

        // TODO: split into two randNext calls
        uint16_t randVal = gPrng1().randNext();
        if (landObj->variationLikelihood <= (randVal >> 8))
        {
            return 0;
        }
        return ((randVal & 0xFF) * landObj->numVariations) >> 8;
    }

    static void applySurfaceStyleToMarkedTiles(HeightMap& heightMap, uint8_t surfaceStyle, bool requireMark)
    {
        for (auto& pos : World::getDrawableTileRange())
        {
            const bool tileIsMarked = heightMap.isMarkerSet({ pos.x, pos.y });
            if (requireMark != tileIsMarked)
            {
                continue;
            }

            auto tile = TileManager::get(pos);
            auto* surface = tile.surface();
            if (surface == nullptr)
            {
                continue;
            }

            surface->setTerrain(surfaceStyle);
            auto res = getRandomTerrainVariation(*surface);
            if (res)
            {
                surface->setVariation(*res);
            }
        }
    }

    // 0x0046A379
    static void generateTerrainFarFromWater(HeightMap& heightMap, uint8_t surfaceStyle)
    {
        heightMap.resetMarkerFlags();

        // Mark tiles near water
        auto seaLevel = getGameState().seaLevel;
        for (auto pos : getWorldRange())
        {
            auto height = heightMap.getHeight({ pos.x, pos.y });
            if (height > seaLevel)
            {
                continue;
            }

            for (auto lookaheadPos : getClampedRange(pos - TilePos2(25, 25), pos + TilePos2(25, 25)))
            {
                heightMap.setMarker({ lookaheadPos.x, lookaheadPos.y });
            }
        }

        // Apply surface style to tiles that have *not* been marked
        applySurfaceStyleToMarkedTiles(heightMap, surfaceStyle, false);
    }

    // 0x0046A439
    static void generateTerrainNearWater(HeightMap& heightMap, uint8_t surfaceStyle)
    {
        heightMap.resetMarkerFlags();

        // Mark tiles near water
        auto seaLevel = getGameState().seaLevel;
        for (auto pos : getWorldRange())
        {
            auto height = heightMap.getHeight({ pos.x, pos.y });
            if (height < seaLevel)
            {
                continue;
            }

            for (auto lookaheadPos : getClampedRange(pos - TilePos2(25, 25), pos + TilePos2(25, 25)))
            {
                heightMap.setMarker({ lookaheadPos.x, lookaheadPos.y });
            }
        }

        // Apply surface style to tiles that have been marked
        applySurfaceStyleToMarkedTiles(heightMap, surfaceStyle, true);
    }

    // 0x0046A5B3
    static void generateTerrainOnMountains(HeightMap& heightMap, uint8_t surfaceStyle)
    {
        heightMap.resetMarkerFlags();

        // Mark tiles above mountain level
        for (auto pos : getWorldRange())
        {
            // NB: this is an inclusive check to match vanilla
            auto height = heightMap.getHeight({ pos.x, pos.y });
            if (height <= kMountainTerrainHeight)
            {
                continue;
            }

            for (auto lookaheadPos : getClampedRange(pos - TilePos2(12, 12), pos + TilePos2(12, 12)))
            {
                heightMap.setMarker({ lookaheadPos.x, lookaheadPos.y });
            }
        }

        // Apply surface style to tiles that have been marked
        applySurfaceStyleToMarkedTiles(heightMap, surfaceStyle, true);
    }

    // 0x0046A4F9
    static void generateTerrainFarFromMountains(HeightMap& heightMap, uint8_t surfaceStyle)
    {
        heightMap.resetMarkerFlags();

        // Mark tiles above mountain level
        for (auto pos : getWorldRange())
        {
            // NB: this is an exclusive check to match vanilla
            auto height = heightMap.getHeight({ pos.x, pos.y });
            if (height < kMountainTerrainHeight)
            {
                continue;
            }

            for (auto lookaheadPos : getClampedRange(pos - TilePos2(25, 25), pos + TilePos2(25, 25)))
            {
                heightMap.setMarker({ lookaheadPos.x, lookaheadPos.y });
            }
        }

        // Apply surface style to tiles that have *not* been marked
        applySurfaceStyleToMarkedTiles(heightMap, surfaceStyle, false);
    }

    // 0x0046A0D8, 0x0046A227
    static void generateTerrainInRandomAreas(uint8_t surfaceStyle, uint16_t minTiles, uint16_t maxTiles)
    {
        auto numberOfAreas = getGameState().rng.randNext(80, 143);
        for (auto i = 0; i < numberOfAreas; i++)
        {
            // TODO: could probably simplify / replace with two randNext(lo, hi) calls
            auto randPos = getGameState().rng.randNext();
            auto xPos = ((randPos & 0xFFFF) * kMapColumns) >> 16;
            auto yPos = ((randPos >> 16) * kMapRows) >> 16;

            auto pos = World::toWorldSpace(TilePos2(xPos, yPos));
            auto numberOfTiles = getGameState().rng.randNext(minTiles, maxTiles - 1);
            for (auto j = 0; j < numberOfTiles; j++)
            {
                if (validCoords(pos))
                {
                    auto tile = TileManager::get(pos);
                    auto surface = tile.surface();
                    if (surface != nullptr)
                    {
                        auto height = TileManager::getSurfaceCornerHeight(*surface) / kMicroToSmallZStep;
                        if (height > getGameState().seaLevel)
                        {
                            surface->setTerrain(surfaceStyle);
                            auto variation = getRandomTerrainVariation(*surface);
                            if (variation.has_value())
                            {
                                surface->setVariation(variation.value());
                            }
                        }
                    }
                }

                auto randRotation = getGameState().rng.randNext(3);
                auto offset = kRotationOffset[randRotation];
                pos += offset;
            }
        }
    }

    // 0x0046A0D8
    static void generateTerrainInSmallRandomAreas([[maybe_unused]] HeightMap& heightMap, uint8_t surfaceStyle)
    {
        generateTerrainInRandomAreas(surfaceStyle, 24, 56);
    }

    // 0x0046A227
    static void generateTerrainInLargeRandomAreas([[maybe_unused]] HeightMap& heightMap, uint8_t surfaceStyle)
    {
        generateTerrainInRandomAreas(surfaceStyle, 128, 388);
    }

    // 0x0046A66D
    static void generateTerrainAroundCliffs(HeightMap& heightMap, uint8_t surfaceStyle)
    {
        heightMap.resetMarkerFlags();

        // Mark tiles with sudden height changes in the next row
        for (auto pos : getDrawableTileRange())
        {
            auto heightA = heightMap.getHeight({ pos + TilePos2{ 0, 1 } });
            auto heightB = heightMap.getHeight({ pos + TilePos2{ 0, 1 } });

            // Find no cliff between A and B?
            if (std::abs(heightB - heightA) < kCliffTerrainHeightDiff)
            {
                auto heightC = heightMap.getHeight({ pos + TilePos2{ 0, 1 } });
                auto heightD = heightMap.getHeight({ pos + TilePos2{ 0, 1 } });

                // Find no cliff between C and D?
                if (std::abs(heightD - heightC) < kCliffTerrainHeightDiff)
                {
                    continue;
                }
            }

            // Found a cliff around this point, so mark the points around it
            for (auto lookaheadPos : getClampedRange(pos - TilePos2(6, 6), pos + TilePos2(6, 6)))
            {
                heightMap.setMarker({ lookaheadPos.x, lookaheadPos.y });
            }
        }

        // Apply surface style to tiles that have been marked
        applySurfaceStyleToMarkedTiles(heightMap, surfaceStyle, true);
    }

    static void generateTerrainNull([[maybe_unused]] HeightMap& heightMap, [[maybe_unused]] uint8_t surfaceStyle) {}

    using GenerateTerrainFunc = void (*)(HeightMap&, uint8_t);
    static const GenerateTerrainFunc _generateFuncs[] = {
        generateTerrainNull,               // LandDistributionPattern::everywhere This is null as it is a special function performed separately
        generateTerrainNull,               // LandDistributionPattern::nowhere
        generateTerrainFarFromWater,       // LandDistributionPattern::farFromWater
        generateTerrainNearWater,          // LandDistributionPattern::nearWater
        generateTerrainOnMountains,        // LandDistributionPattern::onMountains
        generateTerrainFarFromMountains,   // LandDistributionPattern::farFromMountains
        generateTerrainInSmallRandomAreas, // LandDistributionPattern::inSmallRandomAreas
        generateTerrainInLargeRandomAreas, // LandDistributionPattern::inLargeRandomAreas
        generateTerrainAroundCliffs        // LandDistributionPattern::aroundCliffs
    };

    // 0x0046A021
    static void generateTerrain(HeightMap& heightMap)
    {
        const auto style = getEverywhereSurfaceStyle();
        if (!style.has_value())
        {
            return;
        }

        const TilePosRangeView tileLoop = getDrawableTileRange();
        for (const auto& tilePos : tileLoop)
        {
            auto* surface = World::TileManager::get(tilePos).surface();
            if (surface == nullptr)
            {
                continue;
            }
            surface->setTerrain(*style);
            surface->setGrowthStage(0);

            const auto variation = getRandomTerrainVariation(*surface);
            if (variation.has_value())
            {
                surface->setVariation(*variation);
            }
        }

        constexpr std::array landDistributionPatterns = {
            Scenario::LandDistributionPattern::farFromWater,
            Scenario::LandDistributionPattern::nearWater,
            Scenario::LandDistributionPattern::onMountains,
            Scenario::LandDistributionPattern::farFromMountains,
            Scenario::LandDistributionPattern::inSmallRandomAreas,
            Scenario::LandDistributionPattern::inLargeRandomAreas,
            Scenario::LandDistributionPattern::aroundCliffs,
        };

        for (auto i = 0U; i < landDistributionPatterns.size(); i++)
        {
            updateProgress(55 + 12 * i);

            for (uint8_t landObjectIdx = 0; landObjectIdx < ObjectManager::getMaxObjects(ObjectType::land); ++landObjectIdx)
            {
                const auto* landObj = ObjectManager::get<LandObject>(landObjectIdx);
                if (landObj == nullptr)
                {
                    continue;
                }

                const auto typePattern = Scenario::getOptions().landDistributionPatterns[landObjectIdx];
                const auto distPattern = landDistributionPatterns[i];
                if (typePattern != distPattern)
                {
                    continue;
                }

                _generateFuncs[enumValue(distPattern)](heightMap, landObjectIdx);
            }
        }
    }


  void riversFor(const Scenario::Options& o, HeightMap& h) { generateRivers(o, h); }
    static void generateSurfaceVariation()
    {
        for (auto& pos : World::getDrawableTileRange())
        {
            auto tile = TileManager::get(pos);
            auto* surface = tile.surface();

            if (surface == nullptr)
            {
                continue;
            }

            if (!surface->isIndustrial())
            {
                auto* landObj = ObjectManager::get<LandObject>(surface->terrain());
                if (landObj->hasFlags(LandObjectFlags::hasGrowthStages))
                {
                    bool setVariation = false;
                    if (surface->water())
                    {
                        auto waterBaseZ = surface->water() * kMicroToSmallZStep;
                        if (surface->slope())
                        {
                            waterBaseZ -= 4;
                        }

                        if (waterBaseZ > surface->baseZ())
                        {
                            if (surface->terrain() != 0)
                            {
                                surface->setGrowthStage(0);
                                setVariation = true;
                            }
                        }
                    }

                    if (!setVariation)
                    {
                        surface->setGrowthStage(landObj->numGrowthStages - 1);
                    }
                }
            }

            auto snowLine = Scenario::getCurrentSnowLine() / kMicroToSmallZStep;
            MicroZ baseMicroZ = (surface->baseZ() / kMicroToSmallZStep) + 1;
            auto unk = std::clamp(baseMicroZ - snowLine, 0, 5);
            surface->setSnowCoverage(unk);
        }
    }


  void restFor(HeightMap& h) { generateLand(h); generateWater(h); generateTerrain(h); generateSurfaceVariation(); }
}
namespace OpenLoco {
  static GameState gs; GameState& getGameState() { return gs; }
  static HillShapesObject hso; template<> HillShapesObject* ObjectManager::get<HillShapesObject>() { return &hso; }
  static std::vector<Gfx::G1Element> els; static std::vector<std::vector<uint8_t>> bytes;
  Gfx::G1Element* Gfx::getG1Element(uint32_t i) { return &els[i]; }
}
namespace OpenLoco {
  static std::vector<SurfaceElement> tiles(384*384);
  Tile World::TileManager::get(TilePos2 p) { return { &tiles[p.y*384+p.x] }; }
  Tile World::TileManager::get(Pos2 p) { return { &tiles[(p.y/32)*384+p.x/32] }; }
  int World::TileManager::getSurfaceCornerHeight(const SurfaceElement& s) { int b=s.baseZ(); if (s.slope()) { b+=4; if (s.isSlopeDoubleHeight()) b+=4; } return b; }
  static LandObject lands[32]; static bool landOk[32];
  template<> LandObject* ObjectManager::get<LandObject>(size_t i) { return landOk[i] ? &lands[i] : nullptr; }
  static Scenario::GenOptions gopts; Scenario::GenOptions& Scenario::getOptions() { return gopts; }
  static uint8_t snowLine; uint8_t Scenario::getCurrentSnowLine() { return snowLine; }
}
using namespace OpenLoco;
int main(int argc, char** argv) {
  FILE* lf = fopen(argv[2], "r"); unsigned last, sea; fscanf(lf, "%u %u", &last, &sea);
  uint8_t pats[2][32];
  for (int i = 0; i < 32; i++) { unsigned ok, nv, lk, pt, fl, ng; fscanf(lf, "%u %u %u %u %u %u", &ok, &nv, &lk, &pt, &fl, &ng); landOk[i] = ok; lands[i] = { (uint8_t)nv, (uint8_t)lk, (uint8_t)fl, (uint8_t)ng }; pats[0][i] = pt; pats[1][i] = pt; }
  // second pattern set: slots 0..3 far from water, everywhere, on mountains, large random areas
  const uint8_t alt[4] = {2, 0, 4, 7}; for (int i = 0; i < 4; i++) pats[1][i] = alt[i];
  FILE* f = fopen(argv[1], "r"); unsigned hc, mc, fl; fscanf(f, "%u %u %u", &hc, &mc, &fl);
  hso = { (uint8_t)hc, (uint8_t)mc, 0, HillShapeFlags(fl) };
  for (unsigned i = 0; i < hc + mc; i++) { int w, h, ef, n; fscanf(f, "%d %d %d %d", &w, &h, &ef, &n); bytes.emplace_back(n); for (int k = 0; k < n; k++) { unsigned v; fscanf(f, "%u", &v); bytes.back()[k] = v; } els.push_back({ nullptr, (int16_t)w, (int16_t)h }); }
  for (unsigned i = 0; i < hc + mc; i++) els[i].offset = bytes[i].data();
  for (int c = 0; c < 4; c++) {
    int ps = c % 2;
    for (int i = 0; i < 32; i++) gopts.landDistributionPatterns[i] = Scenario::LandDistributionPattern(pats[ps][i]);
    for (auto& t : tiles) { t = SurfaceElement{}; t.t = last; t.bz = 4; }
    gs.rng = Core::Prng(777u + c * 31u, 999u + c * 17u); gs.seaLevel = sea; gs.lastLandOption = last; snowLine = c % 2 ? 44 : 20;
    Scenario::Options o{ 2, Scenario::TopographyStyle(c < 2 ? 2 : 4), uint8_t(c < 2 ? 24 : 60), Scenario::ScenarioFlags::hillsEdgeOfMap, uint8_t(c == 3 ? 1 : 0), 3, 6, 1, 2 };
    World::MapGenerator::HeightMap hm(384, 384); World::MapGenerator::OriginalTerrainGenerator g; g.generate(o, hm);
    World::MapGenerator::riversFor(o, hm); World::MapGenerator::restFor(hm);
    uint64_t hsh = 1469598103934665603ull;
    for (auto& t : tiles) { for (int v : { (int)t.t, (int)t.var, (int)t.w, t.bz, (int)t.sl, (int)t.g, (int)t.snow }) { hsh ^= (uint8_t)v; hsh *= 1099511628211ull; } }
    printf("%d %d %u %u %u %u %u %llx\n", c, ps, c % 2 ? 44 : 20, c < 2 ? 2 : 4, c < 2 ? 24 : 60, 777u + c * 31u, 999u + c * 17u, (unsigned long long)hsh);
  }
}
