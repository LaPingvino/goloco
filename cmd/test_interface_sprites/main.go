package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/LaPingvino/goloco/pkg/assets"
	"github.com/LaPingvino/goloco/pkg/objects"
)

func main() {
	// Find Locomotion data directory
	dataDir := filepath.Join(os.Getenv("HOME"), ".local/share/Steam/steamapps/common/Locomotion")

	// Load G1 for palette
	g1Path := filepath.Join(dataDir, "Data", "g1.DAT")
	log.Printf("Loading G1 from: %s", g1Path)
	g1, err := assets.LoadG1(g1Path)
	if err != nil {
		log.Fatalf("Failed to load G1: %v", err)
	}
	log.Printf("Loaded G1 with %d sprites", g1.GetSpriteCount())

	// Load InterfaceSkin
	interfacePath := filepath.Join(dataDir, "ObjData", "INTERDEF.DAT")
	log.Printf("Loading InterfaceSkin from: %s", interfacePath)

	objMgr := objects.NewObjectManager("")
	objMgr.SetBaseSpriteIndex(uint32(g1.GetSpriteCount()))

	loaded, err := objMgr.LoadObject(interfacePath)
	if err != nil {
		log.Fatalf("Failed to load InterfaceSkin: %v", err)
	}

	skin, ok := loaded.Object.(*objects.InterfaceSkinObject)
	if !ok {
		log.Fatalf("Loaded object is not InterfaceSkinObject")
	}

	log.Printf("InterfaceSkin loaded: %s", skin.DisplayName)
	log.Printf("  Image offset: %d", skin.ImageOffset)
	log.Printf("  Image count: %d", skin.ImageCount)
	log.Printf("  Total sprites: %d", len(skin.Sprites))

	// Try to access toolbar sprites
	testSprites := []struct {
		name string
		id   uint32
	}{
		{"PreviewImage", objects.ISPreviewImage},
		{"ToolbarLoadsave", objects.ISToolbarLoadsave},
		{"ToolbarZoom", objects.ISToolbarZoom},
		{"ToolbarRotate", objects.ISToolbarRotate},
	}

	for _, test := range testSprites {
		if int(test.id) < len(skin.Sprites) {
			sprite := skin.Sprites[test.id]
			if sprite != nil {
				log.Printf("  %s (ID %d): %dx%d, flags=0x%04x, data len=%d",
					test.name, test.id, sprite.Width, sprite.Height, sprite.Flags, len(sprite.Data))
			} else {
				log.Printf("  %s (ID %d): sprite is nil", test.name, test.id)
			}
		} else {
			log.Printf("  %s (ID %d): out of range", test.name, test.id)
		}
	}
}
