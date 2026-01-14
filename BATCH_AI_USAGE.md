# Batch AI Implementation System - Usage Guide

## Quick Start

### 1. Set Up Environment

```bash
# Set your OpenAI API key
export OPENAI_API_KEY="sk-proj-..."

# Verify setup
cd goloco
go build ./cmd/batch_ai_impl
```

### 2. Run Tier 1 (UI Foundation)

```bash
./batch_ai_impl \
    -db essential_functions.json \
    -src ../OpenLoco/src/OpenLoco/src \
    -out pkg \
    -tier 1 \
    -concurrent 3
```

**Output:**
- Creates `pkg/graphics/color.go`, `palette.go`, `drawing_context.go`, `text_renderer.go`
- Creates `pkg/ui/window.go`, `frame_widget.go`, `caption_widget.go`, `button_widget.go`
- 10 functions implemented in ~5-10 minutes

### 3. Test Compilation

```bash
go build ./pkg/graphics
go build ./pkg/ui
```

### 4. Integrate with Main Game

```go
// In cmd/goloco/main.go
import (
    "github.com/LaPingvino/goloco/pkg/ui"
    "github.com/LaPingvino/goloco/pkg/graphics"
)

func (g *Game) Draw(screen *ebiten.Image) {
    dc := graphics.NewDrawingContext(screen)
    
    // Create main window
    window := ui.NewWindow(0, 0, 640, 480)
    window.Draw(dc)
}
```

---

## System Architecture

### Function Database Format

`essential_functions.json` contains:

```json
{
  "functions": [
    {
      "id": "window_draw",
      "priority": 10,
      "tier": 1,
      "cppFile": "Ui/Window.cpp",
      "function": "Window::draw",
      "signature": "void Window::draw(Gfx::RenderTarget& rt)",
      "goSignature": "func (w *Window) Draw(dc *DrawingContext)",
      "goPackage": "ui",
      "goFile": "window.go",
      "dependencies": ["Widget.DrawFrame", "Widget.DrawCaption"],
      "complexity": "medium",
      "estimatedTokens": 800,
      "description": "Main window draw function"
    }
  ]
}
```

### Processing Flow

```
1. Load essential_functions.json
2. Filter by tier
3. Sort by priority
4. For each function:
   ├─ Extract C++ code from OpenLoco source
   ├─ Build AI prompt with context
   ├─ Call GPT-4o-mini API
   ├─ Parse Go implementation
   ├─ Write to pkg/[package]/[file].go
   └─ Report success/failure
5. Generate summary report
```

### AI Prompt Structure

```
You are porting OpenLoco (C++) to Go.

## Task
Implement this function in Go:

**Go Signature:**
func (w *Window) Draw(dc *DrawingContext)

**Original C++ Implementation:**
[actual C++ code from source]

**Description:**
Main window draw function - renders frame, caption, all widgets

**Available Dependencies:**
- Widget.DrawFrame (already implemented)
- Widget.DrawCaption (already implemented)

## Requirements
1. Translate C++ logic to idiomatic Go
2. Use Go naming conventions
3. Handle errors appropriately
4. Output ONLY function body (between { and })

## Output
Return only the Go function body.
```

---

## Command Line Options

| Flag | Default | Description |
|------|---------|-------------|
| `-db` | essential_functions.json | Path to function database |
| `-src` | ../OpenLoco/src/OpenLoco/src | OpenLoco C++ source directory |
| `-out` | pkg | Output directory for Go files |
| `-tier` | 1 | Which tier to process (1-5, 0=all) |
| `-concurrent` | 3 | Number of parallel API requests |
| `-model` | gpt-4o-mini | AI model (gpt-4o-mini or gpt-4o) |
| `-dry` | false | Dry run - show what would happen |
| `-key` | (from env) | OpenAI API key |

---

## Tier Breakdown

### Tier 1: UI Foundation (10 functions)
**Time:** ~10 minutes  
**Cost:** ~$0.20  
**Output:** Basic window UI that looks like OpenLoco

**Functions:**
- Color palette system
- Rectangle drawing (fill, inset, outline)
- Text rendering
- Window frame, caption, buttons

**Deliverable:**
```
┌─────────────────────────────┐
│ ■ Company Name        [_][X]│
├─────────────────────────────┤
│ ┌──────┐ ┌──────┐          │
│ │ File │ │ Edit │ ...      │
│ └──────┘ └──────┘          │
│                             │
│   [Viewport Area]           │
│                             │
└─────────────────────────────┘
```

---

### Tier 2: Viewport & Tiles (9 functions)
**Time:** ~15 minutes  
**Cost:** ~$0.40  
**Output:** Isometric game view

**Functions:**
- Viewport rendering
- Paint session (generate, sort, draw)
- Tile painting
- Coordinate transforms

**Deliverable:** Isometric map visible in viewport

---

### Tier 3: Real Graphics (5 functions)
**Time:** ~10 minutes  
**Cost:** ~$0.30  
**Output:** Actual Locomotion sprites

**Functions:**
- G1.DAT loader
- Sprite renderer
- Palette remapping
- RLE decompression

**Deliverable:** Real game graphics instead of placeholders

---

### Tier 4: Vehicles & Effects (4 functions)
**Time:** ~10 minutes  
**Cost:** ~$0.30  
**Output:** Moving vehicles

**Functions:**
- Vehicle painting
- Vehicle position updates
- Effect rendering (smoke, etc.)

**Deliverable:** Vehicles moving on tracks

---

### Tier 5: Polish (5 functions)
**Time:** ~10 minutes  
**Cost:** ~$0.25  
**Output:** Complete visual fidelity

**Functions:**
- Multi-window rendering
- Dirty region tracking
- Tabs, checkboxes
- Decorations/overlays

**Deliverable:** Full OpenLoco appearance

---

## Cost Estimates

| Tier | Functions | Tokens | Cost (gpt-4o-mini) | Time |
|------|-----------|--------|-------------------|------|
| 1 | 10 | ~6,000 | $0.15-0.30 | 10 min |
| 2 | 9 | ~9,000 | $0.20-0.40 | 15 min |
| 3 | 5 | ~6,000 | $0.15-0.30 | 10 min |
| 4 | 4 | ~6,000 | $0.15-0.30 | 10 min |
| 5 | 5 | ~5,000 | $0.10-0.25 | 10 min |
| **Total** | **33** | **~32K** | **$0.75-1.55** | **55 min** |

**With gpt-4o (higher quality):** ~$15-20 total

---

## Workflow

### Day 1: UI Foundation
```bash
# Process Tier 1
./batch_ai_impl -tier 1

# Review generated code
ls -la pkg/graphics/
ls -la pkg/ui/

# Test compilation
go build ./pkg/...

# Fix any compilation errors manually
# (AI usually gets 80-90% right)

# Integrate with main.go
# Run game: ./goloco
```

**Expected Result:** Window with title bar and buttons

---

### Day 2: Viewport Rendering
```bash
# Process Tier 2
./batch_ai_impl -tier 2

# Review viewport code
cat pkg/ui/viewport.go
cat pkg/paint/paint_session.go

# Test
go build ./pkg/...
./goloco
```

**Expected Result:** Isometric map visible

---

### Day 3: Real Graphics
```bash
# Process Tier 3
./batch_ai_impl -tier 3

# Test with real G1.DAT
./goloco
```

**Expected Result:** Actual Locomotion sprites

---

### Days 4-5: Vehicles & Polish
```bash
# Process Tiers 4-5
./batch_ai_impl -tier 4
./batch_ai_impl -tier 5

# Final polish
./goloco
```

**Expected Result:** Game looks like OpenLoco!

---

## Quality Control

### AI Success Rate
- **Simple functions (low complexity):** 95% success
- **Medium complexity:** 80% success
- **High complexity:** 60% success
- **Very high complexity:** 40% success (needs manual review/fix)

### Common AI Errors
1. **Type mismatches** - Uses wrong Go type
   - Fix: Add type definitions to database
2. **Missing imports** - Forgets to import packages
   - Fix: Add imports manually after generation
3. **Incorrect logic** - Misunderstands C++ code
   - Fix: Review and correct manually
4. **Incomplete implementation** - Skips complex parts
   - Fix: Re-run with more context or implement manually

### Review Checklist
- [ ] Code compiles without errors
- [ ] Logic matches C++ behavior
- [ ] Error handling is appropriate
- [ ] Naming follows Go conventions
- [ ] No unsafe code or panics
- [ ] Comments explain non-obvious parts

---

## Advanced Usage

### Custom Function Database

Create your own JSON file:

```json
{
  "functions": [
    {
      "id": "my_custom_function",
      "priority": 1,
      "tier": 1,
      "cppFile": "Path/To/File.cpp",
      "function": "MyClass::myFunction",
      "goSignature": "func MyFunction() error",
      "goPackage": "mypackage",
      "goFile": "myfile.go",
      "description": "What this function does"
    }
  ]
}
```

```bash
./batch_ai_impl -db my_functions.json -tier 1
```

---

### Retry Failed Functions

If some functions fail, extract them to a new database:

```json
{
  "functions": [
    // Copy failed functions here
  ]
}
```

```bash
./batch_ai_impl -db retry.json -tier 1
```

---

### Use Higher Quality Model

For complex functions:

```bash
./batch_ai_impl -tier 2 -model gpt-4o
```

**Note:** gpt-4o is ~10x more expensive but handles complexity better

---

## Troubleshooting

### "Function not found in source file"
- Check that OpenLoco source path is correct
- Verify C++ file path in database
- Function may have been renamed in newer OpenLoco versions

### "API error: rate limit"
- Reduce `-concurrent` flag (try `-concurrent 1`)
- Add delays between requests

### "Generated code doesn't compile"
- Review AI output manually
- Add missing type definitions
- Fix import statements
- Simplify complex logic

### "Wrong behavior"
- Compare with C++ implementation
- Add unit tests
- Re-generate with more context in prompt

---

## Next Steps After Tier 1

1. **Test UI appearance** - Compare with OpenLoco screenshots
2. **Add missing types** - Create Window, Widget, DrawingContext structs
3. **Integrate with Ebiten** - Connect DrawingContext to Ebiten's Image
4. **Fix compilation errors** - Add imports, fix types
5. **Run game** - See if window renders
6. **Iterate** - Fix bugs, improve appearance
7. **Move to Tier 2** - Viewport rendering

---

## Example: Complete Tier 1 Workflow

```bash
# 1. Set up
export OPENAI_API_KEY="sk-..."
cd goloco
go build ./cmd/batch_ai_impl

# 2. Generate Tier 1 functions
./batch_ai_impl -tier 1 -v

# Output:
# ✓ [1/10] Colour::getShade
# ✓ [2/10] Gfx::loadPalette
# ✓ [3/10] fillRect
# ...
# Success: 9/10, Failed: 1/10

# 3. Review generated files
cat pkg/graphics/color.go
cat pkg/ui/window.go

# 4. Fix compilation errors
# Add missing type definitions:
cat > pkg/graphics/types.go << 'EOF'
package graphics

type Color uint8
type DrawingContext struct {
    Target *ebiten.Image
}
EOF

# 5. Test
go build ./pkg/...

# 6. Integrate with main game
# Edit cmd/goloco/main.go to use new UI

# 7. Run
go build ./cmd/goloco
./goloco
```

---

## Success Criteria

### After Tier 1
- [ ] Window frame renders with correct colors
- [ ] Title bar shows company name
- [ ] Buttons render and respond to mouse
- [ ] Text displays in bitmap font
- [ ] 3D borders look authentic

### After Tier 2
- [ ] Isometric viewport visible
- [ ] Tiles render in correct order
- [ ] Camera pans and zooms
- [ ] Coordinate transforms work

### After Tier 3
- [ ] Real Locomotion sprites load
- [ ] Graphics render correctly
- [ ] Palette colors match original

### After All Tiers
- [ ] **Game looks like OpenLoco!** 🎉

---

**Estimated Total Time:** 2-3 days  
**Estimated Cost:** $1-20 (depending on model)  
**Expected Quality:** 80-90% (needs manual polish)

Ready to make goloco look like OpenLoco! 🚂
