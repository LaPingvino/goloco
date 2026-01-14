// Command cleanup_types fixes extracted type files
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	locoDir = flag.String("dir", "pkg/loco", "Directory containing extracted types")
)

func main() {
	flag.Parse()

	fmt.Println("🧹 Cleaning up extracted types...")

	err := filepath.Walk(*locoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		return cleanupFile(path)
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Cleanup complete!")
}

func cleanupFile(path string) error {
	fmt.Printf("  Cleaning: %s\n", path)

	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	cleaned := cleanContent(string(content))

	return os.WriteFile(path, []byte(cleaned), 0644)
}

var (
	cppCommentRe     = regexp.MustCompile(`(?m)^// (char|std::|return|func |//).*$`)
	invalidConstRe   = regexp.MustCompile(`(?m)^const.*ObjectHeader\{[^}]*\{.*$`)
	stringIdRe       = regexp.MustCompile(`\bStringId\b`)
	loadedObjectIdRe = regexp.MustCompile(`\bLoadedObjectId\b`)
	hexUSuffixRe     = regexp.MustCompile(`0x([0-9A-Fa-f]+)U\b`)
	numericLimitsRe  = regexp.MustCompile(`std\.numeric_limits<[^>]+>\(\)\.max\(\)`)
	charLiteralRe    = regexp.MustCompile(`'\\xFF'`)
)

func cleanContent(content string) string {
	lines := strings.Split(content, "\n")
	var cleaned []string
	inCppBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip C++ comment blocks
		if strings.HasPrefix(trimmed, "// //") {
			inCppBlock = true
			continue
		}
		if inCppBlock {
			if !strings.HasPrefix(trimmed, "//") {
				inCppBlock = false
			} else {
				continue
			}
		}

		// Skip specific C++ remnants
		if strings.HasPrefix(trimmed, "// char ") ||
			strings.HasPrefix(trimmed, "// std::") ||
			strings.HasPrefix(trimmed, "// return ") ||
			strings.HasPrefix(trimmed, "// func ") ||
			strings.HasPrefix(trimmed, "// See below") ||
			strings.HasPrefix(trimmed, "// If object") ||
			strings.HasPrefix(trimmed, "// when looking") ||
			strings.HasPrefix(trimmed, "// This means") ||
			strings.HasPrefix(trimmed, "// Most custom") ||
			strings.HasPrefix(trimmed, "// Use the isVanilla") ||
			strings.HasPrefix(trimmed, "// a lookup") ||
			strings.HasPrefix(trimmed, "// Note: Mods may") ||
			trimmed == "// //" {
			continue
		}

		// Skip invalid const declarations with complex initializers
		if strings.HasPrefix(trimmed, "const") && strings.Contains(trimmed, "ObjectHeader{") {
			continue
		}
		if strings.Contains(trimmed, "std.numeric_limits") {
			continue
		}

		// Fix type references
		line = stringIdRe.ReplaceAllString(line, "uint16")
		line = loadedObjectIdRe.ReplaceAllString(line, "uint32")
		line = hexUSuffixRe.ReplaceAllString(line, "0x$1")
		line = numericLimitsRe.ReplaceAllString(line, "0xFFFFFFFF")
		line = charLiteralRe.ReplaceAllString(line, "0xFF")

		cleaned = append(cleaned, line)
	}

	// Remove excessive blank lines
	result := make([]string, 0, len(cleaned))
	blankCount := 0

	for _, line := range cleaned {
		if strings.TrimSpace(line) == "" {
			blankCount++
			if blankCount <= 2 {
				result = append(result, line)
			}
		} else {
			blankCount = 0
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}
