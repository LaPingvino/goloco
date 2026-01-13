package ui

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Audio/Audio.h"
// #include "Game.h"
// #include "GameState.h"
// #include "Graphics/Colour.h"
// #include "Graphics/ImageIds.h"
// #include "Graphics/RenderTarget.h"
// #include "Graphics/TextRenderer.h"
// #include "Input.h"
// #include "Localisation/Conversion.h"
// #include "Localisation/FormatArguments.hpp"
// #include "Localisation/StringIds.h"
// #include "Map/MapGenerator/MapGenerator.h"
// #include "Objects/HillShapesObject.h"
// #include "Objects/InterfaceSkinObject.h"
// #include "Objects/LandObject.h"
// #include "Objects/ObjectManager.h"
// #include "Objects/WaterObject.h"
// #include "Scenario/Scenario.h"
// #include "Scenario/ScenarioOptions.h"
// #include "Ui/Dropdown.h"
// #include "Ui/ToolManager.h"
// #include "Ui/Widget.h"
// #include "Ui/Widgets/ButtonWidget.h"
// #include "Ui/Widgets/CaptionWidget.h"
// #include "Ui/Widgets/CheckboxWidget.h"
// #include "Ui/Widgets/DropdownWidget.h"
// #include "Ui/Widgets/FrameWidget.h"
// #include "Ui/Widgets/GroupBoxWidget.h"
// #include "Ui/Widgets/ImageButtonWidget.h"
// #include "Ui/Widgets/LabelWidget.h"
// #include "Ui/Widgets/PanelWidget.h"
// #include "Ui/Widgets/ScrollViewWidget.h"
// #include "Ui/Widgets/StepperWidget.h"
// #include "Ui/Widgets/TabWidget.h"
// #include "Ui/WindowManager.h"
// #include "World/IndustryManager.h"
// #include "World/TownManager.h"
// #include <OpenLoco/Diagnostics/Logging.h>
// using namespace OpenLoco::Diagnostics;
// namespace OpenLoco::Ui::Windows::LandscapeGeneration
// static constexpr Ui::Size kWindowSize = { 366, 217 };
// static constexpr Ui::Size kLandTabSize = { 366, 252 };
const RowHeight uint8 = 22
const MaxLandObjects int = ObjectManager.getMaxObjects(ObjectType.land)
// namespace Common
type Widx int

const (
	Frame Widx = 0
	Caption Widx = 1
	Close_button Widx = 2
	Panel Widx = 3
	Tab_options
	Tab_land
	Tab_water
	Tab_forests
	Tab_towns
	Tab_industries
	Generate_now
)
type ResetLandscapeMode int

const (
	Generate_now ResetLandscapeMode = 0
	Use_random_landscape ResetLandscapeMode = 1
)
// func MakeCommonWidgets(frame_height int32, window_caption_id StringId) any
// return makeWidgets(
// Widgets::Frame({ 0, 0 }, { 366, frame_height }, WindowColour::primary),
// Widgets::Caption({ 1, 1 }, { 364, 13 }, Widgets::Caption::Style::whiteText, WindowColour::primary, window_caption_id),
// Widgets::ImageButton({ 351, 2 }, { 13, 13 }, WindowColour::primary, ImageIds::close_button, StringIds::tooltip_close_window),
// Widgets::Panel({ 0, 41 }, { 366, 175 }, WindowColour::secondary),
// Widgets::Tab({ 3, 15 }, { 31, 27 }, WindowColour::secondary, ImageIds::tab, StringIds::tooltip_landscape_generation_options),
// Widgets::Tab({ 34, 15 }, { 31, 27 }, WindowColour::secondary, ImageIds::tab, StringIds::tooltip_landscape_generation_land),
// Widgets::Tab({ 65, 15 }, { 31, 27 }, WindowColour::secondary, ImageIds::tab, StringIds::tooltip_landscape_generation_water),
// Widgets::Tab({ 96, 15 }, { 31, 27 }, WindowColour::secondary, ImageIds::tab, StringIds::tooltip_landscape_generation_forests),
// Widgets::Tab({ 127, 15 }, { 31, 27 }, WindowColour::secondary, ImageIds::tab, StringIds::tooltip_landscape_generation_towns),
// Widgets::Tab({ 158, 15 }, { 31, 27 }, WindowColour::secondary, ImageIds::tab, StringIds::tooltip_landscape_generation_industries),
// Widgets::Button({ 196, frame_height - 17 }, { 160, 12 }, WindowColour::secondary, StringIds::button_generate_landscape, StringIds::tooltip_generate_random_landscape));
// // Defined at the bottom of this file.
// func SwitchTabWidgets(window Window) 
// func SwitchTab(window Window, widgetIndex WidgetIndex_t) 
// func ConfirmResetLandscape(promptType ResetLandscapeMode) 
// if (Scenario::getOptions().madeAnyChanges)
// // 'Are you sure?' confirmation prompt
// orphan member: StringId titleId;
// orphan member: FormatArguments args{};
// if (promptType == ResetLandscapeMode::generate_now)
// titleId = StringIds::title_generate_new_landscape;
// args.push(StringIds::prompt_confirm_generate_landscape);
// else
// titleId = StringIds::title_random_landscape_option;
// args.push(StringIds::prompt_confirm_random_landscape);
// if (!Windows::PromptOkCancel::open(titleId, StringIds::stringid, args, StringIds::label_ok))
// return;
// // Reset the landscape
// if (promptType == ResetLandscapeMode::generate_now)
// Scenario::generateLandscape();
// else
// Scenario::eraseLandscape();
// func OnMouseUp(self Window, widgetIndex WidgetIndex_t, id [[maybe_unused]] WidgetId) 
// switch (widgetIndex)
// case widx::close_button:
// WindowManager::close(&self);
// break;
// case widx::tab_options:
// case widx::tab_land:
// case widx::tab_water:
// case widx::tab_forests:
// case widx::tab_towns:
// case widx::tab_industries:
// switchTab(self, widgetIndex);
// break;
// case widx::generate_now:
// confirmResetLandscape(ResetLandscapeMode::generate_now);
// break;
// // 0x0043ECA4
// func DrawTabs(self Window, drawingCtx Gfx::DrawingContext) 
// auto skin = ObjectManager::get<InterfaceSkinObject>();
// // Options tab
// static constexpr uint32_t optionTabImageIds[] = {
// InterfaceSkin::ImageIds::tab_cogs_frame0,
// InterfaceSkin::ImageIds::tab_cogs_frame1,
// InterfaceSkin::ImageIds::tab_cogs_frame2,
// InterfaceSkin::ImageIds::tab_cogs_frame3,
// uint32_t imageId = skin->img;
// if (self.currentTab == widx::tab_options - widx::tab_options)
// imageId += optionTabImageIds[(self.frameNo / 2) % std::size(optionTabImageIds)];
// else
// imageId += optionTabImageIds[0];
// Widget::drawTab(self, drawingCtx, imageId, widx::tab_options);
// // Land tab
// auto land = ObjectManager::get<LandObject>(getGameState().lastLandOption);
// const uint32_t imageId = land->mapPixelImage + OpenLoco::Land::ImageIds::toolbar_terraform_land;
// Widget::drawTab(self, drawingCtx, imageId, widx::tab_land);
// // Water tab
// const auto waterObj = ObjectManager::get<WaterObject>();
// uint32_t imageId = waterObj->image + OpenLoco::Water::ImageIds::kToolbarTerraformWater;
// if (self.currentTab == widx::tab_water - widx::tab_options)
// imageId += (self.frameNo / 2) % 16;
// Widget::drawTab(self, drawingCtx, imageId, widx::tab_water);
// // Forest tab
// const uint32_t imageId = skin->img + InterfaceSkin::ImageIds::toolbar_menu_plant_trees;
// Widget::drawTab(self, drawingCtx, imageId, widx::tab_forests);
// // Towns tab
// const uint32_t imageId = skin->img + InterfaceSkin::ImageIds::toolbar_menu_towns;
// Widget::drawTab(self, drawingCtx, imageId, widx::tab_towns);
// // Industries tab
// const uint32_t imageId = skin->img + InterfaceSkin::ImageIds::toolbar_menu_industries;
// Widget::drawTab(self, drawingCtx, imageId, widx::tab_industries);
// func Draw(window Window, drawingCtx Gfx::DrawingContext) 
// window.draw(drawingCtx);
// drawTabs(window, drawingCtx);
// func PrepareDraw(window Window) 
// window.widgets[widx::frame].right = window.width - 1;
// window.widgets[widx::frame].bottom = window.height - 1;
// window.widgets[widx::panel].right = window.width - 1;
// window.widgets[widx::panel].bottom = window.height - 1;
// window.widgets[widx::caption].right = window.width - 2;
// window.widgets[widx::close_button].left = window.width - 15;
// window.widgets[widx::close_button].right = window.width - 3;
// auto& options = Scenario::getOptions();
// if (options.generator == Scenario::LandGeneratorType::PngHeightMap)
// if (World::MapGenerator::getPngHeightmapPath().empty())
// window.disabledWidgets |= (1 << widx::generate_now);
// else
// window.disabledWidgets &= ~(1 << widx::generate_now);
// func If(Scenario::ScenarioFlags::landscapeGenerationDone (options.scenarioFlags) else
// window.disabledWidgets |= (1 << widx::generate_now);
// else
// window.disabledWidgets &= ~(1 << widx::generate_now);
// func Update(window Window) 
// window.frameNo++;
// window.callPrepareDraw();
// WindowManager::invalidateWidget(WindowType::landscapeGeneration, window.number, window.currentTab + widx::tab_options);
// namespace Options
type Widx int

const (
	GroupGeneral Widx = Common.widx.generate_now + 1
	Start_year_label
	Start_year
	Start_year_down
	Start_year_up
	HeightMapBoxLabel
	HeightMapBox
	HeightMapDropdown
	GroupGenerator
	HillObjectLabel
	Change_heightmap_btn
	TerrainSmoothingLabel
	TerrainSmoothingNum
	TerrainSmoothingNumDown
	TerrainSmoothingNumUp
	Generate_when_game_starts
	HeightmapFileLabel
	BrowseHeightmapFile
)
// // clang-format off
// const uint64_t holdable_widgets =
// (1 << widx::start_year_up) |
// (1 << widx::start_year_down) |
// (1 << widx::terrainSmoothingNumUp) |
// (1 << widx::terrainSmoothingNumDown);
// // clang-format on
// static constexpr auto widgets = makeWidgets(
// Common::makeCommonWidgets(217, StringIds::title_landscape_generation_options),
// // General options
// Widgets::GroupBox({ 4, 50 }, { 358, 50 }, WindowColour::secondary, StringIds::landscapeOptionsGroupGeneral),
// Widgets::Label({ 10, 65 }, { 240, 12 }, WindowColour::secondary, ContentAlign::left, StringIds::start_year),
// Widgets::stepperWidgets({ 256, 65 }, { 100, 12 }, WindowColour::secondary, StringIds::start_year_value),
// Widgets::Label({ 10, 81 }, { 160, 12 }, WindowColour::secondary, ContentAlign::left, StringIds::height_map_source),
// Widgets::dropdownWidgets({ 176, 81 }, { 180, 12 }, WindowColour::secondary),
// // Generator options
// Widgets::GroupBox({ 4, 105 }, { 358, 50 }, WindowColour::secondary, StringIds::landscapeOptionsGroupGenerator),
// Widgets::Label({ 10, 120 }, { 260, 12 }, WindowColour::secondary, ContentAlign::left, StringIds::landscapeOptionsCurrentHillObject),
// Widgets::Button({ 280, 120 }, { 75, 12 }, WindowColour::secondary, StringIds::change),
// Widgets::Label({ 10, 120 }, { 260, 12 }, WindowColour::secondary, ContentAlign::left, StringIds::landscapeOptionsSmoothingPasses),
// Widgets::stepperWidgets({ 256, 120 }, { 100, 12 }, WindowColour::secondary, StringIds::uint16_raw),
// Widgets::Checkbox({ 10, 136 }, { 346, 12 }, WindowColour::secondary, StringIds::label_generate_random_landscape_when_game_starts, StringIds::tooltip_generate_random_landscape_when_game_starts),
// // PNG browser
// Widgets::Label({ 10, 120 }, { 260, 12 }, WindowColour::secondary, ContentAlign::left, StringIds::currentHeightmapFile),
// Widgets::Button({ 280, 120 }, { 75, 12 }, WindowColour::secondary, StringIds::button_browse)
// );
// static constexpr StringId generatorIds[] = {
// StringIds::generator_original,
// StringIds::generator_simplex,
// StringIds::generator_png_heightmap,
// // TODO: static memory for state is a bit hacky
// static std::string _pngFilename{};
// // 0x0043DB76
// func PrepareDraw(self Window) 
// Common::prepareDraw(self);
// auto& options = Scenario::getOptions();
// // Start year info
// auto args = FormatArguments(self.widgets[widx::start_year].textArgs);
// args.push<uint16_t>(options.scenarioStartYear);
// self.widgets[widx::heightMapBox].text = generatorIds[enumValue(options.generator)];
// bool isOriginal = options.generator == Scenario::LandGeneratorType::Original;
// bool isSimplex = options.generator == Scenario::LandGeneratorType::Simplex;
// bool isPngFile = options.generator == Scenario::LandGeneratorType::PngHeightMap;
// // Hide widgets depending on active generator
// self.widgets[widx::hillObjectLabel].hidden = !isOriginal;
// self.widgets[widx::change_heightmap_btn].hidden = !isOriginal;
// self.widgets[widx::terrainSmoothingLabel].hidden = !isSimplex;
// self.widgets[widx::terrainSmoothingNum].hidden = !isSimplex;
// self.widgets[widx::terrainSmoothingNumUp].hidden = !isSimplex;
// self.widgets[widx::terrainSmoothingNumDown].hidden = !isSimplex;
// self.widgets[widx::heightmapFileLabel].hidden = !isPngFile;
// self.widgets[widx::browseHeightmapFile].hidden = !isPngFile;
// if (isOriginal)
// // Prepare object name
// auto& widget = self.widgets[widx::hillObjectLabel];
// orphan member: FormatArguments args{ widget.textArgs };
// auto* obj = ObjectManager::get<HillShapesObject>();
// args.push(obj->name);
// self.disabledWidgets &= ~(1 << widx::change_heightmap_btn);
// self.disabledWidgets |= ((1 << widx::terrainSmoothingNum) | (1 << widx::terrainSmoothingNumUp) | (1 << widx::terrainSmoothingNumDown));
// self.disabledWidgets |= (1 << widx::browseHeightmapFile);
// func If(isSimplex) else
// // Prepare value
// auto& widget = self.widgets[widx::terrainSmoothingNum];
// orphan member: FormatArguments args{ widget.textArgs };
// args.push<uint16_t>(options.numTerrainSmoothingPasses);
// self.disabledWidgets |= (1 << widx::change_heightmap_btn);
// self.disabledWidgets &= ~((1 << widx::terrainSmoothingNum) | (1 << widx::terrainSmoothingNumUp) | (1 << widx::terrainSmoothingNumDown));
// self.disabledWidgets |= (1 << widx::browseHeightmapFile);
// func If(isPngFile) else
// // Prepare filename label
// auto path = World::MapGenerator::getPngHeightmapPath();
// auto filename = path.filename().make_preferred().u8string();
// auto& widget = self.widgets[widx::heightmapFileLabel];
// orphan member: FormatArguments args{ widget.textArgs };
// if (!filename.empty())
// _pngFilename = Localisation::convertUnicodeToLoco(filename);
// args.push(_pngFilename.c_str());
// else
// args.push(StringManager::getString(StringIds::noneSelected));
// self.disabledWidgets |= (1 << widx::change_heightmap_btn);
// self.disabledWidgets |= ((1 << widx::terrainSmoothingNum) | (1 << widx::terrainSmoothingNumUp) | (1 << widx::terrainSmoothingNumDown));
// self.disabledWidgets &= ~(1 << widx::browseHeightmapFile);
// self.activatedWidgets &= ~(1 << widx::generate_when_game_starts);
// self.disabledWidgets |= (1 << widx::generate_when_game_starts);
// // Enable/disable the 'generate when game starts' checkbox
// if ((options.scenarioFlags & Scenario::ScenarioFlags::landscapeGenerationDone) == Scenario::ScenarioFlags::none)
// self.activatedWidgets |= (1 << widx::generate_when_game_starts);
// self.disabledWidgets &= ~(1 << widx::generate_when_game_starts);
// else
// self.activatedWidgets &= ~(1 << widx::generate_when_game_starts);
// // 0x0043E1BA
// func OnDropdown(window Window, widgetIndex WidgetIndex_t, id [[maybe_unused]] WidgetId, itemIndex int16) 
// switch (widgetIndex)
// case widx::heightMapDropdown:
// if (itemIndex != -1)
// Scenario::getOptions().generator = static_cast<Scenario::LandGeneratorType>(itemIndex);
// window.invalidate();
// break;
// // 0x0043DC83
// func OnMouseDown(window Window, widgetIndex WidgetIndex_t, id [[maybe_unused]] WidgetId) 
// auto& options = Scenario::getOptions();
// switch (widgetIndex)
// case widx::start_year_up:
// if (options.scenarioStartYear + 1 <= Scenario::kMaxYear)
// options.scenarioStartYear += 1;
// window.invalidate();
// break;
// case widx::start_year_down:
// if (options.scenarioStartYear - 1 >= Scenario::kMinYear)
// options.scenarioStartYear -= 1;
// window.invalidate();
// break;
// case widx::terrainSmoothingNumUp:
// options.numTerrainSmoothingPasses = std::clamp(options.numTerrainSmoothingPasses + 1, 1, 5);
// window.invalidate();
// break;
// case widx::terrainSmoothingNumDown:
// options.numTerrainSmoothingPasses = std::clamp(options.numTerrainSmoothingPasses - 1, 1, 5);
// window.invalidate();
// break;
// case widx::heightMapDropdown:
// Widget& target = window.widgets[widx::heightMapBox];
// Dropdown::show(window.x + target.left, window.y + target.top, target.width() - 4, target.height(), window.getColour(WindowColour::secondary), std::size(generatorIds), 0x80);
// for (size_t i = 0; i < std::size(generatorIds); i++)
// Dropdown::add(i, generatorIds[i]);
// Dropdown::setHighlightedItem(static_cast<uint8_t>(options.generator));
// break;
// // 0x0043DC58
// func OnMouseUp(window Window, widgetIndex WidgetIndex_t, id [[maybe_unused]] WidgetId) 
// switch (widgetIndex)
// case widx::generate_when_game_starts:
// if ((Scenario::getOptions().scenarioFlags & Scenario::ScenarioFlags::landscapeGenerationDone) == Scenario::ScenarioFlags::none)
// Scenario::getOptions().scenarioFlags |= Scenario::ScenarioFlags::landscapeGenerationDone;
// Scenario::generateLandscape();
// else
// WindowManager::closeConstructionWindows();
// Common::confirmResetLandscape(Common::ResetLandscapeMode::use_random_landscape);
// break;
// case widx::change_heightmap_btn:
// EditorController::goToPreviousStep();
// ObjectSelectionWindow::openInTab(ObjectType::hillShapes);
// break;
// case widx::browseHeightmapFile:
// if (auto res = Game::loadHeightmapOpen())
// World::MapGenerator::setPngHeightmapPath(fs::u8path(*res));
// window.invalidate();
// break;
// default:
// Common::onMouseUp(window, widgetIndex, id);
// static constexpr WindowEventList kEvents = {
// .onMouseUp = onMouseUp,
// .onMouseDown = onMouseDown,
// .onDropdown = onDropdown,
// .onUpdate = Common::update,
// .prepareDraw = prepareDraw,
// .draw = Common::draw,
// static const WindowEventList& getEvents()
// orphan member: return kEvents;
// // 0x0043DA43
// Window* open()
// auto window = WindowManager::bringToFront(WindowType::landscapeGeneration, 0);
// if (window != nullptr)
// if (ToolManager::isToolActive(window->type, window->number))
// ToolManager::toolCancel();
// window = WindowManager::bringToFront(WindowType::landscapeGeneration, 0);
// // Start of 0x0043DAEA
// if (window == nullptr)
// window = WindowManager::createWindowCentred(WindowType::landscapeGeneration, kWindowSize, WindowFlags::none, Options::getEvents());
// window->setWidgets(Options::widgets);
// window->number = 0;
// window->currentTab = 0;
// window->frameNo = 0;
// window->rowHover = -1;
// auto interface = ObjectManager::get<InterfaceSkinObject>();
// window->setColour(WindowColour::primary, interface->windowTitlebarColour);
// window->setColour(WindowColour::secondary, interface->windowTerraFormColour);
// // End of 0x0043DAEA
// window->width = kWindowSize.width;
// window->height = kWindowSize.height;
// window->invalidate();
// window->activatedWidgets = 0;
// window->holdableWidgets = Options::holdable_widgets;
// window->callOnResize();
// window->callPrepareDraw();
// window->initScrollWidgets();
// orphan member: return window;
// namespace Land
type Widx int

const (
	Topography_style_label Widx = Common.widx.generate_now + 1
	Topography_style
	Topography_style_btn
	Min_land_height_label
	Min_land_height
	Min_land_height_down
	Min_land_height_up
	Hill_density_label
	Hill_density
	Hill_density_down
	Hill_density_up
	HillsEdgeOfMap
	Scrollview
)
// const uint64_t holdable_widgets = (1 << widx::min_land_height_up) | (1 << widx::min_land_height_down) | (1 << widx::hill_density_up) | (1 << widx::hill_density_down);
// static constexpr auto widgets = makeWidgets(
// Common::makeCommonWidgets(252, StringIds::title_landscape_generation_land),
// Widgets::Label({ 10, 52 }, { 160, 12 }, WindowColour::secondary, ContentAlign::left, StringIds::topography_style),
// Widgets::dropdownWidgets({ 176, 52 }, { 180, 12 }, WindowColour::secondary),
// Widgets::Label({ 10, 68 }, { 250, 12 }, WindowColour::secondary, ContentAlign::left, StringIds::min_land_height),
// Widgets::stepperWidgets({ 256, 68 }, { 100, 12 }, WindowColour::secondary, StringIds::min_land_height_units),
// Widgets::Label({ 10, 84 }, { 250, 12 }, WindowColour::secondary, ContentAlign::left, StringIds::hill_density),
// Widgets::stepperWidgets({ 256, 84 }, { 100, 12 }, WindowColour::secondary, StringIds::hill_density_percent),
// Widgets::Checkbox({ 10, 100 }, { 346, 12 }, WindowColour::secondary, StringIds::create_hills_right_up_to_edge_of_map),
// Widgets::ScrollView({ 4, 116 }, { 358, 112 }, WindowColour::secondary, Scrollbars::vertical)
// );
// static constexpr StringId landDistributionLabelIds[] = {
// StringIds::land_distribution_everywhere,
// StringIds::land_distribution_nowhere,
// StringIds::land_distribution_far_from_water,
// StringIds::land_distribution_near_water,
// StringIds::land_distribution_on_mountains,
// StringIds::land_distribution_far_from_mountains,
// StringIds::land_distribution_in_small_random_areas,
// StringIds::land_distribution_in_large_random_areas,
// StringIds::land_distribution_around_cliffs,
var LandDropdownWidth = 190 // auto
var LandDropdownLeft = 150 // auto
var LandDropdownRight = kLandDropdownLeft + kLandDropdownWidth // auto
// // 0x0043E01C
// func DrawScroll(window Ui::Window, drawingCtx Gfx::DrawingContext, scrollIndex [[maybe_unused]] uint32_t) 
// const auto& rt = drawingCtx.currentRenderTarget();
// auto tr = Gfx::TextRenderer(drawingCtx);
// uint16_t yPos = 0;
// for (uint16_t i = 0; i < kMaxLandObjects; i++)
// if (yPos + kRowHeight < rt.y)
// yPos += kRowHeight;
// continue;
// func If(rt.height yPos > rt.y +) else
// break;
// auto landObject = ObjectManager::get<LandObject>(i);
// if (landObject == nullptr)
// continue;
// // Draw tile icon.
// const uint32_t imageId = landObject->mapPixelImage + OpenLoco::Land::ImageIds::landscape_generator_tile_icon;
// drawingCtx.drawImage(2, yPos + 1, imageId);
// // Draw land description.
// orphan member: FormatArguments args{};
// args.push(landObject->name);
// auto point = Point(24, yPos + 5);
// tr.drawStringLeftClipped(point, 121, Colour::black, StringIds::wcolour2_stringid, args);
// // Draw rectangle.
// drawingCtx.fillRectInset(kLandDropdownLeft, yPos + 5, kLandDropdownRight, yPos + 16, window.getColour(WindowColour::secondary), Gfx::RectInsetFlags::borderInset | Gfx::RectInsetFlags::fillDarker);
// // Draw current distribution setting.
// orphan member: FormatArguments args{};
// const StringId distributionId = landDistributionLabelIds[enumValue(Scenario::getOptions().landDistributionPatterns[i])];
// args.push(distributionId);
// auto point = Point(kLandDropdownLeft + 1, yPos + 5);
// tr.drawStringLeftClipped(point, kLandDropdownWidth - 3, Colour::black, StringIds::black_stringid, args);
// // Draw rectangle (knob).
// const Gfx::RectInsetFlags flags = window.rowHover == i ? Gfx::RectInsetFlags::borderInset | Gfx::RectInsetFlags::fillDarker : Gfx::RectInsetFlags::none;
// drawingCtx.fillRectInset(kLandDropdownRight - 11, yPos + 6, kLandDropdownRight - 1, yPos + 15, window.getColour(WindowColour::secondary), flags);
// // Draw triangle (knob).
// auto point = Point(kLandDropdownRight - 10, yPos + 6);
// tr.drawStringLeft(point, Colour::black, StringIds::dropdown);
// yPos += kRowHeight;
// // 0x0043E2AC
// func GetScrollSize(window [[maybe_unused]] Ui::Window, scrollIndex [[maybe_unused]] uint32_t, scrollWidth [[maybe_unused]] int32_t, scrollHeight int32) 
// scrollHeight = 0;
// for (uint16_t i = 0; i < kMaxLandObjects; i++)
// auto landObject = ObjectManager::get<LandObject>(i);
// if (landObject == nullptr)
// continue;
// scrollHeight += kRowHeight;
// static constexpr StringId topographyStyleIds[] = {
// StringIds::flat_land,
// StringIds::small_hills,
// StringIds::mountains,
// StringIds::half_mountains_half_hills,
// StringIds::half_mountains_half_flat,
// // 0x0043E1BA
// func OnDropdown(window Window, widgetIndex WidgetIndex_t, id [[maybe_unused]] WidgetId, itemIndex int16) 
// switch (widgetIndex)
// case widx::topography_style_btn:
// if (itemIndex != -1)
// Scenario::getOptions().topographyStyle = static_cast<Scenario::TopographyStyle>(itemIndex);
// window.invalidate();
// break;
// case widx::scrollview:
// if (itemIndex != -1 && window.rowHover != -1)
// Scenario::getOptions().landDistributionPatterns[window.rowHover] = static_cast<Scenario::LandDistributionPattern>(itemIndex);
// window.invalidate();
// break;
// // 0x0043E173
// func OnMouseDown(window Window, widgetIndex WidgetIndex_t, id [[maybe_unused]] WidgetId) 
// auto& options = Scenario::getOptions();
// switch (widgetIndex)
// case widx::min_land_height_up:
// options.minLandHeight = std::min<int8_t>(options.minLandHeight + 1, Scenario::kMaxBaseLandHeight);
// break;
// case widx::min_land_height_down:
// options.minLandHeight = std::max<int8_t>(Scenario::kMinBaseLandHeight, options.minLandHeight - 1);
// break;
// case widx::topography_style_btn:
// Widget& target = window.widgets[widx::topography_style];
// Dropdown::show(window.x + target.left, window.y + target.top, target.width() - 4, target.height(), window.getColour(WindowColour::secondary), std::size(topographyStyleIds), 0x80);
// for (size_t i = 0; i < std::size(topographyStyleIds); i++)
// Dropdown::add(i, topographyStyleIds[i]);
// Dropdown::setHighlightedItem(static_cast<uint8_t>(options.topographyStyle));
// break;
// case widx::hill_density_up:
// options.hillDensity = std::min<int8_t>(options.hillDensity + 1, Scenario::kMaxHillDensity);
// break;
// case widx::hill_density_down:
// options.hillDensity = std::max<int8_t>(Scenario::kMinHillDensity, options.hillDensity - 1);
// break;
// default:
// // Nothing was changed, don't invalidate.
// return;
// // After changing any of the options, invalidate the window.
// window.invalidate();
// // 0x0043E14E
// func OnMouseUp(window Window, widgetIndex WidgetIndex_t, id [[maybe_unused]] WidgetId) 
// switch (widgetIndex)
// case widx::hillsEdgeOfMap:
// Scenario::getOptions().scenarioFlags ^= Scenario::ScenarioFlags::hillsEdgeOfMap;
// window.invalidate();
// break;
// default:
// Common::onMouseUp(window, widgetIndex, id);
// // 0x0043E421
// func ScrollPosToLandIndex(xPos int16, yPos int16) int16
// if (xPos < kLandDropdownLeft || xPos > kLandDropdownRight)
// return -1;
// for (uint16_t i = 0; i < kMaxLandObjects; i++)
// auto landObject = ObjectManager::get<LandObject>(i);
// if (landObject == nullptr)
// continue;
// yPos -= kRowHeight;
// if (yPos < 0)
// orphan member: return i;
// return -1;
// // 0x0043E1CF
// func ScrollMouseDown(window Window, xPos int16, yPos int16, scrollIndex [[maybe_unused]] uint8_t) 
// int16_t landIndex = scrollPosToLandIndex(xPos, yPos);
// if (landIndex == -1)
// return;
// window.rowHover = landIndex;
// Audio::playSound(Audio::SoundId::clickDown, window.widgets[widx::scrollview].right);
// const Widget& target = window.widgets[widx::scrollview];
// const int16_t dropdownX = window.x + target.left + kLandDropdownLeft + 1;
// const int16_t dropdownY = window.y + target.top + 6 + landIndex * kRowHeight - window.scrollAreas[0].contentOffsetY;
// Dropdown::show(dropdownX, dropdownY, kLandDropdownWidth - 2, 12, window.getColour(WindowColour::secondary), std::size(landDistributionLabelIds), 0x80);
// for (size_t i = 0; i < std::size(landDistributionLabelIds); i++)
// Dropdown::add(i, StringIds::dropdown_stringid, landDistributionLabelIds[i]);
// Dropdown::setItemSelected(enumValue(Scenario::getOptions().landDistributionPatterns[landIndex]));
// // 0x0043DEBF
// func PrepareDraw(window Window) 
// Common::prepareDraw(window);
// auto& options = Scenario::getOptions();
// auto args = FormatArguments(window.widgets[widx::hill_density].textArgs);
// args.push<uint16_t>(options.hillDensity);
// auto args = FormatArguments(window.widgets[widx::min_land_height].textArgs);
// args.push<uint16_t>(options.minLandHeight);
// window.widgets[widx::topography_style].text = topographyStyleIds[static_cast<uint8_t>(options.topographyStyle)];
// if ((options.scenarioFlags & Scenario::ScenarioFlags::hillsEdgeOfMap) != Scenario::ScenarioFlags::none)
// window.activatedWidgets |= (1 << widx::hillsEdgeOfMap);
// else
// window.activatedWidgets &= ~(1 << widx::hillsEdgeOfMap);
// // 0x0043E2A2
// static std::optional<FormatArguments> tooltip([[maybe_unused]] Ui::Window& window, [[maybe_unused]] WidgetIndex_t widgetIndex, [[maybe_unused]] const WidgetId id)
// orphan member: FormatArguments args{};
// args.push(StringIds::tooltip_scroll_list);
// orphan member: return args;
// // 0x0043E3D9
// func Update(window Window) 
// Common::update(window);
// auto dropdown = WindowManager::find(WindowType::dropdown, 0);
// if (dropdown == nullptr && window.rowHover != -1)
// window.rowHover = -1;
// window.invalidate();
// static constexpr WindowEventList kEvents = {
// .onMouseUp = onMouseUp,
// .onMouseDown = onMouseDown,
// .onDropdown = onDropdown,
// .onUpdate = update,
// .getScrollSize = getScrollSize,
// .scrollMouseDown = scrollMouseDown,
// .tooltip = tooltip,
// .prepareDraw = prepareDraw,
// .draw = Common::draw,
// .drawScroll = drawScroll,
// static const WindowEventList& getEvents()
// orphan member: return kEvents;
// namespace Water
type Widx int

const (
	Sea_level_label Widx = Common.widx.generate_now + 1
	Sea_level
	Sea_level_down
	Sea_level_up
	Num_riverbeds_label
	Num_riverbeds
	Num_riverbeds_down
	Num_riverbeds_up
	Min_river_width_label
	Min_river_width
	Min_river_width_down
	Min_river_width_up
	Max_river_width_label
	Max_river_width
	Max_river_width_down
	Max_river_width_up
	Riverbank_width_label
	Riverbank_width
	Riverbank_width_down
	Riverbank_width_up
	Meander_rate_label
	Meander_rate
	Meander_rate_down
	Meander_rate_up
)
// // clang-format off
// const uint64_t holdable_widgets = (1ULL << widx::sea_level_up) | (1ULL << widx::sea_level_down) |
// (1ULL << widx::num_riverbeds_down) | (1ULL << widx::num_riverbeds_up) |
// (1ULL << widx::min_river_width_down) | (1ULL << widx::min_river_width_up) |
// (1ULL << widx::max_river_width_down) | (1ULL << widx::max_river_width_up) |
// (1ULL << widx::riverbank_width_down) | (1ULL << widx::riverbank_width_up) |
// (1ULL << widx::meander_rate_down) | (1ULL << widx::meander_rate_up);
// // clang-format on
// static constexpr auto widgets = makeWidgets(
// Common::makeCommonWidgets(217, StringIds::title_landscape_generation_water),
// Widgets::Label({ 10, 52 }, { 240, 12 }, WindowColour::secondary, ContentAlign::left, StringIds::sea_level),
// Widgets::stepperWidgets({ 256, 52 }, { 100, 12 }, WindowColour::secondary, StringIds::sea_level_units),
// Widgets::Label({ 10, 68 }, { 240, 12 }, WindowColour::secondary, ContentAlign::left, StringIds::number_riverbeds),
// Widgets::stepperWidgets({ 256, 68 }, { 100, 12 }, WindowColour::secondary, StringIds::uint16_raw),
// Widgets::Label({ 10, 84 }, { 240, 12 }, WindowColour::secondary, ContentAlign::left, StringIds::minimum_river_width),
// Widgets::stepperWidgets({ 256, 84 }, { 100, 12 }, WindowColour::secondary, StringIds::min_land_height_units),
// Widgets::Label({ 10, 100 }, { 240, 12 }, WindowColour::secondary, ContentAlign::left, StringIds::maximum_river_width),
// Widgets::stepperWidgets({ 256, 100 }, { 100, 12 }, WindowColour::secondary, StringIds::min_land_height_units),
// Widgets::Label({ 10, 116 }, { 240, 12 }, WindowColour::secondary, ContentAlign::left, StringIds::riverbank_width),
// Widgets::stepperWidgets({ 256, 116 }, { 100, 12 }, WindowColour::secondary, StringIds::min_land_height_units),
// Widgets::Label({ 10, 132 }, { 240, 12 }, WindowColour::secondary, ContentAlign::left, StringIds::meander_rate),
// Widgets::stepperWidgets({ 256, 132 }, { 100, 12 }, WindowColour::secondary, StringIds::min_land_height_units)
// );
// // 0x0043E173
// func OnMouseDown(window Window, widgetIndex WidgetIndex_t, id [[maybe_unused]] WidgetId) 
// auto& gameState = getGameState();
// auto& options = Scenario::getOptions();
// switch (widgetIndex)
// case widx::sea_level_up:
// gameState.seaLevel = std::min<int8_t>(gameState.seaLevel + 1, Scenario::kMaxSeaLevel);
// break;
// case widx::sea_level_down:
// gameState.seaLevel = std::max<int8_t>(Scenario::kMinSeaLevel, gameState.seaLevel - 1);
// break;
// case widx::num_riverbeds_up:
// options.numRiverbeds = std::min<int8_t>(options.numRiverbeds + 1, Scenario::kMaxNumRiverbeds);
// break;
// case widx::num_riverbeds_down:
// options.numRiverbeds = std::max<int8_t>(Scenario::kMinNumRiverbeds, options.numRiverbeds - 1);
// break;
// case widx::min_river_width_up:
// options.minRiverWidth = std::min<int8_t>(options.minRiverWidth + 1, Scenario::kMaxMinRiverWidth);
// break;
// case widx::min_river_width_down:
// options.minRiverWidth = std::max<int8_t>(Scenario::kMinMinRiverWidth, options.minRiverWidth - 1);
// break;
// case widx::max_river_width_up:
// options.maxRiverWidth = std::min<int8_t>(options.maxRiverWidth + 1, Scenario::kMaxMaxRiverWidth);
// break;
// case widx::max_river_width_down:
// options.maxRiverWidth = std::max<int8_t>(Scenario::kMinMaxRiverWidth, options.maxRiverWidth - 1);
// break;
// case widx::riverbank_width_up:
// options.riverbankWidth = std::min<int8_t>(options.riverbankWidth + 1, Scenario::kMaxRiverbankWidth);
// break;
// case widx::riverbank_width_down:
// options.riverbankWidth = std::max<int8_t>(Scenario::kMinRiverbankWidth, options.riverbankWidth - 1);
// break;
// case widx::meander_rate_up:
// options.riverMeanderRate = std::min<int8_t>(options.riverMeanderRate + 1, Scenario::kMaxRiverMeanderRate);
// break;
// case widx::meander_rate_down:
// options.riverMeanderRate = std::max<int8_t>(Scenario::kMinRiverMeanderRate, options.riverMeanderRate - 1);
// break;
// default:
// // Nothing was changed, don't invalidate.
// return;
// // After changing any of the options, invalidate the window.
// window.invalidate();
// // 0x0043DEBF
// func PrepareDraw(window Window) 
// Common::prepareDraw(window);
// auto& gameState = getGameState();
// auto& options = Scenario::getOptions();
// auto args = FormatArguments(window.widgets[widx::sea_level].textArgs);
// args.push(gameState.seaLevel);
// auto args = FormatArguments(window.widgets[widx::num_riverbeds].textArgs);
// args.push<uint16_t>(options.numRiverbeds);
// auto args = FormatArguments(window.widgets[widx::min_river_width].textArgs);
// args.push<uint16_t>(options.minRiverWidth);
// auto args = FormatArguments(window.widgets[widx::max_river_width].textArgs);
// args.push<uint16_t>(options.maxRiverWidth);
// auto args = FormatArguments(window.widgets[widx::riverbank_width].textArgs);
// args.push<uint16_t>(options.riverbankWidth);
// auto args = FormatArguments(window.widgets[widx::meander_rate].textArgs);
// args.push<uint16_t>(options.riverMeanderRate);
// static constexpr WindowEventList kEvents = {
// .onMouseUp = Common::onMouseUp,
// .onMouseDown = onMouseDown,
// .onUpdate = Common::update,
// .prepareDraw = prepareDraw,
// .draw = Common::draw,
// static const WindowEventList& getEvents()
// orphan member: return kEvents;
// namespace Forests
type Widx int

const (
	Number_of_forests_label Widx = Common.widx.generate_now + 1
	Number_of_forests
	Number_of_forests_down
	Number_of_forests_up
	MinForestRadiusLabel
	MinForestRadius
	Min_forest_radius_down
	Min_forest_radius_up
	MaxForestRadiusLabel
	MaxForestRadius
	Max_forest_radius_down
	Max_forest_radius_up
	MinForestDensityLabel
	MinForestDensity
	Min_forest_density_down
	Min_forest_density_up
	MaxForestDensityLabel
	MaxForestDensity
	Max_forest_density_down
	Max_forest_density_up
	Number_random_trees_label
	Number_random_trees
	Number_random_trees_down
	Number_random_trees_up
	Min_altitude_for_trees_label
	Min_altitude_for_trees
	Min_altitude_for_trees_down
	Min_altitude_for_trees_up
	Max_altitude_for_trees_label
	Max_altitude_for_trees
	Max_altitude_for_trees_down
	Max_altitude_for_trees_up
)
// const uint64_t holdable_widgets = (1ULL << widx::number_of_forests_up) | (1ULL << widx::number_of_forests_down) | (1ULL << widx::min_forest_radius_up) | (1ULL << widx::min_forest_radius_down) | (1ULL << widx::max_forest_radius_up) | (1ULL << widx::max_forest_radius_down) | (1ULL << widx::min_forest_density_up) | (1ULL << widx::min_forest_density_down) | (1ULL << widx::max_forest_density_up) | (1 << widx::max_forest_density_down) | (1ULL << widx::number_random_trees_up) | (1ULL << widx::number_random_trees_down) | (1ULL << widx::min_altitude_for_trees_up) | (1ULL << widx::min_altitude_for_trees_down) | (1ULL << widx::max_altitude_for_trees_down) | (1ULL << widx::max_altitude_for_trees_up);
// static constexpr auto widgets = makeWidgets(
// Common::makeCommonWidgets(217, StringIds::title_landscape_generation_forests),
// Widgets::Label({ 10, 52 }, { 240, 12 }, WindowColour::secondary, ContentAlign::left, StringIds::number_of_forests),
// Widgets::stepperWidgets({ 256, 52 }, { 100, 12 }, WindowColour::secondary, StringIds::number_of_forests_value),
// Widgets::Label({ 10, 67 }, { 240, 12 }, WindowColour::secondary, ContentAlign::left, StringIds::min_forest_radius),
// Widgets::stepperWidgets({ 256, 67 }, { 100, 12 }, WindowColour::secondary, StringIds::min_forest_radius_blocks),
// Widgets::Label({ 10, 82 }, { 240, 12 }, WindowColour::secondary, ContentAlign::left, StringIds::max_forest_radius),
// Widgets::stepperWidgets({ 256, 82 }, { 100, 12 }, WindowColour::secondary, StringIds::max_forest_radius_blocks),
// Widgets::Label({ 10, 97 }, { 240, 12 }, WindowColour::secondary, ContentAlign::left, StringIds::min_forest_density),
// Widgets::stepperWidgets({ 256, 97 }, { 100, 12 }, WindowColour::secondary, StringIds::min_forest_density_percent),
// Widgets::Label({ 10, 112 }, { 240, 12 }, WindowColour::secondary, ContentAlign::left, StringIds::max_forest_density),
// Widgets::stepperWidgets({ 256, 112 }, { 100, 12 }, WindowColour::secondary, StringIds::max_forest_density_percent),
// Widgets::Label({ 10, 127 }, { 240, 12 }, WindowColour::secondary, ContentAlign::left, StringIds::number_random_trees),
// Widgets::stepperWidgets({ 256, 127 }, { 100, 12 }, WindowColour::secondary, StringIds::number_random_trees_value),
// Widgets::Label({ 10, 142 }, { 240, 12 }, WindowColour::secondary, ContentAlign::left, StringIds::min_altitude_for_trees),
// Widgets::stepperWidgets({ 256, 142 }, { 100, 12 }, WindowColour::secondary, StringIds::min_altitude_for_trees_height),
// Widgets::Label({ 10, 157 }, { 240, 12 }, WindowColour::secondary, ContentAlign::left, StringIds::max_altitude_for_trees),
// Widgets::stepperWidgets({ 256, 157 }, { 100, 12 }, WindowColour::secondary, StringIds::max_altitude_for_trees_height)
// );
// // 0x0043E670
// func OnMouseDown(window Window, widgetIndex WidgetIndex_t, id [[maybe_unused]] WidgetId) 
// auto& options = Scenario::getOptions();
// switch (widgetIndex)
// case widx::number_of_forests_up:
// options.numberOfForests = std::min<int16_t>(options.numberOfForests + 10, Scenario::kMaxNumForests);
// break;
// case widx::number_of_forests_down:
// options.numberOfForests = std::max<int16_t>(Scenario::kMinNumForests, options.numberOfForests - 10);
// break;
// case widx::min_forest_radius_up:
// options.minForestRadius = std::min<int16_t>(options.minForestRadius + 1, Scenario::kMaxForestRadius);
// if (options.minForestRadius > options.maxForestRadius)
// options.maxForestRadius = options.minForestRadius;
// break;
// case widx::min_forest_radius_down:
// options.minForestRadius = std::max<int8_t>(Scenario::kMinForestRadius, options.minForestRadius - 1);
// break;
// case widx::max_forest_radius_up:
// options.maxForestRadius = std::clamp<int8_t>(options.maxForestRadius + 1, Scenario::kMinForestRadius, Scenario::kMaxForestRadius);
// break;
// case widx::max_forest_radius_down:
// options.maxForestRadius = std::clamp<int8_t>(options.maxForestRadius - 1, Scenario::kMinForestRadius, Scenario::kMaxForestRadius);
// if (options.maxForestRadius < options.minForestRadius)
// options.minForestRadius = options.maxForestRadius;
// break;
// case widx::min_forest_density_up:
// options.minForestDensity = std::min<int8_t>(options.minForestDensity + 1, Scenario::kMaxForestDensity);
// if (options.minForestDensity > options.maxForestDensity)
// options.maxForestDensity = options.minForestDensity;
// break;
// case widx::min_forest_density_down:
// options.minForestDensity = std::max<int8_t>(Scenario::kMinForestDensity, options.minForestDensity - 1);
// break;
// case widx::max_forest_density_up:
// options.maxForestDensity = std::min<int8_t>(options.maxForestDensity + 1, Scenario::kMaxForestDensity);
// break;
// case widx::max_forest_density_down:
// options.maxForestDensity = std::max<int8_t>(Scenario::kMinForestDensity, options.maxForestDensity - 1);
// if (options.maxForestDensity < options.minForestDensity)
// options.minForestDensity = options.maxForestDensity;
// break;
// case widx::number_random_trees_up:
// options.numberRandomTrees = std::min<int16_t>(options.numberRandomTrees + 25, Scenario::kMaxNumTrees);
// break;
// case widx::number_random_trees_down:
// options.numberRandomTrees = std::max<int16_t>(Scenario::kMinNumTrees, options.numberRandomTrees - 25);
// break;
// case widx::min_altitude_for_trees_up:
// options.minAltitudeForTrees = std::min<int8_t>(options.minAltitudeForTrees + 1, Scenario::kMaxAltitudeTrees);
// if (options.minAltitudeForTrees > options.maxAltitudeForTrees)
// options.maxAltitudeForTrees = options.minAltitudeForTrees;
// break;
// case widx::min_altitude_for_trees_down:
// options.minAltitudeForTrees = std::max<int8_t>(Scenario::kMinAltitudeTrees, options.minAltitudeForTrees - 1);
// break;
// case widx::max_altitude_for_trees_up:
// options.maxAltitudeForTrees = std::min<int8_t>(options.maxAltitudeForTrees + 1, Scenario::kMaxAltitudeTrees);
// break;
// case widx::max_altitude_for_trees_down:
// options.maxAltitudeForTrees = std::max<int8_t>(Scenario::kMinAltitudeTrees, options.maxAltitudeForTrees - 1);
// if (options.maxAltitudeForTrees < options.minAltitudeForTrees)
// options.minAltitudeForTrees = options.maxAltitudeForTrees;
// break;
// default:
// // Nothing was changed, don't invalidate.
// return;
// // After changing any of the options, invalidate the window.
// window.invalidate();
// // 0x0043E44F
// func PrepareDraw(window Window) 
// Common::prepareDraw(window);
// auto& options = Scenario::getOptions();
// auto args = FormatArguments(window.widgets[widx::number_of_forests].textArgs);
// args.push<uint16_t>(options.numberOfForests);
// auto args = FormatArguments(window.widgets[widx::maxForestDensity].textArgs);
// args.push<uint16_t>(options.maxForestDensity);
// auto args = FormatArguments(window.widgets[widx::minForestDensity].textArgs);
// args.push<uint16_t>(options.minForestDensity);
// auto args = FormatArguments(window.widgets[widx::minForestRadius].textArgs);
// args.push<uint16_t>(options.minForestRadius);
// auto args = FormatArguments(window.widgets[widx::maxForestRadius].textArgs);
// args.push<uint16_t>(options.maxForestRadius);
// auto args = FormatArguments(window.widgets[widx::minForestDensity].textArgs);
// args.push<uint16_t>(options.minForestDensity * 14);
// auto args = FormatArguments(window.widgets[widx::maxForestDensity].textArgs);
// args.push<uint16_t>(options.maxForestDensity * 14);
// auto args = FormatArguments(window.widgets[widx::number_random_trees].textArgs);
// args.push<uint16_t>(options.numberRandomTrees);
// auto args = FormatArguments(window.widgets[widx::min_altitude_for_trees].textArgs);
// args.push<uint16_t>(options.minAltitudeForTrees);
// auto args = FormatArguments(window.widgets[widx::max_altitude_for_trees].textArgs);
// args.push<uint16_t>(options.maxAltitudeForTrees);
// static constexpr WindowEventList kEvents = {
// .onMouseUp = Common::onMouseUp,
// .onMouseDown = onMouseDown,
// .onUpdate = Common::update,
// .prepareDraw = prepareDraw,
// .draw = Common::draw,
// static const WindowEventList& getEvents()
// orphan member: return kEvents;
// namespace Towns
type Widx int

const (
	Number_of_towns_label Widx = Common.widx.generate_now + 1
	Number_of_towns
	Number_of_towns_down
	Number_of_towns_up
	Max_town_size_label
	Max_town_size
	Max_town_size_btn
)
// const uint64_t holdable_widgets = (1 << widx::number_of_towns_up) | (1 << widx::number_of_towns_down);
// static constexpr auto widgets = makeWidgets(
// Common::makeCommonWidgets(217, StringIds::title_landscape_generation_towns),
// Widgets::Label({ 10, 52 }, { 240, 12 }, WindowColour::secondary, ContentAlign::left, StringIds::number_of_towns),
// Widgets::stepperWidgets({ 256, 52 }, { 100, 12 }, WindowColour::secondary, StringIds::number_of_towns_value),
// Widgets::Label({ 10, 67 }, { 160, 12 }, WindowColour::secondary, ContentAlign::left, StringIds::max_town_size),
// Widgets::dropdownWidgets({ 176, 67 }, { 180, 12 }, WindowColour::secondary)
// );
// static constexpr StringId townSizeLabels[] = {
// StringIds::town_size_1,
// StringIds::town_size_2,
// StringIds::town_size_3,
// StringIds::town_size_4,
// StringIds::town_size_5,
// StringIds::town_size_6,
// StringIds::town_size_7,
// StringIds::town_size_8,
// // 0x0043EA28
// func OnDropdown(window Window, widgetIndex WidgetIndex_t, id [[maybe_unused]] WidgetId, itemIndex int16) 
// if (widgetIndex != widx::max_town_size_btn || itemIndex == -1)
// return;
// Scenario::getOptions().maxTownSize = itemIndex + 1;
// window.invalidate();
// // 0x0043EA0D
// func OnMouseDown(window Window, widgetIndex WidgetIndex_t, id [[maybe_unused]] WidgetId) 
// auto& options = Scenario::getOptions();
// switch (widgetIndex)
// case widx::number_of_towns_up:
// uint16_t newNumTowns = std::min<uint16_t>(options.numberOfTowns + 1, Limits::kMaxTowns);
// options.numberOfTowns = newNumTowns;
// window.invalidate();
// break;
// case widx::number_of_towns_down:
// // Vanilla behaviour: Zero-town map generation is allowed in the scenario editor and
// // is checked on the editor stage progression for non-zero. It is required to be non-zero
// // for gameplay since industries must have at least one associated town. The user must
// // manually place at least one town if they generate a landscape with zero towns.
// if (options.numberOfTowns > 0)
// uint16_t newNumTowns = std::max<uint16_t>(Limits::kMinTowns - 1, options.numberOfTowns - 1);
// options.numberOfTowns = newNumTowns;
// window.invalidate();
// break;
// case widx::max_town_size_btn:
// Widget& target = window.widgets[widx::max_town_size];
// Dropdown::show(window.x + target.left, window.y + target.top, target.width() - 4, target.height(), window.getColour(WindowColour::secondary), std::size(townSizeLabels), 0x80);
// for (size_t i = 0; i < std::size(townSizeLabels); i++)
// Dropdown::add(i, townSizeLabels[i]);
// Dropdown::setHighlightedItem(options.maxTownSize - 1);
// break;
// // 0x0043E90D
// func PrepareDraw(window Window) 
// Common::prepareDraw(window);
// auto args = FormatArguments(window.widgets[widx::number_of_towns].textArgs);
// args.push<uint16_t>(Scenario::getOptions().numberOfTowns);
// window.widgets[widx::max_town_size].text = townSizeLabels[Scenario::getOptions().maxTownSize - 1];
// static constexpr WindowEventList kEvents = {
// .onMouseUp = Common::onMouseUp,
// .onMouseDown = onMouseDown,
// .onDropdown = onDropdown,
// .onUpdate = Common::update,
// .prepareDraw = prepareDraw,
// .draw = Common::draw,
// static const WindowEventList& getEvents()
// orphan member: return kEvents;
// namespace Industries
type Widx int

const (
	Num_industries_label Widx = Common.widx.generate_now + 1
	Num_industries
	Num_industries_btn
	Check_allow_industries_close_down
	Check_allow_industries_start_up
)
// const uint64_t holdable_widgets = 0;
// static constexpr auto widgets = makeWidgets(
// Common::makeCommonWidgets(217, StringIds::title_landscape_generation_industries),
// Widgets::Label({ 10, 52 }, { 160, 12 }, WindowColour::secondary, ContentAlign::left, StringIds::number_of_industries),
// Widgets::dropdownWidgets({ 176, 52 }, { 180, 12 }, WindowColour::secondary),
// Widgets::Checkbox({ 10, 68 }, { 346, 12 }, WindowColour::secondary, StringIds::allow_industries_to_close_down_during_game),
// Widgets::Checkbox({ 10, 83 }, { 346, 12 }, WindowColour::secondary, StringIds::allow_new_industries_to_start_up_during_game)
// );
// static constexpr StringId numIndustriesLabels[] = {
// StringIds::industry_size_low,
// StringIds::industry_size_medium,
// StringIds::industry_size_high,
// // 0x0043EBF8
// func OnDropdown(window Window, widgetIndex WidgetIndex_t, id [[maybe_unused]] WidgetId, itemIndex int16) 
// if (widgetIndex != widx::num_industries_btn || itemIndex == -1)
// return;
// Scenario::getOptions().numberOfIndustries = itemIndex;
// window.invalidate();
// // 0x0043EBF1
// func OnMouseDown(window Window, widgetIndex WidgetIndex_t, id [[maybe_unused]] WidgetId) 
// switch (widgetIndex)
// case widx::num_industries_btn:
// Widget& target = window.widgets[widx::num_industries];
// Dropdown::show(window.x + target.left, window.y + target.top, target.width() - 4, target.height(), window.getColour(WindowColour::secondary), std::size(numIndustriesLabels), 0x80);
// for (size_t i = 0; i < std::size(numIndustriesLabels); i++)
// Dropdown::add(i, numIndustriesLabels[i]);
// Dropdown::setHighlightedItem(Scenario::getOptions().numberOfIndustries);
// break;
// case widx::check_allow_industries_close_down:
// IndustryManager::setFlags(IndustryManager::getFlags() ^ IndustryManager::Flags::disallowIndustriesCloseDown);
// window.invalidate();
// break;
// case widx::check_allow_industries_start_up:
// IndustryManager::setFlags(IndustryManager::getFlags() ^ IndustryManager::Flags::disallowIndustriesStartUp);
// window.invalidate();
// break;
// // 0x0043EAEB
// func PrepareDraw(window Window) 
// Common::prepareDraw(window);
// window.widgets[widx::num_industries].text = numIndustriesLabels[Scenario::getOptions().numberOfIndustries];
// window.activatedWidgets &= ~((1 << widx::check_allow_industries_close_down) | (1 << widx::check_allow_industries_start_up));
// if (!IndustryManager::hasFlags(IndustryManager::Flags::disallowIndustriesCloseDown))
// window.activatedWidgets |= 1 << widx::check_allow_industries_close_down;
// if (!IndustryManager::hasFlags(IndustryManager::Flags::disallowIndustriesStartUp))
// window.activatedWidgets |= 1 << widx::check_allow_industries_start_up;
// static constexpr WindowEventList kEvents = {
// .onMouseUp = Common::onMouseUp,
// .onMouseDown = onMouseDown,
// .onDropdown = onDropdown,
// .onUpdate = Common::update,
// .prepareDraw = prepareDraw,
// .draw = Common::draw,
// static const WindowEventList& getEvents()
// orphan member: return kEvents;
// namespace Common
// func SwitchTabWidgets(window Window) 
// window->activatedWidgets = 0;
// static std::span<const Widget> widgetCollectionsByTabId[] = {
// Options::widgets,
// Land::widgets,
// Water::widgets,
// Forests::widgets,
// Towns::widgets,
// Industries::widgets,
// auto newWidgets = widgetCollectionsByTabId[window->currentTab];
// window->setWidgets(newWidgets);
// window->initScrollWidgets();
// static constexpr widx tabWidgetIdxByTabId[] = {
// tab_options,
// tab_land,
// tab_water,
// tab_forests,
// tab_towns,
// tab_industries,
// window->activatedWidgets &= ~((1 << tab_options) | (1 << tab_land) | (1 << tab_water) | (1 << tab_forests) | (1 << tab_towns) | (1 << tab_industries));
// window->activatedWidgets |= (1ULL << tabWidgetIdxByTabId[window->currentTab]);
// // 0x0043DC98
// func SwitchTab(self Window, widgetIndex WidgetIndex_t) 
// if (ToolManager::isToolActive(self.type, self.number))
// ToolManager::toolCancel();
// self.currentTab = widgetIndex - widx::tab_options;
// self.frameNo = 0;
// self.flags &= ~(WindowFlags::beingResized);
// self.disabledWidgets = 0;
// static const uint64_t* holdableWidgetsByTab[] = {
// &Options::holdable_widgets,
// &Land::holdable_widgets,
// &Water::holdable_widgets,
// &Forests::holdable_widgets,
// &Towns::holdable_widgets,
// &Industries::holdable_widgets,
// self.holdableWidgets = *holdableWidgetsByTab[self.currentTab];
// static const WindowEventList* eventsByTab[] = {
// &Options::getEvents(),
// &Land::getEvents(),
// &Water::getEvents(),
// &Forests::getEvents(),
// &Towns::getEvents(),
// &Industries::getEvents(),
// self.eventHandlers = eventsByTab[self.currentTab];
// switchTabWidgets(&self);
// self.invalidate();
// const auto newSize = [widgetIndex]() {
// if (widgetIndex == widx::tab_land)
// orphan member: return kLandTabSize;
// else
// orphan member: return kWindowSize;
// }();
// self.setSize(newSize);
// self.callOnResize();
// self.callPrepareDraw();
// self.initScrollWidgets();
// self.invalidate();
// self.moveInsideScreenEdges();
