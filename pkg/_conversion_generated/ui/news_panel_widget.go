package ui

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Ui/Widget.h"
// namespace OpenLoco::Ui::Widgets
type NewsPanel struct {
	Public // embedded
type Style int

const (
	Old Style = iota
	New_
)
	// method: constexpr NewsPanel(WidgetId id, Point origin, Size size, WindowColour colour, uint32_t content = Widget::kContentNull, StringId tooltip = StringIds::null, Style newsStyle = Style::old)
// : Widget(id, origin, size, kWidgetType, colour, content, tooltip)
// events.draw = &draw;
// styleData = enumValue(newsStyle);
}
const NewsPanelWidgetType any = WidgetType.newsPanel
// func NewsPanel(origin Point, size Size, colour WindowColour, Widget::kContentNull uint32_t content =, StringIds::null StringId tooltip =, Style::old Style newsStyle =) constexpr
// : NewsPanel(WidgetId::none, origin, size, colour, content, tooltip, newsStyle)
// func Draw(drawingCtx Gfx::DrawingContext, widget Widget, widgetState WidgetState) 
