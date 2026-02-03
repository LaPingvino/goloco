package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/LaPingvino/goloco/pkg/objects"
)

func main() {
	// Find ObjData directory
	objDataPaths := []string{
		"/home/joop/locomotion/ObjData",
		"../locomotion/ObjData",
		"../../locomotion/ObjData",
	}

	var objDataPath string
	for _, p := range objDataPaths {
		if _, err := os.Stat(p); err == nil {
			objDataPath = p
			break
		}
	}

	if objDataPath == "" {
		fmt.Println("Could not find ObjData directory")
		os.Exit(1)
	}

	fmt.Printf("Loading objects from: %s\n\n", objDataPath)

	mgr := objects.NewObjectManager(objDataPath)

	// Load a few sample files first
	sampleFiles := []string{
		"114.DAT",    // A train (Class 114 DMU)
		"A4.DAT",     // A4 steam locomotive
		"747.DAT",    // Boeing 747
		"TRUCK1.DAT", // A truck
	}

	fmt.Println("=== Sample Objects ===")
	for _, name := range sampleFiles {
		path := filepath.Join(objDataPath, name)
		if _, err := os.Stat(path); err != nil {
			continue
		}

		obj, err := mgr.LoadObject(path)
		if err != nil {
			fmt.Printf("%-12s: ERROR - %v\n", name, err)
			continue
		}

		fmt.Printf("%-12s: Type=%s Name=%q\n",
			name, obj.Header.GetType(), obj.Header.GetName())

		if v, ok := obj.Object.(*objects.VehicleObject); ok {
			fmt.Printf("             Vehicle: %s %s, Power=%d, Speed=%.1f km/h\n",
				v.Mode, v.Type, v.Power, v.GetSpeedKmh())
			if v.DisplayName != "" {
				fmt.Printf("             Display: %s\n", v.DisplayName)
			}
		}
	}

	// Now load all and summarize
	fmt.Println("\n=== Loading All Objects ===")

	entries, _ := os.ReadDir(objDataPath)
	typeCounts := make(map[objects.ObjectType]int)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if len(name) < 4 || name[len(name)-4:] != ".DAT" {
			continue
		}

		path := filepath.Join(objDataPath, name)
		obj, err := mgr.LoadObject(path)
		if err != nil {
			continue
		}
		typeCounts[obj.Header.GetType()]++
	}

	fmt.Println("\nObject counts by type:")
	for t := objects.ObjectType(0); t < objects.MaxObjectTypes; t++ {
		if count := typeCounts[t]; count > 0 {
			fmt.Printf("  %-15s: %d\n", t, count)
		}
	}

	// List first 10 vehicles
	fmt.Printf("\n=== First 10 Vehicles (%d total) ===\n", len(mgr.Vehicles))
	for i, v := range mgr.Vehicles {
		if i >= 10 {
			break
		}
		fmt.Printf("  %s\n", v)
	}
}
