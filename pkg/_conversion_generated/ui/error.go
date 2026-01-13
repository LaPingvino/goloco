package ui

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Audio/Audio.h"
// #include "Graphics/Colour.h"
// #include "Graphics/ImageIds.h"
// #include "Graphics/TextRenderer.h"
// #include "Input.h"
// #include "Localisation/FormatArguments.hpp"
// #include "Localisation/Formatting.h"
// #include "Localisation/StringManager.h"
// #include "Objects/CompetitorObject.h"
// #include "Objects/InterfaceSkinObject.h"
// #include "Objects/ObjectManager.h"
// #include "OpenLoco.h"
// #include "Ui/Widget.h"
// #include "Ui/Widgets/PanelWidget.h"
// #include "Ui/Widgets/Wt3Widget.h"
// #include "Ui/WindowManager.h"
// #include "World/CompanyManager.h"
// namespace OpenLoco::Ui::Windows::Error
// static char _errorText[512];         // 0x009C64B3
// static uint16_t _linebreakCount;     // 0x009C66B3
// static CompanyId _errorCompetitorId; // 0x009C68EC
var MinWidth = 70 // auto
var MaxWidth = 250 // auto
var Padding = 4 // auto
var CompetitorSize = 64 // auto
// namespace Common
// static const WindowEventList& getEvents();
// namespace Error
type Widx int

const (
	Frame Widx = iota
)
// static constexpr auto widgets = makeWidgets(
// Widgets::Panel({ 0, 0 }, { 200, 42 }, WindowColour::primary)
// );
// namespace ErrorCompetitor
type Widx int

const (
	Frame Widx = iota
	InnerFrame
)
// static constexpr auto widgets = makeWidgets(
// Widgets::Panel({ 0, 0 }, { 250, 70 }, WindowColour::primary),
// Widgets::Wt3Widget({ 3, 3 }, { 64, 64 }, WindowColour::secondary)
// );
// static char* formatErrorString(StringId title, StringId message, FormatArguments args, char* buffer)
// char* ptr = buffer;
// ptr[0] = ControlCodes::Colour::white;
// ptr++;
// if (title != StringIds::null)
// ptr = StringManager::formatString(ptr, title, args);
// if (message != StringIds::null)
// if (title != StringIds::null)
// *ptr = ControlCodes::newline;
// ptr++;
// StringManager::formatString(ptr, message, args);
// orphan member: return ptr;
// func CreateErrorWindow(title StringId, message StringId, suppressErrorSound bool) 
// WindowManager::close(WindowType::error);
// char* buffer = _errorText;
// auto args = FormatArguments::common();
// buffer = formatErrorString(title, message, args, buffer);
// if (buffer != &_errorText[0])
// // How wide is the error string?
// uint16_t strWidth = Gfx::TextRenderer::getStringWidthNewLined(Gfx::Font::medium_bold, &_errorText[0]);
// strWidth = std::clamp<uint16_t>(strWidth, kMinWidth, kMaxWidth);
// // How many linebreaks?
// uint16_t breakLineCount = 0;
// std::tie(strWidth, breakLineCount) = Gfx::TextRenderer::wrapString(Gfx::Font::medium_bold, &_errorText[0], strWidth + kPadding);
// _linebreakCount = breakLineCount;
// // Calculate window dimensions
// uint16_t width = strWidth + 2 * kPadding;
// uint16_t height = (_linebreakCount + 1) * 10 + 2 * kPadding;
// // Add extra spacing for competitor image
// if (_errorCompetitorId != CompanyId::null)
// width += kCompetitorSize + 22;
// height += kCompetitorSize - 22;
// // Calculate frame size
// uint16_t frameWidth = width - 1;
// uint16_t frameHeight = height - 1;
// // Position error message around the cursor
// auto mousePos = Input::getMouseLocation();
// Ui::Point windowPosition = Ui::Point{ mousePos.x, mousePos.y } + Ui::Point(-width / 2, 26);
// windowPosition.x = std::clamp<int32_t>(windowPosition.x, 0, Ui::width() - width - 40);
// windowPosition.y = std::clamp<int32_t>(windowPosition.y, 22, Ui::height() - height - 40);
// auto error = WindowManager::createWindow(
// WindowType::error,
// windowPosition,
// { width, height },
// WindowFlags::stickToFront | WindowFlags::transparent | WindowFlags::ignoreInFindAt,
// Common::getEvents());
// if (_errorCompetitorId != CompanyId::null)
// error->setWidgets(ErrorCompetitor::widgets);
// else
// error->setWidgets(Error::widgets);
// error->setColour(WindowColour::primary, AdvancedColour(Colour::mutedDarkRed).translucent());
// error->setColour(WindowColour::secondary, AdvancedColour(Colour::mutedDarkRed).translucent());
// error->widgets[Error::widx::frame].right = frameWidth;
// error->widgets[Error::widx::frame].bottom = frameHeight;
// error->var_846 = 0;
// if (!suppressErrorSound)
// int32_t pan = (error->width / 2) + error->x;
// Audio::playSound(Audio::SoundId::error, pan);
// // 0x00431A8A
// func Open(title StringId, message StringId) 
// _errorCompetitorId = CompanyId::null;
// createErrorWindow(title, message, false);
// func OpenQuiet(title StringId, message StringId) 
// _errorCompetitorId = CompanyId::null;
// createErrorWindow(title, message, true);
// // 0x00431908
// func OpenWithCompetitor(title StringId, message StringId, competitorId CompanyId) 
// _errorCompetitorId = competitorId;
// createErrorWindow(title, message, false);
// namespace Common
// // 0x00431C05
// func Draw(self Ui::Window, drawingCtx Gfx::DrawingContext) 
// self.draw(drawingCtx);
// auto tr = Gfx::TextRenderer(drawingCtx);
// auto colour = AdvancedColour(Colour::white).translucent(); // self.colours[0];
// if (_errorCompetitorId == CompanyId::null)
// uint16_t xPos = self.x + self.width / 2;
// uint16_t yPos = self.y + kPadding;
// tr.drawStringCentredRaw(Point(xPos, yPos), _linebreakCount, colour, &_errorText[0]);
// else
// auto xPos = self.x + self.widgets[ErrorCompetitor::widx::innerFrame].left;
// auto yPos = self.y + self.widgets[ErrorCompetitor::widx::innerFrame].top;
// auto company = CompanyManager::get(_errorCompetitorId);
// auto companyObj = ObjectManager::get<CompetitorObject>(company->competitorId);
// auto imageId = companyObj->images[enumValue(company->ownerEmotion)];
// imageId = Gfx::recolour(imageId, company->mainColours.primary);
// imageId++;
// drawingCtx.drawImage(xPos, yPos, imageId);
// if (company->jailStatus != 0)
// drawingCtx.drawImage(xPos, yPos, ImageIds::owner_jailed);
// auto point = Point(self.x + (self.width - kCompetitorSize) / 2 + kCompetitorSize + kPadding, self.y + 20);
// tr.drawStringCentredRaw(point, _linebreakCount, colour, &_errorText[0]);
// // 0x00431E1B
// func OnPeriodicUpdate(self Ui::Window) 
// self.var_846++;
// if (self.var_846 >= 7)
// WindowManager::close(&self);
// static constexpr WindowEventList kEvents = {
// .onPeriodicUpdate = onPeriodicUpdate,
// .draw = draw,
// static const WindowEventList& getEvents()
// orphan member: return kEvents;
