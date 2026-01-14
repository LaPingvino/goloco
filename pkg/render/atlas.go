package render

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
)

// Atlas holds loaded images mapped by filename.
type Atlas struct {
	Images map[string]*ebiten.Image
	Sprites map[uint32]*ebiten.Image // sprite ID to image
}

// Global atlas instance
var globalAtlas *Atlas

// LoadAtlasFromDir loads all PNGs under dir (non-recursive) into an Atlas.
func LoadAtlasFromDir(dir string) (*Atlas, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	at := &Atlas{
		Images: make(map[string]*ebiten.Image),
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
	return a.Sprites[id]
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
