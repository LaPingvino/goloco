package graphics

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include <OpenLoco/Engine/Ui/Rect.hpp>
// #include <cstdint>
// #include <optional>
// namespace OpenLoco::Gfx
// // TODO: Convert this to a handle once everything is implemented.
// // Depending on the rendering engine this could be a buffer on GPU or RAM.
type RenderTarget struct {
// uint8_t* bits;      // 0x00
	X int16
	Y int16
	Width int16
	Height int16
	Pitch int16
	ZoomLevel uint16
// Ui::Rect getUiRect() const;
// Ui::Rect getDrawableRect() const;
}
// std::optional<RenderTarget> clipRenderTarget(const RenderTarget& src, const Ui::Rect& newRect);
