package ui

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Audio/Audio.h"
// #include "Graphics/Colour.h"
// #include "Graphics/TextRenderer.h"
// #include "Input.h"
// #include "Localisation/FormatArguments.hpp"
// #include "Localisation/Formatting.h"
// #include "Localisation/StringIds.h"
// #include "OpenLoco.h"
// #include "Ui.h"
// #include "Ui/Widget.h"
// #include "Ui/Widgets/ButtonWidget.h"
// #include "Ui/Widgets/CaptionWidget.h"
// #include "Ui/Widgets/PanelWidget.h"
// #include "Ui/WindowManager.h"
// #include <SDL2/SDL.h>
// #include <cstring>
// namespace OpenLoco::Ui::Windows::PromptOkCancel
// static char _descriptionBuffer[512];
// static bool _result; // 0x009D1C9A
type Widx int

const (
	Frame Widx = iota
	Caption
	CloseButton
	OkButton
	CancelButton
)
// static constexpr auto _widgets = makeWidgets(
// Widgets::Panel({ 0, 0 }, { 280, 92 }, WindowColour::primary),
// Widgets::Caption({ 1, 1 }, { 278, 13 }, Widgets::Caption::Style::boxed, WindowColour::primary),
// Widgets::Button({ 267, 2 }, { 11, 11 }, WindowColour::primary, StringIds::close_window_cross, StringIds::tooltip_close_window),
// Widgets::Button({ 20, 77 }, { 100, 12 }, WindowColour::primary, StringIds::label_ok),
// Widgets::Button({ 160, 77 }, { 100, 12 }, WindowColour::primary, StringIds::label_button_cancel)
// );
// static const WindowEventList& getEvents();
// // 0x00446F6B
// // eax: okButtonStringId
// // eax: {return}
// func Open(captionId StringId, descriptionId StringId, descriptionArgs FormatArguments, okButtonStringId StringId) bool
// auto window = WindowManager::createWindowCentred(
// WindowType::confirmationPrompt,
// { 280, 92 },
// Ui::WindowFlags::playSoundOnOpen | Ui::WindowFlags::stickToFront,
// getEvents());
// if (window == nullptr)
// orphan member: return false;
// window->setWidgets(_widgets);
// window->widgets[widx::caption].text = captionId;
// window->widgets[widx::okButton].text = okButtonStringId;
// // Prepare description buffer for drawing
// StringManager::formatString(_descriptionBuffer, descriptionId, descriptionArgs);
// window->initScrollWidgets();
// window->setColour(WindowColour::primary, AdvancedColour(Colour::mutedDarkRed).translucent());
// window->setColour(WindowColour::secondary, AdvancedColour(Colour::mutedDarkRed).translucent());
// window->flags |= Ui::WindowFlags::transparent;
// _result = false;
// auto originalModal = WindowManager::getCurrentModalType();
// WindowManager::setCurrentModalType(WindowType::confirmationPrompt);
// promptTickLoop(
// []() {
// Input::handleKeyboard();
// Audio::updateSounds();
// WindowManager::dispatchUpdateAll();
// Input::processKeyboardInput();
// WindowManager::update();
// Ui::minimalHandleInput();
// Gfx::renderAndUpdate();
// return WindowManager::find(WindowType::confirmationPrompt) != nullptr;
// });
// WindowManager::setCurrentModalType(originalModal);
// orphan member: return _result;
// // 0x00447125
// func KeyUp(w Window, charCode [[maybe_unused]] uint32_t, keyCode uint32) bool
// switch (keyCode)
// case SDLK_ESCAPE:
// w.callOnMouseUp(widx::closeButton, w.widgets[widx::closeButton].id);
// orphan member: return true;
// case SDLK_RETURN:
// case SDLK_KP_ENTER:
// w.callOnMouseUp(widx::okButton, w.widgets[widx::okButton].id);
// orphan member: return true;
// default:
// orphan member: return false;
// // 0x00447093
// func PrepareDraw(self [[maybe_unused]] Window) 
// // Prepare description string for drawing.
// char* buffer_2039 = const_cast<char*>(StringManager::getString(StringIds::buffer_2039));
// strncpy(&buffer_2039[0], _descriptionBuffer, 512);
// // 0x004470FD
// func OnMouseUp(self Window, widgetIndex WidgetIndex_t, id [[maybe_unused]] WidgetId) 
// switch (widgetIndex)
// case widx::closeButton:
// case widx::cancelButton:
// WindowManager::close(self.type);
// break;
// case widx::okButton:
// _result = true;
// WindowManager::close(self.type);
// break;
// // 0x004470AA
// func Draw(self Window, drawingCtx Gfx::DrawingContext) 
// auto tr = Gfx::TextRenderer(drawingCtx);
// self.draw(drawingCtx);
// orphan member: FormatArguments args{};
// args.push(StringIds::buffer_2039);
// auto origin = Ui::Point(self.x + self.width / 2, self.y + 41);
// tr.drawStringCentredWrapped(origin, self.width, Colour::black, StringIds::wcolour2_stringid, args);
// static constexpr WindowEventList kEvents = {
// .onMouseUp = onMouseUp,
// .prepareDraw = prepareDraw,
// .draw = draw,
// .keyUp = keyUp,
// static const WindowEventList& getEvents()
// orphan member: return kEvents;
