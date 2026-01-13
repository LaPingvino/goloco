package ui

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Ui/Widget.h"
// #include "Ui/Widgets/ButtonWidget.h"
// namespace OpenLoco::Ui::Widgets
type ComboBox struct {
	Public // embedded
	// method: constexpr ComboBox(WidgetId id, Point origin, Size size, WindowColour colour, StringId content = StringIds::null, StringId tooltip = StringIds::null)
// : Widget(id, origin, size, kWidgetType, colour, content, tooltip)
// events.draw = &draw;
}
const ComboBoxWidgetType any = WidgetType.combobox
// func ComboBox(origin Point, size Size, colour WindowColour, StringIds::null StringId content =, StringIds::null StringId tooltip =) constexpr
// : ComboBox(WidgetId::none, origin, size, colour, content, tooltip)
// events.draw = &draw;
// func Draw(drawingCtx Gfx::DrawingContext, widget Widget, widgetState WidgetState) 
// func DropdownWidgets(origin Ui::Point, size Ui::Size, colour WindowColour, Widget::kContentNull uint32_t content =, StringIds::null StringId tooltip =) any
// const auto makeDropdownButtonWidget = [](Ui::Point origin, Ui::Size size, WindowColour colour) {
// const int16_t xPos = origin.x + size.width - 12;
// const int16_t yPos = origin.y + 1;
// const uint16_t width = 11;
// const uint16_t height = 10;
func Button(xPos {, } yPos, width {, } height, colour, StringIds::dropdown) return {
}
// // TODO: Make this a single widget.
// return makeWidgets(
// ComboBox(origin, size, colour, content, tooltip),
// makeDropdownButtonWidget(origin, size, colour));
