// Generates the getHeight vectors whose SHA-256 TestGetHeightMatchesUpstream checks: TileManager.cpp getHeight + slope helpers (sed-copied) on a stub surface; run ./gh | sha256sum.
#include <cstdio>
#include <cstdint>
#include <initializer_list>
constexpr uint8_t kTileSize = 31; // TileManager.cpp's own (above its slope helpers): the last sub-tile coordinate
namespace World { constexpr int kMapWidth = 384*32, kMapHeight = 384*32; }
struct Pos2 { int32_t x, y; };
struct TileHeight { int16_t landHeight, waterHeight; };
namespace SurfaceSlope { constexpr uint8_t flat = 0, doubleHeight = 1<<4;
 namespace CornerUp { constexpr uint8_t all=0xF, north=1, east=2, south=4, west=8; }
 namespace CornerDown { constexpr uint8_t west = CornerUp::all & ~CornerUp::west, south = CornerUp::all & ~CornerUp::south, east = CornerUp::all & ~CornerUp::east, north = CornerUp::all & ~CornerUp::north; }
 namespace SideUp { constexpr uint8_t northeast = CornerUp::north | CornerUp::east, southeast = CornerUp::south | CornerUp::east, northwest = CornerUp::north | CornerUp::west, southwest = CornerUp::south | CornerUp::west; }
 namespace Valley { constexpr uint8_t westeast = CornerUp::east | CornerUp::west, northsouth = CornerUp::north | CornerUp::south; } }
struct Surf { int bz, sl, w; int16_t waterHeight() const { return w*16; } int16_t baseHeight() const { return bz*4; } uint8_t slopeCorners() const { return sl & 0xF; } bool isSlopeDoubleHeight() const { return sl & 0x10; } };
static Surf cur;
struct Tile { Surf* surface() const { return &cur; } };
namespace TileManager { Tile get(Pos2) { return {}; } }
    static int16_t getOneCornerUpLandHeight(int8_t xl, int8_t yl, uint8_t slope)
    {
        int16_t quad = 0;
        switch (slope)
        {
            case SurfaceSlope::CornerUp::north:
                quad = xl + yl - kTileSize;
                break;
            case SurfaceSlope::CornerUp::east:
                quad = xl - yl;
                break;
            case SurfaceSlope::CornerUp::south:
                quad = kTileSize - yl - xl;
                break;
            case SurfaceSlope::CornerUp::west:
                quad = yl - xl;
                break;
        }
        // If the element is in the quadrant with the slope, raise its height
        if (quad > 0)
        {
            return quad / 2;
        }
        return 0;
    }

    static int16_t getOneSideUpLandHeight(int8_t xl, int8_t yl, uint8_t slope)
    {
        int16_t edge = 0;
        switch (slope)
        {
            case SurfaceSlope::SideUp::northeast:
                edge = xl / 2 + 1;
                break;
            case SurfaceSlope::SideUp::southeast:
                edge = (kTileSize - yl) / 2;
                break;
            case SurfaceSlope::SideUp::northwest:
                edge = yl / 2 + 1;
                break;
            case SurfaceSlope::SideUp::southwest:
                edge = (kTileSize - xl) / 2;
                break;
        }
        return edge;
    }

    // This also takes care of the one corner down and one opposite corner up
    static int16_t getOneCornerDownLandHeight(int8_t xl, int8_t yl, uint8_t slope, bool isDoubleHeight)
    {
        int16_t quadExtra = 0;
        int16_t quad = 0;

        switch (slope)
        {
            case SurfaceSlope::CornerDown::west:
                quadExtra = xl + kTileSize - yl;
                quad = xl - yl;
                break;
            case SurfaceSlope::CornerDown::south:
                quadExtra = xl + yl;
                quad = xl + yl - kTileSize - 1;
                break;
            case SurfaceSlope::CornerDown::east:
                quadExtra = kTileSize - xl + yl;
                quad = yl - xl;
                break;
            case SurfaceSlope::CornerDown::north:
                quadExtra = (kTileSize - xl) + (kTileSize - yl);
                quad = kTileSize - yl - xl - 1;
                break;
        }

        if (isDoubleHeight)
        {
            return quadExtra / 2 + 1;
        }
        else
        {
            // This tile is essentially at the next height level
            // so we move *down* the slope
            return quad / 2 + 16;
        }
    }

    static int16_t getValleyLandHeight(int8_t xl, int8_t yl, uint8_t slope)
    {
        int16_t quad = 0;
        switch (slope)
        {
            case SurfaceSlope::Valley::westeast:
                if (xl + yl > kTileSize + 1)
                {
                    quad = kTileSize - xl - yl;
                }
                break;
            case SurfaceSlope::Valley::northsouth:
                quad = xl - yl;
                break;
        }

        if (quad > 0)
        {
            return quad / 2;
        }
        return 0;
    }

    /**
     * Return the absolute height of an element, given its (x, y) coordinates
     * remember to & with 0xFFFF if you don't want water affecting results
     *
     * @param x @<ax>
     * @param y @<cx>
     * @return height @<edx>
     *
     * 0x00467297 rct2: 0x00662783 (numbers different)
     */
    TileHeight getHeight(const Pos2& pos)
    {
        TileHeight height{ 16, 0 };
        // Off the map
        if ((unsigned)pos.x >= (World::kMapWidth - 1) || (unsigned)pos.y >= (World::kMapHeight - 1))
        {
            return height;
        }

        auto tile = TileManager::get(pos);
        // Get the surface element for the tile
        auto surfaceEl = tile.surface();

        if (surfaceEl == nullptr)
        {
            return height;
        }

        height.waterHeight = surfaceEl->waterHeight();
        height.landHeight = surfaceEl->baseHeight();

        const auto slope = surfaceEl->slopeCorners();

        // Sub-tile coords
        const auto xl = pos.x & 0x1f;
        const auto yl = pos.y & 0x1f;

        // Slope logic:
        // Each of the four bits in slope represents that corner being raised
        // slope == 15 (all four bits) is not used and slope == 0 is flat
        // If the extra_height bit is set, then the slope goes up two z-levels (this happens with one corner down with opposite corner up)

        // We arbitrarily take the SW corner to be closest to the viewer

        switch (slope)
        {
            case SurfaceSlope::flat:
                // Flat surface requires no further calculations.
                break;

            case SurfaceSlope::CornerUp::north:
            case SurfaceSlope::CornerUp::east:
            case SurfaceSlope::CornerUp::south:
            case SurfaceSlope::CornerUp::west:
                height.landHeight += getOneCornerUpLandHeight(xl, yl, slope);
                break;

            case SurfaceSlope::SideUp::northeast:
            case SurfaceSlope::SideUp::southeast:
            case SurfaceSlope::SideUp::northwest:
            case SurfaceSlope::SideUp::southwest:
                height.landHeight += getOneSideUpLandHeight(xl, yl, slope);
                break;

            case SurfaceSlope::CornerDown::north:
            case SurfaceSlope::CornerDown::east:
            case SurfaceSlope::CornerDown::south:
            case SurfaceSlope::CornerDown::west:
                height.landHeight += getOneCornerDownLandHeight(xl, yl, slope, surfaceEl->isSlopeDoubleHeight());
                break;

            case SurfaceSlope::Valley::northsouth:
            case SurfaceSlope::Valley::westeast:
                height.landHeight += getValleyLandHeight(xl, yl, slope);
                break;
        }
        return height;
    }
int main() {
  for (int sl = 0; sl < 32; sl++) for (int bz : {0, 4, 40}) for (int w : {0, 3}) { cur = {bz, sl, w};
    for (int y = 0; y < 32; y++) for (int x = 0; x < 32; x++) { auto h = getHeight(Pos2{ 64 + x, 96 + y }); printf("%d %d %d %d %d %d %d\n", sl, bz, w, x, y, h.landHeight, h.waterHeight); } }
  auto o = getHeight(Pos2{ -1, 5 }); printf("off %d %d\n", o.landHeight, o.waterHeight);
}
