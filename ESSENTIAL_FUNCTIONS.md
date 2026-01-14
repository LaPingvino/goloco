# Essential Functions for OpenLoco Visual Authenticity

## Priority Rankings for AI Implementation

### Tier 1: CRITICAL - Make Game Visually Recognizable (Week 1)
**Goal:** Game window looks like OpenLoco with proper UI layout

| Priority | Function | Source File | Purpose | Complexity |
|----------|----------|-------------|---------|------------|
| 1 | Window::draw() | Ui/Window.cpp | Main window rendering | Medium |
| 2 | Widget::draw() dispatch | Ui/Widget.cpp | Widget rendering system | Medium |
| 3 | Frame::draw() | Ui/Widgets/FrameWidget.cpp | Window frame background | Low |
| 4 | Caption::draw() | Ui/Widgets/CaptionWidget.cpp | Title bar | Low |
| 5 | Button::draw() | Ui/Widgets/ButtonWidget.cpp | UI buttons | Medium |
| 6 | TextRenderer::drawString() | Graphics/TextRenderer.cpp | Text rendering | High |
| 7 | fillRect() | Graphics/SoftwareDrawingContext.cpp | Rectangle fills | Low |
| 8 | fillRectInset() | Graphics/SoftwareDrawingContext.cpp | 3D beveled rectangles | Low |
| 9 | Colours::getShade() | Graphics/Colour.cpp | Color palette lookups | Low |
| 10 | loadPalette() | Graphics/Gfx.cpp | Load 256-color palette | Medium |

**Deliverable:** Window with title bar, buttons, panels that looks like OpenLoco UI

---

### Tier 2: VIEWPORT - Show Game World (Week 2)
**Goal:** Isometric game view displays tiles/terrain

| Priority | Function | Source File | Purpose | Complexity |
|----------|----------|-------------|---------|------------|
| 11 | Viewport::render() | Ui/Viewport.hpp | Main viewport rendering | High |
| 12 | Viewport::paint() | Ui/Viewport.hpp (private) | Internal paint logic | High |
| 13 | PaintSession::generate() | Paint/Paint.cpp | Generate paint structs | High |
| 14 | PaintSession::arrangeStructs() | Paint/Paint.cpp | Z-order sorting | Medium |
| 15 | PaintSession::drawStructs() | Paint/Paint.cpp | Render sorted sprites | Medium |
| 16 | paintTileElements() | Paint/PaintTile.cpp | Paint terrain/tracks | Very High |
| 17 | SoftwareDrawingContext::drawImage() | Graphics/SoftwareDrawingContext.cpp | Sprite rendering | High |
| 18 | screenToViewport() | Ui/Viewport.hpp | Coordinate transformation | Low |
| 19 | viewportToScreen() | Ui/Viewport.hpp | Coordinate transformation | Low |

**Deliverable:** Isometric viewport showing game map with tiles

---

### Tier 3: SPRITES - Load Real Graphics (Week 3)
**Goal:** Display actual Locomotion sprites instead of placeholders

| Priority | Function | Source File | Purpose | Complexity |
|----------|----------|-------------|---------|------------|
| 20 | loadG1() | Graphics/Gfx.cpp | Load G1.DAT sprite database | High |
| 21 | getG1Element() | Graphics/Gfx.cpp | Retrieve sprite by ID | Low |
| 22 | drawImagePaletteSet() | Graphics/SoftwareDrawingContext.cpp | Palette-remapped sprites | Medium |
| 23 | drawSpriteToBuffer<>() | Graphics/DrawSprite.h | Template sprite renderer | Very High |
| 24 | getDrawBlendOp() | Graphics/DrawSprite.h | Determine blend mode | Low |

**Deliverable:** Real Locomotion graphics rendering

---

### Tier 4: ENTITIES - Vehicles & Effects (Week 4)
**Goal:** Moving vehicles and visual effects

| Priority | Function | Source File | Purpose | Complexity |
|----------|----------|-------------|---------|------------|
| 25 | paintVehicleEntity() | Paint/PaintVehicle.cpp | Render vehicles | Very High |
| 26 | Vehicle::update() | Vehicles/Vehicle.cpp | Update vehicle position | High |
| 27 | Vehicle::updatePosition() | Vehicles/Vehicle.cpp | Movement logic | High |
| 28 | paintEffectEntity() | Paint/PaintEffectEntity.cpp | Smoke/explosions | Medium |

**Deliverable:** Vehicles moving on tracks with smoke effects

---

### Tier 5: ADVANCED - Polish & Features (Week 5+)
**Goal:** Complete visual fidelity

| Priority | Function | Source File | Purpose | Complexity |
|----------|----------|-------------|---------|------------|
| 29 | WindowManager::render() | Ui/WindowManager.cpp | Multi-window rendering | High |
| 30 | invalidateRegion() | Graphics/Gfx.cpp | Dirty region tracking | Medium |
| 31 | Tab::draw() | Ui/Widgets/TabWidget.cpp | Tabbed interface | Medium |
| 32 | Checkbox::draw() | Ui/Widgets/CheckboxWidget.cpp | Checkbox widget | Low |
| 33 | paintTileElements2() | Paint/PaintTile.cpp | Decorations/overlays | Very High |

**Deliverable:** Full visual feature parity with OpenLoco

---

## Implementation Strategy

### Phase 1: UI Foundation (Days 1-3)
**Functions:** 1-10 (Tier 1)  
**Output:** Window system that looks like OpenLoco

```go
// Target visual:
┌─────────────────────────────────────────┐
│ ■ OpenLoco                      [_][□][X]│
├─────────────────────────────────────────┤
│ ┌─────┐ ┌─────┐ ┌─────┐                │
│ │FILE │ │EDIT │ │VIEW │  ...           │
│ └─────┘ └─────┘ └─────┘                │
├─────────────────────────────────────────┤
│                                         │
│     [Main Game Viewport Area]          │
│                                         │
│                                         │
└─────────────────────────────────────────┘
```

**AI Implementation Targets:**
- Window frame with proper colors (#003C74 dark blue)
- Title bar with company name
- Menu buttons (File, Edit, View, etc.)
- Status bar at bottom
- 3D inset/outset borders

---

### Phase 2: Viewport & Tiles (Days 4-7)
**Functions:** 11-19 (Tier 2)  
**Output:** Isometric game world visible

```go
// Target visual:
┌─────────────────────────────────────────┐
│        [Viewport showing game map]      │
│                                         │
│           ▓▓   ▓▓   ▓▓                 │
│         ▓▓  ▓▓   ▓▓   ▓▓               │
│       ▓▓ [TILE] [TILE] [TILE]          │
│     ▓▓  ▓▓   ▓▓   ▓▓   ▓▓              │
│       [TILE] [TILE] [TILE]             │
│         ▓▓   ▓▓   ▓▓                   │
└─────────────────────────────────────────┘
```

**AI Implementation Targets:**
- Isometric coordinate transformation
- Tile depth sorting (painter's algorithm)
- Camera pan and zoom
- Viewport clipping

---

### Phase 3: Real Graphics (Days 8-10)
**Functions:** 20-24 (Tier 3)  
**Output:** Actual Locomotion sprites

**AI Implementation Targets:**
- G1.DAT parsing
- RLE sprite decompression
- Palette remapping for company colors
- Sprite scaling for zoom levels

---

### Phase 4: Movement & Life (Days 11-14)
**Functions:** 25-28 (Tier 4)  
**Output:** Moving vehicles

**AI Implementation Targets:**
- Vehicle position updates
- Bogie/body rendering
- Smoke particle effects
- Route following

---

## Batch AI Implementation System

### Function Database Format

Each function has:
```json
{
  "id": "window_draw",
  "priority": 1,
  "tier": 1,
  "cppFile": "src/OpenLoco/src/Ui/Window.cpp",
  "function": "Window::draw()",
  "signature": "void Window::draw(Gfx::RenderTarget& rt)",
  "goSignature": "func (w *Window) Draw(rt *RenderTarget)",
  "dependencies": ["Widget::draw", "Caption::draw", "Frame::draw"],
  "complexity": "medium",
  "estimatedTokens": 800,
  "description": "Renders window frame, caption, and all widgets in order"
}
```

### Batch Processing Workflow

```
1. Load function database (ESSENTIAL_FUNCTIONS.json)
2. Sort by priority
3. For each function:
   a. Extract C++ implementation from source
   b. Identify dependencies (already implemented?)
   c. Generate AI prompt with context
   d. Call GPT-4o-mini API
   e. Parse response and extract Go code
   f. Write to pkg/ui/, pkg/graphics/, pkg/paint/
   g. Test compilation
   h. If fails, retry with error context
4. Generate summary report
```

### AI Prompt Template

```
You are porting OpenLoco (C++) to Go for the goloco project.

## Task
Implement this function in idiomatic Go:

**Go Signature:**
```go
func (w *Window) Draw(rt *RenderTarget)
```

**Original C++ Implementation:**
```cpp
void Window::draw(Gfx::RenderTarget& rt) {
    // [actual C++ code]
}
```

**Available Go Types:**
```go
type Window struct {
    X, Y, Width, Height int32
    Widgets []Widget
    Colors [4]Color
}

type RenderTarget struct {
    Bits []byte
    Width, Height int32
}
```

**Dependencies Already Implemented:**
- Widget.Draw(rt)
- DrawingContext.FillRect(x, y, w, h, color)

## Requirements
1. Translate C++ logic to Go idiomatically
2. Use provided Go types
3. Call implemented dependencies
4. Handle errors with Go error type
5. Output ONLY the function body (no signature)
6. Add comments explaining non-obvious parts

## Output Format
Return only the Go function body between { and }
```

---

## Complexity Ratings Explained

- **Low:** Simple math, lookups, trivial logic (< 50 LOC)
- **Medium:** Conditionals, loops, struct manipulation (50-150 LOC)
- **High:** Complex algorithms, state machines, coordinate transforms (150-300 LOC)
- **Very High:** Large switch statements, template specializations, multi-phase rendering (300+ LOC)

---

## Implementation Files Mapping

| Tier | Go Package | Files to Create |
|------|------------|-----------------|
| 1 | pkg/ui | window.go, widget.go, frame_widget.go, caption_widget.go, button_widget.go |
| 1 | pkg/graphics | drawing_context.go, color.go, text_renderer.go |
| 2 | pkg/ui | viewport.go |
| 2 | pkg/paint | paint_session.go, paint_tile.go |
| 3 | pkg/graphics | g1_loader.go, sprite_renderer.go |
| 3 | pkg/assets | g1.go (upgrade existing) |
| 4 | pkg/paint | paint_vehicle.go, paint_effect.go |
| 4 | pkg/entities | vehicle.go, vehicle_update.go |

---

## Success Criteria by Tier

### Tier 1 Success
- [ ] Window frame renders with correct colors
- [ ] Title bar shows company name
- [ ] Buttons are clickable and highlight on hover
- [ ] Text renders in correct font
- [ ] 3D borders look authentic (inset/outset)

### Tier 2 Success
- [ ] Viewport shows isometric grid
- [ ] Tiles render in correct depth order
- [ ] Camera pan with mouse drag
- [ ] Zoom in/out with mouse wheel
- [ ] Viewport clipping works correctly

### Tier 3 Success
- [ ] Real Locomotion sprites display
- [ ] Company colors apply correctly
- [ ] Zoom levels switch sprite assets
- [ ] Transparency works (RLE decoding)

### Tier 4 Success
- [ ] Vehicles move along tracks
- [ ] Smoke particles animate
- [ ] Vehicle sprites change with direction
- [ ] Multiple vehicles render correctly

---

## Estimated Timeline

| Phase | Duration | Functions | Deliverable |
|-------|----------|-----------|-------------|
| Tier 1 | 3 days | 10 | UI shell |
| Tier 2 | 4 days | 9 | Viewport rendering |
| Tier 3 | 3 days | 5 | Real graphics |
| Tier 4 | 4 days | 4 | Vehicles |
| Tier 5 | 7 days | 5 | Polish |
| **Total** | **21 days** | **33** | **Feature parity** |

With AI assistance: **~14 days** (2 weeks)

---

## Next Steps

1. **Create function database** (ESSENTIAL_FUNCTIONS.json)
2. **Build batch AI processor** (cmd/batch_ai_impl)
3. **Implement Tier 1 functions** (UI foundation)
4. **Test visual output** against OpenLoco screenshots
5. **Iterate through tiers** week by week

**Target:** OpenLoco-looking game in 2 weeks! 🎮
