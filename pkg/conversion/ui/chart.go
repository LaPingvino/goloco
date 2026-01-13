package ui

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Graphics/Colour.h"
// #include "Types.hpp"
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <cstddef>
// namespace OpenLoco::Gfx
// forward: class DrawingContext;
// namespace OpenLoco::Ui
// forward: struct Window;
type GraphFlags int

const (
	None GraphFlags = 0
	ShowNegativeValues GraphFlags = 1 << 0
	DataFrontToBack GraphFlags = 1 << 1
	HideAxesAndLabels GraphFlags = 1 << 2
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(GraphFlags);
type GraphPointFlags int

const (
	None GraphPointFlags = 0
	DrawLines GraphPointFlags = 1 << 0
	DrawPoints GraphPointFlags = 1 << 1
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(GraphPointFlags);
var MaxLines = 32 // auto
type GraphSettings struct {
	Left uint16
	Top uint16
	Width uint16
	Height uint16
	YOffset uint16
	XOffset uint16
	YAxisLabelIncrement uint32
	LineCount uint16
// std::byte* yData[kMaxLines];          // 0x0113DC8C
	DataTypeSize uint32
// uint16_t dataStart[kMaxLines];        // 0x0113DD10
	LinesToExclude uint32
// PaletteIndex_t lineColour[kMaxLines]; // 0x0113DD54
	DataEnd uint16
	XLabel StringId
	XAxisRange uint32
	XAxisStepSize uint32
	XAxisTickIncrement uint16
	XAxisLabelIncrement uint16
	YLabel StringId
	Dword_113DD86 uint32
	YAxisStepSize uint32
	Flags GraphFlags
	CanvasLeft uint16
	CanvasBottom uint16
	CanvasHeight uint16
	NumValueShifts uint8
	PointFlags GraphPointFlags
// uint16_t itemId[kMaxLines];           // 0x0113DD9A
}
// func DrawGraph(gs GraphSettings, self Window, drawingCtx Gfx::DrawingContext) 
