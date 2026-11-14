// gen_signalarrows.cpp — prints PaintSignal.cpp's kOneWayArrowLeft / Right as Go.
#include "../../../OpenLoco/src/OpenLoco/src/Paint/PaintSignal.cpp"
#include <cstdio>

int main()
{
    using namespace OpenLoco::Paint;
    std::printf("\n// SignalOneWayArrowLeft / Right are PaintSignal.cpp's kOneWayArrowLeft / kOneWayArrowRight, by track rotation (0-15).\nvar SignalOneWayArrowLeft = [16]uint32{");
    for (auto v : kOneWayArrowLeft)
        std::printf("%u, ", (unsigned)v);
    std::printf("}\nvar SignalOneWayArrowRight = [16]uint32{");
    for (auto v : kOneWayArrowRight)
        std::printf("%u, ", (unsigned)v);
    std::printf("}\n");
}
