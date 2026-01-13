package ui

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Date.h"
// #include "Economy/Economy.h"
// #include "Graphics/Colour.h"
// #include "Graphics/ImageIds.h"
// #include "Graphics/RenderTarget.h"
// #include "Graphics/TextRenderer.h"
// #include "Input.h"
// #include "Localisation/FormatArguments.hpp"
// #include "Localisation/Formatting.h"
// #include "Localisation/StringIds.h"
// #include "Objects/CargoObject.h"
// #include "Objects/CompetitorObject.h"
// #include "Objects/InterfaceSkinObject.h"
// #include "Objects/ObjectManager.h"
// #include "OpenLoco.h"
// #include "Ui/Chart.h"
// #include "Ui/ToolManager.h"
// #include "Ui/Widget.h"
// #include "Ui/Widgets/CaptionWidget.h"
// #include "Ui/Widgets/FrameWidget.h"
// #include "Ui/Widgets/ImageButtonWidget.h"
// #include "Ui/Widgets/LabelWidget.h"
// #include "Ui/Widgets/PanelWidget.h"
// #include "Ui/Widgets/ScrollViewWidget.h"
// #include "Ui/Widgets/TabWidget.h"
// #include "Ui/Widgets/TableHeaderWidget.h"
// #include "Ui/WindowManager.h"
// #include "World/Company.h"
// #include "World/CompanyManager.h"
// #include "World/CompanyRecords.h"
// #include <OpenLoco/Core/Numerics.hpp>
// using namespace OpenLoco::Literals;
// namespace OpenLoco::Ui::Windows::CompanyList
// static uint16_t _hoverItemTicks;     // 0x009C68C7
// static GraphSettings _graphSettings; // 0x0113DC7A
var LegendMargin = 6 // auto
var LegendWidth = 100 // auto
var WindowPadding = 4 // auto
// namespace Common
// static constexpr Ui::Size kMaxWindowSize = { 800, 940 }; // NB: frame background is only 800px :(
// static constexpr Ui::Size kMinWindowSize = { 300, 272 };
type Widx int

const (
	Frame Widx = iota
	Caption
	Close_button
	Panel
	Tab_company_list
	Tab_performance
	Tab_cargo_units
	Tab_cargo_distance
	Tab_values
	Tab_payment_rates
	Tab_speed_records
)
// func MakeCommonWidgets(frameWidth int32, frameHeight int32, windowCaptionId StringId) any
// return makeWidgets(
// Widgets::Frame({ 0, 0 }, { frameWidth, frameHeight }, WindowColour::primary),
// Widgets::Caption({ 1, 1 }, { frameWidth - 2, 13 }, Widgets::Caption::Style::whiteText, WindowColour::primary, windowCaptionId),
// Widgets::ImageButton({ frameWidth - 15, 2 }, { 13, 13 }, WindowColour::primary, ImageIds::close_button, StringIds::tooltip_close_window),
// Widgets::Panel({ 0, 41 }, { frameWidth, 231 }, WindowColour::secondary),
// Widgets::Tab({ 3, 15 }, { 31, 27 }, WindowColour::secondary, ImageIds::tab, StringIds::tab_compare_companies),
// Widgets::Tab({ 34, 15 }, { 31, 27 }, WindowColour::secondary, ImageIds::tab, StringIds::tab_company_performance),
// Widgets::Tab({ 65, 15 }, { 31, 27 }, WindowColour::secondary, ImageIds::tab, StringIds::tab_cargo_graphs),
// Widgets::Tab({ 96, 15 }, { 31, 27 }, WindowColour::secondary, ImageIds::tab, StringIds::tab_cargo_distance_graphs),
// Widgets::Tab({ 127, 15 }, { 31, 27 }, WindowColour::secondary, ImageIds::tab, StringIds::tab_company_values),
// Widgets::Tab({ 158, 15 }, { 31, 27 }, WindowColour::secondary, ImageIds::tab, StringIds::tab_cargo_payment_rates),
// Widgets::Tab({ 189, 15 }, { 31, 27 }, WindowColour::secondary, ImageIds::tab, StringIds::tab_speed_records));
// func OnMouseUp(self Window, widgetIndex WidgetIndex_t, id [[maybe_unused]] WidgetId) 
// func OnUpdate(self Window) 
// func PrepareDraw(self Window) 
// func SwitchTab(self Window, widgetIndex WidgetIndex_t) 
// func RefreshCompanyList(self Window) 
// func DrawTabs(self Window, drawingCtx Gfx::DrawingContext) 
// func DrawGraphAndLegend(self Window, drawingCtx Gfx::DrawingContext) 
// namespace CompanyList
// static constexpr Ui::Size kWindowSize = { 640, 272 };
const RowHeight uint8 = 25
type Widx int

const (
	Sort_name Widx = 11
	Sort_status
	Sort_performance
	Sort_value
	Scrollview
	Status_bar
)
// static constexpr auto widgets = makeWidgets(
// Common::makeCommonWidgets(640, 272, StringIds::title_company_list),
// Widgets::TableHeader({ 4, 43 }, { 175, 12 }, WindowColour::secondary, ImageIds::null, StringIds::tooltip_sort_company_name),
// Widgets::TableHeader({ 179, 43 }, { 210, 12 }, WindowColour::secondary, ImageIds::null, StringIds::tooltip_sort_company_status),
// Widgets::TableHeader({ 389, 43 }, { 145, 12 }, WindowColour::secondary, ImageIds::null, StringIds::tooltip_sort_company_performance),
// Widgets::TableHeader({ 534, 43 }, { 100, 12 }, WindowColour::secondary, ImageIds::null, StringIds::tooltip_sort_company_value),
// Widgets::ScrollView({ 3, 56 }, { 634, 201 }, WindowColour::secondary, Scrollbars::vertical),
// Widgets::Label({ 3, kWindowSize.height - 17 }, { kWindowSize.width - kResizeHandleSize, 10 }, WindowColour::secondary, ContentAlign::left, StringIds::black_stringid)
// );
type SortMode int

const (
	Name SortMode = iota
	Status
	Performance
	Value
)
// // 0x004360A2
// func OnMouseUp(self Window, widgetIndex WidgetIndex_t, id [[maybe_unused]] WidgetId) 
// switch (widgetIndex)
// case Common::widx::close_button:
// WindowManager::close(&self);
// break;
// case Common::widx::tab_company_list:
// case Common::widx::tab_performance:
// case Common::widx::tab_cargo_units:
// case Common::widx::tab_cargo_distance:
// case Common::widx::tab_values:
// case Common::widx::tab_payment_rates:
// case Common::widx::tab_speed_records:
// Common::switchTab(self, widgetIndex);
// break;
// case sort_name:
// case sort_status:
// case sort_performance:
// case sort_value:
// auto sortMode = widgetIndex - widx::sort_name;
// if (self.sortMode == sortMode)
// return;
// self.sortMode = sortMode;
// self.invalidate();
// self.var_83C = 0;
// self.rowHover = -1;
// Common::refreshCompanyList(self);
// break;
// // 0x004363CB
// func OnResize(self Window) 
// self.setSize(Common::kMinWindowSize, Common::kMaxWindowSize);
// // 0x00437BA0
// func OrderByName(lhs OpenLoco::Company, rhs OpenLoco::Company) bool
// char lhsString[256] = { 0 };
// StringManager::formatString(lhsString, lhs.name);
// char rhsString[256] = { 0 };
// StringManager::formatString(rhsString, rhs.name);
// func Strcmp(lhsString, rhsString) return
// // 0x00437BE1
// func OrderByStatus(lhs OpenLoco::Company, rhs OpenLoco::Company) bool
// char lhsString[256] = { 0 };
// orphan member: FormatArguments args{};
// auto statusString = CompanyManager::getOwnerStatus(lhs.id(), args);
// StringManager::formatString(lhsString, statusString, args);
// char rhsString[256] = { 0 };
// orphan member: FormatArguments args{};
// auto statusString = CompanyManager::getOwnerStatus(rhs.id(), args);
// StringManager::formatString(rhsString, statusString, args);
// func Strcmp(lhsString, rhsString) return
// // 0x00437C53
// func OrderByPerformance(lhs OpenLoco::Company, rhs OpenLoco::Company) bool
// auto lhsPerformance = lhs.performanceIndex;
// auto rhsPerformance = rhs.performanceIndex;
// return rhsPerformance < lhsPerformance;
// // 0x00437C67
// func OrderByValue(lhs OpenLoco::Company, rhs OpenLoco::Company) bool
// return rhs.companyValueHistory[0] < lhs.companyValueHistory[0];
// // 0x00437BA0, 0x00437BE1, 0x00437C53, 0x00437C67
// func GetOrder(mode SortMode, lhs OpenLoco::Company, rhs OpenLoco::Company) bool
// switch (mode)
// case SortMode::Name:
// func OrderByName(lhs, rhs) return
// case SortMode::Status:
// func OrderByStatus(lhs, rhs) return
// case SortMode::Performance:
// func OrderByPerformance(lhs, rhs) return
// case SortMode::Value:
// func OrderByValue(lhs, rhs) return
// orphan member: return false;
// // 0x00437AE2
// func UpdateCompanyList(self Window) 
// CompanyId chosenCompany = CompanyId::null;
// for (auto& company : CompanyManager::companies())
// if ((company.challengeFlags & CompanyFlags::sorted) != CompanyFlags::none)
// continue;
// if (chosenCompany == CompanyId::null)
// chosenCompany = company.id();
// continue;
// if (getOrder(SortMode(self.sortMode), company, *CompanyManager::get(chosenCompany)))
// chosenCompany = company.id();
// if (chosenCompany != CompanyId::null)
// bool shouldInvalidate = false;
// CompanyManager::get(chosenCompany)->challengeFlags |= CompanyFlags::sorted;
// if (chosenCompany != CompanyId(self.rowInfo[self.rowCount]))
// self.rowInfo[self.rowCount] = enumValue(chosenCompany);
// shouldInvalidate = true;
// self.rowCount++;
// if (self.rowCount > self.var_83C)
// self.var_83C = self.rowCount;
// shouldInvalidate = true;
// if (shouldInvalidate)
// self.invalidate();
// else
// if (self.var_83C != self.rowCount)
// self.var_83C = self.rowCount;
// self.invalidate();
// Common::refreshCompanyList(self);
// // 0x004362C0
// func OnUpdate(self Window) 
// self.frameNo++;
// self.callPrepareDraw();
// WindowManager::invalidateWidget(WindowType::companyList, self.number, self.currentTab + Common::widx::tab_company_list);
// _hoverItemTicks++;
// // Add three companies every tick.
// updateCompanyList(self);
// updateCompanyList(self);
// updateCompanyList(self);
// // 0x004362F7
// func Event_08(self Window) 
// self.flags |= WindowFlags::notScrollView;
// // 0x004362FF
// func Event_09(self Window) 
// if (!self.hasFlags(WindowFlags::notScrollView))
// return;
// if (self.rowHover == -1)
// return;
// self.rowHover = -1;
// self.invalidate();
// // 0x00436321
// func GetScrollSize(self Window, scrollIndex [[maybe_unused]] uint32_t, scrollWidth [[maybe_unused]] int32_t, scrollHeight int32) 
// scrollHeight = self.var_83C * kRowHeight;
// // 0x004363A0
// func OnScrollMouseDown(self Window, x [[maybe_unused]] int16_t, y int16, scroll_index [[maybe_unused]] uint8_t) 
// uint16_t currentRow = y / kRowHeight;
// if (currentRow > self.var_83C)
// return;
// CompanyId currentCompany = CompanyId(self.rowInfo[currentRow]);
// if (currentCompany == CompanyId::null)
// return;
// CompanyWindow::open(currentCompany);
// // 0x00436361
// func OnScrollMouseOver(self Window, x [[maybe_unused]] int16_t, y int16, scroll_index [[maybe_unused]] uint8_t) 
// self.flags &= ~(WindowFlags::notScrollView);
// uint16_t currentRow = y / kRowHeight;
// int16_t currentCompany = -1;
// if (currentRow < self.var_83C)
// currentCompany = self.rowInfo[currentRow];
// if (self.rowHover == currentCompany)
// return;
// self.rowHover = currentCompany;
// self.invalidate();
// // 0x004362B6
// static std::optional<FormatArguments> tooltip([[maybe_unused]] Window& self, [[maybe_unused]] WidgetIndex_t widgetIndex, [[maybe_unused]] const WidgetId id)
// orphan member: FormatArguments args{};
// args.push(StringIds::tooltip_scroll_company_list);
// orphan member: return args;
// // 0x0043632C
// static Ui::CursorId cursor(Window& self, WidgetIndex_t widgetIdx, [[maybe_unused]] const WidgetId id, [[maybe_unused]] int16_t xPos, int16_t yPos, Ui::CursorId fallback)
// if (widgetIdx != widx::scrollview)
// orphan member: return fallback;
// uint16_t currentIndex = yPos / kRowHeight;
// if (currentIndex < self.var_83C && self.rowInfo[currentIndex] != -1)
// return CursorId::handPointer;
// orphan member: return fallback;
// // 0x00435D07
// func PrepareDraw(self Window) 
// Common::prepareDraw(self);
// self.widgets[widx::scrollview].right = self.width - 4;
// self.widgets[widx::scrollview].bottom = self.height - 14;
// self.widgets[widx::status_bar].right = self.width - kResizeHandleSize - 1;
// // Reposition header buttons
// self.widgets[widx::sort_name].right = std::min(178, self.width - 8);
// self.widgets[widx::sort_status].left = std::min(179, self.width - 8);
// self.widgets[widx::sort_status].right = std::min(388, self.width - 8);
// self.widgets[widx::sort_performance].left = std::min(389, self.width - 8);
// self.widgets[widx::sort_performance].right = std::min(533, self.width - 8);
// self.widgets[widx::sort_value].left = std::min(534, self.width - 8);
// self.widgets[widx::sort_value].right = std::min(633, self.width - 8);
// // Set header button captions
// self.widgets[widx::sort_name].text = self.sortMode == SortMode::Name ? StringIds::table_header_company_name_desc : StringIds::table_header_company_name;
// self.widgets[widx::sort_status].text = self.sortMode == SortMode::Status ? StringIds::table_header_company_status_desc : StringIds::table_header_company_status;
// self.widgets[widx::sort_performance].text = self.sortMode == SortMode::Performance ? StringIds::table_header_company_performance_desc : StringIds::table_header_company_performance;
// self.widgets[widx::sort_value].text = self.sortMode == SortMode::Value ? StringIds::table_header_company_value_desc : StringIds::table_header_company_value;
// // Reposition status bar
// auto& widget = self.widgets[widx::status_bar];
// widget.top = self.height - 13;
// widget.bottom = self.height - 3;
// // Set status bar text
// orphan member: FormatArguments args{ widget.textArgs };
// args.push(self.var_83C == 1 ? StringIds::company_singular : StringIds::companies_plural);
// args.push(self.var_83C);
// // 0x00435E56
// func Draw(self Window, drawingCtx Gfx::DrawingContext) 
// self.draw(drawingCtx);
// Common::drawTabs(self, drawingCtx);
// // 0x00435EA7
// func DrawScroll(self Window, drawingCtx Gfx::DrawingContext, scrollIndex [[maybe_unused]] uint32_t) 
// const auto& rt = drawingCtx.currentRenderTarget();
// auto tr = Gfx::TextRenderer(drawingCtx);
// auto colour = Colours::getShade(self.getColour(WindowColour::secondary).c(), 3);
// drawingCtx.clearSingle(colour);
// auto yBottom = 0;
// for (auto i = 0; i < self.var_83C; i++, yBottom += 25)
// auto yTop = yBottom + 25;
// if (yTop <= rt.y)
// continue;
// yTop = rt.y + rt.height;
// if (yBottom >= yTop)
// break;
// auto rowItem = self.rowInfo[i];
// if (rowItem == -1)
// continue;
// auto stringId = StringIds::black_stringid;
// if (rowItem == self.rowHover)
// drawingCtx.drawRect(0, yBottom, self.width, 24, enumValue(ExtColour::unk30), Gfx::RectFlags::transparent);
// stringId = StringIds::wcolour2_stringid;
// auto company = CompanyManager::get(CompanyId(rowItem));
// auto competitorObj = ObjectManager::get<CompetitorObject>(company->competitorId);
// auto imageId = Gfx::recolour(competitorObj->images[enumValue(company->ownerEmotion)], company->mainColours.primary);
// orphan member: FormatArguments args{};
// args.push(StringIds::table_item_company);
// args.push(imageId);
// args.push(company->name);
// auto point = Point(0, yBottom - 1);
// tr.drawStringLeftClipped(point, 173, Colour::black, stringId, args);
// orphan member: FormatArguments args{};
// args.skip(sizeof(StringId));
// StringId ownerStatus = CompanyManager::getOwnerStatus(company->id(), args);
// args.rewind();
// args.push(ownerStatus);
// auto point = Point(175, yBottom + 7);
// tr.drawStringLeftClipped(point, 208, Colour::black, stringId, args);
// auto performanceStringId = StringIds::performance_index;
// if ((company->challengeFlags & CompanyFlags::increasedPerformance) != CompanyFlags::none && (company->challengeFlags & CompanyFlags::decreasedPerformance) != CompanyFlags::none)
// performanceStringId = StringIds::performance_index_decrease;
// if ((company->challengeFlags & CompanyFlags::increasedPerformance) != CompanyFlags::none)
// performanceStringId = StringIds::performance_index_increase;
// orphan member: FormatArguments args{};
// args.push(performanceStringId);
// formatPerformanceIndex(company->performanceIndex, args);
// auto point = Point(385, yBottom - 1);
// tr.drawStringLeftClipped(point, 143, Colour::black, stringId, args);
// orphan member: FormatArguments args{};
// args.push(StringIds::company_value_currency);
// args.push(company->companyValueHistory[0]);
// auto point = Point(530, yBottom - 1);
// tr.drawStringLeftClipped(point, 98, Colour::black, stringId, args);
// // 0x00436198
// func TabReset(self Window) 
// self.minWidth = Common::kMinWindowSize.width;
// self.minHeight = Common::kMinWindowSize.height;
// self.maxWidth = Common::kMaxWindowSize.width;
// self.maxHeight = Common::kMaxWindowSize.height;
// self.width = kWindowSize.width;
// self.height = kWindowSize.height;
// self.var_83C = 0;
// self.rowHover = -1;
// Common::refreshCompanyList(self);
// static constexpr WindowEventList kEvents = {
// .onMouseUp = onMouseUp,
// .onResize = onResize,
// .onUpdate = onUpdate,
// .event_08 = event_08,
// .event_09 = event_09,
// .getScrollSize = getScrollSize,
// .scrollMouseDown = onScrollMouseDown,
// .scrollMouseOver = onScrollMouseOver,
// .tooltip = tooltip,
// .cursor = cursor,
// .prepareDraw = prepareDraw,
// .draw = draw,
// .drawScroll = drawScroll,
// static const WindowEventList& getEvents()
// orphan member: return kEvents;
// // 0x00435BC8
// Window* open()
// auto window = WindowManager::bringToFront(WindowType::companyList);
// if (window != nullptr)
// if (ToolManager::isToolActive(ToolManager::getToolWindowType(), ToolManager::getToolWindowNumber()))
// ToolManager::toolCancel();
// window = WindowManager::bringToFront(WindowType::companyList);
// if (window == nullptr)
// static constexpr Ui::Size kWindowSize = { 640, 272 };
// window = WindowManager::createWindow(WindowType::companyList, kWindowSize, WindowFlags::none, CompanyList::getEvents());
// window->frameNo = 0;
// window->savedView.clear();
// window->flags |= WindowFlags::resizable;
// window->sortMode = 2;
// window->var_83C = 0;
// window->rowHover = -1;
// Common::refreshCompanyList(*window);
// auto skin = ObjectManager::get<InterfaceSkinObject>();
// window->setColour(WindowColour::primary, skin->windowTitlebarColour);
// window->setColour(WindowColour::secondary, skin->windowColour);
// window->var_854 = 0;
// window->currentTab = 0;
// window->minWidth = Common::kMinWindowSize.width;
// window->minHeight = Common::kMinWindowSize.height;
// window->maxWidth = Common::kMaxWindowSize.width;
// window->maxHeight = Common::kMaxWindowSize.height;
// window->invalidate();
// window->setWidgets(CompanyList::widgets);
// window->holdableWidgets = 0;
// window->eventHandlers = &CompanyList::getEvents();
// window->activatedWidgets = 0;
// window->initScrollWidgets();
// orphan member: return window;
// func RemoveCompany(id CompanyId) 
// auto* w = WindowManager::find(WindowType::companyList);
// if (w != nullptr)
// for (auto i = 0; i < w->var_83C; i++)
// if (static_cast<CompanyId>(w->rowInfo[i]) == id)
// w->rowInfo[i] = -1;
// WindowManager::invalidate(WindowType::companyList);
// // 0x00435C69
// func OpenPerformanceIndexes() 
// auto window = open();
// window->callOnMouseUp(Common::widx::tab_performance, window->widgets[Common::widx::tab_performance].id);
// namespace CompanyPerformance
// static constexpr Ui::Size kWindowSize = { 635, 322 };
// static constexpr auto widgets = makeWidgets(
// Common::makeCommonWidgets(635, 322, StringIds::title_company_performance)
// );
// // 0x004366D7
// func OnResize(self Window) 
// self.setSize(kWindowSize, Common::kMaxWindowSize);
// // 0x00436490
// func Draw(self Window, drawingCtx Gfx::DrawingContext) 
// self.draw(drawingCtx);
// Common::drawTabs(self, drawingCtx);
// _graphSettings.left = self.x + 4;
// _graphSettings.top = self.y + self.widgets[Common::widx::panel].top + 4;
// _graphSettings.width = self.width - kLegendWidth - kLegendMargin - 2 * kWindowPadding;
// _graphSettings.height = self.height - self.widgets[Common::widx::panel].top - 2 * kWindowPadding;
// _graphSettings.yOffset = 17;
// _graphSettings.xOffset = 40;
// _graphSettings.yAxisLabelIncrement = 20;
// _graphSettings.linesToExclude = 0;
// uint16_t maxHistorySize = 1;
// for (auto& company : CompanyManager::companies())
// if (maxHistorySize < company.historySize)
// maxHistorySize = company.historySize;
// uint8_t count = 0;
// for (auto& company : CompanyManager::companies())
// auto companyId = company.id();
// auto companyColour = CompanyManager::getCompanyColour(companyId);
// _graphSettings.yData[count] = reinterpret_cast<std::byte*>(&company.performanceIndexHistory[0]);
// _graphSettings.dataStart[count] = maxHistorySize - company.historySize;
// _graphSettings.lineColour[count] = Colours::getShade(companyColour, 6);
// _graphSettings.itemId[count] = enumValue(companyId);
// count++;
// _graphSettings.lineCount = count;
// _graphSettings.dataEnd = maxHistorySize;
// _graphSettings.dataTypeSize = 2;
// _graphSettings.xLabel = StringIds::rawdate_short;
// _graphSettings.yLabel = StringIds::percentage_one_decimal_place;
// _graphSettings.xAxisTickIncrement = (_graphSettings.width - _graphSettings.xOffset) / 120;
// _graphSettings.xAxisLabelIncrement = 12;
// _graphSettings.dword_113DD86 = 0;
// _graphSettings.yAxisStepSize = 100;
// _graphSettings.flags = GraphFlags::dataFrontToBack;
// Common::drawGraphAndLegend(self, drawingCtx);
// // 0x004361D8
// func TabReset(self Window) 
// self.minWidth = kWindowSize.width;
// self.minHeight = kWindowSize.height;
// self.maxWidth = kWindowSize.width;
// self.maxHeight = kWindowSize.height;
// self.width = kWindowSize.width;
// self.height = kWindowSize.height;
// static constexpr WindowEventList kEvents = {
// .onMouseUp = Common::onMouseUp,
// .onResize = onResize,
// .onUpdate = Common::onUpdate,
// .prepareDraw = Common::prepareDraw,
// .draw = draw,
// static const WindowEventList& getEvents()
// orphan member: return kEvents;
// namespace CargoUnits
// static constexpr Ui::Size kWindowSize = { 640, 272 };
// static constexpr auto widgets = makeWidgets(
// Common::makeCommonWidgets(635, 322, StringIds::title_company_cargo_units)
// );
// // 0x004369FB
// func OnResize(self Window) 
// self.setSize(kWindowSize, Common::kMaxWindowSize);
// // 0x004367B4
// func Draw(self Window, drawingCtx Gfx::DrawingContext) 
// self.draw(drawingCtx);
// Common::drawTabs(self, drawingCtx);
// _graphSettings.left = self.x + 4;
// _graphSettings.top = self.y + self.widgets[Common::widx::panel].top + 4;
// _graphSettings.width = self.width - kLegendWidth - kLegendMargin - 2 * kWindowPadding;
// _graphSettings.height = self.height - self.widgets[Common::widx::panel].top - 2 * kWindowPadding;
// _graphSettings.yOffset = 17;
// _graphSettings.xOffset = 45;
// _graphSettings.yAxisLabelIncrement = 25;
// _graphSettings.linesToExclude = 0;
// uint16_t maxHistorySize = 1;
// for (auto& company : CompanyManager::companies())
// if (maxHistorySize < company.historySize)
// maxHistorySize = company.historySize;
// uint8_t count = 0;
// for (auto& company : CompanyManager::companies())
// auto companyId = company.id();
// auto companyColour = CompanyManager::getCompanyColour(companyId);
// _graphSettings.yData[count] = reinterpret_cast<std::byte*>(&company.cargoUnitsDeliveredHistory[0]);
// _graphSettings.dataStart[count] = maxHistorySize - company.historySize;
// _graphSettings.lineColour[count] = Colours::getShade(companyColour, 6);
// _graphSettings.itemId[count] = enumValue(companyId);
// count++;
// _graphSettings.lineCount = count;
// _graphSettings.dataEnd = maxHistorySize;
// _graphSettings.dataTypeSize = 4;
// _graphSettings.xLabel = StringIds::rawdate_short;
// _graphSettings.yLabel = StringIds::cargo_units_delivered;
// _graphSettings.xAxisTickIncrement = (_graphSettings.width - _graphSettings.xOffset) / 120;
// _graphSettings.xAxisLabelIncrement = 12;
// _graphSettings.dword_113DD86 = 0;
// _graphSettings.yAxisStepSize = 1000;
// _graphSettings.flags = GraphFlags::dataFrontToBack;
// Common::drawGraphAndLegend(self, drawingCtx);
// // 0x00436201
// func TabReset(self Window) 
// self.minWidth = kWindowSize.width;
// self.minHeight = kWindowSize.height;
// self.maxWidth = kWindowSize.width;
// self.maxHeight = kWindowSize.height;
// self.width = kWindowSize.width;
// self.height = kWindowSize.height;
// static constexpr WindowEventList kEvents = {
// .onMouseUp = Common::onMouseUp,
// .onResize = onResize,
// .onUpdate = Common::onUpdate,
// .prepareDraw = Common::prepareDraw,
// .draw = draw,
// static const WindowEventList& getEvents()
// orphan member: return kEvents;
// namespace CargoDistance
// static constexpr Ui::Size kWindowSize = { 660, 272 };
// static constexpr auto widgets = makeWidgets(
// Common::makeCommonWidgets(635, 322, StringIds::title_cargo_distance_graphs)
// );
// // 0x00436D1F
// func OnResize(self Window) 
// self.setSize(kWindowSize, Common::kMaxWindowSize);
// // 0x00436AD8
// func Draw(self Window, drawingCtx Gfx::DrawingContext) 
// self.draw(drawingCtx);
// Common::drawTabs(self, drawingCtx);
// _graphSettings.left = self.x + 4;
// _graphSettings.top = self.y + self.widgets[Common::widx::panel].top + 4;
// _graphSettings.width = self.width - kLegendWidth - kLegendMargin - 2 * kWindowPadding;
// _graphSettings.height = self.height - self.widgets[Common::widx::panel].top - 2 * kWindowPadding;
// _graphSettings.yOffset = 17;
// _graphSettings.xOffset = 65;
// _graphSettings.yAxisLabelIncrement = 25;
// _graphSettings.linesToExclude = 0;
// uint16_t maxHistorySize = 1;
// for (auto& company : CompanyManager::companies())
// if (maxHistorySize < company.historySize)
// maxHistorySize = company.historySize;
// uint8_t count = 0;
// for (auto& company : CompanyManager::companies())
// auto companyId = company.id();
// auto companyColour = CompanyManager::getCompanyColour(companyId);
// _graphSettings.yData[count] = reinterpret_cast<std::byte*>(&company.cargoUnitsDistanceHistory[0]);
// _graphSettings.dataStart[count] = maxHistorySize - company.historySize;
// _graphSettings.lineColour[count] = Colours::getShade(companyColour, 6);
// _graphSettings.itemId[count] = enumValue(companyId);
// count++;
// _graphSettings.lineCount = count;
// _graphSettings.dataEnd = maxHistorySize;
// _graphSettings.dataTypeSize = 4;
// _graphSettings.xLabel = StringIds::rawdate_short;
// _graphSettings.yLabel = StringIds::cargo_units_delivered;
// _graphSettings.xAxisTickIncrement = (_graphSettings.width - _graphSettings.xOffset) / 120;
// _graphSettings.xAxisLabelIncrement = 12;
// _graphSettings.dword_113DD86 = 0;
// _graphSettings.yAxisStepSize = 1000;
// _graphSettings.flags = GraphFlags::dataFrontToBack;
// Common::drawGraphAndLegend(self, drawingCtx);
// // 0x00436227
// func TabReset(self Window) 
// self.minWidth = kWindowSize.width;
// self.minHeight = kWindowSize.height;
// self.maxWidth = kWindowSize.width;
// self.maxHeight = kWindowSize.height;
// self.width = kWindowSize.width;
// self.height = kWindowSize.height;
// static constexpr WindowEventList kEvents = {
// .onMouseUp = Common::onMouseUp,
// .onResize = onResize,
// .onUpdate = Common::onUpdate,
// .prepareDraw = Common::prepareDraw,
// .draw = draw,
// static const WindowEventList& getEvents()
// orphan member: return kEvents;
// namespace CompanyValues
// static constexpr Ui::Size kWindowSize = { 685, 322 };
// static constexpr auto widgets = makeWidgets(
// Common::makeCommonWidgets(685, 322, StringIds::title_company_values)
// );
// // 0x00437043
// func OnResize(self Window) 
// self.setSize(kWindowSize, Common::kMaxWindowSize);
// // 0x00436DFC
// func Draw(self Window, drawingCtx Gfx::DrawingContext) 
// self.draw(drawingCtx);
// Common::drawTabs(self, drawingCtx);
// _graphSettings.left = self.x + 4;
// _graphSettings.top = self.y + self.widgets[Common::widx::panel].top + 4;
// _graphSettings.width = self.width - kLegendWidth - kLegendMargin - 2 * kWindowPadding;
// _graphSettings.height = self.height - self.widgets[Common::widx::panel].top - 2 * kWindowPadding;
// _graphSettings.yOffset = 17;
// _graphSettings.xOffset = 90;
// _graphSettings.yAxisLabelIncrement = 25;
// _graphSettings.linesToExclude = 0;
// uint16_t maxHistorySize = 1;
// for (auto& company : CompanyManager::companies())
// if (maxHistorySize < company.historySize)
// maxHistorySize = company.historySize;
// uint8_t count = 0;
// for (auto& company : CompanyManager::companies())
// auto companyId = company.id();
// auto companyColour = CompanyManager::getCompanyColour(companyId);
// _graphSettings.yData[count] = reinterpret_cast<std::byte*>(&company.companyValueHistory[0]);
// _graphSettings.dataStart[count] = maxHistorySize - company.historySize;
// _graphSettings.lineColour[count] = Colours::getShade(companyColour, 6);
// _graphSettings.itemId[count] = enumValue(companyId);
// count++;
// _graphSettings.lineCount = count;
// _graphSettings.dataEnd = maxHistorySize;
// _graphSettings.dataTypeSize = 6;
// _graphSettings.xLabel = StringIds::rawdate_short;
// _graphSettings.yLabel = StringIds::small_company_value_currency;
// _graphSettings.xAxisTickIncrement = (_graphSettings.width - _graphSettings.xOffset) / 120;
// _graphSettings.xAxisLabelIncrement = 12;
// _graphSettings.dword_113DD86 = 0;
// _graphSettings.yAxisStepSize = 10000;
// _graphSettings.flags = GraphFlags::dataFrontToBack;
// Common::drawGraphAndLegend(self, drawingCtx);
// // 0x0043624D
// func TabReset(self Window) 
// self.minWidth = kWindowSize.width;
// self.minHeight = kWindowSize.height;
// self.maxWidth = kWindowSize.width;
// self.maxHeight = kWindowSize.height;
// self.width = kWindowSize.width;
// self.height = kWindowSize.height;
// static constexpr WindowEventList kEvents = {
// .onMouseUp = Common::onMouseUp,
// .onResize = onResize,
// .onUpdate = Common::onUpdate,
// .prepareDraw = Common::prepareDraw,
// .draw = draw,
// static const WindowEventList& getEvents()
// orphan member: return kEvents;
// namespace CargoPaymentRates
// static constexpr Ui::Size kWindowSize = { 495, 342 };
// static constexpr auto widgets = makeWidgets(
// Common::makeCommonWidgets(495, 342, StringIds::title_cargo_payment_rates)
// );
// // 0x0043737D
// func OnResize(self Window) 
// self.setSize(kWindowSize, Common::kMaxWindowSize);
// // 0x004F9442
// static constexpr Colour _cargoLineColour[32] = {
// Colour::red,
// Colour::mutedPurple,
// Colour::yellow,
// Colour::blue,
// Colour::orange,
// Colour::green,
// Colour::mutedDarkRed,
// Colour::mutedDarkTeal,
// Colour::mutedDarkYellow,
// Colour::black,
// Colour::white,
// Colour::mutedDarkPurple,
// Colour::purple,
// Colour::darkBlue,
// Colour::mutedTeal,
// Colour::darkGreen,
// Colour::mutedSeaGreen,
// Colour::mutedGrassGreen,
// Colour::mutedAvocadoGreen,
// Colour::mutedOliveGreen,
// Colour::darkYellow,
// Colour::amber,
// Colour::grey,
// Colour::darkOrange,
// Colour::mutedYellow,
// Colour::brown,
// Colour::mutedOrange,
// Colour::darkRed,
// Colour::darkPink,
// Colour::pink,
// Colour::mutedRed,
// Colour::grey,
// // 0x00437949
// func DrawGraphLegend(self Window, drawingCtx Gfx::DrawingContext, x int16, y int16) 
// auto tr = Gfx::TextRenderer(drawingCtx);
// auto cargoCount = 0;
// for (uint8_t i = 0; i < ObjectManager::getMaxObjects(ObjectType::cargo); i++)
// auto cargo = ObjectManager::get<CargoObject>(i);
// if (cargo == nullptr)
// continue;
// auto colour = _cargoLineColour[i];
// auto palette = Colours::getShade(colour, 6);
// auto stringId = StringIds::small_black_string;
// if (self->var_854 & (1 << cargoCount))
// stringId = StringIds::small_white_string;
// if (!(self->var_854 & (1 << cargoCount)) || !(_hoverItemTicks & (1 << 2)))
// drawingCtx.fillRect(x, y + 3, x + 4, y + 7, palette, Gfx::RectFlags::none);
// auto args = FormatArguments();
// args.push(cargo->name);
// auto point = Point(x + 6, y);
// tr.drawStringLeftClipped(point, 94, Colour::black, stringId, args);
// y += 10;
// cargoCount++;
// // 0x00437120
// func Draw(self Window, drawingCtx Gfx::DrawingContext) 
// auto tr = Gfx::TextRenderer(drawingCtx);
// self.draw(drawingCtx);
// Common::drawTabs(self, drawingCtx);
// _graphSettings.left = self.x + 4;
// _graphSettings.top = self.y + self.widgets[Common::widx::panel].top + 14;
// _graphSettings.width = self.width - kLegendWidth - kLegendMargin - 2 * kWindowPadding;
// _graphSettings.height = self.height - self.widgets[Common::widx::panel].top - 20 - 2 * kWindowPadding;
// _graphSettings.yOffset = 17;
// _graphSettings.xOffset = 80;
// _graphSettings.yAxisLabelIncrement = 25;
// _graphSettings.linesToExclude = 0;
// auto count = 0;
// for (uint8_t i = 0; i < ObjectManager::getMaxObjects(ObjectType::cargo); i++)
// auto cargo = ObjectManager::get<CargoObject>(i);
// if (cargo == nullptr)
// continue;
// auto colour = _cargoLineColour[i];
// auto deliveredCargoPayment = Economy::getDeliveryCargoPaymentsTable(i);
// _graphSettings.yData[count] = reinterpret_cast<std::byte*>(deliveredCargoPayment.data());
// _graphSettings.dataStart[count] = 0;
// _graphSettings.lineColour[count] = Colours::getShade(colour, 6);
// _graphSettings.itemId[count] = i;
// count++;
// _graphSettings.lineCount = count;
// _graphSettings.dataEnd = static_cast<uint16_t>(std::size(Economy::getDeliveryCargoPaymentsTable(0)));
// _graphSettings.dataTypeSize = 4;
// _graphSettings.xLabel = StringIds::cargo_delivered_days;
// _graphSettings.yLabel = StringIds::cargo_delivered_currency;
// _graphSettings.xAxisTickIncrement = (_graphSettings.width - _graphSettings.xOffset) / 60;
// _graphSettings.xAxisLabelIncrement = 20;
// _graphSettings.dword_113DD86 = 0;
// _graphSettings.yAxisStepSize = 0;
// _graphSettings.flags = GraphFlags::none;
// _graphSettings.xAxisRange = 2;
// _graphSettings.xAxisStepSize = 2;
// _graphSettings.pointFlags = GraphPointFlags::drawLines;
// Ui::drawGraph(_graphSettings, self, drawingCtx);
// if (self.var_854 != 0)
// auto i = 0;
// while (Numerics::bitScanForward(self.var_854) != _graphSettings.itemId[i])
// i++;
// // Exclude all lines except highlighted data
// _graphSettings.linesToExclude = 0xFFFFFFFF & ~(1 << i);
// if (_hoverItemTicks & 4)
// _graphSettings.lineColour[i] = PaletteIndex::black0;
// _graphSettings.flags |= GraphFlags::hideAxesAndLabels;
// Ui::drawGraph(_graphSettings, self, drawingCtx);
// auto x = self.width + self.x - kLegendWidth - kWindowPadding;
// auto y = self.y + 52;
// drawGraphLegend(&self, drawingCtx, x, y);
// auto canvasMidX = _graphSettings.xOffset + (_graphSettings.width - _graphSettings.xOffset) / 2;
// // Chart title
// auto point = Point(self.x + canvasMidX, self.widgets[Common::widx::panel].top + self.y + 1);
// orphan member: FormatArguments args{};
// args.push<uint16_t>(100);
// args.push<uint16_t>(10);
// tr.drawStringCentred(point, Colour::black, StringIds::cargo_deliver_graph_title, args);
// // X axis label ("Transit time")
// auto point = Point(self.x + canvasMidX, self.height + self.y - 13);
// tr.drawStringCentred(point, Colour::black, StringIds::cargo_transit_time);
// // 0x004379F2
// func SetLegendHover(self Window, x int16, y int16) 
// uint32_t selectedCargo = 0;
// if (!Input::hasFlag(Input::Flags::rightMousePressed))
// const auto location = Input::getMouseLocation2();
// auto* frontWindow = WindowManager::findAt(location);
// const auto xDiff = location.x - x;
// const auto yDiff = location.y - y;
// if (frontWindow != nullptr && frontWindow == self && xDiff <= kLegendWidth && xDiff >= 0 && yDiff < 320 && yDiff >= 0)
// auto listY = yDiff;
// uint8_t cargoItem = 0;
// for (; cargoItem < ObjectManager::getMaxObjects(ObjectType::cargo); ++cargoItem)
// auto* cargoObj = ObjectManager::get<CargoObject>(cargoItem);
// if (cargoObj == nullptr)
// continue;
// listY -= 10;
// if (listY <= 0)
// selectedCargo = 1ULL << cargoItem;
// break;
// if (self->var_854 != selectedCargo)
// // TODO: var_854 is 16 bits but selectedCargo is 32 bits. Only the first 15 cargo types can be selected.
// self->var_854 = selectedCargo;
// self->invalidate();
// if (self->var_854 != 0)
// self->invalidate();
// // 0x00436273
// func TabReset(self Window) 
// self.minWidth = kWindowSize.width;
// self.minHeight = kWindowSize.height;
// self.maxWidth = kWindowSize.width;
// self.maxHeight = kWindowSize.height;
// self.width = kWindowSize.width;
// self.height = kWindowSize.height;
// Economy::buildDeliveredCargoPaymentsTable();
// static constexpr WindowEventList kEvents = {
// .onMouseUp = Common::onMouseUp,
// .onResize = onResize,
// .onUpdate = Common::onUpdate,
// .prepareDraw = Common::prepareDraw,
// .draw = draw,
// static const WindowEventList& getEvents()
// orphan member: return kEvents;
// namespace CompanySpeedRecords
// static constexpr Ui::Size kWindowSize = { 495, 169 };
// static constexpr auto widgets = makeWidgets(
// Common::makeCommonWidgets(495, 169, StringIds::title_speed_records)
// );
// // 0x00437591
// func OnResize(self Window) 
// self.setSize(kWindowSize, kWindowSize);
// // 0x0043745A
// func Draw(self Window, drawingCtx Gfx::DrawingContext) 
// auto tr = Gfx::TextRenderer(drawingCtx);
// self.draw(drawingCtx);
// Common::drawTabs(self, drawingCtx);
// auto y = self.y + 47;
// for (auto i = 0; i < 3; i++)
// auto recordSpeed = CompanyManager::getRecords().speed[i];
// if (recordSpeed == 0_mph)
// continue;
// orphan member: FormatArguments args{};
// args.push(recordSpeed);
// const StringId string[] = {
// StringIds::land_speed_record,
// StringIds::air_speed_record,
// StringIds::water_speed_record,
// auto point = Point(self.x + 4, y);
// tr.drawStringLeft(point, Colour::black, string[i], args);
// y += 11;
// auto companyId = CompanyManager::getRecords().company[i];
// if (companyId != CompanyId::null)
// auto company = CompanyManager::get(companyId);
// auto competitorObj = ObjectManager::get<CompetitorObject>(company->competitorId);
// auto imageId = competitorObj->images[enumValue(company->ownerEmotion)];
// imageId = Gfx::recolour(imageId, company->mainColours.primary);
// auto x = self.x + 4;
// drawingCtx.drawImage(x, y, imageId);
// y += 7;
// auto point = Point(self.x + 33, y);
// orphan member: FormatArguments args{};
// args.push(company->name);
// args.push<uint16_t>(0);
// args.push(CompanyManager::getRecords().date[i]);
// tr.drawStringLeft(point, Colour::black, StringIds::record_date_achieved, args);
// y += 17;
// y += 5;
// static constexpr WindowEventList kEvents = {
// .onMouseUp = Common::onMouseUp,
// .onResize = onResize,
// .onUpdate = Common::onUpdate,
// .prepareDraw = Common::prepareDraw,
// .draw = draw,
// static const WindowEventList& getEvents()
// orphan member: return kEvents;
// namespace Common
type TabInformation struct {
// std::span<const Widget> widgets;
// const widx widgetIndex;
// const WindowEventList& events;
}
// // clang-format off
// static TabInformation tabInformationByTabOffset[] = {
// { CompanyList::widgets,         widx::tab_company_list,   CompanyList::getEvents()         },
// { CompanyPerformance::widgets,  widx::tab_performance,    CompanyPerformance::getEvents()  },
// { CargoUnits::widgets,          widx::tab_cargo_units,    CargoUnits::getEvents()          },
// { CargoDistance::widgets,       widx::tab_cargo_distance, CargoDistance::getEvents()       },
// { CompanyValues::widgets,       widx::tab_values,         CompanyValues::getEvents()       },
// { CargoPaymentRates::widgets,   widx::tab_payment_rates,  CargoPaymentRates::getEvents()   },
// { CompanySpeedRecords::widgets, widx::tab_speed_records,  CompanySpeedRecords::getEvents() },
// // clang-format on
// // 0x0043667B
// func OnMouseUp(self Window, widgetIndex WidgetIndex_t, id [[maybe_unused]] WidgetId) 
// switch (widgetIndex)
// case Common::widx::close_button:
// WindowManager::close(&self);
// break;
// case Common::widx::tab_company_list:
// case Common::widx::tab_performance:
// case Common::widx::tab_cargo_units:
// case Common::widx::tab_cargo_distance:
// case Common::widx::tab_values:
// case Common::widx::tab_payment_rates:
// case Common::widx::tab_speed_records:
// Common::switchTab(self, widgetIndex);
// break;
// // 0x004378BA
// func SetLegendHover(self Window, x int16, y int16) 
// uint32_t selectedCompany = 0;
// if (!Input::hasFlag(Input::Flags::rightMousePressed))
// const auto location = Input::getMouseLocation2();
// auto* frontWindow = WindowManager::findAt(location);
// const auto xDiff = location.x - x;
// const auto yDiff = location.y - y;
// if (frontWindow != nullptr && frontWindow == self && xDiff <= kLegendWidth && xDiff >= 0 && yDiff < 150 && yDiff >= 0)
// auto listY = yDiff;
// for (auto& company : CompanyManager::companies())
// listY -= 10;
// if (listY <= 0)
// selectedCompany = 1ULL << enumValue(company.id());
// break;
// if (self->var_854 != selectedCompany)
// self->var_854 = selectedCompany;
// self->invalidate();
// if (self->var_854 != 0)
// self->invalidate();
// // 0x00437570
// func OnUpdate(self Window) 
// self.frameNo++;
// self.callPrepareDraw();
// WindowManager::invalidateWidget(WindowType::companyList, self.number, self.currentTab + Common::widx::tab_company_list);
// auto legendX = self.x + self.width - kWindowPadding - kLegendWidth;
// auto legendY = self.y + 52;
// switch (self.currentTab + widx::tab_company_list)
// case widx::tab_cargo_distance:
// case widx::tab_cargo_units:
// case widx::tab_performance:
// case widx::tab_values:
// _hoverItemTicks++;
// setLegendHover(&self, legendX, legendY);
// break;
// case widx::tab_payment_rates:
// _hoverItemTicks++;
// CargoPaymentRates::setLegendHover(&self, legendX, legendY);
// break;
// case widx::tab_speed_records:
// break;
// // 0x00436419
// func PrepareDraw(self Window) 
// // Activate the current tab
// self.activatedWidgets &= ~((1ULL << tab_cargo_distance) | (1ULL << tab_cargo_units) | (1ULL << tab_company_list) | (1ULL << tab_payment_rates) | (1ULL << tab_performance) | (1ULL << tab_speed_records) | (1ULL << tab_values));
// self.activatedWidgets |= (1ULL << Common::tabInformationByTabOffset[self.currentTab].widgetIndex);
// self.widgets[Common::widx::frame].right = self.width - 1;
// self.widgets[Common::widx::frame].bottom = self.height - 1;
// self.widgets[Common::widx::panel].right = self.width - 1;
// self.widgets[Common::widx::panel].bottom = self.height - 1;
// self.widgets[Common::widx::caption].right = self.width - 2;
// self.widgets[Common::widx::close_button].left = self.width - 15;
// self.widgets[Common::widx::close_button].right = self.width - 3;
// // 0x004360FA
// func SwitchTab(self Window, widgetIndex WidgetIndex_t) 
// if (ToolManager::isToolActive(self.type, self.number))
// ToolManager::toolCancel();
// self.currentTab = widgetIndex - widx::tab_company_list;
// self.frameNo = 0;
// self.flags &= ~(WindowFlags::beingResized);
// self.viewportRemove(0);
// const auto& tabInfo = tabInformationByTabOffset[widgetIndex - widx::tab_company_list];
// self.holdableWidgets = 0;
// self.eventHandlers = &tabInfo.events;
// self.activatedWidgets = 0;
// self.setWidgets(tabInfo.widgets);
// self.invalidate();
// switch (widgetIndex)
// case widx::tab_company_list:
// CompanyList::tabReset(self);
// break;
// case widx::tab_performance:
// CompanyPerformance::tabReset(self);
// break;
// case widx::tab_cargo_units:
// CargoUnits::tabReset(self);
// break;
// case widx::tab_cargo_distance:
// CargoDistance::tabReset(self);
// break;
// case widx::tab_values:
// CompanyValues::tabReset(self);
// break;
// case widx::tab_payment_rates:
// CargoPaymentRates::tabReset(self);
// break;
// self.callOnResize();
// self.callPrepareDraw();
// self.initScrollWidgets();
// self.invalidate();
// self.moveInsideScreenEdges();
// // 0x00437637
// func DrawTabs(self Window, drawingCtx Gfx::DrawingContext) 
// auto tr = Gfx::TextRenderer(drawingCtx);
// auto skin = ObjectManager::get<InterfaceSkinObject>();
// // Company List Tab
// uint32_t imageId = skin->img;
// imageId += InterfaceSkin::ImageIds::tab_companies;
// Widget::drawTab(self, drawingCtx, imageId, widx::tab_company_list);
// // Performance Index Tab
// static constexpr uint32_t performanceImageIds[] = {
// InterfaceSkin::ImageIds::tab_performance_index_frame0,
// InterfaceSkin::ImageIds::tab_performance_index_frame1,
// InterfaceSkin::ImageIds::tab_performance_index_frame2,
// InterfaceSkin::ImageIds::tab_performance_index_frame3,
// InterfaceSkin::ImageIds::tab_performance_index_frame4,
// InterfaceSkin::ImageIds::tab_performance_index_frame5,
// InterfaceSkin::ImageIds::tab_performance_index_frame6,
// InterfaceSkin::ImageIds::tab_performance_index_frame7,
// uint32_t imageId = skin->img;
// if (self.currentTab == widx::tab_performance - widx::tab_company_list)
// imageId += performanceImageIds[(self.frameNo / 4) % std::size(performanceImageIds)];
// else
// imageId += performanceImageIds[0];
// imageId = Gfx::recolour(imageId, self.getColour(WindowColour::secondary).c());
// Widget::drawTab(self, drawingCtx, imageId, widx::tab_performance);
// // Cargo Unit Tab
// static constexpr uint32_t cargoUnitsImageIds[] = {
// InterfaceSkin::ImageIds::tab_cargo_units_frame0,
// InterfaceSkin::ImageIds::tab_cargo_units_frame1,
// InterfaceSkin::ImageIds::tab_cargo_units_frame2,
// InterfaceSkin::ImageIds::tab_cargo_units_frame3,
// InterfaceSkin::ImageIds::tab_cargo_units_frame4,
// InterfaceSkin::ImageIds::tab_cargo_units_frame5,
// InterfaceSkin::ImageIds::tab_cargo_units_frame6,
// InterfaceSkin::ImageIds::tab_cargo_units_frame7,
// uint32_t imageId = skin->img;
// if (self.currentTab == widx::tab_cargo_units - widx::tab_company_list)
// imageId += cargoUnitsImageIds[(self.frameNo / 4) % std::size(cargoUnitsImageIds)];
// else
// imageId += cargoUnitsImageIds[0];
// imageId = Gfx::recolour(imageId, self.getColour(WindowColour::secondary).c());
// Widget::drawTab(self, drawingCtx, imageId, widx::tab_cargo_units);
// // Cargo Distance Tab
// static constexpr uint32_t cargoDistanceImageIds[] = {
// InterfaceSkin::ImageIds::tab_cargo_distance_frame0,
// InterfaceSkin::ImageIds::tab_cargo_distance_frame1,
// InterfaceSkin::ImageIds::tab_cargo_distance_frame2,
// InterfaceSkin::ImageIds::tab_cargo_distance_frame3,
// InterfaceSkin::ImageIds::tab_cargo_distance_frame4,
// InterfaceSkin::ImageIds::tab_cargo_distance_frame5,
// InterfaceSkin::ImageIds::tab_cargo_distance_frame6,
// InterfaceSkin::ImageIds::tab_cargo_distance_frame7,
// uint32_t imageId = skin->img;
// if (self.currentTab == widx::tab_cargo_distance - widx::tab_company_list)
// imageId += cargoDistanceImageIds[(self.frameNo / 4) % std::size(cargoDistanceImageIds)];
// else
// imageId += cargoDistanceImageIds[0];
// imageId = Gfx::recolour(imageId, self.getColour(WindowColour::secondary).c());
// Widget::drawTab(self, drawingCtx, imageId, widx::tab_cargo_distance);
// // Company Values Tab
// static constexpr uint32_t companyValuesImageIds[] = {
// InterfaceSkin::ImageIds::tab_production_frame0,
// InterfaceSkin::ImageIds::tab_production_frame1,
// InterfaceSkin::ImageIds::tab_production_frame2,
// InterfaceSkin::ImageIds::tab_production_frame3,
// InterfaceSkin::ImageIds::tab_production_frame4,
// InterfaceSkin::ImageIds::tab_production_frame5,
// InterfaceSkin::ImageIds::tab_production_frame6,
// InterfaceSkin::ImageIds::tab_production_frame7,
// uint32_t imageId = skin->img;
// if (self.currentTab == widx::tab_values - widx::tab_company_list)
// imageId += companyValuesImageIds[(self.frameNo / 4) % std::size(companyValuesImageIds)];
// else
// imageId += companyValuesImageIds[0];
// imageId = Gfx::recolour(imageId, self.getColour(WindowColour::secondary).c());
// Widget::drawTab(self, drawingCtx, imageId, widx::tab_values);
// if (!(self.isDisabled(widx::tab_values)))
// auto& widget = self.widgets[widx::tab_values];
// auto point = Point(widget.left + self.x + 28, widget.top + self.y + 14 + 1);
// tr.drawStringRight(point, Colour::black, StringIds::currency_symbol);
// // Payment Rates Tab
// uint32_t imageId = skin->img;
// imageId += InterfaceSkin::ImageIds::tab_cargo_payment_rates;
// Widget::drawTab(self, drawingCtx, imageId, widx::tab_payment_rates);
// if (!(self.isDisabled(widx::tab_payment_rates)))
// auto& widget = self.widgets[widx::tab_payment_rates];
// auto point = Point(widget.left + self.x + 28, widget.top + self.y + 14 + 1);
// tr.drawStringRight(point, Colour::black, StringIds::currency_symbol);
// // Speed Records Tab
// uint32_t imageId = skin->img;
// imageId += InterfaceSkin::ImageIds::tab_awards;
// imageId = Gfx::recolour(imageId, self.getColour(WindowColour::secondary).c());
// Widget::drawTab(self, drawingCtx, imageId, widx::tab_speed_records);
// // 0x00437AB6
// func RefreshCompanyList(self Window) 
// self.rowCount = 0;
// for (auto& company : CompanyManager::companies())
// company.challengeFlags &= ~CompanyFlags::sorted;
// // 0x00437810
// func DrawGraphLegend(self Window, drawingCtx Gfx::DrawingContext, x int16, y int16) 
// auto tr = Gfx::TextRenderer(drawingCtx);
// auto companyCount = 0;
// for (auto& company : CompanyManager::companies())
// auto companyColour = CompanyManager::getCompanyColour(company.id());
// auto colour = Colours::getShade(companyColour, 6);
// auto stringId = StringIds::small_black_string;
// if (self.var_854 & (1 << companyCount))
// stringId = StringIds::small_white_string;
// if (!(self.var_854 & (1 << companyCount)) || !(_hoverItemTicks & (1 << 2)))
// drawingCtx.fillRect(x, y + 3, x + 4, y + 7, colour, Gfx::RectFlags::none);
// orphan member: FormatArguments args{};
// args.push(company.name);
// auto point = Point(x + 6, y);
// tr.drawStringLeftClipped(point, 94, Colour::black, stringId, args);
// y += 10;
// companyCount++;
// // 0x004365E4
// func DrawGraphAndLegend(self Window, drawingCtx Gfx::DrawingContext) 
// auto totalMonths = (getCurrentYear() * 12) + static_cast<uint16_t>(getCurrentMonth());
// _graphSettings.xAxisRange = totalMonths;
// _graphSettings.xAxisStepSize = 1;
// _graphSettings.pointFlags = GraphPointFlags::drawLines;
// Ui::drawGraph(_graphSettings, self, drawingCtx);
// if (self.var_854 != 0)
// auto i = 0;
// auto bitScan = Numerics::bitScanForward(self.var_854);
// while (bitScan != _graphSettings.itemId[i] && bitScan != -1)
// i++;
// // Exclude all except highlighted data
// _graphSettings.linesToExclude = 0xFFFFFFFF & ~(1 << i);
// if (_hoverItemTicks & (1 << 2))
// _graphSettings.lineColour[i] = 10;
// _graphSettings.flags |= GraphFlags::hideAxesAndLabels;
// Ui::drawGraph(_graphSettings, self, drawingCtx);
// auto x = self.width + self.x - kLegendWidth - kWindowPadding;
// auto y = self.y + 52;
// Common::drawGraphLegend(self, drawingCtx, x, y);
