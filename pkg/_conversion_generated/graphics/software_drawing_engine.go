package graphics

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Graphics/Gfx.h"
// #include "InvalidationGrid.h"
// #include "SoftwareDrawingContext.h"
// #include <OpenLoco/Engine/Ui/Rect.hpp>
// #include <algorithm>
// #include <cstddef>
// #include <memory>
// forward: struct SDL_Palette;
// forward: struct SDL_Surface;
// forward: struct SDL_Window;
// forward: struct SDL_Renderer;
// forward: struct SDL_Texture;
// forward: struct SDL_PixelFormat;
// namespace OpenLoco::Ui
// forward: struct ScreenInfo;
// namespace OpenLoco::Gfx
// forward: struct RenderTarget;
type SoftwareDrawingEngine struct {
// SoftwareDrawingEngine();
// ~SoftwareDrawingEngine();
	// method: void initialize(SDL_Window* window);
	// method: void resize(int32_t width, int32_t height);
// // Renders all invalidated regions.
	// method: void render();
// // Renders a specific region.
	// method: void render(const Ui::Rect& rect);
// // Presents the final image to the screen.
	// method: void present();
// // Invalidates a region, this forces it to be rendered next frame.
	// method: void invalidateRegion(int32_t left, int32_t top, int32_t right, int32_t bottom);
	// method: void createPalette();
// SDL_Palette* getPalette() { return _palette; }
	// method: void updatePalette(const PaletteEntry* entries, int32_t index, int32_t count);
// DrawingContext& getDrawingContext();
// const RenderTarget& getScreenRT();
// // Moves the pixels in the specified render target.
// void movePixels(
// const RenderTarget& rt,
// int16_t dstX,
// int16_t dstY,
// int16_t width,
// int16_t height,
// int16_t srcX,
// int16_t srcY);
// const Ui::ScreenInfo& getScreenInfo() const;
	// method: bool setVSync(bool state);
	// method: void renderDirtyRegions();
// SDL_Renderer* _renderer{};
// SDL_Window* _window{};
// SDL_Palette* _palette{};
// SDL_Surface* _screenSurface{};
// SDL_Surface* _screenRGBASurface{};
// SDL_Texture* _screenTexture{};
// SDL_Texture* _scaledScreenTexture{};
// SDL_PixelFormat* _screenTextureFormat{};
// SDL_Texture* _screenRGBATexture{};
	Ctx SoftwareDrawingContext
	InvalidationGrid InvalidationGrid
// bool _vsync = false;
}
