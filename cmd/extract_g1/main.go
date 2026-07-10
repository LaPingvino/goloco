package main

import (
	"flag"
	"fmt"
	"image/png"
	"os"
	"path/filepath"

	"github.com/LaPingvino/goloco/pkg/assets"
)

func main() {
	// Flags
	infoStart := flag.Int("info-start", -1, "Start index for info display")
	infoEnd := flag.Int("info-end", -1, "End index for info display")
	extractStart := flag.Int("start", 0, "Start index for extraction")
	extractEnd := flag.Int("end", 500, "End index for extraction")
	flag.Parse()

	// Default G1 path - try multiple locations
	g1Paths := []string{
		"/home/joop/locomotion/Data/g1.DAT",
		"../locomotion/Data/g1.DAT",
		"../../locomotion/Data/g1.DAT",
	}

	var g1Path string
	for _, p := range g1Paths {
		if _, err := os.Stat(p); err == nil {
			g1Path = p
			break
		}
	}

	if g1Path == "" {
		fmt.Println("Could not find g1.DAT")
		os.Exit(1)
	}

	fmt.Printf("Loading G1 from: %s\n", g1Path)

	g1, err := assets.LoadG1(g1Path)
	if err != nil {
		fmt.Printf("Failed to load G1: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded %d sprites\n", g1.GetSpriteCount())

	// Info mode - just display sprite info
	if *infoStart >= 0 && *infoEnd >= 0 {
		fmt.Printf("\nSprite info for range %d-%d:\n", *infoStart, *infoEnd)
		fmt.Printf("%-6s %-6s %-6s %-8s %-8s %-10s %-8s\n", "Index", "Width", "Height", "XOff", "YOff", "Flags", "DataLen")
		fmt.Println("----------------------------------------------------------------------")
		for i := *infoStart; i <= *infoEnd && i < g1.GetSpriteCount(); i++ {
			elem := &g1.Elements[i]
			flagStr := ""
			if elem.Flags&assets.G1FlagHasTransparency != 0 {
				flagStr += "TRANS "
			}
			if elem.Flags&assets.G1FlagIsRLECompressed != 0 {
				flagStr += "RLE "
			}
			if elem.Flags&assets.G1FlagIsR8G8B8Palette != 0 {
				flagStr += "PAL "
			}
			fmt.Printf("%-6d %-6d %-6d %-8d %-8d %-10s %-8d\n",
				i, elem.Width, elem.Height, elem.XOffset, elem.YOffset, flagStr, len(elem.Data))
		}
		return
	}

	// Extract mode
	outDir := "extracted_sprites"
	os.MkdirAll(outDir, 0755)

	extracted := 0
	skipped := 0
	for i := *extractStart; i < *extractEnd && i < g1.GetSpriteCount(); i++ {
		elem := &g1.Elements[i]

		// Skip empty or invalid sprites
		if elem.Width <= 0 || elem.Height <= 0 || len(elem.Data) == 0 {
			skipped++
			continue
		}

		// Skip palette elements (they're not regular sprites)
		if elem.Flags&assets.G1FlagIsR8G8B8Palette != 0 {
			skipped++
			continue
		}

		img, err := g1.DecodeSprite(i)
		if err != nil {
			fmt.Printf("Sprite %d: decode error: %v\n", i, err)
			skipped++
			continue
		}

		// Save as PNG
		outPath := filepath.Join(outDir, fmt.Sprintf("sprite_%04d.png", i))
		f, err := os.Create(outPath)
		if err != nil {
			fmt.Printf("Sprite %d: create file error: %v\n", i, err)
			continue
		}

		if err := png.Encode(f, img); err != nil {
			fmt.Printf("Sprite %d: encode error: %v\n", i, err)
			f.Close()
			skipped++
			continue
		}
		f.Close()
		extracted++
	}

	fmt.Printf("\nExtracted: %d sprites\n", extracted)
	fmt.Printf("Skipped: %d sprites\n", skipped)
	fmt.Printf("Output directory: %s\n", outDir)
}
