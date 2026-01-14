# Batch AI Implementation with OpenCode + GPT-5-mini

## 🎉 Free Unlimited AI Implementation!

Using `opencode` CLI with your GitHub Pro account, you get **free unlimited access to GPT-5-mini**!

No API keys needed, no costs, unlimited generations! 🚀

---

## Quick Start

### 1. Build the Batch Processor

```bash
cd goloco
go build ./cmd/batch_ai_impl
```

### 2. Run Tier 1 (UI Foundation)

```bash
./batch_ai_impl -tier 1 -v
```

**That's it!** No API key setup, completely free!

---

## What Happens

The batch processor will:
1. Load `essential_functions.json` (10 functions for Tier 1)
2. Extract C++ code from OpenLoco source
3. For each function:
   - Build a detailed prompt
   - Call `opencode run -m github-copilot/gpt-5-mini`
   - Parse the generated Go code
   - Write to `pkg/graphics/*.go` and `pkg/ui/*.go`
4. Report success/failure for each function

**Time:** ~10-15 minutes for 10 functions  
**Cost:** FREE! ✨  
**Concurrency:** 3 parallel requests (adjustable with `-concurrent`)

---

## Command Line Options

```bash
./batch_ai_impl [options]

Options:
  -tier N           Which tier to process (1-5, default: 1)
  -concurrent N     Parallel requests (default: 3)
  -model MODEL      Model to use (default: github-copilot/gpt-5-mini)
  -db FILE          Function database (default: essential_functions.json)
  -src PATH         OpenLoco source (default: ../OpenLoco/src/OpenLoco/src)
  -out PATH         Output directory (default: pkg)
  -dry              Dry run - show what would be done
  -v                Verbose output
```

---

## Examples

### Process Tier 1 (UI)
```bash
./batch_ai_impl -tier 1
```

Output:
```
📚 Loaded 10 functions from database
🎯 Processing tier 1: 10 functions

🤖 Using model: github-copilot/gpt-5-mini
⚡ Concurrency: 3 parallel requests

✅ Colour::getShade
✅ Gfx::loadPalette
✅ fillRect
✅ fillRectInset
✅ drawRectInset
✅ TextRenderer::drawString
✅ Frame::draw
✅ Caption::draw
✅ Button::draw
✅ Window::draw

==================================================
📊 Summary
==================================================
✅ Successful: 10/10
❌ Failed: 0/10

✨ Generated code in: pkg/
🔨 Next step: go build ./pkg/...
```

### Process All Tiers
```bash
./batch_ai_impl -tier 0
```

Processes all 33 functions across all 5 tiers.

### Dry Run
```bash
./batch_ai_impl -tier 1 -dry
```

Shows what would be processed without calling opencode.

### Verbose Mode
```bash
./batch_ai_impl -tier 1 -v
```

Shows progress for each function as it's being processed.

---

## Generated File Structure

After Tier 1:
```
pkg/
├── graphics/
│   ├── color.go            # GetShade() - palette lookups
│   ├── palette.go          # LoadPalette() - 256-color palette
│   ├── drawing_context.go  # FillRect(), FillRectInset(), DrawRectInset()
│   └── text_renderer.go    # DrawString() - text rendering
└── ui/
    ├── frame_widget.go      # DrawFrame() - window backgrounds
    ├── caption_widget.go    # DrawCaption() - title bars
    ├── button_widget.go     # DrawButton() - buttons
    └── window.go            # Window.Draw() - main window rendering
```

---

## How It Works

### 1. Function Database (`essential_functions.json`)

Defines which functions to implement:

```json
{
  "id": "window_draw",
  "priority": 10,
  "tier": 1,
  "cppFile": "Ui/Window.cpp",
  "function": "Window::draw",
  "goSignature": "func (w *Window) Draw(dc *DrawingContext)",
  "goPackage": "ui",
  "goFile": "window.go",
  "dependencies": ["Widget.DrawFrame", "Widget.DrawCaption"],
  "description": "Main window draw function"
}
```

### 2. C++ Code Extraction

The tool finds the function in OpenLoco source:

```cpp
// From OpenLoco/src/Ui/Window.cpp
void Window::draw(Gfx::RenderTarget& rt) {
    // Draw frame
    // Draw caption
    // Draw widgets
    ...
}
```

### 3. Prompt Generation

Creates a detailed prompt for opencode:

```
# Port OpenLoco C++ Function to Go

## Task
Implement this function in idiomatic Go for goloco.

## Go Function Signature
func (w *Window) Draw(dc *DrawingContext)

## Original C++ Implementation
[actual C++ code]

## Description
Main window draw function - renders frame, caption, all widgets

## Available Dependencies
- Widget.DrawFrame (already implemented)
- Widget.DrawCaption (already implemented)

## Requirements
1. Translate C++ logic to idiomatic Go
2. Use Go naming conventions
3. Handle errors appropriately
4. Output ONLY the function body

## Output Format
Return ONLY the Go function body. No signature, no package.
```

### 4. OpenCode Call

```bash
opencode run \
  -m github-copilot/gpt-5-mini \
  --format json \
  --file prompt.txt \
  "Read the attached file and implement..."
```

### 5. Code Generation

GPT-5-mini analyzes the C++ code and generates Go:

```go
// Frame drawing
if w.Frame != nil {
    w.Frame.DrawFrame(w, dc)
}

// Caption (title bar)
if w.Caption != nil {
    w.Caption.DrawCaption(w, dc)
}

// Widgets
for _, widget := range w.Widgets {
    widget.Draw(w, dc)
}
```

### 6. File Writing

Writes to `pkg/ui/window.go`:

```go
package ui

// Auto-generated by batch_ai_impl using opencode + GPT-5-mini
// Original: Ui/Window.cpp

// Main window draw function - renders frame, caption, and all widgets
// Source: Window::draw
func (w *Window) Draw(dc *DrawingContext) {
    // Frame drawing
    if w.Frame != nil {
        w.Frame.DrawFrame(w, dc)
    }
    
    // Caption (title bar)
    if w.Caption != nil {
        w.Caption.DrawCaption(w, dc)
    }
    
    // Widgets
    for _, widget := range w.Widgets {
        widget.Draw(w, dc)
    }
}
```

---

## Advantages of OpenCode vs. Direct API

| Feature | OpenCode | Direct API |
|---------|----------|------------|
| **Cost** | FREE! | $1-20 per run |
| **Setup** | No API key needed | Requires OPENAI_API_KEY |
| **Rate Limits** | None (GitHub Pro) | 10,000 requests/day |
| **Model** | GPT-5-mini | GPT-4o-mini |
| **Quality** | Newer model | Older model |

OpenCode is the clear winner! 🏆

---

## Troubleshooting

### "opencode: command not found"
Install opencode:
```bash
npm install -g @opencode/cli
# or
curl -fsSL https://opencode.sh/install.sh | sh
```

### "Error: No GitHub Pro account"
Make sure you're logged in:
```bash
opencode auth login
```

### Function extraction fails
Check that OpenLoco source path is correct:
```bash
ls -la ../OpenLoco/src/OpenLoco/src/
```

### Generated code doesn't compile
1. Review the generated file
2. Add missing type definitions
3. Fix imports
4. Re-run if needed

---

## Next Steps After Tier 1

1. **Test Compilation**
   ```bash
   go build ./pkg/graphics
   go build ./pkg/ui
   ```

2. **Create Type Definitions**
   ```bash
   # Add missing types that AI referenced
   cat > pkg/graphics/types.go << 'EOF'
   package graphics
   
   type Color uint8
   type DrawingContext struct {
       // TODO: Define fields
   }
   EOF
   ```

3. **Integrate with Main Game**
   ```go
   // In cmd/goloco/main.go
   import "github.com/LaPingvino/goloco/pkg/ui"
   
   window := ui.NewWindow(0, 0, 640, 480)
   window.Draw(dc)
   ```

4. **Run Tier 2**
   ```bash
   ./batch_ai_impl -tier 2
   ```

---

## Full Tier Workflow

### Day 1: Tier 1 (UI Foundation)
```bash
./batch_ai_impl -tier 1
go build ./pkg/...
# Fix compilation errors
# Integrate with main.go
./goloco
```

**Result:** Window with OpenLoco UI appearance

### Day 2: Tier 2 (Viewport)
```bash
./batch_ai_impl -tier 2
go build ./pkg/...
# Fix and integrate
./goloco
```

**Result:** Isometric game view

### Day 3: Tiers 3-5 (Graphics, Vehicles, Polish)
```bash
./batch_ai_impl -tier 3
./batch_ai_impl -tier 4
./batch_ai_impl -tier 5
go build ./pkg/...
./goloco
```

**Result:** Game looks like OpenLoco! 🎉

---

## Quality Expectations

### Success Rate by Complexity
- **Low complexity:** ~95% success (color lookups, rect fills)
- **Medium complexity:** ~85% success (widget drawing, text)
- **High complexity:** ~70% success (viewport, paint session)
- **Very high complexity:** ~50% success (sprite rendering, vehicles)

### What GPT-5-mini Does Well
✅ Simple logic translation  
✅ Type conversions  
✅ Loop and conditional translation  
✅ Function calls and method dispatch  
✅ Error handling patterns  

### What Needs Manual Review
⚠️ Complex algorithms (may need tweaking)  
⚠️ Pointer arithmetic (Go doesn't have it)  
⚠️ Template specializations (Go doesn't have them)  
⚠️ Unsafe memory operations (use Go-safe alternatives)  

---

## Pro Tips

1. **Start with Tier 1** - Build confidence with simple functions
2. **Review Each File** - Don't blindly trust AI output
3. **Iterate on Failures** - Re-run failed functions with more context
4. **Add Type Stubs** - Define types as you encounter missing ones
5. **Test Incrementally** - Compile and test after each tier

---

## Cost Comparison

| Approach | Cost | Time |
|----------|------|------|
| Manual Implementation | $0 | 2-3 weeks |
| Direct OpenAI API | $1-20 | 2-3 days |
| **OpenCode (this!)** | **FREE!** | **2-3 days** |

Best of both worlds! 🎊

---

**Ready to generate Tier 1?**

```bash
cd goloco
go build ./cmd/batch_ai_impl
./batch_ai_impl -tier 1 -v
```

Let's make goloco look like OpenLoco with free AI! 🚂✨
