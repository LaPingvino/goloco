package ui

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Ui/Widget.h"
// namespace OpenLoco::Ui::Widgets
type Caption struct {
	Public // embedded
type Style int

const (
	Boxed Style = iota
	BlackText
	ColourText
	WhiteText
)
	// method: constexpr Caption(WidgetId id, Point origin, Size size, Style captionStyle, WindowColour colour, StringId content = StringIds::empty, StringId tooltip = StringIds::null)
// : Widget(id, origin, size, kWidgetType, colour, content, tooltip)
// events.draw = &draw;
// styleData = enumValue(captionStyle);
}
const CaptionWidgetType any = WidgetType.caption
// func Caption(origin Point, size Size, captionStyle Style, colour WindowColour, StringIds::empty StringId content =, StringIds::null StringId tooltip =) constexpr
// : Caption(WidgetId::none, origin, size, captionStyle, colour, content, tooltip)
// func Draw(drawingCtx Gfx::DrawingContext, widget Widget, widgetState WidgetState) 
