// gen_roadarrows.cpp — prints PaintRoad.cpp's kOneWayArrowImageIndexs as Go.
#include "../../../OpenLoco/src/OpenLoco/src/Paint/PaintRoad.cpp"
#include <cstdio>

int main()
{
    using namespace OpenLoco::Paint;
    std::printf("\n// RoadOneWayArrowImages is PaintRoad.cpp's kOneWayArrowImageIndexs [roadId][isLast][rotation].\nvar RoadOneWayArrowImages = [%zu][2][4]uint32{\n", kOneWayArrowImageIndexs.size());
    for (auto& r : kOneWayArrowImageIndexs)
    {
        std::printf("\t{");
        for (auto& l : r)
            std::printf("{%u, %u, %u, %u}, ", (unsigned)l[0], (unsigned)l[1], (unsigned)l[2], (unsigned)l[3]);
        std::printf("},\n");
    }
    std::printf("}\n");
}
