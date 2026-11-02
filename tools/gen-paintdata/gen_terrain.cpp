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
#include <OpenLoco/Core/Prng.h>
namespace OpenLoco {
  using tile_coord_t = int16_t;
  template<class E> constexpr auto enumValue(E e) { return static_cast<std::underlying_type_t<E>>(e); }
  struct TilePos2 { int32_t x, y; TilePos2(int32_t a, int32_t b) : x(a), y(b) {} };
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
  namespace ObjectManager { template<class T> T* get(); template<> HillShapesObject* get<HillShapesObject>(); }
  struct GameState { Core::Prng rng; uint16_t seaLevel; };
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
namespace OpenLoco::World::MapGenerator {
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


  void riversFor(const Scenario::Options& o, HeightMap& h) { generateRivers(o, h); }
}
namespace OpenLoco {
  static GameState gs; GameState& getGameState() { return gs; }
  static HillShapesObject hso; template<> HillShapesObject* ObjectManager::get<HillShapesObject>() { return &hso; }
  static std::vector<Gfx::G1Element> els; static std::vector<std::vector<uint8_t>> bytes;
  Gfx::G1Element* Gfx::getG1Element(uint32_t i) { return &els[i]; }
}
using namespace OpenLoco;
int main(int argc, char** argv) {
  FILE* f = fopen(argv[1], "r"); unsigned hc, mc, fl; fscanf(f, "%u %u %u", &hc, &mc, &fl);
  hso = { (uint8_t)hc, (uint8_t)mc, 0, HillShapeFlags(fl) };
  for (unsigned i = 0; i < hc + mc; i++) { int w, h, ef, n; fscanf(f, "%d %d %d %d", &w, &h, &ef, &n); bytes.emplace_back(n); for (int k = 0; k < n; k++) { unsigned v; fscanf(f, "%u", &v); bytes.back()[k] = v; } }
  for (unsigned i = 0; i < hc + mc; i++) { els.push_back({ bytes[i].data(), 0, 0 }); }
  rewind(f); fscanf(f, "%u %u %u", &hc, &mc, &fl); for (unsigned i = 0; i < hc + mc; i++) { int w, h, ef, n; fscanf(f, "%d %d %d %d", &w, &h, &ef, &n); els[i].width = w; els[i].height = h; for (int k = 0; k < n; k++) { unsigned v; fscanf(f, "%u", &v); } }
  for (int c = 0; c < 6; c++) { // cases: style, density, edge, seed
    const uint8_t styles[6] = {2, 4, 1, 3, 2, 0}; const uint8_t dens[6] = {24, 50, 10, 80, 30, 24};
    gs.rng = Core::Prng(12345u + c * 7919u, 67890u + c * 104729u); gs.seaLevel = 4;
    Scenario::Options o{ 2, Scenario::TopographyStyle(styles[c]), dens[c], c % 2 ? Scenario::ScenarioFlags::hillsEdgeOfMap : Scenario::ScenarioFlags::none, uint8_t(c % 3), uint8_t(2 + c), uint8_t(5 + c), uint8_t(c % 4), uint8_t(c * 3 % 7) };
    World::MapGenerator::HeightMap hm(384, 384); World::MapGenerator::OriginalTerrainGenerator g; g.generate(o, hm); World::MapGenerator::riversFor(o, hm);
    uint64_t hsh = 1469598103934665603ull; for (auto v : hm.h) { hsh ^= v; hsh *= 1099511628211ull; }
    printf("%d %u %u %d %u %u %u %u %u %u %u %llx\n", c, styles[c], dens[c], c % 2, c % 3, 2 + c, 5 + c, c % 4, c * 3 % 7, 12345u + c * 7919u, 67890u + c * 104729u, (unsigned long long)hsh);
  }
}
