package render

import (
	"fmt"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/LaPingvino/goloco/pkg/assets"
	"github.com/hajimehoshi/ebiten/v2"
)

// Atlas holds loaded images mapped by filename.
type Atlas struct {
	Images  map[string]*ebiten.Image
	Sprites map[uint32]*ebiten.Image // sprite ID to image
	G1      *assets.G1File           // G1 file for on-demand loading
}

// Global atlas instance
var globalAtlas *Atlas

// Track which sprite IDs have been logged to avoid spam
var atlasSpriteLogged = make(map[uint32]bool)

// LoadAtlasFromDir loads all PNGs under dir (non-recursive) into an Atlas.
func LoadAtlasFromDir(dir string) (*Atlas, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	at := &Atlas{
		Images:  make(map[string]*ebiten.Image),
		Sprites: make(map[uint32]*ebiten.Image),
	}
	for _, fi := range files {
		if fi.IsDir() {
			continue
		}
		if filepath.Ext(fi.Name()) != ".png" {
			continue
		}
		fpath := filepath.Join(dir, fi.Name())
		f, err := os.Open(fpath)
		if err != nil {
			return nil, err
		}
		img, err := png.Decode(f)
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("decode %s: %w", fpath, err)
		}
		e := ebiten.NewImageFromImage(img)
		at.Images[fi.Name()] = e

		// Extract sprite ID from filename (sprite_<n>.png -> ID n). The number
		// may be any width: extract_g1 pads to 4 digits but grows past 9999.
		if base, ok := strings.CutPrefix(fi.Name(), "sprite_"); ok {
			if numStr, ok := strings.CutSuffix(base, ".png"); ok {
				if num, err := strconv.Atoi(numStr); err == nil && num >= 0 {
					at.Sprites[uint32(num)] = e
				}
			}
		}
	}

	// Set as global atlas
	globalAtlas = at
	return at, nil
}

// Get returns the ebiten image for the given filename, or nil if not found.
func (a *Atlas) Get(name string) *ebiten.Image {
	if a == nil {
		return nil
	}
	return a.Images[name]
}

// GetSprite returns the ebiten image for the given sprite ID, or nil if not found.
func (a *Atlas) GetSprite(id uint32) *ebiten.Image {
	if a == nil {
		return nil
	}
	// Check cache first
	if img, ok := a.Sprites[id]; ok {
		// Found cached sprite
		return img
	}
	// Try to load from G1
	if a.G1 != nil {
		if int(id) < len(a.G1.Elements) {
			e := &a.G1.Elements[id]
			if rgba := e.ToRGBA(a.G1.Palette); rgba != nil {
				img := ebiten.NewImageFromImage(rgba)
				a.Sprites[id] = img // cache it
				// Log only first few sprite decodes to avoid spam
				if !atlasSpriteLogged[id] && len(atlasSpriteLogged) < 5 {
					log.Printf("[Atlas] Decoded sprite from G1 id=%d size=%dx%d", id, rgba.Rect.Dx(), rgba.Rect.Dy())
					atlasSpriteLogged[id] = true
				}
				return img
			}
			if !atlasSpriteLogged[id] {
				log.Printf("[Atlas] G1 element present but ToRGBA returned nil id=%d", id)
				atlasSpriteLogged[id] = true
			}
		} else {
			if !atlasSpriteLogged[id] {
				log.Printf("[Atlas] G1 present but id out of range id=%d len=%d", id, len(a.G1.Elements))
				atlasSpriteLogged[id] = true
			}
		}
	} else {
		if !atlasSpriteLogged[id] {
			log.Printf("[Atlas] No G1 attached to atlas while looking up id=%d", id)
			atlasSpriteLogged[id] = true
		}
	}
	// Not found - log only once per sprite ID
	if !atlasSpriteLogged[id] {
		log.Printf("[Atlas] Sprite not found id=%d", id)
		atlasSpriteLogged[id] = true
	}
	return nil
}

// GetGlobalSprite returns a sprite from the global atlas
func GetGlobalSprite(id uint32) *ebiten.Image {
	if globalAtlas == nil {
		return nil
	}
	return globalAtlas.GetSprite(id)
}

// SetGlobalAtlas sets the global atlas instance
func SetGlobalAtlas(a *Atlas) {
	globalAtlas = a
}
