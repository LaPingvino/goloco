package ui

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Ui/Widget.h"
// #include "Ui/Widgets/ButtonWidget.h"
// #include "Ui/Widgets/TextBoxWidget.h"
// namespace OpenLoco::Ui::Widgets
// func MakeStepperDecreaseWidget(origin Ui::Point, size Ui::Size, colour WindowColour, Widget::kContentNull [[maybe_unused]] uint32 content =, StringIds::null [[maybe_unused]] StringId tooltip =) Button
// const int16 xPos = origin.x + size.width - 26;
// const int16 yPos = origin.y + 1;
// const uint16 width = 13;
// const uint16 height = size.height - 2;
func Button(xPos {, } yPos, width {, } height, colour, StringIds::stepper_minus, tooltip) return {
}
// func MakeStepperIncreaseWidget(origin Ui::Point, size Ui::Size, colour WindowColour, Widget::kContentNull [[maybe_unused]] uint32 content =, StringIds::null [[maybe_unused]] StringId tooltip =) Button
// const int16 xPos = origin.x + size.width - 13;
// const int16 yPos = origin.y + 1;
// const uint16 width = 12;
// const uint16 height = size.height - 2;
func Button(xPos {, } yPos, width {, } height, colour, StringIds::stepper_plus, tooltip) return {
// // TODO: Make this a single widget.
// func StepperWidgets(origin Ui::Point, size Ui::Size, colour WindowColour, Widget::kContentNull uint32 content =, StringIds::null StringId tooltip =) any
// return makeWidgets(
// TextBox(origin, size, colour, content, tooltip),
// makeStepperDecreaseWidget(origin, size, colour),
// makeStepperIncreaseWidget(origin, size, colour));
